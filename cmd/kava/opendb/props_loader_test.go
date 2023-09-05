//go:build rocksdb
// +build rocksdb

package opendb

import (
	"testing"

	"github.com/stretchr/testify/require"
)

type mockPropsGetter struct {
	props    map[string]string
	intProps map[string]uint64
}

func newMockPropsGetter(
	props map[string]string,
	intProps map[string]uint64,
) *mockPropsGetter {
	return &mockPropsGetter{
		props:    props,
		intProps: intProps,
	}
}

func (m *mockPropsGetter) GetProperty(propName string) string {
	return m.props[propName]
}

func (m *mockPropsGetter) GetIntProperty(propName string) (uint64, bool) {
	prop, ok := m.intProps[propName]
	return prop, ok
}

func TestPropsLoader(t *testing.T) {
	defaultProps := map[string]string{
		"rocksdb.options-statistics": "1",
	}
	defaultIntProps := map[string]uint64{
		"rocksdb.base-level":                 1,
		"rocksdb.block-cache-capacity":       2,
		"rocksdb.block-cache-pinned-usage":   3,
		"rocksdb.block-cache-usage":          4,
		"rocksdb.cur-size-active-mem-table":  5,
		"rocksdb.cur-size-all-mem-tables":    6,
		"rocksdb.estimate-live-data-size":    7,
		"rocksdb.estimate-num-keys":          8,
		"rocksdb.estimate-table-readers-mem": 9,
		"rocksdb.live-sst-files-size":        10,
		"rocksdb.size-all-mem-tables":        11,
	}
	missingProps := make(map[string]string)
	missingIntProps := make(map[string]uint64)
	defaultExpectedProps := properties{
		BaseLevel:               1,
		BlockCacheCapacity:      2,
		BlockCachePinnedUsage:   3,
		BlockCacheUsage:         4,
		CurSizeActiveMemTable:   5,
		CurSizeAllMemTables:     6,
		EstimateLiveDataSize:    7,
		EstimateNumKeys:         8,
		EstimateTableReadersMem: 9,
		LiveSSTFilesSize:        10,
		SizeAllMemTables:        11,
		OptionsStatistics:       "1",
	}

	for _, tc := range []struct {
		desc          string
		props         map[string]string
		intProps      map[string]uint64
		expectedProps *properties
		success       bool
	}{
		{
			desc:          "success case",
			props:         defaultProps,
			intProps:      defaultIntProps,
			expectedProps: &defaultExpectedProps,
			success:       true,
		},
		{
			desc:          "missing props",
			props:         missingProps,
			intProps:      defaultIntProps,
			expectedProps: nil,
			success:       false,
		},
		{
			desc:          "missing integer props",
			props:         defaultProps,
			intProps:      missingIntProps,
			expectedProps: nil,
			success:       false,
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			mockPropsGetter := newMockPropsGetter(tc.props, tc.intProps)

			propsLoader := newPropsLoader(mockPropsGetter)
			actualProps, err := propsLoader.load()
			if tc.success {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
			require.Equal(t, tc.expectedProps, actualProps)
		})
	}
}
