package types

import (
	"errors"
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

var emptyDec = sdk.Dec{}

// Defaults for auction params
const (
	// DefaultMaxAuctionDuration max length of auction
	DefaultMaxAuctionDuration time.Duration = 2 * 24 * time.Hour
	// DefaultForwardBidDuration how long an auction gets extended when someone bids for a forward auction
	DefaultForwardBidDuration time.Duration = 24 * time.Hour
	// DefaultReverseBidDuration how long an auction gets extended when someone bids for a reverse auction
	DefaultReverseBidDuration time.Duration = 1 * time.Hour
)

var (
	// DefaultIncrement is the smallest percent change a new bid must have from the old one
	DefaultIncrement sdk.Dec = sdk.MustNewDecFromStr("0.05")
	// ParamStoreKeyParams Param store key for auction params
	KeyForwardBidDuration  = []byte("ForwardBidDuration")
	KeyReverseBidDuration  = []byte("ReverseBidDuration")
	KeyMaxAuctionDuration  = []byte("MaxAuctionDuration")
	KeyIncrementSurplus    = []byte("IncrementSurplus")
	KeyIncrementDebt       = []byte("IncrementDebt")
	KeyIncrementCollateral = []byte("IncrementCollateral")
)

// NewParams returns a new Params object.
func NewParams(
	maxAuctionDuration, forwardBidDuration, reverseBidDuration time.Duration,
	incrementSurplus,
	incrementDebt,
	incrementCollateral sdk.Dec,
) Params {
	return Params{
		MaxAuctionDuration:  maxAuctionDuration,
		ForwardBidDuration:  forwardBidDuration,
		ReverseBidDuration:  reverseBidDuration,
		IncrementSurplus:    incrementSurplus,
		IncrementDebt:       incrementDebt,
		IncrementCollateral: incrementCollateral,
	}
}

// DefaultParams returns the default parameters for auctions.
func DefaultParams() Params {
	return NewParams(
		DefaultMaxAuctionDuration,
		DefaultForwardBidDuration,
		DefaultReverseBidDuration,
		DefaultIncrement,
		DefaultIncrement,
		DefaultIncrement,
	)
}

// ParamKeyTable Key declaration for parameters
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// ParamSetPairs implements the ParamSet interface and returns all the key/value pairs.
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyForwardBidDuration, &p.ForwardBidDuration, validateBidDurationParam),
		paramtypes.NewParamSetPair(KeyReverseBidDuration, &p.ReverseBidDuration, validateBidDurationParam),
		paramtypes.NewParamSetPair(KeyMaxAuctionDuration, &p.MaxAuctionDuration, validateMaxAuctionDurationParam),
		paramtypes.NewParamSetPair(KeyIncrementSurplus, &p.IncrementSurplus, validateIncrementSurplusParam),
		paramtypes.NewParamSetPair(KeyIncrementDebt, &p.IncrementDebt, validateIncrementDebtParam),
		paramtypes.NewParamSetPair(KeyIncrementCollateral, &p.IncrementCollateral, validateIncrementCollateralParam),
	}
}

// Validate checks that the parameters have valid values.
func (p Params) Validate() error {
	if err := validateBidDurationParam(p.ForwardBidDuration); err != nil {
		return err
	}

	if err := validateBidDurationParam(p.ReverseBidDuration); err != nil {
		return err
	}

	if err := validateMaxAuctionDurationParam(p.MaxAuctionDuration); err != nil {
		return err
	}

	if p.ForwardBidDuration > p.MaxAuctionDuration {
		return errors.New("forward bid duration param cannot be larger than max auction duration")
	}

	if p.ReverseBidDuration > p.MaxAuctionDuration {
		return errors.New("reverse bid duration param cannot be larger than max auction duration")
	}

	if err := validateIncrementSurplusParam(p.IncrementSurplus); err != nil {
		return err
	}

	if err := validateIncrementDebtParam(p.IncrementDebt); err != nil {
		return err
	}

	return validateIncrementCollateralParam(p.IncrementCollateral)
}

func validateBidDurationParam(i interface{}) error {
	bidDuration, ok := i.(time.Duration)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if bidDuration < 0 {
		return fmt.Errorf("bid duration cannot be negative %d", bidDuration)
	}

	return nil
}

func validateMaxAuctionDurationParam(i interface{}) error {
	maxAuctionDuration, ok := i.(time.Duration)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if maxAuctionDuration < 0 {
		return fmt.Errorf("max auction duration cannot be negative %d", maxAuctionDuration)
	}

	return nil
}

func validateIncrementSurplusParam(i interface{}) error {
	incrementSurplus, ok := i.(sdk.Dec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if incrementSurplus == emptyDec || incrementSurplus.IsNil() {
		return errors.New("surplus auction increment cannot be nil or empty")
	}

	if incrementSurplus.IsNegative() {
		return fmt.Errorf("surplus auction increment cannot be less than zero %s", incrementSurplus)
	}

	return nil
}

func validateIncrementDebtParam(i interface{}) error {
	incrementDebt, ok := i.(sdk.Dec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if incrementDebt == emptyDec || incrementDebt.IsNil() {
		return errors.New("debt auction increment cannot be nil or empty")
	}

	if incrementDebt.IsNegative() {
		return fmt.Errorf("debt auction increment cannot be less than zero %s", incrementDebt)
	}

	return nil
}

func validateIncrementCollateralParam(i interface{}) error {
	incrementCollateral, ok := i.(sdk.Dec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if incrementCollateral == emptyDec || incrementCollateral.IsNil() {
		return errors.New("collateral auction increment cannot be nil or empty")
	}

	if incrementCollateral.IsNegative() {
		return fmt.Errorf("collateral auction increment cannot be less than zero %s", incrementCollateral)
	}

	return nil
}
