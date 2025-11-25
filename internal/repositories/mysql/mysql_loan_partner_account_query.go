package mysql

import (
	"bitbucket.org/Amartha/go-accounting/internal/models"
	"bitbucket.org/Amartha/go-accounting/internal/pkg/atime"

	sq "github.com/Masterminds/squirrel"
)

// query to acct_loan_partner_account table
var (
	queryLoanPartnerAccountCreate = `
	INSERT INTO acct_loan_partner_account(
		partner_id,
		loan_kind,
		account_number,
		account_type,
		entity_code,
		loan_sub_category_code
	)
	VALUES(?,?,?,?,?,?)`

	queryBulkLoanPartnerAccountCreate = `
	INSERT INTO acct_loan_partner_account(
		partner_id,
		loan_kind,
		account_number,
		account_type,
		entity_code,
		loan_sub_category_code
	)
	VALUES %s`
)

func queryAccountLoanPartnerUpdate(in models.UpdateLoanPartnerAccount) (sql string, args []interface{}, err error) {
	psql := sq.StatementBuilder.PlaceholderFormat(sq.Question)
	query := psql.Update("acct_loan_partner_account")
	if in.PartnerId != "" {
		query = query.Set("partner_id", in.PartnerId)
	}
	if in.LoanKind != "" {
		query = query.Set("loan_kind", in.LoanKind)
	}
	if in.AccountType != "" {
		query = query.Set("account_type", in.AccountType)
	}
	if in.LoanSubCategoryCode != "" {
		query = query.Set("loan_sub_category_code", in.LoanSubCategoryCode)
	}
	query = query.Set("updated_at", atime.Now())

	query = query.Where(sq.Eq{`account_number`: in.AccountNumber})

	return query.ToSql()
}

func buildQueryGetLoanPartnerAccountByParam(opts models.GetLoanPartnerAccountByParamsIn) (sql string, args []interface{}, err error) {
	psql := sq.StatementBuilder.PlaceholderFormat(sq.Question)
	query := psql.Select([]string{
		`alpa.partner_id`,
		`alpa.loan_kind`,
		`alpa.account_number`,
		`alpa.account_type`,
		`alpa.entity_code`,
		`alpa.loan_sub_category_code`,
		`alpa.created_at`,
		`alpa.updated_at`,
	}...).From("acct_loan_partner_account alpa")
	// query = query.Join(`acct_account aa ON aa.account_number = alpa.account_number`)

	if opts.PartnerId != "" {
		query = query.Where(sq.Eq{`alpa.partner_id`: opts.PartnerId})
	}
	if opts.LoanKind != "" {
		query = query.Where(sq.Eq{`alpa.loan_kind`: opts.LoanKind})
	}
	if opts.AccountNumber != "" {
		query = query.Where(sq.Eq{`alpa.account_number`: opts.AccountNumber})
	}
	if opts.AccountType != "" {
		query = query.Where(sq.Eq{`alpa.account_type`: opts.AccountType})
	}
	if opts.EntityCode != "" {
		query = query.Where(sq.Eq{`alpa.entity_code`: opts.EntityCode})
	}
	if opts.LoanSubCategoryCode != "" {
		query = query.Where(sq.Eq{`alpa.loan_sub_category_code`: opts.LoanSubCategoryCode})
	}

	query = query.OrderBy(`alpa.created_at ASC`)

	return query.ToSql()
}
