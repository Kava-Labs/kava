package types

import (
	"github.com/go-kit/kit/metrics"
	"github.com/go-kit/kit/metrics/discard"
	prometheus "github.com/go-kit/kit/metrics/prometheus"
	stdprometheus "github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/cast"

	servertypes "github.com/cosmos/cosmos-sdk/server/types"
)

// TelemetryOptions defines the app configurations for the x/metrics module
type TelemetryOptions struct {
	// CometBFT config value for instrumentation.prometheus (config.toml)
	PrometheusEnabled bool
	// Namespace for CometBFT metrics. Used to emulate CometBFT metrics.
	CometBFTMetricsNamespace string
	// A list of keys and values used as labels on all metrics
	GlobalLabelsAndValues []string
}

// TelemetryOptionsFromAppOpts creates the TelemetryOptions from server AppOptions
func TelemetryOptionsFromAppOpts(appOpts servertypes.AppOptions) TelemetryOptions {
	prometheusEnabled := cast.ToBool(appOpts.Get("instrumentation.prometheus"))
	if !prometheusEnabled {
		return TelemetryOptions{
			GlobalLabelsAndValues: []string{},
		}
	}

	// parse the app.toml global-labels into a slice of alternating label & value strings
	// the value is expected to be a list of [label, value] tuples.
	rawLabels := cast.ToSlice(appOpts.Get("telemetry.global-labels"))
	globalLabelsAndValues := make([]string, 0, len(rawLabels)*2)
	for _, gl := range rawLabels {
		l := cast.ToStringSlice(gl)
		globalLabelsAndValues = append(globalLabelsAndValues, l[0], l[1])
	}

	return TelemetryOptions{
		PrometheusEnabled:        true,
		CometBFTMetricsNamespace: cast.ToString(appOpts.Get("instrumentation.namespace")),
		GlobalLabelsAndValues:    globalLabelsAndValues,
	}
}

// Metrics contains metrics exposed by this module.
// They use go-kit metrics like CometBFT as opposed to using cosmos-sdk telemetry
// because the sdk's telemetry only supports float32s, whereas go-kit prometheus
// metrics correctly handle float64s (and thus a larger number of int64s)
type Metrics struct {
	// The height of the latest block.
	// This gauges exactly emulates the default blocksync metric in CometBFT v0.38+
	// It should be removed when kava has been updated to CometBFT v0.38+.
	// see https://github.com/cometbft/cometbft/blob/v0.38.0-rc3/blocksync/metrics.gen.go
	LatestBlockHeight metrics.Gauge
}

// NewMetrics creates a new Metrics object based on whether or not prometheus instrumentation is enabled.
func NewMetrics(opts TelemetryOptions) *Metrics {
	if opts.PrometheusEnabled {
		return PrometheusMetrics(opts)
	}
	return NoopMetrics()
}

// PrometheusMetrics returns the gauges for when prometheus instrumentation is enabled.
func PrometheusMetrics(opts TelemetryOptions) *Metrics {
	labels := []string{}
	for i := 0; i < len(opts.GlobalLabelsAndValues); i += 2 {
		labels = append(labels, opts.GlobalLabelsAndValues[i])
	}
	return &Metrics{
		LatestBlockHeight: prometheus.NewGaugeFrom(stdprometheus.GaugeOpts{
			Namespace: opts.CometBFTMetricsNamespace,
			Subsystem: "blocksync",
			Name:      "latest_block_height",
			Help:      "The height of the latest block.",
		}, labels).With(opts.GlobalLabelsAndValues...),
	}
}

// NoopMetrics are a do-nothing placeholder used when prometheus instrumentation is not enabled.
func NoopMetrics() *Metrics {
	return &Metrics{
		LatestBlockHeight: discard.NewGauge(),
	}
}
