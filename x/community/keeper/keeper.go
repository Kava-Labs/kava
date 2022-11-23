package keeper

import (
	"fmt"

	"github.com/tendermint/tendermint/libs/log"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kava-labs/kava/x/community/types"
)

// Keeper of the community store
type Keeper struct {
	bankKeeper    types.BankKeeper
	moduleAddress sdk.AccAddress
}

// NewKeeper creates a new community Keeper instance
func NewKeeper(ak types.AccountKeeper, bk types.BankKeeper) Keeper {
	// ensure community module account is set
	addr := ak.GetModuleAddress(types.ModuleAccountName)
	if addr == nil {
		panic(fmt.Sprintf("%s module account has not been set", types.ModuleAccountName))
	}

	return Keeper{
		bankKeeper:    bk,
		moduleAddress: addr,
	}
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", "x/"+types.ModuleName)
}

// GetModuleAccountBalance returns all the coins held by the community module account
func (k Keeper) GetModuleAccountBalance(ctx sdk.Context) sdk.Coins {
	return k.bankKeeper.GetAllBalances(ctx, k.moduleAddress)
}

// FundCommunityPool transfers coins from the sender to the community module account.
func (k Keeper) FundCommunityPool(ctx sdk.Context, sender sdk.AccAddress, amount sdk.Coins) error {
	return k.bankKeeper.SendCoinsFromAccountToModule(ctx, sender, types.ModuleAccountName, amount)
}

// DistributeFromCommunityPool transfers coins from the community pool to recipient.
func (k Keeper) DistributeFromCommunityPool(ctx sdk.Context, recipient sdk.AccAddress, amount sdk.Coins) error {
	return k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleAccountName, recipient, amount)
}
