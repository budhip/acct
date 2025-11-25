package loanpartneraccount

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

	"bitbucket.org/Amartha/go-accounting/internal/models"
	"bitbucket.org/Amartha/go-accounting/internal/services/mock"
	xlog "bitbucket.org/Amartha/go-x/log"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

type testLoanPartnerAccountHelper struct {
	router      *echo.Echo
	mockCtrl    *gomock.Controller
	mockService *mock.MockLoanPartnerService
}

func loanPartnerAccountTestHelper(t *testing.T) testLoanPartnerAccountHelper {
	t.Helper()

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockSvc := mock.NewMockLoanPartnerService(mockCtrl)

	app := echo.New()
	v1Group := app.Group("/api/v1")
	New(v1Group, mockSvc)

	return testLoanPartnerAccountHelper{
		router:      app,
		mockCtrl:    mockCtrl,
		mockService: mockSvc,
	}
}

func TestMain(m *testing.M) {
	xlog.InitForTest()
	os.Exit(m.Run())
}

func Test_Handler_create(t *testing.T) {
	testHelper := loanPartnerAccountTestHelper(t)
	req := models.DoCreateLoanPartnerAccountRequest{
		PartnerId:           "efishery",
		LoanKind:            "EFISHERY_LOAN",
		AccountNumber:       "22100100000001",
		AccountType:         "INTERNAL_ACCOUNTS_REVENUE_AMARTHA",
		LoanSubCategoryCode: "13101",
	}
	in := models.LoanPartnerAccount{
		PartnerId:           req.PartnerId,
		LoanKind:            req.LoanKind,
		AccountNumber:       req.AccountNumber,
		AccountType:         req.AccountType,
		LoanSubCategoryCode: req.LoanSubCategoryCode,
	}
	type args struct {
		ctx         context.Context
		contentType string
		req         models.DoCreateLoanPartnerAccountRequest
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
				req:         req,
			},
			expectation: expectation{
				wantRes:  `{"kind":"loanPartnerAccount","partnerId":"efishery","loanKind":"EFISHERY_LOAN","accountNumber":"22100100000001","accountType":"INTERNAL_ACCOUNTS_REVENUE_AMARTHA","loanSubCategoryCode":"13101"}`,
				wantCode: 201,
			},
			doMock: func(args args, expectation expectation) {
				testHelper.mockService.EXPECT().Create(args.ctx, in).Return(models.LoanPartnerAccount(in), nil)
			},
		},
		{
			name: "error case - contentType",
			args: args{
				ctx:         context.Background(),
				contentType: "",
			},
			expectation: expectation{
				wantRes:  `{"status":"error","code":400,"message":"code=415, message=Unsupported Media Type"}`,
				wantCode: 400,
			},
		},
		{
			name: "error case - validation failed",
			args: args{
				ctx:         context.Background(),
				contentType: echo.MIMEApplicationJSON,
				req: models.DoCreateLoanPartnerAccountRequest{
					AccountType: req.AccountType,
				},
			},
			expectation: expectation{
				wantRes:  `{"status":"error","message":"validation failed","errors":[{"code":"MISSING_FIELD","field":"partnerId","message":"field is missing"},{"code":"MISSING_FIELD","field":"loanKind","message":"field is missing"},{"code":"MISSING_FIELD","field":"accountNumber","message":"field is missing"},{"code":"MISSING_FIELD","field":"loanSubCategoryCode","message":"field is missing"}]}`,
				wantCode: 422,
			},
		},
		{
			name: "error case - account number not found",
			args: args{
				ctx:         context.Background(),
				contentType: echo.MIMEApplicationJSON,
				req:         req,
			},
			expectation: expectation{
				wantRes:  `{"status":"error","code":"DATA_NOT_FOUND","message":"account number not found"}`,
				wantCode: 404,
			},
			doMock: func(args args, expectation expectation) {
				testHelper.mockService.EXPECT().Create(args.ctx, in).Return(models.LoanPartnerAccount{}, models.GetErrMap(models.ErrKeyAccountNumberNotFound))
			},
		},
		{
			name: "error case - account number is exist",
			args: args{
				ctx:         context.Background(),
				contentType: echo.MIMEApplicationJSON,
				req:         req,
			},
			expectation: expectation{
				wantRes:  `{"status":"error","code":"DATA_IS_EXIST","message":"account number is exist"}`,
				wantCode: 409,
			},
			doMock: func(args args, expectation expectation) {
				testHelper.mockService.EXPECT().Create(args.ctx, in).Return(models.LoanPartnerAccount{}, models.GetErrMap(models.ErrKeyAccountNumberIsExist))
			},
		},
		{
			name: "error case - data is exist",
			args: args{
				ctx:         context.Background(),
				contentType: echo.MIMEApplicationJSON,
				req:         req,
			},
			expectation: expectation{
				wantRes:  `{"status":"error","code":"DATA_IS_EXIST","message":"data is exist"}`,
				wantCode: 409,
			},
			doMock: func(args args, expectation expectation) {
				testHelper.mockService.EXPECT().Create(args.ctx, in).Return(models.LoanPartnerAccount{}, models.GetErrMap(models.ErrKeyDataIsExist))
			},
		},
		{
			name: "error case - database error",
			args: args{
				ctx:         context.Background(),
				contentType: echo.MIMEApplicationJSON,
				req:         req,
			},
			expectation: expectation{
				wantRes:  `{"status":"error","code":"INTERNAL_SERVER_ERROR","message":"unable to create data"}`,
				wantCode: 500,
			},
			doMock: func(args args, expectation expectation) {
				testHelper.mockService.EXPECT().Create(args.ctx, in).Return(models.LoanPartnerAccount{}, models.GetErrMap(models.ErrKeyUnableToCreateData))
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

			r := httptest.NewRequest(http.MethodPost, "/api/v1/loan-partner-accounts", &b)
			r.Header.Set(echo.HeaderContentType, tt.args.contentType)
			w := httptest.NewRecorder()

			testHelper.router.NewContext(r, w)
			testHelper.router.ServeHTTP(w, r)

			require.Equal(t, tt.expectation.wantCode, w.Code)
			require.Equal(t, tt.expectation.wantRes, strings.Trim(w.Body.String(), "\n"))
		})
	}
}

func Test_Handler_update(t *testing.T) {
	testHelper := loanPartnerAccountTestHelper(t)
	req := models.DoUpdateLoanPartnerAccountRequest{
		PartnerId:     "efishery",
		LoanKind:      "EFISHERY_LOAN",
		AccountNumber: "22100100000001",
	}
	in := models.UpdateLoanPartnerAccount{
		PartnerId:     req.PartnerId,
		LoanKind:      req.LoanKind,
		AccountNumber: req.AccountNumber,
	}
	type args struct {
		ctx         context.Context
		contentType string
		req         models.DoUpdateLoanPartnerAccountRequest
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
				req:         req,
			},
			expectation: expectation{
				wantRes:  `{"kind":"loanPartnerAccount","partnerId":"efishery","loanKind":"EFISHERY_LOAN","accountNumber":"22100100000001"}`,
				wantCode: 200,
			},
			doMock: func(args args, expectation expectation) {
				testHelper.mockService.EXPECT().Update(args.ctx, in).Return(models.UpdateLoanPartnerAccount(in), nil)
			},
		},
		{
			name: "error case - contentType",
			args: args{
				ctx:         context.Background(),
				contentType: "",
			},
			expectation: expectation{
				wantRes:  `{"status":"error","code":400,"message":"code=415, message=Unsupported Media Type"}`,
				wantCode: 400,
			},
		},
		{
			name: "error case - validation failed",
			args: args{
				ctx:         context.Background(),
				contentType: echo.MIMEApplicationJSON,
				req: models.DoUpdateLoanPartnerAccountRequest{
					AccountNumber: "123",
				},
			},
			expectation: expectation{
				wantRes:  `{"status":"error","message":"validation failed","errors":[{"code":"MISSING_FIELD","field":"partnerId","message":"required fields at least partnerId"},{"code":"MISSING_FIELD","field":"loanKind","message":"required fields at least loanKind"},{"code":"MISSING_FIELD","field":"accountType","message":"required fields at least accountType"},{"code":"MISSING_FIELD","field":"loanSubCategoryCode","message":"required fields at least loanSubCategoryCode"}]}`,
				wantCode: 422,
			},
		},
		{
			name: "error case - account number not found",
			args: args{
				ctx:         context.Background(),
				contentType: echo.MIMEApplicationJSON,
				req:         req,
			},
			expectation: expectation{
				wantRes:  `{"status":"error","code":"DATA_NOT_FOUND","message":"account number not found"}`,
				wantCode: 404,
			},
			doMock: func(args args, expectation expectation) {
				testHelper.mockService.EXPECT().Update(args.ctx, in).Return(models.UpdateLoanPartnerAccount{}, models.GetErrMap(models.ErrKeyAccountNumberNotFound))
			},
		},
		{
			name: "error case - database error",
			args: args{
				ctx:         context.Background(),
				contentType: echo.MIMEApplicationJSON,
				req:         req,
			},
			expectation: expectation{
				wantRes:  `{"status":"error","code":"DATABASE_ERROR","message":"database error"}`,
				wantCode: 500,
			},
			doMock: func(args args, expectation expectation) {
				testHelper.mockService.EXPECT().Update(args.ctx, in).Return(models.UpdateLoanPartnerAccount{}, models.GetErrMap(models.ErrKeyDatabaseError))
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

			r := httptest.NewRequest(http.MethodPatch, "/api/v1/loan-partner-accounts/:accountNumber", &b)
			r.Header.Set(echo.HeaderContentType, tt.args.contentType)
			w := httptest.NewRecorder()

			testHelper.router.NewContext(r, w)
			testHelper.router.ServeHTTP(w, r)

			require.Equal(t, tt.expectation.wantCode, w.Code)
			require.Equal(t, tt.expectation.wantRes, strings.Trim(w.Body.String(), "\n"))
		})
	}
}

func Test_Handler_getByParams(t *testing.T) {
	testHelper := loanPartnerAccountTestHelper(t)
	req := models.GetLoanPartnerAccountByParamsIn{
		PartnerId:   "efishery",
		LoanKind:    "EFISHERY_LOAN",
		AccountType: "INTERNAL_ACCOUNTS_REVENUE_AMARTHA",
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
			name: "success case",
			args: args{
				ctx:         context.Background(),
				contentType: echo.MIMEApplicationJSON,
				req: url.Values{
					"partnerId":   []string{req.PartnerId},
					"loanKind":    []string{req.LoanKind},
					"accountType": []string{req.AccountType},
				},
			},
			expectation: expectation{
				wantRes:  `{"kind":"collection","contents":[{"kind":"loanPartnerAccount","partnerId":"efishery","loanKind":"EFISHERY_LOAN","accountNumber":"22100100000001","accountType":"INTERNAL_ACCOUNTS_REVENUE_AMARTHA","entityCode":"001","loanSubCategoryCode":"13101","createdAt":"0001-01-01 07:07:12","updatedAt":"0001-01-01 07:07:12"}],"total_rows":1}`,
				wantCode: 200,
			},
			doMock: func(args args, expectation expectation) {
				testHelper.mockService.EXPECT().GetByParams(gomock.Any(), req).Return([]models.LoanPartnerAccount{
					{
						PartnerId:           req.PartnerId,
						LoanKind:            req.LoanKind,
						AccountNumber:       "22100100000001",
						EntityCode:          "001",
						AccountType:         req.AccountType,
						LoanSubCategoryCode: "13101",
					},
				}, nil)
			},
		},
		{
			name: "error case - validation error",
			args: args{
				ctx:         context.Background(),
				contentType: echo.MIMEApplicationJSON,
				req: url.Values{
					"entityCode": []string{"1234567890"},
				},
			},
			expectation: expectation{
				wantCode: 422,
				wantRes:  `{"status":"error","message":"validation failed","errors":[{"code":"INVALID_LENGTH","field":"entityCode","message":"field can have a maximum length of 3 characters"}]}`,
			},
		},
		{
			name: "error case - account number not found",
			args: args{
				ctx:         context.Background(),
				contentType: echo.MIMEApplicationJSON,
				req: url.Values{
					"entityCode": []string{"007"},
				},
			},
			expectation: expectation{
				wantRes:  `{"status":"error","code":"DATA_NOT_FOUND","message":"account number not found"}`,
				wantCode: 404,
			},
			doMock: func(args args, expectation expectation) {
				testHelper.mockService.EXPECT().GetByParams(gomock.Any(), gomock.Any()).Return([]models.LoanPartnerAccount{},
					models.GetErrMap(models.ErrKeyAccountNumberNotFound))
			},
		},
		{
			name: "error case - entity code not found",
			args: args{
				ctx:         context.Background(),
				contentType: echo.MIMEApplicationJSON,
				req: url.Values{
					"entityCode": []string{"007"},
				},
			},
			expectation: expectation{
				wantRes:  `{"status":"error","code":"DATA_NOT_FOUND","message":"entity code not found"}`,
				wantCode: 404,
			},
			doMock: func(args args, expectation expectation) {
				testHelper.mockService.EXPECT().GetByParams(gomock.Any(), gomock.Any()).Return([]models.LoanPartnerAccount{},
					models.GetErrMap(models.ErrKeyEntityCodeNotFound))
			},
		},
		{
			name: "error case - data not found",
			args: args{
				ctx:         context.Background(),
				contentType: echo.MIMEApplicationJSON,
				req: url.Values{
					"entityCode": []string{"007"},
				},
			},
			expectation: expectation{
				wantRes:  `{"status":"error","code":"DATA_NOT_FOUND","message":"data not found"}`,
				wantCode: 404,
			},
			doMock: func(args args, expectation expectation) {
				testHelper.mockService.EXPECT().GetByParams(gomock.Any(), gomock.Any()).Return([]models.LoanPartnerAccount{},
					models.GetErrMap(models.ErrKeyDataNotFound))
			},
		},
		{
			name: "error case - internal server error",
			args: args{
				ctx:         context.Background(),
				contentType: echo.MIMEApplicationJSON,
				req: url.Values{
					"entityCode": []string{"007"},
				},
			},
			expectation: expectation{
				wantRes:  `{"status":"error","code":"DATABASE_ERROR","message":"database error"}`,
				wantCode: 500,
			},
			doMock: func(args args, expectation expectation) {
				testHelper.mockService.EXPECT().GetByParams(gomock.Any(), gomock.Any()).Return([]models.LoanPartnerAccount{},
					models.GetErrMap(models.ErrKeyDatabaseError))
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

			r := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/loan-partner-accounts?%s", tt.args.req.Encode()), nil)
			r.Header.Set(echo.HeaderContentType, tt.args.contentType)
			w := httptest.NewRecorder()

			testHelper.router.NewContext(r, w)
			testHelper.router.ServeHTTP(w, r)

			require.Equal(t, tt.expectation.wantCode, w.Code)
			require.Equal(t, tt.expectation.wantRes, strings.Trim(w.Body.String(), "\n"))
		})
	}
}
