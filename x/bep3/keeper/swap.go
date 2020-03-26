package keeper

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kava-labs/kava/x/bep3/types"
	tmtime "github.com/tendermint/tendermint/types/time"
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

	// Unix timestamp must be in range [-15 mins, 30 mins] of the current time
	pastTimestampLimit := tmtime.Now().Add(time.Duration(-15) * time.Minute).Unix()
	futureTimestampLimit := tmtime.Now().Add(time.Duration(30) * time.Minute).Unix()
	if timestamp < pastTimestampLimit || timestamp >= futureTimestampLimit {
		return types.ErrInvalidTimestamp(k.codespace)
	}

	// Sanity check on recipient address
	if recipient.Empty() {
		return sdk.ErrInvalidAddress("invalid (empty) recipient address")
	}

	if len(amount) != 1 {
		return sdk.ErrInternal("amount must contain exactly one coin")
	}

	// Validate that this asset is supported and active
	err := k.ValidateActiveAsset(ctx, amount[0])
	if err != nil {
		return err
	}

	// If this asset's supply isn't set in the store, set it to 0
	_, assetSupplyFoundInStore := k.GetAssetSupply(ctx, []byte(amount[0].Denom))
	if !assetSupplyFoundInStore {
		k.SetAssetSupply(ctx, sdk.NewInt64Coin(amount[0].Denom, 0), []byte(amount[0].Denom))
	}

	// Validate that the proposed increase will not put asset supply over limit
	err = k.ValidateProposedIncrease(ctx, amount[0])
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
		senderOtherChain, recipientOtherChain, 0, types.Open)

	k.SetAtomicSwap(ctx, atomicSwap)
	k.InsertIntoByBlockIndex(ctx, atomicSwap)

	// Emit 'create_atomic_swap' event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeCreateAtomicSwap,
			sdk.NewAttribute(types.AttributeKeySender, fmt.Sprintf("%s", atomicSwap.Sender)),
			sdk.NewAttribute(types.AttributeKeyRecipient, fmt.Sprintf("%s", atomicSwap.Recipient)),
			sdk.NewAttribute(types.AttributeKeyAtomicSwapID, fmt.Sprintf("%s", hex.EncodeToString(atomicSwap.GetSwapID()))),
			sdk.NewAttribute(types.AttributeKeyRandomNumberHash, fmt.Sprintf("%s", hex.EncodeToString(atomicSwap.RandomNumberHash))),
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
	// Only unexpired AtomicSwaps can be claimed
	if atomicSwap.Status == types.Expired {
		return types.ErrAtomicSwapHasExpired(k.codespace)
	}

	//  Calculate hashed secret using submitted number
	hashedSubmittedNumber := types.CalculateRandomHash(randomNumber, atomicSwap.Timestamp)
	hashedSecret := types.CalculateSwapID(hashedSubmittedNumber, atomicSwap.Sender, atomicSwap.SenderOtherChain)

	// Confirm that secret unlocks the atomic swap
	if !bytes.Equal(hashedSecret, atomicSwap.GetSwapID()) {
		return types.ErrInvalidClaimSecret(k.codespace, hashedSecret, atomicSwap.GetSwapID())
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

	// Emit 'claim_atomic_swap' event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeClaimAtomicSwap,
			sdk.NewAttribute(types.AttributeKeyClaimSender, fmt.Sprintf("%s", from)),
			sdk.NewAttribute(types.AttributeKeyRecipient, fmt.Sprintf("%s", atomicSwap.Recipient)),
			sdk.NewAttribute(types.AttributeKeyAtomicSwapID, fmt.Sprintf("%s", hex.EncodeToString(atomicSwap.GetSwapID()))),
			sdk.NewAttribute(types.AttributeKeyRandomNumberHash, fmt.Sprintf("%s", hex.EncodeToString(atomicSwap.RandomNumberHash))),
			sdk.NewAttribute(types.AttributeKeyRandomNumber, fmt.Sprintf("%s", hex.EncodeToString(randomNumber))),
		),
	)

	// Delete the swap
	k.RemoveAtomicSwap(ctx, atomicSwap.GetSwapID())
	k.RemoveFromByBlockIndex(ctx, atomicSwap)
	return nil
}

// RefundAtomicSwap refunds an AtomicSwap, sending assets to the original sender and closing the AtomicSwap
func (k Keeper) RefundAtomicSwap(ctx sdk.Context, from sdk.AccAddress, swapID []byte) sdk.Error {

	atomicSwap, found := k.GetAtomicSwap(ctx, swapID)
	if !found {
		return types.ErrAtomicSwapNotFound(k.codespace, swapID)
	}
	// Only expired swaps may be refunded
	if atomicSwap.Status != types.Expired {
		return types.ErrSwapNotRefundable(k.codespace)
	}

	// Refund coins to original swap sender
	err := k.supplyKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, atomicSwap.Sender, atomicSwap.Amount)
	if err != nil {
		return err
	}

	// Emit 'refund_atomic_swap' event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeRefundAtomicSwap,
			sdk.NewAttribute(types.AttributeKeyRefundSender, fmt.Sprintf("%s", from)),
			sdk.NewAttribute(types.AttributeKeySender, fmt.Sprintf("%s", atomicSwap.Sender)),
			sdk.NewAttribute(types.AttributeKeyAtomicSwapID, fmt.Sprintf("%s", hex.EncodeToString(atomicSwap.GetSwapID()))),
			sdk.NewAttribute(types.AttributeKeyRandomNumberHash, fmt.Sprintf("%s", hex.EncodeToString(atomicSwap.RandomNumberHash))),
		),
	)

	// Delete the swap
	k.RemoveAtomicSwap(ctx, atomicSwap.GetSwapID())
	return nil
}

// UpdateExpiredAtomicSwaps finds all AtomicSwaps that are past (or at) their ending times and expires them.
func (k Keeper) UpdateExpiredAtomicSwaps(ctx sdk.Context) sdk.Error {
	var expiredSwaps [][]byte
	k.IterateAtomicSwapsByBlock(ctx, uint64(ctx.BlockHeight()), func(id []byte) bool {
		expiredSwaps = append(expiredSwaps, id)
		return false
	})

	// AtomicSwap refunding is in separate loops as db should not be modified during iteration
	for _, id := range expiredSwaps {
		// Update the AtomicSwap's status to expired
		swap, _ := k.GetAtomicSwap(ctx, id)
		swap.Status = types.Expired
		swap.ClosedBlock = ctx.BlockHeight()
		k.SetAtomicSwap(ctx, swap)

		// Remove swap from block index to prevent unnecessary iteration
		k.RemoveFromByBlockIndex(ctx, swap)
	}
	return nil
}
