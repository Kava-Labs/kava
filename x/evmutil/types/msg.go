package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/auth/legacy/legacytx"
	"github.com/ethereum/go-ethereum/common"
)

// ensure Msg interface compliance at compile time
var (
	_ sdk.Msg            = &MsgConvertCoinToERC20{}
	_ sdk.Msg            = &MsgConvertERC20ToCoin{}
	_ legacytx.LegacyMsg = &MsgConvertCoinToERC20{}
	_ legacytx.LegacyMsg = &MsgConvertERC20ToCoin{}
)

// legacy message types
const (
	TypeMsgConvertCoinToERC20 = "evmutil_convert_coin_to_erc20"
	TypeMsgConvertERC20ToCoin = "evmutil_convert_erc20_to_coin"
)

// NewMsgConvertCoinToERC20 returns a new MsgConvertCoinToERC20
func NewMsgConvertCoinToERC20(
	initiator string,
	receiver string,
	amount sdk.Coin,
) MsgConvertCoinToERC20 {
	return MsgConvertCoinToERC20{
		Initiator: initiator,
		Receiver:  receiver,
		Amount:    &amount,
	}
}

// GetSigners returns the addresses of signers that must sign.
func (msg MsgConvertCoinToERC20) GetSigners() []sdk.AccAddress {
	sender, err := sdk.AccAddressFromBech32(msg.Initiator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{sender}
}

// ValidateBasic does a simple validation check that doesn't require access to any other information.
func (msg MsgConvertCoinToERC20) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Initiator)
	if err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, err.Error())
	}

	if !common.IsHexAddress(msg.Receiver) {
		return sdkerrors.Wrap(
			sdkerrors.ErrInvalidAddress,
			"Receiver is not a valid hex address",
		)
	}

	if msg.Amount.IsZero() {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "amount cannot be zero")
	}

	// Checks for negative
	return msg.Amount.Validate()
}

// GetSignBytes implements the LegacyMsg.GetSignBytes method.
func (msg MsgConvertCoinToERC20) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

// Route implements the LegacyMsg.Route method.
func (msg MsgConvertCoinToERC20) Route() string {
	return RouterKey
}

// Type implements the LegacyMsg.Type method.
func (msg MsgConvertCoinToERC20) Type() string {
	return TypeMsgConvertCoinToERC20
}

// NewMsgConvertERC20ToCoin returns a new MsgConvertERC20ToCoin
func NewMsgConvertERC20ToCoin(
	initiator InternalEVMAddress,
	receiver sdk.AccAddress,
	contractAddr InternalEVMAddress,
	amount sdk.Int,
) MsgConvertERC20ToCoin {
	return MsgConvertERC20ToCoin{
		Initiator:        initiator.String(),
		Receiver:         receiver.String(),
		KavaERC20Address: contractAddr.String(),
		Amount:           amount,
	}
}

// GetSigners returns the addresses of signers that must sign.
func (msg MsgConvertERC20ToCoin) GetSigners() []sdk.AccAddress {
	addr := common.HexToAddress(msg.Initiator)
	sender := sdk.AccAddress(addr.Bytes())
	return []sdk.AccAddress{sender}
}

// ValidateBasic does a simple validation check that doesn't require access to any other information.
func (msg MsgConvertERC20ToCoin) ValidateBasic() error {
	if !common.IsHexAddress(msg.Initiator) {
		return sdkerrors.Wrap(
			sdkerrors.ErrInvalidAddress,
			"initiator is not a valid hex address",
		)
	}

	if !common.IsHexAddress(msg.KavaERC20Address) {
		return sdkerrors.Wrap(
			sdkerrors.ErrInvalidAddress,
			"erc20 contract address is not a valid hex address",
		)
	}

	_, err := sdk.AccAddressFromBech32(msg.Receiver)
	if err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "receiver is not a valid bech32 address")
	}

	if msg.Amount.IsNil() || msg.Amount.LTE(sdk.ZeroInt()) {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "amount cannot be zero or less")
	}

	return nil
}

// GetSignBytes implements the LegacyMsg.GetSignBytes method.
func (msg MsgConvertERC20ToCoin) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

// Route implements the LegacyMsg.Route method.
func (msg MsgConvertERC20ToCoin) Route() string {
	return RouterKey
}

// Type implements the LegacyMsg.Type method.
func (msg MsgConvertERC20ToCoin) Type() string {
	return TypeMsgConvertERC20ToCoin
}
