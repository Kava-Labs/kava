package main_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"go.uber.org/zap/zaptest"

	"github.com/strangelove-ventures/interchaintest/v7"
	"github.com/strangelove-ventures/interchaintest/v7/conformance"
	"github.com/strangelove-ventures/interchaintest/v7/ibc"
	"github.com/strangelove-ventures/interchaintest/v7/testreporter"
	"github.com/stretchr/testify/require"

	kavainterchain "github.com/kava-labs/kava/tests/interchain"
)

func TestIbcConformance(t *testing.T) {
	t.Skip("skipping conformance test")

	ctx := context.Background()

	// setup chain factories: must be exactly two chains to run conformance between
	cfs := make([]interchaintest.ChainFactory, 0)
	cfs = append(cfs,
		interchaintest.NewBuiltinChainFactory(zaptest.NewLogger(t), []*interchaintest.ChainSpec{
			{Name: "kava", ChainConfig: kavainterchain.DefaultKavaChainConfig(kavainterchain.KavaTestChainId)},
			{Name: "osmosis", Version: "v24.0.1"},
		}),
	)

	// setup relayer factory
	rfs := make([]interchaintest.RelayerFactory, 0)
	rfs = append(rfs,
		interchaintest.NewBuiltinRelayerFactory(ibc.CosmosRly, zaptest.NewLogger(t)),
	)

	// test reporter
	f, err := interchaintest.CreateLogFile(fmt.Sprintf("%d.json", time.Now().Unix()))
	require.NoError(t, err)
	rep := testreporter.NewReporter(f)

	conformance.Test(t, ctx, cfs, rfs, rep)
}
