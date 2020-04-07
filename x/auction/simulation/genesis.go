package simulation

import (
	"fmt"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authexported "github.com/cosmos/cosmos-sdk/x/auth/exported"
	"github.com/cosmos/cosmos-sdk/x/supply"

	"github.com/kava-labs/kava/x/auction/types"
	cdptypes "github.com/kava-labs/kava/x/cdp/types"
)

// RandomizedGenState generates a random GenesisState for auction
func RandomizedGenState(simState *module.SimulationState) {

	// TODO implement this fully
	// - randomly generating the genesis params
	// - overwriting with genesis provided to simulation
	auctionGenesis := types.DefaultGenesisState()

	// FIXME temporarily add an auction
	a := types.NewDebtAuction(
		cdptypes.LiquidatorMacc, // using cdp account rather than generic test one to avoid having to set permissions on the supply keeper
		sdk.NewInt64Coin("usdx", 100),
		sdk.NewInt64Coin("ukava", 1000000000000),
		simState.GenTimestamp.Add(time.Hour*5),
		sdk.NewInt64Coin("debt", 100), // same as usdx
	) // ID is zero
	auctionGenesis.Auctions = append(auctionGenesis.Auctions, a)
	auctionGenesis.NextAuctionID = 0 + 1

	// Add auction mod account with debt
	// also add usdx coins to the bidder account so that they can actually place a bid
	var authGenesis auth.GenesisState
	simState.Cdc.MustUnmarshalJSON(simState.GenState[auth.ModuleName], &authGenesis)

	auctionModAcc := supply.NewEmptyModuleAccount(types.ModuleName)
	auctionModAcc.SetCoins(sdk.NewCoins(sdk.NewInt64Coin("debt", 100))) // same as sum of auctions.
	authGenesis.Accounts = append(authGenesis.Accounts, auctionModAcc)  // TODO check if it exists first

	bidder, found := getAccount(authGenesis.Accounts, simState.Accounts[0].Address) // 0 is the bidder // FIXME
	if !found {
		panic("bidder not found")
	}
	bidder.SetCoins(bidder.GetCoins().Add(sdk.NewCoins(sdk.NewInt64Coin("usdx", 10000000000))))
	authGenesis.Accounts = replaceAccount(authGenesis.Accounts, bidder)

	simState.GenState[auth.ModuleName] = simState.Cdc.MustMarshalJSON(authGenesis)

	// Also need to update supply's genesis state because everything's terrible
	var supplyGenesis supply.GenesisState
	simState.Cdc.MustUnmarshalJSON(simState.GenState[supply.ModuleName], &supplyGenesis)
	supplyGenesis.Supply = supplyGenesis.Supply.Add(sdk.NewCoins(sdk.NewInt64Coin("debt", 100), sdk.NewInt64Coin("usdx", 10000000000)))
	simState.GenState[supply.ModuleName] = simState.Cdc.MustMarshalJSON(supplyGenesis)

	// TODO liquidator mod account doesn't need to be initialized for this example
	// it just mints kava, doesn't need a balance
	// and supply.GetModuleAccount creates one if it doesn't exist

	// Note: this line prints out the auction genesis state, not just the auction parameters. Some sdk modules print out just the parameters.
	fmt.Printf("Selected randomly generated %s parameters:\n%s\n", types.ModuleName, codec.MustMarshalJSONIndent(simState.Cdc, auctionGenesis))
	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(auctionGenesis)
}

func getAccount(accounts []authexported.GenesisAccount, addr sdk.AccAddress) (authexported.GenesisAccount, bool) {
	for _, a := range accounts {
		if a.GetAddress().Equals(addr) {
			return a, true
		}
	}
	return nil, false
}

func replaceAccount(accounts []authexported.GenesisAccount, acc authexported.GenesisAccount) []authexported.GenesisAccount {
	newAccounts := accounts
	for i, a := range accounts {
		if a.GetAddress().Equals(acc.GetAddress()) {
			newAccounts[i] = acc
			return newAccounts
		}
	}
	return newAccounts
}
