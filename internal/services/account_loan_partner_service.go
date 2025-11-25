package services

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"bitbucket.org/Amartha/go-accounting/internal/models"
	"bitbucket.org/Amartha/go-accounting/internal/repositories/mysql"
	xlog "bitbucket.org/Amartha/go-x/log"
	"github.com/iancoleman/strcase"
)

const (
	keyAdminFee                           = "adminfee"
	keyCashInTransitDisbursementDeduction = "cashintransitdisbursededuction"
	keyCashInTransitRepayment             = "cashintransitrepayment"
	keyAmarthaRevenue                     = "amartharevenue"
	keyWHT2326                            = "wht23_26"
	keyVATOut                             = "vatout"
)

/*

1. get latest product code
2. call create product type

3. get latest sub category code (create new sql in repo sub category)
4. call create sub category

5. process generate account & mapping loan partner account
6. call bulk insert acct_accounts & accounts
7. call bulk insert loan partner accounts

*/

func (as *account) CreateLoanPartnerAccount(ctx context.Context, in models.CreateAccountLoanPartner) (out models.AccountsLoanPartner, err error) {
	var newAccounts []models.CreateAccount

	if err := as.srv.mySqlRepo.Atomic(ctx, func(actx context.Context, r mysql.SQLRepository) (err error) {
		// get latest product code and create new product type
		newProductTypeCode, err := as.createProductType(actx, in)
		if err != nil {
			return
		}

		// get latest sub categ and create new sub category
		newSubCategCode, err := as.createSubCategory(actx, in, newProductTypeCode)
		if err != nil {
			return
		}

		// generate and bulk insert account & loan partner account
		accountNumbers, accounts, err := as.generateLoanPartnerAccount(actx, in, newProductTypeCode, newSubCategCode)
		if err != nil {
			return
		}
		newAccounts = append(newAccounts, accounts...)

		out = models.AccountsLoanPartner{
			AccountNumbers: accountNumbers,
			PartnerName:    in.PartnerName,
			PartnerId:      in.PartnerId,
			Metadata:       in.Metadata,
		}
		return

	}); err != nil {
		return out, err
	}

	// publish event account creation to acuan
	as.publishAccount(ctx, newAccounts)
	if errDelCache := as.srv.cacheRepo.DeleteKeysWithPrefix(ctx, prefixKeyCacheLoanPartner); errDelCache != nil {
		xlog.Warn(ctx, "[DELETE CACHING]", xlog.Any("prefix", prefixKeyCacheLoanPartner), xlog.Err(errDelCache))
	}
	return out, nil
}

func (as *account) createProductType(ctx context.Context, in models.CreateAccountLoanPartner) (productTypeCode string, err error) {
	latestProduct, err := as.srv.mySqlRepo.GetProductTypeRepository().GetLatestProductCode(ctx)
	if err != nil {
		return productTypeCode, err
	}

	latestProductCode, err := strconv.Atoi(latestProduct)
	if err != nil {
		return productTypeCode, err
	}
	newProductCode := latestProductCode + 1
	productTypeCode = strconv.Itoa(newProductCode)

	productTypeRequest := models.CreateProductTypeRequest{
		Code:   productTypeCode,
		Name:   in.PartnerName,
		Status: models.AccountStatusActive,
	}
	newProductType, err := as.srv.ProductType.Create(ctx, productTypeRequest)
	if err != nil {
		return
	}

	return newProductType.Code, nil
}

func (as *account) createSubCategory(ctx context.Context, in models.CreateAccountLoanPartner, productTypeCode string) (newSubCategCode string, err error) {
	latestSubCateg, err := as.srv.mySqlRepo.GetSubCategoryRepository().GetLatestSubCategCode(ctx, models.CategoryCode131)
	if err != nil {
		return
	}

	latestSubCategCode, err := strconv.Atoi(latestSubCateg)
	if err != nil {
		return
	}
	subCategCode := latestSubCategCode + 1
	newSubCategCode = strconv.Itoa(subCategCode)

	partnerName := fmt.Sprintf("Borrower Outstanding - %s", in.PartnerName)
	subCategRequest := models.CreateSubCategory{
		Code:            newSubCategCode,
		Name:            partnerName,
		AccountType:     loanAccountType,
		CategoryCode:    models.CategoryCode131,
		Currency:        models.CurrencyIDR,
		ProductTypeCode: productTypeCode,
		Description:     partnerName,
		Status:          models.StatusActive,
	}
	_, err = as.srv.SubCategory.Create(ctx, subCategRequest)
	if err != nil {
		return
	}
	return
}

func (as *account) generateLoanPartnerAccount(ctx context.Context, in models.CreateAccountLoanPartner, productType, newSubCategCode string) (accountNumbers []models.LoanPartnerAccountNumbers, accounts []models.CreateAccount, err error) {
	var (
		accountsLoanPartner []models.LoanPartnerAccount
	)

	in.LoanKind = strings.ToUpper(in.LoanKind)

	for _, entity := range as.srv.conf.AccountConfig.LoanPartnerAccountEntities {
		account := models.LoanPartnerAccountNumbers{
			Entity: entity,
		}
		for key, v := range as.srv.conf.AccountConfig.LoanPartnerAccountConfig {
			newAccountEntity := models.CreateAccount{
				Name:            fmt.Sprintf("%s %s", v.Name, in.PartnerName),
				CategoryCode:    v.CategoryCode,
				SubCategoryCode: v.SubCategoryCode,
				OwnerID:         strcase.ToCamel(in.PartnerId),
				EntityCode:      entity,
				ProductTypeCode: productType,
				Currency:        v.Currency,
				Metadata:        in.Metadata,
			}
			var newAccountNumber string

			//for cash in transit repayment, will use static account based on its entity
			if key == keyCashInTransitRepayment {

				newAccountNumber = as.srv.conf.AccountConfig.CashInTransitRepaymentEntity[entity]
			} else {
				newAccount, err := as.generateOtherAccount(ctx, newAccountEntity, v.SubCategoryCode)
				if err != nil {
					return nil, nil, err
				}
				newAccount.ProductTypeCode = productType
				accounts = append(accounts, newAccount)
				newAccountNumber = newAccount.AccountNumber
			}

			accountsLoanPartner = append(accountsLoanPartner, models.LoanPartnerAccount{
				PartnerId:           in.PartnerId,
				LoanKind:            in.LoanKind,
				AccountNumber:       newAccountNumber,
				AccountType:         v.AccountType,
				EntityCode:          entity,
				LoanSubCategoryCode: newSubCategCode,
			})

			loanPartnerAccounts, err := as.srv.mySqlRepo.GetLoanPartnerAccountRepository().GetByParams(ctx, models.GetLoanPartnerAccountByParamsIn{
				PartnerId:   in.PartnerId,
				LoanKind:    in.LoanKind,
				AccountType: v.AccountType,
				EntityCode:  entity,
			})
			if err != nil {
				return nil, nil, err
			}
			if len(loanPartnerAccounts) > 0 {
				err = models.GetErrMap(models.ErrKeyDataIsExist)
				return nil, nil, err
			}

			mappingAccountsLoanPartnerResponse(&account, key, newAccountNumber)
		}

		accountNumbers = append(accountNumbers, account)
	}

	if len(accounts) > 0 {

		if err := as.srv.mySqlRepo.GetAccountRepository().BulkInsertAcctAccount(ctx, accounts); err != nil {
			return nil, nil, err
		}
		if err := as.srv.mySqlRepo.GetAccountRepository().BulkInsertAccount(ctx, accounts); err != nil {
			return nil, nil, err
		}
		if err := as.srv.mySqlRepo.GetLoanPartnerAccountRepository().BulkInsertLoanPartnerAccount(ctx, accountsLoanPartner); err != nil {
			return nil, nil, err
		}
	}

	return accountNumbers, accounts, nil
}

func mappingAccountsLoanPartnerResponse(res *models.LoanPartnerAccountNumbers, key, accountNumber string) {
	switch key {
	case keyAdminFee:
		res.AdminFee = accountNumber
	case keyCashInTransitDisbursementDeduction:
		res.CashInTransitDisburseDeduction = accountNumber
	case keyCashInTransitRepayment:
		res.CashInTransitRepayment = accountNumber
	case keyAmarthaRevenue:
		res.AmarthaRevenue = accountNumber
	case keyWHT2326:
		res.WHT2326 = accountNumber
	case keyVATOut:
		res.VATOut = accountNumber
	}
}
