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
	KeyBnbDeputyAddress  = []byte("BnbDeputyAddress")
	KeyBnbDeputyFixedFee = []byte("BnbDeputyFixedFee")
	KeyMinAmount         = []byte("MinAmount")
	KeyMaxAmount         = []byte("MaxAmount")
	KeyMinBlockLock      = []byte("MinBlockLock")
	KeyMaxBlockLock      = []byte("MaxBlockLock")
	KeySupportedAssets   = []byte("SupportedAssets")

	DefaultBnbDeputyFixedFee uint64 = 1000 // 0.00001 BNB
	DefaultMinAmount         uint64
	DefaultMaxAmount         uint64 = 1000000000000 // 10,000 BNB
	AbsoluteMaximumBlockLock uint64 = 10000
	AbsoluteMinimumBlockLock uint64 = 50
	DefaultMinBlockLock      uint64 = 80
	DefaultMaxBlockLock      uint64 = 600
	DefaultSupportedAssets          = AssetParams{
		AssetParam{
			Denom:  "bnb",
			CoinID: 714,
			Limit:  sdk.NewInt(350000000000000), // 3,500,000 BNB
			Active: true,
		},
	}
)

// Params governance parameters for bep3 module
type Params struct {
	BnbDeputyAddress  sdk.AccAddress `json:"bnb_deputy_address" yaml:"bnb_deputy_address"`     // Bnbchain deputy address
	BnbDeputyFixedFee uint64         `json:"bnb_deputy_fixed_fee" yaml:"bnb_deputy_fixed_fee"` // Deputy fixed fee in BNB
	MinAmount         uint64         `json:"min_amount" yaml:"min_amount"`                     // Minimum swap amount
	MaxAmount         uint64         `json:"max_amount" yaml:"max_amount"`                     // Maximum swap amount
	MinBlockLock      uint64         `json:"min_block_lock" yaml:"min_block_lock"`             // Minimum swap block lock
	MaxBlockLock      uint64         `json:"max_block_lock" yaml:"max_block_lock"`             // Maximum swap block lock
	SupportedAssets   AssetParams    `json:"supported_assets" yaml:"supported_assets"`         // Supported assets
}

// String implements fmt.Stringer
func (p Params) String() string {
	return fmt.Sprintf(`Params:
	Bnbchain deputy address: %s,
	Deputy fixed fee (BNB): %d,
	Min amount: %d,
	Max amount: %d,
	Min block lock: %d,
	Max block lock: %d,
	Supported assets: %s`,
		p.BnbDeputyAddress.String(), p.BnbDeputyFixedFee, p.MinAmount,
		p.MaxAmount, p.MinBlockLock, p.MaxBlockLock, p.SupportedAssets)
}

// NewParams returns a new params object
func NewParams(bnbDeputyAddress sdk.AccAddress, bnbDeputyFixedFee, minAmount, maxAmount,
	minBlockLock, maxBlockLock uint64, supportedAssets AssetParams,
) Params {
	return Params{
		BnbDeputyAddress:  bnbDeputyAddress,
		BnbDeputyFixedFee: bnbDeputyFixedFee,
		MinAmount:         minAmount,
		MaxAmount:         maxAmount,
		MinBlockLock:      minBlockLock,
		MaxBlockLock:      maxBlockLock,
		SupportedAssets:   supportedAssets,
	}
}

// DefaultParams returns default params for bep3 module
func DefaultParams() Params {
	defaultBnbDeputyAddress, err := sdk.AccAddressFromBech32("kava1xy7hrjy9r0algz9w3gzm8u6mrpq97kwta747gj")
	if err != nil {
		panic(err)
	}

	return NewParams(defaultBnbDeputyAddress, DefaultBnbDeputyFixedFee, DefaultMinAmount,
		DefaultMaxAmount, DefaultMinBlockLock, DefaultMaxBlockLock, DefaultSupportedAssets)
}

// AssetParam governance parameters for each asset within a supported chain
type AssetParam struct {
	Denom  string  `json:"denom" yaml:"denom"`     // name of the asset
	CoinID int     `json:"coin_id" yaml:"coin_id"` // internationally recognized coin ID
	Limit  sdk.Int `json:"limit" yaml:"limit"`     // asset supply limit
	Active bool    `json:"active" yaml:"active"`   // denotes if asset is available or paused
}

// String implements fmt.Stringer
func (ap AssetParam) String() string {
	return fmt.Sprintf(`Asset:
	Denom: %s
	Coin ID: %d
	Limit: %s
	Active: %t`,
		ap.Denom, ap.CoinID, ap.Limit.String(), ap.Active)
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
		params.NewParamSetPair(KeyBnbDeputyAddress, &p.BnbDeputyAddress, validateBnbDeputyAddressParam),
		params.NewParamSetPair(KeyBnbDeputyFixedFee, &p.BnbDeputyFixedFee, validateBnbDeputyFixedFeeParam),
		params.NewParamSetPair(KeyMinAmount, &p.MinAmount, validateMinAmountParam),
		params.NewParamSetPair(KeyMaxAmount, &p.MaxAmount, validateMaxAmountParam),
		params.NewParamSetPair(KeyMinBlockLock, &p.MinBlockLock, validateMinBlockLockParam),
		params.NewParamSetPair(KeyMaxBlockLock, &p.MaxBlockLock, validateMaxBlockLockParam),
		params.NewParamSetPair(KeySupportedAssets, &p.SupportedAssets, validateSupportedAssetsParams),
	}
}

// Validate ensure that params have valid values
func (p Params) Validate() error {
	if err := validateBnbDeputyAddressParam(p.BnbDeputyAddress); err != nil {
		return err
	}

	if err := validateBnbDeputyFixedFeeParam(p.BnbDeputyFixedFee); err != nil {
		return err
	}

	if err := validateMinAmountParam(p.MinAmount); err != nil {
		return err
	}

	if err := validateMaxAmountParam(p.MaxAmount); err != nil {
		return err
	}

	if p.MinAmount > p.MaxAmount {
		return fmt.Errorf("minimum amount cannot be > maximum amount, got %d > %d", p.MinAmount, p.MaxAmount)
	}

	if err := validateMinBlockLockParam(p.MinBlockLock); err != nil {
		return err
	}

	if err := validateMaxBlockLockParam(p.MaxBlockLock); err != nil {
		return err
	}

	if p.MinBlockLock >= p.MaxBlockLock {
		return fmt.Errorf("minimum block lock cannot be ≥ maximum block lock, got %d ≥ %d", p.MinBlockLock, p.MaxBlockLock)
	}

	return validateSupportedAssetsParams(p.SupportedAssets)
}

func validateBnbDeputyAddressParam(i interface{}) error {
	addr, ok := i.(sdk.AccAddress)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if addr.Empty() {
		return errors.New("bnb deputy address cannot be empty")
	}

	if len(addr.Bytes()) != sdk.AddrLen {
		return fmt.Errorf("bnb deputy address invalid bytes length got %d, want %d", len(addr.Bytes()), sdk.AddrLen)
	}

	return nil
}

func validateBnbDeputyFixedFeeParam(i interface{}) error {
	_, ok := i.(uint64)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	return nil
}

func validateMinAmountParam(i interface{}) error {
	_, ok := i.(uint64)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	return nil
}

func validateMaxAmountParam(i interface{}) error {
	_, ok := i.(uint64)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	return nil
}

func validateMinBlockLockParam(i interface{}) error {
	minBlockLock, ok := i.(uint64)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if minBlockLock < AbsoluteMinimumBlockLock {
		return fmt.Errorf("minimum block lock cannot be less than %d, got %d", AbsoluteMinimumBlockLock, minBlockLock)
	}

	return nil
}

func validateMaxBlockLockParam(i interface{}) error {
	maxBlockLock, ok := i.(uint64)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if maxBlockLock > AbsoluteMaximumBlockLock {
		return fmt.Errorf("maximum block lock cannot be greater than %d, got %d", AbsoluteMaximumBlockLock, maxBlockLock)
	}

	return nil
}

func validateSupportedAssetsParams(i interface{}) error {
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
	}

	return nil
}
