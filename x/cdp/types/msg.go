package types

import (
	"errors"
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// ensure Msg interface compliance at compile time
var (
	_ sdk.Msg = &MsgCreateCDP{}
	_ sdk.Msg = &MsgDeposit{}
	_ sdk.Msg = &MsgWithdraw{}
	_ sdk.Msg = &MsgDrawDebt{}
	_ sdk.Msg = &MsgRepayDebt{}
	_ sdk.Msg = &MsgLiquidate{}
)

// MsgCreateCDP creates a cdp
type MsgCreateCDP struct {
	Sender         sdk.AccAddress `json:"sender" yaml:"sender"`
	Collateral     sdk.Coin       `json:"collateral" yaml:"collateral"`
	Principal      sdk.Coin       `json:"principal" yaml:"principal"`
	CollateralType string         `json:"collateral_type" yaml:"collateral_type"`
}

// NewMsgCreateCDP returns a new MsgPlaceBid.
func NewMsgCreateCDP(sender sdk.AccAddress, collateral sdk.Coin, principal sdk.Coin, collateralType string) MsgCreateCDP {
	return MsgCreateCDP{
		Sender:         sender,
		Collateral:     collateral,
		Principal:      principal,
		CollateralType: collateralType,
	}
}

// Route return the message type used for routing the message.
func (msg MsgCreateCDP) Route() string { return RouterKey }

// Type returns a human-readable string for the message, intended for utilization within tags.
func (msg MsgCreateCDP) Type() string { return "create_cdp" }

// ValidateBasic does a simple validation check that doesn't require access to any other information.
func (msg MsgCreateCDP) ValidateBasic() error {
	if msg.Sender.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "sender address cannot be empty")
	}
	if msg.Collateral.IsZero() || !msg.Collateral.IsValid() {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidCoins, "collateral amount %s", msg.Collateral)
	}
	if msg.Principal.IsZero() || !msg.Principal.IsValid() {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidCoins, "principal amount %s", msg.Principal)
	}
	if strings.TrimSpace(msg.CollateralType) == "" {
		return fmt.Errorf("collateral type cannot be empty")
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
	Collateral Type: %s
`, msg.Sender, msg.Collateral, msg.Principal, msg.CollateralType)
}

// MsgDeposit deposit collateral to an existing cdp.
type MsgDeposit struct {
	Depositor      sdk.AccAddress `json:"depositor" yaml:"depositor"`
	Owner          sdk.AccAddress `json:"owner" yaml:"owner"`
	Collateral     sdk.Coin       `json:"collateral" yaml:"collateral"`
	CollateralType string         `json:"collateral_type" yaml:"collateral_type"`
}

// NewMsgDeposit returns a new MsgDeposit
func NewMsgDeposit(owner sdk.AccAddress, depositor sdk.AccAddress, collateral sdk.Coin, collateralType string) MsgDeposit {
	return MsgDeposit{
		Owner:          owner,
		Depositor:      depositor,
		Collateral:     collateral,
		CollateralType: collateralType,
	}
}

// Route return the message type used for routing the message.
func (msg MsgDeposit) Route() string { return RouterKey }

// Type returns a human-readable string for the message, intended for utilization within tags.
func (msg MsgDeposit) Type() string { return "deposit_cdp" }

// ValidateBasic does a simple validation check that doesn't require access to any other information.
func (msg MsgDeposit) ValidateBasic() error {
	if msg.Owner.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "owner address cannot be empty")
	}
	if msg.Depositor.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "sender address cannot be empty")
	}
	if !msg.Collateral.IsValid() || msg.Collateral.IsZero() {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidCoins, "collateral amount %s", msg.Collateral)
	}
	if strings.TrimSpace(msg.CollateralType) == "" {
		return fmt.Errorf("collateral type cannot be empty")
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
	CollateralType: %s
`, msg.Owner, msg.Owner, msg.Collateral, msg.CollateralType)
}

// MsgWithdraw withdraw collateral from an existing cdp.
type MsgWithdraw struct {
	Depositor      sdk.AccAddress `json:"depositor" yaml:"depositor"`
	Owner          sdk.AccAddress `json:"owner" yaml:"owner"`
	Collateral     sdk.Coin       `json:"collateral" yaml:"collateral"`
	CollateralType string         `json:"collateral_type" yaml:"collateral_type"`
}

// NewMsgWithdraw returns a new MsgDeposit
func NewMsgWithdraw(owner sdk.AccAddress, depositor sdk.AccAddress, collateral sdk.Coin, collateralType string) MsgWithdraw {
	return MsgWithdraw{
		Owner:          owner,
		Depositor:      depositor,
		Collateral:     collateral,
		CollateralType: collateralType,
	}
}

// Route return the message type used for routing the message.
func (msg MsgWithdraw) Route() string { return RouterKey }

// Type returns a human-readable string for the message, intended for utilization within tags.
func (msg MsgWithdraw) Type() string { return "withdraw_cdp" }

// ValidateBasic does a simple validation check that doesn't require access to any other information.
func (msg MsgWithdraw) ValidateBasic() error {
	if msg.Owner.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "owner address cannot be empty")
	}
	if msg.Depositor.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "sender address cannot be empty")
	}
	if !msg.Collateral.IsValid() || msg.Collateral.IsZero() {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidCoins, "collateral amount %s", msg.Collateral)
	}
	if strings.TrimSpace(msg.CollateralType) == "" {
		return fmt.Errorf("collateral type cannot be empty")
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

// MsgDrawDebt draw debt off of collateral in cdp
type MsgDrawDebt struct {
	Sender         sdk.AccAddress `json:"sender" yaml:"sender"`
	CollateralType string         `json:"collateral_type" yaml:"collateral_type"`
	Principal      sdk.Coin       `json:"principal" yaml:"principal"`
}

// NewMsgDrawDebt returns a new MsgDrawDebt
func NewMsgDrawDebt(sender sdk.AccAddress, collateralType string, principal sdk.Coin) MsgDrawDebt {
	return MsgDrawDebt{
		Sender:         sender,
		CollateralType: collateralType,
		Principal:      principal,
	}
}

// Route return the message type used for routing the message.
func (msg MsgDrawDebt) Route() string { return RouterKey }

// Type returns a human-readable string for the message, intended for utilization within tags.
func (msg MsgDrawDebt) Type() string { return "draw_cdp" }

// ValidateBasic does a simple validation check that doesn't require access to any other information.
func (msg MsgDrawDebt) ValidateBasic() error {
	if msg.Sender.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "sender address cannot be empty")
	}
	if strings.TrimSpace(msg.CollateralType) == "" {
		return errors.New("cdp collateral type cannot be blank")
	}
	if msg.Principal.IsZero() || !msg.Principal.IsValid() {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidCoins, "principal amount %s", msg.Principal)
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
	Collateral Type: %s
	Principal: %s
`, msg.Sender, msg.CollateralType, msg.Principal)
}

// MsgRepayDebt repay debt drawn off the collateral in a CDP
type MsgRepayDebt struct {
	Sender         sdk.AccAddress `json:"sender" yaml:"sender"`
	CollateralType string         `json:"collateral_type" yaml:"collateral_type"`
	Payment        sdk.Coin       `json:"payment" yaml:"payment"`
}

// NewMsgRepayDebt returns a new MsgRepayDebt
func NewMsgRepayDebt(sender sdk.AccAddress, collateralType string, payment sdk.Coin) MsgRepayDebt {
	return MsgRepayDebt{
		Sender:         sender,
		CollateralType: collateralType,
		Payment:        payment,
	}
}

// Route return the message type used for routing the message.
func (msg MsgRepayDebt) Route() string { return RouterKey }

// Type returns a human-readable string for the message, intended for utilization within tags.
func (msg MsgRepayDebt) Type() string { return "repay_cdp" }

// ValidateBasic does a simple validation check that doesn't require access to any other information.
func (msg MsgRepayDebt) ValidateBasic() error {
	if msg.Sender.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "sender address cannot be empty")
	}
	if strings.TrimSpace(msg.CollateralType) == "" {
		return errors.New("cdp collateral type cannot be blank")
	}
	if msg.Payment.IsZero() || !msg.Payment.IsValid() {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidCoins, "payment amount %s", msg.Payment)
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
	Collateral Type: %s
	Payment: %s
`, msg.Sender, msg.CollateralType, msg.Payment)
}

// MsgLiquidate attempts to liquidate a borrower's cdp
type MsgLiquidate struct {
	Keeper         sdk.AccAddress `json:"keeper" yaml:"keeper"`
	Borrower       sdk.AccAddress `json:"borrower" yaml:"borrower"`
	CollateralType string         `json:"collateral_type" yaml:"collateral_type"`
}

// NewMsgLiquidate returns a new MsgLiquidate
func NewMsgLiquidate(keeper, borrower sdk.AccAddress, ctype string) MsgLiquidate {
	return MsgLiquidate{
		Keeper:         keeper,
		Borrower:       borrower,
		CollateralType: ctype,
	}
}

// Route return the message type used for routing the message.
func (msg MsgLiquidate) Route() string { return RouterKey }

// Type returns a human-readable string for the message, intended for utilization within tags.
func (msg MsgLiquidate) Type() string { return "liquidate" }

// ValidateBasic does a simple validation check that doesn't require access to any other information.
func (msg MsgLiquidate) ValidateBasic() error {
	if msg.Keeper.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "keeper address cannot be empty")
	}
	if msg.Borrower.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "borrower address cannot be empty")
	}
	if strings.TrimSpace(msg.CollateralType) == "" {
		return sdkerrors.Wrap(ErrInvalidCollateral, "collateral type cannot be empty")
	}
	return nil
}

// GetSignBytes gets the canonical byte representation of the Msg.
func (msg MsgLiquidate) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// GetSigners returns the addresses of signers that must sign.
func (msg MsgLiquidate) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Keeper}
}

// String implements the Stringer interface
func (msg MsgLiquidate) String() string {
	return fmt.Sprintf(`Liquidate Message:
	Keeper:           %s
	Borrower:         %s
	Collateral Type %s
`, msg.Keeper, msg.Borrower, msg.CollateralType)
}
