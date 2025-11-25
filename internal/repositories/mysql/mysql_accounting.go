package mysql

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"bitbucket.org/Amartha/go-accounting/internal/models"
	"bitbucket.org/Amartha/go-accounting/internal/pkg/atime"
	"bitbucket.org/Amartha/go-accounting/internal/pkg/money"

	"github.com/shopspring/decimal"
)

type AccountingRepository interface {
	// TODO[Syldie]: separate TB, GL, etc into multiple repository to prevent bulky interface
	ToggleForeignKeyChecks(ctx context.Context, isEnable bool) (err error)

	// split
	InsertSplit(ctx context.Context, in []models.CreateSplit) (err error)

	// split-account
	InsertSplitAccount(ctx context.Context, in []models.CreateSplitAccount) (err error)
	GetOneSplitAccount(ctx context.Context, accountNumber string) (isExist bool, err error)
	// transaction
	InsertTransaction(ctx context.Context, in []models.CreateTransaction) (err error)
	CheckTransactionIdIsExist(ctx context.Context, transactionId string) (isExist bool, err error)

	// journal
	GetJournalDetailByTransactionId(ctx context.Context, transactionId string) (result []models.GetJournalDetailOut, err error)
	InsertJournalDetail(ctx context.Context, in []models.CreateJournalDetail) (err error)

	// trial-balance
	CalculateFromAccountBalanceDaily(ctx context.Context, in models.CalculateTrialBalance) (out models.AccountTrialBalance, err error)
	CalculateFromTransactions(ctx context.Context, in models.CalculateTrialBalance) (out models.AccountTrialBalance, err error)
	GetOpeningBalanceFromAccountTrialBalance(ctx context.Context, in models.CalculateTrialBalance) (openingBalance decimal.Decimal, err error)
	CalculateOpeningClosingBalanceFromAccountBalance(ctx context.Context, in models.CalculateTrialBalance) (openingBalance decimal.Decimal, err error)
	InsertAccountTrialBalance(ctx context.Context, in []models.AccountTrialBalance) (err error)
	GetTrialBalance(ctx context.Context, opts models.TrialBalanceFilterOptions) (coaCategories map[string][]models.TBCOACategory, coaSubCategories map[models.TBCOACategory][]models.TBSubCategory, err error)
	GetTrialBalanceV2(ctx context.Context, opts models.TrialBalanceFilterOptions) (coaCategories map[string][]models.TBCOACategory, coaSubCategories map[models.TBCOACategory][]models.TBSubCategory, err error)
	GetTrialBalanceSubCategory(ctx context.Context, opts models.TrialBalanceFilterOptions) (out models.TrialBalanceBySubCategoryOut, err error)
	GetTrialBalanceDetails(ctx context.Context, opts models.TrialBalanceDetailsFilterOptions) (tba []models.TrialBalanceDetailOut, err error)

	// sub-ledger
	GetSubLedger(ctx context.Context, opts models.SubLedgerFilterOptions) (result []models.GetSubLedgerOut, err error)
	GetSubLedgerStream(ctx context.Context, opts models.SubLedgerFilterOptions) <-chan models.StreamResult[models.GetSubLedgerOut]
	GetSubLedgerCount(ctx context.Context, opts models.SubLedgerFilterOptions) (total int, err error)
	GetSubLedgerAccounts(ctx context.Context, opts models.SubLedgerAccountsFilterOptions) (result []models.GetSubLedgerAccountsOut, err error)
	GetSubLedgerAccountTotalTransaction(ctx context.Context, opts models.SubLedgerAccountsFilterOptions) (total int, err error)
	GetSubLedgerAccountsCount(ctx context.Context, opts models.SubLedgerAccountsFilterOptions) (total int, err error)

	// account balance daily
	GetAccountBalancePeriodStart(ctx context.Context, accountNumber string, date time.Time) (balance decimal.Decimal, err error)
	GetOpeningBalanceByDate(ctx context.Context, accountNumber string, date time.Time) (openingBalance decimal.Decimal, err error)
	GetLastOpeningBalance(ctx context.Context, accountNumber string, date time.Time) (openingBalance decimal.Decimal, err error)
	CalculateOpeningClosingBalance(ctx context.Context, accountNumber string, date time.Time) (out models.AccountBalanceDaily, err error)
	GetOneAccountBalanceDaily(ctx context.Context, accountNumber string, date time.Time) (out models.AccountBalanceDaily, err error)

	GetAccountTransactionByDate(ctx context.Context, entities []string, date time.Time) <-chan models.StreamResult[models.AccountTransation]
	GetAllAccountDailyBalance(ctx context.Context, entities []string, subCategories *[]models.SubCategory, date time.Time) <-chan models.StreamResult[models.AccountBalanceDaily]
	InsertAccountBalanceDaily(ctx context.Context, in []models.AccountBalanceDaily) (err error)

	// balance sheet
	GetBalanceSheet(ctx context.Context, opts models.BalanceSheetFilterOptions) (out models.BalanceSheetOut, err error)

	GetTransactionsToday(ctx context.Context, transactionDate time.Time) (transactions []string, err error)
}

type accountingRepository sqlRepo

var _ AccountingRepository = (*accountingRepository)(nil)

func (ar *accountingRepository) GetTrialBalance(ctx context.Context, opts models.TrialBalanceFilterOptions) (coaCategories map[string][]models.TBCOACategory, coaSubCategories map[models.TBCOACategory][]models.TBSubCategory, err error) {
	start := atime.Now()

	defer func() {
		logSQL(ctx, err, start)
	}()

	query, args, err := buildGetTrialBalanceQuery(opts)
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
		var v = models.GetTrialBalanceOut{}
		var errScan = rows.Scan(
			&v.CoaTypeCode,
			&v.CoaTypeName,
			&v.CategoryCode,
			&v.CategoryName,
			&v.SubCategoryCode,
			&v.SubCategoryName,
			&v.DebitMovement,
			&v.CreditMovement,
			&v.OpeningBalance,
			&v.ClosingBalance,
		)
		if errScan != nil {
			err = databaseError(errScan)
			return
		}

		coaCategories[v.CategoryCode] = append(coaCategories[v.CategoryCode], models.TBCOACategory{
			Type:                v.CoaTypeName,
			CategoryCode:        v.CategoryCode,
			CategoryName:        v.CategoryName,
			TotalOpeningBalance: v.OpeningBalance,
			TotalDebitMovement:  v.DebitMovement,
			TotalCreditMovement: v.CreditMovement,
			TotalClosingBalance: v.ClosingBalance,
		})

		coaType := strings.ToLower(v.CoaTypeName)
		key := models.TBCOACategory{
			Type:         coaType,
			CoaTypeCode:  v.CoaTypeCode,
			CoaTypeName:  v.CoaTypeName,
			CategoryCode: v.CategoryCode,
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
		return
	}

	return
}

func (ar *accountingRepository) InsertJournalDetail(ctx context.Context, in []models.CreateJournalDetail) (err error) {
	defer func() {
		logSQL(ctx, err)
	}()

	db := ar.r.extractTx(ctx)

	valueStrings := []string{}
	valueArgs := []interface{}{}
	for _, req := range in {
		valueStrings = append(valueStrings, "(?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP(6), CURRENT_TIMESTAMP(6), ?)")
		valueArgs = append(valueArgs, req.JournalId, req.ReferenceNumber, req.OrderType, req.TransactionType, req.TransactionTypeName, req.TransactionDate, req.IsDebit, req.Metadata)
	}

	query := fmt.Sprintf(queryInsertJournalDetail, strings.Join(valueStrings, ","))
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

func (ar *accountingRepository) GetSubLedger(ctx context.Context, opts models.SubLedgerFilterOptions) (result []models.GetSubLedgerOut, err error) {
	start := atime.Now()

	defer func() {
		logSQL(ctx, err, start)
	}()

	db := ar.r.extractTx(ctx)

	query, args, err := buildSubLedgerQuery(opts)
	if err != nil {
		err = fmt.Errorf("failed to build query: %w", err)
		return nil, err
	}

	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		err = databaseError(err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var out = models.GetSubLedgerOut{}
		var err = rows.Scan(
			&out.TransactionID,
			&out.ReferenceNumber,
			&out.TransactionDate,
			&out.OrderType,
			&out.TransactionType,
			&out.Narrative,
			&out.Metadata,
			&out.Debit,
			&out.Credit,
			&out.CreatedAt,
			&out.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		out.Debit = money.FormatBigIntToAmount(out.Debit, CurrencyIDR.Decimals)
		out.Credit = money.FormatBigIntToAmount(out.Credit, CurrencyIDR.Decimals)
		result = append(result, out)
	}
	if rows.Err() != nil {
		err = databaseError(rows.Err())
		return result, err
	}

	return result, nil
}

func (ar *accountingRepository) GetSubLedgerStream(ctx context.Context, opts models.SubLedgerFilterOptions) <-chan models.StreamResult[models.GetSubLedgerOut] {
	db := ar.r.extractTx(ctx)

	ch := make(chan models.StreamResult[models.GetSubLedgerOut], 1)
	go func() {
		defer close(ch)

		opts.Limit = 100_000
		opts.Offset = 0
		for {
			query, args, err := buildSubLedgerQuery(opts)
			if err != nil {
				ch <- models.StreamResult[models.GetSubLedgerOut]{Err: err}
				return
			}

			rows, err := db.QueryContext(ctx, query, args...)
			if err != nil {
				ch <- models.StreamResult[models.GetSubLedgerOut]{Err: err}
				return
			}
			defer rows.Close()

			rowsProcessed := false
			for rows.Next() {
				select {
				case <-ctx.Done():
					return
				default:
					var out models.GetSubLedgerOut
					err := rows.Scan(
						&out.TransactionID,
						&out.ReferenceNumber,
						&out.TransactionDate,
						&out.OrderType,
						&out.TransactionType,
						&out.Narrative,
						&out.Metadata,
						&out.Debit,
						&out.Credit,
						&out.CreatedAt,
						&out.UpdatedAt,
					)
					if err != nil {
						ch <- models.StreamResult[models.GetSubLedgerOut]{Err: err}
						return
					}
					out.Debit = money.FormatBigIntToAmount(out.Debit, CurrencyIDR.Decimals)
					out.Credit = money.FormatBigIntToAmount(out.Credit, CurrencyIDR.Decimals)
					ch <- models.StreamResult[models.GetSubLedgerOut]{Data: out}
					rowsProcessed = true
				}
			}

			if err := rows.Err(); err != nil {
				ch <- models.StreamResult[models.GetSubLedgerOut]{Err: err}
				return
			}

			if !rowsProcessed {
				break
			}
			opts.Offset += opts.Limit
		}
	}()

	return ch
}

func (ar *accountingRepository) GetSubLedgerCount(ctx context.Context, opts models.SubLedgerFilterOptions) (total int, err error) {
	start := atime.Now()

	defer func() {
		logSQL(ctx, err, start)
	}()

	db := ar.r.extractTx(ctx)

	query, args, err := buildCountSubLedgerQuery(opts)
	if err != nil {
		err = fmt.Errorf("failed to build query: %w", err)
		return total, err
	}

	if err = db.QueryRowContext(ctx, query, args...).Scan(&total); err != nil {
		err = databaseError(err)
		return
	}

	return
}

func (ar *accountingRepository) GetAccountBalancePeriodStart(ctx context.Context, accountNumber string, date time.Time) (balance decimal.Decimal, err error) {
	start := atime.Now()

	defer func() {
		logSQL(ctx, err, start)
	}()

	db := ar.r.extractTx(ctx)

	query, args, err := getAccountBalancePeriodStart(accountNumber, date)
	if err != nil {
		err = fmt.Errorf("failed to build query: %w", err)
		return
	}

	if err = db.QueryRowContext(ctx, query, args...).Scan(&balance); err != nil {
		err = databaseError(err)
		return
	}

	return
}

func (ar *accountingRepository) GetJournalDetailByTransactionId(ctx context.Context, transactionId string) (result []models.GetJournalDetailOut, err error) {
	defer func() {
		logSQL(ctx, err)
	}()

	db := ar.r.extractTx(ctx)

	query, args, err := getJournalDetailQuery(transactionId)
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

	for rows.Next() {
		var out = models.GetJournalDetailOut{}
		var err = rows.Scan(
			&out.TransactionId,
			&out.JournalId,
			&out.AccountNumber,
			&out.AccountName,
			&out.AltId,
			&out.EntityCode,
			&out.EntityName,
			&out.SubCategoryCode,
			&out.SubCategoryName,
			&out.TransactionType,
			&out.Amount,
			&out.TransactionDate,
			&out.Narrative,
			&out.IsDebit,
		)
		if err != nil {
			err = databaseError(err)
			return nil, err
		}
		out.Amount = money.FormatBigIntToAmount(out.Amount, CurrencyIDR.Decimals)
		result = append(result, out)
	}
	if rows.Err() != nil {
		err = databaseError(rows.Err())
		return result, err
	}

	return
}

func (ar *accountingRepository) GetSubLedgerAccounts(ctx context.Context, opts models.SubLedgerAccountsFilterOptions) (result []models.GetSubLedgerAccountsOut, err error) {
	start := atime.Now()

	defer func() {
		logSQL(ctx, err, start)
	}()

	_, _, mapSubCategory, err := ar.r.GetAllCategorySubCategoryCOAType(ctx)
	if err != nil {
		return nil, err
	}

	db := ar.r.extractTx(ctx)

	query, args, err := buildSubLedgerAccountsQuery(opts)
	if err != nil {
		err = fmt.Errorf("failed to build query: %w", err)
		return nil, err
	}

	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		err = databaseError(err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var out = models.GetSubLedgerAccountsOut{}
		var err = rows.Scan(
			&out.AccountNumber,
			&out.AccountName,
			&out.AltId,
			&out.SubCategoryCode,
			&out.TotalRowData,
			&out.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		out.SubCategoryName = mapSubCategory[out.SubCategoryCode].SubCategoryName
		result = append(result, out)
	}
	if rows.Err() != nil {
		err = databaseError(rows.Err())
		return result, err
	}

	return result, nil
}

func (ar *accountingRepository) GetSubLedgerAccountsCount(ctx context.Context, opts models.SubLedgerAccountsFilterOptions) (total int, err error) {
	start := atime.Now()

	defer func() {
		logSQL(ctx, err, start)
	}()

	db := ar.r.extractTx(ctx)

	query, args, err := buildCountSubLedgerAccountsQuery(opts)
	if err != nil {
		err = fmt.Errorf("failed to build query: %w", err)
		return total, err
	}

	if err = db.QueryRowContext(ctx, query, args...).Scan(&total); err != nil {
		err = databaseError(err)
		return
	}

	return
}

func (ar *accountingRepository) GetSubLedgerAccountTotalTransaction(ctx context.Context, opts models.SubLedgerAccountsFilterOptions) (total int, err error) {
	start := atime.Now()

	defer func() {
		logSQL(ctx, err, start)
	}()

	db := ar.r.extractTx(ctx)

	query, args, err := buildSubLedgerAccountTotalTransactionQuery(opts)
	if err != nil {
		err = fmt.Errorf("failed to build query: %w", err)
		return
	}

	if err = db.QueryRowContext(ctx, query, args...).Scan(&total); err != nil {
		if errors.Is(err, models.ErrNoRows) {
			err = nil
			return
		}
		err = databaseError(err)
		return
	}

	return
}

func (ar *accountingRepository) GetTrialBalanceDetails(ctx context.Context, opts models.TrialBalanceDetailsFilterOptions) (tba []models.TrialBalanceDetailOut, err error) {
	start := atime.Now()

	defer func() {
		logSQL(ctx, err, start)
	}()

	db := ar.r.extractTx(ctx)

	query, args, err := buildGetListTrialBalanceDetailQuery(opts)
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

	for rows.Next() {
		var out models.TrialBalanceDetailOut
		err = rows.Scan(
			&out.AccountNumber,
			&out.AccountName,
			&out.OpeningBalance,
			&out.ClosingBalance,
			&out.DebitMovement,
			&out.CreditMovement,
		)
		if err != nil {
			return nil, err
		}

		out.OpeningBalance = money.FormatBigIntToAmount(out.OpeningBalance, CurrencyIDR.Decimals)
		out.ClosingBalance = money.FormatBigIntToAmount(out.ClosingBalance, CurrencyIDR.Decimals)
		out.CreditMovement = money.FormatBigIntToAmount(out.CreditMovement, CurrencyIDR.Decimals)
		out.DebitMovement = money.FormatBigIntToAmount(out.DebitMovement, CurrencyIDR.Decimals)

		tba = append(tba, out)
	}

	if rows.Err() != nil {
		err = databaseError(rows.Err())
		return
	}

	return
}

func (ar *accountingRepository) InsertTransaction(ctx context.Context, in []models.CreateTransaction) (err error) {
	defer func() {
		logSQL(ctx, err)
	}()

	db := ar.r.extractTx(ctx)

	valueStrings := []string{}
	valueArgs := []interface{}{}
	for _, req := range in {
		valueStrings = append(valueStrings, "(?, ?, ?)")
		valueArgs = append(valueArgs, req.TransactionID, req.Postdate, req.PosterUserID)
	}

	query := fmt.Sprintf(queryInsertTransaction, strings.Join(valueStrings, ","))
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

func (ar *accountingRepository) ToggleForeignKeyChecks(ctx context.Context, isEnable bool) (err error) {
	defer func() {
		logSQL(ctx, err)
	}()

	db := ar.r.extractTx(ctx)

	query := "SET FOREIGN_KEY_CHECKS = 0;"
	if isEnable {
		query = "SET FOREIGN_KEY_CHECKS = 1;"
	}

	_, err = db.ExecContext(ctx, query)
	if err != nil {
		err = databaseError(err)
		return
	}

	return
}

func (ar *accountingRepository) InsertSplit(ctx context.Context, in []models.CreateSplit) (err error) {
	defer func() {
		logSQL(ctx, err)
	}()

	db := ar.r.extractTx(ctx)

	valueStrings := []string{}
	valueArgs := []interface{}{}
	for _, req := range in {
		valueStrings = append(valueStrings, "(?, ?, ?, ?, ?, ?)")
		valueArgs = append(valueArgs, req.TransactionID, req.SplitID, req.SplitDate, req.Description, req.Currency, req.Amount)
	}

	query := fmt.Sprintf(queryInsertSplit, strings.Join(valueStrings, ","))
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

func (ar *accountingRepository) InsertSplitAccount(ctx context.Context, in []models.CreateSplitAccount) (err error) {
	defer func() {
		logSQL(ctx, err)
	}()

	db := ar.r.extractTx(ctx)

	valueStrings := []string{}
	valueArgs := []interface{}{}
	for _, req := range in {
		valueStrings = append(valueStrings, "(?, ?)")
		valueArgs = append(valueArgs, req.SplitID, req.AccountID)
	}

	query := fmt.Sprintf(queryInsertSplitAccount, strings.Join(valueStrings, ","))
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

func (ar *accountingRepository) GetOneSplitAccount(ctx context.Context, accountNumber string) (isExist bool, err error) {
	defer func() {
		logSQL(ctx, err)
	}()

	db := ar.r.extractTx(ctx)

	if err = db.QueryRowContext(ctx, queryGetOneSplitAccount, accountNumber).Scan(
		&isExist,
	); err != nil {
		return
	}

	return
}

func (ar *accountingRepository) CheckTransactionIdIsExist(ctx context.Context, transactionId string) (isExist bool, err error) {
	defer func() {
		logSQL(ctx, err)
	}()

	db := ar.r.extractTx(ctx)

	if err = db.QueryRowContext(ctx, queryCheckTransactionIdIsExist, transactionId).Scan(
		&isExist,
	); err != nil {
		return
	}

	return
}
