package mysql

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"bitbucket.org/Amartha/go-accounting/internal/models"
	"bitbucket.org/Amartha/go-accounting/internal/pkg/atime"

	"github.com/shopspring/decimal"
)

func (ar *accountingRepository) GetOpeningBalanceByDate(ctx context.Context, accountNumber string, date time.Time) (openingBalance decimal.Decimal, err error) {
	start := atime.Now()
	defer func() {
		logSQL(ctx, err, start)
	}()

	db := ar.r.extractTx(ctx)

	if err = db.QueryRowContext(ctx, queryGetOpeningBalanceDate, accountNumber, date).Scan(
		&openingBalance,
	); err != nil {
		return
	}

	return
}

func (ar *accountingRepository) GetLastOpeningBalance(ctx context.Context, accountNumber string, date time.Time) (openingBalance decimal.Decimal, err error) {
	start := atime.Now()
	defer func() {
		logSQL(ctx, err, start)
	}()

	db := ar.r.extractTx(ctx)

	if err = db.QueryRowContext(ctx, queryGetLastOpeningBalance, accountNumber, date).Scan(
		&openingBalance,
	); err != nil {
		return
	}

	return
}

func (ar *accountingRepository) CalculateOpeningClosingBalance(ctx context.Context, accountNumber string, date time.Time) (out models.AccountBalanceDaily, err error) {
	start := atime.Now()
	defer func() {
		logSQL(ctx, err, start)
	}()

	db := ar.r.extractTx(ctx)

	query, args, err := buildCalculateOpeningClosingBalance(accountNumber, date)
	if err != nil {
		err = fmt.Errorf("failed to build query: %w", err)
		return
	}

	if err = db.QueryRowContext(ctx, query, args...).Scan(
		&out.AccountNumber,
		&out.EntityCode,
		&out.CategoryCode,
		&out.SubCategoryCode,
		&out.OpeningBalance,
		&out.ClosingBalance,
	); err != nil {
		return
	}

	return
}

func (ar *accountingRepository) CalculateOpeningClosingBalanceFromAccountBalance(ctx context.Context, in models.CalculateTrialBalance) (openingBalance decimal.Decimal, err error) {
	start := atime.Now()
	defer func() {
		logSQL(ctx, err, start)
	}()

	db := ar.r.extractTx(ctx)

	if err = db.QueryRowContext(ctx, queryCalculateOpeningClosingBalanceFromAccountBalance, in.EntityCode, in.SubCategoryCode, in.Date).Scan(
		&openingBalance,
	); err != nil {
		return
	}

	return
}

func (ar *accountingRepository) GetOneAccountBalanceDaily(ctx context.Context, accountNumber string, date time.Time) (out models.AccountBalanceDaily, err error) {
	start := atime.Now()
	defer func() {
		logSQL(ctx, err, start)
	}()

	db := ar.r.extractTx(ctx)

	if err = db.QueryRowContext(ctx, queryGetOneAccountBalanceDaily, accountNumber, date).Scan(
		&out.AccountNumber,
		&out.EntityCode,
		&out.CategoryCode,
		&out.SubCategoryCode,
		&out.OpeningBalance,
		&out.ClosingBalance,
	); err != nil {
		return
	}

	return
}

func (ar *accountingRepository) GetAccountTransactionByDate(ctx context.Context, entities []string, date time.Time) <-chan models.StreamResult[models.AccountTransation] {
	db := ar.r.extractTx(ctx)
	startDate, _ := atime.StartDateEndDate(date, date)
	ch := make(chan models.StreamResult[models.AccountTransation], 64) // slightly larger buffer
	wg := new(sync.WaitGroup)
	step := 2 * time.Hour // 2-hour chunks

	for i := 0; i < 24; i += 2 {
		start := startDate.Add(time.Duration(i) * time.Hour)
		end := start.Add(step)
		wg.Add(1)

		go func(start, end time.Time) {
			defer wg.Done()

			// entityPlaceholders, entityArgs := buildInClause(entities)
			// query := fmt.Sprintf(queryGetAccountTransactionByDate, entityPlaceholders)
			// args := append(entityArgs, start, end)
			rows, err := db.QueryContext(ctx, queryGetAccountTransactionByDate, start, end)
			if err != nil {
				ch <- models.StreamResult[models.AccountTransation]{Err: fmt.Errorf("query error %v - %v: %w", start, end, err)}
				return
			}
			defer rows.Close()

			for rows.Next() {
				select {
				case <-ctx.Done():
					return
				default:
					var value models.AccountTransation
					if err := rows.Scan(
						&value.AccountNumber,
						&value.Amount,
						&value.IsDebit,
					); err != nil {
						ch <- models.StreamResult[models.AccountTransation]{Err: fmt.Errorf("scan error: %w", err)}
						return
					}
					ch <- models.StreamResult[models.AccountTransation]{Data: value}
				}
			}

			if err := rows.Err(); err != nil {
				ch <- models.StreamResult[models.AccountTransation]{Err: fmt.Errorf("rows error %v - %v: %w", start, end, err)}
			}
		}(start, end)
	}

	go func() {
		wg.Wait()
		close(ch)
	}()

	return ch
}

func (ar *accountingRepository) GetAllAccountDailyBalance(ctx context.Context, entities []string, subCategories *[]models.SubCategory, date time.Time) <-chan models.StreamResult[models.AccountBalanceDaily] {
	db := ar.r.extractTx(ctx)

	ch := make(chan models.StreamResult[models.AccountBalanceDaily], 64) // Buffered to avoid blocking
	wg := new(sync.WaitGroup)
	// sem := make(chan struct{}, 128) // semaphore to limit concurrency

	for _, subCategory := range *subCategories {
		wg.Add(1)

		go func(wg *sync.WaitGroup, subCategoriesCode string) {
			defer wg.Done()

			// // Acquire semaphore slot
			// select {
			// case sem <- struct{}{}:
			// 	// Acquired
			// case <-ctx.Done():
			// 	return
			// }
			// // Release on exit
			// defer func() { <-sem }()

			entityPlaceholders, entityArgs := buildInClause(entities)
			query := fmt.Sprintf(queryGetAllAccountDailyBalance, entityPlaceholders)
			args := append(entityArgs, subCategoriesCode, date)
			rows, err := db.QueryContext(ctx, query, args...)
			if err != nil {
				ch <- models.StreamResult[models.AccountBalanceDaily]{
					Err: fmt.Errorf("query error for subCategoriesCode %s: %w", subCategoriesCode, err),
				}
				return
			}
			defer rows.Close()

			for rows.Next() {
				select {
				case <-ctx.Done():
					return
				default:
					var v models.AccountBalanceDaily
					if err := rows.Scan(
						&v.AccountNumber,
						&v.EntityCode,
						&v.CategoryCode,
						&v.SubCategoryCode,
						&v.DebitMovement,
						&v.CreditMovement,
						&v.OpeningBalance,
						&v.ClosingBalance,
					); err != nil {
						ch <- models.StreamResult[models.AccountBalanceDaily]{
							Err: fmt.Errorf("scan error for subCategoriesCode %s: %w", subCategoriesCode, err),
						}
						return
					}
					ch <- models.StreamResult[models.AccountBalanceDaily]{Data: v}
				}
			}

			if err := rows.Err(); err != nil {
				ch <- models.StreamResult[models.AccountBalanceDaily]{
					Err: fmt.Errorf("rows error for subCategoriesCode %s: %w", subCategoriesCode, err),
				}
			}
		}(wg, subCategory.Code)
	}

	go func() {
		wg.Wait()
		close(ch)
	}()

	return ch
}

func (ar *accountingRepository) InsertAccountBalanceDaily(ctx context.Context, in []models.AccountBalanceDaily) (err error) {
	start := atime.Now()
	defer func() {
		logSQL(ctx, err, start)
	}()

	db := ar.r.extractTx(ctx)

	valueStrings := []string{}
	valueArgs := []interface{}{}
	for _, req := range in {
		valueStrings = append(valueStrings, `(?, ?, ?, ?, ?, ?, ?, ?, ?)`)
		valueArgs = append(valueArgs, req.BalanceDate)
		valueArgs = append(valueArgs, req.AccountNumber)
		valueArgs = append(valueArgs, req.EntityCode)
		valueArgs = append(valueArgs, req.CategoryCode)
		valueArgs = append(valueArgs, req.SubCategoryCode)
		valueArgs = append(valueArgs, req.DebitMovement)
		valueArgs = append(valueArgs, req.CreditMovement)
		valueArgs = append(valueArgs, req.OpeningBalance)
		valueArgs = append(valueArgs, req.ClosingBalance)
	}

	query := fmt.Sprintf(queryInsertAccountBalanceDailyOld, strings.Join(valueStrings, ","))
	_, err = db.ExecContext(ctx, query, valueArgs...)
	if err != nil {
		err = databaseError(err)
		return
	}

	// affectedRows, err := res.RowsAffected()
	// if err != nil {
	// 	err = databaseError(err)
	// 	return
	// }

	// if affectedRows == 0 {
	// 	err = databaseError(models.ErrNoRowsAffected)
	// 	return
	// }

	return
}
