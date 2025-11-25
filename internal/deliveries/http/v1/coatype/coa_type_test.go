package coatype

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"bitbucket.org/Amartha/go-accounting/internal/models"
	"bitbucket.org/Amartha/go-accounting/internal/services/mock"

	xlog "bitbucket.org/Amartha/go-x/log"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestMain(m *testing.M) {
	xlog.InitForTest()
	os.Exit(m.Run())
}

type testCOATypeHelper struct {
	router      *echo.Echo
	mockCtrl    *gomock.Controller
	mockService *mock.MockCOATypeService
}

func coaTypeTestHelper(t *testing.T) testCOATypeHelper {
	t.Helper()

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockSvc := mock.NewMockCOATypeService(mockCtrl)

	app := echo.New()
	v1Group := app.Group("/api/v1")
	New(v1Group, mockSvc)

	return testCOATypeHelper{
		router:      app,
		mockCtrl:    mockCtrl,
		mockService: mockSvc,
	}
}

func Test_Handler_createCOAType(t *testing.T) {
	testHelper := coaTypeTestHelper(t)

	type args struct {
		ctx         context.Context
		contentType string
		req         models.CreateCOATypeRequest
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
			name: "test success",
			args: args{
				ctx:         context.Background(),
				contentType: echo.MIMEApplicationJSON,
				req: models.CreateCOATypeRequest{
					Code:          "ASS",
					Name:          "Asset",
					NormalBalance: "credit",
					Status:        "active",
				},
			},
			expectation: expectation{
				wantRes:  `{"kind":"coaType","code":"ASS","name":"Asset","normalBalance":"credit","status":"active"}`,
				wantCode: 201,
			},
			doMock: func(args args, expectation expectation) {
				testHelper.mockService.EXPECT().Create(args.ctx, models.CreateCOATypeIn(args.req)).Return(&models.COAType{
					Code:          args.req.Code,
					Name:          args.req.Name,
					NormalBalance: args.req.NormalBalance,
					Status:        args.req.Status,
				}, nil)
			},
		},
		{
			name: "test error contentType",
			args: args{
				ctx:         context.Background(),
				contentType: "",
				req: models.CreateCOATypeRequest{
					Code:          "ASS",
					Name:          "Asset",
					NormalBalance: "credit",
					Status:        "active",
				},
			},
			expectation: expectation{
				wantRes:  `{"status":"error","code":400,"message":"code=415, message=Unsupported Media Type"}`,
				wantCode: 400,
			},
			doMock: func(args args, expectation expectation) {},
		},
		{
			name: "test error validating request",
			args: args{
				ctx:         context.Background(),
				contentType: echo.MIMEApplicationJSON,
				req: models.CreateCOATypeRequest{
					Code:          "AS",
					Name:          "Asset",
					NormalBalance: "credit",
					Status:        "active",
				},
			},
			expectation: expectation{
				wantRes:  `{"status":"error","message":"validation failed","errors":[{"code":"INVALID_LENGTH","field":"code","message":"field must be at least 3 characters"}]}`,
				wantCode: 422,
			},
		},
		{
			name: "test error ErrDataExist",
			args: args{
				ctx:         context.Background(),
				contentType: echo.MIMEApplicationJSON,
				req: models.CreateCOATypeRequest{
					Code:          "ASS",
					Name:          "Asset",
					NormalBalance: "credit",
					Status:        "active",
				},
			},
			expectation: expectation{
				wantRes:  `{"status":"error","code":"DATA_IS_EXIST","message":"data is exist"}`,
				wantCode: 409,
			},
			doMock: func(args args, expectation expectation) {
				testHelper.mockService.EXPECT().Create(args.ctx, models.CreateCOATypeIn(args.req)).Return(&models.COAType{}, models.GetErrMap(models.ErrKeyDataIsExist))
			},
		},
		{
			name: "test error service",
			args: args{
				ctx:         context.Background(),
				contentType: echo.MIMEApplicationJSON,
				req: models.CreateCOATypeRequest{
					Code:          "ASS",
					Name:          "Asset",
					NormalBalance: "credit",
					Status:        "active",
				},
			},
			expectation: expectation{
				wantRes:  `{"status":"error","code":500,"message":"assert.AnError general error for testing"}`,
				wantCode: 500,
			},
			doMock: func(args args, expectation expectation) {
				testHelper.mockService.EXPECT().Create(args.ctx, models.CreateCOATypeIn(args.req)).Return(&models.COAType{}, assert.AnError)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.doMock != nil {
				tt.doMock(tt.args, tt.expectation)
			}

			var b bytes.Buffer
			err := json.NewEncoder(&b).Encode(tt.args.req)
			require.NoError(t, err)

			r := httptest.NewRequest(http.MethodPost, "/api/v1/coa-types", &b)
			r.Header.Set(echo.HeaderContentType, tt.args.contentType)
			w := httptest.NewRecorder()

			testHelper.router.NewContext(r, w)
			testHelper.router.ServeHTTP(w, r)

			require.Equal(t, tt.expectation.wantCode, w.Code)
			require.Equal(t, tt.expectation.wantRes, strings.Trim(w.Body.String(), "\n"))
		})
	}
}

func Test_Handler_getAllCOATypes(t *testing.T) {
	testHelper := coaTypeTestHelper(t)

	type Expectation struct {
		wantRes  string
		wantCode int
	}
	tests := []struct {
		name        string
		expectation Expectation
		doMock      func()
	}{
		{
			name: "happy path",
			expectation: Expectation{
				wantRes:  `{"kind":"collection","contents":[{"kind":"coaType","code":"AST","name":"Asset","normalBalance":"credit","categoryCode":null,"status":"active","createdAt":null,"updatedAt":null}],"total_rows":1}`,
				wantCode: 200,
			},
			doMock: func() {
				testHelper.mockService.EXPECT().GetAll(gomock.AssignableToTypeOf(context.Background())).
					Return([]models.COATypeCategory{{
						Kind:          models.KindCOAType,
						Code:          "AST",
						Name:          "Asset",
						Status:        models.StatusActive,
						NormalBalance: "credit",
					}}, nil)
			},
		},
		{
			name: "error service",
			expectation: Expectation{
				wantRes:  `{"status":"error","code":500,"message":"assert.AnError general error for testing"}`,
				wantCode: 500,
			},
			doMock: func() {
				testHelper.mockService.EXPECT().GetAll(gomock.AssignableToTypeOf(context.Background())).
					Return([]models.COATypeCategory{}, assert.AnError)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.doMock != nil {
				tt.doMock()
			}

			var b bytes.Buffer
			r := httptest.NewRequest(http.MethodGet, "/api/v1/coa-types", &b)
			r.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			w := httptest.NewRecorder()

			testHelper.router.NewContext(r, w)
			testHelper.router.ServeHTTP(w, r)

			require.Equal(t, tt.expectation.wantCode, w.Code)
			require.Equal(t, tt.expectation.wantRes, strings.Trim(w.Body.String(), "\n"))
		})
	}
}

func Test_Handler_updateCOAType(t *testing.T) {
	testHelper := coaTypeTestHelper(t)

	type args struct {
		ctx         context.Context
		contentType string
		req         models.UpdateCOATypeRequest
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
			name:      "success update coa type",
			urlCalled: "/api/v1/coa-types/AST",
			args: args{
				ctx:         context.Background(),
				contentType: echo.MIMEApplicationJSON,
				req: models.UpdateCOATypeRequest{
					Code:          "AST",
					Name:          "test",
					NormalBalance: "debit",
					Status:        "active",
				},
			},
			expectation: expectation{
				wantRes:  `{"kind":"coaType","code":"AST","name":"test","normalBalance":"debit","status":"active"}`,
				wantCode: 200,
			},
			doMock: func(args args, expectation expectation) {
				testHelper.mockService.EXPECT().Update(args.ctx, models.UpdateCOAType{
					Code:          args.req.Code,
					Name:          args.req.Name,
					NormalBalance: args.req.NormalBalance,
					Status:        args.req.Status,
				}).Return(models.UpdateCOAType{
					Code:          args.req.Code,
					Name:          args.req.Name,
					NormalBalance: args.req.NormalBalance,
					Status:        args.req.Status,
				}, nil)
			},
		},
		{
			name:      "error internal server",
			urlCalled: "/api/v1/coa-types/AST",
			args: args{
				ctx:         context.Background(),
				contentType: echo.MIMEApplicationJSON,
				req: models.UpdateCOATypeRequest{
					Code:          "AST",
					Name:          "test",
					NormalBalance: "debit",
					Status:        "active",
				},
			},
			expectation: expectation{
				wantRes:  `{"status":"error","code":500,"message":"internal server error"}`,
				wantCode: 500,
			},
			doMock: func(args args, expectation expectation) {
				testHelper.mockService.EXPECT().Update(args.ctx, models.UpdateCOAType{
					Code:          args.req.Code,
					Name:          args.req.Name,
					NormalBalance: args.req.NormalBalance,
					Status:        args.req.Status,
				}).Return(models.UpdateCOAType{}, models.ErrInternalServerError)
			},
		},
		{
			name:      "error coa type not found",
			urlCalled: "/api/v1/coa-types/AST",
			args: args{
				ctx:         context.Background(),
				contentType: echo.MIMEApplicationJSON,
				req: models.UpdateCOATypeRequest{
					Code:          "AST",
					Name:          "test",
					NormalBalance: "debit",
					Status:        "active",
				},
			},
			expectation: expectation{
				wantRes:  `{"status":"error","code":"DATA_NOT_FOUND","message":"coa type code not found"}`,
				wantCode: 404,
			},
			doMock: func(args args, expectation expectation) {
				testHelper.mockService.EXPECT().Update(args.ctx, models.UpdateCOAType{
					Code:          args.req.Code,
					Name:          args.req.Name,
					NormalBalance: args.req.NormalBalance,
					Status:        args.req.Status,
				}).Return(models.UpdateCOAType{}, models.GetErrMap(models.ErrKeyCoaTypeNotFound))
			},
		},
		{
			name:      "error validation",
			urlCalled: "/api/v1/coa-types/AST",
			args: args{
				ctx:         context.Background(),
				contentType: echo.MIMEApplicationJSON,
				req: models.UpdateCOATypeRequest{
					Code:          "AST",
					Name:          "test",
					NormalBalance: "debits",
					Status:        "active",
				},
			},
			expectation: expectation{
				wantRes:  `{"status":"error","message":"validation failed","errors":[{"code":"INVALID_VALUES","field":"normalBalance","message":"one of debit or credit"}]}`,
				wantCode: 422,
			},
			doMock: func(args args, expectation expectation) {},
		},
		{
			name:      "error contentType",
			urlCalled: "/api/v1/coa-types/AST",
			args: args{
				ctx:         context.Background(),
				contentType: "",
				req: models.UpdateCOATypeRequest{
					Code:          "AST",
					Name:          "test",
					NormalBalance: "debits",
					Status:        "active",
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
