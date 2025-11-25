package accounting

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

type testAccountHelper struct {
	router                *echo.Echo
	mockCtrl              *gomock.Controller
	mockAccountingService *mock.MockAccountingService
}

func accountingTestHelper(t *testing.T) testAccountHelper {
	t.Helper()

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockAccountingSvc := mock.NewMockAccountingService(mockCtrl)

	app := echo.New()
	v2Group := app.Group("/api/v2")
	New(v2Group, mockAccountingSvc)

	return testAccountHelper{
		router:                app,
		mockCtrl:              mockCtrl,
		mockAccountingService: mockAccountingSvc,
	}
}
