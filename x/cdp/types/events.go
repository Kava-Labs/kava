package types

// cdp module event types

const (
	EventTypeCreateCdp      = "create_cdp"
	EventTypeCdpDeposit     = "cdp_deposit"
	EventTypeCdpDraw        = "cdp_draw"
	EventTypeCdpRepay       = "cdp_repayment"
	EventTypeCdpClose       = "cdp_close"
	EventTypeCdpWithdrawal  = "cdp_withdrawal"
	EventTypeCdpLiquidation = "cdp_liquidation"

	AttributeKeyCdpID      = "cdp_id"
	AttributeKeyDepositor  = "depositor"
	AttributeValueCategory = "cdp"
)
