package main_test

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"

	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	sdk "github.com/cosmos/cosmos-sdk/types"
	gov1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	paramsutils "github.com/cosmos/cosmos-sdk/x/params/client/utils"
	"github.com/cosmos/go-bip39"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/evmos/ethermint/crypto/ethsecp256k1"

	"github.com/strangelove-ventures/interchaintest/v7"
	"github.com/strangelove-ventures/interchaintest/v7/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v7/ibc"
	"github.com/strangelove-ventures/interchaintest/v7/testreporter"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/client/erc20"
	"github.com/kava-labs/kava/tests/e2e/runner"
	"github.com/kava-labs/kava/tests/e2e/testutil"
	kavainterchain "github.com/kava-labs/kava/tests/interchain"
	"github.com/kava-labs/kava/tests/util"
	evmutiltypes "github.com/kava-labs/kava/x/evmutil/types"
)

func TestInterchainErc20(t *testing.T) {
	app.SetSDKConfig()
	ctx := context.Background()

	// setup chains
	cf := interchaintest.NewBuiltinChainFactory(zaptest.NewLogger(t), []*interchaintest.ChainSpec{
		{Name: "kava", ChainConfig: kavainterchain.DefaultKavaChainConfig(kavainterchain.KavaTestChainId)},
	})

	chains, err := cf.Chains(t.Name())
	require.NoError(t, err)

	ictKava := chains[0].(*cosmos.CosmosChain)

	client, network := interchaintest.DockerSetup(t)

	// configure interchain
	ic := interchaintest.NewInterchain().AddChain(ictKava)

	// Log location
	f, err := interchaintest.CreateLogFile(fmt.Sprintf("%d.json", time.Now().Unix()))
	require.NoError(t, err)
	// Reporter/logs
	rep := testreporter.NewReporter(f)
	eRep := rep.RelayerExecReporter(t)

	// Build interchain
	err = ic.Build(ctx, eRep, interchaintest.InterchainBuildOptions{
		TestName:         t.Name(),
		Client:           client,
		NetworkID:        network,
		SkipPathCreation: false},
	)
	require.NoError(t, err)

	// Create and Fund User Wallets
	fundAmount := math.NewInt(1e12)

	users := interchaintest.GetAndFundTestUsers(t, ctx, t.Name(), fundAmount, ictKava)
	kavaUser := users[0]

	// wait for new block to ensure initial funding complete
	height, err := ictKava.Height(ctx)
	require.NoError(t, err)
	h := height
	for h <= height {
		h, err = ictKava.Height(ctx)
		require.NoError(t, err)
	}

	// check initial balance
	kavaUserBalInitial, err := ictKava.GetBalance(ctx, kavaUser.FormattedAddress(), ictKava.Config().Denom)
	require.NoError(t, err)
	require.True(t, kavaUserBalInitial.Equal(fundAmount))

	// for simplified management of the chain, use kava's e2e framework for account management
	// this skirts problems in interchaintest with needing coin type 60
	// there are exceptions in the relayer & ibc channel management that complicate setting the chain
	// default coin type to 60 in the chain config.
	// we need to fund an account and then all of kava's e2e testutil chain management will work.

	rpcUrl, err := ictKava.FullNodes[0].GetHostAddress(ctx, "26657/tcp")
	require.NoError(t, err, "failed to find rpc URL")
	grpcUrl, err := ictKava.FullNodes[0].GetHostAddress(ctx, "9090/tcp")
	require.NoError(t, err, "failed to find grpc URL")
	evmUrl, err := ictKava.FullNodes[0].GetHostAddress(ctx, "8545/tcp")
	require.NoError(t, err, "failed to find evm URL")

	evmClient, err := ethclient.Dial(evmUrl)
	require.NoError(t, err, "failed to connect to evm")

	// create a funded evm accountto initialize the testutil.Chain
	deployerMnemonic, evmDeployer, err := newEvmAccount(evmClient, kavainterchain.KavaEvmTestChainId)
	require.NoError(t, err)

	deployerKavaAddr := util.EvmToSdkAddress(evmDeployer.Address())
	err = ictKava.SendFunds(ctx, kavaUser.KeyName(), ibc.WalletAmount{
		Address: deployerKavaAddr.String(),
		Denom:   "ukava",
		Amount:  math.NewInt(1e10),
	})
	require.NoError(t, err)

	// initialize testutil.Chain for account & tx management on both sdk & evm
	kava, err := testutil.NewChain(t, &runner.ChainDetails{
		RpcUrl:       rpcUrl,
		GrpcUrl:      grpcUrl,
		EvmRpcUrl:    evmUrl,
		ChainId:      kavainterchain.KavaTestChainId,
		StakingDenom: "ukava",
	}, deployerMnemonic)
	require.NoError(t, err)

	deployer := kava.GetAccount("whale")

	// deploy ERC20 contract
	usdtAddr, tx, usdt, err := erc20.DeployErc20(
		deployer.EvmAuth, kava.EvmClient,
		"Test Tether USD", "USDT", 6,
	)
	require.NoError(t, err)
	require.NotNil(t, usdtAddr)
	require.NotNil(t, usdt)

	_, err = util.WaitForEvmTxReceipt(kava.EvmClient, tx.Hash(), 10*time.Second)
	require.NoError(t, err)

	////////////////////////////////////////////
	// enable conversion from erc20 -> sdk.Coin
	// (assumes there are none pre-configured!)
	////////////////////////////////////////////
	// 1. Submit Proposal
	sdkDenom := "tether/usdt"
	rawCps, err := json.Marshal(evmutiltypes.NewConversionPairs(
		evmutiltypes.NewConversionPair(
			evmutiltypes.NewInternalEVMAddress(usdtAddr),
			sdkDenom,
		),
	))
	require.NoError(t, err)

	paramChange := paramsutils.ParamChangeProposalJSON{
		Title:       "Enable erc20 conversion to sdk.Coin",
		Description: ".",
		Changes: paramsutils.ParamChangesJSON{
			paramsutils.ParamChangeJSON{
				Subspace: "evmutil",
				Key:      "EnabledConversionPairs",
				Value:    rawCps,
			},
		},
		Deposit: "10000000ukava",
	}

	_, err = legacyParamChangeProposal(ictKava.FullNodes[0], ctx, kavaUser.KeyName(), &paramChange)
	require.NoError(t, err, "error submitting param change proposal tx")

	// TODO: query proposal id. assuming it is 1 here.
	propId := int64(1)

	// 2. Vote on Proposal
	err = ictKava.VoteOnProposalAllValidators(ctx, propId, cosmos.ProposalVoteYes)
	require.NoError(t, err, "failed to submit votes")

	height, _ = ictKava.Height(ctx)
	_, err = cosmos.PollForProposalStatus(ctx, ictKava, height, height+10, propId, gov1beta1.StatusPassed)
	require.NoError(t, err, "proposal status did not change to passed in expected number of blocks")

	// fund a user & mint them some usdt
	user := kava.NewFundedAccount("tether-user", sdk.NewCoins(sdk.NewCoin("ukava", math.NewInt(1e7))))
	erc20FundAmt := big.NewInt(100e6)
	tx, err = usdt.Mint(deployer.EvmAuth, user.EvmAddress, erc20FundAmt)
	require.NoError(t, err)

	_, err = util.WaitForEvmTxReceipt(kava.EvmClient, tx.Hash(), 10*time.Second)
	require.NoError(t, err)
	// verify they have erc20 balance!
	bal, err := usdt.BalanceOf(nil, user.EvmAddress)
	require.NoError(t, err)
	require.Equal(t, erc20FundAmt, bal)

	// convert the erc20 to sdk.Coin!
	amountToConvert := math.NewInt(50e6)
	msg := evmutiltypes.NewMsgConvertERC20ToCoin(
		evmutiltypes.NewInternalEVMAddress(user.EvmAddress),
		user.SdkAddress,
		evmutiltypes.NewInternalEVMAddress(usdtAddr),
		amountToConvert,
	)
	convertTx := util.KavaMsgRequest{
		Msgs:      []sdk.Msg{&msg},
		GasLimit:  4e5,
		FeeAmount: sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(400))),
		Data:      "converting sdk coin to erc20",
	}
	res := user.SignAndBroadcastKavaTx(convertTx)
	require.NoError(t, res.Err)

	// check balance!
	sdkBalance := kava.QuerySdkForBalances(user.SdkAddress)
	require.Equal(t, amountToConvert, sdkBalance.AmountOf(sdkDenom))
}

func newMnemonic() (string, error) {
	entropy, err := bip39.NewEntropy(128)
	if err != nil {
		return "", fmt.Errorf("failed to generate entropy for new mnemonic: %s", err)
	}
	return bip39.NewMnemonic(entropy)
}

func newEvmAccount(evmClient *ethclient.Client, chainId int64) (string, *util.EvmSigner, error) {
	mnemonic, err := newMnemonic()
	if err != nil {
		return mnemonic, nil, err
	}

	hdPath := hd.CreateHDPath(60, 0, 0)
	privKeyBytes, err := hd.Secp256k1.Derive()(mnemonic, "", hdPath.String())
	if err != nil {
		return mnemonic, nil, fmt.Errorf("failed to derive private key from mnemonic: %s", err)
	}
	privKey := &ethsecp256k1.PrivKey{Key: privKeyBytes}
	ecdsaPrivKey, err := crypto.HexToECDSA(hex.EncodeToString(privKey.Bytes()))
	if err != nil {
		return mnemonic, nil, err
	}

	signer, err := util.NewEvmSigner(evmClient, ecdsaPrivKey, big.NewInt(chainId))
	return mnemonic, signer, err
}

// copied from https://github.com/strangelove-ventures/interchaintest/blob/v7/chain/cosmos/chain_node.go#L1270
// but changed "submit-proposal" to "submit-legacy-proposal"
func legacyParamChangeProposal(tn *cosmos.ChainNode, ctx context.Context, keyName string, prop *paramsutils.ParamChangeProposalJSON) (string, error) {
	content, err := json.Marshal(prop)
	if err != nil {
		return "", err
	}

	hash := sha256.Sum256(content)
	proposalFilename := fmt.Sprintf("%x.json", hash)
	err = tn.WriteFile(ctx, content, proposalFilename)
	if err != nil {
		return "", fmt.Errorf("writing param change proposal: %w", err)
	}

	proposalPath := filepath.Join(tn.HomeDir(), proposalFilename)

	command := []string{
		"gov", "submit-legacy-proposal",
		"param-change",
		proposalPath,
	}

	return tn.ExecTx(ctx, keyName, command...)
}
