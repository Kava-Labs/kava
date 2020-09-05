package types

// Event types for cdp module
const (
	EventTypeHarvest           = ModuleName
	EventTypeClaim             = "claim_reward"
	EventTypeRewardPeriod      = "new_reward_period"
	EventTypeClaimPeriod       = "new_claim_period"
	EventTypeClaimPeriodExpiry = "claim_period_expiry"
	EventTypeLPDeposit         = "lp_deposit"
	EventTypeGovDeposit        = "gov_deposit"

	AttributeValueCategory   = ModuleName
	AttributeKeyDeposit      = "deposit"
	AttributeKeyClaimedBy    = "claimed_by"
	AttributeKeyClaimAmount  = "claim_amount"
	AttributeKeyRewardPeriod = "reward_period"
	AttributeKeyClaimPeriod  = "claim_period"
)
