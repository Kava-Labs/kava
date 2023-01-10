package keeper_test

import (
	"fmt"
	"math/big"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/stretchr/testify/suite"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/community/keeper"
	"github.com/kava-labs/kava/x/community/types"
	evmutilkeeper "github.com/kava-labs/kava/x/evmutil/keeper"
	evmutiltestutil "github.com/kava-labs/kava/x/evmutil/testutil"
	evmutiltypes "github.com/kava-labs/kava/x/evmutil/types"
)

// Test suite used for all keeper tests
type proposalHandlerTestSuite struct {
	evmutiltestutil.Suite
	Keeper keeper.Keeper
}

// The default state used by each test
func (suite *proposalHandlerTestSuite) SetupTest() {
	suite.Suite.SetupTest()
	suite.Keeper = suite.App.GetCommunityKeeper()
}

func (suite *proposalHandlerTestSuite) TestHandleCommunityPoolProposal_DifferentMsgs() {
	params := suite.Keeper.GetParams(suite.Ctx)
	params.EnabledProposalMsgUrls = []string{
		sdk.MsgTypeURL(&evmutiltypes.MsgEVMCall{}),
		sdk.MsgTypeURL(&banktypes.MsgMultiSend{}),
		sdk.MsgTypeURL(&banktypes.MsgSend{}),
	}
	userAddr1 := app.RandomAddress()
	userAddr2 := app.RandomAddress()
	userEvmAddr1 := common.BytesToAddress(userAddr1.Bytes())

	authorityModule := types.ModuleAccountName
	authorityAct := suite.App.GetAccountKeeper().GetModuleAccount(suite.Ctx, authorityModule)
	authorityAddr := authorityAct.GetAddress().String()
	authorityStartingUKavaBal := sdk.NewInt(1000)
	authorityCoin := sdk.NewCoin("ukava", authorityStartingUKavaBal)

	type errArgs struct {
		expectPass bool
		contains   string
	}
	tests := []struct {
		name         string
		msgs         []sdk.Msg
		authorityBal sdk.Int
		user1Bal     sdk.Int
		user2Bal     sdk.Int
		errArgs      errArgs
	}{
		{
			"valid - multiple kava transfers",
			[]sdk.Msg{
				&evmutiltypes.MsgEVMCall{
					To:        userEvmAddr1.String(),
					Authority: authorityAddr,
					Amount:    evmutilkeeper.ConversionMultiplier.Mul(sdk.NewInt(120)),
				},
				&banktypes.MsgMultiSend{
					Inputs: []banktypes.Input{
						{Address: authorityAddr, Coins: sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(150)))},
					},
					Outputs: []banktypes.Output{
						{Address: userAddr1.String(), Coins: sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(100)))},
						{Address: userAddr2.String(), Coins: sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(50)))},
					},
				},
				&banktypes.MsgSend{
					FromAddress: authorityAddr,
					ToAddress:   userAddr1.String(),
					Amount:      sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(30))),
				},
			},
			sdk.NewInt(700),
			sdk.NewInt(250),
			sdk.NewInt(50),
			errArgs{
				expectPass: true,
			},
		},
		{
			"fails - insufficient transfers will revert all state changes",
			[]sdk.Msg{
				&evmutiltypes.MsgEVMCall{
					To:        userEvmAddr1.String(),
					Authority: authorityAddr,
					Amount:    evmutilkeeper.ConversionMultiplier.Mul(sdk.NewInt(120)),
				},
				&banktypes.MsgSend{
					FromAddress: authorityAddr,
					ToAddress:   userAddr1.String(),
					Amount:      sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(900))),
				},
			},
			sdk.NewInt(1000),
			sdk.NewInt(0),
			sdk.NewInt(0),
			errArgs{
				expectPass: false,
				contains:   "insufficient funds",
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			suite.SetupTest()
			suite.Keeper.SetParams(suite.Ctx, params)
			err := suite.App.FundModuleAccount(suite.Ctx, authorityModule, sdk.NewCoins(authorityCoin))
			suite.Require().NoError(err)

			proposal, err := types.NewCommunityPoolProposal(
				"title",
				"description",
				tc.msgs,
			)
			suite.Require().NoError(err)

			err = keeper.HandleCommunityPoolProposal(suite.Ctx, suite.Keeper, proposal)

			if tc.errArgs.expectPass {
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
				suite.Require().Contains(err.Error(), tc.errArgs.contains)
			}

			// validate balances
			authorityBal := suite.App.GetBankKeeper().GetBalance(suite.Ctx, authorityAct.GetAddress(), "ukava")
			suite.Require().Equal(tc.authorityBal, authorityBal.Amount, "community pool ukava balance is invalid")
			user1Bal := suite.App.GetBankKeeper().GetBalance(suite.Ctx, userAddr1, "ukava")
			suite.Require().Equal(tc.user1Bal, user1Bal.Amount, "user 1 ukava balance is invalid")
			user2Bal := suite.App.GetBankKeeper().GetBalance(suite.Ctx, userAddr2, "ukava")
			suite.Require().Equal(tc.user2Bal, user2Bal.Amount, "user 2 ukava balance is invalid")
		})
	}
}

func (suite *proposalHandlerTestSuite) TestHandleCommunityPoolProposal_EVMCall() {
	userAddr := app.RandomAddress()
	userEvmAddr := common.BytesToAddress(userAddr.Bytes())
	transferContractAddr := suite.DeployERC20()
	depositContractAddr := suite.DeployERC20()

	authorityModule := types.ModuleAccountName
	authorityAct := suite.App.GetAccountKeeper().GetModuleAccount(suite.Ctx, authorityModule)
	authorityAddr := authorityAct.GetAddress().String()
	authorityEvmAddr := evmutiltypes.NewInternalEVMAddress(common.BytesToAddress(authorityAct.GetAddress().Bytes()))
	erc20StartingBal := big.NewInt(1000)
	authorityStartingUKavaBal := sdk.NewInt(1000)

	transferFnAbi := `{
		"inputs": [
			{ "type": "address", "name": "to" },
			{ "type": "uint256", "name": "amount" }
		],
		"name": "transfer",
		"type": "function"
	}`
	transferData := encodeTransferFn(userEvmAddr, 30)
	depositFnAbi := `{
		"inputs": [],
		"name": "deposit",
		"type": "function"
	}`
	depositData := fmt.Sprintf("0x%s", "d0e30db0")

	type errArgs struct {
		expectPass bool
		contains   string
	}

	tests := []struct {
		name       string
		msgs       []sdk.Msg
		expKavaBal sdk.Int
		expUsdcBal sdk.Int
		errArgs    errArgs
	}{
		{
			"valid - proposal with multiple calls (transfer & deposit)",
			[]sdk.Msg{
				&evmutiltypes.MsgEVMCall{
					To:        transferContractAddr.String(),
					FnAbi:     transferFnAbi,
					Data:      transferData,
					Authority: authorityAddr,
					Amount:    sdk.ZeroInt(),
				},
				&evmutiltypes.MsgEVMCall{
					To:        depositContractAddr.String(),
					FnAbi:     depositFnAbi,
					Data:      depositData,
					Authority: authorityAddr,
					Amount:    evmutilkeeper.ConversionMultiplier.Mul(sdk.NewInt(120)),
				},
				&evmutiltypes.MsgEVMCall{
					To:        transferContractAddr.String(),
					FnAbi:     transferFnAbi,
					Data:      transferData,
					Authority: authorityAddr,
					Amount:    sdk.ZeroInt(),
				},
			},
			sdk.NewInt(880),
			sdk.NewInt(940),
			errArgs{
				expectPass: true,
			},
		},
		{
			"valid - proposal with a single call (deposit)",
			[]sdk.Msg{
				&evmutiltypes.MsgEVMCall{
					To:        depositContractAddr.String(),
					FnAbi:     depositFnAbi,
					Data:      depositData,
					Authority: authorityAddr,
					Amount:    evmutilkeeper.ConversionMultiplier.Mul(sdk.NewInt(120)),
				},
			},
			sdk.NewInt(880),
			sdk.NewInt(1000),
			errArgs{
				expectPass: true,
			},
		},
		{
			"valid - proposal using MsgEVMCall as normal akava transfer",
			[]sdk.Msg{
				&evmutiltypes.MsgEVMCall{
					To:        userEvmAddr.String(),
					Authority: authorityAddr,
					Amount:    evmutilkeeper.ConversionMultiplier.Mul(sdk.NewInt(100)),
				},
			},
			sdk.NewInt(900),
			sdk.NewInt(1000),
			errArgs{
				expectPass: true,
			},
		},
		{
			"invalid - throws if proposal contains an invalid MsgEVMCall",
			[]sdk.Msg{
				&evmutiltypes.MsgEVMCall{
					To:        depositContractAddr.String(),
					FnAbi:     depositFnAbi,
					Data:      depositData,
					Authority: authorityAddr,
					Amount:    evmutilkeeper.ConversionMultiplier.Mul(sdk.NewInt(120)),
				},
				&evmutiltypes.MsgEVMCall{
					To:        depositContractAddr.String(),
					FnAbi:     depositFnAbi,
					Data:      fmt.Sprintf("0x%s", "8aa7c88f"), // call non-existent function
					Authority: authorityAddr,
					Amount:    sdk.ZeroInt(),
				},
			},
			sdk.NewInt(1000), // successful msgs should have have changed state
			sdk.NewInt(1000),
			errArgs{
				expectPass: false,
				contains:   "method not found in fnAbi: 0x8aa7c88f",
			},
		},
		{
			"invalid - contains MsgEVMCall with invalid authority",
			[]sdk.Msg{
				&evmutiltypes.MsgEVMCall{
					To:        depositContractAddr.String(),
					FnAbi:     depositFnAbi,
					Data:      depositData,
					Authority: userAddr.String(),
					Amount:    evmutilkeeper.ConversionMultiplier.Mul(sdk.NewInt(120)),
				},
				&evmutiltypes.MsgEVMCall{
					To:        depositContractAddr.String(),
					FnAbi:     depositFnAbi,
					Data:      depositData,
					Authority: authorityAddr,
					Amount:    evmutilkeeper.ConversionMultiplier.Mul(sdk.NewInt(120)),
				},
			},
			sdk.NewInt(1000),
			sdk.NewInt(1000),
			errArgs{
				expectPass: false,
				contains:   "failed on execution: invalid signer",
			},
		},
		{
			"invalid - msgs data is not hex string",
			[]sdk.Msg{
				&evmutiltypes.MsgEVMCall{
					To:        transferContractAddr.String(),
					FnAbi:     transferFnAbi,
					Data:      "0xinvalid_data",
					Authority: authorityAddr,
					Amount:    sdk.ZeroInt(),
				},
			},
			sdk.NewInt(1000),
			sdk.NewInt(1000),
			errArgs{
				expectPass: false,
				contains:   "invalid data format: invalid hex string",
			},
		},
		{
			"invalid - authority address",
			[]sdk.Msg{
				&evmutiltypes.MsgEVMCall{
					To:        transferContractAddr.String(),
					FnAbi:     transferFnAbi,
					Data:      transferData,
					Authority: authorityEvmAddr.String(),
					Amount:    sdk.ZeroInt(),
				},
			},
			sdk.NewInt(1000),
			sdk.NewInt(1000),
			errArgs{
				expectPass: false,
				contains:   "failed on execution: invalid signer",
			},
		},
		{
			"invalid - evm call 'To' property",
			[]sdk.Msg{
				&evmutiltypes.MsgEVMCall{
					To:        userAddr.String(),
					FnAbi:     transferFnAbi,
					Data:      transferData,
					Authority: authorityAddr,
					Amount:    sdk.ZeroInt(),
				},
			},
			sdk.NewInt(1000),
			sdk.NewInt(1000),
			errArgs{
				expectPass: false,
				contains:   fmt.Sprintf("to '%s' is not hex address", userAddr),
			},
		},
		{
			"invalid - use MsgEVMCall to send kava to nonpayable contract",
			[]sdk.Msg{
				&evmutiltypes.MsgEVMCall{
					To:        depositContractAddr.String(),
					Authority: authorityAddr,
					Amount:    evmutilkeeper.ConversionMultiplier.Mul(sdk.NewInt(100)),
				},
			},
			sdk.NewInt(1000),
			sdk.NewInt(1000),
			errArgs{
				expectPass: false,
				contains:   "evm call failed: execution reverted",
			},
		},
		{
			"invalid - not enabled msgs are included with enabled ones",
			[]sdk.Msg{
				&evmutiltypes.MsgEVMCall{
					To:        userEvmAddr.String(),
					Authority: authorityAddr,
					Amount:    evmutilkeeper.ConversionMultiplier.Mul(sdk.NewInt(50)),
				},
				// MsgSend is not enabled
				&banktypes.MsgSend{
					FromAddress: authorityAddr,
					ToAddress:   userAddr.String(),
					Amount:      sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(100))),
				},
				&evmutiltypes.MsgEVMCall{
					To:        userEvmAddr.String(),
					Authority: authorityAddr,
					Amount:    evmutilkeeper.ConversionMultiplier.Mul(sdk.NewInt(100)),
				},
			},
			sdk.NewInt(1000),
			sdk.NewInt(1000),
			errArgs{
				expectPass: false,
				contains:   "CommunityPoolProposal msg 1 (/cosmos.bank.v1beta1.MsgSend) failed on execution: msg not enabled via params: community pool proposal message execution error",
			},
		},
		{
			"invalid - no changes if no msgs are enabled",
			[]sdk.Msg{
				&banktypes.MsgSend{
					FromAddress: authorityAddr,
					ToAddress:   userAddr.String(),
					Amount:      sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(100))),
				},
				&banktypes.MsgSend{
					FromAddress: authorityAddr,
					ToAddress:   userAddr.String(),
					Amount:      sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(50))),
				},
			},
			sdk.NewInt(1000),
			sdk.NewInt(1000),
			errArgs{
				expectPass: false,
				contains:   "CommunityPoolProposal msg 0 (/cosmos.bank.v1beta1.MsgSend) failed on execution: msg not enabled via params: community pool proposal message execution error",
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			suite.SetupTest()
			transferContractAddr := suite.DeployERC20()
			suite.DeployERC20() // deploy deposit contract

			// give invoker account some erc20 and kava to being with
			err := suite.App.GetEvmutilKeeper().MintERC20(
				suite.Ctx,
				transferContractAddr,
				authorityEvmAddr,
				erc20StartingBal,
			)
			suite.Require().NoError(err)
			authorityCoin := sdk.NewCoin("ukava", authorityStartingUKavaBal)
			err = suite.App.FundModuleAccount(suite.Ctx, authorityModule, sdk.NewCoins(authorityCoin))
			suite.Require().NoError(err)

			proposal, err := types.NewCommunityPoolProposal(
				"title",
				"description",
				tc.msgs,
			)
			suite.Require().NoError(err)

			err = keeper.HandleCommunityPoolProposal(suite.Ctx, suite.Keeper, proposal)

			if tc.errArgs.expectPass {
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
				suite.Require().Contains(err.Error(), tc.errArgs.contains)
			}

			// validate authority kava balance
			coinBal := suite.App.GetBankKeeper().GetBalance(suite.Ctx, authorityAct.GetAddress(), "ukava")
			suite.Require().Equal(tc.expKavaBal, coinBal.Amount, "user ukava balance is invalid")

			// validate authority erc20 balance after msg
			bal, err := suite.App.GetEvmutilKeeper().QueryERC20BalanceOf(
				suite.Ctx,
				transferContractAddr,
				authorityEvmAddr,
			)
			suite.Require().NoError(err)
			suite.Require().Equal(tc.expUsdcBal.BigInt(), bal, "user erc20 balance is invalid")
		})
	}
}

func TestProposalHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(proposalHandlerTestSuite))
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
