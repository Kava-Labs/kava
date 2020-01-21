package types

import (
	"github.com/binance-chain/go-sdk/common/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// HTLT contains the base of an HTLT
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

// KavaHTLT extends BaseHTLT to include an ID and actions/updates to the HTLT
type KavaHTLT struct {
	HTLT    HTLT            `json:"htlt"`
	ID      uint64          `json:"id"`
	Updates UpdateKavaHTLTs `json:"updates"`
}

// KavaHTLTs is a slice of KavaHTLT
type KavaHTLTs []KavaHTLT

// // TODO: Validate verifies that the module account has enough coins to
// //		 complete HTLT and time is less than max time
// func (h HTLT) Validate() error {
// 	if h.EndTime.After(h.MaxEndTime) {
// 		return fmt.Errorf("MaxEndTime < EndTime (%s < %s)", h.MaxEndTime, h.EndTime)
// 	}
// 	return nil
// }

// WithID returns an HTLT with the ID set.
func (h KavaHTLT) WithID(id uint64) KavaHTLT { h.ID = id; return h }

// GetModuleAccountCoins returns the total number of coins held in the module account for this HTLT.
// It is used in genesis initialize the module account correctly.
func (h KavaHTLT) GetModuleAccountCoins() sdk.Coins {
	// We must convert BinanceChain.Coins to Cosmos.Coins
	var coins sdk.Coins
	for _, coin := range h.HTLT.Amount {
		coins = append(coins, sdk.NewCoin(coin.Denom, sdk.NewInt(coin.Amount)))
	}
	return coins
}

// UpdateKavaHTLT is a type shared by all HTLT update structs
type UpdateKavaHTLT struct {
	From   types.AccAddress `json:"from"`
	SwapID types.SwapBytes  `json:"swap_id"`
}

// UpdateKavaHTLTs is a slice of UpdateKavaHTLT
type UpdateKavaHTLTs []UpdateKavaHTLT

// KavaDepositHTLT defines an HTLT deposit
type KavaDepositHTLT struct {
	UpdateKavaHTLT
	Amount types.Coins `json:"amount"`
}

// Name returns a name for this update type. Used to identify updates in event attributes.
func (h KavaDepositHTLT) Name() string { return "deposit" }

// KavaClaimHTLT defines an HTLT claim
type KavaClaimHTLT struct {
	UpdateKavaHTLT
	RandomNumber types.SwapBytes
}

// Name returns a name for this update type. Used to identify updates in event attributes.
func (h KavaClaimHTLT) Name() string { return "claim" }

// KavaRefundHTLT defines an HTLT refund
type KavaRefundHTLT struct {
	UpdateKavaHTLT
}

// Name returns a name for this update type. Used to identify updates in event attributes.
func (h KavaRefundHTLT) Name() string { return "refund" }
