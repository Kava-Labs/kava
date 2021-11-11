package types

import (
	"errors"
	"fmt"
	"strings"
	"time"

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
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidCoins, "collateral %s", cdp.Collateral)
	}
	if !cdp.Principal.IsValid() {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidCoins, "principal %s", cdp.Principal)
	}
	if !cdp.AccumulatedFees.IsValid() {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidCoins, "accumulated fees %s", cdp.AccumulatedFees)
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
	return unsyncedDebt.ToDec().Quo(cdp.InterestFactor), nil
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

// AugmentedCDPs a collection of AugmentedCDP objects
type AugmentedCDPs []AugmentedCDP

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
