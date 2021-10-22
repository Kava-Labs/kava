package types

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"

	tmtime "github.com/tendermint/tendermint/types/time"
	// cdptypes "github.com/kava-labs/kava/x/cdp/types"
)

// Parameter keys and default values
var (
	KeyActive                = []byte("Active")
	KeyPeriods               = []byte("Periods")
	DefaultActive            = false
	DefaultPeriods           = []Period{}
	DefaultPreviousBlockTime = tmtime.Canonical(time.Unix(1, 0))
	// GovDenom                 = cdptypes.DefaultGovDenom
)

// NewPeriod returns a new instance of Period
func NewPeriod(start time.Time, end time.Time, inflation sdk.Dec) Period {
	return Period{
		Start:     start,
		End:       end,
		Inflation: inflation,
	}
}

// String implements fmt.Stringer
func (pr Period) String() string {
	return fmt.Sprintf(`Period:
	Start: %s
	End: %s
	Inflation: %s`, pr.Start, pr.End, pr.Inflation)
}

func NewParams(active bool, periods []Period) Params {
	return Params{
		Active:  active,
		Periods: periods,
	}
}

func DefaultParams() Params {
	return NewParams(DefaultActive, DefaultPeriods)
}

func (p Params) String() string {
	periods := "Periods\n"
	for _, pr := range p.Periods {
		periods += fmt.Sprintf("%s\n", pr)
	}
	return fmt.Sprintf(`Params:
	Active: %t
	Periods %s`, p.Active, periods)
}

func (p Params) Equal(other Params) bool {
	if p.Active != other.Active {
		return false
	}

	if len(p.Periods) != len(other.Periods) {
		return false
	}

	for i, period := range p.Periods {
		if !period.Equal(other.Periods[i]) {
			return false
		}
	}

	return true
}

// ParamKeyTable Key declaration for parameters
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// ParamSetPairs implements the ParamSet interface and returns all the key/value pairs
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyActive, &p.Active, validateActiveParam),
		paramtypes.NewParamSetPair(KeyPeriods, &p.Periods, validatePeriodsParams),
	}
}

// Validate checks that the parameters have valid values.
func (p Params) Validate() error {
	if err := validateActiveParam(p.Active); err != nil {
		return err
	}

	return validatePeriodsParams(p.Periods)
}

func validateActiveParam(i interface{}) error {
	_, ok := i.(bool)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	return nil
}

func validatePeriodsParams(i interface{}) error {
	periods, ok := i.([]Period)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	prevEnd := tmtime.Canonical(time.Unix(0, 0))
	for _, pr := range periods {
		if pr.End.Before(pr.Start) {
			return fmt.Errorf("end time for period is before start time: %s", pr)
		}

		if pr.Start.Before(prevEnd) {
			return fmt.Errorf("periods must be in chronological order: %s", periods)
		}
		prevEnd = pr.End

		if pr.Start.Unix() <= 0 || pr.End.Unix() <= 0 {
			return fmt.Errorf("start or end time cannot be zero: %s", pr)
		}

		//TODO: validate period Inflation?
	}

	return nil
}
