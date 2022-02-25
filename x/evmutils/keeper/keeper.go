package keeper

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/kava-labs/kava/x/evmutils/types"
)

type Keeper struct {
	storeKey sdk.StoreKey
	bk       types.BankKeeper
	ak       types.AccountKeeper
}

func NewKeeper(storeKey sdk.StoreKey, bk types.BankKeeper, ak types.AccountKeeper) Keeper {
	return Keeper{
		storeKey: storeKey,
		bk:       bk,
		ak:       ak,
	}
}

// GetBalance returns the total balance of akava for a given account by address.
func (k Keeper) GetBalance(ctx sdk.Context, addr sdk.AccAddress) sdk.Int {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.AccAddressKeyPrefix)
	var bal sdk.Int
	bz := store.Get(addr)
	if bz == nil {
		return sdk.ZeroInt()
	}
	if err := bal.Unmarshal(bz); err != nil {
		panic(err)
	}
	return bal
}

// SetBalance sets the total balance of akava for a given account by address.
func (k Keeper) SetBalance(ctx sdk.Context, addr sdk.AccAddress, bal sdk.Int) error {
	if bal.IsNegative() {
		return fmt.Errorf("cannot set akava balance to a negative amount: %d", bal)
	}
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.AccAddressKeyPrefix)
	bz, err := bal.Marshal()
	if err != nil {
		panic(err)
	}
	store.Set(addr, bz)
	return nil
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
	finalBal := amt.Sub(bal)
	if finalBal.IsNegative() {
		return sdkerrors.Wrapf(sdkerrors.ErrInsufficientFunds, "insufficient funds to send %s", amt.String())
	}
	return k.SetBalance(ctx, addr, finalBal)
}
