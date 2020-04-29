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

// WeightedOperations creates an operation (with weight) for each type of proposal generator.
// Custom proposal generators can be added for more control over types of proposal submitted, eg to increase likelyhood of particular cdp param changes.
func WeightedOperations(appParams simulation.AppParams, cdc *codec.Codec, ak AccountKeeper,
	k keeper.Keeper, wContents []simulation.WeightedProposalContent) simulation.WeightedOperations {

	var wops simulation.WeightedOperations

	for _, wContent := range wContents {
		wContent := wContent // pin variable
		if wContent.AppParamsKey == OpWeightSubmitCommitteeChangeProposal {
			// don't include committee change/delete proposals as they're not enabled for submission to committees
			continue
		}
		var weight int
		// TODO this doesn't allow weights to be different from what they are in the gov module
		appParams.GetOrGenerate(cdc, wContent.AppParamsKey, &weight, nil,
			func(_ *rand.Rand) { weight = wContent.DefaultWeight })

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

// SimulateMsgSubmitProposal creates a proposal using the passed contentSimulatorFn and tries to find a committee that has permissions for it. If it can't then it uses the fallback committee.
// If the fallback committee isn't there (eg when using an non-generated genesis) and no committee can be found this emits a no-op msg and doesn't do anything.
// For each submit proposal msg, future ops for the vote messages are generated. Sometimes it doesn't run enough votes to allow the proposal to timeout - the likelihood of this happening is controlled by a parameter.
func SimulateMsgSubmitProposal(ak AccountKeeper, k keeper.Keeper, contentSim simulation.ContentSimulatorFn) simulation.Operation {
	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simulation.Account, chainID string) (simulation.OperationMsg, []simulation.FutureOperation, error) {

		// 1) Send  a submit proposal msg

		committees := k.GetCommittees(ctx)
		// shuffle committees to ensure proposals are distributed across them evenly
		r.Shuffle(len(committees), func(i, j int) {
			committees[i], committees[j] = committees[j], committees[i]
		})
		// move fallback committee to the end of slice
		for i, c := range committees {
			if c.ID == FallbackCommitteeID {
				// switch places with last element
				committees[i], committees[len(committees)-1] = committees[len(committees)-1], committees[i]
			}
		}
		// pick a committee that has permissions for proposal
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
			// fallback committee was not present, this should only happen if not using the generated genesis state
			return simulation.NewOperationMsgBasic(types.ModuleName, "no-operation (no committee has permissions for proposal)", "", false, nil), nil, nil
		}

		// create the msg and tx
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
		proposerPrivKey, err := getPrivKey(accs, proposer)
		if err != nil {
			return simulation.NoOpMsg(types.ModuleName), nil, err
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

		// pick the voters
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
			voteTime, err := RandomTime(r, ctx.BlockTime(), proposal.Deadline)
			if err != nil {
				return simulation.NoOpMsg(types.ModuleName), nil, fmt.Errorf("random time generation failed: %w", err)
			}
			fop := simulation.FutureOperation{
				BlockTime: voteTime,
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

		voterPrivKey, err := getPrivKey(accs, voter)
		if err != nil {
			return simulation.NoOpMsg(types.ModuleName), nil, err
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

func getPrivKey(accs []simulation.Account, addr sdk.Address) (crypto.PrivKey, error) {
	for _, a := range accs {
		if a.Address.Equals(addr) {
			return a.PrivKey, nil
		}
	}
	return nil, fmt.Errorf("address not in accounts %s", addr)
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
	simulation.RandIntBetween()
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
