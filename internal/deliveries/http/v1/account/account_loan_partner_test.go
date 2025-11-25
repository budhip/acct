package account

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"bitbucket.org/Amartha/go-accounting/internal/models"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
)

func Test_Handler_createLoanPartnerAccount(t *testing.T) {
	testHelper := accountTestHelper(t)

	type args struct {
		ctx         context.Context
		contentType string
		req         models.DoCreateAccountLoanPartnerRequest
	}
	var mockCreateLoanPartnerOut = models.AccountsLoanPartner{
		PartnerName: "Chickin ABCDE",
		PartnerId:   "9876",
		AccountNumbers: []models.LoanPartnerAccountNumbers{
			{
				Entity:                         "001",
				CashInTransitDisburseDeduction: "121001000000101",
				CashInTransitRepayment:         "121001000000102",
				AmarthaRevenue:                 "221001000000102",
				AdminFee:                       "221001000000101",
				WHT2326:                        "241001000000102",
				VATOut:                         "241001000000101",
			},
			{
				Entity:                         "003",
				CashInTransitDisburseDeduction: "121001000000101",
				CashInTransitRepayment:         "121001000000102",
				AmarthaRevenue:                 "221001000000102",
				AdminFee:                       "221001000000101",
				WHT2326:                        "241001000000102",
				VATOut:                         "241001000000101",
			},
		},
		Metadata: &models.Metadata{},
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
				ctx:         context.Background(),
				contentType: echo.MIMEApplicationJSON,
				req: models.DoCreateAccountLoanPartnerRequest{
					PartnerName: "Chickin ABCDE",
					PartnerId:   "9876",
					LoanKind:    "PartnershipLoan",
				},
			},
			expectation: expectation{
				wantRes:  `{"kind":"loanPartnerAccount","partnerName":"Chickin ABCDE","partnerId":"9876","accountNumbers":[{"entity":"001","cashInTransitDisburseDeduction":"121001000000101","cashInTransitRepayment":"121001000000102","amarthaRevenue":"221001000000102","adminFee":"221001000000101","wht23_26":"241001000000102","vatOut":"241001000000101"},{"entity":"003","cashInTransitDisburseDeduction":"121001000000101","cashInTransitRepayment":"121001000000102","amarthaRevenue":"221001000000102","adminFee":"221001000000101","wht23_26":"241001000000102","vatOut":"241001000000101"}],"metadata":{}}`,
				wantCode: 201,
			},
			doMock: func(args args, expectation expectation) {
				testHelper.mockAccountService.EXPECT().CreateLoanPartnerAccount(args.ctx, models.CreateAccountLoanPartner{
					PartnerName: args.req.PartnerName,
					LoanKind:    args.req.LoanKind,
					PartnerId:   args.req.PartnerId,
					Metadata:    args.req.Metadata,
				}).Return(mockCreateLoanPartnerOut, nil)
			},
		},
		{
			name: "error case - validation failed",
			args: args{
				ctx:         context.Background(),
				contentType: echo.MIMEApplicationJSON,
				req:         models.DoCreateAccountLoanPartnerRequest{},
			},
			expectation: expectation{
				wantRes:  `{"status":"error","message":"validation failed","errors":[{"code":"MISSING_FIELD","field":"partnerName","message":"field is missing"},{"code":"MISSING_FIELD","field":"partnerId","message":"field is missing"},{"code":"MISSING_FIELD","field":"loanKind","message":"field is missing"}]}`,
				wantCode: 422,
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
			name: "error case - data is exist",
			args: args{
				ctx:         context.Background(),
				contentType: echo.MIMEApplicationJSON,
				req: models.DoCreateAccountLoanPartnerRequest{
					PartnerName: "Chickin ABCDE",
					PartnerId:   "9876",
					LoanKind:    "PartnershipLoan",
				},
			},
			expectation: expectation{
				wantRes:  `{"status":"error","code":"DATA_IS_EXIST","message":"data is exist"}`,
				wantCode: 409,
			},
			doMock: func(args args, expectation expectation) {
				testHelper.mockAccountService.EXPECT().CreateLoanPartnerAccount(args.ctx, models.CreateAccountLoanPartner{
					PartnerName: args.req.PartnerName,
					LoanKind:    args.req.LoanKind,
					PartnerId:   args.req.PartnerId,
					Metadata:    args.req.Metadata,
				}).Return(models.AccountsLoanPartner{}, models.GetErrMap(models.ErrKeyDataIsExist))
			},
		},
		{
			name: "error case - database error",
			args: args{
				ctx:         context.Background(),
				contentType: echo.MIMEApplicationJSON,
				req: models.DoCreateAccountLoanPartnerRequest{
					PartnerName: "Chickin ABCDE",
					PartnerId:   "9876",
					LoanKind:    "PartnershipLoan",
				},
			},
			expectation: expectation{
				wantRes:  `{"status":"error","code":"INTERNAL_SERVER_ERROR","message":"unable to create data"}`,
				wantCode: 500,
			},
			doMock: func(args args, expectation expectation) {
				testHelper.mockAccountService.EXPECT().CreateLoanPartnerAccount(args.ctx, models.CreateAccountLoanPartner{
					PartnerName: args.req.PartnerName,
					LoanKind:    args.req.LoanKind,
					PartnerId:   args.req.PartnerId,
					Metadata:    args.req.Metadata,
				}).Return(models.AccountsLoanPartner{}, models.GetErrMap(models.ErrKeyUnableToCreateData))
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

			r := httptest.NewRequest(http.MethodPost, "/api/v1/accounts/loan-partner-accounts", &b)
			r.Header.Set(echo.HeaderContentType, tt.args.contentType)
			w := httptest.NewRecorder()

			testHelper.router.NewContext(r, w)
			testHelper.router.ServeHTTP(w, r)

			require.Equal(t, tt.expectation.wantCode, w.Code)
			require.Equal(t, tt.expectation.wantRes, strings.Trim(w.Body.String(), "\n"))
		})
	}
}
