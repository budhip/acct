package accounting

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"bitbucket.org/Amartha/go-accounting/internal/models"

	"github.com/labstack/echo/v4"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func Test_Handler_getTrialBalance(t *testing.T) {
	testHelper := accountingTestHelper(t)

	type args struct {
		ctx         context.Context
		contentType string
		req         url.Values
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
			name: "success case - empty data",
			args: args{
				ctx:         context.TODO(),
				contentType: echo.MIMEApplicationJSON,
				req: url.Values{
					"entityCode": []string{"001"},
				},
			},
			doMock: func(args args, expectation expectation) {
				testHelper.mockAccountingService.EXPECT().GetTrialBalance(gomock.Any(), gomock.Any()).Return(models.GetTrialBalanceResponses{
					Kind:        models.KindTrialBalance,
					EntityCode:  "001",
					EntityName:  "AMF",
					ClosingDate: "2023-01-31",
				}, nil)
			},
			expectation: expectation{
				wantRes:  `{"kind":"collection","contents":{"kind":"trialBalance","entityCode":"001","entityName":"AMF","closingDate":"2023-01-31","catchAll":{"catchAllOpeningBalance":"","catchAllDebitMovement":"","catchAllCreditMovement":"","catchAllClosingBalance":""},"coaTypes":null},"total_rows":0}`,
				wantCode: 200,
			},
		},
		{
			name: "error case - validation filter failed",
			args: args{
				ctx:         context.TODO(),
				contentType: echo.MIMEApplicationJSON,
				req: url.Values{
					"entityCode": []string{"001"},
					"startDate":  []string{"2023-01-08"},
					"endDate":    []string{"2023-01-07"},
				},
			},
			expectation: expectation{
				wantRes:  `{"status":"error","code":"INVALID_VALUES","message":"end date must be greater than start date"}`,
				wantCode: 400,
			},
		},
		{
			name: "error case - entity code not found",
			args: args{
				ctx:         context.TODO(),
				contentType: echo.MIMEApplicationJSON,
				req: url.Values{
					"entityCode": []string{"000"},
				},
			},
			doMock: func(args args, expectation expectation) {
				testHelper.mockAccountingService.EXPECT().GetTrialBalance(gomock.Any(), gomock.Any()).Return(models.GetTrialBalanceResponses{}, models.GetErrMap(models.ErrKeyEntityCodeNotFound))
			},
			expectation: expectation{
				wantRes:  `{"status":"error","code":"DATA_NOT_FOUND","message":"entity code not found"}`,
				wantCode: 404,
			},
		},
		{
			name: "error case - database error",
			args: args{
				ctx:         context.TODO(),
				contentType: echo.MIMEApplicationJSON,
				req: url.Values{
					"entityCode": []string{"001"},
				},
			},
			expectation: expectation{
				wantRes:  `{"status":"error","code":"DATABASE_ERROR","message":"database error"}`,
				wantCode: 500,
			},
			doMock: func(args args, expectation expectation) {
				testHelper.mockAccountingService.EXPECT().GetTrialBalance(gomock.Any(), gomock.Any()).Return(models.GetTrialBalanceResponses{}, models.GetErrMap(models.ErrKeyDatabaseError))
			},
		},
		{
			name: "error case - request validation failed",
			args: args{
				ctx:         context.TODO(),
				contentType: echo.MIMEApplicationJSON,
				req: url.Values{
					"entityCode": []string{},
				},
			},
			expectation: expectation{
				wantRes:  `{"status":"error","message":"validation failed","errors":[{"code":"MISSING_FIELD","field":"entityCode","message":"field is missing"}]}`,
				wantCode: 422,
			},
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

			r := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/trial-balances?%s", tt.args.req.Encode()), nil)
			r.Header.Set(echo.HeaderContentType, tt.args.contentType)
			w := httptest.NewRecorder()

			testHelper.router.NewContext(r, w)
			testHelper.router.ServeHTTP(w, r)

			require.Equal(t, tt.expectation.wantCode, w.Code)
			require.Equal(t, tt.expectation.wantRes, strings.Trim(w.Body.String(), "\n"))
		})
	}
}

func Test_Handler_downloadCSVgetTrialBalance(t *testing.T) {
	testHelper := accountingTestHelper(t)

	type args struct {
		ctx         context.Context
		contentType string
		req         url.Values
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
				ctx:         context.TODO(),
				contentType: echo.MIMEApplicationJSON,
				req: url.Values{
					"entityCode": []string{"001"},
				},
			},
			expectation: expectation{
				wantRes:  `"processing"`,
				wantCode: 200,
			},
			doMock: func(args args, expectation expectation) {
				testHelper.mockAccountingService.EXPECT().DownloadTrialBalance(gomock.Any(), gomock.Any()).Return(nil)
			},
		},
		{
			name: "error case - validation filter failed",
			args: args{
				ctx:         context.TODO(),
				contentType: echo.MIMEApplicationJSON,
				req: url.Values{
					"entityCode": []string{"001"},
					"startDate":  []string{"2023-01-08"},
					"endDate":    []string{"2023-01-07"},
				},
			},
			expectation: expectation{
				wantRes:  `{"status":"error","code":"INVALID_VALUES","message":"end date must be greater than start date"}`,
				wantCode: 400,
			},
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

			t.Logf(fmt.Sprintf("/api/v1/trial-balances/download?%s", tt.args.req.Encode()))

			r := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/trial-balances/download?%s", tt.args.req.Encode()), nil)
			r.Header.Set(echo.HeaderContentType, tt.args.contentType)
			w := httptest.NewRecorder()

			testHelper.router.NewContext(r, w)
			testHelper.router.ServeHTTP(w, r)

			require.Equal(t, tt.expectation.wantCode, w.Code)
			require.Equal(t, tt.expectation.wantRes, strings.Trim(w.Body.String(), "\n"))
		})
	}
}

func Test_Handler_getTrialBalanceDetailList(t *testing.T) {
	testHelper := accountingTestHelper(t)

	type args struct {
		ctx         context.Context
		contentType string
		req         url.Values
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
			name: "success case - get trial balance details",
			args: args{
				ctx:         context.Background(),
				contentType: echo.MIMEApplicationJSON,
				req: url.Values{
					"subCategoryCode": []string{"21108"},
					"entityCode":      []string{"001"},
				},
			},
			expectation: expectation{
				wantRes:  `{"kind":"collection","contents":[{"kind":"trialBalanceDetail","accountNumber":"123456","accountName":"Shinji Takeru","openingBalance":"100.000,00","debitMovement":"50.000,00","creditMovement":"150.000,00","closingBalance":"0,00"}],"pagination":{"prev":"","next":"","totalEntries":1}}`,
				wantCode: 200,
			},
			doMock: func(args args, expectation expectation) {
				testHelper.mockAccountingService.EXPECT().
					GetTrialBalanceDetails(gomock.Any(), gomock.Any()).
					Return([]models.TrialBalanceDetailOut{
						{
							AccountNumber:  "123456",
							AccountName:    "Shinji Takeru",
							OpeningBalance: decimal.NewFromInt(100_000),
							DebitMovement:  decimal.NewFromInt(50_000),
							CreditMovement: decimal.NewFromInt(150_000),
							ClosingBalance: decimal.Zero,
						},
					}, 1, nil)
			},
		},
		{
			name: "error case - get trial balance details",
			args: args{
				ctx:         context.Background(),
				contentType: echo.MIMEApplicationJSON,
				req: url.Values{
					"subCategoryCode": []string{"21108"},
					"entityCode":      []string{"001"},
				},
			},
			expectation: expectation{
				wantRes:  `{"status":"error","code":500,"message":"assert.AnError general error for testing"}`,
				wantCode: 500,
			},
			doMock: func(args args, expectation expectation) {
				testHelper.mockAccountingService.EXPECT().
					GetTrialBalanceDetails(gomock.Any(), gomock.Any()).
					Return(nil, 1, assert.AnError)
			},
		},
		{
			name: "error case - invalid limit get trial balance details",
			args: args{
				ctx:         context.Background(),
				contentType: echo.MIMEApplicationJSON,
				req: url.Values{
					"subCategoryCode": []string{"21108"},
					"entityCode":      []string{"001"},
					"limit":           []string{"-1"},
				},
			},
			expectation: expectation{
				wantRes:  `{"status":"error","code":"INVALID_VALUES","message":"the limit must be greater than zero"}`,
				wantCode: 400,
			},
			doMock: func(args args, expectation expectation) {},
		},
		{
			name: "error case - validation error get trial balance details",
			args: args{
				ctx:         context.Background(),
				contentType: echo.MIMEApplicationJSON,
				req:         url.Values{},
			},
			expectation: expectation{
				wantCode: 422,
				wantRes:  `{"status":"error","message":"validation failed","errors":[{"code":"MISSING_FIELD","field":"entityCode","message":"field is missing"},{"code":"MISSING_FIELD","field":"subCategoryCode","message":"field is missing"}]}`,
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

			r := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/trial-balances/details?%s", tt.args.req.Encode()), nil)
			r.Header.Set(echo.HeaderContentType, tt.args.contentType)
			w := httptest.NewRecorder()

			testHelper.router.NewContext(r, w)
			testHelper.router.ServeHTTP(w, r)

			require.Equal(t, tt.expectation.wantCode, w.Code)
			require.Equal(t, tt.expectation.wantRes, strings.Trim(w.Body.String(), "\n"))
		})
	}
}

func Test_Handler_getTrialBalanceBySubcategory(t *testing.T) {
	testHelper := accountingTestHelper(t)

	type args struct {
		ctx         context.Context
		contentType string
		req         url.Values
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
			name: "success case - get trial balance by sub category",
			args: args{
				ctx:         context.Background(),
				contentType: echo.MIMEApplicationJSON,
				req:         url.Values{},
			},
			expectation: expectation{
				wantRes:  `{"kind":"trialBalanceSubCategory","subCategoryCode":"21108","subCategoryName":"Lender's Cash - Earn (Fixed Income)","openingBalance":"100.000,00","debitMovement":"50.000,00","creditMovement":"150.000,00","closingBalance":"0,00"}`,
				wantCode: 200,
			},
			doMock: func(args args, expectation expectation) {
				testHelper.mockAccountingService.EXPECT().
					GetTrialBalanceBySubCategoryCode(gomock.Any(), gomock.Any()).
					Return(models.TrialBalanceBySubCategoryOut{
						SubCategoryCode: "21108",
						SubCategoryName: "Lender's Cash - Earn (Fixed Income)",
						OpeningBalance:  decimal.NewFromInt(100_000),
						DebitMovement:   decimal.NewFromInt(50_000),
						CreditMovement:  decimal.NewFromInt(150_000),
						ClosingBalance:  decimal.Zero,
					}, nil)
			},
		},
		{
			name: "error case - get trial balance sub category",
			args: args{
				ctx:         context.Background(),
				contentType: echo.MIMEApplicationJSON,
				req:         url.Values{},
			},
			expectation: expectation{
				wantRes:  `{"status":"error","code":500,"message":"assert.AnError general error for testing"}`,
				wantCode: 500,
			},
			doMock: func(args args, expectation expectation) {
				testHelper.mockAccountingService.EXPECT().
					GetTrialBalanceBySubCategoryCode(gomock.Any(), gomock.Any()).
					Return(models.TrialBalanceBySubCategoryOut{}, assert.AnError)
			},
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

			r := httptest.NewRequest(
				http.MethodGet,
				fmt.Sprintf("/api/v1/trial-balances/sub-categories/%s?%s", "21108", tt.args.req.Encode()),
				nil)
			r.Header.Set(echo.HeaderContentType, tt.args.contentType)
			w := httptest.NewRecorder()

			testHelper.router.NewContext(r, w)
			testHelper.router.ServeHTTP(w, r)

			require.Equal(t, tt.expectation.wantCode, w.Code)
			require.Equal(t, tt.expectation.wantRes, strings.Trim(w.Body.String(), "\n"))
		})
	}
}

func Test_Handler_sendToEmailCSVgetTrialBalanceDetails(t *testing.T) {
	testHelper := accountingTestHelper(t)

	type args struct {
		ctx         context.Context
		contentType string
		req         url.Values
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
				req: url.Values{
					"subCategoryCode": []string{"21108"},
					"entityCode":      []string{"001"},
					"startDate":       []string{"2023-01-01"},
					"endDate":         []string{"2023-01-30"},
					"email":           []string{"tono@amartha.com"},
				},
			},
			expectation: expectation{
				wantRes:  `"processing"`,
				wantCode: 200,
			},
			doMock: func(args args, expectation expectation) {
				testHelper.mockAccountingService.EXPECT().SendToEmailGetTrialBalanceDetails(gomock.Any(), gomock.Any()).Return(nil)
			},
		},
		{
			name: "error case - error convert to filter options",
			args: args{
				ctx:         context.Background(),
				contentType: echo.MIMEApplicationJSON,
				req: url.Values{
					"subCategoryCode": []string{"21108"},
					"entityCode":      []string{"001"},
					"startDate":       []string{"20230101"},
					"endDate":         []string{"20230130"},
					"email":           []string{"tono@amartha.com"},
				},
			},
			expectation: expectation{
				wantRes:  `{"status":"error","code":"INVALID_VALUES","message":"invalid format date caused by date 20230101 format must be YYYY-MM-DD"}`,
				wantCode: 400,
			},
			doMock: func(args args, expectation expectation) {},
		},
		{
			name: "error case - validation",
			args: args{
				ctx:         context.Background(),
				contentType: echo.MIMEApplicationJSON,
				req:         url.Values{},
			},
			expectation: expectation{
				wantRes:  `{"status":"error","message":"validation failed","errors":[{"code":"MISSING_FIELD","field":"entityCode","message":"field is missing"},{"code":"MISSING_FIELD","field":"subCategoryCode","message":"field is missing"},{"code":"UNKNOW","field":"email","message":"required"}]}`,
				wantCode: 422,
			},
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

			r := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/trial-balances/details/download?%s", tt.args.req.Encode()), nil)
			r.Header.Set(echo.HeaderContentType, tt.args.contentType)
			w := httptest.NewRecorder()

			testHelper.router.NewContext(r, w)
			testHelper.router.ServeHTTP(w, r)

			require.Equal(t, tt.expectation.wantCode, w.Code)
			require.Equal(t, tt.expectation.wantRes, strings.Trim(w.Body.String(), "\n"))
		})
	}
}

func Test_Handler_adjustmentTrialBalance(t *testing.T) {
	testHelper := accountingTestHelper(t)
	ctx := context.Background()

	type args struct {
		contentType string
		req         *models.AdjustmentTrialBalanceRequest
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
			name: "success",
			args: args{
				contentType: echo.MIMEApplicationJSON,
				req: &models.AdjustmentTrialBalanceRequest{
					AdjustmentDate: "2025-01-01",
				},
			},
			expectation: expectation{
				wantRes:  `"processing"`,
				wantCode: 200,
			},
			doMock: func(args args, expectation expectation) {
				testHelper.mockAccountingService.EXPECT().GenerateAdjustmentTrialBalanceBigQuery(ctx, gomock.Any()).Return(nil)
			},
		},
		{
			name: "error contentType",
			args: args{
				contentType: "",
			},
			expectation: expectation{
				wantRes:  `{"status":"error","code":400,"message":"code=415, message=Unsupported Media Type"}`,
				wantCode: 400,
			},
			doMock: func(args args, expectation expectation) {},
		},
		{
			name: "error - validation",
			args: args{
				contentType: echo.MIMEApplicationJSON,
			},
			expectation: expectation{
				wantRes:  `{"status":"error","message":"validation failed","errors":[{"code":"MISSING_FIELD","field":"adjustmentDate","message":"field is missing"}]}`,
				wantCode: 422,
			},
		},
		{
			name: "error - invalid format date",
			args: args{
				contentType: echo.MIMEApplicationJSON,
				req: &models.AdjustmentTrialBalanceRequest{
					AdjustmentDate: "2025-01",
				},
			},
			expectation: expectation{
				wantRes:  `{"status":"error","code":"INVALID_VALUES","message":"invalid format date caused by date 2025-01 format must be YYYY-MM-DD"}`,
				wantCode: 400,
			},
		},
		{
			name: "error - internal server error",
			args: args{
				contentType: echo.MIMEApplicationJSON,
				req: &models.AdjustmentTrialBalanceRequest{
					AdjustmentDate: "2025-01-01",
				},
			},
			expectation: expectation{
				wantRes:  `{"status":"error","code":"BQ_ERROR","message":"bq error"}`,
				wantCode: 500,
			},
			doMock: func(args args, expectation expectation) {
				testHelper.mockAccountingService.EXPECT().GenerateAdjustmentTrialBalanceBigQuery(ctx, gomock.Any()).Return(models.GetErrMap(models.ErrKeyBigQueryError))
			},
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
			r := httptest.NewRequest(http.MethodPost, "/api/v1/trial-balances/adjustment", &b)
			r.Header.Set(echo.HeaderContentType, tt.args.contentType)
			w := httptest.NewRecorder()
			testHelper.router.NewContext(r, w)
			testHelper.router.ServeHTTP(w, r)
			require.Equal(t, tt.expectation.wantCode, w.Code)
			require.Equal(t, tt.expectation.wantRes, strings.Trim(w.Body.String(), "\n"))
		})
	}
}
