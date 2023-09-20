//go:build rocksdb
// +build rocksdb

package opendb

import (
	"fmt"
	"strconv"
)

const (
	sum   = "SUM"
	count = "COUNT"
	p50   = "P50"
	p95   = "P95"
	p99   = "P99"
	p100  = "P100"
)

type statLoader struct {
	// statMap contains map of stat objects returned by parseSerializedStats function
	// example of stats:
	// #1: rocksdb.block.cache.miss COUNT : 5
	// #2: rocksdb.compaction.times.micros P50 : 21112 P95 : 21112 P99 : 21112 P100 : 21112 COUNT : 1 SUM : 21112
	// #1 case will be cast into int64
	// #2 case will be cast into float64Histogram
	statMap map[string]*stat

	// NOTE: some methods accumulate errors instead of returning them, these methods are private and not intended to use outside
	errors []error
}

func newStatLoader(statMap map[string]*stat) *statLoader {
	return &statLoader{
		statMap: statMap,
		errors:  make([]error, 0),
	}
}

type stats struct {
	NumberKeysWritten int64
	NumberKeysRead    int64
	NumberKeysUpdated int64

	// total block cache misses
	// BLOCK_CACHE_MISS == BLOCK_CACHE_INDEX_MISS +
	//                     BLOCK_CACHE_FILTER_MISS +
	//                     BLOCK_CACHE_DATA_MISS;
	// BLOCK_CACHE_INDEX_MISS: # of times cache miss when accessing index block from block cache.
	// BLOCK_CACHE_FILTER_MISS: # of times cache miss when accessing filter block from block cache.
	// BLOCK_CACHE_DATA_MISS: # of times cache miss when accessing data block from block cache.
	BlockCacheMiss int64

	// total block cache hit
	// BLOCK_CACHE_HIT == BLOCK_CACHE_INDEX_HIT +
	//                    BLOCK_CACHE_FILTER_HIT +
	//                    BLOCK_CACHE_DATA_HIT;
	// BLOCK_CACHE_INDEX_HIT: # of times cache hit when accessing index block from block cache.
	// BLOCK_CACHE_FILTER_HIT: # of times cache hit when accessing filter block from block cache.
	// BLOCK_CACHE_DATA_HIT: # of times cache hit when accessing data block from block cache.
	BlockCacheHit int64

	// # of blocks added to block cache.
	BlockCacheAdd int64
	// # of failures when adding blocks to block cache.
	BlockCacheAddFailures int64

	BlockCacheIndexMiss         int64
	BlockCacheIndexHit          int64
	BlockCacheIndexBytesInsert  int64
	BlockCacheFilterMiss        int64
	BlockCacheFilterHit         int64
	BlockCacheFilterBytesInsert int64
	BlockCacheDataMiss          int64
	BlockCacheDataHit           int64
	BlockCacheDataBytesInsert   int64

	CompactReadBytes  int64 // Bytes read during compaction
	CompactWriteBytes int64 // Bytes written during compaction

	CompactionTimesMicros      *float64Histogram
	CompactionTimesCPUMicros   *float64Histogram
	NumFilesInSingleCompaction *float64Histogram

	// Read amplification statistics.
	// Read amplification can be calculated using this formula
	// (READ_AMP_TOTAL_READ_BYTES / READ_AMP_ESTIMATE_USEFUL_BYTES)
	//
	// REQUIRES: ReadOptions::read_amp_bytes_per_bit to be enabled
	// TODO(yevhenii): seems not working?
	ReadAmpEstimateUsefulBytes int64 // Estimate of total bytes actually used.
	ReadAmpTotalReadBytes      int64 // Total size of loaded data blocks.

	NumberFileOpens  int64
	NumberFileErrors int64

	// # of times bloom filter has avoided file reads, i.e., negatives.
	BloomFilterUseful int64
	// # of times bloom FullFilter has not avoided the reads.
	BloomFilterFullPositive int64
	// # of times bloom FullFilter has not avoided the reads and data actually
	// exist.
	BloomFilterFullTruePositive int64

	// # of memtable hits.
	MemtableHit int64
	// # of memtable misses.
	MemtableMiss int64

	// # of Get() queries served by L0
	GetHitL0 int64
	// # of Get() queries served by L1
	GetHitL1 int64
	// # of Get() queries served by L2 and up
	GetHitL2AndUp int64

	// The number of uncompressed bytes issued by DB::Put(), DB::Delete(),
	// DB::Merge(), and DB::Write().
	BytesWritten int64
	// The number of uncompressed bytes read from DB::Get().  It could be
	// either from memtables, cache, or table files.
	// For the number of logical bytes read from DB::MultiGet(),
	// please use NUMBER_MULTIGET_BYTES_READ.
	BytesRead int64

	// Writer has to wait for compaction or flush to finish.
	StallMicros           int64
	DBWriteStallHistogram *float64Histogram

	// Last level and non-last level read statistics
	LastLevelReadBytes    int64
	LastLevelReadCount    int64
	NonLastLevelReadBytes int64
	NonLastLevelReadCount int64

	DBGetMicros   *float64Histogram
	DBWriteMicros *float64Histogram

	// Value size distribution in each operation
	BytesPerRead     *float64Histogram
	BytesPerWrite    *float64Histogram
	BytesPerMultiget *float64Histogram

	// Time spent flushing memtable to disk
	FlushMicros *float64Histogram
}

type float64Histogram struct {
	Sum   float64
	Count float64
	P50   float64
	P95   float64
	P99   float64
	P100  float64
}

func (l *statLoader) error() error {
	if len(l.errors) != 0 {
		return fmt.Errorf("%v", l.errors)
	}

	return nil
}

func (l *statLoader) load() (*stats, error) {
	stats := &stats{
		NumberKeysWritten:           l.getInt64StatValue("rocksdb.number.keys.written", count),
		NumberKeysRead:              l.getInt64StatValue("rocksdb.number.keys.read", count),
		NumberKeysUpdated:           l.getInt64StatValue("rocksdb.number.keys.updated", count),
		BlockCacheMiss:              l.getInt64StatValue("rocksdb.block.cache.miss", count),
		BlockCacheHit:               l.getInt64StatValue("rocksdb.block.cache.hit", count),
		BlockCacheAdd:               l.getInt64StatValue("rocksdb.block.cache.add", count),
		BlockCacheAddFailures:       l.getInt64StatValue("rocksdb.block.cache.add.failures", count),
		BlockCacheIndexMiss:         l.getInt64StatValue("rocksdb.block.cache.index.miss", count),
		BlockCacheIndexHit:          l.getInt64StatValue("rocksdb.block.cache.index.hit", count),
		BlockCacheIndexBytesInsert:  l.getInt64StatValue("rocksdb.block.cache.index.bytes.insert", count),
		BlockCacheFilterMiss:        l.getInt64StatValue("rocksdb.block.cache.filter.miss", count),
		BlockCacheFilterHit:         l.getInt64StatValue("rocksdb.block.cache.filter.hit", count),
		BlockCacheFilterBytesInsert: l.getInt64StatValue("rocksdb.block.cache.filter.bytes.insert", count),
		BlockCacheDataMiss:          l.getInt64StatValue("rocksdb.block.cache.data.miss", count),
		BlockCacheDataHit:           l.getInt64StatValue("rocksdb.block.cache.data.hit", count),
		BlockCacheDataBytesInsert:   l.getInt64StatValue("rocksdb.block.cache.data.bytes.insert", count),
		CompactReadBytes:            l.getInt64StatValue("rocksdb.compact.read.bytes", count),
		CompactWriteBytes:           l.getInt64StatValue("rocksdb.compact.write.bytes", count),
		CompactionTimesMicros:       l.getFloat64HistogramStatValue("rocksdb.compaction.times.micros"),
		CompactionTimesCPUMicros:    l.getFloat64HistogramStatValue("rocksdb.compaction.times.cpu_micros"),
		NumFilesInSingleCompaction:  l.getFloat64HistogramStatValue("rocksdb.numfiles.in.singlecompaction"),
		ReadAmpEstimateUsefulBytes:  l.getInt64StatValue("rocksdb.read.amp.estimate.useful.bytes", count),
		ReadAmpTotalReadBytes:       l.getInt64StatValue("rocksdb.read.amp.total.read.bytes", count),
		NumberFileOpens:             l.getInt64StatValue("rocksdb.no.file.opens", count),
		NumberFileErrors:            l.getInt64StatValue("rocksdb.no.file.errors", count),
		BloomFilterUseful:           l.getInt64StatValue("rocksdb.bloom.filter.useful", count),
		BloomFilterFullPositive:     l.getInt64StatValue("rocksdb.bloom.filter.full.positive", count),
		BloomFilterFullTruePositive: l.getInt64StatValue("rocksdb.bloom.filter.full.true.positive", count),
		MemtableHit:                 l.getInt64StatValue("rocksdb.memtable.hit", count),
		MemtableMiss:                l.getInt64StatValue("rocksdb.memtable.miss", count),
		GetHitL0:                    l.getInt64StatValue("rocksdb.l0.hit", count),
		GetHitL1:                    l.getInt64StatValue("rocksdb.l1.hit", count),
		GetHitL2AndUp:               l.getInt64StatValue("rocksdb.l2andup.hit", count),
		BytesWritten:                l.getInt64StatValue("rocksdb.bytes.written", count),
		BytesRead:                   l.getInt64StatValue("rocksdb.bytes.read", count),
		StallMicros:                 l.getInt64StatValue("rocksdb.stall.micros", count),
		DBWriteStallHistogram:       l.getFloat64HistogramStatValue("rocksdb.db.write.stall"),
		LastLevelReadBytes:          l.getInt64StatValue("rocksdb.last.level.read.bytes", count),
		LastLevelReadCount:          l.getInt64StatValue("rocksdb.last.level.read.count", count),
		NonLastLevelReadBytes:       l.getInt64StatValue("rocksdb.non.last.level.read.bytes", count),
		NonLastLevelReadCount:       l.getInt64StatValue("rocksdb.non.last.level.read.count", count),
		DBGetMicros:                 l.getFloat64HistogramStatValue("rocksdb.db.get.micros"),
		DBWriteMicros:               l.getFloat64HistogramStatValue("rocksdb.db.write.micros"),
		BytesPerRead:                l.getFloat64HistogramStatValue("rocksdb.bytes.per.read"),
		BytesPerWrite:               l.getFloat64HistogramStatValue("rocksdb.bytes.per.write"),
		BytesPerMultiget:            l.getFloat64HistogramStatValue("rocksdb.bytes.per.multiget"),
		FlushMicros:                 l.getFloat64HistogramStatValue("rocksdb.db.flush.micros"),
	}

	err := l.error()
	if err != nil {
		return nil, err
	}

	return stats, nil
}

// getFloat64HistogramStatValue converts stat object into float64Histogram
func (l *statLoader) getFloat64HistogramStatValue(statName string) *float64Histogram {
	return &float64Histogram{
		Sum:   l.getFloat64StatValue(statName, sum),
		Count: l.getFloat64StatValue(statName, count),
		P50:   l.getFloat64StatValue(statName, p50),
		P95:   l.getFloat64StatValue(statName, p95),
		P99:   l.getFloat64StatValue(statName, p99),
		P100:  l.getFloat64StatValue(statName, p100),
	}
}

// getInt64StatValue converts property of stat object into int64
func (l *statLoader) getInt64StatValue(statName, propName string) int64 {
	stringVal := l.getStatValue(statName, propName)
	if stringVal == "" {
		l.errors = append(l.errors, fmt.Errorf("can't get stat by name: %v", statName))
		return 0
	}

	intVal, err := strconv.ParseInt(stringVal, 10, 64)
	if err != nil {
		l.errors = append(l.errors, fmt.Errorf("can't parse int: %v", err))
		return 0
	}

	return intVal
}

// getFloat64StatValue converts property of stat object into float64
func (l *statLoader) getFloat64StatValue(statName, propName string) float64 {
	stringVal := l.getStatValue(statName, propName)
	if stringVal == "" {
		l.errors = append(l.errors, fmt.Errorf("can't get stat by name: %v", statName))
		return 0
	}

	floatVal, err := strconv.ParseFloat(stringVal, 64)
	if err != nil {
		l.errors = append(l.errors, fmt.Errorf("can't parse float: %v", err))
		return 0
	}

	return floatVal
}

// getStatValue gets property of stat object
func (l *statLoader) getStatValue(statName, propName string) string {
	stat, ok := l.statMap[statName]
	if !ok {
		l.errors = append(l.errors, fmt.Errorf("stat %v doesn't exist", statName))
		return ""
	}
	prop, ok := stat.props[propName]
	if !ok {
		l.errors = append(l.errors, fmt.Errorf("stat %v doesn't have %v property", statName, propName))
		return ""
	}

	return prop
}
