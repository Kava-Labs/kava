package types

// Events emitted by the issuance module
const (
	EventTypeIssue           = "issue_tokens"
	EventTypeRedeem          = "redeem_tokens"
	EventTypeBlock           = "block_address"
	EventTypeUnblock         = "unblock_address"
	EventTypePause           = "change_pause_status"
	EventTypeSeize           = "seize_coins_from_blocked_address"
	AttributeValueCategory   = ModuleName
	AttributeKeyDenom        = "denom"
	AttributeKeyIssueAmount  = "amount_issued"
	AttributeKeyRedeemAmount = "amount_redeemed"
	AttributeKeyBlock        = "address_blocked"
	AttributeKeyUnblock      = "address_unblocked"
	AttributeKeyAddress      = "address"
	AttributeKeyPauseStatus  = "pause_status"
)
