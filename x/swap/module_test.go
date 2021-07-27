package swap_test

import (
	"testing"

	"github.com/kava-labs/kava/x/swap"
	"github.com/kava-labs/kava/x/swap/testutil"

	"github.com/cosmos/cosmos-sdk/x/crisis"
	"github.com/stretchr/testify/suite"
)

type moduleTestSuite struct {
	testutil.Suite
	crisisKeeper crisis.Keeper
}

func (suite *moduleTestSuite) SetupTest() {
	suite.Suite.SetupTest()
	suite.crisisKeeper = suite.App.GetCrisisKeeper()
}

func (suite *moduleTestSuite) TestRegesterInviarants() {
	swapRoutes := []string{}

	for _, route := range suite.crisisKeeper.Routes() {
		if route.ModuleName == swap.ModuleName {
			swapRoutes = append(swapRoutes, route.Route)
		}
	}

	suite.Contains(swapRoutes, "pool-records")
	suite.Contains(swapRoutes, "share-records")
	suite.Contains(swapRoutes, "pool-reserves")
	suite.Contains(swapRoutes, "pool-shares")
}

func TestModuleTestSuite(t *testing.T) {
	suite.Run(t, new(moduleTestSuite))
}
