package services_test

import (
	"context"
	"mime/multipart"
	"testing"

	"bitbucket.org/Amartha/go-accounting/internal/models"
	"bitbucket.org/Amartha/go-accounting/internal/pkg/atime"
	"bitbucket.org/Amartha/go-accounting/internal/repositories/mysql"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func Test_journalService_ConsumerInsertTransaction(t *testing.T) {
	testHelper := serviceTestHelper(t)
	ctx := context.Background()
	req := models.JournalRequest{
		TransactionId:   uuid.New().String(),
		TransactionDate: atime.DateFormatYYYYMMDDWithTime,
		ProcessingDate:  atime.DateFormatYYYYMMDDWithTime,
		Currency:        "IDR",
		Transactions: []models.Transaction{
			{
				Account:         "TEST1",
				TransactionType: "TEST1",
				Amount:          decimal.NewFromFloat(10000),
			},
			{
				Account:         "TEST2",
				TransactionType: "TEST2",
				Amount:          decimal.NewFromFloat(10000),
			},
		},
	}

	tests := []struct {
		name    string
		req     models.JournalRequest
		doMock  func(ctx context.Context, req models.JournalRequest)
		wantErr bool
	}{
		{
			name: "success case",
			req:  req,
			doMock: func(ctx context.Context, req models.JournalRequest) {
				testHelper.mockAcctRepository.EXPECT().
					CheckTransactionIdIsExist(gomock.Any(), gomock.Any()).
					Return(false, nil)
				testHelper.mockCacheRepository.EXPECT().
					GetIncrement(gomock.Any(), "splitIdCounter").
					Return(int64(1), nil).MaxTimes(2)
				testHelper.mockAccRepository.EXPECT().
					GetOneByAccountNumber(gomock.Any(), gomock.Any()).
					Return(models.GetAccountOut{
						AccountNumber: "TEST",
						EntityCode:    "001",
					}, nil)
				testHelper.mockAccRepository.EXPECT().
					GetOneByAccountNumber(gomock.Any(), gomock.Any()).
					Return(models.GetAccountOut{
						AccountNumber: "TEST1",
						EntityCode:    "001",
					}, nil)
				testHelper.mockMySQLRepository.EXPECT().
					Atomic(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, steps func(ctx context.Context, r mysql.SQLRepository) error) error {
						testHelper.mockAcctRepository.EXPECT().InsertTransaction(gomock.Any(), gomock.Any()).Return(nil)
						testHelper.mockAcctRepository.EXPECT().InsertSplit(gomock.Any(), gomock.Any()).Return(nil)
						testHelper.mockAcctRepository.EXPECT().InsertSplitAccount(gomock.Any(), gomock.Any()).Return(nil)
						testHelper.mockAcctRepository.EXPECT().InsertJournalDetail(gomock.Any(), gomock.Any()).Return(nil)
						return steps(ctx, testHelper.mockMySQLRepository)
					})
				testHelper.mockPublisher.EXPECT().
					PublishSyncWithKeyAndLog(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					AnyTimes()
				testHelper.mockFlag.EXPECT().
					IsEnabled(models.FlagTrialBalanceAutoAdjustment.String()).
					Return(true)
				testHelper.mockTrialBalanceRepository.EXPECT().
					GetFirstPeriodByStatus(gomock.Any(), gomock.Any()).
					Return(&models.TrialBalancePeriod{
						Period: atime.DateFormatYYYYMM,
					}, nil)
				testHelper.mockTrialBalanceRepository.EXPECT().
					UpdateTrialBalanceAdjustment(gomock.Any(), gomock.Any()).
					Return(nil)
				testHelper.mockQueueUnicorn.EXPECT().
					SendJobHTTP(gomock.Any(), gomock.Any()).
					Return(nil)
			},
			wantErr: false,
		},
		{
			name: "error case - transactios empty",
			req: models.JournalRequest{
				TransactionId:   uuid.New().String(),
				TransactionDate: atime.DateFormatYYYYMMDD,
				ProcessingDate:  atime.DateFormatYYYYMMDD,
				Currency:        "IDR",
			},
			doMock: func(ctx context.Context, req models.JournalRequest) {
				testHelper.mockPublisher.EXPECT().
					PublishSyncWithKeyAndLog(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					AnyTimes()
			},
			wantErr: true,
		},
		{
			name: "error case - database error find transaction",
			req:  req,
			doMock: func(ctx context.Context, req models.JournalRequest) {
				testHelper.mockAcctRepository.EXPECT().
					CheckTransactionIdIsExist(gomock.Any(), gomock.Any()).
					Return(false, assert.AnError)
				testHelper.mockPublisher.EXPECT().
					PublishSyncWithKeyAndLog(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					AnyTimes()
			},
			wantErr: true,
		},
		{
			name: "error case - transaction id is exist",
			req:  req,
			doMock: func(ctx context.Context, req models.JournalRequest) {
				testHelper.mockAcctRepository.EXPECT().
					CheckTransactionIdIsExist(gomock.Any(), gomock.Any()).
					Return(true, nil)
				testHelper.mockPublisher.EXPECT().
					PublishSyncWithKeyAndLog(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					AnyTimes()
			},
			wantErr: true,
		},
		{
			name: "error case - invalid format transaction date",
			req: models.JournalRequest{
				TransactionId:   uuid.New().String(),
				TransactionDate: atime.DateFormatYYYYMMDDWithTimeWithoutColon,
				ProcessingDate:  atime.DateFormatYYYYMMDDWithTime,
				Currency:        "IDR",
				Transactions: []models.Transaction{
					{
						Account:         "TEST",
						TransactionType: "TEST",
						Amount:          decimal.NewFromFloat(10000),
					},
				},
			},
			doMock: func(ctx context.Context, req models.JournalRequest) {
				testHelper.mockAcctRepository.EXPECT().
					CheckTransactionIdIsExist(gomock.Any(), gomock.Any()).
					Return(false, nil)
				testHelper.mockPublisher.EXPECT().
					PublishSyncWithKeyAndLog(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					AnyTimes()
			},
			wantErr: true,
		},
		{
			name: "error case - invalid format processing date",
			req: models.JournalRequest{
				TransactionId:   uuid.New().String(),
				TransactionDate: atime.DateFormatYYYYMMDDWithTime,
				ProcessingDate:  atime.DateFormatYYYYMMDDWithTimeWithoutColon,
				Currency:        "IDR",
				Transactions: []models.Transaction{
					{
						Account:         "TEST",
						TransactionType: "TEST",
						Amount:          decimal.NewFromFloat(10000),
					},
				},
			},
			doMock: func(ctx context.Context, req models.JournalRequest) {
				testHelper.mockAcctRepository.EXPECT().
					CheckTransactionIdIsExist(gomock.Any(), gomock.Any()).
					Return(false, nil)
				testHelper.mockPublisher.EXPECT().
					PublishSyncWithKeyAndLog(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					AnyTimes()
			},
			wantErr: true,
		},
		{
			name: "error case - invalid date",
			req: models.JournalRequest{
				TransactionId:   uuid.New().String(),
				TransactionDate: "3006-01-02 15:04:05",
				ProcessingDate:  "3006-01-02 15:04:05",
				Currency:        "IDR",
				Transactions: []models.Transaction{
					{
						Account:         "TEST",
						TransactionType: "TEST",
						Amount:          decimal.NewFromFloat(10000),
					},
				},
			},
			doMock: func(ctx context.Context, req models.JournalRequest) {
				testHelper.mockAcctRepository.EXPECT().
					CheckTransactionIdIsExist(gomock.Any(), gomock.Any()).
					Return(false, nil)
				testHelper.mockPublisher.EXPECT().
					PublishSyncWithKeyAndLog(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					AnyTimes()
			},
			wantErr: true,
		},
		{
			name: "error case - failed get split id Counter",
			req:  req,
			doMock: func(ctx context.Context, req models.JournalRequest) {
				testHelper.mockAcctRepository.EXPECT().
					CheckTransactionIdIsExist(gomock.Any(), gomock.Any()).
					Return(false, nil)
				testHelper.mockCacheRepository.EXPECT().
					GetIncrement(gomock.Any(), "splitIdCounter").
					Return(int64(1), assert.AnError)
				testHelper.mockPublisher.EXPECT().
					PublishSyncWithKeyAndLog(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					AnyTimes()
			},
			wantErr: true,
		},
		{
			name: "error case - database error get account",
			req:  req,
			doMock: func(ctx context.Context, req models.JournalRequest) {
				testHelper.mockAcctRepository.EXPECT().
					CheckTransactionIdIsExist(gomock.Any(), gomock.Any()).
					Return(false, nil)
				testHelper.mockCacheRepository.EXPECT().
					GetIncrement(gomock.Any(), "splitIdCounter").Return(int64(1), nil).
					MaxTimes(2)
				testHelper.mockAccRepository.EXPECT().
					GetOneByAccountNumber(gomock.Any(), gomock.Any()).
					Return(models.GetAccountOut{}, assert.AnError)
				testHelper.mockPublisher.EXPECT().
					PublishSyncWithKeyAndLog(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					AnyTimes()
			},
			wantErr: true,
		},
		{
			name: "error case - account not found",
			req:  req,
			doMock: func(ctx context.Context, req models.JournalRequest) {
				testHelper.mockAcctRepository.EXPECT().
					CheckTransactionIdIsExist(gomock.Any(), gomock.Any()).
					Return(false, nil)
				testHelper.mockCacheRepository.EXPECT().
					GetIncrement(gomock.Any(), "splitIdCounter").
					Return(int64(1), nil).MaxTimes(2)
				testHelper.mockAccRepository.EXPECT().
					GetOneByAccountNumber(gomock.Any(), gomock.Any()).
					Return(models.GetAccountOut{}, models.ErrNoRows)
				testHelper.mockPublisher.EXPECT().
					PublishSyncWithKeyAndLog(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					AnyTimes()
			},
			wantErr: true,
		},
		{
			name: "error case - account entity is diffrent",
			req:  req,
			doMock: func(ctx context.Context, req models.JournalRequest) {
				testHelper.mockAcctRepository.EXPECT().
					CheckTransactionIdIsExist(gomock.Any(), gomock.Any()).
					Return(false, nil)
				testHelper.mockCacheRepository.EXPECT().
					GetIncrement(gomock.Any(), "splitIdCounter").
					Return(int64(1), nil).MaxTimes(2)
				testHelper.mockAccRepository.EXPECT().
					GetOneByAccountNumber(gomock.Any(), gomock.Any()).
					Return(models.GetAccountOut{
						AccountNumber: "TEST",
						EntityCode:    "001",
					}, nil)
				testHelper.mockAccRepository.EXPECT().
					GetOneByAccountNumber(gomock.Any(), gomock.Any()).
					Return(models.GetAccountOut{
						AccountNumber: "TEST1",
						EntityCode:    "002",
					}, nil)
				testHelper.mockPublisher.EXPECT().
					PublishSyncWithKeyAndLog(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					AnyTimes()
			},
			wantErr: true,
		},
		{
			name: "error case - failed insert transaction",
			req:  req,
			doMock: func(ctx context.Context, req models.JournalRequest) {
				testHelper.mockAcctRepository.EXPECT().
					CheckTransactionIdIsExist(gomock.Any(), gomock.Any()).
					Return(false, nil)
				testHelper.mockCacheRepository.EXPECT().
					GetIncrement(gomock.Any(), "splitIdCounter").
					Return(int64(1), nil).MaxTimes(2)
				testHelper.mockAccRepository.EXPECT().
					GetOneByAccountNumber(gomock.Any(), gomock.Any()).
					Return(models.GetAccountOut{
						AccountNumber: "TEST",
						EntityCode:    "001",
					}, nil)
				testHelper.mockAccRepository.EXPECT().
					GetOneByAccountNumber(gomock.Any(), gomock.Any()).
					Return(models.GetAccountOut{
						AccountNumber: "TEST1",
						EntityCode:    "001",
					}, nil)
				testHelper.mockMySQLRepository.EXPECT().
					Atomic(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, steps func(ctx context.Context, r mysql.SQLRepository) error) error {
						testHelper.mockAcctRepository.EXPECT().InsertTransaction(gomock.Any(), gomock.Any()).Return(assert.AnError)
						return steps(ctx, testHelper.mockMySQLRepository)
					})
				testHelper.mockPublisher.EXPECT().
					PublishSyncWithKeyAndLog(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					AnyTimes()
			},
			wantErr: true,
		},
		{
			name: "error case - failed insert splits",
			req:  req,
			doMock: func(ctx context.Context, req models.JournalRequest) {
				testHelper.mockAcctRepository.EXPECT().
					CheckTransactionIdIsExist(gomock.Any(), gomock.Any()).
					Return(false, nil)
				testHelper.mockCacheRepository.EXPECT().
					GetIncrement(gomock.Any(), "splitIdCounter").
					Return(int64(1), nil).MaxTimes(2)
				testHelper.mockAccRepository.EXPECT().
					GetOneByAccountNumber(gomock.Any(), gomock.Any()).
					Return(models.GetAccountOut{
						AccountNumber: "TEST",
						EntityCode:    "001",
					}, nil)
				testHelper.mockAccRepository.EXPECT().
					GetOneByAccountNumber(gomock.Any(), gomock.Any()).
					Return(models.GetAccountOut{
						AccountNumber: "TEST1",
						EntityCode:    "001",
					}, nil)
				testHelper.mockMySQLRepository.EXPECT().
					Atomic(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, steps func(ctx context.Context, r mysql.SQLRepository) error) error {
						testHelper.mockAcctRepository.EXPECT().InsertTransaction(gomock.Any(), gomock.Any()).Return(nil)
						testHelper.mockAcctRepository.EXPECT().InsertSplit(gomock.Any(), gomock.Any()).Return(assert.AnError)
						return steps(ctx, testHelper.mockMySQLRepository)
					})
				testHelper.mockPublisher.EXPECT().
					PublishSyncWithKeyAndLog(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					AnyTimes()
			},
			wantErr: true,
		},
		{
			name: "error case - failed insert split accounts",
			req:  req,
			doMock: func(ctx context.Context, req models.JournalRequest) {
				testHelper.mockAcctRepository.EXPECT().
					CheckTransactionIdIsExist(gomock.Any(), gomock.Any()).
					Return(false, nil)
				testHelper.mockCacheRepository.EXPECT().
					GetIncrement(gomock.Any(), "splitIdCounter").Return(int64(1), nil).
					MaxTimes(2)
				testHelper.mockAccRepository.EXPECT().
					GetOneByAccountNumber(gomock.Any(), gomock.Any()).
					Return(models.GetAccountOut{
						AccountNumber: "TEST",
						EntityCode:    "001",
					}, nil)
				testHelper.mockAccRepository.EXPECT().
					GetOneByAccountNumber(gomock.Any(), gomock.Any()).
					Return(models.GetAccountOut{
						AccountNumber: "TEST1",
						EntityCode:    "001",
					}, nil)
				testHelper.mockMySQLRepository.EXPECT().
					Atomic(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, steps func(ctx context.Context, r mysql.SQLRepository) error) error {
						testHelper.mockAcctRepository.EXPECT().InsertTransaction(gomock.Any(), gomock.Any()).Return(nil)
						testHelper.mockAcctRepository.EXPECT().InsertSplit(gomock.Any(), gomock.Any()).Return(nil)
						testHelper.mockAcctRepository.EXPECT().InsertSplitAccount(gomock.Any(), gomock.Any()).Return(assert.AnError)
						return steps(ctx, testHelper.mockMySQLRepository)
					})
				testHelper.mockPublisher.EXPECT().
					PublishSyncWithKeyAndLog(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					AnyTimes()
			},
			wantErr: true,
		},
		{
			name: "error case - failed insert journal detail",
			req:  req,
			doMock: func(ctx context.Context, req models.JournalRequest) {
				testHelper.mockAcctRepository.EXPECT().
					CheckTransactionIdIsExist(gomock.Any(), gomock.Any()).
					Return(false, nil)
				testHelper.mockCacheRepository.EXPECT().
					GetIncrement(gomock.Any(), "splitIdCounter").
					Return(int64(1), nil).MaxTimes(2)
				testHelper.mockAccRepository.EXPECT().
					GetOneByAccountNumber(gomock.Any(), gomock.Any()).
					Return(models.GetAccountOut{
						AccountNumber: "TEST",
						EntityCode:    "001",
					}, nil)
				testHelper.mockAccRepository.EXPECT().
					GetOneByAccountNumber(gomock.Any(), gomock.Any()).
					Return(models.GetAccountOut{
						AccountNumber: "TEST1",
						EntityCode:    "001",
					}, nil)
				testHelper.mockMySQLRepository.EXPECT().
					Atomic(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, steps func(ctx context.Context, r mysql.SQLRepository) error) error {
						testHelper.mockAcctRepository.EXPECT().InsertTransaction(gomock.Any(), gomock.Any()).Return(nil)
						testHelper.mockAcctRepository.EXPECT().InsertSplit(gomock.Any(), gomock.Any()).Return(nil)
						testHelper.mockAcctRepository.EXPECT().InsertSplitAccount(gomock.Any(), gomock.Any()).Return(nil)
						testHelper.mockAcctRepository.EXPECT().InsertJournalDetail(gomock.Any(), gomock.Any()).Return(assert.AnError)
						return steps(ctx, testHelper.mockMySQLRepository)
					})
				testHelper.mockPublisher.EXPECT().
					PublishSyncWithKeyAndLog(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					AnyTimes()
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if tt.doMock != nil {
				tt.doMock(ctx, tt.req)
			}
			err := testHelper.journalService.ConsumerInsertTransaction(ctx, tt.req)
			assert.Equal(t, tt.wantErr, err != nil)
		})
	}
}

func Test_journalService_PublishJournalTransaction(t *testing.T) {
	testHelper := serviceTestHelper(t)

	req := models.JournalRequest{
		TransactionId:   uuid.New().String(),
		TransactionDate: atime.DateFormatYYYYMMDDWithTime,
		ProcessingDate:  atime.DateFormatYYYYMMDDWithTime,
		Currency:        "IDR",
		Transactions: []models.Transaction{
			{
				Account:         "TEST",
				TransactionType: "TEST",
				Amount:          decimal.NewFromFloat(10000),
			},
		},
	}
	type args struct {
		ctx context.Context
		req models.JournalRequest
	}
	tests := []struct {
		name    string
		args    args
		doMock  func(args args)
		wantErr bool
	}{
		{
			name: "error case - account number not found",
			args: args{
				ctx: context.TODO(),
				req: req,
			},
			doMock: func(args args) {
				testHelper.mockAcctRepository.EXPECT().CheckTransactionIdIsExist(gomock.Any(), gomock.Any()).Return(false, nil)
				testHelper.mockAccRepository.EXPECT().CheckAccountNumberIsExist(args.ctx, gomock.Any()).Return(nil, nil)
			},
			wantErr: true,
		},
		{
			name: "success case",
			args: args{
				ctx: context.TODO(),
				req: req,
			},
			doMock: func(args args) {
				testHelper.mockAcctRepository.EXPECT().CheckTransactionIdIsExist(gomock.Any(), gomock.Any()).Return(false, nil)
				testHelper.mockAccRepository.EXPECT().CheckAccountNumberIsExist(args.ctx, gomock.Any()).AnyTimes().Return(&models.CheckAccountNumberIsExist{}, nil)
				testHelper.mockPublisher.EXPECT().PublishSyncWithKeyAndLog(args.ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any())
			},
			wantErr: false,
		},
		{
			name: "error case - transactios empty",
			args: args{
				ctx: context.TODO(),
				req: models.JournalRequest{
					TransactionId:   uuid.New().String(),
					TransactionDate: atime.DateFormatYYYYMMDD,
					ProcessingDate:  atime.DateFormatYYYYMMDD,
					Currency:        "IDR",
				},
			},
			wantErr: true,
		},
		{
			name: "error case - database error find transaction",
			args: args{
				ctx: context.TODO(),
				req: req,
			},
			doMock: func(args args) {
				testHelper.mockAcctRepository.EXPECT().CheckTransactionIdIsExist(gomock.Any(), gomock.Any()).Return(false, assert.AnError)
			},
			wantErr: true,
		},
		{
			name: "error case - transaction id is exist",
			args: args{
				ctx: context.TODO(),
				req: req,
			},
			doMock: func(args args) {
				testHelper.mockAcctRepository.EXPECT().CheckTransactionIdIsExist(gomock.Any(), gomock.Any()).Return(true, nil)
			},
			wantErr: true,
		},
		{
			name: "error case - invalid format transaction date",
			args: args{
				ctx: context.TODO(),
				req: models.JournalRequest{
					TransactionId:   uuid.New().String(),
					TransactionDate: atime.DateFormatYYYYMMDDWithTimeWithoutColon,
					ProcessingDate:  atime.DateFormatYYYYMMDDWithTime,
					Currency:        "IDR",
					Transactions: []models.Transaction{
						{
							Account:         "TEST",
							TransactionType: "TEST",
							Amount:          decimal.NewFromFloat(10000),
						},
					},
				},
			},
			doMock: func(args args) {
				testHelper.mockAcctRepository.EXPECT().CheckTransactionIdIsExist(gomock.Any(), gomock.Any()).Return(false, nil)
			},
			wantErr: true,
		},
		{
			name: "error case - invalid format processing date",
			args: args{
				ctx: context.TODO(),
				req: models.JournalRequest{
					TransactionId:   uuid.New().String(),
					TransactionDate: atime.DateFormatYYYYMMDDWithTime,
					ProcessingDate:  atime.DateFormatYYYYMMDDWithTimeWithoutColon,
					Currency:        "IDR",
					Transactions: []models.Transaction{
						{
							Account:         "TEST",
							TransactionType: "TEST",
							Amount:          decimal.NewFromFloat(10000),
						},
					},
				},
			},
			doMock: func(args args) {
				testHelper.mockAcctRepository.EXPECT().CheckTransactionIdIsExist(gomock.Any(), gomock.Any()).Return(false, nil)
			},
			wantErr: true,
		},
		{
			name: "error case - failed publish data",
			args: args{
				ctx: context.TODO(),
				req: req,
			},
			doMock: func(args args) {
				testHelper.mockAcctRepository.EXPECT().CheckTransactionIdIsExist(gomock.Any(), gomock.Any()).Return(false, nil)
				testHelper.mockAccRepository.EXPECT().CheckAccountNumberIsExist(args.ctx, gomock.Any()).AnyTimes().Return(&models.CheckAccountNumberIsExist{}, nil)
				testHelper.mockPublisher.EXPECT().PublishSyncWithKeyAndLog(args.ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(models.GetErrMap(models.ErrCodeInternalServerError))
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if tt.doMock != nil {
				tt.doMock(tt.args)
			}
			err := testHelper.journalService.PublishJournalTransaction(tt.args.ctx, tt.args.req)
			assert.Equal(t, tt.wantErr, err != nil)
		})
	}
}

func Test_journalService_ProcessUploadJournal(t *testing.T) {
	testHelper := serviceTestHelper(t)

	f := createMultipartFormData(t)
	o := mustOpen("../../storages/upload_journals.csv")
	csvResp := [][]string{
		{"reference_number", "transaction_id", "order_type", "transaction_date", "processing_date", "currency", "transaction_type1", "transaction_type_name1", "account_debit1", "account_credit1", "narrative1", "amount1", "transaction_type2", "transaction_type_name2", "account_debit2", "account_credit2", "narrative2", "amount2"},
		{"001-TRX-MANUAL-23042024-1", "b13583d7-adba-4aba-9261-3dee86867ec7s", "TUP", "2024-04-23 09:20:00", "2024-04-23 09:20:00", "IDR", "TUPIN", "Amartha Top Up represents Institutional Lender", "142001000000001", "211001000381110", "Amartha Top Up represents Institutional Lender", "3368495938", "TUPLF", "Amartha Top Up represents Institutional Lender", "142001000000001", "211001000381110", "Amartha Top Up represents Institutional Lender", "5000", "{\"test\":\"123\"}"},
	}

	type args struct {
		ctx  context.Context
		file *multipart.FileHeader
	}
	tests := []struct {
		name    string
		args    args
		doMock  func(args args)
		wantErr bool
	}{
		{
			name: "success case",
			args: args{
				ctx:  context.TODO(),
				file: f,
			},
			doMock: func(args args) {
				testHelper.mockFile.EXPECT().CreateFile(gomock.Any()).Return(o, nil)
				testHelper.mockFile.EXPECT().CopyFile(gomock.Any(), gomock.Any()).Return(nil)
				testHelper.mockFile.EXPECT().RemoveFile(gomock.Any()).Return(nil)
				testHelper.mockFile.EXPECT().OpenFile(gomock.Any()).Return(o, nil)
				testHelper.mockFile.EXPECT().CSVReadAll(gomock.Any()).Return(csvResp, nil)

				testHelper.mockAcctRepository.EXPECT().CheckTransactionIdIsExist(gomock.Any(), gomock.Any()).Return(false, nil)
				testHelper.mockPublisher.EXPECT().PublishSyncWithKeyAndLog(args.ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any())
			},
			wantErr: false,
		},
		{
			name: "error case - no such file or directory",
			args: args{
				ctx: context.TODO(),
				file: &multipart.FileHeader{
					Filename: "../../storages/upload_accounts.csv",
				},
			},
			wantErr: true,
		},
		{
			name: "error case - create file",
			args: args{
				ctx:  context.TODO(),
				file: f,
			},
			doMock: func(args args) {
				testHelper.mockFile.EXPECT().CreateFile(gomock.Any()).Return(o, assert.AnError)
			},
			wantErr: true,
		},
		{
			name: "error case - copy file",
			args: args{
				ctx:  context.TODO(),
				file: f,
			},
			doMock: func(args args) {
				testHelper.mockFile.EXPECT().CreateFile(gomock.Any()).Return(o, nil)
				testHelper.mockFile.EXPECT().CopyFile(gomock.Any(), gomock.Any()).Return(assert.AnError)
			},
			wantErr: true,
		},
		{
			name: "error case - open file",
			args: args{
				ctx:  context.TODO(),
				file: f,
			},
			doMock: func(args args) {
				testHelper.mockFile.EXPECT().CreateFile(gomock.Any()).Return(o, nil)
				testHelper.mockFile.EXPECT().CopyFile(gomock.Any(), gomock.Any()).Return(nil)
				testHelper.mockFile.EXPECT().RemoveFile(gomock.Any()).Return(nil)
				testHelper.mockFile.EXPECT().OpenFile(gomock.Any()).Return(o, assert.AnError)
			},
			wantErr: true,
		},
		{
			name: "error case - csv read",
			args: args{
				ctx:  context.TODO(),
				file: f,
			},
			doMock: func(args args) {
				testHelper.mockFile.EXPECT().CreateFile(gomock.Any()).Return(o, nil)
				testHelper.mockFile.EXPECT().CopyFile(gomock.Any(), gomock.Any()).Return(nil)
				testHelper.mockFile.EXPECT().RemoveFile(gomock.Any()).Return(nil)
				testHelper.mockFile.EXPECT().OpenFile(gomock.Any()).Return(o, nil)
				testHelper.mockFile.EXPECT().CSVReadAll(gomock.Any()).Return([][]string{}, assert.AnError)
			},
			wantErr: true,
		},
		{
			name: "error case - invalid format amount1",
			args: args{
				ctx:  context.TODO(),
				file: f,
			},
			doMock: func(args args) {
				testHelper.mockFile.EXPECT().CreateFile(gomock.Any()).Return(o, nil)
				testHelper.mockFile.EXPECT().CopyFile(gomock.Any(), gomock.Any()).Return(nil)
				testHelper.mockFile.EXPECT().RemoveFile(gomock.Any()).Return(nil)
				testHelper.mockFile.EXPECT().OpenFile(gomock.Any()).Return(o, nil)
				testHelper.mockFile.EXPECT().CSVReadAll(gomock.Any()).Return([][]string{
					{"reference_number", "transaction_id", "order_type", "transaction_date", "processing_date", "currency", "transaction_type1", "transaction_type_name1", "account_debit1", "account_credit1", "narrative1", "amount1", "transaction_type2", "transaction_type_name2", "account_debit2", "account_credit2", "narrative2", "amount2", "metadata"},
					{"001-TRX-MANUAL-23042024-1", "b13583d7-adba-4aba-9261-3dee86867ec7s", "TUP", "2024-04-23 09:20:00", "2024-04-23 09:20:00", "IDR", "TUPIN", "Amartha Top Up represents Institutional Lender", "142001000000001", "211001000381110", "Amartha Top Up represents Institutional Lender", "3368495938A", "TUPLF", "Amartha Top Up represents Institutional Lender", "142001000000001", "211001000381110", "Amartha Top Up represents Institutional Lender", "5000", "{\"test\":\"123\"}"},
				}, nil)
			},
			wantErr: true,
		},
		{
			name: "error case - invalid format amount2",
			args: args{
				ctx:  context.TODO(),
				file: f,
			},
			doMock: func(args args) {
				testHelper.mockFile.EXPECT().CreateFile(gomock.Any()).Return(o, nil)
				testHelper.mockFile.EXPECT().CopyFile(gomock.Any(), gomock.Any()).Return(nil)
				testHelper.mockFile.EXPECT().RemoveFile(gomock.Any()).Return(nil)
				testHelper.mockFile.EXPECT().OpenFile(gomock.Any()).Return(o, nil)
				testHelper.mockFile.EXPECT().CSVReadAll(gomock.Any()).Return([][]string{
					{"reference_number", "transaction_id", "order_type", "transaction_date", "processing_date", "currency", "transaction_type1", "transaction_type_name1", "account_debit1", "account_credit1", "narrative1", "amount1", "transaction_type2", "transaction_type_name2", "account_debit2", "account_credit2", "narrative2", "amount2", "metadata"},
					{"001-TRX-MANUAL-23042024-1", "b13583d7-adba-4aba-9261-3dee86867ec7s", "TUP", "2024-04-23 09:20:00", "2024-04-23 09:20:00", "IDR", "TUPIN", "Amartha Top Up represents Institutional Lender", "142001000000001", "211001000381110", "Amartha Top Up represents Institutional Lender", "3368495938", "TUPLF", "Amartha Top Up represents Institutional Lender", "142001000000001", "211001000381110", "Amartha Top Up represents Institutional Lender", "5000A", "{\"test\":\"123\"}"},
				}, nil)
			},
			wantErr: true,
		},
		{
			name: "error case - transaction id is exist",
			args: args{
				ctx:  context.TODO(),
				file: f,
			},
			doMock: func(args args) {
				testHelper.mockFile.EXPECT().CreateFile(gomock.Any()).Return(o, nil)
				testHelper.mockFile.EXPECT().CopyFile(gomock.Any(), gomock.Any()).Return(nil)
				testHelper.mockFile.EXPECT().RemoveFile(gomock.Any()).Return(nil)
				testHelper.mockFile.EXPECT().OpenFile(gomock.Any()).Return(o, nil)
				testHelper.mockFile.EXPECT().CSVReadAll(gomock.Any()).Return(csvResp, nil)

				testHelper.mockAcctRepository.EXPECT().CheckTransactionIdIsExist(gomock.Any(), gomock.Any()).Return(true, nil)
			},
			wantErr: true,
		},
		{
			name: "error case - database error find transaction",
			args: args{
				ctx:  context.TODO(),
				file: f,
			},
			doMock: func(args args) {
				testHelper.mockFile.EXPECT().CreateFile(gomock.Any()).Return(o, nil)
				testHelper.mockFile.EXPECT().CopyFile(gomock.Any(), gomock.Any()).Return(nil)
				testHelper.mockFile.EXPECT().RemoveFile(gomock.Any()).Return(nil)
				testHelper.mockFile.EXPECT().OpenFile(gomock.Any()).Return(o, nil)
				testHelper.mockFile.EXPECT().CSVReadAll(gomock.Any()).Return(csvResp, nil)

				testHelper.mockAcctRepository.EXPECT().CheckTransactionIdIsExist(gomock.Any(), gomock.Any()).Return(true, assert.AnError)
			},
			wantErr: true,
		},
		{
			name: "error case - invalid format transaction date",
			args: args{
				ctx:  context.TODO(),
				file: f,
			},
			doMock: func(args args) {
				testHelper.mockFile.EXPECT().CreateFile(gomock.Any()).Return(o, nil)
				testHelper.mockFile.EXPECT().CopyFile(gomock.Any(), gomock.Any()).Return(nil)
				testHelper.mockFile.EXPECT().RemoveFile(gomock.Any()).Return(nil)
				testHelper.mockFile.EXPECT().OpenFile(gomock.Any()).Return(o, nil)
				testHelper.mockFile.EXPECT().CSVReadAll(gomock.Any()).Return([][]string{
					{"reference_number", "transaction_id", "order_type", "transaction_date", "processing_date", "currency", "transaction_type1", "transaction_type_name1", "account_debit1", "account_credit1", "narrative1", "amount1", "transaction_type2", "transaction_type_name2", "account_debit2", "account_credit2", "narrative2", "amount2", "metadata"},
					{"001-TRX-MANUAL-23042024-1", "b13583d7-adba-4aba-9261-3dee86867ec7s", "TUP", "2024-04-23", "2024-04-23 09:20:00", "IDR", "TUPIN", "Amartha Top Up represents Institutional Lender", "142001000000001", "211001000381110", "Amartha Top Up represents Institutional Lender", "3368495938", "", "", "", "", "", "", ""},
				}, nil)

				testHelper.mockAcctRepository.EXPECT().CheckTransactionIdIsExist(gomock.Any(), gomock.Any()).Return(true, nil)
			},
			wantErr: true,
		},
		{
			name: "error case - invalid format processing date",
			args: args{
				ctx:  context.TODO(),
				file: f,
			},
			doMock: func(args args) {
				testHelper.mockFile.EXPECT().CreateFile(gomock.Any()).Return(o, nil)
				testHelper.mockFile.EXPECT().CopyFile(gomock.Any(), gomock.Any()).Return(nil)
				testHelper.mockFile.EXPECT().RemoveFile(gomock.Any()).Return(nil)
				testHelper.mockFile.EXPECT().OpenFile(gomock.Any()).Return(o, nil)
				testHelper.mockFile.EXPECT().CSVReadAll(gomock.Any()).Return([][]string{
					{"reference_number", "transaction_id", "order_type", "transaction_date", "processing_date", "currency", "transaction_type1", "transaction_type_name1", "account_debit1", "account_credit1", "narrative1", "amount1", "transaction_type2", "transaction_type_name2", "account_debit2", "account_credit2", "narrative2", "amount2", "metadata"},
					{"001-TRX-MANUAL-23042024-1", "b13583d7-adba-4aba-9261-3dee86867ec7s", "TUP", "2024-04-23 09:20:00", "2024-04-23", "IDR", "TUPIN", "Amartha Top Up represents Institutional Lender", "142001000000001", "211001000381110", "Amartha Top Up represents Institutional Lender", "3368495938", "", "", "", "", "", "", ""},
				}, nil)

				testHelper.mockAcctRepository.EXPECT().CheckTransactionIdIsExist(gomock.Any(), gomock.Any()).Return(true, nil)
			},
			wantErr: true,
		},
		{
			name: "error case - failed publish data",
			args: args{
				ctx:  context.TODO(),
				file: f,
			},
			doMock: func(args args) {
				testHelper.mockFile.EXPECT().CreateFile(gomock.Any()).Return(o, nil)
				testHelper.mockFile.EXPECT().CopyFile(gomock.Any(), gomock.Any()).Return(nil)
				testHelper.mockFile.EXPECT().RemoveFile(gomock.Any()).Return(nil)
				testHelper.mockFile.EXPECT().OpenFile(gomock.Any()).Return(o, nil)
				testHelper.mockFile.EXPECT().CSVReadAll(gomock.Any()).Return(csvResp, nil)

				testHelper.mockAcctRepository.EXPECT().CheckTransactionIdIsExist(gomock.Any(), gomock.Any()).Return(false, nil)
				testHelper.mockPublisher.EXPECT().PublishSyncWithKeyAndLog(args.ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(models.GetErrMap(models.ErrCodeInternalServerError))
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if tt.doMock != nil {
				tt.doMock(tt.args)
			}
			err := testHelper.journalService.ProcessUploadJournal(tt.args.ctx, tt.args.file)
			assert.Equal(t, tt.wantErr, err != nil)
		})
	}
}

func Test_journalService_GetJournalByTransactionId(t *testing.T) {
	testHelper := serviceTestHelper(t)

	now, _ := atime.NowZeroTime()
	result := []models.GetJournalDetailOut{
		{
			TransactionId:   "12c5692e-cfbd-4cee-a4ce-86eac1447d48",
			JournalId:       "2024060601093678",
			AccountNumber:   "214003000000194",
			AccountName:     "Mails Morales",
			AltId:           "Mails002",
			EntityCode:      "003",
			EntityName:      "AFA",
			SubCategoryCode: "21401",
			SubCategoryName: "eWallet User",
			Amount:          decimal.NewFromFloat(10000),
			Narrative:       "Repayment Group Loan via Poket jindankarasuno",
			TransactionDate: now,
			IsDebit:         true,
		},
		{

			TransactionId:   "12c5692e-cfbd-4cee-a4ce-86eac1447d48",
			JournalId:       "2024060601093679",
			AccountNumber:   "261003000000005",
			AccountName:     "IA Ewallet Repayment Group Loan",
			AltId:           "",
			EntityCode:      "003",
			EntityName:      "AFA",
			SubCategoryCode: "26101",
			SubCategoryName: "Related Party Payable",
			Amount:          decimal.NewFromFloat(10000),
			TransactionDate: now,
			Narrative:       "Repayment Group Loan via Poket jindankarasuno",
			IsDebit:         false,
		},
		{
			TransactionId:   "12c5692e-cfbd-4cee-a4ce-86eac1447d48",
			JournalId:       "2024060601093680",
			AccountNumber:   "143001000000168",
			AccountName:     "IA Ewallet Repayment Group Loan",
			AltId:           "",
			EntityCode:      "001",
			EntityName:      "AMF",
			SubCategoryCode: "14301",
			SubCategoryName: "Related Parties Receivable",
			Amount:          decimal.NewFromFloat(10000),
			TransactionDate: now,
			Narrative:       "Repayment Group Loan via Poket jindankarasuno",
			IsDebit:         true,
		},
		{
			TransactionId:   "12c5692e-cfbd-4cee-a4ce-86eac1447d48",
			JournalId:       "2024060601093681",
			AccountNumber:   "121001000000008",
			AccountName:     "Cash in Transit - Repayment Group Loan",
			AltId:           "",
			EntityCode:      "001",
			EntityName:      "AMF",
			SubCategoryCode: "12101",
			SubCategoryName: "Cash in Transit",
			Amount:          decimal.NewFromFloat(10000),
			TransactionDate: now,
			Narrative:       "Repayment Group Loan via Poket jindankarasuno",
			IsDebit:         false,
		},
	}
	type args struct {
		ctx context.Context
		req string
	}
	type mockData struct {
	}
	tests := []struct {
		name     string
		args     args
		mockData mockData
		doMock   func(args args, mockData mockData)
		wantOut  []models.GetJournalDetailOut
		wantErr  bool
	}{
		{
			name: "success case",
			args: args{
				ctx: context.Background(),
				req: "12c5692e-cfbd-4cee-a4ce-86eac1447d48",
			},
			mockData: mockData{},
			doMock: func(args args, mockData mockData) {
				testHelper.mockAcctRepository.EXPECT().CheckTransactionIdIsExist(gomock.Any(), gomock.Any()).Return(true, nil)
				testHelper.mockAcctRepository.EXPECT().GetJournalDetailByTransactionId(args.ctx, args.req).Return(result, nil)
			},
			wantErr: false,
			wantOut: result,
		},
		{
			name: "error case - transaction id not found",
			args: args{
				ctx: context.Background(),
				req: "12c5692e-cfbd-4cee-a4ce-86eac1447d48",
			},
			mockData: mockData{},
			doMock: func(args args, mockData mockData) {
				testHelper.mockAcctRepository.EXPECT().CheckTransactionIdIsExist(gomock.Any(), gomock.Any()).Return(false, nil)
			},
			wantErr: true,
		},
		{
			name: "error case - database error",
			args: args{
				ctx: context.Background(),
				req: "12c5692e-cfbd-4cee-a4ce-86eac1447d48",
			},
			mockData: mockData{},
			doMock: func(args args, mockData mockData) {
				testHelper.mockAcctRepository.EXPECT().CheckTransactionIdIsExist(gomock.Any(), gomock.Any()).Return(false, nil)
				testHelper.mockAcctRepository.EXPECT().GetJournalDetailByTransactionId(args.ctx, args.req).Return(nil, models.GetErrMap(models.ErrKeyDatabaseError))
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if tt.doMock != nil {
				tt.doMock(tt.args, tt.mockData)
			}
			res, err := testHelper.journalService.GetJournalByTransactionId(tt.args.ctx, tt.args.req)
			assert.Equal(t, tt.wantErr, err != nil)
			assert.Equal(t, tt.wantOut, res)
		})
	}
}

func Test_journalService_RetryPublishToJournalEntryCreated(t *testing.T) {
	testHelper := serviceTestHelper(t)

	journalEntry := models.JournalEntryCreatedRequest{
		TransactionID: "12c5692e-cfbd-4cee-a4ce-86eac1447d48",
		JournalID:     "2024060601093678",
	}

	type args struct {
		ctx context.Context
		req models.JournalEntryCreatedRequest
	}
	tests := []struct {
		name    string
		args    args
		doMock  func(args args)
		wantErr bool
	}{
		{
			name: "success case",
			args: args{
				ctx: context.Background(),
				req: journalEntry,
			},
			doMock: func(args args) {
				testHelper.mockPublisher.EXPECT().PublishSyncWithKeyAndLog(args.ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if tt.doMock != nil {
				tt.doMock(tt.args)
			}
			err := testHelper.journalService.RetryPublishToJournalEntryCreated(tt.args.ctx, tt.args.req)
			assert.Equal(t, tt.wantErr, err != nil)
		})
	}
}
