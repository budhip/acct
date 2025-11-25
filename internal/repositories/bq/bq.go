package bq

import (
	"context"
	"fmt"
	"time"

	"bitbucket.org/Amartha/go-accounting/internal/config"
	"bitbucket.org/Amartha/go-accounting/internal/models"
	"bitbucket.org/Amartha/go-accounting/internal/pkg/atime"
	xlog "bitbucket.org/Amartha/go-x/log"

	"cloud.google.com/go/bigquery"
	"github.com/google/uuid"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/iterator"
)

type BigQueryRepository interface {
	TableExists(ctx context.Context, baseDate string) (bool, error)
	Close() error

	// exec query to bq
	QueryGenerateTrialBalanceDetail(ctx context.Context, baseDate, previousMonth, periodStartEnd, startDate, endDate string) error
	QueryGenerateTrialBalanceSummary(ctx context.Context, baseDate string) error
	QueryGetTransactions(ctx context.Context, transactionIDs []string) ([]string, error)
	QueryInsertOpeningBalanceCreated(ctx context.Context, sourceTable string, entityCode string) error

	// exec query to bq and export to gcs
	ExportTrialBalanceDetail(ctx context.Context, entityCodes []string, baseDate string, date time.Time, subCategories *[]models.SubCategory) error
	ExportTrialBalanceSummary(ctx context.Context, entityCodes []string, baseDate string, date time.Time) ([]models.CreateTrialBalancePeriod, error)
}

type bigQueryClient struct {
	client    *bigquery.Client
	projectID string
	dataset   string
	cfg       *config.Configuration
}

func NewCloudStorageRepository(cfg *config.Configuration) (BigQueryRepository, error) {
	client, err := bigquery.NewClient(context.Background(), cfg.GcloudProjectID)
	if err != nil {
		return nil, err
	}
	return &bigQueryClient{client, cfg.GcloudProjectID, cfg.BigQueryDataset, cfg}, nil
}

func (bq *bigQueryClient) TableExists(ctx context.Context, baseDate string) (bool, error) {
	tableID := fmt.Sprintf("%s_tb_detail", baseDate)
	table := bq.client.Dataset(bq.dataset).Table(tableID)
	_, err := table.Metadata(ctx)
	if err != nil {
		// Check if it’s a 404 (table not found)
		if gerr, ok := err.(*googleapi.Error); ok && gerr.Code == 404 {
			return false, nil
		}
		return false, fmt.Errorf("failed to get table metadata: %w", err)
	}
	return true, nil
}

func (bq *bigQueryClient) Close() error {
	return bq.client.Close()
}

func (bq *bigQueryClient) QueryGenerateTrialBalanceDetail(ctx context.Context, baseDate, previousMonth, periodStartEnd, startDate, endDate string) error {
	tableID := fmt.Sprintf("%s_tb_detail", baseDate)
	openingBalance := "opening_balance_created"

	query := fmt.Sprintf(`
CREATE OR REPLACE TABLE %s AS
WITH journal_base AS (
	SELECT
		account_number,
		entity_code,
		category_code,
		sub_category_code,
		normal_balance,
		SUM(CASE WHEN is_debit = TRUE THEN amount ELSE 0 END) AS debit_movement,
		SUM(CASE WHEN is_debit = FALSE THEN amount ELSE 0 END) AS credit_movement
	FROM %s
	WHERE PARSE_DATE('%s', transaction_date) BETWEEN '%s' AND '%s'
	GROUP BY account_number, entity_code, category_code, sub_category_code, normal_balance
),
latest_name AS (
	SELECT
		account_number,
		account_name
	FROM (
		SELECT
			account_number,
			account_name,
			created_at,
			ROW_NUMBER() OVER (PARTITION BY account_number ORDER BY created_at DESC) AS rn
		FROM %s
		WHERE PARSE_DATE('%s', transaction_date) BETWEEN '%s' AND '%s'
	)
	WHERE rn = 1
),
opening_base AS (
	SELECT
		account_number,
		account_name,
		entity_code,
		category_code,
		sub_category_code,
		normal_balance,
		COALESCE(opening_balance, 0) AS opening_balance
	FROM %s
	WHERE period_start_date = '%s'
),
combined AS (
	SELECT 
		COALESCE(o.account_number, j.account_number) AS account_number,
		COALESCE(an.account_name, o.account_name) AS account_name,
		COALESCE(o.entity_code, j.entity_code) AS entity_code,
		COALESCE(o.category_code, j.category_code) AS category_code,
		COALESCE(o.sub_category_code, j.sub_category_code) AS sub_category_code,
		COALESCE(o.normal_balance, j.normal_balance) AS normal_balance,
		COALESCE(o.opening_balance, 0) AS opening_balance,
		COALESCE(j.debit_movement, 0) AS debit_movement,
		COALESCE(j.credit_movement, 0) AS credit_movement
	FROM journal_base j
	FULL OUTER JOIN opening_base o ON o.account_number = j.account_number
	LEFT JOIN latest_name an ON an.account_number = j.account_number
)
SELECT
	DATE '%s' AS period_start_end,
	account_number,
	account_name,
	entity_code,
	category_code,
	sub_category_code,
	normal_balance,
	opening_balance,
	debit_movement,
	credit_movement,
	opening_balance +
	CASE
		WHEN normal_balance = 'AST' THEN (debit_movement - credit_movement)
		WHEN normal_balance = 'LIA' THEN (credit_movement - debit_movement)
		ELSE (debit_movement - credit_movement)
	END AS closing_balance
FROM combined;
`, bq.fqTable(tableID),

		bq.fqTable("journal_entry_created"),
		"%Y-%m-%d",
		startDate, endDate,

		bq.fqTable("journal_entry_created"),
		"%Y-%m-%d",
		startDate, endDate,

		bq.fqTable(openingBalance),
		startDate, periodStartEnd,
	)

	if previousMonth != "" {
		openingBalance = fmt.Sprintf("%s_tb_detail", previousMonth)

		query = fmt.Sprintf(`
CREATE OR REPLACE TABLE %s AS
WITH journal_base AS (
	SELECT
		account_number,
		entity_code,
		category_code,
		sub_category_code,
		normal_balance,
		SUM(CASE WHEN is_debit = TRUE THEN amount ELSE 0 END) AS debit_movement,
		SUM(CASE WHEN is_debit = FALSE THEN amount ELSE 0 END) AS credit_movement
	FROM %s
	WHERE PARSE_DATE('%s', transaction_date) BETWEEN '%s' AND '%s'
	GROUP BY account_number, entity_code, category_code, sub_category_code, normal_balance
),
latest_name AS (
	SELECT
		account_number,
		account_name
	FROM (
		SELECT
			account_number,
			account_name,
			created_at,
			ROW_NUMBER() OVER (PARTITION BY account_number ORDER BY created_at DESC) AS rn
		FROM %s
		WHERE PARSE_DATE('%s', transaction_date) BETWEEN '%s' AND '%s'
	)
	WHERE rn = 1
),
opening_base AS (
	SELECT
		account_number,
		account_name,
		entity_code,
		category_code,
		sub_category_code,
		normal_balance,
		COALESCE(closing_balance, 0) AS opening_balance
	FROM %s
	WHERE period_start_end = '%s'
),
combined AS (
	SELECT 
		COALESCE(o.account_number, j.account_number) AS account_number,
		COALESCE(an.account_name, o.account_name) AS account_name,
		COALESCE(o.entity_code, j.entity_code) AS entity_code,
		COALESCE(o.category_code, j.category_code) AS category_code,
		COALESCE(o.sub_category_code, j.sub_category_code) AS sub_category_code,
		COALESCE(o.normal_balance, j.normal_balance) AS normal_balance,
		COALESCE(o.opening_balance, 0) AS opening_balance,
		COALESCE(j.debit_movement, 0) AS debit_movement,
		COALESCE(j.credit_movement, 0) AS credit_movement
	FROM journal_base j
	FULL OUTER JOIN opening_base o ON o.account_number = j.account_number
	LEFT JOIN latest_name an ON an.account_number = j.account_number
)
SELECT
	DATE '%s' AS period_start_end,
	account_number,
	account_name,
	entity_code,
	category_code,
	sub_category_code,
	normal_balance,
	opening_balance,
	debit_movement,
	credit_movement,
	opening_balance +
	CASE
		WHEN normal_balance = 'AST' THEN (debit_movement - credit_movement)
		WHEN normal_balance = 'LIA' THEN (credit_movement - debit_movement)
		ELSE (debit_movement - credit_movement)
	END AS closing_balance
FROM combined;
`, bq.fqTable(tableID),

			bq.fqTable("journal_entry_created"),
			"%Y-%m-%d",
			startDate, endDate,

			bq.fqTable("journal_entry_created"),
			"%Y-%m-%d",
			startDate, endDate,

			bq.fqTable(openingBalance),
			startDate, periodStartEnd,
		)
	}

	return bq.execQuery(ctx, query)
}

func (bq *bigQueryClient) QueryGenerateTrialBalanceSummary(ctx context.Context, baseDate string) error {
	sourceTable := fmt.Sprintf("%s_tb_detail", baseDate)
	destTable := fmt.Sprintf("%s_tb_summary", baseDate)

	query := fmt.Sprintf(`
CREATE OR REPLACE TABLE %s AS
SELECT
	entity_code, 
	category_code, 
	sub_category_code,
	SUM(debit_movement) AS debit_movement,
	SUM(credit_movement) AS credit_movement,
	SUM(opening_balance) AS opening_balance,
	SUM(closing_balance) AS closing_balance
FROM %s
GROUP BY entity_code, category_code, sub_category_code;
`, bq.fqTable(destTable), bq.fqTable(sourceTable))

	return bq.execQuery(ctx, query)
}

func (bq *bigQueryClient) QueryGetTransactions(ctx context.Context, transactionIDs []string) ([]string, error) {
	query := fmt.Sprintf(`
SELECT transaction_id
FROM %s
WHERE transaction_id IN UNNEST(@ids)
`, bq.fqTable("journal_entry_created"))

	q := bq.client.Query(query)
	q.Parameters = []bigquery.QueryParameter{
		{Name: "ids", Value: transactionIDs},
	}

	it, err := q.Read(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}

	var results []string
	for {
		var row struct {
			TransactionID string `bigquery:"transaction_id"`
		}
		if err := it.Next(&row); err == iterator.Done {
			break
		} else if err != nil {
			return nil, fmt.Errorf("iterator error: %w", err)
		}
		results = append(results, row.TransactionID)
	}

	return results, nil
}

func (bq *bigQueryClient) QueryInsertOpeningBalanceCreated(ctx context.Context, sourceTable string, entityCode string) error {
	query := fmt.Sprintf(`
INSERT INTO %s (
	account_number, 
	account_name, 
	entity_code, 
	category_code, 
	sub_category_code,
	normal_balance,
	opening_balance, 
	period_start_date, 
	created_at
)
SELECT
	account_number,
	account_name,
	entity_code,
	category_code,
	sub_category_code,
	normal_balance,
	closing_balance AS opening_balance,
	period_start_end,
	CURRENT_TIMESTAMP()
FROM %s WHERE entity_code = '%s';
`, bq.fqTable("opening_balance_created"), bq.fqTable(sourceTable), entityCode)

	return bq.execQuery(ctx, query)
}

func (bq *bigQueryClient) ExportTrialBalanceDetail(ctx context.Context, entityCodes []string, baseDate string, date time.Time, subCategories *[]models.SubCategory) error {
	sourceTable := fmt.Sprintf("%s_tb_detail", baseDate)

	for _, entity := range entityCodes {
		for _, subCategory := range *subCategories {
			code := subCategory.Code
			basePath := fmt.Sprintf(
				"gs://%s/trial_balances/%s/%s/details/%s/%s",
				bq.cfg.CloudStorageConfig.BucketName,
				date.Format(atime.DateFormatYYYY),
				entity,
				date.Format(atime.DateFormatMM),
				code,
			)

			queries := []struct {
				uri    string
				sql    string
				format bigquery.DataFormat
			}{
				{
					fmt.Sprintf("%s/total_rows.json", basePath),
					fmt.Sprintf(`SELECT COUNT(*) AS total_rows FROM %s WHERE entity_code='%s' AND sub_category_code='%s';`,
						bq.fqTable(sourceTable), entity, code),
					bigquery.JSON,
				},
				{
					fmt.Sprintf("%s/%s_*.csv", basePath, code),
					fmt.Sprintf(`SELECT account_number, account_name, entity_code, category_code, sub_category_code, opening_balance / 100.0 AS opening_balance, debit_movement / 100.0 AS debit_movement, credit_movement / 100.0 AS credit_movement, closing_balance / 100.0 AS closing_balance
					FROM %s WHERE entity_code='%s' AND sub_category_code='%s' ORDER BY account_number ASC;`,
						bq.fqTable(sourceTable), entity, code),
					bigquery.CSV,
				},
			}

			for _, q := range queries {
				if err := bq.queryToGCS(ctx, q.sql, q.uri, q.format); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (bq *bigQueryClient) ExportTrialBalanceSummary(ctx context.Context, entityCodes []string, baseDate string, date time.Time) ([]models.CreateTrialBalancePeriod, error) {
	sourceTable := fmt.Sprintf("%s_tb_summary", baseDate)
	trialBalancePeriods := make([]models.CreateTrialBalancePeriod, 0, len(entityCodes))
	for _, entity := range entityCodes {
		gcsURI := fmt.Sprintf(
			"gs://%s/trial_balances/%s/%s/summaries/%s.csv",
			bq.cfg.CloudStorageConfig.BucketName,
			date.Format(atime.DateFormatYYYY),
			entity,
			date.Format(atime.DateFormatMM),
		)

		query := fmt.Sprintf(`SELECT * FROM %s WHERE entity_code = '%s' ORDER BY sub_category_code ASC;`, bq.fqTable(sourceTable), entity)
		if err := bq.queryToGCS(ctx, query, gcsURI, bigquery.CSV); err != nil {
			return nil, err
		}
		trialBalancePeriods = append(trialBalancePeriods, models.CreateTrialBalancePeriod{
			Period:     date.Format(atime.DateFormatYYYYMM),
			EntityCode: entity,
			TBFilePath: gcsURI,
			Status:     models.TrialBalanceStatusOpen,
		})
	}
	return trialBalancePeriods, nil
}

// ------------------------------------------------------------
// Utility Functions
// ------------------------------------------------------------

func (bq *bigQueryClient) execQuery(ctx context.Context, query string, params ...bigquery.QueryParameter) error {
	var (
		job    *bigquery.Job
		status *bigquery.JobStatus
		err    error
	)

	defer func() {
		if job != nil && status != nil { // prevent nil pointer panic
			logStatus(ctx, job.ID(), status, err)
		}
	}()

	q := bq.client.Query(query)
	q.Parameters = params

	job, err = q.Run(ctx)
	if err != nil {
		err = fmt.Errorf("query.Run: %w", err)
		return err
	}

	status, err = bq.jobStatus(ctx, job)
	if err != nil {
		return err
	}

	return nil
}

func (bq *bigQueryClient) queryToGCS(ctx context.Context, query, gcsURI string, format bigquery.DataFormat) error {
	// Temporary destination table
	dataset := bq.client.DatasetInProject(bq.client.Project(), bq.dataset) // make sure dataset exists
	tmpTable := dataset.Table("tmp_export_" + uuid.New().String())

	var (
		job    *bigquery.Job
		status *bigquery.JobStatus
		err    error
	)

	defer func() {
		if job != nil && status != nil { // prevent nil pointer panic
			logStatus(ctx, job.ID(), status, err)
		}
	}()

	q := bq.client.Query(query)
	q.QueryConfig.Dst = tmpTable
	q.QueryConfig.WriteDisposition = bigquery.WriteTruncate

	job, err = q.Run(ctx)
	if err != nil {
		err = fmt.Errorf("query.Run: %w", err)
		return err
	}

	status, err = bq.jobStatus(ctx, job)
	if err != nil {
		return err
	}

	gcsRef := bigquery.NewGCSReference(gcsURI)
	gcsRef.DestinationFormat = format
	if format == bigquery.CSV {
		gcsRef.FieldDelimiter = ","
	}

	extractor := tmpTable.ExtractorTo(gcsRef)
	job, err = extractor.Run(ctx)
	if err != nil {
		err = fmt.Errorf("extractor.Run: %w", err)
		return err
	}

	status, err = bq.jobStatus(ctx, job)
	if err != nil {
		return err
	}

	if err := tmpTable.Delete(ctx); err != nil {
		err = fmt.Errorf("cleanup temp table: %w", err)
		return err
	}

	return nil
}

func (bq *bigQueryClient) jobStatus(ctx context.Context, job *bigquery.Job) (status *bigquery.JobStatus, err error) {
	status, err = job.Wait(ctx)
	if err != nil {
		err = fmt.Errorf("job.Wait: %w", err)
		return
	}

	if status.Err() != nil {
		err = fmt.Errorf("job completed with error: %w", status.Err())
		return
	}

	return
}

func (bq *bigQueryClient) fqTable(table string) string {
	return fmt.Sprintf("`%s.%s.%s`", bq.projectID, bq.dataset, table)
}

func logStatus(ctx context.Context, jobID string, status *bigquery.JobStatus, err error) {
	logBigQuery := "[BQ-Job-Status]"

	fields := []xlog.Field{
		xlog.String("JobID", jobID),
		xlog.Int64("TotalBytesProcessed", status.Statistics.TotalBytesProcessed),
		xlog.Duration("ExecutionTime", status.Statistics.EndTime.Sub(status.Statistics.StartTime)),
	}

	if qStats, ok := status.Statistics.Details.(*bigquery.QueryStatistics); ok {
		fields = append(fields,
			xlog.String("StatementType", qStats.StatementType),
			xlog.Int64("TotalSlotMs", qStats.SlotMillis),
			xlog.Int64("BytesBilled", qStats.BillingTier),
			xlog.Bool("CacheHit", qStats.CacheHit),
			xlog.Int64("NumDMLAffectedRows", qStats.NumDMLAffectedRows),
		)

		// DML row output
		if qStats.DMLStats != nil {
			rows := qStats.DMLStats.InsertedRowCount +
				qStats.DMLStats.UpdatedRowCount +
				qStats.DMLStats.DeletedRowCount
			fields = append(fields, xlog.Int64("OutputRows", rows))
		}
	}

	if err != nil {
		xlog.Error(ctx, logBigQuery, fields...)
		return
	}

	xlog.Info(ctx, logBigQuery, fields...)
}
