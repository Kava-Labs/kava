package v18de63

// GenesisState - all auth state that must be provided at genesis
type GenesisState struct {
	Params   Params          `json:"params" yaml:"params"`
	Accounts GenesisAccounts `json:"accounts" yaml:"accounts"`
}
