package simulation

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/simulation"

	"github.com/kava-labs/kava/x/committee"
)

/*
	genesis
	- committee, votes, proposals
	msg
	- submit proposal, vote
	invariants
	- validate state like in genesis
	- proposal doesn't exist after end time
	other
	- generate committee change proposals - write content generator, add to gov

*/

const (
	// Block time params are un-exported constants in cosmos-sdk/x/simulation.
	// Copy them here in lieu of importing them.
	minTimePerBlock time.Duration = (10000 / 2) * time.Second
	maxTimePerBlock time.Duration = 10000 * time.Second
	// Calculate the average block time
	AverageBlockTime time.Duration = (maxTimePerBlock - minTimePerBlock) / 2
)

// RandomizedGenState generates a random GenesisState for the module
func RandomizedGenState(simState *module.SimulationState) {
	r := simState.Rand

	// TODO remove this?
	addresses := make([]sdk.AccAddress, len(simState.Accounts))
	for i, acc := range simState.Accounts {
		addresses[i] = acc.Address
	}

	numCommittees := r.Intn(100)
	var committees []committee.Committee
	for i := 0; i < numCommittees; i++ {
		com, err := randomCommittee(r, addresses)
		if err != nil {
			panic(err)
		}
		committees = append(committees, com)
	}

	// TODO Proposals or votes aren't generated. Should these be removed from committee's genesis state?
	genesisState := committee.NewGenesisState(
		committee.DefaultNextProposalID,
		committees,
		[]committee.Proposal{},
		[]committee.Vote{},
	)
	genesisState.Committees = []committee.Committee{com}

	fmt.Printf("Selected randomly generated %s parameters:\n%s\n", committee.ModuleName, codec.MustMarshalJSONIndent(simState.Cdc, genesisState))
	simState.GenState[committee.ModuleName] = simState.Cdc.MustMarshalJSON(genesisState)
}

func randomCommittee(r *rand.Rand, addresses []sdk.AccAddress) (committee.Committee, error) {
	shuffledIndexes := r.Perm(len(addresses))
	memberIndexes := shuffledIndexes[:r.Intn(len(addresses))]
	members := make([]sdk.AccAddress, len(memberIndexes))
	for mi, ai := range memberIndexes {
		members[mi] = addresses[ai]
	}

	dur, err := RandomPositiveDuration(r, 0, AverageBlockTime*100)
	if err != nil {
		return committee.Committee{}, err
	}

	return committee.NewCommittee(
		r.Uint64(),
		simulation.RandStringOfLength(r, committee.MaxCommitteeDescriptionLength),
		members,
		[]committee.Permission{committee.GodPermission{}}, // TODO
		simulation.RandomDecAmount(r, sdk.MustNewDecFromStr("1.00")),
		dur,
	), nil
}

// TODO move to common location
func RandomPositiveDuration(r *rand.Rand, inclusiveMin, exclusiveMax time.Duration) (time.Duration, error) {
	min := int64(inclusiveMin)
	max := int64(exclusiveMax)
	if min < 0 || max < 0 {
		return 0, fmt.Errorf("min and max must be positive")
	}
	if min >= max {
		return 0, fmt.Errorf("max must be < min")
	}
	randPositiveInt64 := r.Int63n(max-min) + min
	return time.Duration(randPositiveInt64), nil
}
