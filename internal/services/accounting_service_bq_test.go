package services_test

import (
	"context"
	"testing"
	"time"

	"bitbucket.org/Amartha/go-accounting/internal/models"
	"bitbucket.org/Amartha/go-accounting/internal/pkg/atime"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func Test_accounting_GenerateTrialBalanceBigQuery(t *testing.T) {
	testHelper := serviceTestHelper(t)
	ctx := context.Background()
	date, _ := atime.NowZeroTime()

	tests := []struct {
		name    string
		req     time.Time
		doMock  func(ctx context.Context, req time.Time)
		wantErr bool
	}{
		{
			name: "success case - trial balance already generate",
			req:  date,
			doMock: func(ctx context.Context, req time.Time) {
				testHelper.mockBigQuery.EXPECT().
					TableExists(ctx, gomock.Any()).
					Return(true, nil)
			},
			wantErr: false,
		},
		{
			name: "success case",
			req:  date,
			doMock: func(ctx context.Context, req time.Time) {
				testHelper.mockBigQuery.EXPECT().
					TableExists(ctx, gomock.Any()).
					Return(false, nil)
				testHelper.mockFlag.EXPECT().
					IsEnabled(models.FlagGetOpeningBalanceFromPreviousMonth.String()).
					Return(true)
				testHelper.mockBigQuery.EXPECT().
					QueryGenerateTrialBalanceDetail(ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil)
				testHelper.mockBigQuery.EXPECT().
					QueryGenerateTrialBalanceSummary(ctx, gomock.Any()).
					Return(nil)
				testHelper.mockBigQuery.EXPECT().
					ExportTrialBalanceSummary(ctx, gomock.Any(), gomock.Any(), gomock.Any()).
					Return([]models.CreateTrialBalancePeriod{}, nil)
				testHelper.mockSubCategoryRepository.EXPECT().
					GetAll(gomock.AssignableToTypeOf(context.Background()), models.GetAllSubCategoryParam{}).
					Return(&[]models.SubCategory{{
						Code: "13101",
					}}, nil)
				testHelper.mockBigQuery.EXPECT().
					ExportTrialBalanceDetail(ctx, gomock.Any(), gomock.Any(), gomock.Any(), &[]models.SubCategory{{
						Code: "13101",
					}}).
					Return(nil)
				testHelper.mockTrialBalanceRepository.EXPECT().
					BulkInsert(ctx, gomock.Any()).
					Return(nil)
			},
			wantErr: false,
		},
		{
			name: "error case - when check table is exist",
			req:  date,
			doMock: func(ctx context.Context, req time.Time) {
				testHelper.mockBigQuery.EXPECT().
					TableExists(ctx, gomock.Any()).
					Return(true, assert.AnError)
			},
			wantErr: true,
		},
		{
			name: "error case - exec query generate trial balance detail",
			req:  date,
			doMock: func(ctx context.Context, req time.Time) {
				testHelper.mockBigQuery.EXPECT().
					TableExists(ctx, gomock.Any()).
					Return(false, nil)
				testHelper.mockFlag.EXPECT().
					IsEnabled(models.FlagGetOpeningBalanceFromPreviousMonth.String()).
					Return(true)
				testHelper.mockBigQuery.EXPECT().
					QueryGenerateTrialBalanceDetail(ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(assert.AnError)
			},
			wantErr: true,
		},
		{
			name: "error case - exec query generate trial balance summary",
			req:  date,
			doMock: func(ctx context.Context, req time.Time) {
				testHelper.mockBigQuery.EXPECT().
					TableExists(ctx, gomock.Any()).
					Return(false, nil)
				testHelper.mockFlag.EXPECT().
					IsEnabled(models.FlagGetOpeningBalanceFromPreviousMonth.String()).
					Return(true)
				testHelper.mockBigQuery.EXPECT().
					QueryGenerateTrialBalanceDetail(ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil)
				testHelper.mockBigQuery.EXPECT().
					QueryGenerateTrialBalanceSummary(ctx, gomock.Any()).
					Return(assert.AnError)
			},
			wantErr: true,
		},
		{
			name: "error case - export trial balance summary",
			req:  date,
			doMock: func(ctx context.Context, req time.Time) {
				testHelper.mockBigQuery.EXPECT().
					TableExists(ctx, gomock.Any()).
					Return(false, nil)
				testHelper.mockFlag.EXPECT().
					IsEnabled(models.FlagGetOpeningBalanceFromPreviousMonth.String()).
					Return(true)
				testHelper.mockBigQuery.EXPECT().
					QueryGenerateTrialBalanceDetail(ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil)
				testHelper.mockBigQuery.EXPECT().
					QueryGenerateTrialBalanceSummary(ctx, gomock.Any()).
					Return(nil)
				testHelper.mockBigQuery.EXPECT().
					ExportTrialBalanceSummary(ctx, gomock.Any(), gomock.Any(), gomock.Any()).
					Return([]models.CreateTrialBalancePeriod{}, assert.AnError)
			},
			wantErr: true,
		},
		{
			name: "error case - get all subcategory",
			req:  date,
			doMock: func(ctx context.Context, req time.Time) {
				testHelper.mockBigQuery.EXPECT().
					TableExists(ctx, gomock.Any()).
					Return(false, nil)
				testHelper.mockFlag.EXPECT().
					IsEnabled(models.FlagGetOpeningBalanceFromPreviousMonth.String()).
					Return(true)
				testHelper.mockBigQuery.EXPECT().
					QueryGenerateTrialBalanceDetail(ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil)
				testHelper.mockBigQuery.EXPECT().
					ExportTrialBalanceSummary(ctx, gomock.Any(), gomock.Any(), gomock.Any()).
					Return([]models.CreateTrialBalancePeriod{}, nil)
				testHelper.mockBigQuery.EXPECT().
					QueryGenerateTrialBalanceSummary(ctx, gomock.Any()).
					Return(nil)
				testHelper.mockSubCategoryRepository.EXPECT().
					GetAll(gomock.AssignableToTypeOf(context.Background()), models.GetAllSubCategoryParam{}).
					Return(nil, assert.AnError)
			},
			wantErr: true,
		},
		{
			name: "error case - export trial balance detail",
			req:  date,
			doMock: func(ctx context.Context, req time.Time) {
				testHelper.mockBigQuery.EXPECT().
					TableExists(ctx, gomock.Any()).
					Return(false, nil)
				testHelper.mockFlag.EXPECT().
					IsEnabled(models.FlagGetOpeningBalanceFromPreviousMonth.String()).
					Return(true)
				testHelper.mockBigQuery.EXPECT().
					QueryGenerateTrialBalanceDetail(ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil)
				testHelper.mockBigQuery.EXPECT().
					QueryGenerateTrialBalanceSummary(ctx, gomock.Any()).
					Return(nil)
				testHelper.mockBigQuery.EXPECT().
					ExportTrialBalanceSummary(ctx, gomock.Any(), gomock.Any(), gomock.Any()).
					Return([]models.CreateTrialBalancePeriod{}, nil)
				testHelper.mockSubCategoryRepository.EXPECT().
					GetAll(gomock.AssignableToTypeOf(context.Background()), models.GetAllSubCategoryParam{}).
					Return(&[]models.SubCategory{{
						Code: "13101",
					}}, nil)
				testHelper.mockBigQuery.EXPECT().
					ExportTrialBalanceDetail(ctx, gomock.Any(), gomock.Any(), gomock.Any(), &[]models.SubCategory{{
						Code: "13101",
					}}).
					Return(assert.AnError)
			},
			wantErr: true,
		},
		{
			name: "error case - bulk insert",
			req:  date,
			doMock: func(ctx context.Context, req time.Time) {
				testHelper.mockBigQuery.EXPECT().
					TableExists(ctx, gomock.Any()).
					Return(false, nil)
				testHelper.mockFlag.EXPECT().
					IsEnabled(models.FlagGetOpeningBalanceFromPreviousMonth.String()).
					Return(true)
				testHelper.mockBigQuery.EXPECT().
					QueryGenerateTrialBalanceDetail(ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil)
				testHelper.mockBigQuery.EXPECT().
					QueryGenerateTrialBalanceSummary(ctx, gomock.Any()).
					Return(nil)
				testHelper.mockBigQuery.EXPECT().
					ExportTrialBalanceSummary(ctx, gomock.Any(), gomock.Any(), gomock.Any()).
					Return([]models.CreateTrialBalancePeriod{}, nil)
				testHelper.mockSubCategoryRepository.EXPECT().
					GetAll(gomock.AssignableToTypeOf(context.Background()), models.GetAllSubCategoryParam{}).
					Return(&[]models.SubCategory{{
						Code: "13101",
					}}, nil)
				testHelper.mockBigQuery.EXPECT().
					ExportTrialBalanceDetail(ctx, gomock.Any(), gomock.Any(), gomock.Any(), &[]models.SubCategory{{
						Code: "13101",
					}}).
					Return(nil)
				testHelper.mockTrialBalanceRepository.EXPECT().
					BulkInsert(ctx, gomock.Any()).
					Return(assert.AnError)
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
			err := testHelper.accountingService.GenerateTrialBalanceBigQuery(ctx, tt.req, false)
			assert.Equal(t, tt.wantErr, err != nil)
		})
	}
}

func Test_accounting_GenerateAdjustmentTrialBalanceBigQuery(t *testing.T) {
	testHelper := serviceTestHelper(t)
	ctx := context.Background()
	ctx2 := context.WithoutCancel(ctx)
	date, _ := atime.NowZeroTime()
	req := models.AdjustmentTrialBalanceFilter{
		AdjustmentDate: date,
		IsManual:       false,
	}

	tests := []struct {
		name    string
		req     models.AdjustmentTrialBalanceFilter
		doMock  func(ctx context.Context, req models.AdjustmentTrialBalanceFilter)
		wantErr bool
	}{
		{
			name: "success case - is manual false",
			req:  req,
			doMock: func(ctx context.Context, req models.AdjustmentTrialBalanceFilter) {
				testHelper.mockAcctRepository.EXPECT().
					GetTransactionsToday(ctx, gomock.Any()).
					Return([]string{""}, nil)
				testHelper.mockBigQuery.EXPECT().
					QueryGetTransactions(ctx, gomock.Any()).
					Return([]string{""}, nil)

				testHelper.mockTrialBalanceRepository.EXPECT().
					GetFirstPeriodByStatus(ctx, gomock.Any()).
					Return(&models.TrialBalancePeriod{
						Period: "2025-01",
					}, nil)
				testHelper.mockTrialBalanceRepository.EXPECT().
					UpdateTrialBalanceAdjustment(ctx, models.CloseTrialBalanceRequest{
						Period: "2025-01",
					}).
					Return(nil)

				//
				testHelper.mockTrialBalanceRepository.EXPECT().
					GetByPeriodStatus(ctx2, gomock.Any(), gomock.Any()).
					Return([]models.TrialBalancePeriod{
						{},
					}, nil)
				testHelper.mockFlag.EXPECT().
					IsEnabled(models.FlagGetOpeningBalanceFromPreviousMonth.String()).
					Return(true)
				testHelper.mockBigQuery.EXPECT().
					QueryGenerateTrialBalanceDetail(ctx2, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil)
				testHelper.mockBigQuery.EXPECT().
					QueryGenerateTrialBalanceSummary(ctx2, gomock.Any()).
					Return(nil)
				testHelper.mockBigQuery.EXPECT().
					ExportTrialBalanceSummary(ctx2, gomock.Any(), gomock.Any(), gomock.Any()).
					Return([]models.CreateTrialBalancePeriod{}, nil)
				testHelper.mockSubCategoryRepository.EXPECT().
					GetAll(gomock.AssignableToTypeOf(ctx2), models.GetAllSubCategoryParam{}).
					Return(&[]models.SubCategory{{
						Code: "13101",
					}}, nil)
				testHelper.mockBigQuery.EXPECT().
					ExportTrialBalanceDetail(ctx2, gomock.Any(), gomock.Any(), gomock.Any(), &[]models.SubCategory{{
						Code: "13101",
					}}).
					Return(nil)
			},
			wantErr: false,
		},
		{
			name: "error case - GetTransactionsToday",
			req:  req,
			doMock: func(ctx context.Context, req models.AdjustmentTrialBalanceFilter) {
				testHelper.mockAcctRepository.EXPECT().
					GetTransactionsToday(ctx, gomock.Any()).
					Return(nil, assert.AnError)
			},
			wantErr: true,
		},
		{
			name: "error case - QueryGetTransactions",
			req:  req,
			doMock: func(ctx context.Context, req models.AdjustmentTrialBalanceFilter) {
				testHelper.mockAcctRepository.EXPECT().
					GetTransactionsToday(ctx, gomock.Any()).
					Return([]string{""}, nil)
				testHelper.mockBigQuery.EXPECT().
					QueryGetTransactions(ctx, gomock.Any()).
					Return(nil, assert.AnError)
			},
			wantErr: true,
		},
		{
			name: "error case - transaction not sync with bq",
			req:  req,
			doMock: func(ctx context.Context, req models.AdjustmentTrialBalanceFilter) {
				testHelper.mockAcctRepository.EXPECT().
					GetTransactionsToday(ctx, gomock.Any()).
					Return([]string{"", ""}, nil)
				testHelper.mockBigQuery.EXPECT().
					QueryGetTransactions(ctx, gomock.Any()).
					Return([]string{""}, nil)
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
			err := testHelper.accountingService.GenerateAdjustmentTrialBalanceBigQuery(ctx, tt.req)
			assert.Equal(t, tt.wantErr, err != nil)
		})
	}
}
