import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kava-labs/kava/x/cdp/types"
)

func (k Keeper) AddCdp(ctx sdk.Context, owner sdk.AccAddress, collateral sdk.Coins, debt sdk.Coins) sdk.Error {
	_, ok := k.GetCdpID(ctx, owner, collateral[0].Denom)
	if ok {
		return types.ErrCdpExists(k.codespace, owner, collateral[0].Denom)
	}
	err := k.ValidateCollateral(ctx, collateral)
	if err != nil {
		return err
	}
	err = k.ValidateDebt(ctx, debt)
	if err != nil {
		return err
	}
	err = k.ValidateCollateralRatio(ctx, collateral, debt)
	if err != nil {
		return err
	}
	id := k.GetNextCdpID(ctx)
	cdp := types.NewCDP(id, owner, collateral, debt, ctx.BlockHeader().Time)
	k.supplyKeeper.SendCoinsFromAccountToModule(ctx, owner, types.ModuleName, collateral)
	k.MintStablecoins(ctx, types.ModuleName, debt)
	k.MintDebtcoins(ctx, types.ModuleName, debt)
	k.supplyKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, owner, debt)
	k.SetCDP(ctx, cdp)
	k.IndexCdbByOwner(ctx, cdp)
	collateralToDebtRatio := k.CalculateCollateralToDebtRatio(ctx, cdp)
	k.IndexCdpByCollateralRatio(ctx, cdp, collateralToDebtRatio)

	return nil
}