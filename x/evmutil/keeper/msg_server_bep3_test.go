package keeper_test

import (
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/x/evmutil/testutil"
	"github.com/kava-labs/kava/x/evmutil/types"
)

func (suite *MsgServerSuite) TestConvertCoinToERC20_Bep3() {
	invoker, err := sdk.AccAddressFromBech32("kava123fxg0l602etulhhcdm0vt7l57qya5wjcrwhzz")
	receiverAddr := testutil.MustNewInternalEVMAddressFromString("0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2")
	suite.Require().NoError(err)

	type errArgs struct {
		expectPass bool
		contains   string
	}

	tests := []struct {
		name                  string
		msg                   types.MsgConvertCoinToERC20
		userBankBalance       sdkmath.Int
		expUserErc20Balance   sdkmath.Int
		moduleErc20Balance    sdkmath.Int
		expModuleErc20Balance sdkmath.Int
		errArgs               errArgs
	}{
		{
			name: "valid",
			msg: types.NewMsgConvertCoinToERC20(
				invoker.String(),
				receiverAddr.String(),
				sdk.NewCoin("bnb", sdkmath.NewInt(1234)),
			),
			userBankBalance:       sdkmath.NewInt(1_234),
			moduleErc20Balance:    sdkmath.NewInt(10e13),
			expUserErc20Balance:   sdkmath.NewInt(1.234e13),
			expModuleErc20Balance: sdkmath.NewInt(8.766e13),
			errArgs: errArgs{
				expectPass: true,
			},
		},
		{
			name: "invalid - invalid receiver address",
			msg: types.NewMsgConvertCoinToERC20(
				invoker.String(),
				"0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc",
				sdk.NewCoin("bnb", sdkmath.NewInt(1234)),
			),
			userBankBalance:    sdkmath.NewInt(1_234),
			moduleErc20Balance: sdkmath.NewInt(1e14),
			errArgs: errArgs{
				expectPass: false,
				contains:   "invalid Receiver address: string is not a hex address",
			},
		},
		{
			name: "invalid - initiator receiver address",
			msg: types.NewMsgConvertCoinToERC20(
				"0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc",
				receiverAddr.String(),
				sdk.NewCoin("bnb", sdkmath.NewInt(1234)),
			),
			userBankBalance:    sdkmath.NewInt(1_234),
			moduleErc20Balance: sdkmath.NewInt(1e14),
			errArgs: errArgs{
				expectPass: false,
				contains:   "invalid Initiator address",
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			suite.SetupTest()
			contractAddr := suite.DeployERC20()

			// enable conversion pair
			params := suite.Keeper.GetParams(suite.Ctx)
			params.EnabledConversionPairs[0].Denom = tc.msg.Amount.Denom
			suite.Keeper.SetParams(suite.Ctx, params)

			pair := types.NewConversionPair(
				contractAddr,
				tc.msg.Amount.Denom,
			)

			// fund user & module account
			if tc.userBankBalance.GT(sdkmath.ZeroInt()) {
				err = suite.App.FundAccount(
					suite.Ctx,
					invoker,
					sdk.NewCoins(sdk.NewCoin(pair.Denom, tc.userBankBalance)),
				)
				suite.Require().NoError(err)
			}
			if tc.moduleErc20Balance.GT(sdkmath.ZeroInt()) {
				err := suite.Keeper.MintERC20(
					suite.Ctx,
					pair.GetAddress(), // contractAddr
					types.NewInternalEVMAddress(types.ModuleEVMAddress), //receiver
					tc.moduleErc20Balance.BigInt(),
				)
				suite.Require().NoError(err)
			}

			_, err = suite.msgServer.ConvertCoinToERC20(sdk.WrapSDKContext(suite.Ctx), &tc.msg)

			if tc.errArgs.expectPass {
				suite.Require().NoError(err)

				// validate user balance
				bal := suite.GetERC20BalanceOf(
					types.ERC20MintableBurnableContract.ABI,
					pair.GetAddress(),
					receiverAddr,
				)
				suite.Require().Equal(
					tc.expUserErc20Balance.BigInt().Int64(),
					bal.Int64(),
					"user erc20 balance should match expected amount",
				)

				// validate module balance
				bal = suite.GetERC20BalanceOf(
					types.ERC20MintableBurnableContract.ABI,
					pair.GetAddress(),
					types.NewInternalEVMAddress(types.ModuleEVMAddress),
				)
				suite.Require().Equal(
					tc.expModuleErc20Balance.BigInt().Int64(),
					bal.Int64(),
					"module erc20 balance should match expected amount",
				)

				// msg server event
				suite.EventsContains(suite.GetEvents(),
					sdk.NewEvent(
						sdk.EventTypeMessage,
						sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
						sdk.NewAttribute(sdk.AttributeKeySender, tc.msg.Initiator),
					))

				// keeper event
				suite.EventsContains(suite.GetEvents(),
					sdk.NewEvent(
						types.EventTypeConvertCoinToERC20,
						sdk.NewAttribute(types.AttributeKeyInitiator, tc.msg.Initiator),
						sdk.NewAttribute(types.AttributeKeyReceiver, tc.msg.Receiver),
						sdk.NewAttribute(types.AttributeKeyERC20Address, pair.GetAddress().String()),
						sdk.NewAttribute(types.AttributeKeyAmount, tc.msg.Amount.String()),
					))

			} else {
				suite.Require().Error(err)
				suite.Require().Contains(err.Error(), tc.errArgs.contains)
			}
		})
	}
}

func (suite *MsgServerSuite) TestConvertERC20ToCoin_Bep3() {
	invoker := testutil.MustNewInternalEVMAddressFromString("0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2")
	invokerCosmosAddr, err := sdk.AccAddressFromHexUnsafe(invoker.String()[2:])
	suite.Require().NoError(err)
	contractAddr := suite.DeployERC20()

	type errArgs struct {
		expectPass bool
		contains   string
	}

	tests := []struct {
		name                string
		msg                 types.MsgConvertERC20ToCoin
		userErc20Balance    sdkmath.Int
		expUserBankBalance  sdkmath.Int
		expUserErc20Balance sdkmath.Int
		errArgs             errArgs
	}{
		{
			name: "valid",
			msg: types.NewMsgConvertERC20ToCoin(
				invoker,
				invokerCosmosAddr,
				contractAddr,
				sdkmath.NewInt(1.0031e18),
			),
			userErc20Balance:    sdkmath.NewInt(1.12e18),
			expUserBankBalance:  sdkmath.NewInt(1.0031e8),
			expUserErc20Balance: sdkmath.NewInt(0.1169e18),
			errArgs: errArgs{
				expectPass: true,
			},
		},
		{
			name: "invalid - invalid initiator address",
			msg: types.MsgConvertERC20ToCoin{
				Initiator:        "0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc",
				Receiver:         invokerCosmosAddr.String(),
				KavaERC20Address: contractAddr.String(),
				Amount:           sdkmath.NewInt(12e8),
			},
			userErc20Balance: sdkmath.NewInt(2e18),
			errArgs: errArgs{
				expectPass: false,
				contains:   "invalid initiator address: string is not a hex address",
			},
		},
		{
			name: "invalid - invalid receiver address",
			msg: types.MsgConvertERC20ToCoin{
				Initiator:        invoker.String(),
				Receiver:         "0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc",
				KavaERC20Address: contractAddr.String(),
				Amount:           sdkmath.NewInt(12e8),
			},
			userErc20Balance: sdkmath.NewInt(2e18),
			errArgs: errArgs{
				expectPass: false,
				contains:   "invalid receiver address",
			},
		},
		{
			name: "invalid - invalid contract address",
			msg: types.MsgConvertERC20ToCoin{
				Initiator:        invoker.String(),
				Receiver:         invokerCosmosAddr.String(),
				KavaERC20Address: invokerCosmosAddr.String(),
				Amount:           sdkmath.NewInt(12e8),
			},
			userErc20Balance: sdkmath.NewInt(2e18),
			errArgs: errArgs{
				expectPass: false,
				contains:   "invalid contract address",
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			suite.SetupTest()
			contractAddr := suite.DeployERC20()

			conversionDenom := "bnb"
			params := suite.Keeper.GetParams(suite.Ctx)
			params.EnabledConversionPairs[0].Denom = conversionDenom
			suite.Keeper.SetParams(suite.Ctx, params)

			pair := types.NewConversionPair(
				contractAddr,
				conversionDenom,
			)

			// fund user erc20 balance
			if tc.userErc20Balance.GT(sdkmath.ZeroInt()) {
				err := suite.Keeper.MintERC20(
					suite.Ctx,
					pair.GetAddress(),
					invoker,
					tc.userErc20Balance.BigInt(),
				)
				suite.Require().NoError(err)
			}

			// create user account, otherwise `CallEVMWithData` will fail due to failing to get user account when finding its sequence.
			err = suite.App.FundAccount(suite.Ctx, invokerCosmosAddr, sdk.NewCoins(sdk.NewCoin(conversionDenom, sdk.ZeroInt())))
			suite.Require().NoError(err)

			_, err := suite.msgServer.ConvertERC20ToCoin(sdk.WrapSDKContext(suite.Ctx), &tc.msg)

			if tc.errArgs.expectPass {
				suite.Require().NoError(err)

				// validate user balance after conversion
				bal := suite.GetERC20BalanceOf(
					types.ERC20MintableBurnableContract.ABI,
					pair.GetAddress(),
					testutil.MustNewInternalEVMAddressFromString(invoker.String()),
				)
				suite.Require().Equal(tc.expUserErc20Balance.BigInt().Int64(), bal.Int64(), "user erc20 balance is invalid")

				// validate user coin balance
				coinBal := suite.App.GetBankKeeper().GetBalance(suite.Ctx, invokerCosmosAddr, pair.Denom)
				suite.Require().Equal(tc.expUserBankBalance, coinBal.Amount, "user coin balance is invalid")

				// msg server event
				suite.EventsContains(suite.GetEvents(),
					sdk.NewEvent(
						sdk.EventTypeMessage,
						sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
						sdk.NewAttribute(sdk.AttributeKeySender, tc.msg.Initiator),
					))

				// keeper event
				suite.EventsContains(suite.GetEvents(),
					sdk.NewEvent(
						types.EventTypeConvertERC20ToCoin,
						sdk.NewAttribute(types.AttributeKeyERC20Address, pair.GetAddress().String()),
						sdk.NewAttribute(types.AttributeKeyInitiator, tc.msg.Initiator),
						sdk.NewAttribute(types.AttributeKeyReceiver, tc.msg.Receiver),
						sdk.NewAttribute(types.AttributeKeyAmount, sdk.NewCoin(pair.Denom, tc.expUserBankBalance).String()),
					))
			} else {
				suite.Require().Error(err)
				suite.Require().Contains(err.Error(), tc.errArgs.contains)
			}
		})
	}
}
