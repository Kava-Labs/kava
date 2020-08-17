package types_test

import (
	"testing"

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

	supply := types.NewAssetSupply(coin, coin, coin)
	suite.supplies = types.AssetSupplies{supply}
}

func (suite *GenesisTestSuite) TestValidate() {
	type args struct {
		swaps types.AtomicSwaps
	}
	testCases := []struct {
		name       string
		args       args
		expectPass bool
	}{
		{
			"default",
			args{
				swaps: types.AtomicSwaps{},
			},
			true,
		},
		{
			"with swaps",
			args{
				swaps: suite.swaps,
			},
			true,
		},
		{
			"duplicate swaps",
			args{
				swaps: types.AtomicSwaps{suite.swaps[2], suite.swaps[2]},
			},
			false,
		},
		{
			"invalid swap",
			args{
				swaps: types.AtomicSwaps{types.AtomicSwap{Amount: sdk.Coins{sdk.Coin{Denom: "Invalid Denom", Amount: sdk.NewInt(-1)}}}},
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
				gs = types.NewGenesisState(types.DefaultParams(), tc.args.swaps, suite.supplies)
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
