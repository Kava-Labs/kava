//go:build rocksdb
// +build rocksdb

package opendb

import (
	"fmt"
	"strings"

	"errors"
)

// stat represents one line from rocksdb statistics data, stat may have one or more properties
// examples:
// - rocksdb.block.cache.miss COUNT : 5
// - rocksdb.compaction.times.micros P50 : 21112 P95 : 21112 P99 : 21112 P100 : 21112 COUNT : 1 SUM : 21112
// `rocksdb.compaction.times.micros` is name of stat, P50, COUNT, SUM, etc... are props of stat
type stat struct {
	name  string
	props map[string]string
}

// parseSerializedStats parses serialisedStats into map of stat objects
// example of serializedStats:
// rocksdb.block.cache.miss COUNT : 5
// rocksdb.compaction.times.micros P50 : 21112 P95 : 21112 P99 : 21112 P100 : 21112 COUNT : 1 SUM : 21112
func parseSerializedStats(serializedStats string) (map[string]*stat, error) {
	stats := make(map[string]*stat, 0)

	serializedStatList := strings.Split(serializedStats, "\n")
	if len(serializedStatList) == 0 {
		return nil, errors.New("serializedStats is empty")
	}
	serializedStatList = serializedStatList[:len(serializedStatList)-1]
	// iterate over stats line by line
	for _, serializedStat := range serializedStatList {
		stat, err := parseSerializedStat(serializedStat)
		if err != nil {
			return nil, err
		}

		stats[stat.name] = stat
	}

	return stats, nil
}

// parseSerializedStat parses serialisedStat into stat object
// example of serializedStat:
// rocksdb.block.cache.miss COUNT : 5
func parseSerializedStat(serializedStat string) (*stat, error) {
	tokens := strings.Split(serializedStat, " ")
	tokensNum := len(tokens)
	if err := validateTokens(tokens); err != nil {
		return nil, fmt.Errorf("tokens are invalid: %v", err)
	}

	props := make(map[string]string)
	for idx := 1; idx < tokensNum; idx += 3 {
		// never should happen, but double check to avoid unexpected panic
		if idx+2 >= tokensNum {
			break
		}

		key := tokens[idx]
		sep := tokens[idx+1]
		value := tokens[idx+2]

		if err := validateStatProperty(key, value, sep); err != nil {
			return nil, fmt.Errorf("invalid stat property: %v", err)
		}

		props[key] = value
	}

	return &stat{
		name:  tokens[0],
		props: props,
	}, nil
}

// validateTokens validates that tokens contains name + N triples (key, sep, value)
func validateTokens(tokens []string) error {
	tokensNum := len(tokens)
	if tokensNum < 4 {
		return fmt.Errorf("invalid number of tokens: %v, tokens: %v", tokensNum, tokens)
	}
	if (tokensNum-1)%3 != 0 {
		return fmt.Errorf("invalid number of tokens: %v, tokens: %v", tokensNum, tokens)
	}
	if tokens[0] == "" {
		return fmt.Errorf("stat name shouldn't be empty")
	}

	return nil
}

// validateStatProperty validates that key and value are divided by separator and aren't empty
func validateStatProperty(key, value, sep string) error {
	if key == "" {
		return fmt.Errorf("key shouldn't be empty")
	}
	if sep != ":" {
		return fmt.Errorf("separator should be :")
	}
	if value == "" {
		return fmt.Errorf("value shouldn't be empty")
	}

	return nil
}
