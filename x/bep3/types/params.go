package types

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/params"

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

// Params governance parameters for bep3 module
type Params struct {
	AssetParams AssetParams `json:"asset_params" yaml:"asset_params"`
}

// String implements fmt.Stringer
func (p Params) String() string {
	return fmt.Sprintf(`Params:
	AssetParams: %s`,
		p.AssetParams)
}

// NewParams returns a new params object
func NewParams(ap AssetParams,
) Params {
	return Params{
		AssetParams: ap,
	}
}

// DefaultParams returns default params for bep3 module
func DefaultParams() Params {
	return NewParams(AssetParams{})
}

// AssetParam parameters that must be specified for each bep3 asset
type AssetParam struct {
	Denom         string         `json:"denom" yaml:"denom"`                     // name of the asset
	CoinID        int            `json:"coin_id" yaml:"coin_id"`                 // SLIP-0044 registered coin type - see https://github.com/satoshilabs/slips/blob/master/slip-0044.md
	SupplyLimit   SupplyLimit    `json:"supply_limit" yaml:"supply_limit"`       // asset supply limit
	Active        bool           `json:"active" yaml:"active"`                   // denotes if asset is available or paused
	DeputyAddress sdk.AccAddress `json:"deputy_address" yaml:"deputy_address"`   // the address of the relayer process
	FixedFee      sdk.Int        `json:"fixed_fee" yaml:"fixed_fee"`             // the fixed fee charged by the relayer process for outgoing swaps
	MinSwapAmount sdk.Int        `json:"min_swap_amount" yaml:"min_swap_amount"` // Minimum swap amount
	MaxSwapAmount sdk.Int        `json:"max_swap_amount" yaml:"max_swap_amount"` // Maximum swap amount
	MinBlockLock  uint64         `json:"min_block_lock" yaml:"min_block_lock"`   // Minimum swap block lock
	MaxBlockLock  uint64         `json:"max_block_lock" yaml:"max_block_lock"`   // Maximum swap block lock
}

// NewAssetParam returns a new AssetParam
func NewAssetParam(
	denom string, coinID int, limit SupplyLimit, active bool,
	deputyAddr sdk.AccAddress, fixedFee sdk.Int, minSwapAmount sdk.Int,
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

// AssetParams array of AssetParam
type AssetParams []AssetParam

// String implements fmt.Stringer
func (aps AssetParams) String() string {
	out := "Asset Params\n"
	for _, ap := range aps {
		out += fmt.Sprintf("%s\n", ap)
	}
	return out
}

// SupplyLimit parameters that control the absolute and time-based limits for an assets's supply
type SupplyLimit struct {
	Limit          sdk.Int       `json:"limit" yaml:"limit"`                       // the absolute supply limit for an asset
	TimeLimited    bool          `json:"time_limited" yaml:"time_limited"`         // boolean for if the supply is also limited by time
	TimePeriod     time.Duration `json:"time_period" yaml:"time_period"`           // the duration for which the supply time limit applies
	TimeBasedLimit sdk.Int       `json:"time_based_limit" yaml:"time_based_limit"` // the supply limit for an asset for each time period
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
func ParamKeyTable() params.KeyTable {
	return params.NewKeyTable().RegisterParamSet(&Params{})
}

// ParamSetPairs implements the ParamSet interface and returns all the key/value pairs
// pairs of bep3 module's parameters.
// nolint
func (p *Params) ParamSetPairs() params.ParamSetPairs {
	return params.ParamSetPairs{
		params.NewParamSetPair(KeyAssetParams, &p.AssetParams, validateAssetParams),
	}
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
