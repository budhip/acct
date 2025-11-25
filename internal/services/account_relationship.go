package services

import (
	"context"
	"errors"
	"fmt"

	"bitbucket.org/Amartha/go-accounting/internal/models"
	"bitbucket.org/Amartha/go-accounting/internal/repositories/mysql"
	xlog "bitbucket.org/Amartha/go-x/log"
)

func (as *account) ConsumerCreateAccountRelationship(ctx context.Context, in models.CreateAccount) (err error) {
	var accounts []models.CreateAccount
	if in.AccountType == "ACCOUNTS" {
		accounts = append(accounts, in)
		if err = as.srv.mySqlRepo.GetAccountRepository().BulkInsertAccount(ctx, accounts); err != nil {
			xlog.Warn(ctx, "[ACCOUNT-RELATIONSHIPS]", xlog.Any("message", in), xlog.Err(err))
			as.srv.Account.publishToAccountMigrationStreamDLQ(ctx, in, err)
			return
		}
	} else if in.AccountType == "CREATE_LOAN_ADVANCE" {
		if err = as.CreateLoanAdvanceAccount(ctx, in); err != nil {
			xlog.Warn(ctx, "[ACCOUNT-RELATIONSHIPS]", xlog.Any("message", in), xlog.Err(err))
			as.srv.Account.publishToAccountMigrationStreamDLQ(ctx, in, err)
			return
		}
	} else {
		if err = as.CreateAccountRelationship(ctx, in); err != nil {
			xlog.Warn(ctx, "[ACCOUNT-RELATIONSHIPS]", xlog.Any("message", in), xlog.Err(err))
			as.srv.Account.publishToAccountMigrationStreamDLQ(ctx, in, err)
			return
		}
	}
	return
}

func (as *account) CreateAccountRelationship(ctx context.Context, in models.CreateAccount) (err error) {
	var (
		newAccounts   = make([]models.CreateAccount, 0, 3)
		lenderAccount models.CreateLenderAccount
		loanAccount   models.CreateLoanAccount
	)

	accountNumber := in.AccountNumber

	account, err := as.srv.mySqlRepo.GetAccountRepository().GetOneByAccountNumber(ctx, accountNumber)
	if err != nil {
		err = checkDatabaseError(err, models.ErrKeyAccountNumberNotFound)
		return
	}

	// generate account number if have invested account
	invested, ok := as.srv.conf.AccountConfig.InvestedAccountNumber[in.SubCategoryCode]
	if ok {
		result, err := as.srv.mySqlRepo.GetAccountRepository().GetAllAccountNumbersByParam(ctx, models.GetAllAccountNumbersByParamIn{
			OwnerId:         in.OwnerID,
			SubCategoryCode: invested,
		})
		if err != nil {
			err = checkDatabaseError(err)
			return err
		}
		accounts := result

		var investedAccount models.CreateAccount
		lenderAccount.CIHAccountNumber = accountNumber
		if len(accounts) == 0 {
			investedAccount, err = as.generateOtherAccount(ctx, in, invested)
			if err != nil {
				return err
			}
			lenderAccount.InvestedAccountNumber = investedAccount.AccountNumber
			newAccounts = append(newAccounts, investedAccount)
		} else if len(accounts) == 1 {
			lenderAccount.InvestedAccountNumber = accounts[0].AccountNumber
		} else {
			date := account.CreatedAt
			for _, v := range accounts {
				if date.Before(v.CreatedAt) || date.Equal(v.CreatedAt) {
					lenderAccount.InvestedAccountNumber = v.AccountNumber
					break
				}
			}
			if lenderAccount.InvestedAccountNumber == "" {
				investedAccount, err = as.generateOtherAccount(ctx, in, invested)
				if err != nil {
					return err
				}
				lenderAccount.InvestedAccountNumber = investedAccount.AccountNumber
				newAccounts = append(newAccounts, investedAccount)
			}
		}
	}

	// generate account number if have receivables account
	receivables, ok := as.srv.conf.AccountConfig.ReceivablesAccountNumber[in.SubCategoryCode]
	if ok {
		result, err := as.srv.mySqlRepo.GetAccountRepository().GetAllAccountNumbersByParam(ctx, models.GetAllAccountNumbersByParamIn{
			OwnerId:         in.OwnerID,
			SubCategoryCode: receivables,
		})
		if err != nil {
			err = checkDatabaseError(err, models.ErrKeyOwnerIdNotFound)
			return err
		}
		accounts := result

		var receivablesAccount models.CreateAccount
		lenderAccount.CIHAccountNumber = accountNumber
		if len(accounts) == 0 {
			receivablesAccount, err = as.generateOtherAccount(ctx, in, receivables)
			if err != nil {
				return err
			}
			lenderAccount.ReceivablesAccountNumber = receivablesAccount.AccountNumber
			newAccounts = append(newAccounts, receivablesAccount)
		} else if len(accounts) == 1 {
			lenderAccount.ReceivablesAccountNumber = accounts[0].AccountNumber
		} else {
			date := account.CreatedAt
			for _, v := range accounts {
				if date.Before(v.CreatedAt) || date.Equal(v.CreatedAt) {
					lenderAccount.ReceivablesAccountNumber = v.AccountNumber
					break
				}
			}
			if lenderAccount.ReceivablesAccountNumber == "" {
				receivablesAccount, err = as.generateOtherAccount(ctx, in, receivables)
				if err != nil {
					return err
				}
				lenderAccount.ReceivablesAccountNumber = receivablesAccount.AccountNumber
				newAccounts = append(newAccounts, receivablesAccount)
			}
		}
	}

	// generate account number if have loan account normal
	loanAdvancePayment, ok := as.srv.conf.AccountConfig.MultiLoanAccount[in.SubCategoryCode]
	if ok {
		result, err := as.srv.mySqlRepo.GetAccountRepository().GetAllAccountNumbersByParam(ctx, models.GetAllAccountNumbersByParamIn{
			OwnerId:         in.OwnerID,
			SubCategoryCode: loanAdvancePayment,
		})
		if err != nil {
			err = checkDatabaseError(err, models.ErrKeyOwnerIdNotFound)
			return err
		}
		accounts := result

		var loanAdvancePaymentAccount models.CreateAccount
		loanAccount.LoanAccountNumber = accountNumber
		if len(accounts) == 0 {
			loanAdvancePaymentAccount, err = as.generateOtherAccount(ctx, in, loanAdvancePayment)
			if err != nil {
				return err
			}
			loanAdvancePaymentAccount.AltId = fmt.Sprintf("%s-%s", accountNumber, in.AltId)
			loanAccount.LoanAdvancePaymentAccountNumber = loanAdvancePaymentAccount.AccountNumber
			newAccounts = append(newAccounts, loanAdvancePaymentAccount)
		} else if len(accounts) == 1 {
			loanAccount.LoanAdvancePaymentAccountNumber = accounts[0].AccountNumber
		} else {
			date := account.CreatedAt
			for _, v := range accounts {
				if date.Before(v.CreatedAt) || date.Equal(v.CreatedAt) {
					loanAccount.LoanAdvancePaymentAccountNumber = v.AccountNumber
					break
				}
			}
			if loanAccount.LoanAdvancePaymentAccountNumber == "" {
				loanAdvancePaymentAccount, err = as.generateOtherAccount(ctx, in, loanAdvancePayment)
				if err != nil {
					return err
				}
				loanAdvancePaymentAccount.AltId = fmt.Sprintf("%s-%s", accountNumber, in.AltId)
				loanAccount.LoanAdvancePaymentAccountNumber = loanAdvancePaymentAccount.AccountNumber
				newAccounts = append(newAccounts, loanAdvancePaymentAccount)
			}
		}
	}

	if err = as.srv.mySqlRepo.Atomic(ctx, func(actx context.Context, r mysql.SQLRepository) (err error) {
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

		if len(newAccounts) > 0 {
			if err = r.GetAccountRepository().BulkInsertAcctAccount(actx, newAccounts); err != nil {
				return
			}

			if err = r.GetAccountRepository().BulkInsertAccount(actx, newAccounts); err != nil {
				return
			}
		}

		return
	}); err != nil {
		return
	}

	if len(newAccounts) > 0 {
		// publish account creation
		as.publishAccount(ctx, newAccounts)
	}
	return
}

func (as *account) CreateLoanAdvanceAccount(ctx context.Context, in models.CreateAccount) (err error) {
	var (
		newAccounts = make([]models.CreateAccount, 0, 1)
		loanAccount models.CreateLoanAccount
	)

	account, err := as.srv.mySqlRepo.GetAccountRepository().GetOneByAccountNumber(ctx, in.AccountNumber)
	if err != nil {
		err = checkDatabaseError(err, models.ErrKeyAccountNumberNotFound)
		return
	}

	_, err = as.srv.mySqlRepo.GetAccountRepository().GetLoanAdvanceAccountByLoanAccount(ctx, account.AccountNumber)
	if err != nil && !errors.Is(err, models.ErrNoRows) {
		return
	}

	if errors.Is(err, models.ErrNoRows) {
		in.Name = account.AccountName
		in.AccountType = account.AccountType
		in.EntityCode = account.EntityCode
		in.OwnerID = account.OwnerID

		var loanAdvancePaymentAccount models.CreateAccount
		loanAccount.LoanAccountNumber = in.AccountNumber
		loanAdvancePaymentAccount, err = as.generateOtherAccount(ctx, in, "21303")
		if err != nil {
			return err
		}
		loanAdvancePaymentAccount.AltId = account.AccountNumber
		if account.AltID != "" {
			loanAdvancePaymentAccount.AltId = fmt.Sprintf("%s-%s", account.AccountNumber, account.AltID)
		}
		loanAccount.LoanAdvancePaymentAccountNumber = loanAdvancePaymentAccount.AccountNumber
		newAccounts = append(newAccounts, loanAdvancePaymentAccount)

		if err = as.srv.mySqlRepo.Atomic(ctx, func(actx context.Context, r mysql.SQLRepository) (err error) {
			if len(newAccounts) > 0 {
				if err = r.GetAccountRepository().BulkInsertAcctAccount(actx, newAccounts); err != nil {
					return
				}

				if err = r.GetAccountRepository().BulkInsertAccount(actx, newAccounts); err != nil {
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

		if len(newAccounts) > 0 {
			// publish account creation
			as.publishAccount(ctx, newAccounts)
		}
	}

	return
}
