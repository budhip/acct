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
	"time"

	"bitbucket.org/Amartha/go-accounting/internal/models"

	"github.com/labstack/echo/v4"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func Test_Handler_getSubLedgerAccounts(t *testing.T) {
	testHelper := accountingTestHelper(t)
	reqUrl := url.Values{
		"entityCode": []string{"001"},
		"search":     []string{"121001000000009"},
		"startDate":  []string{"2023-01-01"},
		"endDate":    []string{"2023-01-30"},
	}
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
			name: "success case - get sub ledger accounts",
			args: args{
				ctx:         context.Background(),
				contentType: echo.MIMEApplicationJSON,
				req:         reqUrl,
			},
			expectation: expectation{
				wantRes:  `{"kind":"collection","contents":[{"kind":"subLedgerAccount","accountNumber":"121001000000009","accountName":"Cash in Transit - Disburse Modal","altId":"","subCategoryCode":"12101","subCategoryName":"Cash in Transit","totalRowData":78,"createdAt":"0001-01-01 07:07:12"}],"pagination":{"prev":"","next":"","totalEntries":1}}`,
				wantCode: 200,
			},
			doMock: func(args args, expectation expectation) {
				testHelper.mockAccountingService.EXPECT().GetSubLedgerAccounts(args.ctx, gomock.Any()).Return(
					[]models.GetSubLedgerAccountsOut{
						{
							AccountNumber:   "121001000000009",
							AccountName:     "Cash in Transit - Disburse Modal",
							AltId:           "",
							SubCategoryCode: "12101",
							SubCategoryName: "Cash in Transit",
							TotalRowData:    78,
						},
					}, 1, nil)
			},
		},
		{
			name: "error case - validation",
			args: args{
				ctx:         context.Background(),
				contentType: echo.MIMEApplicationJSON,
				req: url.Values{
					"entityCode": []string{"001"},
					"search":     []string{"121001000000009"},
				},
			},
			expectation: expectation{
				wantRes:  `{"status":"error","message":"validation failed","errors":[{"code":"MISSING_FIELD","field":"startDate","message":"field is missing"},{"code":"MISSING_FIELD","field":"endDate","message":"field is missing"}]}`,
				wantCode: 422,
			},
		},
		{
			name: "error case - data not found",
			args: args{
				ctx:         context.Background(),
				contentType: echo.MIMEApplicationJSON,
				req:         reqUrl,
			},
			expectation: expectation{
				wantRes:  `{"status":"error","code":"DATA_NOT_FOUND","message":"data not found"}`,
				wantCode: 404,
			},
			doMock: func(args args, expectation expectation) {
				testHelper.mockAccountingService.EXPECT().GetSubLedgerAccounts(args.ctx, gomock.Any()).Return(
					[]models.GetSubLedgerAccountsOut{{}}, 1, models.GetErrMap(models.ErrKeyDataNotFound))
			},
		},
		{
			name: "error case - internal server error",
			args: args{
				ctx:         context.Background(),
				contentType: echo.MIMEApplicationJSON,
				req:         reqUrl,
			},
			expectation: expectation{
				wantRes:  `{"status":"error","code":"DATABASE_ERROR","message":"database error"}`,
				wantCode: 500,
			},
			doMock: func(args args, expectation expectation) {
				testHelper.mockAccountingService.EXPECT().GetSubLedgerAccounts(args.ctx, gomock.Any()).Return(
					[]models.GetSubLedgerAccountsOut{{}}, 1, models.GetErrMap(models.ErrKeyDatabaseError))
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

			r := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/sub-ledgers/accounts?%s", tt.args.req.Encode()), nil)
			r.Header.Set(echo.HeaderContentType, tt.args.contentType)
			w := httptest.NewRecorder()

			testHelper.router.NewContext(r, w)
			testHelper.router.ServeHTTP(w, r)

			require.Equal(t, tt.expectation.wantCode, w.Code)
			require.Equal(t, tt.expectation.wantRes, strings.Trim(w.Body.String(), "\n"))
		})
	}
}

func Test_Handler_getSubLedger(t *testing.T) {
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
			name: "success case - get sub ledgers",
			args: args{
				ctx:         context.Background(),
				contentType: echo.MIMEApplicationJSON,
				req: url.Values{
					"accountNumber": []string{"22201000000008"},
					"startDate":     []string{"2023-01-01"},
					"endDate":       []string{"2023-01-30"},
				},
			},
			expectation: expectation{
				wantRes:  `{"kind":"collection","details":{"kind":"account","accountNumber":"AMF","accountName":"PT AMARTHA MIKRO FINTEK","altId":"12345","coaTypeCode":"AST","entityCode":"AMF","entityName":"PT AMARTHA MIKRO FINTEK","productTypeCode":"1001","productTypeName":"Group Loan","subCategoryCode":"10000","subCategoryName":"RETAIL","currency":"PT AMARTHA MIKRO FINTEK","balancePeriodStart":"500.000,00"},"contents":[{"kind":"subLedger","transactionId":"3c8c389b-abbc-4d17-a718-7bd84721b40f","referenceNumber":"12345","transactionDate":"22 Dec 2023 07:00:00","transactionType":"DSBAC","transactionTypeName":"Admin Fee Partner Loan Deduction","narrative":"Invest To Loan","metadata":null,"debit":"30,00","credit":"0,00"}],"pagination":{"prev":"","next":"","totalEntries":1}}`,
				wantCode: 200,
			},
			doMock: func(args args, expectation expectation) {
				testHelper.mockAccountingService.EXPECT().GetSubLedger(args.ctx, gomock.Any()).Return(
					models.SubLedgerAccountResponse{
						Kind:               "account",
						AccountNumber:      "AMF",
						AccountName:        "PT AMARTHA MIKRO FINTEK",
						AltId:              "12345",
						COATypeCode:        "AST",
						EntityCode:         "AMF",
						EntityName:         "PT AMARTHA MIKRO FINTEK",
						ProductTypeCode:    "1001",
						ProductTypeName:    "Group Loan",
						SubCategoryCode:    "10000",
						SubCategoryName:    "RETAIL",
						Currency:           "PT AMARTHA MIKRO FINTEK",
						BalancePeriodStart: "500.000,00",
					},
					[]models.GetSubLedgerOut{
						{
							TransactionID:       "3c8c389b-abbc-4d17-a718-7bd84721b40f",
							ReferenceNumber:     "12345",
							TransactionDate:     time.Date(2023, time.December, 22, 00, 00, 00, 00, &time.Location{}),
							TransactionType:     "DSBAC",
							TransactionTypeName: "Admin Fee Partner Loan Deduction",
							Narrative:           "Invest To Loan",
							Debit:               decimal.New(30.000, 00),
							Credit:              decimal.New(0, 00),
						},
					}, 1, nil)
			},
		},
		{
			name: "error case - validation",
			args: args{
				ctx:         context.Background(),
				contentType: echo.MIMEApplicationJSON,
				req:         url.Values{},
			},
			expectation: expectation{
				wantRes:  `{"status":"error","message":"validation failed","errors":[{"code":"MISSING_FIELD","field":"accountNumber","message":"field is missing"},{"code":"MISSING_FIELD","field":"startDate","message":"field is missing"},{"code":"MISSING_FIELD","field":"endDate","message":"field is missing"}]}`,
				wantCode: 422,
			},
		},
		{
			name: "error case - limit < 0",
			args: args{
				ctx:         context.Background(),
				contentType: echo.MIMEApplicationJSON,
				req: url.Values{
					"accountNumber": []string{"22201000000008"},
					"startDate":     []string{"2023-01-01"},
					"endDate":       []string{"2023-01-30"},
					"limit":         []string{"-1"},
				},
			},
			expectation: expectation{
				wantRes:  `{"status":"error","code":"INVALID_VALUES","message":"the limit must be greater than zero"}`,
				wantCode: 400,
			},
		},
		{
			name: "error case - account number not found",
			args: args{
				ctx:         context.Background(),
				contentType: echo.MIMEApplicationJSON,
				req: url.Values{
					"accountNumber": []string{"22201000000008"},
					"startDate":     []string{"2023-01-01"},
					"endDate":       []string{"2023-01-30"},
				},
			},
			expectation: expectation{
				wantRes:  `{"status":"error","code":"DATA_NOT_FOUND","message":"account number not found"}`,
				wantCode: 404,
			},
			doMock: func(args args, expectation expectation) {
				testHelper.mockAccountingService.EXPECT().GetSubLedger(args.ctx, gomock.Any()).Return(
					models.SubLedgerAccountResponse{},
					[]models.GetSubLedgerOut{}, 0, models.GetErrMap(models.ErrKeyAccountNumberNotFound))
			},
		},
		{
			name: "error case - data sub ledgers exceed limit will send to email",
			args: args{
				ctx:         context.Background(),
				contentType: echo.MIMEApplicationJSON,
				req: url.Values{
					"accountNumber": []string{"22201000000008"},
					"startDate":     []string{"2023-01-01"},
					"endDate":       []string{"2023-01-30"},
				},
			},
			expectation: expectation{
				wantRes:  `{"status":"error","code":"DATA_IS_EXCEEDS_LIMIT","message":"the data is too big and we will send a csv file to your email"}`,
				wantCode: 200,
			},
			doMock: func(args args, expectation expectation) {
				testHelper.mockAccountingService.EXPECT().GetSubLedger(args.ctx, gomock.Any()).Return(
					models.SubLedgerAccountResponse{},
					[]models.GetSubLedgerOut{}, 100, models.GetErrMap(models.ErrKeyDataGlisExceedsTheLimit))
			},
		},
		{
			name: "error case - internal server error",
			args: args{
				ctx:         context.Background(),
				contentType: echo.MIMEApplicationJSON,
				req: url.Values{
					"accountNumber": []string{"22201000000008"},
					"startDate":     []string{"2023-01-01"},
					"endDate":       []string{"2023-01-30"},
				},
			},
			expectation: expectation{
				wantRes:  `{"status":"error","code":"DATABASE_ERROR","message":"database error"}`,
				wantCode: 500,
			},
			doMock: func(args args, expectation expectation) {
				testHelper.mockAccountingService.EXPECT().GetSubLedger(args.ctx, gomock.Any()).Return(
					models.SubLedgerAccountResponse{},
					[]models.GetSubLedgerOut{}, 0, models.GetErrMap(models.ErrKeyDatabaseError))
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

			r := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/sub-ledgers?%s", tt.args.req.Encode()), nil)
			r.Header.Set(echo.HeaderContentType, tt.args.contentType)
			w := httptest.NewRecorder()

			testHelper.router.NewContext(r, w)
			testHelper.router.ServeHTTP(w, r)

			require.Equal(t, tt.expectation.wantCode, w.Code)
			require.Equal(t, tt.expectation.wantRes, strings.Trim(w.Body.String(), "\n"))
		})
	}
}

func Test_Handler_getSubLedgerCount(t *testing.T) {
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
			name: "success case - get sub ledgers count",
			args: args{
				ctx:         context.Background(),
				contentType: echo.MIMEApplicationJSON,
				req: url.Values{
					"accountNumber": []string{"22201000000008"},
					"startDate":     []string{"2023-01-01"},
					"endDate":       []string{"2023-01-30"},
				},
			},
			expectation: expectation{
				wantRes:  `{"kind":"subLedger","total":10,"isExceedsTheLimit":false}`,
				wantCode: 200,
			},
			doMock: func(args args, expectation expectation) {
				testHelper.mockAccountingService.EXPECT().GetSubLedgerCount(args.ctx, gomock.Any()).Return(
					models.GetSubLedgerCountOut{Total: 10}, nil)
			},
		},
		{
			name: "error case - validation",
			args: args{
				ctx:         context.Background(),
				contentType: echo.MIMEApplicationJSON,
				req:         url.Values{},
			},
			expectation: expectation{
				wantRes:  `{"status":"error","message":"validation failed","errors":[{"code":"MISSING_FIELD","field":"accountNumber","message":"field is missing"},{"code":"MISSING_FIELD","field":"startDate","message":"field is missing"},{"code":"MISSING_FIELD","field":"endDate","message":"field is missing"}]}`,
				wantCode: 422,
			},
		},
		{
			name: "error case - internal server error",
			args: args{
				ctx:         context.Background(),
				contentType: echo.MIMEApplicationJSON,
				req: url.Values{
					"accountNumber": []string{"22201000000008"},
					"startDate":     []string{"2023-01-01"},
					"endDate":       []string{"2023-01-30"},
				},
			},
			expectation: expectation{
				wantRes:  `{"status":"error","code":"DATABASE_ERROR","message":"database error"}`,
				wantCode: 500,
			},
			doMock: func(args args, expectation expectation) {
				testHelper.mockAccountingService.EXPECT().GetSubLedgerCount(args.ctx, gomock.Any()).Return(
					models.GetSubLedgerCountOut{}, models.GetErrMap(models.ErrKeyDatabaseError))
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

			r := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/sub-ledgers/count?%s", tt.args.req.Encode()), nil)
			r.Header.Set(echo.HeaderContentType, tt.args.contentType)
			w := httptest.NewRecorder()

			testHelper.router.NewContext(r, w)
			testHelper.router.ServeHTTP(w, r)

			require.Equal(t, tt.expectation.wantCode, w.Code)
			require.Equal(t, tt.expectation.wantRes, strings.Trim(w.Body.String(), "\n"))
		})
	}
}

func Test_Handler_sendSubLedgerCSVToEmail(t *testing.T) {
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
					"accountNumber": []string{"22201000000008"},
					"startDate":     []string{"2023-01-01"},
					"endDate":       []string{"2023-01-30"},
					"email":         []string{"tono@amartha.com"},
				},
			},
			expectation: expectation{
				wantRes:  `"processing"`,
				wantCode: 200,
			},
			doMock: func(args args, expectation expectation) {
				testHelper.mockAccountingService.EXPECT().SendSubLedgerCSVToEmail(gomock.Any(), gomock.Any()).Return(nil)
			},
		},
		{
			name: "error case - error convert to filter options",
			args: args{
				ctx:         context.Background(),
				contentType: echo.MIMEApplicationJSON,
				req: url.Values{
					"accountNumber": []string{"22201000000008"},
					"startDate":     []string{"20230101"},
					"endDate":       []string{"20230130"},
					"email":         []string{"tono@amartha.com"},
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
				wantRes:  `{"status":"error","message":"validation failed","errors":[{"code":"MISSING_FIELD","field":"accountNumber","message":"field is missing"},{"code":"MISSING_FIELD","field":"startDate","message":"field is missing"},{"code":"MISSING_FIELD","field":"endDate","message":"field is missing"},{"code":"UNKNOW","field":"email","message":"required"}]}`,
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

			r := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/sub-ledgers/send-email?%s", tt.args.req.Encode()), nil)
			r.Header.Set(echo.HeaderContentType, tt.args.contentType)
			w := httptest.NewRecorder()

			testHelper.router.NewContext(r, w)
			testHelper.router.ServeHTTP(w, r)

			require.Equal(t, tt.expectation.wantCode, w.Code)
			require.Equal(t, tt.expectation.wantRes, strings.Trim(w.Body.String(), "\n"))
		})
	}
}
