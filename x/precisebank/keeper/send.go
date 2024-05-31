package keeper

import (
	"errors"
	"fmt"

	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/kava-labs/kava/x/precisebank/types"
)

// IsSendEnabledCoins uses the parent x/bank keeper to check the coins provided
// and returns an ErrSendDisabled if any of the coins are not configured for
// sending. Returns nil if sending is enabled for all provided coin
func (k Keeper) IsSendEnabledCoins(ctx sdk.Context, coins ...sdk.Coin) error {
	// TODO: This does not actually seem to be used by x/evm, so it should be
	// removed from the expected_interface in x/evm.
	return k.bk.IsSendEnabledCoins(ctx, coins...)
}

// SendCoins transfers amt coins from a sending account to a receiving account.
// An error is returned upon failure. This handles transfers including
// ExtendedCoinDenom and supports non-ExtendedCoinDenom transfers by passing
// through to x/bank.
func (k Keeper) SendCoins(
	ctx sdk.Context,
	from, to sdk.AccAddress,
	amt sdk.Coins,
) error {
	// IsSendEnabledCoins() is only used in x/bank in msg server, not in keeper,
	// so we should also not use it here to align with x/bank behavior.

	if !amt.IsValid() {
		return errorsmod.Wrap(sdkerrors.ErrInvalidCoins, amt.String())
	}

	passthroughCoins := amt
	extendedCoinAmount := amt.AmountOf(types.ExtendedCoinDenom)

	// Remove the extended coin amount from the passthrough coins
	if extendedCoinAmount.IsPositive() {
		subCoin := sdk.NewCoin(types.ExtendedCoinDenom, extendedCoinAmount)
		passthroughCoins = amt.Sub(subCoin)
	}

	// Send the passthrough coins through x/bank
	if passthroughCoins.IsAllPositive() {
		if err := k.bk.SendCoins(ctx, from, to, passthroughCoins); err != nil {
			return err
		}
	}

	// If there is no extended coin amount, we are done
	if extendedCoinAmount.IsZero() {
		return nil
	}

	// Send the extended coin amount through x/precisebank
	return k.sendExtendedCoins(ctx, from, to, extendedCoinAmount)
}

// sendExtendedCoins transfers amt extended coins from a sending account to a
// receiving account. An error is returned upon failure. This function is
// called by SendCoins() and should not be called directly.
//
// This method covers 4 cases between two accounts - sender and receiver.
// Depending on the fractional balance and the amount being transferred:
// Sender account:
//  1. Arithmetic borrow 1 integer equivalent amount of fractional coins if the
//     fractional balance is insufficient.
//  2. No borrow if fractional balance is sufficient.
//
// Receiver account:
//  1. Arithmetic carry 1 integer equivalent amount of fractional coins if the
//     received amount exceeds max fractional balance.
//  2. No carry if received amount does not exceed max fractional balance.
//
// The 4 cases are:
// 1. Sender borrow, receiver carry
// 2. Sender borrow, NO receiver carry
// 3. NO sender borrow, receiver carry
// 4. NO sender borrow, NO receiver carry
//
// Truth table:
// | Sender Borrow | Receiver Carry | Direct Transfer |
// | --------------|----------------|-----------------|
// | T             | T              | T               |
// | T             | F              | F               |
// | F             | T              | F               |
// | F             | F              | F               |
func (k Keeper) sendExtendedCoins(
	ctx sdk.Context,
	from, to sdk.AccAddress,
	amt sdkmath.Int,
) error {
	// Sufficient balance check is done by bankkeeper.SendCoins(), for both
	// integer and fractional-only sends. E.g. If fractional balance is
	// insufficient, it will still incur a integer borrow which will fail if the
	// sender does not have sufficient integer balance.

	// Load required state: Account old balances
	senderFracBal := k.GetFractionalBalance(ctx, from)
	recipientFracBal := k.GetFractionalBalance(ctx, to)

	// -------------------------------------------------------------------------
	// Pure stateless calculations
	integerAmt := amt.Quo(types.ConversionFactor())
	fractionalAmt := amt.Mod(types.ConversionFactor())

	// Account new fractional balances
	senderNewFracBal, senderNeedsBorrow := subFromFractionalBalance(senderFracBal, fractionalAmt)
	recipientNewFracBal, recipientNeedsCarry := addToFractionalBalance(recipientFracBal, fractionalAmt)

	// Case #1: Sender borrow, recipient carry
	if senderNeedsBorrow && recipientNeedsCarry {
		// Can directly transfer borrow/carry - increase the direct transfer by 1
		integerAmt = integerAmt.AddRaw(1)
	}

	// -------------------------------------------------------------------------
	// Stateful operations for transfers

	// This includes ALL transfers of >= conversionFactor AND Case #1
	// Full integer amount transfer, including direct transfer of borrow/carry
	// if any.
	if integerAmt.IsPositive() {
		transferCoin := sdk.NewCoin(types.IntegerCoinDenom, integerAmt)
		if err := k.bk.SendCoins(ctx, from, to, sdk.NewCoins(transferCoin)); err != nil {
			return k.wrapError(ctx, from, amt, err)
		}
	}

	// Case #2: Sender borrow, NO recipient carry
	// Sender borrows by transferring 1 integer amount to reserve to account for
	// lack of fractional balance.
	if senderNeedsBorrow && !recipientNeedsCarry {
		borrowCoin := sdk.NewCoin(types.IntegerCoinDenom, sdk.NewInt(1))
		if err := k.bk.SendCoinsFromAccountToModule(
			ctx,
			from, // sender borrowing
			types.ModuleName,
			sdk.NewCoins(borrowCoin),
		); err != nil {
			return k.wrapError(ctx, from, amt, err)
		}
	}

	// Case #3: NO sender borrow, recipient carry.
	// Recipient's fractional balance carries over to integer balance by 1.
	// Always send carry from reserve before receiving borrow from sender to
	// ensure reserve always has sufficient balance starting from 0.
	if !senderNeedsBorrow && recipientNeedsCarry {
		carryCoin := sdk.NewCoin(types.IntegerCoinDenom, sdk.NewInt(1))
		if err := k.bk.SendCoinsFromModuleToAccount(
			ctx,
			types.ModuleName,
			to, // recipient carrying
			sdk.NewCoins(carryCoin),
		); err != nil {
			// Panic instead of returning error, as this will only error
			// with invalid state or logic. Reserve should always have
			// sufficient balance to carry fractional coins.
			panic(fmt.Errorf("failed to carry fractional coins to %s: %w", to, err))
		}
	}

	// Update fractional balances
	k.SetFractionalBalance(ctx, from, senderNewFracBal)
	k.SetFractionalBalance(ctx, to, recipientNewFracBal)

	return nil
}

// subFromFractionalBalance subtracts a fractional amount from the provided
// current fractional balance, returning the new fractional balance and true if
// an integer borrow is required.
func subFromFractionalBalance(
	currentFractionalBalance sdkmath.Int,
	amountToSub sdkmath.Int,
) (sdkmath.Int, bool) {
	// Enforce that currentFractionalBalance is not a full balance.
	if currentFractionalBalance.GTE(types.ConversionFactor()) {
		panic("currentFractionalBalance must be less than ConversionFactor")
	}

	if amountToSub.GTE(types.ConversionFactor()) {
		panic("amountToSub must be less than ConversionFactor")
	}

	newFractionalBalance := currentFractionalBalance.Sub(amountToSub)

	// Insufficient fractional balance, so we need to borrow.
	borrowRequired := newFractionalBalance.IsNegative()

	if borrowRequired {
		// Borrowing 1 integer equivalent amount of fractional coins. We need to
		// add 1 integer equivalent amount to the fractional balance otherwise
		// the new fractional balance will be negative.
		newFractionalBalance = newFractionalBalance.Add(types.ConversionFactor())
	}

	return newFractionalBalance, borrowRequired
}

// addToFractionalBalance adds a fractional amount to the provided current
// fractional balance, returning the new fractional balance and true if a carry
// is required.
func addToFractionalBalance(
	currentFractionalBalance sdkmath.Int,
	amountToAdd sdkmath.Int,
) (sdkmath.Int, bool) {
	// Enforce that currentFractionalBalance is not a full balance.
	if currentFractionalBalance.GTE(types.ConversionFactor()) {
		panic("currentFractionalBalance must be less than ConversionFactor")
	}

	if amountToAdd.GTE(types.ConversionFactor()) {
		panic("amountToAdd must be less than ConversionFactor")
	}

	newFractionalBalance := currentFractionalBalance.Add(amountToAdd)

	// New balance exceeds max fractional balance, so we need to carry it over
	// to the integer balance.
	carryRequired := newFractionalBalance.GTE(types.ConversionFactor())

	if carryRequired {
		// Carry over to integer amount
		newFractionalBalance = newFractionalBalance.Sub(types.ConversionFactor())
	}

	return newFractionalBalance, carryRequired
}

// SendCoinsFromModuleToModule transfers coins from a ModuleAccount to another.
// It will panic if either module account does not exist. An error is returned
// if the recipient module is the x/precisebank module account or if sending the
// tokens fails.
func (k Keeper) SendCoinsFromAccountToModule(
	ctx sdk.Context,
	senderAddr sdk.AccAddress,
	recipientModule string,
	amt sdk.Coins,
) error {
	recipientAcc := k.ak.GetModuleAccount(ctx, recipientModule)
	if recipientAcc == nil {
		panic(errorsmod.Wrapf(sdkerrors.ErrUnknownAddress, "module account %s does not exist", recipientModule))
	}

	if recipientModule == types.ModuleName {
		return errorsmod.Wrapf(sdkerrors.ErrUnauthorized, "module account %s is not allowed to receive funds", types.ModuleName)
	}

	return k.SendCoins(ctx, senderAddr, recipientAcc.GetAddress(), amt)
}

// SendCoinsFromModuleToAccount transfers coins from a ModuleAccount to an AccAddress.
// It will panic if the module account does not exist. An error is returned if
// the recipient address is blocked, if the sender is the x/precisebank module
// account, or if sending the tokens fails.
func (k Keeper) SendCoinsFromModuleToAccount(
	ctx sdk.Context,
	senderModule string,
	recipientAddr sdk.AccAddress,
	amt sdk.Coins,
) error {
	// Identical panics to x/bank
	senderAddr := k.ak.GetModuleAddress(senderModule)
	if senderAddr == nil {
		panic(errorsmod.Wrapf(sdkerrors.ErrUnknownAddress, "module account %s does not exist", senderModule))
	}

	// Custom error to prevent external modules from modifying x/precisebank
	// balances. x/precisebank module account balance is for internal reserve
	// use only.
	if senderModule == types.ModuleName {
		return errorsmod.Wrapf(sdkerrors.ErrUnauthorized, "module account %s is not allowed to send funds", types.ModuleName)
	}

	// Uses x/bank BlockedAddr, no need to modify. x/precisebank should be
	// blocked.
	if k.bk.BlockedAddr(recipientAddr) {
		return errorsmod.Wrapf(sdkerrors.ErrUnauthorized, "%s is not allowed to receive funds", recipientAddr)
	}

	return k.SendCoins(ctx, senderAddr, recipientAddr, amt)
}

// wrapError returns a modified ErrInsufficientFunds with extended coin amounts
// if the error is due to insufficient funds. Otherwise, it returns the original
// error.
func (k Keeper) wrapError(
	ctx sdk.Context,
	addr sdk.AccAddress,
	amt sdkmath.Int,
	err error,
) error {
	if !errors.Is(err, sdkerrors.ErrInsufficientFunds) {
		return err
	}

	// Check balance is sufficient
	bal := k.GetBalance(ctx, addr, types.ExtendedCoinDenom)
	coin := sdk.NewCoin(types.ExtendedCoinDenom, amt)

	// TODO: This checks spendable coins and returns error with spendable
	// coins, not full balance. If GetBalance() is modified to return the
	// full, including locked, balance then this should be updated to deduct
	// locked coins.

	// Use sdk.NewCoins() so that it removes empty balances - ie. prints
	// empty string if balance is 0. This is to match x/bank behavior.
	spendable := sdk.NewCoins(bal)

	return errorsmod.Wrapf(
		sdkerrors.ErrInsufficientFunds,
		"spendable balance %s is smaller than %s",
		spendable, coin,
	)
}
