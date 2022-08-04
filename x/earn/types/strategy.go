package types

import (
	"fmt"
	"strings"
)

// IsValid returns true if the StrategyType status is valid and false otherwise.
func (s StrategyType) IsValid() bool {
	return s == STRATEGY_TYPE_HARD || s == STRATEGY_TYPE_SAVINGS
}

// Validate returns an error if the StrategyType is invalid.
func (s StrategyType) Validate() error {
	if !s.IsValid() {
		return fmt.Errorf("invalid strategy %s", s)
	}

	return nil
}

// NewStrategyTypeFromString converts string to StrategyType type
func NewStrategyTypeFromString(str string) StrategyType {
	switch strings.ToLower(str) {
	case "hard":
		return STRATEGY_TYPE_HARD
	case "savings":
		return STRATEGY_TYPE_SAVINGS
	default:
		return STRATEGY_TYPE_UNSPECIFIED
	}
}

// StrategyTypes defines a slice of StrategyType
type StrategyTypes []StrategyType

// Validate returns an error if StrategyTypes are invalid.
func (strategies StrategyTypes) Validate() error {
	if len(strategies) == 0 {
		return fmt.Errorf("empty StrategyTypes")
	}

	if len(strategies) != 1 {
		return fmt.Errorf("must have exactly one strategy type, multiple strategies are not supported")
	}

	uniqueStrategies := make(map[StrategyType]bool)

	for _, strategy := range strategies {
		if err := strategy.Validate(); err != nil {
			return err
		}

		if _, found := uniqueStrategies[strategy]; found {
			return fmt.Errorf("duplicate strategy %s", strategy)
		}

		uniqueStrategies[strategy] = true
	}

	return nil
}
