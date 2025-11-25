package services

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"time"

	"bitbucket.org/Amartha/go-accounting/internal/models"
	"bitbucket.org/Amartha/go-accounting/internal/pkg/atime"
	"bitbucket.org/Amartha/go-accounting/internal/pkg/flag"
	"bitbucket.org/Amartha/go-accounting/internal/pkg/localstorage"
	xlog "bitbucket.org/Amartha/go-x/log"

	"github.com/google/uuid"
	"golang.org/x/sync/errgroup"
)

/*
go run cmd/job/main.go run -v=v1 -n=GenerateAccountDailyBalanceAndTrialBalance -d=2024-01-01

- prerequisite
1. get all mapping category, sub-category & coa type from acct_category, acct_sub_category, acct_coa_type
2. get all account from acct_account
3. get all account daily balance from acct_account_daily_balance where date = request date - 1
4. get all account transaction from transactions & splits where transaction date = request date
5. loop through accounts then store into local storage
6. loop through account_daily_balance then store into local storage
7. loop through account_journal_transaction, process calculate transactions per account then store into local storage

- process calculate account daily balance
1. loop through acccounts then get by account number
- if account number exist in account_daily_balance & account_transaction then calculate & process
- if account number not exist in account_daily_balance but exist in account_transaction then calculate & process
- if account number exist in account_daily_balance but not exist in account_transaction then process
- if account number not exist in account_daily_balance & account_transaction then process
2. bulk insert into account_daily_balance
*/
func (as *accounting) GenerateAccountDailyBalance(ctx context.Context, date time.Time) (err error) {
	process := atime.Now()
	date = atime.ToZeroTime(date)
	allowedEntities := as.srv.conf.AccountConfig.AllowedEntitiesAccountDailyBalance
	chunkSize := as.srv.conf.AccountConfig.ChunkSizeAccountBalance
	flagChunkSize := models.FlagChunkSizeAccountBalance.String()
	if as.srv.flagger.IsEnabled(flagChunkSize) {
		variant, err := flag.GetVariant[models.ChunkSizeVariant](as.srv.flagger, flagChunkSize)
		if err != nil {
			return err
		}
		chunkSize = variant.Value.ChunkSize
	}

	logMessage := "[JOB-GenerateAccountDailyBalance]"
	message := fmt.Sprintf("Job Generate Account Daily Balance & Trial Balance - %s", date.Format(atime.DateFormatYYYYMMDD))
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
	adbMessage := fmt.Sprintf("Process Generate Account Daily Balance - %s", date.Format(atime.DateFormatYYYYMMDD))
	as.sendMessageToSlack(ctx, adbMessage, fmt.Sprintf("Starting with the execution date - %v", process))

	_, _, mapSubCategory, err := as.srv.mySqlRepo.GetAllCategorySubCategoryCOAType(ctx)
	if err != nil {
		return
	}

	subCategories, err := as.srv.mySqlRepo.GetSubCategoryRepository().GetAll(ctx, models.GetAllSubCategoryParam{})
	if err != nil {
		return
	}

	storageAccountTransaction,
		storageAccountDailyBalance,
		storageAccountTrialBalance,
		storageAccounts,
		err := as.processGetAndSetData(ctx, date, subCategories)
	if err != nil {
		return
	}

	defer func() {
		errCloseAT := storageAccountTransaction.Close()
		errCleanAT := storageAccountTransaction.Clean()

		errCloseADB := storageAccountDailyBalance.Close()
		errCleanADB := storageAccountDailyBalance.Clean()

		errCloseATB := storageAccountTrialBalance.Close()
		errCleanATB := storageAccountTrialBalance.Clean()

		errCloseA := storageAccounts.Close()
		errCleanA := storageAccounts.Clean()

		errCloseClean := errors.Join(errCloseAT, errCleanAT, errCloseADB, errCleanADB, errCloseATB, errCleanATB, errCloseA, errCleanA)
		if errCloseClean != nil {
			xlog.Error(ctx, logMessage, xlog.String("description", "error when close and clean local storage"), xlog.Err(errCloseClean))
		}
	}()

	var (
		onceCalculate sync.Once
		onceInsert    sync.Once
		insertGroup   errgroup.Group
		// sem            = make(chan struct{}, as.srv.conf.AccountConfig.MaxConcurrentAccountBalance)
		accounts       []models.AccountBalanceDaily
		errKeyNotFound = localstorage.ErrKeyNotFound
		act            = as.srv.mySqlRepo.GetAccountingRepository()
	)

	if err = storageAccounts.ForEach(func(accountNumber string, value models.GetAccountOut) error {
		onceCalculate.Do(func() {
			xlog.Info(ctx, logMessage, xlog.String("description", "start processing calculate and insert data"))
		})

		adb, errAdb := storageAccountDailyBalance.Get(accountNumber)
		ajt, errAjt := storageAccountTransaction.Get(accountNumber)

		accountBalanceDaily := models.AccountBalanceDaily{
			BalanceDate:     date,
			AccountNumber:   value.AccountNumber,
			EntityCode:      value.EntityCode,
			CategoryCode:    value.CategoryCode,
			SubCategoryCode: value.SubCategoryCode,
		}
		switch {
		case errors.Is(errAdb, errKeyNotFound) &&
			errors.Is(errAjt, errKeyNotFound):
			// not exist in previous account daily balance and has no transactions
			{
				// do nothing
				return nil
			}
		case errors.Is(errAjt, errKeyNotFound) && errAdb == nil:
			// has no transactions but exist in previous account daily balance
			{
				accountBalanceDaily.OpeningBalance = adb.ClosingBalance
				accountBalanceDaily.ClosingBalance = adb.ClosingBalance
				// return nil
			}
		case errors.Is(errAdb, errKeyNotFound) && errAjt == nil:
			// not exist in previous account daily balance but has a transactions
			{
				accountBalanceDaily.DebitMovement = ajt.DebitMovement
				accountBalanceDaily.CreditMovement = ajt.CreditMovement
				accountBalanceDaily = as.calculateClosingBalance(mapSubCategory, accountBalanceDaily)
			}
		case errAdb == nil && errAjt == nil:
			// both previous daily balance and transactions is exist
			{
				accountBalanceDaily.DebitMovement = ajt.DebitMovement
				accountBalanceDaily.CreditMovement = ajt.CreditMovement
				accountBalanceDaily.OpeningBalance = adb.ClosingBalance
				accountBalanceDaily = as.calculateClosingBalance(mapSubCategory, accountBalanceDaily)
			}
		default:
			// handle unexpected error cases
			err = errors.Join(errAdb, errAjt)
			return err
		}

		key := fmt.Sprintf("%s_%s", accountBalanceDaily.EntityCode, accountBalanceDaily.SubCategoryCode)
		atb, err := storageAccountTrialBalance.Get(key)
		if err != nil && !errors.Is(err, errKeyNotFound) {
			return err
		}
		accountTrialBalance := models.AccountTrialBalance{
			ClosingDate:     date,
			EntityCode:      accountBalanceDaily.EntityCode,
			CategoryCode:    accountBalanceDaily.CategoryCode,
			SubCategoryCode: accountBalanceDaily.SubCategoryCode,
			DebitMovement:   accountBalanceDaily.DebitMovement.Add(atb.DebitMovement),
			CreditMovement:  accountBalanceDaily.CreditMovement.Add(atb.CreditMovement),
			OpeningBalance:  accountBalanceDaily.OpeningBalance.Add(atb.OpeningBalance),
			ClosingBalance:  accountBalanceDaily.ClosingBalance.Add(atb.ClosingBalance),
		}
		if err = storageAccountTrialBalance.Set(key, accountTrialBalance); err != nil {
			return err
		}

		accounts = append(accounts, accountBalanceDaily)
		if len(accounts) == chunkSize {
			onceInsert.Do(func() {
				xlog.Info(ctx, logMessage, xlog.String("description", "first batch insert into acct_account_daily_balance"), xlog.Int("batch-size", len(accounts)))
			})

			// sort.Slice(accounts, func(i, j int) bool {
			// 	if accounts[i].AccountNumber == accounts[j].AccountNumber {
			// 		return accounts[i].BalanceDate.Before(accounts[j].BalanceDate)
			// 	}
			// 	return accounts[i].AccountNumber < accounts[j].AccountNumber
			// })
			insertBatch := make([]models.AccountBalanceDaily, len(accounts))
			copy(insertBatch, accounts)
			accounts = accounts[:0]

			if as.srv.conf.AccountConfig.IsSequentialAccountDailyBalance {
				if err := act.InsertAccountBalanceDaily(ctx, insertBatch); err != nil {
					return err
				}
			} else {
				// sem <- struct{}{}
				insertGroup.Go(func() error {
					// defer func() { <-sem }()
					return as.retryInsert(ctx, insertBatch)
				})
			}
		}

		return nil
	}); err != nil {
		return
	}

	if len(accounts) > 0 {
		insertBatch := make([]models.AccountBalanceDaily, len(accounts))
		copy(insertBatch, accounts)

		if as.srv.conf.AccountConfig.IsSequentialAccountDailyBalance {
			if err := act.InsertAccountBalanceDaily(ctx, insertBatch); err != nil {
				return err
			}
		} else {
			// sem <- struct{}{}
			insertGroup.Go(func() error {
				// defer func() { <-sem }()
				return as.retryInsert(ctx, insertBatch)
			})
		}
	}

	if !as.srv.conf.AccountConfig.IsSequentialAccountDailyBalance {
		if err := insertGroup.Wait(); err != nil {
			return err
		}
	}

	tbMessage := fmt.Sprintf("Process Generate Trial Balance Daily - %s", date.Format(atime.DateFormatYYYYMMDD))
	as.sendMessageToSlack(ctx, tbMessage, fmt.Sprintf("Starting with the execution date - %v", atime.Now()))

	trialBalances := []models.AccountTrialBalance{}
	for _, v := range allowedEntities {
		entityCode := v
		for _, v := range *subCategories {
			subCategoryCode := v.Code
			coaTypeCode := mapSubCategory[subCategoryCode].CoaTypeCode
			categoryCode := mapSubCategory[subCategoryCode].CategoryCode

			key := fmt.Sprintf("%s_%s", entityCode, subCategoryCode)
			accountsTrialBalance, err := storageAccountTrialBalance.Get(key)
			if err != nil && !errors.Is(err, errKeyNotFound) {
				return err
			}
			// if errors.Is(err, errKeyNotFound) {
			// 	xlog.Warn(ctx, "storageAccountTrialBalance",
			// 		xlog.Time("date", date),
			// 		xlog.String("category-code", categoryCode),
			// 		xlog.String("sub-category-code", subCategoryCode),
			// 		xlog.String("coa-type-code", coaTypeCode),
			// 		xlog.Err(err),
			// 	)
			// }

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

func (as *accounting) retryInsert(ctx context.Context, batch []models.AccountBalanceDaily) error {
	const maxRetry = 3
	base := 100 * time.Millisecond

	for i := 0; i < maxRetry; i++ {
		err := as.srv.mySqlRepo.GetAccountingRepository().InsertAccountBalanceDaily(ctx, batch)
		if err == nil {
			return nil
		}

		// Check if it's a deadlock (Error 1213)
		if strings.Contains(err.Error(), "Error 1213") {
			jitter := time.Duration(rand.Intn(100)) * time.Millisecond
			sleep := base*time.Duration(1<<i) + jitter
			xlog.Warn(ctx, "[Retry Insert Deadlock]",
				xlog.Int("attempt", i+1),
				xlog.Duration("sleep", sleep),
				xlog.Int("batch-size", len(batch)),
				xlog.Err(err),
			)
			time.Sleep(sleep)
			continue
		}

		return err
	}

	return fmt.Errorf("insert failed after %d retries due to repeated deadlocks", maxRetry)
}

func (as *accounting) processGetAndSetData(ctx context.Context, date time.Time, subCategories *[]models.SubCategory) (
	storageAccountTransaction *localstorage.BadgerStorage[models.AccountJournalTransation],
	storageAccountDailyBalance *localstorage.BadgerStorage[models.AccountBalanceDaily],
	storageAccountTrialBalance *localstorage.BadgerStorage[models.AccountTrialBalance],
	storageAccounts *localstorage.BadgerStorage[models.GetAccountOut],
	err error,
) {
	now := atime.Now()
	defer func() {
		logDuration(ctx, "ProcessGetAndSetData", now)
	}()

	id := uuid.New().String()
	storageAccountTransaction, err = localstorage.NewBadgerStorage[models.AccountJournalTransation]("account_transaction" + id)
	if err != nil {
		err = models.GetErrMap(models.ErrKeyFailedToCreateStorage, err.Error())
		return
	}

	storageAccountDailyBalance, err = localstorage.NewBadgerStorage[models.AccountBalanceDaily]("account_daily_balance_" + id)
	if err != nil {
		err = models.GetErrMap(models.ErrKeyFailedToCreateStorage, err.Error())
		return
	}

	storageAccountTrialBalance, err = localstorage.NewBadgerStorage[models.AccountTrialBalance]("account_trial_balance_" + id)
	if err != nil {
		err = models.GetErrMap(models.ErrKeyFailedToCreateStorage, err.Error())
		return
	}

	storageAccounts, err = localstorage.NewBadgerStorage[models.GetAccountOut]("account_out_" + id)
	if err != nil {
		err = models.GetErrMap(models.ErrKeyFailedToCreateStorage, err.Error())
		return
	}

	allowedEntities := as.srv.conf.AccountConfig.AllowedEntitiesAccountDailyBalance
	act := as.srv.mySqlRepo.GetAccountingRepository()
	chanAccountTransaction := act.GetAccountTransactionByDate(ctx, allowedEntities, date)
	chanAccountDailyBalance := act.GetAllAccountDailyBalance(ctx, allowedEntities, subCategories, date.AddDate(0, 0, -1))
	chanAccounts := as.srv.mySqlRepo.GetAccountRepository().GetAllAccountNumber(ctx, allowedEntities, subCategories)

	eg, egCtx := errgroup.WithContext(ctx)
	eg.Go(func() error {
		now := atime.Now()
		defer logDuration(ctx, "GetAccountTransactionByDate", now)
		for v := range chanAccountTransaction {
			select {
			case <-egCtx.Done():
				return egCtx.Err()
			default:
				if v.Err != nil {
					err = v.Err
					return err
				}

				accountNumber := v.Data.AccountNumber
				newData := v.Data.Amount

				aj, err := storageAccountTransaction.Get(accountNumber)
				if err != nil && !errors.Is(err, localstorage.ErrKeyNotFound) {
					return err
				}

				accountTransaction := models.AccountJournalTransation{
					AccountNumber:  accountNumber,
					DebitMovement:  aj.DebitMovement,
					CreditMovement: aj.CreditMovement,
				}
				if v.Data.IsDebit {
					accountTransaction.DebitMovement = accountTransaction.DebitMovement.Add(newData)
				} else {
					accountTransaction.CreditMovement = accountTransaction.CreditMovement.Add(newData)
				}

				if err = storageAccountTransaction.Set(accountNumber, accountTransaction); err != nil {
					return err
				}
			}
		}
		return nil
	})

	eg.Go(func() error {
		now := atime.Now()
		defer logDuration(ctx, "GetAllAccountDailyBalance", now)
		for v := range chanAccountDailyBalance {
			select {
			case <-egCtx.Done():
				return egCtx.Err()
			default:
				if v.Err != nil {
					err = v.Err
					return err
				}

				if as.srv.conf.AccountConfig.IsSkipAllZeroAmountAccountDailyBalance {
					if v.Data.DebitMovement.IsZero() &&
						v.Data.CreditMovement.IsZero() &&
						v.Data.OpeningBalance.IsZero() &&
						v.Data.ClosingBalance.IsZero() {
						continue
					}
				}

				if err := storageAccountDailyBalance.Set(v.Data.AccountNumber, v.Data); err != nil {
					return err
				}
			}
		}
		return nil
	})

	eg.Go(func() error {
		now := atime.Now()
		defer logDuration(ctx, "GetAllAccountNumber", now)
		for v := range chanAccounts {
			select {
			case <-egCtx.Done():
				return egCtx.Err()
			default:
				if v.Err != nil {
					err = v.Err
					return err
				}

				if err = storageAccounts.Set(v.Data.AccountNumber, v.Data); err != nil {
					return err
				}
			}
		}
		return nil
	})

	if err = eg.Wait(); err != nil {
		return
	}

	return
}

func (as *accounting) calculateClosingBalance(mapSubCategory map[string]models.CategorySubCategoryCOAType, accountBalanceDaily models.AccountBalanceDaily) models.AccountBalanceDaily {
	coaTypeCode := mapSubCategory[accountBalanceDaily.SubCategoryCode].CoaTypeCode
	accountBalanceDaily.ClosingBalance = accountBalanceDaily.OpeningBalance.Add(accountBalanceDaily.DebitMovement).Sub(accountBalanceDaily.CreditMovement)
	if coaTypeCode == models.COATypeLiability {
		accountBalanceDaily.ClosingBalance = accountBalanceDaily.OpeningBalance.Add(accountBalanceDaily.CreditMovement).Sub(accountBalanceDaily.DebitMovement)
	}
	return accountBalanceDaily
}
