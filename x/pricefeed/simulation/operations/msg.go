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

	tmtime "github.com/tendermint/tendermint/types/time"
)

var (
	noOpMsg     = simulation.NoOpMsg(pricefeed.ModuleName) // TODO QUESTION what is the purpose of noopmsg?
	stableDenom = "usdx"
)

// SimulateMsgUpdatePrices updates the prices of various assets
// TODO - MUST include BNB, USDX, BTC as HARD REQUIREMENTS - asset prices needed by CDP sims
func SimulateMsgUpdatePrices(keeper keeper.Keeper) simulation.Operation {
	// get a pricefeed handler
	handler := pricefeed.NewHandler(keeper)

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

		// get the address for the account
		// TODO QUESTION do we need to check if this address exists on chain or is an oracle for the price or does it not matter
		address := sdk.AccAddress("kava1ffv7nhd3z6sych2qpqkk03ec6hzkmufy0r2s4c")

		// pick a random asset out of BNB and BTC
		assetCode, err := pickRandomAsset(r)
		if err != nil {
			return noOpMsg, nil, fmt.Errorf("Error picking random asset")
		}

		fmt.Printf("Picked asset: %s\n", assetCode)

		// TODO QUESTION - GET THE CURRENT PRICE OF THE ASSET?? HOW TO DO THIS?
		currentPrice := sdk.MustNewDecFromStr("100")

		// generate a new random price based off the current price
		price, err := pickNewRandomPrice(r, currentPrice)
		if err != nil {
			return noOpMsg, nil, fmt.Errorf("Error picking random price")
		}

		// get the expiry time based off the current time
		expiry := getExpiryTime()

		// MSG POST PRICE
		// GENERATE THE MSG TO SEND TO THE KEEPER
		// now create the msg to post price
		msg := types.NewMsgPostPrice(address, assetCode, price, expiry)

		// Perform basic validation of the msg - don't submit errors that fail ValidateBasic, use unit tests for testing ValidateBasic
		if err := msg.ValidateBasic(); err != nil {
			return noOpMsg, nil, fmt.Errorf("expected msg to pass ValidateBasic: %s", msg.GetSignBytes())
		}

		// now we submit the pricefeed update message
		// TODO QUESTION - this is failing for some reason? Any ideas why?
		if ok := submitMsg(ctx, handler, msg); !ok {
			return noOpMsg, nil, fmt.Errorf("could not submit pricefeed msg")
		}
		fmt.Println("Pricefeed update sumbitted!")
		return simulation.NewOperationMsg(msg, true, "pricefeed update submitted"), nil, nil // TODO what should go in comment field?
	}
}

// pickRandomAsset picks a random asset out of BNB and BTC with equal probability
func pickRandomAsset(r *rand.Rand) (assetCode string, err error) {
	// randomly pick an asset
	randomNum, err := simulation.RandPositiveInt(r, sdk.NewInt(100)) // get a random number
	if err != nil {
		return "", err
	}
	// pick each asset with 50% probability
	if randomNum.LTE(sdk.NewInt(50)) {
		assetCode = "BNB"
	} else {
		assetCode = "BTC"
	}
	return assetCode, nil
}

// getExpiryTime gets a price expiry time by taking the current time and adding a delta to it
func getExpiryTime() (t time.Time) {
	t = tmtime.Now().Add(time.Second * 100000)
	return t
}

// pickNewRandomPrice picks a new random price given the current price
// It takes the current price then generates a random number to multiply it by to create variation while
// still being in the similar range. Random walk style.
func pickNewRandomPrice(r *rand.Rand, currentPrice sdk.Dec) (price sdk.Dec, err sdk.Error) {
	// Pick random price
	got := sdk.MustNewDecFromStr("0.4")

	randomPriceMultiplier := simulation.RandomDecAmount(r, got) // get a random number
	if err != nil {
		return sdk.ZeroDec(), err
		// return noOpMsg, nil, fmt.Errorf("Error picking random price")
	}

	offset := sdk.MustNewDecFromStr("0.8")

	randomPriceMultiplier = randomPriceMultiplier.Add(offset) // gives a result in range 0.8-1.2 inclusive

	// MULTIPLY CURRENT PRICE BY RANDOM PRICE MULTIPLER
	price = randomPriceMultiplier.Mul(sdk.NewDec(int64(currentPrice)))
	// return the price
	return price, nil
}

// submitMsg submits a message to the current instance of the keeper and returns a boolean whether the operation completed successfully or not
func submitMsg(ctx sdk.Context, handler sdk.Handler, msg sdk.Msg) (ok bool) {
	ctx, write := ctx.CacheContext()
	ok = handler(ctx, msg).IsOK()
	if ok {
		write()
	}
	return ok
}
