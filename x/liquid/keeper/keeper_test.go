package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/kava-labs/kava/x/liquid/testutil"
)

// Test suite used for all keeper tests
type KeeperTestSuite struct {
	testutil.Suite
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}
