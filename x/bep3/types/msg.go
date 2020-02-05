package types

import (
	"encoding/json"
	"fmt"

	"github.com/binance-chain/go-sdk/common/types"
	"github.com/tendermint/tendermint/crypto"
)

const (
	AtomicSwapRoute = "atomicSwap"
	DepositHTLT     = "depositKavaHTLT"
	ClaimHTLT       = "claimKavaHTLT"
	RefundHTLT      = "refundKavaHTLT"

	Int64Size               = 8
	RandomNumberHashLength  = 32
	RandomNumberLength      = 32
	MaxOtherChainAddrLength = 64
	SwapIDLength            = 32
	MaxExpectedIncomeLength = 64
	MinimumHeightSpan       = 360
	MaximumHeightSpan       = 518400
)

var (
	// bnb prefix address:  bnb1wxeplyw7x8aahy93w96yhwm7xcq3ke4f8ge93u
	// tbnb prefix address: tbnb1wxeplyw7x8aahy93w96yhwm7xcq3ke4ffasp3d
	AtomicSwapCoinsAccAddr = types.AccAddress(crypto.AddressHash([]byte("KavaAtomicSwapCoins")))
)

// MsgCreateHTLT contains an HTLT struct
type MsgCreateHTLT struct {
	From                types.AccAddress `json:"from"`
	To                  types.AccAddress `json:"to"`
	RecipientOtherChain string           `json:"recipient_other_chain"`
	SenderOtherChain    string           `json:"sender_other_chain"`
	RandomNumberHash    types.SwapBytes  `json:"random_number_hash"`
	Timestamp           int64            `json:"timestamp"`
	Amount              types.Coins      `json:"amount"`
	ExpectedIncome      string           `json:"expected_income"`
	HeightSpan          int64            `json:"height_span"`
	CrossChain          bool             `json:"cross_chain"`
}

// NewMsgCreateHTLT initializes a new MsgCreateHTLT
func NewMsgCreateHTLT(from types.AccAddress, to types.AccAddress, recipientOtherChain,
	senderOtherChain string, randomNumberHash types.SwapBytes, timestamp int64,
	amount types.Coins, expectedIncome string, heightSpan int64, crossChain bool) MsgCreateHTLT {
	return MsgCreateHTLT{
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

// Route establishes the route for the MsgCreateHTLT
func (msg MsgCreateHTLT) Route() string { return AtomicSwapRoute }

// Type is the name of MsgCreateHTLT
func (msg MsgCreateHTLT) Type() string { return "KavaHTLT" }

// String prints the MsgCreateHTLT
func (msg MsgCreateHTLT) String() string {
	return fmt.Sprintf("HTLT{%v#%v#%v#%v#%v#%v#%v#%v#%v#%v}",
		msg.From, msg.To, msg.RecipientOtherChain, msg.SenderOtherChain, msg.RandomNumberHash,
		msg.Timestamp, msg.Amount, msg.ExpectedIncome, msg.HeightSpan, msg.CrossChain)
}

// GetInvolvedAddresses gets the addresses involved in a MsgCreateHTLT
func (msg MsgCreateHTLT) GetInvolvedAddresses() []types.AccAddress {
	return append(msg.GetSigners(), AtomicSwapCoinsAccAddr)
}

// GetSigners gets the signers of a MsgCreateHTLT
func (msg MsgCreateHTLT) GetSigners() []types.AccAddress {
	return []types.AccAddress{msg.From}
}

// ValidateBasic validates the MsgCreateHTLT
func (msg MsgCreateHTLT) ValidateBasic() error {
	if len(msg.From) != types.AddrLen {
		return fmt.Errorf("the expected address length is %d, actual length is %d", types.AddrLen, len(msg.From))
	}
	if len(msg.To) != types.AddrLen {
		return fmt.Errorf("the expected address length is %d, actual length is %d", types.AddrLen, len(msg.To))
	}
	if !msg.CrossChain && len(msg.RecipientOtherChain) != 0 {
		return fmt.Errorf("must leave recipient address on other chain to empty for single chain swap")
	}
	if !msg.CrossChain && len(msg.SenderOtherChain) != 0 {
		return fmt.Errorf("must leave sender address on other chain to empty for single chain swap")
	}
	if msg.CrossChain && len(msg.RecipientOtherChain) == 0 {
		return fmt.Errorf("missing recipient address on other chain for cross chain swap")
	}
	if len(msg.RecipientOtherChain) > MaxOtherChainAddrLength {
		return fmt.Errorf("the length of recipient address on other chain should be less than %d", MaxOtherChainAddrLength)
	}
	if len(msg.SenderOtherChain) > MaxOtherChainAddrLength {
		return fmt.Errorf("the length of sender address on other chain should be less than %d", MaxOtherChainAddrLength)
	}
	if len(msg.ExpectedIncome) > MaxExpectedIncomeLength {
		return fmt.Errorf("the length of expected income should be less than %d", MaxExpectedIncomeLength)
	}
	if len(msg.RandomNumberHash) != RandomNumberHashLength {
		return fmt.Errorf("the length of random number hash should be %d", RandomNumberHashLength)
	}
	if !msg.Amount.IsPositive() {
		return fmt.Errorf("the swapped out coin must be positive")
	}
	if msg.HeightSpan < MinimumHeightSpan || msg.HeightSpan > MaximumHeightSpan {
		return fmt.Errorf("the height span should be no less than 360 and no greater than 518400")
	}
	return nil
}

// GetSignBytes gets the sign bytes of a MsgCreateHTLT
func (msg MsgCreateHTLT) GetSignBytes() []byte {
	b, err := json.Marshal(msg)
	if err != nil {
		panic(err)
	}
	return b
}

// MsgDepositHTLT defines an HTLT deposit
type MsgDepositHTLT struct {
	From   types.AccAddress `json:"from"`
	SwapID types.SwapBytes  `json:"swap_id"`
	Amount types.Coins      `json:"amount"`
}

// NewMsgDepositHTLT initializes a new MsgDepositHTLT
func NewMsgDepositHTLT(from types.AccAddress, swapID []byte, amount types.Coins) MsgDepositHTLT {
	return MsgDepositHTLT{
		From:   from,
		SwapID: swapID,
		Amount: amount,
	}
}

// Route establishes the route for the MsgDepositHTLT
func (msg MsgDepositHTLT) Route() string { return AtomicSwapRoute }

// Type is the name of MsgDepositHTLT
func (msg MsgDepositHTLT) Type() string { return DepositHTLT }

// String prints the MsgDepositHTLT
func (msg MsgDepositHTLT) String() string {
	return fmt.Sprintf("depositHTLT{%v#%v#%v}", msg.From, msg.Amount, msg.SwapID)
}

// GetInvolvedAddresses gets the addresses involved in a MsgDepositHTLT
func (msg MsgDepositHTLT) GetInvolvedAddresses() []types.AccAddress {
	return append(msg.GetSigners(), AtomicSwapCoinsAccAddr)
}

// GetSigners gets the signers of a MsgDepositHTLT
func (msg MsgDepositHTLT) GetSigners() []types.AccAddress {
	return []types.AccAddress{msg.From}
}

// ValidateBasic validates the MsgDepositHTLT
func (msg MsgDepositHTLT) ValidateBasic() error {
	if len(msg.From) != types.AddrLen {
		return fmt.Errorf("the expected address length is %d, actual length is %d", types.AddrLen, len(msg.From))
	}
	if len(msg.SwapID) != SwapIDLength {
		return fmt.Errorf("the length of swapID should be %d", SwapIDLength)
	}
	if !msg.Amount.IsPositive() {
		return fmt.Errorf("the swapped out coin must be positive")
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
	From         types.AccAddress `json:"from"`
	SwapID       types.SwapBytes  `json:"swap_id"`
	RandomNumber types.SwapBytes  `json:"random_number"`
}

// NewMsgClaimHTLT initializes a new MsgClaimHTLT
func NewMsgClaimHTLT(from types.AccAddress, swapID, randomNumber []byte) MsgClaimHTLT {
	return MsgClaimHTLT{
		From:         from,
		SwapID:       swapID,
		RandomNumber: randomNumber,
	}
}

// Route establishes the route for the MsgClaimHTLT
func (msg MsgClaimHTLT) Route() string { return AtomicSwapRoute }

// Type is the name of MsgClaimHTLT
func (msg MsgClaimHTLT) Type() string { return ClaimHTLT }

// String prints the MsgClaimHTLT
func (msg MsgClaimHTLT) String() string {
	return fmt.Sprintf("claimHTLT{%v#%v#%v}", msg.From, msg.SwapID, msg.RandomNumber)
}

// GetInvolvedAddresses gets the addresses involved in a MsgClaimHTLT
func (msg MsgClaimHTLT) GetInvolvedAddresses() []types.AccAddress {
	return append(msg.GetSigners(), AtomicSwapCoinsAccAddr)
}

// GetSigners gets the signers of a MsgClaimHTLT
func (msg MsgClaimHTLT) GetSigners() []types.AccAddress { return []types.AccAddress{msg.From} }

// ValidateBasic validates the MsgClaimHTLT
func (msg MsgClaimHTLT) ValidateBasic() error {
	if len(msg.From) != types.AddrLen {
		return fmt.Errorf("the expected address length is %d, actual length is %d", types.AddrLen, len(msg.From))
	}
	if len(msg.SwapID) != SwapIDLength {
		return fmt.Errorf("the length of swapID should be %d", SwapIDLength)
	}
	if len(msg.RandomNumber) != RandomNumberLength {
		return fmt.Errorf("the length of random number should be %d", RandomNumberLength)
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
	From   types.AccAddress `json:"from"`
	SwapID types.SwapBytes  `json:"swap_id"`
}

// NewMsgRefundHTLT initializes a new MsgRefundHTLT
func NewMsgRefundHTLT(from types.AccAddress, swapID []byte) MsgRefundHTLT {
	return MsgRefundHTLT{
		From:   from,
		SwapID: swapID,
	}
}

// Route establishes the route for the MsgRefundHTLT
func (msg MsgRefundHTLT) Route() string { return AtomicSwapRoute }

// Type is the name of MsgRefundHTLT
func (msg MsgRefundHTLT) Type() string { return RefundHTLT }

// String prints the MsgRefundHTLT
func (msg MsgRefundHTLT) String() string {
	return fmt.Sprintf("refundHTLT{%v#%v}", msg.From, msg.SwapID)
}

// GetInvolvedAddresses gets the addresses involved in a MsgRefundHTLT
func (msg MsgRefundHTLT) GetInvolvedAddresses() []types.AccAddress {
	return append(msg.GetSigners(), AtomicSwapCoinsAccAddr)
}

// GetSigners gets the signers of a MsgRefundHTLT
func (msg MsgRefundHTLT) GetSigners() []types.AccAddress { return []types.AccAddress{msg.From} }

// ValidateBasic validates the MsgRefundHTLT
func (msg MsgRefundHTLT) ValidateBasic() error {
	if len(msg.From) != types.AddrLen {
		return fmt.Errorf("the expected address length is %d, actual length is %d", types.AddrLen, len(msg.From))
	}
	if len(msg.SwapID) != SwapIDLength {
		return fmt.Errorf("the length of swapID should be %d", SwapIDLength)
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
