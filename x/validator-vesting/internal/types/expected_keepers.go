package types

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authexported "github.com/cosmos/cosmos-sdk/x/auth/exported"
	stakingexported "github.com/cosmos/cosmos-sdk/x/staking/exported"
	supplyexported "github.com/cosmos/cosmos-sdk/x/supply/exported"
)

// AccountKeeper defines the expected account keeper (noalias)
type AccountKeeper interface {
	GetAccount(sdk.Context, sdk.AccAddress) authexported.Account
	SetAccount(sdk.Context, authexported.Account)
	GetAllAccounts(ctx sdk.Context) (accounts []authexported.Account)
	IterateAccounts(ctx sdk.Context, cb func(account authexported.Account) (stop bool))
}

// BankKeeper defines the expected bank keeper (noalias)
type BankKeeper interface {
	SendCoins(ctx sdk.Context, fromAddr sdk.AccAddress, toAddr sdk.AccAddress, amt sdk.Coins) sdk.Error
}

// StakingKeeper defines the expected staking keeper (noalias)
type StakingKeeper interface {
	IterateDelegations(ctx sdk.Context, delegator sdk.AccAddress,
		fn func(index int64, delegation stakingexported.DelegationI) (stop bool))
	Undelegate(
		ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress, sharesAmount sdk.Dec,
	) (time.Time, sdk.Error)
}

// SupplyKeeper defines the expected supply keeper for module accounts (noalias)
type SupplyKeeper interface {
	SendCoinsFromAccountToModule(ctx sdk.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) sdk.Error
	BurnCoins(ctx sdk.Context, name string, amt sdk.Coins) sdk.Error
	SetModuleAccount(sdk.Context, supplyexported.ModuleAccountI)
}
