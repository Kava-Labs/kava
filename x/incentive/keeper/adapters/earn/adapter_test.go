package earn_test

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	tmprototypes "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/kava-labs/kava/app"
	earntypes "github.com/kava-labs/kava/x/earn/types"
	"github.com/kava-labs/kava/x/incentive/keeper/adapters/earn"
	savingstypes "github.com/kava-labs/kava/x/savings/types"
	"github.com/stretchr/testify/suite"
)

type EarnAdapterTestSuite struct {
	suite.Suite

	app app.TestApp
	ctx sdk.Context

	genesisTime time.Time
	addrs       []sdk.AccAddress
}

func TestEarnAdapterTestSuite(t *testing.T) {
	suite.Run(t, new(EarnAdapterTestSuite))
}

func (suite *EarnAdapterTestSuite) SetupTest() {
	config := sdk.GetConfig()
	app.SetBech32AddressPrefixes(config)

	_, suite.addrs = app.GeneratePrivKeyAddressPairs(5)

	suite.genesisTime = time.Date(2020, 12, 15, 14, 0, 0, 0, time.UTC)
	suite.app = app.NewTestApp()

	suite.ctx = suite.app.NewContext(true, tmprototypes.Header{Time: suite.genesisTime})
}

func (suite *EarnAdapterTestSuite) TestEarnAdapter_OwnerSharesBySource_Empty() {
	ek := suite.app.GetEarnKeeper()
	adapter := earn.NewSourceAdapter(&ek)

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
			"empty vaults are zero",
			suite.addrs[0],
			[]string{
				"vault1",
				"vault2",
				"vault3",
			},
			map[string]sdk.Dec{
				"vault1": sdk.ZeroDec(),
				"vault2": sdk.ZeroDec(),
				"vault3": sdk.ZeroDec(),
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

func (suite *EarnAdapterTestSuite) TestEarnAdapter_OwnerSharesBySource() {
	vaultDenomA := "ukava"
	vaultDenomB := "usdx"

	savingsKeeper := suite.app.GetSavingsKeeper()
	savingsKeeper.SetParams(suite.ctx, savingstypes.NewParams([]string{vaultDenomA, vaultDenomB}))

	earnKeeper := suite.app.GetEarnKeeper()
	earnKeeper.SetParams(suite.ctx, earntypes.NewParams(
		earntypes.AllowedVaults{
			earntypes.NewAllowedVault(
				"ukava",
				earntypes.StrategyTypes{earntypes.STRATEGY_TYPE_SAVINGS},
				false,
				nil,
			),
		},
	))

	suite.NoError(suite.app.FundAccount(
		suite.ctx,
		suite.addrs[0],
		sdk.NewCoins(
			sdk.NewCoin(vaultDenomA, sdk.NewInt(1000000000000)),
			sdk.NewCoin(vaultDenomB, sdk.NewInt(1000000000000)),
		),
	))
	suite.NoError(suite.app.FundAccount(
		suite.ctx,
		suite.addrs[1],
		sdk.NewCoins(
			sdk.NewCoin(vaultDenomA, sdk.NewInt(1000000000000)),
			sdk.NewCoin(vaultDenomB, sdk.NewInt(1000000000000)),
		),
	))

	err := earnKeeper.Deposit(
		suite.ctx,
		suite.addrs[0],
		sdk.NewCoin(vaultDenomA, sdk.NewInt(100)),
		earntypes.STRATEGY_TYPE_SAVINGS,
	)
	suite.NoError(err)

	err = earnKeeper.Deposit(
		suite.ctx,
		suite.addrs[1],
		sdk.NewCoin(vaultDenomA, sdk.NewInt(250)),
		earntypes.STRATEGY_TYPE_SAVINGS,
	)
	suite.NoError(err)

	ek := suite.app.GetEarnKeeper()
	adapter := earn.NewSourceAdapter(&ek)

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
				vaultDenomA,
			},
			map[string]sdk.Dec{
				vaultDenomA: sdk.NewDecWithPrec(100, 0),
			},
		},
		{
			"depositor has shares - including empty deposits",
			suite.addrs[1],
			[]string{
				vaultDenomA,
				"vault2",
			},
			map[string]sdk.Dec{
				vaultDenomA: sdk.NewDecWithPrec(250, 0),
				"vault2":    sdk.ZeroDec(),
			},
		},
		{
			"non-depositor has zero shares",
			suite.addrs[2],
			[]string{
				vaultDenomA,
			},
			map[string]sdk.Dec{
				vaultDenomA: sdk.ZeroDec(),
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

func (suite *EarnAdapterTestSuite) TestEarnAdapter_TotalSharesBySource_Empty() {
	ek := suite.app.GetEarnKeeper()
	adapter := earn.NewSourceAdapter(&ek)

	tests := []struct {
		name         string
		giveSourceID string
		wantShares   sdk.Dec
	}{
		{
			"empty/invalid vaults are zero",
			"vault1",
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

func (suite *EarnAdapterTestSuite) TestEarnAdapter_TotalSharesBySource() {
	vaultDenomA := "ukava"
	vaultDenomB := "usdx"

	savingsKeeper := suite.app.GetSavingsKeeper()
	savingsKeeper.SetParams(suite.ctx, savingstypes.NewParams([]string{vaultDenomA, vaultDenomB}))

	earnKeeper := suite.app.GetEarnKeeper()
	earnKeeper.SetParams(suite.ctx, earntypes.NewParams(
		earntypes.AllowedVaults{
			earntypes.NewAllowedVault(
				vaultDenomA,
				earntypes.StrategyTypes{earntypes.STRATEGY_TYPE_SAVINGS},
				false,
				nil,
			),
		},
	))

	suite.NoError(suite.app.FundAccount(
		suite.ctx,
		suite.addrs[0],
		sdk.NewCoins(
			sdk.NewCoin(vaultDenomA, sdk.NewInt(1000000000000)),
			sdk.NewCoin(vaultDenomB, sdk.NewInt(1000000000000)),
		),
	))
	suite.NoError(suite.app.FundAccount(
		suite.ctx,
		suite.addrs[1],
		sdk.NewCoins(
			sdk.NewCoin(vaultDenomA, sdk.NewInt(1000000000000)),
			sdk.NewCoin(vaultDenomB, sdk.NewInt(1000000000000)),
		),
	))

	err := earnKeeper.Deposit(
		suite.ctx,
		suite.addrs[0],
		sdk.NewCoin(vaultDenomA, sdk.NewInt(100)),
		earntypes.STRATEGY_TYPE_SAVINGS,
	)
	suite.NoError(err)

	err = earnKeeper.Deposit(
		suite.ctx,
		suite.addrs[1],
		sdk.NewCoin(vaultDenomA, sdk.NewInt(250)),
		earntypes.STRATEGY_TYPE_SAVINGS,
	)
	suite.NoError(err)

	ek := suite.app.GetEarnKeeper()
	adapter := earn.NewSourceAdapter(&ek)

	tests := []struct {
		name         string
		giveSourceID string
		wantShares   sdk.Dec
	}{
		{
			"total shares",
			vaultDenomA,
			sdk.NewDecWithPrec(350, 0),
		},
		{
			"empty or invalid vault empty",
			"vault2",
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
