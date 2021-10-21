package simulation

import (
	"fmt"
	"math/rand"

	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/cosmos/cosmos-sdk/x/simulation"

	appparams "github.com/kava-labs/kava/app/params"
	"github.com/kava-labs/kava/x/committee/keeper"
	"github.com/kava-labs/kava/x/committee/types"
)

const OpWeightSubmitCommitteeChangeProposal = "op_weight_submit_committee_change_proposal"

// ProposalContents defines the module weighted proposals' contents
func ProposalContents(k keeper.Keeper, paramChanges []simulation.ParamChange) []simulation.WeightedProposalContent {
	return []simulation.WeightedProposalContent{
		{
			AppParamsKey:       OpWeightSubmitCommitteeChangeProposal,
			DefaultWeight:      appparams.OpWeightSubmitCommitteeChangeProposal,
			ContentSimulatorFn: SimulateCommitteeChangeProposalContent(k, paramChanges),
		},
	}
}

// SimulateCommitteeChangeProposalContent generates gov proposal contents that either:
// - create new committees
// - change existing committees
// - delete committees
// It does not alter the fallback committee.
func SimulateCommitteeChangeProposalContent(k keeper.Keeper, paramChanges []simulation.ParamChange) simulation.ContentSimulatorFn {
	return func(r *rand.Rand, ctx sdk.Context, accs []simulation.Account) govtypes.Content {
		allowedParams := paramChangeToAllowedParams(paramChanges)

		// get current committees, ignoring the fallback committee
		var committees []types.Committee
		k.IterateCommittees(ctx, func(com types.Committee) bool {
			if com.GetID() != FallbackCommitteeID {
				committees = append(committees, com)
			}
			return false
		})
		if len(committees) < 1 { // create a committee if none exist
			com, err := RandomCommittee(r, firstNAccounts(25, accs), allowedParams) // limit num members to avoid overflowing hardcoded gov ops gas limit
			if err != nil {
				panic(err)
			}
			return types.NewCommitteeChangeProposal(
				simulation.RandStringOfLength(r, 10),
				simulation.RandStringOfLength(r, 100),
				com,
			)
		}

		// create a proposal content

		var content govtypes.Content
		switch choice := r.Intn(100); {

		// create committee
		case choice < 20:
			com, err := RandomCommittee(r, firstNAccounts(25, accs), allowedParams) // limit num members to avoid overflowing hardcoded gov ops gas limit
			if err != nil {
				panic(err)
			}
			content = types.NewCommitteeChangeProposal(
				simulation.RandStringOfLength(r, 10),
				simulation.RandStringOfLength(r, 100),
				com,
			)

		// update committee
		case choice < 80:
			com := committees[r.Intn(len(committees))]

			// update members
			if r.Intn(100) < 50 {
				if len(accs) == 0 {
					panic("must have at least one account available to use as committee member")
				}
				var members []sdk.AccAddress
				for len(members) < 1 {
					members = RandomAddresses(r, firstNAccounts(25, accs)) // limit num members to avoid overflowing hardcoded gov ops gas limit
				}
				com.SetMembers(members)
			}
			// update permissions
			if r.Intn(100) < 50 {
				com.SetPermissions(RandomPermissions(r, allowedParams))
			}
			// update proposal duration
			if r.Intn(100) < 50 {
				dur, err := RandomPositiveDuration(r, 0, AverageBlockTime*100)
				if err != nil {
					panic(err)
				}
				com.SetProposalDuration(dur)
			}
			// update vote threshold
			if r.Intn(100) < 50 {
				// VoteThreshold must be in interval (0,1]
				com.SetVoteThreshold(simulation.RandomDecAmount(r, sdk.MustNewDecFromStr("1").Sub(sdk.SmallestDec())).Add(sdk.SmallestDec()))
			}

			content = types.NewCommitteeChangeProposal(
				simulation.RandStringOfLength(r, 10),
				simulation.RandStringOfLength(r, 100),
				com,
			)

		// delete committee
		default:
			com := committees[r.Intn(len(committees))]
			content = types.NewCommitteeDeleteProposal(
				simulation.RandStringOfLength(r, 10),
				simulation.RandStringOfLength(r, 100),
				com.GetID(),
			)
		}

		return content
	}
}

// Example custom ParamChangeProposal generator to only generate change to interesting cdp params.
// This allows more control over what params are changed within a simulation.
/*
func SimulateCDPParamChangeProposalContent(cdpKeeper cdpkeeper.Keeper, paramChangePool []simulation.ParamChange) simulation.ContentSimulatorFn {
	return func(r *rand.Rand, ctx sdk.Context, _ []simulation.Account) govtypes.Content {

		var paramChanges []paramstypes.ParamChange

		// alter sub params
		cp := cdpKeeper.GetParams(ctx).CollateralParams
		if len(cp) == 0 {
			return nil
		}
		cp[0].StabilityFee = sdk.MustNewDecFromStr("0.000001")
		paramChanges = append(
			paramChanges,
			paramstypes.NewParamChange(cdptypes.ModuleName, "?", string(cdptypes.ModuleCdc.MustMarshalJSON(cp))),
		)

		// alter normal param
		for _, pc := range paramChangePool {
			if pc.Subspace == cdptypes.ModuleName && pc.Key == string(cdptypes.KeyGlobalDebtLimit) {
				paramChanges = append(
					paramChanges,
					paramstypes.NewParamChange(pc.Subspace, pc.Key, pc.SimValue(r)),
				)
			}
		}

		return paramstypes.NewParameterChangeProposal(
			simulation.RandStringOfLength(r, 140),  // title
			simulation.RandStringOfLength(r, 5000), // description
			paramChanges,                           // set of changes
		)
	}
}
*/
func firstNAccounts(n int, accs []simulation.Account) []simulation.Account {
	if n < 0 {
		panic(fmt.Sprintf("n must be ≥ 0"))
	}
	if n > len(accs) {
		return accs
	}
	return accs[:n]
}
