package mysql

import (
	"context"
	"fmt"

	"bitbucket.org/Amartha/go-accounting/internal/models"
	"bitbucket.org/Amartha/go-accounting/internal/pkg/dbutil"

	xlog "bitbucket.org/Amartha/go-x/log"
)

type sqlRepo struct {
	r *Repository
}

type Repository struct {
	db     dbutil.DB
	common sqlRepo

	acctr *accountingRepository
	ar    *accountRepository
	coatr *coaTypeRepository
	cr    *categoryRepository
	er    *entityRepository
	lpar  *loanPartnerAccount
	ptr   *productTypeRepository
	scr   *subCategoryRepository
	tbr   *trialBalanceRepository
}

func NewMySQLRepository(db dbutil.DB) *Repository {
	rtx := &Repository{
		db: db,
	}
	rtx.common.r = rtx
	rtx.acctr = (*accountingRepository)(&rtx.common)
	rtx.ar = (*accountRepository)(&rtx.common)
	rtx.coatr = (*coaTypeRepository)(&rtx.common)
	rtx.cr = (*categoryRepository)(&rtx.common)
	rtx.er = (*entityRepository)(&rtx.common)
	rtx.lpar = (*loanPartnerAccount)(&rtx.common)
	rtx.ptr = (*productTypeRepository)(&rtx.common)
	rtx.scr = (*subCategoryRepository)(&rtx.common)
	rtx.tbr = (*trialBalanceRepository)(&rtx.common)
	return rtx
}

type SQLRepository interface {
	Atomic(ctx context.Context, steps func(ctx context.Context, r SQLRepository) error) error
	Ping(ctx context.Context) error
	GetAllCategorySubCategoryCOAType(ctx context.Context) (map[string][]models.CategorySubCategoryCOAType, map[string][]models.CategorySubCategoryCOAType, map[string]models.CategorySubCategoryCOAType, error)
	GetAccountRepository() AccountRepository
	GetAccountingRepository() AccountingRepository
	GetTrialBalanceRepository() TrialBalanceRepository
	GetCategoryRepository() CategoryRepository
	GetCOATypeRepository() COATypeRepository
	GetEntityRepository() EntityRepository
	GetLoanPartnerAccountRepository() LoanPartnerAccountRepository
	GetSubCategoryRepository() SubCategoryRepository
	GetProductTypeRepository() ProductTypeRepository
	CheckAccountTrialBalanceExist(ctx context.Context, entityCode string, subCategoryCode string) (exist bool, err error)
}

var _ SQLRepository = (*Repository)(nil)

func (r *Repository) Atomic(ctx context.Context, steps func(ctx context.Context, r SQLRepository) error) (err error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		err = models.GetErrMap(models.ErrKeyDatabaseError, err.Error())
		return err
	}

	xlog.Info(ctx, "[DATABASE.TRANSACTION.BEGIN]")
	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			err = models.GetErrMap(
				models.ErrKeyDatabaseError,
				fmt.Sprintf("panic happened because: "+fmt.Sprintf("%v", p)),
			)
			xlog.Error(ctx, "[DATABASE.TRANSACTION.PANIC]", xlog.Err(err))
		} else if err != nil {
			if rbErr := tx.Rollback(); rbErr != nil {
				err = models.GetErrMap(
					models.ErrKeyDatabaseError,
					fmt.Sprintf("tx err: %v, rb err: %v", err.Error(), rbErr.Error()),
				)
			}
			xlog.Error(ctx, "[DATABASE.TRANSACTION.ROLLBACK]", xlog.Err(err))
		} else {
			if err = tx.Commit(); err != nil {
				xlog.Error(ctx, "[DATABASE.TRANSACTION.COMMIT]", xlog.Err(err))
			}
			xlog.Info(ctx, "[DATABASE.TRANSACTION.COMMIT]")
		}
	}()
	ctx = injectTx(ctx, tx)
	err = steps(ctx, r)
	return
}

func (r *Repository) Ping(ctx context.Context) error {
	return r.db.PingContext(ctx)
}

func (r *Repository) GetAccountRepository() AccountRepository {
	return r.ar
}

func (r *Repository) GetAccountingRepository() AccountingRepository {
	return r.acctr
}

func (r *Repository) GetCategoryRepository() CategoryRepository {
	return r.cr
}

func (r *Repository) GetAllCategorySubCategoryCOAType(ctx context.Context) (map[string][]models.CategorySubCategoryCOAType, map[string][]models.CategorySubCategoryCOAType, map[string]models.CategorySubCategoryCOAType, error) {
	var err error

	db := r.extractTx(ctx)

	rows, err := db.QueryContext(ctx, queryGetCategorySubCategoryCOAType)
	if err != nil {
		err = databaseError(err)
		return nil, nil, nil, err
	}
	defer rows.Close()

	mapKeyCOATypeCode := make(map[string][]models.CategorySubCategoryCOAType)
	mapKeyCategoryCode := make(map[string][]models.CategorySubCategoryCOAType)
	mapKeySubCategoryCode := make(map[string]models.CategorySubCategoryCOAType)

	for rows.Next() {
		var value models.CategorySubCategoryCOAType
		var err = rows.Scan(
			&value.CategoryCode,
			&value.CategoryName,
			&value.SubCategoryCode,
			&value.SubCategoryName,
			&value.CoaTypeCode,
			&value.CoaTypeName,
		)
		if err != nil {
			err = databaseError(err)
			return nil, nil, nil, err
		}
		mapKeyCOATypeCode[value.CoaTypeCode] = append(mapKeyCOATypeCode[value.CoaTypeCode], value)
		mapKeyCategoryCode[value.CategoryCode] = append(mapKeyCategoryCode[value.CategoryCode], value)
		mapKeySubCategoryCode[value.SubCategoryCode] = value
	}
	if rows.Err() != nil {
		err = databaseError(rows.Err())
		return nil, nil, nil, err
	}

	return mapKeyCOATypeCode, mapKeyCategoryCode, mapKeySubCategoryCode, nil
}

func (r *Repository) GetCOATypeRepository() COATypeRepository {
	return r.coatr
}

func (r *Repository) GetEntityRepository() EntityRepository {
	return r.er
}

func (r *Repository) GetLoanPartnerAccountRepository() LoanPartnerAccountRepository {
	return r.lpar
}

func (r *Repository) GetSubCategoryRepository() SubCategoryRepository {
	return r.scr
}

func (r *Repository) GetProductTypeRepository() ProductTypeRepository {
	return r.ptr
}

func (r *Repository) CheckAccountTrialBalanceExist(ctx context.Context, entityCode string, subCategoryCode string) (exist bool, err error) {
	return r.GetAccountRepository().CheckExistByParam(ctx, models.AccountFilterOptions{EntityCode: entityCode, SubCategoryCode: subCategoryCode})
}

func (r *Repository) GetTrialBalanceRepository() TrialBalanceRepository {
	return r.tbr
}
