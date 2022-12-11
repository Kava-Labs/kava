package keeper_test

import (
	"fmt"
	"math/big"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/kava-labs/kava/app"
	communitytypes "github.com/kava-labs/kava/x/community/types"
	"github.com/kava-labs/kava/x/evmutil/keeper"
	"github.com/kava-labs/kava/x/evmutil/testutil"
	"github.com/kava-labs/kava/x/evmutil/types"
	"github.com/stretchr/testify/suite"
)

type MsgServerSuite struct {
	testutil.Suite

	msgServer types.MsgServer
}

func (suite *MsgServerSuite) SetupTest() {
	suite.Suite.SetupTest()
	suite.msgServer = keeper.NewMsgServerImpl(suite.App.GetEvmutilKeeper())
}

func TestMsgServerSuite(t *testing.T) {
	suite.Run(t, new(MsgServerSuite))
}

func (suite *MsgServerSuite) TestConvertCoinToERC20() {
	invoker, err := sdk.AccAddressFromBech32("kava123fxg0l602etulhhcdm0vt7l57qya5wjcrwhzz")
	suite.Require().NoError(err)

	err = suite.App.FundAccount(suite.Ctx, invoker, sdk.NewCoins(sdk.NewCoin("erc20/usdc", sdk.NewInt(10000))))
	suite.Require().NoError(err)

	contractAddr := suite.DeployERC20()

	pair := types.NewConversionPair(
		contractAddr,
		"erc20/usdc",
	)

	// Module account should have starting balance
	pairStartingBal := big.NewInt(10000)
	err = suite.Keeper.MintERC20(
		suite.Ctx,
		pair.GetAddress(), // contractAddr
		types.NewInternalEVMAddress(types.ModuleEVMAddress), //receiver
		pairStartingBal,
	)
	suite.Require().NoError(err)

	type errArgs struct {
		expectPass bool
		contains   string
	}

	tests := []struct {
		name    string
		msg     types.MsgConvertCoinToERC20
		errArgs errArgs
	}{
		{
			"valid",
			types.NewMsgConvertCoinToERC20(
				invoker.String(),
				"0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2",
				sdk.NewCoin("erc20/usdc", sdk.NewInt(1234)),
			),
			errArgs{
				expectPass: true,
			},
		},
		{
			"invalid - odd length hex address",
			types.NewMsgConvertCoinToERC20(
				invoker.String(),
				"0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc",
				sdk.NewCoin("erc20/usdc", sdk.NewInt(1234)),
			),
			errArgs{
				expectPass: false,
				contains:   "invalid Receiver address: string is not a hex address",
			},
		},
		// Amount coin is not validated by msg_server, but on msg itself
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			_, err := suite.msgServer.ConvertCoinToERC20(sdk.WrapSDKContext(suite.Ctx), &tc.msg)

			if tc.errArgs.expectPass {
				suite.Require().NoError(err)

				bal := suite.GetERC20BalanceOf(
					types.CustomERC20Contract.ABI,
					pair.GetAddress(),
					testutil.MustNewInternalEVMAddressFromString(tc.msg.Receiver),
				)

				suite.Require().Equal(tc.msg.Amount.Amount.BigInt(), bal, "balance should match converted amount")

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

func (suite *MsgServerSuite) TestConvertERC20ToCoin() {
	contractAddr := suite.DeployERC20()
	pair := types.NewConversionPair(
		contractAddr,
		"erc20/usdc",
	)

	// give invoker account some erc20 usdc to begin with
	invoker := testutil.MustNewInternalEVMAddressFromString("0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2")
	pairStartingBal := big.NewInt(10_000_000)
	err := suite.Keeper.MintERC20(
		suite.Ctx,
		pair.GetAddress(), // contractAddr
		invoker,           //receiver
		pairStartingBal,
	)
	suite.Require().NoError(err)

	invokerCosmosAddr, err := sdk.AccAddressFromHex(invoker.String()[2:])
	suite.Require().NoError(err)

	// create user account, otherwise `CallEVMWithData` will fail due to failing to get user account when finding its sequence.
	err = suite.App.FundAccount(suite.Ctx, invokerCosmosAddr, sdk.NewCoins(sdk.NewCoin(pair.Denom, sdk.ZeroInt())))
	suite.Require().NoError(err)

	type errArgs struct {
		expectPass bool
		contains   string
	}

	tests := []struct {
		name           string
		msg            types.MsgConvertERC20ToCoin
		approvalAmount *big.Int
		errArgs        errArgs
	}{
		{
			"valid",
			types.NewMsgConvertERC20ToCoin(
				invoker,
				invokerCosmosAddr,
				contractAddr,
				sdk.NewInt(10_000),
			),
			math.MaxBig256,
			errArgs{
				expectPass: true,
			},
		},
		{
			"invalid - invalid hex address",
			types.MsgConvertERC20ToCoin{
				Initiator:        "0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc",
				Receiver:         invokerCosmosAddr.String(),
				KavaERC20Address: contractAddr.String(),
				Amount:           sdk.NewInt(10_000),
			},
			math.MaxBig256,
			errArgs{
				expectPass: false,
				contains:   "invalid initiator address: string is not a hex address",
			},
		},
		{
			"invalid - insufficient coins",
			types.NewMsgConvertERC20ToCoin(
				invoker,
				invokerCosmosAddr,
				contractAddr,
				sdk.NewIntFromBigInt(pairStartingBal).Add(sdk.OneInt()),
			),
			math.MaxBig256,
			errArgs{
				expectPass: false,
				contains:   "transfer amount exceeds balance",
			},
		},
		{
			"invalid - contract address",
			types.NewMsgConvertERC20ToCoin(
				invoker,
				invokerCosmosAddr,
				testutil.MustNewInternalEVMAddressFromString("0x7Bbf300890857b8c241b219C6a489431669b3aFA"),
				sdk.NewInt(10_000),
			),
			math.MaxBig256,
			errArgs{
				expectPass: false,
				contains:   "ERC20 token not enabled to convert to sdk.Coin",
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			_, err := suite.msgServer.ConvertERC20ToCoin(sdk.WrapSDKContext(suite.Ctx), &tc.msg)

			if tc.errArgs.expectPass {
				suite.Require().NoError(err)

				// validate user balance after conversion
				bal := suite.GetERC20BalanceOf(
					types.CustomERC20Contract.ABI,
					pair.GetAddress(),
					testutil.MustNewInternalEVMAddressFromString(tc.msg.Initiator),
				)
				expectedBal := sdk.NewIntFromBigInt(pairStartingBal).Sub(tc.msg.Amount)
				suite.Require().Equal(expectedBal.BigInt(), bal, "user erc20 balance is invalid")

				// validate user coin balance
				coinBal := suite.App.GetBankKeeper().GetBalance(suite.Ctx, invokerCosmosAddr, pair.Denom)
				suite.Require().Equal(tc.msg.Amount, coinBal.Amount, "user coin balance is invalid")

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
						sdk.NewAttribute(types.AttributeKeyAmount, sdk.NewCoin(pair.Denom, tc.msg.Amount).String()),
					))
			} else {
				suite.Require().Error(err)
				suite.Require().Contains(err.Error(), tc.errArgs.contains)
			}
		})
	}
}

func (suite *MsgServerSuite) TestEVMCall() {
	userAddr := app.RandomAddress()
	userEvmAddr := common.BytesToAddress(userAddr.Bytes())
	contractAddr := suite.DeployERC20()

	authorityModule := communitytypes.ModuleAccountName
	authorityAct := suite.AccountKeeper.GetModuleAccount(suite.Ctx, authorityModule)
	authorityAddr := authorityAct.GetAddress().String()
	authorityEvmAddr := types.NewInternalEVMAddress(common.BytesToAddress(authorityAct.GetAddress().Bytes()))
	erc20StartingBal := big.NewInt(1000)

	validFnAbi := `{
		"inputs": [
			{ "type": "address", "name": "to" },
			{ "type": "uint256", "name": "amount" }
		],
		"name": "transfer",
		"type": "function"
	}`
	validData := encodeTransferFn(userEvmAddr, 30)

	type errArgs struct {
		expectPass bool
		contains   string
	}

	tests := []struct {
		name       string
		msg        types.MsgEVMCall
		expUsdcBal sdk.Int
		errArgs    errArgs
	}{
		{
			"valid - erc20 contract transfer call",
			types.MsgEVMCall{
				To:        contractAddr.String(),
				FnAbi:     validFnAbi,
				Data:      validData,
				Authority: authorityAddr,
				Amount:    sdk.ZeroInt(),
			},
			sdk.NewInt(970),
			errArgs{
				expectPass: true,
			},
		},
		{
			"valid - transfer call",
			types.MsgEVMCall{
				To:        userEvmAddr.String(),
				Amount:    keeper.ConversionMultiplier.Mul(sdk.NewInt(11)),
				Authority: authorityAddr,
			},
			sdk.NewInt(1000),
			errArgs{
				expectPass: true,
			},
		},
		{
			"valid - payable deposit",
			types.MsgEVMCall{
				To:     userEvmAddr.String(),
				Amount: keeper.ConversionMultiplier.Mul(sdk.NewInt(11)),
				FnAbi: `{
					"inputs": [],
					"name": "deposit",
					"type": "function"
				}`,
				Data: fmt.Sprintf(
					"0x%s",
					"d0e30db0", // deposit()
				),
				Authority: authorityAddr,
			},
			sdk.NewInt(1000),
			errArgs{
				expectPass: true,
			},
		},
		{
			"invalid - extra data passed after valid fn data",
			types.MsgEVMCall{
				To:        contractAddr.String(),
				FnAbi:     validFnAbi,
				Data:      validData + "00ab",
				Authority: authorityAddr,
				Amount:    sdk.ZeroInt(),
			},
			sdk.NewInt(970),
			errArgs{
				expectPass: false,
				contains:   "invalid call data: call data does not match unpacked data",
			},
		},
		{
			"invalid - insufficient funds",
			types.MsgEVMCall{
				To:        contractAddr.String(),
				FnAbi:     validFnAbi,
				Data:      encodeTransferFn(userEvmAddr, 2000),
				Authority: authorityAddr,
				Amount:    sdk.ZeroInt(),
			},
			sdk.NewInt(1000),
			errArgs{
				expectPass: false,
				contains:   "transfer amount exceeds balance",
			},
		},
		{
			"invalid - no input data",
			types.MsgEVMCall{
				To:        contractAddr.String(),
				FnAbi:     validFnAbi,
				Authority: authorityAddr,
				Amount:    sdk.ZeroInt(),
			},
			sdk.NewInt(1000),
			errArgs{
				expectPass: false,
				contains:   "evm transaction execution failed",
			},
		},
		{
			"invalid - authority not target module",
			types.MsgEVMCall{
				To:        userEvmAddr.String(),
				Amount:    sdk.NewInt(12),
				Authority: userAddr.String(),
			},
			sdk.NewInt(1000),
			errArgs{
				expectPass: false,
				contains:   "invalid authority;",
			},
		},
		{
			"invalid - to address is not evm address",
			types.MsgEVMCall{
				To:        userAddr.String(),
				Amount:    sdk.NewInt(12),
				Authority: authorityAddr,
			},
			sdk.NewInt(1000),
			errArgs{
				expectPass: false,
				contains:   "to 'kava1r35v2uc9p9slrx3t0sux29t44gwcgfnzpz3uf9' is not hex address",
			},
		},
		{
			"invalid - with data but no fnAbi",
			types.MsgEVMCall{
				To:        contractAddr.String(),
				Data:      validData,
				Authority: authorityAddr,
				Amount:    sdk.ZeroInt(),
			},
			sdk.NewInt(1000),
			errArgs{
				expectPass: false,
				contains:   "fnAbi is not provided",
			},
		},
		{
			"invalid - hex data cannot be parsed",
			types.MsgEVMCall{
				To:        contractAddr.String(),
				FnAbi:     validFnAbi,
				Data:      "hello",
				Authority: authorityAddr,
				Amount:    sdk.ZeroInt(),
			},
			sdk.NewInt(1000),
			errArgs{
				expectPass: false,
				contains:   "invalid data format: hex string without 0x prefix",
			},
		},
		{
			"invalid - calling non-existent contract function",
			types.MsgEVMCall{
				To: contractAddr.String(),
				FnAbi: `{
					"inputs": [
						{ "type": "address", "name": "to" },
						{ "type": "uint256", "name": "amount" }
					],
					"name": "badFn",
					"type": "function"
				}`,
				Data: fmt.Sprintf(
					"0x%s%s%s",
					"dee4aafc", // badFn(address,uint256)
					encodeAddress(userEvmAddr),
					encodeInt(30),
				),
				Amount:    sdk.ZeroInt(),
				Authority: authorityAddr,
			},
			sdk.NewInt(1000),
			errArgs{
				expectPass: false,
				contains:   "evm call failed: execution reverted: evm transaction execution failed",
			},
		},
		{
			"invalid - sending eth to contract fn that is not payable",
			types.MsgEVMCall{
				To:        contractAddr.String(),
				FnAbi:     validFnAbi,
				Data:      validData,
				Authority: authorityAddr,
				Amount:    sdk.NewInt(10),
			},
			sdk.NewInt(1000),
			errArgs{
				expectPass: false,
				contains:   "evm call failed: execution reverted",
			},
		},
		{
			"invalid - contract throws an error",
			types.MsgEVMCall{
				To: contractAddr.String(),
				FnAbi: `{
					"inputs": [],
					"name": "triggerError",
					"type": "function"
				}`,
				Data:      "0xbcffd7cf", // call triggerError
				Authority: authorityAddr,
				Amount:    sdk.ZeroInt(),
			},
			sdk.NewInt(1000),
			errArgs{
				expectPass: false,
				contains:   "this function will always trigger an error",
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			suite.SetupTest()
			contractAddr := suite.DeployERC20()

			// give invoker account some erc20 and kava to being with
			err := suite.Keeper.MintERC20(
				suite.Ctx,
				contractAddr,
				authorityEvmAddr,
				erc20StartingBal,
			)
			suite.Require().NoError(err)
			authorityCoin := sdk.NewCoin("ukava", sdk.NewInt(100))
			err = suite.App.FundModuleAccount(suite.Ctx, authorityModule, sdk.NewCoins(authorityCoin))
			suite.Require().NoError(err)

			// validate msg
			err = tc.msg.ValidateBasic()
			if !tc.errArgs.expectPass && err != nil {
				suite.Require().Error(err)
				suite.Require().Contains(err.Error(), tc.errArgs.contains)
				return
			}

			_, err = suite.msgServer.EVMCall(sdk.WrapSDKContext(suite.Ctx), &tc.msg)
			expectedKavaBal := authorityCoin.Amount.Mul(keeper.ConversionMultiplier)

			if tc.errArgs.expectPass {
				suite.Require().NoError(err)
				if !tc.msg.Amount.IsNil() {
					expectedKavaBal = expectedKavaBal.Sub(tc.msg.Amount)
				}

				// msg server event
				suite.EventsContains(suite.GetEvents(),
					sdk.NewEvent(
						sdk.EventTypeMessage,
						sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
						sdk.NewAttribute(sdk.AttributeKeySender, tc.msg.Authority),
					))
			} else {
				suite.Require().Error(err)
				suite.Require().Contains(err.Error(), tc.errArgs.contains)
			}

			// validate authority kava balance
			coinBal := suite.EvmBankKeeper.GetBalance(suite.Ctx, authorityAct.GetAddress(), keeper.EvmDenom)
			suite.Require().Equal(expectedKavaBal, coinBal.Amount, "user akava balance is invalid")

			// validate authority erc20 balance after msg
			bal := suite.GetERC20BalanceOf(
				types.CustomERC20Contract.ABI,
				contractAddr,
				authorityEvmAddr,
			)
			suite.Require().Equal(tc.expUsdcBal.BigInt(), bal, "user erc20 balance is invalid")
		})
	}
}

func encodeTransferFn(addr common.Address, amt int64) string {
	return fmt.Sprintf(
		"0x%s%s%s",
		"a9059cbb", // transfer(address,uint256)
		encodeAddress(addr),
		encodeInt(amt),
	)
}

func encodeAddress(addr common.Address) string {
	return hexutil.Encode(common.LeftPadBytes(addr.Bytes(), 32))[2:]
}

func encodeInt(amt int64) string {
	return hexutil.Encode(common.LeftPadBytes(big.NewInt(amt).Bytes(), 32))[2:]
}
