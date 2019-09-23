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
	NewValidatorVestingAccountRaw = types.NewValidatorVestingAccountRaw
	NewValidatorVestingAccount    = types.NewValidatorVestingAccount
	NewGenesisState               = types.NewGenesisState
	DefaultGenesisState           = types.DefaultGenesisState
	RegisterCodec                 = types.RegisterCodec
	ValidatorVestingAccountPrefix = types.ValidatorVestingAccountPrefix
	ValidatorVestingAccountKey    = types.ValidatorVestingAccountKey
	NewKeeper                     = keeper.NewKeeper
)

type (
	GenesisState            = types.GenesisState
	Keeper                  = keeper.Keeper
	ValidatorVestingAccount = types.ValidatorVestingAccount
)
