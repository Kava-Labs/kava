package types

// Community module event types
const (
	EventTypeInflationStop      = "inflation_stop"
	EventTypeStakingRewardsPaid = "staking_rewards_paid"

	AttributeKeyStakingRewardAmount  = "staking_reward_amount"
	AttributeKeyInflationDisableTime = "inflation_disable_time"

	AttributeValueFundCommunityPool = "fund_community_pool"
	AttributeValueCategory          = ModuleName
)
