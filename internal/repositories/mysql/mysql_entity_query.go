package mysql

import (
	"bitbucket.org/Amartha/go-accounting/internal/models"

	sq "github.com/Masterminds/squirrel"
)

// query to entity database
var (
	queryEntityIsExistByCode = `SELECT code from acct_entity WHERE code = ?`
	queryEntityAndStatus     = ` and status = ?`
	queryEntityCreate        = `
			INSERT INTO acct_entity(code, name, description, status, created_at, updated_at)
			VALUES(?, ?, ?, ?, now(), now());
		`

	queryEntityGetByCode = `SELECT 
		id, code, name, COALESCE(description, '') as description, COALESCE(status, '') as status, created_at, updated_at
		FROM acct_entity
		WHERE code = ? and status = ?;`

	queryEntityGetByName = `SELECT 
		id, code, name, COALESCE(description, '') as description, COALESCE(status, '') as status, created_at, updated_at
		FROM acct_entity
		WHERE name = ? and status = ?;`

	queryEntityList = `SELECT 
		id, code, name, COALESCE(description, '') as description, COALESCE(status, '') as status, created_at, updated_at 
		FROM acct_entity where status = ? ORDER BY code ASC;`

	queryEntityUpdate = `
		UPDATE acct_entity
		SET
			name = COALESCE(NULLIF(?, ''), name),
			description = COALESCE(NULLIF(?, ''), description),
			status = COALESCE(NULLIF(?, ''), status),
			updated_at = now()
		WHERE code = ?;`
)

func builQueryGetEntityByParam(opts models.GetEntity) (sql string, args []interface{}, err error) {
	psql := sq.StatementBuilder.PlaceholderFormat(sq.Question)
	query := psql.Select([]string{
		`id`,
		`code`,
		`name`,
		`COALESCE(description, '') as description`,
		`COALESCE(status, '') as status`,
		`created_at`,
		`updated_at`,
	}...).From("acct_entity")

	if opts.Code != "" {
		query = query.Where(sq.Eq{`code`: opts.Code})
	}
	if opts.Name != "" {
		query = query.Where(sq.Eq{`name`: opts.Name})
	}
	if opts.Status != "" {
		query = query.Where(sq.Eq{`status`: opts.Status})
	}

	return query.ToSql()
}
