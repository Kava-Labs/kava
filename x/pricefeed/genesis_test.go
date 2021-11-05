package pricefeed_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/pricefeed/keeper"

	"github.com/stretchr/testify/suite"
)

type GenesisTestSuite struct {
	suite.Suite

	ctx    sdk.Context
	keeper keeper.Keeper
}

func (suite *GenesisTestSuite) TestValidGenState() {
	tApp := app.NewTestApp()

	suite.NotPanics(func() {
		tApp.InitializeFromGenesisStates(
			NewPricefeedGenStateMulti(),
		)
	})
	_, addrs := app.GeneratePrivKeyAddressPairs(10)

	// Must create a new TestApp or InitChain will panic with index already set
	tApp = app.NewTestApp()
	suite.NotPanics(func() {
		tApp.InitializeFromGenesisStates(
			NewPricefeedGenStateWithOracles(addrs),
		)
	})
}

func TestGenesisTestSuite(t *testing.T) {
	suite.Run(t, new(GenesisTestSuite))
}
