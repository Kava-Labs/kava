package types

// cdp module event types

const (
	EventTypeCreateCdp     = "create_cdp"
	EventTypeCdpDeposit    = "cdp_deposit"
	EventTypeCdpDraw       = "cdp_draw"
	EventTypeCdpRepay      = "cdp_repayment"
	EventTypeCdpClose      = "cdp_close"
	EventTypeCdpWithdrawal = "cdp_withdrawal"

	AttributeKeyCdpID      = "cdp_id"
	AttributeValueCategory = "cdp"
)
