package v0_8

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	tmtypes "github.com/tendermint/tendermint/types"

	v038auth "github.com/cosmos/cosmos-sdk/x/auth"
	v038authexported "github.com/cosmos/cosmos-sdk/x/auth/exported"
	v038vesting "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	v038dist "github.com/cosmos/cosmos-sdk/x/distribution"
	v038evidence "github.com/cosmos/cosmos-sdk/x/evidence"
	v038genutil "github.com/cosmos/cosmos-sdk/x/genutil"
	v038genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	v038slashing "github.com/cosmos/cosmos-sdk/x/slashing"
	v038staking "github.com/cosmos/cosmos-sdk/x/staking"
	v038supply "github.com/cosmos/cosmos-sdk/x/supply"
	v038upgrade "github.com/cosmos/cosmos-sdk/x/upgrade"

	"github.com/kava-labs/kava/app"
	v18de63auth "github.com/kava-labs/kava/migrate/v0_8/sdk/auth/v18de63"
	v038distcustom "github.com/kava-labs/kava/migrate/v0_8/sdk/distribution/v0_38"
	v18de63dist "github.com/kava-labs/kava/migrate/v0_8/sdk/distribution/v18de63"
	v038evidencecustom "github.com/kava-labs/kava/migrate/v0_8/sdk/evidence/v0_38"
	v038slashingcustom "github.com/kava-labs/kava/migrate/v0_8/sdk/slashing/v0_38"
	v18de63slashing "github.com/kava-labs/kava/migrate/v0_8/sdk/slashing/v18de63"
	v038stakingcustom "github.com/kava-labs/kava/migrate/v0_8/sdk/staking/v0_38"
	v18de63staking "github.com/kava-labs/kava/migrate/v0_8/sdk/staking/v18de63"
	v18de63supply "github.com/kava-labs/kava/migrate/v0_8/sdk/supply/v18de63"
	v032tendermint "github.com/kava-labs/kava/migrate/v0_8/tendermint/v0_32"
	v033tendermint "github.com/kava-labs/kava/migrate/v0_8/tendermint/v0_33"
	"github.com/kava-labs/kava/x/auction"
	"github.com/kava-labs/kava/x/bep3"
	"github.com/kava-labs/kava/x/cdp"
	"github.com/kava-labs/kava/x/committee"
	"github.com/kava-labs/kava/x/incentive"
	"github.com/kava-labs/kava/x/kavadist"
	"github.com/kava-labs/kava/x/pricefeed"
	v0_3validator_vesting "github.com/kava-labs/kava/x/validator-vesting/legacy/v0_3"
	v0_8validator_vesting "github.com/kava-labs/kava/x/validator-vesting/types"
)

// Migrate translates a genesis file from kava v0.3.x format to kava v0.8.x format.
func Migrate(v0_3GenDoc v032tendermint.GenesisDoc) tmtypes.GenesisDoc {

	// migrate app state
	var appStateMap v038genutil.AppMap
	if err := v032tendermint.Cdc.UnmarshalJSON(v0_3GenDoc.AppState, &appStateMap); err != nil {
		panic(err)
	}
	newAppState := MigrateAppState(appStateMap)
	v0_8Codec := app.MakeCodec()
	marshaledNewAppState, err := v0_8Codec.MarshalJSON(newAppState)
	if err != nil {
		panic(err)
	}

	// migrate tendermint state
	newGenDoc := v033tendermint.Migrate(v0_3GenDoc)

	newGenDoc.AppState = marshaledNewAppState
	return newGenDoc
}

func MigrateAppState(v0_3AppState v038genutil.AppMap) v038genutil.AppMap {

	// run sdk migrations for commit 18de63 to v0.38, ignoring auth (validator vesting needs a custom migration)
	v0_8AppState := MigrateSDK(v0_3AppState)

	// create codec for current app version
	v0_8Codec := app.MakeCodec()

	// migrate auth state
	// recreate the codec from v0.3 (note: current codec and crypto packages are backwards compatible) (note this is not a compete recreation of v0.3's codec)
	v0_3Codec := codec.New()
	codec.RegisterCrypto(v0_3Codec)
	v18de63auth.RegisterCodec(v0_3Codec)
	v18de63auth.RegisterCodecVesting(v0_3Codec) // TODO probably split out vesting package
	v18de63supply.RegisterCodec(v0_3Codec)
	v0_3validator_vesting.RegisterCodec(v0_3Codec)

	if v0_3AppState[v18de63auth.ModuleName] != nil {
		var authGenState v18de63auth.GenesisState
		v0_3Codec.MustUnmarshalJSON(v0_3AppState[v18de63auth.ModuleName], &authGenState)

		delete(v0_3AppState, v18de63auth.ModuleName)
		v0_8AppState[v038auth.ModuleName] = v0_8Codec.MustMarshalJSON(MigrateAuth(authGenState))
	}

	// migrate new modules (by adding new gen states)
	v0_8AppState[auction.ModuleName] = v0_8Codec.MustMarshalJSON(auction.DefaultGenesisState())
	v0_8AppState[bep3.ModuleName] = v0_8Codec.MustMarshalJSON(bep3.DefaultGenesisState())
	v0_8AppState[cdp.ModuleName] = v0_8Codec.MustMarshalJSON(cdp.DefaultGenesisState())
	v0_8AppState[committee.ModuleName] = v0_8Codec.MustMarshalJSON(committee.DefaultGenesisState())
	v0_8AppState[incentive.ModuleName] = v0_8Codec.MustMarshalJSON(incentive.DefaultGenesisState())
	v0_8AppState[kavadist.ModuleName] = v0_8Codec.MustMarshalJSON(kavadist.DefaultGenesisState())
	v0_8AppState[pricefeed.ModuleName] = v0_8Codec.MustMarshalJSON(pricefeed.DefaultGenesisState())

	return v0_8AppState
}

// migrate the sdk modules from commit 18de63 (v0.37 and half) to v0.38.3, mostly copying v038.Migrate
func MigrateSDK(appState v038genutil.AppMap) v038genutil.AppMap {

	v18de63Codec := codec.New() // ideally this would use the exact version of amino from kava v0.3
	codec.RegisterCrypto(v18de63Codec)
	v18de63auth.RegisterCodec(v18de63Codec)

	// v038Codec := codec.New()
	// codec.RegisterCrypto(v038Codec)
	// v038auth.RegisterCodec(v038Codec)
	v038Codec := app.MakeCodec() // TODO should we use sim app codec ?

	// for each module, unmarshal old state, run a migrate(genesisStateType) func, marshal returned type into json and set

	// migrate distribution state
	if appState[v18de63dist.ModuleName] != nil {
		var distGenState v18de63dist.GenesisState
		v18de63Codec.MustUnmarshalJSON(appState[v18de63dist.ModuleName], &distGenState)

		fmt.Println(string(appState[v18de63dist.ModuleName]))

		delete(appState, v18de63dist.ModuleName) // delete old key in case the name changed
		appState[v038dist.ModuleName] = v038Codec.MustMarshalJSON(v038distcustom.Migrate(distGenState))
	}

	// migrate slashing and evidence state
	if appState[v18de63slashing.ModuleName] != nil {
		var slashingGenState v18de63slashing.GenesisState
		v18de63Codec.MustUnmarshalJSON(appState[v18de63slashing.ModuleName], &slashingGenState)

		delete(appState, v18de63slashing.ModuleName)
		// remove param
		appState[v038slashing.ModuleName] = v038Codec.MustMarshalJSON(v038slashingcustom.Migrate(slashingGenState))
		// add new evidence module genesis (with above param)
		appState[v038evidence.ModuleName] = v038Codec.MustMarshalJSON(v038evidencecustom.Migrate(slashingGenState))
	}

	// migrate staking state
	if appState[v18de63staking.ModuleName] != nil {
		var stakingGenState v18de63staking.GenesisState
		v18de63Codec.MustUnmarshalJSON(appState[v18de63staking.ModuleName], &stakingGenState)

		delete(appState, v18de63staking.ModuleName)
		appState[v038staking.ModuleName] = v038Codec.MustMarshalJSON(v038stakingcustom.Migrate(stakingGenState))
	}

	// migrate genutil state
	appState[v038genutil.ModuleName] = v038Codec.MustMarshalJSON(v038genutiltypes.DefaultGenesisState())

	// add upgrade state
	appState[v038upgrade.ModuleName] = []byte(`{}`)

	return appState
}

func MigrateAuth(oldGenState v18de63auth.GenesisState) v038auth.GenesisState {
	// old and new struct type are identical but with different (un)marshalJSON methods
	var newAccounts v038authexported.GenesisAccounts
	for _, account := range oldGenState.Accounts {
		switch acc := account.(type) {
		case *v18de63auth.BaseAccount:
			a := v038auth.BaseAccount(*acc)
			newAccounts = append(newAccounts, v038authexported.GenesisAccount(&a))
		case *v18de63auth.BaseVestingAccount:
			ba := v038auth.BaseAccount(*(acc.BaseAccount))
			bva := v038vesting.BaseVestingAccount{
				BaseAccount:      &ba,
				OriginalVesting:  acc.OriginalVesting,
				DelegatedFree:    acc.DelegatedFree,
				DelegatedVesting: acc.DelegatedVesting,
				EndTime:          acc.EndTime,
			}
			newAccounts = append(newAccounts, v038authexported.GenesisAccount(&bva))
		// TODO
		// case *v18de63auth.ContinuousVestingAccount:
		// case *v18de63auth.DelayedVestingAccount:
		case *v18de63auth.PeriodicVestingAccount:
			ba := v038auth.BaseAccount(*(acc.BaseVestingAccount.BaseAccount))
			bva := v038vesting.BaseVestingAccount{
				BaseAccount:      &ba,
				OriginalVesting:  acc.BaseVestingAccount.OriginalVesting,
				DelegatedFree:    acc.BaseVestingAccount.DelegatedFree,
				DelegatedVesting: acc.BaseVestingAccount.DelegatedVesting,
				EndTime:          acc.BaseVestingAccount.EndTime,
			}
			var newPeriods v038vesting.Periods
			for _, p := range acc.VestingPeriods {
				newPeriods = append(newPeriods, v038vesting.Period(p))
			}
			pva := v038vesting.PeriodicVestingAccount{
				BaseVestingAccount: &bva,
				StartTime:          acc.StartTime,
				VestingPeriods:     newPeriods,
			}
			newAccounts = append(newAccounts, v038authexported.GenesisAccount(&pva))
		// TODO does unmarshal json return pointers or concrete types?
		case *v18de63supply.ModuleAccount:
			ba := v038auth.BaseAccount(*(acc.BaseAccount))
			ma := v038supply.ModuleAccount{
				BaseAccount: &ba,
				Name:        acc.Name,
				Permissions: acc.Permissions,
			}
			newAccounts = append(newAccounts, v038authexported.GenesisAccount(&ma))
		case *v0_3validator_vesting.ValidatorVestingAccount:
			ba := v038auth.BaseAccount(*(acc.PeriodicVestingAccount.BaseVestingAccount.BaseAccount))
			bva := v038vesting.BaseVestingAccount{
				BaseAccount:      &ba,
				OriginalVesting:  acc.PeriodicVestingAccount.BaseVestingAccount.OriginalVesting,
				DelegatedFree:    acc.PeriodicVestingAccount.BaseVestingAccount.DelegatedFree,
				DelegatedVesting: acc.PeriodicVestingAccount.BaseVestingAccount.DelegatedVesting,
				EndTime:          acc.PeriodicVestingAccount.BaseVestingAccount.EndTime,
			}
			var newPeriods v038vesting.Periods
			for _, p := range acc.PeriodicVestingAccount.VestingPeriods {
				newPeriods = append(newPeriods, v038vesting.Period(p))
			}
			pva := v038vesting.PeriodicVestingAccount{
				BaseVestingAccount: &bva,
				StartTime:          acc.PeriodicVestingAccount.StartTime,
				VestingPeriods:     newPeriods,
			}
			var newVestingProgress []v0_8validator_vesting.VestingProgress
			for _, p := range acc.VestingPeriodProgress {
				newVestingProgress = append(newVestingProgress, v0_8validator_vesting.VestingProgress(p))
			}
			vva := v0_8validator_vesting.ValidatorVestingAccount{
				PeriodicVestingAccount: &pva,
				ValidatorAddress:       acc.ValidatorAddress,
				ReturnAddress:          acc.ReturnAddress,
				SigningThreshold:       acc.SigningThreshold,
				CurrentPeriodProgress:  v0_8validator_vesting.CurrentPeriodProgress(acc.CurrentPeriodProgress),
				VestingPeriodProgress:  newVestingProgress,
				DebtAfterFailedVesting: acc.DebtAfterFailedVesting,
			}
			newAccounts = append(newAccounts, v038authexported.GenesisAccount(&vva))
		default:
			panic(fmt.Sprintf("unrecognized account type: %T", acc))
		}
	}
	gs := v038auth.GenesisState{
		Params:   v038auth.Params(oldGenState.Params),
		Accounts: newAccounts,
	}
	return gs
}
