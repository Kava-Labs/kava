package types

import (
	"github.com/binance-chain/go-sdk/common/types"
)

// BaseHTLT contains the base of an HTLT
type BaseHTLT struct {
	From                types.AccAddress
	To                  types.AccAddress
	RecipientOtherChain string
	SenderOtherChain    string
	RandomNumberHash    types.SwapBytes
	Timestamp           int64
	Amount              types.Coins
	ExpectedIncome      string
	HeightSpan          int64
	CrossChain          bool
}

// KavaHTLT extends BaseHTLT to include an ID and actions/updates to the HTLT
type KavaHTLT struct {
	BaseHTLT
	ID      uint64
	Updates UpdateHTLTs
}

// KavaHTLTs is a slice of KavaHTLT
type KavaHTLTs []KavaHTLT

// WithID returns an auction with the ID set.
func (h KavaHTLT) WithID(id uint64) KavaHTLT { h.ID = id; return h }

// UpdateHTLT is a type shared by all HTLT update structs
type UpdateHTLT struct {
	From   types.AccAddress
	SwapID types.SwapBytes
}

// UpdateHTLTs is a slice of UpdateHTLT
type UpdateHTLTs []UpdateHTLT

// KavaDepositHTLT defines an HTLT deposit
type KavaDepositHTLT struct {
	UpdateHTLT
	Amount types.Coins
}

// Name returns a name for this update type. Used to identify updates in event attributes.
func (h KavaDepositHTLT) Name() string { return "deposit" }

// KavaClaimHTLT defines an HTLT claim
type KavaClaimHTLT struct {
	UpdateHTLT
	RandomNumber types.SwapBytes
}

// Name returns a name for this update type. Used to identify updates in event attributes.
func (h KavaClaimHTLT) Name() string { return "claim" }

// KavaRefundHTLT defines an HTLT refund
type KavaRefundHTLT struct {
	UpdateHTLT
}

// Name returns a name for this update type. Used to identify updates in event attributes.
func (h KavaRefundHTLT) Name() string { return "refund" }
