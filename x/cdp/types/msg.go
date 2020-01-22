package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// ensure Msg interface compliance at compile time
var (
	_ sdk.Msg = &MsgCreateCDP{}
	_ sdk.Msg = &MsgDeposit{}
	_ sdk.Msg = &MsgWithdraw{}
	_ sdk.Msg = &MsgDrawDebt{}
	_ sdk.Msg = &MsgRepayDebt{}
)

// MsgCreateCDP creates a cdp
type MsgCreateCDP struct {
	Sender     sdk.AccAddress `json:"sender" yaml:"sender"`
	Collateral sdk.Coins      `json:"collateral" yaml:"collateral"`
	Principal  sdk.Coins      `json:"principal" yaml:"principal"`
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
	if len(msg.Collateral) != 1 {
		return sdk.ErrInvalidCoins(fmt.Sprintf("cdps do not support multiple collateral types: received %s", msg.Collateral))
	}
	if !msg.Collateral.IsValid() {
		return sdk.ErrInvalidCoins(msg.Collateral.String())
	}
	if !msg.Collateral.IsAllPositive() {
		return sdk.ErrInvalidCoins(msg.Collateral.String())
	}
	if !msg.Principal.IsValid() {
		return sdk.ErrInvalidCoins(msg.Principal.String())
	}
	if !msg.Principal.IsAllPositive() {
		return sdk.ErrInvalidCoins(msg.Collateral.String())
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
	Depositor  sdk.AccAddress `json:"depositor" yaml:"depositor"`
	Owner      sdk.AccAddress `json:"owner" yaml:"owner"`
	Collateral sdk.Coins      `json:"collateral" yaml:"collateral"`
}

// NewMsgDeposit returns a new MsgDeposit
func NewMsgDeposit(owner sdk.AccAddress, depositor sdk.AccAddress, collateral sdk.Coins) MsgDeposit {
	return MsgDeposit{
		Owner:      owner,
		Depositor:  depositor,
		Collateral: collateral,
	}
}

// Route return the message type used for routing the message.
func (msg MsgDeposit) Route() string { return RouterKey }

// Type returns a human-readable string for the message, intended for utilization within tags.
func (msg MsgDeposit) Type() string { return "deposit_cdp" }

// ValidateBasic does a simple validation check that doesn't require access to any other information.
func (msg MsgDeposit) ValidateBasic() sdk.Error {
	if msg.Owner.Empty() {
		return sdk.ErrInternal("invalid (empty) sender address")
	}
	if msg.Depositor.Empty() {
		return sdk.ErrInternal("invalid (empty) owner address")
	}
	if len(msg.Collateral) != 1 {
		return sdk.ErrInvalidCoins(fmt.Sprintf("cdps do not support multiple collateral types: received %s", msg.Collateral))
	}
	if !msg.Collateral.IsValid() {
		return sdk.ErrInvalidCoins(msg.Collateral.String())
	}
	if !msg.Collateral.IsAllPositive() {
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
	return []sdk.AccAddress{msg.Depositor}
}

// String implements the Stringer interface
func (msg MsgDeposit) String() string {
	return fmt.Sprintf(`Deposit to CDP Message:
	Sender:         %s
	Owner: %s
	Collateral: %s
`, msg.Owner, msg.Owner, msg.Collateral)
}

// MsgWithdraw withdraw collateral from an existing cdp.
type MsgWithdraw struct {
	Depositor  sdk.AccAddress `json:"depositor" yaml:"depositor"`
	Owner      sdk.AccAddress `json:"owner" yaml:"owner"`
	Collateral sdk.Coins      `json:"collateral" yaml:"collateral"`
}

// NewMsgWithdraw returns a new MsgDeposit
func NewMsgWithdraw(owner sdk.AccAddress, depositor sdk.AccAddress, collateral sdk.Coins) MsgWithdraw {
	return MsgWithdraw{
		Owner:      owner,
		Depositor:  depositor,
		Collateral: collateral,
	}
}

// Route return the message type used for routing the message.
func (msg MsgWithdraw) Route() string { return RouterKey }

// Type returns a human-readable string for the message, intended for utilization within tags.
func (msg MsgWithdraw) Type() string { return "withdraw_cdp" }

// ValidateBasic does a simple validation check that doesn't require access to any other information.
func (msg MsgWithdraw) ValidateBasic() sdk.Error {
	if msg.Owner.Empty() {
		return sdk.ErrInternal("invalid (empty) sender address")
	}
	if msg.Depositor.Empty() {
		return sdk.ErrInternal("invalid (empty) owner address")
	}
	if len(msg.Collateral) != 1 {
		return sdk.ErrInvalidCoins(fmt.Sprintf("cdps do not support multiple collateral types: received %s", msg.Collateral))
	}
	if !msg.Collateral.IsValid() {
		return sdk.ErrInvalidCoins(msg.Collateral.String())
	}
	if !msg.Collateral.IsAllPositive() {
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
	return []sdk.AccAddress{msg.Depositor}
}

// String implements the Stringer interface
func (msg MsgWithdraw) String() string {
	return fmt.Sprintf(`Withdraw from CDP Message:
	Owner:         %s
	Depositor: %s
	Collateral: %s
`, msg.Owner, msg.Depositor, msg.Collateral)
}

// MsgDrawDebt draw coins off of collateral in cdp
type MsgDrawDebt struct {
	Sender    sdk.AccAddress `json:"sender" yaml:"sender"`
	CdpDenom  string         `json:"cdp_denom" yaml:"cdp_denom"`
	Principal sdk.Coins      `json:"principal" yaml:"principal"`
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
	if !msg.Principal.IsAllPositive() {
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
	Sender   sdk.AccAddress `json:"sender" yaml:"sender"`
	CdpDenom string         `json:"cdp_denom" yaml:"cdp_denom"`
	Payment  sdk.Coins      `json:"payment" yaml:"payment"`
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
	if !msg.Payment.IsAllPositive() {
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
