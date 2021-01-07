package keeper

import (
	"errors"

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

// AttemptIndexLiquidations attempts to liquidate the lowest LTV borrows
func (k Keeper) AttemptIndexLiquidations(ctx sdk.Context) error {
	params := k.GetParams(ctx)
	borrowers := k.GetLtvIndexSlice(ctx, params.CheckLtvIndexCount)

	for _, borrower := range borrowers {
		_, err := k.AttemptKeeperLiquidation(ctx, sdk.AccAddress(types.LiquidatorAccount), borrower)
		if err != nil {
			if !errors.Is(err, types.ErrBorrowNotLiquidatable) {
				panic(err) // TODO: should this panic?
			}
		}
	}
	return nil
}

// AttemptKeeperLiquidation enables a keeper to liquidate an individual borrower's position
// TODO: does this need to require a bool?
func (k Keeper) AttemptKeeperLiquidation(ctx sdk.Context, keeper sdk.AccAddress, borrower sdk.AccAddress) (bool, error) {
	prevLtv, err := k.GetStoreLTV(ctx, borrower)
	if err != nil {
		return false, err
	}

	// k.SyncSupplyInterest(ctx, borrower) // TODO: must add
	k.SyncBorrowInterest(ctx, borrower)

	deposit, found := k.GetDeposit(ctx, borrower)
	if !found {
		return false, types.ErrDepositNotFound
	}

	borrow, found := k.GetBorrow(ctx, borrower)
	if !found {
		return false, types.ErrBorrowNotFound
	}

	isWithinRange, err := k.IsWithinValidLtvRange(ctx, deposit, borrow)
	if err != nil {
		return false, err
	}
	if isWithinRange {
		return false, sdkerrors.Wrapf(types.ErrBorrowNotLiquidatable, "position is within valid LTV range")
	}

	// Sending coins to auction module with keeper address getting % of the profits
	borrowDenoms := getDenoms(borrow.Amount)
	depositDenoms := getDenoms(deposit.Amount)
	err = k.SeizeDeposits(ctx, keeper, deposit, borrow, depositDenoms, borrowDenoms)
	if err != nil {
		return false, err
	}

	k.DeleteDepositBorrowAndLtvIndex(ctx, deposit, borrow, prevLtv)
	return true, nil
}

// SeizeDeposits seizes a list of deposits and sends them to auction
func (k Keeper) SeizeDeposits(ctx sdk.Context, keeper sdk.AccAddress, deposit types.Deposit,
	borrow types.Borrow, dDenoms, bDenoms []string) error {
	liqMap, err := k.LoadLiquidationData(ctx, deposit, borrow)
	if err != nil {
		return err
	}

	// Seize % of every deposit and send to the keeper
	aucDeposits := sdk.Coins{}
	for _, depCoin := range deposit.Amount {
		denom := depCoin.Denom
		amount := depCoin.Amount
		mm, _ := k.GetMoneyMarket(ctx, denom)

		// No rewards for anyone if liquidated by LTV index
		if !keeper.Equals(sdk.AccAddress(types.LiquidatorAccount)) {
			keeperReward := mm.KeeperRewardPercentage.MulInt(amount).TruncateInt()
			if keeperReward.GT(sdk.ZeroInt()) {
				// Send keeper their reward
				keeperCoin := sdk.NewCoin(denom, keeperReward)
				err := k.supplyKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleAccountName, keeper, sdk.NewCoins(keeperCoin))
				if err != nil {
					return err
				}
				amount = amount.Sub(keeperReward)
			}
		}

		// Add remaining deposit coin to aucDeposits
		aucDeposits = aucDeposits.Add(sdk.NewCoin(denom, amount))
	}

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

	err = k.StartAuctions(ctx, deposit.Depositor, borrow.Amount, aucDeposits, depositCoinValues, borrowCoinValues, ltv, liqMap)
	if err != nil {
		return err
	}

	return nil
}

// StartAuctions attempts to start auctions for seized assets
func (k Keeper) StartAuctions(ctx sdk.Context, borrower sdk.AccAddress, borrows, deposits sdk.Coins,
	depositCoinValues, borrowCoinValues types.ValuationMap, ltv sdk.Dec, liqMap map[string]LiqData) error {
	// Sort keys to ensure deterministic behavior
	bKeys := borrowCoinValues.GetSortedKeys()
	dKeys := depositCoinValues.GetSortedKeys()

	// Set up auction constants
	returnAddrs := []sdk.AccAddress{borrower}
	weights := []sdk.Int{sdk.NewInt(100)}
	debt := sdk.NewCoin("debt", sdk.ZeroInt())

	for _, bKey := range bKeys {
		bValue := borrowCoinValues.Get(bKey)
		maxLotSize := bValue.Quo(ltv)

		for _, dKey := range dKeys {
			dValue := depositCoinValues.Get(dKey)
			if maxLotSize.Equal(sdk.ZeroDec()) {
				break // exit out of the loop if we have cleared the full amount
			}

			if dValue.GTE(maxLotSize) { // We can start an auction for the whole borrow amount
				bid := sdk.NewCoin(bKey, borrows.AmountOf(bKey))
				lotSize := maxLotSize.MulInt(liqMap[dKey].conversionFactor).Quo(liqMap[dKey].price)
				lot := sdk.NewCoin(dKey, lotSize.TruncateInt())
				// Sanity check that we can deliver coins to the liquidator account
				if deposits.AmountOf(dKey).LT(lot.Amount) {
					return types.ErrInsufficientCoins
				}

				// Start auction: bid = full borrow amount, lot = maxLotSize
				err := k.supplyKeeper.SendCoinsFromModuleToModule(ctx, types.ModuleAccountName, types.LiquidatorAccount, sdk.NewCoins(lot))
				if err != nil {
					return err
				}
				_, err = k.auctionKeeper.StartCollateralAuction(ctx, types.LiquidatorAccount, lot, bid, returnAddrs, weights, debt)
				if err != nil {
					return err
				}

				// Update USD valuation maps
				borrowCoinValues.SetZero(bKey)
				depositCoinValues.Decrement(dKey, maxLotSize)
				// Update deposits, borrows
				borrows = borrows.Sub(sdk.NewCoins(bid))
				deposits = deposits.Sub(sdk.NewCoins(lot))
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

				// Sanity check that we can deliver coins to the liquidator account
				if deposits.AmountOf(dKey).LT(lot.Amount) {
					return types.ErrInsufficientCoins
				}

				// Start auction: bid = maxBid, lot = whole deposit amount
				err := k.supplyKeeper.SendCoinsFromModuleToModule(ctx, types.ModuleAccountName, types.LiquidatorAccount, sdk.NewCoins(lot))
				if err != nil {
					return err
				}
				_, err = k.auctionKeeper.StartCollateralAuction(ctx, types.LiquidatorAccount, lot, bid, returnAddrs, weights, debt)
				if err != nil {
					return err
				}

				// Update variables to account for partial auction
				borrowCoinValues.Decrement(bKey, maxBid)
				depositCoinValues.SetZero(dKey)
				// Update deposits, borrows
				borrows = borrows.Sub(sdk.NewCoins(bid))
				deposits = deposits.Sub(sdk.NewCoins(lot))
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
				return err
			}
		}
	}

	return nil
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

// UpdateBorrowAndLtvIndex updates a borrow and its LTV index value in the store
func (k Keeper) UpdateBorrowAndLtvIndex(ctx sdk.Context, borrow types.Borrow, newLtv, oldLtv sdk.Dec) {
	k.RemoveFromLtvIndex(ctx, oldLtv, borrow.Borrower)
	k.SetBorrow(ctx, borrow)
	k.InsertIntoLtvIndex(ctx, newLtv, borrow.Borrower)
}

// UpdateDepositAndLtvIndex updates a deposit and its LTV index value in the store
func (k Keeper) UpdateDepositAndLtvIndex(ctx sdk.Context, deposit types.Deposit, newLtv, oldLtv sdk.Dec) {
	k.RemoveFromLtvIndex(ctx, oldLtv, deposit.Depositor)
	k.SetDeposit(ctx, deposit)
	k.InsertIntoLtvIndex(ctx, newLtv, deposit.Depositor)
}

// DeleteDepositBorrowAndLtvIndex deletes deposit, borrow, and ltv index
func (k Keeper) DeleteDepositBorrowAndLtvIndex(ctx sdk.Context, deposit types.Deposit, borrow types.Borrow, oldLtv sdk.Dec) {
	k.RemoveFromLtvIndex(ctx, oldLtv, deposit.Depositor)
	k.DeleteDeposit(ctx, deposit)
	k.DeleteBorrow(ctx, borrow)
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
