package types

import sdk "github.com/cosmos/cosmos-sdk/types"

/*
Message types for starting various auctions.
Note: these message types are not final and will likely change.
Design options and problems:
 - msgs that only start auctions
	- senders have to pay fees
	- these msgs cannot be bundled into a tx with a PlaceBid msg because PlaceBid requires an auction ID
 - msgs that start auctions and place an initial bid
	- place bid can fail, leaving auction without bids which is similar to first case
 - no msgs, auctions started automatically
	- running this as an endblocker adds complexity and potential vulnerabilities
*/

// MsgSeizeAndStartCollateralAuction siezes a cdp that is below liquidation ratio and starts an auction for the collateral
type MsgSeizeAndStartCollateralAuction struct {
	Sender          sdk.AccAddress // only needed to pay the tx fees
	CdpOwner        sdk.AccAddress
	CollateralDenom string
}

// Route return the message type used for routing the message.
func (msg MsgSeizeAndStartCollateralAuction) Route() string { return "liquidator" }

// Type returns a human-readable string for the message, intended for utilization within tags.
func (msg MsgSeizeAndStartCollateralAuction) Type() string { return "seize_and_start_auction" } // TODO snake case?

// ValidateBasic does a simple validation check that doesn't require access to any other information.
func (msg MsgSeizeAndStartCollateralAuction) ValidateBasic() sdk.Error {
	if msg.Sender.Empty() {
		return sdk.ErrInternal("invalid (empty) sender address")
	}
	if msg.CdpOwner.Empty() {
		return sdk.ErrInternal("invalid (empty) CDP owner address")
	}
	// TODO check coin denoms
	return nil
}

// GetSignBytes gets the canonical byte representation of the Msg.
func (msg MsgSeizeAndStartCollateralAuction) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// GetSigners returns the addresses of signers that must sign.
func (msg MsgSeizeAndStartCollateralAuction) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Sender}
}

// MsgStartDebtAuction starts an auction of gov tokens for stable tokens
type MsgStartDebtAuction struct {
	Sender sdk.AccAddress // only needed to pay the tx fees
}

// Route returns the route for this message
func (msg MsgStartDebtAuction) Route() string { return "liquidator" }

// Type returns the type for this message
func (msg MsgStartDebtAuction) Type() string { return "start_debt_auction" }

// ValidateBasic simple validation check
func (msg MsgStartDebtAuction) ValidateBasic() sdk.Error {
	if msg.Sender.Empty() {
		return sdk.ErrInternal("invalid (empty) sender address")
	}
	return nil
}

// GetSignBytes returns canonical byte representation of the message
func (msg MsgStartDebtAuction) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

// GetSigners returns the addresses of the signers of the message
func (msg MsgStartDebtAuction) GetSigners() []sdk.AccAddress { return []sdk.AccAddress{msg.Sender} }

// With no stability and liquidation fees, surplus auctions can never be run.
// type MsgStartSurplusAuction struct {
// 	Sender sdk.AccAddress // only needed to pay the tx fees
// }

// func (msg MsgStartSurplusAuction) Route() string { return "liquidator" }
// func (msg MsgStartSurplusAuction) Type() string  { return "start_surplus_auction" } // TODO snake case?
// func (msg MsgStartSurplusAuction) ValidateBasic() sdk.Error {
// 	if msg.Sender.Empty() {
// 		return sdk.ErrInternal("invalid (empty) sender address")
// 	}
// 	return nil
// }
// func (msg MsgStartSurplusAuction) GetSignBytes() []byte {
// 	return sdk.MustSortJSON(msgCdc.MustMarshalJSON(msg))
// }
// func (msg MsgStartSurplusAuction) GetSigners() []sdk.AccAddress { return []sdk.AccAddress{msg.Sender} }
