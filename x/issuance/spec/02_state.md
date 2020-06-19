<!--
order: 2
-->

# State

## Parameters and Genesis State

```go // TODO
type Asset struct {
  Owner sdk.AccAddress `json:"owner" yaml:"owner"`
  BlockedAccounts []sdk.AccAddress `json:"blocked_accounts" yaml:"blocked_accounts"`
  Paused bool `json:"paused" yaml:"paused"`
}

type Params struct {
  Assets Assets `json:"assets" yaml:"assets"`
}

type GenesisState struct {
  Assets Assets `json:"assets" yaml:"assets"`
}
```
