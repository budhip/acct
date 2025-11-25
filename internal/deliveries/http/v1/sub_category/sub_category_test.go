package subcategory

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

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

type testSubCategoryHelper struct {
	router      *echo.Echo
	mockCtrl    *gomock.Controller
	mockService *mock.MockSubCategoryService
}

func subCategoryTestHelper(t *testing.T) testSubCategoryHelper {
	t.Helper()

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockSvc := mock.NewMockSubCategoryService(mockCtrl)

	app := echo.New()
	v1Group := app.Group("/api/v1")
	New(v1Group, mockSvc)

	return testSubCategoryHelper{
		router:      app,
		mockCtrl:    mockCtrl,
		mockService: mockSvc,
	}
}

func Test_Handler_createSubCategory(t *testing.T) {
	testHelper := subCategoryTestHelper(t)

	req := models.CreateSubCategoryRequest{
		Code:            "11405",
		Name:            "Borrower Outstanding - Chickin",
		Description:     "Borrower Outstanding - Chickin",
		CategoryCode:    "114",
		AccountType:     "LOAN_ACCOUNT_CHICKIN",
		Status:          "active",
		ProductTypeCode: "1001",
		Currency:        models.CurrencyIDR,
	}
	res := models.SubCategory{
		Code:            "11405",
		Name:            "Borrower Outstanding - Chickin",
		Description:     "Borrower Outstanding - Chickin",
		CategoryCode:    "114",
		AccountType:     "LOAN_ACCOUNT_CHICKIN",
		Status:          "active",
		ProductTypeCode: "1001",
		Currency:        models.CurrencyIDR,
	}
	type args struct {
		ctx         context.Context
		contentType string
		req         models.CreateSubCategoryRequest
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
				wantRes:  `{"kind":"subCategory","code":"11405","name":"Borrower Outstanding - Chickin","description":"Borrower Outstanding - Chickin","categoryCode":"114","accountType":"LOAN_ACCOUNT_CHICKIN","productTypeCode":"1001","currency":"IDR","status":"active"}`,
				wantCode: 201,
			},
			doMock: func(args args, expectation expectation) {
				testHelper.mockService.EXPECT().Create(args.ctx, models.CreateSubCategory(args.req)).
					Return(&res, nil)
			},
		},
		{
			name: "error case -Unsupported Media Type",
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
		{
			name: "error case - request validation failed",
			args: args{
				ctx:         context.Background(),
				contentType: echo.MIMEApplicationJSON,
				req:         models.CreateSubCategoryRequest{},
			},
			expectation: expectation{
				wantRes:  `{"status":"error","message":"validation failed","errors":[{"code":"MISSING_FIELD","field":"code","message":"field is missing"},{"code":"MISSING_FIELD","field":"name","message":"field is missing"},{"code":"MISSING_FIELD","field":"categoryCode","message":"field is missing"},{"code":"MISSING_FIELD","field":"status","message":"field is missing"}]}`,
				wantCode: 422,
			},
		},
		{
			name: "error case - category not found",
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
				testHelper.mockService.EXPECT().Create(args.ctx, models.CreateSubCategory(args.req)).
					Return(&models.SubCategory{}, models.GetErrMap(models.ErrKeyCategoryCodeNotFound))
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
				testHelper.mockService.EXPECT().Create(args.ctx, models.CreateSubCategory(args.req)).
					Return(&models.SubCategory{}, models.GetErrMap(models.ErrKeyDataIsExist))
			},
		},
		{
			name: "error case - internal server error",
			args: args{
				ctx:         context.Background(),
				contentType: echo.MIMEApplicationJSON,
				req: models.CreateSubCategoryRequest{
					CategoryCode:    "001",
					Code:            "00001",
					Name:            "test",
					Description:     "TEST DESC",
					AccountType:     "LENDER_INSTI",
					Status:          "active",
					ProductTypeCode: "101",
					Currency:        models.CurrencyIDR,
				},
			},
			expectation: expectation{
				wantRes:  `{"status":"error","code":"DATABASE_ERROR","message":"database error"}`,
				wantCode: 500,
			},
			doMock: func(args args, expectation expectation) {
				testHelper.mockService.EXPECT().Create(args.ctx, models.CreateSubCategory(args.req)).
					Return(&models.SubCategory{}, models.GetErrMap(models.ErrKeyDatabaseError))
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

			r := httptest.NewRequest(http.MethodPost, "/api/v1/sub-categories", &b)
			r.Header.Set(echo.HeaderContentType, tt.args.contentType)
			w := httptest.NewRecorder()

			testHelper.router.NewContext(r, w)
			testHelper.router.ServeHTTP(w, r)

			require.Equal(t, tt.expectation.wantCode, w.Code)
			require.Equal(t, tt.expectation.wantRes, strings.Trim(w.Body.String(), "\n"))
		})
	}
}

func Test_Handler_getAllSubCategory(t *testing.T) {
	testHelper := subCategoryTestHelper(t)
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
			name: "success get all sub category",
			expectation: Expectation{
				wantRes:  `{"kind":"collection","contents":[{"kind":"subCategory","code":"10000","name":"RETAIL","description":"sub category","categoryCode":"221","accountType":"LENDER_INSTI","productTypeCode":"","productTypeName":"Graduation Loan","currency":"","status":"active","createdAt":"0001-01-01T00:00:00Z","updatedAt":"0001-01-01T00:00:00Z"}],"total_rows":1}`,
				wantCode: 200,
			},
			doMock: func() {
				testHelper.mockService.EXPECT().GetAll(gomock.AssignableToTypeOf(context.Background()), models.GetAllSubCategoryParam{}).
					Return(&[]models.SubCategory{{
						CategoryCode:    "221",
						Code:            "10000",
						Name:            "RETAIL",
						Description:     "sub category",
						AccountType:     "LENDER_INSTI",
						Status:          "active",
						CreatedAt:       &time.Time{},
						UpdatedAt:       &time.Time{},
						ProductTypeName: "Graduation Loan",
					}}, nil)
			},
		},
		{
			name: "failed to get data sub category",
			expectation: Expectation{
				wantRes:  `{"status":"error","code":500,"message":"assert.AnError general error for testing"}`,
				wantCode: 500,
			},
			doMock: func() {
				testHelper.mockService.EXPECT().
					GetAll(gomock.AssignableToTypeOf(context.Background()), models.GetAllSubCategoryParam{}).
					Return(&[]models.SubCategory{}, assert.AnError)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.doMock != nil {
				tt.doMock()
			}
			var b bytes.Buffer
			r := httptest.NewRequest(http.MethodGet, "/api/v1/sub-categories", &b)
			r.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			w := httptest.NewRecorder()

			testHelper.router.NewContext(r, w)
			testHelper.router.ServeHTTP(w, r)

			require.Equal(t, tt.expectation.wantCode, w.Code)
			require.Equal(t, tt.expectation.wantRes, strings.Trim(w.Body.String(), "\n"))
		})
	}
}

func Test_Handler_updateSubCategory(t *testing.T) {
	testHelper := subCategoryTestHelper(t)
	type args struct {
		ctx         context.Context
		contentType string
		req         models.UpdateSubCategoryRequest
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
			name: "success case - update sub category",
			args: args{
				ctx:         context.Background(),
				contentType: echo.MIMEApplicationJSON,
				req: models.UpdateSubCategoryRequest{
					Name:        "RETAIL",
					Description: "Retail",
					Code:        "10000",
					Status:      "active",
				},
			},
			expectation: expectation{
				wantRes:  `{"kind":"subCategory","code":"10000","name":"RETAIL","description":"Retail","status":"active"}`,
				wantCode: 200,
			},
			doMock: func(args args, expectation expectation) {
				testHelper.mockService.EXPECT().Update(args.ctx, models.UpdateSubCategory(args.req)).Return(models.UpdateSubCategory(args.req), nil)
			},
		},
		{
			name: "error case - Unsupported Media Type",
			args: args{
				ctx: context.Background(),
				req: models.UpdateSubCategoryRequest{
					Code: "10000",
				},
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
				req: models.UpdateSubCategoryRequest{
					Name:        "RETAIL",
					Description: "Retail",
					Status:      "actives",
					Code:        "100000",
				},
			},
			expectation: expectation{
				wantRes:  `{"status":"error","message":"validation failed","errors":[{"code":"INVALID_VALUES","field":"status","message":"one of active or inactive"},{"code":"INVALID_LENGTH","field":"code","message":"field can have a maximum length of 3 characters"}]}`,
				wantCode: 422,
			},
		},
		{
			name: "error case - sub category code not found",
			args: args{
				ctx:         context.Background(),
				contentType: echo.MIMEApplicationJSON,
				req: models.UpdateSubCategoryRequest{
					Name:        "RETAIL",
					Description: "Retail",
					Code:        "10000",
					Status:      "active",
				},
			},
			expectation: expectation{
				wantRes:  `{"status":"error","code":"DATA_NOT_FOUND","message":"sub category code not found"}`,
				wantCode: 404,
			},
			doMock: func(args args, expectation expectation) {
				testHelper.mockService.EXPECT().Update(args.ctx, models.UpdateSubCategory(args.req)).Return(models.UpdateSubCategory(args.req), models.GetErrMap(models.ErrKeySubCategoryCodeNotFound))
			},
		},
		{
			name: "error case - product type code not found",
			args: args{
				ctx:         context.Background(),
				contentType: echo.MIMEApplicationJSON,
				req: models.UpdateSubCategoryRequest{
					Name:        "RETAIL",
					Description: "Retail",
					Code:        "10000",
					Status:      "active",
				},
			},
			expectation: expectation{
				wantRes:  `{"status":"error","code":"DATA_NOT_FOUND","message":"product type code not found"}`,
				wantCode: 404,
			},
			doMock: func(args args, expectation expectation) {
				testHelper.mockService.EXPECT().Update(args.ctx, models.UpdateSubCategory(args.req)).Return(models.UpdateSubCategory(args.req), models.GetErrMap(models.ErrKeyProductTypeNotFound))
			},
		},
		{
			name: "error case - internal server error",
			args: args{
				ctx:         context.Background(),
				contentType: echo.MIMEApplicationJSON,
				req: models.UpdateSubCategoryRequest{
					Name:        "RETAIL",
					Description: "Retail",
					Code:        "10000",
					Status:      "active",
				},
			},
			expectation: expectation{
				wantRes:  `{"status":"error","code":"DATABASE_ERROR","message":"database error"}`,
				wantCode: 500,
			},
			doMock: func(args args, expectation expectation) {
				testHelper.mockService.EXPECT().Update(args.ctx, models.UpdateSubCategory(args.req)).Return(models.UpdateSubCategory(args.req), models.GetErrMap(models.ErrKeyDatabaseError))
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

			r := httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/api/v1/sub-categories/%s", tt.args.req.Code), &b)
			r.Header.Set(echo.HeaderContentType, tt.args.contentType)
			w := httptest.NewRecorder()

			testHelper.router.NewContext(r, w)
			testHelper.router.ServeHTTP(w, r)

			require.Equal(t, tt.expectation.wantCode, w.Code)
			require.Equal(t, tt.expectation.wantRes, strings.Trim(w.Body.String(), "\n"))
		})
	}
}
