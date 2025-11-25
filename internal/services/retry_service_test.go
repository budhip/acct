package services_test

import (
	"context"
	"testing"

	"bitbucket.org/Amartha/go-accounting/internal/models"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestRetryService_RetryInsertJournalTransaction(t *testing.T) {
	testHelper := serviceTestHelper(t)

	type args struct {
		ctx context.Context
		req models.JournalError
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
			name: "success case - retry failed get from cache",
			args: args{
				ctx: context.Background(),
				req: models.JournalError{
					JournalRequest: models.JournalRequest{},
					ErrCauser:      models.GetErrMap(models.ErrKeyFailedGetFromCache),
				},
			},
			doMock: func(args args, mockData mockData) {
				testHelper.mockQueueUnicorn.EXPECT().SendJobHTTP(args.ctx, gomock.Any()).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "success case - retry database error",
			args: args{
				ctx: context.Background(),
				req: models.JournalError{
					JournalRequest: models.JournalRequest{},
					ErrCauser:      models.GetErrMap(models.ErrKeyDatabaseError),
				},
			},
			doMock: func(args args, mockData mockData) {
				testHelper.mockQueueUnicorn.EXPECT().SendJobHTTP(args.ctx, gomock.Any()).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "success case - exclude retry transaction id is exist",
			args: args{
				ctx: context.Background(),
				req: models.JournalError{
					JournalRequest: models.JournalRequest{},
					ErrCauser:      models.GetErrMap(models.ErrKeyTransactionIdIsExist),
				},
			},
			wantErr: false,
		},
		{
			name: "error case - not using error mapping",
			args: args{
				ctx: context.Background(),
				req: models.JournalError{
					JournalRequest: models.JournalRequest{},
					ErrCauser:      "",
				},
			},
			doMock: func(args args, mockData mockData) {
				testHelper.mockDDDNotification.EXPECT().SendMessageToSlack(args.ctx, gomock.Any()).Return(nil)
			},
			wantErr: true,
		},
		{
			name: "error case - failed send job",
			args: args{
				ctx: context.Background(),
				req: models.JournalError{
					JournalRequest: models.JournalRequest{},
					ErrCauser:      models.GetErrMap(models.ErrKeyDatabaseError),
				},
			},
			doMock: func(args args, mockData mockData) {
				testHelper.mockDDDNotification.EXPECT().SendMessageToSlack(args.ctx, gomock.Any()).Return(nil)
				testHelper.mockQueueUnicorn.EXPECT().SendJobHTTP(args.ctx, gomock.Any()).Return(assert.AnError)
			},
			wantErr: true,
		},
		{
			name: "error case - failed send message to slack",
			args: args{
				ctx: context.Background(),
				req: models.JournalError{
					JournalRequest: models.JournalRequest{},
					ErrCauser:      models.GetErrMap(models.ErrKeyDatabaseError),
				},
			},
			doMock: func(args args, mockData mockData) {
				testHelper.mockDDDNotification.EXPECT().SendMessageToSlack(args.ctx, gomock.Any()).Return(assert.AnError)
				testHelper.mockQueueUnicorn.EXPECT().SendJobHTTP(args.ctx, gomock.Any()).Return(assert.AnError)
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
			err := testHelper.retryService.RetryInsertJournalTransaction(tt.args.ctx, tt.args.req)
			assert.Equal(t, tt.wantErr, err != nil)
		})
	}
}

func TestRetryService_RetryUpdateLegacyId(t *testing.T) {
	testHelper := serviceTestHelper(t)

	type args struct {
		ctx context.Context
		req models.AccountError
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
			name: "success case - retry external server error",
			args: args{
				ctx: context.Background(),
				req: models.AccountError{
					CreateAccount: models.CreateAccount{},
					ErrCauser:     models.GetErrMap(models.ErrCodeExternalServerError),
				},
			},
			doMock: func(args args, mockData mockData) {
				testHelper.mockQueueUnicorn.EXPECT().SendJobHTTP(args.ctx, gomock.Any()).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "success case - retry database error",
			args: args{
				ctx: context.Background(),
				req: models.AccountError{
					CreateAccount: models.CreateAccount{},
					ErrCauser:     models.GetErrMap(models.ErrKeyDatabaseError),
				},
			},
			doMock: func(args args, mockData mockData) {
				testHelper.mockQueueUnicorn.EXPECT().SendJobHTTP(args.ctx, gomock.Any()).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "success case - exclude retry invalid values from request",
			args: args{
				ctx: context.Background(),
				req: models.AccountError{
					CreateAccount: models.CreateAccount{},
					ErrCauser:     models.GetErrMap(models.ErrKeyFailedMarshal),
				},
			},
			wantErr: false,
		},
		{
			name: "error case - not using error mapping",
			args: args{
				ctx: context.Background(),
				req: models.AccountError{
					CreateAccount: models.CreateAccount{},
					ErrCauser:     "",
				},
			},
			doMock: func(args args, mockData mockData) {
				testHelper.mockDDDNotification.EXPECT().SendMessageToSlack(args.ctx, gomock.Any()).Return(nil)
			},
			wantErr: true,
		},
		{
			name: "error case - failed send job",
			args: args{
				ctx: context.Background(),
				req: models.AccountError{
					CreateAccount: models.CreateAccount{},
					ErrCauser:     models.GetErrMap(models.ErrKeyDatabaseError),
				},
			},
			doMock: func(args args, mockData mockData) {
				testHelper.mockDDDNotification.EXPECT().SendMessageToSlack(args.ctx, gomock.Any()).Return(nil)
				testHelper.mockQueueUnicorn.EXPECT().SendJobHTTP(args.ctx, gomock.Any()).Return(assert.AnError)
			},
			wantErr: true,
		},
		{
			name: "error case - failed send message to slack",
			args: args{
				ctx: context.Background(),
				req: models.AccountError{
					CreateAccount: models.CreateAccount{},
					ErrCauser:     models.GetErrMap(models.ErrKeyDatabaseError),
				},
			},
			doMock: func(args args, mockData mockData) {
				testHelper.mockDDDNotification.EXPECT().SendMessageToSlack(args.ctx, gomock.Any()).Return(assert.AnError)
				testHelper.mockQueueUnicorn.EXPECT().SendJobHTTP(args.ctx, gomock.Any()).Return(assert.AnError)
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
			err := testHelper.retryService.RetryUpdateLegacyId(tt.args.ctx, tt.args.req)
			assert.Equal(t, tt.wantErr, err != nil)
		})
	}
}
