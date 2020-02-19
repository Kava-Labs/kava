package keeper

import (
	"bytes"
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kava-labs/kava/x/bep3/types"
)

// CreateHTLT adds an htlt
func (k Keeper) CreateHTLT(ctx sdk.Context, from sdk.AccAddress, to sdk.AccAddress, recipientOtherChain,
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

	_, found := k.GetHTLT(ctx, expectedSwapID)
	if found {
		return types.ErrHTLTAlreadyExists(k.codespace, types.BytesToHexEncodedString(expectedSwapID))
	}

	if crossChain {
		// Only the deputy may submit cross-chain HTLTs
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
			return sdk.ErrInternal("a same-chain HTLT must have an amount equal to the expected income")
		}

		// Same-chain HTLTs require user to send funds
		err = k.supplyKeeper.SendCoinsFromAccountToModule(ctx, from, types.ModuleName, coins)
		if err != nil {
			return sdk.ErrInternal(err.Error())
		}
	}

	expirationBlock := uint64(ctx.BlockHeight() + heightSpan)

	htlt := types.NewHTLT(expectedSwapID, from, to, recipientOtherChain,
		senderOtherChain, randomNumberHash, timestamp, coins, expectedIncome,
		heightSpan, crossChain, expirationBlock)

	k.StoreNewHTLT(ctx, htlt)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeCreateHtlt,
			sdk.NewAttribute(types.AttributeKeyHtltSwapID, fmt.Sprintf("%s", htlt.SwapID)),
			sdk.NewAttribute(types.AttributeKeyRandomNumberHash, fmt.Sprintf("%s", randomNumberHash)),
			sdk.NewAttribute(types.AttributeKeyFrom, fmt.Sprintf("%s", htlt.From)),
			sdk.NewAttribute(types.AttributeKeyTo, fmt.Sprintf("%s", htlt.To)),
			sdk.NewAttribute(types.AttributeKeyCoinDenom, fmt.Sprintf("%s", htlt.Amount[0].Denom)),
			sdk.NewAttribute(types.AttributeKeyCoinAmount, fmt.Sprintf("%d", htlt.Amount[0].Amount.Int64())),
		),
	)

	return nil
}

// DepositHTLT deposits funds in an existing HTLT
func (k Keeper) DepositHTLT(ctx sdk.Context, from sdk.AccAddress, swapID []byte, coins sdk.Coins) sdk.Error {

	err := k.ValidateCoinDeposit(ctx, coins)
	if err != nil {
		return err
	}

	htlt, found := k.GetHTLT(ctx, swapID)
	if !found {
		return types.ErrHTLTNotFound(k.codespace, swapID)
	}

	// Only unexpired HTLTs can receive deposits
	if uint64(ctx.BlockTime().Unix()) > htlt.ExpirationBlock {
		return types.ErrHTLTHasExpired(k.codespace)
	}

	htltCoin := htlt.Amount[0]
	coin := coins[0]

	// Validate new deposit
	if htlt.CrossChain {
		return types.ErrOnlySameChain(k.codespace)
	}
	if htlt.From.Equals(from) {
		return types.ErrOnlyOriginalCreator(k.codespace, from, htlt.From)
	}
	if htltCoin.Denom != coin.Denom {
		return types.ErrInvalidCoinDenom(k.codespace, htltCoin.Denom, coin.Denom)
	}

	// Send coins from depositor to the bep3 module
	err = k.supplyKeeper.SendCoinsFromAccountToModule(ctx, from, types.ModuleName, coins)
	if err != nil {
		return sdk.ErrInternal(err.Error())
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeDepositHtlt,
			sdk.NewAttribute(types.AttributeKeyHtltSwapID, fmt.Sprintf("%s", swapID)),
			sdk.NewAttribute(types.AttributeKeyCoinDenom, fmt.Sprintf("%s", coin.Denom)),
			sdk.NewAttribute(types.AttributeKeyCoinAmount, fmt.Sprintf("%d", coin.Amount.Int64())),
		),
	)

	// Update HTLT state
	htlt.Amount = htlt.Amount.Add(coins)
	currExpectedIncome, _ := sdk.ParseCoins(htlt.ExpectedIncome)
	htlt.ExpectedIncome = currExpectedIncome.Add(coins).String()

	k.SetHTLT(ctx, htlt)

	return nil
}

// ClaimHTLT validates a claim attempt, and if successful, sends the escrowed amount and closes the HTLT
func (k Keeper) ClaimHTLT(ctx sdk.Context, from sdk.AccAddress, swapID []byte, randomNumber []byte) sdk.Error {

	htlt, found := k.GetHTLT(ctx, swapID)
	if !found {
		return types.ErrHTLTNotFound(k.codespace, swapID)
	}

	// Only unexpired HTLTs can be claimed
	if uint64(ctx.BlockTime().Unix()) > htlt.ExpirationBlock {
		return types.ErrHTLTHasExpired(k.codespace)
	}

	//  Calculate hashed secret using submitted number
	hashedSubmittedNumber := types.CalculateRandomHash(randomNumber, htlt.Timestamp)
	hashedSecret, err := types.CalculateSwapID(hashedSubmittedNumber, htlt.From, htlt.SenderOtherChain)
	if err != nil {
		return sdk.ErrInternal(err.Error())
	}

	if !bytes.Equal(hashedSecret, swapID) {
		return types.ErrInvalidClaimSecret(k.codespace, hashedSecret, swapID)
	}

	// If HTLT is not cross-chain, htlt.ExpectedIncome equals htlt.Amount
	claimerCoins, err := sdk.ParseCoins(htlt.ExpectedIncome)
	if err != nil {
		return sdk.ErrInternal(err.Error())
	}

	if htlt.CrossChain {
		err := k.ValidateCoinMint(ctx, htlt.Amount)
		if err != nil {
			return err
		}

		internalTrackingCoins, err2 := getEqualInternalTrackingCoins(htlt.Amount)
		if err2 != nil {
			return sdk.ErrInternal(err2.Error())
		}

		// Mint full amount of this coin's associated debt coin to bep3 module for internal limit tracking
		err = k.supplyKeeper.SendCoinsFromModuleToModule(ctx, types.ModuleName, types.ModuleName, internalTrackingCoins)
		if err != nil {
			return err
		}
		// Mint coins to claimer
		err = k.supplyKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, from, claimerCoins)
		if err != nil {
			return err
		}
		// Mint remaining coins to deputy
		deputyCoins := htlt.Amount.Sub(claimerCoins)
		err = k.supplyKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, k.GetBnbDeputyAddress(ctx), deputyCoins)
		if err != nil {
			return sdk.ErrInternal(err.Error())
		}
	} else {
		// Send amount from bep3 module to adresss that successfully claimed HTLT
		err = k.supplyKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, from, claimerCoins)
		if err != nil {
			return sdk.ErrInternal(err.Error())
		}
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeClaimHtlt,
			sdk.NewAttribute(types.AttributeKeyHtltSwapID, fmt.Sprintf("%s", types.BytesToHexEncodedString(swapID))),
			sdk.NewAttribute(types.AttributeKeyClaimer, fmt.Sprintf("%s", from)),
			sdk.NewAttribute(types.AttributeKeyCoinDenom, fmt.Sprintf("%s", claimerCoins[0].Denom)),
			sdk.NewAttribute(types.AttributeKeyCoinAmount, fmt.Sprintf("%d", claimerCoins[0].Amount.Int64())),
		),
	)

	// Update HTLT state
	k.DeleteHTLT(ctx, swapID)

	return nil
}

// RefundHTLT refunds an HTLT, sending assets to the original sender and closing the HTLT
func (k Keeper) RefundHTLT(ctx sdk.Context, from sdk.AccAddress, swapID []byte) sdk.Error {

	htlt, found := k.GetHTLT(ctx, swapID)
	if !found {
		return types.ErrHTLTNotFound(k.codespace, swapID)
	}

	// Refund request must come from original creator or bep3 module
	if !from.Equals(htlt.From) && !k.GetBnbDeputyAddress(ctx).Equals(htlt.From) {
		return types.ErrOnlyOriginalCreator(k.codespace, from, htlt.From)
	}

	if !htlt.CrossChain {
		// Refund coins from bep3 module to original creator
		err := k.supplyKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, htlt.From, htlt.Amount)
		if err != nil {
			return sdk.ErrInternal(err.Error())
		}
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeRefundHtlt,
			sdk.NewAttribute(types.AttributeKeyHtltSwapID, fmt.Sprintf("%s", types.BytesToHexEncodedString(swapID))),
			sdk.NewAttribute(types.AttributeKeyFrom, fmt.Sprintf("%s", from)),
			sdk.NewAttribute(types.AttributeKeyCoinDenom, fmt.Sprintf("%s", htlt.Amount[0].Denom)),
			sdk.NewAttribute(types.AttributeKeyCoinAmount, fmt.Sprintf("%d", htlt.Amount[0].Amount.Int64())),
		),
	)

	// Update HTLT state
	k.DeleteHTLT(ctx, swapID)
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

// GetAllHtlts returns all HTLTs from the store
func (k Keeper) GetAllHtlts(ctx sdk.Context) (htlts types.HTLTs) {
	k.IterateHTLTs(ctx, func(htlt types.HTLT) bool {
		htlts = append(htlts, htlt)
		return false
	})
	return
}

// RefundExpiredHTLTs finds all HTLTs that are past (or at) their ending times and closes them.
func (k Keeper) RefundExpiredHTLTs(ctx sdk.Context) sdk.Error {
	var expiredHTLTs [][]byte
	k.IterateHTLTsByTime(ctx, uint64(ctx.BlockTime().Unix()), func(id []byte) bool {
		expiredHTLTs = append(expiredHTLTs, id)
		return false
	})

	sdkAddr := k.supplyKeeper.GetModuleAddress(types.ModuleName)

	// HTLT refunding is in separate loops as db should not be modified during iteration
	for _, id := range expiredHTLTs {
		if err := k.RefundHTLT(ctx, sdkAddr, id); err != nil {
			return err
		}
	}
	return nil
}

// getEqualInternalTrackingCoins returns an equal amount of internal tracking coins
func getEqualInternalTrackingCoins(coins sdk.Coins) (sdk.Coins, error) {
	coin := coins[0]
	internalCoinStr := []string{coin.Amount.String(), coin.Denom, "_INTERNAL_TRACKING_COIN"}
	return sdk.ParseCoins(strings.Join(internalCoinStr, ""))
}
