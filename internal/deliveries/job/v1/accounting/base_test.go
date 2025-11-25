package accounting

import (
	"os"
	"testing"

	"bitbucket.org/Amartha/go-accounting/internal/services/mock"
	xlog "bitbucket.org/Amartha/go-x/log"

	"go.uber.org/mock/gomock"
)

type testAccountingJobHelper struct {
	mockCtrl              *gomock.Controller
	mockAccountingService *mock.MockAccountingService
}

func accountingTestHelper(t *testing.T) testAccountingJobHelper {
	t.Helper()

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockAccountingService := mock.NewMockAccountingService(mockCtrl)

	Routes(mockAccountingService)

	return testAccountingJobHelper{
		mockCtrl:              mockCtrl,
		mockAccountingService: mockAccountingService,
	}
}

func TestMain(m *testing.M) {
	xlog.InitForTest()
	os.Exit(m.Run())
}
