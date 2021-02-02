package v0_13

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const (
	// ModuleName The name that will be used throughout the module
	ModuleName = "issuance"

	// StoreKey Top level store key where all module items will be stored
	StoreKey = ModuleName

	// RouterKey Top level router key
	RouterKey = ModuleName

	// DefaultParamspace default name for parameter store
	DefaultParamspace = ModuleName

	// QuerierRoute route used for abci queries
	QuerierRoute = ModuleName
)

// Parameter keys and default values
var (
	KeyAssets            = []byte("Assets")
	DefaultAssets        = Assets{}
	ModuleAccountName    = ModuleName
	AssetSupplyPrefix    = []byte{0x01}
	PreviousBlockTimeKey = []byte{0x02}
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

// GenesisState is the state that must be provided at genesis for the issuance module
type GenesisState struct {
	Params   Params        `json:"params" yaml:"params"`
	Supplies AssetSupplies `json:"supplies" yaml:"supplies"`
}

// NewGenesisState returns a new GenesisState
func NewGenesisState(params Params, supplies AssetSupplies) GenesisState {
	return GenesisState{
		Params:   params,
		Supplies: supplies,
	}
}

// DefaultGenesisState returns the default GenesisState for the issuance module
func DefaultGenesisState() GenesisState {
	return GenesisState{
		Params:   DefaultParams(),
		Supplies: AssetSupplies{},
	}
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

// Assets array of Asset
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

// AssetSupply contains information about an asset's rate-limited supply (the total supply of the asset is tracked in the top-level supply module)
type AssetSupply struct {
	CurrentSupply sdk.Coin      `json:"current_supply"  yaml:"current_supply"`
	TimeElapsed   time.Duration `json:"time_elapsed" yaml:"time_elapsed"`
}

// NewAssetSupply initializes a new AssetSupply
func NewAssetSupply(currentSupply sdk.Coin, timeElapsed time.Duration) AssetSupply {
	return AssetSupply{
		CurrentSupply: currentSupply,
		TimeElapsed:   timeElapsed,
	}
}

// Validate performs a basic validation of an asset supply fields.
func (a AssetSupply) Validate() error {
	if !a.CurrentSupply.IsValid() {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidCoins, "outgoing supply %s", a.CurrentSupply)
	}
	return nil
}

// String implements stringer
func (a AssetSupply) String() string {
	return fmt.Sprintf(`
	asset supply:
		Current supply:     %s
		Time elapsed: %s
		`,
		a.CurrentSupply, a.TimeElapsed)
}

// GetDenom getter method for the denom of the asset supply
func (a AssetSupply) GetDenom() string {
	return a.CurrentSupply.Denom
}

// AssetSupplies is a slice of AssetSupply
type AssetSupplies []AssetSupply
