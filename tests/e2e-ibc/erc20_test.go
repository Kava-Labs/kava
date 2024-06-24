package main_test

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"math/big"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	gov1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	paramsutils "github.com/cosmos/cosmos-sdk/x/params/client/utils"
	transfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"

	"github.com/ethereum/go-ethereum/ethclient"

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

// This test does the following:
// - set up a network with kava & cosmoshub w/ IBC channels between them
// - deploy an ERC20 to Kava EVM
// - configure Kava to support conversion of ERC20 to Cosmos Coin
// - IBC the sdk.Coin representation of the ERC20 to cosmoshub
func TestInterchainErc20(t *testing.T) {
	app.SetSDKConfig()
	ctx := context.Background()

	// Configure & run chains with interchaintest
	cf := interchaintest.NewBuiltinChainFactory(zaptest.NewLogger(t), []*interchaintest.ChainSpec{
		{Name: "kava", ChainConfig: kavainterchain.DefaultKavaChainConfig(kavainterchain.KavaTestChainId)},
		{Name: "gaia", Version: "v15.2.0", ChainConfig: ibc.ChainConfig{GasPrices: "0.0uatom"}},
	})

	chains, err := cf.Chains(t.Name())
	require.NoError(t, err)

	ictKava := chains[0].(*cosmos.CosmosChain)
	gaia := chains[1].(*cosmos.CosmosChain)

	client, network := interchaintest.DockerSetup(t)

	r := interchaintest.NewBuiltinRelayerFactory(ibc.CosmosRly, zaptest.NewLogger(t)).
		Build(t, client, network)

	// configure interchain
	const kavaGaiaIbcPath = "kava-gaia-ibc"
	ic := interchaintest.NewInterchain().AddChain(ictKava).
		AddChain(gaia).
		AddRelayer(r, "relayer").
		AddLink(interchaintest.InterchainLink{
			Chain1:  ictKava,
			Chain2:  gaia,
			Relayer: r,
			Path:    kavaGaiaIbcPath,
		})

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

	users := interchaintest.GetAndFundTestUsers(t, ctx, t.Name(), fundAmount, ictKava, gaia)
	kavaUser := users[0]
	gaiaUser := users[1]

	// wait for new block to ensure initial funding complete
	height, err := ictKava.Height(ctx)
	require.NoError(t, err)
	h := height
	for h <= height {
		h, err = ictKava.Height(ctx)
		require.NoError(t, err)
	}

	gaiaChannelInfo, err := r.GetChannels(ctx, eRep, gaia.Config().ChainID)
	require.NoError(t, err)
	gaiaToKavaChannelID := gaiaChannelInfo[0].ChannelID
	kavaToGaiaChannelID := gaiaChannelInfo[0].Counterparty.ChannelID

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

	// create a funded evm account to initialize the testutil.Chain
	// use testutil.Chain for account management because the accounts play nicely with EVM & SDK sides
	deployerMnemonic, err := util.RandomMnemonic()
	require.NoError(t, err)
	evmDeployer, err := util.NewEvmSignerFromMnemonic(evmClient, big.NewInt(kavainterchain.KavaEvmTestChainId), deployerMnemonic)
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
	usdtAddr, deployTx, usdt, err := erc20.DeployErc20(
		deployer.EvmAuth, kava.EvmClient,
		"Test Tether USD", "USDT", 6,
	)
	require.NoError(t, err)
	require.NotNil(t, usdtAddr)
	require.NotNil(t, usdt)

	_, err = util.WaitForEvmTxReceipt(kava.EvmClient, deployTx.Hash(), 10*time.Second)
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
	mintTx, err := usdt.Mint(deployer.EvmAuth, user.EvmAddress, erc20FundAmt)
	require.NoError(t, err)

	_, err = util.WaitForEvmTxReceipt(kava.EvmClient, mintTx.Hash(), 10*time.Second)
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

	// check balance of cosmos coin!
	sdkBalance := kava.QuerySdkForBalances(user.SdkAddress)
	require.Equal(t, amountToConvert, sdkBalance.AmountOf(sdkDenom))

	// IBC the newly minted sdk.Coin to gaia
	dstAddress := gaiaUser.FormattedAddress()
	transfer := ibc.WalletAmount{
		Address: dstAddress,
		Denom:   ictKava.Config().Denom,
		Amount:  amountToConvert,
	}
	_, err = ictKava.SendIBCTransfer(ctx, kavaToGaiaChannelID, kavaUser.KeyName(), transfer, ibc.TransferOptions{})
	require.NoError(t, err)

	// manually flush packets so we don't need to wait for the relayer
	require.NoError(t, r.Flush(ctx, eRep, kavaGaiaIbcPath, kavaToGaiaChannelID))

	// determine IBC denom & check gaia balance
	srcDenomTrace := transfertypes.ParseDenomTrace(transfertypes.GetPrefixedDenom("transfer", gaiaToKavaChannelID, ictKava.Config().Denom))
	erc20OnGaiaDenom := srcDenomTrace.IBCDenom()

	gaiaBal, err := gaia.GetBalance(ctx, dstAddress, erc20OnGaiaDenom)
	require.NoError(t, err)

	require.Equal(t, amountToConvert, gaiaBal)
}

// copied from https://github.com/strangelove-ventures/interchaintest/blob/7272afc780da6e2c99b2d2b3d084d5a3b1c6895f/chain/cosmos/chain_node.go#L1292
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
