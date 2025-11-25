package category

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

type testCategoryHelper struct {
	router      *echo.Echo
	mockCtrl    *gomock.Controller
	mockService *mock.MockCategoryService
}

func categoryTestHelper(t *testing.T) testCategoryHelper {
	t.Helper()

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockSvc := mock.NewMockCategoryService(mockCtrl)

	app := echo.New()
	v1Group := app.Group("/api/v1")
	New(v1Group, mockSvc)

	return testCategoryHelper{
		router:      app,
		mockCtrl:    mockCtrl,
		mockService: mockSvc,
	}
}

func Test_Handler_createCategory(t *testing.T) {
	testHelper := categoryTestHelper(t)

	type args struct {
		ctx         context.Context
		contentType string
		req         models.DoCreateCategoryRequest
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
				req: models.DoCreateCategoryRequest{
					Code:        "001",
					Name:        "test",
					Description: "",
					CoaTypeCode: "AST",
					Status:      models.StatusActive,
				},
			},
			expectation: expectation{
				wantRes:  `{"kind":"category","code":"001","name":"test","description":"","coaTypeCode":"","status":""}`,
				wantCode: 201,
			},
			doMock: func(args args, expectation expectation) {
				testHelper.mockService.EXPECT().Create(args.ctx, models.CreateCategoryIn(args.req)).Return(&models.Category{
					Code: args.req.Code,
					Name: args.req.Name,
				}, nil)
			},
		},
		{
			name: "error contentType",
			args: args{
				ctx:         context.Background(),
				contentType: "",
				req: models.DoCreateCategoryRequest{
					Code:        "001",
					Name:        "test",
					Description: "",
					CoaTypeCode: "AST",
					Status:      models.StatusActive,
				},
			},
			expectation: expectation{
				wantRes:  `{"status":"error","code":400,"message":"code=415, message=Unsupported Media Type"}`,
				wantCode: 400,
			},
			doMock: func(args args, expectation expectation) {},
		},
		{
			name: "error validating request",
			args: args{
				ctx:         context.Background(),
				contentType: echo.MIMEApplicationJSON,
				req: models.DoCreateCategoryRequest{
					Code: "12",
				},
			},
			expectation: expectation{
				wantRes:  `{"status":"error","message":"validation failed","errors":[{"code":"INVALID_LENGTH","field":"code","message":"field must be at least 3 characters"},{"code":"MISSING_FIELD","field":"name","message":"field is missing"},{"code":"MISSING_FIELD","field":"coaTypeCode","message":"field is missing"},{"code":"MISSING_FIELD","field":"status","message":"field is missing"}]}`,
				wantCode: 422,
			},
		},
		{
			name: "error data is exist",
			args: args{
				ctx:         context.Background(),
				contentType: echo.MIMEApplicationJSON,
				req: models.DoCreateCategoryRequest{
					Code:        "001",
					Name:        "test",
					Description: "",
					CoaTypeCode: "AST",
					Status:      models.StatusActive,
				},
			},
			expectation: expectation{
				wantRes:  `{"status":"error","code":"DATA_IS_EXIST","message":"data is exist"}`,
				wantCode: 409,
			},
			doMock: func(args args, expectation expectation) {
				testHelper.mockService.EXPECT().Create(args.ctx, models.CreateCategoryIn(args.req)).Return(&models.Category{}, models.GetErrMap(models.ErrKeyDataIsExist))
			},
		},
		{
			name: "error data not found",
			args: args{
				ctx:         context.Background(),
				contentType: echo.MIMEApplicationJSON,
				req: models.DoCreateCategoryRequest{
					Code:        "001",
					Name:        "test",
					Description: "",
					CoaTypeCode: "AST",
					Status:      models.StatusActive,
				},
			},
			expectation: expectation{
				wantRes:  `{"status":"error","code":"DATA_NOT_FOUND","message":"data not found"}`,
				wantCode: 404,
			},
			doMock: func(args args, expectation expectation) {
				testHelper.mockService.EXPECT().Create(args.ctx, models.CreateCategoryIn(args.req)).Return(&models.Category{}, models.GetErrMap(models.ErrKeyDataNotFound))
			},
		},
		{
			name: "error service",
			args: args{
				ctx:         context.Background(),
				contentType: echo.MIMEApplicationJSON,
				req: models.DoCreateCategoryRequest{
					Code:        "001",
					Name:        "test",
					Description: "",
					CoaTypeCode: "AST",
					Status:      models.StatusActive,
				},
			},
			expectation: expectation{
				wantRes:  `{"status":"error","code":500,"message":"assert.AnError general error for testing"}`,
				wantCode: 500,
			},
			doMock: func(args args, expectation expectation) {
				testHelper.mockService.EXPECT().Create(args.ctx, models.CreateCategoryIn(args.req)).Return(&models.Category{}, assert.AnError)
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

			r := httptest.NewRequest(http.MethodPost, "/api/v1/categories", &b)
			r.Header.Set(echo.HeaderContentType, tt.args.contentType)
			w := httptest.NewRecorder()

			testHelper.router.NewContext(r, w)
			testHelper.router.ServeHTTP(w, r)

			require.Equal(t, tt.expectation.wantCode, w.Code)
			require.Equal(t, tt.expectation.wantRes, strings.Trim(w.Body.String(), "\n"))
		})
	}
}

func Test_Handler_getAllCategory(t *testing.T) {
	testHelper := categoryTestHelper(t)

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
				wantRes:  `{"kind":"collection","contents":[{"kind":"category","code":"01","name":"tes","description":"test","coaTypeCode":"AST","status":"active"}],"total_rows":1}`,
				wantCode: 200,
			},
			doMock: func() {
				testHelper.mockService.EXPECT().GetAll(gomock.AssignableToTypeOf(context.Background())).
					Return(&[]models.Category{{
						ID:          0,
						Code:        "01",
						Name:        "tes",
						Description: "test",
						CoaTypeCode: "AST",
						Status:      models.StatusActive,
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
					Return(&[]models.Category{}, assert.AnError)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.doMock != nil {
				tt.doMock()
			}

			var b bytes.Buffer

			r := httptest.NewRequest(http.MethodGet, "/api/v1/categories", &b)
			r.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			w := httptest.NewRecorder()

			testHelper.router.NewContext(r, w)
			testHelper.router.ServeHTTP(w, r)

			require.Equal(t, tt.expectation.wantCode, w.Code)
			require.Equal(t, tt.expectation.wantRes, strings.Trim(w.Body.String(), "\n"))
		})
	}
}

func Test_Handler_updateCategory(t *testing.T) {
	testHelper := categoryTestHelper(t)

	req := models.DoUpdateCategoryRequest{
		Name:        "Kas Teller 111",
		Description: "Kas Teller ceritanya 111",
		CoaTypeCode: "LIA",
		Code:        "111",
	}
	type args struct {
		ctx         context.Context
		contentType string
		req         models.DoUpdateCategoryRequest
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
				wantRes:  `{"kind":"category","code":"111","name":"Kas Teller 111","description":"Kas Teller ceritanya 111","coaTypeCode":"LIA"}`,
				wantCode: 200,
			},
			doMock: func(args args, expectation expectation) {
				testHelper.mockService.EXPECT().Update(args.ctx, models.UpdateCategoryIn(args.req)).Return(models.UpdateCategoryIn(args.req), nil)
			},
		},
		{
			name: "error case - Unsupported Media Type",
			args: args{
				ctx:         context.Background(),
				contentType: "",
				req:         req,
			},
			expectation: expectation{
				wantRes:  `{"status":"error","code":400,"message":"code=415, message=Unsupported Media Type"}`,
				wantCode: 400,
			},
			doMock: func(args args, expectation expectation) {},
		},
		{
			name: "error case - validation failed",
			args: args{
				ctx:         context.Background(),
				contentType: echo.MIMEApplicationJSON,
				req: models.DoUpdateCategoryRequest{
					Name:        "test",
					Description: "testabc",
					CoaTypeCode: "AS",
				},
			},
			expectation: expectation{
				wantRes:  `{"status":"error","message":"validation failed","errors":[{"code":"INVALID_LENGTH","field":"coaTypeCode","message":"field must be at least 3 characters"},{"code":"MISSING_FIELD","field":"code","message":"field is missing"}]}`,
				wantCode: 422,
			},
		},
		{
			name: "error case - data not found",
			args: args{
				ctx:         context.Background(),
				contentType: echo.MIMEApplicationJSON,
				req:         req,
			},
			expectation: expectation{
				wantRes:  `{"status":"error","code":"DATA_NOT_FOUND","message":"category code not found"}`,
				wantCode: 404,
			},
			doMock: func(args args, expectation expectation) {
				testHelper.mockService.EXPECT().Update(args.ctx, models.UpdateCategoryIn(args.req)).Return(models.UpdateCategoryIn(args.req), models.GetErrMap(models.ErrKeyCategoryCodeNotFound))
			},
		},
		{
			name: "error case - internal server error",
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
				testHelper.mockService.EXPECT().Update(args.ctx, models.UpdateCategoryIn(args.req)).Return(models.UpdateCategoryIn(args.req), models.GetErrMap(models.ErrKeyDatabaseError))
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

			r := httptest.NewRequest(http.MethodPatch, "/api/v1/categories/111", &b)
			r.Header.Set(echo.HeaderContentType, tt.args.contentType)
			w := httptest.NewRecorder()

			testHelper.router.NewContext(r, w)
			testHelper.router.ServeHTTP(w, r)

			require.Equal(t, tt.expectation.wantCode, w.Code)
			require.Equal(t, tt.expectation.wantRes, strings.Trim(w.Body.String(), "\n"))
		})
	}
}
