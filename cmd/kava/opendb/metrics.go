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

	// Detailed Cache
	BlockCacheIndexMiss        metrics.Gauge
	BlockCacheIndexHit         metrics.Gauge
	BlockCacheIndexBytesInsert metrics.Gauge

	BlockCacheFilterMiss        metrics.Gauge
	BlockCacheFilterHit         metrics.Gauge
	BlockCacheFilterBytesInsert metrics.Gauge

	BlockCacheDataMiss        metrics.Gauge
	BlockCacheDataHit         metrics.Gauge
	BlockCacheDataBytesInsert metrics.Gauge

	// Latency
	DBGetMicrosP50   metrics.Gauge
	DBGetMicrosP95   metrics.Gauge
	DBGetMicrosP99   metrics.Gauge
	DBGetMicrosP100  metrics.Gauge
	DBGetMicrosCount metrics.Gauge

	DBWriteMicrosP50   metrics.Gauge
	DBWriteMicrosP95   metrics.Gauge
	DBWriteMicrosP99   metrics.Gauge
	DBWriteMicrosP100  metrics.Gauge
	DBWriteMicrosCount metrics.Gauge

	// Write Stall
	StallMicros metrics.Gauge

	DBWriteStallP50   metrics.Gauge
	DBWriteStallP95   metrics.Gauge
	DBWriteStallP99   metrics.Gauge
	DBWriteStallP100  metrics.Gauge
	DBWriteStallCount metrics.Gauge
	DBWriteStallSum   metrics.Gauge

	// Bloom Filter
	BloomFilterUseful           metrics.Gauge
	BloomFilterFullPositive     metrics.Gauge
	BloomFilterFullTruePositive metrics.Gauge

	// LSM Tree Stats
	LastLevelReadBytes    metrics.Gauge
	LastLevelReadCount    metrics.Gauge
	NonLastLevelReadBytes metrics.Gauge
	NonLastLevelReadCount metrics.Gauge

	GetHitL0      metrics.Gauge
	GetHitL1      metrics.Gauge
	GetHitL2AndUp metrics.Gauge
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

		// Detailed Cache
		BlockCacheIndexMiss: prometheus.NewGaugeFrom(stdprometheus.GaugeOpts{
			Namespace: "rocksdb",
			Subsystem: "detailed_cache",
			Name:      "block_cache_index_miss",
			Help:      "",
		}, labels),
		BlockCacheIndexHit: prometheus.NewGaugeFrom(stdprometheus.GaugeOpts{
			Namespace: "rocksdb",
			Subsystem: "detailed_cache",
			Name:      "block_cache_index_hit",
			Help:      "",
		}, labels),
		BlockCacheIndexBytesInsert: prometheus.NewGaugeFrom(stdprometheus.GaugeOpts{
			Namespace: "rocksdb",
			Subsystem: "detailed_cache",
			Name:      "block_cache_index_bytes_insert",
			Help:      "",
		}, labels),

		BlockCacheFilterMiss: prometheus.NewGaugeFrom(stdprometheus.GaugeOpts{
			Namespace: "rocksdb",
			Subsystem: "detailed_cache",
			Name:      "block_cache_filter_miss",
			Help:      "",
		}, labels),
		BlockCacheFilterHit: prometheus.NewGaugeFrom(stdprometheus.GaugeOpts{
			Namespace: "rocksdb",
			Subsystem: "detailed_cache",
			Name:      "block_cache_filter_hit",
			Help:      "",
		}, labels),
		BlockCacheFilterBytesInsert: prometheus.NewGaugeFrom(stdprometheus.GaugeOpts{
			Namespace: "rocksdb",
			Subsystem: "detailed_cache",
			Name:      "block_cache_filter_bytes_insert",
			Help:      "",
		}, labels),

		BlockCacheDataMiss: prometheus.NewGaugeFrom(stdprometheus.GaugeOpts{
			Namespace: "rocksdb",
			Subsystem: "detailed_cache",
			Name:      "block_cache_data_miss",
			Help:      "",
		}, labels),
		BlockCacheDataHit: prometheus.NewGaugeFrom(stdprometheus.GaugeOpts{
			Namespace: "rocksdb",
			Subsystem: "detailed_cache",
			Name:      "block_cache_data_hit",
			Help:      "",
		}, labels),
		BlockCacheDataBytesInsert: prometheus.NewGaugeFrom(stdprometheus.GaugeOpts{
			Namespace: "rocksdb",
			Subsystem: "detailed_cache",
			Name:      "block_cache_data_bytes_insert",
			Help:      "",
		}, labels),

		// Latency
		DBGetMicrosP50: prometheus.NewGaugeFrom(stdprometheus.GaugeOpts{
			Namespace: "rocksdb",
			Subsystem: "latency",
			Name:      "db_get_micros_p50",
			Help:      "",
		}, labels),
		DBGetMicrosP95: prometheus.NewGaugeFrom(stdprometheus.GaugeOpts{
			Namespace: "rocksdb",
			Subsystem: "latency",
			Name:      "db_get_micros_p95",
			Help:      "",
		}, labels),
		DBGetMicrosP99: prometheus.NewGaugeFrom(stdprometheus.GaugeOpts{
			Namespace: "rocksdb",
			Subsystem: "latency",
			Name:      "db_get_micros_p99",
			Help:      "",
		}, labels),
		DBGetMicrosP100: prometheus.NewGaugeFrom(stdprometheus.GaugeOpts{
			Namespace: "rocksdb",
			Subsystem: "latency",
			Name:      "db_get_micros_p100",
			Help:      "",
		}, labels),
		DBGetMicrosCount: prometheus.NewGaugeFrom(stdprometheus.GaugeOpts{
			Namespace: "rocksdb",
			Subsystem: "latency",
			Name:      "db_get_micros_count",
			Help:      "",
		}, labels),

		DBWriteMicrosP50: prometheus.NewGaugeFrom(stdprometheus.GaugeOpts{
			Namespace: "rocksdb",
			Subsystem: "latency",
			Name:      "db_write_micros_p50",
			Help:      "",
		}, labels),
		DBWriteMicrosP95: prometheus.NewGaugeFrom(stdprometheus.GaugeOpts{
			Namespace: "rocksdb",
			Subsystem: "latency",
			Name:      "db_write_micros_p95",
			Help:      "",
		}, labels),
		DBWriteMicrosP99: prometheus.NewGaugeFrom(stdprometheus.GaugeOpts{
			Namespace: "rocksdb",
			Subsystem: "latency",
			Name:      "db_write_micros_p99",
			Help:      "",
		}, labels),
		DBWriteMicrosP100: prometheus.NewGaugeFrom(stdprometheus.GaugeOpts{
			Namespace: "rocksdb",
			Subsystem: "latency",
			Name:      "db_write_micros_p100",
			Help:      "",
		}, labels),
		DBWriteMicrosCount: prometheus.NewGaugeFrom(stdprometheus.GaugeOpts{
			Namespace: "rocksdb",
			Subsystem: "latency",
			Name:      "db_write_micros_count",
			Help:      "",
		}, labels),

		// Write Stall
		StallMicros: prometheus.NewGaugeFrom(stdprometheus.GaugeOpts{
			Namespace: "rocksdb",
			Subsystem: "stall",
			Name:      "stall_micros",
			Help:      "Writer has to wait for compaction or flush to finish.",
		}, labels),

		DBWriteStallP50: prometheus.NewGaugeFrom(stdprometheus.GaugeOpts{
			Namespace: "rocksdb",
			Subsystem: "stall",
			Name:      "db_write_stall_p50",
			Help:      "",
		}, labels),
		DBWriteStallP95: prometheus.NewGaugeFrom(stdprometheus.GaugeOpts{
			Namespace: "rocksdb",
			Subsystem: "stall",
			Name:      "db_write_stall_p95",
			Help:      "",
		}, labels),
		DBWriteStallP99: prometheus.NewGaugeFrom(stdprometheus.GaugeOpts{
			Namespace: "rocksdb",
			Subsystem: "stall",
			Name:      "db_write_stall_p99",
			Help:      "",
		}, labels),
		DBWriteStallP100: prometheus.NewGaugeFrom(stdprometheus.GaugeOpts{
			Namespace: "rocksdb",
			Subsystem: "stall",
			Name:      "db_write_stall_p100",
			Help:      "",
		}, labels),
		DBWriteStallCount: prometheus.NewGaugeFrom(stdprometheus.GaugeOpts{
			Namespace: "rocksdb",
			Subsystem: "stall",
			Name:      "db_write_stall_count",
			Help:      "",
		}, labels),
		DBWriteStallSum: prometheus.NewGaugeFrom(stdprometheus.GaugeOpts{
			Namespace: "rocksdb",
			Subsystem: "stall",
			Name:      "db_write_stall_sum",
			Help:      "",
		}, labels),

		// Bloom Filter
		BloomFilterUseful: prometheus.NewGaugeFrom(stdprometheus.GaugeOpts{
			Namespace: "rocksdb",
			Subsystem: "filter",
			Name:      "bloom_filter_useful",
			Help:      "number of times bloom filter has avoided file reads, i.e., negatives.",
		}, labels),
		BloomFilterFullPositive: prometheus.NewGaugeFrom(stdprometheus.GaugeOpts{
			Namespace: "rocksdb",
			Subsystem: "filter",
			Name:      "bloom_filter_full_positive",
			Help:      "number of times bloom FullFilter has not avoided the reads.",
		}, labels),
		BloomFilterFullTruePositive: prometheus.NewGaugeFrom(stdprometheus.GaugeOpts{
			Namespace: "rocksdb",
			Subsystem: "filter",
			Name:      "bloom_filter_full_true_positive",
			Help:      "number of times bloom FullFilter has not avoided the reads and data actually exist.",
		}, labels),

		// LSM Tree Stats
		LastLevelReadBytes: prometheus.NewGaugeFrom(stdprometheus.GaugeOpts{
			Namespace: "rocksdb",
			Subsystem: "lsm",
			Name:      "last_level_read_bytes",
			Help:      "",
		}, labels),
		LastLevelReadCount: prometheus.NewGaugeFrom(stdprometheus.GaugeOpts{
			Namespace: "rocksdb",
			Subsystem: "lsm",
			Name:      "last_level_read_count",
			Help:      "",
		}, labels),
		NonLastLevelReadBytes: prometheus.NewGaugeFrom(stdprometheus.GaugeOpts{
			Namespace: "rocksdb",
			Subsystem: "lsm",
			Name:      "non_last_level_read_bytes",
			Help:      "",
		}, labels),
		NonLastLevelReadCount: prometheus.NewGaugeFrom(stdprometheus.GaugeOpts{
			Namespace: "rocksdb",
			Subsystem: "lsm",
			Name:      "non_last_level_read_count",
			Help:      "",
		}, labels),

		GetHitL0: prometheus.NewGaugeFrom(stdprometheus.GaugeOpts{
			Namespace: "rocksdb",
			Subsystem: "lsm",
			Name:      "get_hit_l0",
			Help:      "number of Get() queries served by L0",
		}, labels),
		GetHitL1: prometheus.NewGaugeFrom(stdprometheus.GaugeOpts{
			Namespace: "rocksdb",
			Subsystem: "lsm",
			Name:      "get_hit_l1",
			Help:      "number of Get() queries served by L1",
		}, labels),
		GetHitL2AndUp: prometheus.NewGaugeFrom(stdprometheus.GaugeOpts{
			Namespace: "rocksdb",
			Subsystem: "lsm",
			Name:      "get_hit_l2_and_up",
			Help:      "number of Get() queries served by L2 and up",
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

	// Detailed Cache
	m.BlockCacheIndexMiss.Set(float64(stats.BlockCacheIndexMiss))
	m.BlockCacheIndexHit.Set(float64(stats.BlockCacheIndexHit))
	m.BlockCacheIndexBytesInsert.Set(float64(stats.BlockCacheIndexBytesInsert))

	m.BlockCacheFilterMiss.Set(float64(stats.BlockCacheFilterMiss))
	m.BlockCacheFilterHit.Set(float64(stats.BlockCacheFilterHit))
	m.BlockCacheFilterBytesInsert.Set(float64(stats.BlockCacheFilterBytesInsert))

	m.BlockCacheDataMiss.Set(float64(stats.BlockCacheDataMiss))
	m.BlockCacheDataHit.Set(float64(stats.BlockCacheDataHit))
	m.BlockCacheDataBytesInsert.Set(float64(stats.BlockCacheDataBytesInsert))

	// Latency
	m.DBGetMicrosP50.Set(stats.DBGetMicros.P50)
	m.DBGetMicrosP95.Set(stats.DBGetMicros.P95)
	m.DBGetMicrosP99.Set(stats.DBGetMicros.P99)
	m.DBGetMicrosP100.Set(stats.DBGetMicros.P100)
	m.DBGetMicrosCount.Set(stats.DBGetMicros.Count)

	m.DBWriteMicrosP50.Set(stats.DBWriteMicros.P50)
	m.DBWriteMicrosP95.Set(stats.DBWriteMicros.P95)
	m.DBWriteMicrosP99.Set(stats.DBWriteMicros.P99)
	m.DBWriteMicrosP100.Set(stats.DBWriteMicros.P100)
	m.DBWriteMicrosCount.Set(stats.DBWriteMicros.Count)

	// Write Stall
	m.StallMicros.Set(float64(stats.StallMicros))

	m.DBWriteStallP50.Set(stats.DBWriteStallHistogram.P50)
	m.DBWriteStallP95.Set(stats.DBWriteStallHistogram.P95)
	m.DBWriteStallP99.Set(stats.DBWriteStallHistogram.P99)
	m.DBWriteStallP100.Set(stats.DBWriteStallHistogram.P100)
	m.DBWriteStallCount.Set(stats.DBWriteStallHistogram.Count)
	m.DBWriteStallSum.Set(stats.DBWriteStallHistogram.Sum)

	// Bloom Filter
	m.BloomFilterUseful.Set(float64(stats.BloomFilterUseful))
	m.BloomFilterFullPositive.Set(float64(stats.BloomFilterFullPositive))
	m.BloomFilterFullTruePositive.Set(float64(stats.BloomFilterFullTruePositive))

	// LSM Tree Stats
	m.LastLevelReadBytes.Set(float64(stats.LastLevelReadBytes))
	m.LastLevelReadCount.Set(float64(stats.LastLevelReadCount))
	m.NonLastLevelReadBytes.Set(float64(stats.NonLastLevelReadBytes))
	m.NonLastLevelReadCount.Set(float64(stats.NonLastLevelReadCount))

	m.GetHitL0.Set(float64(stats.GetHitL0))
	m.GetHitL1.Set(float64(stats.GetHitL1))
	m.GetHitL2AndUp.Set(float64(stats.GetHitL2AndUp))
}
