package v0_8

import (
	"github.com/cosmos/cosmos-sdk/codec"

	v038auth "github.com/cosmos/cosmos-sdk/x/auth"
	v036distr "github.com/cosmos/cosmos-sdk/x/distribution/legacy/v0_36"
	v038distr "github.com/cosmos/cosmos-sdk/x/distribution/legacy/v0_38"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	"github.com/cosmos/cosmos-sdk/x/upgrade"

	v038authcustom "github.com/kava-labs/kava/migrate/v0_8/sdk/auth/v0_38" // TODO alias?
	v18de63auth "github.com/kava-labs/kava/migrate/v0_8/sdk/auth/v18de63"
	v0_8auction "github.com/kava-labs/kava/x/auction/legacy/v0_8"
)

/*
Migrating between outdated versions:
Any current version doesn't need to be able to migrate state between old versions, install the old software to do that.
Just from previous release to current release.
Otherwise we'll need to eventually keep around almost an entire copy of all old types from previous versions (including dependencies of old versions like the codec)
Not a concern for now I think.

Testing:
- test migrate using types in module migration methods
- test migrate using bytes in this top level migrate
- write some sanity check scripts in python - check balances, vesting times are the same
*/

// Migrate translates a genesis file from kava v0.3.x format to kava v0.8.x format.
func Migrate(v0_3AppState genutil.AppMap) genutil.AppMap {

	// run sdk migrations for v0.36 to v0.38 (at least ones that apply)
	// custom auth
	// reuse distribution
	// custom slashing / evidence
	// custom staking
	// upgrade
	v0_8AppState := MigrateSDK(v0_3AppState) // just move into own function for clarity

	// run our migrations for new modules
	v0_8Codec := codec.New() // TODO what happens when the codec changes in sdk v0.39 ? do we need to vendor amino?
	codec.RegisterCrypto(v0_8Codec)
	// TODO need to register other things

	// auction
	v0_8AppState[v0_8auction.ModuleName] = v0_8Codec.MustMarshalJSON(v0_8auction.EmptyGenesisState())
	// TODO bep3, cdp, committee, incentive, kavadist, pricefeed, same as above

	return v0_8AppState
}

// migrate the sdk modules between 18de630d (v0.37 and half) and v0.38.3, mostly copying v038.Migrate
func MigrateSDK(appState genutil.AppMap) genutil.AppMap {
	// TODO copy appState to avoid mutation?

	// TODO setup codecs or pass in?
	// TODO also need 036 codec for distribution migration

	v18de63Codec := codec.New() // ideally this would use the exact version of amino from kava v0.3.5
	codec.RegisterCrypto(v18de63Codec)
	v18de63auth.RegisterCodec(v18de63Codec)

	v038Codec := codec.New()
	codec.RegisterCrypto(v038Codec)
	v038auth.RegisterCodec(v038Codec)

	// for each module, unmarshal old state, run a migrate(genesisStateType) func, marshal returned type into json and set

	// migrate auth state - different from sdk migration as we migrate from a middle format that never made it into a release
	if appState[v18de63auth.ModuleName] != nil {
		var authGenState v18de63auth.GenesisState
		v18de63Codec.MustUnmarshalJSON(appState[v18de63auth.ModuleName], &authGenState)

		delete(appState, v18de63auth.ModuleName)
		appState[v038auth.ModuleName] = v038Codec.MustMarshalJSON(v038authcustom.Migrate(authGenState))
	}

	// migrate distribution state - copied in from the sdk as these changes happened after 18de630d
	if appState[v036distr.ModuleName] != nil {
		var distrGenState v036distr.GenesisState
		v036Codec.MustUnmarshalJSON(appState[v036distr.ModuleName], &distrGenState)

		delete(appState, v036distr.ModuleName)                                                       // delete old key in case the name changed
		appState[v038distr.ModuleName] = v038Codec.MustMarshalJSON(v038distr.Migrate(distrGenState)) // needs to deal with module and vv accounts as well
		// TODO should probably deal with vv accounts in parent func
	}

	// slashing, evidence
	// TODO
	// if slashing, decode, call migrate (removes param) and marshal and set
	// 		call evidence.Migrate with slashing genState, then marshal and set

	// staking
	// TODO

	// upgrade
	appState[upgrade.ModuleName] = []byte(`{}`) // TODO should use a copy of ModuleName?

	return appState
}
