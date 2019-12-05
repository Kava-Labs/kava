package types

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// TypeMsgPostPrice type of PostPrice msg
	TypeMsgPostPrice = "post_price"
)

// MsgPostPrice struct representing a posted price message.
// Used by oracles to input prices to the pricefeed
type MsgPostPrice struct {
	From      sdk.AccAddress // client that sent in this address
	AssetCode string         // asset code used by exchanges/api
	Price     sdk.Dec        // price in decimal (max precision 18)
	Expiry    time.Time      // expiry time
}

// NewMsgPostPrice creates a new post price msg
func NewMsgPostPrice(
	from sdk.AccAddress,
	assetCode string,
	price sdk.Dec,
	expiry time.Time) MsgPostPrice {
	return MsgPostPrice{
		From:      from,
		AssetCode: assetCode,
		Price:     price,
		Expiry:    expiry,
	}
}

// Route Implements Msg.
func (msg MsgPostPrice) Route() string { return RouterKey }

// Type Implements Msg
func (msg MsgPostPrice) Type() string { return TypeMsgPostPrice }

// GetSignBytes Implements Msg.
func (msg MsgPostPrice) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// GetSigners Implements Msg.
func (msg MsgPostPrice) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.From}
}

// ValidateBasic does a simple validation check that doesn't require access to any other information.
func (msg MsgPostPrice) ValidateBasic() sdk.Error {
	if msg.From.Empty() {
		return sdk.ErrInternal("invalid (empty) bidder address")
	}
	if len(msg.AssetCode) == 0 {
		return sdk.ErrInternal("invalid (empty) asset code")
	}
	if msg.Price.LT(sdk.ZeroDec()) {
		return sdk.ErrInternal("invalid (negative) price")
	}
	// TODO check coin denoms
	return nil
}
