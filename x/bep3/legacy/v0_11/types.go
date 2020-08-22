package v0_11

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/tendermint/tendermint/crypto/tmhash"
	tmbytes "github.com/tendermint/tendermint/libs/bytes"
	tmtime "github.com/tendermint/tendermint/types/time"
)

var (
	AddrByteCount            = 20
	RandomNumberHashLength   = 32
	RandomNumberLength       = 32
	DefaultPreviousBlockTime = tmtime.Canonical(time.Unix(0, 0))
)

// GenesisState - all bep3 state that must be provided at genesis
type GenesisState struct {
	Params            Params        `json:"params" yaml:"params"`
	AtomicSwaps       AtomicSwaps   `json:"atomic_swaps" yaml:"atomic_swaps"`
	Supplies          AssetSupplies `json:"supplies" yaml:"supplies"`
	PreviousBlockTime time.Time     `json:"previous_block_time" yaml:"previous_block_time"`
}

// Validate validates genesis inputs. It returns error if validation of any input fails.
func (gs GenesisState) Validate() error {
	if err := gs.Params.Validate(); err != nil {
		return err
	}

	ids := map[string]bool{}
	for _, swap := range gs.AtomicSwaps {
		if ids[hex.EncodeToString(swap.GetSwapID())] {
			return fmt.Errorf("found duplicate atomic swap ID %s", hex.EncodeToString(swap.GetSwapID()))
		}

		if err := swap.Validate(); err != nil {
			return err
		}

		ids[hex.EncodeToString(swap.GetSwapID())] = true
	}

	supplyDenoms := map[string]bool{}
	for _, supply := range gs.Supplies {
		if err := supply.Validate(); err != nil {
			return err
		}
		if supplyDenoms[supply.GetDenom()] {
			return fmt.Errorf("found duplicate denom in asset supplies %s", supply.GetDenom())
		}
		supplyDenoms[supply.GetDenom()] = true
	}
	return nil
}

// Params governance parameters for the bep3 module
type Params struct {
	AssetParams AssetParams `json:"asset_params" yaml:"asset_params"`
}

// Validate ensure that params have valid values
func (p Params) Validate() error {
	return validateAssetParams(p.AssetParams)
}

func validateAssetParams(i interface{}) error {
	assetParams, ok := i.(AssetParams)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	coinDenoms := make(map[string]bool)
	for _, asset := range assetParams {
		if err := sdk.ValidateDenom(asset.Denom); err != nil {
			return fmt.Errorf(fmt.Sprintf("asset denom invalid: %s", asset.Denom))
		}

		if asset.CoinID < 0 {
			return fmt.Errorf(fmt.Sprintf("asset %s coin id must be a non negative integer", asset.Denom))
		}

		if asset.SupplyLimit.Limit.IsNegative() {
			return fmt.Errorf(fmt.Sprintf("asset %s has invalid (negative) supply limit: %s", asset.Denom, asset.SupplyLimit.Limit))
		}

		if asset.SupplyLimit.TimeBasedLimit.IsNegative() {
			return fmt.Errorf(fmt.Sprintf("asset %s has invalid (negative) supply time limit: %s", asset.Denom, asset.SupplyLimit.TimeBasedLimit))
		}

		if asset.SupplyLimit.TimeBasedLimit.GT(asset.SupplyLimit.Limit) {
			return fmt.Errorf(fmt.Sprintf("asset %s cannot have supply time limit > supply limit: %s>%s", asset.Denom, asset.SupplyLimit.TimeBasedLimit, asset.SupplyLimit.Limit))
		}

		_, found := coinDenoms[asset.Denom]
		if found {
			return fmt.Errorf(fmt.Sprintf("asset %s cannot have duplicate denom", asset.Denom))
		}

		coinDenoms[asset.Denom] = true

		if asset.DeputyAddress.Empty() {
			return fmt.Errorf("deputy address cannot be empty for %s", asset.Denom)
		}

		if len(asset.DeputyAddress.Bytes()) != sdk.AddrLen {
			return fmt.Errorf("%s deputy address invalid bytes length got %d, want %d", asset.Denom, len(asset.DeputyAddress.Bytes()), sdk.AddrLen)
		}

		if asset.FixedFee.IsNegative() {
			return fmt.Errorf("asset %s cannot have a negative fixed fee %s", asset.Denom, asset.FixedFee)
		}

		if asset.MinBlockLock > asset.MaxBlockLock {
			return fmt.Errorf("asset %s has minimum block lock > maximum block lock %d > %d", asset.Denom, asset.MinBlockLock, asset.MaxBlockLock)
		}

		if !asset.MinSwapAmount.IsPositive() {
			return fmt.Errorf(fmt.Sprintf("asset %s must have a positive minimum swap amount, got %s", asset.Denom, asset.MinSwapAmount))
		}

		if !asset.MaxSwapAmount.IsPositive() {
			return fmt.Errorf(fmt.Sprintf("asset %s must have a positive maximum swap amount, got %s", asset.Denom, asset.MaxSwapAmount))
		}

		if asset.MinSwapAmount.GT(asset.MaxSwapAmount) {
			return fmt.Errorf("asset %s has minimum swap amount > maximum swap amount %s > %s", asset.Denom, asset.MinSwapAmount, asset.MaxSwapAmount)
		}
	}

	return nil
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

// CalculateSwapID calculates the hash of a RandomNumberHash, sdk.AccAddress, and string
func CalculateSwapID(randomNumberHash []byte, sender sdk.AccAddress, senderOtherChain string) []byte {
	senderOtherChain = strings.ToLower(senderOtherChain)
	data := randomNumberHash
	data = append(data, sender.Bytes()...)
	data = append(data, []byte(senderOtherChain)...)
	return tmhash.Sum(data)
}

// GetSwapID calculates the ID of an atomic swap
func (a AtomicSwap) GetSwapID() tmbytes.HexBytes {
	return CalculateSwapID(a.RandomNumberHash, a.Sender, a.SenderOtherChain)
}

// GetCoins returns the swap's amount as sdk.Coins
func (a AtomicSwap) GetCoins() sdk.Coins {
	return sdk.NewCoins(a.Amount...)
}

// Validate performs a basic validation of an atomic swap fields.
func (a AtomicSwap) Validate() error {
	if !a.Amount.IsValid() {
		return fmt.Errorf("invalid amount: %s", a.Amount)
	}
	if !a.Amount.IsAllPositive() {
		return fmt.Errorf("the swapped out coin must be positive: %s", a.Amount)
	}
	if len(a.RandomNumberHash) != RandomNumberHashLength {
		return fmt.Errorf("the length of random number hash should be %d", RandomNumberHashLength)
	}
	if a.ExpireHeight == 0 {
		return errors.New("expire height cannot be 0")
	}
	if a.Timestamp == 0 {
		return errors.New("timestamp cannot be 0")
	}
	if a.Sender.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "sender cannot be empty")
	}
	if a.Recipient.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "recipient cannot be empty")
	}
	if len(a.Sender) != AddrByteCount {
		return fmt.Errorf("the expected address length is %d, actual length is %d", AddrByteCount, len(a.Sender))
	}
	if len(a.Recipient) != AddrByteCount {
		return fmt.Errorf("the expected address length is %d, actual length is %d", AddrByteCount, len(a.Recipient))
	}
	// NOTE: These adresses may not have a bech32 prefix.
	if strings.TrimSpace(a.SenderOtherChain) == "" {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "sender other chain cannot be blank")
	}
	if strings.TrimSpace(a.RecipientOtherChain) == "" {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "recipient other chain cannot be blank")
	}
	if a.Status == Completed && a.ClosedBlock == 0 {
		return errors.New("closed block cannot be 0")
	}
	if a.Status == NULL || a.Status > 3 {
		return errors.New("invalid swap status")
	}
	if a.Direction == INVALID || a.Direction > 2 {
		return errors.New("invalid swap direction")
	}
	return nil
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

// MarshalJSON marshals the SwapStatus
func (status SwapStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(status.String())
}

// UnmarshalJSON unmarshals the SwapStatus
func (status *SwapStatus) UnmarshalJSON(data []byte) error {
	var s string
	err := json.Unmarshal(data, &s)
	if err != nil {
		return err
	}
	*status = NewSwapStatusFromString(s)
	return nil
}

// NewSwapStatusFromString converts string to SwapStatus type
func NewSwapStatusFromString(str string) SwapStatus {
	switch str {
	case "Open", "open":
		return Open
	case "Completed", "completed":
		return Completed
	case "Expired", "expired":
		return Expired
	default:
		return NULL
	}
}

// String returns the string representation of a SwapStatus
func (status SwapStatus) String() string {
	switch status {
	case Open:
		return "Open"
	case Completed:
		return "Completed"
	case Expired:
		return "Expired"
	default:
		return "NULL"
	}
}

// SwapDirection is the direction of an AtomicSwap
type SwapDirection byte

const (
	INVALID  SwapDirection = 0x00
	Incoming SwapDirection = 0x01
	Outgoing SwapDirection = 0x02
)

// NewSwapDirectionFromString converts string to SwapDirection type
func NewSwapDirectionFromString(str string) SwapDirection {
	switch str {
	case "Incoming", "incoming", "inc", "I", "i":
		return Incoming
	case "Outgoing", "outgoing", "out", "O", "o":
		return Outgoing
	default:
		return INVALID
	}
}

// String returns the string representation of a SwapDirection
func (direction SwapDirection) String() string {
	switch direction {
	case Incoming:
		return "Incoming"
	case Outgoing:
		return "Outgoing"
	default:
		return "INVALID"
	}
}

// MarshalJSON marshals the SwapDirection
func (direction SwapDirection) MarshalJSON() ([]byte, error) {
	return json.Marshal(direction.String())
}

// UnmarshalJSON unmarshals the SwapDirection
func (direction *SwapDirection) UnmarshalJSON(data []byte) error {
	var s string
	err := json.Unmarshal(data, &s)
	if err != nil {
		return err
	}
	*direction = NewSwapDirectionFromString(s)
	return nil
}

// IsValid returns true if the swap direction is valid and false otherwise.
func (direction SwapDirection) IsValid() bool {
	if direction == Incoming ||
		direction == Outgoing {
		return true
	}
	return false
}
