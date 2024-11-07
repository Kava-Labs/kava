package types

import (
	errorsmod "cosmossdk.io/errors"
	txsigning "cosmossdk.io/x/tx/signing"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/auth/migrations/legacytx"
	protov1 "github.com/golang/protobuf/proto"
	protov2 "google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/protoadapt"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// ensure Msg interface compliance at compile time
var (
	_ sdk.Msg            = &MsgFundCommunityPool{}
	_ legacytx.LegacyMsg = &MsgFundCommunityPool{}
	_ sdk.Msg            = &MsgUpdateParams{}
	_ legacytx.LegacyMsg = &MsgUpdateParams{}
)

var (
	MsgFundCommunityPoolGetSigners = txsigning.CustomGetSigner{
		MsgType: protoreflect.FullName(protov1.MessageName(&MsgFundCommunityPool{})),
		Fn:      GetSignersMsgFundCommunityPool,
	}
	MsgUpdateParamsSigners = txsigning.CustomGetSigner{
		MsgType: protoreflect.FullName(protov1.MessageName(&MsgUpdateParams{})),
		Fn:      GetSignersMsgUpdateParams,
	}
)

// NewMsgFundCommunityPool returns a new MsgFundCommunityPool
func NewMsgFundCommunityPool(depositor sdk.AccAddress, amount sdk.Coins) MsgFundCommunityPool {
	return MsgFundCommunityPool{
		Depositor: depositor.String(),
		Amount:    amount,
	}
}

// Route return the message type used for routing the message.
func (msg MsgFundCommunityPool) Route() string { return ModuleName }

// Type returns a human-readable string for the message, intended for utilization within tags.
func (msg MsgFundCommunityPool) Type() string { return sdk.MsgTypeURL(&msg) }

// ValidateBasic does a simple validation check that doesn't require access to any other information.
func (msg MsgFundCommunityPool) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Depositor)
	if err != nil {
		return errorsmod.Wrap(sdkerrors.ErrInvalidAddress, err.Error())
	}

	if msg.Amount.IsAnyNil() || !msg.Amount.IsValid() || msg.Amount.IsZero() {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidCoins, "'%s'", msg.Amount)
	}

	return nil
}

// GetSignBytes gets the canonical byte representation of the Msg.
func (msg MsgFundCommunityPool) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&msg)
	return sdk.MustSortJSON(bz)
}

// GetSigners returns the addresses of signers that must sign.
func (msg MsgFundCommunityPool) GetSigners() []sdk.AccAddress {
	depositor, err := sdk.AccAddressFromBech32(msg.Depositor)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{depositor}
}

func GetSignersMsgFundCommunityPool(msg protov2.Message) ([][]byte, error) {
	msgV1 := protoadapt.MessageV1Of(msg)

	tryingTypeAnyV1, err := codectypes.NewAnyWithValue(msgV1)
	if err != nil {
		return nil, err
	}

	msgTyped := &MsgFundCommunityPool{}
	err = msgTyped.Unmarshal(tryingTypeAnyV1.Value)
	if err != nil {
		return nil, err
	}

	depositor, err := sdk.AccAddressFromBech32(msgTyped.Depositor)
	if err != nil {
		return nil, err
	}

	return [][]byte{depositor}, nil
}

// NewMsgUpdateParams returns a new MsgUpdateParams
func NewMsgUpdateParams(authority sdk.AccAddress, params Params) MsgUpdateParams {
	return MsgUpdateParams{
		Authority: authority.String(),
		Params:    params,
	}
}

// Route return the message type used for routing the message.
func (msg MsgUpdateParams) Route() string { return ModuleName }

// Type returns a human-readable string for the message, intended for utilization within tags.
func (msg MsgUpdateParams) Type() string { return sdk.MsgTypeURL(&msg) }

// ValidateBasic does a simple validation check that doesn't require access to any other information.
func (msg MsgUpdateParams) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Authority)
	if err != nil {
		return errorsmod.Wrap(sdkerrors.ErrInvalidAddress, err.Error())
	}

	if err := msg.Params.Validate(); err != nil {
		return errorsmod.Wrap(ErrInvalidParams, err.Error())
	}

	return nil
}

// GetSignBytes gets the canonical byte representation of the Msg.
func (msg MsgUpdateParams) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&msg)
	return sdk.MustSortJSON(bz)
}

// GetSigners returns the addresses of signers that must sign.
func (msg MsgUpdateParams) GetSigners() []sdk.AccAddress {
	depositor, err := sdk.AccAddressFromBech32(msg.Authority)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{depositor}
}

func GetSignersMsgUpdateParams(msg protov2.Message) ([][]byte, error) {
	msgV1 := protoadapt.MessageV1Of(msg)

	tryingTypeAnyV1, err := codectypes.NewAnyWithValue(msgV1)
	if err != nil {
		return nil, err
	}

	msgTyped := &MsgUpdateParams{}
	err = msgTyped.Unmarshal(tryingTypeAnyV1.Value)
	if err != nil {
		return nil, err
	}

	depositor, err := sdk.AccAddressFromBech32(msgTyped.Authority)
	if err != nil {
		return nil, err
	}

	return [][]byte{depositor}, nil
}
