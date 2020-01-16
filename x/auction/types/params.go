package types

import (
	"bytes"
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/params/subspace"
)

// Defaults for auction params
const (
	// DefaultMaxAuctionDuration max length of auction
	DefaultMaxAuctionDuration time.Duration = 2 * 24 * time.Hour
	// DefaultBidDuration how long an auction gets extended when someone bids
	DefaultBidDuration time.Duration = 1 * time.Hour
)

// Parameter keys
var (
	// ParamStoreKeyParams Param store key for auction params
	KeyAuctionBidDuration = []byte("BidDuration")
	KeyAuctionDuration    = []byte("MaxAuctionDuration")
)

var _ subspace.ParamSet = &Params{}

// Params is the governance parameters for the auction module.
type Params struct {
	MaxAuctionDuration time.Duration `json:"max_auction_duration" yaml:"max_auction_duration"` // max length of auction
	BidDuration        time.Duration `json:"bid_duration" yaml:"bid_duration"`                 // additional time added to the auction end time after each bid, capped by the expiry.
}

// NewParams returns a new Params object.
func NewParams(maxAuctionDuration time.Duration, bidDuration time.Duration) Params {
	return Params{
		MaxAuctionDuration: maxAuctionDuration,
		BidDuration:        bidDuration,
	}
}

// DefaultParams returns the default parameters for auctions.
func DefaultParams() Params {
	return NewParams(
		DefaultMaxAuctionDuration,
		DefaultBidDuration,
	)
}

// ParamKeyTable Key declaration for parameters
func ParamKeyTable() subspace.KeyTable {
	return subspace.NewKeyTable().RegisterParamSet(&Params{})
}

// ParamSetPairs implements the ParamSet interface and returns all the key/value pairs.
func (p *Params) ParamSetPairs() subspace.ParamSetPairs {
	return subspace.ParamSetPairs{
		{Key: KeyAuctionBidDuration, Value: &p.BidDuration},
		{Key: KeyAuctionDuration, Value: &p.MaxAuctionDuration},
	}
}

// Equal returns a boolean determining if two Params types are identical.
func (p Params) Equal(p2 Params) bool {
	bz1 := ModuleCdc.MustMarshalBinaryLengthPrefixed(&p)
	bz2 := ModuleCdc.MustMarshalBinaryLengthPrefixed(&p2)
	return bytes.Equal(bz1, bz2)
}

// String implements stringer interface
func (p Params) String() string {
	return fmt.Sprintf(`Auction Params:
	Max Auction Duration: %s
	Bid Duration: %s`, p.MaxAuctionDuration, p.BidDuration)
}

// Validate checks that the parameters have valid values.
func (p Params) Validate() error {
	if p.BidDuration < 0 {
		return sdk.ErrInternal("bid duration cannot be negative")
	}
	if p.MaxAuctionDuration < 0 {
		return sdk.ErrInternal("max auction duration cannot be negative")
	}
	if p.BidDuration > p.MaxAuctionDuration {
		return sdk.ErrInternal("bid duration param cannot be larger than max auction duration")
	}
	return nil
}
