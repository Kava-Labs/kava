package keeper_test

import (
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/stretchr/testify/suite"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/community/keeper"
	"github.com/kava-labs/kava/x/community/testutil"
	"github.com/kava-labs/kava/x/community/types"
)

type msgServerTestSuite struct {
	testutil.Suite

	communityPool sdk.AccAddress
	msgServer     types.MsgServer
}

func (suite *msgServerTestSuite) SetupTest() {
	suite.Suite.SetupTest()

	suite.communityPool = suite.App.GetAccountKeeper().GetModuleAddress(types.ModuleAccountName)
	suite.msgServer = keeper.NewMsgServerImpl(suite.Keeper)
}

func TestMsgServerTestSuite(t *testing.T) {
	suite.Run(t, new(msgServerTestSuite))
}

func (suite *msgServerTestSuite) TestMsgFundCommunityPool() {
	singleCoin := sdk.NewCoins(sdk.NewCoin("ukava", sdkmath.NewInt(2e6)))
	multipleCoins := sdk.NewCoins(
		sdk.NewCoin("ukava", sdkmath.NewInt(3e6)),
		sdk.NewCoin("usdx", sdkmath.NewInt(1e7)),
	)
	testCases := []struct {
		name            string
		setup           func() *types.MsgFundCommunityPool
		expectedBalance sdk.Coins
		shouldPass      bool
	}{
		{
			name: "valid funding of single coin",
			setup: func() *types.MsgFundCommunityPool {
				sender := app.RandomAddress()
				suite.App.FundAccount(suite.Ctx, sender, singleCoin)
				return &types.MsgFundCommunityPool{
					Amount:    singleCoin,
					Depositor: sender.String(),
				}
			},
			expectedBalance: singleCoin,
			shouldPass:      true,
		},
		{
			name: "valid funding of multiple coins",
			setup: func() *types.MsgFundCommunityPool {
				sender := app.RandomAddress()
				suite.App.FundAccount(suite.Ctx, sender, multipleCoins)
				return &types.MsgFundCommunityPool{
					Amount:    multipleCoins,
					Depositor: sender.String(),
				}
			},
			expectedBalance: multipleCoins,
			shouldPass:      true,
		},
		{
			name: "invalid - failing message validation",
			setup: func() *types.MsgFundCommunityPool {
				return &types.MsgFundCommunityPool{
					Amount:    sdk.NewCoins(),
					Depositor: app.RandomAddress().String(),
				}
			},
			expectedBalance: sdk.NewCoins(),
			shouldPass:      false,
		},
		{
			name: "invalid - failing tx, insufficient funds",
			setup: func() *types.MsgFundCommunityPool {
				return &types.MsgFundCommunityPool{
					Amount:    singleCoin,
					Depositor: app.RandomAddress().String(),
				}
			},
			expectedBalance: sdk.NewCoins(),
			shouldPass:      false,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.SetupTest()

			msg := tc.setup()
			_, err := suite.msgServer.FundCommunityPool(sdk.WrapSDKContext(suite.Ctx), msg)
			if tc.shouldPass {
				suite.NoError(err)
			} else {
				suite.Error(err)
			}

			balance := suite.Keeper.GetModuleAccountBalance(suite.Ctx)
			suite.App.CheckBalance(suite.T(), suite.Ctx, suite.communityPool, balance)
		})
	}
}

func (suite *msgServerTestSuite) TestMsgUpdateParams() {
	testCases := []struct {
		name        string
		setup       func() *types.MsgUpdateParams
		expectedErr error
	}{
		{
			name: "new params overwrite existing",
			setup: func() *types.MsgUpdateParams {
				newParams := types.DefaultParams()
				newParams.UpgradeTimeDisableInflation = time.Date(2050, 1, 1, 0, 0, 0, 0, time.UTC)
				return &types.MsgUpdateParams{
					Authority: authtypes.NewModuleAddress(govtypes.ModuleName).String(),
					Params:    newParams,
				}
			},
		},
		{
			name: "msg with invalid authority is rejected",
			setup: func() *types.MsgUpdateParams {
				return &types.MsgUpdateParams{
					Authority: authtypes.NewModuleAddress("not gov").String(),
					Params: types.Params{
						UpgradeTimeDisableInflation: time.Date(2050, 1, 1, 0, 0, 0, 0, time.UTC),
					},
				}
			},
			expectedErr: sdkerrors.ErrUnauthorized,
		},
		{
			name: "msg with invalid params is rejected",
			setup: func() *types.MsgUpdateParams {
				return &types.MsgUpdateParams{
					Authority: authtypes.NewModuleAddress(govtypes.ModuleName).String(),
					Params: types.Params{
						UpgradeTimeDisableInflation:           time.Time{},
						StakingRewardsPerSecond:               sdkmath.LegacyNewDec(-5), // invalid
						UpgradeTimeSetStakingRewardsPerSecond: sdkmath.LegacyNewDec(1000),
					},
				}
			},
			expectedErr: types.ErrInvalidParams,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.SetupTest()

			msg := tc.setup()

			oldParams, found := suite.Keeper.GetParams(suite.Ctx)
			suite.Require().True(found)
			_, err := suite.msgServer.UpdateParams(sdk.WrapSDKContext(suite.Ctx), msg)
			newParams, found := suite.Keeper.GetParams(suite.Ctx)
			suite.Require().True(found)

			if tc.expectedErr == nil {
				suite.NoError(err)
				suite.Equal(msg.Params, newParams)
			} else {
				suite.ErrorIs(err, tc.expectedErr)
				suite.Equal(oldParams, newParams)
			}
		})
	}
}
