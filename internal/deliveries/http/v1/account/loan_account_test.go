package account

import (
	"bitbucket.org/Amartha/go-accounting/internal/models"
	"context"
	"fmt"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func Test_Handler_getLoanAdvanceAccountByLoanAccount(t *testing.T) {
	testHelper := accountTestHelper(t)

	req := "21100100000001"

	type args struct {
		req string
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
			args: args{req},
			doMock: func(args args, expectation expectation) {
				testHelper.mockAccountService.EXPECT().GetLoanAdvanceAccountByLoanAccount(gomock.AssignableToTypeOf(context.Background()), args.req).Return(models.AccountLoan{
					LoanAccountNumber:               "21100100000001",
					LoanAdvancePaymentAccountNumber: "21100100000002",
				}, nil)
			},
			expectation: expectation{
				wantRes:  `{"kind":"account","loanAccountNumber":"21100100000001","loanAdvancePaymentAccountNumber":"21100100000002"}`,
				wantCode: 200,
			},
		},
		{
			name: "error case - account number not found",
			args: args{req},
			doMock: func(args args, expectation expectation) {
				testHelper.mockAccountService.EXPECT().GetLoanAdvanceAccountByLoanAccount(gomock.AssignableToTypeOf(context.Background()), args.req).Return(models.AccountLoan{}, models.GetErrMap(models.ErrKeyAccountNumberNotFound))
			},
			expectation: expectation{
				wantRes:  `{"status":"error","code":"DATA_NOT_FOUND","message":"account number not found"}`,
				wantCode: 404,
			},
		},
		{
			name: "error case - internal server error",
			args: args{req},
			doMock: func(args args, expectation expectation) {
				testHelper.mockAccountService.EXPECT().GetLoanAdvanceAccountByLoanAccount(gomock.AssignableToTypeOf(context.Background()), args.req).Return(models.AccountLoan{}, models.GetErrMap(models.ErrKeyDatabaseError))
			},
			expectation: expectation{
				wantRes:  `{"status":"error","code":"DATABASE_ERROR","message":"database error"}`,
				wantCode: 500,
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if tt.doMock != nil {
				tt.doMock(tt.args, tt.expectation)
			}

			url := fmt.Sprintf("%s/%s", "/api/v1/loan-accounts/advance-account", tt.args.req)
			r := httptest.NewRequest(http.MethodGet, url, nil)
			w := httptest.NewRecorder()

			testHelper.router.NewContext(r, w)
			testHelper.router.ServeHTTP(w, r)

			require.Equal(t, tt.expectation.wantCode, w.Code)
			require.Equal(t, tt.expectation.wantRes, strings.Trim(w.Body.String(), "\n"))
		})
	}
}
