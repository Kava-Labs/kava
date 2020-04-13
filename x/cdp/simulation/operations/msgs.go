package operations

import (
	"fmt"
	"math/rand"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authexported "github.com/cosmos/cosmos-sdk/x/auth/exported"
	"github.com/cosmos/cosmos-sdk/x/simulation"
	"github.com/kava-labs/kava/x/cdp"
	"github.com/kava-labs/kava/x/pricefeed"
)

// SimulateMsgCdp generates a MsgCreateCdp or MsgDepositCdp with random values.
func SimulateMsgCdp(ak auth.AccountKeeper, k cdp.Keeper, pfk pricefeed.Keeper) simulation.Operation {
	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simulation.Account) (
		opMsg simulation.OperationMsg, fOps []simulation.FutureOperation, err error) {

		handler := cdp.NewHandler(k)
		simacc := simulation.RandomAcc(r, accs)
		acc := ak.GetAccount(ctx, simacc.Address)
		if acc == nil {
			return simulation.NoOpMsg(cdp.ModuleName), nil, nil
		}
		coins := acc.GetCoins()
		collateralParams := k.GetParams(ctx).CollateralParams
		if len(collateralParams) == 0 {
			return simulation.NoOpMsg(cdp.ModuleName), nil, nil
		}
		randCollateralParam := collateralParams[r.Intn(len(collateralParams))]
		randDebtAsset := randCollateralParam.DebtLimit[r.Intn(len(randCollateralParam.DebtLimit))]
		randDebtParam, _ := k.GetDebtParam(ctx, randDebtAsset.Denom)
		if coins.AmountOf(randCollateralParam.Denom).IsZero() {
			return simulation.NoOpMsg(cdp.ModuleName), nil, nil
		}

		price, err := pfk.GetCurrentPrice(ctx, randCollateralParam.MarketID)
		if err != nil {
			return simulation.NoOpMsg(cdp.ModuleName), nil, err
		}
		// convert the price to the same units as the debt param
		priceShifted := ShiftDec(price.Price, randDebtParam.ConversionFactor)

		existingCDP, found := k.GetCdpByOwnerAndDenom(ctx, acc.GetAddress(), randCollateralParam.Denom)
		if !found {
			// calculate the minimum amount of collateral that is needed to create a cdp with the debt floor amount of debt and the minimum liquidation ratio
			// (debtFloor * liquidationRatio)/priceShifted
			minCollateralDeposit := (sdk.NewDecFromInt(randDebtParam.DebtFloor).Mul(randCollateralParam.LiquidationRatio)).Quo(priceShifted)
			// convert to proper collateral units
			minCollateralDeposit = ShiftDec(minCollateralDeposit, randCollateralParam.ConversionFactor)
			// convert to integer and always round up
			minCollateralDepositRounded := minCollateralDeposit.TruncateInt().Add(sdk.OneInt())
			// if the account has less than the min deposit, return
			if coins.AmountOf(randCollateralParam.Denom).LT(minCollateralDepositRounded) {
				return simulation.NoOpMsg(cdp.ModuleName), nil, nil
			}
			// set the max collateral deposit to the amount of coins in the account
			maxCollateralDeposit := coins.AmountOf(randCollateralParam.Denom)

			// randomly select a collateral deposit amount
			collateralDeposit := sdk.NewInt(int64(simulation.RandIntBetween(r, int(minCollateralDepositRounded.Int64()), int(maxCollateralDeposit.Int64()))))
			// calculate how much the randomly selected deposit is worth
			collateralDepositValue := ShiftDec(sdk.NewDecFromInt(collateralDeposit), randCollateralParam.ConversionFactor.Neg()).Mul(priceShifted)
			// calculate the max amount of debt that could be drawn for the chosen deposit
			maxDebtDraw := collateralDepositValue.Quo(randCollateralParam.LiquidationRatio).TruncateInt()
			// randomly select a debt draw amount
			debtDraw := sdk.NewInt(int64(simulation.RandIntBetween(r, int(randDebtParam.DebtFloor.Int64()), int(maxDebtDraw.Int64()))))
			msg := cdp.NewMsgCreateCDP(acc.GetAddress(), sdk.NewCoins(sdk.NewCoin(randCollateralParam.Denom, collateralDeposit)), sdk.NewCoins(sdk.NewCoin(randDebtParam.Denom, debtDraw)))
			if msg.ValidateBasic() != nil {
				return simulation.NoOpMsg(cdp.ModuleName), nil, fmt.Errorf("expected msg to pass ValidateBasic: %s", msg.GetSignBytes())
			}
			ok := submitMsg(msg, handler, ctx)
			if !ok {
				return simulation.NoOpMsg(cdp.ModuleName), nil, fmt.Errorf("could not submit create cdp msg")
			}
			return simulation.NewOperationMsg(msg, ok, "create cdp"), nil, nil
		}

		// TODO a cdp already exists, deposit to it, draw debt from it, or repay debt to it
		// close 25% of the time
		if canClose(acc, existingCDP, randDebtParam.Denom) && shouldClose(r) {
			repaymentAmount := coins.AmountOf(randDebtParam.Denom)
			msg := cdp.NewMsgRepayDebt(acc.GetAddress(), randCollateralParam.Denom, sdk.NewCoins(sdk.NewCoin(randDebtParam.Denom, repaymentAmount)))
			err := msg.ValidateBasic()
			if err != nil {
				return simulation.NoOpMsg(cdp.ModuleName), nil, fmt.Errorf("expected repay (close) msg to pass ValidateBasic: %v", err)
			}
			ok := submitMsg(msg, handler, ctx)
			if !ok {
				return simulation.NoOpMsg(cdp.ModuleName), nil, fmt.Errorf("could not submit repay (close) msg")
			}
			return simulation.NewOperationMsg(msg, ok, "repay debt (close) cdp"), nil, nil
		}

		// deposit 25% of the time
		if hasCoins(acc, randCollateralParam.Denom) && shouldDeposit(r) {
			randDepositAmount := sdk.NewInt(int64(simulation.RandIntBetween(r, 1, int(acc.GetCoins().AmountOf(randCollateralParam.Denom).Int64()))))
			msg := cdp.NewMsgDeposit(acc.GetAddress(), acc.GetAddress(), sdk.NewCoins(sdk.NewCoin(randCollateralParam.Denom, randDepositAmount)))
			err := msg.ValidateBasic()
			if err != nil {
				return simulation.NoOpMsg(cdp.ModuleName), nil, fmt.Errorf("expected deposit msg to pass ValidateBasic: %v", err)
			}
			ok := submitMsg(msg, handler, ctx)
			if !ok {
				return simulation.NoOpMsg(cdp.ModuleName), nil, fmt.Errorf("could not submit deposit msg")
			}
			return simulation.NewOperationMsg(msg, ok, "deposit to cdp"), nil, nil
		}

		// draw debt 25% of the time
		if shouldDraw(r) {
			collateralShifted := ShiftDec(sdk.NewDecFromInt(existingCDP.Collateral.AmountOf(randCollateralParam.Denom)), randCollateralParam.ConversionFactor.Neg())
			collateralValue := collateralShifted.Mul(priceShifted)
			debt := (existingCDP.Principal.Add(existingCDP.AccumulatedFees)).AmountOf(randDebtParam.Denom)
			maxTotalDebt := collateralValue.Quo(randCollateralParam.LiquidationRatio)
			maxDebt := maxTotalDebt.Sub(sdk.NewDecFromInt(debt)).TruncateInt().Sub(sdk.OneInt())
			randDrawAmount := sdk.NewInt(int64(simulation.RandIntBetween(r, 1, int(maxDebt.Int64()))))
			msg := cdp.NewMsgDrawDebt(acc.GetAddress(), randCollateralParam.Denom, sdk.NewCoins(sdk.NewCoin(randDebtParam.Denom, randDrawAmount)))
			err := msg.ValidateBasic()
			if err != nil {
				return simulation.NoOpMsg(cdp.ModuleName), nil, fmt.Errorf("expected draw msg to pass ValidateBasic: %v", err)
			}
			ok := submitMsg(msg, handler, ctx)
			if !ok {
				return simulation.NoOpMsg(cdp.ModuleName), nil, fmt.Errorf("could not submit draw msg")
			}
			return simulation.NewOperationMsg(msg, ok, "draw debt from cdp"), nil, nil

		}

		// repay debt 25% of the time
		if hasCoins(acc, randDebtParam.Denom) {
			debt := (existingCDP.Principal.Add(existingCDP.AccumulatedFees)).AmountOf(randDebtParam.Denom)
			maxRepay := acc.GetCoins().AmountOf(randDebtParam.Denom)
			payableDebt := debt.Sub(randDebtParam.DebtFloor)
			if maxRepay.GT(payableDebt) {
				maxRepay = payableDebt
			}
			randRepayAmount := sdk.NewInt(int64(simulation.RandIntBetween(r, 1, int(maxRepay.Int64()))))
			if debt.Equal(randDebtParam.DebtFloor) {
				if acc.GetCoins().AmountOf(randDebtParam.Denom).GTE(debt) {
					randRepayAmount = debt
				}
			}
			msg := cdp.NewMsgRepayDebt(acc.GetAddress(), randCollateralParam.Denom, sdk.NewCoins(sdk.NewCoin(randDebtParam.Denom, randRepayAmount)))
			err := msg.ValidateBasic()
			if err != nil {
				return simulation.NoOpMsg(cdp.ModuleName), nil, fmt.Errorf("expected repay msg to pass ValidateBasic: %s", msg.GetSignBytes())
			}
			ok := submitMsg(msg, handler, ctx)
			if !ok {
				return simulation.NoOpMsg(cdp.ModuleName), nil, fmt.Errorf("could not submit repay msg")
			}
			return simulation.NewOperationMsg(msg, ok, "repay debt cdp"), nil, nil
		}

		return simulation.NewOperationMsgBasic(cdp.ModuleName, "no-operation (no valid actions)", "", false, nil), nil, nil
	}
}

func submitMsg(msg sdk.Msg, handler sdk.Handler, ctx sdk.Context) (ok bool) {
	ctx, write := ctx.CacheContext()
	res := handler(ctx, msg)
	if res.IsOK() {
		write()
	} else {
		fmt.Println(res.Log)
	}
	return res.IsOK()
}

func shouldDraw(r *rand.Rand) bool {
	threshold := 50
	value := simulation.RandIntBetween(r, 1, 100)
	if value > threshold {
		return true
	}
	return false
}

func shouldDeposit(r *rand.Rand) bool {
	threshold := 66
	value := simulation.RandIntBetween(r, 1, 100)
	if value > threshold {
		return true
	}
	return false
}

func hasCoins(acc authexported.Account, denom string) bool {
	if acc.GetCoins().AmountOf(denom).IsZero() {
		return false
	}
	return true
}

func shouldClose(r *rand.Rand) bool {
	threshold := 75
	value := simulation.RandIntBetween(r, 1, 100)
	if value > threshold {
		return true
	}
	return false
}

func canClose(acc authexported.Account, c cdp.CDP, denom string) bool {
	repaymentAmount := c.Principal.Add(c.AccumulatedFees).AmountOf(denom)
	if acc.GetCoins().AmountOf(denom).GT(repaymentAmount) {
		return true
	}
	return false
}
