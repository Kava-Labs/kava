package keeper

import (
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/params/subspace"

	"github.com/kava-labs/kava/x/harvest/types"
)

// Keeper keeper for the harvest module
type Keeper struct {
	key             sdk.StoreKey
	cdc             *codec.Codec
	paramSubspace   subspace.Subspace
	accountKeeper   types.AccountKeeper
	supplyKeeper    types.SupplyKeeper
	stakingKeeper   types.StakingKeeper
	pricefeedKeeper types.PricefeedKeeper
}

// NewKeeper creates a new keeper
func NewKeeper(cdc *codec.Codec, key sdk.StoreKey, paramstore subspace.Subspace,
	ak types.AccountKeeper, sk types.SupplyKeeper, stk types.StakingKeeper,
	pfk types.PricefeedKeeper) Keeper {
	if !paramstore.HasKeyTable() {
		paramstore = paramstore.WithKeyTable(types.ParamKeyTable())
	}

	return Keeper{
		key:             key,
		cdc:             cdc,
		paramSubspace:   paramstore,
		accountKeeper:   ak,
		supplyKeeper:    sk,
		stakingKeeper:   stk,
		pricefeedKeeper: pfk,
	}
}

// GetPreviousBlockTime get the blocktime for the previous block
func (k Keeper) GetPreviousBlockTime(ctx sdk.Context) (blockTime time.Time, found bool) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.PreviousBlockTimeKey)
	b := store.Get([]byte{})
	if b == nil {
		return time.Time{}, false
	}
	k.cdc.MustUnmarshalBinaryBare(b, &blockTime)
	return blockTime, true
}

// SetPreviousBlockTime set the time of the previous block
func (k Keeper) SetPreviousBlockTime(ctx sdk.Context, blockTime time.Time) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.PreviousBlockTimeKey)
	store.Set([]byte{}, k.cdc.MustMarshalBinaryBare(blockTime))
}

// GetPreviousDelegatorDistribution get the time of the previous delegator distribution
func (k Keeper) GetPreviousDelegatorDistribution(ctx sdk.Context, denom string) (distTime time.Time, found bool) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.PreviousDelegationDistributionKey)
	bz := store.Get([]byte(denom))
	if bz == nil {
		return time.Time{}, false
	}
	k.cdc.MustUnmarshalBinaryBare(bz, &distTime)
	return distTime, true
}

// SetPreviousDelegationDistribution set the time of the previous delegator distribution
func (k Keeper) SetPreviousDelegationDistribution(ctx sdk.Context, distTime time.Time, denom string) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.PreviousDelegationDistributionKey)
	store.Set([]byte(denom), k.cdc.MustMarshalBinaryBare(distTime))
}

// GetDeposit returns a deposit from the store for a particular depositor address, deposit denom
func (k Keeper) GetDeposit(ctx sdk.Context, depositor sdk.AccAddress, denom string) (types.Deposit, bool) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.DepositsKeyPrefix)
	bz := store.Get(types.DepositKey(denom, depositor))
	if bz == nil {
		return types.Deposit{}, false
	}
	var deposit types.Deposit
	k.cdc.MustUnmarshalBinaryBare(bz, &deposit)
	return deposit, true
}

// SetDeposit sets the input deposit in the store, prefixed by the deposit type, deposit denom, and depositor address, in that order
func (k Keeper) SetDeposit(ctx sdk.Context, deposit types.Deposit) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.DepositsKeyPrefix)
	bz := k.cdc.MustMarshalBinaryBare(deposit)
	store.Set(types.DepositKey(deposit.Amount.Denom, deposit.Depositor), bz)
}

// DeleteDeposit deletes a deposit from the store
func (k Keeper) DeleteDeposit(ctx sdk.Context, deposit types.Deposit) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.DepositsKeyPrefix)
	store.Delete(types.DepositKey(deposit.Amount.Denom, deposit.Depositor))
}

// IterateDeposits iterates over all deposit objects in the store and performs a callback function
func (k Keeper) IterateDeposits(ctx sdk.Context, cb func(deposit types.Deposit) (stop bool)) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.DepositsKeyPrefix)
	iterator := sdk.KVStorePrefixIterator(store, []byte{})
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var deposit types.Deposit
		k.cdc.MustUnmarshalBinaryBare(iterator.Value(), &deposit)
		if cb(deposit) {
			break
		}
	}
}

// IterateDepositsByDenom iterates over all deposit objects in the store with the matching deposit denom and performs a callback function
func (k Keeper) IterateDepositsByDenom(ctx sdk.Context, depositDenom string, cb func(deposit types.Deposit) (stop bool)) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.DepositsKeyPrefix)
	iterator := sdk.KVStorePrefixIterator(store, types.DepositTypeIteratorKey(depositDenom))
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var deposit types.Deposit
		k.cdc.MustUnmarshalBinaryBare(iterator.Value(), &deposit)
		if cb(deposit) {
			break
		}
	}
}

// GetClaim returns a claim from the store for a particular claim owner, deposit denom, and claim type
func (k Keeper) GetClaim(ctx sdk.Context, owner sdk.AccAddress, depositDenom string, claimType types.ClaimType) (types.Claim, bool) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.ClaimsKeyPrefix)
	bz := store.Get(types.ClaimKey(claimType, depositDenom, owner))
	if bz == nil {
		return types.Claim{}, false
	}
	var claim types.Claim
	k.cdc.MustUnmarshalBinaryBare(bz, &claim)
	return claim, true
}

// SetClaim stores the input claim in the store, prefixed by the deposit type, deposit denom, and owner address, in that order
func (k Keeper) SetClaim(ctx sdk.Context, claim types.Claim) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.ClaimsKeyPrefix)
	bz := k.cdc.MustMarshalBinaryBare(claim)
	store.Set(types.ClaimKey(claim.Type, claim.DepositDenom, claim.Owner), bz)
}

// DeleteClaim deletes a claim from the store
func (k Keeper) DeleteClaim(ctx sdk.Context, claim types.Claim) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.ClaimsKeyPrefix)
	store.Delete(types.ClaimKey(claim.Type, claim.DepositDenom, claim.Owner))
}

// IterateClaims iterates over all claim objects in the store and performs a callback function
func (k Keeper) IterateClaims(ctx sdk.Context, cb func(claim types.Claim) (stop bool)) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.ClaimsKeyPrefix)
	iterator := sdk.KVStorePrefixIterator(store, []byte{})
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var claim types.Claim
		k.cdc.MustUnmarshalBinaryBare(iterator.Value(), &claim)
		if cb(claim) {
			break
		}
	}
}

// IterateClaimsByTypeAndDenom iterates over all claim objects in the store with the matching claim type and deposit denom and performs a callback function
func (k Keeper) IterateClaimsByTypeAndDenom(ctx sdk.Context, claimType types.ClaimType, depositDenom string, cb func(claim types.Claim) (stop bool)) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.ClaimsKeyPrefix)
	iterator := sdk.KVStorePrefixIterator(store, types.ClaimTypeIteratorKey(claimType, depositDenom))
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var claim types.Claim
		k.cdc.MustUnmarshalBinaryBare(iterator.Value(), &claim)
		if cb(claim) {
			break
		}
	}
}

// GetDepositsByUser gets all deposits for an individual user
func (k Keeper) GetDepositsByUser(ctx sdk.Context, user sdk.AccAddress) []types.Deposit {
	var deposits []types.Deposit
	k.IterateDeposits(ctx, func(deposit types.Deposit) (stop bool) {
		if deposit.Depositor.Equals(user) {
			deposits = append(deposits, deposit)
		}
		return false
	})
	return deposits
}

// BondDenom returns the bond denom from the staking keeper
func (k Keeper) BondDenom(ctx sdk.Context) string {
	return k.stakingKeeper.BondDenom(ctx)
}

// GetBorrow returns a Borrow from the store for a particular borrower address and borrow denom
func (k Keeper) GetBorrow(ctx sdk.Context, borrower sdk.AccAddress) (types.Borrow, bool) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.BorrowsKeyPrefix)
	bz := store.Get(borrower)
	if bz == nil {
		return types.Borrow{}, false
	}
	var borrow types.Borrow
	k.cdc.MustUnmarshalBinaryBare(bz, &borrow)
	return borrow, true
}

// SetBorrow sets the input borrow in the store, prefixed by the borrower address and borrow denom
func (k Keeper) SetBorrow(ctx sdk.Context, borrow types.Borrow) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.BorrowsKeyPrefix)
	bz := k.cdc.MustMarshalBinaryBare(borrow)
	store.Set(borrow.Borrower, bz)
}

// DeleteBorrow deletes a borrow from the store
func (k Keeper) DeleteBorrow(ctx sdk.Context, borrow types.Borrow) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.BorrowsKeyPrefix)
	store.Delete(borrow.Borrower)
}

// IterateBorrows iterates over all borrow objects in the store and performs a callback function
func (k Keeper) IterateBorrows(ctx sdk.Context, cb func(borrow types.Borrow) (stop bool)) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.BorrowsKeyPrefix)
	iterator := sdk.KVStorePrefixIterator(store, []byte{})
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var borrow types.Borrow
		k.cdc.MustUnmarshalBinaryBare(iterator.Value(), &borrow)
		if cb(borrow) {
			break
		}
	}
}

// SetBorrowedCoins sets the total amount of coins currently borrowed in the store
func (k Keeper) SetBorrowedCoins(ctx sdk.Context, borrowedCoins sdk.Coins) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.BorrowedCoinsPrefix)
	bz := k.cdc.MustMarshalBinaryBare(borrowedCoins)
	store.Set([]byte{}, bz)
}

// GetBorrowedCoins returns an sdk.Coins object from the store representing all currently borrowed coins
func (k Keeper) GetBorrowedCoins(ctx sdk.Context) (sdk.Coins, bool) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.BorrowedCoinsPrefix)
	bz := store.Get([]byte{})
	if bz == nil {
		return sdk.Coins{}, false
	}
	var borrowedCoins sdk.Coins
	k.cdc.MustUnmarshalBinaryBare(bz, &borrowedCoins)
	return borrowedCoins, true
}

// GetInterestRateModel returns an interest rate model from the store for a denom
func (k Keeper) GetInterestRateModel(ctx sdk.Context, denom string) (types.InterestRateModel, bool) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.InterestRateModelsPrefix)
	bz := store.Get([]byte(denom))
	if bz == nil {
		return types.InterestRateModel{}, false
	}
	var interestRateModel types.InterestRateModel
	k.cdc.MustUnmarshalBinaryBare(bz, &interestRateModel)
	return interestRateModel, true
}

// SetInterestRateModel sets an interest rate model in the store for a denom
func (k Keeper) SetInterestRateModel(ctx sdk.Context, denom string, interestRateModel types.InterestRateModel) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.InterestRateModelsPrefix)
	bz := k.cdc.MustMarshalBinaryBare(interestRateModel)
	store.Set([]byte(denom), bz)
}

// IterateInterestRateModels iterates over all interest rate model objects in the store and performs a callback function
func (k Keeper) IterateInterestRateModels(ctx sdk.Context, cb func(interestRateModel types.InterestRateModel) (stop bool)) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.InterestRateModelsPrefix)
	iterator := sdk.KVStorePrefixIterator(store, []byte{})
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var interestRateModel types.InterestRateModel
		k.cdc.MustUnmarshalBinaryBare(iterator.Value(), &interestRateModel)
		if cb(interestRateModel) {
			break
		}
	}
}
