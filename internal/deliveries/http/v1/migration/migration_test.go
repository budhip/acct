package migration

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

type testMigrationHelper struct {
	router      *echo.Echo
	mockCtrl    *gomock.Controller
	mockService *mock.MockMigrationService
}

func migrationTestHelper(t *testing.T) testMigrationHelper {
	t.Helper()

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockSvc := mock.NewMockMigrationService(mockCtrl)

	app := echo.New()
	v1Group := app.Group("/api/v1")
	New(v1Group, mockSvc)

	return testMigrationHelper{
		router:      app,
		mockCtrl:    mockCtrl,
		mockService: mockSvc,
	}
}

func TestMain(m *testing.M) {
	xlog.InitForTest()
	os.Exit(m.Run())
}

func Test_Handler_MigrationJournalLoadBuckets(t *testing.T) {
	testHelper := migrationTestHelper(t)

	type args struct {
		ctx         context.Context
		contentType string
		req         models.MigrationBucketsJournalLoadRequest
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
			urlCalled: "/api/v1/migrations/buckets/journal-load",
			args: args{
				ctx:         context.Background(),
				contentType: echo.MIMEApplicationJSON,
				req:         models.MigrationBucketsJournalLoadRequest{SubFolder: "test"},
			},
			expectation: expectation{
				wantRes:  `{"kind":"file","status":"Migration process started"}`,
				wantCode: 201,
			},
			doMock: func(args args, expectation expectation) {
				testHelper.mockService.EXPECT().BucketsJournalLoad(args.ctx, args.req).Return(nil)
			},
		},
		{
			name:      "error contentType",
			urlCalled: "/api/v1/migrations/buckets/journal-load",
			args: args{
				ctx:         context.Background(),
				contentType: "",
				req:         models.MigrationBucketsJournalLoadRequest{},
			},
			expectation: expectation{
				wantRes:  `{"status":"error","code":400,"message":"code=415, message=Unsupported Media Type"}`,
				wantCode: 400,
			},
			doMock: func(args args, expectation expectation) {},
		},
		{
			name:      "test error validating request",
			urlCalled: "/api/v1/migrations/buckets/journal-load",
			args: args{
				ctx:         context.Background(),
				contentType: echo.MIMEApplicationJSON,
				req:         models.MigrationBucketsJournalLoadRequest{},
			},
			expectation: expectation{
				wantRes:  `{"status":"error","message":"validation failed","errors":[{"code":"UNKNOW","field":"subFolder","message":"required"}]}`,
				wantCode: 422,
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
