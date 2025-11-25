package mysql

// query to sub_category database
var (
	querySubCategoryIsExistByCode = `SELECT 
	id, code, name, COALESCE(description, '') as description, category_code, COALESCE(account_type, '') as account_type, COALESCE(default_product_type_code, '') as product_type_code, COALESCE(default_currency, '') as currency, status, created_at, updated_at
	from acct_sub_category WHERE code = ? AND category_code = ? and status = ?;`

	querySubCategoryGetByCode = `SELECT 
		id, code, name, COALESCE(description, '') as description, category_code, COALESCE(account_type, '') as account_type, COALESCE(default_product_type_code, '') as product_type_code, COALESCE(default_currency, '') as currency, COALESCE(status, '') as status, created_at, updated_at
		FROM acct_sub_category
		WHERE code = ?;`

	querySubCategoryGetByAccountType = `SELECT 
		id, 
		code, 
		name, 
		COALESCE(description, '') as description, 
		COALESCE(account_type, '') as account_type, 
		COALESCE(default_product_type_code, '') as product_type_code,
		category_code,
		COALESCE(default_currency, '') as currency,
		COALESCE(status, '') as status,
		created_at, 
		updated_at
		FROM acct_sub_category
		WHERE account_type = ? and status = ? ORDER BY code ASC;`

	querySubCategoryCreate = `
		INSERT INTO acct_sub_category(
			code, 
			name, 
			description, 
			category_code, 
			account_type,
			default_product_type_code, 
			default_currency, 
			status, 
			created_at, 
			updated_at
		)
		VALUES(
			?, ?, ?, ?, ?, NULLIF(?, ''), NULLIF(?, ''), ?, now(), now()
		);
	`

	queryGetAllSubCategory = `SELECT 
		acs.id, acs.category_code, acs.code, acs.name, COALESCE(acs.description, '') as description, COALESCE(acs.account_type, '') as account_type, 
		COALESCE(acs.default_product_type_code, '') as product_type_code, 
		coalesce(apt.name, '') as product_type_name,
		COALESCE(acs.default_currency, '') as currency, acs.status, 
		acs.created_at, acs.updated_at
		FROM acct_sub_category acs 
		LEFT JOIN acct_product_type apt
		ON acs.default_product_type_code = apt.code
		WHERE acs.status = ? 
		`

	querySubCategoryUpdate = `
		UPDATE acct_sub_category
		SET
			name = COALESCE(NULLIF(?, ''), name),
			description = COALESCE(NULLIF(?, ''), description),
			status = COALESCE(NULLIF(?, ''), status),
	`

	querySubCategoryWhere = `
			updated_at = now()
		WHERE code = ?;
	`
	queryGetLatestSubCategoryCode = `
	SELECT code
	FROM acct_sub_category
	WHERE category_code = ?
	ORDER BY CAST(code AS UNSIGNED) DESC
	LIMIT 1;`
)
