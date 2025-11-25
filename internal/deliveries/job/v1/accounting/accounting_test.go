package accounting

import (
	"context"
	"testing"
	"time"

	"bitbucket.org/Amartha/go-accounting/internal/models"
	"bitbucket.org/Amartha/go-accounting/internal/pkg/atime"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func Test_accountingHandler_GenerateTrialBalanceBigQuery(t *testing.T) {
	testHelper := accountingTestHelper(t)
	type args struct {
		ctx  context.Context
		date time.Time
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
			name: "success case - GenerateTrialBalanceBigQuery",
			args: args{
				ctx:  context.TODO(),
				date: atime.Now(),
			},
			doMock: func(args args, mockData mockData) {
				testHelper.mockAccountingService.EXPECT().GenerateTrialBalanceBigQuery(gomock.AssignableToTypeOf(args.ctx), gomock.Any(), gomock.Any()).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "error case - GenerateTrialBalanceBigQuery",
			args: args{
				ctx:  context.TODO(),
				date: atime.Now(),
			},
			doMock: func(args args, mockData mockData) {
				testHelper.mockAccountingService.EXPECT().GenerateTrialBalanceBigQuery(gomock.AssignableToTypeOf(args.ctx), gomock.Any(), gomock.Any()).Return(models.GetErrMap(models.ErrKeyDatabaseError))
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
			rh := &accountingHandler{
				accountingService: testHelper.mockAccountingService,
			}
			err := rh.GenerateTrialBalanceBigQuery(tt.args.ctx, tt.args.date)
			assert.Equal(t, tt.wantErr, err != nil)
		})
	}
}

func Test_accountingHandler_GenerateAccountDailyBalanceAndTrialBalance(t *testing.T) {
	testHelper := accountingTestHelper(t)
	type args struct {
		ctx  context.Context
		date time.Time
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
			name: "success case - GenerateAccountBalanceDailyTransaction",
			args: args{
				ctx:  context.TODO(),
				date: atime.Now(),
			},
			doMock: func(args args, mockData mockData) {
				testHelper.mockAccountingService.EXPECT().GenerateAccountDailyBalance(gomock.AssignableToTypeOf(args.ctx), gomock.Any()).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "error case - GenerateAccountBalanceDailyTransaction",
			args: args{
				ctx:  context.TODO(),
				date: atime.Now(),
			},
			doMock: func(args args, mockData mockData) {
				testHelper.mockAccountingService.EXPECT().GenerateAccountDailyBalance(gomock.AssignableToTypeOf(args.ctx), gomock.Any()).Return(models.GetErrMap(models.ErrKeyDatabaseError))
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
			rh := &accountingHandler{
				accountingService: testHelper.mockAccountingService,
			}
			err := rh.GenerateAccountDailyBalanceAndTrialBalance(tt.args.ctx, tt.args.date)
			assert.Equal(t, tt.wantErr, err != nil)
		})
	}
}

func Test_accountingHandler_GenerateRangeAccountDailyBalanceAndTrialBalance(t *testing.T) {
	testHelper := accountingTestHelper(t)
	date, _ := atime.ParseStringToDatetime(atime.DateFormatYYYYMMDD, "2024-01-31")
	type args struct {
		ctx  context.Context
		date time.Time
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
			name: "success case - GenerateRangeAccountDailyBalanceAndTrialBalance",
			args: args{
				ctx:  context.TODO(),
				date: date,
			},
			doMock: func(args args, mockData mockData) {
				testHelper.mockAccountingService.EXPECT().GenerateAccountDailyBalance(gomock.AssignableToTypeOf(args.ctx), gomock.Any()).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "error case - GenerateRangeAccountDailyBalanceAndTrialBalance",
			args: args{
				ctx:  context.TODO(),
				date: date,
			},
			doMock: func(args args, mockData mockData) {
				testHelper.mockAccountingService.EXPECT().GenerateAccountDailyBalance(gomock.AssignableToTypeOf(args.ctx), gomock.Any()).Return(models.GetErrMap(models.ErrKeyDatabaseError))
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
			rh := &accountingHandler{
				accountingService: testHelper.mockAccountingService,
			}
			err := rh.GenerateRangeAccountDailyBalanceAndTrialBalance(tt.args.ctx, tt.args.date)
			assert.Equal(t, tt.wantErr, err != nil)
		})
	}
}
