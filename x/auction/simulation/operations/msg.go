package operations

import (
	"fmt"
	"math/rand"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/simulation"

	"github.com/kava-labs/kava/x/auction"
	"github.com/kava-labs/kava/x/auction/keeper"
	"github.com/kava-labs/kava/x/auction/types"
)

var noOpMsg = simulation.NoOpMsg(auction.ModuleName)

func SimulateMsgPlaceBid(authKeeper auth.AccountKeeper, keeper keeper.Keeper) simulation.Operation {
	handler := auction.NewHandler(keeper)

	// There's two error paths
	// - return a OpMessage, but nil error - this will log a message but keep running the simulation
	// - return an error - this will stop the simulation ( I think)
	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simulation.Account) (
		simulation.OperationMsg, []simulation.FutureOperation, error) {

		// Generate a Msg
		// get open auctions
		openAuctions := types.Auctions{}
		keeper.IterateAuctions(ctx, func(a types.Auction) bool { // TODO optimize by using index rather than account objects
			openAuctions = append(openAuctions, a)
			return false
		})
		// randomly pick an auction to bid on
		if len(openAuctions) <= 0 { // protect r.Intn from panicing
			return noOpMsg, nil, nil // don't submit a message if there are no auctions
		}
		auction := openAuctions[r.Intn(len(openAuctions))]
		// randomly pick an amount to bid
		govDenom := "ukava"
		bidder := accs[0].Address // TODO don't panic!
		var amount sdk.Coin
		switch a := auction.(type) {
		case types.DebtAuction:
			balance := authKeeper.GetAccount(ctx, bidder).SpendableCoins(ctx.BlockHeader().Time).AmountOf(govDenom) // TODO use GetSpendableCoins ?
			if balance.LTE(a.Lot.Amount) {
				return noOpMsg, nil, nil // don't place bid if account doesn't have enough coins
			}
			amt, err := randBoundedInt(r, a.Lot.Amount, balance)
			if err != nil {
				return noOpMsg, nil, err
			}
			amount = sdk.NewCoin(govDenom, amt)
		case types.SurplusAuction:
			return noOpMsg, nil, fmt.Errorf("TODO")
		case types.CollateralAuction:
			return noOpMsg, nil, fmt.Errorf("TODO")
		default:
			return noOpMsg, nil, fmt.Errorf("unknown auction type %v", auction)
		}
		// generate msg
		msg := types.NewMsgPlaceBid(auction.GetID(), bidder, amount)
		// don't submit errors that fail ValidateBasic, use unit tests for testing ValidateBasic
		if err := msg.ValidateBasic(); err != nil {
			return noOpMsg, nil, fmt.Errorf("expected msg to pass ValidateBasic: %s", msg.GetSignBytes())
		}

		// submit the msg
		if ok := submitMsg(ctx, handler, msg); !ok {
			return noOpMsg, nil, fmt.Errorf("could not submit place bid msg")
		}
		return simulation.NewOperationMsg(msg, true, "placed bid on auction"), nil, nil // TODO what should go in comment field?
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

// Return a random sdk.Int in the range [lower, upper) // TODO
func randBoundedInt(r *rand.Rand, inclusiveLowerBound, exclusiveUpperBound sdk.Int) (sdk.Int, error) {
	shiftedAmount, err := simulation.RandPositiveInt(r, inclusiveLowerBound.Sub(exclusiveUpperBound)) // TODO doesn't like 1 for some reason
	if err != nil {
		return sdk.Int{}, err
	}
	return shiftedAmount.Add(exclusiveUpperBound), nil
}
