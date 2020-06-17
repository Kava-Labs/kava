package types

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	stakingexported "github.com/cosmos/cosmos-sdk/x/staking/exported"

	bankexported "github.com/cosmos/cosmos-sdk/x/bank/exported"
)

// AccountKeeper defines the expected account keeper (noalias)
type AccountKeeper interface {
	SetModuleAccount(sdk.Context, authtypes.ModuleAccountI)

	GetAccount(sdk.Context, sdk.AccAddress) authtypes.AccountI
	SetAccount(sdk.Context, authtypes.AccountI)
	GetAllAccounts(ctx sdk.Context) (accounts []authtypes.AccountI)
	IterateAccounts(ctx sdk.Context, cb func(account authtypes.AccountI) (stop bool))
}

// TODO: combine

// BankKeeper defines the expected bank keeper (noalias)
type BankKeeper interface {
	SendCoins(ctx sdk.Context, fromAddr sdk.AccAddress, toAddr sdk.AccAddress, amt sdk.Coins) error
	SendCoinsFromAccountToModule(ctx sdk.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error
	BurnCoins(ctx sdk.Context, name string, amt sdk.Coins) error
	GetSupply(ctx sdk.Context) (supply bankexported.SupplyI)
}

// StakingKeeper defines the expected staking keeper (noalias)
type StakingKeeper interface {
	IterateDelegations(ctx sdk.Context, delegator sdk.AccAddress,
		fn func(index int64, delegation stakingexported.DelegationI) (stop bool))
	Undelegate(
		ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress, sharesAmount sdk.Dec,
	) (time.Time, error)
}
