package mysql

import (
	"context"
	"fmt"
	"strings"

	"bitbucket.org/Amartha/go-accounting/internal/models"
)

type LoanPartnerAccountRepository interface {
	Create(ctx context.Context, in models.LoanPartnerAccount) (err error)
	Update(ctx context.Context, in models.UpdateLoanPartnerAccount) (err error)
	GetByParams(ctx context.Context, in models.GetLoanPartnerAccountByParamsIn) (out []models.LoanPartnerAccount, err error)
	BulkInsertLoanPartnerAccount(ctx context.Context, in []models.LoanPartnerAccount) (err error)
}

type loanPartnerAccount sqlRepo

var _ LoanPartnerAccountRepository = (*loanPartnerAccount)(nil)

func (lr *loanPartnerAccount) Create(ctx context.Context, in models.LoanPartnerAccount) (err error) {
	db := lr.r.extractTx(ctx)

	res, err := db.ExecContext(ctx, queryLoanPartnerAccountCreate, in.PartnerId, in.LoanKind, in.AccountNumber, in.AccountType, in.EntityCode, in.LoanSubCategoryCode)
	if err != nil {
		return
	}

	affectedRows, err := res.RowsAffected()
	if err != nil {
		return
	}

	if affectedRows == 0 {
		err = models.ErrNoRowsAffected
		return
	}

	return
}

func (lr *loanPartnerAccount) Update(ctx context.Context, in models.UpdateLoanPartnerAccount) (err error) {
	defer func() {
		logSQL(ctx, err)
	}()

	db := lr.r.extractTx(ctx)
	query, args, err := queryAccountLoanPartnerUpdate(in)
	if err != nil {
		err = fmt.Errorf("failed to build query: %w", err)
		return
	}

	_, err = db.ExecContext(ctx, query, args...)
	if err != nil {
		err = databaseError(err)
		return
	}

	return
}

func (lr *loanPartnerAccount) GetByParams(ctx context.Context, in models.GetLoanPartnerAccountByParamsIn) (out []models.LoanPartnerAccount, err error) {
	db := lr.r.extractTx(ctx)

	query, args, err := buildQueryGetLoanPartnerAccountByParam(in)
	if err != nil {
		err = fmt.Errorf("failed to build query: %w", err)
		return
	}

	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		err = databaseError(err)
		return nil, err
	}
	defer rows.Close()

	var result []models.LoanPartnerAccount
	for rows.Next() {
		var val models.LoanPartnerAccount
		err := rows.Scan(
			&val.PartnerId,
			&val.LoanKind,
			&val.AccountNumber,
			&val.AccountType,
			&val.EntityCode,
			&val.LoanSubCategoryCode,
			&val.CreatedAt,
			&val.UpdatedAt,
		)
		if err != nil {
			err = databaseError(err)
			return nil, err
		}
		result = append(result, val)
	}
	if rows.Err() != nil {
		err = databaseError(rows.Err())
		return nil, err
	}

	return result, err
}

func (lr *loanPartnerAccount) BulkInsertLoanPartnerAccount(ctx context.Context, in []models.LoanPartnerAccount) (err error) {
	defer func() {
		logSQL(ctx, err)
	}()

	db := lr.r.extractTx(ctx)

	valueStrings := []string{}
	valueArgs := []interface{}{}
	for _, req := range in {
		valueStrings = append(valueStrings, "(NULLIF(?, ''), NULLIF(?, ''), NULLIF(?, ''), NULLIF(?, ''), NULLIF(?, ''), NULLIF(?, ''))")
		valueArgs = append(valueArgs, req.PartnerId)
		valueArgs = append(valueArgs, req.LoanKind)
		valueArgs = append(valueArgs, req.AccountNumber)
		valueArgs = append(valueArgs, req.AccountType)
		valueArgs = append(valueArgs, req.EntityCode)
		valueArgs = append(valueArgs, req.LoanSubCategoryCode)
	}

	query := fmt.Sprintf(queryBulkLoanPartnerAccountCreate, strings.Join(valueStrings, ","))
	res, err := db.ExecContext(ctx, query, valueArgs...)
	if err != nil {
		err = databaseError(err)
		return
	}

	affectedRows, err := res.RowsAffected()
	if err != nil {
		err = databaseError(err)
		return
	}

	if affectedRows == 0 {
		err = databaseError(models.ErrNoRowsAffected)
		return
	}

	return
}
