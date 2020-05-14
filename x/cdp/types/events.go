package types

// Event types for cdp module
const (
	EventTypeCreateCdp         = "create_cdp"
	EventTypeCdpDeposit        = "cdp_deposit"
	EventTypeCdpDraw           = "cdp_draw"
	EventTypeCdpRepay          = "cdp_repayment"
	EventTypeCdpClose          = "cdp_close"
	EventTypeCdpWithdrawal     = "cdp_withdrawal"
	EventTypeCdpLiquidation    = "cdp_liquidation"
	EventTypeBeginBlockerFatal = "cdp_begin_block_error"

	AttributeKeyCdpID      = "cdp_id"
	AttributeKeyDeposit    = "deposit"
	AttributeValueCategory = "cdp"
	AttributeKeyError      = "error_message"
)
