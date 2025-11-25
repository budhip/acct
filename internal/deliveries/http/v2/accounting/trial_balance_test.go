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
	"github.com/golang/mock/gomock"
	"github.com/labstack/echo/v4"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_Handler_getTrialBalanceDetail(t *testing.T) {
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
					"period":          []string{"2025-01"},
				},
			},
			expectation: expectation{
				wantRes:  `{"kind":"collection","contents":[{"kind":"trialBalanceDetail","accountNumber":"123456","accountName":"Shinji Takeru","openingBalance":"100.000,00","debitMovement":"50.000,00","creditMovement":"150.000,00","closingBalance":"0,00"}],"summary":{"kind":"trialBalanceSubCategory","subCategoryCode":"21108","subCategoryName":"Cash Point","openingBalance":"100.000,00","debitMovement":"50.000,00","creditMovement":"150.000,00","closingBalance":"0,00"}}`,
				wantCode: 200,
			},
			doMock: func(args args, expectation expectation) {
				testHelper.mockAccountingService.EXPECT().GetTrialBalanceFromGCS(gomock.Any(), gomock.Any()).Return([]models.TrialBalanceDetailOut{
					{
						AccountNumber:  "123456",
						AccountName:    "Shinji Takeru",
						OpeningBalance: decimal.NewFromInt(100_000),
						DebitMovement:  decimal.NewFromInt(50_000),
						CreditMovement: decimal.NewFromInt(150_000),
						ClosingBalance: decimal.Zero,
					},
				}, models.TrialBalanceBySubCategoryOut{
					SubCategoryCode: "21108",
					SubCategoryName: "Cash Point",
					OpeningBalance:  decimal.NewFromInt(100_000),
					DebitMovement:   decimal.NewFromInt(50_000),
					CreditMovement:  decimal.NewFromInt(150_000),
					ClosingBalance:  decimal.Zero,
				}, nil)
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
					"period":          []string{"2025-01"},
				},
			},
			expectation: expectation{
				wantRes:  `{"status":"error","code":500,"message":"assert.AnError general error for testing"}`,
				wantCode: 500,
			},
			doMock: func(args args, expectation expectation) {
				testHelper.mockAccountingService.EXPECT().
					GetTrialBalanceFromGCS(gomock.Any(), gomock.Any()).
					Return(nil, models.TrialBalanceBySubCategoryOut{}, assert.AnError)
			},
		},
		{
			name: "error case - param period is missing",
			args: args{
				ctx:         context.Background(),
				contentType: echo.MIMEApplicationJSON,
				req: url.Values{
					"subCategoryCode": []string{"21108"},
					"entityCode":      []string{"001"},
				},
			},
			expectation: expectation{
				wantRes:  `{"status":"error","message":"validation failed","errors":[{"code":"UNKNOW","field":"period","message":"required"}]}`,
				wantCode: 422,
			},
			doMock: func(args args, expectation expectation) {
				testHelper.mockAccountingService.EXPECT().
					GetTrialBalanceFromGCS(gomock.Any(), gomock.Any()).
					Return(nil, models.TrialBalanceBySubCategoryOut{}, assert.AnError)
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

			r := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v2/trial-balances/details?%s", tt.args.req.Encode()), nil)
			r.Header.Set(echo.HeaderContentType, tt.args.contentType)
			w := httptest.NewRecorder()

			testHelper.router.NewContext(r, w)
			testHelper.router.ServeHTTP(w, r)

			require.Equal(t, tt.expectation.wantCode, w.Code)
			require.Equal(t, tt.expectation.wantRes, strings.Trim(w.Body.String(), "\n"))
		})
	}
}

func Test_Handler_sendTrialBalanceDetailsToEmail(t *testing.T) {
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
			name: "error case - error convert to filter options",
			args: args{
				ctx:         context.Background(),
				contentType: echo.MIMEApplicationJSON,
				req: url.Values{
					"subCategoryCode": []string{"21108"},
					"entityCode":      []string{"001"},
					"period":          []string{"202502"},
					"email":           []string{"tono@amartha.com"},
				},
			},
			expectation: expectation{
				wantRes:  `{"status":"error","code":"INVALID_VALUES","message":"invalid format date caused by date 202502 format must be YYYY-MM"}`,
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
				wantRes:  `{"status":"error","message":"validation failed","errors":[{"code":"MISSING_FIELD","field":"entityCode","message":"field is missing"},{"code":"MISSING_FIELD","field":"subCategoryCode","message":"field is missing"},{"code":"UNKNOW","field":"email","message":"required"},{"code":"UNKNOW","field":"period","message":"required"}]}`,
				wantCode: 422,
			},
		},
		{
			name: "success case",
			args: args{
				ctx:         context.Background(),
				contentType: echo.MIMEApplicationJSON,
				req: url.Values{
					"subCategoryCode": []string{"21108"},
					"entityCode":      []string{"001"},
					"period":          []string{"2023-01"},
					"email":           []string{"tono@amartha.com"},
				},
			},
			expectation: expectation{
				wantRes:  `"processing"`,
				wantCode: 200,
			},
			doMock: func(args args, expectation expectation) {
				testHelper.mockAccountingService.EXPECT().SendEmailTrialBalanceDetails(gomock.Any(), gomock.Any()).Return(nil)
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

			r := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v2/trial-balances/details/download?%s", tt.args.req.Encode()), nil)
			r.Header.Set(echo.HeaderContentType, tt.args.contentType)
			w := httptest.NewRecorder()

			testHelper.router.NewContext(r, w)
			testHelper.router.ServeHTTP(w, r)

			require.Equal(t, tt.expectation.wantCode, w.Code)
			require.Equal(t, tt.expectation.wantRes, strings.Trim(w.Body.String(), "\n"))
		})
	}
}
