package types_test

import (
	"os"
	"testing"

	"github.com/kava-labs/kava/app"
)

func TestMain(m *testing.M) {
	app.SetSDKConfig()
	os.Exit(m.Run())
}
