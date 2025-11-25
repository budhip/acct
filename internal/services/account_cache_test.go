package services_test

import (
	"context"
	"testing"

	"bitbucket.org/Amartha/go-accounting/internal/models"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestAccountService_GetAllCategoryCodeSeq(t *testing.T) {
	testHelper := serviceTestHelper(t)
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		args    args
		doMock  func(args args)
		wantRsp []models.DoGetAllCategoryCodeSeqResponse
		wantErr bool
	}{
		{
			name: "success case - get all category seq",
			args: args{
				ctx: context.Background(),
			},
			doMock: func(args args) {
				testHelper.mockCacheRepository.EXPECT().GetAll(args.ctx, gomock.Any()).Return(
					map[string]string{"category_code_212_seq": "100"},
					nil)
			},
			wantRsp: []models.DoGetAllCategoryCodeSeqResponse{
				{
					Kind:  models.KindAccount,
					Key:   "category_code_212_seq",
					Value: "100",
				},
			},
			wantErr: false,
		},
		{
			name: "error case - redis get all",
			args: args{
				ctx: context.Background(),
			},
			doMock: func(args args) {
				testHelper.mockCacheRepository.EXPECT().GetAll(args.ctx, gomock.Any()).Return(nil,
					models.GetErrMap(models.ErrKeyFailedGetFromCache))
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
			got, err := testHelper.accountService.GetAllCategoryCodeSeq(tt.args.ctx)
			assert.Equal(t, tt.wantErr, err != nil)
			assert.Equal(t, got, tt.wantRsp)
		})
	}
}

func TestAccountService_UpdateCategoryCodeSeq(t *testing.T) {
	testHelper := serviceTestHelper(t)
	type args struct {
		ctx context.Context
		in  models.DoUpdateCategoryCodeSeqRequest
	}
	tests := []struct {
		name    string
		args    args
		doMock  func(args args)
		wantErr bool
	}{
		{
			name: "success case - update category seq",
			args: args{
				ctx: context.TODO(),
				in: models.DoUpdateCategoryCodeSeqRequest{
					Key:   "category_code_212_seq",
					Value: 100,
				},
			},
			doMock: func(args args) {
				testHelper.mockCacheRepository.EXPECT().Get(args.ctx, args.in.Key).Return("", nil)
				testHelper.mockCacheRepository.EXPECT().Set(args.ctx, args.in.Key, args.in.Value, gomock.Any()).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "error case - failed get category seq",
			args: args{
				ctx: context.TODO(),
				in: models.DoUpdateCategoryCodeSeqRequest{
					Key:   "category_code_212_seq",
					Value: 100,
				},
			},
			doMock: func(args args) {
				testHelper.mockCacheRepository.EXPECT().Get(args.ctx, args.in.Key).Return("", models.GetErrMap(models.ErrKeyDataNotFound))
			},
			wantErr: true,
		},
		{
			name: "error case - failed set category seq",
			args: args{
				ctx: context.TODO(),
				in: models.DoUpdateCategoryCodeSeqRequest{
					Key:   "category_code_212_seq",
					Value: 100,
				},
			},
			doMock: func(args args) {
				testHelper.mockCacheRepository.EXPECT().Get(args.ctx, args.in.Key).Return("", nil)
				testHelper.mockCacheRepository.EXPECT().Set(args.ctx, args.in.Key, args.in.Value, gomock.Any()).Return(models.GetErrMap(models.ErrKeyFailedSetToCache))
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
			err := testHelper.accountService.UpdateCategoryCodeSeq(tt.args.ctx, tt.args.in)
			assert.Equal(t, tt.wantErr, err != nil)
		})
	}
}

func TestAccountService_CreateCategoryCodeSeq(t *testing.T) {
	testHelper := serviceTestHelper(t)
	type args struct {
		ctx context.Context
		in  models.DoCreateCategoryCodeSeqRequest
	}
	tests := []struct {
		name    string
		args    args
		doMock  func(args args)
		wantErr bool
	}{
		{
			name: "success case - create category seq",
			args: args{
				ctx: context.TODO(),
				in: models.DoCreateCategoryCodeSeqRequest{
					Key:   "category_code_212_seq",
					Value: 100,
				},
			},
			doMock: func(args args) {
				testHelper.mockCacheRepository.EXPECT().SetIfNotExists(args.ctx, args.in.Key, args.in.Value, gomock.Any()).Return(true,nil)
			},
			wantErr: false,
		},
		{
			name: "error case - failed set category seq",
			args: args{
				ctx: context.TODO(),
				in: models.DoCreateCategoryCodeSeqRequest{
					Key:   "category_code_212_seq",
					Value: 100,
				},
			},
			doMock: func(args args) {
				testHelper.mockCacheRepository.EXPECT().SetIfNotExists(args.ctx, args.in.Key, args.in.Value, gomock.Any()).Return(false,models.GetErrMap(models.ErrKeyFailedSetToCache))
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
			err := testHelper.accountService.CreateCategoryCodeSeq(tt.args.ctx, tt.args.in)
			assert.Equal(t, tt.wantErr, err != nil)
		})
	}
}
