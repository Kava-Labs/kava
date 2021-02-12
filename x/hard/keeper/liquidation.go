package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/kava-labs/kava/x/hard/types"
)

// LiqData holds liquidation-related data
type LiqData struct {
	price            sdk.Dec
	ltv              sdk.Dec
	conversionFactor sdk.Int
}

// AttemptKeeperLiquidation enables a keeper to liquidate an individual borrower's position
func (k Keeper) AttemptKeeperLiquidation(ctx sdk.Context, keeper sdk.AccAddress, borrower sdk.AccAddress) error {
	deposit, found := k.GetDeposit(ctx, borrower)
	if !found {
		return types.ErrDepositNotFound
	}

	borrow, found := k.GetBorrow(ctx, borrower)
	if !found {
		return types.ErrBorrowNotFound
	}

	// Call incentive hooks
	k.BeforeDepositModified(ctx, deposit)
	k.BeforeBorrowModified(ctx, borrow)

	k.SyncBorrowInterest(ctx, borrower)
	k.SyncSupplyInterest(ctx, borrower)

	deposit, found = k.GetDeposit(ctx, borrower)
	if !found {
		return types.ErrDepositNotFound
	}

	borrow, found = k.GetBorrow(ctx, borrower)
	if !found {
		return types.ErrBorrowNotFound
	}

	isWithinRange, err := k.IsWithinValidLtvRange(ctx, deposit, borrow)
	if err != nil {
		return err
	}
	if isWithinRange {
		return sdkerrors.Wrapf(types.ErrBorrowNotLiquidatable, "position is within valid LTV range")
	}

	// Sending coins to auction module with keeper address getting % of the profits
	borrowDenoms := getDenoms(borrow.Amount)
	depositDenoms := getDenoms(deposit.Amount)
	err = k.SeizeDeposits(ctx, keeper, deposit, borrow, depositDenoms, borrowDenoms)
	if err != nil {
		return err
	}

	k.DeleteDeposit(ctx, deposit)
	k.DeleteBorrow(ctx, borrow)
	return nil
}

// SeizeDeposits seizes a list of deposits and sends them to auction
func (k Keeper) SeizeDeposits(ctx sdk.Context, keeper sdk.AccAddress, deposit types.Deposit,
	borrow types.Borrow, dDenoms, bDenoms []string) error {
	liqMap, err := k.LoadLiquidationData(ctx, deposit, borrow)
	if err != nil {
		return err
	}

	// Seize % of every deposit and send to the keeper
	keeperRewardCoins := sdk.Coins{}
	for _, depCoin := range deposit.Amount {
		mm, _ := k.GetMoneyMarket(ctx, depCoin.Denom)
		keeperReward := mm.KeeperRewardPercentage.MulInt(depCoin.Amount).TruncateInt()
		if keeperReward.GT(sdk.ZeroInt()) {
			// Send keeper their reward
			keeperCoin := sdk.NewCoin(depCoin.Denom, keeperReward)
			keeperRewardCoins = append(keeperRewardCoins, keeperCoin)
		}
	}
	if !keeperRewardCoins.Empty() {
		err := k.supplyKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleAccountName, keeper, keeperRewardCoins)
		if err != nil {
			return err
		}
	}

	// All deposit amounts not given to keeper as rewards are eligible to be auctioned off
	aucDeposits := deposit.Amount.Sub(keeperRewardCoins)

	// Build valuation map to hold deposit coin USD valuations
	depositCoinValues := types.NewValuationMap()
	for _, deposit := range aucDeposits {
		dData := liqMap[deposit.Denom]
		dCoinUsdValue := sdk.NewDecFromInt(deposit.Amount).Quo(sdk.NewDecFromInt(dData.conversionFactor)).Mul(dData.price)
		depositCoinValues.Increment(deposit.Denom, dCoinUsdValue)
	}

	// Build valuation map to hold borrow coin USD valuations
	borrowCoinValues := types.NewValuationMap()
	for _, bCoin := range borrow.Amount {
		bData := liqMap[bCoin.Denom]
		bCoinUsdValue := sdk.NewDecFromInt(bCoin.Amount).Quo(sdk.NewDecFromInt(bData.conversionFactor)).Mul(bData.price)
		borrowCoinValues.Increment(bCoin.Denom, bCoinUsdValue)
	}

	// Loan-to-Value ratio after sending keeper their reward
	ltv := borrowCoinValues.Sum().Quo(depositCoinValues.Sum())

	liquidatedCoins, err := k.StartAuctions(ctx, deposit.Depositor, borrow.Amount, aucDeposits, depositCoinValues, borrowCoinValues, ltv, liqMap)
	// If some coins were liquidated and sent to auction prior to error, still need to emit liquidation event
	if !liquidatedCoins.Empty() {
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeHardLiquidation,
				sdk.NewAttribute(types.AttributeKeyLiquidatedOwner, deposit.Depositor.String()),
				sdk.NewAttribute(types.AttributeKeyLiquidatedCoins, liquidatedCoins.String()),
				sdk.NewAttribute(types.AttributeKeyKeeper, keeper.String()),
				sdk.NewAttribute(types.AttributeKeyKeeperRewardCoins, keeperRewardCoins.String()),
			),
		)
	}
	// Returns nil if there's no error
	return err
}

// StartAuctions attempts to start auctions for seized assets
func (k Keeper) StartAuctions(ctx sdk.Context, borrower sdk.AccAddress, borrows, deposits sdk.Coins,
	depositCoinValues, borrowCoinValues types.ValuationMap, ltv sdk.Dec, liqMap map[string]LiqData) (sdk.Coins, error) {
	// Sort keys to ensure deterministic behavior
	bKeys := borrowCoinValues.GetSortedKeys()
	dKeys := depositCoinValues.GetSortedKeys()

	// Set up auction constants
	returnAddrs := []sdk.AccAddress{borrower}
	weights := []sdk.Int{sdk.NewInt(100)}
	debt := sdk.NewCoin("debt", sdk.ZeroInt())

	macc := k.supplyKeeper.GetModuleAccount(ctx, types.ModuleAccountName)
	maccCoins := macc.SpendableCoins(ctx.BlockTime())

	var liquidatedCoins sdk.Coins
	for _, bKey := range bKeys {
		bValue := borrowCoinValues.Get(bKey)
		maxLotSize := bValue.Quo(ltv)

		for _, dKey := range dKeys {
			dValue := depositCoinValues.Get(dKey)
			if maxLotSize.Equal(sdk.ZeroDec()) {
				break // exit out of the loop if we have cleared the full amount
			}

			if dValue.GTE(maxLotSize) { // We can start an auction for the whole borrow amount]
				bid := sdk.NewCoin(bKey, borrows.AmountOf(bKey))

				lotSize := maxLotSize.MulInt(liqMap[dKey].conversionFactor).Quo(liqMap[dKey].price)
				if lotSize.TruncateInt().Equal(sdk.ZeroInt()) {
					continue
				}
				lot := sdk.NewCoin(dKey, lotSize.TruncateInt())

				insufficientLotFunds := false
				if lot.Amount.GT(maccCoins.AmountOf(dKey)) {
					insufficientLotFunds = true
					lot = sdk.NewCoin(lot.Denom, maccCoins.AmountOf(dKey))
				}

				// Sanity check that we can deliver coins to the liquidator account
				if deposits.AmountOf(dKey).LT(lot.Amount) {
					return liquidatedCoins, types.ErrInsufficientCoins
				}

				// Start auction: bid = full borrow amount, lot = maxLotSize
				_, err := k.auctionKeeper.StartCollateralAuction(ctx, types.ModuleAccountName, lot, bid, returnAddrs, weights, debt)
				if err != nil {
					return liquidatedCoins, err
				}
				// Decrement supplied coins and increment borrowed coins optimistically
				k.DecrementSuppliedCoins(ctx, sdk.Coins{lot})
				k.DecrementBorrowedCoins(ctx, sdk.Coins{bid})

				// Add lot to liquidated coins
				liquidatedCoins = liquidatedCoins.Add(lot)

				// Update USD valuation maps
				borrowCoinValues.SetZero(bKey)
				depositCoinValues.Decrement(dKey, maxLotSize)
				// Update deposits, borrows
				borrows = borrows.Sub(sdk.NewCoins(bid))
				if insufficientLotFunds {
					deposits = deposits.Sub(sdk.NewCoins(sdk.NewCoin(dKey, deposits.AmountOf(dKey))))
				} else {
					deposits = deposits.Sub(sdk.NewCoins(lot))
				}
				// Update max lot size
				maxLotSize = sdk.ZeroDec()
			} else { // We can only start an auction for the partial borrow amount
				maxBid := dValue.Mul(ltv)
				bidSize := maxBid.MulInt(liqMap[bKey].conversionFactor).Quo(liqMap[bKey].price)
				bid := sdk.NewCoin(bKey, bidSize.TruncateInt())
				lot := sdk.NewCoin(dKey, deposits.AmountOf(dKey))

				if bid.Amount.Equal(sdk.ZeroInt()) || lot.Amount.Equal(sdk.ZeroInt()) {
					continue
				}

				insufficientLotFunds := false
				if lot.Amount.GT(maccCoins.AmountOf(dKey)) {
					insufficientLotFunds = true
					lot = sdk.NewCoin(lot.Denom, maccCoins.AmountOf(dKey))
				}

				// Sanity check that we can deliver coins to the liquidator account
				if deposits.AmountOf(dKey).LT(lot.Amount) {
					return liquidatedCoins, types.ErrInsufficientCoins
				}

				// Start auction: bid = maxBid, lot = whole deposit amount
				_, err := k.auctionKeeper.StartCollateralAuction(ctx, types.ModuleAccountName, lot, bid, returnAddrs, weights, debt)
				if err != nil {
					return liquidatedCoins, err
				}
				// Decrement supplied coins and increment borrowed coins optimistically
				k.DecrementSuppliedCoins(ctx, sdk.Coins{lot})
				k.DecrementBorrowedCoins(ctx, sdk.Coins{bid})

				// Add lot to liquidated coins
				liquidatedCoins = liquidatedCoins.Add(lot)

				// Update variables to account for partial auction
				borrowCoinValues.Decrement(bKey, maxBid)
				depositCoinValues.SetZero(dKey)

				borrows = borrows.Sub(sdk.NewCoins(bid))
				if insufficientLotFunds {
					deposits = deposits.Sub(sdk.NewCoins(sdk.NewCoin(dKey, deposits.AmountOf(dKey))))
				} else {
					deposits = deposits.Sub(sdk.NewCoins(lot))
				}

				// Update max lot size
				maxLotSize = borrowCoinValues.Get(bKey).Quo(ltv)
			}
		}
	}

	// Send any remaining deposit back to the original borrower
	for _, dKey := range dKeys {
		remaining := deposits.AmountOf(dKey)
		if remaining.GT(sdk.ZeroInt()) {
			returnCoin := sdk.NewCoins(sdk.NewCoin(dKey, remaining))
			err := k.supplyKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleAccountName, borrower, returnCoin)
			if err != nil {
				return liquidatedCoins, err
			}
		}
	}

	return liquidatedCoins, nil
}

// IsWithinValidLtvRange compares a borrow and deposit to see if it's within a valid LTV range at current prices
func (k Keeper) IsWithinValidLtvRange(ctx sdk.Context, deposit types.Deposit, borrow types.Borrow) (bool, error) {
	liqMap, err := k.LoadLiquidationData(ctx, deposit, borrow)
	if err != nil {
		return false, err
	}

	totalBorrowableUSDAmount := sdk.ZeroDec()
	totalDepositedUSDAmount := sdk.ZeroDec()
	for _, depCoin := range deposit.Amount {
		lData := liqMap[depCoin.Denom]
		usdValue := sdk.NewDecFromInt(depCoin.Amount).Quo(sdk.NewDecFromInt(lData.conversionFactor)).Mul(lData.price)
		totalDepositedUSDAmount = totalDepositedUSDAmount.Add(usdValue)
		borrowableUSDAmountForDeposit := usdValue.Mul(lData.ltv)
		totalBorrowableUSDAmount = totalBorrowableUSDAmount.Add(borrowableUSDAmountForDeposit)
	}

	totalBorrowedUSDAmount := sdk.ZeroDec()
	for _, coin := range borrow.Amount {
		lData := liqMap[coin.Denom]
		usdValue := sdk.NewDecFromInt(coin.Amount).Quo(sdk.NewDecFromInt(lData.conversionFactor)).Mul(lData.price)
		totalBorrowedUSDAmount = totalBorrowedUSDAmount.Add(usdValue)
	}

	// Check if the user's has borrowed more than they're allowed to
	if totalBorrowedUSDAmount.GT(totalBorrowableUSDAmount) {
		return false, nil
	}

	return true, nil
}

// GetStoreLTV calculates the user's current LTV based on their deposits/borrows in the store
// and does not include any outsanding interest.
func (k Keeper) GetStoreLTV(ctx sdk.Context, addr sdk.AccAddress) (sdk.Dec, error) {
	// Fetch deposits and parse coin denoms
	deposit, found := k.GetDeposit(ctx, addr)
	if !found {
		return sdk.ZeroDec(), nil
	}

	// Fetch borrow balances and parse coin denoms
	borrow, found := k.GetBorrow(ctx, addr)
	if !found {
		return sdk.ZeroDec(), nil
	}

	return k.CalculateLtv(ctx, deposit, borrow)
}

// CalculateLtv calculates the potential LTV given a user's deposits and borrows.
// The boolean returned indicates if the LTV should be added to the store's LTV index.
func (k Keeper) CalculateLtv(ctx sdk.Context, deposit types.Deposit, borrow types.Borrow) (sdk.Dec, error) {
	// Load required liquidation data for every deposit/borrow denom
	liqMap, err := k.LoadLiquidationData(ctx, deposit, borrow)
	if err != nil {
		return sdk.ZeroDec(), nil
	}

	// Build valuation map to hold deposit coin USD valuations
	depositCoinValues := types.NewValuationMap()
	for _, depCoin := range deposit.Amount {
		dData := liqMap[depCoin.Denom]
		dCoinUsdValue := sdk.NewDecFromInt(depCoin.Amount).Quo(sdk.NewDecFromInt(dData.conversionFactor)).Mul(dData.price)
		depositCoinValues.Increment(depCoin.Denom, dCoinUsdValue)
	}

	// Build valuation map to hold borrow coin USD valuations
	borrowCoinValues := types.NewValuationMap()
	for _, bCoin := range borrow.Amount {
		bData := liqMap[bCoin.Denom]
		bCoinUsdValue := sdk.NewDecFromInt(bCoin.Amount).Quo(sdk.NewDecFromInt(bData.conversionFactor)).Mul(bData.price)
		borrowCoinValues.Increment(bCoin.Denom, bCoinUsdValue)
	}

	// User doesn't have any deposits, catch divide by 0 error
	sumDeposits := depositCoinValues.Sum()
	if sumDeposits.Equal(sdk.ZeroDec()) {
		return sdk.ZeroDec(), nil
	}

	// Loan-to-Value ratio
	return borrowCoinValues.Sum().Quo(sumDeposits), nil
}

// LoadLiquidationData returns liquidation data, deposit, borrow
func (k Keeper) LoadLiquidationData(ctx sdk.Context, deposit types.Deposit, borrow types.Borrow) (map[string]LiqData, error) {
	liqMap := make(map[string]LiqData)

	borrowDenoms := getDenoms(borrow.Amount)
	depositDenoms := getDenoms(deposit.Amount)
	denoms := removeDuplicates(borrowDenoms, depositDenoms)

	// Load required liquidation data for every deposit/borrow denom
	for _, denom := range denoms {
		mm, found := k.GetMoneyMarket(ctx, denom)
		if !found {
			return liqMap, sdkerrors.Wrapf(types.ErrMarketNotFound, "no market found for denom %s", denom)

		}

		priceData, err := k.pricefeedKeeper.GetCurrentPrice(ctx, mm.SpotMarketID)
		if err != nil {
			return liqMap, err
		}

		liqMap[denom] = LiqData{priceData.Price, mm.BorrowLimit.LoanToValue, mm.ConversionFactor}
	}

	return liqMap, nil
}

func getDenoms(coins sdk.Coins) []string {
	denoms := []string{}
	for _, coin := range coins {
		denoms = append(denoms, coin.Denom)
	}
	return denoms
}

func removeDuplicates(one []string, two []string) []string {
	check := make(map[string]int)
	fullList := append(one, two...)

	res := []string{}
	for _, val := range fullList {
		check[val] = 1
	}

	for key := range check {
		res = append(res, key)
	}
	return res
}
