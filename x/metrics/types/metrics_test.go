package types_test

import (
	"testing"

	"github.com/go-kit/kit/metrics"
	"github.com/go-kit/kit/metrics/prometheus"
	"github.com/kava-labs/kava/x/metrics/types"
	"github.com/stretchr/testify/require"
)

func isPrometheusGauge(g metrics.Gauge) bool {
	_, ok := g.(*prometheus.Gauge)
	return ok
}

var (
	disabledOpts = types.TelemetryOptions{
		PrometheusEnabled: false,
	}
	enabledOpts = types.TelemetryOptions{
		PrometheusEnabled:        true,
		CometBFTMetricsNamespace: "cometbft",
		GlobalLabelsAndValues:    []string{"label1", "value1", "label2", "value2"},
	}
)

func TestNewMetrics_DisabledVsEnabled(t *testing.T) {
	myMetrics := types.NewMetrics(disabledOpts)
	require.False(t, isPrometheusGauge(myMetrics.LatestBlockHeight))

	myMetrics = types.NewMetrics(enabledOpts)
	require.True(t, isPrometheusGauge(myMetrics.LatestBlockHeight))
}

type MockAppOpts struct {
	store map[string]interface{}
}

func (mao *MockAppOpts) Get(key string) interface{} {
	return mao.store[key]
}

func TestTelemetryOptionsFromAppOpts(t *testing.T) {
	appOpts := MockAppOpts{store: make(map[string]interface{})}

	// test disabled functionality
	appOpts.store["instrumentation.prometheus"] = false

	opts := types.TelemetryOptionsFromAppOpts(&appOpts)
	require.False(t, opts.PrometheusEnabled)

	// test enabled functionality
	appOpts.store["instrumentation.prometheus"] = true
	appOpts.store["instrumentation.namespace"] = "magic"
	appOpts.store["telemetry.global-labels"] = []interface{}{}

	opts = types.TelemetryOptionsFromAppOpts(&appOpts)
	require.True(t, opts.PrometheusEnabled)
	require.Equal(t, "magic", opts.CometBFTMetricsNamespace)
	require.Len(t, opts.GlobalLabelsAndValues, 0)

	appOpts.store["telemetry.global-labels"] = []interface{}{
		[]interface{}{"label1", "value1"},
		[]interface{}{"label2", "value2"},
	}
	opts = types.TelemetryOptionsFromAppOpts(&appOpts)
	require.True(t, opts.PrometheusEnabled)
	require.Equal(t, "magic", opts.CometBFTMetricsNamespace)
	require.Len(t, opts.GlobalLabelsAndValues, 4)
	require.Equal(t, enabledOpts.GlobalLabelsAndValues, opts.GlobalLabelsAndValues)
}
