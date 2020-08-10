package types_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/bep3/types"
)

type GenesisTestSuite struct {
	suite.Suite
	swaps    types.AtomicSwaps
	supplies types.AssetSupplies
}

func (suite *GenesisTestSuite) SetupTest() {
	config := sdk.GetConfig()
	app.SetBech32AddressPrefixes(config)

	coin := sdk.NewCoin("kava", sdk.OneInt())
	suite.swaps = atomicSwaps(10)

	supply := types.NewAssetSupply(coin, coin, coin, coin, time.Duration(0))
	suite.supplies = types.AssetSupplies{supply}
}

func (suite *GenesisTestSuite) TestValidate() {
	type args struct {
		swaps             types.AtomicSwaps
		supplies          types.AssetSupplies
		previousBlockTime time.Time
	}
	testCases := []struct {
		name       string
		args       args
		expectPass bool
	}{
		{
			"default",
			args{
				swaps:             types.AtomicSwaps{},
				previousBlockTime: types.DefaultPreviousBlockTime,
			},
			true,
		},
		{
			"with swaps",
			args{
				swaps:             suite.swaps,
				previousBlockTime: types.DefaultPreviousBlockTime,
			},
			true,
		},
		{
			"with supplies",
			args{
				swaps:             types.AtomicSwaps{},
				supplies:          suite.supplies,
				previousBlockTime: types.DefaultPreviousBlockTime,
			},
			true,
		},
		{
			"invalid supply",
			args{
				swaps:             types.AtomicSwaps{},
				supplies:          types.AssetSupplies{types.AssetSupply{IncomingSupply: sdk.Coin{"Invalid", sdk.ZeroInt()}}},
				previousBlockTime: types.DefaultPreviousBlockTime,
			},
			false,
		},
		{
			"duplicate swaps",
			args{
				swaps:             types.AtomicSwaps{suite.swaps[2], suite.swaps[2]},
				previousBlockTime: types.DefaultPreviousBlockTime,
			},
			false,
		},
		{
			"invalid swap",
			args{
				swaps:             types.AtomicSwaps{types.AtomicSwap{Amount: sdk.Coins{sdk.Coin{Denom: "Invalid Denom", Amount: sdk.NewInt(-1)}}}},
				previousBlockTime: types.DefaultPreviousBlockTime,
			},
			false,
		},
		{
			"blocktime not set",
			args{
				swaps: types.AtomicSwaps{},
			},
			false,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			var gs types.GenesisState
			if tc.name == "default" {
				gs = types.DefaultGenesisState()
			} else {
				gs = types.NewGenesisState(types.DefaultParams(), tc.args.swaps, tc.args.supplies, tc.args.previousBlockTime)
			}

			err := gs.Validate()
			if tc.expectPass {
				suite.Require().NoError(err, tc.name)
			} else {
				suite.Require().Error(err, tc.name)
			}

		})
	}
}

func TestGenesisTestSuite(t *testing.T) {
	suite.Run(t, new(GenesisTestSuite))
}
