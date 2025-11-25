package services_test

import (
	"context"
	"testing"

	"bitbucket.org/Amartha/go-accounting/internal/models"
	"bitbucket.org/Amartha/go-accounting/internal/pkg/atime"
	"go.uber.org/mock/gomock"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func Test_accounting_GetBalanceSheet(t *testing.T) {
	testHelper := serviceTestHelper(t)
	defaultTime := atime.Now()
	defaultBalanceSheet := make(map[string][]models.BalanceCategory)
	defaultTotalPerCOAType := make(map[string]decimal.Decimal)
	defaultData := models.BalanceSheetOut{
		BalanceSheet:    defaultBalanceSheet,
		TotalPerCOAType: defaultTotalPerCOAType,
	}

	type args struct {
		ctx context.Context
		req models.BalanceSheetFilterOptions
	}
	testCases := []struct {
		name    string
		args    args
		doMock  func(args args)
		wantErr bool
	}{
		{
			name: "success case - get balance sheet",
			args: args{
				ctx: context.Background(),
				req: models.BalanceSheetFilterOptions{
					EntityCode:       "111",
					BalanceSheetDate: defaultTime,
				},
			},
			doMock: func(args args) {
				testHelper.mockEntityRepository.EXPECT().GetByCode(args.ctx, args.req.EntityCode).Return(&models.Entity{}, nil)
				testHelper.mockAcctRepository.EXPECT().GetBalanceSheet(args.ctx, args.req).Return(defaultData, nil)
			},
			wantErr: false,
		},
		{
			name: "error case - get balance sheet error",
			args: args{
				ctx: context.Background(),
				req: models.BalanceSheetFilterOptions{
					EntityCode:       "111",
					BalanceSheetDate: defaultTime,
				},
			},
			doMock: func(args args) {
				testHelper.mockEntityRepository.EXPECT().GetByCode(args.ctx, args.req.EntityCode).Return(&models.Entity{}, nil)
				testHelper.mockAcctRepository.EXPECT().GetBalanceSheet(args.ctx, args.req).Return(defaultData, assert.AnError)
			},
			wantErr: true,
		},
		{
			name: "error case - get entity code not found",
			args: args{
				ctx: context.Background(),
				req: models.BalanceSheetFilterOptions{
					EntityCode:       "123",
					BalanceSheetDate: defaultTime,
				},
			},
			doMock: func(args args) {
				testHelper.mockEntityRepository.EXPECT().GetByCode(args.ctx, args.req.EntityCode).Return(nil, nil)
			},
			wantErr: true,
		},
		{
			name: "error case - get entity code error",
			args: args{
				ctx: context.Background(),
				req: models.BalanceSheetFilterOptions{
					EntityCode:       "123",
					BalanceSheetDate: defaultTime,
				},
			},
			doMock: func(args args) {
				testHelper.mockEntityRepository.EXPECT().GetByCode(args.ctx, args.req.EntityCode).Return(nil, assert.AnError)
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.doMock != nil {
				tc.doMock(tc.args)
			}

			_, err := testHelper.accountingService.GetBalanceSheet(tc.args.ctx, tc.args.req)
			assert.Equal(t, tc.wantErr, err != nil)
		})
	}
}

func Test_accounting_DownloadCSVGetBalanceSheet(t *testing.T) {
	testHelper := serviceTestHelper(t)
	res := models.GetBalanceSheetResponse{
		EntityCode:       "001",
		BalanceSheetDate: "2024-01-01",
		BalanceSheet: models.BalanceSheetData{
			Assets: []models.BalanceCategory{
				{
					CategoryCode: "121",
					CategoryName: "Cash Point",
					Amount:       "0,00",
				},
			},
			Liabilities: []models.BalanceCategory{
				{
					CategoryCode: "211",
					CategoryName: "Cash Point",
					Amount:       "0,00",
				},
			},
			TotalAsset:     "0,00",
			TotalLiability: "0,00",
			CatchAll:       "0,00",
		},
	}
	type args struct {
		opts models.BalanceSheetFilterOptions
		resp models.GetBalanceSheetResponse
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
				opts: models.BalanceSheetFilterOptions{},
				resp: res,
			},
			doMock: func(args args) {
				testHelper.mockFile.EXPECT().NewCSVWriter(gomock.Any())
				testHelper.mockFile.EXPECT().CSVWriteBody(gomock.Any(), gomock.Any()).AnyTimes()
				testHelper.mockFile.EXPECT().CSVProcessWrite(gomock.Any())
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
			_, _, err := testHelper.accountingService.DownloadCSVGetBalanceSheet(context.TODO(), tt.args.opts, tt.args.resp)
			assert.Equal(t, tt.wantErr, err != nil)
		})
	}
}
