package types

import (
	"fmt"
	"sort"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// Valid reward multipliers
const (
	Small  MultiplierName = "small"
	Medium MultiplierName = "medium"
	Large  MultiplierName = "large"
)

// MultiplierName name for valid multiplier
type MultiplierName string

// IsValid checks if the input is one of the expected strings
func (mn MultiplierName) IsValid() error {
	switch mn {
	case Small, Medium, Large:
		return nil
	}
	return sdkerrors.Wrapf(ErrInvalidMultiplier, "invalid multiplier name: %s", mn)
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

// Multipliers slice of Multiplier
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

// String implements fmt.Stringer
func (ms Multipliers) String() string {
	out := "Claim Multipliers\n"
	for _, s := range ms {
		out += fmt.Sprintf("%s\n", s)
	}
	return out
}

type Selection struct {
	Denom          string
	MultiplierName string
}

func NewSelection(denom, multiplierName string) Selection {
	return Selection{
		Denom:          denom,
		MultiplierName: multiplierName,
	}
}

func (s Selection) Validate() error {
	if err := sdk.ValidateDenom(s.Denom); err != nil {
		return sdkerrors.Wrap(ErrInvalidClaimDenoms, err.Error())
	}
	// TODO validate multiplier name? or leave for on chain check
	// if err := MultiplierName(s.MultiplierName).IsValid(); err != nil {
	// 	return err
	// }
	return nil
}

// Selections are a list of denom - multiplier pairs that specify what rewards to claim and with what lockups.
type Selections []Selection

func NewSelectionsFromMap(selectionMap map[string]string) Selections {
	var selections Selections
	for k, v := range selectionMap {
		selections = append(selections, NewSelection(k, v))
	}
	// sort the slice by denom to protect against the random range order causing consensus failures
	sort.SliceStable(selections, func(i, j int) bool {
		return selections[i].Denom > selections[j].Denom
	})
	return selections
}

func (s Selections) Validate() error {
	if len(s) == 0 {
		return sdkerrors.Wrap(ErrInvalidClaimDenoms, "cannot claim 0 denoms")
	}
	if len(s) >= MaxDenomsToClaim {
		return sdkerrors.Wrapf(ErrInvalidClaimDenoms, "cannot claim more than %d denoms", MaxDenomsToClaim)
	}
	for _, d := range s {
		if err := d.Validate(); err != nil {
			return err
		}
	}
	return nil
}

type ConfirmedSelection struct {
	Denom      string
	Multiplier Multiplier
}
