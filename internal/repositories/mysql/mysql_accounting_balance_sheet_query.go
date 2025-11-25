package mysql

import (
	"bitbucket.org/Amartha/go-accounting/internal/models"
	"bitbucket.org/Amartha/go-accounting/internal/pkg/atime"

	sq "github.com/Masterminds/squirrel"
)

func buildGetBalanceSheetQuery(opts models.BalanceSheetFilterOptions) (sql string, args []interface{}, err error) {
	// Change to UTC
	date := atime.ToZeroTime(opts.BalanceSheetDate)
	
	psql := sq.StatementBuilder.PlaceholderFormat(sq.Question)
	query := psql.Select([]string{
		"ac.coa_type_code",
		"aatb.category_code",
		"ac.name",
		"sum(aatb.closing_balance)",
	}...).From("acct_account_trial_balance aatb").
		Join("acct_category ac on aatb.category_code = ac.code").
		Where(sq.Eq{"aatb.entity_code": opts.EntityCode}).
		Where(sq.Eq{"aatb.closing_date": date}).
		Where(`EXISTS (SELECT 1 FROM acct_account aa WHERE aa.entity_code = aatb.entity_code AND aa.sub_category_code = aatb.sub_category_code)`).
		GroupBy("aatb.category_code")

	return query.ToSql()
}
