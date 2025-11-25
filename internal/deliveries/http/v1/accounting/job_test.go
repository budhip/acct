package accounting

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"bitbucket.org/Amartha/go-accounting/internal/models"

	"github.com/golang/mock/gomock"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
)

func Test_Handler_runJob(t *testing.T) {
	testHelper := accountingTestHelper(t)

	type args struct {
		ctx         context.Context
		contentType string
		req         models.RunTrialBalanceRequest
	}
	type expectation struct {
		wantRes  string
		wantCode int
	}
	tests := []struct {
		name        string
		args        args
		expectation expectation
		doMock      func(args args, expectation expectation)
	}{
		{
			name: "success case",
			args: args{
				ctx:         context.Background(),
				contentType: echo.MIMEApplicationJSON,
				req: models.RunTrialBalanceRequest{
					Date:         "2025-01-01",
					IsAdjustment: false,
				},
			},
			expectation: expectation{
				wantRes:  `"processing"`,
				wantCode: 200,
			},
			doMock: func(args args, expectation expectation) {
				testHelper.mockAccountingService.EXPECT().GenerateTrialBalanceBigQuery(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			},
		},
		{
			name: "error case - error convert to filter options",
			args: args{
				ctx:         context.Background(),
				contentType: echo.MIMEApplicationJSON,
				req: models.RunTrialBalanceRequest{
					Date:         "20250101",
					IsAdjustment: false,
				},
			},
			expectation: expectation{
				wantRes:  `{"status":"error","code":"INVALID_VALUES","message":"invalid format date caused by date 20250101 format must be YYYY-MM-DD"}`,
				wantCode: 400,
			},
			doMock: func(args args, expectation expectation) {},
		},
		{
			name: "error case - validation",
			args: args{
				ctx:         context.Background(),
				contentType: echo.MIMEApplicationJSON,
				req:         models.RunTrialBalanceRequest{},
			},
			expectation: expectation{
				wantRes:  `{"status":"error","message":"validation failed","errors":[{"code":"UNKNOW","field":"date","message":"required"}]}`,
				wantCode: 422,
			},
			doMock: func(args args, expectation expectation) {},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if tt.doMock != nil {
				tt.doMock(tt.args, tt.expectation)
			}

			var b bytes.Buffer
			err := json.NewEncoder(&b).Encode(tt.args.req)
			require.NoError(t, err)

			r := httptest.NewRequest(http.MethodPost, "/api/v1/jobs/trial-balance", &b)
			r.Header.Set(echo.HeaderContentType, tt.args.contentType)
			w := httptest.NewRecorder()

			testHelper.router.NewContext(r, w)
			testHelper.router.ServeHTTP(w, r)

			require.Equal(t, tt.expectation.wantCode, w.Code)
			require.Equal(t, tt.expectation.wantRes, strings.Trim(w.Body.String(), "\n"))
		})
	}
}
