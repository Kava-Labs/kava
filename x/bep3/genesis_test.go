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

	type GenState func() []app.GenesisState

	testCases := []struct {
		name       string
		genState   GenState
		expectPass bool
	}{
		{
			name: "default",
			genState: func() []app.GenesisState {
				return []app.GenesisState{NewBep3GenStateMulti(suite.addrs[0])}
			},
			expectPass: true,
		},
		{
			name: "import atomic swaps",
			genState: func() []app.GenesisState {
				gs := baseGenState(suite.addrs[0])
				coins := []sdk.Coins{}
				_, addrs := app.GeneratePrivKeyAddressPairs(3)
				for i := 0; i < len(addrs); i++ {
					coins = append(coins, cs(c("bnb", 50000), c("inc", 50000)))
				}
				authGS := app.NewAuthGenState(addrs, coins)
				var swaps bep3.AtomicSwaps
				for i := 0; i < 2; i++ {
					swap := loadSwap(addrs[i+1], suite.addrs[0], i)
					swaps = append(swaps, swap)
				}
				gs.Params.AssetParams[0].SupplyLimit.IncomingSupply = c("bnb", 50000)
				gs.Params.AssetParams[1].SupplyLimit.IncomingSupply = c("inc", 50000)
				gs.AtomicSwaps = swaps
				return []app.GenesisState{authGS, {"bep3": bep3.ModuleCdc.MustMarshalJSON(gs)}}
			},
			expectPass: true,
		},
		{
			name: "0 deputy fees",
			genState: func() []app.GenesisState {
				gs := baseGenState(suite.addrs[0])
				gs.Params.AssetParams[0].IncomingSwapFixedFee = sdk.ZeroInt()
				return []app.GenesisState{{"bep3": bep3.ModuleCdc.MustMarshalJSON(gs)}}
			},
			expectPass: true,
		},
		{
			name: "atomic swap asset type is unsupported",
			genState: func() []app.GenesisState {
				gs := baseGenState(suite.addrs[0])
				_, addrs := app.GeneratePrivKeyAddressPairs(2)
				timestamp := ts(0)
				randomNumber, _ := bep3.GenerateSecureRandomNumber()
				randomNumberHash := bep3.CalculateRandomHash(randomNumber[:], timestamp)
				swap := bep3.NewAtomicSwap(cs(c("fake", 500000)), randomNumberHash,
					uint64(360), timestamp, suite.addrs[0], addrs[1], TestSenderOtherChain,
					TestRecipientOtherChain, 0, bep3.Open, true, bep3.Incoming)

				gs.AtomicSwaps = bep3.AtomicSwaps{swap}
				return []app.GenesisState{{"bep3": bep3.ModuleCdc.MustMarshalJSON(gs)}}
			},
			expectPass: false,
		},
		{
			name: "atomic swap status is invalid",
			genState: func() []app.GenesisState {
				gs := baseGenState(suite.addrs[0])
				_, addrs := app.GeneratePrivKeyAddressPairs(2)
				timestamp := ts(0)
				randomNumber, _ := bep3.GenerateSecureRandomNumber()
				randomNumberHash := bep3.CalculateRandomHash(randomNumber[:], timestamp)
				swap := bep3.NewAtomicSwap(cs(c("bnb", 5000)), randomNumberHash,
					uint64(360), timestamp, suite.addrs[0], addrs[1], TestSenderOtherChain,
					TestRecipientOtherChain, 0, bep3.NULL, true, bep3.Incoming)

				gs.AtomicSwaps = bep3.AtomicSwaps{swap}
				return []app.GenesisState{{"bep3": bep3.ModuleCdc.MustMarshalJSON(gs)}}
			},
			expectPass: false,
		},
		{
			name: "minimum block lock cannot be > maximum block lock",
			genState: func() []app.GenesisState {
				gs := baseGenState(suite.addrs[0])
				gs.Params.AssetParams[0].MinBlockLock = 201
				gs.Params.AssetParams[0].MaxBlockLock = 200
				return []app.GenesisState{{"bep3": bep3.ModuleCdc.MustMarshalJSON(gs)}}
			},
			expectPass: false,
		},
		{
			name: "empty supported asset denom",
			genState: func() []app.GenesisState {
				gs := baseGenState(suite.addrs[0])
				gs.Params.AssetParams[0].Denom = ""
				return []app.GenesisState{{"bep3": bep3.ModuleCdc.MustMarshalJSON(gs)}}
			},
			expectPass: false,
		},
		{
			name: "negative supported asset limit",
			genState: func() []app.GenesisState {
				gs := baseGenState(suite.addrs[0])
				gs.Params.AssetParams[0].SupplyLimit.SupplyLimit.Amount = i(-100)
				return []app.GenesisState{{"bep3": bep3.ModuleCdc.MustMarshalJSON(gs)}}
			},
			expectPass: false,
		},
		{
			name: "duplicate supported asset denom",
			genState: func() []app.GenesisState {
				gs := baseGenState(suite.addrs[0])
				gs.Params.AssetParams[1].Denom = "bnb"
				return []app.GenesisState{{"bep3": bep3.ModuleCdc.MustMarshalJSON(gs)}}
			},
			expectPass: false,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			if tc.expectPass {
				suite.NotPanics(func() {
					suite.app.InitializeFromGenesisStatesWithTime(suite.ctx.BlockTime(), tc.genState()...)
				}, tc.name)
			} else {
				suite.Panics(func() {
					suite.app.InitializeFromGenesisStatesWithTime(suite.ctx.BlockTime(), tc.genState()...)
				}, tc.name)
			}
		})
	}
}

func TestGenesisTestSuite(t *testing.T) {
	suite.Run(t, new(GenesisTestSuite))
}
