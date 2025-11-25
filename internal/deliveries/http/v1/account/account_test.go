package account

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

	"bitbucket.org/Amartha/go-accounting/internal/models"
	"bitbucket.org/Amartha/go-accounting/internal/pkg/atime"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func Test_Handler_createAccount(t *testing.T) {
	testHelper := accountTestHelper(t)

	type args struct {
		ctx         context.Context
		contentType string
		req         models.DoCreateAccountRequest
	}
	mockCreateAccountOut := models.CreateAccount{
		AccountNumber:   "21100100000001",
		OwnerID:         "12345",
		CategoryCode:    "211",
		SubCategoryCode: "10000",
		EntityCode:      "001",
		Currency:        "IDR",
		Status:          models.AccountStatusActive,
		Name:            "Lender Yang Baik",
		AltId:           "534534534555353523523423423",
		AccountType:     "LENDER_INSTITUTIONAL",
		ProductTypeCode: "1001",
		LegacyId: &models.AccountLegacyId{
			"t24AccountNumber": "111000035909",
			"t24ArrangementId": "AA123",
		},
	}
	type expectation struct {
		wantRes  string
		wantCode int
	}
	tests := []struct {
		name        string
		urlCalled   string
		args        args
		expectation expectation
		doMock      func(args args, expectation expectation)
	}{
		{
			name:      "success",
			urlCalled: "/api/v1/accounts",
			args: args{
				ctx:         context.Background(),
				contentType: echo.MIMEApplicationJSON,
				req: models.DoCreateAccountRequest{
					OwnerID:         "12345",
					EntityCode:      "001",
					Currency:        models.CurrencyIDR,
					Name:            "Lender Yang Baik",
					AltId:           "534534534555353523523423423",
					AccountType:     "LENDER_INSTITUTIONAL",
					ProductTypeCode: "1001",
					CategoryCode:    "001",
					SubCategoryCode: "10001",
					LegacyId: &models.AccountLegacyId{
						"t24AccountNumber": "111000035909",
						"t24ArrangementId": "AA123",
					},
				},
			},
			expectation: expectation{
				wantRes:  `{"kind":"account","name":"Lender Yang Baik","accountNumber":"21100100000001","ownerId":"12345","productTypeCode":"1001","entityCode":"001","categoryCode":"211","subCategoryCode":"10000","currency":"IDR","status":"active","altId":"534534534555353523523423423","legacyId":{"t24AccountNumber":"111000035909","t24ArrangementId":"AA123"}}`,
				wantCode: 201,
			},
			doMock: func(args args, expectation expectation) {
				testHelper.mockAccountService.EXPECT().Create(args.ctx, models.CreateAccount{
					OwnerID:         args.req.OwnerID,
					EntityCode:      args.req.EntityCode,
					Currency:        args.req.Currency,
					Name:            args.req.Name,
					AltId:           args.req.AltId,
					CategoryCode:    "001",
					SubCategoryCode: "10001",
					AccountType:     "LENDER_INSTITUTIONAL",
					ProductTypeCode: "1001",
					LegacyId: &models.AccountLegacyId{
						"t24AccountNumber": "111000035909",
						"t24ArrangementId": "AA123",
					},
				}).Return(mockCreateAccountOut, nil)
			},
		},
		{
			name:      "error contentType",
			urlCalled: "/api/v1/accounts",
			args: args{
				ctx:         context.Background(),
				contentType: "",
				req: models.DoCreateAccountRequest{
					OwnerID:    "12345",
					EntityCode: "001",
					Currency:   "IDR",
					Name:       "Lender Yang Baik",
					AltId:      "534534534555353523523423423",
					LegacyId: &models.AccountLegacyId{
						"t24AccountNumber": "111000035909",
						"t24ArrangementId": "AA123",
					},
				},
			},
			expectation: expectation{
				wantRes:  `{"status":"error","code":400,"message":"code=415, message=Unsupported Media Type"}`,
				wantCode: 400,
			},
			doMock: func(args args, expectation expectation) {},
		},
		{
			name:      "error validating required",
			urlCalled: "/api/v1/accounts",
			args: args{
				ctx:         context.Background(),
				contentType: echo.MIMEApplicationJSON,
				req: models.DoCreateAccountRequest{
					ProductTypeCode: "1001",
				},
			},
			expectation: expectation{
				wantRes:  `{"status":"error","message":"validation failed","errors":[{"code":"MISSING_FIELD","field":"ownerId","message":"field is missing"},{"code":"MISSING_FIELD","field":"entityCode","message":"field is missing"},{"code":"MISSING_FIELD","field":"categoryCode","message":"field is missing"},{"code":"MISSING_FIELD","field":"subCategoryCode","message":"field is missing"},{"code":"MISSING_FIELD","field":"currency","message":"field is missing"}]}`,
				wantCode: 422,
			},
		},
		{
			name:      "error invalid account type",
			urlCalled: "/api/v1/accounts",
			args: args{
				ctx:         context.Background(),
				contentType: echo.MIMEApplicationJSON,
				req: models.DoCreateAccountRequest{
					OwnerID:         "12345",
					EntityCode:      "001",
					Currency:        models.CurrencyIDR,
					Name:            "Lender Yang Baik",
					AltId:           "534534534555353523523423423",
					AccountType:     "LENDER_INSTITUTIONAL",
					ProductTypeCode: "1001",
					CategoryCode:    "001",
					SubCategoryCode: "10001",
					LegacyId: &models.AccountLegacyId{
						"t24AccountNumber": "111000035909",
						"t24ArrangementId": "AA123",
					},
				},
			},
			expectation: expectation{
				wantRes:  `{"status":"error","code":"DATA_NOT_FOUND","message":"account type not valid"}`,
				wantCode: 404,
			},
			doMock: func(args args, expectation expectation) {
				testHelper.mockAccountService.EXPECT().Create(args.ctx, models.CreateAccount{
					OwnerID:         args.req.OwnerID,
					EntityCode:      args.req.EntityCode,
					Currency:        args.req.Currency,
					Name:            args.req.Name,
					AltId:           args.req.AltId,
					CategoryCode:    "001",
					SubCategoryCode: "10001",
					AccountType:     "LENDER_INSTITUTIONAL",
					ProductTypeCode: "1001",
					LegacyId: &models.AccountLegacyId{
						"t24AccountNumber": "111000035909",
						"t24ArrangementId": "AA123",
					},
				}).Return(models.CreateAccount{}, models.GetErrMap(models.ErrKeyAccountTypeNotValid))
			},
		},
		{
			name:      "error alt id exists",
			urlCalled: "/api/v1/accounts",
			args: args{
				ctx:         context.Background(),
				contentType: echo.MIMEApplicationJSON,
				req: models.DoCreateAccountRequest{
					OwnerID:         "12345",
					EntityCode:      "001",
					Currency:        models.CurrencyIDR,
					Name:            "Lender Yang Baik",
					AltId:           "534534534555353523523423423",
					AccountType:     "LENDER_INSTITUTIONAL",
					ProductTypeCode: "1001",
					CategoryCode:    "001",
					SubCategoryCode: "10001",
					LegacyId: &models.AccountLegacyId{
						"t24AccountNumber": "111000035909",
						"t24ArrangementId": "AA123",
					},
				},
			},
			expectation: expectation{
				wantRes:  `{"status":"error","code":"DATA_IS_EXIST","message":"alternate id is exist"}`,
				wantCode: 409,
			},
			doMock: func(args args, expectation expectation) {
				testHelper.mockAccountService.EXPECT().Create(args.ctx, models.CreateAccount{
					OwnerID:         args.req.OwnerID,
					EntityCode:      args.req.EntityCode,
					Currency:        args.req.Currency,
					Name:            args.req.Name,
					AltId:           args.req.AltId,
					CategoryCode:    "001",
					SubCategoryCode: "10001",
					AccountType:     "LENDER_INSTITUTIONAL",
					ProductTypeCode: "1001",
					LegacyId: &models.AccountLegacyId{
						"t24AccountNumber": "111000035909",
						"t24ArrangementId": "AA123",
					},
				}).Return(models.CreateAccount{}, models.GetErrMap(models.ErrKeyAltIdIsExist))
			},
		},
		{
			name:      "internal server error",
			urlCalled: "/api/v1/accounts",
			args: args{
				ctx:         context.Background(),
				contentType: echo.MIMEApplicationJSON,
				req: models.DoCreateAccountRequest{
					OwnerID:         "12345",
					EntityCode:      "001",
					Currency:        models.CurrencyIDR,
					Name:            "Lender Yang Baik",
					AltId:           "534534534555353523523423423",
					AccountType:     "LENDER_INSTITUTIONAL",
					ProductTypeCode: "1001",
					CategoryCode:    "001",
					SubCategoryCode: "10001",
					LegacyId: &models.AccountLegacyId{
						"t24AccountNumber": "111000035909",
						"t24ArrangementId": "AA123",
					},
				},
			},
			expectation: expectation{
				wantRes:  `{"status":"error","code":"DATABASE_ERROR","message":"database error"}`,
				wantCode: 500,
			},
			doMock: func(args args, expectation expectation) {
				testHelper.mockAccountService.EXPECT().Create(args.ctx, models.CreateAccount{
					OwnerID:         args.req.OwnerID,
					EntityCode:      args.req.EntityCode,
					Currency:        args.req.Currency,
					Name:            args.req.Name,
					AltId:           args.req.AltId,
					CategoryCode:    "001",
					SubCategoryCode: "10001",
					AccountType:     "LENDER_INSTITUTIONAL",
					ProductTypeCode: "1001",
					LegacyId: &models.AccountLegacyId{
						"t24AccountNumber": "111000035909",
						"t24ArrangementId": "AA123",
					},
				}).Return(models.CreateAccount{}, models.GetErrMap(models.ErrKeyDatabaseError))
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

			r := httptest.NewRequest(http.MethodPost, tt.urlCalled, &b)
			r.Header.Set(echo.HeaderContentType, tt.args.contentType)
			w := httptest.NewRecorder()

			testHelper.router.NewContext(r, w)
			testHelper.router.ServeHTTP(w, r)

			require.Equal(t, tt.expectation.wantCode, w.Code)
			require.Equal(t, tt.expectation.wantRes, strings.Trim(w.Body.String(), "\n"))
		})
	}
}

func Test_Handler_updateAccount(t *testing.T) {
	testHelper := accountTestHelper(t)
	mockUpdateAccountOut := models.UpdateAccount{
		Name:          "Lender Yang Baik",
		OwnerID:       "12345",
		AltID:         "534534534555353523523423423",
		AccountNumber: "22200100000001",
	}

	type (
		args struct {
			ctx         context.Context
			contentType string
			req         models.DoUpdateAccountRequest
		}
		expectation struct {
			wantRes  string
			wantCode int
		}
	)
	tests := []struct {
		name        string
		urlCalled   string
		args        args
		expectation expectation
		doMock      func(args args, expectation expectation)
	}{
		{
			name:      "test success",
			urlCalled: "/api/v1/accounts/22200100000001",
			args: args{ctx: context.Background(),
				contentType: echo.MIMEApplicationJSON,
				req: models.DoUpdateAccountRequest{
					Name:          "Lender Yang Baik",
					OwnerID:       "12345",
					AltID:         "534534534555353523523423423",
					AccountNumber: "22200100000001",
				},
			},
			expectation: expectation{
				wantRes:  `{"kind":"account","name":"Lender Yang Baik","ownerId":"12345","altId":"534534534555353523523423423"}`,
				wantCode: 200,
			},
			doMock: func(args args, expectation expectation) {
				testHelper.mockAccountService.EXPECT().Update(args.ctx, models.UpdateAccount{
					Name:          "Lender Yang Baik",
					OwnerID:       "12345",
					AltID:         "534534534555353523523423423",
					AccountNumber: "22200100000001",
				}).Return(mockUpdateAccountOut, nil)
			},
		},
		{
			name:      "test internal server error",
			urlCalled: "/api/v1/accounts/22200100000001",
			args: args{
				ctx:         context.Background(),
				contentType: echo.MIMEApplicationJSON,
				req: models.DoUpdateAccountRequest{
					Name:          "Lender Yang Baik",
					OwnerID:       "12345",
					AltID:         "534534534555353523523423423",
					AccountNumber: "22200100000001",
				}},
			expectation: expectation{
				wantRes:  `{"status":"error","code":500,"message":"internal server error"}`,
				wantCode: 500,
			},
			doMock: func(args args, expectation expectation) {
				testHelper.mockAccountService.EXPECT().Update(args.ctx, models.UpdateAccount{
					Name:          "Lender Yang Baik",
					OwnerID:       "12345",
					AltID:         "534534534555353523523423423",
					AccountNumber: "22200100000001",
				}).Return(mockUpdateAccountOut, models.ErrInternalServerError)
			},
		},
		{
			name:      "test error account not found",
			urlCalled: "/api/v1/accounts/222001",
			args: args{
				ctx:         context.Background(),
				contentType: echo.MIMEApplicationJSON,
				req: models.DoUpdateAccountRequest{
					Name:          "Lender Yang Baik",
					OwnerID:       "12345",
					AltID:         "534534534555353523523423423",
					AccountNumber: "222001",
				},
			},
			expectation: expectation{
				wantRes:  `{"status":"error","code":"DATA_NOT_FOUND","message":"account number not found"}`,
				wantCode: 404,
			},
			doMock: func(args args, expectation expectation) {
				testHelper.mockAccountService.EXPECT().Update(args.ctx, models.UpdateAccount{
					Name:          "Lender Yang Baik",
					OwnerID:       "12345",
					AltID:         "534534534555353523523423423",
					AccountNumber: "222001",
				}).Return(mockUpdateAccountOut, models.GetErrMap(models.ErrKeyAccountNumberNotFound))
			},
		},
		{
			name:      "test error validation",
			urlCalled: "/api/v1/accounts/22200100000001",
			args: args{
				ctx:         context.Background(),
				contentType: echo.MIMEApplicationJSON,
				req: models.DoUpdateAccountRequest{
					AccountNumber: "22200100000001",
					Name:          "Lender Yang Baik",
					OwnerID:       "12345543211234554321",
					AltID:         "",
				},
			},
			expectation: expectation{
				wantRes:  `{"status":"error","message":"validation failed","errors":[{"code":"INVALID_LENGTH","field":"ownerId","message":"field can have a maximum length of 15 characters"}]}`,
				wantCode: 422,
			},
			doMock: func(args args, expectation expectation) {},
		},
		{
			name:      "test error contentType",
			urlCalled: "/api/v1/accounts/22200100000001",
			args: args{
				ctx:         context.Background(),
				contentType: "",
				req: models.DoUpdateAccountRequest{
					Name:          "Lender Yang Baik",
					OwnerID:       "12345",
					AltID:         "534534534555353523523423423",
					AccountNumber: "22200100000001",
				},
			},
			expectation: expectation{
				wantRes:  `{"status":"error","code":400,"message":"code=415, message=Unsupported Media Type"}`,
				wantCode: 400,
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

			r := httptest.NewRequest(http.MethodPatch, tt.urlCalled, &b)
			r.Header.Set(echo.HeaderContentType, tt.args.contentType)
			w := httptest.NewRecorder()

			testHelper.router.NewContext(r, w)
			testHelper.router.ServeHTTP(w, r)

			require.Equal(t, tt.expectation.wantCode, w.Code)
			require.Equal(t, tt.expectation.wantRes, strings.Trim(w.Body.String(), "\n"))
		})
	}
}

func Test_Handler_updateAccountEntity(t *testing.T) {
	testHelper := accountTestHelper(t)
	mockReq := models.DoUpdateAccountEntityRequest{
		AccountNumber: "22200100000001",
		EntityCode:    "001",
	}
	mockUpdateAccountEntity := models.UpdateAccountEntity{
		AccountNumber: "22200100000001",
		EntityCode:    "001",
	}

	type (
		args struct {
			ctx         context.Context
			contentType string
			req         models.DoUpdateAccountEntityRequest
		}
		expectation struct {
			wantRes  string
			wantCode int
		}
	)
	tests := []struct {
		name        string
		args        args
		expectation expectation
		doMock      func(args args, expectation expectation)
	}{
		{
			name: "success update",
			args: args{ctx: context.Background(),
				contentType: echo.MIMEApplicationJSON,
				req:         mockReq,
			},
			expectation: expectation{
				wantRes:  `{"kind":"account","entityCode":"001"}`,
				wantCode: 200,
			},
			doMock: func(args args, expectation expectation) {
				testHelper.mockAccountService.EXPECT().UpdateAccountEntity(args.ctx, mockUpdateAccountEntity).Return(mockUpdateAccountEntity, nil)
			},
		},
		{
			name: "error case - account number not found",
			args: args{
				ctx:         context.Background(),
				contentType: echo.MIMEApplicationJSON,
				req:         mockReq,
			},
			expectation: expectation{
				wantRes:  `{"status":"error","code":"DATA_NOT_FOUND","message":"account number not found"}`,
				wantCode: 404,
			},
			doMock: func(args args, expectation expectation) {
				testHelper.mockAccountService.EXPECT().UpdateAccountEntity(args.ctx, mockUpdateAccountEntity).Return(mockUpdateAccountEntity, models.GetErrMap(models.ErrKeyAccountNumberNotFound))
			},
		},
		{
			name: "error case - entity code not found",
			args: args{
				ctx:         context.Background(),
				contentType: echo.MIMEApplicationJSON,
				req:         mockReq,
			},
			expectation: expectation{
				wantRes:  `{"status":"error","code":"DATA_NOT_FOUND","message":"entity code not found"}`,
				wantCode: 404,
			},
			doMock: func(args args, expectation expectation) {
				testHelper.mockAccountService.EXPECT().UpdateAccountEntity(args.ctx, mockUpdateAccountEntity).Return(mockUpdateAccountEntity, models.GetErrMap(models.ErrKeyEntityCodeNotFound))
			},
		},
		{
			name: "error case - entity code not found",
			args: args{
				ctx:         context.Background(),
				contentType: echo.MIMEApplicationJSON,
				req:         mockReq,
			},
			expectation: expectation{
				wantRes:  `{"status":"error","code":"DATA_NOT_FOUND","message":"entity code not found"}`,
				wantCode: 404,
			},
			doMock: func(args args, expectation expectation) {
				testHelper.mockAccountService.EXPECT().UpdateAccountEntity(args.ctx, mockUpdateAccountEntity).Return(mockUpdateAccountEntity, models.GetErrMap(models.ErrKeyEntityCodeNotFound))
			},
		},
		{
			name: "error case - cannot update entity",
			args: args{
				ctx:         context.Background(),
				contentType: echo.MIMEApplicationJSON,
				req:         mockReq,
			},
			expectation: expectation{
				wantRes:  `{"status":"error","code":"DATA_IS_EXIST","message":"unable to change the entity because the account has a transactions"}`,
				wantCode: 409,
			},
			doMock: func(args args, expectation expectation) {
				testHelper.mockAccountService.EXPECT().UpdateAccountEntity(args.ctx, mockUpdateAccountEntity).Return(mockUpdateAccountEntity, models.GetErrMap(models.ErrKeyJournalAccountIsExist))
			},
		},
		{
			name: "error case - database error",
			args: args{
				ctx:         context.Background(),
				contentType: echo.MIMEApplicationJSON,
				req:         mockReq,
			},
			expectation: expectation{
				wantRes:  `{"status":"error","code":"DATABASE_ERROR","message":"database error"}`,
				wantCode: 500,
			},
			doMock: func(args args, expectation expectation) {
				testHelper.mockAccountService.EXPECT().UpdateAccountEntity(args.ctx, mockUpdateAccountEntity).Return(mockUpdateAccountEntity, models.GetErrMap(models.ErrKeyDatabaseError))
			},
		},
		{
			name: "error case - validation failed",
			args: args{
				ctx:         context.Background(),
				contentType: echo.MIMEApplicationJSON,
				req:         models.DoUpdateAccountEntityRequest{},
			},
			expectation: expectation{
				wantRes:  `{"status":"error","message":"validation failed","errors":[{"code":"MISSING_FIELD","field":"accountNumber","message":"field is missing"},{"code":"MISSING_FIELD","field":"entityCode","message":"field is missing"}]}`,
				wantCode: 422,
			},
		},
		{
			name: "test error contentType",
			args: args{
				ctx:         context.Background(),
				contentType: "",
				req:         models.DoUpdateAccountEntityRequest{},
			},
			expectation: expectation{
				wantRes:  `{"status":"error","code":400,"message":"code=415, message=Unsupported Media Type"}`,
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

			r := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/v1/accounts/%s/entity", tt.args.req.AccountNumber), &b)
			r.Header.Set(echo.HeaderContentType, tt.args.contentType)
			w := httptest.NewRecorder()

			testHelper.router.NewContext(r, w)
			testHelper.router.ServeHTTP(w, r)

			require.Equal(t, tt.expectation.wantCode, w.Code)
			require.Equal(t, tt.expectation.wantRes, strings.Trim(w.Body.String(), "\n"))
		})
	}
}

func Test_Handler_getByAccountNumber(t *testing.T) {
	testHelper := accountTestHelper(t)

	type args struct {
		ctx context.Context
		req string
	}

	type expectation struct {
		wantRes  string
		wantCode int
	}
	tests := []struct {
		name        string
		urlCalled   string
		args        args
		expectation expectation
		doMock      func(args args, expectation expectation)
	}{
		{
			name:      "test success",
			urlCalled: "/api/v1/accounts/22200100000001",
			args: args{ctx: context.Background(),
				req: "22200100000001",
			},
			expectation: expectation{
				wantRes:  `{"kind":"account","accountNumber":"21100100000070","accountName":"Lender Balance - Cashout Payable 1","coaTypeCode":"LIA","coaTypeName":"Liability","categoryCode":"211","categoryName":"Marketplace Payable (Lender Balance)","subCategoryCode":"21105","subCategoryName":"Lender Balance - Cashout Payable","altID":"","entityCode":"001","entityName":"AMF","productTypeCode":"1001","productTypeName":"Group Loan","ownerID":"1234567","status":"active","currency":"IDR","accountType":"LENDER_CASHOUT_PAYABLE","createdAt":"0001-01-01 07:07:12","updatedAt":"0001-01-01 07:07:12"}`,
				wantCode: 200,
			},
			doMock: func(args args, expectation expectation) {
				testHelper.mockAccountService.EXPECT().GetOneByAccountNumber(args.ctx, args.req).Return(models.GetAccountOut{
					AccountNumber:   "21100100000070",
					AccountName:     "Lender Balance - Cashout Payable 1",
					CoaTypeCode:     "LIA",
					CoaTypeName:     "Liability",
					CategoryCode:    "211",
					CategoryName:    "Marketplace Payable (Lender Balance)",
					SubCategoryCode: "21105",
					SubCategoryName: "Lender Balance - Cashout Payable",
					AltID:           "",
					EntityCode:      "001",
					EntityName:      "AMF",
					ProductTypeCode: "1001",
					ProductTypeName: "Group Loan",
					OwnerID:         "1234567",
					Status:          "active",
					Currency:        "IDR",
					AccountType:     "LENDER_CASHOUT_PAYABLE",
				}, nil)
			},
		},
		{
			name:      "test error",
			urlCalled: "/api/v1/accounts/22200100000001",
			args: args{
				ctx: context.Background(),
				req: "22200100000001"},
			expectation: expectation{
				wantRes:  `{"status":"error","code":500,"message":"internal server error"}`,
				wantCode: 500,
			},
			doMock: func(args args, expectation expectation) {
				testHelper.mockAccountService.EXPECT().GetOneByAccountNumber(args.ctx, args.req).Return(models.GetAccountOut{}, models.ErrInternalServerError)
			},
		},
		{
			name:      "test not found",
			urlCalled: "/api/v1/accounts/22200100000001",
			args: args{
				ctx: context.Background(),
				req: "22200100000001",
			},
			expectation: expectation{
				wantRes:  `{"status":"error","code":"DATA_NOT_FOUND","message":"account number not found"}`,
				wantCode: 404,
			},
			doMock: func(args args, expectation expectation) {
				testHelper.mockAccountService.EXPECT().GetOneByAccountNumber(args.ctx, args.req).Return(models.GetAccountOut{}, models.GetErrMap(models.ErrKeyAccountNumberNotFound))
			}},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if tt.doMock != nil {
				tt.doMock(tt.args, tt.expectation)
			}
			r := httptest.NewRequest(http.MethodGet, tt.urlCalled, nil)
			w := httptest.NewRecorder()

			testHelper.router.NewContext(r, w)
			testHelper.router.ServeHTTP(w, r)

			require.Equal(t, tt.expectation.wantCode, w.Code)
			require.Equal(t, tt.expectation.wantRes, strings.Trim(w.Body.String(), "\n"))
		})
	}
}

func Test_Handler_getByLegactId(t *testing.T) {
	testHelper := accountTestHelper(t)

	type args struct {
		ctx context.Context
		req string
	}
	type expectation struct {
		wantRes  string
		wantCode int
	}
	tests := []struct {
		name        string
		urlCalled   string
		args        args
		expectation expectation
		doMock      func(args args, expectation expectation)
	}{
		{
			name:      "success case",
			urlCalled: "/api/v1/accounts/t24/22200100000001",
			args: args{ctx: context.Background(),
				req: "22200100000001",
			},
			expectation: expectation{
				wantRes:  `{"kind":"account","accountNumber":"21100100000070","accountName":"Lender Balance - Cashout Payable 1","coaTypeCode":"LIA","coaTypeName":"Liability","categoryCode":"211","categoryName":"Marketplace Payable (Lender Balance)","subCategoryCode":"21105","subCategoryName":"Lender Balance - Cashout Payable","altID":"","entityCode":"001","entityName":"AMF","productTypeCode":"1001","productTypeName":"Group Loan","ownerID":"1234567","status":"active","currency":"IDR","accountType":"LENDER_CASHOUT_PAYABLE","createdAt":"0001-01-01 07:07:12","updatedAt":"0001-01-01 07:07:12"}`,
				wantCode: 200,
			},
			doMock: func(args args, expectation expectation) {
				testHelper.mockAccountService.EXPECT().GetOneByLegacyID(args.ctx, args.req).Return(models.GetAccountOut{
					AccountNumber:   "21100100000070",
					AccountName:     "Lender Balance - Cashout Payable 1",
					CoaTypeCode:     "LIA",
					CoaTypeName:     "Liability",
					CategoryCode:    "211",
					CategoryName:    "Marketplace Payable (Lender Balance)",
					SubCategoryCode: "21105",
					SubCategoryName: "Lender Balance - Cashout Payable",
					AltID:           "",
					EntityCode:      "001",
					EntityName:      "AMF",
					ProductTypeCode: "1001",
					ProductTypeName: "Group Loan",
					OwnerID:         "1234567",
					Status:          "active",
					Currency:        "IDR",
					AccountType:     "LENDER_CASHOUT_PAYABLE",
				}, nil)
			},
		},
		{
			name:      "error case - internal server error",
			urlCalled: "/api/v1/accounts/t24/22200100000001",
			args: args{
				ctx: context.Background(),
				req: "22200100000001",
			},
			expectation: expectation{
				wantRes:  `{"status":"error","code":500,"message":"internal server error"}`,
				wantCode: 500,
			},
			doMock: func(args args, expectation expectation) {
				testHelper.mockAccountService.EXPECT().GetOneByLegacyID(args.ctx, args.req).Return(models.GetAccountOut{}, models.ErrInternalServerError)
			},
		},
		{
			name:      "error case - data not found",
			urlCalled: "/api/v1/accounts/t24/22200100000001",
			args: args{
				ctx: context.Background(),
				req: "22200100000001",
			},
			expectation: expectation{
				wantRes:  `{"status":"error","code":"DATA_NOT_FOUND","message":"legacy id not found"}`,
				wantCode: 404,
			},
			doMock: func(args args, expectation expectation) {
				testHelper.mockAccountService.EXPECT().GetOneByLegacyID(args.ctx, args.req).Return(models.GetAccountOut{}, models.GetErrMap(models.ErrKeyLegacyIdNotFound))
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if tt.doMock != nil {
				tt.doMock(tt.args, tt.expectation)
			}
			r := httptest.NewRequest(http.MethodGet, tt.urlCalled, nil)
			w := httptest.NewRecorder()

			testHelper.router.NewContext(r, w)
			testHelper.router.ServeHTTP(w, r)

			require.Equal(t, tt.expectation.wantCode, w.Code)
			require.Equal(t, tt.expectation.wantRes, strings.Trim(w.Body.String(), "\n"))
		})
	}
}

func Test_Handler_getAccountList(t *testing.T) {
	testHelper := accountTestHelper(t)

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
			name: "success case - get account list",
			args: args{
				ctx:         context.Background(),
				contentType: echo.MIMEApplicationJSON,
				req: url.Values{
					"coaTypeCode":     []string{"AST"},
					"entityCode":      []string{"001"},
					"categoryCode":    []string{"112"},
					"subCategoryCode": []string{"11201"},
					"productTypeCode": []string{"1001"},
				},
			},
			expectation: expectation{
				wantRes:  `{"kind":"collection","contents":[{"kind":"account","accountNumber":"11200100000001","accountName":"Kas BP Love You","coaTypeCode":"AST","coaTypeName":"Liability","categoryCode":"112","categoryName":"KAS BP","subCategoryCode":"11201","subCategoryName":"Kas BP","entityCode":"001","entityName":"AMF","productTypeCode":"1001","productTypeName":"Group Loan","altId":"5345345345553535235234234232","ownerId":"123456789","status":"active","createdAt":"0001-01-01 07:07:12","updatedAt":"0001-01-01 07:07:12","t24AccountNumber":""}],"pagination":{"prev":"","next":"","totalEntries":1}}`,
				wantCode: 200,
			},
			doMock: func(args args, expectation expectation) {
				testHelper.mockAccountService.EXPECT().GetAccountList(gomock.Any(), gomock.Any()).Return([]models.GetAccountOut{
					{
						AccountNumber:   "11200100000001",
						AccountName:     "Kas BP Love You",
						CoaTypeCode:     "AST",
						CoaTypeName:     "Liability",
						CategoryCode:    "112",
						CategoryName:    "KAS BP",
						SubCategoryCode: "11201",
						SubCategoryName: "Kas BP",
						EntityCode:      "001",
						EntityName:      "AMF",
						ProductTypeCode: "1001",
						ProductTypeName: "Group Loan",
						AltID:           "5345345345553535235234234232",
						OwnerID:         "123456789",
						Status:          "active",
						// CreatedAt:       atime.Now(),
						// UpdatedAt:       atime.Now(),
					},
				}, 1, nil)
			},
		},
		{
			name: "error case - get account list",
			args: args{
				ctx:         context.Background(),
				contentType: echo.MIMEApplicationJSON,
				req: url.Values{
					"coaTypeCode":     []string{"AST"},
					"entityCode":      []string{"001"},
					"categoryCode":    []string{"112"},
					"subCategoryCode": []string{"11201"},
					"productTypeCode": []string{"1001"},
				},
			},
			expectation: expectation{
				wantRes:  `{"status":"error","code":500,"message":"assert.AnError general error for testing"}`,
				wantCode: 500,
			},
			doMock: func(args args, expectation expectation) {
				testHelper.mockAccountService.EXPECT().GetAccountList(gomock.Any(), gomock.Any()).Return([]models.GetAccountOut{}, 1, assert.AnError)
			},
		},
		{
			name: "error case - invalid limit get account list",
			args: args{
				ctx:         context.Background(),
				contentType: echo.MIMEApplicationJSON,
				req: url.Values{
					"coaTypeCode":     []string{"AST"},
					"entityCode":      []string{"001"},
					"categoryCode":    []string{"112"},
					"subCategoryCode": []string{"11201"},
					"limit":           []string{"-1"},
					"productTypeCode": []string{"1001"},
				},
			},
			expectation: expectation{
				wantRes:  `{"status":"error","code":"INVALID_VALUES","message":"the limit must be greater than zero"}`,
				wantCode: 400,
			},
			doMock: func(args args, expectation expectation) {},
		},
		{
			name: "error case - validation error get account list",
			args: args{
				ctx:         context.Background(),
				contentType: echo.MIMEApplicationJSON,
				req:         url.Values{},
			},
			expectation: expectation{
				wantCode: 422,
				wantRes:  `{"status":"error","message":"validation failed","errors":[{"code":"MISSING_FIELD","field":"entityCode","message":"field is missing"}]}`,
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

			r := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/accounts?%s", tt.args.req.Encode()), nil)
			r.Header.Set(echo.HeaderContentType, tt.args.contentType)
			w := httptest.NewRecorder()

			testHelper.router.NewContext(r, w)
			testHelper.router.ServeHTTP(w, r)

			require.Equal(t, tt.expectation.wantCode, w.Code)
			require.Equal(t, tt.expectation.wantRes, strings.Trim(w.Body.String(), "\n"))
		})
	}
}

func Test_Handler_downloadCSVGetAccountList(t *testing.T) {
	testHelper := accountTestHelper(t)
	time := atime.Now()
	req := url.Values{
		"coaTypeCode":     []string{"AST"},
		"entityCode":      []string{"001"},
		"categoryCode":    []string{"112"},
		"subCategoryCode": []string{"11201"},
		"productTypeCode": []string{"10001"},
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
			name: "success download account list",
			args: args{
				ctx:         context.Background(),
				contentType: echo.MIMEApplicationJSON,
				req:         req,
			},
			expectation: expectation{
				wantCode: 200,
			},
			doMock: func(args args, expectation expectation) {
				testHelper.mockAccountService.EXPECT().GetAccountList(gomock.Any(), gomock.Any()).Return([]models.GetAccountOut{
					{
						AccountNumber:   "11200100000001",
						AccountName:     "Kas BP Love You",
						CoaTypeCode:     "AST",
						CoaTypeName:     "Liability",
						CategoryCode:    "112",
						CategoryName:    "KAS BP",
						SubCategoryCode: "11201",
						SubCategoryName: "Kas BP",
						EntityCode:      "001",
						EntityName:      "AMF",
						AltID:           "5345345345553535235234234232",
						OwnerID:         "123456789",
						Status:          "active",
						CreatedAt:       time,
						UpdatedAt:       time,
					},
				}, 1, nil)
				testHelper.mockAccountService.EXPECT().DownloadCSVGetAccountList(gomock.Any(), []models.GetAccountOut{
					{
						AccountNumber:   "11200100000001",
						AccountName:     "Kas BP Love You",
						CoaTypeCode:     "AST",
						CoaTypeName:     "Liability",
						CategoryCode:    "112",
						CategoryName:    "KAS BP",
						SubCategoryCode: "11201",
						SubCategoryName: "Kas BP",
						EntityCode:      "001",
						EntityName:      "AMF",
						AltID:           "5345345345553535235234234232",
						OwnerID:         "123456789",
						Status:          "active",
						CreatedAt:       time,
						UpdatedAt:       time,
					},
				}).Return(&bytes.Buffer{}, "", nil)
			},
		},
		{
			name: "error limit get account list",
			args: args{
				ctx:         context.Background(),
				contentType: echo.MIMEApplicationJSON,
				req: url.Values{
					"coaTypeCode":     []string{"AST"},
					"entityCode":      []string{"001"},
					"categoryCode":    []string{"112"},
					"subCategoryCode": []string{"11201"},
					"productTypeCode": []string{"10001"},
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
			name: "error get account list",
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
				testHelper.mockAccountService.EXPECT().GetAccountList(gomock.Any(), gomock.Any()).Return([]models.GetAccountOut{}, 1, assert.AnError)
			},
		},
		{
			name: "error validation download account list",
			args: args{
				ctx:         context.Background(),
				contentType: echo.MIMEApplicationJSON,
				req:         url.Values{},
			},
			expectation: expectation{
				wantCode: 422,
				wantRes:  `{"status":"error","message":"validation failed","errors":[{"code":"MISSING_FIELD","field":"coaTypeCode","message":"field is missing"},{"code":"MISSING_FIELD","field":"entityCode","message":"field is missing"},{"code":"MISSING_FIELD","field":"categoryCode","message":"field is missing"},{"code":"MISSING_FIELD","field":"subCategoryCode","message":"field is missing"}]}`,
			},
			doMock: func(args args, expectation expectation) {},
		},
		{
			name: "error download account list",
			args: args{
				ctx:         context.Background(),
				contentType: echo.MIMEApplicationJSON,
				req:         req,
			},
			expectation: expectation{
				wantCode: 500,
				wantRes:  `{"status":"error","code":500,"message":"assert.AnError general error for testing"}`,
			},
			doMock: func(args args, expectation expectation) {
				testHelper.mockAccountService.EXPECT().GetAccountList(gomock.Any(), gomock.Any()).Return([]models.GetAccountOut{}, 1, nil)
				testHelper.mockAccountService.EXPECT().DownloadCSVGetAccountList(gomock.Any(), []models.GetAccountOut{}).Return(&bytes.Buffer{}, "", assert.AnError)
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

			r := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/accounts/download?%s", tt.args.req.Encode()), nil)
			r.Header.Set(echo.HeaderContentType, tt.args.contentType)
			w := httptest.NewRecorder()

			testHelper.router.NewContext(r, w)
			testHelper.router.ServeHTTP(w, r)

			require.Equal(t, tt.expectation.wantCode, w.Code)
			require.Equal(t, tt.expectation.wantRes, strings.Trim(w.Body.String(), "\n"))
		})
	}
}

func Test_Handler_checkAltIdIsExist(t *testing.T) {
	testHelper := accountTestHelper(t)
	type args struct {
		ctx context.Context
		req string
	}
	type expectation struct {
		wantRes  string
		wantCode int
	}
	tests := []struct {
		name        string
		urlCalled   string
		args        args
		expectation expectation
		doMock      func(args args, expectation expectation)
	}{
		{
			name:      "test success",
			urlCalled: "/api/v1/accounts/alt-ids?altId=22200100000001",
			args: args{ctx: context.Background(),
				req: "22200100000001",
			},
			expectation: expectation{
				wantRes:  `{"kind":"account","altId":"22200100000001","isExist":false}`,
				wantCode: 200,
			},
			doMock: func(args args, expectation expectation) {
				testHelper.mockAccountService.EXPECT().CheckAltIdIsExist(args.ctx, args.req).Return(nil)
			},
		},
		{
			name:      "test error altId required",
			urlCalled: "/api/v1/accounts/alt-ids?altId=",
			args: args{ctx: context.Background(),
				req: "22200100000001",
			},
			expectation: expectation{
				wantRes:  `{"status":"error","message":"validation failed","errors":[{"code":"MISSING_FIELD","field":"altId","message":"field is missing"}]}`,
				wantCode: 422,
			},
			doMock: func(args args, expectation expectation) {},
		},
		{
			name:      "test error data is exist",
			urlCalled: "/api/v1/accounts/alt-ids?altId=22200100000001",
			args: args{ctx: context.Background(),
				req: "22200100000001",
			},
			expectation: expectation{
				wantRes:  `{"status":"error","code":"DATA_IS_EXIST","message":"alternate id is exist"}`,
				wantCode: 409,
			},
			doMock: func(args args, expectation expectation) {
				testHelper.mockAccountService.EXPECT().CheckAltIdIsExist(args.ctx, args.req).Return(models.GetErrMap(models.ErrKeyAltIdIsExist))
			},
		},
		{
			name:      "test internal server error",
			urlCalled: "/api/v1/accounts/alt-ids?altId=22200100000001",
			args: args{ctx: context.Background(),
				req: "22200100000001",
			},
			expectation: expectation{
				wantRes:  `{"status":"error","code":500,"message":"assert.AnError general error for testing"}`,
				wantCode: 500,
			},
			doMock: func(args args, expectation expectation) {
				testHelper.mockAccountService.EXPECT().CheckAltIdIsExist(args.ctx, args.req).Return(assert.AnError)
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if tt.doMock != nil {
				tt.doMock(tt.args, tt.expectation)
			}
			r := httptest.NewRequest(http.MethodGet, tt.urlCalled, nil)
			w := httptest.NewRecorder()

			testHelper.router.NewContext(r, w)
			testHelper.router.ServeHTTP(w, r)

			require.Equal(t, tt.expectation.wantCode, w.Code)
			require.Equal(t, tt.expectation.wantRes, strings.Trim(w.Body.String(), "\n"))
		})
	}
}

func Test_Handler_uploadAccount(t *testing.T) {
	testHelper := accountTestHelper(t)

	type args struct {
		ctx         context.Context
		contentType string
		fileName    string
	}
	type expectation struct {
		wantRes  string
		wantCode int
	}
	tests := []struct {
		name        string
		args        args
		expectation expectation
		doMock      func(args args)
	}{
		{
			name: "happy path",
			expectation: expectation{
				wantRes:  `"success"`,
				wantCode: 200,
			},
			args: args{
				ctx:      context.TODO(),
				fileName: "../../../../../storages/upload_accounts.csv",
			},
			doMock: func(args args) {
				testHelper.mockAccountService.EXPECT().ProcessUploadAccounts(gomock.Any(), gomock.Any()).Return(nil)
			},
		},
		{
			name: "internal server error",
			expectation: expectation{
				wantRes:  `{"status":"error","code":500,"message":"assert.AnError general error for testing"}`,
				wantCode: 500,
			},
			args: args{
				ctx:      context.TODO(),
				fileName: "../../../../../storages/upload_accounts.csv",
			},
			doMock: func(args args) {
				testHelper.mockAccountService.EXPECT().ProcessUploadAccounts(gomock.Any(), gomock.Any()).Return(assert.AnError)
			},
		},
		{
			name: "error contentType not allowed",
			expectation: expectation{
				wantRes:  `{"status":"error","code":400,"message":"request Content-Type isn't multipart/form-data"}`,
				wantCode: 400,
			},
			args: args{
				ctx:         context.TODO(),
				contentType: echo.MIMEApplicationJSON,
				fileName:    "../../../../../storages/upload_accounts.csv",
			},
			doMock: func(args args) {
				testHelper.mockAccountService.EXPECT().ProcessUploadAccounts(gomock.Any(), gomock.Any()).Return(nil)
			},
		},
		{
			name: "error file not csv",
			expectation: expectation{
				wantRes:  `{"status":"error","code":400,"message":"file not csv"}`,
				wantCode: 400,
			},
			args: args{
				ctx:      context.TODO(),
				fileName: "account.go",
			},
			doMock: func(args args) {
				testHelper.mockAccountService.EXPECT().ProcessUploadAccounts(gomock.Any(), gomock.Any()).Return(nil)
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if tt.doMock != nil {
				tt.doMock(tt.args)
			}

			body := new(bytes.Buffer)
			writer := multipart.NewWriter(body)
			dataPart, err := writer.CreateFormFile("file", tt.args.fileName)
			require.NoError(t, err)

			// copy file content into multipart section dataPart
			f, err := os.Open(tt.args.fileName)
			require.NoError(t, err)
			_, err = io.Copy(dataPart, f)
			require.NoError(t, err)
			require.NoError(t, writer.Close())

			r := httptest.NewRequest(http.MethodPost, "/api/v1/accounts/upload", body)
			if tt.args.contentType == "" {
				r.Header.Set(echo.HeaderContentType, writer.FormDataContentType())
			}
			w := httptest.NewRecorder()

			testHelper.router.NewContext(r, w)
			testHelper.router.ServeHTTP(w, r)

			require.Equal(t, tt.expectation.wantCode, w.Code)
			require.Equal(t, tt.expectation.wantRes, strings.Trim(w.Body.String(), "\n"))
		})
	}
}

func Test_Handler_getAllAccountNumbersByParam(t *testing.T) {
	testHelper := accountTestHelper(t)
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
			name: "success case - get all account numbers by param ownerId",
			args: args{
				ctx:         context.Background(),
				contentType: echo.MIMEApplicationJSON,
				req: url.Values{
					"ownerId": []string{"12345678910"},
				},
			},
			expectation: expectation{
				wantRes:  `{"kind":"collection","contents":[{"kind":"account","ownerId":"","accountNumber":"11200100000001","altId":"","name":"","accountType":"LENDER_INSTITUSI_NON_RDL","entityCode":"","productTypeCode":"","categoryCode":"","subCategoryCode":"","currency":"","status":"","legacyId":null,"metadata":null}],"total_rows":1}`,
				wantCode: 200,
			},
			doMock: func(args args, expectation expectation) {
				testHelper.mockAccountService.EXPECT().GetAllAccountNumbersByParam(gomock.Any(), gomock.Any()).
					Return([]models.GetAllAccountNumbersByParamOut{{
						AccountNumber: "11200100000001",
						AccountType:   "LENDER_INSTITUSI_NON_RDL",
					}}, nil)
			},
		},
		{
			name: "error case - request validation error",
			args: args{
				ctx:         context.Background(),
				contentType: echo.MIMEApplicationJSON,
				req:         url.Values{},
			},
			expectation: expectation{
				wantRes:  `{"status":"error","message":"validation failed","errors":[{"code":"MISSING_FIELD","field":"ownerId","message":"required fields at least ownerId"},{"code":"MISSING_FIELD","field":"altId","message":"required fields at least altId"},{"code":"MISSING_FIELD","field":"accountNumbers","message":"required fields at least accountNumbers"},{"code":"MISSING_FIELD","field":"accountType","message":"required fields at least accountType"}]}`,
				wantCode: 422,
			},
		},
		{
			name: "error case - internal server error",
			args: args{
				ctx:         context.Background(),
				contentType: echo.MIMEApplicationJSON,
				req: url.Values{
					"ownerId": []string{"JanganDipaksain"},
				},
			},
			expectation: expectation{
				wantRes:  `{"status":"error","code":500,"message":"assert.AnError general error for testing"}`,
				wantCode: 500,
			},
			doMock: func(args args, expectation expectation) {
				testHelper.mockAccountService.EXPECT().
					GetAllAccountNumbersByParam(gomock.Any(), gomock.Any()).
					Return([]models.GetAllAccountNumbersByParamOut{}, assert.AnError)
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

			r := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/accounts/account-numbers?%s", tt.args.req.Encode()), nil)
			r.Header.Set(echo.HeaderContentType, tt.args.contentType)
			w := httptest.NewRecorder()

			testHelper.router.NewContext(r, w)
			testHelper.router.ServeHTTP(w, r)

			require.Equal(t, tt.expectation.wantCode, w.Code)
			require.Equal(t, tt.expectation.wantRes, strings.Trim(w.Body.String(), "\n"))
		})
	}
}
