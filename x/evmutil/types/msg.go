package types

import (
	"bytes"
	"errors"
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/auth/legacy/legacytx"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

// ensure Msg interface compliance at compile time
var (
	_ sdk.Msg            = &MsgConvertCoinToERC20{}
	_ sdk.Msg            = &MsgConvertERC20ToCoin{}
	_ sdk.Msg            = &MsgEVMCall{}
	_ legacytx.LegacyMsg = &MsgConvertCoinToERC20{}
	_ legacytx.LegacyMsg = &MsgConvertERC20ToCoin{}
	_ legacytx.LegacyMsg = &MsgEVMCall{}
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

// Route implements Msg
func (msg MsgEVMCall) Route() string { return types.RouterKey }

// Type implements Msg
func (msg MsgEVMCall) Type() string { return sdk.MsgTypeURL(&msg) }

// ValidateBasic implements Msg
func (msg MsgEVMCall) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Authority); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid authority address: %s", err)
	}

	if !common.IsHexAddress(msg.To) {
		return sdkerrors.ErrInvalidAddress.Wrapf("to '%s' is not hex address", msg.To)
	}

	if msg.Amount.IsNil() {
		return fmt.Errorf("amount must not be nil")
	}

	if msg.Amount.IsNegative() {
		return fmt.Errorf("amount cannot be negative: %s", msg.Amount)
	}

	// validate data & fnabi
	if len(msg.Data) > 0 {
		if len(msg.FnAbi) == 0 {
			return fmt.Errorf("fnAbi is not provided: this required when passing in data")
		}
		if _, err := msg.Decode(); err != nil {
			return err
		}
	}

	return nil
}

// GetSignBytes implements Msg
func (msg MsgEVMCall) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&msg)
	return sdk.MustSortJSON(bz)
}

// GetSigners implements Msg
func (msg MsgEVMCall) GetSigners() []sdk.AccAddress {
	authority, _ := sdk.AccAddressFromBech32(msg.Authority)
	return []sdk.AccAddress{authority}
}

func (action MsgEVMCall) Decode() ([]string, error) {
	if len(action.FnAbi) == 0 {
		return nil, errors.New("cannot decode MsgEVMCall: empty fnAbi")
	}

	d, err := abi.JSON(strings.NewReader(fmt.Sprintf("[%s]", action.FnAbi)))
	if err != nil {
		return nil, fmt.Errorf("unable to parse fn abi: %s", err)
	}
	data, err := hexutil.Decode(action.Data)
	if err != nil {
		return nil, fmt.Errorf("invalid data format: %s", err)
	}

	if len(d.Methods) != 1 {
		return nil, fmt.Errorf("failed to parse a single method from fnAbi: methods parsed %d", len(d.Methods))
	}

	method, err := d.MethodById(data[:4])
	if err != nil {
		return nil, fmt.Errorf("method not found in fnAbi: %s", hexutil.Encode(data[:4]))
	}

	// attempt to unpack params data by removing the first bytes of the method signature
	val, err := method.Inputs.Unpack(data[4:])
	if err != nil {
		return nil, fmt.Errorf("unable to decode method args: %s", err)
	}

	packedData, err := method.Inputs.Pack(val...)
	if err != nil {
		return nil, fmt.Errorf("unable to pack decoded data: %s", err)
	}

	// verify call data is the same as unpacked data
	if !bytes.Equal(packedData, data[4:]) {
		return nil, fmt.Errorf("invalid call data: call data does not match unpacked data")
	}

	strVals := make([]string, len(val))
	for i, v := range val {
		strVals[i] = fmt.Sprintf("%s", v)
	}

	return strVals, nil
}
