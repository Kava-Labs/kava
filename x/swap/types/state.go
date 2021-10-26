package types

import (
	"errors"
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

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
func NewPoolRecord(reserves sdk.Coins, totalShares sdk.Int) PoolRecord {
	if len(reserves) != 2 {
		panic("reserves must have two denominations")
	}

	poolID := PoolIDFromCoins(reserves)

	return PoolRecord{
		PoolId:      poolID,
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
		PoolId:      poolID,
		ReservesA:   reserves[0],
		ReservesB:   reserves[1],
		TotalShares: pool.TotalShares(),
	}
}

// Validate performs basic validation checks of the record data
func (p PoolRecord) Validate() error {
	if p.PoolId == "" {
		return errors.New("poolID must be set")
	}

	tokens := strings.Split(p.PoolId, PoolIDSep)
	if len(tokens) != 2 || tokens[0] == "" || tokens[1] == "" || tokens[1] < tokens[0] || tokens[0] == tokens[1] {
		return fmt.Errorf("poolID '%s' is invalid", p.PoolId)
	}
	if sdk.ValidateDenom(tokens[0]) != nil || sdk.ValidateDenom(tokens[1]) != nil {
		return fmt.Errorf("poolID '%s' is invalid", p.PoolId)
	}
	if tokens[0] != p.ReservesA.Denom || tokens[1] != p.ReservesB.Denom {
		return fmt.Errorf("poolID '%s' does not match reserves", p.PoolId)
	}

	if !p.ReservesA.IsPositive() {
		return fmt.Errorf("pool '%s' has invalid reserves: %s", p.PoolId, p.ReservesA)
	}

	if !p.ReservesB.IsPositive() {
		return fmt.Errorf("pool '%s' has invalid reserves: %s", p.PoolId, p.ReservesB)
	}

	if !p.TotalShares.IsPositive() {
		return fmt.Errorf("pool '%s' has invalid total shares: %s", p.PoolId, p.TotalShares)
	}

	return nil
}

// Reserves returns the total reserves for a pool
func (p PoolRecord) Reserves() sdk.Coins {
	return sdk.NewCoins(p.ReservesA, p.ReservesB)
}

// ValidatePoolRecords performs basic validation checks on all records in the slice
func ValidatePoolRecords(prs []PoolRecord) error {
	seenPoolIDs := make(map[string]bool)

	for _, p := range prs {
		if err := p.Validate(); err != nil {
			return err
		}

		if seenPoolIDs[p.PoolId] {
			return fmt.Errorf("duplicate poolID '%s'", p.PoolId)
		}

		seenPoolIDs[p.PoolId] = true
	}

	return nil
}

// NewShareRecord takes a depositor, poolID, and shares and returns
// a new share record for storage in state.
func NewShareRecord(depositor sdk.AccAddress, poolID string, sharesOwned sdk.Int) ShareRecord {
	return ShareRecord{
		Depositor:   depositor.String(),
		PoolId:      poolID,
		SharesOwned: sharesOwned,
	}
}

// Validate performs basic validation checks of the record data
func (sr ShareRecord) Validate() error {
	if sr.PoolId == "" {
		return errors.New("poolID must be set")
	}

	tokens := strings.Split(sr.PoolId, PoolIDSep)
	if len(tokens) != 2 || tokens[0] == "" || tokens[1] == "" || tokens[1] < tokens[0] || tokens[0] == tokens[1] {
		return fmt.Errorf("poolID '%s' is invalid", sr.PoolId)
	}
	if sdk.ValidateDenom(tokens[0]) != nil || sdk.ValidateDenom(tokens[1]) != nil {
		return fmt.Errorf("poolID '%s' is invalid", sr.PoolId)
	}

	if len(sr.Depositor) == 0 {
		return fmt.Errorf("share record cannot have empty depositor address")
	}

	if !sr.SharesOwned.IsPositive() {
		return fmt.Errorf("depositor '%s' and pool '%s' has invalid total shares: %s", sr.Depositor, sr.PoolId, sr.SharesOwned.String())
	}

	return nil
}

// ValidateShareRecords performs basic validation checks on all records in the slice
func ValidateShareRecords(srs []ShareRecord) error {
	seenDepositors := make(map[string]map[string]bool)

	for _, sr := range srs {
		if err := sr.Validate(); err != nil {
			return err
		}

		if seenPools, found := seenDepositors[sr.Depositor]; found {
			if seenPools[sr.PoolId] {
				return fmt.Errorf("duplicate depositor '%s' and poolID '%s'", sr.Depositor, sr.PoolId)
			}
			seenPools[sr.PoolId] = true
		} else {
			seenPools := make(map[string]bool)
			seenPools[sr.PoolId] = true
			seenDepositors[sr.Depositor] = seenPools
		}
	}

	return nil
}
