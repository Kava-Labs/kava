package keeper

import (
	"fmt"

	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"

	"github.com/kava-labs/kava/x/community/types"
)

// Keeper of the community store
type Keeper struct {
	bankKeeper    types.BankKeeper
	hardKeeper    types.HardKeeper
	moduleAddress sdk.AccAddress
	paramSubspace paramtypes.Subspace

	// Msg server router
	router *baseapp.MsgServiceRouter
}

// NewKeeper creates a new community Keeper instance
func NewKeeper(params paramtypes.Subspace, ak types.AccountKeeper, bk types.BankKeeper, hk types.HardKeeper, router *baseapp.MsgServiceRouter) Keeper {
	// ensure community module account is set
	addr := ak.GetModuleAddress(types.ModuleAccountName)
	if addr == nil {
		panic(fmt.Sprintf("%s module account has not been set", types.ModuleAccountName))
	}

	if !params.HasKeyTable() {
		params = params.WithKeyTable(types.ParamKeyTable())
	}

	return Keeper{
		bankKeeper:    bk,
		hardKeeper:    hk,
		moduleAddress: addr,
		router:        router,
		paramSubspace: params,
	}
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", "x/"+types.ModuleName)
}

// Router returns the keeper's router
func (keeper Keeper) Router() *baseapp.MsgServiceRouter {
	return keeper.router
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

// GetParams returns the total set of evm parameters.
func (k Keeper) GetParams(ctx sdk.Context) (params types.Params) {
	k.paramSubspace.GetParamSet(ctx, &params)
	return params
}

// SetParams sets the evm parameters to the param space.
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	k.paramSubspace.SetParamSet(ctx, &params)
}
