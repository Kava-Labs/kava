package adapters_test

import (
	"fmt"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	tmprototypes "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/incentive/keeper/adapters/savings"
	"github.com/kava-labs/kava/x/incentive/types"
	savingstypes "github.com/kava-labs/kava/x/savings/types"
	"github.com/stretchr/testify/suite"
)

var testDenoms = []string{
	"ukava",
	"usdx",
	"erc20/multichain/usdc",
}

type AdapterSourceTester interface {
	types.SourceAdapter

	Initialize(ctx sdk.Context) error
	Deposit(ctx sdk.Context, owner sdk.AccAddress, amount sdk.Coins) error
}

type AdapterTestSuite struct {
	suite.Suite

	app app.TestApp
	ctx sdk.Context

	genesisTime time.Time
	addrs       []sdk.AccAddress

	adapters map[types.ClaimType]AdapterSourceTester
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

	suite.ctx = suite.app.NewContext(true, tmprototypes.Header{Time: suite.genesisTime})
	suite.adapters = map[types.ClaimType]AdapterSourceTester{
		types.CLAIM_TYPE_SAVINGS: NewSavingsSourceTester(suite.app.GetSavingsKeeper()),
	}
}

func (suite *AdapterTestSuite) TestAdapter_OwnerSharesBySource_Empty() {
	for claimType, adapter := range suite.adapters {
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

		for _, tt := range tests {
			suite.Run(fmt.Sprintf("%s/%s", claimType, tt.name), func() {
				shares := adapter.OwnerSharesBySource(suite.ctx, tt.giveOwner, tt.giveSourceIDs)

				suite.Equal(tt.wantShares, shares)
			})
		}
	}
}

func (suite *AdapterTestSuite) TestAdapter_OwnerSharesBySource() {
	for claimType, adapter := range suite.adapters {
		adapter.Initialize(suite.ctx)

		err := suite.app.FundAccount(
			suite.ctx,
			suite.addrs[0],
			sdk.NewCoins(
				sdk.NewCoin(testDenoms[0], sdk.NewInt(1000000000000)),
				sdk.NewCoin(testDenoms[1], sdk.NewInt(1000000000000)),
			),
		)
		suite.NoError(err)

		err = suite.app.FundAccount(
			suite.ctx,
			suite.addrs[1],
			sdk.NewCoins(
				sdk.NewCoin(testDenoms[0], sdk.NewInt(1000000000000)),
				sdk.NewCoin(testDenoms[1], sdk.NewInt(1000000000000)),
			),
		)
		suite.NoError(err)

		err = adapter.Deposit(
			suite.ctx,
			suite.addrs[0],
			sdk.NewCoins(
				sdk.NewCoin(testDenoms[0], sdk.NewInt(100)),
				sdk.NewCoin(testDenoms[1], sdk.NewInt(100)),
			),
		)
		suite.NoError(err)

		err = adapter.Deposit(
			suite.ctx,
			suite.addrs[1],
			sdk.NewCoins(
				sdk.NewCoin(testDenoms[0], sdk.NewInt(250)),
				sdk.NewCoin(testDenoms[1], sdk.NewInt(250)),
			),
		)
		suite.NoError(err)

		tests := []struct {
			name          string
			giveOwner     sdk.AccAddress
			giveSourceIDs []string
			wantShares    map[string]sdk.Dec
		}{
			{
				"depositor has shares",
				suite.addrs[0],
				[]string{
					testDenoms[0],
					testDenoms[1],
				},
				map[string]sdk.Dec{
					testDenoms[0]: sdk.NewDecWithPrec(100, 0),
					testDenoms[1]: sdk.NewDecWithPrec(100, 0),
				},
			},
			{
				"depositor has shares - including empty deposits",
				suite.addrs[1],
				[]string{
					testDenoms[0],
					testDenoms[1],
					"unknown",
				},
				map[string]sdk.Dec{
					testDenoms[0]: sdk.NewDecWithPrec(250, 0),
					testDenoms[1]: sdk.NewDecWithPrec(250, 0),
					"unknown":     sdk.ZeroDec(),
				},
			},
			{
				"non-depositor has zero shares",
				suite.addrs[2],
				[]string{
					testDenoms[0],
					testDenoms[1],
				},
				map[string]sdk.Dec{
					testDenoms[0]: sdk.ZeroDec(),
					testDenoms[1]: sdk.ZeroDec(),
				},
			},
		}

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

func (suite *AdapterTestSuite) TestAdapter_TotalSharesBySource() {
	denomA := "ukava"
	denomB := "usdx"

	savingsKeeper := suite.app.GetSavingsKeeper()
	savingsKeeper.SetParams(suite.ctx, savingstypes.NewParams(
		[]string{
			denomA,
			denomB,
		},
	))

	suite.NoError(suite.app.FundAccount(
		suite.ctx,
		suite.addrs[0],
		sdk.NewCoins(
			sdk.NewCoin(denomA, sdk.NewInt(1000000000000)),
			sdk.NewCoin(denomB, sdk.NewInt(1000000000000)),
		),
	))
	suite.NoError(suite.app.FundAccount(
		suite.ctx,
		suite.addrs[1],
		sdk.NewCoins(
			sdk.NewCoin(denomA, sdk.NewInt(1000000000000)),
			sdk.NewCoin(denomB, sdk.NewInt(1000000000000)),
		),
	))

	err := savingsKeeper.Deposit(
		suite.ctx,
		suite.addrs[0],
		sdk.NewCoins(
			sdk.NewCoin(denomA, sdk.NewInt(100)),
			sdk.NewCoin(denomB, sdk.NewInt(100)),
		),
	)
	suite.NoError(err)

	err = savingsKeeper.Deposit(
		suite.ctx,
		suite.addrs[1],
		sdk.NewCoins(
			sdk.NewCoin(denomA, sdk.NewInt(250)),
			sdk.NewCoin(denomB, sdk.NewInt(250)),
		),
	)
	suite.NoError(err)

	adapter := savings.NewSourceAdapter(suite.app.GetSavingsKeeper())

	tests := []struct {
		name         string
		giveSourceID string
		wantShares   sdk.Dec
	}{
		{
			"total shares",
			denomA,
			sdk.NewDecWithPrec(350, 0),
		},
		{
			"empty or invalid coin empty",
			"unknown",
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
