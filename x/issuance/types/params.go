package types

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/params"
)

// Parameter keys and default values
var (
	KeyAssets         = []byte("Assets")
	DefaultAssets     = Assets{}
	ModuleAccountName = ModuleName
)

// Params governance parameters for the issuance module
type Params struct {
	Assets Assets `json:"assets" yaml:"assets"`
}

// NewParams returns a new params object
func NewParams(assets Assets) Params {
	return Params{Assets: assets}
}

// DefaultParams returns default params for issuance module
func DefaultParams() Params {
	return NewParams(DefaultAssets)
}

// ParamKeyTable Key declaration for parameters
func ParamKeyTable() params.KeyTable {
	return params.NewKeyTable().RegisterParamSet(&Params{})
}

// ParamSetPairs implements the ParamSet interface and returns all the key/value pairs
func (p *Params) ParamSetPairs() params.ParamSetPairs {
	return params.ParamSetPairs{
		params.NewParamSetPair(KeyAssets, &p.Assets, validateAssetsParam),
	}
}

// Validate checks that the parameters have valid values.
func (p Params) Validate() error {
	return validateAssetsParam(p.Assets)
}

func validateAssetsParam(i interface{}) error {
	assets, ok := i.(Assets)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	return assets.Validate()
}

// String implements fmt.Stringer
func (p Params) String() string {
	return fmt.Sprintf(`Params:
	Assets: %s
	`, p.Assets)
}

// Asset type for assets in the issuance module
type Asset struct {
	Owner            sdk.AccAddress   `json:"owner" yaml:"owner"`
	Denom            string           `json:"denom" yaml:"denom"`
	BlockedAddresses []sdk.AccAddress `json:"blocked_addresses" yaml:"blocked_addresses"`
	Paused           bool             `json:"paused" yaml:"paused"`
	Blockable        bool             `json:"blockable" yaml:"blockable"`
	RateLimit        RateLimit        `json:"rate_limit" yaml:"rate_limit"`
}

// NewAsset returns a new Asset
func NewAsset(owner sdk.AccAddress, denom string, blockedAddresses []sdk.AccAddress, paused bool, blockable bool, limit RateLimit) Asset {
	return Asset{
		Owner:            owner,
		Denom:            denom,
		BlockedAddresses: blockedAddresses,
		Paused:           paused,
		Blockable:        blockable,
		RateLimit:        limit,
	}
}

// Validate performs a basic check of asset fields
func (a Asset) Validate() error {
	if a.Owner.Empty() {
		return fmt.Errorf("owner must not be empty")
	}
	if !a.Blockable && len(a.BlockedAddresses) > 0 {
		return fmt.Errorf("asset %s does not support blocking, blocked-list should be empty: %s", a.Denom, a.BlockedAddresses)
	}
	for _, address := range a.BlockedAddresses {
		if address.Empty() {
			return fmt.Errorf("blocked address must not be empty")
		}
		if a.Owner.Equals(address) {
			return fmt.Errorf("asset owner cannot be blocked")
		}
	}
	return sdk.ValidateDenom(a.Denom)
}

// String implements fmt.Stringer
func (a Asset) String() string {
	return fmt.Sprintf(`Asset:
	Owner: %s
	Paused: %t
	Denom: %s
	Blocked Addresses: %s
	Rate limits: %s`,
		a.Owner, a.Paused, a.Denom, a.BlockedAddresses, a.RateLimit)
}

// Assets slice of Asset
type Assets []Asset

// Validate checks if all assets are valid and there are no duplicate entries
func (as Assets) Validate() error {
	assetDenoms := make(map[string]bool)
	for _, a := range as {
		if assetDenoms[a.Denom] {
			return fmt.Errorf("cannot have duplicate asset denoms: %s", a.Denom)
		}
		if err := a.Validate(); err != nil {
			return err
		}
		assetDenoms[a.Denom] = true
	}
	return nil
}

// String implements fmt.Stringer
func (as Assets) String() string {
	out := ""
	for _, a := range as {
		out += a.String()
	}
	return out
}

// RateLimit parameters for rate-limiting the supply of an issued asset
type RateLimit struct {
	Active     bool          `json:"active" yaml:"active"`
	Limit      sdk.Int       `json:"limit" yaml:"limit"`
	TimePeriod time.Duration `json:"time_period" yaml:"time_period"`
}

// NewRateLimit initializes a new RateLimit
func NewRateLimit(active bool, limit sdk.Int, timePeriod time.Duration) RateLimit {
	return RateLimit{
		Active:     active,
		Limit:      limit,
		TimePeriod: timePeriod,
	}
}

// String implements fmt.Stringer
func (r RateLimit) String() string {
	return fmt.Sprintf(`
	Active: %t
	Limit: %s
	Time Period: %s`,
		r.Active, r.Limit, r.TimePeriod)
}
