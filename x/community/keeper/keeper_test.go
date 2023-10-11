package keeper_test

import (
	"strings"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/stretchr/testify/suite"

	"github.com/kava-labs/kava/x/community/keeper"
	"github.com/kava-labs/kava/x/community/testutil"
	"github.com/kava-labs/kava/x/community/types"
)

// Test suite used for all keeper tests
type KeeperTestSuite struct {
	testutil.Suite
}

// The default state used by each test
func (suite *KeeperTestSuite) SetupTest() {
	suite.Suite.SetupTest()
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func (suite *KeeperTestSuite) TestCommunityPool() {
	suite.SetupTest()
	maccAddr := suite.App.GetAccountKeeper().GetModuleAddress(types.ModuleAccountName)

	funds := sdk.NewCoins(
		sdk.NewCoin("ukava", sdkmath.NewInt(10000)),
		sdk.NewCoin("usdx", sdkmath.NewInt(100)),
	)
	sender := suite.CreateFundedAccount(funds)

	suite.Run("FundCommunityPool", func() {
		err := suite.Keeper.FundCommunityPool(suite.Ctx, sender, funds)
		suite.Require().NoError(err)

		// check that community pool received balance
		suite.App.CheckBalance(suite.T(), suite.Ctx, maccAddr, funds)
		suite.Equal(funds, suite.Keeper.GetModuleAccountBalance(suite.Ctx))
		// check that sender had balance deducted
		suite.App.CheckBalance(suite.T(), suite.Ctx, sender, sdk.NewCoins())
	})

	// send it back
	suite.Run("DistributeFromCommunityPool - valid", func() {
		err := suite.Keeper.DistributeFromCommunityPool(suite.Ctx, sender, funds)
		suite.Require().NoError(err)

		// community pool has funds deducted
		suite.App.CheckBalance(suite.T(), suite.Ctx, maccAddr, sdk.NewCoins())
		suite.Equal(sdk.NewCoins(), suite.Keeper.GetModuleAccountBalance(suite.Ctx))
		// receiver receives the funds
		suite.App.CheckBalance(suite.T(), suite.Ctx, sender, funds)
	})

	// can't send more than we have!
	suite.Run("DistributeFromCommunityPool - insufficient funds", func() {
		suite.Equal(sdk.NewCoins(), suite.Keeper.GetModuleAccountBalance(suite.Ctx))
		err := suite.Keeper.DistributeFromCommunityPool(suite.Ctx, sender, funds)
		suite.Require().ErrorContains(err, "insufficient funds")
	})
}

func (suite *KeeperTestSuite) TestGetAndSetStakingRewardsState() {
	keeper := suite.Keeper

	defaultParams := keeper.GetStakingRewardsState(suite.Ctx)
	suite.Equal(time.Time{}, defaultParams.LastAccumulationTime, "expected default returned accumulation time to be zero")
	suite.Equal(sdkmath.LegacyZeroDec(), defaultParams.LastTruncationError, "expected default truncation error to be zero")

	suite.NotPanics(func() { keeper.SetStakingRewardsState(suite.Ctx, defaultParams) }, "expected setting default state to not panic")

	invalidParams := defaultParams
	invalidParams.LastTruncationError = sdkmath.LegacyDec{}

	suite.Panics(func() { keeper.SetStakingRewardsState(suite.Ctx, invalidParams) }, "expected setting invalid state to panic")

	validParams := defaultParams
	validParams.LastAccumulationTime = time.Date(2023, 9, 29, 11, 42, 53, 123456789, time.UTC)
	validParams.LastTruncationError = sdkmath.LegacyMustNewDecFromStr("0.50000000000000000")

	suite.NotPanics(func() { keeper.SetStakingRewardsState(suite.Ctx, validParams) }, "expected setting valid state to not panic")

	suite.Equal(validParams, keeper.GetStakingRewardsState(suite.Ctx), "expected fetched state to equal set state")
}

func (suite *KeeperTestSuite) TestGetAuthority_Default() {
	suite.Equal(
		authtypes.NewModuleAddress(govtypes.ModuleName),
		suite.Keeper.GetAuthority(),
		"expected fetched authority to equal x/gov address",
	)
}

func (suite *KeeperTestSuite) TestGetAuthority_Any() {
	tests := []struct {
		name      string
		authority sdk.AccAddress
	}{
		{
			name:      "gov",
			authority: authtypes.NewModuleAddress(govtypes.ModuleName),
		},
		{
			name:      "random",
			authority: sdk.AccAddress("random"),
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			suite.NotPanics(func() {
				suite.Keeper = keeper.NewKeeper(
					suite.App.AppCodec(),
					suite.App.GetKVStoreKey(types.StoreKey),
					suite.App.GetAccountKeeper(),
					suite.App.GetBankKeeper(),
					suite.App.GetCDPKeeper(),
					suite.App.GetDistrKeeper(),
					suite.App.GetHardKeeper(),
					suite.App.GetMintKeeper(),
					suite.App.GetKavadistKeeper(),
					suite.App.GetStakingKeeper(),
					tc.authority,
				)
			})

			suite.Equalf(
				tc.authority,
				suite.Keeper.GetAuthority(),
				"expected fetched authority to equal %s address",
				tc.authority,
			)
		})
	}
}

func (suite *KeeperTestSuite) TestNewKeeper_InvalidAuthority() {
	tests := []struct {
		name      string
		authority sdk.AccAddress
		panicStr  string
	}{
		{
			name:      "empty",
			authority: sdk.AccAddress{},
			panicStr:  "invalid authority address: addresses cannot be empty: unknown address",
		},
		{
			name:      "too long",
			authority: sdk.AccAddress(strings.Repeat("a", address.MaxAddrLen+1)),
			panicStr:  "invalid authority address: address max length is 255, got 256: unknown address",
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			suite.PanicsWithValue(
				tc.panicStr,
				func() {
					suite.Keeper = keeper.NewKeeper(
						suite.App.AppCodec(),
						suite.App.GetKVStoreKey(types.StoreKey),
						suite.App.GetAccountKeeper(),
						suite.App.GetBankKeeper(),
						suite.App.GetCDPKeeper(),
						suite.App.GetDistrKeeper(),
						suite.App.GetHardKeeper(),
						suite.App.GetMintKeeper(),
						suite.App.GetKavadistKeeper(),
						suite.App.GetStakingKeeper(),
						tc.authority,
					)
				})
		})
	}
}
