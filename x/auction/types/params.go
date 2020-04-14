package types

import (
	"bytes"
	"errors"
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/params"
	"github.com/cosmos/cosmos-sdk/x/params/subspace"
)

// Defaults for auction params
const (
	// DefaultMaxAuctionDuration max length of auction
	DefaultMaxAuctionDuration time.Duration = 2 * 24 * time.Hour
	// DefaultBidDuration how long an auction gets extended when someone bids
	DefaultBidDuration time.Duration = 1 * time.Hour
)

var (
	// DefaultIncrement is the smallest percent change a new bid must have from the old one
	DefaultIncrement sdk.Dec = sdk.MustNewDecFromStr("0.05")
	// ParamStoreKeyParams Param store key for auction params
	KeyBidDuration         = []byte("BidDuration")
	KeyMaxAuctionDuration  = []byte("MaxAuctionDuration")
	KeyIncrementSurplus    = []byte("IncrementSurplus")
	KeyIncrementDebt       = []byte("IncrementDebt")
	KeyIncrementCollateral = []byte("IncrementCollateral")
)

var _ subspace.ParamSet = &Params{}

// Params is the governance parameters for the auction module.
type Params struct {
	MaxAuctionDuration  time.Duration `json:"max_auction_duration" yaml:"max_auction_duration"` // max length of auction
	BidDuration         time.Duration `json:"bid_duration" yaml:"bid_duration"`                 // additional time added to the auction end time after each bid, capped by the expiry.
	IncrementSurplus    sdk.Dec       `json:"increment_surplus" yaml:"increment_surplus"`       // percentage change (of auc.Bid) required for a new bid on a surplus auction
	IncrementDebt       sdk.Dec       `json:"increment_debt" yaml:"increment_debt"`             // percentage change (of auc.Lot) required for a new bid on a debt auction
	IncrementCollateral sdk.Dec       `json:"increment_collateral" yaml:"increment_collateral"` // percentage change (of auc.Bid or auc.Lot) required for a new bid on a collateral auction
}

// NewParams returns a new Params object.
func NewParams(maxAuctionDuration, bidDuration time.Duration, incrementSurplus, incrementDebt, incrementCollateral sdk.Dec) Params {
	return Params{
		MaxAuctionDuration:  maxAuctionDuration,
		BidDuration:         bidDuration,
		IncrementSurplus:    incrementSurplus,
		IncrementDebt:       incrementDebt,
		IncrementCollateral: incrementCollateral,
	}
}

// DefaultParams returns the default parameters for auctions.
func DefaultParams() Params {
	return NewParams(
		DefaultMaxAuctionDuration,
		DefaultBidDuration,
		DefaultIncrement,
		DefaultIncrement,
		DefaultIncrement,
	)
}

// ParamKeyTable Key declaration for parameters
func ParamKeyTable() subspace.KeyTable {
	return subspace.NewKeyTable().RegisterParamSet(&Params{})
}

// ParamSetPairs implements the ParamSet interface and returns all the key/value pairs.
func (p *Params) ParamSetPairs() subspace.ParamSetPairs {
	// TODO: Write validation functions
	return subspace.ParamSetPairs{
		params.NewParamSetPair(KeyBidDuration, &p.BidDuration, validateFn),
		params.NewParamSetPair(KeyMaxAuctionDuration, &p.MaxAuctionDuration, validateFn),
		params.NewParamSetPair(KeyIncrementSurplus, &p.IncrementSurplus, validateFn),
		params.NewParamSetPair(KeyIncrementDebt, &p.IncrementDebt, validateFn),
		params.NewParamSetPair(KeyIncrementCollateral, &p.IncrementCollateral, validateFn),
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
	Bid Duration: %s
	Increment Surplus: %s
	Increment Debt: %s
	Increment Collateral: %s`,
		p.MaxAuctionDuration, p.BidDuration, p.IncrementSurplus, p.IncrementDebt, p.IncrementCollateral)
}

// Validate checks that the parameters have valid values.
func (p Params) Validate() error {
	if p.BidDuration < 0 {
		return errors.New("bid duration cannot be negative")
	}
	if p.MaxAuctionDuration < 0 {
		return errors.New("max auction duration cannot be negative")
	}
	if p.BidDuration > p.MaxAuctionDuration {
		return errors.New("bid duration param cannot be larger than max auction duration")
	}
	if p.IncrementSurplus == (sdk.Dec{}) || p.IncrementDebt == (sdk.Dec{}) || p.IncrementCollateral == (sdk.Dec{}) {
		return errors.New("auction increment values cannot be nil")
	}
	if p.IncrementSurplus.IsNegative() {
		return errors.New("surplus auction increment cannot be less than zero")
	}
	if p.IncrementDebt.IsNegative() {
		return errors.New("debt auction increment cannot be less than zero")
	}
	if p.IncrementCollateral.IsNegative() {
		return errors.New("collateral auction increment cannot be less than zero")
	}
	return nil
}

func validateFn(i interface{}) error {
	return nil
}
