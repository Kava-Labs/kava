package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// MsgCreateCDP creates a cdp
type MsgCreateCDP struct {
	Sender     sdk.AccAddress
	Collateral sdk.Coins
	Principal  sdk.Coins
}

// NewMsgCreateCDP returns a new MsgPlaceBid.
func NewMsgCreateCDP(sender sdk.AccAddress, collateral sdk.Coins, principal sdk.Coins) MsgCreateCDP {
	return MsgCreateCDP{
		Sender:     sender,
		Collateral: collateral,
		Principal:  principal,
	}
}

// Route return the message type used for routing the message.
func (msg MsgCreateCDP) Route() string { return RouterKey }

// Type returns a human-readable string for the message, intended for utilization within tags.
func (msg MsgCreateCDP) Type() string { return "create_cdp" }

// ValidateBasic does a simple validation check that doesn't require access to any other information.
func (msg MsgCreateCDP) ValidateBasic() sdk.Error {
	if msg.Sender.Empty() {
		return sdk.ErrInternal("invalid (empty) sender address")
	}
	if !msg.Collateral.IsValid() {
		return sdk.ErrInvalidCoins(msg.Collateral.String())
	}
	if !msg.Principal.IsValid() {
		return sdk.ErrInvalidCoins(msg.Principal.String())
	}
	if msg.Collateral.IsAnyNegative() {
		return sdk.ErrInvalidCoins(msg.Collateral.String())
	}
	if msg.Principal.IsAnyNegative() {
		return sdk.ErrInvalidCoins(msg.Principal.String())
	}
	return nil
}

// GetSignBytes gets the canonical byte representation of the Msg.
func (msg MsgCreateCDP) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// GetSigners returns the addresses of signers that must sign.
func (msg MsgCreateCDP) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Sender}
}

// String implements the Stringer interface
func (msg MsgCreateCDP) String() string {
	return fmt.Sprintf(`Create CDP Message:
  Sender:         %s
	Collateral: %s
	Principal: %s
`, msg.Sender, msg.Collateral, msg.Principal)
}

// MsgDeposit deposit collateral to an existing cdp.
type MsgDeposit struct {
	Sender     sdk.AccAddress
	Owner      sdk.AccAddress
	Collateral sdk.Coins
}

// NewMsgDeposit returns a new MsgDeposit
func NewMsgDeposit(sender sdk.AccAddress, owner sdk.AccAddress, collateral sdk.Coins) MsgDeposit {
	return MsgDeposit{
		Sender:     sender,
		Owner:      owner,
		Collateral: collateral,
	}
}

// Route return the message type used for routing the message.
func (msg MsgDeposit) Route() string { return RouterKey }

// Type returns a human-readable string for the message, intended for utilization within tags.
func (msg MsgDeposit) Type() string { return "deposit_cdp" }

// ValidateBasic does a simple validation check that doesn't require access to any other information.
func (msg MsgDeposit) ValidateBasic() sdk.Error {
	if msg.Sender.Empty() {
		return sdk.ErrInternal("invalid (empty) sender address")
	}
	if msg.Owner.Empty() {
		return sdk.ErrInternal("invalid (empty) owner address")
	}
	if !msg.Collateral.IsValid() {
		return sdk.ErrInvalidCoins(msg.Collateral.String())
	}
	if msg.Collateral.IsAnyNegative() {
		return sdk.ErrInvalidCoins(msg.Collateral.String())
	}
	return nil
}

// GetSignBytes gets the canonical byte representation of the Msg.
func (msg MsgDeposit) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// GetSigners returns the addresses of signers that must sign.
func (msg MsgDeposit) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Sender}
}

// String implements the Stringer interface
func (msg MsgDeposit) String() string {
	return fmt.Sprintf(`Deposit to CDP Message:
	Sender:         %s
	Owner: %s
	Collateral: %s
`, msg.Sender, msg.Owner, msg.Collateral)
}

// MsgWithdraw withdraw collateral from an existing cdp.
type MsgWithdraw struct {
	Sender     sdk.AccAddress
	Owner      sdk.AccAddress
	Collateral sdk.Coins
}

// NewMsgWithdraw returns a new MsgDeposit
func NewMsgWithdraw(sender sdk.AccAddress, owner sdk.AccAddress, collateral sdk.Coins) MsgDeposit {
	return MsgDeposit{
		Sender:     sender,
		Owner:      owner,
		Collateral: collateral,
	}
}

// Route return the message type used for routing the message.
func (msg MsgWithdraw) Route() string { return RouterKey }

// Type returns a human-readable string for the message, intended for utilization within tags.
func (msg MsgWithdraw) Type() string { return "withdraw_cdp" }

// ValidateBasic does a simple validation check that doesn't require access to any other information.
func (msg MsgWithdraw) ValidateBasic() sdk.Error {
	if msg.Sender.Empty() {
		return sdk.ErrInternal("invalid (empty) sender address")
	}
	if msg.Owner.Empty() {
		return sdk.ErrInternal("invalid (empty) owner address")
	}
	if !msg.Collateral.IsValid() {
		return sdk.ErrInvalidCoins(msg.Collateral.String())
	}
	if msg.Collateral.IsAnyNegative() {
		return sdk.ErrInvalidCoins(msg.Collateral.String())
	}
	return nil
}

// GetSignBytes gets the canonical byte representation of the Msg.
func (msg MsgWithdraw) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// GetSigners returns the addresses of signers that must sign.
func (msg MsgWithdraw) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Sender}
}

// String implements the Stringer interface
func (msg MsgWithdraw) String() string {
	return fmt.Sprintf(`Withdraw from CDP Message:
	Sender:         %s
	Owner: %s
	Collateral: %s
`, msg.Sender, msg.Owner, msg.Collateral)
}

// MsgDrawDebt draw coins off of collateral in cdp
type MsgDrawDebt struct {
	Sender    sdk.AccAddress
	CdpDenom  string
	Principal sdk.Coins
}

// NewMsgDrawDebt returns a new MsgDrawDebt
func NewMsgDrawDebt(sender sdk.AccAddress, denom string, principal sdk.Coins) MsgDrawDebt {
	return MsgDrawDebt{
		Sender:    sender,
		CdpDenom:  denom,
		Principal: principal,
	}
}

// Route return the message type used for routing the message.
func (msg MsgDrawDebt) Route() string { return RouterKey }

// Type returns a human-readable string for the message, intended for utilization within tags.
func (msg MsgDrawDebt) Type() string { return "draw_cdp" }

// ValidateBasic does a simple validation check that doesn't require access to any other information.
func (msg MsgDrawDebt) ValidateBasic() sdk.Error {
	if msg.Sender.Empty() {
		return sdk.ErrInternal("invalid (empty) sender address")
	}
	if msg.CdpDenom == "" {
		return sdk.ErrInternal("invalid (empty) cdp denom")
	}
	if !msg.Principal.IsValid() {
		return sdk.ErrInvalidCoins(msg.Principal.String())
	}
	if msg.Principal.IsAnyNegative() {
		return sdk.ErrInvalidCoins(msg.Principal.String())
	}
	return nil
}

// GetSignBytes gets the canonical byte representation of the Msg.
func (msg MsgDrawDebt) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// GetSigners returns the addresses of signers that must sign.
func (msg MsgDrawDebt) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Sender}
}

// String implements the Stringer interface
func (msg MsgDrawDebt) String() string {
	return fmt.Sprintf(`Draw debt from CDP Message:
	Sender:         %s
	CDP Denom: %s
	Principal: %s
`, msg.Sender, msg.CdpDenom, msg.Principal)
}

// MsgRepayDebt repay debt drawn off the collateral in a CDP
type MsgRepayDebt struct {
	Sender   sdk.AccAddress
	CdpDenom string
	Payment  sdk.Coins
}

// NewMsgRepayDebt returns a new MsgRepayDebt
func NewMsgRepayDebt(sender sdk.AccAddress, denom string, payment sdk.Coins) MsgRepayDebt {
	return MsgRepayDebt{
		Sender:   sender,
		CdpDenom: denom,
		Payment:  payment,
	}
}

// Route return the message type used for routing the message.
func (msg MsgRepayDebt) Route() string { return RouterKey }

// Type returns a human-readable string for the message, intended for utilization within tags.
func (msg MsgRepayDebt) Type() string { return "repay_cdp" }

// ValidateBasic does a simple validation check that doesn't require access to any other information.
func (msg MsgRepayDebt) ValidateBasic() sdk.Error {
	if msg.Sender.Empty() {
		return sdk.ErrInternal("invalid (empty) sender address")
	}
	if msg.CdpDenom == "" {
		return sdk.ErrInternal("invalid (empty) cdp denom")
	}
	if !msg.Payment.IsValid() {
		return sdk.ErrInvalidCoins(msg.Payment.String())
	}
	if msg.Payment.IsAnyNegative() {
		return sdk.ErrInvalidCoins(msg.Payment.String())
	}
	return nil
}

// GetSignBytes gets the canonical byte representation of the Msg.
func (msg MsgRepayDebt) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// GetSigners returns the addresses of signers that must sign.
func (msg MsgRepayDebt) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Sender}
}

// String implements the Stringer interface
func (msg MsgRepayDebt) String() string {
	return fmt.Sprintf(`Draw debt from CDP Message:
	Sender:         %s
	CDP Denom: %s
	Payment: %s
`, msg.Sender, msg.CdpDenom, msg.Payment)
}

// MsgTransferCDP changes the ownership of a cdp
type MsgTransferCDP struct {
	// TODO
}
