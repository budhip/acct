package mysql

// query to acct_category table
var (
	queryCategoryIsExistByCode = `SELECT code from acct_category WHERE code = ? and status = ?;`

	queryCategoryCreate = `
		INSERT INTO acct_category(
			code, name, description, coa_type_code, status, created_at, updated_at
		)
		VALUES(
			?, ?, ?, ?, ?, now(), now()
		);
	`

	queryCategoryGetByCode = `SELECT 
		id, code, name, COALESCE(description, '') as description, COALESCE(coa_type_code, '') as coa_type_code, COALESCE(status, '') as status, created_at, updated_at
	FROM acct_category
	WHERE code = ?;`

	queryCategoryList = `SELECT id, code, name, COALESCE(description, '') as description, COALESCE(coa_type_code, '') as coa_type_code, COALESCE(status, '') as status, created_at, updated_at FROM acct_category 
	WHERE status = ? ORDER BY code ASC;`

	queryCategoryListByCoa = `SELECT  code, name FROM acct_category 
	WHERE coa_type_code = ? ORDER BY code ASC;`

	queryCeCategoryUpdate = `
		UPDATE accounting.acct_category
		SET name = COALESCE(NULLIF(?, ''), name),
		description = COALESCE(NULLIF(?, ''), description),
		coa_type_code= COALESCE(NULLIF(?, ''), coa_type_code),
		updated_at = now()
		WHERE code= ? ;
	`
)
