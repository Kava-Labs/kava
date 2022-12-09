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

type KeeperI interface {
	GetParams(ctx sdk.Context) (params types.Params)
	SetParams(ctx sdk.Context, params types.Params)

	BondDenom(ctx sdk.Context) string
	TotalSupply(ctx sdk.Context) sdk.Int
	TotalBondedTokens(ctx sdk.Context) sdk.Int

	MintCoins(ctx sdk.Context, newCoins sdk.Coins) error
	AddCollectedFees(ctx sdk.Context, fees sdk.Coins) error
	FundCommunityPool(ctx sdk.Context, funds sdk.Coins) error

	GetPreviousBlockTime(ctx sdk.Context) (blockTime time.Time)
	SetPreviousBlockTime(ctx sdk.Context, blockTime time.Time)

	CumulativeInflation(ctx sdk.Context) sdk.Dec
	AccumulateInflation(
		ctx sdk.Context,
		rate sdk.Dec,
		basis sdk.Int,
		secondsSinceLastMint float64,
	) (sdk.Coins, error)
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

// BondDenom implements an alias call to the underlying staking keeper's BondDenom.
func (k Keeper) BondDenom(ctx sdk.Context) string {
	return k.stakingKeeper.BondDenom(ctx)
}

// TotalBondedTokens implements an alias call to the underlying staking keeper's
// TotalBondedTokens to be used in BeginBlocker.
func (k Keeper) TotalBondedTokens(ctx sdk.Context) sdk.Int {
	return k.stakingKeeper.TotalBondedTokens(ctx)
}

// MintCoins implements an alias call to the underlying supply keeper's
// MintCoins to be used in BeginBlocker.
func (k Keeper) MintCoins(ctx sdk.Context, newCoins sdk.Coins) error {
	if newCoins.Empty() {
		// skip as no coins need to be minted
		return nil
	}

	return k.bankKeeper.MintCoins(ctx, types.ModuleName, newCoins)
}

// AddCollectedFees implements an alias call to the underlying supply keeper's
// AddCollectedFees to be used in BeginBlocker.
func (k Keeper) AddCollectedFees(ctx sdk.Context, fees sdk.Coins) error {
	return k.bankKeeper.SendCoinsFromModuleToModule(ctx, types.ModuleName, k.stakingRewardsFeeCollectorName, fees)
}

// FundCommunityPool implements an alias call to the underlying supply keeper's
// FundCommunityPool to be used in BeginBlocker.
func (k Keeper) FundCommunityPool(ctx sdk.Context, funds sdk.Coins) error {
	return k.bankKeeper.SendCoinsFromModuleToModule(ctx, types.ModuleName, k.communityPoolModuleAccountName, funds)
}

// TotalSupply implements an alias call to the underlying supply keeper's
// GetSupply for the mint denom to be used in calculating cumulative inflation.
func (k Keeper) TotalSupply(ctx sdk.Context) sdk.Int {
	return k.bankKeeper.GetSupply(ctx, k.BondDenom(ctx)).Amount
}

func (k Keeper) CumulativeInflation(ctx sdk.Context) sdk.Dec {
	params := k.GetParams(ctx)
	totalInflation := sdk.NewDec(0)

	// community pool contribution is simply the inflation param
	totalInflation = totalInflation.Add(params.CommunityPoolInflation)

	// staking rewards contribution is the apy * bonded_ratio
	bondedSupply := k.TotalBondedTokens(ctx)
	totalSupply := k.TotalSupply(ctx)
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
