<!--
order: 2
-->

# State

## Parameters and genesis state

`Paramaters` define the rules according to which swaps are executed. Parameter updates can be made via on-chain parameter update proposals.

```go
// Params governance parameters for bep3 module
type Params struct {
	BnbDeputyAddress  sdk.AccAddress `json:"bnb_deputy_address" yaml:"bnb_deputy_address"`     // Bnbchain deputy address
	BnbDeputyFixedFee sdk.Int        `json:"bnb_deputy_fixed_fee" yaml:"bnb_deputy_fixed_fee"` // Deputy fixed fee in BNB
	MinAmount         sdk.Int        `json:"min_amount" yaml:"min_amount"`                     // Minimum swap amount
	MaxAmount         sdk.Int        `json:"max_amount" yaml:"max_amount"`                     // Maximum swap amount
	MinBlockLock      uint64         `json:"min_block_lock" yaml:"min_block_lock"`             // Minimum swap block lock
	MaxBlockLock      uint64         `json:"max_block_lock" yaml:"max_block_lock"`             // Maximum swap block lock
	SupportedAssets   AssetParams    `json:"supported_assets" yaml:"supported_assets"`         // Supported assets
}

// AssetParam governance parameters for each asset within a supported chain
type AssetParam struct {
	Denom  string  `json:"denom" yaml:"denom"`     // name of the asset
	CoinID int     `json:"coin_id" yaml:"coin_id"` // internationally recognized coin ID
	Limit  sdk.Int `json:"limit" yaml:"limit"`     // asset supply limit
	Active bool    `json:"active" yaml:"active"`   // denotes if asset is active or paused
}
```

`GenesisState` defines the state that must be persisted when the blockchain stops/restarts in order for normal function of the bep3 module to resume.

```go
// GenesisState - all bep3 state that must be provided at genesis
type GenesisState struct {
	Params        Params        `json:"params" yaml:"params"`
	AtomicSwaps   AtomicSwaps   `json:"atomic_swaps" yaml:"atomic_swaps"`
	AssetSupplies AssetSupplies `json:"assets_supplies" yaml:"assets_supplies"`
}
```

## Types

```go
// AtomicSwap contains the information for an atomic swap
type AtomicSwap struct {
	Amount              sdk.Coins        `json:"amount"  yaml:"amount"`
	RandomNumberHash    tmbytes.HexBytes `json:"random_number_hash"  yaml:"random_number_hash"`
	ExpireHeight        int64            `json:"expire_height"  yaml:"expire_height"`
	Timestamp           int64            `json:"timestamp"  yaml:"timestamp"`
	Sender              sdk.AccAddress   `json:"sender"  yaml:"sender"`
	Recipient           sdk.AccAddress   `json:"recipient"  yaml:"recipient"`
	SenderOtherChain    string           `json:"sender_other_chain"  yaml:"sender_other_chain"`
	RecipientOtherChain string           `json:"recipient_other_chain"  yaml:"recipient_other_chain"`
	ClosedBlock         int64            `json:"closed_block"  yaml:"closed_block"`
	Status              SwapStatus       `json:"status"  yaml:"status"`
	Direction           SwapDirection    `json:"direction"  yaml:"direction"`
}

// SwapStatus is the status of an AtomicSwap
type SwapStatus byte

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

// AssetSupply contains information about an asset's supply
type AssetSupply struct {
	Denom          string   `json:"denom"  yaml:"denom"`
	IncomingSupply sdk.Coin `json:"incoming_supply"  yaml:"incoming_supply"`
	OutgoingSupply sdk.Coin `json:"outgoing_supply"  yaml:"outgoing_supply"`
	CurrentSupply  sdk.Coin `json:"current_supply"  yaml:"current_supply"`
	SupplyLimit    sdk.Coin `json:"supply_limit"  yaml:"supply_limit"`
}
```