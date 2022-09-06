package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	// "github.com/cosmos/cosmos-sdk/x/auth/legacy/legacytx"
)

var (
	_ sdk.Msg = &MsgMintDeposit{}
	// TODO _ legacytx.LegacyMsg = &MsgMintDeposit{}
	_ sdk.Msg = &MsgDelegateMintDeposit{}
	_ sdk.Msg = &MsgWithdrawBurn{}
	_ sdk.Msg = &MsgWithdrawBurnUndelegate{}
)

// ValidateBasic does a simple validation check that doesn't require access to any other information.
func (msg MsgMintDeposit) ValidateBasic() error {
	// TODO
	return nil
}

// GetSigners returns the addresses of signers that must sign.
func (msg MsgMintDeposit) GetSigners() []sdk.AccAddress {
	depositor, _ := sdk.AccAddressFromBech32(msg.Depositor)
	return []sdk.AccAddress{depositor}
}

// ValidateBasic does a simple validation check that doesn't require access to any other information.
func (msg MsgDelegateMintDeposit) ValidateBasic() error {
	// TODO
	return nil
}

// GetSigners returns the addresses of signers that must sign.
func (msg MsgDelegateMintDeposit) GetSigners() []sdk.AccAddress {
	depositor, _ := sdk.AccAddressFromBech32(msg.Depositor)
	return []sdk.AccAddress{depositor}
}

// ValidateBasic does a simple validation check that doesn't require access to any other information.
func (msg MsgWithdrawBurn) ValidateBasic() error {
	// TODO
	return nil
}

// GetSigners returns the addresses of signers that must sign.
func (msg MsgWithdrawBurn) GetSigners() []sdk.AccAddress {
	from, _ := sdk.AccAddressFromBech32(msg.From)
	return []sdk.AccAddress{from}
}

// ValidateBasic does a simple validation check that doesn't require access to any other information.
func (msg MsgWithdrawBurnUndelegate) ValidateBasic() error {
	// TODO
	return nil
}

// GetSigners returns the addresses of signers that must sign.
func (msg MsgWithdrawBurnUndelegate) GetSigners() []sdk.AccAddress {
	from, _ := sdk.AccAddressFromBech32(msg.From)
	return []sdk.AccAddress{from}
}
