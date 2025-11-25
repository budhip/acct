package mysql

var (
	queryCOATypeCreate = `
		INSERT INTO acct_coa_type(
			code, coa_type_name, normal_balance, status, created_at, updated_at
		)
		VALUES(
			?, ?, ?, ?, now(), now()
		);
	`

	queryGetCOATypeByCode = `
		SELECT 
			code, coa_type_name, normal_balance, status, created_at, updated_at
		FROM acct_coa_type
		WHERE code = ?;
	`

	queryCOATypeIsExistByCode = `SELECT code from acct_coa_type WHERE code = ? and status = ?;`

	queryGetAllCOAType = `
	SELECT 
		code, coa_type_name,normal_balance, status, created_at, updated_at
	FROM acct_coa_type
	WHERE status = ?;
	`

	queryCOATypeUpdate = `
		UPDATE acct_coa_type
		SET
			coa_type_name = COALESCE(NULLIF(?, ''), coa_type_name),
			normal_balance = COALESCE(NULLIF(?, ''), normal_balance),
			status = COALESCE(NULLIF(?, ''), status),
			updated_at = now()
		WHERE code = ?;
	`
)
