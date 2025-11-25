package cache

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
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestMain(m *testing.M) {
	xlog.InitForTest()
	os.Exit(m.Run())
}

type testCacheHelper struct {
	router      *echo.Echo
	mockCtrl    *gomock.Controller
	mockService *mock.MockAccountService
}

func cacheTestHelper(t *testing.T) testCacheHelper {
	t.Helper()

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockAccountSvc := mock.NewMockAccountService(mockCtrl)

	app := echo.New()
	v1Group := app.Group("/api/v1")
	New(v1Group, mockAccountSvc)

	return testCacheHelper{
		router:      app,
		mockCtrl:    mockCtrl,
		mockService: mockAccountSvc,
	}
}

func Test_Handler_getAllCategoryCodeSeq(t *testing.T) {
	testHelper := cacheTestHelper(t)

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
			name: "success case",
			expectation: Expectation{
				wantRes:  `{"kind":"collection","contents":[{"kind":"account","key":"category_code_212_seq","value":"100"}],"total_rows":1}`,
				wantCode: 200,
			},
			doMock: func() {
				testHelper.mockService.EXPECT().GetAllCategoryCodeSeq(gomock.AssignableToTypeOf(context.Background())).
					Return([]models.DoGetAllCategoryCodeSeqResponse{{
						Kind:  models.KindAccount,
						Key:   "category_code_212_seq",
						Value: "100",
					}}, nil)
			},
		},
		{
			name: "error case - database error",
			expectation: Expectation{
				wantRes:  `{"status":"error","code":"DATABASE_ERROR","message":"database error"}`,
				wantCode: 500,
			},
			doMock: func() {
				testHelper.mockService.EXPECT().GetAllCategoryCodeSeq(gomock.AssignableToTypeOf(context.Background())).
					Return([]models.DoGetAllCategoryCodeSeqResponse{}, models.GetErrMap(models.ErrKeyDatabaseError))
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.doMock != nil {
				tt.doMock()
			}

			r := httptest.NewRequest(http.MethodGet, "/api/v1/cache/accounts/category-codes-seq", nil)
			r.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			w := httptest.NewRecorder()

			testHelper.router.NewContext(r, w)
			testHelper.router.ServeHTTP(w, r)

			require.Equal(t, tt.expectation.wantCode, w.Code)
			require.Equal(t, tt.expectation.wantRes, strings.Trim(w.Body.String(), "\n"))
		})
	}
}

func Test_Handler_updateCategoryCodeSeq(t *testing.T) {
	testHelper := cacheTestHelper(t)

	type args struct {
		contentType string
		req         models.DoUpdateCategoryCodeSeqRequest
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
			name: "success case - update category sequence",
			args: args{
				contentType: echo.MIMEApplicationJSON,
				req: models.DoUpdateCategoryCodeSeqRequest{
					Key:   "category_code_212_seq",
					Value: 100,
				},
			},
			expectation: expectation{
				wantRes:  `{"kind":"account","key":"category_code_212_seq","value":100}`,
				wantCode: 200,
			},
			doMock: func(args args, expectation expectation) {
				testHelper.mockService.EXPECT().UpdateCategoryCodeSeq(gomock.AssignableToTypeOf(context.Background()), args.req).Return(nil)
			},
		},
		{
			name: "error case - category sequence not found",
			args: args{
				contentType: echo.MIMEApplicationJSON,
				req: models.DoUpdateCategoryCodeSeqRequest{
					Key:   "category_code_212_seq",
					Value: 100,
				},
			},
			expectation: expectation{
				wantRes:  `{"status":"error","code":"DATA_NOT_FOUND","message":"data not found"}`,
				wantCode: 404,
			},
			doMock: func(args args, expectation expectation) {
				testHelper.mockService.EXPECT().UpdateCategoryCodeSeq(gomock.AssignableToTypeOf(context.Background()), args.req).Return(models.GetErrMap(models.ErrKeyDataNotFound))
			},
		},
		{
			name: "error case - internal server error",
			args: args{
				contentType: echo.MIMEApplicationJSON,
				req: models.DoUpdateCategoryCodeSeqRequest{
					Key:   "category_code_212_seq",
					Value: 100,
				},
			},
			expectation: expectation{
				wantRes:  `{"status":"error","code":"CACHE_ERROR","message":"failed set to cache"}`,
				wantCode: 500,
			},
			doMock: func(args args, expectation expectation) {
				testHelper.mockService.EXPECT().UpdateCategoryCodeSeq(gomock.AssignableToTypeOf(context.Background()), args.req).Return(models.GetErrMap(models.ErrKeyFailedSetToCache))
			},
		},
		{
			name: "error case - validation error",
			args: args{
				contentType: echo.MIMEApplicationJSON,
				req:         models.DoUpdateCategoryCodeSeqRequest{},
			},
			expectation: expectation{
				wantRes:  `{"status":"error","message":"validation failed","errors":[{"code":"MISSING_FIELD","field":"key","message":"field is missing"},{"code":"MISSING_FIELD","field":"value","message":"field is missing"}]}`,
				wantCode: 422,
			},
		},
		{
			name: "error case - Unsupported Media Type",
			args: args{
				contentType: "",
			},
			expectation: expectation{
				wantRes:  `{"status":"error","code":400,"message":"code=415, message=Unsupported Media Type"}`,
				wantCode: 400,
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

			r := httptest.NewRequest(http.MethodPut, "/api/v1/cache/accounts/category-codes-seq", &b)
			r.Header.Set(echo.HeaderContentType, tt.args.contentType)
			w := httptest.NewRecorder()

			testHelper.router.NewContext(r, w)
			testHelper.router.ServeHTTP(w, r)

			require.Equal(t, tt.expectation.wantCode, w.Code)
			require.Equal(t, tt.expectation.wantRes, strings.Trim(w.Body.String(), "\n"))
		})
	}
}

func Test_Handler_createCategoryCodeSeq(t *testing.T) {
	testHelper := cacheTestHelper(t)

	type args struct {
		contentType string
		req         models.DoCreateCategoryCodeSeqRequest
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
			name: "success case - create category sequence",
			args: args{
				contentType: echo.MIMEApplicationJSON,
				req: models.DoCreateCategoryCodeSeqRequest{
					Key:   "category_code_212_seq",
					Value: 100,
				},
			},
			expectation: expectation{
				wantRes:  `{"kind":"account","key":"category_code_212_seq","value":100}`,
				wantCode: 200,
			},
			doMock: func(args args, expectation expectation) {
				testHelper.mockService.EXPECT().CreateCategoryCodeSeq(gomock.AssignableToTypeOf(context.Background()), args.req).Return(nil)
			},
		},
		{
			name: "error case - category sequence found",
			args: args{
				contentType: echo.MIMEApplicationJSON,
				req: models.DoCreateCategoryCodeSeqRequest{
					Key:   "category_code_212_seq",
					Value: 100,
				},
			},
			expectation: expectation{
				wantRes:  `{"status":"error","code":"DATA_IS_EXIST","message":"data is exist"}`,
				wantCode: 409,
			},
			doMock: func(args args, expectation expectation) {
				testHelper.mockService.EXPECT().CreateCategoryCodeSeq(gomock.AssignableToTypeOf(context.Background()), args.req).Return(models.GetErrMap(models.ErrKeyDataIsExist))
			},
		},
		{
			name: "error case - internal server error",
			args: args{
				contentType: echo.MIMEApplicationJSON,
				req: models.DoCreateCategoryCodeSeqRequest{
					Key:   "category_code_212_seq",
					Value: 100,
				},
			},
			expectation: expectation{
				wantRes:  `{"status":"error","code":"CACHE_ERROR","message":"failed set to cache"}`,
				wantCode: 500,
			},
			doMock: func(args args, expectation expectation) {
				testHelper.mockService.EXPECT().CreateCategoryCodeSeq(gomock.AssignableToTypeOf(context.Background()), args.req).Return(models.GetErrMap(models.ErrKeyFailedSetToCache))
			},
		},
		{
			name: "error case - validation error",
			args: args{
				contentType: echo.MIMEApplicationJSON,
				req:         models.DoCreateCategoryCodeSeqRequest{},
			},
			expectation: expectation{
				wantRes:  `{"status":"error","message":"validation failed","errors":[{"code":"MISSING_FIELD","field":"key","message":"field is missing"},{"code":"MISSING_FIELD","field":"value","message":"field is missing"}]}`,
				wantCode: 422,
			},
		},
		{
			name: "error case - Unsupported Media Type",
			args: args{
				contentType: "",
			},
			expectation: expectation{
				wantRes:  `{"status":"error","code":400,"message":"code=415, message=Unsupported Media Type"}`,
				wantCode: 400,
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

			r := httptest.NewRequest(http.MethodPost, "/api/v1/cache/accounts/category-codes-seq", &b)
			r.Header.Set(echo.HeaderContentType, tt.args.contentType)
			w := httptest.NewRecorder()

			testHelper.router.NewContext(r, w)
			testHelper.router.ServeHTTP(w, r)

			require.Equal(t, tt.expectation.wantCode, w.Code)
			require.Equal(t, tt.expectation.wantRes, strings.Trim(w.Body.String(), "\n"))
		})
	}
}
