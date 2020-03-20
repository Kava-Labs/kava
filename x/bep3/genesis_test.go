package bep3_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/bep3"
	"github.com/kava-labs/kava/x/bep3/types"
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
		{
			name: "import atomic swaps and asset supplies",
			genState: func() app.GenesisState {
				gs := baseGenState()
				_, addrs := app.GeneratePrivKeyAddressPairs(3)
				var swaps types.AtomicSwaps
				for i := 0; i < len(addrs); i++ {
					swap := atomicSwapFromAddress(addrs[i], i)
					swaps = append(swaps, swap)
				}
				gs.AtomicSwaps = swaps
				gs.AssetSupplies = cs(c("bnb", 7654321))
				return app.GenesisState{"bep3": bep3.ModuleCdc.MustMarshalJSON(gs)}
			},
			expectPass: true,
		},
		{
			name: "asset supply is not a supported asset",
			genState: func() app.GenesisState {
				gs := baseGenState()
				gs.AssetSupplies = []sdk.Coin{c("fake", 500000)}
				return app.GenesisState{"bep3": bep3.ModuleCdc.MustMarshalJSON(gs)}
			},
			expectPass: false,
		},
		{
			name: "asset supply above supply limit",
			genState: func() app.GenesisState {
				gs := baseGenState()
				assetParam, _ := suite.keeper.GetAssetByDenom(suite.ctx, "bnb")
				overLimitAmount := assetParam.Limit.Add(i(1))
				gs.AssetSupplies = []sdk.Coin{c("bnb", overLimitAmount.Int64())}
				return app.GenesisState{"bep3": bep3.ModuleCdc.MustMarshalJSON(gs)}
			},
			expectPass: false,
		},
		{
			name: "asset in swap supply above supply limit",
			genState: func() app.GenesisState {
				gs := baseGenState()
				_, addrs := app.GeneratePrivKeyAddressPairs(2)
				timestamp := ts(0)
				randomNumber, _ := types.GenerateSecureRandomNumber()
				randomNumberHash := types.CalculateRandomHash(randomNumber.Bytes(), timestamp)
				assetParam, _ := suite.keeper.GetAssetByDenom(suite.ctx, "bnb")
				overLimitAmount := assetParam.Limit.Add(i(1))
				swap := types.NewAtomicSwap(cs(c("bnb", overLimitAmount.Int64())), randomNumberHash,
					int64(360), timestamp, addrs[0], addrs[1], TestSenderOtherChain,
					TestRecipientOtherChain, 0, types.Open, true)

				gs.AtomicSwaps = types.AtomicSwaps{swap}
				return app.GenesisState{"bep3": bep3.ModuleCdc.MustMarshalJSON(gs)}
			},
			expectPass: false,
		},
		{
			name: "atomic swap asset type is unsupported",
			genState: func() app.GenesisState {
				gs := baseGenState()
				_, addrs := app.GeneratePrivKeyAddressPairs(2)
				timestamp := ts(0)
				randomNumber, _ := types.GenerateSecureRandomNumber()
				randomNumberHash := types.CalculateRandomHash(randomNumber.Bytes(), timestamp)
				swap := types.NewAtomicSwap(cs(c("fake", 500000)), randomNumberHash,
					int64(360), timestamp, addrs[0], addrs[1], TestSenderOtherChain,
					TestRecipientOtherChain, 0, types.Open, true)

				gs.AtomicSwaps = types.AtomicSwaps{swap}
				return app.GenesisState{"bep3": bep3.ModuleCdc.MustMarshalJSON(gs)}
			},
			expectPass: false,
		},
		{
			name: "atomic swap status is invalid",
			genState: func() app.GenesisState {
				gs := baseGenState()
				_, addrs := app.GeneratePrivKeyAddressPairs(2)
				timestamp := ts(0)
				randomNumber, _ := types.GenerateSecureRandomNumber()
				randomNumberHash := types.CalculateRandomHash(randomNumber.Bytes(), timestamp)
				swap := types.NewAtomicSwap(cs(c("bnb", 5000)), randomNumberHash,
					int64(360), timestamp, addrs[0], addrs[1], TestSenderOtherChain,
					TestRecipientOtherChain, 0, types.NULL, true)

				gs.AtomicSwaps = types.AtomicSwaps{swap}
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
		{
			name: "duplicate supported asset denom",
			genState: func() app.GenesisState {
				gs := baseGenState()
				gs.Params.SupportedAssets[1].Denom = "bnb"
				return app.GenesisState{"bep3": bep3.ModuleCdc.MustMarshalJSON(gs)}
			},
			expectPass: false,
		},
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
