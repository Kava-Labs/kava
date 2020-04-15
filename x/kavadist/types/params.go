package types

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/params"
	cdptypes "github.com/kava-labs/kava/x/cdp/types"
	tmtime "github.com/tendermint/tendermint/types/time"
)

// Parameter keys and default values
var (
	KeyActive                = []byte("Active")
	KeyPeriods               = []byte("Periods")
	DefaultActive            = false
	DefaultPeriods           = Periods{}
	DefaultPreviousBlockTime = tmtime.Canonical(time.Unix(0, 0))
	GovDenom                 = cdptypes.DefaultGovDenom
)

// Params governance parameters for kavadist module
type Params struct {
	Active  bool    `json:"active" yaml:"active"`
	Periods Periods `json:"periods" yaml:"periods"`
}

// Period stores the specified start and end dates, and the inflation, expressed as a decimal representing the yearly APR of KAVA tokens that will be minted during that period
type Period struct {
	Start     time.Time `json:"start" yaml:"start"`         // example "2020-03-01T15:20:00Z"
	End       time.Time `json:"end" yaml:"end"`             // example "2020-06-01T15:20:00Z"
	Inflation sdk.Dec   `json:"inflation" yaml:"inflation"` // example "1.000000003022265980"  - 10% inflation
}

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

// Periods array of Period
type Periods []Period

// String implements fmt.Stringer
func (prs Periods) String() string {
	out := "Periods\n"
	for _, pr := range prs {
		out += fmt.Sprintf("%s\n", pr)
	}
	return out
}

// NewParams returns a new params object
func NewParams(active bool, periods Periods) Params {
	return Params{
		Active:  active,
		Periods: periods,
	}
}

// DefaultParams returns default params for kavadist module
func DefaultParams() Params {
	return NewParams(DefaultActive, DefaultPeriods)
}

// String implements fmt.Stringer
func (p Params) String() string {
	return fmt.Sprintf(`Params:
	Active: %t
	Periods %s`, p.Active, p.Periods)
}

// ParamKeyTable Key declaration for parameters
func ParamKeyTable() params.KeyTable {
	return params.NewKeyTable().RegisterParamSet(&Params{})
}

// ParamSetPairs implements the ParamSet interface and returns all the key/value pairs
func (p *Params) ParamSetPairs() params.ParamSetPairs {
	return params.ParamSetPairs{
		{Key: KeyActive, Value: &p.Active},
		{Key: KeyPeriods, Value: &p.Periods},
	}
}

// Validate checks that the parameters have valid values.
func (p Params) Validate() error {
	prevEnd := tmtime.Canonical(time.Unix(0, 0))
	for _, pr := range p.Periods {
		if pr.End.Before(pr.Start) {
			return fmt.Errorf("end time for period is before start time: %s", pr)
		}
		if pr.Start.Before(prevEnd) {
			return fmt.Errorf("periods must be in chronological order: %s", p.Periods)
		}
		prevEnd = pr.End
	}
	return nil
}
