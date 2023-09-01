//go:build rocksdb
// +build rocksdb

package opendb

import (
	"fmt"
	"strings"

	"errors"
)

type propsGetter interface {
	GetProperty(propName string) (value string)
	GetIntProperty(propName string) (value uint64, success bool)
}

type propsLoader struct {
	db        propsGetter
	errorMsgs []string
}

func newPropsLoader(db propsGetter) *propsLoader {
	return &propsLoader{
		db:        db,
		errorMsgs: make([]string, 0),
	}
}

func (l *propsLoader) load() (*properties, error) {
	props := &properties{
		BaseLevel:               l.getIntProperty("rocksdb.base-level"),
		BlockCacheCapacity:      l.getIntProperty("rocksdb.block-cache-capacity"),
		BlockCachePinnedUsage:   l.getIntProperty("rocksdb.block-cache-pinned-usage"),
		BlockCacheUsage:         l.getIntProperty("rocksdb.block-cache-usage"),
		CurSizeActiveMemTable:   l.getIntProperty("rocksdb.cur-size-active-mem-table"),
		CurSizeAllMemTables:     l.getIntProperty("rocksdb.cur-size-all-mem-tables"),
		EstimateLiveDataSize:    l.getIntProperty("rocksdb.estimate-live-data-size"),
		EstimateNumKeys:         l.getIntProperty("rocksdb.estimate-num-keys"),
		EstimateTableReadersMem: l.getIntProperty("rocksdb.estimate-table-readers-mem"),
		LiveSSTFilesSize:        l.getIntProperty("rocksdb.live-sst-files-size"),
		SizeAllMemTables:        l.getIntProperty("rocksdb.size-all-mem-tables"),
		OptionsStatistics:       l.getProperty("rocksdb.options-statistics"),
	}

	if len(l.errorMsgs) != 0 {
		errorMsg := strings.Join(l.errorMsgs, ";")
		return nil, errors.New(errorMsg)
	}

	return props, nil
}

func (l *propsLoader) getProperty(propName string) string {
	value := l.db.GetProperty(propName)
	if value == "" {
		l.errorMsgs = append(l.errorMsgs, fmt.Sprintf("property %v is empty", propName))
		return ""
	}

	return value
}

func (l *propsLoader) getIntProperty(propName string) uint64 {
	value, ok := l.db.GetIntProperty(propName)
	if !ok {
		l.errorMsgs = append(l.errorMsgs, fmt.Sprintf("can't get %v int property", propName))
		return 0
	}

	return value
}

type properties struct {
	BaseLevel               uint64
	BlockCacheCapacity      uint64
	BlockCachePinnedUsage   uint64
	BlockCacheUsage         uint64
	CurSizeActiveMemTable   uint64
	CurSizeAllMemTables     uint64
	EstimateLiveDataSize    uint64
	EstimateNumKeys         uint64
	EstimateTableReadersMem uint64
	LiveSSTFilesSize        uint64
	SizeAllMemTables        uint64
	OptionsStatistics       string
}
