package operations

import (
	"fmt"
	"math/big"
	"math/rand"
	"time"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/simulation"

	"github.com/kava-labs/kava/x/committee"
)

var (
	noOpMsg                = simulation.NoOpMsg(committee.ModuleName)
	proposalPassPercentage = 0.8
)

func SimulateSubmittingVotingForProposal(keeper committee.Keeper, pubProposalSimulator PubProposalSimulator) simulation.Operation {
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
		if len(coms) < 1 {
			return simulation.NewOperationMsgBasic(committee.ModuleName, "no-operation (no committees)", "", false, nil), nil, nil
		}
		randomCommittee := coms[r.Intn(len(coms))]

		// pick a permission randomly
		if len(randomCommittee.Permissions) < 1 {
			return simulation.NewOperationMsgBasic(committee.ModuleName, "no-operation (committee has no permissions)", "", false, nil), nil, nil
		}
		perm := randomCommittee.Permissions[r.Intn(len(randomCommittee.Permissions))]

		// generate a proposal and submit-proposal msg
		proposer := randomCommittee.Members[r.Intn(len(randomCommittee.Members))] // won't panic as committees must have ≥ 1 members
		pubProposal := pubProposalSimulator(r, ctx, accs, perm)
		if pubProposal == nil {
			return simulation.NewOperationMsgBasic(committee.ModuleName, "no-operation (couldn't generate pubproposal)", "", false, nil), nil, nil
		}
		msg := committee.NewMsgSubmitProposal(
			pubProposal,
			proposer,
			randomCommittee.ID,
		)
		if err := msg.ValidateBasic(); err != nil {
			return noOpMsg, nil, fmt.Errorf("expected msg to pass ValidateBasic: %w, %s", err, msg.GetSignBytes()) // TODO formatting
		}

		// submit the msg
		response := submitMsg(ctx, handler, msg)
		submitOpMsg := simulation.NewOperationMsg(msg, response.IsOK(), response.Log)
		if !response.IsOK() { // don't schedule votes if proposal failed
			return submitOpMsg, nil, nil
		}

		// 2) Schedule vote operations

		// get submitted proposal
		proposalID := committee.Uint64FromBytes(response.Data)
		proposal, found := keeper.GetProposal(ctx, proposalID)
		if !found {
			return noOpMsg, nil, fmt.Errorf("can't find proposal with ID '%d'", proposalID)
		}

		// pick the voters (addresses that will submit vote msgs)
		// num voters determined by whether the proposal should pass or not
		numMembers := int64(len(randomCommittee.Members))
		majority := randomCommittee.VoteThreshold.Mul(sdk.NewInt(numMembers).ToDec()).Ceil().TruncateInt64()
		var numVoters int64 = 0
		// Catch the edge case where vote threshold is zero.
		// This can be interpreted as meaning a vote should pass immediately after submition even with no votes.
		// There is no numVoters that will make the proposal not pass, so just leave it at zero and let the proposal pass.
		if majority > 0 {
			numVoters = r.Int63n(majority) // in interval [0, majority)
		}

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

// TODO move random funcs to common location

func RandomTime(r *rand.Rand, inclusiveMin, exclusiveMax time.Time) (time.Time, error) {
	if exclusiveMax.Before(inclusiveMin) {
		return time.Time{}, fmt.Errorf("max must be > min")
	}
	period := exclusiveMax.Sub(inclusiveMin)
	subPeriod, err := RandomPositiveDuration(r, 0, period)
	if err != nil {
		return time.Time{}, err
	}
	return inclusiveMin.Add(subPeriod), nil
}

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
