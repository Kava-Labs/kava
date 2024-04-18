package main_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"cosmossdk.io/math"
	transfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"

	"github.com/strangelove-ventures/interchaintest/v8"
	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v8/ibc"
	"github.com/strangelove-ventures/interchaintest/v8/testreporter"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

func TestInterchainIBC(t *testing.T) {
	ctx := context.Background()

	// setup chains
	cf := interchaintest.NewBuiltinChainFactory(zaptest.NewLogger(t), []*interchaintest.ChainSpec{
		{
			Name: "kava",
			ChainConfig: ibc.ChainConfig{
				Type:           "cosmos",
				ChainID:        "kava_8888-1",
				Images:         []ibc.DockerImage{{Repository: "ghcr.io/strangelove-ventures/heighliner/kava", Version: "v0.26.0", UidGid: "1025:1025"}},
				Bin:            "kava",
				Bech32Prefix:   "kava",
				Denom:          "ukava",
				GasPrices:      "0ukava", // 0 gas price makes calculating expected balances simpler
				GasAdjustment:  1.5,
				TrustingPeriod: "168h0m0s",
				// ModifyGenesis:  cosmos.ModifyGenesis(genesis),
				// CoinType: "60", // might need this to sign evm txs. will need to override decimals to be 6 again.
			},
		},
		{Name: "gaia", Version: "v7.0.0", ChainConfig: ibc.ChainConfig{
			GasPrices: "0.0uatom",
		}},
		// {Name: "osmosis", Version: "v11.0.0"},
	})

	chains, err := cf.Chains(t.Name())
	require.NoError(t, err)

	kava, gaia := chains[0].(*cosmos.CosmosChain), chains[1].(*cosmos.CosmosChain)

	// setup relayer
	client, network := interchaintest.DockerSetup(t)
	r := interchaintest.NewBuiltinRelayerFactory(ibc.CosmosRly, zaptest.NewLogger(t)).
		Build(t, client, network)

	// configure interchain
	const kavaGaiaIbcPath = "kava-gaia-demo"
	// const kavaOsmosisIbcPath = "kava-osmo-demo"
	ic := interchaintest.NewInterchain().
		AddChain(kava).
		AddChain(gaia).
		// AddChain(osmosis).
		AddRelayer(r, "relayer").
		AddLink(interchaintest.InterchainLink{
			Chain1:  kava,
			Chain2:  gaia,
			Relayer: r,
			Path:    kavaGaiaIbcPath,
		})
		// AddLink(interchaintest.InterchainLink{
		// 	Chain1:  kava,
		// 	Chain2:  osmosis,
		// 	Relayer: r,
		// 	Path:    kavaOsmosisIbcPath,
		// })

	// Log location
	f, err := interchaintest.CreateLogFile(fmt.Sprintf("%d.json", time.Now().Unix()))
	require.NoError(t, err)
	// Reporter/logs
	rep := testreporter.NewReporter(f)
	eRep := rep.RelayerExecReporter(t)

	// Build interchain
	require.NoError(t, ic.Build(ctx, eRep, interchaintest.InterchainBuildOptions{
		TestName:         t.Name(),
		Client:           client,
		NetworkID:        network,
		SkipPathCreation: false},
	),
	)

	// Create and Fund User Wallets
	fundAmount := math.NewInt(10_000_000)

	users := interchaintest.GetAndFundTestUsers(t, ctx, "default", fundAmount, kava, gaia)
	kavaUser := users[0]
	gaiaUser := users[1]
	// osmosisUser := users[2]

	// wait for new block to ensure initial funding complete
	height, err := kava.Height(ctx)
	require.NoError(t, err)
	h := height
	for h <= height {
		h, err = kava.Height(ctx)
		require.NoError(t, err)
	}

	// check initial balance
	kavaUserBalInitial, err := kava.GetBalance(ctx, kavaUser.FormattedAddress(), kava.Config().Denom)
	require.NoError(t, err)
	require.True(t, kavaUserBalInitial.Equal(fundAmount))

	// get ibc channel ids
	kavaChannelInfo, err := r.GetChannels(ctx, eRep, kava.Config().ChainID)
	require.NoError(t, err)
	kavaChannelID := kavaChannelInfo[0].ChannelID

	gaiaChannelInfo, err := r.GetChannels(ctx, eRep, gaia.Config().ChainID)
	require.NoError(t, err)
	gaiaChannelID := gaiaChannelInfo[0].ChannelID

	// IBC transfer kava to cosmoshub
	// inspired by https://github.com/strangelove-ventures/interchaintest/blob/main/examples/ibc/learn_ibc_test.go
	t.Run("ibc transfer kava -> cosmoshub", func(t *testing.T) {
		amountToSend := math.NewInt(1_000_000)
		dstAddress := gaiaUser.FormattedAddress()
		transfer := ibc.WalletAmount{
			Address: dstAddress,
			Denom:   kava.Config().Denom,
			Amount:  amountToSend,
		}

		tx, err := kava.SendIBCTransfer(ctx, kavaChannelID, kavaUser.KeyName(), transfer, ibc.TransferOptions{})
		require.NoError(t, err)
		require.NoError(t, tx.Validate())

		// relay MsgRecvPacket to gaia, then MsgAcknowledgement back to kava
		require.NoError(t, r.Flush(ctx, eRep, kavaGaiaIbcPath, kavaChannelID))

		// verify balance deducted from kava account
		expectedBal := kavaUserBalInitial.Sub(amountToSend)
		kavaUserBalNew, err := kava.GetBalance(ctx, kavaUser.FormattedAddress(), kava.Config().Denom)
		require.NoError(t, err)
		require.True(t, kavaUserBalNew.Equal(expectedBal))

		// determine ibc denom on cosmoshub
		srcDenomTrace := transfertypes.ParseDenomTrace(transfertypes.GetPrefixedDenom("transfer", gaiaChannelID, kava.Config().Denom))
		dstIbcDenom := srcDenomTrace.IBCDenom()

		// verify cosmoshub account received funds
		gaiaUserBalNew, err := gaia.GetBalance(ctx, gaiaUser.FormattedAddress(), dstIbcDenom)
		require.NoError(t, err)
		require.True(t, gaiaUserBalNew.Equal(amountToSend))
	})
}
