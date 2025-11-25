package accounting

import (
	"context"
	"os"
	"time"

	"bitbucket.org/Amartha/go-accounting/internal/pkg/atime"
	"bitbucket.org/Amartha/go-accounting/internal/services"
)

type accountingHandler struct {
	accountingService services.AccountingService
}

func Routes(acts services.AccountingService) map[string]func(ctx context.Context, date time.Time) error {
	handler := accountingHandler{
		accountingService: acts,
	}
	return map[string]func(ctx context.Context, date time.Time) error{
		"GenerateTrialBalanceBigQuery":                          handler.GenerateTrialBalanceBigQuery,
		"GenerateAccountDailyBalanceAndTrialBalance":            handler.GenerateAccountDailyBalanceAndTrialBalance,
		"GenerateRangeAccountDailyBalanceAndTrialBalance":       handler.GenerateRangeAccountDailyBalanceAndTrialBalance,
		"GenerateRangeAccountDailyBalanceAndTrialBalanceCustom": handler.GenerateRangeAccountDailyBalanceAndTrialBalanceCustom,
	}
}

// go run cmd/job/main.go run -v=v1 -n=GenerateTrialBalanceBigQuery -d=2025-02-01
func (rh *accountingHandler) GenerateTrialBalanceBigQuery(ctx context.Context, date time.Time) error {
	if err := rh.accountingService.GenerateTrialBalanceBigQuery(ctx, date, false); err != nil {
		return err
	}

	return nil
}

// go run cmd/job/main.go run -v=v1 -n=GenerateAccountDailyBalanceAndTrialBalance -d=2024-01-01
func (rh *accountingHandler) GenerateAccountDailyBalanceAndTrialBalance(ctx context.Context, date time.Time) error {
	if atime.DateEqualToday(date) {
		date = date.AddDate(0, 0, -1)
	}
	if err := rh.accountingService.GenerateAccountDailyBalance(ctx, date); err != nil {
		return err
	}

	return nil
}

// go run cmd/job/main.go run -v=v1 -n=GenerateRangeAccountDailyBalanceAndTrialBalance -d=2024-01-01
func (rh *accountingHandler) GenerateRangeAccountDailyBalanceAndTrialBalance(ctx context.Context, date time.Time) error {
	start := atime.ToZeroTime(date)
	end := atime.ToZeroTime(atime.EndOfMonth(date))

	dates, err := atime.GenerateNextDate(start, end)
	if err != nil {
		return err
	}

	for _, v := range dates {
		if err := rh.accountingService.GenerateAccountDailyBalance(ctx, v); err != nil {
			return err
		}
		time.Sleep(10 * time.Second)
	}

	return nil
}

func (rh *accountingHandler) GenerateRangeAccountDailyBalanceAndTrialBalanceCustom(ctx context.Context, date time.Time) error {
	start := atime.ToZeroTime(date)
	de, _ := atime.ParseStringToDatetime(atime.DateFormatYYYYMMDD, os.Getenv("DATE_END"))
	end := atime.ToZeroTime(atime.EndOfMonth(de))

	dates, err := atime.GenerateNextDate(start, end)
	if err != nil {
		return err
	}

	for _, v := range dates {
		if err := rh.accountingService.GenerateAccountDailyBalance(ctx, v); err != nil {
			return err
		}
		time.Sleep(10 * time.Second)
	}

	return nil
}
