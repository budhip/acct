package health

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"bitbucket.org/Amartha/go-accounting/internal/services/mock"
	xlog "bitbucket.org/Amartha/go-x/log"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

type testHelper struct {
	mockCtrl          *gomock.Controller
	router            *echo.Echo
	mockHealthService *mock.MockHealthService
}

func testHealthHelper(t *testing.T) testHelper {
	t.Helper()
	t.Parallel()

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockHealthService := mock.NewMockHealthService(mockCtrl)

	app := echo.New()
	apiGroup := app.Group("/api")

	New(apiGroup, mockHealthService)

	return testHelper{
		mockCtrl:          mockCtrl,
		router:            app,
		mockHealthService: mockHealthService,
	}
}

func TestMain(m *testing.M) {
	xlog.InitForTest()
	os.Exit(m.Run())
}

func Test_Handler_readinessCheck(t *testing.T) {
	testHelper := testHealthHelper(t)

	type args struct{}
	type mockData struct {
		wantRes  string
		wantCode int
	}
	tests := []struct {
		name      string
		urlCalled string
		args      args
		mockData  mockData
		doMock    func(args args, mockData mockData)
	}{
		{
			name:      "success",
			urlCalled: "/api/health/readiness",
			args:      args{},
			mockData: mockData{
				wantRes:  `{"kind":"health","status":"server is up and running"}`,
				wantCode: 200,
			},
			doMock: func(args args, mockData mockData) {
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if tt.doMock != nil {
				tt.doMock(tt.args, tt.mockData)
			}
			r := httptest.NewRequest(http.MethodGet, tt.urlCalled, nil)
			r.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			w := httptest.NewRecorder()

			testHelper.router.NewContext(r, w)
			testHelper.router.ServeHTTP(w, r)

			require.Equal(t, tt.mockData.wantCode, w.Code)
			require.Equal(t, tt.mockData.wantRes, strings.Trim(w.Body.String(), "\n"))
		})
	}
}

func Test_Handler_livenessCheck(t *testing.T) {
	testHelper := testHealthHelper(t)

	type mockData struct {
		wantRes  string
		wantCode int
	}
	tests := []struct {
		name     string
		mockData mockData
		doMock   func(mockData mockData)
	}{
		{
			name: "success",
			mockData: mockData{
				wantRes:  `{"kind":"health","status":{"mysql":"mysql is up and running","redis":"redis is up and running"}}`,
				wantCode: 200,
			},
			doMock: func(mockData mockData) {
				testHelper.mockHealthService.EXPECT().GetHealth(gomock.Any()).Return(map[string]string{
					"redis": "redis is up and running",
					"mysql": "mysql is up and running",
				})
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if tt.doMock != nil {
				tt.doMock(tt.mockData)
			}
			r := httptest.NewRequest(http.MethodGet, "/api/health/liveness", nil)
			r.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			w := httptest.NewRecorder()

			testHelper.router.NewContext(r, w)
			testHelper.router.ServeHTTP(w, r)

			require.Equal(t, tt.mockData.wantCode, w.Code)
			require.Equal(t, tt.mockData.wantRes, strings.Trim(w.Body.String(), "\n"))
		})
	}
}
