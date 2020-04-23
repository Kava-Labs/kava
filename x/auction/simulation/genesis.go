package simulation

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authexported "github.com/cosmos/cosmos-sdk/x/auth/exported"
	"github.com/cosmos/cosmos-sdk/x/simulation"
	"github.com/cosmos/cosmos-sdk/x/supply"

	"github.com/kava-labs/kava/x/auction/types"
	cdptypes "github.com/kava-labs/kava/x/cdp/types"
)

const (
	// Block time params are un-exported constants in cosmos-sdk/x/simulation.
	// Copy them here in lieu of importing them.
	minTimePerBlock time.Duration = (10000 / 2) * time.Second
	maxTimePerBlock time.Duration = 10000 * time.Second

	// Calculate the average block time
	AverageBlockTime time.Duration = (maxTimePerBlock - minTimePerBlock) / 2
	// MaxBidDuration is a crude way of ensuring that BidDuration ≤ MaxAuctionDuration for all generated params
	MaxBidDuration time.Duration = AverageBlockTime * 50
)

func GenBidDuration(r *rand.Rand) time.Duration {
	d, err := RandomPositiveDuration(r, 0, MaxBidDuration)
	if err != nil {
		panic(err)
	}
	return d
}
func GenMaxAuctionDuration(r *rand.Rand) time.Duration {
	d, err := RandomPositiveDuration(r, MaxBidDuration, AverageBlockTime*200)
	if err != nil {
		panic(err)
	}
	return d
}

func GenIncrementCollateral(r *rand.Rand) sdk.Dec {
	return simulation.RandomDecAmount(r, sdk.MustNewDecFromStr("1"))
}

var GenIncrementDebt = GenIncrementCollateral
var GenIncrementSurplus = GenIncrementCollateral

// RandomizedGenState generates a random GenesisState for auction
func RandomizedGenState(simState *module.SimulationState) {

	p := types.NewParams(
		GenMaxAuctionDuration(simState.Rand),
		GenBidDuration(simState.Rand),
		GenIncrementSurplus(simState.Rand),
		GenIncrementDebt(simState.Rand),
		GenIncrementCollateral(simState.Rand),
	)
	if err := p.Validate(); err != nil {
		panic(err)
	}
	auctionGenesis := types.NewGenesisState(
		types.DefaultNextAuctionID,
		p,
		nil,
	)

	// Add auctions
	auctions := types.GenesisAuctions{
		types.NewDebtAuction(
			cdptypes.LiquidatorMacc, // using cdp account rather than generic test one to avoid having to set permissions on the supply keeper
			sdk.NewInt64Coin("usdx", 100),
			sdk.NewInt64Coin("ukava", 1000000000000),
			simState.GenTimestamp.Add(time.Hour*5),
			sdk.NewInt64Coin("debt", 100), // same as usdx
		),
	}
	var startingID = auctionGenesis.NextAuctionID
	var ok bool
	var totalAuctionCoins sdk.Coins
	for i, a := range auctions {
		auctions[i], ok = a.WithID(uint64(i) + startingID).(types.GenesisAuction)
		if !ok {
			panic("can't convert Auction to GenesisAuction")
		}
		totalAuctionCoins = totalAuctionCoins.Add(a.GetModuleAccountCoins()...)
	}
	auctionGenesis.NextAuctionID = startingID + uint64(len(auctions))
	auctionGenesis.Auctions = append(auctionGenesis.Auctions, auctions...)

	// Also need to update the auction module account (to reflect the coins held in the auctions)
	var authGenesis auth.GenesisState
	simState.Cdc.MustUnmarshalJSON(simState.GenState[auth.ModuleName], &authGenesis)

	auctionModAcc, found := getAccount(authGenesis.Accounts, supply.NewModuleAddress(types.ModuleName))
	if !found {
		auctionModAcc = supply.NewEmptyModuleAccount(types.ModuleName)
	}
	if err := auctionModAcc.SetCoins(totalAuctionCoins); err != nil {
		panic(err)
	}
	authGenesis.Accounts = replaceOrAppendAccount(authGenesis.Accounts, auctionModAcc)

	// TODO adding bidder coins as well - this should be moved elsewhere
	bidder, found := getAccount(authGenesis.Accounts, simState.Accounts[0].Address) // 0 is the bidder // FIXME
	if !found {
		panic("bidder not found")
	}
	bidderCoins := sdk.NewCoins(sdk.NewInt64Coin("usdx", 10000000000))
	if err := bidder.SetCoins(bidder.GetCoins().Add(bidderCoins...)); err != nil {
		panic(err)
	}
	authGenesis.Accounts = replaceOrAppendAccount(authGenesis.Accounts, bidder)

	simState.GenState[auth.ModuleName] = simState.Cdc.MustMarshalJSON(authGenesis)

	// Update the supply genesis state to reflect the new coins
	// TODO find some way for this to happen automatically / move it elsewhere
	var supplyGenesis supply.GenesisState
	simState.Cdc.MustUnmarshalJSON(simState.GenState[supply.ModuleName], &supplyGenesis)
	supplyGenesis.Supply = supplyGenesis.Supply.Add(totalAuctionCoins...).Add(bidderCoins...)
	simState.GenState[supply.ModuleName] = simState.Cdc.MustMarshalJSON(supplyGenesis)

	// TODO liquidator mod account doesn't need to be initialized for this example
	// - it just mints kava, doesn't need a starting balance
	// - and supply.GetModuleAccount creates one if it doesn't exist

	// Note: this line prints out the auction genesis state, not just the auction parameters. Some sdk modules print out just the parameters.
	fmt.Printf("Selected randomly generated %s parameters:\n%s\n", types.ModuleName, codec.MustMarshalJSONIndent(simState.Cdc, auctionGenesis))
	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(auctionGenesis)
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

func RandomPositiveDuration(r *rand.Rand, inclusiveMin, exclusiveMax time.Duration) (time.Duration, error) {
	min := int64(inclusiveMin)
	max := int64(exclusiveMax)
	if min < 0 || max < 0 {
		return 0, fmt.Errorf("min and max must be positive")
	}
	if min >= max {
		return 0, fmt.Errorf("max must be < min")
	}
	randPositiveInt64 := r.Int63n(max-min) + min
	return time.Duration(randPositiveInt64), nil
}
