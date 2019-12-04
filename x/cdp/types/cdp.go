package types

import (
	"fmt"
	"strings"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// CDP is the state of a single Collateralized Debt Position.
type CDP struct {
	ID               uint64         `json:"id" yaml:"id"`                               // unique id for cdp
	Owner            sdk.AccAddress `json:"owner" yaml:"owner"`                         // Account that authorizes changes to the CDP
	CollateralAmount sdk.Coins      `json:"collateral_amount" yaml:"collateral_amount"` // Amount of collateral stored in this CDP
	Debt             sdk.Coins      `json:"debt" yaml:"debt"`
	AccumulatedFees  sdk.Coins      `json:"accumulated_fees" yaml:"accumulated_fees"`
	FeesUpdated      time.Time      `json:"fees_updated" yaml:"fees_updated"` // Amount of stable coin drawn from this CDP
}

func NewCDP(ID uint64)

// String implements fmt.stringer
func (cdp CDP) String() string {
	return strings.TrimSpace(fmt.Sprintf(`CDP:
	Owner:      %s
	ID: %d
	Collateral Type: %s
	Collateral: %s
	Debt: %s
	Fees: %s
	Fees Last Updated: %s`,
		cdp.Owner,
		cdp.ID,
		cdp.CollateralAmount[0].Denom,
		cdp.CollateralAmount,
		cdp.Debt,
		cdp.AccumulatedFees,
		cdp.FeesUpdated,
	))
}

// CDPs array of CDP
type CDPs []CDP

// String implements stringer
func (cdps CDPs) String() string {
	out := ""
	for _, cdp := range cdps {
		out += cdp.String() + "\n"
	}
	return out
}
