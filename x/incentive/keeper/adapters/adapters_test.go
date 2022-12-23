package adapters_test

import (
	"fmt"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	tmprototypes "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/incentive/keeper/adapters/cdp"
	"github.com/kava-labs/kava/x/incentive/keeper/adapters/earn"
	"github.com/kava-labs/kava/x/incentive/keeper/adapters/swap"
	"github.com/kava-labs/kava/x/incentive/types"
	"github.com/stretchr/testify/suite"
)

type AdapterTestSuite struct {
	suite.Suite

	app app.TestApp
	ctx sdk.Context

	genesisTime time.Time
	addrs       []sdk.AccAddress

	adapters map[types.ClaimType]types.SourceAdapter
}

func TestAdapterTestSuite(t *testing.T) {
	suite.Run(t, new(AdapterTestSuite))
}

func (suite *AdapterTestSuite) SetupTest() {
	config := sdk.GetConfig()
	app.SetBech32AddressPrefixes(config)

	_, suite.addrs = app.GeneratePrivKeyAddressPairs(5)

	suite.genesisTime = time.Date(2020, 12, 15, 14, 0, 0, 0, time.UTC)
	suite.app = app.NewTestApp()

	ek := suite.app.GetEarnKeeper()

	suite.ctx = suite.app.NewContext(true, tmprototypes.Header{Time: suite.genesisTime})
	suite.adapters = map[types.ClaimType]types.SourceAdapter{
		types.CLAIM_TYPE_EARN:         earn.NewSourceAdapter(&ek),
		types.CLAIM_TYPE_SWAP:         swap.NewSourceAdapter(suite.app.GetSwapKeeper()),
		types.CLAIM_TYPE_USDX_MINTING: cdp.NewSourceAdapter(suite.app.GetCDPKeeper()),
	}
}

func (suite *AdapterTestSuite) TestAdapter_OwnerSharesBySource_Empty() {
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
			"empty source ids are zero",
			suite.addrs[0],
			[]string{
				"usdx",
				"ukava",
				"erc20/multichain/usdc",
			},
			map[string]sdk.Dec{
				"usdx":                  sdk.ZeroDec(),
				"ukava":                 sdk.ZeroDec(),
				"erc20/multichain/usdc": sdk.ZeroDec(),
			},
		},
	}

	for claimType, adapter := range suite.adapters {
		for _, tt := range tests {
			suite.Run(fmt.Sprintf("%s/%s", claimType, tt.name), func() {
				shares := adapter.OwnerSharesBySource(suite.ctx, tt.giveOwner, tt.giveSourceIDs)

				suite.Equal(tt.wantShares, shares)
			})
		}
	}
}

func (suite *AdapterTestSuite) TestAdapter_TotalSharesBySource_Empty() {
	for claimType, adapter := range suite.adapters {
		tests := []struct {
			name         string
			giveSourceID string
			wantShares   sdk.Dec
		}{
			{
				"empty/invalid claimIDs are zero",
				"unknown",
				sdk.ZeroDec(),
			},
		}

		for _, tt := range tests {
			suite.Run(fmt.Sprintf("%s/%s", claimType, tt.name), func() {
				shares := adapter.TotalSharesBySource(suite.ctx, tt.giveSourceID)

				suite.Equal(tt.wantShares, shares)
			})
		}
	}
}
