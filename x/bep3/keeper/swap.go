package keeper

import (
	"bytes"
	"encoding/hex"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kava-labs/kava/x/bep3/types"
)

// CreateAtomicSwap creates a new AtomicSwap
func (k Keeper) CreateAtomicSwap(ctx sdk.Context, randomNumberHash []byte, timestamp int64, heightSpan int64,
	sender sdk.AccAddress, recipient sdk.AccAddress, senderOtherChain, recipientOtherChain string,
	amount sdk.Coins, expectedIncome string) sdk.Error {

	swapID := types.CalculateSwapID(randomNumberHash, sender, senderOtherChain)

	// Confirm that this swap is valid
	_, found := k.GetAtomicSwap(ctx, swapID)
	if found {
		return types.ErrAtomicSwapAlreadyExists(k.codespace, swapID)
	}

	// The heightSpan period should be more than 10 minutes and less than one week
	// Assume average block time interval is 10 second. 10 mins = 60 blocks, 1 week = 60480 blocks
	if heightSpan < k.GetMinBlockLock(ctx) || heightSpan > k.GetMaxBlockLock(ctx) {
		return types.ErrInvalidHeightSpan(k.codespace, heightSpan, k.GetMinBlockLock(ctx), k.GetMaxBlockLock(ctx))
	}

	// Validate that timestamp is within reasonable bounds
	if ctx.BlockHeight() > 1800 {
		if timestamp > ctx.BlockHeight()-1800 || timestamp < ctx.BlockHeight()+900 {
			return types.ErrInvalidTimestamp(k.codespace)
		}
	} else {
		if timestamp >= 1800 {
			return types.ErrInvalidTimestamp(k.codespace)
		}
	}

	// Sanity check on recipient address
	if recipient.Empty() {
		return sdk.ErrInvalidAddress("invalid (empty) bidder address")
	}

	err := k.validateCoinDeposit(ctx, amount)
	if err != nil {
		return err
	}

	// Transfer coins to module
	err = k.supplyKeeper.SendCoinsFromAccountToModule(ctx, sender, types.ModuleName, amount)
	if err != nil {
		return err
	}

	// Store the details of the swap.
	atomicSwap := types.NewAtomicSwap(amount, randomNumberHash,
		ctx.BlockHeight()+heightSpan, timestamp, sender, recipient,
		senderOtherChain, 0, types.Open)

	k.StoreNewAtomicSwap(ctx, atomicSwap, swapID)

	// Emit 'create_atomic_swap' event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeCreateAtomicSwap,
			sdk.NewAttribute(types.AttributeKeySender, fmt.Sprintf("%s", atomicSwap.Sender)),
			sdk.NewAttribute(types.AttributeKeyRecipient, fmt.Sprintf("%s", atomicSwap.Recipient)),
			sdk.NewAttribute(types.AttributeKeyAtomicSwapID, fmt.Sprintf("%s", hex.EncodeToString(swapID))),
			sdk.NewAttribute(types.AttributeKeyRandomNumberHash, fmt.Sprintf("%s", atomicSwap.RandomNumberHash)),
			sdk.NewAttribute(types.AttributeKeyTimestamp, fmt.Sprintf("%d", atomicSwap.Timestamp)),
			sdk.NewAttribute(types.AttributeKeySenderOtherChain, fmt.Sprintf("%s", atomicSwap.SenderOtherChain)),
			sdk.NewAttribute(types.AttributeKeyExpireHeight, fmt.Sprintf("%d", atomicSwap.ExpireHeight)),
			sdk.NewAttribute(types.AttributeKeyAmount, fmt.Sprintf("%s", atomicSwap.Amount[0].String())),
			sdk.NewAttribute(types.AttributeKeyExpectedIncome, fmt.Sprintf("%s", expectedIncome)),
		),
	)

	return nil
}

// ClaimAtomicSwap validates a claim attempt, and if successful, sends the escrowed amount and closes the AtomicSwap
func (k Keeper) ClaimAtomicSwap(ctx sdk.Context, from sdk.AccAddress, swapID []byte, randomNumber []byte) sdk.Error {

	atomicSwap, found := k.GetAtomicSwap(ctx, swapID)
	if !found {
		return types.ErrAtomicSwapNotFound(k.codespace, swapID)
	}
	if atomicSwap.Status != types.Open {
		return types.ErrSwapNotOpen(k.codespace)
	}
	// Only unexpired AtomicSwaps can be claimed
	if ctx.BlockHeight() > atomicSwap.ExpireHeight {
		return types.ErrAtomicSwapHasExpired(k.codespace)
	}

	//  Calculate hashed secret using submitted number
	hashedSubmittedNumber := types.CalculateRandomHash(randomNumber, atomicSwap.Timestamp)
	hashedSecret := types.CalculateSwapID(hashedSubmittedNumber, atomicSwap.Sender, atomicSwap.SenderOtherChain)

	// Confirm that secret unlocks the atomic swap
	if !bytes.Equal(hashedSecret, swapID) {
		return types.ErrInvalidClaimSecret(k.codespace, hashedSecret, swapID)
	}

	// Increment the asset's total supply (if valid)
	err := k.IncrementAssetSupply(ctx, atomicSwap.Amount[0])
	if err != nil {
		return err
	}

	// Send intended recipient coins
	err = k.supplyKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, atomicSwap.Recipient, atomicSwap.Amount)
	if err != nil {
		return err
	}

	// Complete the swap
	atomicSwap.Status = types.Completed
	atomicSwap.ClosedBlock = ctx.BlockHeight()
	k.SetAtomicSwap(ctx, atomicSwap, swapID)

	// Emit "claim_atomic_swap" event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeClaimAtomicSwap,
			sdk.NewAttribute(types.AttributeKeyClaimSender, fmt.Sprintf("%s", from)),
			sdk.NewAttribute(types.AttributeKeyRecipient, fmt.Sprintf("%s", atomicSwap.Recipient)),
			sdk.NewAttribute(types.AttributeKeyAtomicSwapID, fmt.Sprintf("%s", hex.EncodeToString(swapID))),
			sdk.NewAttribute(types.AttributeKeyRandomNumberHash, fmt.Sprintf("%s", atomicSwap.RandomNumberHash)),
			sdk.NewAttribute(types.AttributeKeyRandomNumber, fmt.Sprintf("%s", randomNumber)),
		),
	)

	return nil
}

// RefundAtomicSwap refunds an AtomicSwap, sending assets to the original sender and closing the AtomicSwap
func (k Keeper) RefundAtomicSwap(ctx sdk.Context, from sdk.AccAddress, swapID []byte) sdk.Error {

	atomicSwap, found := k.GetAtomicSwap(ctx, swapID)
	if !found {
		return types.ErrAtomicSwapNotFound(k.codespace, swapID)
	}
	if atomicSwap.Status != types.Open {
		return types.ErrSwapNotOpen(k.codespace)
	}
	// Only expired swaps may be refunded
	if ctx.BlockHeight() <= atomicSwap.ExpireHeight {
		return types.ErrSwapNotRefundable(k.codespace)
	}

	// Refund coins to original swap sender
	err := k.supplyKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, atomicSwap.Sender, atomicSwap.Amount)
	if err != nil {
		return err
	}

	// Expire the swap
	atomicSwap.Status = types.Expired
	atomicSwap.ClosedBlock = ctx.BlockHeight()
	k.SetAtomicSwap(ctx, atomicSwap, swapID)

	// Emit 'refund_atomic_swap' event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeRefundAtomicSwap,
			sdk.NewAttribute(types.AttributeKeyRefundSender, fmt.Sprintf("%s", from)),
			sdk.NewAttribute(types.AttributeKeySender, fmt.Sprintf("%s", atomicSwap.Sender)),
			sdk.NewAttribute(types.AttributeKeyAtomicSwapID, fmt.Sprintf("%s", hex.EncodeToString(swapID))),
			sdk.NewAttribute(types.AttributeKeyRandomNumberHash, fmt.Sprintf("%s", atomicSwap.RandomNumberHash)),
		),
	)

	return nil
}

// GetAllAtomicSwaps returns all AtomicSwaps from the store
func (k Keeper) GetAllAtomicSwaps(ctx sdk.Context) (atomicSwaps types.AtomicSwaps) {
	k.IterateAtomicSwaps(ctx, func(atomicSwap types.AtomicSwap) bool {
		atomicSwaps = append(atomicSwaps, atomicSwap)
		return false
	})
	return
}

// validateCoinDeposit validates that coins can be deposited into an atomic swap
func (k Keeper) validateCoinDeposit(ctx sdk.Context, coins sdk.Coins) sdk.Error {
	if len(coins) != 1 {
		return sdk.ErrInternal("amount must contain exactly one coin")
	}

	err := k.ValidateActiveAsset(ctx, coins[0])
	if err != nil {
		return err
	}
	if coins[0].IsZero() {
		return types.ErrAmountTooSmall(k.codespace, coins[0])
	}
	return nil
}
