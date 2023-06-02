package keeper_test

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/suite"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/math"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/evmutil/keeper"
	"github.com/kava-labs/kava/x/evmutil/testutil"
	"github.com/kava-labs/kava/x/evmutil/types"
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

	err = suite.App.FundAccount(suite.Ctx, invoker, sdk.NewCoins(sdk.NewCoin("erc20/usdc", sdkmath.NewInt(10000))))
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
				sdk.NewCoin("erc20/usdc", sdkmath.NewInt(1234)),
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
				sdk.NewCoin("erc20/usdc", sdkmath.NewInt(1234)),
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
					types.ERC20MintableBurnableContract.ABI,
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

	invokerCosmosAddr, err := sdk.AccAddressFromHexUnsafe(invoker.String()[2:])
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
				sdkmath.NewInt(10_000),
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
				Amount:           sdkmath.NewInt(10_000),
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
				sdkmath.NewIntFromBigInt(pairStartingBal).Add(sdk.OneInt()),
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
				sdkmath.NewInt(10_000),
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
					types.ERC20MintableBurnableContract.ABI,
					pair.GetAddress(),
					testutil.MustNewInternalEVMAddressFromString(tc.msg.Initiator),
				)
				expectedBal := sdkmath.NewIntFromBigInt(pairStartingBal).Sub(tc.msg.Amount)
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

func (suite *MsgServerSuite) TestConvertCosmosCoinToERC20_InitialContractDeploy() {
	allowedDenom := "ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2"
	initialFunding := int64(1e10)
	fundedAccount := app.RandomAddress()

	setup := func() {
		suite.SetupTest()

		// make the denom allowed for conversion
		params := suite.Keeper.GetParams(suite.Ctx)
		params.AllowedCosmosDenoms = types.NewAllowedCosmosCoinERC20Tokens(
			types.NewAllowedCosmosCoinERC20Token(allowedDenom, "Kava EVM Atom", "ATOM", 6),
		)
		suite.Keeper.SetParams(suite.Ctx, params)

		// fund account
		err := suite.App.FundAccount(suite.Ctx, fundedAccount, sdk.NewCoins(
			sdk.NewInt64Coin(allowedDenom, initialFunding),
		))
		suite.NoError(err, "failed to initially fund account")
	}

	testCases := []struct {
		name            string
		msg             types.MsgConvertCosmosCoinToERC20
		amountConverted sdkmath.Int
		expectedErr     string
	}{
		{
			name: "valid - first conversion deploys contract, send to self",
			msg: types.NewMsgConvertCosmosCoinToERC20(
				fundedAccount.String(),
				common.BytesToAddress(fundedAccount.Bytes()).Hex(), // it's me!
				sdk.NewInt64Coin(allowedDenom, 5e7),
			),
			amountConverted: sdkmath.NewInt(5e7),
			expectedErr:     "",
		},
		{
			name: "valid - first conversion deploys contract, send to other",
			msg: types.NewMsgConvertCosmosCoinToERC20(
				fundedAccount.String(),
				testutil.RandomEvmAddress().Hex(), // someone else!
				sdk.NewInt64Coin(allowedDenom, 9993317),
			),
			amountConverted: sdkmath.NewInt(9993317),
			expectedErr:     "",
		},
		{
			name: "invalid - un-allowed denom",
			msg: types.NewMsgConvertCosmosCoinToERC20(
				app.RandomAddress().String(),
				testutil.RandomEvmAddress().Hex(),
				sdk.NewInt64Coin("not-allowed-denom", 1e4),
			),
			expectedErr: "sdk.Coin not enabled to convert to ERC20 token",
		},
		{
			name: "invalid - bad initiator",
			msg: types.NewMsgConvertCosmosCoinToERC20(
				"invalid-kava-address",
				testutil.RandomEvmAddress().Hex(),
				sdk.NewInt64Coin(allowedDenom, 1e4),
			),
			expectedErr: "invalid initiator address",
		},
		{
			name: "invalid - bad receiver",
			msg: types.NewMsgConvertCosmosCoinToERC20(
				app.RandomAddress().String(),
				"invalid-0x-address",
				sdk.NewInt64Coin(allowedDenom, 1e4),
			),
			expectedErr: "invalid receiver address",
		},
		{
			name: "invalid - bad receiver",
			msg: types.NewMsgConvertCosmosCoinToERC20(
				app.RandomAddress().String(),
				"invalid-0x-address",
				sdk.NewInt64Coin(allowedDenom, 1e4),
			),
			expectedErr: "invalid receiver address",
		},
		{
			name: "invalid - insufficient balance",
			msg: types.NewMsgConvertCosmosCoinToERC20(
				fundedAccount.String(),
				testutil.RandomEvmAddress().Hex(),
				sdk.NewInt64Coin(allowedDenom, initialFunding+1),
			),
			expectedErr: "insufficient funds",
		},
		// NOTE: a zero amount tx passes in this scope but will fail to pass ValidateBasic()
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			// initial setup
			setup()

			moduleBalanceBefore := suite.ModuleBalance(allowedDenom)

			// submit message
			_, err := suite.msgServer.ConvertCosmosCoinToERC20(suite.Ctx, &tc.msg)

			// verify error, if expected
			if tc.expectedErr != "" {
				suite.ErrorContains(err, tc.expectedErr)
				// the contract wasn't previously deployed, so still shouldn't be
				_, found := suite.Keeper.GetDeployedCosmosCoinContract(suite.Ctx, allowedDenom)
				suite.False(found)
				return
			}

			// verify success
			suite.NoError(err)

			initiator := sdk.MustAccAddressFromBech32(tc.msg.Initiator)
			receiver := testutil.MustNewInternalEVMAddressFromString(tc.msg.Receiver)

			// initiator no longer has sdk coins
			cosmosBalanceAfter := suite.BankKeeper.GetBalance(suite.Ctx, initiator, allowedDenom)
			suite.Equal(
				sdkmath.NewInt(initialFunding).Sub(tc.amountConverted),
				cosmosBalanceAfter.Amount,
				"unexpected sdk.Coin balance of initiator",
			)

			// sdk coins are locked into module
			moduleBalanceAfter := suite.ModuleBalance(allowedDenom)
			suite.Equal(
				moduleBalanceBefore.Add(tc.amountConverted),
				moduleBalanceAfter,
				"unexpected module balance",
			)

			// deployed contract address is registered in module store
			contractAddress, found := suite.Keeper.GetDeployedCosmosCoinContract(suite.Ctx, allowedDenom)
			suite.True(found, "expected deployed contract address to be registered, found none")

			// receiver has been minted correct number of tokens
			erc20Balance, err := suite.Keeper.QueryERC20BalanceOf(suite.Ctx, contractAddress, receiver)
			suite.NoError(err)
			suite.Equal(tc.amountConverted.BigInt(), erc20Balance, "unexpected erc20 balance for receiver")

			// msg server event
			suite.EventsContains(suite.GetEvents(),
				sdk.NewEvent(
					sdk.EventTypeMessage,
					sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
					sdk.NewAttribute(sdk.AttributeKeySender, initiator.String()),
				))

			// keeper event
			suite.EventsContains(suite.GetEvents(),
				sdk.NewEvent(
					types.EventTypeConvertCosmosCoinToERC20,
					sdk.NewAttribute(types.AttributeKeyInitiator, initiator.String()),
					sdk.NewAttribute(types.AttributeKeyReceiver, receiver.String()),
					sdk.NewAttribute(types.AttributeKeyERC20Address, contractAddress.Hex()),
					sdk.NewAttribute(types.AttributeKeyAmount, tc.msg.Amount.String()),
				))
		})
	}
}

func (suite *MsgServerSuite) TestConvertCosmosCoinToERC20_AlreadyDeployedContract() {
	allowedDenom := "ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2"
	initialFunding := int64(1e10)
	fundedAccount := app.RandomAddress()

	amount := sdkmath.NewInt(6e8)
	receiver1 := types.BytesToInternalEVMAddress(app.RandomAddress().Bytes())
	receiver2 := types.BytesToInternalEVMAddress(app.RandomAddress().Bytes())

	suite.SetupTest()

	// make the denom allowed for conversion
	params := suite.Keeper.GetParams(suite.Ctx)
	params.AllowedCosmosDenoms = types.NewAllowedCosmosCoinERC20Tokens(
		types.NewAllowedCosmosCoinERC20Token(allowedDenom, "Kava EVM Atom", "ATOM", 6),
	)
	suite.Keeper.SetParams(suite.Ctx, params)

	// fund account
	err := suite.App.FundAccount(suite.Ctx, fundedAccount, sdk.NewCoins(
		sdk.NewInt64Coin(allowedDenom, initialFunding),
	))
	suite.NoError(err, "failed to initially fund account")

	// verify contract is not deployed
	_, found := suite.Keeper.GetDeployedCosmosCoinContract(suite.Ctx, allowedDenom)
	suite.False(found)

	// initial convert deploys contract
	msg := types.NewMsgConvertCosmosCoinToERC20(
		fundedAccount.String(),
		receiver1.Hex(),
		sdk.NewCoin(allowedDenom, amount),
	)
	_, err = suite.msgServer.ConvertCosmosCoinToERC20(suite.Ctx, &msg)
	suite.NoError(err)

	contractAddress, found := suite.Keeper.GetDeployedCosmosCoinContract(suite.Ctx, allowedDenom)
	suite.True(found)

	// second convert uses same contract
	msg.Receiver = receiver2.Hex()
	_, err = suite.msgServer.ConvertCosmosCoinToERC20(suite.Ctx, &msg)
	suite.NoError(err)

	after2ndUseAddress, found := suite.Keeper.GetDeployedCosmosCoinContract(suite.Ctx, allowedDenom)
	suite.True(found)
	suite.Equal(contractAddress, after2ndUseAddress, "contract address should remain the same")

	// check balances
	bal1, err := suite.Keeper.QueryERC20BalanceOf(suite.Ctx, contractAddress, receiver1)
	suite.NoError(err)
	suite.Equal(amount.BigInt(), bal1)

	bal2, err := suite.Keeper.QueryERC20BalanceOf(suite.Ctx, contractAddress, receiver2)
	suite.NoError(err)
	suite.Equal(amount.BigInt(), bal2)

	// check total supply
	caller, key := testutil.RandomEvmAccount()
	totalSupply, err := suite.QueryContract(
		types.ERC20KavaWrappedCosmosCoinContract.ABI,
		caller,
		key,
		contractAddress,
		"totalSupply",
	)
	suite.NoError(err)
	suite.Len(totalSupply, 1)
	suite.Equal(amount.MulRaw(2).BigInt(), totalSupply[0].(*big.Int))
}

func (suite *MsgServerSuite) TestConvertCosmosCoinFromERC20() {
	denom := "magic"
	tokenInfo := types.NewAllowedCosmosCoinERC20Token(denom, "Cosmos Coin", "MAGIC", 6)
	initialPosition := sdk.NewInt64Coin(denom, 1e10)
	initiator := testutil.RandomInternalEVMAddress()

	var contractAddress types.InternalEVMAddress
	setup := func() {
		suite.SetupTest()

		// allow conversion to the denom
		params := suite.Keeper.GetParams(suite.Ctx)
		params.AllowedCosmosDenoms = append(params.AllowedCosmosDenoms, tokenInfo)
		suite.Keeper.SetParams(suite.Ctx, params)

		// setup initial position
		addr := app.RandomAddress()
		err := suite.App.FundAccount(suite.Ctx, addr, sdk.NewCoins(initialPosition))
		suite.NoError(err)
		err = suite.Keeper.ConvertCosmosCoinToERC20(suite.Ctx, addr, initiator, initialPosition)
		suite.NoError(err)

		contractAddress, _ = suite.Keeper.GetDeployedCosmosCoinContract(suite.Ctx, denom)
	}

	testCases := []struct {
		name            string
		msg             types.MsgConvertCosmosCoinFromERC20
		amountConverted sdkmath.Int
		expectedErr     string
	}{
		{
			name: "valid - full convert",
			msg: types.NewMsgConvertCosmosCoinFromERC20(
				initiator.Hex(),
				app.RandomAddress().String(),
				initialPosition,
			),
			amountConverted: initialPosition.Amount,
			expectedErr:     "",
		},
		{
			name: "valid - partial convert",
			msg: types.NewMsgConvertCosmosCoinFromERC20(
				initiator.Hex(),
				app.RandomAddress().String(),
				sdk.NewInt64Coin(denom, 123456),
			),
			amountConverted: sdkmath.NewInt(123456),
			expectedErr:     "",
		},
		{
			name: "invalid - bad initiator",
			msg: types.NewMsgConvertCosmosCoinFromERC20(
				"invalid-address",
				app.RandomAddress().String(),
				sdk.NewInt64Coin(denom, 123456),
			),
			amountConverted: sdkmath.ZeroInt(),
			expectedErr:     "invalid initiator address",
		},
		{
			name: "invalid - bad receiver",
			msg: types.NewMsgConvertCosmosCoinFromERC20(
				testutil.RandomEvmAddress().Hex(),
				"invalid-address",
				sdk.NewInt64Coin(denom, 123456),
			),
			amountConverted: sdkmath.ZeroInt(),
			expectedErr:     "invalid receiver address",
		},
		{
			name: "invalid - unsupported asset",
			msg: types.NewMsgConvertCosmosCoinFromERC20(
				initiator.Hex(),
				app.RandomAddress().String(),
				sdk.NewInt64Coin("not-supported", 123456),
			),
			amountConverted: sdkmath.ZeroInt(),
			expectedErr:     "no erc20 contract found",
		},
		{
			name: "invalid - insufficient funds",
			msg: types.NewMsgConvertCosmosCoinFromERC20(
				initiator.Hex(),
				app.RandomAddress().String(),
				initialPosition.AddAmount(sdkmath.OneInt()),
			),
			amountConverted: sdkmath.ZeroInt(),
			expectedErr:     "failed to convert to cosmos coins: insufficient funds",
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			setup()

			_, err := suite.msgServer.ConvertCosmosCoinFromERC20(suite.Ctx, &tc.msg)

			if tc.expectedErr != "" {
				suite.ErrorContains(err, tc.expectedErr)
				// expect no change in erc20 balance
				balance, err := suite.Keeper.QueryERC20BalanceOf(suite.Ctx, contractAddress, initiator)
				suite.NoError(err)
				suite.BigIntsEqual(initialPosition.Amount.BigInt(), balance, "expected no change in initiator's erc20 balance")
				// expect no change in module balance
				suite.Equal(initialPosition.Amount, suite.ModuleBalance(denom), "expected no change in module balance")
				return
			}

			suite.NoError(err)

			receiver := sdk.MustAccAddressFromBech32(tc.msg.Receiver)
			// expect receiver to have the sdk coins
			sdkBalance := suite.BankKeeper.GetBalance(suite.Ctx, receiver, denom)
			suite.Equal(tc.amountConverted, sdkBalance.Amount)

			newEvmBalance := initialPosition.SubAmount(tc.amountConverted)
			// expect initiator to have the balance deducted
			evmBalance, err := suite.Keeper.QueryERC20BalanceOf(suite.Ctx, contractAddress, initiator)
			suite.NoError(err)
			suite.BigIntsEqual(newEvmBalance.Amount.BigInt(), evmBalance, "unexpected initiator final erc20 balance")

			// expect tokens to be deducted from module account
			suite.True(newEvmBalance.Amount.Equal(suite.ModuleBalance(denom)), "unexpected module balance")

			// expect erc20 total supply to reflect new value
			caller, key := testutil.RandomEvmAccount()
			totalSupply, err := suite.QueryContract(
				types.ERC20KavaWrappedCosmosCoinContract.ABI,
				caller,
				key,
				contractAddress,
				"totalSupply",
			)
			suite.NoError(err)
			suite.BigIntsEqual(newEvmBalance.Amount.BigInt(), totalSupply[0].(*big.Int), "unexpected total supply")
		})
	}
}

// the test verifies the behavior for when a denom is removed from the params list
// after conversions have been made:
// - it should prevent more conversions from sdk -> evm for that denom
// - existing erc20s should be allowed to get converted back to a sdk.Coins
// - allowing the denom again should use existing contract
func (suite *MsgServerSuite) TestConvertCosmosCoinForRemovedDenom() {
	denom := "magic"
	tokenInfo := types.NewAllowedCosmosCoinERC20Token(denom, "MAGIC COIN", "MAGIC", 6)
	account := app.RandomAddress()
	evmAddr := types.BytesToInternalEVMAddress(account.Bytes())
	coin := func(amt int64) sdk.Coin { return sdk.NewInt64Coin(denom, amt) }

	// fund account
	suite.NoError(suite.App.FundAccount(suite.Ctx, account, sdk.NewCoins(coin(1e10))))

	// setup the token as allowed
	params := suite.Keeper.GetParams(suite.Ctx)
	params.AllowedCosmosDenoms = append(params.AllowedCosmosDenoms, tokenInfo)
	suite.Keeper.SetParams(suite.Ctx, params)

	// convert some coins while its allowed
	msg := types.NewMsgConvertCosmosCoinToERC20(account.String(), evmAddr.Hex(), coin(5e9))
	_, err := suite.msgServer.ConvertCosmosCoinToERC20(suite.Ctx, &msg)
	suite.NoError(err)

	// expect contract registered
	contractAddress, isRegistered := suite.Keeper.GetDeployedCosmosCoinContract(suite.Ctx, denom)
	suite.True(isRegistered)
	suite.False(contractAddress.IsNil())

	// unregister contract
	params.AllowedCosmosDenoms = []types.AllowedCosmosCoinERC20Token{}
	suite.Keeper.SetParams(suite.Ctx, params)

	suite.Run("disallows sdk -> evm when removed", func() {
		msg := types.NewMsgConvertCosmosCoinToERC20(account.String(), evmAddr.Hex(), coin(5e9))
		_, err := suite.msgServer.ConvertCosmosCoinToERC20(suite.Ctx, &msg)
		suite.ErrorContains(err, "sdk.Coin not enabled to convert to ERC20 token")
	})

	suite.Run("allows conversion of existing ERC20s", func() {
		msg := types.NewMsgConvertCosmosCoinFromERC20(evmAddr.Hex(), account.String(), coin(5e9))
		_, err := suite.msgServer.ConvertCosmosCoinFromERC20(suite.Ctx, &msg)
		suite.NoError(err)

		// should be fully withdrawn
		erc20Bal, err := suite.Keeper.QueryERC20BalanceOf(suite.Ctx, contractAddress, evmAddr)
		suite.NoError(err)
		suite.BigIntsEqual(big.NewInt(0), erc20Bal, "cosmos coins were not converted back")
		sdkBal := suite.BankKeeper.GetBalance(suite.Ctx, account, denom)
		suite.Equal(coin(1e10), sdkBal)
	})

	suite.Run("contract stays registered", func() {
		postDisableContractAddress, found := suite.Keeper.GetDeployedCosmosCoinContract(suite.Ctx, denom)
		suite.True(found)
		suite.Equal(contractAddress, postDisableContractAddress)
	})

	suite.Run("re-enable uses original contract", func() {
		// re-enable contract
		params.AllowedCosmosDenoms = append(params.AllowedCosmosDenoms, tokenInfo)
		suite.Keeper.SetParams(suite.Ctx, params)

		// attempt conversion
		msg := types.NewMsgConvertCosmosCoinToERC20(account.String(), evmAddr.Hex(), coin(1e10))
		_, err := suite.msgServer.ConvertCosmosCoinToERC20(suite.Ctx, &msg)
		suite.NoError(err)

		// should have balance on original ERC20 contract
		erc20Bal, err := suite.Keeper.QueryERC20BalanceOf(suite.Ctx, contractAddress, evmAddr)
		suite.NoError(err)
		suite.BigIntsEqual(big.NewInt(1e10), erc20Bal, "cosmos coins were not converted")
		sdkBal := suite.BankKeeper.GetBalance(suite.Ctx, account, denom)
		suite.True(sdkBal.IsZero())
	})
}
