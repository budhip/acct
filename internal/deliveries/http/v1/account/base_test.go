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

type testAccountHelper struct {
	router             *echo.Echo
	mockCtrl           *gomock.Controller
	mockAccountService *mock.MockAccountService
}

func accountTestHelper(t *testing.T) testAccountHelper {
	t.Helper()

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockAccountSvc := mock.NewMockAccountService(mockCtrl)

	app := echo.New()
	v1Group := app.Group("/api/v1")
	New(v1Group, mockAccountSvc)

	return testAccountHelper{
		router:             app,
		mockCtrl:           mockCtrl,
		mockAccountService: mockAccountSvc,
	}
}
