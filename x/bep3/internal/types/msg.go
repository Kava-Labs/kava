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

// MsgCreateKHTLT is an HTLT struct, additionally containing an ID and record of updates
type MsgCreateKHTLT struct {
	HTLT
	OriginChain string `json:"origin_chain"`
	// Updates UpdateKavaHTLTs `json:"updates"`
}

// NewMsgCreateKHTLT initializes a new MsgCreateKHTLT
func NewMsgCreateKHTLT(originChain string, from types.AccAddress, to types.AccAddress, recipientOtherChain, senderOtherChain string, randomNumberHash types.SwapBytes, timestamp int64,
	amount types.Coins, expectedIncome string, heightSpan int64, crossChain bool) MsgCreateKHTLT {
	return MsgCreateKHTLT{
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
		OriginChain: originChain,
	}
}

// Route establishes the route for the MsgCreateKHTLT
func (msg MsgCreateKHTLT) Route() string { return AtomicSwapRoute }

// Type is the name of MsgCreateKHTLT
func (msg MsgCreateKHTLT) Type() string { return "KavaHTLT" }

// String prints the MsgCreateKHTLT
func (msg MsgCreateKHTLT) String() string {
	return fmt.Sprintf("KavaHTLT{%v#%v#%v#%v#%v#%v#%v#%v#%v#%v}", msg.From, msg.To, msg.RecipientOtherChain, msg.SenderOtherChain, msg.RandomNumberHash,
		msg.Timestamp, msg.Amount, msg.ExpectedIncome, msg.HeightSpan, msg.CrossChain)
}

// GetInvolvedAddresses gets the addresses involved in a MsgCreateKHTLT
func (msg MsgCreateKHTLT) GetInvolvedAddresses() []types.AccAddress {
	return append(msg.GetSigners(), AtomicSwapCoinsAccAddr)
}

// GetSigners gets the signers of a MsgCreateKHTLT
func (msg MsgCreateKHTLT) GetSigners() []types.AccAddress {
	return []types.AccAddress{msg.From}
}

// ValidateBasic validates the MsgCreateKHTLT
func (msg MsgCreateKHTLT) ValidateBasic() error {
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

// GetSignBytes gets the sign bytes of a MsgCreateKHTLT
func (msg MsgCreateKHTLT) GetSignBytes() []byte {
	b, err := json.Marshal(msg)
	if err != nil {
		panic(err)
	}
	return b
}

// MsgDepositKTHLT defines an KHTLT deposit
type MsgDepositKTHLT struct {
	From   types.AccAddress `json:"from"`
	SwapID types.SwapBytes  `json:"swap_id"`
	Amount types.Coins      `json:"amount"`
}

// NewMsgDepositKTHLT initializes a new MsgDepositKTHLT
func NewMsgDepositKTHLT(from types.AccAddress, swapID []byte, amount types.Coins) MsgDepositKTHLT {
	return MsgDepositKTHLT{
		From:   from,
		SwapID: swapID,
		Amount: amount,
	}
}

// Route establishes the route for the MsgDepositKTHLT
func (msg MsgDepositKTHLT) Route() string { return AtomicSwapRoute }

// Type is the name of MsgDepositKTHLT
func (msg MsgDepositKTHLT) Type() string { return DepositHTLT }

// String prints the MsgDepositKTHLT
func (msg MsgDepositKTHLT) String() string {
	return fmt.Sprintf("depositHTLT{%v#%v#%v}", msg.From, msg.Amount, msg.SwapID)
}

// GetInvolvedAddresses gets the addresses involved in a MsgDepositKTHLT
func (msg MsgDepositKTHLT) GetInvolvedAddresses() []types.AccAddress {
	return append(msg.GetSigners(), AtomicSwapCoinsAccAddr)
}

// GetSigners gets the signers of a MsgDepositKTHLT
func (msg MsgDepositKTHLT) GetSigners() []types.AccAddress {
	return []types.AccAddress{msg.From}
}

// ValidateBasic validates the MsgDepositKTHLT
func (msg MsgDepositKTHLT) ValidateBasic() error {
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

// GetSignBytes gets the sign bytes of a MsgDepositKTHLT
func (msg MsgDepositKTHLT) GetSignBytes() []byte {
	b, err := json.Marshal(msg)
	if err != nil {
		panic(err)
	}
	return b
}

// MsgClaimKHTLT defines a KHTLT claim
type MsgClaimKHTLT struct {
	From         types.AccAddress `json:"from"`
	SwapID       types.SwapBytes  `json:"swap_id"`
	RandomNumber types.SwapBytes  `json:"random_number"`
}

// NewMsgClaimKHTLT initializes a new MsgClaimKHTLT
func NewMsgClaimKHTLT(from types.AccAddress, swapID, randomNumber []byte) MsgClaimKHTLT {
	return MsgClaimKHTLT{
		From:         from,
		SwapID:       swapID,
		RandomNumber: randomNumber,
	}
}

// Route establishes the route for the MsgClaimKHTLT
func (msg MsgClaimKHTLT) Route() string { return AtomicSwapRoute }

// Type is the name of MsgClaimKHTLT
func (msg MsgClaimKHTLT) Type() string { return ClaimHTLT }

// String prints the MsgClaimKHTLT
func (msg MsgClaimKHTLT) String() string {
	return fmt.Sprintf("claimHTLT{%v#%v#%v}", msg.From, msg.SwapID, msg.RandomNumber)
}

// GetInvolvedAddresses gets the addresses involved in a MsgClaimKHTLT
func (msg MsgClaimKHTLT) GetInvolvedAddresses() []types.AccAddress {
	return append(msg.GetSigners(), AtomicSwapCoinsAccAddr)
}

// GetSigners gets the signers of a MsgClaimKHTLT
func (msg MsgClaimKHTLT) GetSigners() []types.AccAddress { return []types.AccAddress{msg.From} }

// ValidateBasic validates the MsgClaimKHTLT
func (msg MsgClaimKHTLT) ValidateBasic() error {
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

// GetSignBytes gets the sign bytes of a MsgClaimKHTLT
func (msg MsgClaimKHTLT) GetSignBytes() []byte {
	b, err := json.Marshal(msg)
	if err != nil {
		panic(err)
	}
	return b
}

// MsgRefundKHTLT defines a refund msg
type MsgRefundKHTLT struct {
	From   types.AccAddress `json:"from"`
	SwapID types.SwapBytes  `json:"swap_id"`
}

// NewMsgRefundKHTLT initializes a new MsgRefundKHTLT
func NewMsgRefundKHTLT(from types.AccAddress, swapID []byte) MsgRefundKHTLT {
	return MsgRefundKHTLT{
		From:   from,
		SwapID: swapID,
	}
}

// Route establishes the route for the MsgRefundKHTLT
func (msg MsgRefundKHTLT) Route() string { return AtomicSwapRoute }

// Type is the name of MsgRefundKHTLT
func (msg MsgRefundKHTLT) Type() string { return RefundHTLT }

// String prints the MsgRefundKHTLT
func (msg MsgRefundKHTLT) String() string {
	return fmt.Sprintf("refundHTLT{%v#%v}", msg.From, msg.SwapID)
}

// GetInvolvedAddresses gets the addresses involved in a MsgRefundKHTLT
func (msg MsgRefundKHTLT) GetInvolvedAddresses() []types.AccAddress {
	return append(msg.GetSigners(), AtomicSwapCoinsAccAddr)
}

// GetSigners gets the signers of a MsgRefundKHTLT
func (msg MsgRefundKHTLT) GetSigners() []types.AccAddress { return []types.AccAddress{msg.From} }

// ValidateBasic validates the MsgRefundKHTLT
func (msg MsgRefundKHTLT) ValidateBasic() error {
	if len(msg.From) != types.AddrLen {
		return fmt.Errorf("the expected address length is %d, actual length is %d", types.AddrLen, len(msg.From))
	}
	if len(msg.SwapID) != SwapIDLength {
		return fmt.Errorf("the length of swapID should be %d", SwapIDLength)
	}
	return nil
}

// GetSignBytes gets the sign bytes of a MsgRefundKHTLT
func (msg MsgRefundKHTLT) GetSignBytes() []byte {
	b, err := json.Marshal(msg)
	if err != nil {
		panic(err)
	}
	return b
}
