package services_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"bitbucket.org/Amartha/go-accounting/internal/models"
	"bitbucket.org/Amartha/go-accounting/internal/pkg/atime"
	"bitbucket.org/Amartha/go-accounting/internal/pkg/gofptransaction"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func Test_accounting_GetSubLedgerAccounts(t *testing.T) {
	testHelper := serviceTestHelper(t)

	type args struct {
		ctx context.Context
		req models.SubLedgerAccountsFilterOptions
	}
	type mockData struct {
	}
	tests := []struct {
		name     string
		args     args
		mockData mockData
		doMock   func(args args, mockData mockData)
		wantErr  bool
	}{
		{
			name: "success case - get sub ledger accounts no filter",
			args: args{
				ctx: context.Background(),
				req: models.SubLedgerAccountsFilterOptions{},
			},
			doMock: func(args args, mockData mockData) {
				testHelper.mockFlag.EXPECT().IsEnabled(models.FlagGuestModePayment.String()).Return(false)
				testHelper.mockAcctRepository.EXPECT().GetSubLedgerAccounts(args.ctx, args.req).Return([]models.GetSubLedgerAccountsOut{
					{
						AccountNumber:   "121001000000009",
						AccountName:     "Cash in Transit - Disburse Modal",
						AltId:           "",
						SubCategoryCode: "12101",
						SubCategoryName: "Cash in Transit",
						TotalRowData:    0,
					},
				}, nil)
				testHelper.mockCacheRepository.EXPECT().Get(args.ctx, gomock.Any()).Return("", assert.AnError)
				testHelper.mockAcctRepository.EXPECT().GetSubLedgerAccountsCount(args.ctx, args.req).Return(10, nil)
				testHelper.mockCacheRepository.EXPECT().Set(args.ctx, gomock.Any(), 10, gomock.Any()).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "success case - get sub ledger accounts no filter - cache",
			args: args{
				ctx: context.Background(),
				req: models.SubLedgerAccountsFilterOptions{},
			},
			doMock: func(args args, mockData mockData) {
				testHelper.mockFlag.EXPECT().IsEnabled(models.FlagGuestModePayment.String()).Return(false)
				testHelper.mockAcctRepository.EXPECT().GetSubLedgerAccounts(args.ctx, args.req).Return([]models.GetSubLedgerAccountsOut{
					{
						AccountNumber:   "121001000000009",
						AccountName:     "Cash in Transit - Disburse Modal",
						AltId:           "",
						SubCategoryCode: "12101",
						SubCategoryName: "Cash in Transit",
						TotalRowData:    0,
					},
				}, nil)
				testHelper.mockCacheRepository.EXPECT().Get(args.ctx, gomock.Any()).Return("10", nil)
			},
			wantErr: false,
		},
		{
			name: "success case - get sub ledger accounts with filter",
			args: args{
				ctx: context.Background(),
				req: models.SubLedgerAccountsFilterOptions{
					EntityCode: "001",
					Search:     "121001000000009",
					StartDate:  atime.Now(),
					EndDate:    atime.Now(),
				},
			},
			doMock: func(args args, mockData mockData) {
				testHelper.mockFlag.EXPECT().IsEnabled(models.FlagGuestModePayment.String()).Return(false)
				testHelper.mockAcctRepository.EXPECT().GetSubLedgerAccounts(args.ctx, args.req).Return([]models.GetSubLedgerAccountsOut{
					{
						AccountNumber:   "121001000000009",
						AccountName:     "Cash in Transit - Disburse Modal",
						AltId:           "",
						SubCategoryCode: "12101",
						SubCategoryName: "Cash in Transit",
						TotalRowData:    0,
					},
				}, nil)
			},
			wantErr: false,
		},
		{
			name: "error case - data not found get sub ledger accounts",
			args: args{
				ctx: context.Background(),
				req: models.SubLedgerAccountsFilterOptions{},
			},
			doMock: func(args args, mockData mockData) {
				testHelper.mockFlag.EXPECT().IsEnabled(models.FlagGuestModePayment.String()).Return(false)
				testHelper.mockAcctRepository.EXPECT().GetSubLedgerAccounts(args.ctx, args.req).Return([]models.GetSubLedgerAccountsOut{}, nil)
			},
			wantErr: true,
		},
		{
			name: "error case - database error get sub ledger accounts",
			args: args{
				ctx: context.Background(),
				req: models.SubLedgerAccountsFilterOptions{},
			},
			doMock: func(args args, mockData mockData) {
				testHelper.mockFlag.EXPECT().IsEnabled(models.FlagGuestModePayment.String()).Return(false)
				testHelper.mockAcctRepository.EXPECT().GetSubLedgerAccounts(args.ctx, args.req).Return([]models.GetSubLedgerAccountsOut{}, models.GetErrMap(models.ErrKeyDatabaseError))
			},
			wantErr: true,
		},
		{
			name: "error case - database error count get sub ledger accounts",
			args: args{
				ctx: context.Background(),
				req: models.SubLedgerAccountsFilterOptions{},
			},
			doMock: func(args args, mockData mockData) {
				testHelper.mockFlag.EXPECT().IsEnabled(models.FlagGuestModePayment.String()).Return(false)
				testHelper.mockAcctRepository.EXPECT().GetSubLedgerAccounts(args.ctx, args.req).Return([]models.GetSubLedgerAccountsOut{
					{
						AccountNumber:   "121001000000009",
						AccountName:     "Cash in Transit - Disburse Modal",
						AltId:           "",
						SubCategoryCode: "12101",
						SubCategoryName: "Cash in Transit",
						TotalRowData:    0,
					},
				}, nil)
				testHelper.mockCacheRepository.EXPECT().Get(args.ctx, gomock.Any()).Return("", assert.AnError)
				testHelper.mockAcctRepository.EXPECT().GetSubLedgerAccountsCount(args.ctx, args.req).Return(1, models.GetErrMap(models.ErrKeyDatabaseError))
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
			_, _, err := testHelper.accountingService.GetSubLedgerAccounts(tt.args.ctx, tt.args.req)
			assert.Equal(t, tt.wantErr, err != nil)
		})
	}
}

func Test_accounting_GetSubLedger(t *testing.T) {
	testHelper := serviceTestHelper(t)

	type args struct {
		ctx context.Context
		req models.SubLedgerFilterOptions
	}
	type mockData struct {
	}
	tests := []struct {
		name     string
		args     args
		mockData mockData
		doMock   func(args args, mockData mockData)
		wantErr  bool
	}{
		{
			name: "success case - get sub ledger",
			args: args{
				ctx: context.Background(),
				req: models.SubLedgerFilterOptions{},
			},
			doMock: func(args args, mockData mockData) {
				testHelper.mockAccRepository.EXPECT().GetOneByAccountNumber(args.ctx, args.req.AccountNumber).Return(models.GetAccountOut{}, nil)
				testHelper.mockAcctRepository.EXPECT().GetSubLedgerCount(args.ctx, gomock.Any()).Return(1, nil)
				testHelper.mockAcctRepository.EXPECT().GetSubLedger(args.ctx, args.req).Return([]models.GetSubLedgerOut{}, nil)
				testHelper.mockFlag.EXPECT().IsEnabled(models.FlagGetOpeningBalanceV2.String()).Return(false)
				testHelper.mockAcctRepository.EXPECT().GetAccountBalancePeriodStart(args.ctx, args.req.AccountNumber, args.req.StartDate).Return(decimal.New(0, 0), nil)
			},
			wantErr: false,
		},
		{
			name: "error case - account number not found",
			args: args{
				ctx: context.Background(),
				req: models.SubLedgerFilterOptions{},
			},
			doMock: func(args args, mockData mockData) {
				testHelper.mockAccRepository.EXPECT().GetOneByAccountNumber(args.ctx, args.req.AccountNumber).Return(models.GetAccountOut{}, models.ErrNoRows)
			},
			wantErr: true,
		},
		{
			name: "error case - data sub ledger is exceeds the limit",
			args: args{
				ctx: context.Background(),
				req: models.SubLedgerFilterOptions{},
			},
			doMock: func(args args, mockData mockData) {
				testHelper.mockAccRepository.EXPECT().GetOneByAccountNumber(args.ctx, args.req.AccountNumber).Return(models.GetAccountOut{}, nil)
				testHelper.mockAcctRepository.EXPECT().GetSubLedgerCount(args.ctx, gomock.Any()).Return(6, nil)
			},
			wantErr: true,
		},
		{
			name: "error case - database error count get sub ledger",
			args: args{
				ctx: context.Background(),
				req: models.SubLedgerFilterOptions{},
			},
			doMock: func(args args, mockData mockData) {
				testHelper.mockAccRepository.EXPECT().GetOneByAccountNumber(args.ctx, args.req.AccountNumber).Return(models.GetAccountOut{}, nil)
				testHelper.mockAcctRepository.EXPECT().GetSubLedgerCount(args.ctx, gomock.Any()).Return(0, models.GetErrMap(models.ErrKeyDatabaseError))
			},
			wantErr: true,
		},

		{
			name: "error case - database error get sub ledger",
			args: args{
				ctx: context.Background(),
				req: models.SubLedgerFilterOptions{},
			},
			doMock: func(args args, mockData mockData) {
				testHelper.mockAccRepository.EXPECT().GetOneByAccountNumber(args.ctx, args.req.AccountNumber).Return(models.GetAccountOut{}, nil)
				testHelper.mockAcctRepository.EXPECT().GetSubLedgerCount(args.ctx, gomock.Any()).Return(1, nil)
				testHelper.mockAcctRepository.EXPECT().GetSubLedger(args.ctx, args.req).Return([]models.GetSubLedgerOut{}, models.GetErrMap(models.ErrKeyDatabaseError))
			},
			wantErr: true,
		},
		{
			name: "error case - database error get account balance period start",
			args: args{
				ctx: context.Background(),
				req: models.SubLedgerFilterOptions{},
			},
			doMock: func(args args, mockData mockData) {
				testHelper.mockAccRepository.EXPECT().GetOneByAccountNumber(args.ctx, args.req.AccountNumber).Return(models.GetAccountOut{}, nil)
				testHelper.mockAcctRepository.EXPECT().GetSubLedgerCount(args.ctx, gomock.Any()).Return(1, nil)
				testHelper.mockAcctRepository.EXPECT().GetSubLedger(args.ctx, args.req).Return([]models.GetSubLedgerOut{}, nil)
				testHelper.mockFlag.EXPECT().IsEnabled(models.FlagGetOpeningBalanceV2.String()).Return(false)
				testHelper.mockAcctRepository.EXPECT().GetAccountBalancePeriodStart(args.ctx, args.req.AccountNumber, args.req.StartDate).Return(decimal.NewFromFloat(0), models.GetErrMap(models.ErrKeyDatabaseError))
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
			_, _, _, err := testHelper.accountingService.GetSubLedger(tt.args.ctx, tt.args.req)
			assert.Equal(t, tt.wantErr, err != nil)
		})
	}
}

func Test_accounting_GetSubLedgerCount(t *testing.T) {
	testHelper := serviceTestHelper(t)

	type args struct {
		ctx context.Context
		req models.SubLedgerFilterOptions
	}
	type mockData struct {
	}
	tests := []struct {
		name     string
		args     args
		mockData mockData
		doMock   func(args args, mockData mockData)
		wantErr  bool
	}{
		{
			name: "success case - get sub ledger count",
			args: args{
				ctx: context.Background(),
				req: models.SubLedgerFilterOptions{},
			},
			doMock: func(args args, mockData mockData) {
				testHelper.mockAcctRepository.EXPECT().GetSubLedgerCount(args.ctx, gomock.Any()).Return(1, nil)
			},
			wantErr: false,
		},
		{
			name: "error case - get sub ledger count",
			args: args{
				ctx: context.Background(),
				req: models.SubLedgerFilterOptions{},
			},
			doMock: func(args args, mockData mockData) {
				testHelper.mockAcctRepository.EXPECT().GetSubLedgerCount(args.ctx, gomock.Any()).Return(1, models.GetErrMap(models.ErrKeyDatabaseError))
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
			_, err := testHelper.accountingService.GetSubLedgerCount(tt.args.ctx, tt.args.req)
			assert.Equal(t, tt.wantErr, err != nil)
		})
	}
}

func Test_accounting_DownloadSubLedgerCSV(t *testing.T) {
	testHelper := serviceTestHelper(t)

	type args struct {
		ctx context.Context
		req models.SubLedgerFilterOptions
	}
	type mockData struct {
	}
	tests := []struct {
		name     string
		args     args
		mockData mockData
		doMock   func(args args, mockData mockData)
		wantErr  bool
	}{
		{
			name: "error case - account number not found",
			args: args{
				ctx: context.Background(),
				req: models.SubLedgerFilterOptions{},
			},
			doMock: func(args args, mockData mockData) {
				testHelper.mockAccRepository.EXPECT().GetOneByAccountNumber(args.ctx, args.req.AccountNumber).Return(models.GetAccountOut{}, models.ErrNoRows)
			},
			wantErr: true,
		},
		{
			name: "error case - data sub ledger is exceeds the limit",
			args: args{
				ctx: context.Background(),
				req: models.SubLedgerFilterOptions{},
			},
			doMock: func(args args, mockData mockData) {
				testHelper.mockAccRepository.EXPECT().GetOneByAccountNumber(args.ctx, args.req.AccountNumber).Return(models.GetAccountOut{}, nil)
				testHelper.mockAcctRepository.EXPECT().GetSubLedgerCount(args.ctx, gomock.Any()).Return(6, nil)
			},
			wantErr: true,
		},
		{
			name: "error case - database error count get sub ledger",
			args: args{
				ctx: context.Background(),
				req: models.SubLedgerFilterOptions{},
			},
			doMock: func(args args, mockData mockData) {
				testHelper.mockAccRepository.EXPECT().GetOneByAccountNumber(args.ctx, args.req.AccountNumber).Return(models.GetAccountOut{}, nil)
				testHelper.mockAcctRepository.EXPECT().GetSubLedgerCount(args.ctx, gomock.Any()).Return(0, models.GetErrMap(models.ErrKeyDatabaseError))
			},
			wantErr: true,
		},
		{
			name: "error case - database error get sub ledger",
			args: args{
				ctx: context.Background(),
				req: models.SubLedgerFilterOptions{},
			},
			doMock: func(args args, mockData mockData) {
				testHelper.mockAccRepository.EXPECT().GetOneByAccountNumber(args.ctx, args.req.AccountNumber).Return(models.GetAccountOut{}, nil)
				testHelper.mockAcctRepository.EXPECT().GetSubLedgerCount(args.ctx, gomock.Any()).Return(1, nil)
				testHelper.mockAcctRepository.EXPECT().GetSubLedger(args.ctx, args.req).Return([]models.GetSubLedgerOut{}, models.GetErrMap(models.ErrKeyDatabaseError))
			},
			wantErr: true,
		},
		{
			name: "error case - database error get account balance period start",
			args: args{
				ctx: context.Background(),
				req: models.SubLedgerFilterOptions{},
			},
			doMock: func(args args, mockData mockData) {
				testHelper.mockAccRepository.EXPECT().GetOneByAccountNumber(args.ctx, args.req.AccountNumber).Return(models.GetAccountOut{}, nil)
				testHelper.mockAcctRepository.EXPECT().GetSubLedgerCount(args.ctx, gomock.Any()).Return(1, nil)
				testHelper.mockAcctRepository.EXPECT().GetSubLedger(args.ctx, args.req).Return([]models.GetSubLedgerOut{}, nil)
				testHelper.mockFlag.EXPECT().IsEnabled(models.FlagGetOpeningBalanceV2.String()).Return(false)
				testHelper.mockAcctRepository.EXPECT().GetAccountBalancePeriodStart(args.ctx, args.req.AccountNumber, args.req.StartDate).Return(decimal.NewFromFloat(0), models.GetErrMap(models.ErrKeyDatabaseError))
			},
			wantErr: true,
		},
		{
			name: "error case - error get all order types",
			args: args{
				ctx: context.Background(),
				req: models.SubLedgerFilterOptions{},
			},
			doMock: func(args args, mockData mockData) {
				testHelper.mockAccRepository.EXPECT().GetOneByAccountNumber(args.ctx, args.req.AccountNumber).Return(models.GetAccountOut{}, nil)
				testHelper.mockAcctRepository.EXPECT().GetSubLedgerCount(args.ctx, gomock.Any()).Return(1, nil)
				testHelper.mockAcctRepository.EXPECT().GetSubLedger(args.ctx, args.req).Return([]models.GetSubLedgerOut{
					{}, {}, {}, {}, {}, {},
				}, nil)
				testHelper.mockFlag.EXPECT().IsEnabled(models.FlagGetOpeningBalanceV2.String()).Return(false)
				testHelper.mockAcctRepository.EXPECT().GetAccountBalancePeriodStart(args.ctx, args.req.AccountNumber, args.req.StartDate).Return(decimal.NewFromFloat(0), nil)
				testHelper.mockGoFpTransaction.EXPECT().GetAllOrderTypes(args.ctx, "", "").Return(gofptransaction.GetAllOrderTypesOut{}, assert.AnError)
			},
			wantErr: true,
		},
		{
			name: "error case - create csv write all",
			args: args{
				ctx: context.Background(),
				req: models.SubLedgerFilterOptions{},
			},
			doMock: func(args args, mockData mockData) {
				testHelper.mockAccRepository.EXPECT().GetOneByAccountNumber(args.ctx, args.req.AccountNumber).Return(models.GetAccountOut{}, nil)
				testHelper.mockAcctRepository.EXPECT().GetSubLedgerCount(args.ctx, gomock.Any()).Return(1, nil)
				testHelper.mockAcctRepository.EXPECT().GetSubLedger(args.ctx, args.req).Return([]models.GetSubLedgerOut{
					{}, {}, {}, {}, {}, {},
				}, nil)
				testHelper.mockFlag.EXPECT().IsEnabled(models.FlagGetOpeningBalanceV2.String()).Return(false)
				testHelper.mockAcctRepository.EXPECT().GetAccountBalancePeriodStart(args.ctx, args.req.AccountNumber, args.req.StartDate).Return(decimal.NewFromFloat(0), nil)
				testHelper.mockGoFpTransaction.EXPECT().GetAllOrderTypes(args.ctx, "", "").Return(gofptransaction.GetAllOrderTypesOut{}, nil)
				testHelper.mockFile.EXPECT().NewCSVWriter(gomock.Any())
				testHelper.mockFile.EXPECT().CSVWriteAll(gomock.Any(), gomock.Any()).Return(assert.AnError)
			},
			wantErr: true,
		},
		{
			name: "error case - create csv write header 1",
			args: args{
				ctx: context.Background(),
				req: models.SubLedgerFilterOptions{},
			},
			doMock: func(args args, mockData mockData) {
				testHelper.mockAccRepository.EXPECT().GetOneByAccountNumber(args.ctx, args.req.AccountNumber).Return(models.GetAccountOut{}, nil)
				testHelper.mockAcctRepository.EXPECT().GetSubLedgerCount(args.ctx, gomock.Any()).Return(5, nil)
				testHelper.mockAcctRepository.EXPECT().GetSubLedger(args.ctx, args.req).Return([]models.GetSubLedgerOut{
					{}, {}, {}, {}, {},
				}, nil)
				testHelper.mockFlag.EXPECT().IsEnabled(models.FlagGetOpeningBalanceV2.String()).Return(false)
				testHelper.mockAcctRepository.EXPECT().GetAccountBalancePeriodStart(args.ctx, args.req.AccountNumber, args.req.StartDate).Return(decimal.NewFromFloat(0), nil)
				testHelper.mockGoFpTransaction.EXPECT().GetAllOrderTypes(args.ctx, "", "").Return(gofptransaction.GetAllOrderTypesOut{}, nil)
				testHelper.mockFile.EXPECT().NewCSVWriter(gomock.Any())
				testHelper.mockFile.EXPECT().CSVWriteAll(gomock.Any(), gomock.Any()).Return(nil)
				testHelper.mockFile.EXPECT().CSVWriteHeader(gomock.Any(), gomock.Any()).Return(assert.AnError)
			},
			wantErr: true,
		},
		{
			name: "error case - create csv write header 2",
			args: args{
				ctx: context.Background(),
				req: models.SubLedgerFilterOptions{},
			},
			doMock: func(args args, mockData mockData) {
				testHelper.mockAccRepository.EXPECT().GetOneByAccountNumber(args.ctx, args.req.AccountNumber).Return(models.GetAccountOut{}, nil)
				testHelper.mockAcctRepository.EXPECT().GetSubLedgerCount(args.ctx, gomock.Any()).Return(1, nil)
				testHelper.mockAcctRepository.EXPECT().GetSubLedger(args.ctx, args.req).Return([]models.GetSubLedgerOut{
					{}, {}, {}, {}, {}, {},
				}, nil)
				testHelper.mockFlag.EXPECT().IsEnabled(models.FlagGetOpeningBalanceV2.String()).Return(false)
				testHelper.mockAcctRepository.EXPECT().GetAccountBalancePeriodStart(args.ctx, args.req.AccountNumber, args.req.StartDate).Return(decimal.NewFromFloat(0), nil)
				testHelper.mockGoFpTransaction.EXPECT().GetAllOrderTypes(args.ctx, "", "").Return(gofptransaction.GetAllOrderTypesOut{}, nil)
				testHelper.mockFile.EXPECT().NewCSVWriter(gomock.Any())
				testHelper.mockFile.EXPECT().CSVWriteAll(gomock.Any(), gomock.Any()).Return(nil)
				testHelper.mockFile.EXPECT().CSVWriteHeader(gomock.Any(), gomock.Any()).Return(nil)
				testHelper.mockFile.EXPECT().CSVWriteHeader(gomock.Any(), gomock.Any()).Return(assert.AnError)
			},
			wantErr: true,
		},
		{
			name: "error case - create csv write body",
			args: args{
				ctx: context.Background(),
				req: models.SubLedgerFilterOptions{},
			},
			doMock: func(args args, mockData mockData) {
				testHelper.mockAccRepository.EXPECT().GetOneByAccountNumber(args.ctx, args.req.AccountNumber).Return(models.GetAccountOut{
					CoaTypeCode: models.COATypeLiability,
				}, nil)
				testHelper.mockAcctRepository.EXPECT().GetSubLedgerCount(args.ctx, gomock.Any()).Return(1, nil)
				testHelper.mockAcctRepository.EXPECT().GetSubLedger(args.ctx, args.req).Return([]models.GetSubLedgerOut{
					{},
				}, nil)
				testHelper.mockFlag.EXPECT().IsEnabled(models.FlagGetOpeningBalanceV2.String()).Return(false)
				testHelper.mockAcctRepository.EXPECT().GetAccountBalancePeriodStart(args.ctx, args.req.AccountNumber, args.req.StartDate).Return(decimal.NewFromFloat(0), nil)
				testHelper.mockGoFpTransaction.EXPECT().GetAllOrderTypes(args.ctx, "", "").Return(gofptransaction.GetAllOrderTypesOut{}, nil)
				testHelper.mockFile.EXPECT().NewCSVWriter(gomock.Any())
				testHelper.mockFile.EXPECT().CSVWriteAll(args.ctx, gomock.Any()).Return(nil)
				testHelper.mockFile.EXPECT().CSVWriteHeader(args.ctx, gomock.Any()).Return(nil).MaxTimes(2)
				testHelper.mockFile.EXPECT().CSVWriteBody(args.ctx, gomock.Any()).Return(assert.AnError)
			},
			wantErr: true,
		},
		{
			name: "error case - csv write header 3",
			args: args{
				ctx: context.Background(),
				req: models.SubLedgerFilterOptions{},
			},
			doMock: func(args args, mockData mockData) {
				testHelper.mockAccRepository.EXPECT().GetOneByAccountNumber(args.ctx, args.req.AccountNumber).Return(models.GetAccountOut{
					CoaTypeCode: models.COATypeLiability,
				}, nil)
				testHelper.mockAcctRepository.EXPECT().GetSubLedgerCount(args.ctx, gomock.Any()).Return(1, nil)
				testHelper.mockAcctRepository.EXPECT().GetSubLedger(args.ctx, args.req).Return([]models.GetSubLedgerOut{
					{}, {}, {}, {}, {}, {},
				}, nil)
				testHelper.mockFlag.EXPECT().IsEnabled(models.FlagGetOpeningBalanceV2.String()).Return(false)
				testHelper.mockAcctRepository.EXPECT().GetAccountBalancePeriodStart(args.ctx, args.req.AccountNumber, args.req.StartDate).Return(decimal.NewFromFloat(0), nil)
				testHelper.mockGoFpTransaction.EXPECT().GetAllOrderTypes(args.ctx, "", "").Return(gofptransaction.GetAllOrderTypesOut{}, nil)
				testHelper.mockFile.EXPECT().NewCSVWriter(gomock.Any())
				testHelper.mockFile.EXPECT().CSVWriteAll(args.ctx, gomock.Any()).Return(nil)
				testHelper.mockFile.EXPECT().CSVWriteHeader(args.ctx, gomock.Any()).Return(nil).MaxTimes(2)
				testHelper.mockFile.EXPECT().CSVWriteBody(args.ctx, gomock.Any()).Return(nil).MaxTimes(7)
				testHelper.mockFile.EXPECT().CSVWriteHeader(args.ctx, gomock.Any()).Return(assert.AnError)
			},
			wantErr: true,
		},
		{
			name: "error case - csv process write",
			args: args{
				ctx: context.Background(),
				req: models.SubLedgerFilterOptions{},
			},
			doMock: func(args args, mockData mockData) {
				testHelper.mockAccRepository.EXPECT().GetOneByAccountNumber(args.ctx, args.req.AccountNumber).Return(models.GetAccountOut{
					CoaTypeCode: models.COATypeLiability,
				}, nil)
				testHelper.mockAcctRepository.EXPECT().GetSubLedgerCount(args.ctx, gomock.Any()).Return(1, nil)
				testHelper.mockAcctRepository.EXPECT().GetSubLedger(args.ctx, args.req).Return([]models.GetSubLedgerOut{
					{}, {}, {}, {}, {}, {},
				}, nil)
				testHelper.mockFlag.EXPECT().IsEnabled(models.FlagGetOpeningBalanceV2.String()).Return(false)
				testHelper.mockAcctRepository.EXPECT().GetAccountBalancePeriodStart(args.ctx, args.req.AccountNumber, args.req.StartDate).Return(decimal.NewFromFloat(0), nil)
				testHelper.mockGoFpTransaction.EXPECT().GetAllOrderTypes(args.ctx, "", "").Return(gofptransaction.GetAllOrderTypesOut{}, nil)
				testHelper.mockFile.EXPECT().NewCSVWriter(gomock.Any())
				testHelper.mockFile.EXPECT().CSVWriteAll(args.ctx, gomock.Any()).Return(nil)
				testHelper.mockFile.EXPECT().CSVWriteHeader(args.ctx, gomock.Any()).Return(nil).MaxTimes(2)
				testHelper.mockFile.EXPECT().CSVWriteBody(args.ctx, gomock.Any()).Return(nil).MaxTimes(7)
				testHelper.mockFile.EXPECT().CSVWriteHeader(args.ctx, gomock.Any()).Return(nil)
				testHelper.mockFile.EXPECT().CSVProcessWrite(args.ctx).Return(assert.AnError)
			},
			wantErr: true,
		},
		{
			name: "success case - download sub ledger",
			args: args{
				ctx: context.Background(),
				req: models.SubLedgerFilterOptions{},
			},
			doMock: func(args args, mockData mockData) {
				testHelper.mockAccRepository.EXPECT().GetOneByAccountNumber(args.ctx, args.req.AccountNumber).Return(models.GetAccountOut{
					CoaTypeCode: models.COATypeAsset,
				}, nil)
				testHelper.mockAcctRepository.EXPECT().GetSubLedgerCount(args.ctx, gomock.Any()).Return(1, nil)
				testHelper.mockAcctRepository.EXPECT().GetSubLedger(args.ctx, args.req).Return([]models.GetSubLedgerOut{
					{}, {}, {}, {}, {}, {},
				}, nil)
				testHelper.mockFlag.EXPECT().IsEnabled(models.FlagGetOpeningBalanceV2.String()).Return(false)
				testHelper.mockAcctRepository.EXPECT().GetAccountBalancePeriodStart(args.ctx, args.req.AccountNumber, args.req.StartDate).Return(decimal.NewFromFloat(0), nil)
				testHelper.mockGoFpTransaction.EXPECT().GetAllOrderTypes(args.ctx, "", "").Return(gofptransaction.GetAllOrderTypesOut{}, nil)
				testHelper.mockFile.EXPECT().NewCSVWriter(gomock.Any())
				testHelper.mockFile.EXPECT().CSVWriteAll(args.ctx, gomock.Any()).Return(nil)
				testHelper.mockFile.EXPECT().CSVWriteHeader(args.ctx, gomock.Any()).Return(nil).MaxTimes(2)
				testHelper.mockFile.EXPECT().CSVWriteBody(args.ctx, gomock.Any()).Return(nil).MaxTimes(6)
				testHelper.mockFile.EXPECT().CSVWriteHeader(args.ctx, gomock.Any()).Return(nil)
				testHelper.mockFile.EXPECT().CSVProcessWrite(args.ctx).Return(nil)
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if tt.doMock != nil {
				tt.doMock(tt.args, tt.mockData)
			}
			_, _, err := testHelper.accountingService.DownloadSubLedgerCSV(tt.args.ctx, tt.args.req)
			assert.Equal(t, tt.wantErr, err != nil)
		})
	}
}

func Test_accounting_SendSubLedgerCSVToEmail(t *testing.T) {
	testHelper := serviceTestHelper(t)

	type args struct {
		ctx context.Context
		req models.SubLedgerFilterOptions
	}
	type mockData struct {
	}
	tests := []struct {
		name     string
		args     args
		mockData mockData
		doMock   func(args args, mockData mockData)
		wantErr  bool
	}{
		{
			name: "error case - account number not found",
			args: args{
				ctx: context.Background(),
				req: models.SubLedgerFilterOptions{},
			},
			doMock: func(args args, mockData mockData) {
				testHelper.mockAccRepository.EXPECT().GetOneByAccountNumber(args.ctx, args.req.AccountNumber).Return(models.GetAccountOut{}, models.ErrNoRows)
			},
			wantErr: true,
		},
		{
			name: "error case - database error get entity",
			args: args{
				ctx: context.Background(),
				req: models.SubLedgerFilterOptions{},
			},
			doMock: func(args args, mockData mockData) {
				testHelper.mockAccRepository.EXPECT().GetOneByAccountNumber(args.ctx, args.req.AccountNumber).Return(models.GetAccountOut{}, nil)
				testHelper.mockEntityRepository.EXPECT().GetByCode(args.ctx, gomock.Any()).Return(&models.Entity{}, models.GetErrMap(models.ErrKeyDatabaseError))
			},
			wantErr: true,
		},
		{
			name: "error case - error data not found get entity",
			args: args{
				ctx: context.Background(),
				req: models.SubLedgerFilterOptions{},
			},
			doMock: func(args args, mockData mockData) {
				testHelper.mockAccRepository.EXPECT().GetOneByAccountNumber(args.ctx, args.req.AccountNumber).Return(models.GetAccountOut{}, nil)
				testHelper.mockEntityRepository.EXPECT().GetByCode(args.ctx, gomock.Any()).Return(nil, nil)
			},
			wantErr: true,
		},
		{
			name: "error case - database error get account balance period start",
			args: args{
				ctx: context.Background(),
				req: models.SubLedgerFilterOptions{},
			},
			doMock: func(args args, mockData mockData) {
				testHelper.mockAccRepository.EXPECT().GetOneByAccountNumber(args.ctx, args.req.AccountNumber).Return(models.GetAccountOut{}, nil)
				testHelper.mockEntityRepository.EXPECT().GetByCode(args.ctx, gomock.Any()).Return(&models.Entity{}, nil)
				testHelper.mockFlag.EXPECT().IsEnabled(models.FlagGetOpeningBalanceV2.String()).Return(false)
				testHelper.mockAcctRepository.EXPECT().GetAccountBalancePeriodStart(args.ctx, args.req.AccountNumber, args.req.StartDate).Return(decimal.NewFromFloat(0), models.GetErrMap(models.ErrKeyDatabaseError))
			},
			wantErr: true,
		},
		{
			name: "error case - error get all order types",
			args: args{
				ctx: context.Background(),
				req: models.SubLedgerFilterOptions{},
			},
			doMock: func(args args, mockData mockData) {
				testHelper.mockAccRepository.EXPECT().GetOneByAccountNumber(args.ctx, args.req.AccountNumber).Return(models.GetAccountOut{}, nil)
				testHelper.mockEntityRepository.EXPECT().GetByCode(args.ctx, gomock.Any()).Return(&models.Entity{}, nil)
				testHelper.mockFlag.EXPECT().IsEnabled(models.FlagGetOpeningBalanceV2.String()).Return(false)
				testHelper.mockAcctRepository.EXPECT().GetAccountBalancePeriodStart(args.ctx, args.req.AccountNumber, args.req.StartDate).Return(decimal.NewFromFloat(0), nil)
				testHelper.mockGoFpTransaction.EXPECT().GetAllOrderTypes(args.ctx, "", "").Return(gofptransaction.GetAllOrderTypesOut{}, assert.AnError)
			},
			wantErr: true,
		},
		{
			name: "error case - create csv write all",
			args: args{
				ctx: context.Background(),
				req: models.SubLedgerFilterOptions{},
			},
			doMock: func(args args, mockData mockData) {
				testHelper.mockAccRepository.EXPECT().GetOneByAccountNumber(args.ctx, args.req.AccountNumber).Return(models.GetAccountOut{}, nil)
				testHelper.mockEntityRepository.EXPECT().GetByCode(args.ctx, gomock.Any()).Return(&models.Entity{}, nil)
				testHelper.mockFlag.EXPECT().IsEnabled(models.FlagGetOpeningBalanceV2.String()).Return(false)
				testHelper.mockAcctRepository.EXPECT().GetAccountBalancePeriodStart(args.ctx, args.req.AccountNumber, args.req.StartDate).Return(decimal.NewFromFloat(0), nil)
				testHelper.mockGoFpTransaction.EXPECT().GetAllOrderTypes(args.ctx, "", "").Return(gofptransaction.GetAllOrderTypesOut{}, nil)
				testHelper.mockFile.EXPECT().NewCSVWriter(gomock.Any())

				testHelper.mockFile.EXPECT().CSVWriteAll(gomock.Any(), gomock.Any()).Return(assert.AnError)
				testHelper.mockAcctRepository.EXPECT().GetSubLedgerStream(args.ctx, args.req).
					DoAndReturn(func(ctx context.Context, opts models.SubLedgerFilterOptions) <-chan models.StreamResult[models.GetSubLedgerOut] {
						resultCh := make(chan models.StreamResult[models.GetSubLedgerOut], 1)
						go func() {
							defer close(resultCh)
							resultCh <- models.StreamResult[models.GetSubLedgerOut]{Data: models.GetSubLedgerOut{
								TransactionID:       "3c8c389b-abbc-4d17-a718-7bd84721b40f",
								ReferenceNumber:     "12345",
								TransactionDate:     time.Date(2023, time.December, 22, 00, 00, 00, 00, &time.Location{}),
								TransactionType:     "DSBAC",
								TransactionTypeName: "Admin Fee Partner Loan Deduction",
								Narrative:           "Invest To Loan",
								Debit:               decimal.New(30.000, 00),
								Credit:              decimal.New(0, 00),
							}}
						}()
						return resultCh
					})
			},
			wantErr: true,
		},
		{
			name: "error case - create csv write header 1",
			args: args{
				ctx: context.Background(),
				req: models.SubLedgerFilterOptions{},
			},
			doMock: func(args args, mockData mockData) {
				testHelper.mockAccRepository.EXPECT().GetOneByAccountNumber(args.ctx, args.req.AccountNumber).Return(models.GetAccountOut{}, nil)
				testHelper.mockEntityRepository.EXPECT().GetByCode(args.ctx, gomock.Any()).Return(&models.Entity{}, nil)
				testHelper.mockFlag.EXPECT().IsEnabled(models.FlagGetOpeningBalanceV2.String()).Return(false)
				testHelper.mockAcctRepository.EXPECT().GetAccountBalancePeriodStart(args.ctx, args.req.AccountNumber, args.req.StartDate).Return(decimal.NewFromFloat(0), nil)
				testHelper.mockGoFpTransaction.EXPECT().GetAllOrderTypes(args.ctx, "", "").Return(gofptransaction.GetAllOrderTypesOut{}, nil)
				testHelper.mockFile.EXPECT().NewCSVWriter(gomock.Any())

				testHelper.mockFile.EXPECT().CSVWriteAll(gomock.Any(), gomock.Any()).Return(nil)
				testHelper.mockFile.EXPECT().CSVWriteHeader(gomock.Any(), gomock.Any()).Return(assert.AnError)
				testHelper.mockAcctRepository.EXPECT().GetSubLedgerStream(args.ctx, args.req).
					DoAndReturn(func(ctx context.Context, opts models.SubLedgerFilterOptions) <-chan models.StreamResult[models.GetSubLedgerOut] {
						resultCh := make(chan models.StreamResult[models.GetSubLedgerOut], 1)
						go func() {
							defer close(resultCh)
							resultCh <- models.StreamResult[models.GetSubLedgerOut]{Data: models.GetSubLedgerOut{
								TransactionID:       "3c8c389b-abbc-4d17-a718-7bd84721b40f",
								ReferenceNumber:     "12345",
								TransactionDate:     time.Date(2023, time.December, 22, 00, 00, 00, 00, &time.Location{}),
								TransactionType:     "DSBAC",
								TransactionTypeName: "Admin Fee Partner Loan Deduction",
								Narrative:           "Invest To Loan",
								Debit:               decimal.New(30.000, 00),
								Credit:              decimal.New(0, 00),
							}}
						}()
						return resultCh
					})
			},
			wantErr: true,
		},
		{
			name: "error case - create csv write header 2",
			args: args{
				ctx: context.Background(),
				req: models.SubLedgerFilterOptions{},
			},
			doMock: func(args args, mockData mockData) {
				testHelper.mockAccRepository.EXPECT().GetOneByAccountNumber(args.ctx, args.req.AccountNumber).Return(models.GetAccountOut{}, nil)
				testHelper.mockEntityRepository.EXPECT().GetByCode(args.ctx, gomock.Any()).Return(&models.Entity{}, nil)
				testHelper.mockFlag.EXPECT().IsEnabled(models.FlagGetOpeningBalanceV2.String()).Return(false)
				testHelper.mockAcctRepository.EXPECT().GetAccountBalancePeriodStart(args.ctx, args.req.AccountNumber, args.req.StartDate).Return(decimal.NewFromFloat(0), nil)
				testHelper.mockGoFpTransaction.EXPECT().GetAllOrderTypes(args.ctx, "", "").Return(gofptransaction.GetAllOrderTypesOut{}, nil)
				testHelper.mockFile.EXPECT().NewCSVWriter(gomock.Any())

				testHelper.mockFile.EXPECT().CSVWriteAll(gomock.Any(), gomock.Any()).Return(nil)
				testHelper.mockFile.EXPECT().CSVWriteHeader(gomock.Any(), gomock.Any()).Return(nil)
				testHelper.mockFile.EXPECT().CSVWriteHeader(gomock.Any(), gomock.Any()).Return(assert.AnError)
				testHelper.mockAcctRepository.EXPECT().GetSubLedgerStream(args.ctx, args.req).
					DoAndReturn(func(ctx context.Context, opts models.SubLedgerFilterOptions) <-chan models.StreamResult[models.GetSubLedgerOut] {
						resultCh := make(chan models.StreamResult[models.GetSubLedgerOut], 1)
						go func() {
							defer close(resultCh)
							resultCh <- models.StreamResult[models.GetSubLedgerOut]{Data: models.GetSubLedgerOut{
								TransactionID:       "3c8c389b-abbc-4d17-a718-7bd84721b40f",
								ReferenceNumber:     "12345",
								TransactionDate:     time.Date(2023, time.December, 22, 00, 00, 00, 00, &time.Location{}),
								TransactionType:     "DSBAC",
								TransactionTypeName: "Admin Fee Partner Loan Deduction",
								Narrative:           "Invest To Loan",
								Debit:               decimal.New(30.000, 00),
								Credit:              decimal.New(0, 00),
							}}
						}()
						return resultCh
					})
			},
			wantErr: true,
		},
		{
			name: "error case - create csv write body",
			args: args{
				ctx: context.Background(),
				req: models.SubLedgerFilterOptions{},
			},
			doMock: func(args args, mockData mockData) {
				testHelper.mockAccRepository.EXPECT().GetOneByAccountNumber(args.ctx, args.req.AccountNumber).Return(models.GetAccountOut{
					CoaTypeCode: models.COATypeLiability,
				}, nil)
				testHelper.mockEntityRepository.EXPECT().GetByCode(args.ctx, gomock.Any()).Return(&models.Entity{}, nil)
				testHelper.mockFlag.EXPECT().IsEnabled(models.FlagGetOpeningBalanceV2.String()).Return(false)
				testHelper.mockAcctRepository.EXPECT().GetAccountBalancePeriodStart(args.ctx, args.req.AccountNumber, args.req.StartDate).Return(decimal.NewFromFloat(0), nil)
				testHelper.mockGoFpTransaction.EXPECT().GetAllOrderTypes(args.ctx, "", "").Return(gofptransaction.GetAllOrderTypesOut{}, nil)
				testHelper.mockFile.EXPECT().NewCSVWriter(gomock.Any())
				testHelper.mockFile.EXPECT().CSVWriteAll(gomock.Any(), gomock.Any()).Return(nil)
				testHelper.mockFile.EXPECT().CSVWriteHeader(gomock.Any(), gomock.Any()).Return(nil).MaxTimes(2)

				testHelper.mockAcctRepository.EXPECT().GetSubLedgerStream(args.ctx, args.req).
					DoAndReturn(func(ctx context.Context, opts models.SubLedgerFilterOptions) <-chan models.StreamResult[models.GetSubLedgerOut] {
						resultCh := make(chan models.StreamResult[models.GetSubLedgerOut], 1)
						go func() {
							defer close(resultCh)
							resultCh <- models.StreamResult[models.GetSubLedgerOut]{Data: models.GetSubLedgerOut{
								TransactionID:       "3c8c389b-abbc-4d17-a718-7bd84721b40f",
								ReferenceNumber:     "12345",
								TransactionDate:     time.Date(2023, time.December, 22, 00, 00, 00, 00, &time.Location{}),
								TransactionType:     "DSBAC",
								TransactionTypeName: "Admin Fee Partner Loan Deduction",
								Narrative:           "Invest To Loan",
								Debit:               decimal.New(30.000, 00),
								Credit:              decimal.New(0, 00),
							}}
						}()
						return resultCh
					})
				testHelper.mockFile.EXPECT().CSVWriteBody(gomock.Any(), gomock.Any()).Return(assert.AnError)
			},
			wantErr: true,
		},
		{
			name: "error case - csv write header 3",
			args: args{
				ctx: context.Background(),
				req: models.SubLedgerFilterOptions{},
			},
			doMock: func(args args, mockData mockData) {
				testHelper.mockAccRepository.EXPECT().GetOneByAccountNumber(args.ctx, args.req.AccountNumber).Return(models.GetAccountOut{
					CoaTypeCode: models.COATypeLiability,
				}, nil)
				testHelper.mockEntityRepository.EXPECT().GetByCode(args.ctx, gomock.Any()).Return(&models.Entity{}, nil)
				testHelper.mockFlag.EXPECT().IsEnabled(models.FlagGetOpeningBalanceV2.String()).Return(false)
				testHelper.mockAcctRepository.EXPECT().GetAccountBalancePeriodStart(args.ctx, args.req.AccountNumber, args.req.StartDate).Return(decimal.NewFromFloat(0), nil)
				testHelper.mockGoFpTransaction.EXPECT().GetAllOrderTypes(args.ctx, "", "").Return(gofptransaction.GetAllOrderTypesOut{}, nil)
				testHelper.mockFile.EXPECT().NewCSVWriter(gomock.Any())

				testHelper.mockFile.EXPECT().CSVWriteAll(gomock.Any(), gomock.Any()).Return(nil)
				testHelper.mockFile.EXPECT().CSVWriteHeader(gomock.Any(), gomock.Any()).Return(nil).MaxTimes(2)
				testHelper.mockFile.EXPECT().CSVWriteBody(gomock.Any(), gomock.Any()).Return(nil).MaxTimes(7)
				testHelper.mockAcctRepository.EXPECT().GetSubLedgerStream(args.ctx, args.req).
					DoAndReturn(func(ctx context.Context, opts models.SubLedgerFilterOptions) <-chan models.StreamResult[models.GetSubLedgerOut] {
						resultCh := make(chan models.StreamResult[models.GetSubLedgerOut], 1)
						go func() {
							defer close(resultCh)
							resultCh <- models.StreamResult[models.GetSubLedgerOut]{Data: models.GetSubLedgerOut{
								TransactionID:       "3c8c389b-abbc-4d17-a718-7bd84721b40f",
								ReferenceNumber:     "12345",
								TransactionDate:     time.Date(2023, time.December, 22, 00, 00, 00, 00, &time.Location{}),
								TransactionType:     "DSBAC",
								TransactionTypeName: "Admin Fee Partner Loan Deduction",
								Narrative:           "Invest To Loan",
								Debit:               decimal.New(30.000, 00),
								Credit:              decimal.New(0, 00),
							}}
						}()
						return resultCh
					})
				testHelper.mockFile.EXPECT().CSVWriteHeader(gomock.Any(), gomock.Any()).Return(assert.AnError)
			},
			wantErr: true,
		},
		{
			name: "error case - csv process write",
			args: args{
				ctx: context.Background(),
				req: models.SubLedgerFilterOptions{},
			},
			doMock: func(args args, mockData mockData) {
				testHelper.mockAccRepository.EXPECT().GetOneByAccountNumber(args.ctx, args.req.AccountNumber).Return(models.GetAccountOut{
					CoaTypeCode: models.COATypeLiability,
				}, nil)
				testHelper.mockEntityRepository.EXPECT().GetByCode(args.ctx, gomock.Any()).Return(&models.Entity{}, nil)
				testHelper.mockFlag.EXPECT().IsEnabled(models.FlagGetOpeningBalanceV2.String()).Return(false)
				testHelper.mockAcctRepository.EXPECT().GetAccountBalancePeriodStart(args.ctx, args.req.AccountNumber, args.req.StartDate).Return(decimal.NewFromFloat(0), nil)
				testHelper.mockGoFpTransaction.EXPECT().GetAllOrderTypes(args.ctx, "", "").Return(gofptransaction.GetAllOrderTypesOut{}, nil)
				testHelper.mockFile.EXPECT().NewCSVWriter(gomock.Any())

				testHelper.mockFile.EXPECT().CSVWriteAll(gomock.Any(), gomock.Any()).Return(nil)
				testHelper.mockFile.EXPECT().CSVWriteHeader(gomock.Any(), gomock.Any()).Return(nil).MaxTimes(2)
				testHelper.mockFile.EXPECT().CSVWriteBody(gomock.Any(), gomock.Any()).Return(nil).MaxTimes(7)
				testHelper.mockFile.EXPECT().CSVWriteHeader(gomock.Any(), gomock.Any()).Return(nil)
				testHelper.mockAcctRepository.EXPECT().GetSubLedgerStream(args.ctx, args.req).
					DoAndReturn(func(ctx context.Context, opts models.SubLedgerFilterOptions) <-chan models.StreamResult[models.GetSubLedgerOut] {
						resultCh := make(chan models.StreamResult[models.GetSubLedgerOut], 1)
						go func() {
							defer close(resultCh)
							resultCh <- models.StreamResult[models.GetSubLedgerOut]{Data: models.GetSubLedgerOut{
								TransactionID:       "3c8c389b-abbc-4d17-a718-7bd84721b40f",
								ReferenceNumber:     "12345",
								TransactionDate:     time.Date(2023, time.December, 22, 00, 00, 00, 00, &time.Location{}),
								TransactionType:     "DSBAC",
								TransactionTypeName: "Admin Fee Partner Loan Deduction",
								Narrative:           "Invest To Loan",
								Debit:               decimal.New(30.000, 00),
								Credit:              decimal.New(0, 00),
							}}
						}()
						return resultCh
					})
				testHelper.mockFile.EXPECT().CSVProcessWrite(gomock.Any()).Return(assert.AnError)
			},
			wantErr: true,
		},
		{
			name: "error case - get signer url",
			args: args{
				ctx: context.Background(),
				req: models.SubLedgerFilterOptions{},
			},
			doMock: func(args args, mockData mockData) {
				testHelper.mockAccRepository.EXPECT().GetOneByAccountNumber(args.ctx, args.req.AccountNumber).Return(models.GetAccountOut{
					CoaTypeCode: models.COATypeAsset,
				}, nil)
				testHelper.mockEntityRepository.EXPECT().GetByCode(args.ctx, gomock.Any()).Return(&models.Entity{}, nil)
				testHelper.mockFlag.EXPECT().IsEnabled(models.FlagGetOpeningBalanceV2.String()).Return(false)
				testHelper.mockAcctRepository.EXPECT().GetAccountBalancePeriodStart(args.ctx, args.req.AccountNumber, args.req.StartDate).Return(decimal.NewFromFloat(0), nil)
				testHelper.mockGoFpTransaction.EXPECT().GetAllOrderTypes(args.ctx, "", "").Return(gofptransaction.GetAllOrderTypesOut{}, nil)
				testHelper.mockFile.EXPECT().NewCSVWriter(gomock.Any())

				testHelper.mockFile.EXPECT().CSVWriteAll(gomock.Any(), gomock.Any()).Return(nil)
				testHelper.mockFile.EXPECT().CSVWriteHeader(gomock.Any(), gomock.Any()).Return(nil).MaxTimes(2)
				testHelper.mockAcctRepository.EXPECT().GetSubLedgerStream(args.ctx, args.req).
					DoAndReturn(func(ctx context.Context, opts models.SubLedgerFilterOptions) <-chan models.StreamResult[models.GetSubLedgerOut] {
						resultCh := make(chan models.StreamResult[models.GetSubLedgerOut], 1)
						go func() {
							defer close(resultCh)
							resultCh <- models.StreamResult[models.GetSubLedgerOut]{Data: models.GetSubLedgerOut{
								TransactionID:       "3c8c389b-abbc-4d17-a718-7bd84721b40f",
								ReferenceNumber:     "12345",
								TransactionDate:     time.Date(2023, time.December, 22, 00, 00, 00, 00, &time.Location{}),
								TransactionType:     "DSBAC",
								TransactionTypeName: "Admin Fee Partner Loan Deduction",
								Narrative:           "Invest To Loan",
								Debit:               decimal.New(30.000, 00),
								Credit:              decimal.New(0, 00),
							}}
						}()
						return resultCh
					})
				testHelper.mockFile.EXPECT().CSVWriteBody(gomock.Any(), gomock.Any()).Return(nil).MaxTimes(7)
				testHelper.mockFile.EXPECT().CSVWriteHeader(gomock.Any(), gomock.Any()).Return(nil)
				testHelper.mockFile.EXPECT().CSVProcessWrite(gomock.Any()).Return(nil)

				gcsPayload := &models.CloudStoragePayload{
					Filename: fmt.Sprintf("SubLedger-%s-%s-%s-%s-%d.csv", "", "", args.req.StartDate.Format(atime.DateFormatYYYYMMDDWithoutDash), args.req.EndDate.Format(atime.DateFormatYYYYMMDDWithoutDash), 1),
					Path:     string(models.SubLedgerDir),
				}
				tempFile, _ := os.CreateTemp("", "test_mock_gcs")

				testHelper.mockCloudStorageRepository.EXPECT().NewWriter(args.ctx, gcsPayload).Return(tempFile)
				testHelper.mockCloudStorageRepository.EXPECT().GetSignedURL(*gcsPayload, gomock.Any()).Return("https://test.com", assert.AnError)
			},
			wantErr: true,
		},
		{
			name: "error case - send email",
			args: args{
				ctx: context.Background(),
				req: models.SubLedgerFilterOptions{},
			},
			doMock: func(args args, mockData mockData) {
				testHelper.mockAccRepository.EXPECT().GetOneByAccountNumber(args.ctx, args.req.AccountNumber).Return(models.GetAccountOut{
					CoaTypeCode: models.COATypeAsset,
				}, nil)
				testHelper.mockEntityRepository.EXPECT().GetByCode(args.ctx, gomock.Any()).Return(&models.Entity{}, nil)
				testHelper.mockFlag.EXPECT().IsEnabled(models.FlagGetOpeningBalanceV2.String()).Return(false)
				testHelper.mockAcctRepository.EXPECT().GetAccountBalancePeriodStart(args.ctx, args.req.AccountNumber, args.req.StartDate).Return(decimal.NewFromFloat(0), nil)
				testHelper.mockGoFpTransaction.EXPECT().GetAllOrderTypes(args.ctx, "", "").Return(gofptransaction.GetAllOrderTypesOut{}, nil)
				testHelper.mockFile.EXPECT().NewCSVWriter(gomock.Any())

				testHelper.mockFile.EXPECT().CSVWriteAll(gomock.Any(), gomock.Any()).Return(nil)
				testHelper.mockFile.EXPECT().CSVWriteHeader(gomock.Any(), gomock.Any()).Return(nil).MaxTimes(2)
				testHelper.mockAcctRepository.EXPECT().GetSubLedgerStream(args.ctx, args.req).
					DoAndReturn(func(ctx context.Context, opts models.SubLedgerFilterOptions) <-chan models.StreamResult[models.GetSubLedgerOut] {
						resultCh := make(chan models.StreamResult[models.GetSubLedgerOut], 1)
						go func() {
							defer close(resultCh)
							resultCh <- models.StreamResult[models.GetSubLedgerOut]{Data: models.GetSubLedgerOut{
								TransactionID:       "3c8c389b-abbc-4d17-a718-7bd84721b40f",
								ReferenceNumber:     "12345",
								TransactionDate:     time.Date(2023, time.December, 22, 00, 00, 00, 00, &time.Location{}),
								TransactionType:     "DSBAC",
								TransactionTypeName: "Admin Fee Partner Loan Deduction",
								Narrative:           "Invest To Loan",
								Debit:               decimal.New(30.000, 00),
								Credit:              decimal.New(0, 00),
							}}
						}()
						return resultCh
					})
				testHelper.mockFile.EXPECT().CSVWriteBody(gomock.Any(), gomock.Any()).Return(nil).MaxTimes(7)
				testHelper.mockFile.EXPECT().CSVWriteHeader(gomock.Any(), gomock.Any()).Return(nil)
				testHelper.mockFile.EXPECT().CSVProcessWrite(gomock.Any()).Return(nil)

				gcsPayload := &models.CloudStoragePayload{
					Filename: fmt.Sprintf("SubLedger-%s-%s-%s-%s-%d.csv", "", "", args.req.StartDate.Format(atime.DateFormatYYYYMMDDWithoutDash), args.req.EndDate.Format(atime.DateFormatYYYYMMDDWithoutDash), 1),
					Path:     string(models.SubLedgerDir),
				}
				tempFile, _ := os.CreateTemp("", "test_mock_gcs")

				testHelper.mockCloudStorageRepository.EXPECT().NewWriter(args.ctx, gcsPayload).Return(tempFile)
				testHelper.mockCloudStorageRepository.EXPECT().GetSignedURL(*gcsPayload, gomock.Any()).Return("https://test.com", nil)
				testHelper.mockDDDNotification.EXPECT().SendEmail(args.ctx, gomock.Any()).Return(assert.AnError)
			},
			wantErr: true,
		},
		{
			name: "success case - download sub ledger",
			args: args{
				ctx: context.Background(),
				req: models.SubLedgerFilterOptions{},
			},
			doMock: func(args args, mockData mockData) {
				testHelper.mockAccRepository.EXPECT().GetOneByAccountNumber(args.ctx, args.req.AccountNumber).Return(models.GetAccountOut{
					CoaTypeCode: models.COATypeAsset,
				}, nil)
				testHelper.mockEntityRepository.EXPECT().GetByCode(args.ctx, gomock.Any()).Return(&models.Entity{}, nil)
				testHelper.mockFlag.EXPECT().IsEnabled(models.FlagGetOpeningBalanceV2.String()).Return(false)
				testHelper.mockAcctRepository.EXPECT().GetAccountBalancePeriodStart(args.ctx, args.req.AccountNumber, args.req.StartDate).Return(decimal.NewFromFloat(0), nil)
				testHelper.mockGoFpTransaction.EXPECT().GetAllOrderTypes(args.ctx, "", "").Return(gofptransaction.GetAllOrderTypesOut{}, nil)
				testHelper.mockFile.EXPECT().NewCSVWriter(gomock.Any())

				testHelper.mockFile.EXPECT().CSVWriteAll(gomock.Any(), gomock.Any()).Return(nil)
				testHelper.mockFile.EXPECT().CSVWriteHeader(gomock.Any(), gomock.Any()).Return(nil).MaxTimes(2)
				testHelper.mockAcctRepository.EXPECT().GetSubLedgerStream(args.ctx, args.req).
					DoAndReturn(func(ctx context.Context, opts models.SubLedgerFilterOptions) <-chan models.StreamResult[models.GetSubLedgerOut] {
						resultCh := make(chan models.StreamResult[models.GetSubLedgerOut], 1)
						go func() {
							defer close(resultCh)
							resultCh <- models.StreamResult[models.GetSubLedgerOut]{Data: models.GetSubLedgerOut{
								TransactionID:       "3c8c389b-abbc-4d17-a718-7bd84721b40f",
								ReferenceNumber:     "12345",
								TransactionDate:     time.Date(2023, time.December, 22, 00, 00, 00, 00, &time.Location{}),
								TransactionType:     "DSBAC",
								TransactionTypeName: "Admin Fee Partner Loan Deduction",
								Narrative:           "Invest To Loan",
								Debit:               decimal.New(30.000, 00),
								Credit:              decimal.New(0, 00),
							}}
						}()
						return resultCh
					})
				testHelper.mockFile.EXPECT().CSVWriteBody(gomock.Any(), gomock.Any()).Return(nil).MaxTimes(7)
				testHelper.mockFile.EXPECT().CSVWriteHeader(gomock.Any(), gomock.Any()).Return(nil)
				testHelper.mockFile.EXPECT().CSVProcessWrite(gomock.Any()).Return(nil)

				gcsPayload := &models.CloudStoragePayload{
					Filename: fmt.Sprintf("SubLedger-%s-%s-%s-%s-%d.csv", "", "", args.req.StartDate.Format(atime.DateFormatYYYYMMDDWithoutDash), args.req.EndDate.Format(atime.DateFormatYYYYMMDDWithoutDash), 1),
					Path:     string(models.SubLedgerDir),
				}
				tempFile, _ := os.CreateTemp("", "test_mock_gcs")

				testHelper.mockCloudStorageRepository.EXPECT().NewWriter(args.ctx, gcsPayload).Return(tempFile)
				testHelper.mockCloudStorageRepository.EXPECT().GetSignedURL(*gcsPayload, gomock.Any()).Return("https://test.com", nil)
				testHelper.mockDDDNotification.EXPECT().SendEmail(args.ctx, gomock.Any()).Return(nil)
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
			err := testHelper.accountingService.SendSubLedgerCSVToEmail(tt.args.ctx, tt.args.req)
			assert.Equal(t, tt.wantErr, err != nil)
		})
	}
}
