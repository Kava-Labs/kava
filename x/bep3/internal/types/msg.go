package types

import (
	"encoding/json"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	// "github.com/binance-chain/go-sdk/common/types"
	"github.com/tendermint/tendermint/crypto"
)

const (
	AtomicSwapRoute = "atomicSwap"
	HTLT            = "HTLT"
	DepositHTLT     = "depositHTLT"
	ClaimHTLT       = "claimHTLT"
	RefundHTLT      = "refundHTLT"

	Int64Size               = 8
	RandomNumberHashLength  = 32
	RandomNumberLength      = 32
	MaxOtherChainAddrLength = 64
	SwapIDLength            = 32
	MaxExpectedIncomeLength = 64
	MinimumHeightSpan       = 360
	MaximumHeightSpan       = 518400
)

// TODO: change this...
var (
	// bnb prefix address:  bnb1wxeplyw7x8aahy93w96yhwm7xcq3ke4f8ge93u
	// tbnb prefix address: tbnb1wxeplyw7x8aahy93w96yhwm7xcq3ke4ffasp3d
	AtomicSwapCoinsAccAddr = types.AccAddress(crypto.AddressHash([]byte("BinanceChainAtomicSwapCoins")))
)

var (
	_ sdk.Msg = &HTLTMsg{}
	_ sdk.Msg = &DepositHTLTMsg{}
	_ sdk.Msg = &ClaimHTLTMsg{}
	_ sdk.Msg = &RefundHTLTMsg{}
)


type HTLTMsg struct {
	From                sdk.AccAddress `json:"from"`
	To                  sdk.AccAddress `json:"to"`
	RecipientOtherChain string           `json:"recipient_other_chain"`
	SenderOtherChain    string           `json:"sender_other_chain"`
	RandomNumberHash    types.SwapBytes  `json:"random_number_hash"`
	Timestamp           int64            `json:"timestamp"`
	Amount              sdk.Coins      `json:"amount"`
	ExpectedIncome      string           `json:"expected_income"`
	HeightSpan          int64            `json:"height_span"`
	CrossChain          bool             `json:"cross_chain"`
}

func NewHTLTMsg(from, to types.AccAddress, recipientOtherChain, senderOtherChain string, randomNumberHash types.SwapBytes, timestamp int64,
	amount types.Coins, expectedIncome string, heightSpan int64, crossChain bool) HTLTMsg {
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

func (msg HTLTMsg) Route() string { return AtomicSwapRoute }
func (msg HTLTMsg) Type() string  { return HTLT }
func (msg HTLTMsg) String() string {
	return fmt.Sprintf("HTLT{%v#%v#%v#%v#%v#%v#%v#%v#%v#%v}", msg.From, msg.To, msg.RecipientOtherChain, msg.SenderOtherChain, msg.RandomNumberHash,
		msg.Timestamp, msg.Amount, msg.ExpectedIncome, msg.HeightSpan, msg.CrossChain)
}
func (msg HTLTMsg) GetInvolvedAddresses() []types.AccAddress {
	return append(msg.GetSigners(), AtomicSwapCoinsAccAddr)
}
func (msg HTLTMsg) GetSigners() []types.AccAddress {
	return []types.AccAddress{msg.From}
}

func (msg HTLTMsg) ValidateBasic() error {
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

func (msg HTLTMsg) GetSignBytes() []byte {
	b, err := json.Marshal(msg)
	if err != nil {
		panic(err)
	}
	return b
}

type DepositHTLTMsg struct {
	From   types.AccAddress `json:"from"`
	Amount types.Coins      `json:"amount"`
	SwapID types.SwapBytes  `json:"swap_id"`
}

func NewDepositHTLTMsg(from types.AccAddress, swapID []byte, amount types.Coins) DepositHTLTMsg {
	return DepositHTLTMsg{
		From:   from,
		SwapID: swapID,
		Amount: amount,
	}
}

func (msg DepositHTLTMsg) Route() string { return AtomicSwapRoute }
func (msg DepositHTLTMsg) Type() string  { return DepositHTLT }
func (msg DepositHTLTMsg) String() string {
	return fmt.Sprintf("depositHTLT{%v#%v#%v}", msg.From, msg.Amount, msg.SwapID)
}
func (msg DepositHTLTMsg) GetInvolvedAddresses() []types.AccAddress {
	return append(msg.GetSigners(), AtomicSwapCoinsAccAddr)
}
func (msg DepositHTLTMsg) GetSigners() []types.AccAddress {
	return []types.AccAddress{msg.From}
}

func (msg DepositHTLTMsg) ValidateBasic() error {
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

func (msg DepositHTLTMsg) GetSignBytes() []byte {
	b, err := json.Marshal(msg)
	if err != nil {
		panic(err)
	}
	return b
}

type ClaimHTLTMsg struct {
	From         types.AccAddress `json:"from"`
	SwapID       types.SwapBytes  `json:"swap_id"`
	RandomNumber types.SwapBytes  `json:"random_number"`
}

func NewClaimHTLTMsg(from types.AccAddress, swapID, randomNumber []byte) ClaimHTLTMsg {
	return ClaimHTLTMsg{
		From:         from,
		SwapID:       swapID,
		RandomNumber: randomNumber,
	}
}

func (msg ClaimHTLTMsg) Route() string { return AtomicSwapRoute }
func (msg ClaimHTLTMsg) Type() string  { return ClaimHTLT }
func (msg ClaimHTLTMsg) String() string {
	return fmt.Sprintf("claimHTLT{%v#%v#%v}", msg.From, msg.SwapID, msg.RandomNumber)
}
func (msg ClaimHTLTMsg) GetInvolvedAddresses() []types.AccAddress {
	return append(msg.GetSigners(), AtomicSwapCoinsAccAddr)
}
func (msg ClaimHTLTMsg) GetSigners() []types.AccAddress { return []types.AccAddress{msg.From} }

func (msg ClaimHTLTMsg) ValidateBasic() error {
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

func (msg ClaimHTLTMsg) GetSignBytes() []byte {
	b, err := json.Marshal(msg)
	if err != nil {
		panic(err)
	}
	return b
}

type RefundHTLTMsg struct {
	From   types.AccAddress `json:"from"`
	SwapID types.SwapBytes  `json:"swap_id"`
}

func NewRefundHTLTMsg(from types.AccAddress, swapID []byte) RefundHTLTMsg {
	return RefundHTLTMsg{
		From:   from,
		SwapID: swapID,
	}
}

func (msg RefundHTLTMsg) Route() string { return AtomicSwapRoute }
func (msg RefundHTLTMsg) Type() string  { return RefundHTLT }
func (msg RefundHTLTMsg) String() string {
	return fmt.Sprintf("refundHTLT{%v#%v}", msg.From, msg.SwapID)
}
func (msg RefundHTLTMsg) GetInvolvedAddresses() []types.AccAddress {
	return append(msg.GetSigners(), AtomicSwapCoinsAccAddr)
}
func (msg RefundHTLTMsg) GetSigners() []types.AccAddress { return []types.AccAddress{msg.From} }

func (msg RefundHTLTMsg) ValidateBasic() error {
	if len(msg.From) != types.AddrLen {
		return fmt.Errorf("the expected address length is %d, actual length is %d", types.AddrLen, len(msg.From))
	}
	if len(msg.SwapID) != SwapIDLength {
		return fmt.Errorf("the length of swapID should be %d", SwapIDLength)
	}
	return nil
}

func (msg RefundHTLTMsg) GetSignBytes() []byte {
	b, err := json.Marshal(msg)
	if err != nil {
		panic(err)
	}
	return b
}
