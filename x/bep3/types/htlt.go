package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// HTLT contains the base of an HTLT as implemented by BinanceChain
type HTLT struct {
	From                sdk.AccAddress `json:"from"`
	To                  sdk.AccAddress `json:"to"`
	RecipientOtherChain string         `json:"recipient_other_chain"`
	SenderOtherChain    string         `json:"sender_other_chain"`
	RandomNumberHash    SwapBytes      `json:"random_number_hash"`
	Timestamp           int64          `json:"timestamp"`
	Amount              sdk.Coins      `json:"amount"`
	ExpectedIncome      string         `json:"expected_income"`
	HeightSpan          int64          `json:"height_span"`
	CrossChain          bool           `json:"cross_chain"`
}

// NewHTLT returns a new HTLT.
func NewHTLT(from sdk.AccAddress, to sdk.AccAddress, recipientOtherChain,
	senderOtherChain string, randomNumberHash SwapBytes, timestamp int64,
	amount sdk.Coins, expectedIncome string, heightSpan int64, crossChain bool) HTLT {
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
// func (h HTLT) GetModuleAccountCoins() sdk.Coins {
// }
