package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// PoolIDFromCoins returns a poolID from a coins object
func PoolIDFromCoins(coins sdk.Coins) string {
	return PoolID(coins[0].Denom, coins[1].Denom)
}

// PoolID returns an alphabetically sorted pool name from two denoms.
// The name is commutative for any all pairs A,B: f(A,B) == f(B,A).
func PoolID(denomA string, denomB string) string {
	if denomB < denomA {
		return fmt.Sprintf("%s/%s", denomB, denomA)
	}

	return fmt.Sprintf("%s/%s", denomA, denomB)
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

func (p PoolRecord) Validate() error {
	if !p.ReservesA.IsValid() {
		return fmt.Errorf("pool %s has invalid reserves: %s", p.PoolID, p.ReservesA)
	}
	if !p.ReservesB.IsValid() {
		return fmt.Errorf("pool %s has invalid reserves: %s", p.PoolID, p.ReservesB)
	}
	if p.TotalShares.IsNegative() {
		return fmt.Errorf("pool %s has negative shares: %s", p.PoolID, p.TotalShares)
	}
	return nil
}

// Reserves returns the total reserves for a pool
func (p PoolRecord) Reserves() sdk.Coins {
	return sdk.NewCoins(p.ReservesA, p.ReservesB)
}

// PoolRecords is a slice of PoolRecord
type PoolRecords []PoolRecord

func (prs PoolRecords) Validate() error {
	for _, p := range prs {
		if err := p.Validate(); err != nil {
			return err
		}
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

func (sr ShareRecord) Validate() error {
	if sr.Depositor.Empty() {
		return fmt.Errorf("share record cannot have empty depositor address")
	}
	if sr.SharesOwned.IsNegative() {
		return fmt.Errorf("pool %s depositor %s has negative shares: %s", sr.PoolID, sr.Depositor, sr.SharesOwned)
	}
	return nil
}

// ShareRecords is a slice of ShareRecord
type ShareRecords []ShareRecord

func (srs ShareRecords) Validate() error {
	for _, sr := range srs {
		if err := sr.Validate(); err != nil {
			return err
		}
	}
	return nil
}
