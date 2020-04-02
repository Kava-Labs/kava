package simulation

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/simulation"
	"github.com/kava-labs/kava/x/bep3/types"
)

// Simulation parameter constants
const (
	BnbDeputyAddress = "bnb_deputy_address"
	MinBlockLock     = "min_block_lock"
	MaxBlockLock     = "max_block_lock"
	SupportedAssets  = "supported_assets"
)

var (
	MaxSupplyLimit  = sdk.NewInt(10000000000000000)
	BondedAddresses []sdk.AccAddress
)

// GenBnbDeputyAddress randomized BnbDeputyAddress
func GenBnbDeputyAddress(r *rand.Rand) sdk.AccAddress {
	return BondedAddresses[r.Intn(len(BondedAddresses))]
}

// GenMinBlockLock randomized MinBlockLock
func GenMinBlockLock(r *rand.Rand) int64 {
	min := int(types.AbsoluteMinimumBlockLock)
	max := int(types.AbsoluteMaximumBlockLock)
	return int64(r.Intn(max-min) + min)
}

// GenMaxBlockLock randomized MaxBlockLock
func GenMaxBlockLock(r *rand.Rand, minBlockLock int64) int64 {
	min := int(minBlockLock)
	max := int(types.AbsoluteMaximumBlockLock)
	return int64(r.Intn(max-min) + min)
}

// GenSupportedAssets gets randomized SupportedAssets
func GenSupportedAssets(r *rand.Rand) types.AssetParams {
	var assets types.AssetParams
	for i := 0; i < (r.Intn(10) + 1); i++ {
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		asset := genSupportedAsset(r)
		assets = append(assets, asset)
	}
	return assets
}

func genSupportedAsset(r *rand.Rand) types.AssetParam {
	denom := strings.ToLower(simulation.RandStringOfLength(r, (r.Intn(3) + 3)))
	coinID, _ := simulation.RandPositiveInt(r, sdk.NewInt(100000))
	limit, _ := simulation.RandPositiveInt(r, MaxSupplyLimit)
	active := func() bool {
		if r.Int()%2 == 0 {
			return true
		}
		return false
	}
	return types.AssetParam{
		Denom:  denom,
		CoinID: int(coinID.Int64()),
		Limit:  limit,
		Active: active(),
	}
}

// RandomizedGenState generates a random GenesisState
func RandomizedGenState(simState *module.SimulationState) {
	BondedAddresses = loadBondedAddresses(simState)

	var bnbDeputyAddress sdk.AccAddress
	simState.AppParams.GetOrGenerate(
		simState.Cdc, BnbDeputyAddress, &bnbDeputyAddress, simState.Rand,
		func(r *rand.Rand) { bnbDeputyAddress = GenBnbDeputyAddress(r) },
	)

	var minBlockLock int64
	simState.AppParams.GetOrGenerate(
		simState.Cdc, MinBlockLock, &minBlockLock, simState.Rand,
		func(r *rand.Rand) { minBlockLock = GenMinBlockLock(r) },
	)

	var maxBlockLock int64
	simState.AppParams.GetOrGenerate(
		simState.Cdc, MaxBlockLock, &maxBlockLock, simState.Rand,
		func(r *rand.Rand) { maxBlockLock = GenMaxBlockLock(r, minBlockLock) },
	)

	var supportedAssets types.AssetParams
	simState.AppParams.GetOrGenerate(
		simState.Cdc, SupportedAssets, &supportedAssets, simState.Rand,
		func(r *rand.Rand) { supportedAssets = GenSupportedAssets(r) },
	)

	bep3Genesis := types.GenesisState{
		Params: types.Params{
			BnbDeputyAddress: bnbDeputyAddress,
			MinBlockLock:     minBlockLock,
			MaxBlockLock:     maxBlockLock,
			SupportedAssets:  supportedAssets,
		},
	}

	fmt.Printf("Selected randomly generated %s parameters:\n%s\n", types.ModuleName, codec.MustMarshalJSONIndent(simState.Cdc, bep3Genesis))
	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(bep3Genesis)
}

// loadBondedAddresses loads an array of bonded account addresses
func loadBondedAddresses(simState *module.SimulationState) []sdk.AccAddress {
	addrs := make([]sdk.AccAddress, simState.NumBonded)
	for i := 0; i < int(simState.NumBonded); i++ {
		addr := simState.Accounts[i].Address
		addrs[i] = addr
	}
	return addrs
}
