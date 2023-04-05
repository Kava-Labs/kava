<!--
order: 2
-->

# State

## Parameters and genesis state

`Paramaters` define the rules according to which auctions are run. There is only one active parameter set at any given time. Updates to the parameter set can be made via on-chain parameter update proposals.

```go
// Params governance parameters for auction module
type Params struct {
	MaxAuctionDuration time.Duration `json:"max_auction_duration" yaml:"max_auction_duration"` // max length of auction
	MaxBidDuration     time.Duration `json:"max_bid_duration" yaml:"max_bid_duration"` // additional time added to the auction end time after each bid, capped by the expiry.
	IncrementSurplus    sdk.Dec       `json:"increment_surplus" yaml:"increment_surplus"`       // percentage change (of auc.Bid) required for a new bid on a surplus auction
	IncrementDebt       sdk.Dec       `json:"increment_debt" yaml:"increment_debt"`             // percentage change (of auc.Lot) required for a new bid on a debt auction
	IncrementCollateral sdk.Dec       `json:"increment_collateral" yaml:"increment_collateral"` // percentage change (of auc.Bid or auc.Lot) required for a new bid on a collateral auction
}
```

`GenesisState` defines the state that must be persisted when the blockchain stops/restarts in order for normal function of the auction module to resume.

```go
// GenesisState - auction state that must be provided at genesis
type GenesisState struct {
	NextAuctionID uint64          `json:"next_auction_id" yaml:"next_auction_id"` // auctionID that will be used for the next created auction
	Params        Params          `json:"auction_params" yaml:"auction_params"` // auction params
	Auctions      Auctions `json:"genesis_auctions" yaml:"genesis_auctions"` // auctions currently in the store
}
```

## Base types

```go
// Auction is an interface to several types of auction.
type Auction interface {
	GetID() uint64
	WithID(uint64) Auction
	GetEndTime() time.Time
}

// BaseAuction is a common type shared by all Auctions.
type BaseAuction struct {
	ID         uint64
	Initiator  string         // Module name that starts the auction. Pays out Lot.
	Lot        sdk.Coin       // Coins that will paid out by Initiator to the winning bidder.
	Bidder     sdk.AccAddress // Latest bidder. Receiver of Lot.
	Bid        sdk.Coin       // Coins paid into the auction the bidder.
	EndTime    time.Time      // Current auction closing time. Triggers at the end of the block with time ≥ EndTime.
	MaxEndTime time.Time      // Maximum closing time. Auctions can close before this but never after.
}

// SurplusAuction is a forward auction that burns what it receives from bids.
// It is normally used to sell off excess pegged asset acquired by the CDP system.
type SurplusAuction struct {
	BaseAuction
}

// DebtAuction is a reverse auction that mints what it pays out.
// It is normally used to acquire pegged asset to cover the CDP system's debts that were not covered by selling collateral.
type DebtAuction struct {
	BaseAuction
}

// WeightedAddresses is a type for storing some addresses and associated weights.
type WeightedAddresses struct {
	Addresses []sdk.AccAddress
	Weights   []sdkmath.Int
}

// CollateralAuction is a two phase auction.
// Initially, in forward auction phase, bids can be placed up to a max bid.
// Then it switches to a reverse auction phase, where the initial amount up for auction is bid down.
// Unsold Lot is sent to LotReturns, being divided among the addresses by weight.
// Collateral auctions are normally used to sell off collateral seized from CDPs.
type CollateralAuction struct {
	BaseAuction
	MaxBid     sdk.Coin
	LotReturns WeightedAddresses
}
```
