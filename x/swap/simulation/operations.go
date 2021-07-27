package simulation

import (
	"errors"
	"fmt"
	"math/big"
	"math/rand"
	"time"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/simapp/helpers"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authexported "github.com/cosmos/cosmos-sdk/x/auth/exported"
	"github.com/cosmos/cosmos-sdk/x/simulation"

	appparams "github.com/kava-labs/kava/app/params"
	"github.com/kava-labs/kava/x/swap/keeper"
	"github.com/kava-labs/kava/x/swap/types"
)

var (
	//nolint
	noOpMsg             = simulation.NoOpMsg(types.ModuleName)
	errorNotEnoughCoins = errors.New("account doesn't have enough coins")
)

// Simulation operation weights constants
const (
	OpWeightMsgDeposit            = "op_weight_msg_deposit"
	OpWeightMsgWithdraw           = "op_weight_msg_withdraw"
	OpWeightMsgSwapExactForTokens = "op_weight_msg_swap_exact_for_tokens"
	OpWeightMsgSwapForExactTokens = "op_weight_msg_swap_for_exact_tokens"
)

// WeightedOperations returns all the operations from the module with their respective weights
func WeightedOperations(
	appParams simulation.AppParams, cdc *codec.Codec, ak types.AccountKeeper, k keeper.Keeper,
) simulation.WeightedOperations {
	var weightMsgDeposit int
	var weightMsgWithdraw int
	var weightMsgSwapExactForTokens int
	var weightMsgSwapForExactTokens int

	appParams.GetOrGenerate(cdc, OpWeightMsgDeposit, &weightMsgDeposit, nil,
		func(_ *rand.Rand) {
			weightMsgDeposit = appparams.DefaultWeightMsgDeposit
		},
	)

	appParams.GetOrGenerate(cdc, OpWeightMsgWithdraw, &weightMsgWithdraw, nil,
		func(_ *rand.Rand) {
			weightMsgWithdraw = appparams.DefaultWeightMsgWithdraw
		},
	)

	appParams.GetOrGenerate(cdc, OpWeightMsgSwapExactForTokens, &weightMsgSwapExactForTokens, nil,
		func(_ *rand.Rand) {
			weightMsgSwapExactForTokens = appparams.DefaultWeightMsgSwapExactForTokens
		},
	)

	appParams.GetOrGenerate(cdc, OpWeightMsgSwapForExactTokens, &weightMsgSwapForExactTokens, nil,
		func(_ *rand.Rand) {
			weightMsgSwapForExactTokens = appparams.DefaultWeightMsgSwapForExactTokens
		},
	)

	return simulation.WeightedOperations{
		simulation.NewWeightedOperation(
			weightMsgDeposit,
			SimulateMsgDeposit(ak, k),
		),
		simulation.NewWeightedOperation(
			weightMsgWithdraw,
			SimulateMsgWithdraw(ak, k),
		),
		simulation.NewWeightedOperation(
			weightMsgSwapExactForTokens,
			SimulateMsgSwapExactForTokens(ak, k),
		),
		simulation.NewWeightedOperation(
			weightMsgSwapForExactTokens,
			SimulateMsgSwapForExactTokens(ak, k),
		),
	}
}

// SimulateMsgDeposit generates a MsgDeposit
func SimulateMsgDeposit(ak types.AccountKeeper, k keeper.Keeper) simulation.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simulation.Account, chainID string,
	) (simulation.OperationMsg, []simulation.FutureOperation, error) {
		// Get possible pools and shuffle so that deposits are evenly distributed across pools
		params := k.GetParams(ctx)
		allowedPools := params.AllowedPools
		r.Shuffle(len(allowedPools), func(i, j int) {
			allowedPools[i], allowedPools[j] = allowedPools[j], allowedPools[i]
		})

		// Find an account-pool pair that is likely to result in a successful deposit
		blockTime := ctx.BlockHeader().Time
		depositor, allowedPool, found := findValidAccountAllowedPoolPair(accs, allowedPools, func(acc simulation.Account, pool types.AllowedPool) bool {
			account := ak.GetAccount(ctx, acc.Address)

			err := validateDepositor(ctx, k, pool, account, blockTime)
			if err == errorNotEnoughCoins {
				return false // keep searching
			} else if err != nil {
				panic(err) // raise errors
			}
			return true // found valid pair
		})
		if !found {
			return simulation.NewOperationMsgBasic(types.ModuleName, "no-operation (no valid allowed pool and depositor)", "", false, nil), nil, nil
		}

		// Get random slippage amount between 1-99%
		slippageRaw, err := RandIntInclusive(r, sdk.OneInt(), sdk.NewInt(99))
		if err != nil {
			panic(err)
		}
		slippage := slippageRaw.ToDec().Quo(sdk.NewDec(100))

		// Generate random deadline
		deadline := genRandDeadline(r, blockTime)

		depositorAcc := ak.GetAccount(ctx, depositor.Address)
		depositorCoins := depositorAcc.SpendableCoins(blockTime)

		// Construct initial msg (without coin amounts)
		msg := types.NewMsgDeposit(depositorAcc.GetAddress(), sdk.Coin{}, sdk.Coin{}, slippage, deadline)

		// Populate msg with randomized token amounts
		pool, found := k.GetPool(ctx, allowedPool.Name())
		if !found { // Pool doesn't exist: first deposit
			depositTokenA := randCoinFromCoins(r, depositorCoins, allowedPool.TokenA)
			msg.TokenA = depositTokenA

			depositTokenB := randCoinFromCoins(r, depositorCoins, allowedPool.TokenB)
			msg.TokenB = depositTokenB
		} else { // Pool exists: successive deposit
			var denomX string // Denom X is the token denom in the pool with the larger amount
			var denomY string //  Denom Y is the token denom in the pool with the larger amount
			if pool.ReservesA.Amount.GTE(pool.ReservesB.Amount) {
				denomX = pool.ReservesA.Denom
				denomY = pool.ReservesB.Denom
			} else {
				denomX = pool.ReservesB.Denom
				denomY = pool.ReservesA.Denom
			}
			depositTokenY := randCoinFromCoins(r, depositorCoins, denomY)
			msg.TokenA = depositTokenY

			// Calculate the pool's slippage ratio and use it to build other coin
			ratio := pool.Reserves().AmountOf(denomX).ToDec().Quo(pool.Reserves().AmountOf(denomY).ToDec())
			amtTokenX := depositTokenY.Amount.ToDec().Mul(ratio).RoundInt()
			depositTokenX := sdk.NewCoin(denomX, amtTokenX)
			if depositorCoins.AmountOf(denomX).LT(amtTokenX) {
				return simulation.NewOperationMsgBasic(types.ModuleName, "no-operation (depositor has insufficient coins)", "", false, nil), nil, nil
			}
			msg.TokenB = depositTokenX
		}

		err = msg.ValidateBasic()
		if err != nil {
			return noOpMsg, nil, nil
		}

		tx := helpers.GenTx(
			[]sdk.Msg{msg},
			sdk.NewCoins(),
			helpers.DefaultGenTxGas,
			chainID,
			[]uint64{depositorAcc.GetAccountNumber()},
			[]uint64{depositorAcc.GetSequence()},
			depositor.PrivKey,
		)

		_, result, err := app.Deliver(tx)
		if err != nil {
			// to aid debugging, add the stack trace to the comment field of the returned opMsg
			return simulation.NewOperationMsg(msg, false, fmt.Sprintf("%+v", err)), nil, err
		}
		return simulation.NewOperationMsg(msg, true, result.Log), nil, nil
	}
}

// SimulateMsgWithdraw generates a MsgWithdraw
func SimulateMsgWithdraw(ak types.AccountKeeper, k keeper.Keeper) simulation.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simulation.Account, chainID string,
	) (simulation.OperationMsg, []simulation.FutureOperation, error) {

		poolRecords := k.GetAllPools(ctx)
		r.Shuffle(len(poolRecords), func(i, j int) {
			poolRecords[i], poolRecords[j] = poolRecords[j], poolRecords[i]
		})

		// Find an account-pool pair for which withdraw is possible
		withdrawer, poolRecord, found := findValidAccountPoolRecordPair(accs, poolRecords, func(acc simulation.Account, poolRecord types.PoolRecord) bool {
			_, found := k.GetDepositorShares(ctx, acc.Address, poolRecord.PoolID)
			if !found {
				return false // keep searching
			}
			return true
		})
		if !found {
			return simulation.NewOperationMsgBasic(types.ModuleName, "no-operation (no valid pool record and withdrawer)", "", false, nil), nil, nil
		}

		withdrawerAcc := ak.GetAccount(ctx, withdrawer.Address)
		shareRecord, _ := k.GetDepositorShares(ctx, withdrawerAcc.GetAddress(), poolRecord.PoolID)
		denominatedPool, err := types.NewDenominatedPoolWithExistingShares(poolRecord.Reserves(), poolRecord.TotalShares)
		if err != nil {
			return noOpMsg, nil, nil
		}
		coinsOwned := denominatedPool.ShareValue(shareRecord.SharesOwned)

		// Get random amount of shares between 2-10% of the total
		sharePercentage, err := RandIntInclusive(r, sdk.NewInt(2), sdk.NewInt(10))
		if err != nil {
			panic(err)
		}
		shares := shareRecord.SharesOwned.Mul(sharePercentage).Quo(sdk.NewInt(100))

		// Expect minimum token amounts relative to the % of shares owned and withdrawn
		oneLessThanSharePercentage := sharePercentage.Sub(sdk.OneInt())

		amtTokenAOwned := coinsOwned.AmountOf(poolRecord.ReservesA.Denom)
		minAmtTokenA := amtTokenAOwned.Mul(oneLessThanSharePercentage).Quo(sdk.NewInt(100))
		minTokenA := sdk.NewCoin(poolRecord.ReservesA.Denom, minAmtTokenA)

		amtTokenBOwned := coinsOwned.AmountOf(poolRecord.ReservesB.Denom)
		minTokenAmtB := amtTokenBOwned.Mul(oneLessThanSharePercentage).Quo(sdk.NewInt(100))
		minTokenB := sdk.NewCoin(poolRecord.ReservesB.Denom, minTokenAmtB)

		// Generate random deadline
		blockTime := ctx.BlockHeader().Time
		deadline := genRandDeadline(r, blockTime)

		// Construct MsgWithdraw
		msg := types.NewMsgWithdraw(withdrawerAcc.GetAddress(), shares, minTokenA, minTokenB, deadline)
		err = msg.ValidateBasic()
		if err != nil {
			return noOpMsg, nil, nil
		}

		tx := helpers.GenTx(
			[]sdk.Msg{msg},
			sdk.NewCoins(),
			helpers.DefaultGenTxGas,
			chainID,
			[]uint64{withdrawerAcc.GetAccountNumber()},
			[]uint64{withdrawerAcc.GetSequence()},
			withdrawer.PrivKey,
		)

		_, result, err := app.Deliver(tx)
		if err != nil {
			// to aid debugging, add the stack trace to the comment field of the returned opMsg
			return simulation.NewOperationMsg(msg, false, fmt.Sprintf("%+v", err)), nil, err
		}
		return simulation.NewOperationMsg(msg, true, result.Log), nil, nil

	}
}

// SimulateMsgSwapExactForTokens generates a MsgSwapExactForTokens
func SimulateMsgSwapExactForTokens(ak types.AccountKeeper, k keeper.Keeper) simulation.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simulation.Account, chainID string,
	) (simulation.OperationMsg, []simulation.FutureOperation, error) {

		poolRecords := k.GetAllPools(ctx)
		r.Shuffle(len(poolRecords), func(i, j int) {
			poolRecords[i], poolRecords[j] = poolRecords[j], poolRecords[i]
		})

		// Find an account-pool pair for which trade is possible
		trader, poolRecord, found := findValidAccountPoolRecordPair(accs, poolRecords, func(acc simulation.Account, poolRecord types.PoolRecord) bool {
			traderAcc := ak.GetAccount(ctx, acc.Address)
			balanceTokenA := traderAcc.GetCoins().AmountOf(poolRecord.ReservesA.Denom)
			balanceTokenB := traderAcc.GetCoins().AmountOf(poolRecord.ReservesB.Denom)
			if !balanceTokenA.IsPositive() || !balanceTokenB.IsPositive() {
				return false
			}
			return true
		})
		if !found {
			return simulation.NewOperationMsgBasic(types.ModuleName, "no-operation (no valid pool record and trader)", "", false, nil), nil, nil
		}

		// Select input token
		randInt, err := RandInt(r, sdk.OneInt(), sdk.NewInt(9))
		if err != nil {
			panic(err)
		}
		inputToken := poolRecord.ReservesA
		outputToken := poolRecord.ReservesB
		if randInt.Int64()%2 == 0 {
			inputToken = poolRecord.ReservesB
			outputToken = poolRecord.ReservesA
		}

		// Select entity (trader account or pool) with smaller token amount
		traderAcc := ak.GetAccount(ctx, trader.Address)
		maxTradeAmount := inputToken.Amount
		if traderAcc.GetCoins().AmountOf(inputToken.Denom).LT(inputToken.Amount) {
			maxTradeAmount = traderAcc.GetCoins().AmountOf(inputToken.Denom)
		}

		// Exact input token is between 2-10% of the max trade amount
		percentage, err := RandIntInclusive(r, sdk.NewInt(2), sdk.NewInt(10))
		if err != nil {
			panic(err)
		}
		tradeAmount := maxTradeAmount.Mul(percentage).Quo(sdk.NewInt(100))
		exactInputToken := sdk.NewCoin(inputToken.Denom, tradeAmount)

		// Calculate expected output coin
		globalSwapFee := k.GetSwapFee(ctx)
		tradeAmountAfterFee := exactInputToken.Amount.ToDec().Mul(sdk.OneDec().Sub(globalSwapFee)).TruncateInt()

		var outputAmt big.Int
		outputAmt.Mul(outputToken.Amount.BigInt(), tradeAmountAfterFee.BigInt())
		outputAmt.Quo(&outputAmt, inputToken.Amount.Add(tradeAmountAfterFee).BigInt())
		expectedOutTokenAmount := sdk.NewIntFromBigInt(&outputAmt)
		expectedOutputToken := sdk.NewCoin(outputToken.Denom, expectedOutTokenAmount)

		// Get random slippage amount between 50-100%
		slippageRaw, err := RandIntInclusive(r, sdk.NewInt(50), sdk.NewInt(99))
		if err != nil {
			panic(err)
		}
		slippage := slippageRaw.ToDec().Quo(sdk.NewDec(100))

		// Generate random deadline
		blockTime := ctx.BlockHeader().Time
		deadline := genRandDeadline(r, blockTime)

		// Construct MsgSwapExactForTokens
		msg := types.NewMsgSwapExactForTokens(traderAcc.GetAddress(), exactInputToken, expectedOutputToken, slippage, deadline)
		err = msg.ValidateBasic()
		if err != nil {
			return noOpMsg, nil, nil
		}

		tx := helpers.GenTx(
			[]sdk.Msg{msg},
			sdk.NewCoins(),
			helpers.DefaultGenTxGas,
			chainID,
			[]uint64{traderAcc.GetAccountNumber()},
			[]uint64{traderAcc.GetSequence()},
			trader.PrivKey,
		)

		_, result, err := app.Deliver(tx)
		if err != nil {
			// to aid debugging, add the stack trace to the comment field of the returned opMsg
			return simulation.NewOperationMsg(msg, false, fmt.Sprintf("%+v", err)), nil, err
		}
		return simulation.NewOperationMsg(msg, true, result.Log), nil, nil
	}
}

// SimulateMsgSwapForExactTokens generates a MsgSwapForExactTokens
func SimulateMsgSwapForExactTokens(ak types.AccountKeeper, k keeper.Keeper) simulation.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simulation.Account, chainID string,
	) (simulation.OperationMsg, []simulation.FutureOperation, error) {

		poolRecords := k.GetAllPools(ctx)
		r.Shuffle(len(poolRecords), func(i, j int) {
			poolRecords[i], poolRecords[j] = poolRecords[j], poolRecords[i]
		})

		// Find an account-pool pair for which trade is possible
		trader, poolRecord, found := findValidAccountPoolRecordPair(accs, poolRecords, func(acc simulation.Account, poolRecord types.PoolRecord) bool {
			traderAcc := ak.GetAccount(ctx, acc.Address)
			balanceTokenA := traderAcc.GetCoins().AmountOf(poolRecord.ReservesA.Denom)
			balanceTokenB := traderAcc.GetCoins().AmountOf(poolRecord.ReservesB.Denom)
			if !balanceTokenA.IsPositive() || !balanceTokenB.IsPositive() {
				return false
			}
			return true
		})
		if !found {
			return simulation.NewOperationMsgBasic(types.ModuleName, "no-operation (no valid pool record and trader)", "", false, nil), nil, nil
		}

		// Select input token
		randInt, err := RandInt(r, sdk.OneInt(), sdk.NewInt(9))
		if err != nil {
			panic(err)
		}
		inputToken := poolRecord.ReservesA
		outputToken := poolRecord.ReservesB
		if randInt.Int64()%2 == 0 {
			inputToken = poolRecord.ReservesB
			outputToken = poolRecord.ReservesA
		}

		// Select entity (trader account or pool) with smaller token amount
		traderAcc := ak.GetAccount(ctx, trader.Address)
		maxTradeAmount := inputToken.Amount
		if traderAcc.GetCoins().AmountOf(inputToken.Denom).LT(inputToken.Amount) {
			maxTradeAmount = traderAcc.GetCoins().AmountOf(inputToken.Denom)
		}

		// Expected input token is between 2-10% of the max trade amount
		percentage, err := RandIntInclusive(r, sdk.NewInt(2), sdk.NewInt(10))
		if err != nil {
			panic(err)
		}
		tradeAmount := maxTradeAmount.Mul(percentage).Quo(sdk.NewInt(100))
		expectedInputToken := sdk.NewCoin(inputToken.Denom, tradeAmount)

		// Calculate exact output coin
		globalSwapFee := k.GetSwapFee(ctx)
		tradeAmountAfterFee := expectedInputToken.Amount.ToDec().Mul(sdk.OneDec().Sub(globalSwapFee)).TruncateInt()

		var outputAmt big.Int
		outputAmt.Mul(outputToken.Amount.BigInt(), tradeAmountAfterFee.BigInt())
		outputAmt.Quo(&outputAmt, inputToken.Amount.Add(tradeAmountAfterFee).BigInt())
		outputTokenAmount := sdk.NewIntFromBigInt(&outputAmt)
		exactOutputToken := sdk.NewCoin(outputToken.Denom, outputTokenAmount)

		// Get random slippage amount between 50-100%
		slippageRaw, err := RandIntInclusive(r, sdk.NewInt(50), sdk.NewInt(99))
		if err != nil {
			panic(err)
		}
		slippage := slippageRaw.ToDec().Quo(sdk.NewDec(100))

		// Generate random deadline
		blockTime := ctx.BlockHeader().Time
		deadline := genRandDeadline(r, blockTime)

		// Construct MsgSwapForExactTokens
		msg := types.NewMsgSwapForExactTokens(traderAcc.GetAddress(), expectedInputToken, exactOutputToken, slippage, deadline)
		err = msg.ValidateBasic()
		if err != nil {
			return noOpMsg, nil, nil
		}

		tx := helpers.GenTx(
			[]sdk.Msg{msg},
			sdk.NewCoins(),
			helpers.DefaultGenTxGas,
			chainID,
			[]uint64{traderAcc.GetAccountNumber()},
			[]uint64{traderAcc.GetSequence()},
			trader.PrivKey,
		)

		_, result, err := app.Deliver(tx)
		if err != nil {
			// to aid debugging, add the stack trace to the comment field of the returned opMsg
			return simulation.NewOperationMsg(msg, false, fmt.Sprintf("%+v", err)), nil, err
		}
		return simulation.NewOperationMsg(msg, true, result.Log), nil, nil
	}
}

// From a set of coins return a coin of the specified denom with 1-10% of the total amount
func randCoinFromCoins(r *rand.Rand, coins sdk.Coins, denom string) sdk.Coin {
	percentOfBalance, err := RandIntInclusive(r, sdk.OneInt(), sdk.NewInt(10))
	if err != nil {
		panic(err)
	}
	balance := coins.AmountOf(denom)
	amtToken := balance.Mul(percentOfBalance).Quo(sdk.NewInt(100))
	return sdk.NewCoin(denom, amtToken)
}

func validateDepositor(ctx sdk.Context, k keeper.Keeper, allowedPool types.AllowedPool,
	depositor authexported.Account, blockTime time.Time) error {
	depositorCoins := depositor.SpendableCoins(blockTime)
	tokenABalance := depositorCoins.AmountOf(allowedPool.TokenA)
	tokenBBalance := depositorCoins.AmountOf(allowedPool.TokenB)

	oneThousand := sdk.NewInt(1000)
	if tokenABalance.LT(oneThousand) || tokenBBalance.LT(oneThousand) {
		return errorNotEnoughCoins
	}

	return nil
}

// findValidAccountAllowedPoolPair finds an account for which the callback func returns true
func findValidAccountAllowedPoolPair(accounts []simulation.Account, pools types.AllowedPools,
	cb func(simulation.Account, types.AllowedPool) bool) (simulation.Account, types.AllowedPool, bool) {
	for _, pool := range pools {
		for _, acc := range accounts {
			if isValid := cb(acc, pool); isValid {
				return acc, pool, true
			}
		}
	}
	return simulation.Account{}, types.AllowedPool{}, false
}

// findValidAccountPoolRecordPair finds an account for which the callback func returns true
func findValidAccountPoolRecordPair(accounts []simulation.Account, pools types.PoolRecords,
	cb func(simulation.Account, types.PoolRecord) bool) (simulation.Account, types.PoolRecord, bool) {
	for _, pool := range pools {
		for _, acc := range accounts {
			if isValid := cb(acc, pool); isValid {
				return acc, pool, true
			}
		}
	}
	return simulation.Account{}, types.PoolRecord{}, false
}

func genRandDeadline(r *rand.Rand, blockTime time.Time) int64 {
	// Set up deadline
	durationNanoseconds, err := RandIntInclusive(r,
		sdk.NewInt((time.Second * 10).Nanoseconds()), // ten seconds
		sdk.NewInt((time.Hour * 24).Nanoseconds()),   // one day
	)
	if err != nil {
		panic(err)
	}
	extraTime := time.Duration(durationNanoseconds.Int64())
	return blockTime.Add(extraTime).Unix()
}

// RandIntInclusive randomly generates an sdk.Int in the range [inclusiveMin, inclusiveMax]. It works for negative and positive integers.
func RandIntInclusive(r *rand.Rand, inclusiveMin, inclusiveMax sdk.Int) (sdk.Int, error) {
	if inclusiveMin.GT(inclusiveMax) {
		return sdk.Int{}, fmt.Errorf("min larger than max")
	}
	return RandInt(r, inclusiveMin, inclusiveMax.Add(sdk.OneInt()))
}

// RandInt randomly generates an sdk.Int in the range [inclusiveMin, exclusiveMax). It works for negative and positive integers.
func RandInt(r *rand.Rand, inclusiveMin, exclusiveMax sdk.Int) (sdk.Int, error) {
	// validate input
	if inclusiveMin.GTE(exclusiveMax) {
		return sdk.Int{}, fmt.Errorf("min larger or equal to max")
	}
	// shift the range to start at 0
	shiftedRange := exclusiveMax.Sub(inclusiveMin) // should always be positive given the check above
	// randomly pick from the shifted range
	shiftedRandInt := sdk.NewIntFromBigInt(new(big.Int).Rand(r, shiftedRange.BigInt()))
	// shift back to the original range
	return shiftedRandInt.Add(inclusiveMin), nil
}
