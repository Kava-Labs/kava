package swap_test

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	tmprototypes "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/incentive/keeper/adapters/swap"
	"github.com/stretchr/testify/suite"
)

type SwapAdapterTestSuite struct {
	suite.Suite

	app app.TestApp
	ctx sdk.Context

	genesisTime time.Time
	addrs       []sdk.AccAddress
}

func TestSwapAdapterTestSuite(t *testing.T) {
	suite.Run(t, new(SwapAdapterTestSuite))
}

func (suite *SwapAdapterTestSuite) SetupTest() {
	config := sdk.GetConfig()
	app.SetBech32AddressPrefixes(config)

	_, suite.addrs = app.GeneratePrivKeyAddressPairs(5)

	suite.genesisTime = time.Date(2020, 12, 15, 14, 0, 0, 0, time.UTC)
	suite.app = app.NewTestApp()

	suite.ctx = suite.app.NewContext(true, tmprototypes.Header{Time: suite.genesisTime})
}

func (suite *SwapAdapterTestSuite) TestSwapAdapter_OwnerSharesBySource_Empty() {
	adapter := swap.NewSourceAdapter(suite.app.GetSwapKeeper())

	tests := []struct {
		name          string
		giveOwner     sdk.AccAddress
		giveSourceIDs []string
		wantShares    map[string]sdk.Dec
	}{
		{
			"empty requests",
			suite.addrs[0],
			[]string{},
			map[string]sdk.Dec{},
		},
		{
			"empty pools are zero",
			suite.addrs[0],
			[]string{
				"pool1",
				"pool2",
				"pool3",
			},
			map[string]sdk.Dec{
				"pool1": sdk.ZeroDec(),
				"pool2": sdk.ZeroDec(),
				"pool3": sdk.ZeroDec(),
			},
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			shares := adapter.OwnerSharesBySource(suite.ctx, tt.giveOwner, tt.giveSourceIDs)

			suite.Equal(tt.wantShares, shares)
		})
	}
}

func (suite *SwapAdapterTestSuite) TestSwapAdapter_TotalSharesBySource_Empty() {
	adapter := swap.NewSourceAdapter(suite.app.GetSwapKeeper())

	tests := []struct {
		name         string
		giveSourceID string
		wantShares   sdk.Dec
	}{
		{
			"empty/invalid pools are zero",
			"pool1",
			sdk.ZeroDec(),
		},
		{
			"invalid request returns zero",
			"",
			sdk.ZeroDec(),
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			shares := adapter.TotalSharesBySource(suite.ctx, tt.giveSourceID)

			suite.Equal(tt.wantShares, shares)
		})
	}
}
