package pricefeed_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/pricefeed"

	"github.com/stretchr/testify/suite"
)

type GenesisTestSuite struct {
	suite.Suite

	ctx    sdk.Context
	keeper pricefeed.Keeper
}

func (suite *GenesisTestSuite) TestValidGenState() {
	tApp := app.NewTestApp()

	suite.NotPanics(func() {
		tApp.InitializeFromGenesisStates(
			NewPricefeedGenStateMulti(),
		)
	})
	_, addrs := app.GeneratePrivKeyAddressPairs(10)

	suite.NotPanics(func() {
		tApp.InitializeFromGenesisStates(
			NewPricefeedGenStateWithOracles(addrs),
		)
	})
}

func TestGenesisTestSuite(t *testing.T) {
	suite.Run(t, new(GenesisTestSuite))
}
