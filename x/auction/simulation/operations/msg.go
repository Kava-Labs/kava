package operations

import (
	"fmt"
	"math/big"
	"math/rand"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authexported "github.com/cosmos/cosmos-sdk/x/auth/exported"
	"github.com/cosmos/cosmos-sdk/x/simulation"

	"github.com/kava-labs/kava/x/auction"
	"github.com/kava-labs/kava/x/auction/keeper"
	"github.com/kava-labs/kava/x/auction/types"
)

var (
	noOpMsg = simulation.NoOpMsg(auction.ModuleName)
)

func SimulateMsgPlaceBid(authKeeper auth.AccountKeeper, keeper keeper.Keeper) simulation.Operation {
	handler := auction.NewHandler(keeper)

	// There's two error paths
	// - return a OpMessage, but nil error - this will log a message but keep running the simulation
	// - return an error - this will stop the simulation ( I think)
	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simulation.Account) (
		simulation.OperationMsg, []simulation.FutureOperation, error) {

		// get open auctions
		openAuctions := types.Auctions{}
		keeper.IterateAuctions(ctx, func(a types.Auction) bool { // TODO optimize by using index rather than account objects
			openAuctions = append(openAuctions, a)
			return false
		})

		// randomly pick an auction to bid on
		if len(openAuctions) <= 0 { // protect r.Intn from panicing
			return noOpMsg, nil, fmt.Errorf("TODO no auctions") // don't submit a message if there are no auctions
		}
		auction := openAuctions[r.Intn(len(openAuctions))]

		// randomly pick bidder and amount to bid
		bidder := authKeeper.GetAccount(ctx, accs[0].Address) // TODO don't panic!
		amount, err := generateBidAmount(r, ctx, auction, bidder)
		if err != nil {
			// TODO check for not enough coins error and pick new bidder
			return noOpMsg, nil, err
		}

		// generate msg
		msg := types.NewMsgPlaceBid(auction.GetID(), bidder.GetAddress(), amount)
		if err := msg.ValidateBasic(); err != nil { // don't submit errors that fail ValidateBasic, use unit tests for testing ValidateBasic
			return noOpMsg, nil, fmt.Errorf("expected msg to pass ValidateBasic: %s", msg.GetSignBytes())
		}

		// submit the msg
		if ok := submitMsg(ctx, handler, msg); !ok {
			return noOpMsg, nil, fmt.Errorf("could not submit place bid msg")
		}
		fmt.Println("bid sumbitted!")                                                   // FIXME
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

func generateBidAmount(r *rand.Rand, ctx sdk.Context, auction types.Auction, bidder authexported.Account) (sdk.Coin, error) {
	balance := bidder.SpendableCoins(ctx.BlockHeader().Time)

	switch a := auction.(type) {

	case types.DebtAuction:
		if balance.AmountOf(a.Bid.Denom).LT(a.Bid.Amount) { // stable coin
			return sdk.Coin{}, fmt.Errorf("account doesn't have enough coins") // don't place bid if account doesn't have enough coins
		}
		amt, err := RandInt(r, sdk.ZeroInt(), a.Lot.Amount) // pick amount less than current lot amount // TODO min bid increments
		if err != nil {
			return sdk.Coin{}, err
		}
		return sdk.NewCoin(a.Lot.Denom, amt), nil // gov coin

	case types.SurplusAuction:
		if balance.AmountOf(a.Bid.Denom).LT(a.Bid.Amount) { // gov coin // TODO account for bid increments
			return sdk.Coin{}, fmt.Errorf("account doesn't have enough coins") // don't place bid if account doesn't have enough coins
		}
		amt, _ := RandInt(r, a.Bid.Amount, balance.AmountOf(a.Bid.Denom))
		return sdk.NewCoin(a.Bid.Denom, amt), nil // gov coin

	case types.CollateralAuction:
		if balance.AmountOf(a.Bid.Denom).LT(a.Bid.Amount) { // stable coin // TODO account for bid increments (in forward phase)
			return sdk.Coin{}, fmt.Errorf("account doesn't have enough coins") // don't place bid if account doesn't have enough coins
		}
		if a.IsReversePhase() {
			amt, err := RandInt(r, sdk.ZeroInt(), a.Lot.Amount) // pick amount less than current lot amount
			if err != nil {
				return sdk.Coin{}, err
			}
			return sdk.NewCoin(a.Lot.Denom, amt), nil // collateral coin
		} else {
			amt, _ := RandInt(r, a.Bid.Amount, sdk.MinInt(balance.AmountOf(a.Bid.Denom), a.MaxBid.Amount))
			// TODO pick the MaxBid amount often to flip the auction phase
			return sdk.NewCoin(a.Bid.Denom, amt), nil // stable coin
		}

	default:
		return sdk.Coin{}, fmt.Errorf("unknown auction type")
	}
}

// TODO change to inclusive bounds?
// RandInt randomly generates an sdk.Int in the range [inclusiveMin, exclusiveMax). It works for negative and positive integers.
func RandInt(r *rand.Rand, inclusiveMin, exclusiveMax sdk.Int) (sdk.Int, error) {
	// validate input
	if inclusiveMin.GTE(exclusiveMax) {
		return sdk.Int{}, fmt.Errorf("invalid bounds")
	}
	// shift the range to start at 0
	shiftedRange := exclusiveMax.Sub(inclusiveMin) // should always be positive given the check above
	// randomly pick from the shifted range
	shiftedRandInt := sdk.NewIntFromBigInt(new(big.Int).Rand(r, shiftedRange.BigInt()))
	// shift back to the original range
	return shiftedRandInt.Add(inclusiveMin), nil
}
