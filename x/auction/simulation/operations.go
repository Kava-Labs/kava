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
	"github.com/cosmos/cosmos-sdk/x/auth"
	authexported "github.com/cosmos/cosmos-sdk/x/auth/exported"
	"github.com/cosmos/cosmos-sdk/x/simulation"

	appparams "github.com/kava-labs/kava/app/params"
	"github.com/kava-labs/kava/x/auction/keeper"
	"github.com/kava-labs/kava/x/auction/types"
)

var (
	noOpMsg             = simulation.NoOpMsg(types.ModuleName)
	ErrorNotEnoughCoins = errors.New("account doesn't have enough coins")
)

// Simulation operation weights constants
const (
	OpWeightMsgPlaceBid = "op_weight_msg_place_bid"
)

// WeightedOperations returns all the operations from the module with their respective weights
func WeightedOperations(
	appParams simulation.AppParams, cdc *codec.Codec, ak auth.AccountKeeper, k keeper.Keeper,
) simulation.WeightedOperations {
	var weightMsgPlaceBid int

	appParams.GetOrGenerate(cdc, OpWeightMsgPlaceBid, &weightMsgPlaceBid, nil,
		func(_ *rand.Rand) {
			weightMsgPlaceBid = appparams.DefaultWeightMsgPlaceBid
		},
	)

	return simulation.WeightedOperations{
		simulation.NewWeightedOperation(
			weightMsgPlaceBid,
			SimulateMsgPlaceBid(ak, k),
		),
	}
}

// SimulateMsgPlaceBid returns a function that runs a random state change on the module keeper.
// There's two error paths
// - return a OpMessage, but nil error - this will log a message but keep running the simulation
// - return an error - this will stop the simulation
func SimulateMsgPlaceBid(ak auth.AccountKeeper, keeper keeper.Keeper) simulation.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simulation.Account, chainID string,
	) (simulation.OperationMsg, []simulation.FutureOperation, error) {
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

		// search through auctions and an accounts to find a pair where a bid can be placed (ie account has enough coins to place bid on auction)
		blockTime := ctx.BlockHeader().Time
		bidder, openAuction, found := findValidAccountAuctionPair(accs, openAuctions, func(acc simulation.Account, auc types.Auction) bool {
			account := ak.GetAccount(ctx, acc.Address)
			_, err := generateBidAmount(r, auc, account, blockTime)
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

		bidderAcc := ak.GetAccount(ctx, bidder.Address)
		if bidderAcc == nil {
			return simulation.NoOpMsg(types.ModuleName), nil, nil
		}

		// pick a bid amount for the chosen auction and bidder
		amount, err := generateBidAmount(r, openAuction, bidderAcc, blockTime)
		if err != nil {
			return simulation.NoOpMsg(types.ModuleName), nil, err
		}

		// create a msg
		msg := types.NewMsgPlaceBid(openAuction.GetID(), bidder.Address, amount)

		spendable := bidderAcc.SpendableCoins(ctx.BlockTime())
		fees, err := simulation.RandomFees(r, ctx, spendable)
		if err != nil {
			return simulation.NoOpMsg(types.ModuleName), nil, err
		}

		tx := helpers.GenTx(
			[]sdk.Msg{msg},
			fees,
			helpers.DefaultGenTxGas,
			chainID,
			[]uint64{bidderAcc.GetAccountNumber()},
			[]uint64{bidderAcc.GetSequence()},
			bidder.PrivKey,
		)

		_, result, err := app.Deliver(tx)
		if err != nil {
			return simulation.NoOpMsg(types.ModuleName), nil, err
		}

		// Return an operationMsg indicating whether the msg was submitted successfully
		// Using result.Log as the comment field as it contains any error message emitted by the keeper
		return simulation.NewOperationMsg(msg, true, result.Log), nil, nil
	}
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
		}
		amt, err := RandIntInclusive(r, a.Bid.Amount, sdk.MinInt(bidderBalance.AmountOf(a.Bid.Denom), a.MaxBid.Amount))
		if err != nil {
			panic(err)
		}
		// pick the MaxBid amount more frequently to increase chance auctions phase get into reverse phase
		if r.Intn(10) == 0 { // 10%
			amt = a.MaxBid.Amount
		}
		return sdk.NewCoin(a.Bid.Denom, amt), nil // stable coin

	default:
		return sdk.Coin{}, fmt.Errorf("unknown auction type")
	}
}

// findValidAccountAuctionPair finds an auction and account for which the callback func returns true
func findValidAccountAuctionPair(accounts []simulation.Account, auctions types.Auctions, cb func(simulation.Account, types.Auction) bool) (simulation.Account, types.Auction, bool) {
	for _, auc := range auctions {
		for _, acc := range accounts {
			if isValid := cb(acc, auc); isValid {
				return acc, auc, true
			}
		}
	}
	return simulation.Account{}, nil, false
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
