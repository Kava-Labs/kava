package simulation

import (
	"fmt"
	"math/big"
	"math/rand"
	"time"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/simapp/helpers"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authexported "github.com/cosmos/cosmos-sdk/x/auth/exported"
	"github.com/cosmos/cosmos-sdk/x/simulation"
	"github.com/tendermint/tendermint/crypto"

	"github.com/kava-labs/kava/x/committee/keeper"
	"github.com/kava-labs/kava/x/committee/types"
)

var (
	proposalPassPercentage = 0.9
)

type AccountKeeper interface {
	GetAccount(sdk.Context, sdk.AccAddress) authexported.Account
}

// WeightedOperations returns all the operations from the module with their respective weights
func WeightedOperations(appParams simulation.AppParams, cdc *codec.Codec, ak AccountKeeper,
	k keeper.Keeper, wContents []simulation.WeightedProposalContent) simulation.WeightedOperations {
	/*
		create a list of weighted operations
		- submit proposal

		gov creates a different op for every content simulator

		calling the submitproposal op creates future ops that send votes

		gov: pick weights of proposals, submit according to those weights
		com: available proposals to submit depends on proposals available and committees' permissions (former static, later not)
		solution: only export one wop, it tries it's best to pick proposals according to their weights, but accounts for available committes.
	*/
	// create the equivalent of ContenSimulatorFns but allow permissions to be passed in
	// - text content - just convert func call type
	// - community pool spend -
	// respect the content weights

	// TODO create wops for each pubproposal simulator (we'll have some custom ones)
	// for example a paramchange pubproposal generator that only changes interesing cdp params

	var wops simulation.WeightedOperations

	for _, wContent := range wContents {
		wContent := wContent // pin variable
		if wContent.AppParamsKey == OpWeightSubmitCommitteeChangeProposal {
			// don't include committee change/delete proposals as they're not enabled for submission to committees
			continue
		}
		var weight int
		appParams.GetOrGenerate(cdc, wContent.AppParamsKey, &weight, nil,
			func(_ *rand.Rand) { weight = wContent.DefaultWeight }) // we might want different weights

		wops = append(
			wops,
			simulation.NewWeightedOperation(
				weight,
				SimulateMsgSubmitProposal(ak, k, wContent.ContentSimulatorFn),
			),
		)
	}
	return wops
}

// proposal type -> []permissiontypes

/*
Original gov sets the propabilities various proposals are run, then picks the type of proposal accordingly, and puts it into a submitProposal msg.
This model doesn't work in committees as they have permissions - some proposal types will not be able to be submitted as there may currently be no committees with permissions for it.

The simulator picks an operation with probability = op_weight / total_op_weights

Solutions
- ignore the problem, pick a proposal according to the weights, try and submit it anyway. - this will lead to a lot of failures, wasted time, and will result in skewed probabilities
- ignore the weights, pick a committee at random, then randomly pick a proposal that fits - no wasted time, but no control over likelihood of various proposals

*/

/*
   - pick random committee
   - generate a pubproposal (sending in the permissions)
   - genrate and deliver the msg
   - if successful schedule future operations

   - pick a random proposal type (according to weights)
   - find a committee that allows this type
   - generate proposal (pass in permissions)
   - ...
*/

// This func tries to find an accepting committee (ignoring the special one), and falls back to the special one
// for each submit proposal it also generates voting future ops
// if the special committee isn't there (like when using an existing genesis) just noop instead of falling back (same as naive solution)
func SimulateMsgSubmitProposal(ak AccountKeeper, k keeper.Keeper, contentSim simulation.ContentSimulatorFn) simulation.Operation {
	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simulation.Account, chainID string) (simulation.OperationMsg, []simulation.FutureOperation, error) {

		// 1) submit proposal

		committees := k.GetCommittees(ctx)

		// shuffle committees to ensure proposals are distributed across them evenly
		r.Shuffle(len(committees), func(i, j int) {
			committees[i], committees[j] = committees[j], committees[i]
		})

		// move fallback committee to the end
		for i, c := range committees {
			if c.ID == FallbackCommitteeID {
				// switch places with last element
				committees[i], committees[len(committees)-1] = committees[len(committees)-1], committees[i]
			}
		}

		pp := types.PubProposal(contentSim(r, ctx, accs))
		var selectedCommittee types.Committee
		var found bool
		for _, c := range committees {
			if c.HasPermissionsFor(pp) {
				selectedCommittee = c
				found = true
				break
			}
		}
		if !found {
			// This should only happen if not using the generated genesis state
			return simulation.NewOperationMsgBasic(types.ModuleName, "no-operation (no committee has permissions for proposal)", "", false, nil), nil, nil
		}

		proposer := selectedCommittee.Members[r.Intn(len(selectedCommittee.Members))] // won't panic as committees must have ≥ 1 members
		msg := types.NewMsgSubmitProposal(
			pp,
			proposer,
			selectedCommittee.ID,
		)

		account := ak.GetAccount(ctx, proposer)
		fees, err := simulation.RandomFees(r, ctx, account.SpendableCoins(ctx.BlockTime()))
		if err != nil {
			return simulation.NoOpMsg(types.ModuleName), nil, err
		}

		var proposerPrivKey crypto.PrivKey
		for _, a := range accs {
			if a.Address.Equals(proposer) {
				proposerPrivKey = a.PrivKey
				break
			}
		}
		tx := helpers.GenTx(
			[]sdk.Msg{msg},
			fees,
			helpers.DefaultGenTxGas,
			chainID,
			[]uint64{account.GetAccountNumber()},
			[]uint64{account.GetSequence()},
			proposerPrivKey,
		)
		// submit tx
		_, result, err := app.Deliver(tx)
		if err != nil {
			// to aid debugging, add the stack trace to the comment field of the returned opMsg
			return simulation.NewOperationMsg(msg, false, fmt.Sprintf("%+v", err)), nil, err
		}
		// to aid debugging, add the result log to the comment field
		submitOpMsg := simulation.NewOperationMsg(msg, true, result.Log)

		// 2) Schedule vote operations

		// get submitted proposal
		proposalID := types.Uint64FromBytes(result.Data)
		proposal, found := k.GetProposal(ctx, proposalID)
		if !found {
			return simulation.NoOpMsg(types.ModuleName), nil, fmt.Errorf("can't find proposal with ID '%d'", proposalID)
		}

		// pick the voters (addresses that will submit vote msgs)
		// num voters determined by whether the proposal should pass or not
		numMembers := int64(len(selectedCommittee.Members))
		majority := selectedCommittee.VoteThreshold.Mul(sdk.NewInt(numMembers).ToDec()).Ceil().TruncateInt64()

		numVoters := r.Int63n(majority) // in interval [0, majority)
		shouldPass := r.Float64() < proposalPassPercentage
		if shouldPass {
			numVoters = majority + r.Int63n(numMembers-majority+1) // in interval [majority, numMembers]
		}
		voters := selectedCommittee.Members[:numVoters]

		// schedule vote operations
		var futureOps []simulation.FutureOperation
		for _, v := range voters {
			rt, err := RandomTime(r, ctx.BlockTime(), proposal.Deadline)
			if err != nil {
				return simulation.NoOpMsg(types.ModuleName), nil, fmt.Errorf("random time generation failed: %w", err)
			}
			fop := simulation.FutureOperation{
				BlockTime: rt,
				Op:        SimulateMsgVote(k, ak, v, proposal.ID),
			}
			futureOps = append(futureOps, fop)
		}

		return submitOpMsg, futureOps, nil
	}
}

func SimulateMsgVote(k keeper.Keeper, ak AccountKeeper, voter sdk.AccAddress, proposalID uint64) simulation.Operation {

	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simulation.Account, chainID string) (
		opMsg simulation.OperationMsg, fOps []simulation.FutureOperation, err error) {

		msg := types.NewMsgVote(voter, proposalID)

		account := ak.GetAccount(ctx, voter)
		fees, err := simulation.RandomFees(r, ctx, account.SpendableCoins(ctx.BlockTime()))
		if err != nil {
			return simulation.NoOpMsg(types.ModuleName), nil, err
		}

		var voterPrivKey crypto.PrivKey
		for _, a := range accs {
			if a.Address.Equals(voter) {
				voterPrivKey = a.PrivKey
				break
			}
		}

		tx := helpers.GenTx(
			[]sdk.Msg{msg},
			fees,
			helpers.DefaultGenTxGas,
			chainID,
			[]uint64{account.GetAccountNumber()},
			[]uint64{account.GetSequence()},
			voterPrivKey,
		)

		_, result, err := app.Deliver(tx)
		if err != nil {
			// to aid debugging, add the stack trace to the comment field of the returned opMsg
			return simulation.NewOperationMsg(msg, false, fmt.Sprintf("%+v", err)), nil, err
		}
		// to aid debugging, add the result log to the comment field
		return simulation.NewOperationMsg(msg, true, result.Log), nil, nil
	}
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
