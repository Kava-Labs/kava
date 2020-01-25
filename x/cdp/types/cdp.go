package types

import (
	"fmt"
	"strings"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// CDP is the state of a single collateralized debt position.
type CDP struct {
	ID              uint64         `json:"id" yaml:"id"`                 // unique id for cdp
	Owner           sdk.AccAddress `json:"owner" yaml:"owner"`           // Account that authorizes changes to the CDP
	Collateral      sdk.Coins      `json:"collateral" yaml:"collateral"` // Amount of collateral stored in this CDP
	Principal       sdk.Coins      `json:"principal" yaml:"principal"`
	AccumulatedFees sdk.Coins      `json:"accumulated_fees" yaml:"accumulated_fees"`
	FeesUpdated     time.Time      `json:"fees_updated" yaml:"fees_updated"` // Amount of stable coin drawn from this CDP
}

// NewCDP creates a new CDP object
func NewCDP(id uint64, owner sdk.AccAddress, collateral sdk.Coins, principal sdk.Coins, time time.Time) CDP {
	var fees sdk.Coins
	return CDP{
		ID:              id,
		Owner:           owner,
		Collateral:      collateral,
		Principal:       principal,
		AccumulatedFees: fees,
		FeesUpdated:     time,
	}
}

// String implements fmt.stringer
func (cdp CDP) String() string {
	return strings.TrimSpace(fmt.Sprintf(`CDP:
	Owner:      %s
	ID: %d
	Collateral Type: %s
	Collateral: %s
	Principal: %s
	Fees: %s
	Fees Last Updated: %s`,
		cdp.Owner,
		cdp.ID,
		cdp.Collateral[0].Denom,
		cdp.Collateral,
		cdp.Principal,
		cdp.AccumulatedFees,
		cdp.FeesUpdated,
	))
}

// CDPs a collection of CDP objects
type CDPs []CDP

// String implements stringer
func (cdps CDPs) String() string {
	out := ""
	for _, cdp := range cdps {
		out += cdp.String() + "\n"
	}
	return out
}

// AugmentedCDP provides additional information about an active CDP
type AugmentedCDP struct {
	CDP                    `json:"cdp" yaml:"cdp"`
	CollateralValue        sdk.Dec `json:"collateral_value" yaml:"collateral_value"`               // collateral's market value (quantity * price)
	CollateralizationRatio sdk.Dec `json:"collateralization_ratio" yaml:"collateralization_ratio"` // current collateralization ratio
}

// NewAugmentedCDP creates a new AugmentedCDP object
func NewAugmentedCDP(cdp CDP, collateralValue sdk.Dec, collateralizationRatio sdk.Dec) AugmentedCDP {
	augmentedCDP := AugmentedCDP{
		CDP: CDP{
			ID:              cdp.ID,
			Owner:           cdp.Owner,
			Collateral:      cdp.Collateral,
			Principal:       cdp.Principal,
			AccumulatedFees: cdp.AccumulatedFees,
			FeesUpdated:     cdp.FeesUpdated,
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
	Collateralization ratio: %s`,
		augCDP.Owner,
		augCDP.ID,
		augCDP.Collateral[0].Denom,
		augCDP.Collateral,
		augCDP.CollateralValue,
		augCDP.Principal,
		augCDP.AccumulatedFees,
		augCDP.FeesUpdated,
		augCDP.CollateralizationRatio,
	))
}
