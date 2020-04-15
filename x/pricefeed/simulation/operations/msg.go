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
	noOpMsg  = simulation.NoOpMsg(pricefeed.ModuleName)
	btcPrice = sdk.MustNewDecFromStr("7000")
	bnbPrice = sdk.MustNewDecFromStr("15")
	xrpPrice = sdk.MustNewDecFromStr("0.25")
)

func getStartPrice(marketID string) (startPrice sdk.Dec) {
	switch marketID {
	case "btc:usd":
		return sdk.MustNewDecFromStr("7000")
	case "bnb:usd":
		return sdk.MustNewDecFromStr("15")
	case "xrp:usd":
		return sdk.MustNewDecFromStr("0.25")
	}
	return sdk.MustNewDecFromStr("100")
}

func getRecentPrice(marketID string) (prevPrice sdk.Dec) {
	switch marketID {
	case "btc:usd":
		return btcPrice
	case "bnb:usd":
		return bnbPrice
	case "xrp:usd":
		return xrpPrice
	}
	return sdk.MustNewDecFromStr("100")
}

func setRecentPrice(marketID string, newPrice sdk.Dec) {
	switch marketID {
	case "btc:usd":
		btcPrice = newPrice
		return
	case "bnb:usd":
		bnbPrice = newPrice
		return
	case "xrp:usd":
		xrpPrice = newPrice
		return
	}
}

func getIncrement(marketID string) (increment sdk.Dec) {
	startPrice := getStartPrice(marketID)
	divisor := sdk.MustNewDecFromStr("20")
	increment = startPrice.Quo(divisor)
	return increment
}

// SimulateMsgUpdatePrices updates the prices of various assets by randomly varying them based on current price
func SimulateMsgUpdatePrices(keeper keeper.Keeper) simulation.Operation {
	// get a pricefeed handler
	handler := pricefeed.NewHandler(keeper)

	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simulation.Account) (
		simulation.OperationMsg, []simulation.FutureOperation, error) {

		randomMarket := pickRandomAsset(ctx, keeper, r)
		marketID := randomMarket.MarketID
		address := getRandomOracle(r, randomMarket)
		price := pickNewRandomPrice(r, marketID)

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

func pickNewRandomPrice(r *rand.Rand, marketID string) (newPrice sdk.Dec) {
	startPrice := getStartPrice(marketID)
	maxPrice := sdk.MustNewDecFromStr("10.0").Mul(startPrice)
	recentPrice := getRecentPrice(marketID)
	increment := getIncrement(marketID)

	upDown := r.Intn(2)
	if upDown == 0 {
		if maxPrice.Sub(increment).LTE(recentPrice) {
			newPrice = maxPrice
		} else {
			newPrice = recentPrice.Add(increment)
		}
	} else {
		if sdk.SmallestDec().Add(increment).GTE(recentPrice) {
			newPrice = sdk.SmallestDec()
		} else {
			newPrice = recentPrice.Sub(increment)
		}
	}
	setRecentPrice(marketID, newPrice)

	return newPrice
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
