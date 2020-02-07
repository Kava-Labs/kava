package types

// bep3 module event types
const (
	// EventTypeCreateHtlt  = "HTLT"
	EventTypeCreateHtlt  = "createHTLT"
	EventTypeDepositHtlt = "depositHTLT"
	EventTypeRefundHtlt  = "claimHTLT"
	EventTypeClaimHtlt   = "refundHTLT"

	AttributeKeyHtltSwapID = "htlt_swap_id"
	AttributeKeyFrom       = "htlt_from"
	AttributeKeyTo         = "htlt_to"
	AttributeKeyCoinDenom  = "coin_denom"
	AttributeKeyCoinAmount = "coin_amount"

	AttributeValueCategory = ModuleName
)
