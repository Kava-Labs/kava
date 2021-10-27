package types

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"

	tmtime "github.com/tendermint/tendermint/types/time"
)

const (
	bech32MainPrefix = "kava"
)

// Parameter keys
var (
	KeyAssetParams = []byte("AssetParams")

	DefaultBnbDeputyFixedFee sdk.Int = sdk.NewInt(1000) // 0.00001 BNB
	DefaultMinAmount         sdk.Int = sdk.ZeroInt()
	DefaultMaxAmount         sdk.Int = sdk.NewInt(1000000000000) // 10,000 BNB
	DefaultMinBlockLock      uint64  = 220
	DefaultMaxBlockLock      uint64  = 270
	DefaultPreviousBlockTime         = tmtime.Canonical(time.Unix(1, 0))
)

// String implements fmt.Stringer
func (p Params) String() string {
	return fmt.Sprintf(`Params:
	AssetParams: %s`,
		p.AssetParams)
}

// NewParams returns a new params object
func NewParams(ap []AssetParam) Params {
	return Params{
		AssetParams: ap,
	}
}

// DefaultParams returns default params for bep3 module
func DefaultParams() Params {
	return NewParams([]AssetParam{})
}

// NewAssetParam returns a new AssetParam
func NewAssetParam(
	denom string, coinID int64, limit SupplyLimit, active bool,
	deputyAddr string, fixedFee sdk.Int, minSwapAmount sdk.Int,
	maxSwapAmount sdk.Int, minBlockLock uint64, maxBlockLock uint64,
) AssetParam {
	return AssetParam{
		Denom:         denom,
		CoinID:        coinID,
		SupplyLimit:   limit,
		Active:        active,
		DeputyAddress: deputyAddr,
		FixedFee:      fixedFee,
		MinSwapAmount: minSwapAmount,
		MaxSwapAmount: maxSwapAmount,
		MinBlockLock:  minBlockLock,
		MaxBlockLock:  maxBlockLock,
	}
}

// String implements fmt.Stringer
func (ap AssetParam) String() string {
	return fmt.Sprintf(`Asset:
	Denom: %s
	Coin ID: %d
	Limit: %s
	Active: %t
	Deputy Address: %s
	Fixed Fee: %s
	Min Swap Amount: %s
	Max Swap Amount: %s
	Min Block Lock: %d
	Max Block Lock: %d`,
		ap.Denom, ap.CoinID, ap.SupplyLimit, ap.Active, ap.DeputyAddress, ap.FixedFee,
		ap.MinSwapAmount, ap.MaxSwapAmount, ap.MinBlockLock, ap.MaxBlockLock)
}

// String implements fmt.Stringer
func (sl SupplyLimit) String() string {
	return fmt.Sprintf(`%s
	%t
	%s
	%s
	`, sl.Limit, sl.TimeLimited, sl.TimePeriod, sl.TimeBasedLimit)
}

// Equals returns true if two supply limits are equal
func (sl SupplyLimit) Equals(sl2 SupplyLimit) bool {
	return sl.Limit.Equal(sl2.Limit) && sl.TimeLimited == sl2.TimeLimited && sl.TimePeriod == sl2.TimePeriod && sl.TimeBasedLimit.Equal(sl2.TimeBasedLimit)
}

// ParamKeyTable Key declaration for parameters
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// ParamSetPairs implements the ParamSet interface and returns all the key/value pairs
// pairs of bep3 module's parameters.
// nolint
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyAssetParams, &p.AssetParams, validateAssetParams),
	}
}

// Validate ensure that params have valid values
func (p Params) Validate() error {
	return validateAssetParams(p.AssetParams)
}

func validateAssetParams(i interface{}) error {
	assetParams, ok := i.([]AssetParam)
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

		if len(asset.DeputyAddress) == 0 {
			return fmt.Errorf("deputy address cannot be empty for %s", asset.Denom)
		}

		// Verify address format
		if _, err := sdk.AccAddressFromBech32(asset.DeputyAddress); err != nil {
			return fmt.Errorf("%s deputy address invalid %w", asset.Denom, err)
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
