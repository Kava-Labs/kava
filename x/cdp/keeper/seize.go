package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/kava-labs/kava/x/cdp/types"
)

// AttemptKeeperLiquidation liquidates the cdp with the input collateral type and owner if it is below the required collateralization ratio
// if the cdp is liquidated, the keeper that sent the transaction is rewarded a percentage of the collateral according to that collateral types'
// keeper reward percentage.
func (k Keeper) AttemptKeeperLiquidation(ctx sdk.Context, keeper, owner sdk.AccAddress, collateralType string) error {
	cdp, found := k.GetCdpByOwnerAndCollateralType(ctx, owner, collateralType)
	if !found {
		return sdkerrors.Wrapf(types.ErrCdpNotFound, "owner %s, denom %s", owner, collateralType)
	}
	k.hooks.BeforeCDPModified(ctx, cdp)
	cdp = k.SynchronizeInterest(ctx, cdp)

	err := k.ValidateLiquidation(ctx, cdp.Collateral, cdp.Type, cdp.Principal, cdp.AccumulatedFees)
	if err != nil {
		return err
	}
	cdp, err = k.payoutKeeperLiquidationReward(ctx, keeper, cdp)
	if err != nil {
		return err
	}
	return k.SeizeCollateral(ctx, cdp)
}

// SeizeCollateral liquidates the collateral in the input cdp.
// the following operations are performed:
// 1. Collateral for all deposits is sent from the cdp module to the liquidator module account
// 2. The liquidation penalty is applied
// 3. Debt coins are sent from the cdp module to the liquidator module account
// 4. The total amount of principal outstanding for that collateral type is decremented
// (this is the equivalent of saying that fees are no longer accumulated by a cdp once it gets liquidated)
func (k Keeper) SeizeCollateral(ctx sdk.Context, cdp types.CDP) error {
	// Calculate the previous collateral ratio
	oldCollateralToDebtRatio := k.CalculateCollateralToDebtRatio(ctx, cdp.Collateral, cdp.Type, cdp.GetTotalPrincipal())

	// Move debt coins from cdp to liquidator account
	deposits := k.GetDeposits(ctx, cdp.ID)
	debt := cdp.GetTotalPrincipal().Amount
	modAccountDebt := k.getModAccountDebt(ctx, types.ModuleName)
	debt = sdk.MinInt(debt, modAccountDebt)
	debtCoin := sdk.NewCoin(k.GetDebtDenom(ctx), debt)
	err := k.supplyKeeper.SendCoinsFromModuleToModule(ctx, types.ModuleName, types.LiquidatorMacc, sdk.NewCoins(debtCoin))
	if err != nil {
		return err
	}

	// liquidate deposits and send collateral from cdp to liquidator
	for _, dep := range deposits {
		err := k.supplyKeeper.SendCoinsFromModuleToModule(ctx, types.ModuleName, types.LiquidatorMacc, sdk.NewCoins(dep.Amount))
		if err != nil {
			return err
		}
		k.DeleteDeposit(ctx, dep.CdpID, dep.Depositor)

		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeCdpLiquidation,
				sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
				sdk.NewAttribute(types.AttributeKeyCdpID, fmt.Sprintf("%d", cdp.ID)),
				sdk.NewAttribute(types.AttributeKeyDeposit, dep.String()),
			),
		)
	}

	err = k.AuctionCollateral(ctx, deposits, cdp.Type, debt, cdp.Principal.Denom)
	if err != nil {
		return err
	}

	// Decrement total principal for this collateral type
	coinsToDecrement := cdp.GetTotalPrincipal()
	k.DecrementTotalPrincipal(ctx, cdp.Type, coinsToDecrement)

	// Delete CDP from state
	k.RemoveCdpOwnerIndex(ctx, cdp)
	k.RemoveCdpCollateralRatioIndex(ctx, cdp.Type, cdp.ID, oldCollateralToDebtRatio)
	return k.DeleteCDP(ctx, cdp)
}

// LiquidateCdps seizes collateral from all CDPs below the input liquidation ratio
func (k Keeper) LiquidateCdps(ctx sdk.Context, marketID string, collateralType string, liquidationRatio sdk.Dec, slice sdk.Int) error {
	price, err := k.pricefeedKeeper.GetCurrentPrice(ctx, marketID)
	if err != nil {
		return err
	}
	priceDivLiqRatio := price.Price.Quo(liquidationRatio)
	if priceDivLiqRatio.IsZero() {
		priceDivLiqRatio = sdk.SmallestDec()
	}
	// price = $0.5
	// liquidation ratio = 1.5
	// normalizedRatio = (1/(0.5/1.5)) = 3
	normalizedRatio := sdk.OneDec().Quo(priceDivLiqRatio)
	cdpsToLiquidate := k.GetSliceOfCDPsByRatioAndType(ctx, slice, normalizedRatio, collateralType)
	for _, c := range cdpsToLiquidate {
		k.hooks.BeforeCDPModified(ctx, c)
		err := k.SeizeCollateral(ctx, c)
		if err != nil {
			return err
		}
	}
	return nil
}

// ApplyLiquidationPenalty multiplies the input debt amount by the liquidation penalty
func (k Keeper) ApplyLiquidationPenalty(ctx sdk.Context, collateralType string, debt sdk.Int) sdk.Int {
	penalty := k.getLiquidationPenalty(ctx, collateralType)
	return sdk.NewDecFromInt(debt).Mul(penalty).RoundInt()
}

// ValidateLiquidation validate that adding the input principal puts the cdp below the liquidation ratio
func (k Keeper) ValidateLiquidation(ctx sdk.Context, collateral sdk.Coin, collateralType string, principal sdk.Coin, fees sdk.Coin) error {
	collateralizationRatio, err := k.CalculateCollateralizationRatio(ctx, collateral, collateralType, principal, fees, spot)
	if err != nil {
		return err
	}
	liquidationRatio := k.getLiquidationRatio(ctx, collateralType)
	if collateralizationRatio.GT(liquidationRatio) {
		return sdkerrors.Wrapf(types.ErrNotLiquidatable, "collateral %s, collateral ratio %s, liquidation ratio %s", collateral.Denom, collateralizationRatio, liquidationRatio)
	}
	return nil
}

func (k Keeper) getModAccountDebt(ctx sdk.Context, accountName string) sdk.Int {
	macc := k.supplyKeeper.GetModuleAccount(ctx, accountName)
	return macc.GetCoins().AmountOf(k.GetDebtDenom(ctx))
}

func (k Keeper) payoutKeeperLiquidationReward(ctx sdk.Context, keeper sdk.AccAddress, cdp types.CDP) (types.CDP, error) {
	collateralParam, found := k.GetCollateral(ctx, cdp.Type)
	if !found {
		return types.CDP{}, sdkerrors.Wrapf(types.ErrInvalidCollateral, "%s", cdp.Type)
	}
	reward := cdp.Collateral.Amount.ToDec().Mul(collateralParam.KeeperRewardPercentage).RoundInt()
	rewardCoin := sdk.NewCoin(cdp.Collateral.Denom, reward)
	paidReward := false
	deposits := k.GetDeposits(ctx, cdp.ID)
	for _, dep := range deposits {
		if dep.Amount.IsGTE(rewardCoin) {
			dep.Amount = dep.Amount.Sub(rewardCoin)
			k.SetDeposit(ctx, dep)
			paidReward = true
			break
		}
	}
	if !paidReward {
		return cdp, nil
	}
	err := k.supplyKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, keeper, sdk.NewCoins(rewardCoin))
	if err != nil {
		return types.CDP{}, err
	}
	cdp.Collateral = cdp.Collateral.Sub(rewardCoin)
	ratio := k.CalculateCollateralToDebtRatio(ctx, cdp.Collateral, cdp.Type, cdp.GetTotalPrincipal())
	err = k.UpdateCdpAndCollateralRatioIndex(ctx, cdp, ratio)
	if err != nil {
		return types.CDP{}, err
	}
	return cdp, nil
}
