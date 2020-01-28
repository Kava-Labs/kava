package types

import (
	"time"

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

// KHTLT extends BaseHTLT to include an ID and actions/updates to the HTLT
type KHTLT struct {
	HTLT        HTLT         `json:"htlt"`
	ID          uint64       `json:"id"`
	OriginChain string       `json:"origin_chain"`
	EndTime     time.Time    `json:"end_time"`
	Updates     UpdatesKHTLT `json:"updates"`
	Status      string       `json:"status"` // TODO: Use enum for status
}

// KHTLTs is a slice of KHTLTs
type KHTLTs []KHTLT

// // TODO: Validate verifies that the module account has enough coins to
// //		 complete HTLT and time is less than max time
// func (h HTLT) Validate() error {
// 	if h.EndTime.After(h.MaxEndTime) {
// 		return fmt.Errorf("MaxEndTime < EndTime (%s < %s)", h.MaxEndTime, h.EndTime)
// 	}
// 	return nil
// }

// UpdateKHTLT is an interface for handling common actions on a KHTLT
type UpdateKHTLT interface {
	GetFrom() []byte
	GetSwapID(types.SwapBytes) KHTLT // TODO: Is this correct?
	GetInitiator() string
	GetParentID() uint64
}

// UpdatesKHTLT is a slice of UpdateKHTLT.
type UpdatesKHTLT []UpdateKHTLT

// WithID returns an HTLT with the ID set.
func (h KHTLT) WithID(id uint64) KHTLT { h.ID = id; return h }

// GetModuleAccountCoins returns the total number of coins held in the module account for this HTLT.
// It is used in genesis initialize the module account correctly.
func (h KHTLT) GetModuleAccountCoins() sdk.Coins {
	// We must convert BinanceChain.Coins to Cosmos.Coins
	var coins sdk.Coins
	for _, coin := range h.HTLT.Amount {
		coins = append(coins, sdk.NewCoin(coin.Denom, sdk.NewInt(coin.Amount)))
	}
	return coins
}

// TODO: UpdateKHTLT is a type shared by all HTLT update structs
// type UpdateKHTLT struct {
// 	From   types.AccAddress `json:"from"`
// 	SwapID types.SwapBytes  `json:"swap_id"`
// }
// TODO: UpdateKHTLTs is a slice of UpdateKHTLT
// type UpdateKHTLTs []UpdateKHTLT

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
