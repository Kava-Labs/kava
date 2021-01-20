package keeper

import (
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/params/subspace"

	"github.com/kava-labs/kava/x/hard/types"
)

// Keeper keeper for the hard module
type Keeper struct {
	key             sdk.StoreKey
	cdc             *codec.Codec
	paramSubspace   subspace.Subspace
	accountKeeper   types.AccountKeeper
	supplyKeeper    types.SupplyKeeper
	stakingKeeper   types.StakingKeeper
	pricefeedKeeper types.PricefeedKeeper
	auctionKeeper   types.AuctionKeeper
}

// NewKeeper creates a new keeper
func NewKeeper(cdc *codec.Codec, key sdk.StoreKey, paramstore subspace.Subspace,
	ak types.AccountKeeper, sk types.SupplyKeeper, stk types.StakingKeeper,
	pfk types.PricefeedKeeper, auk types.AuctionKeeper) Keeper {
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
		auctionKeeper:   auk,
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

// GetDeposit returns a deposit from the store for a particular depositor address, deposit denom
func (k Keeper) GetDeposit(ctx sdk.Context, depositor sdk.AccAddress) (types.Deposit, bool) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.DepositsKeyPrefix)
	bz := store.Get(depositor.Bytes())
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
	store.Set(deposit.Depositor.Bytes(), bz)
}

// DeleteDeposit deletes a deposit from the store
func (k Keeper) DeleteDeposit(ctx sdk.Context, deposit types.Deposit) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.DepositsKeyPrefix)
	store.Delete(deposit.Depositor.Bytes())
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
	if borrowedCoins.Empty() {
		store.Set([]byte{}, []byte{})
	} else {
		bz := k.cdc.MustMarshalBinaryBare(borrowedCoins)
		store.Set([]byte{}, bz)
	}
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

// SetSuppliedCoins sets the total amount of coins currently supplied in the store
func (k Keeper) SetSuppliedCoins(ctx sdk.Context, suppliedCoins sdk.Coins) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.SuppliedCoinsPrefix)
	if suppliedCoins.Empty() {
		store.Set([]byte{}, []byte{})
	} else {
		bz := k.cdc.MustMarshalBinaryBare(suppliedCoins)
		store.Set([]byte{}, bz)
	}
}

// GetSuppliedCoins returns an sdk.Coins object from the store representing all currently supplied coins
func (k Keeper) GetSuppliedCoins(ctx sdk.Context) (sdk.Coins, bool) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.SuppliedCoinsPrefix)
	bz := store.Get([]byte{})
	if bz == nil {
		return sdk.Coins{}, false
	}
	var suppliedCoins sdk.Coins
	k.cdc.MustUnmarshalBinaryBare(bz, &suppliedCoins)
	return suppliedCoins, true
}

// GetMoneyMarket returns a money market from the store for a denom
func (k Keeper) GetMoneyMarket(ctx sdk.Context, denom string) (types.MoneyMarket, bool) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.MoneyMarketsPrefix)
	bz := store.Get([]byte(denom))
	if bz == nil {
		return types.MoneyMarket{}, false
	}
	var moneyMarket types.MoneyMarket
	k.cdc.MustUnmarshalBinaryBare(bz, &moneyMarket)
	return moneyMarket, true
}

// SetMoneyMarket sets a money market in the store for a denom
func (k Keeper) SetMoneyMarket(ctx sdk.Context, denom string, moneyMarket types.MoneyMarket) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.MoneyMarketsPrefix)
	bz := k.cdc.MustMarshalBinaryBare(moneyMarket)
	store.Set([]byte(denom), bz)
}

// DeleteMoneyMarket deletes a money market from the store
func (k Keeper) DeleteMoneyMarket(ctx sdk.Context, denom string) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.MoneyMarketsPrefix)
	store.Delete([]byte(denom))
}

// IterateMoneyMarkets iterates over all money markets objects in the store and performs a callback function
// 		that returns both the money market and the key (denom) it's stored under
func (k Keeper) IterateMoneyMarkets(ctx sdk.Context, cb func(denom string, moneyMarket types.MoneyMarket) (stop bool)) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.MoneyMarketsPrefix)
	iterator := sdk.KVStorePrefixIterator(store, []byte{})
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var moneyMarket types.MoneyMarket
		k.cdc.MustUnmarshalBinaryBare(iterator.Value(), &moneyMarket)
		if cb(string(iterator.Key()), moneyMarket) {
			break
		}
	}
}

// GetPreviousAccrualTime returns the last time an individual market accrued interest
func (k Keeper) GetPreviousAccrualTime(ctx sdk.Context, denom string) (time.Time, bool) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.PreviousAccrualTimePrefix)
	bz := store.Get([]byte(denom))
	if bz == nil {
		return time.Time{}, false
	}
	var previousAccrualTime time.Time
	k.cdc.MustUnmarshalBinaryBare(bz, &previousAccrualTime)
	return previousAccrualTime, true
}

// SetPreviousAccrualTime sets the most recent accrual time for a particular market
func (k Keeper) SetPreviousAccrualTime(ctx sdk.Context, denom string, previousAccrualTime time.Time) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.PreviousAccrualTimePrefix)
	bz := k.cdc.MustMarshalBinaryBare(previousAccrualTime)
	store.Set([]byte(denom), bz)
}

// GetTotalReserves returns the total reserves for an individual market
func (k Keeper) GetTotalReserves(ctx sdk.Context, denom string) (sdk.Coin, bool) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.TotalReservesPrefix)
	bz := store.Get([]byte(denom))
	if bz == nil {
		return sdk.Coin{}, false
	}
	var totalReserves sdk.Coin
	k.cdc.MustUnmarshalBinaryBare(bz, &totalReserves)
	return totalReserves, true
}

// SetTotalReserves sets the total reserves for an individual market
func (k Keeper) SetTotalReserves(ctx sdk.Context, denom string, coin sdk.Coin) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.TotalReservesPrefix)
	bz := k.cdc.MustMarshalBinaryBare(coin)
	store.Set([]byte(denom), bz)
}

// GetBorrowInterestFactor returns the current borrow interest factor for an individual market
func (k Keeper) GetBorrowInterestFactor(ctx sdk.Context, denom string) (sdk.Dec, bool) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.BorrowInterestFactorPrefix)
	bz := store.Get([]byte(denom))
	if bz == nil {
		return sdk.ZeroDec(), false
	}
	var borrowInterestFactor sdk.Dec
	k.cdc.MustUnmarshalBinaryBare(bz, &borrowInterestFactor)
	return borrowInterestFactor, true
}

// SetBorrowInterestFactor sets the current borrow interest factor for an individual market
func (k Keeper) SetBorrowInterestFactor(ctx sdk.Context, denom string, borrowInterestFactor sdk.Dec) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.BorrowInterestFactorPrefix)
	bz := k.cdc.MustMarshalBinaryBare(borrowInterestFactor)
	store.Set([]byte(denom), bz)
}

// GetSupplyInterestFactor returns the current supply interest factor for an individual market
func (k Keeper) GetSupplyInterestFactor(ctx sdk.Context, denom string) (sdk.Dec, bool) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.SupplyInterestFactorPrefix)
	bz := store.Get([]byte(denom))
	if bz == nil {
		return sdk.ZeroDec(), false
	}
	var supplyInterestFactor sdk.Dec
	k.cdc.MustUnmarshalBinaryBare(bz, &supplyInterestFactor)
	return supplyInterestFactor, true
}

// SetSupplyInterestFactor sets the current supply interest factor for an individual market
func (k Keeper) SetSupplyInterestFactor(ctx sdk.Context, denom string, supplyInterestFactor sdk.Dec) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.SupplyInterestFactorPrefix)
	bz := k.cdc.MustMarshalBinaryBare(supplyInterestFactor)
	store.Set([]byte(denom), bz)
}

// InsertIntoLtvIndex indexes a user's borrow object by its current LTV
func (k Keeper) InsertIntoLtvIndex(ctx sdk.Context, ltv sdk.Dec, borrower sdk.AccAddress) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.LtvIndexPrefix)
	store.Set(types.GetBorrowByLtvKey(ltv, borrower), borrower)
}

// RemoveFromLtvIndex removes a user's borrow object from the LTV index
func (k Keeper) RemoveFromLtvIndex(ctx sdk.Context, ltv sdk.Dec, borrower sdk.AccAddress) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.LtvIndexPrefix)
	store.Delete(types.GetBorrowByLtvKey(ltv, borrower))
}

// IterateLtvIndex provides an iterator over the borrowers ordered by LTV.
// For results found before the cutoff count, the cb will be called and the item returned.
func (k Keeper) IterateLtvIndex(ctx sdk.Context, cutoffCount int,
	cb func(addr sdk.AccAddress) (stop bool)) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.LtvIndexPrefix)
	iterator := store.ReverseIterator(nil, nil)
	count := 0

	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {

		// Stop iteration after first 10 items
		count = count + 1
		if count > cutoffCount {
			break
		}

		id := iterator.Value()
		cb(id)
	}
}

// GetLtvIndexSlice returns the first 10 items in the LTV index from the store
func (k Keeper) GetLtvIndexSlice(ctx sdk.Context, count int) (addrs []sdk.AccAddress) {
	k.IterateLtvIndex(ctx, count, func(addr sdk.AccAddress) bool {
		addrs = append(addrs, addr)
		return false
	})
	return
}
