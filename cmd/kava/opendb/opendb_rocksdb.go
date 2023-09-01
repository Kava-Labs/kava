//go:build rocksdb
// +build rocksdb

// Copyright 2023 Kava Labs, Inc.
// Copyright 2023 Cronos Labs, Inc.
//
// Derived from https://github.com/crypto-org-chain/cronos@496ce7e
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package opendb

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/cosmos/cosmos-sdk/server/types"
	"github.com/linxGnu/grocksdb"
	"github.com/spf13/cast"
	dbm "github.com/tendermint/tm-db"
)

var ErrUnexpectedConfiguration = errors.New("unexpected rocksdb configuration, rocksdb should have only one column family named default")

const (
	// default tm-db block cache size for RocksDB
	blockCacheSize = 1 << 30

	defaultColumnFamilyName = "default"

	enableMetricsOptName             = "rocksdb.enable-metrics"
	reportMetricsIntervalSecsOptName = "rocksdb.report-metrics-interval-secs"
	defaultReportMetricsIntervalSecs = 15

	maxOpenFilesDBOptName          = "rocksdb.max-open-files"
	maxFileOpeningThreadsDBOptName = "rocksdb.max-file-opening-threads"

	writeBufferSizeCFOptName = "rocksdb.write-buffer-size"
	numLevelsCFOptName       = "rocksdb.num-levels"
)

func OpenDB(appOpts types.AppOptions, home string, backendType dbm.BackendType) (dbm.DB, error) {
	dataDir := filepath.Join(home, "data")
	if backendType == dbm.RocksDBBackend {
		return openRocksdb(filepath.Join(dataDir, "application.db"), appOpts)
	}

	return dbm.NewDB("application", backendType, dataDir)
}

// openRocksdb loads existing options, overrides some of them with appOpts and opens database
// option will be overridden only in case if it explicitly specified in appOpts
func openRocksdb(dir string, appOpts types.AppOptions) (dbm.DB, error) {
	dbOpts, cfOpts, err := loadLatestOptions(dir)
	if err != nil {
		return nil, err
	}
	// customize rocksdb options
	dbOpts = overrideDBOpts(dbOpts, appOpts)
	cfOpts = overrideCFOpts(cfOpts, appOpts)

	enableMetrics := cast.ToBool(appOpts.Get(enableMetricsOptName))
	reportMetricsIntervalSecs := cast.ToInt64(appOpts.Get(reportMetricsIntervalSecsOptName))
	if reportMetricsIntervalSecs == 0 {
		reportMetricsIntervalSecs = defaultReportMetricsIntervalSecs
	}

	return newRocksDBWithOptions("application", dir, dbOpts, cfOpts, enableMetrics, reportMetricsIntervalSecs)
}

// loadLatestOptions loads and returns database and column family options
// if options file not found, it means database isn't created yet, in such case default tm-db options will be returned
// if database exists it should have only one column family named default
func loadLatestOptions(dir string) (*grocksdb.Options, *grocksdb.Options, error) {
	latestOpts, err := grocksdb.LoadLatestOptions(dir, grocksdb.NewDefaultEnv(), true, grocksdb.NewLRUCache(blockCacheSize))
	if err != nil && strings.HasPrefix(err.Error(), "NotFound: ") {
		return newDefaultOptions(), newDefaultOptions(), nil
	}
	if err != nil {
		return nil, nil, err
	}

	cfNames := latestOpts.ColumnFamilyNames()
	cfOpts := latestOpts.ColumnFamilyOpts()
	// db should have only one column family named default
	ok := len(cfNames) == 1 && cfNames[0] == defaultColumnFamilyName
	if !ok {
		return nil, nil, ErrUnexpectedConfiguration
	}

	// return db and cf opts
	return latestOpts.Options(), &cfOpts[0], nil
}

// overrideDBOpts merges dbOpts and appOpts, appOpts takes precedence
func overrideDBOpts(dbOpts *grocksdb.Options, appOpts types.AppOptions) *grocksdb.Options {
	maxOpenFiles := appOpts.Get(maxOpenFilesDBOptName)
	if maxOpenFiles != nil {
		dbOpts.SetMaxOpenFiles(cast.ToInt(maxOpenFiles))
	}

	maxFileOpeningThreads := appOpts.Get(maxFileOpeningThreadsDBOptName)
	if maxFileOpeningThreads != nil {
		dbOpts.SetMaxFileOpeningThreads(cast.ToInt(maxFileOpeningThreads))
	}

	return dbOpts
}

// overrideCFOpts merges cfOpts and appOpts, appOpts takes precedence
func overrideCFOpts(cfOpts *grocksdb.Options, appOpts types.AppOptions) *grocksdb.Options {
	writeBufferSize := appOpts.Get(writeBufferSizeCFOptName)
	if writeBufferSize != nil {
		cfOpts.SetWriteBufferSize(cast.ToUint64(writeBufferSize))
	}

	numLevels := appOpts.Get(numLevelsCFOptName)
	if numLevels != nil {
		cfOpts.SetNumLevels(cast.ToInt(numLevels))
	}

	return cfOpts
}

// newRocksDBWithOptions opens rocksdb with provided database and column family options
// newRocksDBWithOptions expects that db has only one column family named default
func newRocksDBWithOptions(
	name string,
	dir string,
	dbOpts *grocksdb.Options,
	cfOpts *grocksdb.Options,
	enableMetrics bool,
	reportMetricsIntervalSecs int64,
) (*dbm.RocksDB, error) {
	dbPath := filepath.Join(dir, name+".db")

	// Ensure path exists
	if err := os.MkdirAll(dbPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create db path: %w", err)
	}

	// EnableStatistics adds overhead so shouldn't be enabled in production
	if enableMetrics {
		dbOpts.EnableStatistics()
	}

	db, _, err := grocksdb.OpenDbColumnFamilies(dbOpts, dbPath, []string{defaultColumnFamilyName}, []*grocksdb.Options{cfOpts})
	if err != nil {
		return nil, err
	}

	if enableMetrics {
		registerMetrics()
		go reportMetrics(db, time.Second*time.Duration(reportMetricsIntervalSecs))
	}

	ro := grocksdb.NewDefaultReadOptions()
	wo := grocksdb.NewDefaultWriteOptions()
	woSync := grocksdb.NewDefaultWriteOptions()
	woSync.SetSync(true)
	return dbm.NewRocksDBWithRawDB(db, ro, wo, woSync), nil
}

// newDefaultOptions returns default tm-db options for RocksDB, see for details:
// https://github.com/Kava-Labs/tm-db/blob/94ff76d31724965f8883cddebabe91e0d01bc03f/rocksdb.go#L30
func newDefaultOptions() *grocksdb.Options {
	// default rocksdb option, good enough for most cases, including heavy workloads.
	// 1GB table cache, 512MB write buffer(may use 50% more on heavy workloads).
	// compression: snappy as default, need to -lsnappy to enable.
	bbto := grocksdb.NewDefaultBlockBasedTableOptions()
	bbto.SetBlockCache(grocksdb.NewLRUCache(1 << 30))
	bbto.SetFilterPolicy(grocksdb.NewBloomFilter(10))

	opts := grocksdb.NewDefaultOptions()
	opts.SetBlockBasedTableFactory(bbto)
	// SetMaxOpenFiles to 4096 seems to provide a reliable performance boost
	opts.SetMaxOpenFiles(4096)
	opts.SetCreateIfMissing(true)
	opts.IncreaseParallelism(runtime.NumCPU())
	// 1.5GB maximum memory use for writebuffer.
	opts.OptimizeLevelStyleCompaction(512 * 1024 * 1024)

	return opts
}

// reportMetrics periodically requests stats from rocksdb and reports to prometheus
// NOTE: should be launched as a goroutine
func reportMetrics(db *grocksdb.DB, interval time.Duration) {
	ticker := time.NewTicker(interval)
	for {
		select {
		case <-ticker.C:
			props, stats, err := getPropsAndStats(db)
			if err != nil {
				continue
			}

			rocksdbMetrics.report(props, stats)
		}
	}
}

// getPropsAndStats gets statistics from rocksdb
func getPropsAndStats(db *grocksdb.DB) (*properties, *stats, error) {
	propsLoader := newPropsLoader(db)
	props, err := propsLoader.load()
	if err != nil {
		return nil, nil, err
	}

	statMap, err := parseSerializedStats(props.OptionsStatistics)
	if err != nil {
		return nil, nil, err
	}

	statLoader := newStatLoader(statMap)
	stats, err := statLoader.load()
	if err != nil {
		return nil, nil, err
	}

	return props, stats, nil
}
