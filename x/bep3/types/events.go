package types

// Events for bep3 module
const (
	EventTypeCreateAtomicSwap = "create_atomic_swap"
	EventTypeClaimAtomicSwap  = "claim_atomic_swap"
	EventTypeRefundAtomicSwap = "refund_atomic_swap"
	EventTypeSwapsExpired     = "swaps_expired"

	AttributeValueCategory       = ModuleName
	AttributeKeySender           = "sender"
	AttributeKeyRecipient        = "recipient"
	AttributeKeyAtomicSwapID     = "atomic_swap_id"
	AttributeKeyRandomNumberHash = "random_number_hash"
	AttributeKeyTimestamp        = "timestamp"
	AttributeKeySenderOtherChain = "sender_other_chain"
	AttributeKeyExpireHeight     = "expire_height"
	AttributeKeyAmount           = "amount"
	AttributeKeyDirection        = "direction"
	AttributeKeyClaimSender      = "claim_sender"
	AttributeKeyRandomNumber     = "random_number"
	AttributeKeyRefundSender     = "refund_sender"
	AttributeKeyAtomicSwapIDs    = "atomic_swap_ids"
	AttributeExpirationBlock     = "expiration_block"
)
