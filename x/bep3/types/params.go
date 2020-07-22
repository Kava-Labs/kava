package types

import (
	"errors"
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/params"
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
	Denom                string         `json:"denom" yaml:"denom"`                                     // name of the asset
	CoinID               int            `json:"coin_id" yaml:"coin_id"`                                 // SLIP-0044 registered coin type - see https://github.com/satoshilabs/slips/blob/master/slip-0044.md
	Limit                sdk.Int        `json:"limit" yaml:"limit"`                                     // asset supply limit
	Active               bool           `json:"active" yaml:"active"`                                   // denotes if asset is available or paused
	DeputyAddress        sdk.AccAddress `json:"deputy_address" yaml:"deputy_address"`                   // the address of the relayer process
	IncomingSwapFixedFee sdk.Int        `json:"incoming_swap_fixed_fee" yaml:"incoming_swap_fixed_fee"` // the fixed fee charged by the relayer process for incoming swaps
	MinSwapAmount        sdk.Int        `json:"min_swap_amount" yaml:"min_swap_amount"`                 // Minimum swap amount
	MaxSwapAmount        sdk.Int        `json:"max_swap_amount" yaml:"max_swap_amount"`                 // Maximum swap amount
	MinBlockLock         uint64         `json:"min_block_lock" yaml:"min_block_lock"`                   // Minimum swap block lock
	MaxBlockLock         uint64         `json:"max_block_lock" yaml:"max_block_lock"`                   // Maximum swap block lock
}

// NewAssetParam returns a new AssetParam
func NewAssetParam(
	denom string, coinID int, limit sdk.Int, active bool,
	deputyAddr sdk.AccAddress, incomingFee sdk.Int, minSwapAmount sdk.Int,
	maxSwapAmount sdk.Int, minBlockLock uint64, maxBlockLock uint64,
) AssetParam {
	return AssetParam{
		Denom:                denom,
		CoinID:               coinID,
		Limit:                limit,
		Active:               active,
		DeputyAddress:        deputyAddr,
		IncomingSwapFixedFee: incomingFee,
		MinSwapAmount:        minSwapAmount,
		MaxSwapAmount:        maxSwapAmount,
		MinBlockLock:         minBlockLock,
		MaxBlockLock:         maxBlockLock,
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
	Incoming Swap Fee: %s
	Min Swap Amount: %s
	Max Swap Amount: %s
	Min Block Lock: %d
	Max Block Lock: %d`,
		ap.Denom, ap.CoinID, ap.Limit.String(), ap.Active, ap.DeputyAddress, ap.IncomingSwapFixedFee,
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

	coinIDs := make(map[int]bool)
	coinDenoms := make(map[string]bool)
	for _, asset := range assetParams {
		if strings.TrimSpace(asset.Denom) == "" {
			return errors.New("asset denom cannot be empty")
		}

		if asset.CoinID < 0 {
			return fmt.Errorf(fmt.Sprintf("asset %s must be a non negative integer", asset.Denom))
		}

		if !asset.Limit.IsPositive() {
			return fmt.Errorf(fmt.Sprintf("asset %s must have a positive supply limit", asset.Denom))
		}

		_, found := coinDenoms[asset.Denom]
		if found {
			return fmt.Errorf(fmt.Sprintf("asset %s cannot have duplicate denom", asset.Denom))
		}

		coinDenoms[asset.Denom] = true

		_, found = coinIDs[asset.CoinID]
		if found {
			return fmt.Errorf(fmt.Sprintf("asset %s cannot have duplicate coin id %d", asset.Denom, asset.CoinID))
		}

		coinIDs[asset.CoinID] = true
		if asset.DeputyAddress.Empty() {
			return errors.New("bnb deputy address cannot be empty")
		}

		if len(asset.DeputyAddress.Bytes()) != sdk.AddrLen {
			return fmt.Errorf("bnb deputy address invalid bytes length got %d, want %d", len(asset.DeputyAddress.Bytes()), sdk.AddrLen)
		}

		if asset.IncomingSwapFixedFee.IsNegative() {
			return fmt.Errorf("asset %s cannot have negative incoming swap fee %s", asset.Denom, asset.IncomingSwapFixedFee)
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
