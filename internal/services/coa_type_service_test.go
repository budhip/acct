package services_test

import (
	"context"
	"testing"

	"bitbucket.org/Amartha/go-accounting/internal/models"
	"go.uber.org/mock/gomock"

	"github.com/stretchr/testify/assert"
)

func TestCOATypeService_Create(t *testing.T) {
	testHelper := serviceTestHelper(t)

	type args struct {
		ctx context.Context
		req models.CreateCOATypeIn
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
			name: "test success",
			args: args{
				ctx: context.Background(),
				req: models.CreateCOATypeIn{},
			},
			doMock: func(args args, mockData mockData) {
				testHelper.mockCOATypeRepository.EXPECT().GetCOATypeByCode(args.ctx, args.req.Code).Return(nil, nil)
				testHelper.mockCOATypeRepository.EXPECT().Create(args.ctx, &args.req).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "test code is exist",
			args: args{
				ctx: context.Background(),
				req: models.CreateCOATypeIn{},
			},
			mockData: mockData{},
			doMock: func(args args, mockData mockData) {
				testHelper.mockCOATypeRepository.EXPECT().GetCOATypeByCode(args.ctx, args.req.Code).Return(&models.COAType{}, nil)
			},
			wantErr: true,
		},
		{
			name: "test error CheckCOATypeByCode",
			args: args{
				ctx: context.Background(),
				req: models.CreateCOATypeIn{},
			},
			mockData: mockData{},
			doMock: func(args args, mockData mockData) {
				testHelper.mockCOATypeRepository.EXPECT().GetCOATypeByCode(args.ctx, args.req.Code).Return(nil, assert.AnError)
			},
			wantErr: true,
		},
		{
			name: "test error Create",
			args: args{
				ctx: context.Background(),
				req: models.CreateCOATypeIn{},
			},
			mockData: mockData{},
			doMock: func(args args, mockData mockData) {
				testHelper.mockCOATypeRepository.EXPECT().GetCOATypeByCode(args.ctx, args.req.Code).Return(nil, nil)
				testHelper.mockCOATypeRepository.EXPECT().Create(args.ctx, &args.req).Return(assert.AnError)
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.doMock != nil {
				tt.doMock(tt.args, tt.mockData)
			}

			_, err := testHelper.coaTypeService.Create(tt.args.ctx, tt.args.req)
			assert.Equal(t, tt.wantErr, err != nil)
		})
	}
}

func TestCOATypeService_GetAll(t *testing.T) {
	testHelper := serviceTestHelper(t)

	type args struct {
		ctx context.Context
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
			name: "test success",
			doMock: func(args args, mockData mockData) {
				testHelper.mockCOATypeRepository.EXPECT().GetAll(args.ctx).Return([]models.COAType{
					{},
				}, nil)
				testHelper.mockCategoryRepository.EXPECT().GetByCoaCode(args.ctx, gomock.Any()).Return(&[]models.CategoryCOA{
					{},
				}, nil)
			},
			wantErr: false,
		},
		{
			name: "test error GetAll",
			doMock: func(args args, mockData mockData) {
				testHelper.mockCOATypeRepository.EXPECT().GetAll(args.ctx).Return([]models.COAType{}, assert.AnError)
			},
			wantErr: true,
		},
		{
			name: "test error GetByCoaCode",
			doMock: func(args args, mockData mockData) {
				testHelper.mockCOATypeRepository.EXPECT().GetAll(args.ctx).Return([]models.COAType{
					{},
				}, nil)
				testHelper.mockCategoryRepository.EXPECT().GetByCoaCode(args.ctx, gomock.Any()).Return(&[]models.CategoryCOA{},
					assert.AnError)
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.doMock != nil {
				tt.doMock(tt.args, tt.mockData)
			}
			_, err := testHelper.coaTypeService.GetAll(tt.args.ctx)
			assert.Equal(t, tt.wantErr, err != nil)
		})
	}
}

func TestCOATypeService_Update(t *testing.T) {
	testHelper := serviceTestHelper(t)

	type (
		args struct {
			ctx context.Context
			req models.UpdateCOAType
		}
	)
	tests := []struct {
		name    string
		args    args
		doMock  func(args args)
		wantErr bool
	}{
		{
			name: "success case - update coa type",
			args: args{
				ctx: context.Background(),
				req: models.UpdateCOAType{
					Name:          "Lender Yang Baik",
					NormalBalance: "debit",
					Status:        "active",
					Code:          "001",
				},
			},
			doMock: func(args args) {
				testHelper.mockCOATypeRepository.EXPECT().GetCOATypeByCode(args.ctx, args.req.Code).Return(&models.COAType{}, nil)
				testHelper.mockCOATypeRepository.EXPECT().Update(args.ctx, args.req).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "success case - empty request",
			args: args{
				ctx: context.Background(),
				req: models.UpdateCOAType{
					Code: "001",
				},
			},
			doMock: func(args args) {
				testHelper.mockCOATypeRepository.EXPECT().GetCOATypeByCode(args.ctx, args.req.Code).Return(&models.COAType{}, nil)
			},
			wantErr: false,
		},
		{
			name: "error case - database error get coa type",
			args: args{
				ctx: context.Background(),
				req: models.UpdateCOAType{
					Name:          "Lender Yang Baik",
					NormalBalance: "debit",
					Status:        "active",
					Code:          "001",
				},
			},
			doMock: func(args args) {
				testHelper.mockCOATypeRepository.EXPECT().GetCOATypeByCode(args.ctx, args.req.Code).Return(&models.COAType{}, assert.AnError)
			},
			wantErr: true,
		},
		{
			name: "error case - coa type not found",
			args: args{
				ctx: context.Background(),
				req: models.UpdateCOAType{
					Name:          "Lender Yang Baik",
					NormalBalance: "debit",
					Status:        "active",
					Code:          "001",
				},
			},
			doMock: func(args args) {
				testHelper.mockCOATypeRepository.EXPECT().GetCOATypeByCode(args.ctx, args.req.Code).Return(nil, nil)
			},
			wantErr: true,
		},
		{
			name: "error case - database error update coa type",
			args: args{
				ctx: context.Background(),
				req: models.UpdateCOAType{
					Name:          "Lender Yang Baik",
					NormalBalance: "debit",
					Status:        "active",
					Code:          "001",
				},
			},
			doMock: func(args args) {
				testHelper.mockCOATypeRepository.EXPECT().GetCOATypeByCode(args.ctx, args.req.Code).Return(&models.COAType{}, nil)
				testHelper.mockCOATypeRepository.EXPECT().Update(args.ctx, args.req).Return(models.GetErrMap(models.ErrKeyDatabaseError))
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
			_, err := testHelper.coaTypeService.Update(tt.args.ctx, tt.args.req)
			assert.Equal(t, tt.wantErr, err != nil)
		})
	}
}
