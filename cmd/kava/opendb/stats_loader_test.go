//go:build rocksdb
// +build rocksdb

package opendb

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStatsLoader(t *testing.T) {
	defaultStat := stat{
		props: map[string]string{
			"COUNT": "1",
		},
	}
	defaultHistogramStat := stat{
		props: map[string]string{
			"P50":   "1",
			"P95":   "2",
			"P99":   "3",
			"P100":  "4",
			"COUNT": "5",
			"SUM":   "6",
		},
	}
	defaultStatMap := map[string]*stat{
		"rocksdb.number.keys.written":             &defaultStat,
		"rocksdb.number.keys.read":                &defaultStat,
		"rocksdb.number.keys.updated":             &defaultStat,
		"rocksdb.block.cache.miss":                &defaultStat,
		"rocksdb.block.cache.hit":                 &defaultStat,
		"rocksdb.block.cache.add":                 &defaultStat,
		"rocksdb.block.cache.add.failures":        &defaultStat,
		"rocksdb.block.cache.index.miss":          &defaultStat,
		"rocksdb.block.cache.index.hit":           &defaultStat,
		"rocksdb.block.cache.index.bytes.insert":  &defaultStat,
		"rocksdb.block.cache.filter.miss":         &defaultStat,
		"rocksdb.block.cache.filter.hit":          &defaultStat,
		"rocksdb.block.cache.filter.bytes.insert": &defaultStat,
		"rocksdb.block.cache.data.miss":           &defaultStat,
		"rocksdb.block.cache.data.hit":            &defaultStat,
		"rocksdb.block.cache.data.bytes.insert":   &defaultStat,
		"rocksdb.compact.read.bytes":              &defaultStat,
		"rocksdb.compact.write.bytes":             &defaultStat,
		"rocksdb.compaction.times.micros":         &defaultHistogramStat,
		"rocksdb.compaction.times.cpu_micros":     &defaultHistogramStat,
		"rocksdb.numfiles.in.singlecompaction":    &defaultHistogramStat,
		"rocksdb.read.amp.estimate.useful.bytes":  &defaultStat,
		"rocksdb.read.amp.total.read.bytes":       &defaultStat,
		"rocksdb.no.file.opens":                   &defaultStat,
		"rocksdb.no.file.errors":                  &defaultStat,
		"rocksdb.bloom.filter.useful":             &defaultStat,
		"rocksdb.bloom.filter.full.positive":      &defaultStat,
		"rocksdb.bloom.filter.full.true.positive": &defaultStat,
		"rocksdb.memtable.hit":                    &defaultStat,
		"rocksdb.memtable.miss":                   &defaultStat,
		"rocksdb.l0.hit":                          &defaultStat,
		"rocksdb.l1.hit":                          &defaultStat,
		"rocksdb.l2andup.hit":                     &defaultStat,
		"rocksdb.bytes.written":                   &defaultStat,
		"rocksdb.bytes.read":                      &defaultStat,
		"rocksdb.stall.micros":                    &defaultStat,
		"rocksdb.db.write.stall":                  &defaultHistogramStat,
		"rocksdb.last.level.read.bytes":           &defaultStat,
		"rocksdb.last.level.read.count":           &defaultStat,
		"rocksdb.non.last.level.read.bytes":       &defaultStat,
		"rocksdb.non.last.level.read.count":       &defaultStat,
		"rocksdb.db.get.micros":                   &defaultHistogramStat,
		"rocksdb.db.write.micros":                 &defaultHistogramStat,
		"rocksdb.bytes.per.read":                  &defaultHistogramStat,
		"rocksdb.bytes.per.write":                 &defaultHistogramStat,
		"rocksdb.bytes.per.multiget":              &defaultHistogramStat,
		"rocksdb.db.flush.micros":                 &defaultHistogramStat,
	}

	statLoader := newStatLoader(defaultStatMap)
	stats, err := statLoader.load()
	require.NoError(t, err)

	require.Equal(t, stats.NumberKeysWritten, int64(1))
	require.Equal(t, stats.NumberKeysRead, int64(1))
	require.Equal(t, stats.CompactionTimesMicros.P50, float64(1))
	require.Equal(t, stats.CompactionTimesMicros.P95, float64(2))
	require.Equal(t, stats.CompactionTimesMicros.P99, float64(3))
	require.Equal(t, stats.CompactionTimesMicros.P100, float64(4))
	require.Equal(t, stats.CompactionTimesMicros.Count, float64(5))
	require.Equal(t, stats.CompactionTimesMicros.Sum, float64(6))
}
