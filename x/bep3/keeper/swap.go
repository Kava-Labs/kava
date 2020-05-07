package keeper

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/kava-labs/kava/x/bep3/types"
)

// CreateAtomicSwap creates a new atomic swap.
func (k Keeper) CreateAtomicSwap(ctx sdk.Context, randomNumberHash []byte, timestamp int64, heightSpan uint64,
	sender sdk.AccAddress, recipient sdk.AccAddress, senderOtherChain, recipientOtherChain string,
	amount sdk.Coins, crossChain bool) error {
	// Confirm that this is not a duplicate swap
	swapID := types.CalculateSwapID(randomNumberHash, sender, senderOtherChain)
	_, found := k.GetAtomicSwap(ctx, swapID)
	if found {
		return sdkerrors.Wrap(types.ErrAtomicSwapAlreadyExists, hex.EncodeToString(swapID))
	}

	// The heightSpan period should be more than 10 minutes and less than one week
	// Assume average block time interval is 10 second. 10 mins = 60 blocks, 1 week = 60480 blocks
	if heightSpan < k.GetMinBlockLock(ctx) || heightSpan > k.GetMaxBlockLock(ctx) {
		return sdkerrors.Wrapf(types.ErrInvalidHeightSpan, "height span %d, range %d - %d", heightSpan, k.GetMinBlockLock(ctx), k.GetMaxBlockLock(ctx))
	}

	// Unix timestamp must be in range [-15 mins, 30 mins] of the current time
	pastTimestampLimit := ctx.BlockTime().Add(time.Duration(-15) * time.Minute).Unix()
	futureTimestampLimit := ctx.BlockTime().Add(time.Duration(30) * time.Minute).Unix()
	if timestamp < pastTimestampLimit || timestamp >= futureTimestampLimit {
		return sdkerrors.Wrap(types.ErrInvalidTimestamp, time.Unix(timestamp, 0).UTC().String())
	}

	if len(amount) != 1 {
		return fmt.Errorf("amount must contain exactly one coin")
	}

	err := k.ValidateLiveAsset(ctx, amount[0])
	if err != nil {
		return err
	}

	var direction types.SwapDirection
	deputy := k.GetBnbDeputyAddress(ctx)
	if sender.Equals(deputy) {
		direction = types.Incoming
	} else {
		direction = types.Outgoing
	}

	switch direction {
	case types.Incoming:
		// If recipient's account doesn't exist, register it in state so that the address can send
		// a claim swap tx without needing to be registered in state by receiving a coin transfer.
		recipientAcc := k.accountKeeper.GetAccount(ctx, recipient)
		if recipientAcc == nil {
			newAcc := k.accountKeeper.NewAccountWithAddress(ctx, recipient)
			k.accountKeeper.SetAccount(ctx, newAcc)
		}
		err = k.IncrementIncomingAssetSupply(ctx, amount[0])
	case types.Outgoing:
		err = k.IncrementOutgoingAssetSupply(ctx, amount[0])
	default:
		err = fmt.Errorf("invalid swap direction: %s", direction.String())
	}

	if err != nil {
		return err
	}

	// Transfer coins to module
	err = k.supplyKeeper.SendCoinsFromAccountToModule(ctx, sender, types.ModuleName, amount)
	if err != nil {
		return err
	}

	// Store the details of the swap
	expireHeight := uint64(ctx.BlockHeight()) + heightSpan
	atomicSwap := types.NewAtomicSwap(amount, randomNumberHash, expireHeight, timestamp, sender,
		recipient, senderOtherChain, recipientOtherChain, 0, types.Open, crossChain, direction)

	// Insert the atomic swap under both keys
	k.SetAtomicSwap(ctx, atomicSwap)
	k.InsertIntoByBlockIndex(ctx, atomicSwap)

	// Emit 'create_atomic_swap' event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeCreateAtomicSwap,
			sdk.NewAttribute(types.AttributeKeySender, atomicSwap.Sender.String()),
			sdk.NewAttribute(types.AttributeKeyRecipient, atomicSwap.Recipient.String()),
			sdk.NewAttribute(types.AttributeKeyAtomicSwapID, hex.EncodeToString(atomicSwap.GetSwapID())),
			sdk.NewAttribute(types.AttributeKeyRandomNumberHash, hex.EncodeToString(atomicSwap.RandomNumberHash)),
			sdk.NewAttribute(types.AttributeKeyTimestamp, fmt.Sprintf("%d", atomicSwap.Timestamp)),
			sdk.NewAttribute(types.AttributeKeySenderOtherChain, atomicSwap.SenderOtherChain),
			sdk.NewAttribute(types.AttributeKeyExpireHeight, fmt.Sprintf("%d", atomicSwap.ExpireHeight)),
			sdk.NewAttribute(types.AttributeKeyAmount, atomicSwap.Amount.String()),
			sdk.NewAttribute(types.AttributeKeyDirection, atomicSwap.Direction.String()),
		),
	)

	return nil
}

// ClaimAtomicSwap validates a claim attempt, and if successful, sends the escrowed amount and closes the AtomicSwap.
func (k Keeper) ClaimAtomicSwap(ctx sdk.Context, from sdk.AccAddress, swapID []byte, randomNumber []byte) error {
	atomicSwap, found := k.GetAtomicSwap(ctx, swapID)
	if !found {
		return sdkerrors.Wrapf(types.ErrAtomicSwapNotFound, "%s", swapID)
	}

	// Only open atomic swaps can be claimed
	if atomicSwap.Status != types.Open {
		return sdkerrors.Wrapf(types.ErrSwapNotClaimable, "status %s", atomicSwap.Status.String())
	}

	//  Calculate hashed secret using submitted number
	hashedSubmittedNumber := types.CalculateRandomHash(randomNumber, atomicSwap.Timestamp)
	hashedSecret := types.CalculateSwapID(hashedSubmittedNumber, atomicSwap.Sender, atomicSwap.SenderOtherChain)

	// Confirm that secret unlocks the atomic swap
	if !bytes.Equal(hashedSecret, atomicSwap.GetSwapID()) {
		return sdkerrors.Wrapf(types.ErrInvalidClaimSecret, "the submitted random number is incorrect")
	}

	var err error
	switch atomicSwap.Direction {
	case types.Incoming:
		err = k.DecrementIncomingAssetSupply(ctx, atomicSwap.Amount[0])
		if err != nil {
			return err
		}
		err = k.IncrementCurrentAssetSupply(ctx, atomicSwap.Amount[0])
		if err != nil {
			return err
		}
	case types.Outgoing:
		err = k.DecrementOutgoingAssetSupply(ctx, atomicSwap.Amount[0])
		if err != nil {
			return err
		}
		err = k.DecrementCurrentAssetSupply(ctx, atomicSwap.Amount[0])
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("invalid swap direction: %s", atomicSwap.Direction.String())
	}

	// Send intended recipient coins
	err = k.supplyKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, atomicSwap.Recipient, atomicSwap.Amount)
	if err != nil {
		return err
	}

	// Complete swap
	atomicSwap.Status = types.Completed
	atomicSwap.ClosedBlock = ctx.BlockHeight()
	k.SetAtomicSwap(ctx, atomicSwap)

	// Remove from byBlock index and transition to longterm storage
	k.RemoveFromByBlockIndex(ctx, atomicSwap)
	k.InsertIntoLongtermStorage(ctx, atomicSwap)

	// Emit 'claim_atomic_swap' event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeClaimAtomicSwap,
			sdk.NewAttribute(types.AttributeKeyClaimSender, from.String()),
			sdk.NewAttribute(types.AttributeKeyRecipient, atomicSwap.Recipient.String()),
			sdk.NewAttribute(types.AttributeKeyAtomicSwapID, hex.EncodeToString(atomicSwap.GetSwapID())),
			sdk.NewAttribute(types.AttributeKeyRandomNumberHash, hex.EncodeToString(atomicSwap.RandomNumberHash)),
			sdk.NewAttribute(types.AttributeKeyRandomNumber, string(randomNumber)),
		),
	)

	return nil
}

// RefundAtomicSwap refunds an AtomicSwap, sending assets to the original sender and closing the AtomicSwap.
func (k Keeper) RefundAtomicSwap(ctx sdk.Context, from sdk.AccAddress, swapID []byte) error {
	atomicSwap, found := k.GetAtomicSwap(ctx, swapID)
	if !found {
		return sdkerrors.Wrapf(types.ErrAtomicSwapNotFound, "%s", swapID)
	}
	// Only expired swaps may be refunded
	if atomicSwap.Status != types.Expired {
		return sdkerrors.Wrapf(types.ErrSwapNotRefundable, "status %s", atomicSwap.Status.String())
	}

	var err error
	switch atomicSwap.Direction {
	case types.Incoming:
		err = k.DecrementIncomingAssetSupply(ctx, atomicSwap.Amount[0])
	case types.Outgoing:
		err = k.DecrementOutgoingAssetSupply(ctx, atomicSwap.Amount[0])
	default:
		err = fmt.Errorf("invalid swap direction: %s", atomicSwap.Direction.String())
	}

	if err != nil {
		return err
	}

	// Refund coins to original swap sender
	err = k.supplyKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, atomicSwap.Sender, atomicSwap.Amount)
	if err != nil {
		return err
	}

	// Complete swap
	atomicSwap.Status = types.Completed
	atomicSwap.ClosedBlock = ctx.BlockHeight()
	k.SetAtomicSwap(ctx, atomicSwap)

	// Transition to longterm storage
	k.InsertIntoLongtermStorage(ctx, atomicSwap)

	// Emit 'refund_atomic_swap' event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeRefundAtomicSwap,
			sdk.NewAttribute(types.AttributeKeyRefundSender, from.String()),
			sdk.NewAttribute(types.AttributeKeySender, atomicSwap.Sender.String()),
			sdk.NewAttribute(types.AttributeKeyAtomicSwapID, hex.EncodeToString(atomicSwap.GetSwapID())),
			sdk.NewAttribute(types.AttributeKeyRandomNumberHash, hex.EncodeToString(atomicSwap.RandomNumberHash)),
		),
	)

	return nil
}

// UpdateExpiredAtomicSwaps finds all AtomicSwaps that are past (or at) their ending times and expires them.
func (k Keeper) UpdateExpiredAtomicSwaps(ctx sdk.Context) {
	var expiredSwapIDs []string
	k.IterateAtomicSwapsByBlock(ctx, uint64(ctx.BlockHeight()), func(id []byte) bool {
		atomicSwap, found := k.GetAtomicSwap(ctx, id)
		if !found {
			// NOTE: shouldn't happen. Continue to next item.
			return false
		}
		// Expire the uncompleted swap and update both indexes
		atomicSwap.Status = types.Expired
		// Note: claimed swaps have already been removed from byBlock index.
		k.RemoveFromByBlockIndex(ctx, atomicSwap)
		k.SetAtomicSwap(ctx, atomicSwap)
		expiredSwapIDs = append(expiredSwapIDs, hex.EncodeToString(atomicSwap.GetSwapID()))
		return false
	})

	// Emit 'swaps_expired' event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeSwapsExpired,
			sdk.NewAttribute(types.AttributeKeyAtomicSwapIDs, fmt.Sprintf("%s", expiredSwapIDs)),
			sdk.NewAttribute(types.AttributeExpirationBlock, fmt.Sprintf("%d", ctx.BlockHeight())),
		),
	)
}

// DeleteClosedAtomicSwapsFromLongtermStorage removes swaps one week after completion.
func (k Keeper) DeleteClosedAtomicSwapsFromLongtermStorage(ctx sdk.Context) {
	k.IterateAtomicSwapsLongtermStorage(ctx, uint64(ctx.BlockHeight()), func(id []byte) bool {
		swap, found := k.GetAtomicSwap(ctx, id)
		if !found {
			// NOTE: shouldn't happen. Continue to next item.
			return false
		}
		k.RemoveAtomicSwap(ctx, swap.GetSwapID())
		k.RemoveFromLongtermStorage(ctx, swap)
		return false
	})
}
