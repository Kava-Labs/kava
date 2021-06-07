<!--
order: 2
-->

# State

## Parameters and Genesis State

`Parameters` define the governance parameters and default behavior of the swap module.

```go
// Params are governance parameters for the swap module
type Params struct {
	AllowedPools   AllowedPools   `json:"allowedPools" yaml:"allowedPools"`
	SwapFee sdk.Dec `json:"swap_fee" yaml:"swap_fee"`
}

// AllowedPool defines a tradable pool
type AllowedPool struct {
	TokenA string `json:"token_a" yaml:"token_a"`
	TokenB string `json:"token_b" yaml:"token_b"`
}

// AllowedPools is a slice of AllowedPool
type AllowedPools []AllowedPool
```

`GenesisState` defines the state that must be persisted when the blockchain stops/restarts in order for the normal function of the swap module to resume.

```go
// GenesisState is the state that must be provided at genesis.
type GenesisState struct {
	Params Params `json:"params" yaml:"params"`
}
```
