package services_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"bitbucket.org/Amartha/go-accounting/internal/models"
	"bitbucket.org/Amartha/go-accounting/internal/pkg/atime"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func dummyResponse() (map[string][]models.TBCOACategory, map[models.TBCOACategory][]models.TBSubCategory) {
	result := []models.GetTrialBalanceOut{
		{
			CoaTypeCode:     "000",
			CoaTypeName:     "asset",
			CategoryCode:    "A.111",
			CategoryName:    "Kas Teller",
			SubCategoryCode: "A.111.01",
			SubCategoryName: "Kas Teller Point",
			OpeningBalance:  decimal.NewFromFloat(100000),
			DebitMovement:   decimal.NewFromFloat(20000),
			CreditMovement:  decimal.NewFromFloat(-20000),
			ClosingBalance:  decimal.NewFromFloat(140000),
		},
		{
			CoaTypeCode:     "000",
			CoaTypeName:     "asset",
			CategoryCode:    "A.111",
			CategoryName:    "Kas Teller",
			SubCategoryCode: "A.111.02",
			SubCategoryName: "Kas Teller Point",
			OpeningBalance:  decimal.NewFromFloat(100000),
			DebitMovement:   decimal.NewFromFloat(20000),
			CreditMovement:  decimal.NewFromFloat(-20000),
			ClosingBalance:  decimal.NewFromFloat(340000),
		},
		{
			CoaTypeCode:     "000",
			CoaTypeName:     "asset",
			CategoryCode:    "A.112",
			CategoryName:    "Kas Teller",
			SubCategoryCode: "A.112.01",
			SubCategoryName: "Kas Teller Point",
			OpeningBalance:  decimal.NewFromFloat(200000),
			DebitMovement:   decimal.NewFromFloat(20000),
			CreditMovement:  decimal.NewFromFloat(-20000),
			ClosingBalance:  decimal.NewFromFloat(240000),
		},
		{
			CoaTypeCode:     "001",
			CoaTypeName:     "liability",
			CategoryCode:    "B.211",
			CategoryName:    "Marketplace Payable",
			SubCategoryCode: "B.211.01",
			SubCategoryName: "Lender Balance - Individual Non RDL",
			OpeningBalance:  decimal.NewFromFloat(100000),
			DebitMovement:   decimal.NewFromFloat(20000),
			CreditMovement:  decimal.NewFromFloat(20000),
			ClosingBalance:  decimal.NewFromFloat(140000),
		},
	}

	resC := make(map[string][]models.TBCOACategory)
	resSC := make(map[models.TBCOACategory][]models.TBSubCategory)

	for _, v := range result {
		coaType := strings.ToLower(v.CoaTypeName)

		resC[v.CategoryCode] = append(resC[v.CategoryCode], models.TBCOACategory{
			Type:                v.CoaTypeName,
			CategoryCode:        v.CategoryCode,
			CategoryName:        v.CategoryName,
			TotalOpeningBalance: v.OpeningBalance,
			TotalDebitMovement:  v.DebitMovement,
			TotalCreditMovement: v.CreditMovement,
			TotalClosingBalance: v.ClosingBalance,
		})

		key := models.TBCOACategory{
			Type:         coaType,
			CoaTypeCode:  v.CoaTypeCode,
			CoaTypeName:  v.CoaTypeName,
			CategoryCode: v.CategoryCode,
			CategoryName: v.CategoryName,
		}
		resSC[key] = append(resSC[key], models.TBSubCategory{
			Kind:            models.KindSubCategory,
			SubCategoryCode: v.SubCategoryCode,
			SubCategoryName: v.SubCategoryName,
			OpeningBalance:  v.OpeningBalance,
			DebitMovement:   v.DebitMovement,
			CreditMovement:  v.CreditMovement,
			ClosingBalance:  v.ClosingBalance,
		})
	}

	return resC, resSC
}

func Test_accounting_GetTrialBalance(t *testing.T) {
	testHelper := serviceTestHelper(t)
	resC, resSC := dummyResponse()

	type args struct {
		ctx context.Context
		req models.TrialBalanceFilterOptions
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
			name: "success - Old Flow",
			args: args{
				ctx: context.Background(),
				req: models.TrialBalanceFilterOptions{},
			},
			doMock: func(args args, mockData mockData) {
				testHelper.mockFlag.EXPECT().IsEnabled(models.FlagGetTrialBalanceGCS.String()).Return(false)
				testHelper.mockFlag.EXPECT().IsEnabled(models.FlagGetTrialBalanceV2.String()).Return(false)
				testHelper.mockEntityRepository.EXPECT().GetByCode(args.ctx, args.req.EntityCode).Return(&models.Entity{}, nil)
				testHelper.mockAcctRepository.EXPECT().GetTrialBalance(gomock.Any(), gomock.Any()).Return(resC, resSC, nil)
			},
			wantErr: false,
		},
		{
			name: "success - New Flow",
			args: args{
				ctx: context.Background(),
				req: models.TrialBalanceFilterOptions{
					EntityCode: "AMF",
					Period:     time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
				},
			},
			doMock: func(args args, mockData mockData) {
				testHelper.mockFlag.EXPECT().
					IsEnabled(models.FlagGetTrialBalanceGCS.String()).
					Return(true)
				testHelper.mockFlag.EXPECT().
					IsEnabled(models.FlagGetTrialBalanceV2.String()).
					Return(false)

				testHelper.mockEntityRepository.EXPECT().
					GetByCode(args.ctx, args.req.EntityCode).
					Return(&models.Entity{
						Code: "AMF",
						Name: "Amartha Finance",
					}, nil)

				testHelper.mockTrialBalanceRepository.EXPECT().
					GetByPeriod(args.ctx, args.req.Period.Format(atime.DateFormatYYYYMM), args.req.EntityCode).
					Return(&models.TrialBalancePeriod{
						Status: "OPEN",
					}, nil)

				csvContent := `account_code,account_name,debit,credit
1001,Cash,1000,0
2001,Payable,0,1000`
				reader := io.NopCloser(bytes.NewReader([]byte(csvContent)))

				testHelper.mockCloudStorageRepository.EXPECT().
					NewReader(args.ctx, gomock.AssignableToTypeOf(&models.CloudStoragePayload{})).
					Return(reader, nil)

				testHelper.mockMySQLRepository.EXPECT().
					GetAllCategorySubCategoryCOAType(args.ctx).
					Return(nil, nil, map[string]models.CategorySubCategoryCOAType{
						"1001": {
							CoaTypeCode:     "ASSET",
							CoaTypeName:     "Asset",
							CategoryName:    "Cash",
							SubCategoryName: "Cash Account",
						},
						"2001": {
							CoaTypeCode:     "LIAB",
							CoaTypeName:     "Liability",
							CategoryName:    "Payable",
							SubCategoryName: "Accounts Payable",
						},
					}, nil)
			},
			wantErr: false,
		},
		{
			name: "error case - GetTrialBalance",
			args: args{
				ctx: context.Background(),
				req: models.TrialBalanceFilterOptions{},
			},
			doMock: func(args args, mockData mockData) {
				testHelper.mockFlag.EXPECT().IsEnabled(models.FlagGetTrialBalanceGCS.String()).Return(false)
				testHelper.mockFlag.EXPECT().IsEnabled(models.FlagGetTrialBalanceV2.String()).Return(false)
				testHelper.mockEntityRepository.EXPECT().GetByCode(args.ctx, args.req.EntityCode).Return(&models.Entity{}, nil)
				testHelper.mockAcctRepository.EXPECT().GetTrialBalance(gomock.Any(), gomock.Any()).Return(nil, nil, assert.AnError)
			},
			wantErr: true,
		},
		{
			name: "error case - get entity - database error",
			args: args{
				ctx: context.Background(),
				req: models.TrialBalanceFilterOptions{},
			},
			doMock: func(args args, mockData mockData) {
				testHelper.mockFlag.EXPECT().IsEnabled(models.FlagGetTrialBalanceV2.String()).Return(false)
				testHelper.mockEntityRepository.EXPECT().GetByCode(args.ctx, args.req.EntityCode).Return(nil, assert.AnError)
			},
			wantErr: true,
		},
		{
			name: "error case - get entity - data not found",
			args: args{
				ctx: context.Background(),
				req: models.TrialBalanceFilterOptions{},
			},
			doMock: func(args args, mockData mockData) {
				testHelper.mockFlag.EXPECT().IsEnabled(models.FlagGetTrialBalanceV2.String()).Return(false)
				testHelper.mockEntityRepository.EXPECT().GetByCode(args.ctx, args.req.EntityCode).Return(nil, nil)
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

			_, err := testHelper.accountingService.GetTrialBalance(tt.args.ctx, tt.args.req)
			assert.Equal(t, tt.wantErr, err != nil)
		})
	}
}

func Test_accounting_DownloadCSVGetTrialBalance(t *testing.T) {
	testHelper := serviceTestHelper(t)
	resC, resSC := dummyResponse()
	type args struct {
		ctx context.Context
		req models.TrialBalanceFilterOptions
	}
	tests := []struct {
		name    string
		args    args
		doMock  func(args args)
		wantErr bool
	}{
		{
			name: "error case - GetTrialBalance",
			args: args{
				ctx: context.Background(),
				req: models.TrialBalanceFilterOptions{},
			},
			doMock: func(args args) {
				testHelper.mockFlag.EXPECT().IsEnabled(models.FlagGetTrialBalanceGCS.String()).Return(false)
				testHelper.mockFlag.EXPECT().IsEnabled(models.FlagGetTrialBalanceV2.String()).Return(false)
				testHelper.mockEntityRepository.EXPECT().GetByCode(args.ctx, args.req.EntityCode).Return(&models.Entity{}, nil)
				testHelper.mockAcctRepository.EXPECT().GetTrialBalance(gomock.Any(), gomock.Any()).Return(nil, nil, assert.AnError)
			},
			wantErr: true,
		},
		{
			name: "error case - get entity - data not found",
			args: args{
				ctx: context.Background(),
				req: models.TrialBalanceFilterOptions{},
			},
			doMock: func(args args) {
				testHelper.mockFlag.EXPECT().IsEnabled(models.FlagGetTrialBalanceGCS.String()).Return(false)
				testHelper.mockFlag.EXPECT().IsEnabled(models.FlagGetTrialBalanceV2.String()).Return(false)
				testHelper.mockEntityRepository.EXPECT().GetByCode(args.ctx, args.req.EntityCode).Return(&models.Entity{}, nil)
				testHelper.mockAcctRepository.EXPECT().GetTrialBalance(gomock.Any(), gomock.Any()).Return(nil, nil, nil)

				testHelper.mockEntityRepository.EXPECT().GetByCode(args.ctx, args.req.EntityCode).Return(nil, nil)
			},
			wantErr: true,
		},
		{
			name: "error case - get entity - database error",
			args: args{
				ctx: context.Background(),
				req: models.TrialBalanceFilterOptions{},
			},
			doMock: func(args args) {
				testHelper.mockFlag.EXPECT().IsEnabled(models.FlagGetTrialBalanceGCS.String()).Return(false)
				testHelper.mockFlag.EXPECT().IsEnabled(models.FlagGetTrialBalanceV2.String()).Return(false)
				testHelper.mockEntityRepository.EXPECT().GetByCode(args.ctx, args.req.EntityCode).Return(&models.Entity{}, nil)
				testHelper.mockAcctRepository.EXPECT().GetTrialBalance(gomock.Any(), gomock.Any()).Return(nil, nil, nil)

				testHelper.mockEntityRepository.EXPECT().GetByCode(args.ctx, args.req.EntityCode).Return(nil, assert.AnError)
			},
			wantErr: true,
		},
		{
			name: "error case - create csv write title 1",
			args: args{
				ctx: context.Background(),
				req: models.TrialBalanceFilterOptions{},
			},
			doMock: func(args args) {
				testHelper.mockFlag.EXPECT().IsEnabled(models.FlagGetTrialBalanceGCS.String()).Return(false)
				testHelper.mockFlag.EXPECT().IsEnabled(models.FlagGetTrialBalanceV2.String()).Return(false)
				testHelper.mockEntityRepository.EXPECT().GetByCode(args.ctx, args.req.EntityCode).Return(&models.Entity{}, nil)
				testHelper.mockAcctRepository.EXPECT().GetTrialBalance(gomock.Any(), gomock.Any()).Return(resC, resSC, nil)

				testHelper.mockEntityRepository.EXPECT().GetByCode(args.ctx, args.req.EntityCode).Return(&models.Entity{}, nil)
				testHelper.mockFile.EXPECT().NewCSVWriter(gomock.Any())
				testHelper.mockFile.EXPECT().CSVWriteBody(args.ctx, gomock.Any()).Return(assert.AnError)
			},
			wantErr: true,
		},
		{
			name: "error case - create csv write title 2",
			args: args{
				ctx: context.Background(),
				req: models.TrialBalanceFilterOptions{},
			},
			doMock: func(args args) {
				testHelper.mockFlag.EXPECT().IsEnabled(models.FlagGetTrialBalanceGCS.String()).Return(false)
				testHelper.mockFlag.EXPECT().IsEnabled(models.FlagGetTrialBalanceV2.String()).Return(false)
				testHelper.mockEntityRepository.EXPECT().GetByCode(args.ctx, args.req.EntityCode).Return(&models.Entity{}, nil)
				testHelper.mockAcctRepository.EXPECT().GetTrialBalance(gomock.Any(), gomock.Any()).Return(resC, resSC, nil)

				testHelper.mockEntityRepository.EXPECT().GetByCode(args.ctx, args.req.EntityCode).Return(&models.Entity{}, nil)
				testHelper.mockFile.EXPECT().NewCSVWriter(gomock.Any())
				successCSVWriteBody := testHelper.mockFile.EXPECT().CSVWriteBody(args.ctx, gomock.Any()).Return(nil)
				errorCSVWriteBody := testHelper.mockFile.EXPECT().CSVWriteBody(args.ctx, gomock.Any()).Return(assert.AnError)
				gomock.InOrder(
					successCSVWriteBody,
					errorCSVWriteBody,
				)
			},
			wantErr: true,
		},
		{
			name: "error case - create csv write title 3",
			args: args{
				ctx: context.Background(),
				req: models.TrialBalanceFilterOptions{},
			},
			doMock: func(args args) {
				testHelper.mockFlag.EXPECT().IsEnabled(models.FlagGetTrialBalanceGCS.String()).Return(false)
				testHelper.mockFlag.EXPECT().IsEnabled(models.FlagGetTrialBalanceV2.String()).Return(false)
				testHelper.mockEntityRepository.EXPECT().GetByCode(args.ctx, args.req.EntityCode).Return(&models.Entity{}, nil)
				testHelper.mockAcctRepository.EXPECT().GetTrialBalance(gomock.Any(), gomock.Any()).Return(resC, resSC, nil)

				testHelper.mockEntityRepository.EXPECT().GetByCode(args.ctx, args.req.EntityCode).Return(&models.Entity{}, nil)
				testHelper.mockFile.EXPECT().NewCSVWriter(gomock.Any())
				successCSVWriteBody := testHelper.mockFile.EXPECT().CSVWriteBody(args.ctx, gomock.Any()).Return(nil).MaxTimes(2)
				errorCSVWriteBody := testHelper.mockFile.EXPECT().CSVWriteBody(args.ctx, gomock.Any()).Return(assert.AnError)
				gomock.InOrder(
					successCSVWriteBody,
					errorCSVWriteBody,
				)
			},
			wantErr: true,
		},
		{
			name: "error case - create csv write title 4",
			args: args{
				ctx: context.Background(),
				req: models.TrialBalanceFilterOptions{},
			},
			doMock: func(args args) {
				testHelper.mockFlag.EXPECT().IsEnabled(models.FlagGetTrialBalanceGCS.String()).Return(false)
				testHelper.mockFlag.EXPECT().IsEnabled(models.FlagGetTrialBalanceV2.String()).Return(false)
				testHelper.mockEntityRepository.EXPECT().GetByCode(args.ctx, args.req.EntityCode).Return(&models.Entity{}, nil)
				testHelper.mockAcctRepository.EXPECT().GetTrialBalance(gomock.Any(), gomock.Any()).Return(resC, resSC, nil)

				testHelper.mockEntityRepository.EXPECT().GetByCode(args.ctx, args.req.EntityCode).Return(&models.Entity{}, nil)
				testHelper.mockFile.EXPECT().NewCSVWriter(gomock.Any())
				successCSVWriteBody := testHelper.mockFile.EXPECT().CSVWriteBody(args.ctx, gomock.Any()).Return(nil).MaxTimes(3)
				errorCSVWriteBody := testHelper.mockFile.EXPECT().CSVWriteBody(args.ctx, gomock.Any()).Return(assert.AnError)
				gomock.InOrder(
					successCSVWriteBody,
					errorCSVWriteBody,
				)
			},
			wantErr: true,
		},
		{
			name: "error case - create csv write header",
			args: args{
				ctx: context.Background(),
				req: models.TrialBalanceFilterOptions{},
			},
			doMock: func(args args) {
				testHelper.mockFlag.EXPECT().IsEnabled(models.FlagGetTrialBalanceGCS.String()).Return(false)
				testHelper.mockFlag.EXPECT().IsEnabled(models.FlagGetTrialBalanceV2.String()).Return(false)
				testHelper.mockEntityRepository.EXPECT().GetByCode(args.ctx, args.req.EntityCode).Return(&models.Entity{}, nil)
				testHelper.mockAcctRepository.EXPECT().GetTrialBalance(gomock.Any(), gomock.Any()).Return(resC, resSC, nil)

				testHelper.mockEntityRepository.EXPECT().GetByCode(args.ctx, args.req.EntityCode).Return(&models.Entity{}, nil)
				testHelper.mockFile.EXPECT().NewCSVWriter(gomock.Any())
				testHelper.mockFile.EXPECT().CSVWriteBody(args.ctx, gomock.Any()).Return(nil).MaxTimes(4)
				testHelper.mockFile.EXPECT().CSVWriteHeader(args.ctx, gomock.Any()).Return(assert.AnError)
			},
			wantErr: true,
		},
		{
			name: "error case - create csv write body 1",
			args: args{
				ctx: context.Background(),
				req: models.TrialBalanceFilterOptions{},
			},
			doMock: func(args args) {
				testHelper.mockFlag.EXPECT().IsEnabled(models.FlagGetTrialBalanceGCS.String()).Return(false)
				testHelper.mockFlag.EXPECT().IsEnabled(models.FlagGetTrialBalanceV2.String()).Return(false)
				testHelper.mockEntityRepository.EXPECT().GetByCode(args.ctx, args.req.EntityCode).Return(&models.Entity{}, nil)
				testHelper.mockAcctRepository.EXPECT().GetTrialBalance(gomock.Any(), gomock.Any()).Return(resC, resSC, nil)

				testHelper.mockEntityRepository.EXPECT().GetByCode(args.ctx, gomock.Any()).Return(&models.Entity{}, nil)
				testHelper.mockFile.EXPECT().NewCSVWriter(gomock.Any())
				testHelper.mockFile.EXPECT().CSVWriteAll(args.ctx, gomock.Any()).Return(nil)
				testHelper.mockFile.EXPECT().CSVWriteHeader(args.ctx, gomock.Any()).Return(nil)
				successCSVWriteBody := testHelper.mockFile.EXPECT().CSVWriteBody(args.ctx, gomock.Any()).Return(nil).MaxTimes(4)
				errorCSVWriteBody := testHelper.mockFile.EXPECT().CSVWriteBody(args.ctx, gomock.Any()).Return(assert.AnError)
				gomock.InOrder(
					successCSVWriteBody,
					errorCSVWriteBody,
				)
			},
			wantErr: true,
		},
		{
			name: "error case - create csv write body 2",
			args: args{
				ctx: context.Background(),
				req: models.TrialBalanceFilterOptions{},
			},
			doMock: func(args args) {
				testHelper.mockFlag.EXPECT().IsEnabled(models.FlagGetTrialBalanceGCS.String()).Return(false)
				testHelper.mockFlag.EXPECT().IsEnabled(models.FlagGetTrialBalanceV2.String()).Return(false)
				testHelper.mockEntityRepository.EXPECT().GetByCode(args.ctx, args.req.EntityCode).Return(&models.Entity{}, nil)
				testHelper.mockAcctRepository.EXPECT().GetTrialBalance(gomock.Any(), gomock.Any()).Return(resC, resSC, nil)

				testHelper.mockEntityRepository.EXPECT().GetByCode(args.ctx, gomock.Any()).Return(&models.Entity{}, nil)
				testHelper.mockFile.EXPECT().NewCSVWriter(gomock.Any())
				testHelper.mockFile.EXPECT().CSVWriteAll(args.ctx, gomock.Any()).Return(nil)
				testHelper.mockFile.EXPECT().CSVWriteHeader(args.ctx, gomock.Any()).Return(nil)

				successCSVWriteBody := testHelper.mockFile.EXPECT().CSVWriteBody(args.ctx, gomock.Any()).Return(nil).MaxTimes(5)
				errorCSVWriteBody := testHelper.mockFile.EXPECT().CSVWriteBody(args.ctx, gomock.Any()).Return(assert.AnError)
				gomock.InOrder(
					successCSVWriteBody,
					errorCSVWriteBody,
				)
			},
			wantErr: true,
		},
		{
			name: "error case - create csv write body 3",
			args: args{
				ctx: context.Background(),
				req: models.TrialBalanceFilterOptions{},
			},
			doMock: func(args args) {
				testHelper.mockFlag.EXPECT().IsEnabled(models.FlagGetTrialBalanceGCS.String()).Return(false)
				testHelper.mockFlag.EXPECT().IsEnabled(models.FlagGetTrialBalanceV2.String()).Return(false)
				testHelper.mockEntityRepository.EXPECT().GetByCode(args.ctx, args.req.EntityCode).Return(&models.Entity{}, nil)
				testHelper.mockAcctRepository.EXPECT().GetTrialBalance(gomock.Any(), gomock.Any()).Return(resC, resSC, nil)

				testHelper.mockEntityRepository.EXPECT().GetByCode(args.ctx, gomock.Any()).Return(&models.Entity{}, nil)
				testHelper.mockFile.EXPECT().NewCSVWriter(gomock.Any())
				testHelper.mockFile.EXPECT().CSVWriteAll(args.ctx, gomock.Any()).Return(nil)
				testHelper.mockFile.EXPECT().CSVWriteHeader(args.ctx, gomock.Any()).Return(nil)

				successCSVWriteBody := testHelper.mockFile.EXPECT().CSVWriteBody(args.ctx, gomock.Any()).Return(nil).MaxTimes(6)
				errorCSVWriteBody := testHelper.mockFile.EXPECT().CSVWriteBody(args.ctx, gomock.Any()).Return(assert.AnError)
				gomock.InOrder(
					successCSVWriteBody,
					errorCSVWriteBody,
				)
			},
			wantErr: true,
		},
		{
			name: "error case - create csv write body 4",
			args: args{
				ctx: context.Background(),
				req: models.TrialBalanceFilterOptions{},
			},
			doMock: func(args args) {
				testHelper.mockFlag.EXPECT().IsEnabled(models.FlagGetTrialBalanceGCS.String()).Return(false)
				testHelper.mockFlag.EXPECT().IsEnabled(models.FlagGetTrialBalanceV2.String()).Return(false)
				testHelper.mockEntityRepository.EXPECT().GetByCode(args.ctx, args.req.EntityCode).Return(&models.Entity{}, nil)
				testHelper.mockAcctRepository.EXPECT().GetTrialBalance(gomock.Any(), gomock.Any()).Return(resC, resSC, nil)

				testHelper.mockEntityRepository.EXPECT().GetByCode(args.ctx, gomock.Any()).Return(&models.Entity{}, nil)
				testHelper.mockFile.EXPECT().NewCSVWriter(gomock.Any())
				testHelper.mockFile.EXPECT().CSVWriteAll(args.ctx, gomock.Any()).Return(nil)
				testHelper.mockFile.EXPECT().CSVWriteHeader(args.ctx, gomock.Any()).Return(nil)

				successCSVWriteBody := testHelper.mockFile.EXPECT().CSVWriteBody(args.ctx, gomock.Any()).Return(nil).MaxTimes(8)
				errorCSVWriteBody := testHelper.mockFile.EXPECT().CSVWriteBody(args.ctx, gomock.Any()).Return(assert.AnError)
				gomock.InOrder(
					successCSVWriteBody,
					errorCSVWriteBody,
				)
			},
			wantErr: true,
		},
		{
			name: "error case - create csv write body 5",
			args: args{
				ctx: context.Background(),
				req: models.TrialBalanceFilterOptions{},
			},
			doMock: func(args args) {
				testHelper.mockFlag.EXPECT().IsEnabled(models.FlagGetTrialBalanceGCS.String()).Return(false)
				testHelper.mockFlag.EXPECT().IsEnabled(models.FlagGetTrialBalanceV2.String()).Return(false)
				testHelper.mockEntityRepository.EXPECT().GetByCode(args.ctx, args.req.EntityCode).Return(&models.Entity{}, nil)
				testHelper.mockAcctRepository.EXPECT().GetTrialBalance(gomock.Any(), gomock.Any()).Return(resC, resSC, nil)

				testHelper.mockEntityRepository.EXPECT().GetByCode(args.ctx, gomock.Any()).Return(&models.Entity{}, nil)
				testHelper.mockFile.EXPECT().NewCSVWriter(gomock.Any())
				testHelper.mockFile.EXPECT().CSVWriteAll(args.ctx, gomock.Any()).Return(nil)
				testHelper.mockFile.EXPECT().CSVWriteHeader(args.ctx, gomock.Any()).Return(nil)

				successCSVWriteBody := testHelper.mockFile.EXPECT().CSVWriteBody(args.ctx, gomock.Any()).Return(nil).MaxTimes(10)
				errorCSVWriteBody := testHelper.mockFile.EXPECT().CSVWriteBody(args.ctx, gomock.Any()).Return(assert.AnError)
				gomock.InOrder(
					successCSVWriteBody,
					errorCSVWriteBody,
				)
			},
			wantErr: true,
		},
		{
			name: "error case - create csv write body 6",
			args: args{
				ctx: context.Background(),
				req: models.TrialBalanceFilterOptions{},
			},
			doMock: func(args args) {
				testHelper.mockFlag.EXPECT().IsEnabled(models.FlagGetTrialBalanceGCS.String()).Return(false)
				testHelper.mockFlag.EXPECT().IsEnabled(models.FlagGetTrialBalanceV2.String()).Return(false)
				testHelper.mockEntityRepository.EXPECT().GetByCode(args.ctx, args.req.EntityCode).Return(&models.Entity{}, nil)
				testHelper.mockAcctRepository.EXPECT().GetTrialBalance(gomock.Any(), gomock.Any()).Return(resC, resSC, nil)

				testHelper.mockEntityRepository.EXPECT().GetByCode(args.ctx, gomock.Any()).Return(&models.Entity{}, nil)
				testHelper.mockFile.EXPECT().NewCSVWriter(gomock.Any())
				testHelper.mockFile.EXPECT().CSVWriteAll(args.ctx, gomock.Any()).Return(nil)
				testHelper.mockFile.EXPECT().CSVWriteHeader(args.ctx, gomock.Any()).Return(nil)

				successCSVWriteBody := testHelper.mockFile.EXPECT().CSVWriteBody(args.ctx, gomock.Any()).Return(nil).MaxTimes(11)
				errorCSVWriteBody := testHelper.mockFile.EXPECT().CSVWriteBody(args.ctx, gomock.Any()).Return(assert.AnError)
				gomock.InOrder(
					successCSVWriteBody,
					errorCSVWriteBody,
				)
			},
			wantErr: true,
		},
		{
			name: "error case - create csv process write",
			args: args{
				ctx: context.Background(),
				req: models.TrialBalanceFilterOptions{},
			},
			doMock: func(args args) {
				testHelper.mockFlag.EXPECT().IsEnabled(models.FlagGetTrialBalanceGCS.String()).Return(false)
				testHelper.mockFlag.EXPECT().IsEnabled(models.FlagGetTrialBalanceV2.String()).Return(false)
				testHelper.mockEntityRepository.EXPECT().GetByCode(args.ctx, args.req.EntityCode).Return(&models.Entity{}, nil)
				testHelper.mockAcctRepository.EXPECT().GetTrialBalance(gomock.Any(), gomock.Any()).Return(resC, resSC, nil)

				testHelper.mockEntityRepository.EXPECT().GetByCode(args.ctx, gomock.Any()).Return(&models.Entity{}, nil)
				testHelper.mockFile.EXPECT().NewCSVWriter(gomock.Any())
				testHelper.mockFile.EXPECT().CSVWriteAll(args.ctx, gomock.Any()).Return(nil)
				testHelper.mockFile.EXPECT().CSVWriteHeader(args.ctx, gomock.Any()).Return(nil).AnyTimes()
				testHelper.mockFile.EXPECT().CSVWriteBody(args.ctx, gomock.Any()).Return(nil).AnyTimes()
				testHelper.mockFile.EXPECT().CSVProcessWrite(args.ctx).Return(assert.AnError)
			},
			wantErr: true,
		},
		{
			name: "error case - send email",
			args: args{
				ctx: context.Background(),
				req: models.TrialBalanceFilterOptions{},
			},
			doMock: func(args args) {
				testHelper.mockFlag.EXPECT().IsEnabled(models.FlagGetTrialBalanceGCS.String()).Return(false)
				testHelper.mockFlag.EXPECT().IsEnabled(models.FlagGetTrialBalanceV2.String()).Return(false)
				testHelper.mockEntityRepository.EXPECT().GetByCode(args.ctx, args.req.EntityCode).Return(&models.Entity{}, nil)
				testHelper.mockAcctRepository.EXPECT().GetTrialBalance(gomock.Any(), gomock.Any()).Return(resC, resSC, nil)

				testHelper.mockEntityRepository.EXPECT().GetByCode(args.ctx, gomock.Any()).Return(&models.Entity{}, nil)
				testHelper.mockFile.EXPECT().NewCSVWriter(gomock.Any())
				testHelper.mockFile.EXPECT().CSVWriteAll(args.ctx, gomock.Any()).Return(nil)
				testHelper.mockFile.EXPECT().CSVWriteHeader(args.ctx, gomock.Any()).Return(nil).AnyTimes()
				testHelper.mockFile.EXPECT().CSVWriteBody(args.ctx, gomock.Any()).Return(nil).AnyTimes()
				testHelper.mockFile.EXPECT().CSVProcessWrite(args.ctx).Return(nil)
				testHelper.mockDDDNotification.EXPECT().SendEmail(args.ctx, gomock.Any()).Return(assert.AnError)
			},
			wantErr: true,
		},
		{
			name: "success",
			args: args{
				ctx: context.Background(),
				req: models.TrialBalanceFilterOptions{},
			},
			doMock: func(args args) {
				testHelper.mockFlag.EXPECT().IsEnabled(models.FlagGetTrialBalanceGCS.String()).Return(false)
				testHelper.mockFlag.EXPECT().IsEnabled(models.FlagGetTrialBalanceV2.String()).Return(false)
				testHelper.mockEntityRepository.EXPECT().GetByCode(args.ctx, args.req.EntityCode).Return(&models.Entity{}, nil)
				testHelper.mockAcctRepository.EXPECT().GetTrialBalance(gomock.Any(), gomock.Any()).Return(resC, resSC, nil)

				testHelper.mockEntityRepository.EXPECT().GetByCode(args.ctx, gomock.Any()).Return(&models.Entity{}, nil)
				testHelper.mockFile.EXPECT().NewCSVWriter(gomock.Any())
				testHelper.mockFile.EXPECT().CSVWriteAll(args.ctx, gomock.Any()).Return(nil)
				testHelper.mockFile.EXPECT().CSVWriteHeader(args.ctx, gomock.Any()).Return(nil).AnyTimes()
				testHelper.mockFile.EXPECT().CSVWriteBody(args.ctx, gomock.Any()).Return(nil).AnyTimes()
				testHelper.mockFile.EXPECT().CSVProcessWrite(args.ctx).Return(nil)
				testHelper.mockDDDNotification.EXPECT().SendEmail(args.ctx, gomock.Any()).Return(nil)
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
			err := testHelper.accountingService.DownloadTrialBalance(tt.args.ctx, tt.args.req)
			assert.Equal(t, tt.wantErr, err != nil)
		})
	}
}

func Test_accounting_GetTrialBalanceDetails(t *testing.T) {
	testHelper := serviceTestHelper(t)

	type args struct {
		ctx context.Context
		req models.TrialBalanceDetailsFilterOptions
	}

	type mockData struct{}

	tests := []struct {
		name     string
		args     args
		mockData mockData
		doMock   func(args args, mockData mockData)
		wantErr  bool
	}{
		{
			name: "success",
			args: args{
				ctx: context.Background(),
				req: models.TrialBalanceDetailsFilterOptions{},
			},
			doMock: func(args args, mockData mockData) {
				testHelper.mockEntityRepository.EXPECT().
					GetByCode(args.ctx, args.req.EntityCode).
					Return(&models.Entity{}, nil)
				testHelper.mockSubCategoryRepository.EXPECT().
					GetByCode(args.ctx, args.req.SubCategoryCode).
					Return(&models.SubCategory{}, nil)
				testHelper.mockAcctRepository.EXPECT().
					GetTrialBalanceDetails(gomock.Any(), gomock.Any()).
					Return([]models.TrialBalanceDetailOut{}, nil)
				testHelper.mockAccRepository.EXPECT().
					GetAccountListCount(gomock.Any(), gomock.Any()).
					Return(0, nil)

				testHelper.mockCacheRepository.EXPECT().
					Get(gomock.Any(), gomock.Any()).
					Return("0", nil)
			},
			wantErr: false,
		},
		{
			name: "error case - unable get trial balance account",
			args: args{
				ctx: context.Background(),
				req: models.TrialBalanceDetailsFilterOptions{},
			},
			doMock: func(args args, mockData mockData) {
				testHelper.mockEntityRepository.EXPECT().
					GetByCode(args.ctx, args.req.EntityCode).
					Return(&models.Entity{}, nil)
				testHelper.mockSubCategoryRepository.EXPECT().
					GetByCode(args.ctx, args.req.SubCategoryCode).
					Return(&models.SubCategory{}, nil)
				testHelper.mockAcctRepository.EXPECT().
					GetTrialBalanceDetails(gomock.Any(), gomock.Any()).
					Return(nil, assert.AnError)
			},
			wantErr: true,
		},
		{
			name: "error case - database error when get by sub category code",
			args: args{
				ctx: context.Background(),
				req: models.TrialBalanceDetailsFilterOptions{},
			},
			doMock: func(args args, mockData mockData) {
				testHelper.mockEntityRepository.EXPECT().
					GetByCode(args.ctx, args.req.EntityCode).
					Return(&models.Entity{}, nil)
				testHelper.mockSubCategoryRepository.EXPECT().
					GetByCode(args.ctx, args.req.SubCategoryCode).
					Return(nil, assert.AnError)
			},
			wantErr: true,
		},
		{
			name: "error case - sub category code not found",
			args: args{
				ctx: context.Background(),
				req: models.TrialBalanceDetailsFilterOptions{},
			},
			doMock: func(args args, mockData mockData) {
				testHelper.mockEntityRepository.EXPECT().
					GetByCode(args.ctx, args.req.EntityCode).
					Return(&models.Entity{}, nil)
				testHelper.mockSubCategoryRepository.EXPECT().
					GetByCode(args.ctx, args.req.SubCategoryCode).
					Return(nil, nil)
			},
			wantErr: true,
		},
		{
			name: "error case - database error when get by entity code",
			args: args{
				ctx: context.Background(),
				req: models.TrialBalanceDetailsFilterOptions{},
			},
			doMock: func(args args, mockData mockData) {
				testHelper.mockEntityRepository.EXPECT().
					GetByCode(args.ctx, args.req.EntityCode).
					Return(nil, assert.AnError)
			},
			wantErr: true,
		},
		{
			name: "error case - entity code not found",
			args: args{
				ctx: context.Background(),
				req: models.TrialBalanceDetailsFilterOptions{},
			},
			doMock: func(args args, mockData mockData) {
				testHelper.mockEntityRepository.EXPECT().
					GetByCode(args.ctx, args.req.EntityCode).
					Return(nil, nil)
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

			_, _, err := testHelper.accountingService.GetTrialBalanceDetails(tt.args.ctx, tt.args.req)
			assert.Equal(t, tt.wantErr, err != nil)
		})
	}
}

func Test_accounting_SendToEmailGetTrialBalanceDetails(t *testing.T) {
	testHelper := serviceTestHelper(t)
	header := []string{"Account Number", "Account Name", "Opening Balance (Rp)", "Debit Movement (Rp)", "Credit Movement (Rp)", "Closing Balance (Rp)"}
	dummyTBDResult := []models.TrialBalanceDetailOut{{
		AccountNumber:  "12345",
		AccountName:    "Test",
		OpeningBalance: decimal.NewFromInt(0),
		DebitMovement:  decimal.NewFromInt(0),
		CreditMovement: decimal.NewFromInt(0),
		ClosingBalance: decimal.NewFromInt(0),
	}}
	dummyCSVBody := []string{"12345", "Test", "0", "0", "0", "0"}
	type args struct {
		ctx context.Context
		req models.TrialBalanceDetailsFilterOptions
	}
	tests := []struct {
		name    string
		args    args
		doMock  func(args args)
		wantErr bool
	}{
		{
			name: "error case - get entity by code",
			args: args{
				ctx: context.Background(),
				req: models.TrialBalanceDetailsFilterOptions{},
			},
			doMock: func(args args) {
				testHelper.mockEntityRepository.EXPECT().GetByCode(args.ctx, args.req.EntityCode).Return(nil, assert.AnError)
			},
			wantErr: true,
		},
		{
			name: "not found case - get entity by code",
			args: args{
				ctx: context.Background(),
				req: models.TrialBalanceDetailsFilterOptions{},
			},
			doMock: func(args args) {
				testHelper.mockEntityRepository.EXPECT().GetByCode(args.ctx, args.req.EntityCode).Return(nil, nil)
			},
			wantErr: true,
		},
		{
			name: "error case - get subcategory by code",
			args: args{
				ctx: context.Background(),
				req: models.TrialBalanceDetailsFilterOptions{},
			},
			doMock: func(args args) {
				testHelper.mockEntityRepository.EXPECT().GetByCode(args.ctx, args.req.EntityCode).Return(&models.Entity{}, nil)
				testHelper.mockSubCategoryRepository.EXPECT().GetByCode(args.ctx, args.req.SubCategoryCode).Return(nil, assert.AnError)
			},
			wantErr: true,
		},
		{
			name: "not found case - get subcategory by code",
			args: args{
				ctx: context.Background(),
				req: models.TrialBalanceDetailsFilterOptions{},
			},
			doMock: func(args args) {
				testHelper.mockEntityRepository.EXPECT().GetByCode(args.ctx, args.req.EntityCode).Return(&models.Entity{}, nil)
				testHelper.mockSubCategoryRepository.EXPECT().GetByCode(args.ctx, args.req.SubCategoryCode).Return(nil, nil)
			},
			wantErr: true,
		},
		{
			name: "error case - get trial balance details",
			args: args{
				ctx: context.Background(),
				req: models.TrialBalanceDetailsFilterOptions{},
			},
			doMock: func(args args) {
				testHelper.mockEntityRepository.EXPECT().GetByCode(args.ctx, args.req.EntityCode).Return(&models.Entity{}, nil)
				testHelper.mockSubCategoryRepository.EXPECT().GetByCode(args.ctx, args.req.SubCategoryCode).Return(&models.SubCategory{}, nil)
				testHelper.mockAcctRepository.EXPECT().GetTrialBalanceDetails(args.ctx, args.req).Return(nil, assert.AnError)
			},
			wantErr: true,
		},
		{
			name: "error case - csv write header",
			args: args{
				ctx: context.Background(),
				req: models.TrialBalanceDetailsFilterOptions{},
			},
			doMock: func(args args) {
				testHelper.mockEntityRepository.EXPECT().GetByCode(args.ctx, args.req.EntityCode).Return(&models.Entity{}, nil)
				testHelper.mockSubCategoryRepository.EXPECT().GetByCode(args.ctx, args.req.SubCategoryCode).Return(&models.SubCategory{}, nil)
				testHelper.mockAcctRepository.EXPECT().GetTrialBalanceDetails(args.ctx, args.req).Return([]models.TrialBalanceDetailOut{}, nil)

				testHelper.mockFile.EXPECT().NewCSVWriter(gomock.Any())
				testHelper.mockFile.EXPECT().CSVWriteHeader(args.ctx, header).Return(assert.AnError)

			},
			wantErr: true,
		},
		{
			name: "error case - csv write body",
			args: args{
				ctx: context.Background(),
				req: models.TrialBalanceDetailsFilterOptions{},
			},
			doMock: func(args args) {
				testHelper.mockEntityRepository.EXPECT().GetByCode(args.ctx, args.req.EntityCode).Return(&models.Entity{}, nil)
				testHelper.mockSubCategoryRepository.EXPECT().GetByCode(args.ctx, args.req.SubCategoryCode).Return(&models.SubCategory{}, nil)
				testHelper.mockAcctRepository.EXPECT().GetTrialBalanceDetails(args.ctx, args.req).Return(dummyTBDResult, nil)

				testHelper.mockFile.EXPECT().NewCSVWriter(gomock.Any())
				testHelper.mockFile.EXPECT().CSVWriteHeader(args.ctx, header).Return(nil)
				testHelper.mockFile.EXPECT().CSVWriteBody(args.ctx, dummyCSVBody).Return(assert.AnError)
			},
			wantErr: true,
		},
		{
			name: "error case - csv write body",
			args: args{
				ctx: context.Background(),
				req: models.TrialBalanceDetailsFilterOptions{},
			},
			doMock: func(args args) {
				testHelper.mockEntityRepository.EXPECT().GetByCode(args.ctx, args.req.EntityCode).Return(&models.Entity{}, nil)
				testHelper.mockSubCategoryRepository.EXPECT().GetByCode(args.ctx, args.req.SubCategoryCode).Return(&models.SubCategory{}, nil)
				testHelper.mockAcctRepository.EXPECT().GetTrialBalanceDetails(args.ctx, args.req).Return(dummyTBDResult, nil)

				testHelper.mockFile.EXPECT().NewCSVWriter(gomock.Any())
				testHelper.mockFile.EXPECT().CSVWriteHeader(args.ctx, header).Return(nil)
				testHelper.mockFile.EXPECT().CSVWriteBody(args.ctx, dummyCSVBody).Return(assert.AnError)
			},
			wantErr: true,
		},
		{
			name: "error case - csv process write",
			args: args{
				ctx: context.Background(),
				req: models.TrialBalanceDetailsFilterOptions{},
			},
			doMock: func(args args) {
				testHelper.mockEntityRepository.EXPECT().GetByCode(args.ctx, args.req.EntityCode).Return(&models.Entity{}, nil)
				testHelper.mockSubCategoryRepository.EXPECT().GetByCode(args.ctx, args.req.SubCategoryCode).Return(&models.SubCategory{}, nil)
				testHelper.mockAcctRepository.EXPECT().GetTrialBalanceDetails(args.ctx, args.req).Return(dummyTBDResult, nil)

				testHelper.mockFile.EXPECT().NewCSVWriter(gomock.Any())
				testHelper.mockFile.EXPECT().CSVWriteHeader(args.ctx, header).Return(nil)
				testHelper.mockFile.EXPECT().CSVWriteBody(args.ctx, dummyCSVBody).Return(nil)
				testHelper.mockFile.EXPECT().CSVProcessWrite(args.ctx).Return(assert.AnError)
			},
			wantErr: true,
		},
		{
			name: "error case - get signed url",
			args: args{
				ctx: context.Background(),
				req: models.TrialBalanceDetailsFilterOptions{},
			},
			doMock: func(args args) {
				testHelper.mockEntityRepository.EXPECT().GetByCode(args.ctx, args.req.EntityCode).Return(&models.Entity{}, nil)
				testHelper.mockSubCategoryRepository.EXPECT().GetByCode(args.ctx, args.req.SubCategoryCode).Return(&models.SubCategory{}, nil)
				testHelper.mockAcctRepository.EXPECT().GetTrialBalanceDetails(args.ctx, args.req).Return(dummyTBDResult, nil)

				testHelper.mockFile.EXPECT().NewCSVWriter(gomock.Any())
				testHelper.mockFile.EXPECT().CSVWriteHeader(args.ctx, header).Return(nil)
				testHelper.mockFile.EXPECT().CSVWriteBody(args.ctx, dummyCSVBody).Return(nil)
				testHelper.mockFile.EXPECT().CSVProcessWrite(args.ctx).Return(nil)

				gcsPayload := &models.CloudStoragePayload{
					Filename: fmt.Sprintf("trialBalanceDetail-%s-%s-%s-%d.csv", args.req.StartDate.Format(atime.DateFormatYYYYMMDD), args.req.EndDate.Format(atime.DateFormatYYYYMMDD), args.req.SubCategoryName, 1),
					Path:     fmt.Sprintf("%s", models.TrialBalanceDetailDir),
				}
				tempFile, _ := os.CreateTemp("", "test_mock_gcs")

				testHelper.mockCloudStorageRepository.EXPECT().NewWriter(args.ctx, gcsPayload).Return(tempFile)
				testHelper.mockCloudStorageRepository.EXPECT().GetSignedURL(*gcsPayload, gomock.Any()).Return("", assert.AnError)
			},
			wantErr: true,
		},
		{
			name: "error case - send email",
			args: args{
				ctx: context.Background(),
				req: models.TrialBalanceDetailsFilterOptions{},
			},
			doMock: func(args args) {
				testHelper.mockEntityRepository.EXPECT().GetByCode(args.ctx, args.req.EntityCode).Return(&models.Entity{}, nil)
				testHelper.mockSubCategoryRepository.EXPECT().GetByCode(args.ctx, args.req.SubCategoryCode).Return(&models.SubCategory{}, nil)
				testHelper.mockAcctRepository.EXPECT().GetTrialBalanceDetails(args.ctx, args.req).Return(dummyTBDResult, nil)

				testHelper.mockFile.EXPECT().NewCSVWriter(gomock.Any())
				testHelper.mockFile.EXPECT().CSVWriteHeader(args.ctx, header).Return(nil)
				testHelper.mockFile.EXPECT().CSVWriteBody(args.ctx, dummyCSVBody).Return(nil)
				testHelper.mockFile.EXPECT().CSVProcessWrite(args.ctx).Return(nil)

				gcsPayload := &models.CloudStoragePayload{
					Filename: fmt.Sprintf("trialBalanceDetail-%s-%s-%s-%d.csv", args.req.StartDate.Format(atime.DateFormatYYYYMMDD), args.req.EndDate.Format(atime.DateFormatYYYYMMDD), args.req.SubCategoryName, 1),
					Path:     fmt.Sprintf("%s", models.TrialBalanceDetailDir),
				}
				tempFile, _ := os.CreateTemp("", "test_mock_gcs")

				testHelper.mockCloudStorageRepository.EXPECT().NewWriter(args.ctx, gcsPayload).Return(tempFile)
				testHelper.mockCloudStorageRepository.EXPECT().GetSignedURL(*gcsPayload, gomock.Any()).Return("https://test.com", nil)

				testHelper.mockDDDNotification.EXPECT().SendEmail(args.ctx, gomock.Any()).Return(assert.AnError)

			},
			wantErr: true,
		},
		{
			name: "success",
			args: args{
				ctx: context.Background(),
				req: models.TrialBalanceDetailsFilterOptions{},
			},
			doMock: func(args args) {
				testHelper.mockEntityRepository.EXPECT().GetByCode(args.ctx, args.req.EntityCode).Return(&models.Entity{}, nil)
				testHelper.mockSubCategoryRepository.EXPECT().GetByCode(args.ctx, args.req.SubCategoryCode).Return(&models.SubCategory{}, nil)
				testHelper.mockAcctRepository.EXPECT().GetTrialBalanceDetails(args.ctx, args.req).Return(dummyTBDResult, nil)

				testHelper.mockFile.EXPECT().NewCSVWriter(gomock.Any())
				testHelper.mockFile.EXPECT().CSVWriteHeader(args.ctx, header).Return(nil)
				testHelper.mockFile.EXPECT().CSVWriteBody(args.ctx, dummyCSVBody).Return(nil)
				testHelper.mockFile.EXPECT().CSVProcessWrite(args.ctx).Return(nil)

				gcsPayload := &models.CloudStoragePayload{
					Filename: fmt.Sprintf("trialBalanceDetail-%s-%s-%s-%d.csv", args.req.StartDate.Format(atime.DateFormatYYYYMMDD), args.req.EndDate.Format(atime.DateFormatYYYYMMDD), args.req.SubCategoryName, 1),
					Path:     fmt.Sprintf("%s", models.TrialBalanceDetailDir),
				}
				tempFile, _ := os.CreateTemp("", "test_mock_gcs")

				testHelper.mockCloudStorageRepository.EXPECT().NewWriter(args.ctx, gcsPayload).Return(tempFile)
				testHelper.mockCloudStorageRepository.EXPECT().GetSignedURL(*gcsPayload, gomock.Any()).Return("https://test.com", nil)

				testHelper.mockDDDNotification.EXPECT().SendEmail(args.ctx, gomock.Any()).Return(nil)

			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.doMock != nil {
				tt.doMock(tt.args)
			}
			err := testHelper.accountingService.SendToEmailGetTrialBalanceDetails(tt.args.ctx, tt.args.req)
			assert.Equal(t, tt.wantErr, err != nil)
		})
	}
}

func Test_accounting_GetTrialBalanceBySubCategoryCode(t *testing.T) {
	testHelper := serviceTestHelper(t)

	type args struct {
		ctx context.Context
		req models.TrialBalanceFilterOptions
	}

	type mockData struct{}

	tests := []struct {
		name     string
		args     args
		mockData mockData
		doMock   func(args args, mockData mockData)
		wantErr  bool
	}{
		{
			name: "success",
			args: args{
				ctx: context.Background(),
				req: models.TrialBalanceFilterOptions{},
			},
			doMock: func(args args, mockData mockData) {
				testHelper.mockAcctRepository.EXPECT().
					GetTrialBalanceSubCategory(gomock.Any(), gomock.Any()).
					Return(models.TrialBalanceBySubCategoryOut{}, nil)
			},
			wantErr: false,
		},
		{
			name: "error case - unable get trial balance account",
			args: args{
				ctx: context.Background(),
				req: models.TrialBalanceFilterOptions{},
			},
			doMock: func(args args, mockData mockData) {
				testHelper.mockAcctRepository.EXPECT().
					GetTrialBalanceSubCategory(gomock.Any(), gomock.Any()).
					Return(models.TrialBalanceBySubCategoryOut{}, assert.AnError)
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

			_, err := testHelper.accountingService.GetTrialBalanceBySubCategoryCode(tt.args.ctx, tt.args.req)
			assert.Equal(t, tt.wantErr, err != nil)
		})
	}
}

func Test_accounting_GetTrialBalanceFromGCS(t *testing.T) {
	testHelper := serviceTestHelper(t)

	type args struct {
		ctx context.Context
		req models.TrialBalanceDetailsFilterOptions
	}
	type mockData struct{}

	tests := []struct {
		name     string
		args     args
		mockData mockData
		doMock   func(args args, mockData mockData)
		wantErr  bool
	}{
		{
			name: "error case - failed to read total rows",
			args: args{
				ctx: context.Background(),
				req: models.TrialBalanceDetailsFilterOptions{},
			},

			doMock: func(args args, mockData mockData) {
				gcsPayload := &models.CloudStoragePayload{
					Path:     fmt.Sprintf("%s/%s/%s/details/%s/%s", models.TrialBalanceDir, args.req.Year, args.req.EntityCode, args.req.Month, args.req.SubCategoryCode),
					Filename: "total_rows.json",
				}
				testHelper.mockCloudStorageRepository.EXPECT().NewReader(args.ctx, gcsPayload).Return(nil, assert.AnError)
			},
			wantErr: true,
		},
		{
			name: "error case - exceed limit",
			args: args{
				ctx: context.Background(),
				req: models.TrialBalanceDetailsFilterOptions{},
			},
			doMock: func(args args, mockData mockData) {
				totalRowsReader := io.NopCloser(strings.NewReader(`{"total_rows":"1000"}`))
				testHelper.mockCloudStorageRepository.EXPECT().
					NewReader(args.ctx, gomock.Any()).
					Return(totalRowsReader, nil)
			},
			wantErr: true,
		},
		{
			name: "error case - failed to list file",
			args: args{
				ctx: context.Background(),
				req: models.TrialBalanceDetailsFilterOptions{},
			},
			doMock: func(args args, mockData mockData) {
				totalRowsReader := io.NopCloser(strings.NewReader(`{"total_rows":"2"}`))
				testHelper.mockCloudStorageRepository.EXPECT().
					NewReader(args.ctx, gomock.Any()).
					Return(totalRowsReader, nil)
				testHelper.mockCloudStorageRepository.EXPECT().
					ListFiles(args.ctx, gomock.Any()).
					Return(nil, assert.AnError)

			},
			wantErr: true,
		},
		{
			name: "error case - failed to read file",
			args: args{
				ctx: context.Background(),
				req: models.TrialBalanceDetailsFilterOptions{},
			},
			doMock: func(args args, mockData mockData) {
				totalRowsReader := io.NopCloser(strings.NewReader(`{"total_rows":"2"}`))
				testHelper.mockCloudStorageRepository.EXPECT().
					NewReader(args.ctx, gomock.Any()).
					Return(totalRowsReader, nil)
				testHelper.mockCloudStorageRepository.EXPECT().
					ListFiles(args.ctx, gomock.Any()).
					Return([]string{"details_1.csv", "total_rows.json"}, nil)
				testHelper.mockCloudStorageRepository.EXPECT().NewReader(args.ctx, gomock.Any()).Return(nil, assert.AnError)

			},
			wantErr: true,
		},
		{
			name: "error case - failed to get detail sub category",
			args: args{
				ctx: context.Background(),
				req: models.TrialBalanceDetailsFilterOptions{},
			},
			doMock: func(args args, mockData mockData) {
				totalRowsReader := io.NopCloser(strings.NewReader(`{"total_rows":"2"}`))
				testHelper.mockCloudStorageRepository.EXPECT().
					NewReader(args.ctx, gomock.Any()).
					Return(totalRowsReader, nil)
				testHelper.mockCloudStorageRepository.EXPECT().
					ListFiles(args.ctx, gomock.Any()).
					Return([]string{"details_1.csv", "total_rows.json"}, nil)
				csvContent := `account_number,account_name,entity_code,category_code,sub_category_code,opening_balance,debit_movement,credit_movement,closing_balance
					0001,Cash,001,11,11032,100,50,20,130
					0002,Receivable,001,11,11032,200,30,40,190`
				csvReader := io.NopCloser(strings.NewReader(csvContent))
				testHelper.mockCloudStorageRepository.EXPECT().NewReader(args.ctx, gomock.Any()).Return(csvReader, nil)

				testHelper.mockSubCategoryRepository.EXPECT().GetByCode(args.ctx, gomock.Any()).Return(nil, assert.AnError)

			},
			wantErr: true,
		},
		{
			name: "success case",
			args: args{
				ctx: context.Background(),
				req: models.TrialBalanceDetailsFilterOptions{},
			},
			doMock: func(args args, mockData mockData) {
				totalRowsReader := io.NopCloser(strings.NewReader(`{"total_rows":"2"}`))
				testHelper.mockCloudStorageRepository.EXPECT().
					NewReader(args.ctx, gomock.Any()).
					Return(totalRowsReader, nil)
				testHelper.mockCloudStorageRepository.EXPECT().
					ListFiles(args.ctx, gomock.Any()).
					Return([]string{"details_1.csv", "total_rows.json"}, nil)
				csvContent := `account_number,account_name,entity_code,category_code,sub_category_code,opening_balance,debit_movement,credit_movement,closing_balance
					0001,Cash,001,11,11032,100,50,20,130
					0002,Receivable,001,11,11032,200,30,40,190`
				csvReader := io.NopCloser(strings.NewReader(csvContent))
				testHelper.mockCloudStorageRepository.EXPECT().NewReader(args.ctx, gomock.Any()).Return(csvReader, nil)

				testHelper.mockSubCategoryRepository.EXPECT().GetByCode(args.ctx, gomock.Any()).Return(&models.SubCategory{
					Name: "Cash Point",
					Code: "1111",
				}, nil)

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

			_, _, err := testHelper.accountingService.GetTrialBalanceFromGCS(tt.args.ctx, tt.args.req)
			assert.Equal(t, tt.wantErr, err != nil)
		})
	}

}

func Test_accounting_SendEmailTrialBalanceDetails(t *testing.T) {
	testHelper := serviceTestHelper(t)

	entityData := &models.Entity{
		Code:        "001",
		Name:        "AMF",
		Description: "Amartha",
	}

	subCategData := &models.SubCategory{
		Code: "11022",
		Name: "Cash Point",
	}

	req := models.TrialBalanceDetailsFilterOptions{
		Year:            "2025",
		Month:           "01",
		SubCategoryCode: "11022",
		EntityCode:      "001",
		Email:           "test@amartha.com",
	}

	type args struct {
		ctx context.Context
		req models.TrialBalanceDetailsFilterOptions
	}

	tests := []struct {
		name    string
		args    args
		doMock  func(args args)
		wantErr bool
	}{
		{
			name: "error case - failed to get entity data",
			args: args{
				ctx: context.Background(),
				req: req,
			},
			doMock: func(args args) {
				testHelper.mockEntityRepository.EXPECT().GetByCode(args.ctx, gomock.Any()).Return(nil, assert.AnError)
			},
			wantErr: true,
		},
		{
			name: "error case - entity data not found",
			args: args{
				ctx: context.Background(),
				req: req,
			},
			doMock: func(args args) {
				testHelper.mockEntityRepository.EXPECT().GetByCode(args.ctx, gomock.Any()).Return(nil, nil)
			},
			wantErr: true,
		},
		{
			name: "error case - failed to get subcategory data",
			args: args{
				ctx: context.Background(),
				req: req,
			},
			doMock: func(args args) {
				testHelper.mockEntityRepository.EXPECT().GetByCode(args.ctx, gomock.Any()).Return(entityData, nil)
				testHelper.mockSubCategoryRepository.EXPECT().GetByCode(args.ctx, gomock.Any()).Return(nil, assert.AnError)
			},
			wantErr: true,
		},
		{
			name: "error case - subcategory data not found",
			args: args{
				ctx: context.Background(),
				req: req,
			},
			doMock: func(args args) {
				testHelper.mockEntityRepository.EXPECT().GetByCode(args.ctx, gomock.Any()).Return(entityData, nil)
				testHelper.mockSubCategoryRepository.EXPECT().GetByCode(args.ctx, gomock.Any()).Return(nil, nil)
			},
			wantErr: true,
		},
		{
			name: "error case - failed to get from gcs",
			args: args{
				ctx: context.Background(),
				req: req,
			},
			doMock: func(args args) {
				testHelper.mockEntityRepository.EXPECT().GetByCode(args.ctx, gomock.Any()).Return(entityData, nil)
				testHelper.mockSubCategoryRepository.EXPECT().GetByCode(args.ctx, gomock.Any()).Return(subCategData, nil)
				testHelper.mockCloudStorageRepository.EXPECT().ListFiles(args.ctx, gomock.Any()).Return(nil, assert.AnError)
			},
			wantErr: true,
		},
		{
			name: "error case - failed to get signed url",
			args: args{
				ctx: context.Background(),
				req: req,
			},
			doMock: func(args args) {
				testHelper.mockEntityRepository.EXPECT().GetByCode(args.ctx, gomock.Any()).Return(entityData, nil)
				testHelper.mockSubCategoryRepository.EXPECT().GetByCode(args.ctx, gomock.Any()).Return(subCategData, nil)
				testHelper.mockCloudStorageRepository.EXPECT().ListFiles(args.ctx, gomock.Any()).Return([]string{"11231.csv"}, nil)
				testHelper.mockCloudStorageRepository.EXPECT().GetSignedURL(gomock.Any(), gomock.Any()).Return("", assert.AnError)
			},
			wantErr: true,
		},
		{
			name: "error case - failed send email",
			args: args{
				ctx: context.Background(),
				req: req,
			},
			doMock: func(args args) {
				testHelper.mockEntityRepository.EXPECT().GetByCode(args.ctx, gomock.Any()).Return(entityData, nil)
				testHelper.mockSubCategoryRepository.EXPECT().GetByCode(args.ctx, gomock.Any()).Return(subCategData, nil)
				testHelper.mockCloudStorageRepository.EXPECT().ListFiles(args.ctx, gomock.Any()).Return([]string{"11231.csv"}, nil)
				testHelper.mockCloudStorageRepository.EXPECT().GetSignedURL(gomock.Any(), gomock.Any()).Return("https://test.com", nil)
				testHelper.mockDDDNotification.EXPECT().SendEmail(args.ctx, gomock.Any()).Return(assert.AnError)

			},
			wantErr: true,
		},
		{
			name: "success",
			args: args{
				ctx: context.Background(),
				req: req,
			},
			doMock: func(args args) {
				testHelper.mockEntityRepository.EXPECT().GetByCode(args.ctx, gomock.Any()).Return(entityData, nil)
				testHelper.mockSubCategoryRepository.EXPECT().GetByCode(args.ctx, gomock.Any()).Return(subCategData, nil)
				testHelper.mockCloudStorageRepository.EXPECT().ListFiles(args.ctx, gomock.Any()).Return([]string{"11231.csv"}, nil)
				testHelper.mockCloudStorageRepository.EXPECT().GetSignedURL(gomock.Any(), gomock.Any()).Return("https://test.com", nil)
				testHelper.mockDDDNotification.EXPECT().SendEmail(args.ctx, gomock.Any()).Return(nil)

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
			err := testHelper.accountingService.SendEmailTrialBalanceDetails(tt.args.ctx, tt.args.req)
			assert.Equal(t, tt.wantErr, err != nil)
		})
	}

}
