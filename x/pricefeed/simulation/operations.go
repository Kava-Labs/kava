package simulation

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/simulation"

	"github.com/kava-labs/kava/x/pricefeed/keeper"
	"github.com/kava-labs/kava/x/pricefeed/types"
)

var (
	noOpMsg   = simulation.NoOpMsg(types.ModuleName)
	btcPrices = []sdk.Dec{}
	bnbPrices = []sdk.Dec{}
	xrpPrices = []sdk.Dec{}
	genPrices sync.Once
)

// SimulateMsgUpdatePrices updates the prices of various assets by randomly varying them based on current price
func SimulateMsgUpdatePrices(keeper keeper.Keeper, blocks int) simulation.Operation {
	// get a pricefeed handler
	// handler := pricefeed.NewHandler(keeper)

	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simulation.Account) (
		simulation.OperationMsg, []simulation.FutureOperation, error) {
		genPrices.Do(func() {
			// generate a random walk for each asset exactly once, with observations equal to the number of blocks in the sim
			for _, m := range keeper.GetMarkets(ctx) {
				startPrice := getStartPrice(m.MarketID)
				// allow prices to fluctuate from 10x GAINZ to 100x REKT
				maxPrice := sdk.MustNewDecFromStr("10.0").Mul(startPrice)
				minPrice := sdk.MustNewDecFromStr("0.01").Mul(startPrice)
				previousPrice := startPrice
				for i := 0; i < blocks; i++ {
					increment := getIncrement(m.MarketID)
					// note calling r instead of rand here breaks determinism
					upDown := rand.Intn(2)
					if upDown == 0 {
						if previousPrice.Add(increment).GT(maxPrice) {
							previousPrice = maxPrice
						} else {
							previousPrice = previousPrice.Add(increment)
						}
					} else {
						if previousPrice.Sub(increment).LT(minPrice) {
							previousPrice = minPrice
						} else {
							previousPrice = previousPrice.Sub(increment)
						}
					}
					setPrice(m.MarketID, previousPrice)
				}
			}
		})

		randomMarket := pickRandomAsset(ctx, keeper, r)
		marketID := randomMarket.MarketID
		address := getRandomOracle(r, randomMarket)
		price := pickNewRandomPrice(marketID, int(ctx.BlockHeight()))

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

func getIncrement(marketID string) (increment sdk.Dec) {
	startPrice := getStartPrice(marketID)
	divisor := sdk.MustNewDecFromStr("20")
	increment = startPrice.Quo(divisor)
	return increment
}

func setPrice(marketID string, price sdk.Dec) {
	switch marketID {
	case "btc:usd":
		btcPrices = append(btcPrices, price)
		return
	case "bnb:usd":
		bnbPrices = append(bnbPrices, price)
		return
	case "xrp:usd":
		xrpPrices = append(xrpPrices, price)
	}
	return
}

func pickNewRandomPrice(marketID string, blockHeight int) (newPrice sdk.Dec) {
	switch marketID {
	case "btc:usd":
		return btcPrices[blockHeight-1]
	case "bnb:usd":
		return bnbPrices[blockHeight-1]
	case "xrp:usd":
		return xrpPrices[blockHeight-1]
	}
	panic("invalid price request")
}

// getRandomOracle picks a random oracle from the list of oracles
func getRandomOracle(r *rand.Rand, market types.Market) sdk.AccAddress {
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

// submitMsg submits a message to the current instance of the keeper and returns a boolean whether the operation completed successfully or not
func submitMsg(ctx sdk.Context, handler sdk.Handler, msg sdk.Msg) (ok bool) {
	ctx, write := ctx.CacheContext()
	got, err := handler(ctx, msg)
	if err != nil {
		return err == nil
	}

	ok = got.IsOK()
	if ok {
		write()
	}
	return ok
}
