package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/params/subspace"
)

/*
How this uses the sdk params module:
 - Put all the params for this module in one struct `CDPModuleParams`
 - Store this in the keeper's paramSubspace under one key
 - Provide a function to load the param struct all at once `keeper.GetParams(ctx)`
It's possible to set individual key value pairs within a paramSubspace, but reading and setting them is awkward (an empty variable needs to be created, then Get writes the value into it)
This approach will be awkward if we ever need to write individual parameters (because they're stored all together). If this happens do as the sdk modules do - store parameters separately with custom get/set func for each.
*/

// CdpParams governance parameters for cdp module
type CdpParams struct {
	GlobalDebtLimit  sdk.Int
	CollateralParams []CollateralParams
	StableDenoms     []string
}

// CollateralParams governance parameters for each collateral type within the cdp module
type CollateralParams struct {
	Denom            string  // Coin name of collateral type
	LiquidationRatio sdk.Dec // The ratio (Collateral (priced in stable coin) / Debt) under which a CDP will be liquidated
	DebtLimit        sdk.Int // Maximum amount of debt allowed to be drawn from this collateral type
	//DebtFloor        sdk.Int // used to prevent dust
}

// Parameter keys
var (
	// ParamStoreKeyAuctionParams Param store key for auction params
	KeyGlobalDebtLimit  = []byte("GlobalDebtLimit")
	KeyCollateralParams = []byte("CollateralParams")
	KeyStableDenoms     = []byte("StableDenoms")
)

// ParamKeyTable Key declaration for parameters
func ParamKeyTable() subspace.KeyTable {
	return subspace.NewKeyTable().RegisterParamSet(&CdpParams{})
}

// ParamSetPairs implements the ParamSet interface and returns all the key/value pairs
// pairs of auth module's parameters.
// nolint
func (p *CdpParams) ParamSetPairs() subspace.ParamSetPairs {
	return subspace.ParamSetPairs{
		{KeyGlobalDebtLimit, &p.GlobalDebtLimit},
		{KeyCollateralParams, &p.CollateralParams},
		{KeyStableDenoms, &p.StableDenoms},
	}
}

// String implements fmt.Stringer
func (p CdpParams) String() string {
	out := fmt.Sprintf(`Params:
	Global Debt Limit: %s
	Collateral Params:`,
		p.GlobalDebtLimit,
	)
	for _, cp := range p.CollateralParams {
		out += fmt.Sprintf(`
		%s
			Liquidation Ratio: %s
			Debt Limit:        %s`,
			cp.Denom,
			cp.LiquidationRatio,
			cp.DebtLimit,
		)
	}
	return out
}

// GetCollateralParams returns params for a specific collateral denom
func (p CdpParams) GetCollateralParams(collateralDenom string) CollateralParams {
	// search for matching denom, return
	for _, cp := range p.CollateralParams {
		if cp.Denom == collateralDenom {
			return cp
		}
	}
	// panic if not found, to be safe
	panic("collateral params not found in module params")
}

// IsCollateralPresent returns true if the denom is among the collaterals in cdp module
func (p CdpParams) IsCollateralPresent(collateralDenom string) bool {
	// search for matching denom, return
	for _, cp := range p.CollateralParams {
		if cp.Denom == collateralDenom {
			return true
		}
	}
	return false
}

// Validate checks that the parameters have valid values.
func (p CdpParams) Validate() error {
	collateralDupMap := make(map[string]int)
	denomDupMap := make(map[string]int)
	for _, collateral := range p.CollateralParams {
		_, found := collateralDupMap[collateral.Denom]
		if found {
			return fmt.Errorf("duplicate denom: %s", collateral.Denom)
		}
		collateralDupMap[collateral.Denom] = 1

		if collateral.DebtLimit.IsNegative() {
			return fmt.Errorf("debt limit should be positive, is %s for %s", collateral.DebtLimit, collateral.Denom)
		}

		// TODO do we want to enforce overcollateralization at this level? -- probably not, as it's technically a governance thing (kevin)
	}
	if p.GlobalDebtLimit.IsNegative() {
		return fmt.Errorf("global debt limit should be positive, is %s", p.GlobalDebtLimit)
	}

	for _, denom := range p.StableDenoms {
		_, found := denomDupMap[denom]
		if found {
			return fmt.Errorf("duplicate stable denom: %s", denom)
		}
		denomDupMap[denom] = 1
	}
	return nil
}

func DefaultParams() CdpParams {
	return CdpParams{
		GlobalDebtLimit:  sdk.NewInt(0),
		CollateralParams: []CollateralParams{},
		StableDenoms:     []string{"usdx"},
	}
}
