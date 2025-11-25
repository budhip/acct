package services_test

import (
	"context"
	"fmt"
	"testing"

	"bitbucket.org/Amartha/go-accounting/internal/models"
	"bitbucket.org/Amartha/go-accounting/internal/pkg/dddnotification"
	"github.com/stretchr/testify/assert"
)

func TestDLQProcessorService_SendNotificationJournalFailure(t *testing.T) {
	testHelper := serviceTestHelper(t)
	type args struct {
		ctx context.Context
		in  models.JournalError
	}

	var (
		journalErrData = models.JournalError{
			JournalRequest: models.JournalRequest{TransactionId: "12345", Transactions: []models.Transaction{
				{
					TransactionType: "TUPVN",
				},
			}},
			ErrCauser: assert.AnError,
		}
	)

	tests := []struct {
		name    string
		args    args
		doMock  func(args args)
		wantErr bool
	}{
		{
			name: "error send notification",
			args: args{
				ctx: context.Background(),
				in:  journalErrData,
			},
			doMock: func(args args) {
				testHelper.mockDDDNotification.EXPECT().SendMessageToSlack(args.ctx, dddnotification.MessageData{
					Operation: "Process Journal",
					Message:   fmt.Sprintf("<!channel> Failed with trx id: %s, trx type: %s, err: %v", args.in.TransactionId, args.in.Transactions[0].TransactionType, args.in.ErrCauser),
				}).Return(assert.AnError)
			},
			wantErr: true,
		},
		{
			name: "success send notification",
			args: args{
				ctx: context.Background(),
				in:  journalErrData,
			},
			doMock: func(args args) {
				testHelper.mockDDDNotification.EXPECT().SendMessageToSlack(args.ctx, dddnotification.MessageData{
					Operation: "Process Journal",
					Message:   fmt.Sprintf("<!channel> Failed with trx id: %s, trx type: %s, err: %v", args.in.TransactionId, args.in.Transactions[0].TransactionType, args.in.ErrCauser),
				}).Return(nil)
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.doMock(tt.args)

			err := testHelper.dlqProcessorService.SendNotificationJournalFailure(tt.args.ctx, tt.args.in)
			assert.Equal(t, tt.wantErr, err != nil)
		})
	}
}

func TestDLQProcessorService_SendNotificationAccountT24Failure(t *testing.T) {
	testHelper := serviceTestHelper(t)
	type args struct {
		ctx context.Context
		in  models.AccountError
	}

	var (
		accountErrData = models.AccountError{
			CreateAccount: models.CreateAccount{AccountNumber: "12345"},
			ErrCauser:     assert.AnError,
		}
	)

	tests := []struct {
		name    string
		args    args
		doMock  func(args args)
		wantErr bool
	}{
		{
			name: "error send notification",
			args: args{
				ctx: context.Background(),
				in:  accountErrData,
			},
			doMock: func(args args) {
				testHelper.mockDDDNotification.EXPECT().SendMessageToSlack(args.ctx, dddnotification.MessageData{
					Operation: "Process Account to T24",
					Message:   fmt.Sprintf("<!channel> Failed with acc number: %s. err: %v", args.in.AccountNumber, args.in.ErrCauser),
				}).Return(assert.AnError)
			},
			wantErr: true,
		},
		{
			name: "success send notification",
			args: args{
				ctx: context.Background(),
				in:  accountErrData,
			},
			doMock: func(args args) {
				testHelper.mockDDDNotification.EXPECT().SendMessageToSlack(args.ctx, dddnotification.MessageData{
					Operation: "Process Account to T24",
					Message:   fmt.Sprintf("<!channel> Failed with acc number: %s. err: %v", args.in.AccountNumber, args.in.ErrCauser),
				}).Return(nil)
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.doMock(tt.args)
			err := testHelper.dlqProcessorService.SendNotificationAccountT24Failure(tt.args.ctx, tt.args.in)
			assert.Equal(t, tt.wantErr, err != nil)
		})
	}
}
