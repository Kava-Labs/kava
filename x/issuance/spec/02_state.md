<!--
order: 2
-->

# State

## Parameters and Genesis State

```go

// Asset type for assets in the issuance module
type Asset struct {
  Owner            sdk.AccAddress   `json:"owner" yaml:"owner"`
  Denom            string           `json:"denom" yaml:"denom"`
  BlockedAddresses []sdk.AccAddress `json:"blocked_addresses" yaml:"blocked_addresses"`
  Paused           bool             `json:"paused" yaml:"paused"`
}

// Assets array of Asset
type Assets []Asset

// Params governance parameters for the issuance module
type Params struct {
  Assets Assets `json:"assets" yaml:"assets"`
}

// GenesisState state that must be provided at genesis
type GenesisState struct {
  Assets Assets `json:"assets" yaml:"assets"`
}
```
