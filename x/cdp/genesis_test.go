package cdp_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/cdp"

	"github.com/stretchr/testify/suite"
)

type GenesisTestSuite struct {
	suite.Suite

	ctx    sdk.Context
	keeper cdp.Keeper
}

func (suite *GenesisTestSuite) TestInvalidGenState() {
	tApp := app.NewTestApp()
	for _, gs := range badGenStates() {

		appGS := app.GenesisState{"cdp": cdp.ModuleCdc.MustMarshalJSON(gs.Genesis)}
		suite.Panics(func() {
			tApp.InitializeFromGenesisStates(
				NewPricefeedGenStateMulti(),
				appGS,
			)
		}, gs.Reason)
	}
}

func (suite *GenesisTestSuite) TestValidGenState() {
	tApp := app.NewTestApp()

	suite.NotPanics(func() {
		tApp.InitializeFromGenesisStates(
			NewPricefeedGenStateMulti(),
			NewCDPGenStateMulti(),
		)
	})

	cdpGS := NewCDPGenStateMulti()
	gs := cdp.GenesisState{}
	cdp.ModuleCdc.UnmarshalJSON(cdpGS["cdp"], &gs)
	gs.CDPs = cdps()
	gs.StartingCdpID = uint64(5)
	appGS := app.GenesisState{"cdp": cdp.ModuleCdc.MustMarshalJSON(gs)}
	suite.NotPanics(func() {
		tApp.InitializeFromGenesisStates(
			NewPricefeedGenStateMulti(),
			appGS,
		)
	})

}

func TestGenesisTestSuite(t *testing.T) {
	suite.Run(t, new(GenesisTestSuite))
}
