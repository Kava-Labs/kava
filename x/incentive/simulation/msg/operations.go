package operations

import (
	"fmt"
	"math/rand"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authexported "github.com/cosmos/cosmos-sdk/x/auth/exported"
	"github.com/cosmos/cosmos-sdk/x/simulation"

	"github.com/kava-labs/kava/x/incentive"
	"github.com/kava-labs/kava/x/incentive/keeper"
	"github.com/kava-labs/kava/x/incentive/types"
)

var (
	noOpMsg = simulation.NoOpMsg(incentive.ModuleName)
)

// SimulateMsgClaimReward generates a MsgClaimReward
func SimulateMsgClaimReward(ak auth.AccountKeeper, k keeper.Keeper) simulation.Operation {
	// TODO: create CDPs

	handler := incentive.NewHandler(k)

	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simulation.Account) (
		simulation.OperationMsg, []simulation.FutureOperation, error) {

		var accounts []authexported.Account
		for _, acc := range accs {
			accounts = append(accounts, ak.GetAccount(ctx, acc.Address))
		}

		claimer := GetRandomClaimer(r, accounts).GetAddress()

		fmt.Println("claimer:", claimer)
		// timestamp := ctx.BlockTime().Unix()

		// Randomly select a reward's collateral type from rewards
		params := k.GetParams(ctx)
		if len(params.Rewards) == 0 {
			return noOpMsg, nil, fmt.Errorf("no rewards found in incentive module params")
		}
		reward := params.Rewards[r.Intn(len(params.Rewards))]

		fmt.Println("reward:", reward)

		// Check that the sender has coins of this type
		availableAmount := ak.GetAccount(ctx, claimer).GetCoins().AmountOf(reward.Denom)
		if !availableAmount.IsPositive() {
			return noOpMsg, nil, fmt.Errorf("claimer doesn't have available amount")
		}
		fmt.Println("availableAmount:", availableAmount)

		// Get a random amount of the available coins
		// TODO: need minimum CDP principal for this type
		// simulation.RandIntBetween()
		// amount, err := simulation.RandPositiveInt(r, availableAmount)
		// if err != nil {
		// 	return noOpMsg, nil, err
		// }

		msg := types.NewMsgClaimReward(claimer, reward.Denom)

		fmt.Println("msg:", msg)

		if err := msg.ValidateBasic(); err != nil {
			return noOpMsg, nil, fmt.Errorf("expected MsgClaimReward to pass ValidateBasic: %s", err)
		}

		// Submit msg
		ok := submitMsg(ctx, handler, msg)
		fmt.Println("ok:", ok)

		// var futureOp simulation.FutureOperation
		// if ok {
		// 	acc := simulation.RandomAcc(r, accs)
		// 	executionBlock := ctx.BlockHeight() +
		// 	futureOp = loadClaimFutureOp(acc.Address, denom, executionBlock, handler)
		// }

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
