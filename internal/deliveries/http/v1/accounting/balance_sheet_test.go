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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func Test_Handler_getBalanceSheet(t *testing.T) {
	testHelper := accountingTestHelper(t)
	defaultQueryFilter := new(models.GetBalanceSheetRequest)
	defaultQueryFilter.EntityCode = "111"
	defaultQueryFilter.BalanceSheetDate = "2023-12-31"
	defaultOpts, _ := defaultQueryFilter.ToFilterOpts()

	type args struct {
		ctx         context.Context
		contentType string
		req         url.Values
	}
	type expectation struct {
		wantRes  string
		wantCode int
	}
	testCases := []struct {
		name        string
		args        args
		expectation expectation
		doMock      func(args args, expectation expectation)
	}{
		{
			name: "success case - get balance sheet",
			args: args{
				ctx:         context.Background(),
				contentType: echo.MIMEApplicationJSON,
				req: url.Values{
					"entityCode":       []string{"111"},
					"balanceSheetDate": []string{"2023-12-31"},
				},
			},
			expectation: expectation{
				wantRes:  `{"kind":"","entityCode":"","entityName":"","entityDesc":"","balanceSheetDate":"","balanceSheet":{"kind":"","assets":null,"totalAsset":"","liabilities":null,"totalLiability":"","catchAll":""}}`,
				wantCode: 200,
			},
			doMock: func(args args, expectation expectation) {
				testHelper.mockAccountingService.EXPECT().GetBalanceSheet(args.ctx, *defaultOpts).Return(models.GetBalanceSheetResponse{}, nil)
			},
		},
		{
			name: "error case - database error",
			args: args{
				ctx:         context.Background(),
				contentType: echo.MIMEApplicationJSON,
				req: url.Values{
					"entityCode":       []string{"111"},
					"balanceSheetDate": []string{"2023-12-31"},
				},
			},
			expectation: expectation{
				wantRes:  `{"status":"error","code":"DATABASE_ERROR","message":"database error"}`,
				wantCode: 500,
			},
			doMock: func(args args, expectation expectation) {
				testHelper.mockAccountingService.EXPECT().GetBalanceSheet(args.ctx, *defaultOpts).Return(models.GetBalanceSheetResponse{}, models.GetErrMap(models.ErrKeyDatabaseError))
			},
		},
		{
			name: "error case - entity code not found",
			args: args{
				ctx:         context.Background(),
				contentType: echo.MIMEApplicationJSON,
				req: url.Values{
					"entityCode":       []string{"000"},
					"balanceSheetDate": []string{"2023-12-31"},
				},
			},
			expectation: expectation{
				wantRes:  `{"status":"error","code":"DATA_NOT_FOUND","message":"entity code not found"}`,
				wantCode: 404,
			},
			doMock: func(args args, expectation expectation) {
				testHelper.mockAccountingService.EXPECT().GetBalanceSheet(args.ctx, gomock.Any()).Return(models.GetBalanceSheetResponse{}, models.GetErrMap(models.ErrKeyEntityCodeNotFound))
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.doMock != nil {
				tc.doMock(tc.args, tc.expectation)
			}

			var b bytes.Buffer
			err := json.NewEncoder(&b).Encode(tc.args.req)
			require.NoError(t, err)

			r := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/balance-sheets?%s", tc.args.req.Encode()), nil)
			r.Header.Set(echo.HeaderContentType, tc.args.contentType)
			w := httptest.NewRecorder()

			testHelper.router.NewContext(r, w)
			testHelper.router.ServeHTTP(w, r)

			require.Equal(t, tc.expectation.wantCode, w.Code)
			require.Equal(t, tc.expectation.wantRes, strings.Trim(w.Body.String(), "\n"))
		})
	}
}

func Test_Handler_downloadCSVBalanceSheet(t *testing.T) {
	testHelper := accountingTestHelper(t)
	defaultQueryFilter := new(models.GetBalanceSheetRequest)
	defaultQueryFilter.EntityCode = "111"
	defaultQueryFilter.BalanceSheetDate = "2023-12-31"
	defaultOpts, _ := defaultQueryFilter.ToFilterOpts()
	req := url.Values{
		"entityCode":       []string{"111"},
		"balanceSheetDate": []string{"2023-12-31"},
	}
	res := models.GetBalanceSheetResponse{
		EntityCode:       "001",
		BalanceSheetDate: "2024-01-01",
		BalanceSheet: models.BalanceSheetData{
			Assets: []models.BalanceCategory{
				{
					CategoryCode: "121",
					CategoryName: "Cash Point",
					Amount:       "0,00",
				},
			},
			Liabilities: []models.BalanceCategory{
				{
					CategoryCode: "211",
					CategoryName: "Cash Point",
					Amount:       "0,00",
				},
			},
			TotalAsset:     "0,00",
			TotalLiability: "0,00",
			CatchAll:       "0,00",
		},
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
			name: "success case - download balance sheet",
			args: args{
				ctx:         context.Background(),
				contentType: echo.MIMEApplicationJSON,
				req:         req,
			},
			expectation: expectation{
				wantCode: 200,
			},
			doMock: func(args args, expectation expectation) {
				testHelper.mockAccountingService.EXPECT().GetBalanceSheet(args.ctx, *defaultOpts).Return(res, nil)
				testHelper.mockAccountingService.EXPECT().DownloadCSVGetBalanceSheet(gomock.Any(), *defaultOpts, res).Return(&bytes.Buffer{}, "", nil)
			},
		},
		{
			name: "error case - entity not found - get balance sheet",
			args: args{
				ctx:         context.Background(),
				contentType: echo.MIMEApplicationJSON,
				req:         req,
			},
			expectation: expectation{
				wantRes:  `{"status":"error","code":"DATA_NOT_FOUND","message":"entity code not found"}`,
				wantCode: 404,
			},
			doMock: func(args args, expectation expectation) {
				testHelper.mockAccountingService.EXPECT().GetBalanceSheet(args.ctx, *defaultOpts).Return(res, models.GetErrMap(models.ErrKeyEntityCodeNotFound))
			},
		},
		{
			name: "error case - internal server error - get balance sheet",
			args: args{
				ctx:         context.Background(),
				contentType: echo.MIMEApplicationJSON,
				req:         req,
			},
			expectation: expectation{
				wantRes:  `{"status":"error","code":500,"message":"assert.AnError general error for testing"}`,
				wantCode: 500,
			},
			doMock: func(args args, expectation expectation) {
				testHelper.mockAccountingService.EXPECT().GetBalanceSheet(args.ctx, *defaultOpts).Return(res, assert.AnError)
			},
		},
		{
			name: "error case - download csv balance sheet",
			args: args{
				ctx:         context.Background(),
				contentType: echo.MIMEApplicationJSON,
				req:         req,
			},
			expectation: expectation{
				wantRes:  `{"status":"error","code":500,"message":"assert.AnError general error for testing"}`,
				wantCode: 500,
			},
			doMock: func(args args, expectation expectation) {
				testHelper.mockAccountingService.EXPECT().GetBalanceSheet(args.ctx, *defaultOpts).Return(res, nil)
				testHelper.mockAccountingService.EXPECT().DownloadCSVGetBalanceSheet(gomock.Any(), *defaultOpts, res).Return(&bytes.Buffer{}, "", assert.AnError)
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

			r := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/balance-sheets/download?%s", tt.args.req.Encode()), nil)
			r.Header.Set(echo.HeaderContentType, tt.args.contentType)
			w := httptest.NewRecorder()

			testHelper.router.NewContext(r, w)
			testHelper.router.ServeHTTP(w, r)

			require.Equal(t, tt.expectation.wantCode, w.Code)
			require.Equal(t, tt.expectation.wantRes, strings.Trim(w.Body.String(), "\n"))
		})
	}
}
