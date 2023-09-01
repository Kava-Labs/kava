//go:build rocksdb
// +build rocksdb

package opendb

import (
	"github.com/go-kit/kit/metrics"
	"github.com/go-kit/kit/metrics/prometheus"
	stdprometheus "github.com/prometheus/client_golang/prometheus"
)

// rocksdbMetrics will be initialized in registerMetrics() if enableRocksdbMetrics flag set to true
var rocksdbMetrics *Metrics

// Metrics contains all rocksdb metrics which will be reported to prometheus
type Metrics struct {
	// Keys
	NumberKeysWritten metrics.Gauge
	NumberKeysRead    metrics.Gauge
	NumberKeysUpdated metrics.Gauge
	EstimateNumKeys   metrics.Gauge

	// Files
	NumberFileOpens  metrics.Gauge
	NumberFileErrors metrics.Gauge

	// Memory
	BlockCacheUsage         metrics.Gauge
	EstimateTableReadersMem metrics.Gauge
	CurSizeAllMemTables     metrics.Gauge
	BlockCachePinnedUsage   metrics.Gauge

	// Cache
	BlockCacheMiss        metrics.Gauge
	BlockCacheHit         metrics.Gauge
	BlockCacheAdd         metrics.Gauge
	BlockCacheAddFailures metrics.Gauge
}

// registerMetrics registers metrics in prometheus and initializes rocksdbMetrics variable
func registerMetrics() {
	if rocksdbMetrics != nil {
		// metrics already registered
		return
	}

	labels := make([]string, 0)
	rocksdbMetrics = &Metrics{
		// Keys
		NumberKeysWritten: prometheus.NewGaugeFrom(stdprometheus.GaugeOpts{
			Namespace: "rocksdb",
			Subsystem: "key",
			Name:      "number_keys_written",
			Help:      "",
		}, labels),
		NumberKeysRead: prometheus.NewGaugeFrom(stdprometheus.GaugeOpts{
			Namespace: "rocksdb",
			Subsystem: "key",
			Name:      "number_keys_read",
			Help:      "",
		}, labels),
		NumberKeysUpdated: prometheus.NewGaugeFrom(stdprometheus.GaugeOpts{
			Namespace: "rocksdb",
			Subsystem: "key",
			Name:      "number_keys_updated",
			Help:      "",
		}, labels),
		EstimateNumKeys: prometheus.NewGaugeFrom(stdprometheus.GaugeOpts{
			Namespace: "rocksdb",
			Subsystem: "key",
			Name:      "estimate_num_keys",
			Help:      "estimated number of total keys in the active and unflushed immutable memtables and storage",
		}, labels),

		// Files
		NumberFileOpens: prometheus.NewGaugeFrom(stdprometheus.GaugeOpts{
			Namespace: "rocksdb",
			Subsystem: "file",
			Name:      "number_file_opens",
			Help:      "",
		}, labels),
		NumberFileErrors: prometheus.NewGaugeFrom(stdprometheus.GaugeOpts{
			Namespace: "rocksdb",
			Subsystem: "file",
			Name:      "number_file_errors",
			Help:      "",
		}, labels),

		// Memory
		BlockCacheUsage: prometheus.NewGaugeFrom(stdprometheus.GaugeOpts{
			Namespace: "rocksdb",
			Subsystem: "memory",
			Name:      "block_cache_usage",
			Help:      "memory size for the entries residing in block cache",
		}, labels),
		EstimateTableReadersMem: prometheus.NewGaugeFrom(stdprometheus.GaugeOpts{
			Namespace: "rocksdb",
			Subsystem: "memory",
			Name:      "estimate_table_readers_mem",
			Help:      "estimated memory used for reading SST tables, excluding memory used in block cache (e.g., filter and index blocks)",
		}, labels),
		CurSizeAllMemTables: prometheus.NewGaugeFrom(stdprometheus.GaugeOpts{
			Namespace: "rocksdb",
			Subsystem: "memory",
			Name:      "cur_size_all_mem_tables",
			Help:      "approximate size of active and unflushed immutable memtables (bytes)",
		}, labels),
		BlockCachePinnedUsage: prometheus.NewGaugeFrom(stdprometheus.GaugeOpts{
			Namespace: "rocksdb",
			Subsystem: "memory",
			Name:      "block_cache_pinned_usage",
			Help:      "returns the memory size for the entries being pinned",
		}, labels),

		// Cache
		BlockCacheMiss: prometheus.NewGaugeFrom(stdprometheus.GaugeOpts{
			Namespace: "rocksdb",
			Subsystem: "cache",
			Name:      "block_cache_miss",
			Help:      "block_cache_miss == block_cache_index_miss + block_cache_filter_miss + block_cache_data_miss",
		}, labels),
		BlockCacheHit: prometheus.NewGaugeFrom(stdprometheus.GaugeOpts{
			Namespace: "rocksdb",
			Subsystem: "cache",
			Name:      "block_cache_hit",
			Help:      "block_cache_hit == block_cache_index_hit + block_cache_filter_hit + block_cache_data_hit",
		}, labels),
		BlockCacheAdd: prometheus.NewGaugeFrom(stdprometheus.GaugeOpts{
			Namespace: "rocksdb",
			Subsystem: "cache",
			Name:      "block_cache_add",
			Help:      "number of blocks added to block cache",
		}, labels),
		BlockCacheAddFailures: prometheus.NewGaugeFrom(stdprometheus.GaugeOpts{
			Namespace: "rocksdb",
			Subsystem: "cache",
			Name:      "block_cache_add_failures",
			Help:      "number of failures when adding blocks to block cache",
		}, labels),
	}
}

// report reports metrics to prometheus based on rocksdb props and stats
func (m *Metrics) report(props *properties, stats *stats) {
	// Keys
	m.NumberKeysWritten.Set(float64(stats.NumberKeysWritten))
	m.NumberKeysRead.Set(float64(stats.NumberKeysRead))
	m.NumberKeysUpdated.Set(float64(stats.NumberKeysUpdated))
	m.EstimateNumKeys.Set(float64(props.EstimateNumKeys))

	// Files
	m.NumberFileOpens.Set(float64(stats.NumberFileOpens))
	m.NumberFileErrors.Set(float64(stats.NumberFileErrors))

	// Memory
	m.BlockCacheUsage.Set(float64(props.BlockCacheUsage))
	m.EstimateTableReadersMem.Set(float64(props.EstimateTableReadersMem))
	m.CurSizeAllMemTables.Set(float64(props.CurSizeAllMemTables))
	m.BlockCachePinnedUsage.Set(float64(props.BlockCachePinnedUsage))

	// Cache
	m.BlockCacheMiss.Set(float64(stats.BlockCacheMiss))
	m.BlockCacheHit.Set(float64(stats.BlockCacheHit))
	m.BlockCacheAdd.Set(float64(stats.BlockCacheAdd))
	m.BlockCacheAddFailures.Set(float64(stats.BlockCacheAddFailures))
}
