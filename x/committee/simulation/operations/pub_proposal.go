package operations

import (
	"fmt"
	"math/rand"

	sdk "github.com/cosmos/cosmos-sdk/types"
	govsimops "github.com/cosmos/cosmos-sdk/x/gov/simulation/operations"
	"github.com/cosmos/cosmos-sdk/x/params"
	"github.com/cosmos/cosmos-sdk/x/simulation"

	"github.com/kava-labs/kava/x/committee"
)

type PubProposalSimulator func(r *rand.Rand, ctx sdk.Context, accs []simulation.Account, perm committee.Permission) committee.PubProposal

// Generating pubProposals is different for committees compared to gov. Committees have permissions that limit the valid pubproposals.
// Solution here passes a permission into the pubproposal generation function. This function converts existing gov content generators into PubProposal simulators.
func SimulateAnyPubProposal(textSimulator govsimops.ContentSimulator, paramSimulator govsimops.ContentSimulator, paramFromPermSimulator PubProposalSimulator) PubProposalSimulator {
	return func(r *rand.Rand, ctx sdk.Context, accs []simulation.Account, permission committee.Permission) committee.PubProposal {
		switch perm := permission.(type) {
		case committee.GodPermission:
			switch r.Intn(2) {
			case 0:
				return textSimulator(r, ctx, accs)
			default:
				return paramSimulator(r, ctx, accs)
			}
		case committee.ParamChangePermission:
			return paramFromPermSimulator(r, ctx, accs, perm)
		case committee.TextPermission:
			return textSimulator(r, ctx, accs)
		default:
			panic(fmt.Sprintf("unexpected permission type %T", permission))
		}
	}
}

func SimulateParamChangePubProposal(paramChanges []simulation.ParamChange) PubProposalSimulator {
	return func(r *rand.Rand, ctx sdk.Context, accs []simulation.Account, permission committee.Permission) committee.PubProposal {
		perm, ok := permission.(committee.ParamChangePermission)
		if !ok {
			panic("expected permission to be of type ParamChangePermission")
		}

		// get available params that are allowed for the committee
		var availableParamChanges []simulation.ParamChange
		for _, pc := range paramChanges {
			for _, ap := range perm.AllowedParams {
				if paramChangeAllowed(pc, ap) {
					availableParamChanges = append(availableParamChanges, pc)
				}
				// this could produce duplicate if paramChanges, or AllowedParams contain duplicates
			}
		}

		// generate param changes
		numChanges := r.Intn(len(availableParamChanges))
		indexes := r.Perm(len(availableParamChanges))[:numChanges]

		var changes []params.ParamChange
		for _, i := range indexes {
			pc := availableParamChanges[i]
			changes = append(changes, params.NewParamChangeWithSubkey(pc.Subspace, pc.Key, pc.Subkey, pc.SimValue(r)))
		}

		return params.NewParameterChangeProposal(
			simulation.RandStringOfLength(r, 140),
			simulation.RandStringOfLength(r, 5000),
			changes,
		)
	}
}

func paramChangeAllowed(pc simulation.ParamChange, ap committee.AllowedParam) bool {
	return ap.Subspace+ap.Key+ap.Subkey == pc.ComposedKey()
}
