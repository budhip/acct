package services_test

import (
	"context"
	"encoding/json"
	"testing"

	"bitbucket.org/Amartha/go-accounting/internal/models"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestEntityService_Create(t *testing.T) {
	testHelper := serviceTestHelper(t)

	type args struct {
		ctx context.Context
		req models.CreateEntityIn
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
				req: models.CreateEntityIn{
					Code:        "",
					Name:        "",
					Description: "",
					Status:      "",
				},
			},
			doMock: func(args args, mockData mockData) {
				testHelper.mockEntityRepository.EXPECT().GetByCode(args.ctx, args.req.Code).Return(nil, nil)
				testHelper.mockEntityRepository.EXPECT().Create(args.ctx, &args.req).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "code is exist",
			args: args{
				ctx: context.Background(),
				req: models.CreateEntityIn{},
			},
			mockData: mockData{},
			doMock: func(args args, mockData mockData) {
				testHelper.mockEntityRepository.EXPECT().GetByCode(args.ctx, args.req.Code).Return(&models.Entity{}, nil)
			},
			wantErr: true,
		},
		{
			name: "error GetByCode",
			args: args{
				ctx: context.Background(),
				req: models.CreateEntityIn{},
			},
			mockData: mockData{},
			doMock: func(args args, mockData mockData) {
				testHelper.mockEntityRepository.EXPECT().GetByCode(args.ctx, args.req.Code).Return(&models.Entity{}, assert.AnError)
			},
			wantErr: true,
		},
		{
			name: "test error",
			args: args{
				ctx: context.Background(),
				req: models.CreateEntityIn{},
			},
			mockData: mockData{},
			doMock: func(args args, mockData mockData) {
				testHelper.mockEntityRepository.EXPECT().GetByCode(args.ctx, args.req.Code).Return(nil, nil)
				testHelper.mockEntityRepository.EXPECT().Create(args.ctx, &args.req).Return(assert.AnError)
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

			_, err := testHelper.entityService.Create(tt.args.ctx, tt.args.req)
			assert.Equal(t, tt.wantErr, err != nil)
		})
	}
}

func TestEntityService_GetAll(t *testing.T) {
	testHelper := serviceTestHelper(t)

	tests := []struct {
		name    string
		doMock  func()
		wantErr bool
	}{
		{
			name: "success get all entities",
			doMock: func() {
				testHelper.mockEntityRepository.EXPECT().
					List(gomock.AssignableToTypeOf(context.Background())).
					Return(&[]models.Entity{}, nil)
			},
			wantErr: false,
		},
		{
			name: "error get data from repository",
			doMock: func() {
				testHelper.mockEntityRepository.EXPECT().
					List(gomock.AssignableToTypeOf(context.Background())).
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

			_, err := testHelper.entityService.GetAll(context.Background())
			assert.Equal(t, tc.wantErr, err != nil)
		})
	}
}

func TestEntityService_Update(t *testing.T) {
	testHelper := serviceTestHelper(t)
	type (
		args struct {
			ctx context.Context
			req models.UpdateEntity
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
			name: "success case - update entity",
			args: args{
				ctx: context.Background(),
				req: models.UpdateEntity{
					Name:        "Lender Yang Baik",
					Description: "Testing",
					Status:      "active",
					Code:        "001",
				},
			},
			doMock: func(args args, mockData mockData) {
				testHelper.mockEntityRepository.EXPECT().CheckEntityByCode(args.ctx, args.req.Code, "").Return(nil)
				testHelper.mockEntityRepository.EXPECT().Update(args.ctx, args.req).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "success case - empty request",
			args: args{
				ctx: context.Background(),
				req: models.UpdateEntity{
					Code: "001",
				},
			},
			doMock: func(args args, mockData mockData) {
				testHelper.mockEntityRepository.EXPECT().CheckEntityByCode(args.ctx, args.req.Code, "").Return(nil)
			},
			wantErr: false,
		},
		{
			name: "error case - database error get entity",
			args: args{
				ctx: context.Background(),
				req: models.UpdateEntity{
					Name:        "Lender Yang Baik",
					Description: "Testing",
					Status:      "active",
					Code:        "001",
				},
			},
			doMock: func(args args, mockData mockData) {
				testHelper.mockEntityRepository.EXPECT().CheckEntityByCode(args.ctx, args.req.Code, "").Return(assert.AnError)
			},
			wantErr: true,
		},
		{
			name: "error case - database error update entity",
			args: args{
				ctx: context.Background(),
				req: models.UpdateEntity{
					Name:        "Lender Yang Baik",
					Description: "Testing",
					Code:        "001",
					Status:      "active",
				},
			},
			doMock: func(args args, mockData mockData) {
				testHelper.mockEntityRepository.EXPECT().CheckEntityByCode(args.ctx, args.req.Code, "").Return(nil)
				testHelper.mockEntityRepository.EXPECT().Update(args.ctx, args.req).Return(models.GetErrMap(models.ErrKeyDatabaseError))
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
			_, err := testHelper.entityService.Update(tt.args.ctx, tt.args.req)
			assert.Equal(t, tt.wantErr, err != nil)
		})
	}
}

func TestEntityService_GetEntityByParams(t *testing.T) {
	testHelper := serviceTestHelper(t)
	type args struct {
		ctx context.Context
		req models.GetEntity
	}

	validEntity := &models.Entity{Code: "001", Name: "AMF"}
	b, _ := json.Marshal(validEntity)

	tests := []struct {
		name    string
		doMock  func()
		wantErr bool
		args    args
	}{
		{
			name: "success from cache",
			doMock: func() {
				testHelper.mockCacheRepository.EXPECT().
					Get(gomock.Any(), gomock.Any()).
					Return(string(b), nil)
			},
			wantErr: false,
		},
		{
			name: "cache miss, success from DB and set cache",
			doMock: func() {
				testHelper.mockCacheRepository.EXPECT().
					Get(gomock.Any(), gomock.Any()).
					Return("", models.GetErrMap(models.ErrKeyDataNotFound))

				testHelper.mockEntityRepository.EXPECT().
					GetByParams(gomock.Any(), gomock.Any()).
					Return(validEntity, nil)

				testHelper.mockCacheRepository.EXPECT().
					Set(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil)
			},
			wantErr: false,
		},
		{
			name: "db error after cache miss",
			doMock: func() {
				testHelper.mockCacheRepository.EXPECT().
					Get(gomock.Any(), gomock.Any()).
					Return("", models.GetErrMap(models.ErrKeyDataNotFound))

				testHelper.mockEntityRepository.EXPECT().
					GetByParams(gomock.Any(), gomock.Any()).
					Return(nil, models.GetErrMap(models.ErrKeyDatabaseError))
			},
			wantErr: true,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.doMock != nil {
				tc.doMock()
			}

			_, err := testHelper.entityService.GetByParam(tc.args.ctx, tc.args.req)
			assert.Equal(t, tc.wantErr, err != nil)
		})
	}
}
