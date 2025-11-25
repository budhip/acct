package services_test

import (
	"context"
	"testing"

	"bitbucket.org/Amartha/go-accounting/internal/models"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestProductTypeService_GetAll(t *testing.T) {
	testHelper := serviceTestHelper(t)
	tests := []struct {
		name    string
		doMock  func()
		wantErr bool
	}{
		{
			name: "success case - get all data product type",
			doMock: func() {
				testHelper.mockProductTypeRepository.EXPECT().
					List(gomock.AssignableToTypeOf(context.Background())).
					Return([]models.ProductType{}, nil)
			},
			wantErr: false,
		},
		{
			name: "error case - get data data product type",
			doMock: func() {
				testHelper.mockProductTypeRepository.EXPECT().
					List(gomock.AssignableToTypeOf(context.Background())).
					Return([]models.ProductType{}, assert.AnError)
			},
			wantErr: true,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.doMock != nil {
				tc.doMock()
			}

			_, err := testHelper.productTypeService.GetAll(context.Background())
			assert.Equal(t, tc.wantErr, err != nil)
		})
	}
}

func TestProductTypeService_Create(t *testing.T) {
	testHelper := serviceTestHelper(t)
	databaseError := models.GetErrMap(models.ErrKeyDatabaseError)

	type args struct {
		ctx context.Context
		req models.CreateProductTypeRequest
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
			name: "success case",
			args: args{
				ctx: context.Background(),
				req: models.CreateProductTypeRequest{
					Code:       "test",
					Name:       "test",
					Status:     "active",
					EntityCode: "test",
				},
			},
			doMock: func(args args, mockData mockData) {
				testHelper.mockProductTypeRepository.EXPECT().GetByCode(args.ctx, args.req.Code).Return(nil, nil)
				testHelper.mockEntityRepository.EXPECT().GetByCode(args.ctx, args.req.EntityCode).Return(&models.Entity{}, nil)
				testHelper.mockProductTypeRepository.EXPECT().Create(args.ctx, &args.req).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "error case - product type code exist",
			args: args{
				ctx: context.Background(),
				req: models.CreateProductTypeRequest{
					Code: "211",
				},
			},
			doMock: func(args args, mockData mockData) {
				testHelper.mockProductTypeRepository.EXPECT().GetByCode(args.ctx, args.req.Code).Return(&models.ProductType{}, nil)
			},
			wantErr: true,
		},
		{
			name: "error case - entity code exist",
			args: args{
				ctx: context.Background(),
				req: models.CreateProductTypeRequest{
					Code:       "211",
					EntityCode: "test",
				},
			},
			doMock: func(args args, mockData mockData) {
				testHelper.mockProductTypeRepository.EXPECT().GetByCode(args.ctx, args.req.Code).Return(nil, nil)
				testHelper.mockEntityRepository.EXPECT().GetByCode(args.ctx, args.req.EntityCode).Return(nil, nil)
			},
			wantErr: true,
		},
		{
			name: "error case - database error get product",
			args: args{
				ctx: context.Background(),
				req: models.CreateProductTypeRequest{
					Code: "211",
				},
			},
			doMock: func(args args, mockData mockData) {
				testHelper.mockProductTypeRepository.EXPECT().GetByCode(args.ctx, args.req.Code).Return(nil, databaseError)
			},
			wantErr: true,
		},
		{
			name: "error case - database error get entity",
			args: args{
				ctx: context.Background(),
				req: models.CreateProductTypeRequest{
					Code:       "211",
					EntityCode: "test",
				},
			},
			doMock: func(args args, mockData mockData) {
				testHelper.mockProductTypeRepository.EXPECT().GetByCode(args.ctx, args.req.Code).Return(nil, nil)
				testHelper.mockEntityRepository.EXPECT().GetByCode(args.ctx, args.req.EntityCode).Return(nil, databaseError)
			},
			wantErr: true,
		},
		{
			name: "error case - database error insert product",
			args: args{
				ctx: context.Background(),
				req: models.CreateProductTypeRequest{
					Code:       "test",
					Name:       "test",
					Status:     "active",
					EntityCode: "test",
				},
			},
			doMock: func(args args, mockData mockData) {
				testHelper.mockProductTypeRepository.EXPECT().GetByCode(args.ctx, args.req.Code).Return(nil, nil)
				testHelper.mockEntityRepository.EXPECT().GetByCode(args.ctx, args.req.EntityCode).Return(&models.Entity{}, nil)
				testHelper.mockProductTypeRepository.EXPECT().Create(args.ctx, &args.req).Return(databaseError)
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
			_, err := testHelper.productTypeService.Create(tt.args.ctx, tt.args.req)
			assert.Equal(t, tt.wantErr, err != nil)
		})
	}
}

func TestProductTypeService_Update(t *testing.T) {
	testHelper := serviceTestHelper(t)

	req := models.UpdateProductType{
		Name:       "Chickin",
		Status:     "active",
		Code:       "1001",
		EntityCode: "001",
	}
	type (
		args struct {
			ctx context.Context
			req models.UpdateProductType
		}
		mockData struct{}
	)
	tests := []struct {
		name     string
		args     args
		mockData mockData
		doMock   func(args args, mockData mockData)
		wantErr  bool
	}{
		{
			name: "success case - update product type",
			args: args{
				ctx: context.Background(),
				req: req,
			},
			doMock: func(args args, mockData mockData) {
				testHelper.mockProductTypeRepository.EXPECT().CheckProductTypeIsExist(args.ctx, args.req.Code).Return(nil)
				testHelper.mockEntityRepository.EXPECT().CheckEntityByCode(args.ctx, args.req.EntityCode, models.StatusActive).Return(nil)
				testHelper.mockProductTypeRepository.EXPECT().Update(args.ctx, args.req).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "success case - empty request",
			args: args{
				ctx: context.Background(),
				req: models.UpdateProductType{
					Code:       "1001",
					EntityCode: "",
				},
			},
			doMock: func(args args, mockData mockData) {
				testHelper.mockProductTypeRepository.EXPECT().CheckProductTypeIsExist(args.ctx, args.req.Code).Return(nil)
				testHelper.mockProductTypeRepository.EXPECT().Update(args.ctx, args.req).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "error case - database error get product type",
			args: args{
				ctx: context.Background(),
				req: req,
			},
			doMock: func(args args, mockData mockData) {
				testHelper.mockProductTypeRepository.EXPECT().CheckProductTypeIsExist(args.ctx, args.req.Code).Return(models.GetErrMap(models.ErrKeyDatabaseError))
			},
			wantErr: true,
		},
		{
			name: "error case - database error get entity code",
			args: args{
				ctx: context.Background(),
				req: req,
			},
			doMock: func(args args, mockData mockData) {
				testHelper.mockProductTypeRepository.EXPECT().CheckProductTypeIsExist(args.ctx, args.req.Code).Return(nil)
				testHelper.mockEntityRepository.EXPECT().CheckEntityByCode(args.ctx, args.req.EntityCode, models.StatusActive).Return(models.GetErrMap(models.ErrKeyDatabaseError))
			},
			wantErr: true,
		},
		{
			name: "error case - database error update product type",
			args: args{
				ctx: context.Background(),
				req: req,
			},
			doMock: func(args args, mockData mockData) {
				testHelper.mockProductTypeRepository.EXPECT().CheckProductTypeIsExist(args.ctx, args.req.Code).Return(nil)
				testHelper.mockEntityRepository.EXPECT().CheckEntityByCode(args.ctx, args.req.EntityCode, models.StatusActive).Return(nil)
				testHelper.mockProductTypeRepository.EXPECT().Update(args.ctx, args.req).Return(models.GetErrMap(models.ErrKeyDatabaseError))
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
			_, err := testHelper.productTypeService.Update(tt.args.ctx, tt.args.req)
			assert.Equal(t, tt.wantErr, err != nil)
		})
	}
}
