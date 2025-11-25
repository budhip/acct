package producttype

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

type testProductTypeHelper struct {
	router      *echo.Echo
	mockCtrl    *gomock.Controller
	mockService *mock.MockProductTypeService
}

func productTypeTestHelper(t *testing.T) testProductTypeHelper {
	t.Helper()

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockSvc := mock.NewMockProductTypeService(mockCtrl)

	app := echo.New()
	v1Group := app.Group("/api/v1")
	New(v1Group, mockSvc)

	return testProductTypeHelper{
		router:      app,
		mockCtrl:    mockCtrl,
		mockService: mockSvc,
	}
}

func TestMain(m *testing.M) {
	xlog.InitForTest()
	os.Exit(m.Run())
}

func Test_Handler_getAllProductType(t *testing.T) {
	testHelper := productTypeTestHelper(t)

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
			name: "success get all data product type",
			expectation: Expectation{
				wantRes:  `{"kind":"collection","contents":[{"kind":"productType","code":"1001","name":"Chickin","status":"active","entityCode":"001"}],"total_rows":1}`,
				wantCode: 200,
			},
			doMock: func() {
				testHelper.mockService.EXPECT().GetAll(gomock.AssignableToTypeOf(context.Background())).
					Return([]models.ProductType{{
						ID:         1,
						Code:       "1001",
						Name:       "Chickin",
						Status:     models.StatusActive,
						EntityCode: "001",
					}}, nil)
			},
		},
		{
			name: "failed to get data product type",
			expectation: Expectation{
				wantRes:  `{"status":"error","code":500,"message":"assert.AnError general error for testing"}`,
				wantCode: 500,
			},
			doMock: func() {
				testHelper.mockService.EXPECT().
					GetAll(gomock.AssignableToTypeOf(context.Background())).
					Return([]models.ProductType{}, assert.AnError)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.doMock != nil {
				tt.doMock()
			}

			var b bytes.Buffer
			r := httptest.NewRequest(http.MethodGet, "/api/v1/product-types", &b)
			r.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			w := httptest.NewRecorder()

			testHelper.router.NewContext(r, w)
			testHelper.router.ServeHTTP(w, r)

			require.Equal(t, tt.expectation.wantCode, w.Code)
			require.Equal(t, tt.expectation.wantRes, strings.Trim(w.Body.String(), "\n"))
		})
	}
}

func Test_Handler_createProductType(t *testing.T) {
	testHelper := productTypeTestHelper(t)

	type args struct {
		ctx         context.Context
		contentType string
		req         models.CreateProductTypeRequest
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
				req: models.CreateProductTypeRequest{
					Code:       "1001",
					Name:       "Chickin",
					Status:     "active",
					EntityCode: "001",
				},
			},
			expectation: expectation{
				wantRes:  `{"kind":"productType","code":"1001","name":"Chickin","status":"active","entityCode":"001"}`,
				wantCode: 201,
			},
			doMock: func(args args, expectation expectation) {
				testHelper.mockService.EXPECT().Create(args.ctx, args.req).
					Return(&models.ProductType{
						Code:       "1001",
						Name:       "Chickin",
						Status:     "active",
						EntityCode: "001",
					}, nil)
			},
		},
		{
			name: "error case - validation failed",
			args: args{
				ctx:         context.Background(),
				contentType: echo.MIMEApplicationJSON,
				req: models.CreateProductTypeRequest{
					Code: "12",
				},
			},
			expectation: expectation{
				wantRes:  `{"status":"error","message":"validation failed","errors":[{"code":"INVALID_LENGTH","field":"code","message":"field must be at least 3 characters"},{"code":"MISSING_FIELD","field":"name","message":"field is missing"},{"code":"MISSING_FIELD","field":"status","message":"field is missing"}]}`,
				wantCode: 422,
			},
		},
		{
			name: "error case - unsupported media type",
			args: args{
				ctx:         context.Background(),
				contentType: "",
				req: models.CreateProductTypeRequest{
					Code:       "1001",
					Name:       "Chinkin",
					Status:     "active",
					EntityCode: "001",
				},
			},
			expectation: expectation{
				wantRes:  `{"status":"error","code":400,"message":"code=415, message=Unsupported Media Type"}`,
				wantCode: 400,
			},
			doMock: func(args args, expectation expectation) {},
		},
		{
			name: "error case - product type code is exist",
			args: args{
				ctx:         context.Background(),
				contentType: echo.MIMEApplicationJSON,
				req: models.CreateProductTypeRequest{
					Code:       "1001",
					Name:       "Chinkin",
					Status:     "active",
					EntityCode: "001",
				},
			},
			expectation: expectation{
				wantRes:  `{"status":"error","code":"DATA_IS_EXIST","message":"product type code is exist"}`,
				wantCode: 409,
			},
			doMock: func(args args, expectation expectation) {
				testHelper.mockService.EXPECT().Create(args.ctx, args.req).
					Return(&models.ProductType{}, models.GetErrMap(models.ErrKeyProductTypeCodeIsExist))
			},
		},
		{
			name: "error case - entity code not found",
			args: args{
				ctx:         context.Background(),
				contentType: echo.MIMEApplicationJSON,
				req: models.CreateProductTypeRequest{
					Code:       "1001",
					Name:       "Chinkin",
					Status:     "active",
					EntityCode: "001",
				},
			},
			expectation: expectation{
				wantRes:  `{"status":"error","code":"DATA_NOT_FOUND","message":"entity code not found"}`,
				wantCode: 404,
			},
			doMock: func(args args, expectation expectation) {
				testHelper.mockService.EXPECT().Create(args.ctx, args.req).
					Return(&models.ProductType{}, models.GetErrMap(models.ErrKeyEntityCodeNotFound))
			},
		},
		{
			name: "error case - internal server error",
			args: args{
				ctx:         context.Background(),
				contentType: echo.MIMEApplicationJSON,
				req: models.CreateProductTypeRequest{
					Code:       "1001",
					Name:       "Chinkin",
					Status:     "active",
					EntityCode: "001",
				},
			},
			expectation: expectation{
				wantRes:  `{"status":"error","code":500,"message":"assert.AnError general error for testing"}`,
				wantCode: 500,
			},
			doMock: func(args args, expectation expectation) {
				testHelper.mockService.EXPECT().Create(args.ctx, args.req).
					Return(&models.ProductType{}, assert.AnError)
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

			r := httptest.NewRequest(http.MethodPost, "/api/v1/product-types", &b)
			r.Header.Set(echo.HeaderContentType, tt.args.contentType)
			w := httptest.NewRecorder()

			testHelper.router.NewContext(r, w)
			testHelper.router.ServeHTTP(w, r)

			require.Equal(t, tt.expectation.wantCode, w.Code)
			require.Equal(t, tt.expectation.wantRes, strings.Trim(w.Body.String(), "\n"))
		})
	}
}

func Test_Handler_updateProductType(t *testing.T) {
	testHelper := productTypeTestHelper(t)

	req := models.UpdateProductTypeRequest{
		Code:   "1001",
		Name:   "Chickin",
		Status: "active",
	}
	type args struct {
		ctx         context.Context
		contentType string
		req         models.UpdateProductTypeRequest
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
			name:      "success case - update product type",
			urlCalled: "/api/v1/product-types/1001",
			args: args{
				ctx:         context.Background(),
				contentType: echo.MIMEApplicationJSON,
				req:         req,
			},
			expectation: expectation{
				wantRes:  `{"kind":"productType","code":"1001","name":"Chickin","status":"active"}`,
				wantCode: 200,
			},
			doMock: func(args args, expectation expectation) {
				testHelper.mockService.EXPECT().Update(args.ctx, models.UpdateProductType(args.req)).Return(models.UpdateProductType(args.req), nil)
			},
		},
		{
			name:      "error case - validation failed",
			urlCalled: "/api/v1/product-types/1001",
			args: args{
				ctx:         context.Background(),
				contentType: echo.MIMEApplicationJSON,
				req: models.UpdateProductTypeRequest{
					Code:   "001",
					Name:   "tes_test",
					Status: "test",
				},
			},
			expectation: expectation{
				wantRes:  `{"status":"error","message":"validation failed","errors":[{"code":"INVALID_VALUES","field":"name","message":"input field no special characters"},{"code":"INVALID_VALUES","field":"status","message":"one of active or inactive"}]}`,
				wantCode: 422,
			},
		},
		{
			name:      "error case - product type not found",
			urlCalled: "/api/v1/product-types/1001",
			args: args{
				ctx:         context.Background(),
				contentType: echo.MIMEApplicationJSON,
				req:         req,
			},
			expectation: expectation{
				wantRes:  `{"status":"error","code":"DATA_NOT_FOUND","message":"product type code not found"}`,
				wantCode: 404,
			},
			doMock: func(args args, expectation expectation) {
				testHelper.mockService.EXPECT().Update(args.ctx, models.UpdateProductType(args.req)).Return(models.UpdateProductType{}, models.GetErrMap(models.ErrKeyProductTypeNotFound))
			},
		},
		{
			name:      "error case - internal server error",
			urlCalled: "/api/v1/product-types/1001",
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
				testHelper.mockService.EXPECT().Update(args.ctx, models.UpdateProductType(args.req)).
					Return(models.UpdateProductType{}, models.GetErrMap(models.ErrKeyDatabaseError))

			},
		},
		{
			name:      "error contentType",
			urlCalled: "/api/v1/product-types/1001",
			args: args{
				ctx:         context.Background(),
				contentType: "",
				req:         req,
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
