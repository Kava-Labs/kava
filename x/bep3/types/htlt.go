package types

import (
	"github.com/binance-chain/go-sdk/common/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// HTLT contains the base of an HTLT as implemented by BinanceChain
type HTLT struct {
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

// NewHTLT returns a new HTLT.
func NewHTLT(from types.AccAddress, to types.AccAddress, recipientOtherChain,
	senderOtherChain string, randomNumberHash types.SwapBytes, timestamp int64,
	amount types.Coins, expectedIncome string, heightSpan int64, crossChain bool) HTLT {
	return HTLT{
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

// HTLTs is a slice of HTLT
type HTLTs []HTLT

// // TODO: Validate verifies that the module account has enough coins to
// //		 complete HTLT and time is less than max time
// func (h HTLT) Validate() error {
// 	if h.EndTime.After(h.MaxEndTime) {
// 		return fmt.Errorf("MaxEndTime < EndTime (%s < %s)", h.MaxEndTime, h.EndTime)
// 	}
// 	return nil
// }

// GetModuleAccountCoins returns the total number of coins held in the module account for this HTLT.
// It is used in genesis initialize the module account correctly.
func (h HTLT) GetModuleAccountCoins() sdk.Coins {
	// We must convert BinanceChain.Coins to Cosmos.Coins
	var coins sdk.Coins
	for _, coin := range h.Amount {
		coins = append(coins, sdk.NewCoin(coin.Denom, sdk.NewInt(coin.Amount)))
	}
	return coins
}

// ClaimKHTLT defines an HTLT claim
type ClaimKHTLT struct {
	From         types.AccAddress `json:"from"`
	SwapID       types.SwapBytes  `json:"swap_id"`
	RandomNumber types.SwapBytes  `json:"random_number"`
}

// Name returns a name for this update type. Used to identify updates in event attributes.
func (h ClaimKHTLT) Name() string { return "claim" }

// DepositKHTLT defines an HTLT deposit
type DepositKHTLT struct {
	From   types.AccAddress `json:"from"`
	SwapID types.SwapBytes  `json:"swap_id"`
	Amount types.Coins      `json:"amount"`
}

// Name returns a name for this update type. Used to identify updates in event attributes.
func (h DepositKHTLT) Name() string { return "deposit" }

// RefundKHTLT defines an HTLT refund
type RefundKHTLT struct {
	From   types.AccAddress `json:"from"`
	SwapID types.SwapBytes  `json:"swap_id"`
}

// Name returns a name for this update type. Used to identify updates in event attributes.
func (h RefundKHTLT) Name() string { return "refund" }
