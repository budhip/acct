package account

import (
	"os"
	"testing"

	"bitbucket.org/Amartha/go-accounting/internal/services/mock"
	xlog "bitbucket.org/Amartha/go-x/log"

	"github.com/labstack/echo/v4"
	"go.uber.org/mock/gomock"
)

func TestMain(m *testing.M) {
	xlog.InitForTest()
	os.Exit(m.Run())
}

type testPublisherHelper struct {
	router               *echo.Echo
	mockCtrl             *gomock.Controller
	mockPublisherService *mock.MockPublisherService
}

func publisherTestHelper(t *testing.T) testPublisherHelper {
	t.Helper()

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockPublisherService := mock.NewMockPublisherService(mockCtrl)

	app := echo.New()
	v1Group := app.Group("/api/v1")
	New(v1Group, mockPublisherService)

	return testPublisherHelper{
		router:               app,
		mockCtrl:             mockCtrl,
		mockPublisherService: mockPublisherService,
	}
}
