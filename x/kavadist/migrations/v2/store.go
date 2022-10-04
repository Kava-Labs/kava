package v2

import (
	"encoding/json"

	sdk "github.com/cosmos/cosmos-sdk/types"

	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/kava-labs/kava/x/kavadist/types"
)

// MigrateStore performs in-place migrations from kava 10 to kava 11.
// The migration includes:
//
// - Sets the default InfrastructureParams parameter.
func MigrateStore(ctx sdk.Context, ss paramtypes.Subspace) error {
	if !ss.HasKeyTable() {
		ss = ss.WithKeyTable(types.ParamKeyTable())
	}
	infraParamsJson := `{
		"infrastructure_periods": [
			{
				"start": "2022-10-13T14:00:00Z",
				"end": "2026-10-13T14:00:00Z",
				"inflation": "1.000000007075835620"
			}
		],
		"core_rewards": [
			{
				"address": "kava19zhewrsuqjjqgmhamk6cz6nzsnadg8jv5raalp",
				"weight": "1.0"
			}
		],
		"partner_rewards": [
			{
				"address": "kava1w27q3jmv3lyx6t2ttedxdcfmxuk9ve4gl4rvlu",
				"rewards_per_second": {
					"denom": "ukava",
					"amount": "31710"
				}
			},
			{
				"address": "kava1g2r44etkdh756urh9gaf26zzujms34v6j4afgh",
				"rewards_per_second": {
					"denom": "ukava",
					"amount": "3171"
				}
			},
			{
				"address": "kava125ls9zymjl2739ulj5znmhevyq9lcj5m4fjnkc",
				"rewards_per_second": {
					"denom": "ukava",
					"amount": "634"
				}
			},
			{
				"address": "kava13u8cqwhds2qghga0vajjpu6r7gk0323whkjkzm",
				"rewards_per_second": {
					"denom": "ukava",
					"amount": "317"
				}
			},
			{
				"address": "kava1pjd09vjvy5trdmydqn9k36mhee3vtejc6tj6xy",
				"rewards_per_second": {
					"denom": "ukava",
					"amount": "317"
				}
			},
			{
				"address": "kava140g8fnnl46mlvfhygj3zvjqlku6x0fwuhfj3uf",
				"rewards_per_second": {
					"denom": "ukava",
					"amount": "6342"
				}
			},
			{
				"address": "kava1wf4mqk2k62puxh0apvr4ztnnvmrzwam30e7w6t",
				"rewards_per_second": {
					"denom": "ukava",
					"amount": "317"
				}
			},
			{
				"address": "kava155svs6sgxe55rnvs6ghprtqu0mh69kehlxmsky",
				"rewards_per_second": {
					"denom": "ukava",
					"amount": "634"
				}
			},
			{
				"address": "kava1fkshuyjja57yakf63welsy76zdlcc48dgxp7an",
				"rewards_per_second": {
					"denom": "ukava",
					"amount": "317"
				}
			},
			{
				"address": "kava1d94gy2s6w88fsk7k664um53q3p6mhc0y9dhu5f",
				"rewards_per_second": {
					"denom": "ukava",
					"amount": "159"
				}
			},
			{
				"address": "kava1tsfvkjmmn7eelflyx7s0av3leqz55pjf49vp8j",
				"rewards_per_second": {
					"denom": "ukava",
					"amount": "1585"
				}
			},
			{
				"address": "kava1pr2u860fgn8fn0g8t2e7jc92jn22jldzqpde4r",
				"rewards_per_second": {
					"denom": "ukava",
					"amount": "634"
				}
			},
			{
				"address": "kava17sfh0uh4hk537c7njcwvtz5036jumk8msmjyyw",
				"rewards_per_second": {
					"denom": "ukava",
					"amount": "317"
				}
			},
			{
				"address": "kava1x8vwczqxp6l0f7j332ddrgcwer2nh42etjjspv",
				"rewards_per_second": {
					"denom": "ukava",
					"amount": "317"
				}
			},
			{
				"address": "kava1zqktfkj34xfkhg3f005vjv95efjctpk6mv9x7l",
				"rewards_per_second": {
					"denom": "ukava",
					"amount": "317"
				}
			},
			{
				"address": "kava1c9ye54e3pzwm3e0zpdlel6pnavrj9qqv6e8r4h",
				"rewards_per_second": {
					"denom": "ukava",
					"amount": "317"
				}
			},
			{
				"address": "kava1spjfdd4cxat53sv6fpglz4qfqcuk9gl3k2wya2",
				"rewards_per_second": {
					"denom": "ukava",
					"amount": "634"
				}
			},
			{
				"address": "kava16t00cvmr20eegltl8rhygcl0ljqe4783qzypk0",
				"rewards_per_second": {
					"denom": "ukava",
					"amount": "634"
				}
			}
		]
	}`
	infraParams := types.DefaultInfraParams
	if err := json.Unmarshal([]byte(infraParamsJson), &infraParams); err != nil {
		return err
	}
	ss.Set(ctx, types.KeyInfra, infraParams)
	return nil
}
