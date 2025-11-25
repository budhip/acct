package services_test

import (
	"context"
	"fmt"
	"testing"

	"bitbucket.org/Amartha/go-accounting/internal/models"
	"bitbucket.org/Amartha/go-accounting/internal/pkg/dddnotification"
	"bitbucket.org/Amartha/go-accounting/internal/repositories/mysql"
	mockmysql "bitbucket.org/Amartha/go-accounting/internal/repositories/mysql/mock"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func Test_migrationService_BucketsJournalLoad(t *testing.T) {
	testHelper := serviceTestHelper(t)

	type args struct {
		req models.MigrationBucketsJournalLoadRequest
	}
	tests := []struct {
		name    string
		args    args
		doMock  func(args args)
		wantErr bool
	}{
		{
			name: "failed - error insert journal detail",
			args: args{
				req: models.MigrationBucketsJournalLoadRequest{SubFolder: "part test"},
			},
			doMock: func(args args) {
				testHelper.mockDDDNotification.EXPECT().SendMessageToSlack(
					gomock.AssignableToTypeOf(context.Background()), gomock.AssignableToTypeOf(dddnotification.MessageData{})).
					Return(nil).Times(2)
				testHelper.mockMySQLRepository.EXPECT().
					Atomic(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, f func(ctx context.Context, r mysql.SQLRepository) error) error {
						atomicAccountingRepo := mockmysql.NewMockAccountingRepository(testHelper.mockCtrl)
						atomicRepo := mockmysql.NewMockSQLRepository(testHelper.mockCtrl)
						atomicRepo.EXPECT().GetAccountingRepository().Return(atomicAccountingRepo).AnyTimes()

						atomicAccountingRepo.EXPECT().ToggleForeignKeyChecks(ctx, false).Return(nil).Times(1)
						atomicAccountingRepo.EXPECT().ToggleForeignKeyChecks(ctx, true).Return(nil).Times(1)
						testHelper.mockStorageRepository.EXPECT().Stream(ctx, testHelper.config.Migration.Buckets, fmt.Sprint(args.req.SubFolder, "/5_pas_acct_journal_detail.csv")).
							DoAndReturn(func(ctx context.Context, bucketName, objectName string) (chan models.StreamResult[string], error) {
								streamResult := make(chan models.StreamResult[string])
								go func() {
									defer close(streamResult)
									streamResult <- models.StreamResult[string]{Data: "1,2,3,2024-08-19T13:36:00.619713+07:00,4,true"}
								}()
								return streamResult, nil
							}).Times(1)
						testHelper.mockStorageRepository.EXPECT().Stream(ctx, testHelper.config.Migration.Buckets, gomock.AssignableToTypeOf(string(""))).
							Return(make(chan models.StreamResult[string]), assert.AnError).Times(3)
						atomicAccountingRepo.EXPECT().InsertJournalDetail(ctx, gomock.AssignableToTypeOf([]models.CreateJournalDetail{})).Return(assert.AnError).Times(1)

						return f(ctx, atomicRepo)
					})
			},
			wantErr: true,
		},
		{
			name: "failed - invalid date journal detail",
			args: args{
				req: models.MigrationBucketsJournalLoadRequest{SubFolder: "part test"},
			},
			doMock: func(args args) {
				testHelper.mockDDDNotification.EXPECT().SendMessageToSlack(
					gomock.AssignableToTypeOf(context.Background()), gomock.AssignableToTypeOf(dddnotification.MessageData{})).
					Return(nil).Times(2)
				testHelper.mockMySQLRepository.EXPECT().
					Atomic(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, f func(ctx context.Context, r mysql.SQLRepository) error) error {
						atomicAccountingRepo := mockmysql.NewMockAccountingRepository(testHelper.mockCtrl)
						atomicRepo := mockmysql.NewMockSQLRepository(testHelper.mockCtrl)
						atomicRepo.EXPECT().GetAccountingRepository().Return(atomicAccountingRepo).AnyTimes()

						atomicAccountingRepo.EXPECT().ToggleForeignKeyChecks(ctx, false).Return(nil).Times(1)
						atomicAccountingRepo.EXPECT().ToggleForeignKeyChecks(ctx, true).Return(nil).Times(1)
						testHelper.mockStorageRepository.EXPECT().Stream(ctx, testHelper.config.Migration.Buckets, fmt.Sprint(args.req.SubFolder, "/5_pas_acct_journal_detail.csv")).
							DoAndReturn(func(ctx context.Context, bucketName, objectName string) (chan models.StreamResult[string], error) {
								streamResult := make(chan models.StreamResult[string])
								go func() {
									defer close(streamResult)
									streamResult <- models.StreamResult[string]{Data: "1,2,3,INVALID_DATE,4,true"}
								}()
								return streamResult, nil
							}).Times(1)
						testHelper.mockStorageRepository.EXPECT().Stream(ctx, testHelper.config.Migration.Buckets, gomock.AssignableToTypeOf(string(""))).
							Return(make(chan models.StreamResult[string]), assert.AnError).Times(3)

						return f(ctx, atomicRepo)
					})
			},
			wantErr: true,
		},
		{
			name: "failed - invalid data",
			args: args{
				req: models.MigrationBucketsJournalLoadRequest{SubFolder: "part test"},
			},
			doMock: func(args args) {
				testHelper.mockDDDNotification.EXPECT().SendMessageToSlack(
					gomock.AssignableToTypeOf(context.Background()), gomock.AssignableToTypeOf(dddnotification.MessageData{})).
					Return(nil).Times(2)
				testHelper.mockMySQLRepository.EXPECT().
					Atomic(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, f func(ctx context.Context, r mysql.SQLRepository) error) error {
						atomicAccountingRepo := mockmysql.NewMockAccountingRepository(testHelper.mockCtrl)
						atomicRepo := mockmysql.NewMockSQLRepository(testHelper.mockCtrl)
						atomicRepo.EXPECT().GetAccountingRepository().Return(atomicAccountingRepo).AnyTimes()

						atomicAccountingRepo.EXPECT().ToggleForeignKeyChecks(ctx, false).Return(nil).Times(1)
						atomicAccountingRepo.EXPECT().ToggleForeignKeyChecks(ctx, true).Return(nil).Times(1)
						testHelper.mockStorageRepository.EXPECT().Stream(ctx, testHelper.config.Migration.Buckets, gomock.AssignableToTypeOf(string(""))).
							DoAndReturn(func(ctx context.Context, bucketName, objectName string) (chan models.StreamResult[string], error) {
								streamResult := make(chan models.StreamResult[string])
								go func() {
									defer close(streamResult)
									streamResult <- models.StreamResult[string]{Data: "1,2,3,4,5,6,7,8,9,0"}
								}()
								return streamResult, nil
							}).
							Times(4)

						return f(ctx, atomicRepo)
					})
			},
			wantErr: true,
		},
		{
			name: "failed - read error",
			args: args{
				req: models.MigrationBucketsJournalLoadRequest{SubFolder: "part test"},
			},
			doMock: func(args args) {
				testHelper.mockDDDNotification.EXPECT().SendMessageToSlack(
					gomock.AssignableToTypeOf(context.Background()), gomock.AssignableToTypeOf(dddnotification.MessageData{})).
					Return(nil).Times(2)
				testHelper.mockMySQLRepository.EXPECT().
					Atomic(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, f func(ctx context.Context, r mysql.SQLRepository) error) error {
						atomicAccountingRepo := mockmysql.NewMockAccountingRepository(testHelper.mockCtrl)
						atomicRepo := mockmysql.NewMockSQLRepository(testHelper.mockCtrl)
						atomicRepo.EXPECT().GetAccountingRepository().Return(atomicAccountingRepo).AnyTimes()

						atomicAccountingRepo.EXPECT().ToggleForeignKeyChecks(ctx, false).Return(nil).Times(1)
						atomicAccountingRepo.EXPECT().ToggleForeignKeyChecks(ctx, true).Return(nil).Times(1)
						testHelper.mockStorageRepository.EXPECT().Stream(ctx, testHelper.config.Migration.Buckets, gomock.AssignableToTypeOf(string(""))).
							DoAndReturn(func(ctx context.Context, bucketName, objectName string) (chan models.StreamResult[string], error) {
								streamResult := make(chan models.StreamResult[string])
								go func() {
									defer close(streamResult)
									streamResult <- models.StreamResult[string]{Err: assert.AnError}
								}()
								return streamResult, nil
							}).
							Times(4)

						return f(ctx, atomicRepo)
					})
			},
			wantErr: true,
		},
		{
			name: "failed - stream",
			args: args{
				req: models.MigrationBucketsJournalLoadRequest{SubFolder: "part test"},
			},
			doMock: func(args args) {
				testHelper.mockDDDNotification.EXPECT().SendMessageToSlack(
					gomock.AssignableToTypeOf(context.Background()), gomock.AssignableToTypeOf(dddnotification.MessageData{})).
					Return(nil).Times(2)
				testHelper.mockMySQLRepository.EXPECT().
					Atomic(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, f func(ctx context.Context, r mysql.SQLRepository) error) error {
						atomicAccountingRepo := mockmysql.NewMockAccountingRepository(testHelper.mockCtrl)
						atomicRepo := mockmysql.NewMockSQLRepository(testHelper.mockCtrl)
						atomicRepo.EXPECT().GetAccountingRepository().Return(atomicAccountingRepo).AnyTimes()

						atomicAccountingRepo.EXPECT().ToggleForeignKeyChecks(ctx, false).Return(nil).Times(1)
						atomicAccountingRepo.EXPECT().ToggleForeignKeyChecks(ctx, true).Return(nil).Times(1)
						testHelper.mockStorageRepository.EXPECT().Stream(ctx, testHelper.config.Migration.Buckets, gomock.AssignableToTypeOf(string(""))).
							Return(nil, assert.AnError).Times(4)

						return f(ctx, atomicRepo)
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
			err := testHelper.migrationService.BucketsJournalLoad(context.Background(), tt.args.req)
			assert.Equal(t, tt.wantErr, err != nil)
		})
	}
}
