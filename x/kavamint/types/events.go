package types

// Minting module event types
const (
	EventTypeMint = ModuleName

	AttributeKeyTotalSupply       = "total_supply"
	AttributeKeyTotalBonded       = "total_bonded"
	AttributeSecondsPassed        = "seconds_since_last_mint"
	AttributeKeyCommunityPoolMint = "minted_community_pool_inflation"
	AttributeKeyStakingRewardMint = "minted_staking_rewards"
)
