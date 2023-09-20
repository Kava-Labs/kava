//go:build rocksdb
// +build rocksdb

package opendb

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseSerializedStats(t *testing.T) {
	defaultSerializedStats := `rocksdb.block.cache.miss COUNT : 1
rocksdb.block.cache.hit COUNT : 2
rocksdb.block.cache.add COUNT : 3
rocksdb.block.cache.add.failures COUNT : 4
rocksdb.compaction.times.micros P50 : 1 P95 : 2 P99 : 3 P100 : 4 COUNT : 5 SUM : 6
rocksdb.compaction.times.cpu_micros P50 : 7 P95 : 8 P99 : 9 P100 : 10 COUNT : 11 SUM : 12
`
	defaultExpectedStatMap := map[string]*stat{
		"rocksdb.block.cache.miss": {
			name: "rocksdb.block.cache.miss",
			props: map[string]string{
				"COUNT": "1",
			},
		},
		"rocksdb.block.cache.hit": {
			name: "rocksdb.block.cache.hit",
			props: map[string]string{
				"COUNT": "2",
			},
		},
		"rocksdb.block.cache.add": {
			name: "rocksdb.block.cache.add",
			props: map[string]string{
				"COUNT": "3",
			},
		},
		"rocksdb.block.cache.add.failures": {
			name: "rocksdb.block.cache.add.failures",
			props: map[string]string{
				"COUNT": "4",
			},
		},
		"rocksdb.compaction.times.micros": {
			name: "rocksdb.compaction.times.micros",
			props: map[string]string{
				"P50":   "1",
				"P95":   "2",
				"P99":   "3",
				"P100":  "4",
				"COUNT": "5",
				"SUM":   "6",
			},
		},
		"rocksdb.compaction.times.cpu_micros": {
			name: "rocksdb.compaction.times.cpu_micros",
			props: map[string]string{
				"P50":   "7",
				"P95":   "8",
				"P99":   "9",
				"P100":  "10",
				"COUNT": "11",
				"SUM":   "12",
			},
		},
	}

	for _, tc := range []struct {
		desc            string
		serializedStats string
		expectedStatMap map[string]*stat
		errMsg          string
	}{
		{
			desc:            "success case",
			serializedStats: defaultSerializedStats,
			expectedStatMap: defaultExpectedStatMap,
			errMsg:          "",
		},
		{
			desc: "missing value #1",
			serializedStats: `rocksdb.block.cache.miss COUNT :
`,
			expectedStatMap: nil,
			errMsg:          "invalid number of tokens",
		},
		{
			desc: "missing value #2",
			serializedStats: `rocksdb.compaction.times.micros P50 : 1 P95 :
`,
			expectedStatMap: nil,
			errMsg:          "invalid number of tokens",
		},
		{
			desc: "missing stat name",
			serializedStats: ` COUNT : 1
`,
			expectedStatMap: nil,
			errMsg:          "stat name shouldn't be empty",
		},
		{
			desc:            "empty stat",
			serializedStats: ``,
			expectedStatMap: make(map[string]*stat),
			errMsg:          "",
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			actualStatMap, err := parseSerializedStats(tc.serializedStats)
			if tc.errMsg == "" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.errMsg)
			}
			require.Equal(t, tc.expectedStatMap, actualStatMap)
		})
	}
}

func TestValidateTokens(t *testing.T) {
	for _, tc := range []struct {
		desc   string
		tokens []string
		errMsg string
	}{
		{
			desc:   "success case",
			tokens: []string{"name", "key", ":", "value"},
			errMsg: "",
		},
		{
			desc:   "missing value #1",
			tokens: []string{"name", "key", ":"},
			errMsg: "invalid number of tokens",
		},
		{
			desc:   "missing value #2",
			tokens: []string{"name", "key", ":", "value", "key2", ":"},
			errMsg: "invalid number of tokens",
		},
		{
			desc:   "empty stat name",
			tokens: []string{"", "key", ":", "value"},
			errMsg: "stat name shouldn't be empty",
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			err := validateTokens(tc.tokens)
			if tc.errMsg == "" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.errMsg)
			}
		})
	}
}

func TestValidateStatProperty(t *testing.T) {
	for _, tc := range []struct {
		desc   string
		key    string
		value  string
		sep    string
		errMsg string
	}{
		{
			desc:   "success case",
			key:    "key",
			value:  "value",
			sep:    ":",
			errMsg: "",
		},
		{
			desc:   "missing key",
			key:    "",
			value:  "value",
			sep:    ":",
			errMsg: "key shouldn't be empty",
		},
		{
			desc:   "missing value",
			key:    "key",
			value:  "",
			sep:    ":",
			errMsg: "value shouldn't be empty",
		},
		{
			desc:   "invalid separator",
			key:    "key",
			value:  "value",
			sep:    "#",
			errMsg: "separator should be :",
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			err := validateStatProperty(tc.key, tc.value, tc.sep)
			if tc.errMsg == "" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.errMsg)
			}
		})
	}
}
