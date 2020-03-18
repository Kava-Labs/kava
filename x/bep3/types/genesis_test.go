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
	swaps types.AtomicSwaps
}

func (suite *GenesisTestSuite) SetupTest() {
	config := sdk.GetConfig()
	app.SetBech32AddressPrefixes(config)
	suite.swaps = atomicSwaps(10)
	return
}

func (suite *GenesisTestSuite) TestValidate() {
	type args struct {
		swaps  types.AtomicSwaps
		assets []sdk.Coin
	}
	testCases := []struct {
		name       string
		args       args
		expectPass bool
	}{
		{
			"default",
			args{
				swaps:  make(types.AtomicSwaps, 0),
				assets: []sdk.Coin{},
			},
			true,
		},
		{
			"with swaps",
			args{
				swaps:  suite.swaps,
				assets: []sdk.Coin{},
			},
			true,
		},
		{
			"duplicate swaps",
			args{
				swaps:  types.AtomicSwaps{suite.swaps[2], suite.swaps[2]},
				assets: []sdk.Coin{},
			},
			false,
		},
	}

	for _, tc := range testCases {
		var gs types.GenesisState
		if tc.name == "default" {
			gs = types.DefaultGenesisState()
		} else {
			gs = types.NewGenesisState(types.DefaultParams(), tc.args.swaps, tc.args.assets)
		}

		err := gs.Validate()
		if tc.expectPass {
			suite.Nil(err)
		} else {
			suite.Error(err)
		}
	}
}

func TestGenesisTestSuite(t *testing.T) {
	suite.Run(t, new(GenesisTestSuite))
}
