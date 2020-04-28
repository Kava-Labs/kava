package types_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/bep3/types"
	"github.com/stretchr/testify/suite"
)

type GenesisTestSuite struct {
	suite.Suite
	swaps    types.AtomicSwaps
	supplies types.AssetSupplies
}

func (suite *GenesisTestSuite) SetupTest() {
	config := sdk.GetConfig()
	app.SetBech32AddressPrefixes(config)

	count := 10
	suite.swaps = atomicSwaps(count)

	incomingSupply := int64(count * 50000)
	supply := types.NewAssetSupply("bnb", c("bnb", incomingSupply),
		c("bnb", 0), c("bnb", 0), c("bnb", 100000000000))
	suite.supplies = types.AssetSupplies{supply}
}

func (suite *GenesisTestSuite) TestValidate() {
	type args struct {
		swaps    types.AtomicSwaps
		supplies types.AssetSupplies
	}
	testCases := []struct {
		name       string
		args       args
		expectPass bool
	}{
		{
			"default",
			args{
				swaps:    types.AtomicSwaps{},
				supplies: types.AssetSupplies{},
			},
			true,
		},
		{
			"with swaps",
			args{
				swaps:    suite.swaps,
				supplies: types.AssetSupplies{},
			},
			true,
		},
		{
			"with supplies",
			args{
				swaps:    types.AtomicSwaps{},
				supplies: suite.supplies,
			},
			true,
		},
		{
			"duplicate swaps",
			args{
				swaps:    types.AtomicSwaps{suite.swaps[2], suite.swaps[2]},
				supplies: types.AssetSupplies{},
			},
			false,
		},
		{
			"duplicate supplies",
			args{
				swaps:    types.AtomicSwaps{},
				supplies: types.AssetSupplies{suite.supplies[0], suite.supplies[0]},
			},
			false,
		}}

	for _, tc := range testCases {
		suite.SetupTest()
		suite.Run(tc.name, func() {
			var gs types.GenesisState
			if tc.name == "default" {
				gs = types.DefaultGenesisState()
			} else {
				gs = types.NewGenesisState(types.DefaultParams(), tc.args.swaps, tc.args.supplies)
			}

			err := gs.Validate()
			if tc.expectPass {
				suite.Nil(err)
			} else {
				suite.Error(err)
			}
		})
	}
}

func TestGenesisTestSuite(t *testing.T) {
	suite.Run(t, new(GenesisTestSuite))
}
