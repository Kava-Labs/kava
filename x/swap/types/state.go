package types

import (
	"errors"
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const PoolKeySep = "/"

// PoolIDFromCoins returns a poolID from a coins object
func PoolIDFromCoins(coins sdk.Coins) string {
	return PoolID(coins[0].Denom, coins[1].Denom)
}

// PoolID returns an alphabetically sorted pool name from two denoms.
// The name is commutative for any all pairs A,B: f(A,B) == f(B,A).
func PoolID(denomA string, denomB string) string {
	if denomB < denomA {
		return fmt.Sprintf("%s%s%s", denomB, PoolKeySep, denomA)
	}

	return fmt.Sprintf("%s%s%s", denomA, PoolKeySep, denomB)
}

// PoolRecord represents the state of a liquidity pool
// and is used to store the state of a denominated pool
type PoolRecord struct {
	// primary key
	PoolID      string   `json:"pool_id" yaml:"pool_id"`
	ReservesA   sdk.Coin `json:"reserves_a" yaml:"reserves_a"`
	ReservesB   sdk.Coin `json:"reserves_b" yaml:"reserves_b"`
	TotalShares sdk.Int  `json:"total_shares" yaml:"total_shares"`
}

// NewPoolRecord takes reserve coins and total shares, returning
// a new pool record with a id
func NewPoolRecord(reserves sdk.Coins, totalShares sdk.Int) PoolRecord {
	if len(reserves) != 2 {
		panic("reserves must have two denominations")
	}

	poolID := PoolIDFromCoins(reserves)

	return PoolRecord{
		PoolID:      poolID,
		ReservesA:   reserves[0],
		ReservesB:   reserves[1],
		TotalShares: totalShares,
	}
}

// NewPoolRecordFromPool takes a pointer to a denominated pool and returns a
// pool record for storage in state.
func NewPoolRecordFromPool(pool *DenominatedPool) PoolRecord {
	reserves := pool.Reserves()
	poolID := PoolIDFromCoins(reserves)

	return PoolRecord{
		PoolID:      poolID,
		ReservesA:   reserves[0],
		ReservesB:   reserves[1],
		TotalShares: pool.TotalShares(),
	}
}

// Validate performs basic validation checks of the record data
func (p PoolRecord) Validate() error {
	if p.PoolID == "" {
		return errors.New("poolID must be set")
	}

	tokens := strings.Split(p.PoolID, "/")
	if len(tokens) != 2 || tokens[0] == "" || tokens[1] == "" || tokens[1] < tokens[0] || tokens[0] == tokens[1] {
		return fmt.Errorf("poolID '%s' is invalid", p.PoolID)
	}
	if sdk.ValidateDenom(tokens[0]) != nil || sdk.ValidateDenom(tokens[1]) != nil {
		return fmt.Errorf("poolID '%s' is invalid", p.PoolID)
	}
	if tokens[0] != p.ReservesA.Denom || tokens[1] != p.ReservesB.Denom {
		return fmt.Errorf("poolID '%s' does not match reserves", p.PoolID)
	}

	if !p.ReservesA.IsPositive() {
		return fmt.Errorf("pool '%s' has invalid reserves: %s", p.PoolID, p.ReservesA)
	}

	if !p.ReservesB.IsPositive() {
		return fmt.Errorf("pool '%s' has invalid reserves: %s", p.PoolID, p.ReservesB)
	}

	if !p.TotalShares.IsPositive() {
		return fmt.Errorf("pool '%s' has invalid total shares: %s", p.PoolID, p.TotalShares)
	}

	return nil
}

// Reserves returns the total reserves for a pool
func (p PoolRecord) Reserves() sdk.Coins {
	return sdk.NewCoins(p.ReservesA, p.ReservesB)
}

// PoolRecords is a slice of PoolRecord
type PoolRecords []PoolRecord

// Validate performs basic validation checks on all records in the slice
func (prs PoolRecords) Validate() error {
	seenPoolIDs := make(map[string]bool)

	for _, p := range prs {
		if err := p.Validate(); err != nil {
			return err
		}

		if seenPoolIDs[p.PoolID] {
			return fmt.Errorf("duplicate poolID '%s'", p.PoolID)
		}

		seenPoolIDs[p.PoolID] = true
	}

	return nil
}

// ShareRecord stores the shares owned for a depositor and pool
type ShareRecord struct {
	// primary key
	Depositor sdk.AccAddress `json:"depositor" yaml:"depositor"`
	// secondary / sort key
	PoolID      string  `json:"pool_id" yaml:"pool_id"`
	SharesOwned sdk.Int `json:"shares_owned" yaml:"shares_owned"`
}

// NewShareRecord takes a depositor, poolID, and shares and returns
// a new share record for storage in state.
func NewShareRecord(depositor sdk.AccAddress, poolID string, sharesOwned sdk.Int) ShareRecord {
	return ShareRecord{
		Depositor:   depositor,
		PoolID:      poolID,
		SharesOwned: sharesOwned,
	}
}

// Validate performs basic validation checks of the record data
func (sr ShareRecord) Validate() error {
	if sr.PoolID == "" {
		return errors.New("poolID must be set")
	}

	tokens := strings.Split(sr.PoolID, "/")
	if len(tokens) != 2 || tokens[0] == "" || tokens[1] == "" || tokens[1] < tokens[0] || tokens[0] == tokens[1] {
		return fmt.Errorf("poolID '%s' is invalid", sr.PoolID)
	}
	if sdk.ValidateDenom(tokens[0]) != nil || sdk.ValidateDenom(tokens[1]) != nil {
		return fmt.Errorf("poolID '%s' is invalid", sr.PoolID)
	}

	if sr.Depositor.Empty() {
		return fmt.Errorf("share record cannot have empty depositor address")
	}

	if !sr.SharesOwned.IsPositive() {
		return fmt.Errorf("depositor '%s' and pool '%s' has invalid total shares: %s", sr.Depositor.String(), sr.PoolID, sr.SharesOwned.String())
	}

	return nil
}

// ShareRecords is a slice of ShareRecord
type ShareRecords []ShareRecord

// Validate performs basic validation checks on all records in the slice
func (srs ShareRecords) Validate() error {
	seenDepositors := make(map[string]map[string]bool)

	for _, sr := range srs {
		if err := sr.Validate(); err != nil {
			return err
		}

		if seenPools, found := seenDepositors[sr.Depositor.String()]; found {
			if seenPools[sr.PoolID] {
				return fmt.Errorf("duplicate depositor '%s' and poolID '%s'", sr.Depositor.String(), sr.PoolID)
			}
			seenPools[sr.PoolID] = true
		} else {
			seenPools := make(map[string]bool)
			seenPools[sr.PoolID] = true
			seenDepositors[sr.Depositor.String()] = seenPools
		}
	}

	return nil
}
