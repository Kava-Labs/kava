package operations

import (
	"fmt"
	"math/big"
	"math/rand"
	"time"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govsimops "github.com/cosmos/cosmos-sdk/x/gov/simulation/operations"
	"github.com/cosmos/cosmos-sdk/x/simulation"

	"github.com/kava-labs/kava/x/committee"
)

var (
	noOpMsg                = simulation.NoOpMsg(committee.ModuleName)
	proposalPassPercentage = 0.8
)

/*
each permission type has a ContentSimulator that create valid pubproposals
pick committee, pick a permission, run generator for permission to get pubproposal
param - write own, needs list of possible param changes
god - could use any existing ContentSimulator
text - use existing Text ContentSimulator

unlike gov, can't just send in a pubp generator and say submit a message. Might be no committees, or no committees with the right permissions.
could say, given committees, try and get someone to submit a valid proposal.
pick committee, pick permission, generate proposal valid for perission - could use existing funcs

really want to control amount each type of proposal is submitted

top level - control of likelihood of different proposal types
create new param change proposal generator - one that accepts a list of param generators with weights

*/

type PubProposalGenerator func(r *rand.Rand, ctx sdk.Context, perm committee.Permission, accs []simulation.Account) committee.PubProposal {

}

func SimulateMsgSubmitProposal(keeper committee.Keeper, pubProposalGenerator PubProposalGenerator) simulation.Operation {
	handler := committee.NewHandler(keeper)

	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simulation.Account) (
		simulation.OperationMsg, []simulation.FutureOperation, error) {

		// 1) submit proposal

		// pick a committee randomly
		var coms []committee.Committee
		keeper.IterateCommittees(ctx, func(com committee.Committee) bool {
			coms = append(coms, com)
			return false
		})
		randomCommittee := coms[r.Intn(len(coms))]

		// pick a committee with permissions that fullfi

		// generate a proposal and submit-proposal msg
		proposer := randomCommittee.Members[len(randomCommittee.Members)]
		perm := randomCommittee.Permissions[r.Intn(len(randomCommittee.Permissions))]
		getPubProposalGenerator(perm)
		pubProposal := pubProposalGenerator(r, ctx, perm, accs)
		msg := committee.NewMsgSubmitProposal(
			pubProposal,
			proposer,
			randomCommittee.ID,
		)
		if err := msg.ValidateBasic(); err != nil {
			return noOpMsg, nil, err
		}

		// submit the msg
		res := submitMsg(ctx, handler, msg)
		submitOpMsg := simulation.NewOperationMsg(msg, res.IsOK(), res.Log)
		if !res.IsOK() { // don't schedule votes if proposal failed
			return submitOpMsg, nil, nil
		}

		// 2) Schedule vote operations

		// get submitted proposal
		proposalID := committee.Uint64FromBytes(res.Data)
		proposal, found := keeper.GetProposal(ctx, proposalID)
		if !found {
			return noOpMsg, nil, fmt.Errorf("can't find proposal")
		}

		// pick the voters - addresses that will submit vote msgs
		// note: not all proposals will get enough votes (number of passing proposals determined by `proposalPassPercentage`)
		numMembers := int64(len(randomCommittee.Members))
		majority := randomCommittee.VoteThreshold.Mul(sdk.NewInt(numMembers).ToDec()).Ceil().TruncateInt64()
		numVoters := r.Int63n(majority) // in interval [0, majority)

		shouldPass := r.Float64() < proposalPassPercentage
		if shouldPass {
			numVoters = majority + r.Int63n(numMembers-majority+1) // in interval [majority, numMembers]
		}
		voters := randomCommittee.Members[:numVoters]

		// schedule vote operations
		var futureOps []simulation.FutureOperation
		for _, v := range voters {
			rt, err := RandomTime(r, ctx.BlockTime(), proposal.Deadline)
			if err != nil {
				return noOpMsg, nil, fmt.Errorf("random time generation failed: %w", err)
			}
			fop := simulation.FutureOperation{
				BlockTime: rt,
				Op:        SimulateMsgVote(keeper, v, proposal.ID),
			}
			futureOps = append(futureOps, fop)
		}

		return submitOpMsg, futureOps, nil
	}
}

func SimulateMsgVote(keeper committee.Keeper, voter sdk.AccAddress, proposalID uint64) simulation.Operation {
	handler := committee.NewHandler(keeper)

	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simulation.Account) (
		opMsg simulation.OperationMsg, fOps []simulation.FutureOperation, err error) {

		msg := committee.NewMsgVote(voter, proposalID)
		if msg.ValidateBasic() != nil {
			return noOpMsg, nil, fmt.Errorf("expected msg to pass ValidateBasic: %s", msg.GetSignBytes())
		}

		res := submitMsg(ctx, handler, msg)
		return simulation.NewOperationMsg(msg, res.IsOK(), res.Log), nil, nil
	}
}

func submitMsg(ctx sdk.Context, handler sdk.Handler, msg sdk.Msg) sdk.Result {
	ctx, write := ctx.CacheContext()
	result := handler(ctx, msg)
	if result.IsOK() {
		write()
	}
	return result
}

func RandomTime(r *rand.Rand, inclusiveMin, exclusiveMax time.Time) (time.Time, error) {
	if exclusiveMax.Before(inclusiveMin) {
		return time.Time{}, fmt.Errorf("max must be > min")
	}
	period := exclusiveMax.Sub(inclusiveMin)
	subPeriod, err := RandomPositiveDuration(r, 0, period)
	if err != nil {
		return err
	}
	return inclusiveMin.Add(subPeriod), nil
}

// TODO move to common location
func RandomPositiveDuration(r *rand.Rand, inclusiveMin, exclusiveMax time.Duration) (time.Duration, error) {
	min := int64(inclusiveMin)
	max := int64(exclusiveMax)
	if min < 0 || max < 0 {
		return 0, fmt.Errorf("min and max must be positive")
	}
	if min >= max {
		return 0, fmt.Errorf("max must be > min")
	}
	randPositiveInt64 := r.Int63n(max-min) + min
	return time.Duration(randPositiveInt64), nil
}

// // Return a function that runs a random state change on the module keeper.
// // There's two error paths
// // - return a OpMessage, but nil error - this will log a message but keep running the simulation
// // - return an error - this will stop the simulation
// func SimulateMsgPlaceBid(authKeeper auth.AccountKeeper, keeper auction.Keeper) simulation.Operation {
// 	handler := auction.NewHandler(keeper)

// 	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simulation.Account) (
// 		simulation.OperationMsg, []simulation.FutureOperation, error) {

// 		// get open auctions
// 		openAuctions := auction.Auctions{}
// 		keeper.IterateAuctions(ctx, func(a auction.Auction) bool {
// 			openAuctions = append(openAuctions, a)
// 			return false
// 		})

// 		// shuffle auctions slice so that bids are evenly distributed across auctions
// 		rand.Shuffle(len(openAuctions), func(i, j int) {
// 			openAuctions[i], openAuctions[j] = openAuctions[j], openAuctions[i]
// 		})
// 		// TODO do the same for accounts?
// 		var accounts []authexported.Account
// 		for _, acc := range accs {
// 			accounts = append(accounts, authKeeper.GetAccount(ctx, acc.Address))
// 		}

// 		// search through auctions and an accounts to find a pair where a bid can be placed (ie account has enough coins to place bid on auction)
// 		blockTime := ctx.BlockHeader().Time
// 		bidder, openAuction, found := findValidAccountAuctionPair(accounts, openAuctions, func(acc authexported.Account, auc auction.Auction) bool {
// 			_, err := generateBidAmount(r, auc, acc, blockTime)
// 			if err == ErrorNotEnoughCoins {
// 				return false // keep searching
// 			} else if err != nil {
// 				panic(err) // raise errors
// 			}
// 			return true // found valid pair
// 		})
// 		if !found {
// 			return simulation.NewOperationMsgBasic(auction.ModuleName, "no-operation (no valid auction and bidder)", "", false, nil), nil, nil
// 		}

// 		// pick a bid amount for the chosen auction and bidder
// 		amount, _ := generateBidAmount(r, openAuction, bidder, blockTime)

// 		// create a msg
// 		msg := auction.NewMsgPlaceBid(openAuction.GetID(), bidder.GetAddress(), amount)
// 		if err := msg.ValidateBasic(); err != nil { // don't submit errors that fail ValidateBasic, use unit tests for testing ValidateBasic
// 			return noOpMsg, nil, fmt.Errorf("expected msg to pass ValidateBasic: %s", msg.GetSignBytes())
// 		}

// 		// submit the msg
// 		result := submitMsg(ctx, handler, msg)
// 		// Return an operationMsg indicating whether the msg was submitted successfully
// 		// Using result.Log as the comment field as it contains any error message emitted by the keeper
// 		return simulation.NewOperationMsg(msg, result.IsOK(), result.Log), nil, nil
// 	}
// }

// func generateBidAmount(r *rand.Rand, auc auction.Auction, bidder authexported.Account, blockTime time.Time) (sdk.Coin, error) {
// 	bidderBalance := bidder.SpendableCoins(blockTime)

// 	switch a := auc.(type) {

// 	case auction.DebtAuction:
// 		if bidderBalance.AmountOf(a.Bid.Denom).LT(a.Bid.Amount) { // stable coin
// 			return sdk.Coin{}, ErrorNotEnoughCoins
// 		}
// 		amt, err := RandIntInclusive(r, sdk.ZeroInt(), a.Lot.Amount) // pick amount less than current lot amount // TODO min bid increments
// 		if err != nil {
// 			panic(err)
// 		}
// 		return sdk.NewCoin(a.Lot.Denom, amt), nil // gov coin

// 	case auction.SurplusAuction:
// 		if bidderBalance.AmountOf(a.Bid.Denom).LT(a.Bid.Amount) { // gov coin // TODO account for bid increments
// 			return sdk.Coin{}, ErrorNotEnoughCoins
// 		}
// 		amt, err := RandIntInclusive(r, a.Bid.Amount, bidderBalance.AmountOf(a.Bid.Denom))
// 		if err != nil {
// 			panic(err)
// 		}
// 		return sdk.NewCoin(a.Bid.Denom, amt), nil // gov coin

// 	case auction.CollateralAuction:
// 		if bidderBalance.AmountOf(a.Bid.Denom).LT(a.Bid.Amount) { // stable coin // TODO account for bid increments (in forward phase)
// 			return sdk.Coin{}, ErrorNotEnoughCoins
// 		}
// 		if a.IsReversePhase() {
// 			amt, err := RandIntInclusive(r, sdk.ZeroInt(), a.Lot.Amount) // pick amount less than current lot amount
// 			if err != nil {
// 				panic(err)
// 			}
// 			return sdk.NewCoin(a.Lot.Denom, amt), nil // collateral coin
// 		} else {
// 			amt, err := RandIntInclusive(r, a.Bid.Amount, sdk.MinInt(bidderBalance.AmountOf(a.Bid.Denom), a.MaxBid.Amount))
// 			if err != nil {
// 				panic(err)
// 			}
// 			// pick the MaxBid amount more frequently to increase chance auctions phase get into reverse phase
// 			if r.Intn(10) == 0 { // 10%
// 				amt = a.MaxBid.Amount
// 			}
// 			return sdk.NewCoin(a.Bid.Denom, amt), nil // stable coin
// 		}

// 	default:
// 		return sdk.Coin{}, fmt.Errorf("unknown auction type")
// 	}
// }

// // findValidAccountAuctionPair finds an auction and account for which the callback func returns true
// func findValidAccountAuctionPair(accounts []authexported.Account, auctions auction.Auctions, cb func(authexported.Account, auction.Auction) bool) (authexported.Account, auction.Auction, bool) {
// 	for _, auc := range auctions {
// 		for _, acc := range accounts {
// 			if isValid := cb(acc, auc); isValid {
// 				return acc, auc, true
// 			}

// 		}
// 	}
// 	return nil, nil, false
// }

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
