package services_test

import (
	"context"
	"fmt"
	"testing"

	"bitbucket.org/Amartha/go-accounting/internal/models"
	"bitbucket.org/Amartha/go-accounting/internal/pkg/godbledger"
	"bitbucket.org/Amartha/go-accounting/internal/repositories/mysql"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestAccountService_CreateLoanPartnerAccount(t *testing.T) {
	testHelper := serviceTestHelper(t)
	var (
		latestProductCode = "1011"
		newestProductCode = "1012"
		latestCateg       = "131"
		latestSubCateg    = "13101"
		newestSubCateg    = "13102"
	)
	req := models.CreateAccountLoanPartner{
		PartnerName: "Chickin X",
		PartnerId:   "1234",
		LoanKind:    "PartnershipLoan",
		Metadata:    &models.Metadata{},
	}
	reqProductType := models.CreateProductTypeRequest{
		Code:   newestProductCode,
		Name:   req.PartnerName,
		Status: models.AccountStatusActive,
	}
	reqSubCategory := models.CreateSubCategory{
		Code:            newestSubCateg,
		Name:            fmt.Sprintf("Borrower Outstanding - %s", req.PartnerName),
		AccountType:     "PARTNERSHIP_LOAN",
		CategoryCode:    models.CategoryCode131,
		Currency:        models.CurrencyIDR,
		ProductTypeCode: newestProductCode,
		Description:     fmt.Sprintf("Borrower Outstanding - %s", req.PartnerName),
		Status:          models.StatusActive,
	}
	type args struct {
		ctx context.Context
		in  models.CreateAccountLoanPartner
	}
	tests := []struct {
		name    string
		args    args
		doMock  func(args args)
		wantErr bool
	}{
		{
			name: "error case - get latest product",
			args: args{
				ctx: context.Background(),
				in:  req,
			},
			doMock: func(args args) {
				testHelper.mockMySQLRepository.EXPECT().Atomic(gomock.Any(), gomock.Any()).DoAndReturn(
					func(ctx context.Context, steps func(ctx context.Context, r mysql.SQLRepository) error) error {
						testHelper.mockProductTypeRepository.EXPECT().GetLatestProductCode(args.ctx).Return(latestProductCode, models.ErrDataNotFound)
						return steps(ctx, testHelper.mockMySQLRepository)
					}).Return(models.ErrDataNotFound)
				testHelper.mockCacheRepository.EXPECT().
					DeleteKeysWithPrefix(args.ctx, gomock.Any()).Return(nil)
			},
			wantErr: true,
		},
		{
			name: "error case - convert latest product",
			args: args{
				ctx: context.Background(),
				in:  req,
			},
			doMock: func(args args) {
				testHelper.mockMySQLRepository.EXPECT().Atomic(gomock.Any(), gomock.Any()).DoAndReturn(
					func(ctx context.Context, steps func(ctx context.Context, r mysql.SQLRepository) error) error {
						testHelper.mockProductTypeRepository.EXPECT().GetLatestProductCode(args.ctx).Return("abc", nil)
						return steps(ctx, testHelper.mockMySQLRepository)
					}).Return(models.ErrValidation)
				testHelper.mockCacheRepository.EXPECT().
					DeleteKeysWithPrefix(args.ctx, gomock.Any()).Return(nil)
			},
			wantErr: true,
		},
		{
			name: "error case - product type code exist",
			args: args{
				ctx: context.Background(),
				in: models.CreateAccountLoanPartner{
					PartnerName: "Chickin X",
					PartnerId:   "1234",
					LoanKind:    "PartnershipLoan",
					Metadata:    &models.Metadata{},
				},
			},
			doMock: func(args args) {
				testHelper.mockMySQLRepository.EXPECT().Atomic(gomock.Any(), gomock.Any()).DoAndReturn(
					func(ctx context.Context, steps func(ctx context.Context, r mysql.SQLRepository) error) error {
						testHelper.mockProductTypeRepository.EXPECT().GetLatestProductCode(args.ctx).Return(latestProductCode, nil)
						testHelper.mockProductTypeRepository.EXPECT().GetByCode(args.ctx, reqProductType.Code).Return(&models.ProductType{}, nil)
						return steps(ctx, testHelper.mockMySQLRepository)
					}).Return(models.ErrDataExist)
				testHelper.mockCacheRepository.EXPECT().
					DeleteKeysWithPrefix(args.ctx, gomock.Any()).Return(nil)
			},
			wantErr: true,
		},
		{
			name: "error case - create product",
			args: args{
				ctx: context.Background(),
				in:  req,
			},
			doMock: func(args args) {
				testHelper.mockMySQLRepository.EXPECT().Atomic(gomock.Any(), gomock.Any()).DoAndReturn(
					func(ctx context.Context, steps func(ctx context.Context, r mysql.SQLRepository) error) error {
						testHelper.mockProductTypeRepository.EXPECT().GetLatestProductCode(args.ctx).Return(latestProductCode, nil)
						testHelper.mockProductTypeRepository.EXPECT().GetByCode(args.ctx, reqProductType.Code).Return(nil, nil)
						testHelper.mockProductTypeRepository.EXPECT().Create(args.ctx, &reqProductType).Return(models.ErrNoRowsAffected)
						return steps(ctx, testHelper.mockMySQLRepository)
					}).Return(models.ErrNoRowsAffected)
				testHelper.mockCacheRepository.EXPECT().
					DeleteKeysWithPrefix(args.ctx, gomock.Any()).Return(nil)
			},
			wantErr: true,
		},
		{
			name: "error case - get latest sub categ",
			args: args{
				ctx: context.Background(),
				in:  req,
			},
			doMock: func(args args) {
				testHelper.mockMySQLRepository.EXPECT().Atomic(gomock.Any(), gomock.Any()).DoAndReturn(
					func(ctx context.Context, steps func(ctx context.Context, r mysql.SQLRepository) error) error {
						testHelper.mockProductTypeRepository.EXPECT().GetLatestProductCode(args.ctx).Return(latestProductCode, nil)
						testHelper.mockProductTypeRepository.EXPECT().GetByCode(args.ctx, reqProductType.Code).Return(nil, nil)
						testHelper.mockProductTypeRepository.EXPECT().Create(args.ctx, &reqProductType).Return(nil)

						testHelper.mockSubCategoryRepository.EXPECT().GetLatestSubCategCode(args.ctx, latestCateg).Return(latestSubCateg, models.ErrNoRows)
						return steps(ctx, testHelper.mockMySQLRepository)
					}).Return(models.ErrNoRows)
				testHelper.mockCacheRepository.EXPECT().
					DeleteKeysWithPrefix(args.ctx, gomock.Any()).Return(nil)
			},
			wantErr: true,
		},
		{
			name: "error case - convert sub categ",
			args: args{
				ctx: context.Background(),
				in:  req,
			},
			doMock: func(args args) {
				testHelper.mockMySQLRepository.EXPECT().Atomic(gomock.Any(), gomock.Any()).DoAndReturn(
					func(ctx context.Context, steps func(ctx context.Context, r mysql.SQLRepository) error) error {
						testHelper.mockProductTypeRepository.EXPECT().GetLatestProductCode(args.ctx).Return(latestProductCode, nil)
						testHelper.mockProductTypeRepository.EXPECT().GetByCode(args.ctx, reqProductType.Code).Return(nil, nil)
						testHelper.mockProductTypeRepository.EXPECT().Create(args.ctx, &reqProductType).Return(nil)

						testHelper.mockSubCategoryRepository.EXPECT().GetLatestSubCategCode(args.ctx, "131").Return("131abc", nil)
						return steps(ctx, testHelper.mockMySQLRepository)
					}).Return(models.ErrValidation)
				testHelper.mockCacheRepository.EXPECT().
					DeleteKeysWithPrefix(args.ctx, gomock.Any()).Return(nil)
			},
			wantErr: true,
		},
		{
			name: "error case - create sub categ",
			args: args{
				ctx: context.Background(),
				in:  req,
			},
			doMock: func(args args) {
				testHelper.mockMySQLRepository.EXPECT().Atomic(gomock.Any(), gomock.Any()).DoAndReturn(
					func(ctx context.Context, steps func(ctx context.Context, r mysql.SQLRepository) error) error {
						testHelper.mockProductTypeRepository.EXPECT().GetLatestProductCode(args.ctx).Return(latestProductCode, nil)
						testHelper.mockProductTypeRepository.EXPECT().GetByCode(args.ctx, reqProductType.Code).Return(nil, nil)
						testHelper.mockProductTypeRepository.EXPECT().Create(args.ctx, &reqProductType).Return(nil)

						testHelper.mockSubCategoryRepository.EXPECT().GetLatestSubCategCode(args.ctx, latestCateg).Return(latestSubCateg, nil)
						testHelper.mockCategoryRepository.EXPECT().GetByCode(args.ctx, reqSubCategory.CategoryCode).Return(&models.Category{}, nil)
						testHelper.mockSubCategoryRepository.EXPECT().GetByCode(args.ctx, reqSubCategory.Code).Return(nil, nil)
						testHelper.mockSubCategoryRepository.EXPECT().GetByAccountType(args.ctx, reqSubCategory.AccountType).Return(nil, nil)
						testHelper.mockProductTypeRepository.EXPECT().GetByCode(args.ctx, reqSubCategory.ProductTypeCode).Return(&models.ProductType{}, nil)
						testHelper.mockGoDbLedger.EXPECT().GetCurrency(args.ctx, reqSubCategory.Currency).Return(godbledger.CurrencyIDR, nil)
						testHelper.mockSubCategoryRepository.EXPECT().Create(args.ctx, &reqSubCategory).Return(models.ErrNoRowsAffected)

						return steps(ctx, testHelper.mockMySQLRepository)
					}).Return(models.ErrNoRowsAffected)
			},
			wantErr: true,
		},
		{
			name: "error case - generate account number",
			args: args{
				ctx: context.Background(),
				in:  req,
			},
			doMock: func(args args) {
				testHelper.mockMySQLRepository.EXPECT().Atomic(gomock.Any(), gomock.Any()).DoAndReturn(
					func(ctx context.Context, steps func(ctx context.Context, r mysql.SQLRepository) error) error {
						testHelper.mockProductTypeRepository.EXPECT().GetLatestProductCode(args.ctx).Return(latestProductCode, nil)
						testHelper.mockProductTypeRepository.EXPECT().GetByCode(args.ctx, reqProductType.Code).Return(nil, nil)
						testHelper.mockProductTypeRepository.EXPECT().Create(args.ctx, &reqProductType).Return(nil)

						testHelper.mockSubCategoryRepository.EXPECT().GetLatestSubCategCode(args.ctx, latestCateg).Return(latestSubCateg, nil)
						testHelper.mockCategoryRepository.EXPECT().GetByCode(args.ctx, reqSubCategory.CategoryCode).Return(&models.Category{}, nil)
						testHelper.mockSubCategoryRepository.EXPECT().GetByCode(args.ctx, reqSubCategory.Code).Return(nil, nil)
						testHelper.mockSubCategoryRepository.EXPECT().GetByAccountType(args.ctx, reqSubCategory.AccountType).Return(nil, nil)
						testHelper.mockProductTypeRepository.EXPECT().GetByCode(args.ctx, reqSubCategory.ProductTypeCode).Return(&models.ProductType{}, nil)
						testHelper.mockGoDbLedger.EXPECT().GetCurrency(args.ctx, reqSubCategory.Currency).Return(godbledger.CurrencyIDR, nil)
						testHelper.mockSubCategoryRepository.EXPECT().Create(args.ctx, &reqSubCategory).Return(nil)

						testHelper.mockSubCategoryRepository.EXPECT().GetByCode(args.ctx, "12101").Return(&models.SubCategory{}, assert.AnError)

						return steps(ctx, testHelper.mockMySQLRepository)
					}).Return(assert.AnError)
			},
			wantErr: true,
		},
		{
			name: "error case - get loan partner",
			args: args{
				ctx: context.Background(),
				in:  req,
			},
			doMock: func(args args) {
				testHelper.mockMySQLRepository.EXPECT().Atomic(gomock.Any(), gomock.Any()).DoAndReturn(
					func(ctx context.Context, steps func(ctx context.Context, r mysql.SQLRepository) error) error {
						testHelper.mockProductTypeRepository.EXPECT().GetLatestProductCode(args.ctx).Return(latestProductCode, nil)
						testHelper.mockProductTypeRepository.EXPECT().GetByCode(args.ctx, reqProductType.Code).Return(nil, nil)
						testHelper.mockProductTypeRepository.EXPECT().Create(args.ctx, &reqProductType).Return(nil)

						testHelper.mockSubCategoryRepository.EXPECT().GetLatestSubCategCode(args.ctx, latestCateg).Return(latestSubCateg, nil)
						testHelper.mockCategoryRepository.EXPECT().GetByCode(args.ctx, reqSubCategory.CategoryCode).Return(&models.Category{}, nil)
						testHelper.mockSubCategoryRepository.EXPECT().GetByCode(args.ctx, reqSubCategory.Code).Return(nil, nil)
						testHelper.mockSubCategoryRepository.EXPECT().GetByAccountType(args.ctx, reqSubCategory.AccountType).Return(nil, nil)
						testHelper.mockProductTypeRepository.EXPECT().GetByCode(args.ctx, reqSubCategory.ProductTypeCode).Return(&models.ProductType{}, nil)
						testHelper.mockGoDbLedger.EXPECT().GetCurrency(args.ctx, reqSubCategory.Currency).Return(godbledger.CurrencyIDR, nil)
						testHelper.mockSubCategoryRepository.EXPECT().Create(args.ctx, &reqSubCategory).Return(nil)

						testHelper.mockSubCategoryRepository.EXPECT().GetByCode(args.ctx, "12101").Return(&models.SubCategory{}, nil).AnyTimes()
						testHelper.mockCacheRepository.EXPECT().GetIncrement(args.ctx, gomock.Any()).Return(int64(1), nil).AnyTimes()
						testHelper.mockLoanPartnerAccountRepository.EXPECT().GetByParams(args.ctx, gomock.Any()).Return([]models.LoanPartnerAccount{}, assert.AnError)

						return steps(ctx, testHelper.mockMySQLRepository)
					}).Return(assert.AnError)
			},
			wantErr: true,
		},
		{
			name: "error case - loan partner account is exist",
			args: args{
				ctx: context.Background(),
				in:  req,
			},
			doMock: func(args args) {
				testHelper.mockMySQLRepository.EXPECT().Atomic(gomock.Any(), gomock.Any()).DoAndReturn(
					func(ctx context.Context, steps func(ctx context.Context, r mysql.SQLRepository) error) error {
						testHelper.mockProductTypeRepository.EXPECT().GetLatestProductCode(args.ctx).Return(latestProductCode, nil)
						testHelper.mockProductTypeRepository.EXPECT().GetByCode(args.ctx, reqProductType.Code).Return(nil, nil)
						testHelper.mockProductTypeRepository.EXPECT().Create(args.ctx, &reqProductType).Return(nil)

						testHelper.mockSubCategoryRepository.EXPECT().GetLatestSubCategCode(args.ctx, latestCateg).Return(latestSubCateg, nil)
						testHelper.mockCategoryRepository.EXPECT().GetByCode(args.ctx, reqSubCategory.CategoryCode).Return(&models.Category{}, nil)
						testHelper.mockSubCategoryRepository.EXPECT().GetByCode(args.ctx, reqSubCategory.Code).Return(nil, nil)
						testHelper.mockSubCategoryRepository.EXPECT().GetByAccountType(args.ctx, reqSubCategory.AccountType).Return(nil, nil)
						testHelper.mockProductTypeRepository.EXPECT().GetByCode(args.ctx, reqSubCategory.ProductTypeCode).Return(&models.ProductType{}, nil)
						testHelper.mockGoDbLedger.EXPECT().GetCurrency(args.ctx, reqSubCategory.Currency).Return(godbledger.CurrencyIDR, nil)
						testHelper.mockSubCategoryRepository.EXPECT().Create(args.ctx, &reqSubCategory).Return(nil)

						testHelper.mockSubCategoryRepository.EXPECT().GetByCode(args.ctx, "12101").Return(&models.SubCategory{}, nil).AnyTimes()
						testHelper.mockCacheRepository.EXPECT().GetIncrement(args.ctx, gomock.Any()).Return(int64(1), nil).AnyTimes()
						testHelper.mockLoanPartnerAccountRepository.EXPECT().GetByParams(args.ctx, gomock.Any()).Return([]models.LoanPartnerAccount{
							{
								PartnerId: req.PartnerId,
							},
						}, nil)

						return steps(ctx, testHelper.mockMySQLRepository)
					}).Return(assert.AnError)
			},
			wantErr: true,
		},
		{
			name: "error case - insert acct account",
			args: args{
				ctx: context.Background(),
				in:  req,
			},
			doMock: func(args args) {
				testHelper.mockMySQLRepository.EXPECT().Atomic(gomock.Any(), gomock.Any()).DoAndReturn(
					func(ctx context.Context, steps func(ctx context.Context, r mysql.SQLRepository) error) error {
						testHelper.mockProductTypeRepository.EXPECT().GetLatestProductCode(args.ctx).Return(latestProductCode, nil)
						testHelper.mockProductTypeRepository.EXPECT().GetByCode(args.ctx, reqProductType.Code).Return(nil, nil)
						testHelper.mockProductTypeRepository.EXPECT().Create(args.ctx, &reqProductType).Return(nil)

						testHelper.mockSubCategoryRepository.EXPECT().GetLatestSubCategCode(args.ctx, latestCateg).Return(latestSubCateg, nil)
						testHelper.mockCategoryRepository.EXPECT().GetByCode(args.ctx, reqSubCategory.CategoryCode).Return(&models.Category{}, nil)
						testHelper.mockSubCategoryRepository.EXPECT().GetByCode(args.ctx, reqSubCategory.Code).Return(nil, nil)
						testHelper.mockSubCategoryRepository.EXPECT().GetByAccountType(args.ctx, reqSubCategory.AccountType).Return(nil, nil)
						testHelper.mockProductTypeRepository.EXPECT().GetByCode(args.ctx, reqSubCategory.ProductTypeCode).Return(&models.ProductType{}, nil)
						testHelper.mockGoDbLedger.EXPECT().GetCurrency(args.ctx, reqSubCategory.Currency).Return(godbledger.CurrencyIDR, nil)
						testHelper.mockSubCategoryRepository.EXPECT().Create(args.ctx, &reqSubCategory).Return(nil)

						testHelper.mockLoanPartnerAccountRepository.EXPECT().GetByParams(args.ctx, gomock.Any()).Return([]models.LoanPartnerAccount{}, nil).AnyTimes()
						testHelper.mockSubCategoryRepository.EXPECT().GetByCode(args.ctx, "12101").Return(&models.SubCategory{}, nil).AnyTimes()
						testHelper.mockCacheRepository.EXPECT().GetIncrement(args.ctx, gomock.Any()).Return(int64(1), nil).AnyTimes()

						testHelper.mockAccRepository.EXPECT().BulkInsertAcctAccount(args.ctx, gomock.Any()).Return(models.ErrUnableToCreate)
						return steps(ctx, testHelper.mockMySQLRepository)
					}).Return(models.ErrUnableToCreate)
			},
			wantErr: true,
		},
		{
			name: "error case - insert account",
			args: args{
				ctx: context.Background(),
				in:  req,
			},
			doMock: func(args args) {
				testHelper.mockMySQLRepository.EXPECT().Atomic(gomock.Any(), gomock.Any()).DoAndReturn(
					func(ctx context.Context, steps func(ctx context.Context, r mysql.SQLRepository) error) error {
						testHelper.mockProductTypeRepository.EXPECT().GetLatestProductCode(args.ctx).Return(latestProductCode, nil)
						testHelper.mockProductTypeRepository.EXPECT().GetByCode(args.ctx, reqProductType.Code).Return(nil, nil)
						testHelper.mockProductTypeRepository.EXPECT().Create(args.ctx, &reqProductType).Return(nil)

						testHelper.mockSubCategoryRepository.EXPECT().GetLatestSubCategCode(args.ctx, latestCateg).Return(latestSubCateg, nil)
						testHelper.mockCategoryRepository.EXPECT().GetByCode(args.ctx, reqSubCategory.CategoryCode).Return(&models.Category{}, nil)
						testHelper.mockSubCategoryRepository.EXPECT().GetByCode(args.ctx, reqSubCategory.Code).Return(nil, nil)
						testHelper.mockSubCategoryRepository.EXPECT().GetByAccountType(args.ctx, reqSubCategory.AccountType).Return(nil, nil)
						testHelper.mockProductTypeRepository.EXPECT().GetByCode(args.ctx, reqSubCategory.ProductTypeCode).Return(&models.ProductType{}, nil)
						testHelper.mockGoDbLedger.EXPECT().GetCurrency(args.ctx, reqSubCategory.Currency).Return(godbledger.CurrencyIDR, nil)
						testHelper.mockSubCategoryRepository.EXPECT().Create(args.ctx, &reqSubCategory).Return(nil)

						testHelper.mockLoanPartnerAccountRepository.EXPECT().GetByParams(args.ctx, gomock.Any()).Return([]models.LoanPartnerAccount{}, nil).AnyTimes()
						testHelper.mockSubCategoryRepository.EXPECT().GetByCode(args.ctx, "12101").Return(&models.SubCategory{}, nil).AnyTimes()
						testHelper.mockCacheRepository.EXPECT().GetIncrement(args.ctx, gomock.Any()).Return(int64(1), nil).AnyTimes()

						testHelper.mockAccRepository.EXPECT().BulkInsertAcctAccount(args.ctx, gomock.Any()).Return(nil)
						testHelper.mockAccRepository.EXPECT().BulkInsertAccount(args.ctx, gomock.Any()).Return(models.ErrUnableToCreate)
						return steps(ctx, testHelper.mockMySQLRepository)
					}).Return(models.ErrUnableToCreate)
			},
			wantErr: true,
		},
		{
			name: "error case - insert loan partner account",
			args: args{
				ctx: context.Background(),
				in:  req,
			},
			doMock: func(args args) {
				testHelper.mockMySQLRepository.EXPECT().Atomic(gomock.Any(), gomock.Any()).DoAndReturn(
					func(ctx context.Context, steps func(ctx context.Context, r mysql.SQLRepository) error) error {
						testHelper.mockProductTypeRepository.EXPECT().GetLatestProductCode(args.ctx).Return(latestProductCode, nil)
						testHelper.mockProductTypeRepository.EXPECT().GetByCode(args.ctx, reqProductType.Code).Return(nil, nil)
						testHelper.mockProductTypeRepository.EXPECT().Create(args.ctx, &reqProductType).Return(nil)

						testHelper.mockSubCategoryRepository.EXPECT().GetLatestSubCategCode(args.ctx, latestCateg).Return(latestSubCateg, nil)
						testHelper.mockCategoryRepository.EXPECT().GetByCode(args.ctx, reqSubCategory.CategoryCode).Return(&models.Category{}, nil)
						testHelper.mockSubCategoryRepository.EXPECT().GetByCode(args.ctx, reqSubCategory.Code).Return(nil, nil)
						testHelper.mockSubCategoryRepository.EXPECT().GetByAccountType(args.ctx, reqSubCategory.AccountType).Return(nil, nil)
						testHelper.mockProductTypeRepository.EXPECT().GetByCode(args.ctx, reqSubCategory.ProductTypeCode).Return(&models.ProductType{}, nil)
						testHelper.mockGoDbLedger.EXPECT().GetCurrency(args.ctx, reqSubCategory.Currency).Return(godbledger.CurrencyIDR, nil)
						testHelper.mockSubCategoryRepository.EXPECT().Create(args.ctx, &reqSubCategory).Return(nil)

						testHelper.mockLoanPartnerAccountRepository.EXPECT().GetByParams(args.ctx, gomock.Any()).Return([]models.LoanPartnerAccount{}, nil).AnyTimes()
						testHelper.mockSubCategoryRepository.EXPECT().GetByCode(args.ctx, "12101").Return(&models.SubCategory{}, nil).AnyTimes()
						testHelper.mockCacheRepository.EXPECT().GetIncrement(args.ctx, gomock.Any()).Return(int64(1), nil).AnyTimes()

						testHelper.mockAccRepository.EXPECT().BulkInsertAcctAccount(args.ctx, gomock.Any()).Return(nil)
						testHelper.mockAccRepository.EXPECT().BulkInsertAccount(args.ctx, gomock.Any()).Return(nil)
						testHelper.mockLoanPartnerAccountRepository.EXPECT().BulkInsertLoanPartnerAccount(args.ctx, gomock.Any()).Return(models.ErrUnableToCreate)
						return steps(ctx, testHelper.mockMySQLRepository)
					}).Return(models.ErrUnableToCreate)
			},
			wantErr: true,
		},
		{
			name: "success case - create loan partner account",
			args: args{
				ctx: context.Background(),
				in:  req,
			},
			doMock: func(args args) {
				testHelper.mockMySQLRepository.EXPECT().Atomic(gomock.Any(), gomock.Any()).DoAndReturn(
					func(ctx context.Context, steps func(ctx context.Context, r mysql.SQLRepository) error) error {
						testHelper.mockProductTypeRepository.EXPECT().GetLatestProductCode(args.ctx).Return(latestProductCode, nil)
						testHelper.mockProductTypeRepository.EXPECT().GetByCode(args.ctx, reqProductType.Code).Return(nil, nil)
						testHelper.mockProductTypeRepository.EXPECT().Create(args.ctx, &reqProductType).Return(nil)

						testHelper.mockSubCategoryRepository.EXPECT().GetLatestSubCategCode(args.ctx, latestCateg).Return(latestSubCateg, nil)
						testHelper.mockCategoryRepository.EXPECT().GetByCode(args.ctx, reqSubCategory.CategoryCode).Return(&models.Category{}, nil)
						testHelper.mockSubCategoryRepository.EXPECT().GetByCode(args.ctx, reqSubCategory.Code).Return(nil, nil)
						testHelper.mockSubCategoryRepository.EXPECT().GetByAccountType(args.ctx, reqSubCategory.AccountType).Return(nil, nil)
						testHelper.mockProductTypeRepository.EXPECT().GetByCode(args.ctx, reqSubCategory.ProductTypeCode).Return(&models.ProductType{}, nil)
						testHelper.mockGoDbLedger.EXPECT().GetCurrency(args.ctx, reqSubCategory.Currency).Return(godbledger.CurrencyIDR, nil)
						testHelper.mockSubCategoryRepository.EXPECT().Create(args.ctx, &reqSubCategory).Return(nil)

						testHelper.mockLoanPartnerAccountRepository.EXPECT().GetByParams(args.ctx, gomock.Any()).Return([]models.LoanPartnerAccount{}, nil).AnyTimes()
						testHelper.mockSubCategoryRepository.EXPECT().GetByCode(args.ctx, gomock.Any()).Return(&models.SubCategory{}, nil).AnyTimes()
						testHelper.mockCacheRepository.EXPECT().GetIncrement(args.ctx, gomock.Any()).Return(int64(1), nil).AnyTimes()

						testHelper.mockAccRepository.EXPECT().BulkInsertAcctAccount(args.ctx, gomock.Any()).Return(nil)
						testHelper.mockAccRepository.EXPECT().BulkInsertAccount(args.ctx, gomock.Any()).Return(nil)
						testHelper.mockLoanPartnerAccountRepository.EXPECT().BulkInsertLoanPartnerAccount(args.ctx, gomock.Any()).Return(nil)
						return steps(ctx, testHelper.mockMySQLRepository)
					})
				testHelper.mockAcuanClient.EXPECT().PublishAccount(args.ctx, gomock.Any()).AnyTimes()
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
			_, err := testHelper.accountService.CreateLoanPartnerAccount(tt.args.ctx, tt.args.in)
			assert.Equal(t, tt.wantErr, err != nil)
		})
	}
}
