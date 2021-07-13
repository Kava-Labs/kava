package types

// Event types for swap module
const (
	AttributeValueCategory     = ModuleName
	EventTypeSwapDeposit       = "swap_deposit"
	EventTypeSwapWithdraw      = "swap_withdraw"
	EventTypeSwapTrade         = "swap_trade"
	AttributeKeyPoolID         = "pool_id"
	AttributeKeyDepositor      = "depositor"
	AttributeKeyShares         = "shares"
	AttributeKeyOwner          = "owner"
	AttributeKeyRequester      = "requester"
	AttributeKeySwapInput      = "input"
	AttributeKeySwapOutput     = "output"
	AttributeKeyFeePaid        = "fee"
	AttributeKeyExactDirection = "exact"
)
