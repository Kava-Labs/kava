package keeper

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/kava-labs/kava/x/community/types"
)

// Keeper of the community store
type Keeper struct {
	key storetypes.StoreKey
	cdc codec.Codec

	accountKeeper  types.AccountKeeper
	bankKeeper     types.BankKeeper
	cdpKeeper      types.CdpKeeper
	distrKeeper    types.DistributionKeeper
	hardKeeper     types.HardKeeper
	moduleAddress  sdk.AccAddress
	mintKeeper     types.MintKeeper
	kavadistKeeper types.KavadistKeeper
	stakingKeeper  types.StakingKeeper

	// the address capable of executing a MsgUpdateParams message. Typically, this
	// should be the x/gov module account.
	authority sdk.AccAddress

	legacyCommunityPoolAddress sdk.AccAddress
}

// NewKeeper creates a new community Keeper instance
func NewKeeper(
	cdc codec.Codec,
	key storetypes.StoreKey,
	ak types.AccountKeeper,
	bk types.BankKeeper,
	ck types.CdpKeeper,
	dk types.DistributionKeeper,
	hk types.HardKeeper,
	mk types.MintKeeper,
	kk types.KavadistKeeper,
	sk types.StakingKeeper,
	authority sdk.AccAddress,
) Keeper {
	// ensure community module account is set
	addr := ak.GetModuleAddress(types.ModuleAccountName)
	if addr == nil {
		panic(fmt.Sprintf("%s module account has not been set", types.ModuleAccountName))
	}
	legacyAddr := ak.GetModuleAddress(types.LegacyCommunityPoolModuleName)
	if addr == nil {
		panic("legacy community pool address not found")
	}
	if err := sdk.VerifyAddressFormat(authority); err != nil {
		panic(fmt.Sprintf("invalid authority address: %s", err))
	}

	return Keeper{
		key: key,
		cdc: cdc,

		accountKeeper:  ak,
		bankKeeper:     bk,
		cdpKeeper:      ck,
		distrKeeper:    dk,
		hardKeeper:     hk,
		mintKeeper:     mk,
		kavadistKeeper: kk,
		stakingKeeper:  sk,
		moduleAddress:  addr,

		authority:                  authority,
		legacyCommunityPoolAddress: legacyAddr,
	}
}

// GetAuthority returns the x/community module's authority.
func (k Keeper) GetAuthority() sdk.AccAddress {
	return k.authority
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

// GetStakingRewardsState returns the staking reward state or the default state if not set
func (k Keeper) GetStakingRewardsState(ctx sdk.Context) types.StakingRewardsState {
	store := ctx.KVStore(k.key)

	b := store.Get(types.StakingRewardsStateKey)
	if b == nil {
		return types.DefaultStakingRewardsState()
	}

	state := types.StakingRewardsState{}
	k.cdc.MustUnmarshal(b, &state)

	return state
}

// SetStakingRewardsState validates and sets the staking rewards state in the store
func (k Keeper) SetStakingRewardsState(ctx sdk.Context, state types.StakingRewardsState) {
	if err := state.Validate(); err != nil {
		panic(fmt.Sprintf("invalid state: %s", err))
	}

	store := ctx.KVStore(k.key)
	b := k.cdc.MustMarshal(&state)

	store.Set(types.StakingRewardsStateKey, b)
}
