package account

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"bitbucket.org/Amartha/go-accounting/internal/models"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func Test_Handler_getInvestedAccountByCIHAccount(t *testing.T) {
	testHelper := accountTestHelper(t)

	req := "211001000000417"

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
				testHelper.mockAccountService.EXPECT().GetLenderAccountByCIHAccountNumber(gomock.AssignableToTypeOf(context.Background()), args.req).Return(models.AccountLender{
					CIHAccountNumber:         "211001000000417",
					InvestedAccountNumber:    "212001000000327",
					ReceivablesAccountNumber: "142001000000005",
				}, nil)
			},
			expectation: expectation{
				wantRes:  `{"kind":"account","cihAccountNumber":"211001000000417","investedAccountNumber":"212001000000327","receivablesAccountNumber":"142001000000005"}`,
				wantCode: 200,
			},
		},
		{
			name: "error case - account number not found",
			args: args{req},
			doMock: func(args args, expectation expectation) {
				testHelper.mockAccountService.EXPECT().GetLenderAccountByCIHAccountNumber(gomock.AssignableToTypeOf(context.Background()), args.req).Return(models.AccountLender{}, models.GetErrMap(models.ErrKeyAccountNumberNotFound))
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
				testHelper.mockAccountService.EXPECT().GetLenderAccountByCIHAccountNumber(gomock.AssignableToTypeOf(context.Background()), args.req).Return(models.AccountLender{}, models.GetErrMap(models.ErrKeyDatabaseError))
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

			url := fmt.Sprintf("%s/%s", "/api/v1/lender-accounts", tt.args.req)
			r := httptest.NewRequest(http.MethodGet, url, nil)
			w := httptest.NewRecorder()

			testHelper.router.NewContext(r, w)
			testHelper.router.ServeHTTP(w, r)

			require.Equal(t, tt.expectation.wantCode, w.Code)
			require.Equal(t, tt.expectation.wantRes, strings.Trim(w.Body.String(), "\n"))
		})
	}
}
