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
	GetBidder() sdk.AccAddress
	GetBid() sdk.Coin
	GetLot() sdk.Coin
	GetEndTime() time.Time
}

// BaseAuction type shared by all Auctions
type BaseAuction struct {
	ID         uint64
	Initiator  string         // Module that starts the auction. Giving away Lot (aka seller in a forward auction). Restricted to being a module account name rather than any account.
	Lot        sdk.Coin       // Amount of coins up being given by initiator (FA - amount for sale by seller, RA - cost of good by buyer (bid))
	Bidder     sdk.AccAddress // Person who bids in the auction. Receiver of Lot. (aka buyer in forward auction, seller in RA)
	Bid        sdk.Coin       // Amount of coins being given by the bidder (FA - bid, RA - amount being sold)
	EndTime    time.Time      // Auction closing time. Triggers at the end of the block with time ≥ endTime (bids placed in that block are valid) // TODO ensure everything is consistent with this
	MaxEndTime time.Time      // Maximum closing time. Auctions can close before this but never after.
}

//SurplusAuction type for forward auctions
typeSurplusAuction struct {
	BaseAuction
}

// DebtAuction type for reverse auctions
type DebtAuction struct {
	BaseAuction
}

// WeightedAddresses type for storing an address and its associated weight
type WeightedAddresses struct {
	Addresses []sdk.AccAddress
	Weights   []sdk.Int
}

// CollateralAuction type for forward reverse auction
type CollateralAuction struct {
	BaseAuction
	MaxBid     sdk.Coin
	LotReturns WeightedAddresses // return addresses to pay out reductions in the lot amount to. Lot is bid down during reverse phase.
}
```
