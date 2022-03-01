package keeper

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/kava-labs/kava/x/evmutil/types"
)

type Keeper struct {
	cdc      codec.Codec
	storeKey sdk.StoreKey
}

func NewKeeper(cdc codec.Codec, storeKey sdk.StoreKey) Keeper {
	return Keeper{
		cdc:      cdc,
		storeKey: storeKey,
	}
}

// GetAllAccounts returns all the account balances for the given account address.
func (k Keeper) GetAllAccounts(ctx sdk.Context) (accounts []types.Account) {
	k.IterateAllAccounts(ctx, func(account types.Account) bool {
		accounts = append(accounts, account)
		return false
	})
	return accounts
}

// IterateAllAccounts iterates over all accounts. If true is returned from the
// callback, iteration is halted.
func (k Keeper) IterateAllAccounts(ctx sdk.Context, cb func(types.Account) bool) {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.AccountStoreKeyPrefix)

	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var acc types.Account
		if err := k.cdc.Unmarshal(iterator.Value(), &acc); err != nil {
			panic(err)
		}
		if cb(acc) {
			break
		}
	}
}

func (k Keeper) GetAccount(ctx sdk.Context, addr sdk.AccAddress) *types.Account {
	store := ctx.KVStore(k.storeKey)
	var account types.Account
	bz := store.Get(types.AccountStoreKey(addr))
	if bz == nil {
		return nil
	}
	if err := account.Unmarshal(bz); err != nil {
		panic(err)
	}
	return &account
}

// SetBalance sets the total balance of akava for a given account by address.
func (k Keeper) SetAccount(ctx sdk.Context, account types.Account) error {
	if err := account.Validate(); err != nil {
		return err
	}
	store := ctx.KVStore(k.storeKey)
	accountKey := types.AccountStoreKey(account.Address)

	bz, err := k.cdc.Marshal(&account)
	if err != nil {
		panic(err)
	}
	store.Set(accountKey, bz)
	return nil
}

// GetBalance returns the total balance of akava for a given account by address.
func (k Keeper) GetBalance(ctx sdk.Context, addr sdk.AccAddress) sdk.Int {
	account := k.GetAccount(ctx, addr)
	if account == nil {
		return sdk.ZeroInt()
	}
	return account.Balance
}

// SetBalance sets the total balance of akava for a given account by address.
func (k Keeper) SetBalance(ctx sdk.Context, addr sdk.AccAddress, bal sdk.Int) error {
	account := k.GetAccount(ctx, addr)
	if account == nil {
		account = types.NewAccount(addr, bal)
	} else {
		account.Balance = bal
	}

	if err := account.Validate(); err != nil {
		return err
	}

	return k.SetAccount(ctx, *account)
}

// SendBalance transfers the akava balance from sender addr to recipient addr.
func (k Keeper) SendBalance(ctx sdk.Context, senderAddr sdk.AccAddress, recipientAddr sdk.AccAddress, amt sdk.Int) error {
	if amt.IsNegative() {
		return fmt.Errorf("cannot send a negative amount of akava: %d", amt)
	}

	if amt.IsZero() {
		return nil
	}

	senderBal := k.GetBalance(ctx, senderAddr)
	if senderBal.LT(amt) {
		return sdkerrors.Wrapf(sdkerrors.ErrInsufficientFunds, "insufficient funds to send %s", amt.String())
	}
	if err := k.SetBalance(ctx, senderAddr, senderBal.Sub(amt)); err != nil {
		return err
	}

	receiverBal := k.GetBalance(ctx, recipientAddr).Add(amt)
	return k.SetBalance(ctx, recipientAddr, receiverBal)
}

// AddBalance increments the akava balance of an address.
func (k Keeper) AddBalance(ctx sdk.Context, addr sdk.AccAddress, amt sdk.Int) error {
	bal := k.GetBalance(ctx, addr)
	return k.SetBalance(ctx, addr, amt.Add(bal))
}

// RemoveBalance decrements the akava balance of an address.
func (k Keeper) RemoveBalance(ctx sdk.Context, addr sdk.AccAddress, amt sdk.Int) error {
	bal := k.GetBalance(ctx, addr)
	finalBal := bal.Sub(amt)
	if finalBal.IsNegative() {
		return sdkerrors.Wrapf(sdkerrors.ErrInsufficientFunds, "insufficient funds to send %s", amt.String())
	}
	return k.SetBalance(ctx, addr, finalBal)
}
