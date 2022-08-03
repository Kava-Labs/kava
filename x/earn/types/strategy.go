package types

import "fmt"

// Validate returns an error if the StrategyType is invalid.
func (s StrategyType) Validate() error {
	if s == STRATEGY_TYPE_UNSPECIFIED {
		return ErrInvalidVaultStrategy
	}

	// Check if out of range
	_, ok := StrategyType_name[int32(s)]
	if !ok {
		return fmt.Errorf("invalid strategy %s", s)
	}

	return nil
}

// StrategyTypes defines a slice of StrategyType
type StrategyTypes []StrategyType

// Validate returns an error if StrategyTypes are invalid.
func (strategies StrategyTypes) Validate() error {
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
