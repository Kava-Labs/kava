package bep3_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/bep3"
	"github.com/stretchr/testify/suite"
)

type GenesisTestSuite struct {
	suite.Suite

	ctx    sdk.Context
	keeper bep3.Keeper
}

func (suite *GenesisTestSuite) TestGenesisState() {
	tApp := app.NewTestApp()

	type GenState func() app.GenesisState

	testCases := []struct {
		name       string
		genState   GenState
		expectPass bool
	}{
		{
			name: "default",
			genState: func() app.GenesisState {
				return NewBep3GenStateMulti()
			},
			expectPass: true,
		},
		// TODO:
		// {
		// 	name: "import atomic swaps and asset supplies",
		// 	genState: func() app.GenesisState {
		// 		gs := baseGenState()
		// 		_, addrs := app.GeneratePrivKeyAddressPairs(3)
		// 		atomicSwaps, assetSupply := atomicSwapsWithAssetSupply(addrs, "bnb")
		// 		gs.AtomicSwaps = atomicSwaps
		// 		gs.AssetSupplies = []sdk.Coin{assetSupply}
		// 		return app.GenesisState{"bep3": bep3.ModuleCdc.MustMarshalJSON(gs)}
		// 	},
		// 	expectPass: true,
		// },
		{
			name: "atomic swap total balance doesn't equal asset supply total balance",
			genState: func() app.GenesisState {
				gs := baseGenState()
				_, addrs := app.GeneratePrivKeyAddressPairs(3)
				atomicSwaps, _ := atomicSwapsWithAssetSupply(addrs, "bnb")
				gs.AtomicSwaps = atomicSwaps
				return app.GenesisState{"bep3": bep3.ModuleCdc.MustMarshalJSON(gs)}
			},
			expectPass: false,
		},
		{
			name: "minimum block lock below limit",
			genState: func() app.GenesisState {
				gs := baseGenState()
				gs.Params.MinBlockLock = 1
				return app.GenesisState{"bep3": bep3.ModuleCdc.MustMarshalJSON(gs)}
			},
			expectPass: false,
		},
		{
			name: "minimum block lock above limit",
			genState: func() app.GenesisState {
				gs := baseGenState()
				gs.Params.MinBlockLock = 500000
				return app.GenesisState{"bep3": bep3.ModuleCdc.MustMarshalJSON(gs)}
			},
			expectPass: false,
		},
		{
			name: "maximum block lock below limit",
			genState: func() app.GenesisState {
				gs := baseGenState()
				gs.Params.MaxBlockLock = 1
				return app.GenesisState{"bep3": bep3.ModuleCdc.MustMarshalJSON(gs)}
			},
			expectPass: false,
		},
		{
			name: "maximum block lock above limit",
			genState: func() app.GenesisState {
				gs := baseGenState()
				gs.Params.MaxBlockLock = 100000000
				return app.GenesisState{"bep3": bep3.ModuleCdc.MustMarshalJSON(gs)}
			},
			expectPass: false,
		},
		{
			name: "empty supported asset denom",
			genState: func() app.GenesisState {
				gs := baseGenState()
				gs.Params.SupportedAssets[0].Denom = ""
				return app.GenesisState{"bep3": bep3.ModuleCdc.MustMarshalJSON(gs)}
			},
			expectPass: false,
		},
		{
			name: "negative supported asset limit",
			genState: func() app.GenesisState {
				gs := baseGenState()
				gs.Params.SupportedAssets[0].Limit = i(-100)
				return app.GenesisState{"bep3": bep3.ModuleCdc.MustMarshalJSON(gs)}
			},
			expectPass: false,
		},
		// TODO:
		// {
		// 	name: "duplicate supported asset denom",
		// 	genState: func() app.GenesisState {
		// 		gs := baseGenState()
		// 		gs.Params.SupportedAssets[1].Denom = "bnb"
		// 		return app.GenesisState{"bep3": bep3.ModuleCdc.MustMarshalJSON(gs)}
		// 	},
		// 	expectPass: false,
		// },
		// TODO:
		// {
		// 	name: "invalid deputy address",
		// 	genState: func() app.GenesisState {
		// 		gs := baseGenState()
		// 		gs.Params.BnbDeputyAddress = sdk.AccAddress{}
		// 		return app.GenesisState{"bep3": bep3.ModuleCdc.MustMarshalJSON(gs)}
		// 	},
		// 	expectPass: false,
		// },
	}

	for _, tc := range testCases {
		if tc.expectPass {
			suite.NotPanics(func() {
				tApp.InitializeFromGenesisStates(tc.genState())
			})
		} else {
			suite.Panics(func() {
				tApp.InitializeFromGenesisStates(tc.genState())
			}, tc.name)
		}
	}
}

func TestGenesisTestSuite(t *testing.T) {
	suite.Run(t, new(GenesisTestSuite))
}
