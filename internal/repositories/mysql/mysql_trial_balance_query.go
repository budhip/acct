package mysql

// query to acct_trial_balance_periods table
var (
	queryBulkInsertTrialBalancePeriod = `
		INSERT INTO acct_trial_balance_periods(
			period,
			entity_code,
			tb_file_path,
			status,
			is_adjustment
		) VALUES %s`

	queryTrialBalanceClose = `
		UPDATE 
			acct_trial_balance_periods
		SET 
			status = ?,
			closed_by = ?,
			updated_at = CURRENT_TIMESTAMP(6)
		WHERE 
			period = ? AND entity_code = ?`

	queryUpdateTrialBalanceAdjustment = `
		UPDATE 
			acct_trial_balance_periods
		SET 
			is_adjustment = TRUE,
			updated_at = CURRENT_TIMESTAMP(6)
		WHERE 
			period = ?`

	queryTrialBalanceGetByPeriod = `
		SELECT 
			id,
			period,
			COALESCE(tb_file_path, '') as tb_file_path,
			status,
			COALESCE(closed_by, '') as closed_by,
			is_adjustment,
			created_at,
			updated_at
		FROM 
			acct_trial_balance_periods
		WHERE 
			period = ? AND entity_code = ?`

	queryGetFirstPeriodByStatus = `
		SELECT 
			id,
			period,
			COALESCE(tb_file_path, '') as tb_file_path,
			status,
			COALESCE(closed_by, '') as closed_by,
			is_adjustment,
			created_at,
			updated_at
		FROM 
			acct_trial_balance_periods
		WHERE 
			status = ?
			ORDER BY period ASC
			LIMIT 1`

	queryGetByPeriodStatus = `
		SELECT 
			period,
			entity_code,
			COALESCE(tb_file_path, '') as tb_file_path,
			status,
			COALESCE(closed_by, '') as closed_by,
			is_adjustment,
			created_at,
			updated_at
		FROM 
			acct_trial_balance_periods
		WHERE 
			period = ? AND status = ?
			`
)
