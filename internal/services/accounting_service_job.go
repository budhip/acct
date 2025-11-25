package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"bitbucket.org/Amartha/go-accounting/internal/models"
	"bitbucket.org/Amartha/go-accounting/internal/pkg/atime"
	"bitbucket.org/Amartha/go-accounting/internal/pkg/dddnotification"
	"bitbucket.org/Amartha/go-accounting/internal/pkg/localstorage"
	xlog "bitbucket.org/Amartha/go-x/log"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"golang.org/x/sync/errgroup"
)

/*
go run cmd/job/main.go run -v=v1 -n=GenerateAccountTrialBalanceDaily -d=2024-01-01

- prerequisite
1. get all entity from acct_entity
2. get all sub-category from acct_sub_category
3. get all mapping category, sub-category & coa type from acct_category, acct_sub_category, acct_coa_type
3. get all account daily balance from acct_account_daily_balance where date = request date
4. loop through account_daily_balance, process calculate transactions per entity & sub-category then store into local storage

- process calculate trial balance
1. loop through entities & loop through sub-category by entity
- get account_daily_balance by entity and sub-category then calculate & process
2. bulk insert into account_trial_balance
*/
func (as *accounting) GenerateAccountTrialBalanceDaily(ctx context.Context, date time.Time) (err error) {
	process := atime.Now()
	date = atime.ToZeroTime(date)

	logMessage := "[JOB-GenerateTrialBalanceDaily]"
	message := fmt.Sprintf("Job Generate Trial Balance Daily - %s", date.Format(atime.DateFormatYYYYMMDD))
	defer func() {
		logService(ctx, err)
		elapsed := time.Since(process)
		if err != nil {
			as.sendMessageToSlack(ctx, message, err.Error())
			xlog.Error(ctx, logMessage, xlog.String("description", message), xlog.Duration("elapsed-time", elapsed), xlog.Err(err))
			return
		}
		as.sendMessageToSlack(ctx, message, fmt.Sprintf("Finished, Elapsed Time: %v", elapsed))
		xlog.Info(ctx, logMessage, xlog.String("description", message), xlog.Duration("elapsed-time", elapsed))
	}()

	xlog.Info(ctx, logMessage, xlog.Time("date", date), xlog.Time("execution-date", process))
	as.sendMessageToSlack(ctx, message, fmt.Sprintf("Starting with the execution date - %v", process))

	subCategories, err := as.srv.mySqlRepo.GetSubCategoryRepository().GetAll(ctx, models.GetAllSubCategoryParam{})
	if err != nil {
		return
	}

	_, _, mapSubCategory, err := as.srv.mySqlRepo.GetAllCategorySubCategoryCOAType(ctx)
	if err != nil {
		return
	}

	allowedEntities := as.srv.conf.AccountConfig.AllowedEntitiesAccountDailyBalance

	act := as.srv.mySqlRepo.GetAccountingRepository()
	chanAccountDailyBalance := act.GetAllAccountDailyBalance(ctx, allowedEntities, subCategories, date)

	id := uuid.New().String()
	storageAccountDailyBalance, err := localstorage.NewBadgerStorage[models.AccountTrialBalance]("account_trial_balance_" + id)
	if err != nil {
		err = models.GetErrMap(models.ErrKeyFailedToCreateStorage, err.Error())
		return
	}
	defer func() {
		storageAccountDailyBalance.Close()
		storageAccountDailyBalance.Clean()
	}()

	eg, egCtx := errgroup.WithContext(ctx)
	eg.Go(func() error {
		for v := range chanAccountDailyBalance {
			select {
			case <-egCtx.Done():
				return egCtx.Err()
			default:
				if v.Err != nil {
					err = v.Err
					return err
				}

				if v.Data.DebitMovement.IsZero() &&
					v.Data.CreditMovement.IsZero() &&
					v.Data.OpeningBalance.IsZero() &&
					v.Data.ClosingBalance.IsZero() {
					continue
				}

				key := fmt.Sprintf("%s_%s", v.Data.EntityCode, v.Data.SubCategoryCode)
				adb, err := storageAccountDailyBalance.Get(key)
				if err != nil && !errors.Is(err, localstorage.ErrKeyNotFound) {
					return err
				}

				accountTrialBalance := models.AccountTrialBalance{
					ClosingDate:     date,
					EntityCode:      v.Data.EntityCode,
					CategoryCode:    v.Data.CategoryCode,
					SubCategoryCode: v.Data.SubCategoryCode,
					DebitMovement:   v.Data.DebitMovement.Add(adb.DebitMovement),
					CreditMovement:  v.Data.CreditMovement.Add(adb.CreditMovement),
					OpeningBalance:  v.Data.OpeningBalance.Add(adb.OpeningBalance),
					ClosingBalance:  v.Data.ClosingBalance.Add(adb.ClosingBalance),
				}
				if err = storageAccountDailyBalance.Set(key, accountTrialBalance); err != nil {
					return err
				}
			}
		}
		return nil
	})
	if err = eg.Wait(); err != nil {
		return
	}
	logDuration(ctx, "ProcessGetAndSetData", process)

	trialBalances := []models.AccountTrialBalance{}
	for _, v := range allowedEntities {
		entityCode := v
		for _, v := range *subCategories {
			subCategoryCode := v.Code
			coaTypeCode := mapSubCategory[subCategoryCode].CoaTypeCode
			categoryCode := mapSubCategory[subCategoryCode].CategoryCode

			key := fmt.Sprintf("%s_%s", entityCode, subCategoryCode)
			accountsTrialBalance, err := storageAccountDailyBalance.Get(key)
			if err != nil && !errors.Is(err, localstorage.ErrKeyNotFound) {
				err = checkDatabaseError(err)
				return err
			}
			if errors.Is(err, localstorage.ErrKeyNotFound) {
				xlog.Warn(ctx, "storageAccountDailyBalance",
					xlog.Time("date", date),
					xlog.String("category-code", categoryCode),
					xlog.String("sub-category-code", subCategoryCode),
					xlog.String("coa-type-code", coaTypeCode),
					xlog.Err(err),
				)
			}

			accountsTrialBalance.ClosingDate = date
			accountsTrialBalance.EntityCode = entityCode
			accountsTrialBalance.CategoryCode = categoryCode
			accountsTrialBalance.SubCategoryCode = subCategoryCode
			accountsTrialBalance.ClosingBalance = accountsTrialBalance.OpeningBalance.Add(accountsTrialBalance.DebitMovement).Sub(accountsTrialBalance.CreditMovement)
			if coaTypeCode == models.COATypeLiability {
				accountsTrialBalance.ClosingBalance = accountsTrialBalance.OpeningBalance.Add(accountsTrialBalance.CreditMovement).Sub(accountsTrialBalance.DebitMovement)
			}
			trialBalances = append(trialBalances, accountsTrialBalance)
		}
	}

	if len(trialBalances) > 0 {
		if err = act.InsertAccountTrialBalance(ctx, trialBalances); err != nil {
			return err
		}
	}

	return
}

func (as *accounting) getOpeningBalance(ctx context.Context, accountNumber string, date time.Time) (openingBalance decimal.Decimal, err error) {
	act := as.srv.mySqlRepo.GetAccountingRepository()

	openingBalance, err = act.GetOpeningBalanceByDate(ctx, accountNumber, date)
	if err != nil && !errors.Is(err, models.ErrNoRows) {
		return
	}
	if errors.Is(err, models.ErrNoRows) {
		openingBalance, err = act.GetLastOpeningBalance(ctx, accountNumber, date)
		if err != nil && !errors.Is(err, models.ErrNoRows) {
			return
		}
	}
	return openingBalance, nil
}

func (as *accounting) sendMessageToSlack(ctx context.Context, operation, message string) {
	if err := as.srv.dddNotificationClient.SendMessageToSlack(ctx, dddnotification.MessageData{
		Operation: operation,
		Message:   message,
	}); err != nil {
		xlog.Error(ctx, "[PROCESS-JOB]", xlog.String("operation", operation), xlog.String("description", "failed to send slack message"), xlog.Err(err))
	}
}
