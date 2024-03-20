package keeper_test

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	"github.com/kava-labs/kava/x/evmutil/testutil"
	"github.com/kava-labs/kava/x/evmutil/types"
)

type Bep3ConversionTestSuite struct {
	testutil.Suite
}

func TestBep3ConversionTestSuite(t *testing.T) {
	suite.Run(t, new(Bep3ConversionTestSuite))
}

func (suite *Bep3ConversionTestSuite) setEnabledConversionPairDenom(denom string) {
	params := suite.Keeper.GetParams(suite.Ctx)
	params.EnabledConversionPairs[0].Denom = denom
	suite.Keeper.SetParams(suite.Ctx, params)
}

func (suite *Bep3ConversionTestSuite) TestConvertCoinToERC20_Bep3() {
	invoker, err := sdk.AccAddressFromBech32("kava123fxg0l602etulhhcdm0vt7l57qya5wjcrwhzz")
	receiverAddr := testutil.MustNewInternalEVMAddressFromString("0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2")
	suite.Require().NoError(err)

	type errArgs struct {
		expectPass bool
		contains   string
	}

	tests := []struct {
		name                  string
		pairDenom             string
		userBankBalance       sdkmath.Int
		expUserErc20Balance   sdkmath.Int
		moduleErc20Balance    sdkmath.Int
		expModuleErc20Balance sdkmath.Int
		coinToConvert         sdk.Coin
		errArgs               errArgs
	}{
		{
			name:                  "success - btcb conversion",
			pairDenom:             "btcb",
			userBankBalance:       sdkmath.NewInt(1_234),
			moduleErc20Balance:    sdkmath.NewInt(10e13),
			expUserErc20Balance:   sdkmath.NewInt(1.234e13),
			expModuleErc20Balance: sdkmath.NewInt(8.766e13),
			coinToConvert:         sdk.NewCoin("btcb", sdkmath.NewInt(1_234)),
			errArgs: errArgs{
				expectPass: true,
			},
		},
		{
			name:                  "success - xrpb convert smallest unit",
			pairDenom:             "xrpb",
			userBankBalance:       sdkmath.NewInt(2),
			moduleErc20Balance:    sdkmath.NewInt(10e13),
			expUserErc20Balance:   sdkmath.NewInt(1e10),
			expModuleErc20Balance: sdkmath.NewInt(9.999e13),
			coinToConvert:         sdk.NewCoin("xrpb", sdkmath.NewInt(1)),
			errArgs: errArgs{
				expectPass: true,
			},
		},
		{
			name:                  "success - no change when 0 amount converted",
			pairDenom:             "btcb",
			userBankBalance:       sdkmath.NewInt(1234),
			moduleErc20Balance:    sdkmath.NewInt(1e14),
			expUserErc20Balance:   sdkmath.ZeroInt(),
			expModuleErc20Balance: sdkmath.NewInt(1e14),
			coinToConvert:         sdk.NewCoin("btcb", sdkmath.NewInt(0)),
			errArgs: errArgs{
				expectPass: true,
			},
		},
		{
			name:               "error - bep3 not enabled",
			pairDenom:          "btcb",
			userBankBalance:    sdkmath.NewInt(1_234),
			moduleErc20Balance: sdkmath.NewInt(1e14),
			coinToConvert:      sdk.NewCoin("bnb", sdkmath.NewInt(1_234)), // bnb is not enabled
			errArgs: errArgs{
				expectPass: false,
				contains:   "bnb: ERC20 token not enabled to convert to sdk.Coin",
			},
		},
		{
			name:               "error - module account does not have enough balance to unlock",
			pairDenom:          "btcb",
			userBankBalance:    sdkmath.NewInt(1_234),
			moduleErc20Balance: sdkmath.NewInt(1e9),
			coinToConvert:      sdk.NewCoin("btcb", sdkmath.NewInt(1)),
			errArgs: errArgs{
				expectPass: false,
				contains:   "execution reverted: ERC20: transfer amount exceeds balance",
			},
		},
		{
			name:               "error - user converting more than user balance",
			pairDenom:          "btcb",
			userBankBalance:    sdkmath.NewInt(1_234),
			moduleErc20Balance: sdkmath.NewInt(1e14),
			coinToConvert:      sdk.NewCoin("btcb", sdkmath.NewInt(20_000)),
			errArgs: errArgs{
				expectPass: false,
				contains:   "spendable balance 1234btcb is smaller than 20000btcb: insufficient funds",
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			suite.SetupTest()
			contractAddr := suite.DeployERC20()
			suite.setEnabledConversionPairDenom(tc.pairDenom)
			pair := types.NewConversionPair(
				contractAddr,
				tc.pairDenom,
			)

			// fund user & module account
			if tc.userBankBalance.GT(sdkmath.ZeroInt()) {
				err = suite.App.FundAccount(
					suite.Ctx,
					invoker,
					sdk.NewCoins(sdk.NewCoin(tc.pairDenom, tc.userBankBalance)),
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

			// execute bep3 conversion
			err := suite.Keeper.ConvertCoinToERC20(suite.Ctx, invoker, receiverAddr, tc.coinToConvert)

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

				// keeper event
				suite.EventsContains(suite.GetEvents(),
					sdk.NewEvent(
						types.EventTypeConvertCoinToERC20,
						sdk.NewAttribute(types.AttributeKeyInitiator, invoker.String()),
						sdk.NewAttribute(types.AttributeKeyReceiver, receiverAddr.String()),
						sdk.NewAttribute(types.AttributeKeyERC20Address, pair.GetAddress().String()),
						sdk.NewAttribute(types.AttributeKeyAmount, tc.coinToConvert.String()),
					))
			} else {
				suite.Require().Error(err)
				suite.Require().Contains(err.Error(), tc.errArgs.contains)
			}
		})
	}
}

func (suite *Bep3ConversionTestSuite) TestConvertERC20ToCoin_Bep3() {
	invoker := testutil.MustNewInternalEVMAddressFromString("0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2")
	invokerCosmosAddr, err := sdk.AccAddressFromHexUnsafe(invoker.String()[2:])
	suite.Require().NoError(err)

	type errArgs struct {
		expectPass bool
		contains   string
	}

	tests := []struct {
		name                string
		pairDenom           string
		contractAddr        string
		userErc20Balance    sdkmath.Int
		expUserBankBalance  sdkmath.Int
		expUserErc20Balance sdkmath.Int
		convertAmount       sdkmath.Int
		errArgs             errArgs
	}{
		{
			name:                "success - btcb conversion with no dust",
			pairDenom:           "btcb",
			userErc20Balance:    sdkmath.NewInt(1.12e18),
			expUserBankBalance:  sdkmath.NewInt(1.0031e8),
			expUserErc20Balance: sdkmath.NewInt(0.1169e18),
			convertAmount:       sdkmath.NewInt(1.0031e18),
			errArgs: errArgs{
				expectPass: true,
			},
		},
		{
			name:                "success - xrpb convert smallest bank unit",
			pairDenom:           "xrpb",
			userErc20Balance:    sdkmath.NewInt(2e18), // 2xrpb
			expUserBankBalance:  sdkmath.NewInt(1),
			expUserErc20Balance: sdkmath.NewInt(1.99999999e18),
			convertAmount:       sdkmath.NewInt(1.12e10),
			errArgs: errArgs{
				expectPass: true,
			},
		},
		{
			name:                "success - bnb conversion with dust",
			pairDenom:           "bnb",
			userErc20Balance:    sdkmath.NewInt(2e18),
			expUserBankBalance:  sdkmath.NewInt(12),
			expUserErc20Balance: sdkmath.NewInt(1.99999988e18),
			convertAmount:       sdkmath.NewInt(12.123e10),
			errArgs: errArgs{
				expectPass: true,
			},
		},
		{
			name:             "fail - converting less than 1 bank unit",
			pairDenom:        "bnb",
			userErc20Balance: sdkmath.NewInt(2e18),
			convertAmount:    sdkmath.NewInt(12e8),
			errArgs: errArgs{
				expectPass: false,
				contains:   "unable to convert bep3 coin due converting less than 1bnb",
			},
		},
		{
			name:             "fail - contract not enabled",
			pairDenom:        "bnb",
			contractAddr:     "0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2",
			userErc20Balance: sdkmath.NewInt(2e18),
			convertAmount:    sdkmath.NewInt(2e18),
			errArgs: errArgs{
				expectPass: false,
				contains:   "ERC20 token not enabled to convert",
			},
		},
		{
			name:             "fail - converting 0 amount of bep3 erc20 token",
			pairDenom:        "bnb",
			userErc20Balance: sdkmath.NewInt(2e18),
			convertAmount:    sdkmath.NewInt(0),
			errArgs: errArgs{
				expectPass: false,
				contains:   "unable to convert bep3 coin due converting less than 1bnb",
			},
		},
		{
			name:                "success - not bep3 conversion",
			pairDenom:           "xrp",
			userErc20Balance:    sdkmath.NewInt(2.5e18),
			expUserBankBalance:  sdkmath.NewInt(2.1e18),
			expUserErc20Balance: sdkmath.NewInt(0.4e18),
			convertAmount:       sdkmath.NewInt(2.1e18),
			errArgs: errArgs{
				expectPass: true,
			},
		},
		{
			name:             "fail - user converting more than user balance",
			pairDenom:        "bnb",
			userErc20Balance: sdkmath.NewInt(2e18),
			convertAmount:    sdkmath.NewInt(2.3e18),
			errArgs: errArgs{
				expectPass: false,
				contains:   "transfer amount exceeds balance",
			},
		},
		{
			name:                "success - user converting more than balance but only by dust amount",
			pairDenom:           "bnb",
			userErc20Balance:    sdkmath.NewInt(2e18),
			expUserBankBalance:  sdkmath.NewInt(2e8),
			expUserErc20Balance: sdkmath.NewInt(0),
			convertAmount:       sdkmath.NewInt(2.0000000091e18),
			errArgs: errArgs{
				expectPass: true,
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			suite.SetupTest()
			contractAddr := suite.DeployERC20()
			suite.setEnabledConversionPairDenom(tc.pairDenom)
			if tc.contractAddr != "" {
				contractAddr = testutil.MustNewInternalEVMAddressFromString(tc.contractAddr)
			}
			pair := types.NewConversionPair(
				contractAddr,
				tc.pairDenom,
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
			err = suite.App.FundAccount(suite.Ctx, invokerCosmosAddr, sdk.NewCoins(sdk.NewCoin(tc.pairDenom, sdk.ZeroInt())))
			suite.Require().NoError(err)

			// execute bep3 conversion
			err := suite.Keeper.ConvertERC20ToCoin(suite.Ctx, invoker, invokerCosmosAddr, pair.GetAddress(), tc.convertAmount)

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

				// keeper event
				suite.EventsContains(suite.GetEvents(),
					sdk.NewEvent(
						types.EventTypeConvertERC20ToCoin,
						sdk.NewAttribute(types.AttributeKeyERC20Address, pair.GetAddress().String()),
						sdk.NewAttribute(types.AttributeKeyInitiator, invoker.String()),
						sdk.NewAttribute(types.AttributeKeyReceiver, invokerCosmosAddr.String()),
						sdk.NewAttribute(types.AttributeKeyAmount, sdk.NewCoin(pair.Denom, tc.expUserBankBalance).String()),
					))
			} else {
				suite.Require().Error(err)
				suite.Require().Contains(err.Error(), tc.errArgs.contains)
			}
		})
	}
}
