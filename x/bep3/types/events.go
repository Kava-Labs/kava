package types

// bep3 module event types
const (
	EventTypeCreateAtomicSwap  = "createAtomicSwap"
	EventTypeDepositAtomicSwap = "depositAtomicSwap"
	EventTypeClaimAtomicSwap   = "claimAtomicSwap"
	EventTypeRefundAtomicSwap  = "refundAtomicSwap"

	AttributeKeyAtomicSwapID     = "swap_id"
	AttributeKeyRandomNumberHash = "random_number_hash"
	AttributeKeyFrom             = "from"
	AttributeKeyTo               = "to"
	AttributeKeyCoinDenom        = "coin_denom"
	AttributeKeyCoinAmount       = "coin_amount"
	AttributeKeyClaimer          = "claimer"

	AttributeValueCategory = ModuleName
)
