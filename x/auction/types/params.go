package types

import (
	"bytes"
	"fmt"
	"time"

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
	KeyAuctionBidDuration = []byte("MaxBidDuration")
	KeyAuctionDuration    = []byte("MaxAuctionDuration")
)

var _ subspace.ParamSet = &Params{}

// Params is the governance parameters for the auction module.
type Params struct {
	MaxAuctionDuration time.Duration `json:"max_auction_duration" yaml:"max_auction_duration"` // max length of auction
	MaxBidDuration     time.Duration `json:"max_bid_duration" yaml:"max_bid_duration"`         // additional time added to the auction end time after each bid, capped by the expiry.
}

// NewParams returns a new Params object.
func NewParams(maxAuctionDuration time.Duration, bidDuration time.Duration) Params {
	return Params{
		MaxAuctionDuration: maxAuctionDuration,
		MaxBidDuration:     bidDuration,
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
// nolint
func (ap *Params) ParamSetPairs() subspace.ParamSetPairs {
	return subspace.ParamSetPairs{
		{KeyAuctionBidDuration, &ap.MaxBidDuration},
		{KeyAuctionDuration, &ap.MaxAuctionDuration},
	}
}

// Equal returns a boolean determining if two Params types are identical.
func (ap Params) Equal(ap2 Params) bool {
	bz1 := ModuleCdc.MustMarshalBinaryLengthPrefixed(&ap)
	bz2 := ModuleCdc.MustMarshalBinaryLengthPrefixed(&ap2)
	return bytes.Equal(bz1, bz2)
}

// String implements stringer interface
func (ap Params) String() string {
	return fmt.Sprintf(`Auction Params:
	Max Auction Duration: %s
	Max Bid Duration: %s`, ap.MaxAuctionDuration, ap.MaxBidDuration)
}

// Validate checks that the parameters have valid values.
func (ap Params) Validate() error {
	// TODO check durations are within acceptable limits, if needed
	return nil
}
