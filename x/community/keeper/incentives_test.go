package keeper_test

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	"github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmtime "github.com/tendermint/tendermint/types/time"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/community/keeper"
	types "github.com/kava-labs/kava/x/community/types"
	kavadisttypes "github.com/kava-labs/kava/x/kavadist/types"
)

// Test suite used for all incentive tests
type IncentivesTestSuite struct {
	suite.Suite

	App    app.TestApp
	Ctx    sdk.Context
	Keeper keeper.Keeper
}

// The default state used by each test
func (suite *IncentivesTestSuite) SetupTest() {
	app.SetSDKConfig()
	tApp := app.NewTestApp()
	suite.App = tApp
	suite.Ctx = suite.App.NewContext(true, tmproto.Header{Height: 1, Time: tmtime.Now()})
	suite.Keeper = suite.App.GetCommunityKeeper()

	// Set up x/mint and x/kavadist gen state
	mintGen := minttypes.DefaultGenesisState()
	mintGen.Params.InflationMax = sdk.NewDecWithPrec(595, 3)
	mintGen.Params.InflationMin = sdk.NewDecWithPrec(595, 3)
	kavadistGen := kavadisttypes.DefaultGenesisState()
	kavadistGen.Params.Active = true
	appCodec := tApp.AppCodec()
	suite.App.InitializeFromGenesisStates(
		app.GenesisState{minttypes.ModuleName: appCodec.MustMarshalJSON(mintGen)},
		app.GenesisState{kavadisttypes.ModuleName: appCodec.MustMarshalJSON(kavadistGen)},
	)
}

func TestIncentivesTestSuite(t *testing.T) {
	suite.Run(t, new(IncentivesTestSuite))
}

func (suite *IncentivesTestSuite) TestCanPayoutStakingRewards() {
	suite.Run("returns false if we have not upgrade", func() {
		suite.SetupTest()
		suite.False(suite.Keeper.CanPayoutStakingRewards(suite.Ctx))
	})

	suite.Run("returns true if we have upgraded", func() {
		suite.SetupTest()
		suite.Keeper.StartDisableInflationUpgrade(suite.Ctx)
		suite.True(suite.Keeper.CanPayoutStakingRewards(suite.Ctx))
	})
}

func (suite *IncentivesTestSuite) TestPayoutStakingRewards() {
	suite.Run("errors if params are not set", func() {
		suite.SetupTest()
		store := suite.Ctx.KVStore(suite.App.GetKVStoreKey(types.StoreKey))
		store.Delete(types.ParamsKey)
		err := suite.Keeper.PayoutStakingRewards(suite.Ctx)
		suite.Error(err)
		suite.ErrorIs(err, types.ErrStakingRewardsPayout)
	})

	suite.Run("errors if previous block time is not set", func() {
		suite.SetupTest()
		_, found := suite.Keeper.GetPreviousBlockTime(suite.Ctx)
		suite.False(found)
		err := suite.Keeper.PayoutStakingRewards(suite.Ctx)
		suite.Error(err)
		suite.ErrorIs(err, types.ErrStakingRewardsPayout)
	})

	suite.Run("errors if community account is out of funds", func() {
		suite.SetupTest()
		suite.Keeper.StartDisableInflationUpgrade(suite.Ctx)
		ctx := suite.Ctx.WithBlockTime(suite.Ctx.BlockTime().Add(5 * time.Second))
		err := suite.Keeper.PayoutStakingRewards(ctx)
		suite.Error(err)
		suite.ErrorIs(err, sdkerrors.ErrInsufficientFunds)
	})

	suite.Run("errors if new block time is less than previous block", func() {
		suite.SetupTest()
		suite.Keeper.StartDisableInflationUpgrade(suite.Ctx)
		ctx := suite.Ctx.WithBlockTime(suite.Ctx.BlockTime().Add(-1 * time.Second))
		err := suite.Keeper.PayoutStakingRewards(ctx)
		suite.Error(err)
		suite.ErrorIs(err, types.ErrStakingRewardsPayout)
	})

	suite.Run("emits payout rewards event", func() {
		suite.SetupTest()
		suite.Keeper.StartDisableInflationUpgrade(suite.Ctx)
		suite.fundCommunityAccount(sdkmath.NewInt(1_000_000_000))
		ctx := suite.Ctx.WithBlockTime(suite.Ctx.BlockTime().Add(5 * time.Second))
		err := suite.Keeper.PayoutStakingRewards(ctx)
		suite.NoError(err)
		events := suite.Ctx.EventManager().Events()
		suite.EventsContains(
			events,
			sdk.NewEvent(
				types.EventTypePayoutRewards,
				sdk.NewAttribute(
					types.AttributeKeyRewardsPerSecond, "744191",
				),
				sdk.NewAttribute(
					sdk.AttributeKeyAmount, "3720955",
				),
			),
		)
	})

	// pays out the correct staking rewards
	testCases := []struct {
		name          string
		blockDuration time.Duration
		expRewards    int64
	}{
		{
			name:          "6.2 seconds block time",
			blockDuration: suite.mustParseDuration("6.2s"),
			expRewards:    4_465_146,
		},
		{
			name:          "21.9 seconds block time",
			blockDuration: suite.mustParseDuration("21.9s"),
			expRewards:    15_628_011,
		},
		{
			name:          "same block time as previous block",
			blockDuration: suite.mustParseDuration("0s"),
			expRewards:    0,
		},
	}
	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.SetupTest()
			bk := suite.App.GetBankKeeper()
			feeCollectorAcc := suite.App.GetAccountKeeper().GetModuleAccount(suite.Ctx, authtypes.FeeCollectorName)

			feeCollectorBal := bk.GetBalance(suite.Ctx, feeCollectorAcc.GetAddress(), "ukava")
			suite.Equal(sdk.ZeroInt(), feeCollectorBal.Amount)

			suite.Keeper.StartDisableInflationUpgrade(suite.Ctx)
			ctx := suite.Ctx.WithBlockTime(suite.Ctx.BlockTime().Add(tc.blockDuration))
			suite.fundCommunityAccount(sdkmath.NewInt(1_000_000_000))
			err := suite.Keeper.PayoutStakingRewards(ctx)
			suite.NoError(err)

			// check that the fee collector account has the correct amount of funds
			feeCollectorBal = bk.GetBalance(ctx, feeCollectorAcc.GetAddress(), "ukava")
			suite.Equal(tc.expRewards, feeCollectorBal.Amount.Int64())
		})
	}
}

func (suite *IncentivesTestSuite) TestShouldStartDisableInflationUpgrade() {
	shouldUpgrade := func() bool {
		return suite.Keeper.ShouldStartDisableInflationUpgrade(suite.Ctx)
	}

	suite.Run("skips upgrade if community params does not exist", func() {
		suite.SetupTest()

		// remove param from store
		store := suite.Ctx.KVStore(suite.App.GetKVStoreKey(types.StoreKey))
		store.Delete(types.ParamsKey)

		_, found := suite.Keeper.GetParams(suite.Ctx)
		suite.False(found)
		suite.False(shouldUpgrade())
	})

	suite.Run("skips upgrade if upgrade time is set in the future", func() {
		suite.SetupTest()

		suite.setUpgradeTimeFromNow(1 * time.Hour)
		suite.False(shouldUpgrade())
	})

	suite.Run("upgrades if params are set to the default", func() {
		suite.SetupTest()

		suite.Keeper.SetParams(suite.Ctx, types.DefaultParams())
		param, found := suite.Keeper.GetParams(suite.Ctx)
		suite.True(found)
		suite.Equal(time.Time{}, param.UpgradeTimeDisableInflation)
		suite.True(shouldUpgrade())
	})

	suite.Run("upgrades if blockTime is at or after upgrade time", func() {
		suite.SetupTest()

		suite.setUpgradeTimeFromNow(0)
		suite.True(shouldUpgrade())
		suite.setUpgradeTimeFromNow(-2 * time.Minute)
		suite.True(shouldUpgrade())
	})

	suite.Run("skips upgrade if already upgraded", func() {
		suite.SetupTest()

		suite.setUpgradeTimeFromNow(-2 * time.Minute)
		suite.True(shouldUpgrade())
		suite.Keeper.StartDisableInflationUpgrade(suite.Ctx)
		suite.False(shouldUpgrade())
	})
}

func (suite *IncentivesTestSuite) TestStartDisableInflationUpgrade() {
	suite.Run("upgrade should set mint and kavadist inflation to 0", func() {
		suite.SetupTest()

		mintParams := suite.App.GetMintKeeper().GetParams(suite.Ctx)
		suite.Equal(sdk.NewDecWithPrec(595, 3), mintParams.InflationMax)
		kavadistParams := suite.App.GetKavadistKeeper().GetParams(suite.Ctx)
		suite.True(kavadistParams.Active)

		suite.setUpgradeTimeFromNow(-2 * time.Minute)
		suite.Keeper.StartDisableInflationUpgrade(suite.Ctx)

		mintParams = suite.App.GetMintKeeper().GetParams(suite.Ctx)
		suite.Equal(sdk.ZeroDec(), mintParams.InflationMax)
		suite.Equal(sdk.ZeroDec(), mintParams.InflationMin)

		kavadistParams = suite.App.GetKavadistKeeper().GetParams(suite.Ctx)
		suite.False(kavadistParams.Active)
	})

	suite.Run("upgrade should set previous block time", func() {
		suite.SetupTest()

		suite.setUpgradeTimeFromNow(-2 * time.Minute)
		suite.Keeper.StartDisableInflationUpgrade(suite.Ctx)
		prevBlockTime, found := suite.Keeper.GetPreviousBlockTime(suite.Ctx)
		suite.True(found)
		suite.Equal(suite.Ctx.BlockTime(), prevBlockTime)
	})
}

func (suite *IncentivesTestSuite) setUpgradeTimeFromNow(t time.Duration) {
	params, found := suite.Keeper.GetParams(suite.Ctx)
	suite.True(found)
	params.UpgradeTimeDisableInflation = suite.Ctx.BlockTime().Add(t)
	suite.Keeper.SetParams(suite.Ctx, params)
}

// EventsContains asserts that the expected event is in the provided events
func (suite *IncentivesTestSuite) EventsContains(events sdk.Events, expectedEvent sdk.Event) {
	foundMatch := false
	for _, event := range events {
		if event.Type == expectedEvent.Type {
			if reflect.DeepEqual(attrsToMap(expectedEvent.Attributes), attrsToMap(event.Attributes)) {
				foundMatch = true
			}
		}
	}

	suite.True(foundMatch, fmt.Sprintf("event of type %s not found or did not match", expectedEvent.Type))
}

func attrsToMap(attrs []abci.EventAttribute) []sdk.Attribute { // new cosmos changed the event attribute type
	out := []sdk.Attribute{}

	for _, attr := range attrs {
		out = append(out, sdk.NewAttribute(string(attr.Key), string(attr.Value)))
	}

	return out
}

func (suite *IncentivesTestSuite) fundCommunityAccount(amt sdkmath.Int) {
	err := suite.App.FundModuleAccount(
		suite.Ctx,
		types.ModuleAccountName,
		sdk.NewCoins(sdk.NewCoin("ukava", amt)),
	)
	suite.NoError(err)
}

func (suite *IncentivesTestSuite) mustParseDuration(dur string) time.Duration {
	ret, err := time.ParseDuration(dur)
	suite.NoError(err)
	return ret
}
