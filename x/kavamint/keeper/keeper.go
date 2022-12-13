package keeper

import (
	"time"

	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/kava-labs/kava/x/kavamint/types"
)

// KeeperI is the required keeper interface for x/kavamint's begin blocker
type KeeperI interface {
	AccumulateAndMintInflation(ctx sdk.Context) error
}

// Keeper of the kavamint store
type Keeper struct {
	cdc                            codec.BinaryCodec
	storeKey                       sdk.StoreKey
	paramSpace                     paramtypes.Subspace
	stakingKeeper                  types.StakingKeeper
	bankKeeper                     types.BankKeeper
	stakingRewardsFeeCollectorName string
	communityPoolModuleAccountName string
}

var _ KeeperI = Keeper{}

// NewKeeper creates a new kavamint Keeper instance
func NewKeeper(
	cdc codec.BinaryCodec, key sdk.StoreKey, paramSpace paramtypes.Subspace,
	sk types.StakingKeeper, ak types.AccountKeeper, bk types.BankKeeper,
	stakingRewardsFeeCollectorName string, communityPoolModuleAccountName string,
) Keeper {
	// set KeyTable if it has not already been set
	if !paramSpace.HasKeyTable() {
		paramSpace = paramSpace.WithKeyTable(types.ParamKeyTable())
	}

	return Keeper{
		cdc:                            cdc,
		storeKey:                       key,
		paramSpace:                     paramSpace,
		stakingKeeper:                  sk,
		bankKeeper:                     bk,
		stakingRewardsFeeCollectorName: stakingRewardsFeeCollectorName,
		communityPoolModuleAccountName: communityPoolModuleAccountName,
	}
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", "x/"+types.ModuleName)
}

// GetParams returns the total set of minting parameters.
func (k Keeper) GetParams(ctx sdk.Context) (params types.Params) {
	k.paramSpace.GetParamSet(ctx, &params)
	return params
}

// SetParams sets the total set of minting parameters.
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	k.paramSpace.SetParamSet(ctx, &params)
}

// GetStakingApy returns the APY minted for staking rewards
func (k Keeper) GetStakingApy(ctx sdk.Context) sdk.Dec {
	params := k.GetParams(ctx)
	return params.StakingRewardsApy
}

// bondDenom implements an alias call to the underlying staking keeper's BondDenom.
func (k Keeper) bondDenom(ctx sdk.Context) string {
	return k.stakingKeeper.BondDenom(ctx)
}

// totalBondedTokens implements an alias call to the underlying staking keeper's
// TotalBondedTokens to be used in BeginBlocker.
func (k Keeper) totalBondedTokens(ctx sdk.Context) sdk.Int {
	return k.stakingKeeper.TotalBondedTokens(ctx)
}

// mintCoinsToModule mints teh desired coins to the x/kavamint module account and then
// transfers them to the designated module account.
// if `newCoins` is empty or zero, this method is a noop.
func (k Keeper) mintCoinsToModule(ctx sdk.Context, newCoins sdk.Coins, destMaccName string) error {
	if newCoins.IsZero() {
		// skip as no coins need to be minted
		return nil
	}

	// mint the coins
	err := k.bankKeeper.MintCoins(ctx, types.ModuleAccountName, newCoins)
	if err != nil {
		return nil
	}

	// transfer them to the desired destination module account
	return k.bankKeeper.SendCoinsFromModuleToModule(ctx, types.ModuleAccountName, destMaccName, newCoins)
}

// totalSupply implements an alias call to the underlying supply keeper's
// GetSupply for the mint denom to be used in calculating cumulative inflation.
func (k Keeper) totalSupply(ctx sdk.Context) sdk.Int {
	return k.bankKeeper.GetSupply(ctx, k.bondDenom(ctx)).Amount
}

func (k Keeper) CumulativeInflation(ctx sdk.Context) sdk.Dec {
	params := k.GetParams(ctx)
	totalInflation := sdk.NewDec(0)

	// community pool contribution is simply the inflation param
	totalInflation = totalInflation.Add(params.CommunityPoolInflation)

	// staking rewards contribution is the apy * bonded_ratio
	bondedSupply := k.totalBondedTokens(ctx)
	totalSupply := k.totalSupply(ctx)
	bondedRatio := sdk.NewDecFromInt(bondedSupply).QuoInt(totalSupply)
	inflationFromStakingRewards := params.StakingRewardsApy.Mul(bondedRatio)

	totalInflation = totalInflation.Add(inflationFromStakingRewards)

	return totalInflation
}

// GetPreviousBlockTime get the blocktime for the previous block
func (k Keeper) GetPreviousBlockTime(ctx sdk.Context) (blockTime time.Time) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.PreviousBlockTimeKey)

	b := store.Get(types.PreviousBlockTimeKey)
	if b == nil {
		return blockTime
	}

	if err := blockTime.UnmarshalBinary(b); err != nil {
		panic(err)
	}

	return
}

// SetPreviousBlockTime set the time of the previous block
func (k Keeper) SetPreviousBlockTime(ctx sdk.Context, blockTime time.Time) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.PreviousBlockTimeKey)

	if blockTime.IsZero() {
		store.Delete(types.PreviousBlockTimeKey)
		return
	}

	b, err := blockTime.MarshalBinary()
	if err != nil {
		panic(err)
	}
	store.Set(types.PreviousBlockTimeKey, b)
}
