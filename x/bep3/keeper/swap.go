package keeper

import (
	"bytes"
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kava-labs/kava/x/bep3/types"
)

// CreateAtomicSwap adds an atomicSwap
func (k Keeper) CreateAtomicSwap(ctx sdk.Context, from sdk.AccAddress, to sdk.AccAddress, recipientOtherChain,
	senderOtherChain string, randomNumberHash []byte, timestamp int64, coins sdk.Coins,
	expectedIncome string, heightSpan int64, crossChain bool) sdk.Error {

	if heightSpan < k.GetMinLockTime(ctx) || heightSpan > k.GetMaxLockTime(ctx) {
		return types.ErrInvalidHeightSpan(k.codespace, heightSpan, k.GetMinLockTime(ctx), k.GetMaxLockTime(ctx))
	}

	err := k.ValidateCoinDeposit(ctx, coins)
	if err != nil {
		return err
	}

	expectedSwapID, err2 := types.CalculateSwapID(randomNumberHash, from, senderOtherChain)
	if err2 != nil {
		return sdk.ErrInternal(err2.Error())
	}

	existingAtomicSwap, found := k.GetAtomicSwap(ctx, expectedSwapID)
	if found {
		return types.ErrAtomicSwapAlreadyExists(k.codespace, existingAtomicSwap.SwapID)
	}

	if crossChain {
		// Only the deputy may submit cross-chain AtomicSwaps
		deputyAddress := k.GetBnbDeputyAddress(ctx)
		if !deputyAddress.Equals(from) {
			return types.ErrOnlyDeputy(k.codespace, from, deputyAddress)
		}
	} else {
		expectedIncomeCoins, err := sdk.ParseCoins(expectedIncome)
		if err != nil {
			return sdk.ErrInternal(err.Error())
		}
		if !coins.IsEqual(expectedIncomeCoins) {
			return sdk.ErrInternal("a same-chain AtomicSwap must have an amount equal to the expected income")
		}

		// Same-chain AtomicSwaps require user to send funds
		err = k.supplyKeeper.SendCoinsFromAccountToModule(ctx, from, types.ModuleName, coins)
		if err != nil {
			return sdk.ErrInternal(err.Error())
		}
	}

	expirationBlock := uint64(ctx.BlockHeight() + heightSpan)

	atomicSwap := types.NewAtomicSwap(expectedSwapID, from, to, recipientOtherChain,
		senderOtherChain, randomNumberHash, timestamp, coins, expectedIncome,
		crossChain, expirationBlock)

	k.StoreNewAtomicSwap(ctx, atomicSwap)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeCreateAtomicSwap,
			sdk.NewAttribute(types.AttributeKeyAtomicSwapID, fmt.Sprintf("%s", atomicSwap.SwapID)),
			sdk.NewAttribute(types.AttributeKeyRandomNumberHash, fmt.Sprintf("%s", atomicSwap.RandomNumberHash)),
			sdk.NewAttribute(types.AttributeKeyFrom, fmt.Sprintf("%s", atomicSwap.From)),
			sdk.NewAttribute(types.AttributeKeyTo, fmt.Sprintf("%s", atomicSwap.To)),
			sdk.NewAttribute(types.AttributeKeyCoinDenom, fmt.Sprintf("%s", atomicSwap.Amount[0].Denom)),
			sdk.NewAttribute(types.AttributeKeyCoinAmount, fmt.Sprintf("%d", atomicSwap.Amount[0].Amount.Int64())),
		),
	)

	return nil
}

// DepositAtomicSwap deposits funds in an existing AtomicSwap
func (k Keeper) DepositAtomicSwap(ctx sdk.Context, from sdk.AccAddress, swapID []byte, coins sdk.Coins) sdk.Error {

	err := k.ValidateCoinDeposit(ctx, coins)
	if err != nil {
		return err
	}

	atomicSwap, found := k.GetAtomicSwap(ctx, swapID)
	if !found {
		return types.ErrAtomicSwapNotFound(k.codespace, swapID)
	}

	// Only unexpired AtomicSwaps can receive deposits
	if uint64(ctx.BlockHeight()) > atomicSwap.ExpirationBlock {
		return types.ErrAtomicSwapHasExpired(k.codespace)
	}

	atomicSwapCoin := atomicSwap.Amount[0]
	coin := coins[0]

	// Validate new deposit
	if atomicSwap.CrossChain {
		return types.ErrOnlySameChain(k.codespace)
	}
	if !atomicSwap.From.Equals(from) {
		return types.ErrOnlyOriginalCreator(k.codespace, from, atomicSwap.From)
	}
	if atomicSwapCoin.Denom != coin.Denom {
		return types.ErrInvalidCoinDenom(k.codespace, atomicSwapCoin.Denom, coin.Denom)
	}

	// Send coins from depositor to the bep3 module
	err = k.supplyKeeper.SendCoinsFromAccountToModule(ctx, from, types.ModuleName, coins)
	if err != nil {
		return sdk.ErrInternal(err.Error())
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeDepositAtomicSwap,
			sdk.NewAttribute(types.AttributeKeyAtomicSwapID, fmt.Sprintf("%s", atomicSwap.SwapID)),
			sdk.NewAttribute(types.AttributeKeyCoinDenom, fmt.Sprintf("%s", coin.Denom)),
			sdk.NewAttribute(types.AttributeKeyCoinAmount, fmt.Sprintf("%d", coin.Amount.Int64())),
		),
	)

	// Update AtomicSwap state
	atomicSwap.Amount = atomicSwap.Amount.Add(coins)
	currExpectedIncome, _ := sdk.ParseCoins(atomicSwap.ExpectedIncome)
	atomicSwap.ExpectedIncome = currExpectedIncome.Add(coins).String()

	k.SetAtomicSwap(ctx, atomicSwap)

	return nil
}

// ClaimAtomicSwap validates a claim attempt, and if successful, sends the escrowed amount and closes the AtomicSwap
func (k Keeper) ClaimAtomicSwap(ctx sdk.Context, from sdk.AccAddress, swapID []byte, randomNumber []byte) sdk.Error {

	atomicSwap, found := k.GetAtomicSwap(ctx, swapID)
	if !found {
		return types.ErrAtomicSwapNotFound(k.codespace, swapID)
	}

	// Only unexpired AtomicSwaps can be claimed
	if uint64(ctx.BlockHeight()) > atomicSwap.ExpirationBlock {
		return types.ErrAtomicSwapHasExpired(k.codespace)
	}

	//  Calculate hashed secret using submitted number
	hashedSubmittedNumber := types.CalculateRandomHash(randomNumber, atomicSwap.Timestamp)
	hashedSecret, err := types.CalculateSwapID(hashedSubmittedNumber, atomicSwap.From, atomicSwap.SenderOtherChain)
	if err != nil {
		return sdk.ErrInternal(err.Error())
	}

	if !bytes.Equal(hashedSecret, swapID) {
		return types.ErrInvalidClaimSecret(k.codespace, hashedSecret, swapID)
	}

	// If AtomicSwap is not cross-chain, atomicSwap.ExpectedIncome equals atomicSwap.Amount
	claimerCoins, err := sdk.ParseCoins(atomicSwap.ExpectedIncome)
	if err != nil {
		return sdk.ErrInternal(err.Error())
	}

	if atomicSwap.CrossChain {
		err := k.ValidateCoinMint(ctx, atomicSwap.Amount)
		if err != nil {
			return err
		}

		// Mint full amount of this coin's associated debt coin to bep3 module for internal limit tracking
		internalTrackingCoins, err2 := getEqualInternalTrackingCoins(atomicSwap.Amount)
		if err2 != nil {
			return sdk.ErrInternal(err2.Error())
		}

		err = k.supplyKeeper.MintCoins(ctx, types.ModuleName, internalTrackingCoins)
		if err != nil {
			return err
		}

		// Mint coins for distribution
		err = k.supplyKeeper.MintCoins(ctx, types.ModuleName, atomicSwap.Amount)
		if err != nil {
			return err
		}
		// Send claimer their portion of coins
		err = k.supplyKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, from, claimerCoins)
		if err != nil {
			return err
		}
		// Send deputy remaining coins
		deputyCoins := atomicSwap.Amount.Sub(claimerCoins)
		err = k.supplyKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, k.GetBnbDeputyAddress(ctx), deputyCoins)
		if err != nil {
			return err
		}
	} else {
		// Send amount from bep3 module to adresss that successfully claimed AtomicSwap
		err = k.supplyKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, from, claimerCoins)
		if err != nil {
			return sdk.ErrInternal(err.Error())
		}
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeClaimAtomicSwap,
			sdk.NewAttribute(types.AttributeKeyAtomicSwapID, fmt.Sprintf("%s", atomicSwap.SwapID)),
			sdk.NewAttribute(types.AttributeKeyClaimer, fmt.Sprintf("%s", from)),
			sdk.NewAttribute(types.AttributeKeyCoinDenom, fmt.Sprintf("%s", claimerCoins[0].Denom)),
			sdk.NewAttribute(types.AttributeKeyCoinAmount, fmt.Sprintf("%d", claimerCoins[0].Amount.Int64())),
		),
	)

	// Update AtomicSwap state
	k.DeleteAtomicSwap(ctx, swapID)

	return nil
}

// RefundAtomicSwap refunds an AtomicSwap, sending assets to the original sender and closing the AtomicSwap
func (k Keeper) RefundAtomicSwap(ctx sdk.Context, from sdk.AccAddress, swapID []byte) sdk.Error {

	atomicSwap, found := k.GetAtomicSwap(ctx, swapID)
	if !found {
		return types.ErrAtomicSwapNotFound(k.codespace, swapID)
	}

	// Refund request must come from original creator or bep3 module
	if !from.Equals(atomicSwap.From) && !k.GetBnbDeputyAddress(ctx).Equals(atomicSwap.From) {
		return types.ErrOnlyOriginalCreator(k.codespace, from, atomicSwap.From)
	}

	if !atomicSwap.CrossChain {
		// Refund coins from bep3 module to original creator
		err := k.supplyKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, atomicSwap.From, atomicSwap.Amount)
		if err != nil {
			return sdk.ErrInternal(err.Error())
		}
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeRefundAtomicSwap,
			sdk.NewAttribute(types.AttributeKeyAtomicSwapID, fmt.Sprintf("%s", atomicSwap.SwapID)),
			sdk.NewAttribute(types.AttributeKeyFrom, fmt.Sprintf("%s", from)),
			sdk.NewAttribute(types.AttributeKeyCoinDenom, fmt.Sprintf("%s", atomicSwap.Amount[0].Denom)),
			sdk.NewAttribute(types.AttributeKeyCoinAmount, fmt.Sprintf("%d", atomicSwap.Amount[0].Amount.Int64())),
		),
	)

	// Update AtomicSwap state
	k.DeleteAtomicSwap(ctx, swapID)
	return nil

}

// ValidateCoinDeposit validates that an asset can be accepted
func (k Keeper) ValidateCoinDeposit(ctx sdk.Context, coins sdk.Coins) sdk.Error {
	if len(coins) != 1 {
		return sdk.ErrInternal("amount must contain exactly one coin")
	}
	coin := coins[0]
	if coin.Amount.IsZero() {
		return types.ErrAmountTooSmall(k.codespace, coin)
	}
	asset, found := k.GetAssetByDenom(ctx, coin.Denom)
	if !found {
		return types.ErrAssetNotSupported(k.codespace, coin.Denom)
	}
	if !asset.Active {
		return types.ErrAssetNotActive(k.codespace, asset.Denom)
	}
	return nil
}

// ValidateCoinMint validates that an asset minted
func (k Keeper) ValidateCoinMint(ctx sdk.Context, coins sdk.Coins) sdk.Error {
	coin := coins[0]
	asset, found := k.GetAssetByDenom(ctx, coin.Denom)
	if !found {
		return types.ErrAssetNotSupported(k.codespace, coin.Denom)
	}
	if !asset.Active {
		return types.ErrAssetNotActive(k.codespace, asset.Denom)
	}
	// Confirm that mint does not surpass asset limit
	internalTrackingCoin, err := getEqualInternalTrackingCoins(coins)
	if err != nil {
		return sdk.ErrInternal(err.Error())
	}
	skAcc := k.supplyKeeper.GetModuleAccount(ctx, types.ModuleName)
	for _, skCoin := range skAcc.GetCoins() {
		if skCoin.Denom == internalTrackingCoin[0].Denom {
			if skCoin.Amount.Add(internalTrackingCoin[0].Amount).Int64() > asset.Limit {
				return types.ErrAmountTooLarge(k.codespace, coin)
			}
		}
	}
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

// RefundExpiredAtomicSwaps finds all AtomicSwaps that are past (or at) their ending times and closes them.
func (k Keeper) RefundExpiredAtomicSwaps(ctx sdk.Context) sdk.Error {
	var expiredAtomicSwaps [][]byte
	k.IterateAtomicSwapsByBlock(ctx, uint64(ctx.BlockHeight()), func(id []byte) bool {
		expiredAtomicSwaps = append(expiredAtomicSwaps, id)
		return false
	})

	sdkAddr := k.supplyKeeper.GetModuleAddress(types.ModuleName)

	// AtomicSwap refunding is in separate loops as db should not be modified during iteration
	for _, id := range expiredAtomicSwaps {
		if err := k.RefundAtomicSwap(ctx, sdkAddr, id); err != nil {
			return err
		}
	}
	return nil
}

// getEqualInternalTrackingCoins returns an equal amount of internal tracking coins
func getEqualInternalTrackingCoins(coins sdk.Coins) (sdk.Coins, error) {
	coin := coins[0]
	internalCoinStr := []string{coin.Amount.String(), coin.Denom, "debt"}
	return sdk.ParseCoins(strings.Join(internalCoinStr, ""))
}
