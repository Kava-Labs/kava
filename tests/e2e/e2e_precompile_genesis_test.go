package e2e_test

import (
	"context"
	"fmt"
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	ethermint "github.com/evmos/ethermint/types"
	evmtypes "github.com/evmos/ethermint/x/evm/types"

	"github.com/kava-labs/kava/precompile/registry"
)

// TestPrecompileGenesis tests that the the following is true for enabled precompiles:
//
//   - An enabled precompile has an EthAccount with matching code hash,
//     sequence of 1, and no public key
//   - An enabled precompile has code equal to 0x01
//   - An enabled precompile has a nonce of 1
//
// This is important to ensure the genesis setup for precompiles is correct.
func (suite *IntegrationTestSuite) TestPrecompileGenesis() {
	type fixture struct {
		address            string
		expectIsEnabled    bool
		expectIsEthAccount bool
		expectCode         []byte
		expectNonce        uint64
	}

	// enabled represnets the expected state for a registered precompile
	// that is enabled at genesis
	enabled := func(address string) func() fixture {
		return func() fixture {
			return fixture{
				address:            address,
				expectIsEnabled:    true,
				expectIsEthAccount: true,
				expectCode:         []byte{0x01},
				expectNonce:        uint64(1),
			}
		}
	}

	// disabled represents the expected state for a registered precompile
	// that is not enabled at genesis
	disabled := func(address string) func() fixture {
		return func() fixture {
			return fixture{
				address:            address,
				expectIsEnabled:    false,
				expectIsEthAccount: false,
				expectCode:         []byte{},
				expectNonce:        uint64(0),
			}
		}
	}

	testCases := []struct {
		name       string
		genFixture func() fixture
	}{
		{
			name:       "noop contract address is enabled and initialized",
			genFixture: enabled(registry.NoopContractAddress),
		},
		{
			name:       "noop contract address second address is disabled and not initialized",
			genFixture: disabled(registry.NoopContractAddress2),
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			tf := tc.genFixture()

			fmt.Println("tf", tf)

			//
			// Addresses
			//
			evmAddress := common.HexToAddress(tf.address)
			sdkAddress := sdk.AccAddress(evmAddress.Bytes())

			//
			// Heights
			//
			// We ensure all queries happen at block 1 after genesis
			// and help ensure determisitc behavior
			genesisHeight := big.NewInt(1)
			grpcGenesisContext := suite.Kava.Grpc.CtxAtHeight(genesisHeight.Int64())

			//
			// Queries
			//
			evmParamsResp, err := suite.Kava.Grpc.Query.Evm.Params(grpcGenesisContext, &evmtypes.QueryParamsRequest{})
			suite.Require().NoError(err)

			fmt.Println("evmParamsResp", evmParamsResp.Params)

			// accountErr is checked during in the assertions below if tf.expectIsEthAccount is true
			// This is due to the service returning a not found error for address that are not enabled,
			// which can be ignored when we do not expect an eth account to exist.
			accountResponse, accountErr := suite.Kava.Grpc.Query.Auth.Account(
				grpcGenesisContext, &authtypes.QueryAccountRequest{Address: sdkAddress.String()})
			var account authtypes.AccountI
			if accountErr == nil {
				err = suite.Kava.EncodingConfig.Marshaler.UnpackAny(accountResponse.Account, &account)
				suite.Require().NoError(err)
			}

			// We ensure both the evm json rpc and x/evm grpc code endpoints return the same value
			// in the assertions below
			grpcCodeResponse, err := suite.Kava.Grpc.Query.Evm.Code(grpcGenesisContext,
				&evmtypes.QueryCodeRequest{Address: evmAddress.String()})
			suite.Require().NoError(err)
			rpcCode, err := suite.Kava.EvmClient.CodeAt(context.Background(), evmAddress, genesisHeight)
			suite.Require().NoError(err)

			nonce, err := suite.Kava.EvmClient.NonceAt(context.Background(), evmAddress, genesisHeight)
			suite.Require().NoError(err)

			fmt.Println("nonce", nonce)

			// evmParamsResp params:<evm_denom:"akava" enable_create:true enable_call:true chain_config:<homestead_block:"0" dao_fork_block:"0" dao_fork_support:true eip150_block:"0" eip150_hash:"0x0000000000000000000000000000000000000000000000000000000000000000" eip155_block:"0" eip158_block:"0" byzantium_block:"0" constantinople_block:"0" petersburg_block:"0" istanbul_block:"0" muir_glacier_block:"0" berlin_block:"0" > eip712_allowed_msgs:<msg_type_url:"/kava.evmutil.v1beta1.MsgConvertERC20ToCoin" msg_value_type_name:"MsgValueEVMConvertERC20ToCoin" value_types:<name:"initiator" type:"string" > value_types:<name:"receiver" type:"string" > value_types:<name:"kava_erc20_address" type:"string" > value_types:<name:"amount" type:"string" > > eip712_allowed_msgs:<msg_type_url:"/kava.evmutil.v1beta1.MsgConvertCoinToERC20" msg_value_type_name:"MsgValueEVMConvertCoinToERC20" value_types:<name:"initiator" type:"string" > value_types:<name:"receiver" type:"string" > value_types:<name:"amount" type:"Coin" > > eip712_allowed_msgs:<msg_type_url:"/kava.earn.v1beta1.MsgDeposit" msg_value_type_name:"MsgValueEarnDeposit" value_types:<name:"depositor" type:"string" > value_types:<name:"amount" type:"Coin" > value_types:<name:"strategy" type:"int32" > > eip712_allowed_msgs:<msg_type_url:"/kava.earn.v1beta1.MsgWithdraw" msg_value_type_name:"MsgValueEarnWithdraw" value_types:<name:"from" type:"string" > value_types:<name:"amount" type:"Coin" > value_types:<name:"strategy" type:"int32" > > eip712_allowed_msgs:<msg_type_url:"/cosmos.staking.v1beta1.MsgDelegate" msg_value_type_name:"MsgValueStakingDelegate" value_types:<name:"delegator_address" type:"string" > value_types:<name:"validator_address" type:"string" > value_types:<name:"amount" type:"Coin" > > eip712_allowed_msgs:<msg_type_url:"/cosmos.staking.v1beta1.MsgUndelegate" msg_value_type_name:"MsgValueStakingUndelegate" value_types:<name:"delegator_address" type:"string" > value_types:<name:"validator_address" type:"string" > value_types:<name:"amount" type:"Coin" > > eip712_allowed_msgs:<msg_type_url:"/cosmos.staking.v1beta1.MsgBeginRedelegate" msg_value_type_name:"MsgValueStakingBeginRedelegate" value_types:<name:"delegator_address" type:"string" > value_types:<name:"validator_src_address" type:"string" > value_types:<name:"validator_dst_address" type:"string" > value_types:<name:"amount" type:"Coin" > > eip712_allowed_msgs:<msg_type_url:"/kava.incentive.v1beta1.MsgClaimUSDXMintingReward" msg_value_type_name:"MsgValueIncentiveClaimUSDXMintingReward" value_types:<name:"sender" type:"string" > value_types:<name:"multiplier_name" type:"string" > > eip712_allowed_msgs:<msg_type_url:"/kava.incentive.v1beta1.MsgClaimHardReward" msg_value_type_name:"MsgValueIncentiveClaimHardReward" value_types:<name:"sender" type:"string" > value_types:<name:"denoms_to_claim" type:"IncentiveSelection[]" > nested_types:<name:"IncentiveSelection" attrs:<name:"denom" type:"string" > attrs:<name:"multiplier_name" type:"string" > > > eip712_allowed_msgs:<msg_type_url:"/kava.incentive.v1beta1.MsgClaimDelegatorReward" msg_value_type_name:"MsgValueIncentiveClaimDelegatorReward" value_types:<name:"sender" type:"string" > value_types:<name:"denoms_to_claim" type:"IncentiveSelection[]" > nested_types:<name:"IncentiveSelection" attrs:<name:"denom" type:"string" > attrs:<name:"multiplier_name" type:"string" > > > eip712_allowed_msgs:<msg_type_url:"/kava.incentive.v1beta1.MsgClaimSwapReward" msg_value_type_name:"MsgValueIncentiveClaimSwapReward" value_types:<name:"sender" type:"string" > value_types:<name:"denoms_to_claim" type:"IncentiveSelection[]" > nested_types:<name:"IncentiveSelection" attrs:<name:"denom" type:"string" > attrs:<name:"multiplier_name" type:"string" > > > eip712_allowed_msgs:<msg_type_url:"/kava.incentive.v1beta1.MsgClaimSavingsReward" msg_value_type_name:"MsgValueIncentiveClaimSavingsReward" value_types:<name:"sender" type:"string" > value_types:<name:"denoms_to_claim" type:"IncentiveSelection[]" > nested_types:<name:"IncentiveSelection" attrs:<name:"denom" type:"string" > attrs:<name:"multiplier_name" type:"string" > > > eip712_allowed_msgs:<msg_type_url:"/kava.incentive.v1beta1.MsgClaimEarnReward" msg_value_type_name:"MsgValueIncentiveClaimEarnReward" value_types:<name:"sender" type:"string" > value_types:<name:"denoms_to_claim" type:"IncentiveSelection[]" > nested_types:<name:"IncentiveSelection" attrs:<name:"denom" type:"string" > attrs:<name:"multiplier_name" type:"string" > > > eip712_allowed_msgs:<msg_type_url:"/kava.router.v1beta1.MsgMintDeposit" msg_value_type_name:"MsgValueRouterMintDeposit" value_types:<name:"depositor" type:"string" > value_types:<name:"validator" type:"string" > value_types:<name:"amount" type:"Coin" > > eip712_allowed_msgs:<msg_type_url:"/kava.router.v1beta1.MsgDelegateMintDeposit" msg_value_type_name:"MsgValueRouterDelegateMintDeposit" value_types:<name:"depositor" type:"string" > value_types:<name:"validator" type:"string" > value_types:<name:"amount" type:"Coin" > > eip712_allowed_msgs:<msg_type_url:"/kava.router.v1beta1.MsgWithdrawBurn" msg_value_type_name:"MsgValueRouterWithdrawBurn" value_types:<name:"from" type:"string" > value_types:<name:"validator" type:"string" > value_types:<name:"amount" type:"Coin" > > eip712_allowed_msgs:<msg_type_url:"/kava.router.v1beta1.MsgWithdrawBurnUndelegate" msg_value_type_name:"MsgValueRouterWithdrawBurnUndelegate" value_types:<name:"from" type:"string" > value_types:<name:"validator" type:"string" > value_types:<name:"amount" type:"Coin" > > eip712_allowed_msgs:<msg_type_url:"/cosmos.gov.v1beta1.MsgVote" msg_value_type_name:"MsgValueGovVote" value_types:<name:"proposal_id" type:"uint64" > value_types:<name:"voter" type:"string" > value_types:<name:"option" type:"int32" > > eip712_allowed_msgs:<msg_type_url:"/cosmos.bank.v1beta1.MsgSend" msg_value_type_name:"MsgValueBankSend" value_types:<name:"from_address" type:"string" > value_types:<name:"to_address" type:"string" > value_types:<name:"amount" type:"Coin[]" > > eip712_allowed_msgs:<msg_type_url:"/kava.liquid.v1beta1.MsgMintDerivative" msg_value_type_name:"MsgValueMintDerivative" value_types:<name:"sender" type:"string" > value_types:<name:"validator" type:"string" > value_types:<name:"amount" type:"Coin" > > eip712_allowed_msgs:<msg_type_url:"/kava.liquid.v1beta1.MsgBurnDerivative" msg_value_type_name:"MsgValueBurnDerivative" value_types:<name:"sender" type:"string" > value_types:<name:"validator" type:"string" > value_types:<name:"amount" type:"Coin" > > eip712_allowed_msgs:<msg_type_url:"/kava.hard.v1beta1.MsgWithdraw" msg_value_type_name:"MsgValueHardWithdraw" value_types:<name:"depositor" type:"string" > value_types:<name:"amount" type:"Coin[]" > > eip712_allowed_msgs:<msg_type_url:"/kava.hard.v1beta1.MsgDeposit" msg_value_type_name:"MsgValueHardDeposit" value_types:<name:"depositor" type:"string" > value_types:<name:"amount" type:"Coin[]" > > eip712_allowed_msgs:<msg_type_url:"/kava.evmutil.v1beta1.MsgConvertCosmosCoinToERC20" msg_value_type_name:"MsgConvertCosmosCoinToERC20" value_types:<name:"initiator" type:"string" > value_types:<name:"receiver" type:"string" > value_types:<name:"amount" type:"Coin" > > eip712_allowed_msgs:<msg_type_url:"/kava.evmutil.v1beta1.MsgConvertCosmosCoinFromERC20" msg_value_type_name:"MsgConvertCosmosCoinFromERC20" value_types:<name:"initiator" type:"string" > value_types:<name:"receiver" type:"string" > value_types:<name:"amount" type:"Coin" > > eip712_allowed_msgs:<msg_type_url:"/cosmos.distribution.v1beta1.MsgWithdrawDelegatorReward" msg_value_type_name:"MsgValueWithdrawDelegatorReward" value_types:<name:"delegator_address" type:"string" > value_types:<name:"validator_address" type:"string" > > eip712_allowed_msgs:<msg_type_url:"/ibc.applications.transfer.v1.MsgTransfer" msg_value_type_name:"MsgValueIBCTransfer" value_types:<name:"source_port" type:"string" > value_types:<name:"source_channel" type:"string" > value_types:<name:"token" type:"Coin" > value_types:<name:"sender" type:"string" > value_types:<name:"receiver" type:"string" > value_types:<name:"timeout_height" type:"Height" > value_types:<name:"timeout_timestamp" type:"uint64" > value_types:<name:"memo" type:"string" > nested_types:<name:"Height" attrs:<name:"revision_number" type:"uint64" > attrs:<name:"revision_height" type:"uint64" > > > eip712_allowed_msgs:<msg_type_url:"/kava.cdp.v1beta1.MsgCreateCDP" msg_value_type_name:"MsgValueCdpCreateCDP" value_types:<name:"sender" type:"string" > value_types:<name:"collateral" type:"Coin" > value_types:<name:"principal" type:"Coin" > value_types:<name:"collateral_type" type:"string" > > eip712_allowed_msgs:<msg_type_url:"/kava.cdp.v1beta1.MsgDeposit" msg_value_type_name:"MsgValueCdpDeposit" value_types:<name:"depositor" type:"string" > value_types:<name:"owner" type:"string" > value_types:<name:"collateral" type:"Coin" > value_types:<name:"collateral_type" type:"string" > > eip712_allowed_msgs:<msg_type_url:"/kava.cdp.v1beta1.MsgWithdraw" msg_value_type_name:"MsgValueCdpWithdraw" value_types:<name:"depositor" type:"string" > value_types:<name:"owner" type:"string" > value_types:<name:"collateral" type:"Coin" > value_types:<name:"collateral_type" type:"string" > > eip712_allowed_msgs:<msg_type_url:"/kava.cdp.v1beta1.MsgDrawDebt" msg_value_type_name:"MsgValueCdpDrawDebt" value_types:<name:"sender" type:"string" > value_types:<name:"collateral_type" type:"string" > value_types:<name:"principal" type:"Coin" > > eip712_allowed_msgs:<msg_type_url:"/kava.cdp.v1beta1.MsgRepayDebt" msg_value_type_name:"MsgValueCdpRepayDebt" value_types:<name:"sender" type:"string" > value_types:<name:"collateral_type" type:"string" > value_types:<name:"payment" type:"Coin" > > >

			//
			// Assertions
			//
			if tf.expectIsEnabled {
				suite.Containsf(evmParamsResp.Params.EnabledPrecompiles, tf.address,
					"expected %s to be enabled in evm params", tf.address)
			} else {
				suite.NotContainsf(evmParamsResp.Params.EnabledPrecompiles, tf.address,
					"expected %s to not be enabled in evm params", tf.address)
			}

			// Fail if fixuture configuration is invalid
			if tf.expectIsEthAccount && len(tf.expectCode) == 0 {
				suite.Failf("an eth account must have expected code for address %s", tf.address)
			}

			// Run addition EthAccount assertions
			if tf.expectIsEthAccount {
				suite.Require().NoErrorf(accountErr, "expected account query to not error for address %s", tf.address)

				// All contracts including precompiles must be EthAccount's
				ethAccount, isEthAccount := account.(*ethermint.EthAccount)
				suite.Require().Truef(isEthAccount, "expected account at address %s to be an eth account", tf.address)

				// Code hash must always match the EthAccount
				codeHash := ethAccount.GetCodeHash()
				suite.Equalf(crypto.Keccak256Hash(tf.expectCode), codeHash,
					"expected codehash for account %s to match expected code", tf.address)

				// A precompile (and contract) should never have a public key set
				suite.Nilf(ethAccount.PubKey, "expected account %s to have no public key", tf.address)

				// Assert the account sequence matches the expected nonce
				// This a duplicate of the nonce assertion below, but also helps ensure these
				// two sources agree
				suite.Equal(tf.expectNonce, ethAccount.GetSequence())
			}

			fmt.Println("tf.expectCode", tf.expectCode)
			fmt.Println("rpcCode", rpcCode)
			fmt.Println("tf.address", tf.address)
			// We assert both methods of code retrieval report the same value
			suite.Equalf(tf.expectCode, rpcCode, "expected code for address %s to match expected", tf.address)
			// The GRPC endpoint returns []byte(nil) when the code is not in state, which is different from
			// the rpc endpoint that returns []byte{}.
			grpcExpectedCode := tf.expectCode
			if len(grpcExpectedCode) == 0 {
				grpcExpectedCode = []byte(nil)
			}
			suite.Equalf(grpcExpectedCode, grpcCodeResponse.Code, "expected code for address %s to match expected", tf.address)
			// We assert this outside of the account context since the evm rpc always returns a nonce
			suite.Equalf(tf.expectNonce, nonce, "expected nonce for address %s to match expected", tf.address)
		})
	}
}
