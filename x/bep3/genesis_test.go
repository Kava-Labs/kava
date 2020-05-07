package bep3_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"

	abci "github.com/tendermint/tendermint/abci/types"
	tmtime "github.com/tendermint/tendermint/types/time"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/bep3"
)

type GenesisTestSuite struct {
	suite.Suite

	app    app.TestApp
	ctx    sdk.Context
	keeper bep3.Keeper
	addrs  []sdk.AccAddress
}

func (suite *GenesisTestSuite) SetupTest() {
	tApp := app.NewTestApp()
	suite.ctx = tApp.NewContext(true, abci.Header{Height: 1, Time: tmtime.Now()})
	suite.keeper = tApp.GetBep3Keeper()
	suite.app = tApp

	_, addrs := app.GeneratePrivKeyAddressPairs(3)
	suite.addrs = addrs
}

func (suite *GenesisTestSuite) TestGenesisState() {

	type GenState func() app.GenesisState

	testCases := []struct {
		name       string
		genState   GenState
		expectPass bool
	}{
		{
			name: "default",
			genState: func() app.GenesisState {
				return NewBep3GenStateMulti(suite.addrs[0])
			},
			expectPass: true,
		},
		{
			name: "import atomic swaps and asset supplies",
			genState: func() app.GenesisState {
				gs := baseGenState(suite.addrs[0])
				_, addrs := app.GeneratePrivKeyAddressPairs(3)
				var swaps bep3.AtomicSwaps
				var supplies bep3.AssetSupplies
				for i := 0; i < 3; i++ {
					swap, supply := loadSwapAndSupply(addrs[i], i)
					swaps = append(swaps, swap)
					supplies = append(supplies, supply)
				}
				gs.AtomicSwaps = swaps
				gs.AssetSupplies = supplies
				return app.GenesisState{"bep3": bep3.ModuleCdc.MustMarshalJSON(gs)}
			},
			expectPass: true,
		},
		{
			name: "incoming supply doesn't match amount in incoming atomic swaps",
			genState: func() app.GenesisState {
				gs := baseGenState(suite.addrs[0])
				_, addrs := app.GeneratePrivKeyAddressPairs(1)
				swap, _ := loadSwapAndSupply(addrs[0], 2)
				gs.AtomicSwaps = bep3.AtomicSwaps{swap}
				return app.GenesisState{"bep3": bep3.ModuleCdc.MustMarshalJSON(gs)}
			},
			expectPass: false,
		},
		{
			name: "current supply above limit",
			genState: func() app.GenesisState {
				gs := baseGenState(suite.addrs[0])
				assetParam, _ := suite.keeper.GetAssetByDenom(suite.ctx, "bnb")
				gs.AssetSupplies = bep3.AssetSupplies{
					bep3.AssetSupply{
						Denom:          "bnb",
						IncomingSupply: c("bnb", 0),
						OutgoingSupply: c("bnb", 0),
						CurrentSupply:  c("bnb", assetParam.Limit.Add(i(1)).Int64()),
						SupplyLimit:    c("bnb", assetParam.Limit.Int64()),
					},
				}
				return app.GenesisState{"bep3": bep3.ModuleCdc.MustMarshalJSON(gs)}
			},
			expectPass: false,
		},
		{
			name: "incoming supply above limit",
			genState: func() app.GenesisState {
				gs := baseGenState(suite.addrs[0])
				// Set up overlimit amount
				assetParam, _ := suite.keeper.GetAssetByDenom(suite.ctx, "bnb")
				overLimitAmount := assetParam.Limit.Add(i(1))

				// Set up an atomic swap with amount equal to the currently asset supply
				_, addrs := app.GeneratePrivKeyAddressPairs(2)
				timestamp := ts(0)
				randomNumber, _ := bep3.GenerateSecureRandomNumber()
				randomNumberHash := bep3.CalculateRandomHash(randomNumber, timestamp)
				swap := bep3.NewAtomicSwap(cs(c("bnb", overLimitAmount.Int64())), randomNumberHash,
					uint64(360), timestamp, suite.addrs[0], addrs[1], TestSenderOtherChain,
					TestRecipientOtherChain, 0, bep3.Open, true, bep3.Incoming)
				gs.AtomicSwaps = bep3.AtomicSwaps{swap}

				// Set up asset supply with overlimit current supply
				gs.AssetSupplies = bep3.AssetSupplies{
					bep3.AssetSupply{
						Denom:          "bnb",
						IncomingSupply: c("bnb", assetParam.Limit.Add(i(1)).Int64()),
						OutgoingSupply: c("bnb", 0),
						CurrentSupply:  c("bnb", 0),
						SupplyLimit:    c("bnb", assetParam.Limit.Int64()),
					},
				}
				return app.GenesisState{"bep3": bep3.ModuleCdc.MustMarshalJSON(gs)}
			},
			expectPass: false,
		},
		{
			name: "incoming supply + current supply above limit",
			genState: func() app.GenesisState {
				gs := baseGenState(suite.addrs[0])
				// Set up overlimit amount
				assetParam, _ := suite.keeper.GetAssetByDenom(suite.ctx, "bnb")
				halfLimit := assetParam.Limit.Int64() / 2
				overHalfLimit := halfLimit + 1

				// Set up an atomic swap with amount equal to the currently asset supply
				_, addrs := app.GeneratePrivKeyAddressPairs(2)
				timestamp := ts(0)
				randomNumber, _ := bep3.GenerateSecureRandomNumber()
				randomNumberHash := bep3.CalculateRandomHash(randomNumber, timestamp)
				swap := bep3.NewAtomicSwap(cs(c("bnb", halfLimit)), randomNumberHash,
					uint64(360), timestamp, suite.addrs[0], addrs[1], TestSenderOtherChain,
					TestRecipientOtherChain, 0, bep3.Open, true, bep3.Incoming)
				gs.AtomicSwaps = bep3.AtomicSwaps{swap}

				// Set up asset supply with overlimit current supply
				gs.AssetSupplies = bep3.AssetSupplies{
					bep3.AssetSupply{
						Denom:          "bnb",
						IncomingSupply: c("bnb", halfLimit),
						OutgoingSupply: c("bnb", 0),
						CurrentSupply:  c("bnb", overHalfLimit),
						SupplyLimit:    c("bnb", assetParam.Limit.Int64()),
					},
				}
				return app.GenesisState{"bep3": bep3.ModuleCdc.MustMarshalJSON(gs)}
			},
			expectPass: false,
		},
		{
			name: "asset supply denom is not a supported asset",
			genState: func() app.GenesisState {
				gs := baseGenState(suite.addrs[0])
				gs.AssetSupplies = bep3.AssetSupplies{
					bep3.AssetSupply{
						Denom:          "fake",
						IncomingSupply: c("fake", 0),
						OutgoingSupply: c("fake", 0),
						CurrentSupply:  c("fake", 0),
						SupplyLimit:    c("fake", StandardSupplyLimit.Int64()),
					},
				}
				return app.GenesisState{"bep3": bep3.ModuleCdc.MustMarshalJSON(gs)}
			},
			expectPass: false,
		},
		{
			name: "atomic swap asset type is unsupported",
			genState: func() app.GenesisState {
				gs := baseGenState(suite.addrs[0])
				_, addrs := app.GeneratePrivKeyAddressPairs(2)
				timestamp := ts(0)
				randomNumber, _ := bep3.GenerateSecureRandomNumber()
				randomNumberHash := bep3.CalculateRandomHash(randomNumber, timestamp)
				swap := bep3.NewAtomicSwap(cs(c("fake", 500000)), randomNumberHash,
					uint64(360), timestamp, suite.addrs[0], addrs[1], TestSenderOtherChain,
					TestRecipientOtherChain, 0, bep3.Open, true, bep3.Incoming)

				gs.AtomicSwaps = bep3.AtomicSwaps{swap}
				return app.GenesisState{"bep3": bep3.ModuleCdc.MustMarshalJSON(gs)}
			},
			expectPass: false,
		},
		{
			name: "atomic swap status is invalid",
			genState: func() app.GenesisState {
				gs := baseGenState(suite.addrs[0])
				_, addrs := app.GeneratePrivKeyAddressPairs(2)
				timestamp := ts(0)
				randomNumber, _ := bep3.GenerateSecureRandomNumber()
				randomNumberHash := bep3.CalculateRandomHash(randomNumber, timestamp)
				swap := bep3.NewAtomicSwap(cs(c("bnb", 5000)), randomNumberHash,
					uint64(360), timestamp, suite.addrs[0], addrs[1], TestSenderOtherChain,
					TestRecipientOtherChain, 0, bep3.NULL, true, bep3.Incoming)

				gs.AtomicSwaps = bep3.AtomicSwaps{swap}
				return app.GenesisState{"bep3": bep3.ModuleCdc.MustMarshalJSON(gs)}
			},
			expectPass: false,
		},
		{
			name: "minimum block lock below limit",
			genState: func() app.GenesisState {
				gs := baseGenState(suite.addrs[0])
				gs.Params.MinBlockLock = 1
				return app.GenesisState{"bep3": bep3.ModuleCdc.MustMarshalJSON(gs)}
			},
			expectPass: false,
		},
		{
			name: "minimum block lock above limit",
			genState: func() app.GenesisState {
				gs := baseGenState(suite.addrs[0])
				gs.Params.MinBlockLock = 500000
				return app.GenesisState{"bep3": bep3.ModuleCdc.MustMarshalJSON(gs)}
			},
			expectPass: false,
		},
		{
			name: "maximum block lock below limit",
			genState: func() app.GenesisState {
				gs := baseGenState(suite.addrs[0])
				gs.Params.MaxBlockLock = 1
				return app.GenesisState{"bep3": bep3.ModuleCdc.MustMarshalJSON(gs)}
			},
			expectPass: false,
		},
		{
			name: "maximum block lock above limit",
			genState: func() app.GenesisState {
				gs := baseGenState(suite.addrs[0])
				gs.Params.MaxBlockLock = 100000000
				return app.GenesisState{"bep3": bep3.ModuleCdc.MustMarshalJSON(gs)}
			},
			expectPass: false,
		},
		{
			name: "empty supported asset denom",
			genState: func() app.GenesisState {
				gs := baseGenState(suite.addrs[0])
				gs.Params.SupportedAssets[0].Denom = ""
				return app.GenesisState{"bep3": bep3.ModuleCdc.MustMarshalJSON(gs)}
			},
			expectPass: false,
		},
		{
			name: "negative supported asset limit",
			genState: func() app.GenesisState {
				gs := baseGenState(suite.addrs[0])
				gs.Params.SupportedAssets[0].Limit = i(-100)
				return app.GenesisState{"bep3": bep3.ModuleCdc.MustMarshalJSON(gs)}
			},
			expectPass: false,
		},
		{
			name: "duplicate supported asset denom",
			genState: func() app.GenesisState {
				gs := baseGenState(suite.addrs[0])
				gs.Params.SupportedAssets[1].Denom = "bnb"
				return app.GenesisState{"bep3": bep3.ModuleCdc.MustMarshalJSON(gs)}
			},
			expectPass: false,
		},
	}

	for _, tc := range testCases {
		if tc.expectPass {
			suite.NotPanics(func() {
				suite.app.InitializeFromGenesisStates(tc.genState())
			}, tc.name)
		} else {
			suite.Panics(func() {
				suite.app.InitializeFromGenesisStates(tc.genState())
			}, tc.name)
		}
	}
}

func TestGenesisTestSuite(t *testing.T) {
	suite.Run(t, new(GenesisTestSuite))
}
