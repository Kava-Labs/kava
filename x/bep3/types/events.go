package types

// bep3 module event types
const (
	EventTypeCreateHtlt  = "createHTLT"
	EventTypeDepositHtlt = "depositHTLT"
	EventTypeClaimHtlt   = "claimHTLT"
	EventTypeRefundHtlt  = "refundHTLT"

	AttributeKeyHtltSwapID       = "htlt_swap_id"
	AttributeKeyRandomNumberHash = "htlt_random_number_hash"
	AttributeKeyFrom             = "htlt_from"
	AttributeKeyTo               = "htlt_to"
	AttributeKeyCoinDenom        = "coin_denom"
	AttributeKeyCoinAmount       = "coin_amount"
	AttributeKeyClaimer          = "htlt_claimer"

	AttributeValueCategory = ModuleName
)
