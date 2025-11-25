package mysql

import (
	"fmt"
	"time"

	"bitbucket.org/Amartha/go-accounting/internal/models"
	"bitbucket.org/Amartha/go-accounting/internal/pkg/atime"

	sq "github.com/Masterminds/squirrel"
)

var (
	queryCheckTransactionIdIsExist = `
	SELECT EXISTS (
		SELECT 1
		FROM transactions sa
		WHERE transaction_id = ? LIMIT 1
	) AS is_exist;`

	queryInsertJournalDetail = `
		INSERT INTO acct_journal_detail(journal_id, reference_number, order_type, transaction_type, transaction_type_name, transaction_date, is_debit, created_at, updated_at, metadata) VALUES %s`

	queryInsertTransaction = `
		INSERT INTO transactions(transaction_id, postdate, poster_user_id) VALUES %s`

	queryInsertSplit = `
		INSERT INTO splits(transaction_id, split_id, split_date, description, currency, amount) VALUES %s`

	queryInsertSplitAccount = `
		INSERT INTO split_accounts(split_id, account_id) VALUES %s`

	queryGetAccountTransactionByDate = `
	SELECT 		
		sa.account_id,
		s.amount,
		ajd.is_debit
	FROM splits s
	JOIN split_accounts sa ON sa.split_id = s.split_id
	JOIN acct_journal_detail ajd ON ajd.journal_id = s.split_id
	JOIN transactions t ON t.transaction_id = s.transaction_id
		WHERE ajd.transaction_date >= ? AND ajd.transaction_date < ?`

	queryGetOpeningBalanceDate = `
    SELECT 
        aadb.closing_balance
    FROM acct_account_daily_balance aadb
    WHERE aadb.account_number = ? AND aadb.balance_date = ?
    `
	queryGetOneAccountBalanceDaily = `
    SELECT 
		aadb.account_number,
		aadb.entity_code,
        aadb.category_code,
        aadb.sub_category_code,
        aadb.closing_balance,
        aadb.closing_balance
    FROM acct_account_daily_balance aadb
    WHERE aadb.account_number = ? AND aadb.balance_date = ?
    `

	queryGetLastOpeningBalance = `
    SELECT 
        aadb.closing_balance
    FROM acct_account_daily_balance aadb
    WHERE aadb.account_number = ? AND aadb.balance_date < ?
	ORDER BY aadb.balance_date DESC LIMIT 1
    `

	queryInsertAccountBalanceDailyOld = `
    INSERT INTO acct_account_daily_balance (
    balance_date,
    account_number,
    entity_code,
    category_code,
    sub_category_code,
    debit_movement,
    credit_movement,
    opening_balance,
    closing_balance
) VALUES %s AS new_data
ON DUPLICATE KEY UPDATE
    entity_code = new_data.entity_code,
	category_code = new_data.category_code,
    sub_category_code = new_data.sub_category_code,
    debit_movement = new_data.debit_movement,
    credit_movement = new_data.credit_movement,
    opening_balance = new_data.opening_balance,
    closing_balance = new_data.closing_balance,
	updated_at = CURRENT_TIMESTAMP(6)`

	queryInsertAccountBalanceDaily = `
    INSERT INTO acct_account_daily_balance (
    balance_date,
    account_number,
    entity_code,
    category_code,
    sub_category_code,
    debit_movement,
    credit_movement,
    opening_balance,
    closing_balance
) VALUES %s AS new_data
ON DUPLICATE KEY UPDATE
	entity_code = IF(
		acct_account_daily_balance.entity_code <> new_data.entity_code, 
		new_data.entity_code, 
		acct_account_daily_balance.entity_code
	),
	category_code = IF(
		acct_account_daily_balance.category_code <> new_data.category_code, 
		new_data.category_code, 
		acct_account_daily_balance.category_code
	),
	sub_category_code = IF(
		acct_account_daily_balance.sub_category_code <> new_data.sub_category_code, 
		new_data.sub_category_code, 
		acct_account_daily_balance.sub_category_code
	),
    debit_movement = IF(
        acct_account_daily_balance.debit_movement <> new_data.debit_movement,
        new_data.debit_movement,
        acct_account_daily_balance.debit_movement
    ),
    credit_movement = IF(
        acct_account_daily_balance.credit_movement <> new_data.credit_movement,
        new_data.credit_movement,
        acct_account_daily_balance.credit_movement
    ),
    opening_balance = IF(
        acct_account_daily_balance.opening_balance <> new_data.opening_balance,
        new_data.opening_balance,
        acct_account_daily_balance.opening_balance
    ),
    closing_balance = IF(
        acct_account_daily_balance.closing_balance <> new_data.closing_balance,
        new_data.closing_balance,
        acct_account_daily_balance.closing_balance
    ),
    updated_at = IF(
        acct_account_daily_balance.closing_balance <> new_data.closing_balance,
        CURRENT_TIMESTAMP(6),
        acct_account_daily_balance.updated_at
    );`

	queryGetCategorySubCategoryCOAType = `
    SELECT 
        ac.code as category_code,
        ac.name as category_name,
        asc2.code as sub_category_code,
        asc2.name as sub_category_name,
        act.code as coa_type_code,
        act.coa_type_name
    FROM acct_coa_type act
        JOIN acct_category ac ON ac.coa_type_code = act.code
        JOIN acct_sub_category asc2 ON asc2.category_code = ac.code
    ORDER BY asc2.code ASC
    `

	queryGetOpeingBalanceFromAccountTrialBalance = `
    SELECT
        aatb.closing_balance
    FROM acct_account_trial_balance aatb
    WHERE aatb.entity_code = ? 
		AND aatb.sub_category_code = ? 
		AND aatb.closing_date = ?`

	queryCalculateOpeningClosingBalanceFromAccountBalance = `
	SELECT 
		COALESCE(SUM(aadb.opening_balance), 0) opening_balance
	FROM acct_account_daily_balance aadb
	WHERE aadb.entity_code = ?
	AND aadb.sub_category_code = ?
	AND aadb.balance_date = ?`

	queryInsertAccountTrialBalance = `
    INSERT INTO acct_account_trial_balance(
        closing_date,
		entity_code,
        category_code,
        sub_category_code,
        debit_movement,
        credit_movement,
        opening_balance,
        closing_balance) VALUES %s 
		AS new_data
        ON DUPLICATE KEY UPDATE 
            closing_date = new_data.closing_date,
            entity_code = new_data.entity_code,
			category_code = new_data.category_code,
            sub_category_code = new_data.sub_category_code,
            debit_movement = new_data.debit_movement,
            credit_movement = new_data.credit_movement,
            opening_balance = new_data.opening_balance,
            closing_balance = new_data.closing_balance,
			updated_at = CURRENT_TIMESTAMP(6)
        `

	queryGetAllAccountDailyBalance = `
    SELECT 
		aadb.account_number,
		aadb.entity_code,
        aadb.category_code,
        aadb.sub_category_code,
		aadb.debit_movement,
		aadb.credit_movement,
        aadb.opening_balance,
        aadb.closing_balance
    FROM acct_account_daily_balance aadb
    WHERE aadb.entity_code IN (%s) AND aadb.sub_category_code = ? AND aadb.balance_date = ?
    `

	queryGetOneSplitAccount = `
	SELECT EXISTS (
		SELECT 1
		FROM split_accounts sa
		WHERE sa.account_id = ? LIMIT 1
	) AS is_exist;`
)

/*
SELECT
COALESCE(act.id, ”) as coa_id,
COALESCE(act.coa_type_name, ”) as coa_type_name,
ac.code as category_code,
ac.name as category_name,
asc2.code as sub_category_code,
asc2.name as sub_category_name,
SUM(CASE WHEN s.amount>0 and t.postdate >= '2023-10-01 00:00:00' and t.postdate <= '2023-10-31 23:59:59' THEN s.amount ELSE 0 END) as debit,
SUM(CASE WHEN s.amount<0 and t.postdate >= '2023-10-01 00:00:00' and t.postdate <= '2023-10-31 23:59:59' THEN s.amount ELSE 0 END) as credit,
sum(case when s.amount>0 and t.postdate <= '2023-10-01 00:00:00' then s.amount ELSE 0 END) +
sum(case when s.amount<0 and t.postdate <= '2023-10-01 00:00:00' then s.amount ELSE 0 END)
as opening,
sum(case when s.amount>0 and t.postdate <= '2023-10-31 23:59:59' then s.amount ELSE 0 END) +
sum(case when s.amount<0 and t.postdate <= '2023-10-31 23:59:59' then s.amount ELSE 0 END)
as closing
FROM acct_category ac
INNER JOIN acct_sub_category asc2 ON ac.code = asc2.category_code
LEFT JOIN acct_coa_type act ON act.category_code  = ac.code
LEFT JOIN acct_account aa ON asc2.code = aa.sub_category_code
LEFT JOIN split_accounts sa ON sa.account_id = aa.account_number
LEFT JOIN splits s ON s.split_id = sa.split_id
LEFT JOIN transactions t ON t.transaction_id = s.transaction_id
GROUP BY asc2.code,ac.code,act.id,act.coa_type_name;
*/
func buildGetTrialBalanceQuery(opts models.TrialBalanceFilterOptions) (sql string, args []interface{}, err error) {
	columns := []string{
		`COALESCE(act.code, '') as coa_id`,
		`COALESCE(act.coa_type_name, '') as coa_type_name`,
		`ac.code as category_code`,
		`ac.name as category_name`,
		`asc2.code as sub_category_code`,
		`asc2.name as sub_category_name`,
	}
	query := buildFilteredTrialBalanceQuery(columns, opts)

	return query.ToSql()
}

func buildFilteredTrialBalanceQuery(cols []string, opts models.TrialBalanceFilterOptions) sq.SelectBuilder {
	psql := sq.StatementBuilder.PlaceholderFormat(sq.Question)
	query := psql.Select(cols...).From("acct_category ac")
	query = query.InnerJoin("acct_sub_category asc2 ON ac.code = asc2.category_code")
	query = query.LeftJoin("acct_coa_type act ON act.code = ac.coa_type_code")
	query = query.LeftJoin("acct_account aa ON aa.sub_category_code = asc2.code")
	query = query.LeftJoin("split_accounts sa ON sa.account_id = aa.account_number")
	query = query.LeftJoin("splits s ON s.split_id = sa.split_id")
	query = query.LeftJoin("acct_journal_detail dtl on dtl.journal_id = s.split_id")
	query = query.LeftJoin("transactions t ON t.transaction_id = s.transaction_id")

	query = query.Column(
		`SUM(CASE WHEN t.postdate >= ? and t.postdate <= ? and dtl.is_debit=1 THEN s.amount ELSE 0 END) as debit`,
		opts.StartDate, opts.EndDate)
	query = query.Column(
		`SUM(CASE WHEN t.postdate >= ? and t.postdate <= ? and dtl.is_debit=0 THEN s.amount ELSE 0 END) as credit`,
		opts.StartDate, opts.EndDate)

	query = query.Column(
		`SUM(CASE WHEN t.postdate < ? and (dtl.is_debit=1 and act.code = 'AST' or dtl.is_debit=0 and act.code = 'LIA') then s.amount ELSE 0 END) - SUM(CASE WHEN t.postdate < ? and (dtl.is_debit=0 and act.code = 'AST' or dtl.is_debit=1 and act.code = 'LIA') then s.amount ELSE 0 END) as opening`,
		opts.StartDate, opts.StartDate)
	query = query.Column(
		`SUM(CASE WHEN t.postdate <= ? and (dtl.is_debit=1 and act.code = 'AST' or dtl.is_debit=0 and act.code = 'LIA') then s.amount ELSE 0 END) - SUM(CASE WHEN t.postdate <= ? and (dtl.is_debit=0 and act.code = 'AST' or dtl.is_debit=1 and act.code = 'LIA') then s.amount ELSE 0 END) as closing`,
		opts.EndDate, opts.EndDate)

	query = query.Where(sq.Eq{`aa.entity_code`: opts.EntityCode})
	query = query.GroupBy(`ac.code, asc2.code, act.code, act.coa_type_name`)
	query = query.OrderBy(`act.code, ac.code, asc2.code ASC`)

	return query
}

/*
SELECT

	t.transaction_id,
	COALESCE(ajd.reference_number, ''),
	ajd.transaction_date,
	ajd.transaction_type,
	COALESCE(ajd.transaction_type_name, ''),
	COALESCE(s.description, '') narrative,
	case when ajd.is_debit = TRUE then s.amount else 0 end as debit,
	case when ajd.is_debit = FALSE then s.amount else 0 end as credit,
	ajd.created_at,
	ajd.updated_at

FROM splits s
LEFT JOIN transactions t ON t.transaction_id = s.transaction_id
LEFT JOIN acct_journal_detail ajd ON ajd.journal_id = s.split_id
LEFT JOIN split_accounts sa ON s.split_id = sa.split_id
WHERE sa.account_id = '121001000000003'
AND ajd.transaction_date >= '2024-05-01 00:00:00'
AND ajd.transaction_date <= '2024-05-30 23:59:59.999999999'
ORDER BY ajd.transaction_date ASC
LIMIT 11
*/
func buildSubLedgerQuery(opts models.SubLedgerFilterOptions) (sql string, args []interface{}, err error) {
	columns := []string{
		`COALESCE(t.transaction_id, '')`,
		`COALESCE(ajd.reference_number, '')`,
		`ajd.transaction_date`,
		`ajd.order_type`,
		`ajd.transaction_type`,
		`COALESCE(s.description, '') narrative`,
		`COALESCE(ajd.metadata, '{}') metadata`,
		`case when ajd.is_debit = TRUE then s.amount else 0 end as debit`,
		`case when ajd.is_debit = FALSE then s.amount else 0 end as credit`,
		`ajd.created_at`,
		`ajd.updated_at`,
	}
	query := buildFilteredSubLedgerQuery(columns, opts)

	if opts.AfterCreatedAt != nil {
		query = query.Where(sq.Lt{`ajd.transaction_date`: opts.AfterCreatedAt})
	}
	if opts.BeforeCreatedAt != nil {
		query = query.Where(sq.Gt{`ajd.transaction_date`: opts.BeforeCreatedAt})
	}

	if opts.AscendingOrder {
		query = query.OrderBy(`ajd.transaction_date DESC`)
	} else {
		query = query.OrderBy(`ajd.transaction_date ASC`)
	}

	if opts.Limit > 0 {
		query = query.Limit(uint64(opts.Limit))
	}
	if opts.Offset > 0 {
		query = query.Offset(uint64(opts.Offset))
	}
	return query.ToSql()
}

func buildFilteredSubLedgerQuery(cols []string, opts models.SubLedgerFilterOptions) sq.SelectBuilder {
	psql := sq.StatementBuilder.PlaceholderFormat(sq.Question)
	query := psql.Select(cols...).From("splits s")
	query = query.LeftJoin("acct_journal_detail ajd ON ajd.journal_id = s.split_id")
	query = query.LeftJoin("split_accounts sa ON sa.split_id = s.split_id")
	query = query.LeftJoin("transactions t ON t.transaction_id = s.transaction_id")
	query = query.Where(sq.Eq{`sa.account_id`: opts.AccountNumber})
	query = query.Where(sq.GtOrEq{`ajd.transaction_date`: opts.StartDate})
	query = query.Where(sq.LtOrEq{`ajd.transaction_date`: opts.EndDate})
	return query
}

func buildCountSubLedgerQuery(opts models.SubLedgerFilterOptions) (sql string, args []interface{}, err error) {
	columns := []string{`count(*)`}

	query := buildFilteredCountSubLedgerQuery(columns, opts)
	return query.ToSql()
}

func buildFilteredCountSubLedgerQuery(cols []string, opts models.SubLedgerFilterOptions) sq.SelectBuilder {
	psql := sq.StatementBuilder.PlaceholderFormat(sq.Question)
	subQuery := psql.Select([]string{`1`}...).From("acct_journal_detail ajd")
	subQuery = subQuery.Join("split_accounts sa ON sa.split_id = ajd.journal_id")
	subQuery = subQuery.Where(sq.Eq{`sa.account_id`: opts.AccountNumber})
	subQuery = subQuery.Where(sq.GtOrEq{`ajd.transaction_date`: opts.StartDate})
	subQuery = subQuery.Where(sq.LtOrEq{`ajd.transaction_date`: opts.EndDate})
	subQuery = subQuery.Limit(uint64(opts.Limit))

	query := psql.Select(cols...).FromSelect(subQuery, "limited_count")

	return query
}

/*
SELECT

	SUM(CASE WHEN t.postdate < '2024-05-01 00:00:00' and (ajd.is_debit=1 and ac.coa_type_code = 'AST' or ajd.is_debit=0 and ac.coa_type_code = 'LIA') then s.amount ELSE 0 END) -
	SUM(CASE WHEN t.postdate < '2024-05-01 00:00:00' and (ajd.is_debit=0 and ac.coa_type_code = 'AST' or ajd.is_debit=1 and ac.coa_type_code = 'LIA') then s.amount ELSE 0 END) as balance_period_start

FROM acct_account aa
INNER JOIN acct_category ac ON ac.code = aa.category_code
LEFT JOIN split_accounts sa ON sa.account_id = aa.account_number
LEFT JOIN splits s ON s.split_id = sa.split_id
LEFT JOIN transactions t ON t.transaction_id = s.transaction_id
LEFT JOIN acct_journal_detail ajd ON ajd.journal_id = s.split_id
WHERE aa.account_number = '121001000000003'
*/
func getAccountBalancePeriodStart(accountNumber string, date time.Time) (sql string, args []interface{}, err error) {
	psql := sq.StatementBuilder.PlaceholderFormat(sq.Question)
	query := psql.Select([]string{}...).From("acct_account aa")
	query = query.InnerJoin("acct_category ac ON ac.code = aa.category_code")
	query = query.LeftJoin("split_accounts sa ON sa.account_id = aa.account_number")
	query = query.LeftJoin("splits s ON s.split_id = sa.split_id")
	query = query.LeftJoin("transactions t ON t.transaction_id = s.transaction_id")
	query = query.LeftJoin("acct_journal_detail ajd ON ajd.journal_id = s.split_id")
	query = query.Column(`
		SUM(CASE WHEN t.postdate < ? and (ajd.is_debit=1 and ac.coa_type_code = 'AST' or ajd.is_debit=0 and ac.coa_type_code = 'LIA') then s.amount ELSE 0 END) - 
		SUM(CASE WHEN t.postdate < ? and (ajd.is_debit=0 and ac.coa_type_code = 'AST' or ajd.is_debit=1 and ac.coa_type_code = 'LIA') then s.amount ELSE 0 END) as balance_period_start`,
		date,
		date,
	)
	query = query.Where(sq.Eq{`aa.account_number`: accountNumber})

	return query.ToSql()
}

func getJournalDetailQuery(transactionId string) (sql string, args []interface{}, err error) {
	subQuery := sq.StatementBuilder.PlaceholderFormat(sq.Question).Select([]string{
		`aa.account_number`,
		`coalesce(aa.name, '') name`,
		`aa.alt_id`,
		`aa.entity_code`,
		`ae.name as entity_name`,
		`aa.sub_category_code`,
		`asc2.name as sub_category_name`,
	}...).From("acct_account aa")
	subQuery = subQuery.Join("acct_entity ae ON ae.code = aa.entity_code")
	subQuery = subQuery.Join("acct_sub_category asc2 ON asc2.code = aa.sub_category_code")

	psql := sq.StatementBuilder.PlaceholderFormat(sq.Question)
	query := psql.Select([]string{
		`t.transaction_id`,
		`ajd.journal_id`,
		`aa.account_number`,
		`COALESCE(aa.name, '') account_name`,
		`COALESCE(aa.alt_id, '') alt_id`,
		`aa.entity_code`,
		`aa.entity_name`,
		`aa.sub_category_code`,
		`aa.sub_category_name`,
		`ajd.transaction_type`,
		`s.amount`,
		`ajd.transaction_date`,
		`COALESCE(s.description, '') AS narrative`,
		`ajd.is_debit`,
	}...).FromSelect(subQuery, "aa")
	query = query.Join("split_accounts sa ON sa.account_id = aa.account_number")
	query = query.Join("splits s ON s.split_id = sa.split_id")
	query = query.Join("acct_journal_detail ajd ON ajd.journal_id = s.split_id")
	query = query.Join("transactions t ON t.transaction_id = s.transaction_id")
	query = query.Where(sq.Eq{`t.transaction_id `: transactionId})
	query = query.OrderBy(`ajd.journal_id ASC`)

	return query.ToSql()
}

func buildSubLedgerAccountsQuery(opts models.SubLedgerAccountsFilterOptions) (sql string, args []interface{}, err error) {
	query := sq.StatementBuilder.PlaceholderFormat(sq.Question).
		Select([]string{
			`aa.account_number`,
			`COALESCE(aa.name, '') account_name`,
			`coalesce(aa.alt_id, "") alt_id`,
			`aa.sub_category_code`,
			`0 as number_of_data`,
			`aa.created_at`,
		}...).From("acct_account aa")
	if opts.Search != "" {
		query = query.Where(sq.Eq{fmt.Sprintf(`aa.%s`, opts.SearchBy): opts.Search})
	}
	query = query.Where(sq.Eq{`aa.entity_code`: opts.EntityCode})

	if opts.AfterCreatedAt != nil {
		query = query.Where(sq.Gt{`aa.created_at`: opts.AfterCreatedAt})
	}
	if opts.BeforeCreatedAt != nil {
		query = query.Where(sq.Lt{`aa.created_at`: opts.BeforeCreatedAt})
	}

	if opts.GuestMode {
		query = query.Where(sq.Eq{"aa.sub_category_code": models.ExcludedSubCategoryForGuestMode})
	}

	if opts.AscendingOrder {
		query = query.OrderBy(`aa.created_at ASC`)
	} else {
		query = query.OrderBy(`aa.created_at DESC`)
	}

	if opts.Limit > 0 {
		query = query.Limit(uint64(opts.Limit))
	}

	return query.ToSql()
}

func buildCountSubLedgerAccountsQuery(opts models.SubLedgerAccountsFilterOptions) (sql string, args []interface{}, err error) {
	query := sq.StatementBuilder.PlaceholderFormat(sq.Question).
		Select([]string{`count(1)`}...).
		From("acct_account aa")
	query = query.Where(sq.Eq{`aa.entity_code`: opts.EntityCode})

	return query.ToSql()
}

func buildSubLedgerAccountTotalTransactionQuery(opts models.SubLedgerAccountsFilterOptions) (sql string, args []interface{}, err error) {
	opts.EndDate = time.Date(opts.EndDate.Year(), opts.EndDate.Month(), opts.EndDate.Day(), 23, 59, 59, 0, opts.EndDate.Location())
	subQuery := sq.StatementBuilder.PlaceholderFormat(sq.Question).
		Select([]string{
			`aa.account_number`,
		}...).
		From("acct_account aa")
	subQuery = subQuery.Where(sq.Eq{`aa.entity_code`: opts.EntityCode})
	subQuery = subQuery.Where(sq.Eq{fmt.Sprintf(`aa.%s`, opts.SearchBy): opts.Search})

	query := sq.StatementBuilder.PlaceholderFormat(sq.Question).
		Select([]string{
			`count(t.transaction_id) as number_of_data`,
		}...).
		FromSelect(subQuery, "aa")
	query = query.Join("split_accounts sa ON sa.account_id = aa.account_number")
	query = query.Join("splits s ON s.split_id = sa.split_id")
	query = query.Join("transactions t ON t.transaction_id = s.transaction_id")
	query = query.Where(sq.GtOrEq{`t.postdate `: opts.StartDate})
	query = query.Where(sq.LtOrEq{`t.postdate `: opts.EndDate})

	return query.ToSql()
}

func buildCalculateOpeningClosingBalance(accountNumber string, date time.Time) (sql string, args []interface{}, err error) {
	psql := sq.StatementBuilder.PlaceholderFormat(sq.Question)
	_, end := atime.StartDateEndDate(date, date)

	subQuery := psql.Select([]string{
		`aa.account_number`,
		`aa.entity_code`,
		`aa.category_code`,
		`aa.sub_category_code`,
		`ac.coa_type_code`,
	}...).From("acct_account aa")
	subQuery = subQuery.Join("acct_category ac ON ac.code = aa.category_code")
	subQuery = subQuery.Where(sq.Eq{`aa.account_number`: accountNumber})

	query := psql.Select([]string{
		`ac.account_number`,
		`ac.entity_code`,
		`ac.category_code`,
		`ac.sub_category_code`,
	}...).FromSelect(subQuery, "ac")
	query = query.Join("split_accounts sa ON sa.account_id = ac.account_number")
	query = query.Join("splits s ON s.split_id = sa.split_id")
	query = query.Join("transactions t ON t.transaction_id = s.transaction_id")
	query = query.Join("acct_journal_detail ajd on ajd.journal_id = s.split_id")
	query = query.Where(sq.Eq{`ac.account_number`: accountNumber})
	query = query.GroupBy(`ac.account_number`)

	query = query.Column(
		`SUM(CASE WHEN t.postdate < ? AND (ajd.is_debit=1 AND ac.coa_type_code = 'AST' OR ajd.is_debit=0 AND ac.coa_type_code = 'LIA') THEN s.amount ELSE 0 END) - SUM(CASE WHEN t.postdate < ? AND (ajd.is_debit=0 AND ac.coa_type_code = 'AST' OR ajd.is_debit=1 AND ac.coa_type_code = 'LIA') THEN s.amount ELSE 0 END) opening_balance`,
		end, end)
	query = query.Column(
		`SUM(CASE WHEN t.postdate <= ? AND (ajd.is_debit=1 AND ac.coa_type_code = 'AST' OR ajd.is_debit=0 AND ac.coa_type_code = 'LIA') THEN s.amount ELSE 0 END) - SUM(CASE WHEN t.postdate <= ? AND (ajd.is_debit=0 AND ac.coa_type_code = 'AST' OR ajd.is_debit=1 AND ac.coa_type_code = 'LIA') THEN s.amount ELSE 0 END) opening_balance`,
		end, end)

	return query.ToSql()
}

func buildCalculateFromAccountBalanceDaily(in models.CalculateTrialBalance) (sql string, args []interface{}, err error) {
	psql := sq.StatementBuilder.PlaceholderFormat(sq.Question)
	query := psql.Select([]string{
		`aadb.entity_code`,
		`aadb.category_code`,
		`aadb.sub_category_code`,
		`SUM(aadb.debit_movement) debit_movement`,
		`SUM(aadb.credit_movement) credit_movement`,
	}...).From("acct_account_daily_balance aadb")
	query = query.Where(sq.Eq{`aadb.entity_code`: in.EntityCode})
	query = query.Where(sq.Eq{`aadb.sub_category_code`: in.SubCategoryCode})
	query = query.Where(sq.Eq{`aadb.balance_date`: in.Date})
	query = query.GroupBy(`aadb.entity_code, aadb.category_code, sub_category_code`)

	return query.ToSql()
}

func buildCalculateOpeningClosingBalanceFromTransactions(in models.CalculateTrialBalance) (sql string, args []interface{}, err error) {
	_, end := atime.StartDateEndDate(in.Date, in.Date)

	psql := sq.StatementBuilder.PlaceholderFormat(sq.Question)
	query := psql.Select([]string{
		`aa.sub_category_code`,
	}...).From("acct_category ac")
	query = query.Join("acct_account aa ON aa.category_code = ac.code")
	query = query.LeftJoin("split_accounts sa ON sa.account_id = aa.account_number")
	query = query.LeftJoin("splits s ON s.split_id = sa.split_id")
	query = query.LeftJoin("acct_journal_detail ajd on ajd.journal_id = s.split_id")
	query = query.LeftJoin("transactions t ON t.transaction_id = s.transaction_id")
	query = query.Where(sq.Eq{`aa.sub_category_code`: in.SubCategoryCode})
	query = query.Where(sq.Eq{`aa.entity_code`: in.EntityCode})
	query = query.GroupBy(`aa.sub_category_code`)

	query = query.Column(
		`SUM(CASE WHEN t.postdate < ? AND (ajd.is_debit=1 AND ac.coa_type_code = 'AST' OR ajd.is_debit=0 AND ac.coa_type_code = 'LIA') THEN s.amount ELSE 0 END) - SUM(CASE WHEN t.postdate < ? AND (ajd.is_debit=0 AND ac.coa_type_code = 'AST' OR ajd.is_debit=1 AND ac.coa_type_code = 'LIA') THEN s.amount ELSE 0 END) opening_balance`,
		end, end)
	query = query.Column(
		`SUM(CASE WHEN t.postdate <= ? AND (ajd.is_debit=1 AND ac.coa_type_code = 'AST' OR ajd.is_debit=0 AND ac.coa_type_code = 'LIA') THEN s.amount ELSE 0 END) - SUM(CASE WHEN t.postdate <= ? AND (ajd.is_debit=0 AND ac.coa_type_code = 'AST' OR ajd.is_debit=1 AND ac.coa_type_code = 'LIA') THEN s.amount ELSE 0 END) opening_balance`,
		end, end)
	return query.ToSql()
}

func buildCalculateFromTransactions(in models.CalculateTrialBalance) (sql string, args []interface{}, err error) {
	start, end := atime.StartDateEndDate(in.Date, in.Date)

	psql := sq.StatementBuilder.PlaceholderFormat(sq.Question)
	query := psql.Select([]string{
		`aa.entity_code`,
		`aa.category_code`,
		`aa.sub_category_code`,
	}...).From("acct_category ac")
	query = query.Join("acct_account aa ON aa.category_code = ac.code")
	query = query.LeftJoin("split_accounts sa ON sa.account_id = aa.account_number")
	query = query.LeftJoin("splits s ON s.split_id = sa.split_id")
	query = query.LeftJoin("acct_journal_detail ajd on ajd.journal_id = s.split_id")
	query = query.LeftJoin("transactions t ON t.transaction_id = s.transaction_id")
	query = query.Where(sq.Eq{`aa.sub_category_code`: in.SubCategoryCode})
	query = query.Where(sq.Eq{`aa.entity_code`: in.EntityCode})
	query = query.GroupBy(`aa.entity_code, aa.category_code, aa.sub_category_code`)

	query = query.Column(
		`SUM(CASE WHEN t.postdate >= ? AND t.postdate <= ? AND ajd.is_debit=1 THEN s.amount ELSE 0 END) debit_movement`, start, end)
	query = query.Column(
		`SUM(CASE WHEN t.postdate >= ? AND t.postdate <= ? and ajd.is_debit=0 THEN s.amount ELSE 0 END) credit_movement`, start, end)
	return query.ToSql()
}

/*
buildGetListTrialBalanceDetailQuery create tb accounts query
finalize query:

	with tb_account as (
		select
			account_number,
			name
		from acct_account
		order by account_number
		limit 10
	)
	select
		tb_account.account_number,
		tb_account.name,
		sum(debit_movement) as debit_movement,
		sum(credit_movement) as credit_movement,
		substring_index(group_concat(cast(opening_balance as CHAR) order by balance_date), ',', 1 ) as first_opening_balance,
		substring_index(group_concat(cast(closing_balance as CHAR) order by balance_date desc), ',', 1 ) as last_closing_balance
	from
		tb_account
	left join acct_account_daily_balance balance_daily on tb_account.account_number = balance_daily.account_number and balance_date >= '2024-04-30 17:00:00' and balance_date < '2024-08-30 17:00:00'
	group by tb_account.account_number, tb_account.name;
*/
func buildGetListTrialBalanceDetailQuery(opts models.TrialBalanceDetailsFilterOptions) (sql string, args []interface{}, err error) {
	start := atime.ToZeroTime(opts.StartDate)
	end := atime.ToZeroTime(opts.EndDate)

	psql := sq.StatementBuilder.PlaceholderFormat(sq.Question)

	cteAccount := psql.
		Select("account_number", "coalesce(name,'') as name").
		From("acct_account")

	if opts.Search != "" {
		cteAccount = cteAccount.
			Where(sq.Eq{"account_number": opts.Search})
	}

	if opts.SubCategoryCode != "" {
		cteAccount = cteAccount.
			Where(sq.Eq{"sub_category_code": opts.SubCategoryCode})
	}

	if opts.EntityCode != "" {
		cteAccount = cteAccount.
			Where(sq.Eq{"entity_code": opts.EntityCode})
	}

	if opts.CursorValue != nil {
		if opts.IsBackward {
			cteAccount = cteAccount.Where(sq.Lt{"account_number": opts.CursorValue})
		} else {
			cteAccount = cteAccount.Where(sq.Gt{"account_number": opts.CursorValue})
		}
	}

	if opts.CursorValue != nil && opts.IsBackward {
		cteAccount = cteAccount.OrderBy("account_number DESC")
	} else {
		cteAccount = cteAccount.OrderBy("account_number ASC")
	}

	if opts.Limit > 0 {
		cteAccount = cteAccount.Limit(uint64(opts.Limit))
	}

	cteQuery, cteArgs, cteErr := cteAccount.ToSql()
	if cteErr != nil {
		return "", nil, cteErr
	}

	aggregatedCols := []string{
		"tb_account.account_number",
		"tb_account.name",
		"substring_index(group_concat(cast(coalesce(opening_balance, 0) as CHAR) order by balance_date), ',', 1 ) as first_opening_balance",
		"substring_index(group_concat(cast(coalesce(closing_balance, 0) as CHAR) order by balance_date desc), ',', 1 ) as last_closing_balance",
		"sum(coalesce(debit_movement, 0)) as debit_movement",
		"sum(coalesce(credit_movement, 0)) as credit_movement",
	}

	query := psql.Select(aggregatedCols...).
		Prefix(fmt.Sprintf("with tb_account as (%s) ", cteQuery), cteArgs...).
		From("tb_account").
		LeftJoin("acct_account_daily_balance balance_daily "+
			"on tb_account.account_number = balance_daily.account_number and "+
			"balance_date >= ? and "+
			"balance_date <= ?", start, end).
		GroupBy("tb_account.account_number", "tb_account.name")

	if opts.CursorValue != nil && opts.IsBackward {
		query = query.OrderBy("tb_account.account_number DESC")
	} else {
		query = query.OrderBy("tb_account.account_number ASC")
	}

	return query.ToSql()
}

func buildGetTrialBalanceV2Query(opts models.TrialBalanceFilterOptions) (sql string, args []interface{}, err error) {
	start := atime.ToZeroTime(opts.StartDate)
	end := atime.ToZeroTime(opts.EndDate)

	psql := sq.StatementBuilder.PlaceholderFormat(sq.Question)
	query := psql.Select([]string{
		`aatb.entity_code`,
		`aatb.category_code`,
		`aatb.sub_category_code`,
		`SUM(aatb.debit_movement) debit_movement`,
		`SUM(aatb.credit_movement) credit_movement`,
	}...).From("acct_account_trial_balance aatb")
	query = query.Column(
		`SUM(CASE WHEN aatb.closing_date = ? THEN aatb.opening_balance ELSE 0 END) opening_balance`,
		start)
	query = query.Column(
		`SUM(CASE WHEN aatb.closing_date = ? THEN aatb.closing_balance ELSE 0 END) closing_balance`,
		end)

	query = query.Where(sq.Eq{`aatb.entity_code`: opts.EntityCode})
	query = query.Where(sq.GtOrEq{`aatb.closing_date`: start})
	query = query.Where(sq.LtOrEq{`aatb.closing_date`: end})
	query = query.Where(`EXISTS (SELECT 1 FROM acct_account aa WHERE aa.sub_category_code = aatb.sub_category_code AND aa.entity_code = aatb.entity_code)`)
	query = query.GroupBy(`aatb.entity_code, aatb.category_code, aatb.sub_category_code`)
	query = query.OrderBy(`aatb.sub_category_code ASC`)

	return query.ToSql()
}

func buildGetTrialBalanceSubCategory(opts models.TrialBalanceFilterOptions) (sql string, args []interface{}, err error) {
	start := atime.ToZeroTime(opts.StartDate)
	end := atime.ToZeroTime(opts.EndDate)

	psql := sq.StatementBuilder.PlaceholderFormat(sq.Question)
	query := psql.
		Select([]string{
			"sub_category.code",
			"sub_category.name",
			"SUM(coalesce(aatb.debit_movement, 0)) debit_movement",
			"SUM(coalesce(aatb.credit_movement, 0)) credit_movement"}...,
		).
		Column(`SUM(CASE WHEN aatb.closing_date = ? THEN aatb.opening_balance ELSE 0 END) opening_balance`, start).
		Column(`SUM(CASE WHEN aatb.closing_date = ? THEN aatb.closing_balance ELSE 0 END) closing_balance`, end).
		From("acct_sub_category sub_category")

	if opts.EntityCode != "" {
		query = query.
			LeftJoin("acct_account_trial_balance aatb on "+
				"sub_category.code = aatb.sub_category_code and "+
				"aatb.entity_code = ? and "+
				"aatb.closing_date >= ? and "+
				"aatb.closing_date <= ?", opts.EntityCode, start, end)
	} else {
		query = query.
			LeftJoin("acct_account_trial_balance aatb on "+
				"sub_category.code = aatb.sub_category_code and "+
				"aatb.closing_date >= ? and "+
				"aatb.closing_date <= ?", start, end)
	}

	if opts.SubCategoryCode != "" {
		query = query.Where(sq.Eq{`sub_category.code`: opts.SubCategoryCode})
	}

	query = query.GroupBy("sub_category.code", "sub_category.name")

	return query.ToSql()
}

func buildGetTransactionsToday(transactionDate time.Time) (sql string, args []interface{}, err error) {
	startTrx, endTrx := atime.StartDateEndDate(transactionDate, transactionDate)
	startCreated, endCreated := atime.StartDateEndDate(atime.Now(), atime.Now())

	psql := sq.StatementBuilder.PlaceholderFormat(sq.Question)
	query := psql.Select([]string{
		`s.transaction_id`,
	}...).From("splits s")
	query = query.Join("transactions t ON t.transaction_id = s.transaction_id")
	query = query.Join("split_accounts sa ON sa.split_id = s.split_id")
	query = query.Join("acct_journal_detail ajd on ajd.journal_id = s.split_id")
	query = query.Where(sq.GtOrEq{`ajd.transaction_date `: startTrx})
	query = query.Where(sq.LtOrEq{`ajd.transaction_date `: endTrx})
	query = query.Where(sq.GtOrEq{`ajd.created_at `: startCreated})
	query = query.Where(sq.LtOrEq{`ajd.created_at `: endCreated})

	return query.ToSql()
}
