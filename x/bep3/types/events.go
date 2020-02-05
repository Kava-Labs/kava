package types

// bep3 module event types
const (
	EventTypeCreateHtlt  = "create_htlt"
	EventTypeDepositHtlt = "deposit_htlt"
	EventTypeRefundHtlt  = "refund_htlt"
	EventTypeClaimHtlt   = "claim_htlt"

	AttributeKeyHtltSwapID = "htlt_swap_id"
	AttributeKeyFrom       = "htlt_from"
	AttributeKeyTo         = "htlt_to"
	AttributeKeyCoinDenom  = "coin_denom"
	AttributeKeyCoinAmount = "coin_amount"

	AttributeValueCategory = ModuleName
)
