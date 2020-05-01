package simulation

import (
	"fmt"
	"math"
	"math/rand"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/simapp/helpers"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/simulation"

	appparams "github.com/kava-labs/kava/app/params"
	"github.com/kava-labs/kava/x/bep3/keeper"
	"github.com/kava-labs/kava/x/bep3/types"
)

var (
	noOpMsg = simulation.NoOpMsg(types.ModuleName)
)

// Simulation operation weights constants
const (
	OpWeightMsgCreateAtomicSwap = "op_weight_msg_create_atomic_swap"
)

// WeightedOperations returns all the operations from the module with their respective weights
func WeightedOperations(
	appParams simulation.AppParams, cdc *codec.Codec, ak types.AccountKeeper, k keeper.Keeper,
) simulation.WeightedOperations {
	var weightCreateAtomicSwap int

	appParams.GetOrGenerate(cdc, OpWeightMsgCreateAtomicSwap, &weightCreateAtomicSwap, nil,
		func(_ *rand.Rand) {
			weightCreateAtomicSwap = appparams.DefaultWeightMsgCreateAtomicSwap
		},
	)

	return simulation.WeightedOperations{
		simulation.NewWeightedOperation(
			weightCreateAtomicSwap,
			SimulateMsgCreateAtomicSwap(ak, k),
		),
	}
}

// SimulateMsgCreateAtomicSwap generates a MsgCreateAtomicSwap with random values
func SimulateMsgCreateAtomicSwap(ak types.AccountKeeper, k keeper.Keeper) simulation.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simulation.Account, chainID string,
	) (simulation.OperationMsg, []simulation.FutureOperation, error) {

		senderAddr := k.GetBnbDeputyAddress(ctx)

		sender, found := simulation.FindAccount(accs, senderAddr)
		if !found {
			return noOpMsg, nil, nil
		}

		recipient, _ := simulation.RandomAcc(r, accs)

		recipientOtherChain := simulation.RandStringOfLength(r, 43)
		senderOtherChain := simulation.RandStringOfLength(r, 43)

		// Generate cryptographically strong pseudo-random number
		randomNumber, err := simulation.RandPositiveInt(r, sdk.NewInt(math.MaxInt64))
		if err != nil {
			return noOpMsg, nil, err
		}
		// Must use current blocktime instead of 'now' since initial blocktime was randomly generated
		timestamp := ctx.BlockTime().Unix()
		randomNumberHash := types.CalculateRandomHash(randomNumber.BigInt().Bytes(), timestamp)

		// Randomly select an asset from supported assets
		assets, found := k.GetAssets(ctx)
		if !found {
			return noOpMsg, nil, fmt.Errorf("no supported assets found")
		}
		asset := assets[r.Intn(len(assets))]

		// Check that the sender has coins of this type
		senderAcc := ak.GetAccount(ctx, senderAddr)
		fees, err := simulation.RandomFees(r, ctx, senderAcc.SpendableCoins(ctx.BlockTime()))
		if err != nil {
			return simulation.NoOpMsg(types.ModuleName), nil, err
		}
		availableAmount := senderAcc.SpendableCoins(ctx.BlockTime()).Sub(fees).AmountOf(asset.Denom)
		// Get an amount of coins between 0.1 and 2% of total coins
		amount := availableAmount.Quo(sdk.NewInt(int64(simulation.RandIntBetween(r, 50, 1000))))
		if amount.IsZero() {
			return simulation.NewOperationMsgBasic(types.ModuleName, fmt.Sprintf("no-operation (all funds exhausted for asset %s)", asset.Denom), "", false, nil), nil, nil
		}
		coin := sdk.NewCoin(asset.Denom, amount)
		coins := sdk.NewCoins(coin)
		expectedIncome := coin.String()

		// We're assuming that sims are run with -NumBlocks=100
		heightSpan := int64(55)
		crossChain := true

		msg := types.NewMsgCreateAtomicSwap(
			senderAddr, recipient.Address, recipientOtherChain, senderOtherChain, randomNumberHash,
			timestamp, coins, expectedIncome, heightSpan, crossChain,
		)

		tx := helpers.GenTx(
			[]sdk.Msg{msg},
			fees,
			helpers.DefaultGenTxGas,
			chainID,
			[]uint64{senderAcc.GetAccountNumber()},
			[]uint64{senderAcc.GetSequence()},
			sender.PrivKey,
		)

		_, result, err := app.Deliver(tx)
		if err != nil {
			return simulation.NoOpMsg(types.ModuleName), nil, err
		}

		// If created, construct a MsgClaimAtomicSwap or MsgRefundAtomicSwap future operation
		var futureOp simulation.FutureOperation
		swapID := types.CalculateSwapID(msg.RandomNumberHash, msg.From, msg.SenderOtherChain)
		if r.Intn(100) < 50 {
			// Claim future operation
			executionBlock := ctx.BlockHeight() + (msg.HeightSpan / 2)
			futureOp = simulation.FutureOperation{
				BlockHeight: int(executionBlock),
				Op:          operationClaimAtomicSwap(ak, k, swapID, randomNumber.BigInt().Bytes()),
			}
		} else {
			// Refund future operation
			executionBlock := ctx.BlockHeight() + msg.HeightSpan
			futureOp = simulation.FutureOperation{
				BlockHeight: int(executionBlock),
				Op:          operationRefundAtomicSwap(ak, k, swapID),
			}
		}

		return simulation.NewOperationMsg(msg, true, result.Log), []simulation.FutureOperation{futureOp}, nil
	}
}

func operationClaimAtomicSwap(ak types.AccountKeeper, k keeper.Keeper, swapID []byte, randomNumber []byte) simulation.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simulation.Account, chainID string,
	) (simulation.OperationMsg, []simulation.FutureOperation, error) {
		simAccount, _ := simulation.RandomAcc(r, accs)
		acc := ak.GetAccount(ctx, simAccount.Address)

		msg := types.NewMsgClaimAtomicSwap(acc.GetAddress(), swapID, randomNumber)

		fees, err := simulation.RandomFees(r, ctx, acc.SpendableCoins(ctx.BlockTime()))
		if err != nil {
			return simulation.NoOpMsg(types.ModuleName), nil, err
		}

		tx := helpers.GenTx(
			[]sdk.Msg{msg},
			fees,
			helpers.DefaultGenTxGas,
			chainID,
			[]uint64{acc.GetAccountNumber()},
			[]uint64{acc.GetSequence()},
			simAccount.PrivKey,
		)

		_, result, err := app.Deliver(tx)
		if err != nil {
			return simulation.NoOpMsg(types.ModuleName), nil, err
		}

		return simulation.NewOperationMsg(msg, true, result.Log), nil, nil
	}
}

func operationRefundAtomicSwap(ak types.AccountKeeper, k keeper.Keeper, swapID []byte) simulation.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simulation.Account, chainID string,
	) (simulation.OperationMsg, []simulation.FutureOperation, error) {
		simAccount, _ := simulation.RandomAcc(r, accs)
		acc := ak.GetAccount(ctx, simAccount.Address)

		msg := types.NewMsgRefundAtomicSwap(acc.GetAddress(), swapID)

		fees, err := simulation.RandomFees(r, ctx, acc.SpendableCoins(ctx.BlockTime()))
		if err != nil {
			return simulation.NoOpMsg(types.ModuleName), nil, err
		}

		tx := helpers.GenTx(
			[]sdk.Msg{msg},
			fees,
			helpers.DefaultGenTxGas,
			chainID,
			[]uint64{acc.GetAccountNumber()},
			[]uint64{acc.GetSequence()},
			simAccount.PrivKey,
		)

		_, result, err := app.Deliver(tx)
		if err != nil {
			return simulation.NoOpMsg(types.ModuleName), nil, err
		}

		return simulation.NewOperationMsg(msg, true, result.Log), nil, nil
	}
}
