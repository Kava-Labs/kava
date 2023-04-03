package types

import (
	"errors"
	"fmt"
	"strings"
	time "time"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const (
	// TypeMsgPostPrice type of PostPrice msg
	TypeMsgPostPrice = "post_price"

	// MaxExpiry defines the max expiry time defined as UNIX time (9999-12-31 23:59:59 +0000 UTC)
	MaxExpiry = 253402300799
)

// ensure Msg interface compliance at compile time
var _ sdk.Msg = &MsgPostPrice{}

// NewMsgPostPrice returns a new MsgPostPrice
func NewMsgPostPrice(from string, marketID string, price sdk.Dec, expiry time.Time) *MsgPostPrice {
	return &MsgPostPrice{
		From:     from,
		MarketID: marketID,
		Price:    price,
		Expiry:   expiry,
	}
}

// Route Implements Msg.
func (msg MsgPostPrice) Route() string { return RouterKey }

// Type Implements Msg
func (msg MsgPostPrice) Type() string { return TypeMsgPostPrice }

// GetSignBytes Implements Msg.
func (msg MsgPostPrice) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&msg)
	return sdk.MustSortJSON(bz)
}

// GetSigners Implements Msg.
func (msg MsgPostPrice) GetSigners() []sdk.AccAddress {
	from, err := sdk.AccAddressFromBech32(msg.From)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{from}
}

// ValidateBasic does a simple validation check that doesn't require access to any other information.
func (msg MsgPostPrice) ValidateBasic() error {
	if len(msg.From) == 0 {
		return errorsmod.Wrap(sdkerrors.ErrInvalidAddress, "sender address cannot be empty")
	}
	if strings.TrimSpace(msg.MarketID) == "" {
		return errors.New("market id cannot be blank")
	}
	if msg.Price.IsNegative() {
		return fmt.Errorf("price cannot be negative: %s", msg.Price.String())
	}
	if msg.Expiry.Unix() <= 0 {
		return errors.New("must set an expiration time")
	}
	return nil
}
