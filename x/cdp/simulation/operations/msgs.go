package operations

import (
	"fmt"
	"math/rand"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/simulation"
	"github.com/kava-labs/kava/x/cdp"
	"github.com/kava-labs/kava/x/pricefeed"
)

// SimulateMsgCreateOrDepositCdp generates a MsgCreateCdp or MsgDepositCdp with random values.
func SimulateMsgCreateOrDepositCdp(ak auth.AccountKeeper, k cdp.Keeper, pfk pricefeed.Keeper) simulation.Operation {
	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simulation.Account) (
		opMsg simulation.OperationMsg, fOps []simulation.FutureOperation, err error) {

		handler := cdp.NewHandler(k)
		acc := simulation.RandomAcc(r, accs)
		coins := ak.GetAccount(ctx, acc.Address).GetCoins()
		collateralParams := k.GetParams(ctx).CollateralParams
		if len(collateralParams) == 0 {
			return simulation.NoOpMsg(cdp.ModuleName), nil, nil
		}
		randCollateralParams := collateralParams[r.Intn(len(collateralParams))]
		randDebtAsset := randCollateralParams.DebtLimit[r.Intn(len(randCollateralParams.DebtLimit))]
		randDebtParam, _ := k.GetDebtParam(ctx, randDebtAsset.Denom)
		if coins.AmountOf(randCollateralParams.Denom).IsZero() {
			return simulation.NoOpMsg(cdp.ModuleName), nil, nil
		}

		_, found := k.GetCdpByOwnerAndDenom(ctx, acc.Address, randCollateralParams.Denom)
		if !found {
			price, err := pfk.GetCurrentPrice(ctx, randCollateralParams.MarketID)
			if err != nil {
				return simulation.NoOpMsg(cdp.ModuleName), nil, err
			}
			// convert the price to the same units as the debt param
			priceShifted := ShiftDec(price.Price, randDebtParam.ConversionFactor)
			// calculate the minimum amount of collateral that is needed to create a cdp with the debt floor amount of debt and the minimum liquidation ratio
			// (debtFloor * liquidationRatio)/priceShifted
			minCollateralDeposit := (sdk.NewDecFromInt(randDebtParam.DebtFloor).Mul(randCollateralParams.LiquidationRatio)).Quo(priceShifted)
			// convert to proper collateral units
			minCollateralDeposit = ShiftDec(minCollateralDeposit, randCollateralParams.ConversionFactor)
			// convert to integer and always round up
			minCollateralDepositRounded := minCollateralDeposit.TruncateInt().Add(sdk.OneInt())
			// if the account has less than the min deposit, return
			if coins.AmountOf(randCollateralParams.Denom).LT(minCollateralDepositRounded) {
				return simulation.NoOpMsg(cdp.ModuleName), nil, nil
			}
			// set the max collateral deposit to the amount of coins in the account
			maxCollateralDeposit := coins.AmountOf(randCollateralParams.Denom)

			// randomly select a collateral deposit amount
			collateralDeposit := sdk.NewInt(int64(simulation.RandIntBetween(r, int(minCollateralDepositRounded.Int64()), int(maxCollateralDeposit.Int64()))))
			fmt.Printf("%s\n", collateralDeposit)
			// calculate how much the randomly selected deposit is worth
			collateralDepositValue := ShiftDec(sdk.NewDecFromInt(collateralDeposit), randCollateralParams.ConversionFactor.Neg()).Mul(priceShifted)
			// calculate the max amount of debt that could be drawn for the chosen deposit
			maxDebtDraw := collateralDepositValue.Quo(randCollateralParams.LiquidationRatio).TruncateInt()
			// randomly select a debt draw amount
			debtDraw := sdk.NewInt(int64(simulation.RandIntBetween(r, int(randDebtParam.DebtFloor.Int64()), int(maxDebtDraw.Int64()))))
			msg := cdp.NewMsgCreateCDP(acc.Address, sdk.NewCoins(sdk.NewCoin(randCollateralParams.Denom, collateralDeposit)), sdk.NewCoins(sdk.NewCoin(randDebtParam.Denom, debtDraw)))
			if msg.ValidateBasic() != nil {
				return simulation.NoOpMsg(cdp.ModuleName), nil, fmt.Errorf("expected msg to pass ValidateBasic: %s", msg.GetSignBytes())
			}
			ok := simulateHandleMsgCreateCdp(msg, handler, ctx)
			return simulation.NewOperationMsg(msg, ok, "create cdp"), nil, nil
		}

		// TODO a cdp already exists, deposit to it, draw debt from it, or repay debt to it
		return simulation.NoOpMsg(cdp.ModuleName), nil, nil
	}
}

func simulateHandleMsgCreateCdp(msg cdp.MsgCreateCDP, handler sdk.Handler, ctx sdk.Context) (ok bool) {
	ctx, write := ctx.CacheContext()
	ok = handler(ctx, msg).IsOK()
	if ok {
		write()
	}
	return ok
}
