package services

import (
	"context"
	"fmt"
	"time"

	"bitbucket.org/Amartha/go-accounting/internal/models"
	"bitbucket.org/Amartha/go-accounting/internal/pkg/atime"
)

type TrialBalanceService interface {
	CloseTrialBalance(ctx context.Context, in models.CloseTrialBalanceRequest) (models.TrialBalancePeriod, error)
}

type trialBalance service

var _ TrialBalanceService = (*trialBalance)(nil)

func (ts *trialBalance) CloseTrialBalance(ctx context.Context, in models.CloseTrialBalanceRequest) (out models.TrialBalancePeriod, err error) {
	defer func() {
		logService(ctx, err)
	}()

	periodDate, err := time.Parse(atime.DateFormatYYYYMM, in.Period)
	if err != nil {
		err = models.GetErrMap(models.ErrKeyInvalidFormatDate, fmt.Sprintf("date %s format must be YYYY-MM", in.Period))
		return out, err
	}

	tb, err := ts.srv.mySqlRepo.GetTrialBalanceRepository().GetByPeriod(ctx, in.Period, in.EntityCode)
	if err != nil {
		err = checkDatabaseError(err, models.ErrKeyClosedPeriodNotFound)
		return out, err
	}

	if tb.Status == models.TrialBalanceStatusClosed {
		err = models.GetErrMap(models.ErrKeyPeriodAlreadyClosed)
		return out, err
	}

	if err = ts.srv.mySqlRepo.GetTrialBalanceRepository().Close(ctx, in); err != nil {
		err = checkDatabaseError(err, models.ErrKeyClosedPeriodNotFound)
		return out, err
	}

	tb.ClosedBy = in.ClosedBy
	tb.Status = models.TrialBalanceStatusClosed
	out = *tb

	sourceTable := fmt.Sprintf("%s_tb_detail", periodDate.Format(atime.DateFormatYYYYMMWithUnderscore))
	if err = ts.srv.bigQueryRepo.QueryInsertOpeningBalanceCreated(ctx, sourceTable, in.EntityCode); err != nil {
		err = models.GetErrMap(models.ErrKeyBigQueryError, err.Error())
		return
	}

	return out, nil
}
