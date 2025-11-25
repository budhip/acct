package mysql

import (
	"bitbucket.org/Amartha/go-accounting/internal/models"
	sq "github.com/Masterminds/squirrel"
)

// query to acct_product_type database
var (
	queryProductTypeGetByCode = `SELECT 
		id, code, name, COALESCE(status, '') as status, COALESCE(entity_code, '') as entity_code, created_at, updated_at
		FROM acct_product_type
		WHERE code = ? and status = ?;`

	queryProductTypeList = `SELECT 
	id, code, name, COALESCE(status, '') as status, COALESCE(entity_code, '') as entity_code, created_at, updated_at
		FROM acct_product_type where status = ? ORDER BY code ASC;`

	queryProductTypeCreate = `
		INSERT INTO acct_product_type(
			code, 
			name,
			status,
			entity_code, 
			created_at, 
			updated_at
		)
		VALUES(
			?, ?, ?, ?, now(), now()
		);
	`

	queryCheckProductTypeIsExist = `SELECT id
		FROM acct_product_type
		WHERE code = ?`

	queryGetLatestProductCode = `SELECT 
		code
		FROM acct_product_type
		ORDER BY CAST(code AS UNSIGNED) DESC
		LIMIT 1;`
)

func queryProductTypeUpdate(in models.UpdateProductType) (sql string, args []interface{}, err error) {
	psql := sq.StatementBuilder.PlaceholderFormat(sq.Question)
	query := psql.Update("acct_product_type")
	if in.Name != "" {
		query = query.Set("name", in.Name)
	}
	if in.Status != "" {
		query = query.Set("status", in.Status)
	}
	query = query.Set("entity_code", in.EntityCode)

	query = query.Where(sq.Eq{`code`: in.Code})

	return query.ToSql()
}
