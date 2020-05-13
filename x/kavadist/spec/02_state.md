# State

## Parameters and Genesis State

`Parameters` define the rate at which inflationary coins are minted and for how long inflationary periods last.

```go
// Params governance parameters for kavadist module
type Params struct {
	Active  bool    `json:"active" yaml:"active"`
	Periods Periods `json:"periods" yaml:"periods"`
}

// Period stores the specified start and end dates, and the inflation, expressed as a decimal representing the yearly APR of tokens that will be minted during that period
type Period struct {
	Start     time.Time `json:"start" yaml:"start"`         // example "2020-03-01T15:20:00Z"
	End       time.Time `json:"end" yaml:"end"`             // example "2020-06-01T15:20:00Z"
	Inflation sdk.Dec   `json:"inflation" yaml:"inflation"` // example "1.000000003022265980"  - 10% inflation
}
```

`GenesisState` defines the state that must be persisted when the blockchain stops/restarts in order for normal function of the kavadist module to resume.

```go
// GenesisState is the state that must be provided at genesis.
type GenesisState struct {
	Params            Params    `json:"params" yaml:"params"`
	PreviousBlockTime time.Time `json:"previous_block_time" yaml:"previous_block_time"`
}
```
