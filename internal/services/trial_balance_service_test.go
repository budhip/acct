package services_test

import (
	"context"
	"testing"
	"time"

	"bitbucket.org/Amartha/go-accounting/internal/models"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestTrialBalanceService_CloseTrialBalance(t *testing.T) {
	testHelper := serviceTestHelper(t)
	databaseError := models.GetErrMap(models.ErrKeyDatabaseError)

	type args struct {
		ctx context.Context
		req models.CloseTrialBalanceRequest
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
			name: "error case - invalid format period",
			args: args{
				ctx: context.Background(),
				req: models.CloseTrialBalanceRequest{
					Period:   "2023-10-01",
					ClosedBy: "dummy@amartha.com",
				},
			},
			wantErr: true,
		},
		{
			name: "error get trial balance by period",
			args: args{
				ctx: context.Background(),
				req: models.CloseTrialBalanceRequest{
					Period:   "2023-10",
					ClosedBy: "dummy@amartha.com",
				},
			},
			doMock: func(args args, mockData mockData) {
				testHelper.mockTrialBalanceRepository.EXPECT().GetByPeriod(args.ctx, args.req.Period, args.req.EntityCode).Return(nil, databaseError)
			},
			wantErr: true,
		},
		{
			name: "error close trial balance",
			args: args{
				ctx: context.Background(),
				req: models.CloseTrialBalanceRequest{
					Period:   "2023-10",
					ClosedBy: "dummy@amartha.com",
				},
			},
			doMock: func(args args, mockData mockData) {
				testHelper.mockTrialBalanceRepository.EXPECT().GetByPeriod(args.ctx, args.req.Period, args.req.EntityCode).Return(&models.TrialBalancePeriod{
					ID:           1,
					Period:       args.req.Period,
					TBFilePath:   "dummy_file_path",
					Status:       models.TrialBalanceStatusOpen,
					ClosedBy:     args.req.ClosedBy,
					IsAdjustment: false,
					CreatedAt:    time.Now(),
					UpdatedAt:    time.Now(),
				}, nil)
				testHelper.mockTrialBalanceRepository.EXPECT().Close(args.ctx, args.req).Return(databaseError)
			},
			wantErr: true,
		},
		{
			name: "success",
			args: args{
				ctx: context.Background(),
				req: models.CloseTrialBalanceRequest{
					Period:   "2023-10",
					ClosedBy: "dummy@amartha.com",
				},
			},
			doMock: func(args args, mockData mockData) {
				testHelper.mockTrialBalanceRepository.EXPECT().GetByPeriod(args.ctx, args.req.Period, args.req.EntityCode).Return(&models.TrialBalancePeriod{
					ID:           1,
					Period:       args.req.Period,
					TBFilePath:   "dummy_file_path",
					Status:       models.TrialBalanceStatusOpen,
					ClosedBy:     args.req.ClosedBy,
					IsAdjustment: false,
					CreatedAt:    time.Now(),
					UpdatedAt:    time.Now(),
				}, nil)
				testHelper.mockTrialBalanceRepository.EXPECT().Close(args.ctx, args.req).Return(nil)
				testHelper.mockBigQuery.EXPECT().
					QueryInsertOpeningBalanceCreated(args.ctx, gomock.Any(), gomock.Any()).
					Return(nil)
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if tt.doMock != nil {
				tt.doMock(tt.args, tt.mockData)
			}
			_, err := testHelper.trialBalanceService.CloseTrialBalance(tt.args.ctx, tt.args.req)
			assert.Equal(t, tt.wantErr, err != nil)
		})
	}
}
