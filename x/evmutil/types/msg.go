package types

import (
	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	txsigning "cosmossdk.io/x/tx/signing"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/auth/migrations/legacytx"
	"github.com/ethereum/go-ethereum/common"
	"google.golang.org/protobuf/protoadapt"
	"google.golang.org/protobuf/reflect/protoreflect"

	protov1 "github.com/golang/protobuf/proto" //nolint:staticcheck
	protov2 "google.golang.org/protobuf/proto"
)

// ensure Msg interface compliance at compile time
var (
	_ sdk.Msg = &MsgConvertCoinToERC20{}
	// TODO(boodyvo): do we want to switch to Msg?
	_ legacytx.LegacyMsg = &MsgConvertCoinToERC20{}
	_ sdk.Msg            = &MsgConvertERC20ToCoin{}
	_ legacytx.LegacyMsg = &MsgConvertERC20ToCoin{}

	_ sdk.Msg            = &MsgConvertCosmosCoinToERC20{}
	_ legacytx.LegacyMsg = &MsgConvertCosmosCoinToERC20{}
	_ sdk.Msg            = &MsgConvertCosmosCoinFromERC20{}
	_ legacytx.LegacyMsg = &MsgConvertCosmosCoinFromERC20{}
)

//var MsgsGetSigners = map[protoreflect.FullName]txsigning.GetSignersFunc{
//	protoreflect.FullName(protov1.MessageName(&MsgConvertCoinToERC20{})):         GetSignersMsgConvertCoinToERC20,
//	protoreflect.FullName(protov1.MessageName(&MsgConvertERC20ToCoin{})):         GetSignersMsgConvertERC20ToCoin,
//	protoreflect.FullName(protov1.MessageName(&MsgConvertCosmosCoinToERC20{})):   GetSignersMsgConvertCosmosCoinToERC20,
//	protoreflect.FullName(protov1.MessageName(&MsgConvertCosmosCoinFromERC20{})): GetSignersMsgConvertCosmosCoinFromERC20,
//}

var (
	MsgConvertCoinToERC20GetSigners = txsigning.CustomGetSigner{
		MsgType: protoreflect.FullName(protov1.MessageName(&MsgConvertCoinToERC20{})),
		Fn:      GetSignersMsgConvertCoinToERC20,
	}
	MsgConvertERC20ToCoinGetSigners = txsigning.CustomGetSigner{
		MsgType: protoreflect.FullName(protov1.MessageName(&MsgConvertERC20ToCoin{})),
		Fn:      GetSignersMsgConvertERC20ToCoin,
	}
	MsgConvertCosmosCoinToERC20GetSigners = txsigning.CustomGetSigner{
		MsgType: protoreflect.FullName(protov1.MessageName(&MsgConvertCosmosCoinToERC20{})),
		Fn:      GetSignersMsgConvertCosmosCoinToERC20,
	}
	MsgConvertCosmosCoinFromERC20GetSigners = txsigning.CustomGetSigner{
		MsgType: protoreflect.FullName(protov1.MessageName(&MsgConvertCosmosCoinFromERC20{})),
		Fn:      GetSignersMsgConvertCosmosCoinFromERC20,
	}
)

// legacy message types
const (
	TypeMsgConvertCoinToERC20 = "evmutil_convert_coin_to_erc20"
	TypeMsgConvertERC20ToCoin = "evmutil_convert_erc20_to_coin"

	TypeMsgConvertCosmosCoinToERC20   = "evmutil_convert_cosmos_coin_to_erc20"
	TypeMsgConvertCosmosCoinFromERC20 = "evmutil_convert_cosmos_coin_from_erc20"
)

////////////////////////////
// EVM-native assets -> Cosmos SDK
////////////////////////////

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

func GetSignersMsgConvertCoinToERC20(msg protov2.Message) ([][]byte, error) {
	msgV1 := protoadapt.MessageV1Of(msg)

	tryingTypeAnyV1, err := codectypes.NewAnyWithValue(msgV1)
	if err != nil {
		return nil, err
	}

	msgTyped := &MsgConvertCoinToERC20{}
	err = msgTyped.Unmarshal(tryingTypeAnyV1.Value)
	if err != nil {
		return nil, err
	}

	sender, err := sdk.AccAddressFromBech32(msgTyped.Initiator)
	if err != nil {
		return nil, err
	}

	return [][]byte{sender}, nil
}

// ValidateBasic does a simple validation check that doesn't require access to any other information.
func (msg MsgConvertCoinToERC20) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Initiator)
	if err != nil {
		return errorsmod.Wrap(sdkerrors.ErrInvalidAddress, err.Error())
	}

	if !common.IsHexAddress(msg.Receiver) {
		return errorsmod.Wrap(
			sdkerrors.ErrInvalidAddress,
			"Receiver is not a valid hex address",
		)
	}

	if msg.Amount.IsZero() {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "amount cannot be zero")
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
	amount sdkmath.Int,
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

func GetSignersMsgConvertERC20ToCoin(msg protov2.Message) ([][]byte, error) {
	msgV1 := protoadapt.MessageV1Of(msg)

	tryingTypeAnyV1, err := codectypes.NewAnyWithValue(msgV1)
	if err != nil {
		return nil, err
	}

	msgTyped := &MsgConvertERC20ToCoin{}
	err = msgTyped.Unmarshal(tryingTypeAnyV1.Value)
	if err != nil {
		return nil, err
	}

	addr := common.HexToAddress(msgTyped.Initiator)
	sender := sdk.AccAddress(addr.Bytes())

	return [][]byte{sender}, nil
}

// ValidateBasic does a simple validation check that doesn't require access to any other information.
func (msg MsgConvertERC20ToCoin) ValidateBasic() error {
	if !common.IsHexAddress(msg.Initiator) {
		return errorsmod.Wrap(
			sdkerrors.ErrInvalidAddress,
			"initiator is not a valid hex address",
		)
	}

	if !common.IsHexAddress(msg.KavaERC20Address) {
		return errorsmod.Wrap(
			sdkerrors.ErrInvalidAddress,
			"erc20 contract address is not a valid hex address",
		)
	}

	_, err := sdk.AccAddressFromBech32(msg.Receiver)
	if err != nil {
		return errorsmod.Wrap(sdkerrors.ErrInvalidAddress, "receiver is not a valid bech32 address")
	}

	if msg.Amount.IsNil() || msg.Amount.LTE(sdkmath.ZeroInt()) {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "amount cannot be zero or less")
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

////////////////////////////
// Cosmos SDK-native assets -> EVM
////////////////////////////

// NewMsgConvertCosmosCoinToERC20 returns a new MsgConvertCosmosCoinToERC20
func NewMsgConvertCosmosCoinToERC20(
	initiator string,
	receiver string,
	amount sdk.Coin,
) MsgConvertCosmosCoinToERC20 {
	return MsgConvertCosmosCoinToERC20{
		Initiator: initiator,
		Receiver:  receiver,
		Amount:    &amount,
	}
}

// GetSigners implements types.Msg
func (msg MsgConvertCosmosCoinToERC20) GetSigners() []sdk.AccAddress {
	sender, err := sdk.AccAddressFromBech32(msg.Initiator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{sender}
}

func GetSignersMsgConvertCosmosCoinToERC20(msg protov2.Message) ([][]byte, error) {
	msgV1 := protoadapt.MessageV1Of(msg)

	tryingTypeAnyV1, err := codectypes.NewAnyWithValue(msgV1)
	if err != nil {
		return nil, err
	}

	msgTyped := &MsgConvertERC20ToCoin{}
	err = msgTyped.Unmarshal(tryingTypeAnyV1.Value)
	if err != nil {
		return nil, err
	}

	sender, err := sdk.AccAddressFromBech32(msgTyped.Initiator)
	if err != nil {
		return nil, err
	}

	return [][]byte{sender}, nil
}

// ValidateBasic implements types.Msg
func (msg MsgConvertCosmosCoinToERC20) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Initiator)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid initiator address (%s): %s", msg.Initiator, err.Error())
	}

	if !common.IsHexAddress(msg.Receiver) {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "receiver is not a valid hex address (%s)", msg.Receiver)
	}

	if msg.Amount.IsNil() || !msg.Amount.IsValid() || msg.Amount.IsZero() {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidCoins, "'%s'", msg.Amount)
	}

	return nil
}

// GetSignBytes implements legacytx.LegacyMsg
func (msg MsgConvertCosmosCoinToERC20) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

// Route implements legacytx.LegacyMsg
func (MsgConvertCosmosCoinToERC20) Route() string { return RouterKey }

// Type implements legacytx.LegacyMsg
func (MsgConvertCosmosCoinToERC20) Type() string { return TypeMsgConvertCosmosCoinToERC20 }

// NewMsgConvertCosmosCoinFromERC20 returns a new MsgConvertCosmosCoinToERC20
func NewMsgConvertCosmosCoinFromERC20(
	initiator string,
	receiver string,
	amount sdk.Coin,
) MsgConvertCosmosCoinFromERC20 {
	return MsgConvertCosmosCoinFromERC20{
		Initiator: initiator,
		Receiver:  receiver,
		Amount:    &amount,
	}
}

// GetSigners implements types.Msg
func (msg MsgConvertCosmosCoinFromERC20) GetSigners() []sdk.AccAddress {
	sender0x, err := NewInternalEVMAddressFromString(msg.Initiator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{sender0x.Bytes()}
}

func GetSignersMsgConvertCosmosCoinFromERC20(msg protov2.Message) ([][]byte, error) {
	msgV1 := protoadapt.MessageV1Of(msg)

	tryingTypeAnyV1, err := codectypes.NewAnyWithValue(msgV1)
	if err != nil {
		return nil, err
	}

	msgTyped := &MsgConvertERC20ToCoin{}
	err = msgTyped.Unmarshal(tryingTypeAnyV1.Value)
	if err != nil {
		return nil, err
	}

	sender0x, err := NewInternalEVMAddressFromString(msgTyped.Initiator)
	if err != nil {
		return nil, err
	}

	return [][]byte{sender0x.Bytes()}, nil
}

// ValidateBasic implements types.Msg
func (msg MsgConvertCosmosCoinFromERC20) ValidateBasic() error {
	if !common.IsHexAddress(msg.Initiator) {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "initiator is not a valid hex address (%s)", msg.Initiator)
	}

	_, err := sdk.AccAddressFromBech32(msg.Receiver)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid receiver address (%s): %s", msg.Receiver, err.Error())
	}

	if msg.Amount.IsNil() || !msg.Amount.IsValid() || msg.Amount.IsZero() {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidCoins, "'%s'", msg.Amount)
	}

	return nil
}

// GetSignBytes implements legacytx.LegacyMsg
func (msg MsgConvertCosmosCoinFromERC20) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

// Route implements legacytx.LegacyMsg
func (MsgConvertCosmosCoinFromERC20) Route() string { return RouterKey }

// Type implements legacytx.LegacyMsg
func (MsgConvertCosmosCoinFromERC20) Type() string { return TypeMsgConvertCosmosCoinFromERC20 }
