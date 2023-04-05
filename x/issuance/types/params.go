package types

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkmath "cosmossdk.io/math"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

// Parameter keys and default values
var (
	KeyAssets         = []byte("Assets")
	DefaultAssets     = []Asset{}
	ModuleAccountName = ModuleName
)

// NewParams returns a new params object
func NewParams(assets []Asset) Params {
	return Params{
		Assets: assets,
	}
}

// DefaultParams returns default params for issuance module
func DefaultParams() Params {
	return NewParams(DefaultAssets)
}

// ParamKeyTable Key declaration for parameters
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// ParamSetPairs implements the ParamSet interface and returns all the key/value pairs
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyAssets, &p.Assets, validateAssetsParam),
	}
}

// Validate checks that the parameters have valid values.
func (p Params) Validate() error {
	return validateAssetsParam(p.Assets)
}

func validateAssetsParam(i interface{}) error {
	assets, ok := i.([]Asset)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	return ValidateAssets(assets)
}

// String implements fmt.Stringer
func (p Params) String() string {
	return fmt.Sprintf(`Params:
	Assets: %s
	`, p.Assets)
}

// NewAsset returns a new Asset
func NewAsset(owner string, denom string, blockedAddresses []string, paused bool, blockable bool, limit RateLimit) Asset {
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
	if len(a.Owner) == 0 {
		return fmt.Errorf("owner must not be empty")
	}
	if !a.Blockable && len(a.BlockedAddresses) > 0 {
		return fmt.Errorf("asset %s does not support blocking, blocked-list should be empty: %s", a.Denom, a.BlockedAddresses)
	}
	for _, address := range a.BlockedAddresses {
		if len(address) == 0 {
			return fmt.Errorf("blocked address must not be empty")
		}
		if a.Owner == address {
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
		a.Owner, a.Paused, a.Denom, a.BlockedAddresses, a.RateLimit.String())
}

// Validate checks if all assets are valid and there are no duplicate entries
func ValidateAssets(as []Asset) error {
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

// NewRateLimit initializes a new RateLimit
func NewRateLimit(active bool, limit sdkmath.Int, timePeriod time.Duration) RateLimit {
	return RateLimit{
		Active:     active,
		Limit:      limit,
		TimePeriod: timePeriod,
	}
}
