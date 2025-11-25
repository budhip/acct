package services_test

import (
	"context"
	"mime/multipart"
	"testing"

	"bitbucket.org/Amartha/go-accounting/internal/models"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func Test_publisher_PublishMessage(t *testing.T) {
	testHelper := serviceTestHelper(t)

	f := createMultipartFormData(t)
	data := `[
    {
        "accountNumber": "230007600963",
        "accountType": "CREATE_LOAN_ADVANCE"
    },
    {
        "accountNumber": "230007600941",
        "accountType": "CREATE_LOAN_ADVANCE"
    },
    {
        "accountNumber": "230007600986",
        "accountType": "CREATE_LOAN_ADVANCE"
    }
]`
	o, _ := f.Open()
	type args struct {
		ctx context.Context
		in  models.PublishRequest
	}
	tests := []struct {
		name    string
		args    args
		doMock  func(args args)
		wantErr bool
	}{
		{
			name: "success case",
			args: args{
				ctx: context.Background(),
				in: models.PublishRequest{
					Topic:   "",
					Message: f,
				},
			},
			doMock: func(args args) {
				testHelper.mockFile.EXPECT().ReadAll(o).Return([]byte(data), nil)
				testHelper.mockPublisher.EXPECT().PublishSyncWithKeyAndLog(
					args.ctx,
					gomock.Any(),
					gomock.Any(),
					gomock.Any(),
					gomock.Any(),
				).AnyTimes()
			},
			wantErr: false,
		},
		{
			name: "error case - no such file or directory",
			args: args{
				ctx: context.Background(),
				in: models.PublishRequest{
					Topic: "",
					Message: &multipart.FileHeader{
						Filename: "../../storages/upload_accounts.csv",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "error case - empty data",
			args: args{
				ctx: context.Background(),
				in: models.PublishRequest{
					Topic:   "",
					Message: f,
				},
			},
			doMock: func(args args) {
				testHelper.mockFile.EXPECT().ReadAll(o).Return([]byte(`[]`), nil)
			},
			wantErr: true,
		},
		{
			name: "error case - failed read data",
			args: args{
				ctx: context.Background(),
				in: models.PublishRequest{
					Topic:   "",
					Message: &multipart.FileHeader{},
				},
			},
			doMock: func(args args) {
				testHelper.mockFile.EXPECT().ReadAll(o).Return([]byte(data), assert.AnError)
			},
			wantErr: true,
		},
		{
			name: "error case - failed publish data",
			args: args{
				ctx: context.Background(),
				in: models.PublishRequest{
					Topic:   "",
					Message: f,
				},
			},
			doMock: func(args args) {
				testHelper.mockFile.EXPECT().ReadAll(o).Return([]byte(data), nil)
				testHelper.mockPublisher.EXPECT().PublishSyncWithKeyAndLog(
					gomock.Any(),
					gomock.Any(),
					gomock.Any(),
					gomock.Any(),
					gomock.Any(),
				).Return(models.GetErrMap(models.ErrCodeInternalServerError))
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
			err := testHelper.publisherService.PublishMessage(tt.args.ctx, tt.args.in)
			assert.Equal(t, tt.wantErr, err != nil)
		})
	}
}
