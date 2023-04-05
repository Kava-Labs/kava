package types

import (
	"errors"
	"fmt"
	"strings"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// PoolIDSep represents the separator used in pool ids to separate two denominations
const PoolIDSep = ":"

// PoolIDFromCoins returns a poolID from a coins object
func PoolIDFromCoins(coins sdk.Coins) string {
	return PoolID(coins[0].Denom, coins[1].Denom)
}

// PoolID returns an alphabetically sorted pool name from two denoms.
// The name is commutative for any all pairs A,B: f(A,B) == f(B,A).
func PoolID(denomA string, denomB string) string {
	if denomB < denomA {
		return fmt.Sprintf("%s%s%s", denomB, PoolIDSep, denomA)
	}

	return fmt.Sprintf("%s%s%s", denomA, PoolIDSep, denomB)
}

// NewPoolRecord takes reserve coins and total shares, returning
// a new pool record with a id
func NewPoolRecord(reserves sdk.Coins, totalShares sdkmath.Int) PoolRecord {
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

	tokens := strings.Split(p.PoolID, PoolIDSep)
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

// NewShareRecord takes a depositor, poolID, and shares and returns
// a new share record for storage in state.
func NewShareRecord(depositor sdk.AccAddress, poolID string, sharesOwned sdkmath.Int) ShareRecord {
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

	tokens := strings.Split(sr.PoolID, PoolIDSep)
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
		return fmt.Errorf("depositor '%s' and pool '%s' has invalid total shares: %s", sr.Depositor, sr.PoolID, sr.SharesOwned.String())
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
				return fmt.Errorf("duplicate depositor '%s' and poolID '%s'", sr.Depositor, sr.PoolID)
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
