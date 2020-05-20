package v18de63

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
)

const ModuleName = staking.ModuleName // TODO copy in?

// most types on this struct are unchanged between v18de63 and 0.38.3, so we can import types from the current v0.38
// should we copy them in anyway?
type GenesisState struct {
	Params               Params                        `json:"params" yaml:"params"`
	LastTotalPower       sdk.Int                       `json:"last_total_power" yaml:"last_total_power"`
	LastValidatorPowers  []staking.LastValidatorPower  `json:"last_validator_powers" yaml:"last_validator_powers"`
	Validators           staking.Validators            `json:"validators" yaml:"validators"`
	Delegations          staking.Delegations           `json:"delegations" yaml:"delegations"`
	UnbondingDelegations []staking.UnbondingDelegation `json:"unbonding_delegations" yaml:"unbonding_delegations"`
	Redelegations        []staking.Redelegation        `json:"redelegations" yaml:"redelegations"`
	Exported             bool                          `json:"exported" yaml:"exported"`
}

type Params struct {
	UnbondingTime time.Duration `json:"unbonding_time" yaml:"unbonding_time"` // time duration of unbonding
	MaxValidators uint16        `json:"max_validators" yaml:"max_validators"` // maximum number of validators (max uint16 = 65535)
	MaxEntries    uint16        `json:"max_entries" yaml:"max_entries"`       // max entries for either unbonding delegation or redelegation (per pair/trio)
	// note: we need to be a bit careful about potential overflow here, since this is user-determined
	BondDenom string `json:"bond_denom" yaml:"bond_denom"` // bondable coin denomination
}
