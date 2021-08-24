package v0_15

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"sort"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/distribution"
	"github.com/cosmos/cosmos-sdk/x/slashing"
	"github.com/cosmos/cosmos-sdk/x/staking"

	"github.com/tendermint/tendermint/privval"
)

var (
	ValidatorKeysDir = "/Users/kjydavis/go/src/github.com/kava-labs/kava-testnet-scripts/kava-8-preview/data/validator-keys"
)

func Staking(
	cdc *codec.Codec, stakingGenState staking.GenesisState, distGenState distribution.GenesisState,
	slashingGenState slashing.GenesisState, validatorKeysDir string) (staking.GenesisState, distribution.GenesisState, slashing.GenesisState) {

	validatorInfos, err := LoadValidatorInfo(cdc, validatorKeysDir)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%s\n", validatorInfos[0].Address)
	// sort validators by delegator share so that when we access them by index, we are accessing the validators with the most delegations first
	sort.Slice(stakingGenState.Validators, func(i, j int) bool {
		return stakingGenState.Validators[i].DelegatorShares.LTE(stakingGenState.Validators[i].DelegatorShares)
	})

	updatedStakingGenState := ReplaceValidatorsInStakingState(stakingGenState, validatorInfos)
	return updatedStakingGenState, distGenState, slashingGenState
}

func ReplaceValidatorsInStakingState(stakingGenesisState staking.GenesisState, validatorInfos ValidatorInfos) staking.GenesisState {
	for idx, vi := range validatorInfos {
		newConsPub := vi.KeyInfo.PubKey
		stakingGenesisState.Validators[idx].ConsPubKey = newConsPub
	}
	return stakingGenesisState
}

type ValidatorInfo struct {
	Address sdk.ConsAddress
	KeyInfo privval.FilePVKey
}

type ValidatorInfos []ValidatorInfo

func LoadValidatorInfo(cdc *codec.Codec, fp string) (ValidatorInfos, error) {
	var validatorInfos ValidatorInfos

	files, err := ioutil.ReadDir(fp)
	if err != nil {
		return validatorInfos, err
	}
	for _, f := range files {
		bz, err := ioutil.ReadFile(filepath.Join(fp, f.Name()))
		if err != nil {
			return validatorInfos, err
		}
		var pv privval.FilePVKey
		err = cdc.UnmarshalJSON(bz, &pv)
		if err != nil {
			return validatorInfos, err
		}
		vi := ValidatorInfo{
			Address: sdk.ConsAddress(pv.Address),
			KeyInfo: pv,
		}
		validatorInfos = append(validatorInfos, vi)
	}
	return validatorInfos, nil
}
