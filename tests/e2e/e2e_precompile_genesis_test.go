package e2e_test

import (
	"context"
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	ethermint "github.com/evmos/ethermint/types"
	evmtypes "github.com/evmos/ethermint/x/evm/types"

	"github.com/kava-labs/kava/precompile/registry"
)

// TestPrecompileGenesis tests that the following is true for enabled precompiles:
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
