package simulation

import (
	authsim "github.com/cosmos/cosmos-sdk/x/auth/simulation"
	banksim "github.com/cosmos/cosmos-sdk/x/bank/simulation"
	distributionsim "github.com/cosmos/cosmos-sdk/x/distribution/simulation"
	govsim "github.com/cosmos/cosmos-sdk/x/gov/simulation"
	mintsim "github.com/cosmos/cosmos-sdk/x/mint/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"
	slashingsim "github.com/cosmos/cosmos-sdk/x/slashing/simulation"
	stakingsim "github.com/cosmos/cosmos-sdk/x/staking/simulation"

	auctionsim "github.com/kava-labs/kava/x/auction/simulation"
	cdpsim "github.com/kava-labs/kava/x/cdp/simulation"
	pricefeedsim "github.com/kava-labs/kava/x/pricefeed/simulation"

	// TODO kavadistsim "github.com/kava-labs/kava/x/kavadist/simulation"
	// TODO incentivesim "github.com/kava-labs/kava/x/incentive/simulation"
	"github.com/kava-labs/kava/x/committee/types"
)

// getAllowedParamKeys collects up and returns the keys of all the params that can be changed during a simulation
// TODO This only exists because the random genesis needs access to the available params to generate committee permissions
// and there was no way to pass that information down from the simulation test.
func GetAllowedParamKeys() []types.AllowedParam {
	paramChanges := [][]simulation.ParamChange{
		authsim.ParamChanges(nil),
		banksim.ParamChanges(nil),
		distributionsim.ParamChanges(nil),
		govsim.ParamChanges(nil),
		mintsim.ParamChanges(nil),
		slashingsim.ParamChanges(nil),
		stakingsim.ParamChanges(nil),
		auctionsim.ParamChanges(nil),
		cdpsim.ParamChanges(nil),
		pricefeedsim.ParamChanges(nil),
	}

	var allowedParams []types.AllowedParam
	for _, pcs := range paramChanges {
		for _, pc := range pcs {
			allowedParams = append(
				allowedParams,
				types.AllowedParam{
					Subspace: pc.Subspace,
					Key:      pc.Key,
					Subkey:   pc.Subkey,
				},
			)
		}
	}
	return allowedParams
}
