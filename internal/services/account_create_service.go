package services

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"bitbucket.org/Amartha/go-accounting/internal/models"
	"bitbucket.org/Amartha/go-accounting/internal/pkg/acuanclient"
	"bitbucket.org/Amartha/go-accounting/internal/repositories/mysql"
	xlog "bitbucket.org/Amartha/go-x/log"
)

/*
1. check create account with account type or not and validate request
2. get last sequence by category code
3. generate account number
4. if create account with account type
- check invested and receivables account
- get last sequence by category code
- generate account number
5. bulk Insert into accounts and acct_account
6. if account have invested and receivables account insert into acct_lender_account
7. publish account to kafka
- if create account with account type and request metadata not empty and legacy id is empty then publish it to account_stream_t24 after that account will publish to acuan
- if create account with account type and request legacy id is not empty then publish it to acuan
- other than that publish it to acuan
8. transform to response
*/

var loanAccountType = "PARTNERSHIP_LOAN"

func (as *account) Create(ctx context.Context, in models.CreateAccount) (out models.CreateAccount, err error) {
	var (
		accounts      = make([]models.CreateAccount, 0, 3)
		lenderAccount models.CreateLenderAccount
		loanAccount   models.CreateLoanAccount
		legacyID      models.LegacyID
	)

	defer func() {
		logService(ctx, err)
	}()

	// validate request
	_, err = as.validateInput(ctx, &in)
	if err != nil {
		return out, err
	}

	if in.LegacyId != nil {
		legacy, legacyErr := in.LegacyId.Value()
		if legacyErr != nil {
			err = models.GetErrMap(models.ErrKeyFailedMarshal, legacyErr.Error())
			return out, err
		}
		if err = json.Unmarshal(legacy.([]byte), &legacyID); err != nil {
			err = models.GetErrMap(models.ErrKeyFailedUnmarshal, err.Error())
			return out, err
		}
		in.AccountNumber = legacyID.T24AccountNumber
	}

	if in.AccountNumber == "" || in.AccountNumber == "0" {
		// get last sequence
		lastSequence, lastSequenceErr := as.getCategoryCodeSeq(ctx, in.CategoryCode)
		if lastSequenceErr != nil {
			err = lastSequenceErr
			return out, err
		}
		// generate account number
		in.AccountNumber, err = generateAccountNumber(
			in.CategoryCode,
			in.EntityCode,
			as.srv.conf.AccountConfig.AccountNumberPadWidth,
			lastSequence,
		)
		if err != nil {
			return out, err
		}
	}

	in.Status = models.MapAccountStatus[models.ACCOUNT_STATUS_ACTIVE]
	accounts = append(accounts, in)
	out = in

	/*
		we add validation here to handle the case the account already
		created on point this api is called. it happen because the account
		creation in t24 exist first and the consumer create the account,
		the same account being used to call this api. we need to revisit
		again after the t24 turn off.
	*/

	accData, err := as.srv.mySqlRepo.GetAccountRepository().GetOneByAccountNumber(ctx, out.AccountNumber)
	if err == nil {
		if out.AltId != "" {
			if err = as.srv.mySqlRepo.GetAccountRepository().UpdateAltId(ctx, models.UpdateAltId{
				AltId:         out.AltId,
				AccountNumber: out.AccountNumber,
			}); err != nil {
				xlog.Error(ctx, "[CREATE-ACCOUNT]", xlog.String("status", "failed to update altId"), xlog.String("alt_id", out.AltId), xlog.String("account_number", out.AccountNumber), xlog.Err(err))
			} else {
				accData.AltID = out.AltId
			}
		}
		return models.CreateAccount{
			AccountNumber:   accData.AccountNumber,
			OwnerID:         accData.OwnerID,
			AccountType:     accData.AccountType,
			ProductTypeCode: accData.ProductTypeCode,
			EntityCode:      accData.EntityCode,
			CategoryCode:    accData.CategoryCode,
			SubCategoryCode: accData.SubCategoryCode,
			Currency:        accData.Currency,
			Status:          accData.Status,
			Name:            accData.AccountName,
			AltId:           accData.AltID,
			LegacyId:        accData.LegacyId,
			Metadata:        accData.Metadata,
			ProductTypeName: accData.ProductTypeName,
		}, err
	}

	// return error if got database error
	err = checkDatabaseError(err, models.ErrKeyAccountNumberNotFound)
	if err != models.GetErrMap(models.ErrKeyAccountNumberNotFound) {
		return
	}

	// generate account number if have invested account
	invested, ok := as.srv.conf.AccountConfig.InvestedAccountNumber[in.SubCategoryCode]
	if ok {
		xlog.Info(ctx, "invested", xlog.String("sub-category-code", invested))
		var investedAccount models.CreateAccount
		investedAccount, err = as.generateOtherAccount(ctx, in, invested)
		if err != nil {
			return out, err
		}
		lenderAccount.CIHAccountNumber = in.AccountNumber
		lenderAccount.InvestedAccountNumber = investedAccount.AccountNumber
		accounts = append(accounts, investedAccount)
	}

	// generate account number if have receivables account
	receivables, ok := as.srv.conf.AccountConfig.ReceivablesAccountNumber[in.SubCategoryCode]
	if ok {
		xlog.Info(ctx, "receivables", xlog.String("sub-category-code", receivables))
		var receivablesAccount models.CreateAccount
		receivablesAccount, err = as.generateOtherAccount(ctx, in, receivables)
		if err != nil {
			return out, err
		}
		lenderAccount.CIHAccountNumber = in.AccountNumber
		lenderAccount.ReceivablesAccountNumber = receivablesAccount.AccountNumber
		accounts = append(accounts, receivablesAccount)
	}

	// generate account number if have loan account normal
	loanAdvancePayment, ok := as.srv.conf.AccountConfig.MultiLoanAccount[in.SubCategoryCode]
	if ok {
		xlog.Info(ctx, "loanAdvancePayment", xlog.String("sub-category-code", loanAdvancePayment))
		var loanAdvancePaymentAccount models.CreateAccount
		loanAdvancePaymentAccount, err = as.generateOtherAccount(ctx, in, loanAdvancePayment)
		if err != nil {
			return out, err
		}
		loanAdvancePaymentAccount.AltId = in.AccountNumber
		if in.AltId != "" {
			loanAdvancePaymentAccount.AltId = fmt.Sprintf("%s-%s", in.AccountNumber, in.AltId)
		}
		loanAccount.LoanAccountNumber = in.AccountNumber
		loanAccount.LoanAdvancePaymentAccountNumber = loanAdvancePaymentAccount.AccountNumber
		accounts = append(accounts, loanAdvancePaymentAccount)
	}

	if err = as.srv.mySqlRepo.Atomic(ctx, func(actx context.Context, r mysql.SQLRepository) (err error) {
		if err = r.GetAccountRepository().BulkInsertAcctAccount(actx, accounts); err != nil {
			return
		}

		if err = r.GetAccountRepository().BulkInsertAccount(actx, accounts); err != nil {
			return
		}

		if lenderAccount.CIHAccountNumber != "" {
			if err = r.GetAccountRepository().CreateLenderAccount(actx, lenderAccount); err != nil {
				return
			}
		}

		if loanAccount.LoanAccountNumber != "" {
			if err = r.GetAccountRepository().CreateLoanAccount(actx, loanAccount); err != nil {
				return
			}
		}
		return
	}); err != nil {
		return
	}

	// publish account creation to acuan
	as.publishAccount(ctx, accounts)

	// delete caching
	keys := deletePasAccountsKey(models.GetAllAccountNumbersByParamIn{
		OwnerId:         in.OwnerID,
		AltId:           in.AltId,
		SubCategoryCode: in.SubCategoryCode,
		AccountType:     in.AccountType,
	})
	as.deleteCaching(ctx, keys)

	return
}

func (as *account) validateInput(ctx context.Context, in *models.CreateAccount) (bool, error) {
	isWithAccountType := in.AccountType != "" && in.AccountType != loanAccountType
	if isWithAccountType {
		out, err := as.srv.mySqlRepo.GetSubCategoryRepository().GetByAccountType(ctx, in.AccountType)
		if err != nil {
			err = checkDatabaseError(err, models.ErrKeyAccountTypeNotValid)
			return isWithAccountType, err
		}
		in.ProductTypeCode = out.ProductTypeCode
		in.CategoryCode = out.CategoryCode
		in.SubCategoryCode = out.Code
		in.Currency = out.Currency

		productType, err := as.srv.mySqlRepo.GetProductTypeRepository().GetByCode(ctx, in.ProductTypeCode)
		if err != nil {
			return isWithAccountType, err
		}
		if productType == nil {
			err = models.GetErrMap(models.ErrKeyProductTypeNotFound)
			return isWithAccountType, err
		}
		in.EntityCode = productType.EntityCode
		in.ProductTypeName = productType.Name
	} else if strings.ToUpper(in.AccountType) == loanAccountType {
		var metadataLoan models.MetadataLoan
		loanKind := loanAccountType
		metadata, err := in.Metadata.Value()
		if err != nil {
			err = models.GetErrMap(models.ErrKeyFailedMarshal, err.Error())
			return isWithAccountType, err
		}

		if err = json.Unmarshal(metadata.([]byte), &metadataLoan); err != nil {
			err = models.GetErrMap(models.ErrKeyFailedUnmarshal, err.Error())
			return isWithAccountType, err
		}

		if metadataLoan.PartnerId == "" {
			err = models.GetErrMap(models.ErrKeyPartnerIdRequired, "partnerId missing or empty value")
			return isWithAccountType, err
		}

		loanPartnerAccounts, err := as.srv.mySqlRepo.GetLoanPartnerAccountRepository().GetByParams(ctx, models.GetLoanPartnerAccountByParamsIn{
			PartnerId: metadataLoan.PartnerId,
			LoanKind:  loanKind,
		})
		if err != nil {
			return isWithAccountType, err
		}
		if len(loanPartnerAccounts) == 0 {
			err = models.GetErrMap(models.ErrKeyLoanPartnerAccountNotFound, fmt.Sprintf("partnerId:%v loanKind:%v", metadataLoan.PartnerId, loanKind))
			return isWithAccountType, err
		}

		subCategoryCode := loanPartnerAccounts[0].LoanSubCategoryCode
		subCategory, err := as.srv.mySqlRepo.GetSubCategoryRepository().GetByCode(ctx, subCategoryCode)
		if err != nil {
			return isWithAccountType, err
		}
		if subCategory == nil {
			err = models.GetErrMap(models.ErrKeySubCategoryCodeNotFound)
			return isWithAccountType, err
		}

		in.ProductTypeCode = subCategory.ProductTypeCode
		in.CategoryCode = subCategory.CategoryCode
		in.SubCategoryCode = subCategory.Code
		in.Currency = subCategory.Currency

		productType, err := as.srv.mySqlRepo.GetProductTypeRepository().GetByCode(ctx, in.ProductTypeCode)
		if err != nil {
			return isWithAccountType, err
		}
		if productType == nil {
			err = models.GetErrMap(models.ErrKeyLoanPartnerAccountNotFound)
			return isWithAccountType, err
		}
		in.EntityCode = productType.EntityCode
		in.ProductTypeName = productType.Name

	} else {
		if err := as.srv.mySqlRepo.GetCategoryRepository().CheckCategoryByCode(ctx, in.CategoryCode); err != nil {
			err = checkDatabaseError(err, models.ErrKeyCategoryCodeNotFound)
			return isWithAccountType, err
		}

		subCategory, err := as.srv.mySqlRepo.GetSubCategoryRepository().CheckSubCategoryByCodeAndCategoryCode(ctx, in.SubCategoryCode, in.CategoryCode)
		if err != nil {
			err = checkDatabaseError(err, models.ErrKeySubCategoryCodeNotFound)
			return isWithAccountType, err
		}
		in.AccountType = subCategory.AccountType

		if in.ProductTypeCode != "" {
			productType, err := as.srv.mySqlRepo.GetProductTypeRepository().GetByCode(ctx, in.ProductTypeCode)
			if err != nil {
				return isWithAccountType, err
			}
			if productType == nil {
				err = models.GetErrMap(models.ErrKeyProductTypeNotFound)
				return isWithAccountType, err
			}
			in.ProductTypeName = productType.Name
		}

		if err := as.srv.mySqlRepo.GetEntityRepository().CheckEntityByCode(ctx, in.EntityCode, models.StatusActive); err != nil {
			err = checkDatabaseError(err, models.ErrKeyEntityCodeNotFound)
			return isWithAccountType, err
		}

		if _, err := as.srv.goDBLedger.GetCurrency(ctx, in.Currency); err != nil {
			err = checkDatabaseError(err, models.ErrKeyCurrencyNotFound)
			return isWithAccountType, err
		}
	}

	in.Name = removeSpecialChars(in.Name)

	return isWithAccountType, nil
}

func (as *account) getCategoryCodeSeq(ctx context.Context, categoryCode string) (int64, error) {
	// get last sequence
	sequenceName := fmt.Sprintf("category_code_%s_seq", categoryCode)
	lastSequence, err := as.srv.cacheRepo.GetIncrement(ctx, sequenceName)
	if err != nil {
		return lastSequence, models.GetErrMap(models.ErrKeyFailedGetFromCache, err.Error())
	}
	return lastSequence, nil
}

func (as *account) generateOtherAccount(ctx context.Context, in models.CreateAccount, subCategoryCode string) (account models.CreateAccount, err error) {
	subCategory, err := as.srv.mySqlRepo.GetSubCategoryRepository().GetByCode(ctx, subCategoryCode)
	if err != nil {
		return
	}
	if subCategory == nil {
		err = models.GetErrMap(models.ErrKeySubCategoryCodeNotFound)
		return
	}

	// get last sequence
	lastSequence, err := as.getCategoryCodeSeq(ctx, subCategory.CategoryCode)
	if err != nil {
		return
	}

	// generate account number
	accountNumber, err := generateAccountNumber(
		subCategory.CategoryCode,
		in.EntityCode,
		as.srv.conf.AccountConfig.AccountNumberPadWidth,
		lastSequence,
	)
	if err != nil {
		return
	}

	currency := models.CurrencyIDR
	if subCategory.Currency != "" {
		currency = subCategory.Currency
	}
	account = models.CreateAccount{
		AccountNumber:   accountNumber,
		OwnerID:         in.OwnerID,
		AccountType:     subCategory.AccountType,
		ProductTypeCode: subCategory.ProductTypeCode,
		EntityCode:      in.EntityCode,
		CategoryCode:    subCategory.CategoryCode,
		SubCategoryCode: subCategory.Code,
		Currency:        currency,
		Status:          models.MapAccountStatus[models.ACCOUNT_STATUS_ACTIVE],
		Name:            in.Name,
		Metadata:        in.Metadata,
	}

	return
}

func (as *account) publishAccount(ctx context.Context, accounts []models.CreateAccount) {
	for _, v := range accounts {
		// publish event account creation to acuan
		as.publishAccountToAcuan(ctx, models.TypeAccountCreated, v)
	}
}

func (as *account) publishAccountToAcuan(ctx context.Context, messageType string, in models.CreateAccount) {
	as.srv.acuanClient.PublishAccount(ctx, acuanclient.PublishAccountData{
		Type:            messageType,
		AccountNumber:   in.AccountNumber,
		ProductTypeName: in.ProductTypeName,
		OwnerId:         in.OwnerID,
		CategoryCode:    in.CategoryCode,
		SubCategoryCode: in.SubCategoryCode,
		EntityCode:      in.EntityCode,
		Currency:        in.Currency,
		Status:          in.Status,
		Name:            in.Name,
		AltId:           in.AltId,
		LegacyId:        in.LegacyId,
		Metadata:        in.Metadata,
	})
}

func removeSpecialChars(input string) string {
	re := regexp.MustCompile(`[^a-zA-Z0-9\s.,/\-'()]`)
	input = re.ReplaceAllString(input, "")

	reSpace := regexp.MustCompile(`\s+`)
	return reSpace.ReplaceAllString(input, " ")
}
