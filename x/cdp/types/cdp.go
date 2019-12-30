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
