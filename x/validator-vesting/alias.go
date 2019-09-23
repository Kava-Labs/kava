// nolint
package validatorvesting

import (
	"github.com/cosmos/cosmos-sdk/x/validator-vesting/internal/keeper"
	"github.com/cosmos/cosmos-sdk/x/validator-vesting/internal/types"
)

const (
	ModuleName = types.ModuleName
	StoreKey   = types.StoreKey
)

var (
	NewGenesisState = types.NewGenesisState
)

type (
	GenesisState            = types.GenesisState
	Keeper                  = keeper.Keeper
	ValidatorVestingAccount = types.ValidatorVestingAccount
)
