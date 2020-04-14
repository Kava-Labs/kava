package operations

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/simulation"

	"github.com/kava-labs/kava/x/pricefeed"
	"github.com/kava-labs/kava/x/pricefeed/keeper"
	"github.com/kava-labs/kava/x/pricefeed/types"
)

var (
	noOpMsg = simulation.NoOpMsg(pricefeed.ModuleName)
)

// SimulateMsgUpdatePrices updates the prices of various assets by randomly varying them based on current price
func SimulateMsgUpdatePrices(keeper keeper.Keeper) simulation.Operation {
	// get a pricefeed handler
	handler := pricefeed.NewHandler(keeper)

	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simulation.Account) (
		simulation.OperationMsg, []simulation.FutureOperation, error) {

		// OVERALL LOGIC:
		// (1) RANDOMLY PICK AN ASSET OUT OF BNB AN BTC [TODO QUESTION - USDX IS EXCLUDED AS IT IS A STABLE DENOM
		// (2) GET THE CURRENT PRICE OF THAT ASSET IN USD
		// (3) GENERATE A RANDOM NUMBER IN THE RANGE 0.8-1.2 (UNIFORM DISTRIBUTION)
		// (4) MULTIPLY THE CURRENT PRICE BY THE RANDOM NUMBER
		// (5) POST THE NEW PRICE TO THE KEEPER

		// pick a random asset out of BNB and BTC
		randomMarket := pickRandomAsset(ctx, keeper, r)

		marketID := randomMarket.MarketID

		// Get the current price of the asset
		currentPrice, err := keeper.GetCurrentPrice(ctx, marketID) //  Note this is marketID AND **NOT** just the base asset
		if err != nil {
			return noOpMsg, nil, fmt.Errorf("Error getting current price")
		}

		// get the genesis price for an asset
		genesisPrice, err := keeper.GetCurrentPrice(ctx.WithBlockHeight(1), marketID)
		if err != nil {
			return noOpMsg, nil, fmt.Errorf("Error getting genesis price")
		}

		// get the address for the account
		// this address needs to be an oracle and also exist. genesis should add all the accounts as oracles.
		address := getRandomOracle(r, randomMarket)

		// generate a new random price based off the current price
		price, err := pickNewRandomPrice(r, currentPrice.Price, genesisPrice.Price)
		if err != nil {
			return noOpMsg, nil, fmt.Errorf("Error picking random price")
		}

		// get the expiry time based off the current time
		expiry := getExpiryTime(ctx)

		// now create the msg to post price
		msg := types.NewMsgPostPrice(address, marketID, price, expiry)

		// Perform basic validation of the msg - don't submit errors that fail ValidateBasic, use unit tests for testing ValidateBasic
		if err := msg.ValidateBasic(); err != nil {
			return noOpMsg, nil, fmt.Errorf("expected msg to pass ValidateBasic: %s", msg.GetSignBytes())
		}

		// now we submit the pricefeed update message
		if ok := submitMsg(ctx, handler, msg); !ok {
			return noOpMsg, nil, fmt.Errorf("could not submit pricefeed msg")
		}
		return simulation.NewOperationMsg(msg, true, "pricefeed update submitted"), nil, nil
	}
}

// getRandomOracle picks a random oracle from the list of oracles
func getRandomOracle(r *rand.Rand, market pricefeed.Market) sdk.AccAddress {
	randomIndex := simulation.RandIntBetween(r, 0, len(market.Oracles))
	oracle := market.Oracles[randomIndex]
	return oracle
}

// pickRandomAsset picks a random asset out of the assets with equal probability
// it returns the Market which includes the base asset as one of its fields
func pickRandomAsset(ctx sdk.Context, keeper keeper.Keeper, r *rand.Rand) (market types.Market) {
	// get the params
	params := keeper.GetParams(ctx)
	// now pick a random asset
	randomIndex := simulation.RandIntBetween(r, 0, len(params.Markets))
	market = params.Markets[randomIndex]
	return market
}

// getExpiryTime gets a price expiry time by taking the current time and adding a delta to it
func getExpiryTime(ctx sdk.Context) (t time.Time) {
	// need to use the blocktime from the context as the context generates random start time when running simulations
	t = ctx.BlockTime().Add(time.Second * 1000000)
	return t
}

// pickNewRandomPrice picks a new random price given the current price
// It takes the current price then generates a random number to multiply it by to create variation while
// still being in the similar range. Random walk style.
// originalPrice is the starting price for the asset
func pickNewRandomPrice(r *rand.Rand, currentPrice sdk.Dec, originalPrice sdk.Dec) (newPrice sdk.Dec, err sdk.Error) {
	// Pick random price multiplier
	limit := sdk.MustNewDecFromStr("0.2")
	got := simulation.RandomDecAmount(r, limit) // random in the range 0-0.2

	if err != nil {
		fmt.Errorf("error generating random price multiplier")
		return sdk.ZeroDec(), err
	}
	// random amount to deviate by
	deviationAmount := got.Mul(originalPrice)

	// now flip a coin
	upDown := r.Intn(2)
	// upDown := simulation.RandIntBetween(r, 0, 1) // WARNING THIS IS BIASED TOWARDS ZERO DO NOT USE

	// either add or subtract the deviation amount with random probability
	if upDown == 0 {
		if currentPrice.Sub(deviationAmount).LTE(sdk.ZeroDec()) {
			newPrice = sdk.ZeroDec()
		} else {
			newPrice = currentPrice.Sub(deviationAmount)
		}
	} else {
		newPrice = currentPrice.Add(deviationAmount)
	}

	// return the price
	return newPrice, nil
}

// submitMsg submits a message to the current instance of the keeper and returns a boolean whether the operation completed successfully or not
func submitMsg(ctx sdk.Context, handler sdk.Handler, msg sdk.Msg) (ok bool) {
	ctx, write := ctx.CacheContext()
	got := handler(ctx, msg)

	ok = got.IsOK()
	if ok {
		write()
	}
	return ok
}
