package simulation

import (
	"fmt"
	"math/rand"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/simapp/helpers"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authexported "github.com/cosmos/cosmos-sdk/x/auth/exported"
	distsim "github.com/cosmos/cosmos-sdk/x/distribution/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"

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
		if wContent.AppParamsKey == OpWeightSubmitCommitteeChangeProposal || wContent.AppParamsKey == distsim.OpWeightSubmitCommunitySpendProposal {
			// don't include committee change/delete proposals as they're not enabled for submission to committees
			// don't include community pool proposals as the generator func sometimes returns nil // TODO replace generator with a better one
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
				SimulateMsgSubmitProposal(cdc, ak, k, wContent.ContentSimulatorFn),
			),
		)
	}
	return wops
}

// SimulateMsgSubmitProposal creates a proposal using the passed contentSimulatorFn and tries to find a committee that has permissions for it. If it can't then it uses the fallback committee.
// If the fallback committee isn't there (eg when using an non-generated genesis) and no committee can be found this emits a no-op msg and doesn't do anything.
// For each submit proposal msg, future ops for the vote messages are generated. Sometimes it doesn't run enough votes to allow the proposal to timeout - the likelihood of this happening is controlled by a parameter.
func SimulateMsgSubmitProposal(cdc *codec.Codec, ak AccountKeeper, k keeper.Keeper, contentSim simulation.ContentSimulatorFn) simulation.Operation {
	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simulation.Account, chainID string) (simulation.OperationMsg, []simulation.FutureOperation, error) {

		// 1) Send  a submit proposal msg

		committees := k.GetCommittees(ctx)
		// shuffle committees to ensure proposals are distributed across them evenly
		r.Shuffle(len(committees), func(i, j int) {
			committees[i], committees[j] = committees[j], committees[i]
		})
		// move fallback committee to the end of slice
		for i, c := range committees {
			if c.GetID() == FallbackCommitteeID {
				// switch places with last element
				committees[i], committees[len(committees)-1] = committees[len(committees)-1], committees[i]
			}
		}
		// pick a committee that has permissions for proposal
		pp := types.PubProposal(contentSim(r, ctx, accs))
		if pp == nil {
			return simulation.NewOperationMsgBasic(types.ModuleName, "no-operation (conent generation function returned nil)", "", false, nil), nil, nil
		}
		var selectedCommittee types.Committee
		var found bool
		for _, c := range committees {
			if c.HasPermissionsFor(ctx, cdc, k.ParamKeeper, pp) {
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
		proposer := selectedCommittee.GetMembers()[r.Intn(len(selectedCommittee.GetMembers()))] // won't panic as committees must have ≥ 1 members
		msg := types.NewMsgSubmitProposal(
			pp,
			proposer,
			selectedCommittee.GetID(),
		)
		account := ak.GetAccount(ctx, proposer)
		fees, err := simulation.RandomFees(r, ctx, account.SpendableCoins(ctx.BlockTime()))
		if err != nil {
			return simulation.NoOpMsg(types.ModuleName), nil, err
		}
		proposerAcc, found := simulation.FindAccount(accs, proposer)
		if !found {
			return simulation.NoOpMsg(types.ModuleName), nil, fmt.Errorf("address not in account list")
		}
		tx := helpers.GenTx(
			[]sdk.Msg{msg},
			fees,
			helpers.DefaultGenTxGas,
			chainID,
			[]uint64{account.GetAccountNumber()},
			[]uint64{account.GetSequence()},
			proposerAcc.PrivKey,
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
		numMembers := int64(len(selectedCommittee.GetMembers()))
		majority := selectedCommittee.GetVoteThreshold().Mul(sdk.NewInt(numMembers).ToDec()).Ceil().TruncateInt64()

		numVoters := r.Int63n(majority) // in interval [0, majority)
		shouldPass := r.Float64() < proposalPassPercentage
		if shouldPass {
			numVoters = majority + r.Int63n(numMembers-majority+1) // in interval [majority, numMembers]
		}
		voters := selectedCommittee.GetMembers()[:numVoters]

		// schedule vote operations
		var futureOps []simulation.FutureOperation
		for _, v := range voters {
			voteTime, err := RandomTime(r, ctx.BlockTime(), proposal.Deadline)
			if err != nil {
				return simulation.NoOpMsg(types.ModuleName), nil, fmt.Errorf("random time generation failed: %w", err)
			}

			// Valid vote types: 0, 1, 2
			randInt, err := RandIntInclusive(r, sdk.ZeroInt(), sdk.NewInt(2))
			if err != nil {
				return simulation.NoOpMsg(types.ModuleName), nil, fmt.Errorf("random vote type generation failed: %w", err)
			}
			voteType := types.VoteType(randInt.Int64())

			fop := simulation.FutureOperation{
				BlockTime: voteTime,
				Op:        SimulateMsgVote(k, ak, v, proposal.ID, voteType),
			}
			futureOps = append(futureOps, fop)
		}

		return submitOpMsg, futureOps, nil
	}
}

func SimulateMsgVote(k keeper.Keeper, ak AccountKeeper, voter sdk.AccAddress, proposalID uint64, voteType types.VoteType) simulation.Operation {
	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simulation.Account, chainID string) (
		opMsg simulation.OperationMsg, fOps []simulation.FutureOperation, err error) {

		msg := types.NewMsgVote(voter, proposalID, voteType)

		account := ak.GetAccount(ctx, voter)
		fees, err := simulation.RandomFees(r, ctx, account.SpendableCoins(ctx.BlockTime()))
		if err != nil {
			return simulation.NoOpMsg(types.ModuleName), nil, err
		}

		voterAcc, found := simulation.FindAccount(accs, voter)
		if !found {
			return simulation.NoOpMsg(types.ModuleName), nil, fmt.Errorf("address not in account list")
		}

		tx := helpers.GenTx(
			[]sdk.Msg{msg},
			fees,
			helpers.DefaultGenTxGas,
			chainID,
			[]uint64{account.GetAccountNumber()},
			[]uint64{account.GetSequence()},
			voterAcc.PrivKey,
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
