package v0_13

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	cdptypes "github.com/kava-labs/kava/x/cdp/legacy/v0_13"

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

// GenesisState is the state that must be provided at genesis.
type GenesisState struct {
	Params            Params    `json:"params" yaml:"params"`
	PreviousBlockTime time.Time `json:"previous_block_time" yaml:"previous_block_time"`
}

// NewGenesisState returns a new genesis state
func NewGenesisState(params Params, previousBlockTime time.Time) GenesisState {
	return GenesisState{
		Params:            params,
		PreviousBlockTime: previousBlockTime,
	}
}

// DefaultGenesisState returns a default genesis state
func DefaultGenesisState() GenesisState {
	return GenesisState{
		Params:            DefaultParams(),
		PreviousBlockTime: DefaultPreviousBlockTime,
	}
}

// Params governance parameters for kavadist module
type Params struct {
	Active  bool    `json:"active" yaml:"active"`
	Periods Periods `json:"periods" yaml:"periods"`
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

// Periods array of Period
type Periods []Period
