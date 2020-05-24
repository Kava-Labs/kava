package v18de63

import (
	authtypes "github.com/kava-labs/kava/migrate/v0_8/sdk/auth/v18de63"
)

// ModuleAccount defines an account for modules that holds coins on a pool
type ModuleAccount struct {
	*authtypes.BaseAccount
	Name        string   `json:"name" yaml:"name"`               // name of the module
	Permissions []string `json:"permissions" yaml:"permissions"` // permissions of module account
}
