package keeper

import (
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

func (k Keeper) sendExtendedCoins(
	ctx sdk.Context,
	from, to sdk.AccAddress,
	amt sdkmath.Int,
) error {
	// Check balance is sufficient
	bal := k.GetBalance(ctx, from, types.ExtendedCoinDenom)
	if bal.Amount.LT(amt) {
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

	integerAmt := amt.Quo(types.ConversionFactor())
	fractionalAmt := amt.Mod(types.ConversionFactor())

	// Account old balances
	senderFractionalBal := k.GetFractionalBalance(ctx, from)
	recipientFractionalBal := k.GetFractionalBalance(ctx, to)

	// Account new fractional balances (NOT YET carried)
	senderNewFractionalBal := senderFractionalBal.Sub(fractionalAmt)
	recipientNewFractionalBal := recipientFractionalBal.Add(fractionalAmt)

	// Check if sender needs to borrow and recipient needs to carry
	senderNeedsBorrow := senderNewFractionalBal.IsNegative()
	recipientNeedsCarry := recipientNewFractionalBal.GTE(types.ConversionFactor())

	// Update fractional balances to account for carried and borrowed integer
	// amounts. This needs to be always done after the integer transfer, no
	// matter if we used direct transfer or reserve exchange.
	if senderNeedsBorrow {
		// Increase fractional balance by 1 integer equivalent amount.
		// No longer negative after adding. SetFractionalBalance will panic
		// if the amount is invalid.
		senderNewFractionalBal = senderNewFractionalBal.Add(types.ConversionFactor())
	}

	if recipientNeedsCarry {
		// Decrease fractional balance by 1 integer equivalent amount.
		// No longer > conversionFactor after subtracting
		recipientNewFractionalBal = recipientNewFractionalBal.Sub(types.ConversionFactor())
	}

	// Integer balance needs to be deducted from sender and increased for
	// recipient. Instead of using reserve exchange, we can directly transfer
	// between the account
	canDirectTransferCarry := senderNeedsBorrow && recipientNeedsCarry
	if canDirectTransferCarry {
		integerAmt = integerAmt.AddRaw(1)
	}

	// Direct integer transfer, including carry if possible.
	if integerAmt.IsPositive() {
		transferCoin := sdk.NewCoin(types.IntegerCoinDenom, integerAmt)
		if err := k.bk.SendCoins(ctx, from, to, sdk.NewCoins(transferCoin)); err != nil {
			return err
		}
	}

	// ------------------------------------------------
	// Use reserve to borrow and carry fractional coins if we there is no direct
	// transfer from sender to recipient.
	if !canDirectTransferCarry {
		// Send to reserve if sender needs to borrow
		if senderNeedsBorrow {
			borrowCoin := sdk.NewCoin(types.IntegerCoinDenom, sdk.NewInt(1))
			if err := k.bk.SendCoinsFromAccountToModule(ctx, from, types.ModuleName, sdk.NewCoins(borrowCoin)); err != nil {
				return err
			}
		}

		if recipientNeedsCarry {
			carryCoin := sdk.NewCoin(types.IntegerCoinDenom, sdk.NewInt(1))
			if err := k.bk.SendCoinsFromModuleToAccount(ctx, types.ModuleName, to, sdk.NewCoins(carryCoin)); err != nil {
				return err
			}
		}
	}

	// Update fractional balances
	k.SetFractionalBalance(ctx, from, senderNewFractionalBal)
	k.SetFractionalBalance(ctx, to, recipientNewFractionalBal)

	return nil
}

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

	return k.SendCoins(ctx, senderAddr, recipientAcc.GetAddress(), amt)
}

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

	// Uses x/bank BlockedAddr, no need to modify
	if k.bk.BlockedAddr(recipientAddr) {
		return errorsmod.Wrapf(sdkerrors.ErrUnauthorized, "%s is not allowed to receive funds", recipientAddr)
	}

	return k.SendCoins(ctx, senderAddr, recipientAddr, amt)
}
