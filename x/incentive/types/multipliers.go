package types

import (
	"fmt"
	"sort"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// Available reward multipliers names
const (
	Small  MultiplierName = "small"
	Medium MultiplierName = "medium"
	Large  MultiplierName = "large"
)

// MultiplierName is the user facing ID for a multiplier. There is a restricted set of possible values.
type MultiplierName string

// IsValid checks if the input is one of the expected strings
func (mn MultiplierName) IsValid() error {
	switch mn {
	case Small, Medium, Large:
		return nil
	}
	return sdkerrors.Wrapf(ErrInvalidMultiplier, "invalid multiplier name: %s", mn)
}

// ParseMultiplierName converts a string into a valid MultiplierName value.
func ParseMultiplierName(unparsedName string) (MultiplierName, error) {
	name := MultiplierName(unparsedName)
	if err := name.IsValid(); err != nil {
		return "", err
	}
	return name, nil
}

// Multiplier amount the claim rewards get increased by, along with how long the claim rewards are locked
type Multiplier struct {
	Name         MultiplierName `json:"name" yaml:"name"`
	MonthsLockup int64          `json:"months_lockup" yaml:"months_lockup"`
	Factor       sdk.Dec        `json:"factor" yaml:"factor"`
}

// NewMultiplier returns a new Multiplier
func NewMultiplier(name MultiplierName, lockup int64, factor sdk.Dec) Multiplier {
	return Multiplier{
		Name:         name,
		MonthsLockup: lockup,
		Factor:       factor,
	}
}

// Validate multiplier param
func (m Multiplier) Validate() error {
	if err := m.Name.IsValid(); err != nil {
		return err
	}
	if m.MonthsLockup < 0 {
		return fmt.Errorf("expected non-negative lockup, got %d", m.MonthsLockup)
	}
	if m.Factor.IsNegative() {
		return fmt.Errorf("expected non-negative factor, got %s", m.Factor.String())
	}

	return nil
}

// String implements fmt.Stringer
func (m Multiplier) String() string {
	return fmt.Sprintf(`Claim Multiplier:
	Name: %s
	Months Lockup %d
	Factor %s
	`, m.Name, m.MonthsLockup, m.Factor)
}

// Multipliers is a slice of Multiplier
type Multipliers []Multiplier

// Validate validates each multiplier
func (ms Multipliers) Validate() error {
	for _, m := range ms {
		if err := m.Validate(); err != nil {
			return err
		}
	}
	return nil
}

// Get returns a multiplier with a matching name
func (ms Multipliers) Get(name MultiplierName) (Multiplier, bool) {
	for _, m := range ms {
		if m.Name == name {
			return m, true
		}
	}
	return Multiplier{}, false
}

// String implements fmt.Stringer
func (ms Multipliers) String() string {
	out := "Claim Multipliers\n"
	for _, s := range ms {
		out += fmt.Sprintf("%s\n", s)
	}
	return out
}

// MultipliersPerDenom is a map of denoms to a set of multipliers
type MultipliersPerDenom []struct {
	Denom       string
	Multipliers Multipliers
}

// Validate checks each denom and multipliers for invalid values.
func (mpd MultipliersPerDenom) Validate() error {
	foundDenoms := map[string]bool{}

	for _, item := range mpd {
		if err := sdk.ValidateDenom(item.Denom); err != nil {
			return err
		}
		if err := item.Multipliers.Validate(); err != nil {
			return err
		}

		if foundDenoms[item.Denom] {
			return fmt.Errorf("")
		}
		foundDenoms[item.Denom] = true
	}
	return nil
}

// Selection a pair of denom and multiplier name. It holds the choice of multiplier a user makes when they claim a denom.
type Selection struct {
	Denom          string
	MultiplierName string
}

// NewSelection returns a new Selection
func NewSelection(denom, multiplierName string) Selection {
	return Selection{
		Denom:          denom,
		MultiplierName: multiplierName,
	}
}

// Validate performs basic validation checks
func (s Selection) Validate() error {
	if err := sdk.ValidateDenom(s.Denom); err != nil {
		return sdkerrors.Wrap(ErrInvalidClaimDenoms, err.Error())
	}
	if _, err := ParseMultiplierName(s.MultiplierName); err != nil {
		return err
	}
	return nil
}

// Selections are a list of denom - multiplier pairs that specify what rewards to claim and with what lockups.
type Selections []Selection

// NewSelectionsFromMap creates a new set of selections from a string to string map.
// It sorts the output before returning.
func NewSelectionsFromMap(selectionMap map[string]string) Selections {
	var selections Selections
	for k, v := range selectionMap {
		selections = append(selections, NewSelection(k, v))
	}
	// deterministically sort the slice to protect against the random range order causing consensus failures
	sort.Slice(selections, func(i, j int) bool {
		if selections[i].Denom != selections[j].Denom {
			return selections[i].Denom < selections[j].Denom
		}
		return selections[i].MultiplierName < selections[j].MultiplierName
	})
	return selections
}

// Valdate performs basic validaton checks
func (ss Selections) Validate() error {
	if len(ss) == 0 {
		return sdkerrors.Wrap(ErrInvalidClaimDenoms, "cannot claim 0 denoms")
	}
	if len(ss) >= MaxDenomsToClaim {
		return sdkerrors.Wrapf(ErrInvalidClaimDenoms, "cannot claim more than %d denoms", MaxDenomsToClaim)
	}
	foundDenoms := map[string]bool{}
	for _, s := range ss {
		if err := s.Validate(); err != nil {
			return err
		}
		if foundDenoms[s.Denom] {
			return sdkerrors.Wrapf(ErrInvalidClaimDenoms, "cannot claim denom '%s' more than once", s.Denom)
		}
		foundDenoms[s.Denom] = true
	}
	return nil
}
