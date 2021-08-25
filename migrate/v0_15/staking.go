package v0_15

import (
	"io/ioutil"
	"path/filepath"
	"sort"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/distribution"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	"github.com/cosmos/cosmos-sdk/x/slashing"
	"github.com/cosmos/cosmos-sdk/x/staking"
	"github.com/kava-labs/kava/app"

	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/libs/bytes"
	"github.com/tendermint/tendermint/privval"
	tmtypes "github.com/tendermint/tendermint/types"
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
	// sort validators by delegator share so that when we access them by index, we are accessing the validators with the most delegations first
	sort.Slice(stakingGenState.Validators, func(i, j int) bool {
		return stakingGenState.Validators[i].DelegatorShares.GT(stakingGenState.Validators[j].DelegatorShares)
	})

	updatedStakingGenState := ReplaceValidatorsInStakingState(stakingGenState, validatorInfos)
	updatedDistributionGenState := ReplacePreviousProposerInDistributionState(stakingGenState, distGenState, validatorInfos)
	updatedSlashingGenState := ReplaceConsAddressesInSlashingState(stakingGenState, slashingGenState, validatorInfos)
	return updatedStakingGenState, updatedDistributionGenState, updatedSlashingGenState
}

func ReplaceValidatorsInStakingState(stakingGenesisState staking.GenesisState, validatorInfos ValidatorInfos) staking.GenesisState {
	newValidators := make(staking.Validators, len(stakingGenesisState.Validators))
	copy(newValidators, stakingGenesisState.Validators)
	updatedStakingGenesisState := staking.GenesisState(stakingGenesisState)
	updatedStakingGenesisState.Validators = newValidators
	sort.Slice(updatedStakingGenesisState.Validators, func(i, j int) bool {
		return updatedStakingGenesisState.Validators[i].DelegatorShares.GT(updatedStakingGenesisState.Validators[j].DelegatorShares)
	})
	for idx, vi := range validatorInfos {
		newConsPub := vi.KeyInfo.PubKey
		updatedStakingGenesisState.Validators[idx].ConsPubKey = newConsPub
	}
	return updatedStakingGenesisState
}

func ReplacePreviousProposerInDistributionState(stakingGenesisState staking.GenesisState, distributionGenesisState distribution.GenesisState, validatorInfos ValidatorInfos) distribution.GenesisState {
	validatorConsAddrMapping := make(map[string]string)
	for idx, vi := range validatorInfos {
		validatorConsAddrMapping[sdk.GetConsAddress(stakingGenesisState.Validators[idx].ConsPubKey).String()] = vi.Address.String()
	}
	prevProposer, found := validatorConsAddrMapping[distributionGenesisState.PreviousProposer.String()]
	if found {
		prevProposerReplacementAddr, err := sdk.ConsAddressFromBech32(prevProposer)
		if err != nil {
			panic(err)
		}
		distributionGenesisState.PreviousProposer = prevProposerReplacementAddr
	}
	return distributionGenesisState

}

func ReplaceConsAddressesInSlashingState(stakingGenesisState staking.GenesisState, slashingGenesisState slashing.GenesisState, validatorInfos ValidatorInfos) slashing.GenesisState {
	validatorConsAddrMapping := make(map[string]string)
	for idx, vi := range validatorInfos {
		validatorConsAddrMapping[sdk.GetConsAddress(stakingGenesisState.Validators[idx].ConsPubKey).String()] = vi.Address.String()
	}
	updatedMissedBlocks := make(map[string][]slashing.MissedBlock)
	updatedSigningInfos := make(map[string]slashing.ValidatorSigningInfo)
	for consAddrString, mb := range slashingGenesisState.MissedBlocks {
		consAddrReplacementAddr, found := validatorConsAddrMapping[consAddrString]
		if found {
			updatedMissedBlocks[consAddrReplacementAddr] = mb
		} else {
			updatedMissedBlocks[consAddrString] = mb
		}
	}
	for consAddrString, vsi := range slashingGenesisState.SigningInfos {
		consAddrReplacementAddr, found := validatorConsAddrMapping[consAddrString]
		if found {
			updatedSigningInfos[consAddrReplacementAddr] = vsi
		} else {
			updatedSigningInfos[consAddrString] = vsi
		}
	}
	return slashing.NewGenesisState(slashingGenesisState.Params, updatedSigningInfos, updatedMissedBlocks)
}

func MigrateStaking(v0_14AppState genutil.AppMap, validators []tmtypes.GenesisValidator, validatorInfos ValidatorInfos) {
	v0_14Codec := makeV014Codec()
	v0_15Codec := app.MakeCodec()
	if v0_14AppState[staking.ModuleName] != nil {
		var stakingGenState staking.GenesisState
		v0_14Codec.MustUnmarshalJSON(v0_14AppState[staking.ModuleName], &stakingGenState)

		var slashingGenState slashing.GenesisState
		v0_14Codec.MustUnmarshalJSON(v0_14AppState[slashing.ModuleName], &slashingGenState)

		var distributionGenState distribution.GenesisState // v0_14 hard genesis state is the same as v0_15
		v0_15Codec.MustUnmarshalJSON(v0_14AppState[distribution.ModuleName], &distributionGenState)

		newStakingGenState, newDistributionGenState, newSlashingGenState := Staking(v0_15Codec, stakingGenState, distributionGenState, slashingGenState, ValidatorKeysDir)
		delete(v0_14AppState, staking.ModuleName)
		delete(v0_14AppState, distribution.ModuleName)
		delete(v0_14AppState, slashing.ModuleName)

		v0_14AppState[staking.ModuleName] = v0_15Codec.MustMarshalJSON(newStakingGenState)
		v0_14AppState[distribution.ModuleName] = v0_15Codec.MustMarshalJSON(newDistributionGenState)
		v0_14AppState[slashing.ModuleName] = v0_15Codec.MustMarshalJSON(newSlashingGenState)
	}
	validatorPubkeyMapping := make(map[string]crypto.PubKey)
	sort.Slice(validators, func(i, j int) bool {
		return validators[i].Power > validators[j].Power
	})
	for idx, vi := range validatorInfos {
		validatorPubkeyMapping[sdk.GetConsAddress(validators[idx].PubKey).String()] = vi.KeyInfo.PubKey
	}
	for idx, validator := range validators {
		pubKeyReplacement, found := validatorPubkeyMapping[sdk.GetConsAddress(validator.PubKey).String()]
		if found {
			validators[idx].PubKey = pubKeyReplacement
			validators[idx].Address = bytes.HexBytes{}
		}
	}
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
			Address: sdk.GetConsAddress(pv.PubKey),
			KeyInfo: pv,
		}
		validatorInfos = append(validatorInfos, vi)
	}
	return validatorInfos, nil
}
