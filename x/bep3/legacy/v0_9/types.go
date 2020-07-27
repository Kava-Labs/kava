package v0_9

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	tmbytes "github.com/tendermint/tendermint/libs/bytes"
)

const (
	ModuleName = "bep3"
)

var (
	DefaultBnbDeputyFixedFee sdk.Int = sdk.NewInt(1000) // 0.00001 BNB
	DefaultMinAmount         sdk.Int = sdk.ZeroInt()
	DefaultMaxAmount         sdk.Int = sdk.NewInt(1000000000000) // 10,000 BNB
	DefaultMinBlockLock      uint64  = 220
	DefaultMaxBlockLock      uint64  = 270
	DefaultSupportedAssets           = AssetParams{
		AssetParam{
			Denom:  "bnb",
			CoinID: 714,
			Limit:  sdk.NewInt(350000000000000), // 3,500,000 BNB
			Active: true,
		},
	}
	KeySupportedAssets = []byte("SupportedAssets")
)

// Params v0.9 governance parameters for bep3 module
type Params struct {
	BnbDeputyAddress  sdk.AccAddress `json:"bnb_deputy_address" yaml:"bnb_deputy_address"`     // Bnbchain deputy address
	BnbDeputyFixedFee sdk.Int        `json:"bnb_deputy_fixed_fee" yaml:"bnb_deputy_fixed_fee"` // Deputy fixed fee in BNB
	MinAmount         sdk.Int        `json:"min_amount" yaml:"min_amount"`                     // Minimum swap amount
	MaxAmount         sdk.Int        `json:"max_amount" yaml:"max_amount"`                     // Maximum swap amount
	MinBlockLock      uint64         `json:"min_block_lock" yaml:"min_block_lock"`             // Minimum swap block lock
	MaxBlockLock      uint64         `json:"max_block_lock" yaml:"max_block_lock"`             // Maximum swap block lock
	SupportedAssets   AssetParams    `json:"supported_assets" yaml:"supported_assets"`         // Supported assets
}

// NewParams returns a new params object
func NewParams(bnbDeputyAddress sdk.AccAddress, bnbDeputyFixedFee, minAmount,
	maxAmount sdk.Int, minBlockLock, maxBlockLock uint64, supportedAssets AssetParams,
) Params {
	return Params{
		BnbDeputyAddress:  bnbDeputyAddress,
		BnbDeputyFixedFee: bnbDeputyFixedFee,
		MinAmount:         minAmount,
		MaxAmount:         maxAmount,
		MinBlockLock:      minBlockLock,
		MaxBlockLock:      maxBlockLock,
		SupportedAssets:   supportedAssets,
	}
}

// AssetParam v0.9 governance parameters for each asset supported on kava chain
type AssetParam struct {
	Denom  string  `json:"denom" yaml:"denom"`     // name of the asset
	CoinID int     `json:"coin_id" yaml:"coin_id"` // internationally recognized coin ID
	Limit  sdk.Int `json:"limit" yaml:"limit"`     // asset supply limit
	Active bool    `json:"active" yaml:"active"`   // denotes if asset is available or paused
}

// AssetParams is a slice of AssetParams
type AssetParams []AssetParam

// GenesisState - all bep3 state that must be provided at genesis
type GenesisState struct {
	Params        Params        `json:"params" yaml:"params"`
	AtomicSwaps   AtomicSwaps   `json:"atomic_swaps" yaml:"atomic_swaps"`
	AssetSupplies AssetSupplies `json:"assets_supplies" yaml:"assets_supplies"`
}

// NewGenesisState creates a new GenesisState object
func NewGenesisState(params Params, swaps AtomicSwaps, supplies AssetSupplies) GenesisState {
	return GenesisState{
		Params:        params,
		AtomicSwaps:   swaps,
		AssetSupplies: supplies,
	}
}

// DefaultGenesisState - default GenesisState used by Cosmos Hub
func DefaultGenesisState() GenesisState {
	return NewGenesisState(
		DefaultParams(),
		AtomicSwaps{},
		AssetSupplies{},
	)
}

// DefaultParams returns default params for bep3 module
func DefaultParams() Params {
	defaultBnbDeputyAddress, err := sdk.AccAddressFromBech32("kava1r4v2zdhdalfj2ydazallqvrus9fkphmglhn6u6")
	if err != nil {
		panic(err)
	}

	return NewParams(defaultBnbDeputyAddress, DefaultBnbDeputyFixedFee, DefaultMinAmount,
		DefaultMaxAmount, DefaultMinBlockLock, DefaultMaxBlockLock, DefaultSupportedAssets)
}

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

// AssetSupply contains information about an asset's supply
type AssetSupply struct {
	Denom          string   `json:"denom"  yaml:"denom"`
	IncomingSupply sdk.Coin `json:"incoming_supply"  yaml:"incoming_supply"`
	OutgoingSupply sdk.Coin `json:"outgoing_supply"  yaml:"outgoing_supply"`
	CurrentSupply  sdk.Coin `json:"current_supply"  yaml:"current_supply"`
	SupplyLimit    sdk.Coin `json:"supply_limit"  yaml:"supply_limit"`
}

// AssetSupplies is a slice of AssetSupply
type AssetSupplies []AssetSupply

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
