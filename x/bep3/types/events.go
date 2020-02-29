package types

// bep3 module event types
const (
	EventTypeCreateAtomicSwap  = "createAtomicSwap"
	EventTypeDepositAtomicSwap = "depositAtomicSwap"
	EventTypeClaimAtomicSwap   = "claimAtomicSwap"
	EventTypeRefundAtomicSwap  = "refundAtomicSwap"

	// Common
	AttributeKeySender           = "sender"
	AttributeKeyRecipient        = "recipient"
	AttributeKeyAtomicSwapID     = "atomic_swap_id"
	AttributeKeyRandomNumberHash = "random_number_hash"
	AttributeKeyTimestamp        = "timestamp"
	AttributeKeySenderOtherChain = "sender_other_chain"
	AttributeKeyExpireHeight     = "expire_height"
	AttributeKeyAmount           = "amount"
	AttributeKeyExpectedIncome   = "expected_income"

	// Claim
	AttributeKeyClaimSender  = "claim_sender"
	AttributeKeyRandomNumber = "random_number"

	// Refund
	AttributeKeyRefundSender = "refund_sender"

	AttributeValueCategory = ModuleName
)
