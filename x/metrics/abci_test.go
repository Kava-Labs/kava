package metrics_test

import (
	"testing"

	kitmetrics "github.com/go-kit/kit/metrics"
	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/metrics"
	"github.com/kava-labs/kava/x/metrics/types"
)

type MockGauge struct {
	value float64
}

func (mg *MockGauge) With(labelValues ...string) kitmetrics.Gauge { return mg }
func (mg *MockGauge) Set(value float64)                           { mg.value = value }
func (*MockGauge) Add(_ float64)                                  {}

func ctxWithHeight(height int64) sdk.Context {
	tApp := app.NewTestApp()
	tApp.InitializeFromGenesisStates()
	return tApp.NewContext(false, tmproto.Header{Height: height})
}

func TestBeginBlockEmitsLatestHeight(t *testing.T) {
	gauge := MockGauge{}
	myMetrics := &types.Metrics{
		LatestBlockHeight: &gauge,
	}

	metrics.BeginBlocker(ctxWithHeight(1), myMetrics)
	require.EqualValues(t, 1, gauge.value)

	metrics.BeginBlocker(ctxWithHeight(100), myMetrics)
	require.EqualValues(t, 100, gauge.value)

	metrics.BeginBlocker(ctxWithHeight(17e6), myMetrics)
	require.EqualValues(t, 17e6, gauge.value)
}
