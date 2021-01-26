package types

// Events emitted by the incentive module
const (
	EventTypeClaim             = "claim_reward"
	EventTypeRewardPeriod      = "new_reward_period"
	EventTypeClaimPeriod       = "new_claim_period"
	EventTypeClaimPeriodExpiry = "claim_period_expiry"

	AttributeValueCategory   = ModuleName
	AttributeKeyClaimedBy    = "claimed_by"
	AttributeKeyClaimAmount  = "claim_amount"
	AttributeKeyClaimType    = "claim_type"
	AttributeKeyRewardPeriod = "reward_period"
	AttributeKeyClaimPeriod  = "claim_period"
)
