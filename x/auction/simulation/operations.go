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
	errorNotEnoughCoins  = errors.New("account doesn't have enough coins")
	errorCantReceiveBids = errors.New("auction can't receive bids (lot = 0 in reverse auction)")
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
		params := keeper.GetParams(ctx)
		bidder, openAuction, found := findValidAccountAuctionPair(accs, openAuctions, func(acc simulation.Account, auc types.Auction) bool {
			account := ak.GetAccount(ctx, acc.Address)
			_, err := generateBidAmount(r, params, auc, account, blockTime)
			if err == errorNotEnoughCoins || err == errorCantReceiveBids {
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
			return simulation.NoOpMsg(types.ModuleName), nil, fmt.Errorf("couldn't find account %s", bidder.Address)
		}

		// pick a bid amount for the chosen auction and bidder
		amount, err := generateBidAmount(r, params, openAuction, bidderAcc, blockTime)
		if err != nil { // shouldn't happen given the checks above
			return simulation.NoOpMsg(types.ModuleName), nil, err
		}

		// create and deliver a tx
		msg := types.NewMsgPlaceBid(openAuction.GetID(), bidder.Address, amount)

		tx := helpers.GenTx(
			[]sdk.Msg{msg},
			sdk.NewCoins(), // TODO pick a random amount fees
			helpers.DefaultGenTxGas,
			chainID,
			[]uint64{bidderAcc.GetAccountNumber()},
			[]uint64{bidderAcc.GetSequence()},
			bidder.PrivKey,
		)

		_, _, err = app.Deliver(tx)
		if err != nil {
			// to aid debugging, add the stack trace to the comment field of the returned opMsg
			return simulation.NewOperationMsg(msg, false, fmt.Sprintf("%+v", err)), nil, err
		}
		return simulation.NewOperationMsg(msg, true, ""), nil, nil
	}
}

func generateBidAmount(
	r *rand.Rand, params types.Params, auc types.Auction,
	bidder authexported.Account, blockTime time.Time) (sdk.Coin, error) {
	bidderBalance := bidder.SpendableCoins(blockTime)

	switch a := auc.(type) {

	case types.DebtAuction:
		// Check bidder has enough (stable coin) to pay in
		if bidderBalance.AmountOf(a.Bid.Denom).LT(a.Bid.Amount) { // stable coin
			return sdk.Coin{}, errorNotEnoughCoins
		}
		// Check auction can still receive new bids
		if a.Lot.Amount.Equal(sdk.ZeroInt()) {
			return sdk.Coin{}, errorCantReceiveBids
		}
		// Generate a new lot amount (gov coin)
		maxNewLotAmt := a.Lot.Amount.Sub( // new lot must be some % less than old lot, and at least 1 smaller to avoid replacing an old bid at no cost
			sdk.MaxInt(
				sdk.NewInt(1),
				sdk.NewDecFromInt(a.Lot.Amount).Mul(params.IncrementDebt).RoundInt(),
			),
		)
		amt, err := RandIntInclusive(r, sdk.ZeroInt(), maxNewLotAmt) // maxNewLotAmt shouldn't be < 0 given the check above
		if err != nil {
			panic(err)
		}
		return sdk.NewCoin(a.Lot.Denom, amt), nil // gov coin

	case types.SurplusAuction:
		// Check the bidder has enough (gov coin) to pay in
		minNewBidAmt := a.Bid.Amount.Add( // new bids must be some % greater than old bid, and at least 1 larger to avoid replacing an old bid at no cost
			sdk.MaxInt(
				sdk.NewInt(1),
				sdk.NewDecFromInt(a.Bid.Amount).Mul(params.IncrementSurplus).RoundInt(),
			),
		)
		if bidderBalance.AmountOf(a.Bid.Denom).LT(minNewBidAmt) { // gov coin
			return sdk.Coin{}, errorNotEnoughCoins
		}
		// Generate a new bid amount (gov coin)
		amt, err := RandIntInclusive(r, minNewBidAmt, bidderBalance.AmountOf(a.Bid.Denom))
		if err != nil {
			panic(err)
		}
		return sdk.NewCoin(a.Bid.Denom, amt), nil // gov coin

	case types.CollateralAuction:
		// Check the bidder has enough (stable coin) to pay in
		minNewBidAmt := a.Bid.Amount.Add( // new bids must be some % greater than old bid, and at least 1 larger to avoid replacing an old bid at no cost
			sdk.MaxInt(
				sdk.NewInt(1),
				sdk.NewDecFromInt(a.Bid.Amount).Mul(params.IncrementCollateral).RoundInt(),
			),
		)
		minNewBidAmt = sdk.MinInt(minNewBidAmt, a.MaxBid.Amount) // allow new bids to hit MaxBid even though it may be less than the increment %
		if bidderBalance.AmountOf(a.Bid.Denom).LT(minNewBidAmt) {
			return sdk.Coin{}, errorNotEnoughCoins
		}
		// Check auction can still receive new bids
		if a.IsReversePhase() && a.Lot.Amount.Equal(sdk.ZeroInt()) {
			return sdk.Coin{}, errorCantReceiveBids
		}
		// Generate a new bid amount (collateral coin in reverse phase)
		if a.IsReversePhase() {
			maxNewLotAmt := a.Lot.Amount.Sub( // new lot must be some % less than old lot, and at least 1 smaller to avoid replacing an old bid at no cost
				sdk.MaxInt(
					sdk.NewInt(1),
					sdk.NewDecFromInt(a.Lot.Amount).Mul(params.IncrementCollateral).RoundInt(),
				),
			)
			amt, err := RandIntInclusive(r, sdk.ZeroInt(), maxNewLotAmt) // maxNewLotAmt shouldn't be < 0 given the check above
			if err != nil {
				panic(err)
			}
			return sdk.NewCoin(a.Lot.Denom, amt), nil // collateral coin

			// Generate a new bid amount (stable coin in forward phase)
		} else {
			amt, err := RandIntInclusive(r, minNewBidAmt, sdk.MinInt(bidderBalance.AmountOf(a.Bid.Denom), a.MaxBid.Amount))
			if err != nil {
				panic(err)
			}
			// when the bidder has enough coins, pick the MaxBid amount more frequently to increase chance auctions phase get into reverse phase
			if r.Intn(2) == 0 && bidderBalance.AmountOf(a.Bid.Denom).GTE(a.MaxBid.Amount) { // 50%
				amt = a.MaxBid.Amount
			}
			return sdk.NewCoin(a.Bid.Denom, amt), nil // stable coin
		}

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
