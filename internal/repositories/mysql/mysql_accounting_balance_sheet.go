package mysql

import (
	"context"
	"fmt"

	"bitbucket.org/Amartha/go-accounting/internal/models"
	"bitbucket.org/Amartha/go-accounting/internal/pkg/money"

	"github.com/shopspring/decimal"
)

func (ar *accountingRepository) GetBalanceSheet(ctx context.Context, opts models.BalanceSheetFilterOptions) (out models.BalanceSheetOut, err error) {
	defer func() {
		logSQL(ctx, err)
	}()

	query, args, err := buildGetBalanceSheetQuery(opts)
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

	out.BalanceSheet = make(map[string][]models.BalanceCategory)
	out.TotalPerCOAType = make(map[string]decimal.Decimal)

	for rows.Next() {
		var data = models.BalancePerCategoryOut{}
		var errScan = rows.Scan(
			&data.COAType,
			&data.CategoryCode,
			&data.CategoryName,
			&data.SumClosingBalance,
		)
		if errScan != nil {
			err = databaseError(errScan)
			return
		}

		out.BalanceSheet[data.COAType] = append(out.BalanceSheet[data.COAType], models.BalanceCategory{
			CategoryCode: data.CategoryCode,
			CategoryName: data.CategoryName,
			Amount:       money.FormatAmountToIDRFromDecimal(data.SumClosingBalance),
			DecAmount:    data.SumClosingBalance,
		})

		out.TotalPerCOAType[data.COAType] = decimal.Sum(out.TotalPerCOAType[data.COAType], data.SumClosingBalance)
	}
	if rows.Err() != nil {
		err = databaseError(rows.Err())
		return
	}

	return
}
