package services

import (
	"context"

	"bitbucket.org/Amartha/go-accounting/internal/models"
	"bitbucket.org/Amartha/go-accounting/internal/repositories/mysql"
	xlog "bitbucket.org/Amartha/go-x/log"
)

/*
1. check legacy id if exist skipped
2. check account number if exist do update
3. validate request
4. generate account number if have invested account
4. generate account number if have receivables account
5. generate account number if have loan account normal
6. insert to database
8. publish account to acuan
*/
func (as *account) ConsumerCreateAccountMigration(ctx context.Context, in models.CreateAccount) (err error) {
	_, err = as.createAccountMigration(ctx, in)
	if err != nil {
		xlog.Warn(ctx, "[ACCOUNT-MIGRATION]", xlog.Any("message", in), xlog.Err(err))
		as.srv.Account.publishToAccountMigrationStreamDLQ(ctx, in, err)
		return
	}
	return
}

func (as *account) publishToAccountMigrationStreamDLQ(ctx context.Context, data models.CreateAccount, err error) error {
	messages := models.AccountError{
		CreateAccount: data,
		ErrCauser:     err,
	}
	return as.srv.publisher.PublishSyncWithKeyAndLog(ctx, "publish account to account_migration_stream.dlq", as.srv.conf.Kafka.Publishers.AccountMigrationStreamDLQ.Topic, data.AccountNumber, messages)
}

func (as *account) createAccountMigration(ctx context.Context, in models.CreateAccount) (out models.CreateAccount, err error) {
	var (
		accounts      = make([]models.CreateAccount, 0, 3)
		lenderAccount models.CreateLenderAccount
		loanAccount   models.CreateLoanAccount
	)

	defer func() {
		logService(ctx, err)
	}()

	isExistLegacyId, err := as.srv.mySqlRepo.GetAccountRepository().CheckLegacyIdIsExist(ctx, in.AccountNumber)
	if err != nil {
		err = checkDatabaseError(err)
		return
	}
	if isExistLegacyId {
		xlog.Info(ctx, "[ACCOUNT-MIGRATION]", xlog.String("status", "legacy id is exist"), xlog.Any("message", in))
		return
	}

	isExistAccountNumber, err := as.srv.mySqlRepo.GetAccountRepository().CheckAccountNumberIsExist(ctx, in.AccountNumber)
	if err != nil {
		err = checkDatabaseError(err)
		return
	}

	if isExistAccountNumber != nil {
		if err = as.srv.mySqlRepo.GetAccountRepository().Update(ctx, models.UpdateAccount{
			Name:          in.Name,
			OwnerID:       in.OwnerID,
			AltID:         in.AltId,
			LegacyId:      in.LegacyId,
			AccountNumber: in.AccountNumber,
		}); err != nil {
			err = checkDatabaseError(err)
			return
		}
		xlog.Info(ctx, "[ACCOUNT-MIGRATION]", xlog.String("status", "account number is exist do update"), xlog.Any("message", in))
		return
	}

	_, err = as.validateInput(ctx, &in)
	if err != nil {
		return out, err
	}

	accountNumber := in.AccountNumber
	in.Status = models.MapAccountStatus[models.ACCOUNT_STATUS_ACTIVE]
	accounts = append(accounts, in)
	out = in

	// generate account number if have invested account
	invested, ok := as.srv.conf.AccountConfig.InvestedAccountNumber[in.SubCategoryCode]
	if ok {
		lenderAccount.CIHAccountNumber = accountNumber
		account, err := as.generateOtherAccount(ctx, in, invested)
		if err != nil {
			return out, err
		}
		accounts = append(accounts, account)
		lenderAccount.InvestedAccountNumber = account.AccountNumber
	}

	// generate account number if have receivables account
	receivables, ok := as.srv.conf.AccountConfig.ReceivablesAccountNumber[in.SubCategoryCode]
	if ok {
		if in.SubCategoryCode == "21103" {
			lenderAccount.CIHAccountNumber = accountNumber
			account, err := as.generateOtherAccountReceivables(ctx, in, receivables)
			if err != nil {
				return out, err
			}
			accounts = append(accounts, account)
			lenderAccount.ReceivablesAccountNumber = account.AccountNumber
		} else {
			lenderAccount.CIHAccountNumber = accountNumber
			account, err := as.generateOtherAccount(ctx, in, receivables)
			if err != nil {
				return out, err
			}
			accounts = append(accounts, account)
			lenderAccount.ReceivablesAccountNumber = account.AccountNumber
		}
	}

	// generate account number if have loan account normal
	loanAdvancePayment, ok := as.srv.conf.AccountConfig.MultiLoanAccount[in.SubCategoryCode]
	if ok {
		loanAccount.LoanAccountNumber = accountNumber
		account, err := as.generateOtherAccount(ctx, in, loanAdvancePayment)
		if err != nil {
			return out, err
		}
		accounts = append(accounts, account)
		loanAccount.LoanAdvancePaymentAccountNumber = account.AccountNumber
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

	for _, v := range accounts {
		// publish event account creation to acuan
		as.publishAccountToAcuan(ctx, "migration_database", v)
	}

	xlog.Info(ctx, "[ACCOUNT-MIGRATION]", xlog.String("status", "success"), xlog.Any("message", in))

	return
}

func (as *account) generateOtherAccountReceivables(ctx context.Context, in models.CreateAccount, subCategoryCode string) (account models.CreateAccount, err error) {
	var lastSequence int64
	subCategory, err := as.srv.mySqlRepo.GetSubCategoryRepository().GetByCode(ctx, subCategoryCode)
	if err != nil {
		return
	}
	if subCategory == nil {
		err = models.GetErrMap(models.ErrKeySubCategoryCodeNotFound)
		return
	}

	accountNumber, ok := as.srv.conf.AccountConfig.LenderInstiReceivablesAccount[in.AccountNumber]
	if !ok {
		// get last sequence
		lastSequence, err = as.getCategoryCodeSeq(ctx, subCategory.CategoryCode)
		if err != nil {
			return account, err
		}

		// generate account number
		accountNumber, err = generateAccountNumber(
			subCategory.CategoryCode,
			in.EntityCode,
			as.srv.conf.AccountConfig.AccountNumberPadWidth,
			lastSequence,
		)
		if err != nil {
			return account, err
		}
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
	}

	return
}
