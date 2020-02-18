package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kava-labs/kava/x/bep3/types"
)

// CreateHTLT adds an htlt
func (k Keeper) CreateHTLT(ctx sdk.Context, from sdk.AccAddress, to sdk.AccAddress, recipientOtherChain,
	senderOtherChain string, randomNumberHash string, timestamp int64, coins sdk.Coins,
	expectedIncome string, heightSpan int64, crossChain bool) (string, sdk.Error) {

	if heightSpan < k.GetMinLockTime(ctx) || heightSpan > k.GetMaxLockTime(ctx) {
		return "", types.ErrInvalidHeightSpan(k.codespace, heightSpan, k.GetMinLockTime(ctx), k.GetMaxLockTime(ctx))
	}

	err := k.ValidateAsset(ctx, coins)
	if err != nil {
		return "", err
	}

	expectedSwapID, err2 := types.CalculateSwapID(randomNumberHash, from, senderOtherChain)
	if err2 != nil {
		return "", sdk.ErrInternal(err2.Error())
	}

	_, found := k.GetHTLT(ctx, expectedSwapID)
	if found {
		encodedExpectedSwapID := types.BytesToHexEncodedString(expectedSwapID)
		return "", types.ErrHTLTAlreadyExists(k.codespace, encodedExpectedSwapID)
	}

	htlt := types.NewHTLT(from, to, recipientOtherChain, senderOtherChain,
		randomNumberHash, timestamp, coins, expectedIncome, heightSpan,
		crossChain)

	// Send coins from sender to the bep3 module
	err = k.supplyKeeper.SendCoinsFromAccountToModule(ctx, from, types.ModuleName, coins)
	if err != nil {
		return "", sdk.ErrInternal(err.Error())
	}

	swapID, sdkErr := k.StoreNewHTLT(ctx, htlt)
	if sdkErr != nil {
		return "", sdk.ErrInternal(sdkErr.Error())
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeCreateHtlt,
			sdk.NewAttribute(types.AttributeKeyHtltSwapID, fmt.Sprintf("%s", swapID)),
			sdk.NewAttribute(types.AttributeKeyRandomNumberHash, fmt.Sprintf("%s", randomNumberHash)),
			sdk.NewAttribute(types.AttributeKeyFrom, fmt.Sprintf("%s", htlt.From)),
			sdk.NewAttribute(types.AttributeKeyTo, fmt.Sprintf("%s", htlt.To)),
			sdk.NewAttribute(types.AttributeKeyCoinDenom, fmt.Sprintf("%s", htlt.Amount[0].Denom)),
			sdk.NewAttribute(types.AttributeKeyCoinAmount, fmt.Sprintf("%d", htlt.Amount[0].Amount.Int64())),
		),
	)

	return swapID, nil
}

// DepositHTLT deposits funds in an existing HTLT
func (k Keeper) DepositHTLT(ctx sdk.Context, from sdk.AccAddress, swapID string, coins sdk.Coins) sdk.Error {

	err := k.ValidateAsset(ctx, coins)
	if err != nil {
		return err
	}

	decodedSwapID, err2 := types.HexEncodedStringToBytes(swapID)
	if err2 != nil {
		return sdk.ErrInternal(err2.Error())
	}

	htlt, found := k.GetHTLT(ctx, decodedSwapID)
	if !found {
		return types.ErrHTLTNotFound(k.codespace, swapID)
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

	k.SetHTLT(ctx, htlt, decodedSwapID)

	return nil
}

// ClaimHTLT validates a claim attempt, and if successful, sends the escrowed amount and closes the HTLT
func (k Keeper) ClaimHTLT(ctx sdk.Context, from sdk.AccAddress, encodedSwapID string, randomNumber []byte) sdk.Error {
	decodedSwapID, err := types.HexEncodedStringToBytes(encodedSwapID)
	if err != nil {
		return sdk.ErrInternal(err.Error())
	}

	htlt, found := k.GetHTLT(ctx, decodedSwapID)
	if !found {
		return types.ErrHTLTNotFound(k.codespace, encodedSwapID)
	}

	// Calculate hashed random number with param number and timestamp
	hashedRandomNumber := types.CalculateRandomHash(randomNumber, htlt.Timestamp)
	stringRandomNumber := types.BytesToHexEncodedString(hashedRandomNumber)

	// Calculate hashed secret hashed random number, htlt sender, and sender other chain
	hashedSecret, err := types.CalculateSwapID(stringRandomNumber, htlt.From, htlt.SenderOtherChain)
	if err != nil {
		return sdk.ErrInternal(err.Error())
	}

	encodedHashedSecret := types.BytesToHexEncodedString(hashedSecret)
	if encodedHashedSecret != encodedSwapID {
		return types.ErrInvalidClaimSecret(k.codespace, encodedHashedSecret, encodedSwapID)
	}

	// If HTLT is not cross-chain, htlt.ExpectedIncome should equal htlt.Amount
	claimerCoins, err := sdk.ParseCoins(htlt.ExpectedIncome)
	if err != nil {
		return sdk.ErrInternal(err.Error())
	}
	deputyCoins := htlt.Amount.Sub(claimerCoins)

	// Send expected income from bep3 module to claiming address
	err = k.supplyKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, from, claimerCoins)
	if err != nil {
		return sdk.ErrInternal(err.Error())
	}

	// Send remaining amount from bep3 module to deputy
	if htlt.CrossChain {
		err = k.supplyKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, htlt.From, deputyCoins)
		if err != nil {
			return sdk.ErrInternal(err.Error())
		}
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeClaimHtlt,
			sdk.NewAttribute(types.AttributeKeyHtltSwapID, fmt.Sprintf("%s", encodedSwapID)),
			sdk.NewAttribute(types.AttributeKeyClaimer, fmt.Sprintf("%s", from)),
			sdk.NewAttribute(types.AttributeKeyCoinDenom, fmt.Sprintf("%s", claimerCoins[0].Denom)),
			sdk.NewAttribute(types.AttributeKeyCoinAmount, fmt.Sprintf("%d", claimerCoins[0].Amount.Int64())),
		),
	)

	// Update HTLT state
	k.DeleteHTLT(ctx, decodedSwapID)

	return nil
}

// RefundHTLT refunds an HTLT, sending assets to the original sender and closing the HTLT
func (k Keeper) RefundHTLT(ctx sdk.Context, from sdk.AccAddress, encodedSwapID string) sdk.Error {
	decodedSwapID, err := types.HexEncodedStringToBytes(encodedSwapID)
	if err != nil {
		return sdk.ErrInternal(err.Error())
	}

	htlt, found := k.GetHTLT(ctx, decodedSwapID)
	if !found {
		return types.ErrHTLTNotFound(k.codespace, encodedSwapID)
	}

	if !htlt.From.Equals(from) {
		return types.ErrOnlyOriginalCreator(k.codespace, from, htlt.From)
	}

	// Send coins from bep3 module to original creator
	err = k.supplyKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, from, htlt.Amount)
	if err != nil {
		return sdk.ErrInternal(err.Error())
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeRefundHtlt,
			sdk.NewAttribute(types.AttributeKeyHtltSwapID, fmt.Sprintf("%s", encodedSwapID)),
			sdk.NewAttribute(types.AttributeKeyFrom, fmt.Sprintf("%s", from)),
			sdk.NewAttribute(types.AttributeKeyCoinDenom, fmt.Sprintf("%s", htlt.Amount[0].Denom)),
			sdk.NewAttribute(types.AttributeKeyCoinAmount, fmt.Sprintf("%d", htlt.Amount[0].Amount.Int64())),
		),
	)

	// Update HTLT state
	k.DeleteHTLT(ctx, decodedSwapID)
	return nil

}

// ValidateAsset validates that an asset can be accepted
func (k Keeper) ValidateAsset(ctx sdk.Context, coins sdk.Coins) sdk.Error {
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
	skAcc := k.supplyKeeper.GetModuleAccount(ctx, types.ModuleName)
	for _, skCoin := range skAcc.GetCoins() {
		if skCoin.Denom == asset.Denom {
			if skCoin.Amount.Add(coin.Amount).Int64() > asset.Limit {
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
	fmt.Println("sdkAddr: ", sdkAddr)

	// TODO: is this correct? gov modifies during iteration
	// HTLT refunding is in separate loops as db should not be modified during iteration
	for _, id := range expiredHTLTs {
		if err := k.RefundHTLT(ctx, sdkAddr, types.BytesToHexEncodedString(id)); err != nil {
			return err
		}
	}
	return nil
}
