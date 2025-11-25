package services

import (
	"context"
	"reflect"

	"bitbucket.org/Amartha/go-accounting/internal/models"
	"bitbucket.org/Amartha/go-accounting/internal/repositories/cache"
	xlog "bitbucket.org/Amartha/go-x/log"
)

type LoanPartnerService interface {
	Create(ctx context.Context, in models.LoanPartnerAccount) (out models.LoanPartnerAccount, err error)
	Update(ctx context.Context, in models.UpdateLoanPartnerAccount) (out models.UpdateLoanPartnerAccount, err error)
	GetByParams(ctx context.Context, in models.GetLoanPartnerAccountByParamsIn) (out []models.LoanPartnerAccount, err error)
}

type loanPartnerAccount service

const (
	prefixKeyCacheLoanPartner = "pas_loan_partner_key_*"
)

var _ LoanPartnerService = (*loanPartnerAccount)(nil)

func (l *loanPartnerAccount) Create(ctx context.Context, in models.LoanPartnerAccount) (out models.LoanPartnerAccount, err error) {
	defer func() {
		logService(ctx, err)
	}()

	acct, loanPartnerAccountNumber, err := l.validateAccountNumber(ctx, in.AccountNumber, in.LoanSubCategoryCode)
	if err != nil {
		return
	}
	if len(loanPartnerAccountNumber) > 0 {
		err = models.GetErrMap(models.ErrKeyAccountNumberIsExist)
		return
	}

	in.EntityCode = acct.EntityCode

	if err = l.validateLoanPartnerAccount(ctx, models.GetLoanPartnerAccountByParamsIn{
		PartnerId:           in.PartnerId,
		LoanKind:            in.LoanKind,
		AccountType:         in.AccountType,
		EntityCode:          in.EntityCode,
		LoanSubCategoryCode: in.LoanSubCategoryCode,
	}); err != nil {
		return
	}

	if err = l.srv.mySqlRepo.GetLoanPartnerAccountRepository().Create(ctx, in); err != nil {
		err = models.GetErrMap(models.ErrKeyUnableToCreateData, err.Error())
		return
	}

	l.deleteCache(ctx)

	out = in

	return
}

func (l *loanPartnerAccount) Update(ctx context.Context, in models.UpdateLoanPartnerAccount) (out models.UpdateLoanPartnerAccount, err error) {
	defer func() {
		logService(ctx, err)
	}()

	acct, loanPartnerAccountNumber, err := l.validateAccountNumber(ctx, in.AccountNumber, in.LoanSubCategoryCode)
	if err != nil {
		return
	}
	if len(loanPartnerAccountNumber) == 0 {
		err = models.GetErrMap(models.ErrKeyAccountNumberNotFound)
		return
	}

	if err = l.validateLoanPartnerAccount(ctx, models.GetLoanPartnerAccountByParamsIn{
		PartnerId:           in.PartnerId,
		LoanKind:            in.LoanKind,
		AccountType:         in.AccountType,
		EntityCode:          acct.EntityCode,
		LoanSubCategoryCode: in.LoanSubCategoryCode,
	}); err != nil {
		return
	}

	if err = l.srv.mySqlRepo.GetLoanPartnerAccountRepository().Update(ctx, models.UpdateLoanPartnerAccount{
		PartnerId:           in.PartnerId,
		LoanKind:            in.LoanKind,
		AccountNumber:       in.AccountNumber,
		AccountType:         in.AccountType,
		LoanSubCategoryCode: in.LoanSubCategoryCode,
	}); err != nil {
		err = models.GetErrMap(models.ErrKeyUnableToCreateData, err.Error())
		return
	}

	l.deleteCache(ctx)

	out = in

	return
}

func (l *loanPartnerAccount) GetByParams(ctx context.Context, in models.GetLoanPartnerAccountByParamsIn) (out []models.LoanPartnerAccount, err error) {
	defer func() {
		logService(ctx, err)
	}()

	isEmpty := reflect.DeepEqual(in, models.GetLoanPartnerAccountByParamsIn{})
	if isEmpty {
		out, err = l.srv.mySqlRepo.GetLoanPartnerAccountRepository().GetByParams(ctx, in)
		return
	}

	if in.EntityCode != "" && in.AccountNumber == "" {
		var exist *models.Entity
		exist, err = l.srv.mySqlRepo.GetEntityRepository().GetByCode(ctx, in.EntityCode)
		if err != nil {
			return nil, err
		}
		if exist == nil {
			err = models.GetErrMap(models.ErrKeyEntityCodeNotFound)
			return nil, err
		}
	}

	if in.AccountNumber != "" {
		account, err := l.srv.mySqlRepo.GetAccountRepository().GetOneByAccountNumber(ctx, in.AccountNumber)
		if err != nil {
			err = checkDatabaseError(err, models.ErrKeyAccountNumberNotFound)
			return nil, err
		}
		in.EntityCode = account.EntityCode
	}

	if in.LoanAccountNumber != "" {
		account, err := l.srv.mySqlRepo.GetAccountRepository().GetOneByAccountNumber(ctx, in.LoanAccountNumber)
		if err != nil {
			err = checkDatabaseError(err, models.ErrKeyAccountNumberNotFound)
			return nil, err
		}
		in.EntityCode = account.EntityCode
		in.LoanSubCategoryCode = account.SubCategoryCode
	}

	out, err = cache.GetOrSet(l.srv.cacheRepo, models.GetOrSetCacheOpts[[]models.LoanPartnerAccount]{
		Ctx: ctx,
		Key: pasLoanPartnerAccountKey(in),
		TTL: l.srv.conf.CacheTTL.GetLoanPartnerAccount,
		Callback: func() ([]models.LoanPartnerAccount, error) {
			return l.srv.mySqlRepo.GetLoanPartnerAccountRepository().GetByParams(ctx, in)
		},
	})
	if err != nil {
		return nil, err
	}

	if len(out) == 0 {
		err = models.GetErrMap(models.ErrKeyDataNotFound)
		return
	}

	return
}

func pasLoanPartnerAccountKey(in models.GetLoanPartnerAccountByParamsIn) string {
	const prefix = "pas_loan_partner_key"
	return prefix + "_" + in.AccountNumber + in.PartnerId + in.LoanKind + in.AccountType + in.EntityCode + in.LoanAccountNumber + in.LoanSubCategoryCode
}

func (l *loanPartnerAccount) validateAccountNumber(ctx context.Context, accountNumber, subCategoryCode string) (account models.GetAccountOut, loanPartnerAccountNumber []models.LoanPartnerAccount, err error) {
	account, err = l.srv.mySqlRepo.GetAccountRepository().GetOneByAccountNumber(ctx, accountNumber)
	if err != nil {
		err = checkDatabaseError(err, models.ErrKeyAccountNumberNotFound)
		return
	}

	loanPartnerAccountNumber, err = l.srv.mySqlRepo.GetLoanPartnerAccountRepository().GetByParams(ctx, models.GetLoanPartnerAccountByParamsIn{
		AccountNumber:       accountNumber,
		LoanSubCategoryCode: subCategoryCode,
	})
	if err != nil {
		return
	}
	return
}

func (l *loanPartnerAccount) validateLoanPartnerAccount(ctx context.Context, in models.GetLoanPartnerAccountByParamsIn) (err error) {
	loanPartnerAccounts, err := l.srv.mySqlRepo.GetLoanPartnerAccountRepository().GetByParams(ctx, models.GetLoanPartnerAccountByParamsIn{
		PartnerId:           in.PartnerId,
		LoanKind:            in.LoanKind,
		AccountType:         in.AccountType,
		EntityCode:          in.EntityCode,
		LoanSubCategoryCode: in.LoanSubCategoryCode,
	})
	if err != nil {
		return
	}
	if len(loanPartnerAccounts) > 0 {
		err = models.GetErrMap(models.ErrKeyDataIsExist)
		return
	}

	return
}

func (l *loanPartnerAccount) deleteCache(ctx context.Context) {
	if errDelCache := l.srv.cacheRepo.DeleteKeysWithPrefix(ctx, prefixKeyCacheLoanPartner); errDelCache != nil {
		xlog.Warn(ctx, "[DELETE CACHING]", xlog.Any("prefix", prefixKeyCacheLoanPartner), xlog.Err(errDelCache))
	}
}
