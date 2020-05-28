package simulation

import (
	"fmt"
	"math/rand"
	"strings"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authexported "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/simulation"
	"github.com/cosmos/cosmos-sdk/x/supply"

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
	MaxSupplyLimit   = sdk.NewInt(1000000000000)
	accs             []simulation.Account
	ConsistentDenoms = [3]string{"bnb", "xrp", "btc"}
)

// GenRandBnbDeputy randomized BnbDeputyAddress
func GenRandBnbDeputy(r *rand.Rand) simulation.Account {
	acc, _ := simulation.RandomAcc(r, accs)
	return acc
}

// GenRandBnbDeputyFixedFee randomized BnbDeputyFixedFee in range [2, 10000]
func GenRandBnbDeputyFixedFee(r *rand.Rand) uint64 {
	min := int(2)
	max := int(10000)
	return uint64(r.Intn(max-min) + min)
}

// GenMinBlockLock randomized MinBlockLock
func GenMinBlockLock(r *rand.Rand) uint64 {
	min := int(types.AbsoluteMinimumBlockLock)
	max := int(types.AbsoluteMaximumBlockLock)
	return uint64(r.Intn(max-min) + min)
}

// GenMaxBlockLock randomized MaxBlockLock
func GenMaxBlockLock(r *rand.Rand, minBlockLock uint64) uint64 {
	min := int(minBlockLock)
	max := int(types.AbsoluteMaximumBlockLock)
	return uint64(r.Intn(max-min) + min)
}

// GenSupportedAssets gets randomized SupportedAssets
func GenSupportedAssets(r *rand.Rand) types.AssetParams {

	numAssets := (r.Intn(10) + 1)
	assets := make(types.AssetParams, numAssets+1)
	for i := 0; i < numAssets; i++ {
		denom := strings.ToLower(simulation.RandStringOfLength(r, (r.Intn(3) + 3)))
		asset := genSupportedAsset(r, denom)
		assets[i] = asset
	}
	// Add bnb, btc, or xrp as a supported asset for interactions with other modules
	assets[len(assets)-1] = genSupportedAsset(r, ConsistentDenoms[r.Intn(3)])

	return assets
}

func genSupportedAsset(r *rand.Rand, denom string) types.AssetParam {
	coinID, _ := simulation.RandPositiveInt(r, sdk.NewInt(100000))
	limit, _ := simulation.RandPositiveInt(r, MaxSupplyLimit)
	return types.AssetParam{
		Denom:  denom,
		CoinID: int(coinID.Int64()),
		Limit:  limit,
		Active: true,
	}
}

// RandomizedGenState generates a random GenesisState
func RandomizedGenState(simState *module.SimulationState) {
	accs = simState.Accounts

	bep3Genesis := loadRandomBep3GenState(simState)
	fmt.Printf("Selected randomly generated %s parameters:\n%s\n", types.ModuleName, codec.MustMarshalJSONIndent(simState.Cdc, bep3Genesis))
	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(bep3Genesis)

	authGenesis, totalCoins := loadAuthGenState(simState, bep3Genesis)
	simState.GenState[auth.ModuleName] = simState.Cdc.MustMarshalJSON(authGenesis)

	// Update supply to match amount of coins in auth
	var supplyGenesis supply.GenesisState
	simState.Cdc.MustUnmarshalJSON(simState.GenState[supply.ModuleName], &supplyGenesis)

	for _, deputyCoin := range totalCoins {
		supplyGenesis.Supply = supplyGenesis.Supply.Add(deputyCoin...)
	}
	simState.GenState[supply.ModuleName] = simState.Cdc.MustMarshalJSON(supplyGenesis)
}

func loadRandomBep3GenState(simState *module.SimulationState) types.GenesisState {
	bnbDeputy := GenRandBnbDeputy(simState.Rand)
	bnbDeputyFixedFee := GenRandBnbDeputyFixedFee(simState.Rand)

	// min/max block lock are hardcoded to 50/100 for expected -NumBlocks=100
	minBlockLock := types.AbsoluteMinimumBlockLock
	maxBlockLock := minBlockLock * 2

	var supportedAssets types.AssetParams
	simState.AppParams.GetOrGenerate(
		simState.Cdc, SupportedAssets, &supportedAssets, simState.Rand,
		func(r *rand.Rand) { supportedAssets = GenSupportedAssets(r) },
	)

	bep3Genesis := types.GenesisState{
		Params: types.Params{
			BnbDeputyAddress:  bnbDeputy.Address,
			BnbDeputyFixedFee: bnbDeputyFixedFee,
			MinBlockLock:      minBlockLock,
			MaxBlockLock:      maxBlockLock,
			SupportedAssets:   supportedAssets,
		},
	}

	return bep3Genesis
}

func loadAuthGenState(simState *module.SimulationState, bep3Genesis types.GenesisState) (auth.GenesisState, []sdk.Coins) {
	var authGenesis auth.GenesisState
	simState.Cdc.MustUnmarshalJSON(simState.GenState[auth.ModuleName], &authGenesis)

	deputy, found := getAccount(authGenesis.Accounts, bep3Genesis.Params.BnbDeputyAddress)
	if !found {
		panic("deputy address not found in available accounts")
	}

	// Load total limit of each supported asset to deputy's account
	var totalCoins []sdk.Coins
	for _, asset := range bep3Genesis.Params.SupportedAssets {
		assetCoin := sdk.NewCoins(sdk.NewCoin(asset.Denom, asset.Limit))
		if err := deputy.SetCoins(deputy.GetCoins().Add(assetCoin...)); err != nil {
			panic(err)
		}
		totalCoins = append(totalCoins, assetCoin)
	}
	authGenesis.Accounts = replaceOrAppendAccount(authGenesis.Accounts, deputy)

	return authGenesis, totalCoins
}

// Return an account from a list of accounts that matches an address.
func getAccount(accounts []authexported.GenesisAccount, addr sdk.AccAddress) (authexported.GenesisAccount, bool) {
	for _, a := range accounts {
		if a.GetAddress().Equals(addr) {
			return a, true
		}
	}
	return nil, false
}

// In a list of accounts, replace the first account found with the same address. If not found, append the account.
func replaceOrAppendAccount(accounts []authexported.GenesisAccount, acc authexported.GenesisAccount) []authexported.GenesisAccount {
	newAccounts := accounts
	for i, a := range accounts {
		if a.GetAddress().Equals(acc.GetAddress()) {
			newAccounts[i] = acc
			return newAccounts
		}
	}
	return append(newAccounts, acc)
}
