package operations

import (
	"fmt"
	"math"
	"math/rand"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authexported "github.com/cosmos/cosmos-sdk/x/auth/exported"
	"github.com/cosmos/cosmos-sdk/x/simulation"

	"github.com/kava-labs/kava/x/cdp"
	"github.com/kava-labs/kava/x/pricefeed"

	"github.com/kava-labs/kava/x/incentive"
	"github.com/kava-labs/kava/x/incentive/keeper"
	"github.com/kava-labs/kava/x/incentive/types"
)

var (
	noOpMsg = simulation.NoOpMsg(incentive.ModuleName)
)

// SimulateMsgClaimReward generates a MsgClaimReward
func SimulateMsgClaimReward(ak auth.AccountKeeper, k keeper.Keeper, ck cdp.Keeper, pk pricefeed.Keeper) simulation.Operation {
	handler := incentive.NewHandler(k)
	cdpHandler := cdp.NewHandler(ck)

	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simulation.Account) (
		simulation.OperationMsg, []simulation.FutureOperation, error) {

		var accounts []authexported.Account
		for _, acc := range accs {
			accounts = append(accounts, ak.GetAccount(ctx, acc.Address))
		}

		claimer := GetRandomClaimer(r, accounts).GetAddress()

		// Randomly select a reward's collateral type from rewards
		params := k.GetParams(ctx)
		if len(params.Rewards) == 0 {
			return noOpMsg, nil, fmt.Errorf("no rewards found in incentive module params")
		}
		reward := params.Rewards[r.Intn(len(params.Rewards))]

		// -------------------------------------------------------------------------------
		//	 				TODO: move CDP creation to separate func
		// -------------------------------------------------------------------------------
		//		simOpMsg, simFutureOp, err := simulateClaimerMsgCreateCDP(claimer, reward, ak, ck, pk)
		//		 if err != nil {
		// 			return noOpMsg, nil, fmt.Errorf("failed to submit new MsgCreateCDP")
		// 		}
		// 		return simOpMsg, simFutureOp, err
		// -------------------------------------------------------------------------------
		_, found := ck.GetCdpByOwnerAndDenom(ctx, claimer, reward.Denom)
		if !found {
			collateralParam, found := ck.GetCollateral(ctx, reward.Denom)
			if !found {
				return noOpMsg, nil, fmt.Errorf(fmt.Sprintf("collateral %s not supported by cdp module", reward.Denom))
			}

			collateralPrice, err := pk.GetCurrentPrice(ctx, fmt.Sprintf("%s:usd", collateralParam.Denom))
			if err != nil {
				return noOpMsg, nil, fmt.Errorf(fmt.Sprintf("couldn't fetch current price from pricefeed, err: %s", err.Error()))
			}
			debtAsset := collateralParam.DebtLimit[r.Intn(len(collateralParam.DebtLimit))] // Only 1 available
			debtParam, found := ck.GetDebtParam(ctx, debtAsset.Denom)
			if !found {
				return noOpMsg, nil, fmt.Errorf(fmt.Sprintf("debt asset %s not found in cdp module's debt params", debtAsset.Denom))
			}

			// Check that the sender has coins of this type
			claimerBalance := ak.GetAccount(ctx, claimer).GetCoins()
			availableAmount := claimerBalance.AmountOf(reward.Denom)

			debtConversionFactor := sdk.NewInt(int64(math.Pow(10, float64(debtParam.ConversionFactor.Int64()))))
			debtLimitInUSD := sdk.NewDec(debtParam.DebtFloor.Quo(debtConversionFactor).Int64())
			debtLimitInCollateralValue := debtLimitInUSD.Quo(collateralPrice.Price)

			minDeposit := debtLimitInCollateralValue.Mul(collateralParam.LiquidationRatio)
			truncatedMinDeposit := minDeposit.TruncateInt().Add(sdk.OneInt())
			if !availableAmount.GTE(truncatedMinDeposit) {
				return noOpMsg, nil, fmt.Errorf("claimer doesn't have enough to make the minimum deposit")
			}

			// Max deposit is 1/1000 total available coins
			maxDeposit := claimerBalance.AmountOf(reward.Denom).Quo(sdk.NewInt(1000))
			// Calculate random collateral deposit size within range
			deposit := simulation.RandIntBetween(r, int(truncatedMinDeposit.Int64()), int(maxDeposit.Int64()))
			collateral := sdk.NewCoins(sdk.NewInt64Coin(reward.Denom, int64(deposit)))

			// Convert deposit to value in debt denom and reverse apply liquidation ratio
			priceShifted := ShiftDec(collateralPrice.Price, debtParam.ConversionFactor)
			collateralDepositValue := ShiftDec(sdk.NewDec(int64(deposit)), debtParam.ConversionFactor.Neg()).Mul(priceShifted)
			principalAmount := collateralDepositValue.Quo(collateralParam.LiquidationRatio).TruncateInt()
			principal := sdk.NewCoins(sdk.NewCoin(debtParam.Denom, principalAmount))

			// Create MsgCreateCDP
			msgCreateCDP := cdp.NewMsgCreateCDP(claimer, collateral, principal)
			err = msgCreateCDP.ValidateBasic()
			if err != nil {
				return simulation.NoOpMsg(cdp.ModuleName), nil, fmt.Errorf("expected msg to pass ValidateBasic: %v", err)
			}
			// Submit msg
			ok := submitMsg(ctx, cdpHandler, msgCreateCDP)
			return simulation.NewOperationMsg(msgCreateCDP, ok, ""), nil, nil
		}

		// -------------------------------------------------------------------------------
		// 		TODO: submit as future op or move ctx forward to valid block time
		// -------------------------------------------------------------------------------
		// var futureOp simulation.FutureOperation
		// if ok {
		// 	acc := simulation.RandomAcc(r, accs)
		// 	executionBlock := ctx.BlockHeight() +
		// 	futureOp = loadClaimFutureOp(acc.Address, denom, executionBlock, handler)
		// }
		// -------------------------------------------------------------------------------
		msg := types.NewMsgClaimReward(claimer, reward.Denom)
		if err := msg.ValidateBasic(); err != nil {
			return noOpMsg, nil, fmt.Errorf("expected MsgClaimReward to pass ValidateBasic: %s", err)
		}
		ok := submitMsg(ctx, handler, msg)
		return simulation.NewOperationMsg(msg, ok, ""), nil, nil
	}
}

func loadClaimRewardFutureOp(sender sdk.AccAddress, denom string, height int64, handler sdk.Handler) simulation.FutureOperation {
	claimOp := func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simulation.Account) (
		simulation.OperationMsg, []simulation.FutureOperation, error) {

		// Build the refund msg and validate basic
		claimRewardMsg := types.NewMsgClaimReward(sender, denom)
		if err := claimRewardMsg.ValidateBasic(); err != nil {
			return noOpMsg, nil, fmt.Errorf("expected MsgClaimReward to pass ValidateBasic: %s", err)
		}

		// Test msg submission at target block height
		ok := handler(ctx.WithBlockHeight(height), claimRewardMsg).IsOK()
		return simulation.NewOperationMsg(claimRewardMsg, ok, ""), nil, nil
	}

	return simulation.FutureOperation{
		BlockHeight: int(height),
		Op:          claimOp,
	}
}

func submitMsg(ctx sdk.Context, handler sdk.Handler, msg sdk.Msg) (ok bool) {
	ctx, write := ctx.CacheContext()
	ok = handler(ctx, msg).IsOK()
	if ok {
		write()
	}
	return ok
}

// GetRandomClaimer gets a random account from the set of claimer accounts
func GetRandomClaimer(r *rand.Rand, accounts []authexported.Account) authexported.Account {
	claimers := LoadOpClaimers(accounts)
	return claimers[r.Intn(len(claimers))]
}

// LoadOpClaimers loads the first 10 accounts from auth
func LoadOpClaimers(accounts []authexported.Account) []authexported.Account {
	var claimers []authexported.Account
	for i, acc := range accounts {
		if i < 10 {
			claimers = append(claimers, acc)
		} else {
			break
		}
	}
	return claimers
}

// ShiftDec applies conversion factor (taken from x/cdp/simulation/operations)
func ShiftDec(x sdk.Dec, places sdk.Int) sdk.Dec {
	neg := places.IsNegative()
	for i := 0; i < int(abs(places.Int64())); i++ {
		if neg {
			x = x.Mul(sdk.MustNewDecFromStr("0.1"))
		} else {
			x = x.Mul(sdk.NewDecFromInt(sdk.NewInt(10)))
		}

	}
	return x
}

// abs returns the absolute value of x.
func abs(x int64) int64 {
	if x < 0 {
		return -x
	}
	return x
}
