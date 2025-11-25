package godbledger

import (
	"os"
	"testing"

	xlog "bitbucket.org/Amartha/go-x/log"
)

func TestMain(m *testing.M) {
	xlog.InitForTest()
	os.Exit(m.Run())
}
