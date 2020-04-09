package operations

import (
	"fmt"
	"math"
	"math/rand"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/simulation"

	"github.com/kava-labs/kava/x/bep3"
	"github.com/kava-labs/kava/x/bep3/keeper"
	"github.com/kava-labs/kava/x/bep3/types"
)

var (
	noOpMsg = simulation.NoOpMsg(bep3.ModuleName)
)

// SimulateMsgCreateAtomicSwap generates a MsgCreateAtomicSwap with random values
func SimulateMsgCreateAtomicSwap(ak auth.AccountKeeper, k keeper.Keeper) simulation.Operation {
	handler := bep3.NewHandler(k)

	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simulation.Account) (
		simulation.OperationMsg, []simulation.FutureOperation, error) {

		sender := k.GetBnbDeputyAddress(ctx)
		recipient := simulation.RandomAcc(r, accs).Address

		recipientOtherChain := simulation.RandStringOfLength(r, 43)
		senderOtherChain := simulation.RandStringOfLength(r, 43)

		// Generate cryptographically strong pseudo-random number
		randomNumber, err := types.GenerateSecureRandomNumber()
		if err != nil {
			return noOpMsg, nil, err
		}

		// Must use current blocktime instead of 'now' since initial blocktime was randomly generated
		timestamp := ctx.BlockTime().Unix()

		randomNumberHash := types.CalculateRandomHash(randomNumber.Bytes(), timestamp)

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

		minLock := int(k.GetMinBlockLock(ctx))
		maxLock := int(k.GetMaxBlockLock(ctx))
		heightSpan := int64(r.Intn(maxLock-minLock) + minLock)
		crossChain := true

		msg := types.NewMsgCreateAtomicSwap(
			sender, recipient, recipientOtherChain, senderOtherChain, randomNumberHash,
			timestamp, coins, expectedIncome, heightSpan, crossChain)

		if err := msg.ValidateBasic(); err != nil {
			return noOpMsg, nil, fmt.Errorf("expected msg to pass ValidateBasic: %s", err)
		}

		// Submit msg
		ok := submitMsg(ctx, handler, msg)

		// TODO: Create FutureOperation claimOp/refundOp
		return simulation.NewOperationMsg(msg, ok, ""), nil, nil
	}
}

func submitMsg(ctx sdk.Context, handler sdk.Handler, msg sdk.Msg) (ok bool) {
	ctx, write := ctx.CacheContext()
	// TODO: I really want to get '.Log' out of handler(), but can't because it duplicates msg submission
	ok = handler(ctx, msg).IsOK()
	if ok {
		write()
	}
	return ok
}
