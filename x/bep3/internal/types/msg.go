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

// MsgCreateKavaHTLT is an HTLT struct, additionally containing an ID and record of updates
type MsgCreateKavaHTLT struct {
	HTLT
	ID      uint64          `json:"id"`
	Updates UpdateKavaHTLTs `json:"updates"`
}

// NewMsgCreateKavaHTLT initializes a new MsgCreateKavaHTLT
func NewMsgCreateKavaHTLT(from types.AccAddress, to types.AccAddress, recipientOtherChain, senderOtherChain string, randomNumberHash types.SwapBytes, timestamp int64,
	amount types.Coins, expectedIncome string, heightSpan int64, crossChain bool) MsgCreateKavaHTLT {
	return MsgCreateKavaHTLT{
		// No id
		// No updates
		HTLT: HTLT{
			From:                from,
			To:                  to,
			RecipientOtherChain: recipientOtherChain,
			SenderOtherChain:    senderOtherChain,
			RandomNumberHash:    randomNumberHash,
			Timestamp:           timestamp,
			Amount:              amount,
			ExpectedIncome:      expectedIncome,
			HeightSpan:          heightSpan,
			CrossChain:          crossChain},
	}
}

// Route establishes the route for the MsgCreateKavaHTLT
func (msg MsgCreateKavaHTLT) Route() string { return AtomicSwapRoute }

// Type is the name of MsgCreateKavaHTLT
func (msg MsgCreateKavaHTLT) Type() string { return "KavaHTLT" }

// String prints the MsgCreateKavaHTLT
func (msg MsgCreateKavaHTLT) String() string {
	return fmt.Sprintf("KavaHTLT{%v#%v#%v#%v#%v#%v#%v#%v#%v#%v}", msg.From, msg.To, msg.RecipientOtherChain, msg.SenderOtherChain, msg.RandomNumberHash,
		msg.Timestamp, msg.Amount, msg.ExpectedIncome, msg.HeightSpan, msg.CrossChain)
}

// GetInvolvedAddresses gets the addresses involved in a MsgCreateKavaHTLT
func (msg MsgCreateKavaHTLT) GetInvolvedAddresses() []types.AccAddress {
	return append(msg.GetSigners(), AtomicSwapCoinsAccAddr)
}

// GetSigners gets the signers of a MsgCreateKavaHTLT
func (msg MsgCreateKavaHTLT) GetSigners() []types.AccAddress {
	return []types.AccAddress{msg.From}
}

// ValidateBasic validates the MsgCreateKavaHTLT
func (msg MsgCreateKavaHTLT) ValidateBasic() error {
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

// GetSignBytes gets the sign bytes of a MsgCreateKavaHTLT
func (msg MsgCreateKavaHTLT) GetSignBytes() []byte {
	b, err := json.Marshal(msg)
	if err != nil {
		panic(err)
	}
	return b
}

// MsgDepositKavaHTLT defines an HTLT deposit
type MsgDepositKavaHTLT struct {
	UpdateKavaHTLT
	Amount types.Coins `json:"amount"`
}

// NewMsgDepositKavaHTLT initializes a new MsgDepositKavaHTLT
func NewMsgDepositKavaHTLT(from types.AccAddress, swapID []byte, amount types.Coins) MsgDepositKavaHTLT {
	return MsgDepositKavaHTLT{
		From:   from,
		SwapID: swapID,
		Amount: amount,
	}
}

// Route establishes the route for the MsgDepositKavaHTLT
func (msg MsgDepositKavaHTLT) Route() string { return AtomicSwapRoute }

// Type is the name of MsgDepositKavaHTLT
func (msg MsgDepositKavaHTLT) Type() string { return DepositHTLT }

// String prints the MsgDepositKavaHTLT
func (msg MsgDepositKavaHTLT) String() string {
	return fmt.Sprintf("depositHTLT{%v#%v#%v}", msg.From, msg.Amount, msg.SwapID)
}

// GetInvolvedAddresses gets the addresses involved in a MsgDepositKavaHTLT
func (msg MsgDepositKavaHTLT) GetInvolvedAddresses() []types.AccAddress {
	return append(msg.GetSigners(), AtomicSwapCoinsAccAddr)
}

// GetSigners gets the signers of a MsgDepositKavaHTLT
func (msg MsgDepositKavaHTLT) GetSigners() []types.AccAddress {
	return []types.AccAddress{msg.From}
}

// ValidateBasic validates the MsgDepositKavaHTLT
func (msg MsgDepositKavaHTLT) ValidateBasic() error {
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

// GetSignBytes gets the sign bytes of a MsgDepositKavaHTLT
func (msg MsgDepositKavaHTLT) GetSignBytes() []byte {
	b, err := json.Marshal(msg)
	if err != nil {
		panic(err)
	}
	return b
}

// MsgClaimKavaHTLT defines an HTLT claim
type MsgClaimKavaHTLT struct {
	UpdateKavaHTLT
	RandomNumber types.SwapBytes `json:"random_number"`
}

// NewMsgClaimKavaHTLT initializes a new MsgClaimKavaHTLT
func NewMsgClaimKavaHTLT(from types.AccAddress, swapID, randomNumber []byte) MsgClaimKavaHTLT {
	return MsgClaimKavaHTLT{
		From:         from,
		SwapID:       swapID,
		RandomNumber: randomNumber,
	}
}

// Route establishes the route for the MsgClaimKavaHTLT
func (msg MsgClaimKavaHTLT) Route() string { return AtomicSwapRoute }

// Type is the name of MsgClaimKavaHTLT
func (msg MsgClaimKavaHTLT) Type() string { return ClaimHTLT }

// String prints the MsgClaimKavaHTLT
func (msg MsgClaimKavaHTLT) String() string {
	return fmt.Sprintf("claimHTLT{%v#%v#%v}", msg.From, msg.SwapID, msg.RandomNumber)
}

// GetInvolvedAddresses gets the addresses involved in a MsgClaimKavaHTLT
func (msg MsgClaimKavaHTLT) GetInvolvedAddresses() []types.AccAddress {
	return append(msg.GetSigners(), AtomicSwapCoinsAccAddr)
}

// GetSigners gets the signers of a MsgClaimKavaHTLT
func (msg MsgClaimKavaHTLT) GetSigners() []types.AccAddress { return []types.AccAddress{msg.From} }

// ValidateBasic validates the MsgClaimKavaHTLT
func (msg MsgClaimKavaHTLT) ValidateBasic() error {
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

// GetSignBytes gets the sign bytes of a MsgClaimKavaHTLT
func (msg MsgClaimKavaHTLT) GetSignBytes() []byte {
	b, err := json.Marshal(msg)
	if err != nil {
		panic(err)
	}
	return b
}

// MsgKavaRefundHTLT defines a refund msg
type MsgKavaRefundHTLT struct {
	UpdateKavaHTLT
}

// NewMsgKavaRefundHTLT initializes a new MsgKavaRefundHTLT
func NewMsgKavaRefundHTLT(from types.AccAddress, swapID []byte) MsgKavaRefundHTLT {
	return MsgKavaRefundHTLT{
		From:   from,
		SwapID: swapID,
	}
}

// Route establishes the route for the MsgKavaRefundHTLT
func (msg MsgKavaRefundHTLT) Route() string { return AtomicSwapRoute }

// Type is the name of MsgKavaRefundHTLT
func (msg MsgKavaRefundHTLT) Type() string { return RefundHTLT }

// String prints the MsgKavaRefundHTLT
func (msg MsgKavaRefundHTLT) String() string {
	return fmt.Sprintf("refundHTLT{%v#%v}", msg.From, msg.SwapID)
}

// GetInvolvedAddresses gets the addresses involved in a MsgKavaRefundHTLT
func (msg MsgKavaRefundHTLT) GetInvolvedAddresses() []types.AccAddress {
	return append(msg.GetSigners(), AtomicSwapCoinsAccAddr)
}

// GetSigners gets the signers of a MsgKavaRefundHTLT
func (msg MsgKavaRefundHTLT) GetSigners() []types.AccAddress { return []types.AccAddress{msg.From} }

// ValidateBasic validates the MsgKavaRefundHTLT
func (msg MsgKavaRefundHTLT) ValidateBasic() error {
	if len(msg.From) != types.AddrLen {
		return fmt.Errorf("the expected address length is %d, actual length is %d", types.AddrLen, len(msg.From))
	}
	if len(msg.SwapID) != SwapIDLength {
		return fmt.Errorf("the length of swapID should be %d", SwapIDLength)
	}
	return nil
}

// GetSignBytes gets the sign bytes of a MsgKavaRefundHTLT
func (msg MsgKavaRefundHTLT) GetSignBytes() []byte {
	b, err := json.Marshal(msg)
	if err != nil {
		panic(err)
	}
	return b
}
