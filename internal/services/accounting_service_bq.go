package services

import (
	"context"
	"fmt"
	"time"

	"bitbucket.org/Amartha/go-accounting/internal/models"
	"bitbucket.org/Amartha/go-accounting/internal/pkg/atime"
	xlog "bitbucket.org/Amartha/go-x/log"
)

func (as *accounting) GenerateTrialBalanceBigQuery(ctx context.Context, date time.Time, isAdjustment bool) (err error) {
	const logPrefix = "[JOB-Generate-TrialBalance-BQ]"

	defer func() {
		logService(ctx, err)
	}()

	jobStart := atime.Now()
	periodStartEnd := atime.BeginningOfMonth(date).Format(atime.DateFormatYYYYMMDD)
	start, end := atime.PrevMonth(date)
	startDate, endDate := start.Format(atime.DateFormatYYYYMMDD), end.Format(atime.DateFormatYYYYMMDD)
	baseDate := end.Format(atime.DateFormatYYYYMMWithUnderscore)

	xlog.Info(ctx, logPrefix,
		xlog.String("description", "start processing"),
		xlog.String("period-start", startDate),
		xlog.String("period-end", endDate),
		xlog.Bool("is-adjustment", isAdjustment),
	)

	var trialBalancePeriods []models.CreateTrialBalancePeriod
	entityCodes := as.srv.conf.AccountConfig.AllowedEntitiesAccountDailyBalance

	if !isAdjustment {
		if exists, err := as.srv.bigQueryRepo.TableExists(ctx, baseDate); err != nil {
			err = models.GetErrMap(models.ErrKeyBigQueryError, err.Error())
			return err
		} else if exists {
			xlog.Info(ctx, logPrefix,
				xlog.String("description", "tb already generate"),
			)
			return nil
		}
	} else {
		tb, err := as.srv.mySqlRepo.GetTrialBalanceRepository().GetByPeriodStatus(ctx, start.Format(atime.DateFormatYYYYMM), models.TrialBalanceStatusOpen)
		if err != nil {
			return err
		}

		if len(tb) == 0 {
			return nil
		}

		adjustmentEntityCodes := make([]string, 0, len(tb))
		for _, v := range tb {
			adjustmentEntityCodes = append(adjustmentEntityCodes, v.EntityCode)
		}
		entityCodes = adjustmentEntityCodes
	}

	previousMonth := ""
	if as.srv.flagger.IsEnabled(models.FlagGetOpeningBalanceFromPreviousMonth.String()) {
		previousMonth = start.AddDate(0, -1, 0).Format(atime.DateFormatYYYYMMWithUnderscore)
	}

	steps := []struct {
		desc string
		run  func() error
	}{
		{
			"generate trial balance detail", func() error {
				return as.srv.bigQueryRepo.QueryGenerateTrialBalanceDetail(ctx, baseDate, previousMonth, periodStartEnd, startDate, endDate)
			}},
		{
			"generate trial balance summary", func() error {
				return as.srv.bigQueryRepo.QueryGenerateTrialBalanceSummary(ctx, baseDate)
			}},
		{
			"export trial balance summary to gcs", func() error {
				result, err := as.srv.bigQueryRepo.ExportTrialBalanceSummary(ctx, entityCodes, baseDate, start)
				trialBalancePeriods = append(trialBalancePeriods, result...)
				return err
			}},
		{
			"export trial balance detail to gcs", func() error {
				return as.exportTrialBalanceDetail(ctx, entityCodes, baseDate, start)
			}},
	}

	for _, step := range steps {
		stepStart := atime.Now()

		xlog.Info(ctx, logPrefix,
			xlog.String("description", "start step"),
			xlog.String("step", step.desc),
			xlog.Time("process-time", stepStart))

		if err := step.run(); err != nil {
			xlog.Error(ctx, logPrefix,
				xlog.String("description", "error"),
				xlog.String("step", step.desc),
				xlog.Err(err),
				xlog.Duration("elapsed-time", time.Since(stepStart)))

			err = models.GetErrMap(models.ErrKeyBigQueryError, err.Error())
			return err
		}

		xlog.Info(ctx, logPrefix,
			xlog.String("description", "complete"),
			xlog.String("step", step.desc),
			xlog.Duration("elapsed-time", time.Since(stepStart)))
	}

	if !isAdjustment {
		if err = as.srv.mySqlRepo.GetTrialBalanceRepository().BulkInsert(ctx, trialBalancePeriods); err != nil {
			return err
		}
	}

	xlog.Info(ctx, logPrefix,
		xlog.String("description", "job complete"),
		xlog.Duration("total-duration", time.Since(jobStart)),
	)

	return nil
}

func (as *accounting) GenerateAdjustmentTrialBalanceBigQuery(ctx context.Context, in models.AdjustmentTrialBalanceFilter) (err error) {
	defer func() {
		logService(ctx, err)
	}()

	adjustmentDate := in.AdjustmentDate
	if !in.IsManual {
		transactionsIDs, err := as.srv.mySqlRepo.GetAccountingRepository().GetTransactionsToday(ctx, adjustmentDate)
		if err != nil {
			return err
		}

		bqTransactionsIDs, err := as.srv.bigQueryRepo.QueryGetTransactions(ctx, transactionsIDs)
		if err != nil {
			err = models.GetErrMap(models.ErrKeyBigQueryError, err.Error())
			return err
		}

		if len(transactionsIDs) != len(bqTransactionsIDs) {
			// retry in gqu
			err = models.GetErrMap(models.ErrKeyTransactionIdsNotSync)
			return err
		}
	}

	// recalc tb
	go as.GenerateTrialBalanceBigQuery(context.WithoutCancel(ctx), atime.FirstOfNextMonth(adjustmentDate), true)

	return nil
}

func (as *accounting) exportTrialBalanceDetail(ctx context.Context, entityCodes []string, baseDate string, date time.Time) error {
	subCategories, err := as.srv.mySqlRepo.GetSubCategoryRepository().GetAll(ctx, models.GetAllSubCategoryParam{})
	if err != nil {
		return fmt.Errorf("get subcategories: %w", err)
	}

	return as.srv.bigQueryRepo.ExportTrialBalanceDetail(ctx, entityCodes, baseDate, date, subCategories)
}
