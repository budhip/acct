package services_test

import (
	"context"
	"errors"
	"testing"

	"bitbucket.org/Amartha/go-accounting/internal/repositories/mysql"
	"github.com/darcys22/godbledger/godbledger/core"

	"bitbucket.org/Amartha/go-accounting/internal/models"
	"bitbucket.org/Amartha/go-accounting/internal/pkg/godbledger"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestSubCategoryService_Create(t *testing.T) {
	testHelper := serviceTestHelper(t)
	databaseError := models.GetErrMap(models.ErrKeyDatabaseError)

	type args struct {
		ctx context.Context
		req models.CreateSubCategory
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
			name: "success case - account type and product type code empty",
			args: args{
				ctx: context.Background(),
				req: models.CreateSubCategory{
					CategoryCode: "211",
					Code:         "21101",
				},
			},
			doMock: func(args args, mockData mockData) {
				testHelper.mockCategoryRepository.EXPECT().GetByCode(args.ctx, args.req.CategoryCode).Return(&models.Category{}, nil)
				testHelper.mockSubCategoryRepository.EXPECT().GetByCode(args.ctx, args.req.Code).Return(nil, nil)
				testHelper.mockSubCategoryRepository.EXPECT().Create(args.ctx, &args.req).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "success case - account type not empty",
			args: args{
				ctx: context.Background(),
				req: models.CreateSubCategory{
					CategoryCode: "211",
					Code:         "21101",
					AccountType:  "LENDER_ACCOUNT",
				},
			},
			doMock: func(args args, mockData mockData) {
				testHelper.mockCategoryRepository.EXPECT().GetByCode(args.ctx, args.req.CategoryCode).Return(&models.Category{}, nil)
				testHelper.mockSubCategoryRepository.EXPECT().GetByCode(args.ctx, args.req.Code).Return(nil, nil)
				testHelper.mockSubCategoryRepository.EXPECT().Create(args.ctx, &args.req).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "success case - product type code not empty",
			args: args{
				ctx: context.Background(),
				req: models.CreateSubCategory{
					CategoryCode:    "211",
					Code:            "21101",
					ProductTypeCode: "1001",
				},
			},
			doMock: func(args args, mockData mockData) {
				testHelper.mockCategoryRepository.EXPECT().GetByCode(args.ctx, args.req.CategoryCode).Return(&models.Category{}, nil)
				testHelper.mockSubCategoryRepository.EXPECT().GetByCode(args.ctx, args.req.Code).Return(nil, nil)
				testHelper.mockProductTypeRepository.EXPECT().GetByCode(args.ctx, args.req.ProductTypeCode).Return(&models.ProductType{}, nil)
				testHelper.mockSubCategoryRepository.EXPECT().Create(args.ctx, &args.req).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "success case - currency not empty",
			args: args{
				ctx: context.Background(),
				req: models.CreateSubCategory{
					CategoryCode:    "211",
					Code:            "21101",
					ProductTypeCode: "1001",
					Currency:        "IDR",
				},
			},
			doMock: func(args args, mockData mockData) {
				testHelper.mockCategoryRepository.EXPECT().GetByCode(args.ctx, args.req.CategoryCode).Return(&models.Category{}, nil)
				testHelper.mockSubCategoryRepository.EXPECT().GetByCode(args.ctx, args.req.Code).Return(nil, nil)
				testHelper.mockProductTypeRepository.EXPECT().GetByCode(args.ctx, args.req.ProductTypeCode).Return(&models.ProductType{}, nil)
				testHelper.mockGoDbLedger.EXPECT().GetCurrency(args.ctx, args.req.Currency).Return(godbledger.CurrencyIDR, nil)
				testHelper.mockSubCategoryRepository.EXPECT().Create(args.ctx, &args.req).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "success case - currency not found",
			args: args{
				ctx: context.Background(),
				req: models.CreateSubCategory{
					CategoryCode:    "211",
					Code:            "21101",
					ProductTypeCode: "1001",
					Currency:        "IDR",
				},
			},
			doMock: func(args args, mockData mockData) {
				testHelper.mockCategoryRepository.EXPECT().GetByCode(args.ctx, args.req.CategoryCode).Return(&models.Category{}, nil)
				testHelper.mockSubCategoryRepository.EXPECT().GetByCode(args.ctx, args.req.Code).Return(nil, nil)
				testHelper.mockProductTypeRepository.EXPECT().GetByCode(args.ctx, args.req.ProductTypeCode).Return(&models.ProductType{}, nil)
				testHelper.mockGoDbLedger.EXPECT().GetCurrency(args.ctx, args.req.Currency).Return(nil, models.ErrNoRows)
				testHelper.mockSubCategoryRepository.EXPECT().Create(args.ctx, &args.req).Return(nil)
			},
			wantErr: true,
		},
		{
			name: "error case - category not exist",
			args: args{
				ctx: context.Background(),
				req: models.CreateSubCategory{
					CategoryCode: "211",
				},
			},
			doMock: func(args args, mockData mockData) {
				testHelper.mockCategoryRepository.EXPECT().GetByCode(args.ctx, args.req.CategoryCode).Return(nil, nil)
			},
			wantErr: true,
		},
		{
			name: "error case - database error get category",
			args: args{
				ctx: context.Background(),
				req: models.CreateSubCategory{
					CategoryCode: "211",
				},
			},
			doMock: func(args args, mockData mockData) {
				testHelper.mockCategoryRepository.EXPECT().GetByCode(args.ctx, args.req.CategoryCode).Return(nil, databaseError)
			},
			wantErr: true,
		},
		{
			name: "error case - sub category not exist",
			args: args{
				ctx: context.Background(),
				req: models.CreateSubCategory{
					CategoryCode: "211",
					Code:         "21101",
				},
			},
			doMock: func(args args, mockData mockData) {
				testHelper.mockCategoryRepository.EXPECT().GetByCode(args.ctx, args.req.CategoryCode).Return(&models.Category{}, nil)
				testHelper.mockSubCategoryRepository.EXPECT().GetByCode(args.ctx, args.req.Code).Return(&models.SubCategory{}, nil)
			},
			wantErr: true,
		},
		{
			name: "error case - database error get sub category",
			args: args{
				ctx: context.Background(),
				req: models.CreateSubCategory{
					CategoryCode: "211",
					Code:         "21101",
				},
			},
			doMock: func(args args, mockData mockData) {
				testHelper.mockCategoryRepository.EXPECT().GetByCode(args.ctx, args.req.CategoryCode).Return(&models.Category{}, nil)
				testHelper.mockSubCategoryRepository.EXPECT().GetByCode(args.ctx, args.req.Code).Return(nil, databaseError)
			},
			wantErr: true,
		},
		{
			name: "error case - product type code not exist",
			args: args{
				ctx: context.Background(),
				req: models.CreateSubCategory{
					CategoryCode:    "211",
					Code:            "21101",
					ProductTypeCode: "1001",
				},
			},
			doMock: func(args args, mockData mockData) {
				testHelper.mockCategoryRepository.EXPECT().GetByCode(args.ctx, args.req.CategoryCode).Return(&models.Category{}, nil)
				testHelper.mockSubCategoryRepository.EXPECT().GetByCode(args.ctx, args.req.Code).Return(nil, nil)
				testHelper.mockProductTypeRepository.EXPECT().GetByCode(args.ctx, args.req.ProductTypeCode).Return(nil, nil)
			},
			wantErr: true,
		},
		{
			name: "error case - database error get product type code",
			args: args{
				ctx: context.Background(),
				req: models.CreateSubCategory{
					CategoryCode:    "211",
					Code:            "21101",
					ProductTypeCode: "1001",
				},
			},
			doMock: func(args args, mockData mockData) {
				testHelper.mockCategoryRepository.EXPECT().GetByCode(args.ctx, args.req.CategoryCode).Return(&models.Category{}, nil)
				testHelper.mockSubCategoryRepository.EXPECT().GetByCode(args.ctx, args.req.Code).Return(nil, nil)
				testHelper.mockProductTypeRepository.EXPECT().GetByCode(args.ctx, args.req.ProductTypeCode).Return(nil, databaseError)
			},
			wantErr: true,
		},
		{
			name: "error case - database error create sub category",
			args: args{
				ctx: context.Background(),
				req: models.CreateSubCategory{
					CategoryCode: "211",
					Code:         "21101",
				},
			},
			doMock: func(args args, mockData mockData) {
				testHelper.mockCategoryRepository.EXPECT().GetByCode(args.ctx, args.req.CategoryCode).Return(&models.Category{}, nil)
				testHelper.mockSubCategoryRepository.EXPECT().GetByCode(args.ctx, args.req.Code).Return(nil, nil)
				testHelper.mockSubCategoryRepository.EXPECT().Create(args.ctx, &args.req).Return(databaseError)
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
			_, err := testHelper.subCategoryService.Create(tt.args.ctx, tt.args.req)
			assert.Equal(t, tt.wantErr, err != nil)
		})
	}
}

func TestSubCategoryService_GetAll(t *testing.T) {
	testHelper := serviceTestHelper(t)
	tests := []struct {
		name    string
		doMock  func()
		wantErr bool
	}{
		{
			name: "success - get all entities",
			doMock: func() {
				testHelper.mockSubCategoryRepository.EXPECT().
					GetAll(gomock.AssignableToTypeOf(context.Background()), models.GetAllSubCategoryParam{}).
					Return(&[]models.SubCategory{}, nil)
			},
			wantErr: false,
		},
		{
			name: "error - get data from repository",
			doMock: func() {
				testHelper.mockSubCategoryRepository.EXPECT().
					GetAll(gomock.AssignableToTypeOf(context.Background()), models.GetAllSubCategoryParam{}).
					Return(nil, assert.AnError)
			},
			wantErr: true,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.doMock != nil {
				tc.doMock()
			}
			_, err := testHelper.subCategoryService.GetAll(context.Background(), models.GetAllSubCategoryParam{})
			assert.Equal(t, tc.wantErr, err != nil)
		})
	}
}

func TestSubCategoryService_GetByAccountType(t *testing.T) {
	testHelper := serviceTestHelper(t)
	type args struct {
		ctx         context.Context
		accountType string
	}
	tests := []struct {
		name    string
		args    args
		doMock  func(a *args)
		wantErr bool
	}{
		{
			name: "success - get sub category by account type",
			args: args{
				ctx:         context.TODO(),
				accountType: "LENDER_INSTITUTIONAL",
			},
			doMock: func(a *args) {
				testHelper.mockSubCategoryRepository.EXPECT().
					GetByAccountType(a.ctx, a.accountType).
					Return(&models.SubCategory{}, nil)
			},
			wantErr: false,
		},
		{
			name: "error - failed get data from repository",
			args: args{
				ctx:         context.TODO(),
				accountType: "LENDER_DELUSIONAL",
			},
			doMock: func(a *args) {
				testHelper.mockSubCategoryRepository.EXPECT().
					GetByAccountType(a.ctx, a.accountType).
					Return(nil, assert.AnError)
			},
			wantErr: true,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.doMock != nil {
				tc.doMock(&tc.args)
			}
			_, err := testHelper.subCategoryService.GetByAccountType(tc.args.ctx, tc.args.accountType)
			assert.Equal(t, tc.wantErr, err != nil)
		})
	}
}

func TestSubCategoryService_Update(t *testing.T) {
	testHelper := serviceTestHelper(t)

	req := models.UpdateSubCategory{
		Name:            "RETAIL",
		Description:     "Retail",
		Status:          "active",
		Code:            "10000",
		ProductTypeCode: &[]string{"1001"}[0],
		Currency:        &[]string{"IDR"}[0],
	}
	type (
		args struct {
			ctx context.Context
			req models.UpdateSubCategory
		}
	)
	tests := []struct {
		name    string
		args    args
		doMock  func(args args)
		wantErr bool
	}{
		{
			name: "success case - empty request",
			args: args{
				ctx: context.Background(),
				req: models.UpdateSubCategory{
					ProductTypeCode: nil,
					Name:            "",
					Description:     "",
					Status:          "",
					Currency:        nil,
					Code:            "10000",
				},
			},
			doMock: func(args args) {
				testHelper.mockSubCategoryRepository.EXPECT().GetByCode(args.ctx, args.req.Code).Return(&models.SubCategory{}, nil)
			},
			wantErr: false,
		},
		{
			name: "success case - update sub category",
			args: args{
				ctx: context.Background(),
				req: req,
			},
			doMock: func(args args) {
				testHelper.mockSubCategoryRepository.EXPECT().GetByCode(args.ctx, args.req.Code).Return(&models.SubCategory{}, nil)
				testHelper.mockProductTypeRepository.EXPECT().GetByCode(args.ctx, *args.req.ProductTypeCode).Return(&models.ProductType{}, nil)
				testHelper.mockGoDbLedger.EXPECT().GetCurrency(args.ctx, *args.req.Currency).Return(godbledger.Currency(&core.Currency{Name: "IDR", Decimals: 2}), nil)
				testHelper.mockMySQLRepository.EXPECT().Atomic(gomock.Any(), gomock.Any()).DoAndReturn(
					func(ctx context.Context, steps func(ctx context.Context, r mysql.SQLRepository) error) error {
						testHelper.mockSubCategoryRepository.EXPECT().Update(args.ctx, args.req).Return(nil)
						testHelper.mockAccRepository.EXPECT().UpdateBySubCategory(args.ctx, gomock.Any()).Return(nil)
						testHelper.mockGoFpTransaction.EXPECT().UpdateAccountBySubCategory(args.ctx, gomock.Any()).Return(nil)
						return steps(ctx, testHelper.mockMySQLRepository)
					})
			},
			wantErr: false,
		},
		{
			name: "error case - sub category code not found",
			args: args{
				ctx: context.Background(),
				req: req,
			},
			doMock: func(args args) {
				testHelper.mockSubCategoryRepository.EXPECT().GetByCode(args.ctx, args.req.Code).Return(nil, nil)
			},
			wantErr: true,
		},
		{
			name: "error case - database error get sub category code",
			args: args{
				ctx: context.Background(),
				req: req,
			},
			doMock: func(args args) {
				testHelper.mockSubCategoryRepository.EXPECT().GetByCode(args.ctx, args.req.Code).Return(nil, models.GetErrMap(models.ErrKeyDatabaseError))
			},
			wantErr: true,
		},
		{
			name: "error case - product type code not found",
			args: args{
				ctx: context.Background(),
				req: req,
			},
			doMock: func(args args) {
				testHelper.mockSubCategoryRepository.EXPECT().GetByCode(args.ctx, args.req.Code).Return(&models.SubCategory{}, nil)
				testHelper.mockProductTypeRepository.EXPECT().GetByCode(args.ctx, *args.req.ProductTypeCode).Return(nil, nil)
			},
			wantErr: true,
		},
		{
			name: "error case - database error get product type code",
			args: args{
				ctx: context.Background(),
				req: req,
			},
			doMock: func(args args) {
				testHelper.mockSubCategoryRepository.EXPECT().GetByCode(args.ctx, args.req.Code).Return(&models.SubCategory{}, nil)
				testHelper.mockProductTypeRepository.EXPECT().GetByCode(args.ctx, *args.req.ProductTypeCode).Return(&models.ProductType{}, models.GetErrMap(models.ErrKeyDatabaseError))
			},
			wantErr: true,
		},
		{
			name: "error case - currency not found",
			args: args{
				ctx: context.Background(),
				req: req,
			},
			doMock: func(args args) {
				testHelper.mockSubCategoryRepository.EXPECT().GetByCode(args.ctx, args.req.Code).Return(&models.SubCategory{}, nil)
				testHelper.mockProductTypeRepository.EXPECT().GetByCode(args.ctx, *args.req.ProductTypeCode).Return(&models.ProductType{}, nil)
				testHelper.mockGoDbLedger.EXPECT().GetCurrency(args.ctx, *args.req.Currency).Return(nil, models.GetErrMap(models.ErrKeyCurrencyNotFound))
			},
			wantErr: true,
		},
		{
			name: "error case - database error get currency",
			args: args{
				ctx: context.Background(),
				req: req,
			},
			doMock: func(args args) {
				testHelper.mockSubCategoryRepository.EXPECT().GetByCode(args.ctx, args.req.Code).Return(&models.SubCategory{}, nil)
				testHelper.mockProductTypeRepository.EXPECT().GetByCode(args.ctx, *args.req.ProductTypeCode).Return(&models.ProductType{}, nil)
				testHelper.mockGoDbLedger.EXPECT().GetCurrency(args.ctx, *args.req.Currency).Return(nil, models.GetErrMap(models.ErrKeyDatabaseError))
			},
			wantErr: true,
		},
		{
			name: "error case - database error update sub category",
			args: args{
				ctx: context.Background(),
				req: req,
			},
			doMock: func(args args) {
				testHelper.mockSubCategoryRepository.EXPECT().GetByCode(args.ctx, args.req.Code).Return(&models.SubCategory{}, nil)
				testHelper.mockProductTypeRepository.EXPECT().GetByCode(args.ctx, *args.req.ProductTypeCode).Return(&models.ProductType{}, nil)
				testHelper.mockGoDbLedger.EXPECT().GetCurrency(args.ctx, *args.req.Currency).Return(godbledger.Currency(&core.Currency{Name: "IDR", Decimals: 2}), nil)
				testHelper.mockMySQLRepository.EXPECT().Atomic(gomock.Any(), gomock.Any()).DoAndReturn(
					func(ctx context.Context, steps func(ctx context.Context, r mysql.SQLRepository) error) error {
						testHelper.mockSubCategoryRepository.EXPECT().Update(args.ctx, args.req).Return(models.GetErrMap(models.ErrKeyDatabaseError))
						return steps(ctx, testHelper.mockMySQLRepository)
					})
			},
			wantErr: true,
		},
		{
			name: "error case - database error update account by sub category",
			args: args{
				ctx: context.Background(),
				req: req,
			},
			doMock: func(args args) {
				testHelper.mockSubCategoryRepository.EXPECT().GetByCode(args.ctx, args.req.Code).Return(&models.SubCategory{}, nil)
				testHelper.mockProductTypeRepository.EXPECT().GetByCode(args.ctx, *args.req.ProductTypeCode).Return(&models.ProductType{}, nil)
				testHelper.mockGoDbLedger.EXPECT().GetCurrency(args.ctx, *args.req.Currency).Return(godbledger.Currency(&core.Currency{Name: "IDR", Decimals: 2}), nil)
				testHelper.mockMySQLRepository.EXPECT().Atomic(gomock.Any(), gomock.Any()).DoAndReturn(
					func(ctx context.Context, steps func(ctx context.Context, r mysql.SQLRepository) error) error {
						testHelper.mockSubCategoryRepository.EXPECT().Update(args.ctx, args.req).Return(nil)
						testHelper.mockAccRepository.EXPECT().UpdateBySubCategory(args.ctx, gomock.Any()).Return(models.GetErrMap(models.ErrKeyDatabaseError))
						return steps(ctx, testHelper.mockMySQLRepository)
					})
			},
			wantErr: true,
		},
		{
			name: "error case - database error update account fp transaction by sub category",
			args: args{
				ctx: context.Background(),
				req: req,
			},
			doMock: func(args args) {
				testHelper.mockSubCategoryRepository.EXPECT().GetByCode(args.ctx, args.req.Code).Return(&models.SubCategory{}, nil)
				testHelper.mockProductTypeRepository.EXPECT().GetByCode(args.ctx, *args.req.ProductTypeCode).Return(&models.ProductType{}, nil)
				testHelper.mockGoDbLedger.EXPECT().GetCurrency(args.ctx, *args.req.Currency).Return(godbledger.Currency(&core.Currency{Name: "IDR", Decimals: 2}), nil)
				testHelper.mockMySQLRepository.EXPECT().Atomic(gomock.Any(), gomock.Any()).DoAndReturn(
					func(ctx context.Context, steps func(ctx context.Context, r mysql.SQLRepository) error) error {
						testHelper.mockSubCategoryRepository.EXPECT().Update(args.ctx, args.req).Return(nil)
						testHelper.mockAccRepository.EXPECT().UpdateBySubCategory(args.ctx, gomock.Any()).Return(nil)
						testHelper.mockGoFpTransaction.EXPECT().UpdateAccountBySubCategory(args.ctx, gomock.Any()).Return(errors.New("error"))
						return steps(ctx, testHelper.mockMySQLRepository)
					})
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
			_, err := testHelper.subCategoryService.Update(tt.args.ctx, tt.args.req)
			assert.Equal(t, tt.wantErr, err != nil)
		})
	}
}
