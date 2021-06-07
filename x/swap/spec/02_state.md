<!--
order: 2
-->

# State

## Parameters and Genesis State

`Parameters` define the governance parameters and default behavior of the swap module.

```go
// Params are governance parameters for the swap module
type Params struct {
	Pairs   Pairs   `json:"pairs" yaml:"pairs"`
	SwapFee sdk.Dec `json:"swap_fee" yaml:"swap_fee"`
}

// Pair defines a tradable token pair
type Pair struct {
	TokenA string `json:"token_a" yaml:"token_a"`
	TokenB string `json:"token_b" yaml:"token_b"`
}

// Pairs is a slice of Pair
type Pairs []Pair
```

`GenesisState` defines the state that must be persisted when the blockchain stops/restarts in order for the normal function of the swap module to resume.

```go
// GenesisState is the state that must be provided at genesis.
type GenesisState struct {
	Params Params `json:"params" yaml:"params"`
}
```
