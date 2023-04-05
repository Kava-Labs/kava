package types

import (
	"errors"
	"fmt"
	"strings"
	"time"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// NewCDP creates a new CDP object
func NewCDP(id uint64, owner sdk.AccAddress, collateral sdk.Coin, collateralType string, principal sdk.Coin, time time.Time, interestFactor sdk.Dec) CDP {
	fees := sdk.NewCoin(principal.Denom, sdk.ZeroInt())
	return CDP{
		ID:              id,
		Owner:           owner,
		Type:            collateralType,
		Collateral:      collateral,
		Principal:       principal,
		AccumulatedFees: fees,
		FeesUpdated:     time,
		InterestFactor:  interestFactor,
	}
}

// NewCDPWithFees creates a new CDP object, for use during migration
func NewCDPWithFees(id uint64, owner sdk.AccAddress, collateral sdk.Coin, collateralType string, principal, fees sdk.Coin, time time.Time, interestFactor sdk.Dec) CDP {
	return CDP{
		ID:              id,
		Owner:           owner,
		Type:            collateralType,
		Collateral:      collateral,
		Principal:       principal,
		AccumulatedFees: fees,
		FeesUpdated:     time,
		InterestFactor:  interestFactor,
	}
}

// Validate performs a basic validation of the CDP fields.
func (cdp CDP) Validate() error {
	if cdp.ID == 0 {
		return errors.New("cdp id cannot be 0")
	}
	if cdp.Owner.Empty() {
		return errors.New("cdp owner cannot be empty")
	}
	if !cdp.Collateral.IsValid() {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidCoins, "collateral %s", cdp.Collateral)
	}
	if !cdp.Principal.IsValid() {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidCoins, "principal %s", cdp.Principal)
	}
	if !cdp.AccumulatedFees.IsValid() {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidCoins, "accumulated fees %s", cdp.AccumulatedFees)
	}
	if cdp.FeesUpdated.Unix() <= 0 {
		return errors.New("cdp updated fee time cannot be zero")
	}
	if strings.TrimSpace(cdp.Type) == "" {
		return fmt.Errorf("cdp type cannot be empty")
	}
	return nil
}

// GetTotalPrincipal returns the total principle for the cdp
func (cdp CDP) GetTotalPrincipal() sdk.Coin {
	return cdp.Principal.Add(cdp.AccumulatedFees)
}

// GetNormalizedPrincipal returns the total cdp principal divided by the interest factor.
//
// Multiplying the normalized principal by the current global factor gives the current debt (ie including all interest, ie a synced cdp).
// The normalized principal is effectively how big the principal would have been if it had been borrowed at time 0 and not touched since.
//
// An error is returned if the cdp interest factor is in an invalid state.
func (cdp CDP) GetNormalizedPrincipal() (sdk.Dec, error) {
	unsyncedDebt := cdp.GetTotalPrincipal().Amount
	if cdp.InterestFactor.LT(sdk.OneDec()) {
		return sdk.Dec{}, fmt.Errorf("interest factor '%s' must be ≥ 1", cdp.InterestFactor)
	}
	return sdk.NewDecFromInt(unsyncedDebt).Quo(cdp.InterestFactor), nil
}

// CDPs a collection of CDP objects
type CDPs []CDP

// Validate validates each CDP
func (cdps CDPs) Validate() error {
	for _, cdp := range cdps {
		if err := cdp.Validate(); err != nil {
			return err
		}
	}
	return nil
}

// AugmentedCDP provides additional information about an active CDP.
// This is only used for the legacy querier and legacy rest endpoints.
type AugmentedCDP struct {
	CDP                    `json:"cdp" yaml:"cdp"`
	CollateralValue        sdk.Coin `json:"collateral_value" yaml:"collateral_value"`               // collateral's market value in debt coin
	CollateralizationRatio sdk.Dec  `json:"collateralization_ratio" yaml:"collateralization_ratio"` // current collateralization ratio
}

// NewAugmentedCDP creates a new AugmentedCDP object
func NewAugmentedCDP(cdp CDP, collateralValue sdk.Coin, collateralizationRatio sdk.Dec) AugmentedCDP {
	augmentedCDP := AugmentedCDP{
		CDP: CDP{
			ID:              cdp.ID,
			Owner:           cdp.Owner,
			Type:            cdp.Type,
			Collateral:      cdp.Collateral,
			Principal:       cdp.Principal,
			AccumulatedFees: cdp.AccumulatedFees,
			FeesUpdated:     cdp.FeesUpdated,
			InterestFactor:  cdp.InterestFactor,
		},
		CollateralValue:        collateralValue,
		CollateralizationRatio: collateralizationRatio,
	}
	return augmentedCDP
}

// String implements fmt.stringer
func (augCDP AugmentedCDP) String() string {
	return strings.TrimSpace(fmt.Sprintf(`AugmentedCDP:
	Owner:      %s
	ID: %d
	Collateral Type: %s
	Collateral: %s
	Collateral Value: %s
	Principal: %s
	Fees: %s
	Fees Last Updated: %s
	Interest Factor: %s
	Collateralization ratio: %s`,
		augCDP.Owner,
		augCDP.ID,
		augCDP.Type,
		augCDP.Collateral,
		augCDP.CollateralValue,
		augCDP.Principal,
		augCDP.AccumulatedFees,
		augCDP.FeesUpdated,
		augCDP.InterestFactor,
		augCDP.CollateralizationRatio,
	))
}

// AugmentedCDPs a collection of AugmentedCDP objects
type AugmentedCDPs []AugmentedCDP

// String implements stringer
func (augcdps AugmentedCDPs) String() string {
	out := ""
	for _, augcdp := range augcdps {
		out += augcdp.String() + "\n"
	}
	return out
}

// NewCDPResponse creates a new CDPResponse object
func NewCDPResponse(cdp CDP, collateralValue sdk.Coin, collateralizationRatio sdk.Dec) CDPResponse {
	return CDPResponse{
		ID:                     cdp.ID,
		Owner:                  cdp.Owner.String(),
		Type:                   cdp.Type,
		Collateral:             cdp.Collateral,
		Principal:              cdp.Principal,
		AccumulatedFees:        cdp.AccumulatedFees,
		FeesUpdated:            cdp.FeesUpdated,
		InterestFactor:         cdp.InterestFactor.String(),
		CollateralValue:        collateralValue,
		CollateralizationRatio: collateralizationRatio.String(),
	}
}

// CDPResponses a collection of CDPResponse objects
type CDPResponses []CDPResponse

// TotalPrincipals a collection of TotalPrincipal objects
type TotalPrincipals []TotalPrincipal

// TotalPrincipal returns a new TotalPrincipal
func NewTotalPrincipal(collateralType string, amount sdk.Coin) TotalPrincipal {
	return TotalPrincipal{
		CollateralType: collateralType,
		Amount:         amount,
	}
}

// TotalCollaterals a collection of TotalCollateral objects
type TotalCollaterals []TotalCollateral

// TotalCollateral returns a new TotalCollateral
func NewTotalCollateral(collateralType string, amount sdk.Coin) TotalCollateral {
	return TotalCollateral{
		CollateralType: collateralType,
		Amount:         amount,
	}
}
