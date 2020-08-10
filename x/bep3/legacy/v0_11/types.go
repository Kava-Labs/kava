package v0_11

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	tmbytes "github.com/tendermint/tendermint/libs/bytes"
	tmtime "github.com/tendermint/tendermint/types/time"
)

// GenesisState - all bep3 state that must be provided at genesis
type GenesisState struct {
	Params            Params        `json:"params" yaml:"params"`
	AtomicSwaps       AtomicSwaps   `json:"atomic_swaps" yaml:"atomic_swaps"`
	Supplies          AssetSupplies `json:"supplies" yaml:"supplies"`
	PreviousBlockTime time.Time     `json:"previous_block_time" yaml:"previous_block_time"`
}

// Params governance parameters for the bep3 module
type Params struct {
	AssetParams AssetParams `json:"asset_params" yaml:"asset_params"`
}

// AssetParam parameters that must be specified for each bep3 asset
type AssetParam struct {
	Denom         string         `json:"denom" yaml:"denom"`                                     // name of the asset
	CoinID        int            `json:"coin_id" yaml:"coin_id"`                                 // SLIP-0044 registered coin type - see https://github.com/satoshilabs/slips/blob/master/slip-0044.md
	SupplyLimit   SupplyLimit    `json:"supply_limit" yaml:"supply_limit"`                       // asset supply limit
	Active        bool           `json:"active" yaml:"active"`                                   // denotes if asset is available or paused
	DeputyAddress sdk.AccAddress `json:"deputy_address" yaml:"deputy_address"`                   // the address of the relayer process
	FixedFee      sdk.Int        `json:"incoming_swap_fixed_fee" yaml:"incoming_swap_fixed_fee"` // the fixed fee charged by the relayer process for incoming swaps
	MinSwapAmount sdk.Int        `json:"min_swap_amount" yaml:"min_swap_amount"`                 // Minimum swap amount
	MaxSwapAmount sdk.Int        `json:"max_swap_amount" yaml:"max_swap_amount"`                 // Maximum swap amount
	MinBlockLock  uint64         `json:"min_block_lock" yaml:"min_block_lock"`                   // Minimum swap block lock
	MaxBlockLock  uint64         `json:"max_block_lock" yaml:"max_block_lock"`                   // Maximum swap block lock
}

// SupplyLimit parameters that control the absolute and time-based limits for an assets's supply
type SupplyLimit struct {
	Limit          sdk.Int       `json:"limit" yaml:"limit"`                       // the absolute supply limit for an asset
	TimeLimited    bool          `json:"time_limited" yaml:"time_limited"`         // boolean for if the supply is also limited by time
	TimePeriod     time.Duration `json:"time_period" yaml:"time_period"`           // the duration for which the supply time limit applies
	TimeBasedLimit sdk.Int       `json:"time_based_limit" yaml:"time_based_limit"` // the supply limit for an asset for each time period
}

// AssetParams array of AssetParam
type AssetParams []AssetParam

// AssetSupply contains information about an asset's supply
type AssetSupply struct {
	IncomingSupply           sdk.Coin      `json:"incoming_supply"  yaml:"incoming_supply"`
	OutgoingSupply           sdk.Coin      `json:"outgoing_supply"  yaml:"outgoing_supply"`
	CurrentSupply            sdk.Coin      `json:"current_supply"  yaml:"current_supply"`
	TimeLimitedCurrentSupply sdk.Coin      `json:"time_limited_current_supply" yaml:"time_limited_current_supply"`
	TimeElapsed              time.Duration `json:"time_elapsed" yaml:"time_elapsed"`
}

// NewAssetSupply initializes a new AssetSupply
func NewAssetSupply(incomingSupply, outgoingSupply, currentSupply, timeLimitedSupply sdk.Coin, timeElapsed time.Duration) AssetSupply {
	return AssetSupply{
		IncomingSupply:           incomingSupply,
		OutgoingSupply:           outgoingSupply,
		CurrentSupply:            currentSupply,
		TimeLimitedCurrentSupply: timeLimitedSupply,
		TimeElapsed:              timeElapsed,
	}
}

// Validate performs a basic validation of an asset supply fields.
func (a AssetSupply) Validate() error {
	if !a.IncomingSupply.IsValid() {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidCoins, "incoming supply %s", a.IncomingSupply)
	}
	if !a.OutgoingSupply.IsValid() {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidCoins, "outgoing supply %s", a.OutgoingSupply)
	}
	if !a.CurrentSupply.IsValid() {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidCoins, "current supply %s", a.CurrentSupply)
	}
	if !a.TimeLimitedCurrentSupply.IsValid() {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidCoins, "time-limited current supply %s", a.CurrentSupply)
	}
	denom := a.CurrentSupply.Denom
	if (a.IncomingSupply.Denom != denom) ||
		(a.OutgoingSupply.Denom != denom) ||
		(a.TimeLimitedCurrentSupply.Denom != denom) {
		return fmt.Errorf("asset supply denoms do not match %s %s %s %s", a.CurrentSupply.Denom, a.IncomingSupply.Denom, a.OutgoingSupply.Denom, a.TimeLimitedCurrentSupply.Denom)
	}
	return nil
}

// Equal returns if two asset supplies are equal
func (a AssetSupply) Equal(b AssetSupply) bool {
	if a.GetDenom() != b.GetDenom() {
		return false
	}
	return (a.IncomingSupply.IsEqual(b.IncomingSupply) &&
		a.CurrentSupply.IsEqual(b.CurrentSupply) &&
		a.OutgoingSupply.IsEqual(b.OutgoingSupply) &&
		a.TimeLimitedCurrentSupply.IsEqual(b.TimeLimitedCurrentSupply) &&
		a.TimeElapsed == b.TimeElapsed)
}

// String implements stringer
func (a AssetSupply) String() string {
	return fmt.Sprintf(`
	asset supply:
		Incoming supply:    %s
		Outgoing supply:    %s
		Current supply:     %s
		Time-limited current cupply: %s
		Time elapsed: %s
		`,
		a.IncomingSupply, a.OutgoingSupply, a.CurrentSupply, a.TimeLimitedCurrentSupply, a.TimeElapsed)
}

// GetDenom getter method for the denom of the asset supply
func (a AssetSupply) GetDenom() string {
	return a.CurrentSupply.Denom
}

// AssetSupplies is a slice of AssetSupply
type AssetSupplies []AssetSupply

// AtomicSwap contains the information for an atomic swap
type AtomicSwap struct {
	Amount              sdk.Coins        `json:"amount"  yaml:"amount"`
	RandomNumberHash    tmbytes.HexBytes `json:"random_number_hash"  yaml:"random_number_hash"`
	ExpireHeight        uint64           `json:"expire_height"  yaml:"expire_height"`
	Timestamp           int64            `json:"timestamp"  yaml:"timestamp"`
	Sender              sdk.AccAddress   `json:"sender"  yaml:"sender"`
	Recipient           sdk.AccAddress   `json:"recipient"  yaml:"recipient"`
	SenderOtherChain    string           `json:"sender_other_chain"  yaml:"sender_other_chain"`
	RecipientOtherChain string           `json:"recipient_other_chain"  yaml:"recipient_other_chain"`
	ClosedBlock         int64            `json:"closed_block"  yaml:"closed_block"`
	Status              SwapStatus       `json:"status"  yaml:"status"`
	CrossChain          bool             `json:"cross_chain"  yaml:"cross_chain"`
	Direction           SwapDirection    `json:"direction"  yaml:"direction"`
}

// AtomicSwaps is a slice of AtomicSwap
type AtomicSwaps []AtomicSwap

// SwapStatus is the status of an AtomicSwap
type SwapStatus byte

// swap statuses
const (
	NULL      SwapStatus = 0x00
	Open      SwapStatus = 0x01
	Completed SwapStatus = 0x02
	Expired   SwapStatus = 0x03
)

// SwapDirection is the direction of an AtomicSwap
type SwapDirection byte

const (
	INVALID  SwapDirection = 0x00
	Incoming SwapDirection = 0x01
	Outgoing SwapDirection = 0x02
)

var (
	DefaultPreviousBlockTime = tmtime.Canonical(time.Unix(0, 0))
)
