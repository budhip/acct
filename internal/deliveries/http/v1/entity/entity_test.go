package entity

import (
	"bytes"
	"context"
	"encoding/json"
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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

type testEntityHelper struct {
	router      *echo.Echo
	mockCtrl    *gomock.Controller
	mockService *mock.MockEntityService
}

func entityTestHelper(t *testing.T) testEntityHelper {
	t.Helper()

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockSvc := mock.NewMockEntityService(mockCtrl)

	app := echo.New()
	v1Group := app.Group("/api/v1")
	New(v1Group, mockSvc)

	return testEntityHelper{
		router:      app,
		mockCtrl:    mockCtrl,
		mockService: mockSvc,
	}
}

func TestMain(m *testing.M) {
	xlog.InitForTest()
	os.Exit(m.Run())
}

func Test_Handler_createEntity(t *testing.T) {
	testHelper := entityTestHelper(t)

	type args struct {
		ctx         context.Context
		contentType string
		req         models.CreateEntityRequest
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
			urlCalled: "/api/v1/entities",
			args: args{
				ctx:         context.Background(),
				contentType: echo.MIMEApplicationJSON,
				req: models.CreateEntityRequest{
					Code:        "001",
					Name:        "test",
					Description: "",
					Status:      models.StatusActive,
				},
			},
			expectation: expectation{
				wantRes:  `{"kind":"entity","code":"001","name":"test","description":"","status":""}`,
				wantCode: 201,
			},
			doMock: func(args args, expectation expectation) {
				testHelper.mockService.EXPECT().Create(args.ctx, models.CreateEntityIn(args.req)).Return(&models.Entity{
					Code: args.req.Code,
					Name: args.req.Name,
				}, nil)
			},
		},
		{
			name:      "error contentType",
			urlCalled: "/api/v1/entities",
			args: args{
				ctx:         context.Background(),
				contentType: "",
				req: models.CreateEntityRequest{
					Code:        "001",
					Name:        "test",
					Description: "",
				},
			},
			expectation: expectation{
				wantRes:  `{"status":"error","code":400,"message":"code=415, message=Unsupported Media Type"}`,
				wantCode: 400,
			},
			doMock: func(args args, expectation expectation) {},
		},
		{
			name:      "test error validating request",
			urlCalled: "/api/v1/entities",
			args: args{
				ctx:         context.Background(),
				contentType: echo.MIMEApplicationJSON,
				req: models.CreateEntityRequest{
					Code:   "12",
					Status: models.StatusActive,
				},
			},
			expectation: expectation{
				wantRes:  `{"status":"error","message":"validation failed","errors":[{"code":"INVALID_LENGTH","field":"code","message":"field must be at least 3 characters"},{"code":"MISSING_FIELD","field":"name","message":"field is missing"}]}`,
				wantCode: 422,
			},
		},
		{
			name:      "test error",
			urlCalled: "/api/v1/entities",
			args: args{
				ctx:         context.Background(),
				contentType: echo.MIMEApplicationJSON,
				req: models.CreateEntityRequest{
					Code:        "001",
					Name:        "test",
					Description: "",
					Status:      models.StatusActive,
				},
			},
			expectation: expectation{
				wantRes:  `{"status":"error","code":"DATA_IS_EXIST","message":"data is exist"}`,
				wantCode: 409,
			},
			doMock: func(args args, expectation expectation) {
				testHelper.mockService.EXPECT().Create(args.ctx, models.CreateEntityIn(args.req)).Return(&models.Entity{}, models.ErrDataExist)
			},
		},
		{
			name:      "test error",
			urlCalled: "/api/v1/entities",
			args: args{
				ctx:         context.Background(),
				contentType: echo.MIMEApplicationJSON,
				req: models.CreateEntityRequest{
					Code:        "001",
					Name:        "test",
					Description: "",
					Status:      models.StatusActive,
				},
			},
			expectation: expectation{
				wantRes:  `{"status":"error","code":500,"message":"assert.AnError general error for testing"}`,
				wantCode: 500,
			},
			doMock: func(args args, expectation expectation) {
				testHelper.mockService.EXPECT().Create(args.ctx, models.CreateEntityIn(args.req)).Return(&models.Entity{}, assert.AnError)
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

func Test_Handler_getAllEntity(t *testing.T) {
	testHelper := entityTestHelper(t)

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
			name: "success get all entity",
			expectation: Expectation{
				wantRes:  `{"kind":"collection","contents":[{"kind":"entity","code":"666","name":"ENT","description":"ini entity","status":"active"}],"total_rows":1}`,
				wantCode: 200,
			},
			doMock: func() {
				testHelper.mockService.EXPECT().GetAll(gomock.AssignableToTypeOf(context.Background())).
					Return(&[]models.Entity{{
						Code:        "666",
						Name:        "ENT",
						Description: "ini entity",
						Status:      models.StatusActive,
					}}, nil)
			},
		},
		{
			name: "failed to get data entity",
			expectation: Expectation{
				wantRes:  `{"status":"error","code":500,"message":"assert.AnError general error for testing"}`,
				wantCode: 500,
			},
			doMock: func() {
				testHelper.mockService.EXPECT().
					GetAll(gomock.AssignableToTypeOf(context.Background())).
					Return(&[]models.Entity{}, assert.AnError)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.doMock != nil {
				tt.doMock()
			}

			var b bytes.Buffer

			r := httptest.NewRequest(http.MethodGet, "/api/v1/entities", &b)
			r.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			w := httptest.NewRecorder()

			testHelper.router.NewContext(r, w)
			testHelper.router.ServeHTTP(w, r)

			require.Equal(t, tt.expectation.wantCode, w.Code)
			require.Equal(t, tt.expectation.wantRes, strings.Trim(w.Body.String(), "\n"))
		})
	}
}

func Test_Handler_updateEntity(t *testing.T) {
	testHelper := entityTestHelper(t)

	type args struct {
		ctx         context.Context
		contentType string
		req         models.UpdateEntityRequest
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
			name:      "success update entity",
			urlCalled: "/api/v1/entities/001",
			args: args{
				ctx:         context.Background(),
				contentType: echo.MIMEApplicationJSON,
				req: models.UpdateEntityRequest{
					Code:        "001",
					Name:        "test",
					Description: "",
					Status:      "active",
				},
			},
			expectation: expectation{
				wantRes:  `{"kind":"entity","code":"001","name":"test","status":"active"}`,
				wantCode: 200,
			},
			doMock: func(args args, expectation expectation) {
				testHelper.mockService.EXPECT().Update(args.ctx, models.UpdateEntity{
					Name:        args.req.Name,
					Description: args.req.Description,
					Code:        args.req.Code,
					Status:      args.req.Status,
				}).Return(models.UpdateEntity{
					Name:        args.req.Name,
					Description: args.req.Description,
					Code:        args.req.Code,
					Status:      args.req.Status,
				}, nil)
			},
		},
		{
			name:      "error internal server",
			urlCalled: "/api/v1/entities/001",
			args: args{
				ctx:         context.Background(),
				contentType: echo.MIMEApplicationJSON,
				req: models.UpdateEntityRequest{
					Code:        "001",
					Name:        "test",
					Description: "",
					Status:      "active",
				},
			},
			expectation: expectation{
				wantRes:  `{"status":"error","code":500,"message":"internal server error"}`,
				wantCode: 500,
			},
			doMock: func(args args, expectation expectation) {
				testHelper.mockService.EXPECT().Update(args.ctx, models.UpdateEntity{
					Name:        args.req.Name,
					Description: args.req.Description,
					Code:        args.req.Code,
					Status:      args.req.Status,
				}).Return(models.UpdateEntity{}, models.ErrInternalServerError)
			},
		},
		{
			name:      "error entity not found",
			urlCalled: "/api/v1/entities/001",
			args: args{
				ctx:         context.Background(),
				contentType: echo.MIMEApplicationJSON,
				req: models.UpdateEntityRequest{
					Code:        "001",
					Name:        "test",
					Description: "",
					Status:      "active",
				},
			},
			expectation: expectation{
				wantRes:  `{"status":"error","code":"DATA_NOT_FOUND","message":"entity code not found"}`,
				wantCode: 404,
			},
			doMock: func(args args, expectation expectation) {
				testHelper.mockService.EXPECT().Update(args.ctx, models.UpdateEntity{
					Name:        args.req.Name,
					Description: args.req.Description,
					Code:        args.req.Code,
					Status:      args.req.Status,
				}).Return(models.UpdateEntity{}, models.GetErrMap(models.ErrKeyEntityCodeNotFound))
			},
		},
		{
			name:      "error validation",
			urlCalled: "/api/v1/entities/001",
			args: args{
				ctx:         context.Background(),
				contentType: echo.MIMEApplicationJSON,
				req: models.UpdateEntityRequest{
					Code:        "001",
					Name:        "tes_test",
					Description: "",
					Status:      "active",
				},
			},
			expectation: expectation{
				wantRes:  `{"status":"error","message":"validation failed","errors":[{"code":"INVALID_VALUES","field":"name","message":"input field no special characters"}]}`,
				wantCode: 422,
			},
			doMock: func(args args, expectation expectation) {
				testHelper.mockService.EXPECT().Update(args.ctx, models.UpdateEntity{
					Name:        args.req.Name,
					Description: args.req.Description,
					Code:        args.req.Code,
					Status:      args.req.Status,
				}).Return(models.UpdateEntity{}, models.GetErrMap(models.ErrKeyEntityCodeNotFound))
			},
		},
		{
			name:      "error contentType",
			urlCalled: "/api/v1/entities/001",
			args: args{
				ctx:         context.Background(),
				contentType: "",
				req: models.UpdateEntityRequest{
					Code:        "001",
					Name:        "tes_test",
					Description: "",
					Status:      "active",
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

func Test_Handler_getEntityByParam(t *testing.T) {
	testHelper := entityTestHelper(t)

	type Expectation struct {
		wantRes  string
		wantCode int
	}
	type args struct {
		ctx         context.Context
		contentType string
		req         url.Values
	}
	tests := []struct {
		name        string
		urlCalled   string
		args        args
		expectation Expectation
		doMock      func()
	}{
		{
			name:      "success get entity",
			urlCalled: "/api/v1/entities/search?entityCode=002&name=ANR",
			args: args{
				ctx:         context.Background(),
				contentType: echo.MIMEApplicationJSON,
				req: url.Values{
					"entityCode": []string{"002"},
					"name":       []string{"ANR"},
				},
			},
			expectation: Expectation{
				wantRes:  `{"kind":"entity","code":"002","name":"ANR","description":"PT. Amartha Nusantara Raya","status":"active"}`,
				wantCode: 200,
			},
			doMock: func() {
				testHelper.mockService.EXPECT().GetByParam(gomock.Any(), gomock.Any()).
					Return(&models.Entity{
						Code:        "002",
						Name:        "ANR",
						Description: "PT. Amartha Nusantara Raya",
						Status:      models.StatusActive,
					}, nil)
			},
		},
		{
			name:      "error validation entity",
			urlCalled: "/api/v1/entities/search",
			args: args{
				ctx:         context.Background(),
				contentType: echo.MIMEApplicationJSON,
				req:         url.Values{},
			},
			expectation: Expectation{
				wantRes:  `{"status":"error","message":"validation failed","errors":[{"code":"MISSING_FIELD","field":"entityCode","message":"required fields at least entityCode"},{"code":"MISSING_FIELD","field":"name","message":"required fields at least name"}]}`,
				wantCode: 422,
			},
			doMock: func() {},
		},
		{
			name:      "error get entity not found",
			urlCalled: "/api/v1/entities/search?entityCode=002&name=AN",
			args: args{
				ctx:         context.Background(),
				contentType: echo.MIMEApplicationJSON,
				req: url.Values{
					"entityCode": []string{"002"},
					"name":       []string{"AN"},
				},
			},
			expectation: Expectation{
				wantRes:  `{"status":"error","code":"DATA_NOT_FOUND","message":"data not found"}`,
				wantCode: 404,
			},
			doMock: func() {
				testHelper.mockService.EXPECT().GetByParam(gomock.Any(), gomock.Any()).
					Return(&models.Entity{}, models.GetErrMap(models.ErrKeyDataNotFound))
			},
		},
		{
			name:      "error get entity internal server error",
			urlCalled: "/api/v1/entities/search?entityCode=002&name=AN",
			args: args{
				ctx:         context.Background(),
				contentType: echo.MIMEApplicationJSON,
				req: url.Values{
					"entityCode": []string{"002"},
					"name":       []string{"ANR"},
				},
			},
			expectation: Expectation{
				wantRes:  `{"status":"error","code":500,"message":"internal server error"}`,
				wantCode: 500,
			},
			doMock: func() {
				testHelper.mockService.EXPECT().GetByParam(gomock.Any(), gomock.Any()).
					Return(&models.Entity{}, models.ErrInternalServerError)
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if tt.doMock != nil {
				tt.doMock()
			}

			var b bytes.Buffer
			err := json.NewEncoder(&b).Encode(tt.args.req)
			require.NoError(t, err)

			r := httptest.NewRequest(http.MethodGet, tt.urlCalled, nil)
			r.Header.Set(echo.HeaderContentType, tt.args.contentType)
			w := httptest.NewRecorder()

			testHelper.router.NewContext(r, w)
			testHelper.router.ServeHTTP(w, r)

			require.Equal(t, tt.expectation.wantCode, w.Code)
			require.Equal(t, tt.expectation.wantRes, strings.Trim(w.Body.String(), "\n"))
		})
	}
}
