package simulation

import (
	"fmt"
	"math"
	"math/rand"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/simulation"

	"github.com/kava-labs/kava/x/bep3/keeper"
	"github.com/kava-labs/kava/x/bep3/types"
)

var (
	noOpMsg = simulation.NoOpMsg(types.ModuleName)
)

// SimulateMsgCreateAtomicSwap generates a MsgCreateAtomicSwap with random values
func SimulateMsgCreateAtomicSwap(ak auth.AccountKeeper, k keeper.Keeper) simulation.Operation {
	// handler := bep3.NewHandler(k)

	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simulation.Account) (
		simulation.OperationMsg, []simulation.FutureOperation, error) {

		sender := k.GetBnbDeputyAddress(ctx)
		recipient := simulation.RandomAcc(r, accs).Address

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
		availableAmount := ak.GetAccount(ctx, sender).GetCoins().AmountOf(asset.Denom)
		if !availableAmount.IsPositive() {
			return noOpMsg, nil, fmt.Errorf("available amount must be positive")
		}

		// Get a random amount of the available coins
		amount, err := simulation.RandPositiveInt(r, availableAmount)
		if err != nil {
			return noOpMsg, nil, err
		}

		// If we don't adjust the conversion factor, we'll be out of funds soon
		adjustedAmount := amount.Int64() / int64(math.Pow10(8))
		coin := sdk.NewInt64Coin(asset.Denom, adjustedAmount)
		coins := sdk.NewCoins(coin)
		expectedIncome := coin.String()

		// We're assuming that sims are run with -NumBlocks=100
		heightSpan := int64(55)
		crossChain := true

		msg := types.NewMsgCreateAtomicSwap(
			sender, recipient, recipientOtherChain, senderOtherChain, randomNumberHash,
			timestamp, coins, expectedIncome, heightSpan, crossChain)

		if err := msg.ValidateBasic(); err != nil {
			return noOpMsg, nil, fmt.Errorf("expected MsgCreateAtomicSwap to pass ValidateBasic: %s", err)
		}

		// Submit msg
		ok := submitMsg(ctx, handler, msg)

		// If created, construct a MsgClaimAtomicSwap or MsgRefundAtomicSwap future operation
		var futureOp simulation.FutureOperation
		if ok {
			swapID := types.CalculateSwapID(msg.RandomNumberHash, msg.From, msg.SenderOtherChain)
			acc := simulation.RandomAcc(r, accs)
			evenOdd := r.Intn(2) + 1
			if evenOdd%2 == 0 {
				// Claim future operation
				executionBlock := ctx.BlockHeight() + (msg.HeightSpan / 2)
				futureOp = loadClaimFutureOp(acc.Address, swapID, randomNumber.BigInt().Bytes(), executionBlock, handler)
			} else {
				// Refund future operation
				executionBlock := ctx.BlockHeight() + msg.HeightSpan
				futureOp = loadRefundFutureOp(acc.Address, swapID, executionBlock, handler)
			}
		}

		return simulation.NewOperationMsg(msg, ok, ""), []simulation.FutureOperation{futureOp}, nil
	}
}

func loadClaimFutureOp(sender sdk.AccAddress, swapID []byte, randomNumber []byte, height int64, handler sdk.Handler) simulation.FutureOperation {
	claimOp := func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simulation.Account) (
		simulation.OperationMsg, []simulation.FutureOperation, error) {

		// Build the refund msg and validate basic
		claimMsg := types.NewMsgClaimAtomicSwap(sender, swapID, randomNumber)
		if err := claimMsg.ValidateBasic(); err != nil {
			return noOpMsg, nil, fmt.Errorf("expected MsgClaimAtomicSwap to pass ValidateBasic: %s", err)
		}

		// Test msg submission at target block height
		ok := handler(ctx.WithBlockHeight(height), claimMsg).IsOK()
		return simulation.NewOperationMsg(claimMsg, ok, ""), nil, nil
	}

	return simulation.FutureOperation{
		BlockHeight: int(height),
		Op:          claimOp,
	}
}

func loadRefundFutureOp(sender sdk.AccAddress, swapID []byte, height int64, handler sdk.Handler) simulation.FutureOperation {
	refundOp := func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simulation.Account) (
		simulation.OperationMsg, []simulation.FutureOperation, error) {
		// Build the refund msg and validate basic
		refundMsg := types.NewMsgRefundAtomicSwap(sender, swapID)
		if err := refundMsg.ValidateBasic(); err != nil {
			return noOpMsg, nil, fmt.Errorf("expected MsgRefundAtomicSwap to pass ValidateBasic: %s", err)
		}

		// Test msg submission at target block height
		ok := handler(ctx.WithBlockHeight(height), refundMsg).IsOK()
		return simulation.NewOperationMsg(refundMsg, ok, ""), nil, nil
	}

	return simulation.FutureOperation{
		BlockHeight: int(height),
		Op:          refundOp,
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
