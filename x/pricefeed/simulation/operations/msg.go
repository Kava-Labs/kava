package operations

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/simulation"

	"github.com/kava-labs/kava/x/pricefeed"
	"github.com/kava-labs/kava/x/pricefeed/types"
)

var (
	noOpMsg     = simulation.NoOpMsg(pricefeed.ModuleName) // TODO QUESTION what is the purpose of noopmsg?
	govDenom    = "ukava"                                  // TODO QUESTION do I need these in pricefeed?
	stableDenom = "usdx"
)

// SimulateMsgUpdatePrices updates the prices of various assets
// TODO - MUST include BNB, USDX, BTC as HARD REQUIREMENTS - asset prices needed by CDP sims
func SimulateMsgUpdatePrices(authKeeper auth.AccountKeeper, pfk pricefeed.Keeper) simulation.Operation {
	// get a pricefeed handler
	handler := pricefeed.NewHandler(pfk)

	// There's two error paths
	// - return a OpMessage, but nil error - this will log a message but keep running the simulation
	// - return an error - this will stop the simulation (I think)
	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simulation.Account) (
		simulation.OperationMsg, []simulation.FutureOperation, error) {

		// OVERALL LOGIC IDEA:
		// (1) RANDOMLY PICK AN ASSET OUT OF BNB AN BTC [TODO QUESTION - PRESUMABLY USDX IS EXCLUDED AS IT IS A STABLE DENOM?]
		// (2) GET THE CURRENT PRICE OF THAT ASSET IN USD
		// (3) GENERATE A RANDOM NUMBER IN THE RANGE 0.8-1.2 (UNIFORM DISTRIBUTION) [TODO POTENTIAL UPDATE MAKE IT NORMALLY DISTRIBUTED ACROSS A WIDER RANGE?]
		// (4) MULTIPLY THE CURRENT PRICE BY THE RANDOM NUMBER
		// (5) POST THE NEW PRICE TO THE KEEPER

		// pick a random asset out of BNB and BTC
		market := pickRandomMarket(ctx, pfk, r)

		fmt.Printf("Picked market: %s\n", market.MarketID)

		// TODO QUESTION - GET THE CURRENT PRICE OF THE ASSET?? HOW TO DO THIS?
		currentPrice, err := pfk.GetCurrentPrice(ctx, market.MarketID)
		if err != nil {
			return noOpMsg, nil, fmt.Errorf("Error getting current price: %v", err)
		}

		oracle := getRandomOracle(r, market)

		// generate a new random price based off the current price
		price := pickNewRandomPrice(r, currentPrice.Price)

		// get the expiry time based off the current time
		expiry := getExpiryTime(ctx)

		// MSG POST PRICE
		// GENERATE THE MSG TO SEND TO THE KEEPER
		// now create the msg to post price
		msg := types.NewMsgPostPrice(oracle, market.MarketID, price, expiry)

		// Perform basic validation of the msg - don't submit errors that fail ValidateBasic, use unit tests for testing ValidateBasic
		if err := msg.ValidateBasic(); err != nil {
			return noOpMsg, nil, fmt.Errorf("expected msg to pass ValidateBasic: %v", err)
		}

		if ok := submitMsg(ctx, handler, msg); !ok {
			return noOpMsg, nil, fmt.Errorf("could not submit pricefeed msg")
		}
		return simulation.NewOperationMsg(msg, true, "pricefeed update price"), nil, nil
	}
}

// pickRandomMarket picks a random asset from params
func pickRandomMarket(ctx sdk.Context, pfk pricefeed.Keeper, r *rand.Rand) pricefeed.Market {
	// randomly pick an asset
	markets := pfk.GetMarkets(ctx)
	return markets[simulation.RandIntBetween(r, 0, len(markets))]
}

// getExpiryTime gets a price expiry time by taking the current time and adding a delta to it
func getExpiryTime(ctx sdk.Context) (t time.Time) {
	t = ctx.BlockTime().Add(time.Hour * 24)
	return t
}

// pickNewRandomPrice picks a new random price given the current price
// It takes the current price then generates a random number to multiply it by to create variation while
// still being in the similar range. Random walk style.
func pickNewRandomPrice(r *rand.Rand, currentPrice sdk.Dec) (price sdk.Dec) {
	// Pick random price
	randomNoise := sdk.MustNewDecFromStr("0.4")

	randomPriceMultiplier := simulation.RandomDecAmount(r, randomNoise) // get a random number

	offset := sdk.MustNewDecFromStr("0.8")

	randomPriceMultiplier = randomPriceMultiplier.Add(offset) // gives a result in range 0.8-1.2 inclusive

	// MULTIPLY CURRENT PRICE BY RANDOM PRICE MULTIPLER
	price = randomPriceMultiplier.Mul(currentPrice)
	// return the price
	return price
}

func getRandomOracle(r *rand.Rand, market pricefeed.Market) sdk.AccAddress {
	fmt.Printf("picking random oracle for %v\n", market)
	randn := simulation.RandIntBetween(r, 0, len(market.Oracles))
	oracle := market.Oracles[randn]
	return oracle
}

// submitMsg submits a message to the current instance of the keeper and returns a boolean whether the operation completed successfully or not
func submitMsg(ctx sdk.Context, handler sdk.Handler, msg sdk.Msg) (ok bool) {
	ctx, write := ctx.CacheContext()
	ok = handler(ctx, msg).IsOK()
	if ok {
		write()
	} else {
		log := handler(ctx, msg).Log
		fmt.Printf("Error when submitting msg to handler: %v\n", log)
	}
	return ok
}
