package mysql

import (
	"fmt"
	"strings"

	"bitbucket.org/Amartha/go-accounting/internal/models"

	sq "github.com/Masterminds/squirrel"
)

const (
	orderAsc  = "asc"
	orderDesc = "desc"
)

// query to acct_account database
var (
	queryAccountCreate = `
	INSERT INTO acct_account(
		account_number,
		owner_id,
		account_type,
		product_type_code,
		entity_code,
		category_code,
		sub_category_code,
		currency,
		status,
		name,
		alt_id,
		legacy_id,
		metadata,
		created_at,
		updated_at
	)
	VALUES(
		NULLIF(?, ''),
		NULLIF(?, ''),
		NULLIF(?, ''),
		NULLIF(?, ''),
		NULLIF(?, ''),
		NULLIF(?, ''),
		NULLIF(?, ''),
		NULLIF(?, ''),
		NULLIF(?, ''),
		NULLIF(?, ''),
		NULLIF(?, ''),
		NULLIF(?, '{}'),
		NULLIF(?, '{}'),
		CURRENT_TIMESTAMP(6), CURRENT_TIMESTAMP(6)
	)`

	queryUpdateAccountEntity = `
	UPDATE acct_account
	SET
		entity_code = ?,
		updated_at = CURRENT_TIMESTAMP(6)
	WHERE account_number = ?;`

	queryAccountUpdateLegacyId = `
	UPDATE acct_account
	SET
		legacy_id = COALESCE(NULLIF(?, ''),legacy_id),
		updated_at = CURRENT_TIMESTAMP(6)
	WHERE account_number = ?;`

	queryAccountUpdateAltId = `
	UPDATE acct_account
	SET
		alt_id = ?,
		updated_at = CURRENT_TIMESTAMP(6)
	WHERE account_number = ?;`

	queryAccountUpdateBySubCategory = `
	UPDATE acct_account
	SET
`
	queryAccountUpdateBySubCategoryWhere = `
		updated_at = CURRENT_TIMESTAMP(6)
	WHERE sub_category_code = ?;`

	queryAccountByNumber = `
	SELECT aa.account_number,
		   coalesce(aa.name,"") account_name,
		   coalesce(aa.owner_id,"") owner_id,
		   coalesce(aa.category_code,"") category_code,
		   ac.name category_name,
		   coalesce(act.code,"") coa_type_code,
		   coalesce(act.coa_type_name,"") coa_type_name,
		   coalesce(aa.sub_category_code,"") sub_category_code,
		   coalesce(asuc.name,"") sub_category_name,
		   coalesce(aa.entity_code,"") entity_code,
		   coalesce(ae.name,"") entity_name,
		   coalesce(aa.product_type_code, "") product_type_code,
		   coalesce(apt.name, "") product_type_name,
		   coalesce(aa.currency,"") currency,
		   coalesce(aa.status,"") status,
		   coalesce(aa.alt_id, "") alt_id,
		   aa.created_at,
		   aa.updated_at,
		   coalesce(aa.legacy_id, "{}") legacy_id,
		   coalesce(aa.metadata, "{}") metadata,
		   coalesce(asuc.account_type, "") account_type
	FROM acct_account aa
		LEFT JOIN acct_category ac ON aa.category_code = ac.code
		LEFT JOIN acct_coa_type act ON ac.coa_type_code = act.code
		LEFT JOIN acct_sub_category asuc ON aa.sub_category_code = asuc.code
		LEFT JOIN acct_entity ae ON aa.entity_code = ae.code
		LEFT JOIN acct_product_type apt on apt.code = aa.product_type_code
	WHERE account_number = ? or alt_id = ?;`

	queryAccountByLegacyID = `
	SELECT aa.account_number,
		   coalesce(aa.name,"") account_name,
		   coalesce(aa.owner_id,"") owner_id,
		   coalesce(aa.category_code,"") category_code,
		   ac.name category_name,
		   coalesce(act.code,"") coa_type_code,
		   coalesce(act.coa_type_name,"") coa_type_name,
		   coalesce(aa.sub_category_code,"") sub_category_code,
		   coalesce(asuc.name,"") sub_category_name,
		   coalesce(aa.entity_code,"") entity_code,
		   coalesce(ae.name,"") entity_name,
		   coalesce(aa.product_type_code, "") product_type_code,
		   coalesce(apt.name, "") product_type_name,
		   coalesce(aa.currency,"") currency,
		   coalesce(aa.status,"") status,
		   coalesce(aa.alt_id, "") alt_id,
		   aa.created_at,
		   aa.updated_at,
		   coalesce(aa.legacy_id, "{}") legacy_id,
		   coalesce(aa.metadata, "{}") metadata,
		   coalesce(asuc.account_type, "") account_type
	FROM acct_account aa
		LEFT JOIN acct_category ac ON aa.category_code = ac.code
		LEFT JOIN acct_coa_type act ON ac.coa_type_code = act.code
		LEFT JOIN acct_sub_category asuc ON aa.sub_category_code = asuc.code
		LEFT JOIN acct_entity ae ON aa.entity_code = ae.code
		LEFT JOIN acct_product_type apt on apt.code = aa.product_type_code
	WHERE (cast(legacy_id->>"$.t24AccountNumber" as char(255))) = ?;`

	queryBulkInsertAccount     = `INSERT INTO accounts(account_id, name) VALUES %s`
	queryBulkInsertAcctAccount = `INSERT INTO acct_account(
		account_number, owner_id, account_type, product_type_code, entity_code, category_code, sub_category_code, currency, status, name, alt_id, legacy_id, metadata, created_at, updated_at) 
		VALUES %s`

	queryCreateLenderAccount = `
		INSERT INTO acct_lender_account(
			cih_account_number, invested_account_number, receivables_account_number
		)
		VALUES(
			?, NULLIF(?, ''), NULLIF(?, '')
		);`

	queryGetAllAccountNumbersByParam = `
	SELECT aa.account_number, coalesce(asc2.account_type, "") account_type, aa.sub_category_code, aa.created_at
	FROM acct_account aa
	JOIN acct_sub_category asc2 ON asc2.code = aa.sub_category_code
	WHERE aa.owner_id = ?`

	queryGetLenderAccountByCIHAccountNumber = `
	SELECT acl.cih_account_number, 
		coalesce(acl.invested_account_number, ""),
		coalesce(acl.receivables_account_number, "")
	FROM acct_lender_account acl
	WHERE acl.cih_account_number = ?;`

	queryCreateLoanAccount = `
		INSERT INTO acct_loan_account(
			loan_account_number, loan_advance_payment_account_number
		)
		VALUES(
			?, NULLIF(?, '')
		);`

	queryCheckLegacyId = `
	SELECT legacy_id->>"$.t24AccountNumber" as t24_account_number
	FROM acct_account
	WHERE (cast(legacy_id->>"$.t24AccountNumber" as char(255))) = ?;`

	queryCheckAccountNumber = `
	SELECT
		aa.account_number,
		aa.entity_code
	FROM acct_account aa
	JOIN accounts a ON a.account_id = aa.account_number
		WHERE aa.account_number = ?;`

	queryGetAccountNumberByLegacyId = `
	SELECT account_number 
	FROM acct_account 
	WHERE (cast(legacy_id->>"$.t24AccountNumber" as char(255))) = ?;`

	queryGetLoanAdvanceAccountByLoanAccount = `
	SELECT ala.loan_account_number, 
		coalesce(ala.loan_advance_payment_account_number, "")
	FROM acct_loan_account ala
	WHERE ala.loan_account_number = ?;`

	queryGetAllAccountNumber = `
	SELECT 
		aa.account_number,
		aa.entity_code,
		aa.category_code,
		aa.sub_category_code
		FROM acct_account aa 
		WHERE aa.sub_category_code = ? AND aa.entity_code IN (%s)`
)

func buildUpdateAccountQuery(input models.UpdateAccount) (sql string, args []interface{}, err error) {
	psql := sq.StatementBuilder.PlaceholderFormat(sq.Question)

	query := psql.Update("acct_account").
		Set("owner_id", input.OwnerID).
		Set("updated_at", sq.Expr("CURRENT_TIMESTAMP(6)")).
		Where(sq.Eq{"account_number": input.AccountNumber})

	if input.Name != "" {
		query = query.Set("name", input.Name)
	}

	if input.AltID != "" {
		query = query.Set("alt_id", input.AltID)
	}

	if input.LegacyId != nil {
		query = query.Set("legacy_id", input.LegacyId)
	}

	return query.ToSql()
}

func buildAccountListQuery(opts models.AccountFilterOptions) (sql string, args []interface{}, err error) {
	columns := []string{
		`aa.account_number`,
		`coalesce(aa.name, '') account_name`,
		`coalesce(aa.category_code, '') category_code`,
		`coalesce(aa.sub_category_code, '') sub_category_code`,
		`coalesce(aa.entity_code, '') entity_code`,
		`coalesce(ae.name, '') entity_name`,
		`coalesce(aa.product_type_code, '') product_type_code`,
		`coalesce(apt.name, '') product_type_name`,
		`coalesce(aa.alt_id,  '') alt_id`,
		`coalesce(aa.owner_id, '') owner_id`,
		`coalesce(aa.status, '') status`,
		`coalesce(aa.legacy_id, "{}") legacy_id`,
		`aa.created_at`,
		`aa.updated_at`,
		`coalesce(aa.t24_account_number, '') t24_account_number`,
	}
	query := buildFilteredAccountListQuery(columns, opts)

	return query.ToSql()
}

func buildFilteredAccountListQuery(cols []string, opts models.AccountFilterOptions) sq.SelectBuilder {
	subQuery := sq.StatementBuilder.PlaceholderFormat(sq.Question).Select([]string{
		`aa.account_number`,
		`coalesce(aa.name,'') name`,
		`aa.category_code`,
		`aa.sub_category_code`,
		`aa.entity_code`,
		`aa.product_type_code`,
		`aa.alt_id`,
		`aa.owner_id`,
		`aa.status`,
		`aa.legacy_id`,
		`aa.created_at`,
		`aa.updated_at`,
		`cast(legacy_id ->> "$.t24AccountNumber" as char(255)) t24_account_number`,
	}...).From("acct_account aa")

	if opts.EntityCode != "" {
		subQuery = subQuery.Where(sq.Eq{`aa.entity_code`: opts.EntityCode})
	}

	if opts.CoaTypeCode != "" {
		subQuery = subQuery.Join("acct_category ac ON ac.code = aa.category_code")
		subQuery = subQuery.Where(sq.Eq{`ac.coa_type_code`: opts.CoaTypeCode})
	}
	if opts.Search != "" {
		if opts.SearchBy == "t24AccountNumber" {
			subQuery = subQuery.Where(sq.Eq{`cast(legacy_id ->> "$.t24AccountNumber" as char(255))`: opts.Search})
		} else {
			subQuery = subQuery.Where(sq.Eq{fmt.Sprintf(`aa.%s`, opts.SearchBy): opts.Search})
		}
	}
	if opts.CategoryCode != "" {
		subQuery = subQuery.Where(sq.Eq{`aa.category_code`: opts.CategoryCode})
	}
	if opts.SubCategoryCode != "" {
		subQuery = subQuery.Where(sq.Eq{`aa.sub_category_code`: opts.SubCategoryCode})
	}
	if opts.ProductTypeCode != "" {
		subQuery = subQuery.Where(sq.Eq{`aa.product_type_code`: opts.ProductTypeCode})
	}

	if opts.AfterCreatedAt != nil {
		subQuery = subQuery.Where(sq.Lt{`aa.created_at`: opts.AfterCreatedAt})
	}
	if opts.BeforeCreatedAt != nil {
		subQuery = subQuery.Where(sq.Gt{`aa.created_at`: opts.BeforeCreatedAt})
	}

	if opts.GuestMode {
		subQuery = subQuery.Where(sq.Eq{"aa.sub_category_code": models.ExcludedSubCategoryForGuestMode})
	}

	if opts.AscendingOrder {
		subQuery = subQuery.OrderBy(`aa.created_at ASC`)
	} else {
		subQuery = subQuery.OrderBy(`aa.created_at DESC`)
	}

	if opts.Limit > 0 {
		subQuery = subQuery.Limit(uint64(opts.Limit))
	}

	psql := sq.StatementBuilder.PlaceholderFormat(sq.Question)
	query := psql.Select(cols...).FromSelect(subQuery, "aa")
	query = query.LeftJoin("acct_entity ae ON ae.code = aa.entity_code")
	query = query.LeftJoin("acct_product_type apt on apt.code = aa.product_type_code")

	return query
}

func buildCountAccountListQuery(opts models.AccountFilterOptions) (sql string, args []interface{}, err error) {
	query := countAccountListQuery([]string{`count(1)`}, opts)
	return query.ToSql()
}

func countAccountListQuery(cols []string, opts models.AccountFilterOptions) sq.SelectBuilder {
	psql := sq.StatementBuilder.PlaceholderFormat(sq.Question)
	query := psql.Select(cols...).From("acct_account aa")

	if opts.CoaTypeCode != "" {
		query = query.Join("acct_category ac ON ac.code = aa.category_code")
		query = query.Where(sq.Eq{`ac.coa_type_code`: opts.CoaTypeCode})
	}
	if opts.Search != "" {
		if opts.SearchBy == "t24AccountNumber" {
			query = query.Where(sq.Eq{`cast(legacy_id ->> "$.t24AccountNumber" as char(255))`: opts.Search})
		} else {
			query = query.Where(sq.Eq{fmt.Sprintf(`aa.%s`, opts.SearchBy): opts.Search})
		}
	}
	if opts.EntityCode != "" {
		query = query.Where(sq.Eq{`aa.entity_code`: opts.EntityCode})
	}
	if opts.CategoryCode != "" {
		query = query.Where(sq.Eq{`aa.category_code`: opts.CategoryCode})
	}
	if opts.SubCategoryCode != "" {
		query = query.Where(sq.Eq{`aa.sub_category_code`: opts.SubCategoryCode})
	}
	if opts.ProductTypeCode != "" {
		query = query.Where(sq.Eq{`aa.product_type_code`: opts.ProductTypeCode})
	}

	return query
}

func buildQueryCheckExistByParam(opts models.AccountFilterOptions) (sql string, args []interface{}, err error) {
	psql := sq.StatementBuilder.PlaceholderFormat(sq.Question)
	query := psql.Select([]string{`id`}...).From("acct_account")

	if opts.AltID != "" {
		query = query.Where(sq.Eq{`alt_id`: opts.AltID})
	}
	if opts.SubCategoryCode != "" {
		query = query.Where(sq.Eq{`sub_category_code`: opts.SubCategoryCode})
	}
	if opts.EntityCode != "" {
		query = query.Where(sq.Eq{`entity_code`: opts.EntityCode})
	}

	query.Limit(1)
	return query.ToSql()
}

func buildQueryGetAllAccountNumbersByParam(opts models.GetAllAccountNumbersByParamIn) (sql string, args []interface{}, err error) {
	psql := sq.StatementBuilder.PlaceholderFormat(sq.Question)
	query := psql.Select([]string{
		`coalesce(aa.owner_id, "") owner_id`,
		`aa.account_number`,
		`coalesce(aa.alt_id, "") alt_id`,
		`coalesce(aa.name,"") account_name`,
		`coalesce(asc2.account_type,'') account_type`,
		`coalesce(aa.entity_code,"") entity_code`,
		`coalesce(aa.product_type_code, "") product_type_code`,
		`coalesce(aa.category_code,"") category_code`,
		`coalesce(aa.sub_category_code,"") sub_category_code`,
		`coalesce(aa.currency,"") currency`,
		`coalesce(aa.status,"") status`,
		`coalesce(aa.legacy_id, "{}") legacy_id`,
		`coalesce(aa.metadata, "{}") metadata`,
		`aa.created_at`,
	}...).From("acct_account aa")
	query = query.Join(`acct_sub_category asc2 ON asc2.code = aa.sub_category_code`)

	if opts.OwnerId != "" {
		query = query.Where(sq.Eq{`aa.owner_id`: opts.OwnerId})
	}
	if opts.AltId != "" {
		query = query.Where(sq.Eq{`aa.alt_id`: opts.AltId})
	}
	if opts.SubCategoryCode != "" {
		query = query.Where(sq.Eq{`aa.sub_category_code`: opts.SubCategoryCode})
	}
	if opts.AccountType != "" {
		query = query.Where(sq.Eq{`asc2.account_type`: opts.AccountType})
	}
	if opts.AccountNumbers != "" {
		accountNumbers := strings.Split(opts.AccountNumbers, ",")
		placeholders := toPlaceholders(accountNumbers)
		args := toInterface(accountNumbers)
		query = query.Where(sq.Expr(fmt.Sprintf("aa.account_number IN (%s)", strings.Join(placeholders, ",")), args...))
	}
	if opts.Limit > 0 {
		query = query.Limit(uint64(opts.Limit))
	}
	query = query.OrderBy(`aa.created_at ASC`)

	return query.ToSql()
}
