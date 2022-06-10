package types

// MultiLiquidStakingHooks combine multiple LiquidStaking hooks, all hook functions are run in array sequence
type MultiLiquidStakingHooks []LiquidStakingHooks

// NewMultiLiquidStakingHooks returns a new MultiLiquidStakingHooks
func NewMultiLiquidStakingHooks(hooks ...LiquidStakingHooks) MultiLiquidStakingHooks {
	return hooks
}
