package types

// Events emitted by the issuance module
const (
	EventTypeIssue           = "issue_tokens"
	EventTypeRedeem          = "redeem_tokens"
	EventTypeBlock           = "block_address"
	EventTypePause           = "change_pause_status"
	AttributeValueCategory   = ModuleName
	AttributeKeyDenom        = "denom"
	AttributeKeyIssueAmount  = "amount_issued"
	AttributeKeyRedeemAmount = "amount_redeemed"
	AttributeKeyBlock        = "address_blocked"
	AttributeKeyPauseStatus  = "pause_status"
)
