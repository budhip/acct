package mysql

import (
	"context"
	"fmt"
	"strings"
	"time"

	"bitbucket.org/Amartha/go-accounting/internal/models"
	"bitbucket.org/Amartha/go-accounting/internal/pkg/atime"
	"bitbucket.org/Amartha/go-accounting/internal/pkg/money"

	"github.com/shopspring/decimal"
)

func (ar *accountingRepository) CalculateFromAccountBalanceDaily(ctx context.Context, in models.CalculateTrialBalance) (out models.AccountTrialBalance, err error) {
	query, args, err := buildCalculateFromAccountBalanceDaily(in)
	if err != nil {
		err = fmt.Errorf("failed to build query: %w", err)
		return
	}

	db := ar.r.extractTx(ctx)
	if err = db.QueryRowContext(ctx, query, args...).Scan(
		&out.EntityCode,
		&out.CategoryCode,
		&out.SubCategoryCode,
		&out.DebitMovement,
		&out.CreditMovement,
	); err != nil {
		return
	}

	return
}

func (ar *accountingRepository) CalculateFromTransactions(ctx context.Context, in models.CalculateTrialBalance) (out models.AccountTrialBalance, err error) {
	query, args, err := buildCalculateFromTransactions(in)
	if err != nil {
		err = fmt.Errorf("failed to build query: %w", err)
		return
	}

	db := ar.r.extractTx(ctx)
	if err = db.QueryRowContext(ctx, query, args...).Scan(
		&out.EntityCode,
		&out.CategoryCode,
		&out.SubCategoryCode,
		&out.DebitMovement,
		&out.CreditMovement,
	); err != nil {
		return
	}

	return
}

func (ar *accountingRepository) GetOpeningBalanceFromAccountTrialBalance(ctx context.Context, in models.CalculateTrialBalance) (openingBalance decimal.Decimal, err error) {
	db := ar.r.extractTx(ctx)

	if err = db.QueryRowContext(ctx, queryGetOpeingBalanceFromAccountTrialBalance, in.EntityCode, in.SubCategoryCode, in.Date).Scan(
		&openingBalance,
	); err != nil {
		return
	}

	return
}

func (ar *accountingRepository) InsertAccountTrialBalance(ctx context.Context, in []models.AccountTrialBalance) (err error) {
	start := atime.Now()
	defer func() {
		logSQL(ctx, err, start)
	}()
	db := ar.r.extractTx(ctx)

	valueStrings := []string{}
	valueArgs := []interface{}{}
	for _, req := range in {
		valueStrings = append(valueStrings, `(?, ?, ?, ?, ?, ?, ?, ?)`)
		valueArgs = append(valueArgs, req.ClosingDate)
		valueArgs = append(valueArgs, req.EntityCode)
		valueArgs = append(valueArgs, req.CategoryCode)
		valueArgs = append(valueArgs, req.SubCategoryCode)
		valueArgs = append(valueArgs, req.DebitMovement)
		valueArgs = append(valueArgs, req.CreditMovement)
		valueArgs = append(valueArgs, req.OpeningBalance)
		valueArgs = append(valueArgs, req.ClosingBalance)
	}

	query := fmt.Sprintf(queryInsertAccountTrialBalance, strings.Join(valueStrings, ","))
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

func (ar *accountingRepository) GetTrialBalanceV2(ctx context.Context, opts models.TrialBalanceFilterOptions) (coaCategories map[string][]models.TBCOACategory, coaSubCategories map[models.TBCOACategory][]models.TBSubCategory, err error) {
	start := atime.Now()

	defer func() {
		logSQL(ctx, err, start)
	}()

	_, _, mapSubCategory, err := ar.r.GetAllCategorySubCategoryCOAType(ctx)
	if err != nil {
		return nil, nil, err
	}

	query, args, err := buildGetTrialBalanceV2Query(opts)
	if err != nil {
		err = fmt.Errorf("failed to build query: %w", err)
		return nil, nil, err
	}

	db := ar.r.extractTx(ctx)
	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		err = databaseError(err)
		return nil, nil, err
	}
	defer rows.Close()

	coaCategories = make(map[string][]models.TBCOACategory)
	coaSubCategories = make(map[models.TBCOACategory][]models.TBSubCategory)

	for rows.Next() {
		var v = models.GetTrialBalanceV2Out{}
		var errScan = rows.Scan(
			&v.EntityCode,
			&v.CategoryCode,
			&v.SubCategoryCode,
			&v.DebitMovement,
			&v.CreditMovement,
			&v.OpeningBalance,
			&v.ClosingBalance,
		)
		if errScan != nil {
			err = databaseError(errScan)
			return
		}

		categoryCode := v.CategoryCode
		subCategoryCode := v.SubCategoryCode

		v.CoaTypeCode = mapSubCategory[subCategoryCode].CoaTypeCode
		v.CoaTypeName = mapSubCategory[subCategoryCode].CoaTypeName
		v.CategoryName = mapSubCategory[subCategoryCode].CategoryName
		v.SubCategoryName = mapSubCategory[subCategoryCode].SubCategoryName

		coaCategories[categoryCode] = append(coaCategories[categoryCode], models.TBCOACategory{
			Type:                v.CoaTypeName,
			CategoryCode:        categoryCode,
			CategoryName:        v.CategoryName,
			TotalOpeningBalance: v.OpeningBalance,
			TotalDebitMovement:  v.DebitMovement,
			TotalCreditMovement: v.CreditMovement,
			TotalClosingBalance: v.ClosingBalance,
		})

		coaTypeName := strings.ToLower(v.CoaTypeName)
		key := models.TBCOACategory{
			Type:         coaTypeName,
			CoaTypeCode:  v.CoaTypeCode,
			CoaTypeName:  v.CoaTypeName,
			CategoryCode: categoryCode,
			CategoryName: v.CategoryName,
		}
		coaSubCategories[key] = append(coaSubCategories[key], models.TBSubCategory{
			Kind:            models.KindSubCategory,
			SubCategoryCode: v.SubCategoryCode,
			SubCategoryName: v.SubCategoryName,
			OpeningBalance:  v.OpeningBalance,
			DebitMovement:   v.DebitMovement,
			CreditMovement:  v.CreditMovement,
			ClosingBalance:  v.ClosingBalance,

			IDRFormatOpeningBalance: money.FormatAmountToIDRFromDecimal(v.OpeningBalance),
			IDRFormatDebitMovement:  money.FormatAmountToIDRFromDecimal(v.DebitMovement),
			IDRFormatCreditMovement: money.FormatAmountToIDRFromDecimal(v.CreditMovement),
			IDRFormatClosingBalance: money.FormatAmountToIDRFromDecimal(v.ClosingBalance),
		})
	}
	if rows.Err() != nil {
		err = databaseError(rows.Err())
		return nil, nil, err
	}

	return
}

func (ar *accountingRepository) GetTrialBalanceSubCategory(ctx context.Context, opts models.TrialBalanceFilterOptions) (out models.TrialBalanceBySubCategoryOut, err error) {
	start := atime.Now()

	defer func() {
		logSQL(ctx, err, start)
	}()

	db := ar.r.extractTx(ctx)

	if opts.SubCategoryCode == "" {
		err = fmt.Errorf("sub category code is required")
		return
	}

	query, args, err := buildGetTrialBalanceSubCategory(opts)
	if err != nil {
		err = fmt.Errorf("failed to build query: %w", err)
		return
	}

	err = db.
		QueryRowContext(ctx, query, args...).
		Scan(
			&out.SubCategoryCode,
			&out.SubCategoryName,
			&out.DebitMovement,
			&out.CreditMovement,
			&out.OpeningBalance,
			&out.ClosingBalance,
		)
	if err != nil {
		return
	}

	out.OpeningBalance = money.FormatBigIntToAmount(out.OpeningBalance, CurrencyIDR.Decimals)
	out.ClosingBalance = money.FormatBigIntToAmount(out.ClosingBalance, CurrencyIDR.Decimals)
	out.CreditMovement = money.FormatBigIntToAmount(out.CreditMovement, CurrencyIDR.Decimals)
	out.DebitMovement = money.FormatBigIntToAmount(out.DebitMovement, CurrencyIDR.Decimals)

	return
}

func (ar *accountingRepository) GetTransactionsToday(ctx context.Context, transactionDate time.Time) (transactions []string, err error) {
	defer func() {
		logSQL(ctx, err)
	}()

	query, args, err := buildGetTransactionsToday(transactionDate)
	if err != nil {
		err = fmt.Errorf("failed to build query: %w", err)
		return
	}

	db := ar.r.extractTx(ctx)
	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		err = databaseError(err)
		return
	}
	defer rows.Close()

	var transactionIDs []string
	for rows.Next() {
		var transactionID string
		var errScan = rows.Scan(
			&transactionID,
		)
		if errScan != nil {
			err = databaseError(errScan)
			return
		}
		transactionIDs = append(transactionIDs, transactionID)
	}
	if rows.Err() != nil {
		err = databaseError(rows.Err())
		return
	}

	return transactionIDs, nil
}
