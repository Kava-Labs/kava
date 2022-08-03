package types_test

import (
	"testing"

	"github.com/kava-labs/kava/x/earn/types"
	"github.com/stretchr/testify/require"
)

func TestNewStrategyTypeFromString(t *testing.T) {
	tests := []struct {
		name     string
		strategy string
		expected types.StrategyType
	}{
		{
			name:     "hard",
			strategy: "hard",
			expected: types.STRATEGY_TYPE_HARD,
		},
		{
			name:     "savings",
			strategy: "savings",
			expected: types.STRATEGY_TYPE_SAVINGS,
		},
		{
			name:     "unspecified",
			strategy: "not a valid strategy name",
			expected: types.STRATEGY_TYPE_UNSPECIFIED,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actual := types.NewStrategyTypeFromString(tc.strategy)
			if actual != tc.expected {
				t.Errorf("expected %s, got %s", tc.expected, actual)
			}
		})
	}
}

func TestValidateStrategyTypes(t *testing.T) {
	type errArgs struct {
		expectPass bool
		contains   string
	}

	tests := []struct {
		name       string
		strategies types.StrategyTypes
		errArgs    errArgs
	}{
		{
			name:       "valid - hard",
			strategies: types.StrategyTypes{types.STRATEGY_TYPE_HARD},
			errArgs: errArgs{
				expectPass: true,
			},
		},
		{
			name:       "valid - savings",
			strategies: types.StrategyTypes{types.STRATEGY_TYPE_SAVINGS},
			errArgs: errArgs{
				expectPass: true,
			},
		},
		{
			name: "invalid - duplicate",
			strategies: types.StrategyTypes{
				types.STRATEGY_TYPE_SAVINGS,
				types.STRATEGY_TYPE_SAVINGS,
			},
			errArgs: errArgs{
				expectPass: false,
				// This will change to duplicate error if multiple strategies are supported
				contains: "must have exactly one strategy type, multiple strategies are not supported",
			},
		},
		{
			name:       "invalid - unspecified",
			strategies: types.StrategyTypes{types.STRATEGY_TYPE_UNSPECIFIED},
			errArgs: errArgs{
				expectPass: false,
				contains:   "invalid strategy",
			},
		},
		{
			name:       "invalid - zero",
			strategies: types.StrategyTypes{},
			errArgs: errArgs{
				expectPass: false,
				contains:   "empty StrategyTypes",
			},
		},
		{
			name: "invalid - more than 1",
			strategies: types.StrategyTypes{
				types.STRATEGY_TYPE_HARD,
				types.STRATEGY_TYPE_SAVINGS,
			},
			errArgs: errArgs{
				expectPass: false,
				contains:   "must have exactly one strategy type, multiple strategies are not supported",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.strategies.Validate()
			if tc.errArgs.expectPass {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.errArgs.contains)
			}
		})
	}
}
