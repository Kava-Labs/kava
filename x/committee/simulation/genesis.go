package simulation

import (
	"fmt"
	"math/rand"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/simulation"

	"github.com/kava-labs/kava/x/committee/types"
)

const (
	// Block time params are un-exported constants in cosmos-sdk/x/simulation.
	// Copy them here in lieu of importing them.
	minTimePerBlock time.Duration = (10000 / 2) * time.Second
	maxTimePerBlock time.Duration = 10000 * time.Second
	// Calculate the average block time
	AverageBlockTime time.Duration = (maxTimePerBlock - minTimePerBlock) / 2

	FallbackCommitteeID uint64 = 0
)

// RandomizedGenState generates a random GenesisState for the module
func RandomizedGenState(simState *module.SimulationState) {
	r := simState.Rand

	// Create a always present committee with god permissions to ensure any randomly generated proposal can always be submitted.
	fallbackCommittee := types.NewCommittee(
		FallbackCommitteeID,
		"A committee with god permissions that will always be in state and not deleted. It ensures any generated proposal can always be submitted and voted on.",
		RandomAddresses(r, simState.Accounts),
		[]types.Permission{types.GodPermission{}},
		sdk.MustNewDecFromStr("0.5"),
		AverageBlockTime*10,
	)

	// Create other committees
	numCommittees := r.Intn(100)
	committees := []types.Committee{fallbackCommittee}
	for i := 0; i < numCommittees; i++ {
		com, err := RandomCommittee(r, firstNAccounts(25, simState.Accounts), paramChangeToAllowedParams(simState.ParamChanges))
		if err != nil {
			panic(err)
		}
		committees = append(committees, com)
	}

	// Add genesis state to simState
	genesisState := types.NewGenesisState(
		types.DefaultNextProposalID,
		committees,
		[]types.Proposal{},
		[]types.Vote{},
	)
	fmt.Printf("Selected randomly generated %s parameters:\n%s\n", types.ModuleName, []byte{})
	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(genesisState)
}

func RandomCommittee(r *rand.Rand, availableAccs []simulation.Account, allowedParams []types.AllowedParam) (types.Committee, error) {
	// pick committee members
	if len(availableAccs) < 1 {
		return types.Committee{}, fmt.Errorf("must be ≥ 1 addresses")
	}
	var members []sdk.AccAddress
	for len(members) < 1 {
		members = RandomAddresses(r, availableAccs)
	}

	// pick proposal duration
	dur, err := RandomPositiveDuration(r, 0, AverageBlockTime*100)
	if err != nil {
		return types.Committee{}, err
	}

	// pick committee vote threshold, must be in interval (0,1]
	threshold := simulation.RandomDecAmount(r, sdk.MustNewDecFromStr("1").Sub(sdk.SmallestDec())).Add(sdk.SmallestDec())

	return types.NewCommittee(
		r.Uint64(), // could collide with other committees, but unlikely
		simulation.RandStringOfLength(r, r.Intn(types.MaxCommitteeDescriptionLength+1)),
		members,
		RandomPermissions(r, allowedParams),
		threshold,
		dur,
	), nil
}

func RandomAddresses(r *rand.Rand, accs []simulation.Account) []sdk.AccAddress {
	r.Shuffle(len(accs), func(i, j int) {
		accs[i], accs[j] = accs[j], accs[i]
	})

	var addresses []sdk.AccAddress
	numAddresses := r.Intn(len(accs) + 1)
	for i := 0; i < numAddresses; i++ {
		addresses = append(addresses, accs[i].Address)
	}
	return addresses
}

func RandomPermissions(r *rand.Rand, allowedParams []types.AllowedParam) []types.Permission {
	var permissions []types.Permission
	if r.Intn(100) < 50 {
		permissions = append(permissions, types.TextPermission{})
	}
	if r.Intn(100) < 50 {
		r.Shuffle(len(allowedParams), func(i, j int) {
			allowedParams[i], allowedParams[j] = allowedParams[j], allowedParams[i]
		})
		permissions = append(permissions,
			types.ParamChangePermission{
				AllowedParams: allowedParams[:r.Intn(len(allowedParams)+1)],
			})
	}
	return permissions
}

func paramChangeToAllowedParams(paramChanges []simulation.ParamChange) []types.AllowedParam {
	var allowedParams []types.AllowedParam
	for _, pc := range paramChanges {
		allowedParams = append(
			allowedParams,
			types.AllowedParam{
				Subspace: pc.Subspace,
				Key:      pc.Key,
			},
		)
	}
	return allowedParams
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
