package services

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"bitbucket.org/Amartha/go-accounting/internal/models"
	"bitbucket.org/Amartha/go-accounting/internal/pkg/atime"
	"bitbucket.org/Amartha/go-accounting/internal/pkg/dddnotification"
	"bitbucket.org/Amartha/go-accounting/internal/repositories/mysql"
	xlog "bitbucket.org/Amartha/go-x/log"
)

type MigrationService interface {
	BucketsJournalLoad(ctx context.Context, req models.MigrationBucketsJournalLoadRequest) (err error)
}

type migrationService service

var _ MigrationService = (*migrationService)(nil)

func (s *migrationService) BucketsJournalLoad(ctx context.Context, req models.MigrationBucketsJournalLoadRequest) (err error) {
	startTime := time.Now()
	operation := fmt.Sprintf("Load Journal From Buckets: %s/%s", s.srv.conf.Migration.Buckets, req.SubFolder)
	s.sendMessageToSlack(ctx, operation, "Start")
	defer func() {
		logService(ctx, err)

		message := fmt.Sprintf("Finished; Elapsed Time: %s\n", time.Since(startTime).String())
		if err != nil {
			message = fmt.Sprintf("Failed with err: %s", err.Error())
		}
		s.sendMessageToSlack(ctx, operation, message)
	}()

	if err = s.srv.mySqlRepo.Atomic(ctx, func(atomicCtx context.Context, r mysql.SQLRepository) (err error) {
		// Disable dependency check, so we can insert all table at once
		r.GetAccountingRepository().ToggleForeignKeyChecks(atomicCtx, false)
		defer r.GetAccountingRepository().ToggleForeignKeyChecks(atomicCtx, true)

		wg := new(sync.WaitGroup)
		mu := sync.Mutex{}

		// Prevent race condition when got error at same time
		updateErr := func(newErr error) {
			mu.Lock()
			defer mu.Unlock()
			err = newErr
		}

		// Journal Detail
		wg.Add(1)
		go func() {
			defer wg.Done()
			for data := range s.populateJournalDetail(atomicCtx, req) {
				if data.Err != nil {
					updateErr(fmt.Errorf("failed to populate journal detail: %w", data.Err))
					return
				}

				errNested := r.GetAccountingRepository().InsertJournalDetail(atomicCtx, data.Data)
				if errNested != nil {
					updateErr(fmt.Errorf("unable to insert journal detail: %w", errNested))
					return
				}
			}
		}()

		// Transaction
		wg.Add(1)
		go func() {
			defer wg.Done()
			for data := range s.populateTransaction(atomicCtx, req) {
				if data.Err != nil {
					updateErr(fmt.Errorf("failed to populate transaction: %w", data.Err))
					return
				}

				errNested := r.GetAccountingRepository().InsertTransaction(atomicCtx, data.Data)
				if errNested != nil {
					updateErr(fmt.Errorf("unable to insert transaction: %w", errNested))
					return
				}
			}
		}()

		// Splits
		wg.Add(1)
		go func() {
			defer wg.Done()
			for data := range s.populateSplit(atomicCtx, req) {
				if data.Err != nil {
					updateErr(fmt.Errorf("failed to populate split: %w", data.Err))
					return
				}

				errNested := r.GetAccountingRepository().InsertSplit(atomicCtx, data.Data)
				if errNested != nil {
					updateErr(fmt.Errorf("unable to insert split: %w", errNested))
					return
				}
			}
		}()

		// Split Account
		wg.Add(1)
		go func() {
			defer wg.Done()
			for data := range s.populateSplitAccount(atomicCtx, req) {
				if data.Err != nil {
					updateErr(fmt.Errorf("failed to populate split account: %w", data.Err))
					return
				}

				errNested := r.GetAccountingRepository().InsertSplitAccount(atomicCtx, data.Data)
				if errNested != nil {
					updateErr(fmt.Errorf("unable to insert split account: %w", errNested))
					return
				}
			}
		}()

		wg.Wait()
		return
	}); err != nil {
		fmt.Printf("err: %v\n", err)
		return
	}

	return nil
}

func (s *migrationService) populateJournalDetail(ctx context.Context, req models.MigrationBucketsJournalLoadRequest) chan models.StreamArrayResult[models.CreateJournalDetail] {
	result := make(chan models.StreamArrayResult[models.CreateJournalDetail])
	go func() {
		defer close(result)

		streamResult, err := s.srv.storageRepo.Stream(ctx, s.srv.conf.Migration.Buckets, fmt.Sprint(req.SubFolder, "/5_pas_acct_journal_detail.csv"))
		if err != nil {
			result <- models.StreamArrayResult[models.CreateJournalDetail]{Err: fmt.Errorf("unable to stream: %w", err)}
			return
		}

		journalDetails := []models.CreateJournalDetail{}
		counter := 0
		for res := range streamResult {
			if res.Err != nil {
				result <- models.StreamArrayResult[models.CreateJournalDetail]{Err: fmt.Errorf("failed to read file: %w", res.Err)}
				return
			}

			// Process data

			// split string into array
			dataArr := strings.Split(res.Data, ",")
			if len(dataArr) != 6 {
				result <- models.StreamArrayResult[models.CreateJournalDetail]{Err: fmt.Errorf("invalid data: %s", res.Data)}
				return
			}

			// parse string into date
			trxDate, err := time.Parse(atime.DateFormatRFC3339NanoWithTimeZone, dataArr[3])
			if err != nil {
				result <- models.StreamArrayResult[models.CreateJournalDetail]{Err: fmt.Errorf("invalid date (%s): %w", dataArr[4], err)}
				return
			}

			// parse string into bool
			isDebit, err := strconv.ParseBool(dataArr[5])
			if err != nil {
				result <- models.StreamArrayResult[models.CreateJournalDetail]{Err: fmt.Errorf("invalid bool (%s): %w", dataArr[5], err)}
				return
			}

			// Insert bulk data
			journalDetails = append(journalDetails, models.CreateJournalDetail{
				JournalId:       dataArr[0],
				OrderType:       dataArr[1],
				TransactionType: dataArr[2],
				TransactionDate: trxDate,
				ReferenceNumber: dataArr[4],
				IsDebit:         isDebit,
			})

			counter++
			if counter == s.srv.conf.SQLTransaction.BulkLimit {
				result <- models.StreamArrayResult[models.CreateJournalDetail]{Data: journalDetails}
				journalDetails = []models.CreateJournalDetail{}
				counter = 0
			}
		}

		result <- models.StreamArrayResult[models.CreateJournalDetail]{Data: journalDetails}
	}()

	return result
}

func (s *migrationService) populateTransaction(ctx context.Context, req models.MigrationBucketsJournalLoadRequest) chan models.StreamArrayResult[models.CreateTransaction] {
	result := make(chan models.StreamArrayResult[models.CreateTransaction])
	go func() {
		defer close(result)

		streamResult, err := s.srv.storageRepo.Stream(ctx, s.srv.conf.Migration.Buckets, fmt.Sprint(req.SubFolder, "/2_pas_transactions.csv"))
		if err != nil {
			result <- models.StreamArrayResult[models.CreateTransaction]{Err: fmt.Errorf("unable to stream: %w", err)}
			return
		}

		newData := []models.CreateTransaction{}
		counter := 0
		for res := range streamResult {
			if res.Err != nil {
				result <- models.StreamArrayResult[models.CreateTransaction]{Err: fmt.Errorf("failed to read file: %w", res.Err)}
				return
			}

			// Process data

			// split string into array
			dataArr := strings.Split(res.Data, ",")
			if len(dataArr) != 3 {
				result <- models.StreamArrayResult[models.CreateTransaction]{Err: fmt.Errorf("invalid data: %s", res.Data)}
				return
			}

			// parse string into date
			postDate, err := time.Parse(atime.DateFormatRFC3339NanoWithTimeZone, dataArr[1])
			if err != nil {
				result <- models.StreamArrayResult[models.CreateTransaction]{Err: fmt.Errorf("invalid date (%s): %w", dataArr[1], err)}
				return
			}

			// Insert bulk data
			newData = append(newData, models.CreateTransaction{
				TransactionID: dataArr[0],
				Postdate:      postDate,
				PosterUserID:  dataArr[2],
			})

			counter++
			if counter == s.srv.conf.SQLTransaction.BulkLimit {
				result <- models.StreamArrayResult[models.CreateTransaction]{Data: newData}
				newData = []models.CreateTransaction{}
				counter = 0
			}
		}

		result <- models.StreamArrayResult[models.CreateTransaction]{Data: newData}
	}()

	return result
}

func (s *migrationService) populateSplit(ctx context.Context, req models.MigrationBucketsJournalLoadRequest) chan models.StreamArrayResult[models.CreateSplit] {
	result := make(chan models.StreamArrayResult[models.CreateSplit])
	go func() {
		defer close(result)

		streamResult, err := s.srv.storageRepo.Stream(ctx, s.srv.conf.Migration.Buckets, fmt.Sprint(req.SubFolder, "/3_pas_splits.csv"))
		if err != nil {
			result <- models.StreamArrayResult[models.CreateSplit]{Err: fmt.Errorf("unable to stream: %w", err)}
			return
		}

		newData := []models.CreateSplit{}
		counter := 0
		for res := range streamResult {
			if res.Err != nil {
				result <- models.StreamArrayResult[models.CreateSplit]{Err: fmt.Errorf("failed to read file: %w", res.Err)}
				return
			}

			// Process data

			// split string into array
			dataArr := strings.Split(res.Data, ",")
			if len(dataArr) != 6 {
				result <- models.StreamArrayResult[models.CreateSplit]{Err: fmt.Errorf("invalid data: %s", res.Data)}
				return
			}

			// parse string to date
			splitDate, err := time.Parse(atime.DateFormatRFC3339NanoWithTimeZone, dataArr[1])
			if err != nil {
				result <- models.StreamArrayResult[models.CreateSplit]{Err: fmt.Errorf("invalid date (%s): %w", dataArr[1], err)}
				return
			}

			// parse string to int64
			amount, err := strconv.ParseInt(dataArr[4], 10, 64)
			if err != nil {
				result <- models.StreamArrayResult[models.CreateSplit]{Err: fmt.Errorf("invalid amount (%s): %w", dataArr[4], err)}
				return
			}

			// Insert bulk data
			newData = append(newData, models.CreateSplit{
				SplitID:       dataArr[0],
				SplitDate:     splitDate,
				Description:   dataArr[2],
				Currency:      dataArr[3],
				Amount:        amount,
				TransactionID: dataArr[5],
			})

			counter++
			if counter == s.srv.conf.SQLTransaction.BulkLimit {
				result <- models.StreamArrayResult[models.CreateSplit]{Data: newData}
				newData = []models.CreateSplit{}
				counter = 0
			}
		}

		result <- models.StreamArrayResult[models.CreateSplit]{Data: newData}
	}()

	return result
}

func (s *migrationService) populateSplitAccount(ctx context.Context, req models.MigrationBucketsJournalLoadRequest) chan models.StreamArrayResult[models.CreateSplitAccount] {
	result := make(chan models.StreamArrayResult[models.CreateSplitAccount])
	go func() {
		defer close(result)

		streamResult, err := s.srv.storageRepo.Stream(ctx, s.srv.conf.Migration.Buckets, fmt.Sprint(req.SubFolder, "/4_pas_split_accounts.csv"))

		if err != nil {
			result <- models.StreamArrayResult[models.CreateSplitAccount]{Err: fmt.Errorf("unable to stream: %w", err)}
			return
		}

		newData := []models.CreateSplitAccount{}
		counter := 0
		for res := range streamResult {
			if res.Err != nil {
				result <- models.StreamArrayResult[models.CreateSplitAccount]{Err: fmt.Errorf("failed to read file: %w", res.Err)}
				return
			}

			// Process data

			// split string into array
			dataArr := strings.Split(res.Data, ",")
			if len(dataArr) != 2 {
				result <- models.StreamArrayResult[models.CreateSplitAccount]{Err: fmt.Errorf("invalid data: %s", res.Data)}
				return
			}

			// Insert bulk data
			newData = append(newData, models.CreateSplitAccount{
				SplitID:   dataArr[0],
				AccountID: dataArr[1],
			})

			counter++
			if counter == s.srv.conf.SQLTransaction.BulkLimit {
				result <- models.StreamArrayResult[models.CreateSplitAccount]{Data: newData}
				newData = []models.CreateSplitAccount{}
				counter = 0
			}
		}

		result <- models.StreamArrayResult[models.CreateSplitAccount]{Data: newData}
	}()

	return result
}

func (s *migrationService) sendMessageToSlack(ctx context.Context, operation, message string) {
	if err := s.srv.dddNotificationClient.SendMessageToSlack(ctx, dddnotification.MessageData{
		Operation: operation,
		Message:   message,
	}); err != nil {
		xlog.Error(ctx, "[PROCESS-JOB]", xlog.String("operation", operation), xlog.String("description", "failed to send slack message"), xlog.Err(err))
	}
}
