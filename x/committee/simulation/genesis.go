package simulation

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
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
)

// RandomizedGenState generates a random GenesisState for the module
func RandomizedGenState(simState *module.SimulationState) {
	r := simState.Rand
	allowedParams := GetAllowedParamKeys()

	numCommittees := r.Intn(100)
	var committees []types.Committee
	for i := 0; i < numCommittees; i++ {
		com, err := RandomCommittee(r, simState.Accounts, allowedParams)
		if err != nil {
			panic(err)
		}
		committees = append(committees, com)
	}

	// TODO Proposals or votes aren't generated. Should these be removed from committee's genesis state?
	genesisState := types.NewGenesisState(
		types.DefaultNextProposalID,
		committees,
		[]types.Proposal{},
		[]types.Vote{},
	)

	fmt.Printf("Selected randomly generated %s parameters:\n%s\n", types.ModuleName, codec.MustMarshalJSONIndent(simState.Cdc, genesisState))
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

	// pick committee duration
	dur, err := RandomPositiveDuration(r, 0, AverageBlockTime*100)
	if err != nil {
		return types.Committee{}, err
	}

	return types.NewCommittee(
		r.Uint64(),
		simulation.RandStringOfLength(r, r.Intn(types.MaxCommitteeDescriptionLength+1)),
		members,
		RandomPermissions(r, allowedParams),
		simulation.RandomDecAmount(r, sdk.MustNewDecFromStr("1.00")),
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
