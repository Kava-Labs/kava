package simulation

import (
	"errors"
	"fmt"
	"math/big"
	"math/rand"
	"time"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authexported "github.com/cosmos/cosmos-sdk/x/auth/exported"
	"github.com/cosmos/cosmos-sdk/x/simulation"

	"github.com/kava-labs/kava/x/auction/keeper"
	"github.com/kava-labs/kava/x/auction/types"
)

var (
	noOpMsg             = simulation.NoOpMsg(types.ModuleName)
	ErrorNotEnoughCoins = errors.New("account doesn't have enough coins")
)

// Return a function that runs a random state change on the module keeper.
// There's two error paths
// - return a OpMessage, but nil error - this will log a message but keep running the simulation
// - return an error - this will stop the simulation
func SimulateMsgPlaceBid(authKeeper auth.AccountKeeper, keeper keeper.Keeper) simulation.Operation {
	handler := types.NewHandler(keeper)

	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simulation.Account) (
		simulation.OperationMsg, []simulation.FutureOperation, error) {

		// get open auctions
		openAuctions := types.Auctions{}
		keeper.IterateAuctions(ctx, func(a types.Auction) bool {
			openAuctions = append(openAuctions, a)
			return false
		})

		// shuffle auctions slice so that bids are evenly distributed across auctions
		r.Shuffle(len(openAuctions), func(i, j int) {
			openAuctions[i], openAuctions[j] = openAuctions[j], openAuctions[i]
		})
		// TODO do the same for accounts?
		var accounts []authexported.Account
		for _, acc := range accs {
			accounts = append(accounts, authKeeper.GetAccount(ctx, acc.Address))
		}

		// search through auctions and an accounts to find a pair where a bid can be placed (ie account has enough coins to place bid on auction)
		blockTime := ctx.BlockHeader().Time
		bidder, openAuction, found := findValidAccountAuctionPair(accounts, openAuctions, func(acc authexported.Account, auc types.Auction) bool {
			_, err := generateBidAmount(r, auc, acc, blockTime)
			if err == ErrorNotEnoughCoins {
				return false // keep searching
			} else if err != nil {
				panic(err) // raise errors
			}
			return true // found valid pair
		})
		if !found {
			return simulation.NewOperationMsgBasic(types.ModuleName, "no-operation (no valid auction and bidder)", "", false, nil), nil, nil
		}

		// pick a bid amount for the chosen auction and bidder
		amount, _ := generateBidAmount(r, openAuction, bidder, blockTime)

		// create a msg
		msg := types.NewMsgPlaceBid(openAuction.GetID(), bidder.GetAddress(), amount)
		if err := msg.ValidateBasic(); err != nil { // don't submit errors that fail ValidateBasic, use unit tests for testing ValidateBasic
			return noOpMsg, nil, fmt.Errorf("expected msg to pass ValidateBasic: %s", msg.GetSignBytes())
		}

		// submit the msg
		result := submitMsg(ctx, handler, msg)
		// Return an operationMsg indicating whether the msg was submitted successfully
		// Using result.Log as the comment field as it contains any error message emitted by the keeper
		return simulation.NewOperationMsg(msg, result.IsOK(), result.Log), nil, nil
	}
}

func submitMsg(ctx sdk.Context, handler sdk.Handler, msg sdk.Msg) sdk.Result {
	ctx, write := ctx.CacheContext()
	result, err := handler(ctx, msg)
	if err != nil {
		write()
	}
	return result
}

func generateBidAmount(r *rand.Rand, auc types.Auction, bidder authexported.Account, blockTime time.Time) (sdk.Coin, error) {
	bidderBalance := bidder.SpendableCoins(blockTime)

	switch a := auc.(type) {

	case types.DebtAuction:
		if bidderBalance.AmountOf(a.Bid.Denom).LT(a.Bid.Amount) { // stable coin
			return sdk.Coin{}, ErrorNotEnoughCoins
		}
		amt, err := RandIntInclusive(r, sdk.ZeroInt(), a.Lot.Amount) // pick amount less than current lot amount // TODO min bid increments
		if err != nil {
			panic(err)
		}
		return sdk.NewCoin(a.Lot.Denom, amt), nil // gov coin

	case types.SurplusAuction:
		if bidderBalance.AmountOf(a.Bid.Denom).LT(a.Bid.Amount) { // gov coin // TODO account for bid increments
			return sdk.Coin{}, ErrorNotEnoughCoins
		}
		amt, err := RandIntInclusive(r, a.Bid.Amount, bidderBalance.AmountOf(a.Bid.Denom))
		if err != nil {
			panic(err)
		}
		return sdk.NewCoin(a.Bid.Denom, amt), nil // gov coin

	case types.CollateralAuction:
		if bidderBalance.AmountOf(a.Bid.Denom).LT(a.Bid.Amount) { // stable coin // TODO account for bid increments (in forward phase)
			return sdk.Coin{}, ErrorNotEnoughCoins
		}
		if a.IsReversePhase() {
			amt, err := RandIntInclusive(r, sdk.ZeroInt(), a.Lot.Amount) // pick amount less than current lot amount
			if err != nil {
				panic(err)
			}
			return sdk.NewCoin(a.Lot.Denom, amt), nil // collateral coin
		} else {
			amt, err := RandIntInclusive(r, a.Bid.Amount, sdk.MinInt(bidderBalance.AmountOf(a.Bid.Denom), a.MaxBid.Amount))
			if err != nil {
				panic(err)
			}
			// pick the MaxBid amount more frequently to increase chance auctions phase get into reverse phase
			if r.Intn(10) == 0 { // 10%
				amt = a.MaxBid.Amount
			}
			return sdk.NewCoin(a.Bid.Denom, amt), nil // stable coin
		}

	default:
		return sdk.Coin{}, fmt.Errorf("unknown auction type")
	}
}

// findValidAccountAuctionPair finds an auction and account for which the callback func returns true
func findValidAccountAuctionPair(accounts []authexported.Account, auctions types.Auctions, cb func(authexported.Account, types.Auction) bool) (authexported.Account, types.Auction, bool) {
	for _, auc := range auctions {
		for _, acc := range accounts {
			if isValid := cb(acc, auc); isValid {
				return acc, auc, true
			}

		}
	}
	return nil, nil, false
}

// RandInt randomly generates an sdk.Int in the range [inclusiveMin, inclusiveMax]. It works for negative and positive integers.
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
