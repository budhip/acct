package services

import (
	"context"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"strings"
	"time"

	"bitbucket.org/Amartha/go-accounting/internal/models"
	"bitbucket.org/Amartha/go-accounting/internal/pkg/atime"
	"bitbucket.org/Amartha/go-accounting/internal/pkg/dbutil"
	"bitbucket.org/Amartha/go-accounting/internal/pkg/godbledger"
	"bitbucket.org/Amartha/go-accounting/internal/pkg/money"
	"bitbucket.org/Amartha/go-accounting/internal/pkg/queueunicorn"
	"bitbucket.org/Amartha/go-accounting/internal/repositories/mysql"

	xlog "bitbucket.org/Amartha/go-x/log"

	"github.com/hashicorp/go-multierror"
	"github.com/shopspring/decimal"
)

type JournalService interface {
	ConsumerInsertTransaction(ctx context.Context, req models.JournalRequest) (err error)
	InsertJournalTransaction(ctx context.Context, request models.JournalRequest) (out []models.JournalEntryCreatedRequest, err error)
	ProcessUploadJournal(ctx context.Context, file *multipart.FileHeader) error
	PublishJournalTransaction(ctx context.Context, req models.JournalRequest) (err error)
	GetJournalByTransactionId(ctx context.Context, transactionId string) (out []models.GetJournalDetailOut, err error)
	RetryPublishToJournalEntryCreated(ctx context.Context, data models.JournalEntryCreatedRequest) (err error)
}

type journalService service

var _ JournalService = (*journalService)(nil)

/*
1. validations
- check transaction data empty or not
- check transaction is exist or not
- validate format date
- validate transaction date & processing date
2. transform journals to transactions, splits, split_accounts & acct_journal_detail
- get account is exist or not
- generate splitId
- check account entity
5. insert into transactions, splits, split_accounts & acct_journal_detail
6. if error publish to dlq
7. if not error publish to journal_entry_created
8. if error publish to journal_entry_created_dlq
*/
func (js *journalService) ConsumerInsertTransaction(ctx context.Context, req models.JournalRequest) (err error) {
	journalEntries, err := js.InsertJournalTransaction(ctx, req)
	if err != nil {
		js.publishToJournalStreamDLQ(ctx, req, err)
		return err
	}

	for _, journal := range journalEntries {
		if err = js.publishToJournalEntryCreated(ctx, journal); err != nil {
			js.publishToJournalEntryCreatedDLQ(ctx, journal)
			return err
		}
	}

	if js.srv.flagger.IsEnabled(models.FlagTrialBalanceAutoAdjustment.String()) {
		js.detectAdjustmentTransaction(ctx, req)
	}

	return nil
}

func (js *journalService) InsertJournalTransaction(ctx context.Context, req models.JournalRequest) (journalEntries []models.JournalEntryCreatedRequest, err error) {
	ctx = dbutil.NewContextUsePrimaryDB(ctx) // make sure all query direct to primary db
	defer func() {
		logService(ctx, err)
	}()

	xlog.Info(ctx, "[JOURNAL-STREAM]",
		xlog.String("transaction-id", req.TransactionId),
		xlog.Any("data", req),
	)

	trxDate, processingDate, err := js.validateTransaction(ctx, req)
	if err != nil {
		return
	}

	now := atime.Now()
	trxDate = time.Date(trxDate.Year(), trxDate.Month(), trxDate.Day(), trxDate.Hour(), trxDate.Minute(), trxDate.Second(), now.Nanosecond(), now.Location())

	transactions, splits, splitAccounts, journals, journalEntries, err := js.changeTrxToSplits(ctx, req, trxDate, processingDate)
	if err != nil {
		return
	}

	err = js.srv.mySqlRepo.Atomic(ctx, func(actx context.Context, r mysql.SQLRepository) (err error) {
		if err = r.GetAccountingRepository().InsertTransaction(actx, transactions); err != nil {
			return err
		}

		if err = r.GetAccountingRepository().InsertSplit(actx, splits); err != nil {
			return err
		}

		if err = r.GetAccountingRepository().InsertSplitAccount(actx, splitAccounts); err != nil {
			return err
		}

		if err = r.GetAccountingRepository().InsertJournalDetail(actx, journals); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return []models.JournalEntryCreatedRequest{}, err
	}

	trx := req.Transactions[0]
	amountFloat, _ := trx.Amount.Float64()
	xlog.LogFtm(
		amountFloat,
		"journal-stream:transaction",
		trx.TransactionType,
	)

	return journalEntries, nil
}

func (js *journalService) validateTransaction(ctx context.Context, req models.JournalRequest) (trxDate, processingDate time.Time, err error) {
	if len(req.Transactions) == 0 {
		err = models.GetErrMap(models.ErrKeyTransactionsEmpty)
		return
	}

	isExist, err := js.srv.mySqlRepo.GetAccountingRepository().CheckTransactionIdIsExist(ctx, req.TransactionId)
	if err != nil {
		err = checkDatabaseError(err)
		return
	}
	if isExist {
		err = models.GetErrMap(models.ErrKeyTransactionIdIsExist)
		return
	}

	trxDate, err = atime.ParseStringToDatetime(atime.DateFormatYYYYMMDDWithTime, req.TransactionDate)
	if err != nil {
		err = models.GetErrMap(models.ErrKeyInvalidFormatDate, fmt.Sprintf("date %s format must be %s", req.TransactionDate, atime.DateFormatYYYYMMDDWithTime))
		return
	}

	processingDate, err = atime.ParseStringToDatetime(atime.DateFormatYYYYMMDDWithTime, req.ProcessingDate)
	if err != nil {
		err = models.GetErrMap(models.ErrKeyInvalidFormatDate, fmt.Sprintf("date %s format must be %s", req.ProcessingDate, atime.DateFormatYYYYMMDDWithTime))
		return
	}

	now := atime.Now()
	curDate := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 999999999, now.Location())
	if trxDate.After(curDate) || processingDate.After(curDate) {
		err = models.GetErrMap(models.ErrKeyDateIsGreaterThanToday)
		return
	}

	return
}

func (js *journalService) changeTrxToSplits(
	ctx context.Context,
	req models.JournalRequest,
	trxDate, processingDate time.Time,
) ([]models.CreateTransaction, []models.CreateSplit, []models.CreateSplitAccount, []models.CreateJournalDetail, []models.JournalEntryCreatedRequest, error) {
	len := len(req.Transactions)
	transactions := make([]models.CreateTransaction, 0, len)
	splits := make([]models.CreateSplit, 0, len)
	splitAccounts := make([]models.CreateSplitAccount, 0, len)
	journals := make([]models.CreateJournalDetail, 0, len)
	journalEntries := make([]models.JournalEntryCreatedRequest, 0, len)
	arrEntity := make([]string, 0, len)
	currency := godbledger.CurrencyIDR

	transactions = append(transactions, models.CreateTransaction{
		TransactionID: req.TransactionId,
		Postdate:      trxDate,
		PosterUserID:  godbledger.UserSystem.Id,
	})
	for _, v := range req.Transactions {
		splitId, err := js.generateSplitId(ctx)
		if err != nil {
			return nil, nil, nil, nil, nil, err
		}

		account, err := js.srv.mySqlRepo.GetAccountRepository().GetOneByAccountNumber(ctx, v.Account)
		if err != nil {
			err = checkDatabaseError(err, models.ErrKeyAccountNumberNotFound)
			return nil, nil, nil, nil, nil, err
		}

		arrEntity = append(arrEntity, account.EntityCode)

		amount := money.FormatAmountToBigInt(v.Amount, currency.Decimals)
		splits = append(splits, models.CreateSplit{
			SplitID:       splitId,
			SplitDate:     processingDate,
			Description:   v.Narrative,
			Currency:      currency.Name,
			Amount:        amount.BigInt().Int64(),
			TransactionID: req.TransactionId,
		})

		splitAccounts = append(splitAccounts, models.CreateSplitAccount{
			SplitID:   splitId,
			AccountID: account.AccountNumber,
		})

		journals = append(journals, models.CreateJournalDetail{
			JournalId:       splitId,
			ReferenceNumber: req.ReferenceNumber,
			OrderType:       req.OrderType,
			TransactionType: v.TransactionType,
			TransactionDate: trxDate,
			IsDebit:         v.IsDebit,
			Metadata:        req.Metadata,
		})

		journalEntries = append(journalEntries, models.JournalEntryCreatedRequest{
			TransactionID:   req.TransactionId,
			ReferenceNumber: req.ReferenceNumber,
			JournalID:       splitId,
			AccountNumber:   account.AccountNumber,
			AccountName:     account.AccountName,
			Amount:          amount.BigInt().Int64(),
			IsDebit:         v.IsDebit,
			OrderType:       req.OrderType,
			TransactionDate: trxDate.Format(atime.DateFormatYYYYMMDD),
			TransactionType: v.TransactionType,
			EntityCode:      account.EntityCode,
			CategoryCode:    account.CategoryCode,
			SubCategoryCode: account.SubCategoryCode,
			NormalBalance:   account.CoaTypeCode,
			CreatedAt:       atime.Now().UTC(),
		})
	}

	if isSame := js.allSameEntity(arrEntity); !isSame {
		return nil, nil, nil, nil, nil, models.GetErrMap(models.ErrKeyAccountNumberDifferentEntity)
	}

	return transactions, splits, splitAccounts, journals, journalEntries, nil
}

func (js *journalService) generateSplitId(ctx context.Context) (string, error) {
	lastSequence, err := js.srv.cacheRepo.GetIncrement(ctx, "splitIdCounter")
	if err != nil {
		return "", models.GetErrMap(models.ErrKeyFailedGetFromCache, err.Error())
	}
	splitId := fmt.Sprintf("%s%s", atime.Now().Format(atime.DateFormatYYYYMMDDWithoutDash), leftZeroPad(lastSequence, js.srv.conf.JournalConfig.SplitIdPadWidth))
	return splitId, nil
}

func (js *journalService) PublishJournalTransaction(ctx context.Context, req models.JournalRequest) (err error) {
	defer func() {
		logService(ctx, err)
	}()

	_, _, err = js.validateTransaction(ctx, req)
	if err != nil {
		return
	}

	for _, v := range req.Transactions {
		isExistAccountNumber, err := js.srv.mySqlRepo.GetAccountRepository().CheckAccountNumberIsExist(ctx, v.Account)
		if err != nil {
			err = checkDatabaseError(err)
			return err
		}
		if isExistAccountNumber == nil {
			err = models.GetErrMap(models.ErrKeyAccountNumberNotFound, v.Account)
			return err
		}
	}

	if err = js.publishToJournalStream(ctx, req); err != nil {
		return
	}
	return
}

func (js *journalService) RetryPublishToJournalEntryCreated(ctx context.Context, data models.JournalEntryCreatedRequest) (err error) {
	defer func() {
		logService(ctx, err)
	}()

	if err = js.publishToJournalEntryCreated(ctx, data); err != nil {
		return js.publishToJournalEntryCreatedDLQ(ctx, data)
	}

	return nil
}

func (js *journalService) publishToJournalStream(ctx context.Context, data models.JournalRequest) error {
	return js.srv.publisher.PublishSyncWithKeyAndLog(ctx, "publish transaction to journal stream", js.srv.conf.Kafka.Publishers.JournalStream.Topic, data.TransactionId, data)
}

func (js *journalService) publishToJournalEntryCreated(ctx context.Context, data models.JournalEntryCreatedRequest) error {
	return js.srv.publisher.PublishSyncWithKeyAndLog(ctx, "publish transaction to journal_entry_created", js.srv.conf.Kafka.Publishers.JournalEntryCreated.Topic, data.JournalID, data)
}

func (js *journalService) publishToJournalEntryCreatedDLQ(ctx context.Context, data models.JournalEntryCreatedRequest) error {
	return js.srv.publisher.PublishSyncWithKeyAndLog(ctx, "publish transaction to journal_entry_created.dlq", js.srv.conf.Kafka.Publishers.JournalEntryCreatedDLQ.Topic, data.JournalID, data)
}

func (js *journalService) publishToJournalStreamDLQ(ctx context.Context, data models.JournalRequest, err error) error {
	messages := models.JournalError{
		JournalRequest: data,
		ErrCauser:      err,
	}
	return js.srv.publisher.PublishSyncWithKeyAndLog(ctx, "publish transaction to journal_stream.dlq", js.srv.conf.Kafka.Publishers.JournalStreamDLQ.Topic, data.TransactionId, messages)
}

// TODO: will be deleted (temporary for testing purposes)
func (js *journalService) ProcessUploadJournal(ctx context.Context, file *multipart.FileHeader) (err error) {
	defer func() {
		logService(ctx, err)
	}()

	path := fmt.Sprintf("./%s_%s", atime.Now().Format(atime.DateFormatYYYYMMDDWithTimeWithoutDash), file.Filename)
	src, err := file.Open()
	if err != nil {
		return
	}
	defer src.Close()

	dst, err := js.srv.file.CreateFile(path)
	if err != nil {
		return
	}
	defer dst.Close()

	if err = js.srv.file.CopyFile(dst, src); err != nil {
		return err
	}

	defer js.srv.file.RemoveFile(dst.Name())
	fs, err := js.srv.file.OpenFile(dst.Name())
	if err != nil {
		return
	}

	records, err := js.srv.file.CSVReadAll(fs)
	if err != nil {
		return
	}

	fmtErr := func(data []string, err error) error {
		return fmt.Errorf("error process data %s caused by - %v ", strings.Join(data, ","), err)
	}

	var (
		journals []models.JournalRequest
		errs     *multierror.Error
	)
	for _, r := range records[1:] {
		amount, err := decimal.NewFromString(r[11])
		if err != nil {
			errs = multierror.Append(errs, fmtErr(r, err))
			continue
		}
		journal := models.JournalRequest{
			ReferenceNumber: strings.TrimSpace(r[0]),
			TransactionId:   strings.TrimSpace(r[1]),
			OrderType:       strings.TrimSpace(r[2]),
			TransactionDate: strings.TrimSpace(r[3]),
			ProcessingDate:  strings.TrimSpace(r[4]),
			Currency:        strings.TrimSpace(r[5]),
			Transactions: []models.Transaction{
				{
					TransactionType:     strings.TrimSpace(r[6]),
					TransactionTypeName: strings.TrimSpace(r[7]),
					Account:             strings.TrimSpace(r[8]),
					Narrative:           strings.TrimSpace(r[10]),
					Amount:              amount,
					IsDebit:             true,
				},
				{
					TransactionType:     strings.TrimSpace(r[6]),
					TransactionTypeName: strings.TrimSpace(r[7]),
					Account:             strings.TrimSpace(r[9]),
					Narrative:           strings.TrimSpace(r[10]),
					Amount:              amount,
					IsDebit:             false,
				},
			},
		}

		if r[18] != "" {
			if err := json.Unmarshal([]byte(strings.TrimSpace(r[18])), &journal.Metadata); err != nil {
				errs = multierror.Append(errs, fmtErr(r, err))
				continue
			}
		}

		if r[12] != "" && r[14] != "" && r[15] != "" && r[17] != "" {
			amount, err := decimal.NewFromString(r[17])
			if err != nil {
				errs = multierror.Append(errs, fmtErr(r, err))
				continue
			}
			journal.Transactions = append(journal.Transactions,
				models.Transaction{
					TransactionType:     strings.TrimSpace(r[12]),
					TransactionTypeName: strings.TrimSpace(r[13]),
					Account:             strings.TrimSpace(r[14]),
					Narrative:           strings.TrimSpace(r[16]),
					Amount:              amount,
					IsDebit:             true,
				}, models.Transaction{
					TransactionType:     strings.TrimSpace(r[12]),
					TransactionTypeName: strings.TrimSpace(r[13]),
					Account:             strings.TrimSpace(r[15]),
					Narrative:           strings.TrimSpace(r[16]),
					Amount:              amount,
					IsDebit:             false,
				})
		}
		_, _, err = js.validateTransaction(ctx, journal)
		if err != nil {
			errs = multierror.Append(errs, err)
			continue
		}
		journals = append(journals, journal)
	}
	if errs.ErrorOrNil() != nil {
		err = errs
		return
	}

	for _, v := range journals {
		if err = js.publishToJournalStream(ctx, v); err != nil {
			errs = multierror.Append(errs, err)
			xlog.Error(ctx, "[PROCESS.UPLOAD.JOURNAL]", xlog.Any("data", v), xlog.Err(err))
		}
	}
	if errs.ErrorOrNil() != nil {
		err = errs
		return
	}

	return
}

func (js *journalService) GetJournalByTransactionId(ctx context.Context, transactionId string) (out []models.GetJournalDetailOut, err error) {
	defer func() {
		logService(ctx, err)
	}()

	isExist, err := js.srv.mySqlRepo.GetAccountingRepository().CheckTransactionIdIsExist(ctx, transactionId)
	if err != nil {
		err = checkDatabaseError(err)
		return
	}
	if !isExist {
		err = models.GetErrMap(models.ErrKeyTransactionIdNotFound)
		return
	}

	out, err = js.srv.mySqlRepo.GetAccountingRepository().GetJournalDetailByTransactionId(
		ctx, transactionId,
	)
	if err != nil {
		err = checkDatabaseError(err)
		return
	}

	return out, err
}

func (js *journalService) detectAdjustmentTransaction(ctx context.Context, req models.JournalRequest) {
	var err error
	logPrefix := "[Adjustment-TrialBalance]"

	defer func() {
		logService(ctx, err)
	}()

	xlog.Info(ctx, logPrefix, xlog.String("adjustment-date", req.TransactionDate))

	// check adjustment trx date not equal today
	trxDate, err := atime.ParseStringToDatetime(atime.DateFormatYYYYMMDDWithTime, req.TransactionDate)
	if err != nil {
		return
	}

	if isEqual := atime.DateEqualToday(trxDate); isEqual {
		return
	}

	periodDate := trxDate.Format(atime.DateFormatYYYYMM)
	// get first period with status open
	tb, err := js.srv.mySqlRepo.GetTrialBalanceRepository().GetFirstPeriodByStatus(ctx, models.TrialBalanceStatusOpen)
	if err != nil {
		err = checkDatabaseError(err)
		return
	}

	// skip if no period with status open OR skip if period_date not equal with transaction_date
	if tb == nil ||
		(tb.IsAdjustment && atime.DateEqualToday(tb.UpdatedAt)) ||
		tb.Period != periodDate {
		xlog.Info(ctx, logPrefix, xlog.String("description", "no period date with status open OR period_date not equal with transaction_date"))
		return
	}

	// update period adjustment is true
	if err = js.srv.mySqlRepo.GetTrialBalanceRepository().UpdateTrialBalanceAdjustment(ctx, models.CloseTrialBalanceRequest{
		Period: periodDate,
	}); err != nil {
		err = checkDatabaseError(err, models.ErrKeyClosedPeriodNotFound)
		return
	}

	// schedule
	reqGQU := queueunicorn.RequestJobHTTP{
		Name: queueunicorn.HttpRequestJobKey,
		Payload: queueunicorn.PayloadJob{
			Host:   fmt.Sprintf("%s/%s", js.srv.conf.HostGoAccounting, "v1/trial-balances/adjustment"),
			Method: http.MethodPost,
			Body: models.AdjustmentTrialBalanceRequest{
				AdjustmentDate: trxDate.Format(atime.DateFormatYYYYMMDD),
				IsManual:       false,
			},
			Headers: queueunicorn.RequestHeaderJob(js.srv.conf.SecretKey, req.TransactionId),
		},
		Options: queueunicorn.Options{
			ProcessIn: js.srv.conf.GQUTrialBalance.ProcessIn,
			MaxRetry:  js.srv.conf.GQUTrialBalance.MaxRetry,
			// Timeout:   10,
			// Deadline:  1000,
		},
	}
	if err = js.srv.queueUnicornClient.SendJobHTTP(ctx, reqGQU); err != nil {
		return
	}
}

func (js *journalService) allSameEntity(arr []string) bool {
	if len(arr)%2 != 0 || len(arr) == 0 {
		return false
	}

	for i := 0; i < len(arr); i += 2 {
		if arr[i] != arr[i+1] {
			return false
		}
	}

	return true
}
