package types

import (
	fmt "fmt"
	"sort"
	"strings"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NewVaultShare returns a new VaultShare
func NewVaultShare(denom string, amount sdk.Dec) VaultShare {
	share := VaultShare{
		Denom:  denom,
		Amount: amount,
	}

	if err := share.Validate(); err != nil {
		panic(err)
	}

	return share
}

// Validate returns an error if a VaultShare is invalid.
func (share VaultShare) Validate() error {
	if err := sdk.ValidateDenom(share.Denom); err != nil {
		return errorsmod.Wrap(ErrInvalidVaultDenom, err.Error())
	}

	if share.Amount.IsNil() {
		return fmt.Errorf("nil share amount: %s", share.Amount)
	}

	if share.Amount.IsNegative() {
		return fmt.Errorf("vault share amount %v is negative", share.Amount)
	}

	return nil
}

// IsValid returns true if the VaultShare is valid
func (share VaultShare) IsValid() bool {
	return share.Validate() == nil
}

func (share VaultShare) IsPositive() bool {
	return share.Amount.IsPositive()
}

// Add adds amounts of two vault shares with same denom. If the shares differ in
// denom then it panics.
func (share VaultShare) Add(vsB VaultShare) VaultShare {
	if share.Denom != vsB.Denom {
		panic(fmt.Sprintf("invalid share denominations; %s, %s", share.Denom, vsB.Denom))
	}

	return NewVaultShare(share.Denom, share.Amount.Add(vsB.Amount))
}

// IsZero returns if this represents no shares
func (share VaultShare) IsZero() bool {
	return share.Amount.IsZero()
}

// IsNegative returns true if the share amount is negative and false otherwise.
func (share VaultShare) IsNegative() bool {
	return share.Amount.IsNegative()
}

// Sub subtracts amounts of two vault shares with same denom. If the shares
// differ in denom then it panics.
func (share VaultShare) Sub(vsB VaultShare) VaultShare {
	if share.Denom != vsB.Denom {
		panic(fmt.Sprintf("invalid share denominations; %s, %s", share.Denom, vsB.Denom))
	}

	res := NewVaultShare(share.Denom, share.Amount.Sub(vsB.Amount))
	if res.Amount.IsNegative() {
		panic("negative share amount")
	}

	return res
}

func (share VaultShare) String() string {
	return fmt.Sprintf("%v%v", share.Amount, share.Denom)
}

// VaultShares is a slice of VaultShare.
type VaultShares []VaultShare

// NewVaultShares returns new VaultShares
func NewVaultShares(shares ...VaultShare) VaultShares {
	newVaultShares := sanitizeVaultShares(shares)
	if err := newVaultShares.Validate(); err != nil {
		panic(fmt.Errorf("invalid share set %s: %w", newVaultShares, err))
	}

	return newVaultShares
}

func sanitizeVaultShares(shares VaultShares) VaultShares {
	newVaultShares := removeZeroShares(shares)
	if len(newVaultShares) == 0 {
		return VaultShares{}
	}

	return newVaultShares.Sort()
}

// Validate returns an error if a slice of VaultShares is invalid.
func (shares VaultShares) Validate() error {
	switch len(shares) {
	case 0:
		return nil

	case 1:
		if err := sdk.ValidateDenom(shares[0].Denom); err != nil {
			return err
		}
		if !shares[0].IsPositive() {
			return fmt.Errorf("share %s amount is not positive", shares[0])
		}
		return nil
	default:
		// check single share case
		if err := (VaultShares{shares[0]}).Validate(); err != nil {
			return err
		}

		lowDenom := shares[0].Denom
		seenDenoms := make(map[string]bool)
		seenDenoms[lowDenom] = true

		for _, share := range shares[1:] {
			if seenDenoms[share.Denom] {
				return fmt.Errorf("duplicate denomination %s", share.Denom)
			}
			if err := sdk.ValidateDenom(share.Denom); err != nil {
				return err
			}
			if share.Denom <= lowDenom {
				return fmt.Errorf("denomination %s is not sorted", share.Denom)
			}
			if !share.IsPositive() {
				return fmt.Errorf("share %s amount is not positive", share.Denom)
			}

			// we compare each share against the last denom
			lowDenom = share.Denom
			seenDenoms[share.Denom] = true
		}

		return nil
	}
}

// IsValid returns true if the VaultShares are valid
func (shares VaultShares) IsValid() bool {
	return shares.Validate() == nil
}

// Add adds two sets of VaultShares.
func (shares VaultShares) Add(sharesB ...VaultShare) VaultShares {
	return shares.safeAdd(sharesB)
}

// safeAdd will perform addition of two shares sets. If both share sets are
// empty, then an empty set is returned. If only a single set is empty, the
// other set is returned. Otherwise, the shares are compared in order of their
// denomination and addition only occurs when the denominations match, otherwise
// the share is simply added to the sum assuming it's not zero.
// The function panics if `shares` or  `sharesB` are not sorted (ascending).
func (shares VaultShares) safeAdd(sharesB VaultShares) VaultShares {
	// probably the best way will be to make Shares and interface and hide the structure
	// definition (type alias)
	if !shares.isSorted() {
		panic("Shares (self) must be sorted")
	}
	if !sharesB.isSorted() {
		panic("Wrong argument: shares must be sorted")
	}

	sum := (VaultShares)(nil)
	indexA, indexB := 0, 0
	lenA, lenB := len(shares), len(sharesB)

	for {
		if indexA == lenA {
			if indexB == lenB {
				// return nil shares if both sets are empty
				return sum
			}

			// return set B (excluding zero shares) if set A is empty
			return append(sum, removeZeroShares(sharesB[indexB:])...)
		} else if indexB == lenB {
			// return set A (excluding zero shares) if set B is empty
			return append(sum, removeZeroShares(shares[indexA:])...)
		}

		shareA, shareB := shares[indexA], sharesB[indexB]

		switch strings.Compare(shareA.Denom, shareB.Denom) {
		case -1: // share A denom < share B denom
			if !shareA.IsZero() {
				sum = append(sum, shareA)
			}

			indexA++

		case 0: // share A denom == share B denom
			res := shareA.Add(shareB)
			if !res.IsZero() {
				sum = append(sum, res)
			}

			indexA++
			indexB++

		case 1: // share A denom > share B denom
			if !shareB.IsZero() {
				sum = append(sum, shareB)
			}

			indexB++
		}
	}
}

// Sub subtracts a set of shares from another.
//
// e.g.
// {2A, 3B} - {A} = {A, 3B}
// {2A} - {0B} = {2A}
// {A, B} - {A} = {B}
//
// CONTRACT: Sub will never return Shares where one Share has a non-positive
// amount. In otherwords, IsValid will always return true.
func (shares VaultShares) Sub(sharesB ...VaultShare) VaultShares {
	diff, hasNeg := shares.SafeSub(sharesB)
	if hasNeg {
		panic("negative share amount")
	}

	return diff
}

// SafeSub performs the same arithmetic as Sub but returns a boolean if any
// negative share amount was returned.
// The function panics if `shares` or  `sharesB` are not sorted (ascending).
func (shares VaultShares) SafeSub(sharesB VaultShares) (VaultShares, bool) {
	diff := shares.safeAdd(sharesB.negative())
	return diff, diff.IsAnyNegative()
}

// IsAnyNegative returns true if there is at least one share whose amount
// is negative; returns false otherwise. It returns false if the share set
// is empty too.
func (shares VaultShares) IsAnyNegative() bool {
	for _, share := range shares {
		if share.IsNegative() {
			return true
		}
	}

	return false
}

// negative returns a set of shares with all amount negative.
func (shares VaultShares) negative() VaultShares {
	res := make(VaultShares, 0, len(shares))

	for _, share := range shares {
		res = append(res, VaultShare{
			Denom:  share.Denom,
			Amount: share.Amount.Neg(),
		})
	}

	return res
}

// AmountOf returns the amount of shares of the given denom.
func (shares VaultShares) AmountOf(denom string) sdk.Dec {
	for _, s := range shares {
		if s.Denom == denom {
			return s.Amount
		}
	}

	return sdk.ZeroDec()
}

// GetShare the single share of the given denom.
func (shares VaultShares) GetShare(denom string) VaultShare {
	for _, s := range shares {
		if s.Denom == denom {
			return s
		}
	}

	return NewVaultShare(denom, sdk.ZeroDec())
}

// IsZero returns true if the VaultShares is empty.
func (shares VaultShares) IsZero() bool {
	for _, s := range shares {
		// If any amount is non-zero, false
		if !s.Amount.IsZero() {
			return false
		}
	}

	return true
}

func (shares VaultShares) isSorted() bool {
	for i := 1; i < len(shares); i++ {
		if shares[i-1].Denom > shares[i].Denom {
			return false
		}
	}
	return true
}

func (shares VaultShares) String() string {
	if len(shares) == 0 {
		return ""
	}

	out := ""
	for _, share := range shares {
		out += fmt.Sprintf("%v,", share.String())
	}

	return out[:len(out)-1]
}

// removeZeroShares removes all zero shares from the given share set in-place.
func removeZeroShares(shares VaultShares) VaultShares {
	for i := 0; i < len(shares); i++ {
		if shares[i].IsZero() {
			break
		} else if i == len(shares)-1 {
			return shares
		}
	}

	var result VaultShares
	if len(shares) > 0 {
		result = make(VaultShares, 0, len(shares)-1)
	}

	for _, share := range shares {
		if !share.IsZero() {
			result = append(result, share)
		}
	}

	return result
}

// ----------------------------------------------------------------------------
// VaultShares sort interface

func (shares VaultShares) Len() int { return len(shares) }

// Less implements sort.Interface for VaultShares
func (shares VaultShares) Less(i, j int) bool { return shares[i].Denom < shares[j].Denom }

// Swap implements sort.Interface for VaultShares
func (shares VaultShares) Swap(i, j int) { shares[i], shares[j] = shares[j], shares[i] }

var _ sort.Interface = VaultShares{}

// Sort is a helper function to sort the set of vault shares in-place
func (shares VaultShares) Sort() VaultShares {
	sort.Sort(shares)
	return shares
}
