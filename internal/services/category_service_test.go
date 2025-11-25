package services_test

import (
	"context"
	"testing"

	"bitbucket.org/Amartha/go-accounting/internal/models"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestCategoryService_Create(t *testing.T) {
	testHelper := serviceTestHelper(t)

	type args struct {
		ctx context.Context
		req models.CreateCategoryIn
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
				req: models.CreateCategoryIn{
					Code:        "",
					Name:        "",
					Description: "",
				},
			},
			doMock: func(args args, mockData mockData) {
				testHelper.mockCategoryRepository.EXPECT().GetByCode(args.ctx, args.req.Code).Return(nil, nil)
				testHelper.mockCOATypeRepository.EXPECT().CheckCOATypeByCode(args.ctx, args.req.Code).Return(nil)
				testHelper.mockCategoryRepository.EXPECT().Create(args.ctx, &args.req).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "category code is exist",
			args: args{
				ctx: context.Background(),
				req: models.CreateCategoryIn{},
			},
			mockData: mockData{},
			doMock: func(args args, mockData mockData) {
				testHelper.mockCategoryRepository.EXPECT().GetByCode(args.ctx, args.req.Code).Return(&models.Category{}, nil)
			},
			wantErr: true,
		},
		{
			name: "coa type code is exist",
			args: args{
				ctx: context.Background(),
				req: models.CreateCategoryIn{},
			},
			mockData: mockData{},
			doMock: func(args args, mockData mockData) {
				testHelper.mockCategoryRepository.EXPECT().GetByCode(args.ctx, args.req.Code).Return(nil, nil)
				testHelper.mockCOATypeRepository.EXPECT().CheckCOATypeByCode(args.ctx, args.req.Code).Return(assert.AnError)
			},
			wantErr: true,
		},
		{
			name: "error GetByCode",
			args: args{
				ctx: context.Background(),
				req: models.CreateCategoryIn{},
			},
			mockData: mockData{},
			doMock: func(args args, mockData mockData) {
				testHelper.mockCategoryRepository.EXPECT().GetByCode(args.ctx, args.req.Code).Return(&models.Category{}, assert.AnError)
			},
			wantErr: true,
		},
		{
			name: "test error",
			args: args{
				ctx: context.Background(),
				req: models.CreateCategoryIn{},
			},
			mockData: mockData{},
			doMock: func(args args, mockData mockData) {
				testHelper.mockCategoryRepository.EXPECT().GetByCode(args.ctx, args.req.Code).Return(nil, nil)
				testHelper.mockCOATypeRepository.EXPECT().CheckCOATypeByCode(args.ctx, args.req.Code).Return(nil)
				testHelper.mockCategoryRepository.EXPECT().Create(args.ctx, &args.req).Return(assert.AnError)
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

			_, err := testHelper.categoryService.Create(tt.args.ctx, tt.args.req)
			assert.Equal(t, tt.wantErr, err != nil)
		})
	}
}

func TestCategoryService_GetAll(t *testing.T) {
	testHelper := serviceTestHelper(t)
	testHelper.mockMySQLRepository.EXPECT().GetCategoryRepository().Return(testHelper.mockCategoryRepository).AnyTimes()

	tests := []struct {
		name    string
		doMock  func()
		wantErr bool
	}{
		{
			name: "happy path",
			doMock: func() {
				testHelper.mockCategoryRepository.EXPECT().List(gomock.AssignableToTypeOf(context.Background())).
					Return([]models.Category{}, nil)
			},
			wantErr: false,
		},
		{
			name: "error repository",
			doMock: func() {
				testHelper.mockCategoryRepository.EXPECT().List(gomock.AssignableToTypeOf(context.Background())).
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

			_, err := testHelper.categoryService.GetAll(context.Background())
			assert.Equal(t, tc.wantErr, err != nil)
		})
	}
}

func TestCategoryService_Update(t *testing.T) {
	testHelper := serviceTestHelper(t)

	req := models.UpdateCategoryIn{
		Code:        "111",
		Name:        "Kas Teller 111",
		Description: "Kas Teller ceritanya 111",
		CoaTypeCode: "LIA",
	}
	type args struct {
		ctx context.Context
		req models.UpdateCategoryIn
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
				req: req,
			},
			doMock: func(args args, mockData mockData) {
				testHelper.mockCategoryRepository.EXPECT().GetByCode(args.ctx, args.req.Code).Return(&models.Category{}, nil)
				testHelper.mockCOATypeRepository.EXPECT().CheckCOATypeByCode(args.ctx, args.req.CoaTypeCode).Return(nil)
				testHelper.mockCategoryRepository.EXPECT().Update(args.ctx, args.req).Return(nil)
			},
			wantErr: false,
		},
		{
			args: args{
				ctx: context.Background(),
				req: models.UpdateCategoryIn{
					Code: req.Code,
				},
			},
			doMock: func(args args, mockData mockData) {
				testHelper.mockCategoryRepository.EXPECT().GetByCode(args.ctx, args.req.Code).Return(&models.Category{}, nil)
				testHelper.mockCOATypeRepository.EXPECT().CheckCOATypeByCode(args.ctx, args.req.CoaTypeCode).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "error case - category code not found",
			args: args{
				ctx: context.Background(),
				req: req,
			},
			mockData: mockData{},
			doMock: func(args args, mockData mockData) {
				testHelper.mockCategoryRepository.EXPECT().GetByCode(args.ctx, args.req.Code).Return(nil, nil)
			},
			wantErr: true,
		},
		{
			name: "error case - database error get category code",
			args: args{
				ctx: context.Background(),
				req: req,
			},
			mockData: mockData{},
			doMock: func(args args, mockData mockData) {
				testHelper.mockCategoryRepository.EXPECT().GetByCode(args.ctx, args.req.Code).Return(nil, assert.AnError)
			},
			wantErr: true,
		},
		{
			name: "error case - coa type code not found",
			args: args{
				ctx: context.Background(),
				req: req,
			},
			mockData: mockData{},
			doMock: func(args args, mockData mockData) {
				testHelper.mockCategoryRepository.EXPECT().GetByCode(args.ctx, args.req.Code).Return(&models.Category{}, nil)
				testHelper.mockCOATypeRepository.EXPECT().CheckCOATypeByCode(args.ctx, args.req.CoaTypeCode).Return(models.GetErrMap(models.ErrKeyCoaTypeNotFound))
			},
			wantErr: true,
		},
		{
			name: "error case - database error update category",
			args: args{
				ctx: context.Background(),
				req: req,
			},
			mockData: mockData{},
			doMock: func(args args, mockData mockData) {
				testHelper.mockCategoryRepository.EXPECT().GetByCode(args.ctx, args.req.Code).Return(&models.Category{}, nil)
				testHelper.mockCOATypeRepository.EXPECT().CheckCOATypeByCode(args.ctx, args.req.CoaTypeCode).Return(nil)
				testHelper.mockCategoryRepository.EXPECT().Update(args.ctx, args.req).Return(assert.AnError)
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
			_, err := testHelper.categoryService.Update(tt.args.ctx, tt.args.req)
			assert.Equal(t, tt.wantErr, err != nil)
		})
	}
}
