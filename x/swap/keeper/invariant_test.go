package keeper_test

import (
	"testing"

	//"github.com/kava-labs/kava/x/swap"
	"github.com/kava-labs/kava/x/swap/testutil"
	//"github.com/kava-labs/kava/x/swap/types"
	"github.com/stretchr/testify/suite"
	//sdk "github.com/cosmos/cosmos-sdk/types"
)

type invariantTestSuite struct {
	testutil.Suite
}

func TestGenesisTestSuite(t *testing.T) {
	suite.Run(t, new(invariantTestSuite))
}
