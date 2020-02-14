package types

import (
	"encoding/json"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/tendermint/crypto"
)

const (
	Htlt        = "HTLT"
	DepositHTLT = "depositHTLT"
	ClaimHTLT   = "claimHTLT"
	RefundHTLT  = "refundHTLT"
	CalcSwapID  = "calcSwapID"

	Int64Size               = 8
	RandomNumberHashLength  = 64
	RandomNumberLength      = 64
	AddrByteCount           = 20
	MaxOtherChainAddrLength = 64
	SwapIDLength            = 64
	MaxExpectedIncomeLength = 64
	MinimumHeightSpan       = 360
	MaximumHeightSpan       = 518400
)

var (
	// kava prefix address:  [INSERT BEP3-DEPUTY ADDRESS]
	// tkava prefix address: [INSERT BEP3-DEPUTY ADDRESS]
	AtomicSwapCoinsAccAddr = sdk.AccAddress(crypto.AddressHash([]byte("KavaAtomicSwapCoins")))
)

// HTLTMsg contains an HTLT struct
type HTLTMsg struct {
	From                sdk.AccAddress `json:"from"`
	To                  sdk.AccAddress `json:"to"`
	RecipientOtherChain string         `json:"recipient_other_chain"`
	SenderOtherChain    string         `json:"sender_other_chain"`
	RandomNumberHash    string         `json:"random_number_hash"`
	Timestamp           int64          `json:"timestamp"`
	Amount              sdk.Coins      `json:"amount"`
	ExpectedIncome      string         `json:"expected_income"`
	HeightSpan          int64          `json:"height_span"`
	CrossChain          bool           `json:"cross_chain"`
}

// NewHTLTMsg initializes a new HTLTMsg
func NewHTLTMsg(from sdk.AccAddress, to sdk.AccAddress, recipientOtherChain,
	senderOtherChain string, randomNumberHash string, timestamp int64,
	amount sdk.Coins, expectedIncome string, heightSpan int64, crossChain bool) HTLTMsg {
	return HTLTMsg{
		From:                from,
		To:                  to,
		RecipientOtherChain: recipientOtherChain,
		SenderOtherChain:    senderOtherChain,
		RandomNumberHash:    randomNumberHash,
		Timestamp:           timestamp,
		Amount:              amount,
		ExpectedIncome:      expectedIncome,
		HeightSpan:          heightSpan,
		CrossChain:          crossChain,
	}
}

// Route establishes the route for the HTLTMsg
func (msg HTLTMsg) Route() string { return RouterKey }

// Type is the name of HTLTMsg
func (msg HTLTMsg) Type() string { return Htlt }

// String prints the HTLTMsg
func (msg HTLTMsg) String() string {
	return fmt.Sprintf("HTLT{%v#%v#%v#%v#%v#%v#%v#%v#%v#%v}",
		msg.From, msg.To, msg.RecipientOtherChain, msg.SenderOtherChain, msg.RandomNumberHash,
		msg.Timestamp, msg.Amount, msg.ExpectedIncome, msg.HeightSpan, msg.CrossChain)
}

// GetInvolvedAddresses gets the addresses involved in a HTLTMsg
func (msg HTLTMsg) GetInvolvedAddresses() []sdk.AccAddress {
	return append(msg.GetSigners(), AtomicSwapCoinsAccAddr)
}

// GetSigners gets the signers of a HTLTMsg
func (msg HTLTMsg) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.From}
}

// ValidateBasic validates the HTLTMsg
func (msg HTLTMsg) ValidateBasic() sdk.Error {
	if len(msg.From) != AddrByteCount {
		return sdk.ErrInternal(fmt.Sprintf("the expected address length is %d, actual length is %d", AddrByteCount, len(msg.From)))
	}
	if len(msg.To) != AddrByteCount {
		return sdk.ErrInternal(fmt.Sprintf("the expected address length is %d, actual length is %d", AddrByteCount, len(msg.To)))
	}
	if !msg.CrossChain && len(msg.RecipientOtherChain) != 0 {
		return sdk.ErrInternal(fmt.Sprintf("must leave recipient address on other chain to empty for single chain swap"))
	}
	if !msg.CrossChain && len(msg.SenderOtherChain) != 0 {
		return sdk.ErrInternal(fmt.Sprintf("must leave sender address on other chain to empty for single chain swap"))
	}
	if msg.CrossChain && len(msg.RecipientOtherChain) == 0 {
		return sdk.ErrInternal(fmt.Sprintf("missing recipient address on other chain for cross chain swap"))
	}
	if len(msg.RecipientOtherChain) > MaxOtherChainAddrLength {
		return sdk.ErrInternal(fmt.Sprintf("the length of recipient address on other chain should be less than %d", MaxOtherChainAddrLength))
	}
	if len(msg.SenderOtherChain) > MaxOtherChainAddrLength {
		return sdk.ErrInternal(fmt.Sprintf("the length of sender address on other chain should be less than %d", MaxOtherChainAddrLength))
	}
	if len(msg.ExpectedIncome) > MaxExpectedIncomeLength {
		return sdk.ErrInternal(fmt.Sprintf("the length of expected income should be less than %d", MaxExpectedIncomeLength))
	}
	if len(msg.RandomNumberHash) != RandomNumberHashLength {
		return sdk.ErrInternal(fmt.Sprintf("the length of random number hash should be %d", RandomNumberHashLength))
	}
	if !msg.Amount.IsAllPositive() {
		return sdk.ErrInternal(fmt.Sprintf("the swapped out coin must be positive"))
	}
	if msg.HeightSpan < MinimumHeightSpan || msg.HeightSpan > MaximumHeightSpan {
		return sdk.ErrInternal(fmt.Sprintf("the height span should be no less than 360 and no greater than 518400"))
	}
	return nil
}

// GetSignBytes gets the sign bytes of a HTLTMsg
func (msg HTLTMsg) GetSignBytes() []byte {
	b, err := json.Marshal(msg)
	if err != nil {
		panic(err)
	}
	return b
}

// MsgDepositHTLT defines an HTLT deposit
type MsgDepositHTLT struct {
	From   sdk.AccAddress `json:"from"`
	SwapID string         `json:"swap_id"`
	Amount sdk.Coins      `json:"amount"`
}

// NewMsgDepositHTLT initializes a new MsgDepositHTLT
func NewMsgDepositHTLT(from sdk.AccAddress, swapID string, amount sdk.Coins) MsgDepositHTLT {
	return MsgDepositHTLT{
		From:   from,
		SwapID: swapID,
		Amount: amount,
	}
}

// Route establishes the route for the MsgDepositHTLT
func (msg MsgDepositHTLT) Route() string { return RouterKey }

// Type is the name of MsgDepositHTLT
func (msg MsgDepositHTLT) Type() string { return DepositHTLT }

// String prints the MsgDepositHTLT
func (msg MsgDepositHTLT) String() string {
	return fmt.Sprintf("depositHTLT{%v#%v#%v}", msg.From, msg.Amount, msg.SwapID)
}

// GetInvolvedAddresses gets the addresses involved in a MsgDepositHTLT
func (msg MsgDepositHTLT) GetInvolvedAddresses() []sdk.AccAddress {
	return append(msg.GetSigners(), AtomicSwapCoinsAccAddr)
}

// GetSigners gets the signers of a MsgDepositHTLT
func (msg MsgDepositHTLT) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.From}
}

// ValidateBasic validates the MsgDepositHTLT
func (msg MsgDepositHTLT) ValidateBasic() sdk.Error {
	if len(msg.From) != AddrByteCount {
		return sdk.ErrInternal(fmt.Sprintf("the expected address length is %d, actual length is %d", AddrByteCount, len(msg.From)))
	}
	if len(msg.SwapID) != SwapIDLength {
		return sdk.ErrInternal(fmt.Sprintf("the length of swapID should be %d", SwapIDLength))
	}
	if !msg.Amount.IsAllPositive() {
		return sdk.ErrInternal(fmt.Sprintf("the swapped out coin must be positive"))
	}
	return nil
}

// GetSignBytes gets the sign bytes of a MsgDepositHTLT
func (msg MsgDepositHTLT) GetSignBytes() []byte {
	b, err := json.Marshal(msg)
	if err != nil {
		panic(err)
	}
	return b
}

// MsgClaimHTLT defines a HTLT claim
type MsgClaimHTLT struct {
	From         sdk.AccAddress `json:"from"`
	SwapID       string         `json:"swap_id"`
	RandomNumber SwapBytes      `json:"random_number"`
}

// NewMsgClaimHTLT initializes a new MsgClaimHTLT
func NewMsgClaimHTLT(from sdk.AccAddress, swapID string, randomNumber SwapBytes) MsgClaimHTLT {
	return MsgClaimHTLT{
		From:         from,
		SwapID:       swapID,
		RandomNumber: randomNumber,
	}
}

// Route establishes the route for the MsgClaimHTLT
func (msg MsgClaimHTLT) Route() string { return RouterKey }

// Type is the name of MsgClaimHTLT
func (msg MsgClaimHTLT) Type() string { return ClaimHTLT }

// String prints the MsgClaimHTLT
func (msg MsgClaimHTLT) String() string {
	return fmt.Sprintf("claimHTLT{%v#%v#%v}", msg.From, msg.SwapID, msg.RandomNumber)
}

// GetInvolvedAddresses gets the addresses involved in a MsgClaimHTLT
func (msg MsgClaimHTLT) GetInvolvedAddresses() []sdk.AccAddress {
	return append(msg.GetSigners(), AtomicSwapCoinsAccAddr)
}

// GetSigners gets the signers of a MsgClaimHTLT
func (msg MsgClaimHTLT) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.From}
}

// ValidateBasic validates the MsgClaimHTLT
func (msg MsgClaimHTLT) ValidateBasic() sdk.Error {
	if len(msg.From) != AddrByteCount {
		return sdk.ErrInternal(fmt.Sprintf("the expected address length is %d, actual length is %d", AddrByteCount, len(msg.From)))
	}
	if len(msg.SwapID) != SwapIDLength {
		return sdk.ErrInternal(fmt.Sprintf("the length of swapID should be %d", SwapIDLength))
	}
	if len(msg.RandomNumber) == 0 {
		return sdk.ErrInternal("the length of random number cannot be 0")
	}
	return nil
}

// GetSignBytes gets the sign bytes of a MsgClaimHTLT
func (msg MsgClaimHTLT) GetSignBytes() []byte {
	b, err := json.Marshal(msg)
	if err != nil {
		panic(err)
	}
	return b
}

// MsgRefundHTLT defines a refund msg
type MsgRefundHTLT struct {
	From   sdk.AccAddress `json:"from"`
	SwapID string         `json:"swap_id"`
}

// NewMsgRefundHTLT initializes a new MsgRefundHTLT
func NewMsgRefundHTLT(from sdk.AccAddress, swapID string) MsgRefundHTLT {
	return MsgRefundHTLT{
		From:   from,
		SwapID: swapID,
	}
}

// Route establishes the route for the MsgRefundHTLT
func (msg MsgRefundHTLT) Route() string { return RouterKey }

// Type is the name of MsgRefundHTLT
func (msg MsgRefundHTLT) Type() string { return RefundHTLT }

// String prints the MsgRefundHTLT
func (msg MsgRefundHTLT) String() string {
	return fmt.Sprintf("refundHTLT{%v#%v}", msg.From, msg.SwapID)
}

// GetInvolvedAddresses gets the addresses involved in a MsgRefundHTLT
func (msg MsgRefundHTLT) GetInvolvedAddresses() []sdk.AccAddress {
	return append(msg.GetSigners(), AtomicSwapCoinsAccAddr)
}

// GetSigners gets the signers of a MsgRefundHTLT
func (msg MsgRefundHTLT) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.From}
}

// ValidateBasic validates the MsgRefundHTLT
func (msg MsgRefundHTLT) ValidateBasic() sdk.Error {
	if len(msg.From) != AddrByteCount {
		return sdk.ErrInternal(fmt.Sprintf("the expected address length is %d, actual length is %d", AddrByteCount, len(msg.From)))
	}
	if len(msg.SwapID) != SwapIDLength {
		return sdk.ErrInternal(fmt.Sprintf("the length of swapID should be %d", SwapIDLength))
	}
	return nil
}

// GetSignBytes gets the sign bytes of a MsgRefundHTLT
func (msg MsgRefundHTLT) GetSignBytes() []byte {
	b, err := json.Marshal(msg)
	if err != nil {
		panic(err)
	}
	return b
}
