package simulation

import (
	"math/rand"
	"sort"
	"time"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/simapp/helpers"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/simulation"

	appparams "github.com/kava-labs/kava/app/params"
	"github.com/kava-labs/kava/x/pricefeed/keeper"
	"github.com/kava-labs/kava/x/pricefeed/types"
)

// Simulation operation weights constants
const (
	OpWeightMsgUpdatePrices = "op_weight_msg_update_prices"
)

// WeightedOperations returns all the operations from the module with their respective weights
func WeightedOperations(
	appParams simulation.AppParams, cdc *codec.Codec, ak auth.AccountKeeper, k keeper.Keeper,
) simulation.WeightedOperations {
	var weightMsgUpdatePrices int
	// var numBlocks int

	appParams.GetOrGenerate(cdc, OpWeightMsgUpdatePrices, &weightMsgUpdatePrices, nil,
		func(_ *rand.Rand) {
			weightMsgUpdatePrices = appparams.DefaultWeightMsgUpdatePrices
		},
	)

	return simulation.WeightedOperations{
		simulation.NewWeightedOperation(
			weightMsgUpdatePrices,
			SimulateMsgUpdatePrices(ak, k, 10000),
		),
	}
}

// SimulateMsgUpdatePrices updates the prices of various assets by randomly varying them based on current price
func SimulateMsgUpdatePrices(ak auth.AccountKeeper, keeper keeper.Keeper, blocks int) simulation.Operation {
	// runs one at the start of each simulation
	startingPrices := map[string]sdk.Dec{
		"btc:usd": sdk.MustNewDecFromStr("7000"),
		"bnb:usd": sdk.MustNewDecFromStr("15"),
		"xrp:usd": sdk.MustNewDecFromStr("0.25"),
	}

	// creates the new price generator from starting prices - resets for each sim
	priceGenerator := NewPriceGenerator(startingPrices)

	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simulation.Account, chainID string,
	) (simulation.OperationMsg, []simulation.FutureOperation, error) {
		// walk prices to current block height, noop if already called for current height
		priceGenerator.Step(r, ctx.BlockHeight())

		randomMarket := pickRandomAsset(ctx, keeper, r)
		marketID := randomMarket.MarketID
		address := getRandomOracle(r, randomMarket)

		oracle, found := simulation.FindAccount(accs, address)
		if !found {
			return simulation.NoOpMsg(types.ModuleName), nil, nil
		}

		oracleAcc := ak.GetAccount(ctx, oracle.Address)
		if oracleAcc == nil {
			return simulation.NoOpMsg(types.ModuleName), nil, nil
		}

		// get price for marketID and current block height set in Step
		price := priceGenerator.GetCurrentPrice(marketID)

		// get the expiry time based off the current time
		expiry := getExpiryTime(ctx)

		// now create the msg to post price
		msg := types.NewMsgPostPrice(oracle.Address, marketID, price, expiry)

		spendable := oracleAcc.SpendableCoins(ctx.BlockTime())
		fees, err := simulation.RandomFees(r, ctx, spendable)
		if err != nil {
			return simulation.NoOpMsg(types.ModuleName), nil, err
		}

		tx := helpers.GenTx(
			[]sdk.Msg{msg},
			fees,
			helpers.DefaultGenTxGas,
			chainID,
			[]uint64{oracleAcc.GetAccountNumber()},
			[]uint64{oracleAcc.GetSequence()},
			oracle.PrivKey,
		)

		_, result, err := app.Deliver(tx)
		if err != nil {
			return simulation.NoOpMsg(types.ModuleName), nil, err
		}
		return simulation.NewOperationMsg(msg, true, result.Log), nil, nil
	}
}

// getRandomOracle picks a random oracle from the list of oracles
func getRandomOracle(r *rand.Rand, market types.Market) sdk.AccAddress {
	randomIndex := simulation.RandIntBetween(r, 0, len(market.Oracles))
	return market.Oracles[randomIndex]
}

// pickRandomAsset picks a random asset out of the assets with equal probability
// it returns the Market which includes the base asset as one of its fields
func pickRandomAsset(ctx sdk.Context, keeper keeper.Keeper, r *rand.Rand) (market types.Market) {
	// get the params
	params := keeper.GetParams(ctx)
	// now pick a random asset
	randomIndex := simulation.RandIntBetween(r, 0, len(params.Markets))
	return params.Markets[randomIndex]
}

// getExpiryTime gets a price expiry time by taking the current time and adding a delta to it
func getExpiryTime(ctx sdk.Context) (t time.Time) {
	// need to use the blocktime from the context as the context generates random start time when running simulations
	return ctx.BlockTime().Add(time.Second * 1000000)
}

// PriceGenerator allows deterministic price generation in simulations
type PriceGenerator struct {
	markets            []string
	currentPrice       map[string]sdk.Dec
	maxPrice           map[string]sdk.Dec
	minPrice           map[string]sdk.Dec
	increment          map[string]sdk.Dec
	currentBlockHeight int64
}

// NewPriceGenerator returns a new market price generator from starting values
func NewPriceGenerator(startingPrice map[string]sdk.Dec) *PriceGenerator {
	p := &PriceGenerator{
		markets:            []string{},
		currentPrice:       startingPrice,
		maxPrice:           map[string]sdk.Dec{},
		minPrice:           map[string]sdk.Dec{},
		increment:          map[string]sdk.Dec{},
		currentBlockHeight: 0,
	}

	divisor := sdk.MustNewDecFromStr("20")

	for marketID, startPrice := range startingPrice {
		p.markets = append(p.markets, marketID)
		// allow 10x price increase
		p.maxPrice[marketID] = sdk.MustNewDecFromStr("10.0").Mul(startPrice)
		// allow 100x price decrease
		p.minPrice[marketID] = sdk.MustNewDecFromStr("0.01").Mul(startPrice)
		// set increment - should we use a random increment?
		p.increment[marketID] = startPrice.Quo(divisor)
	}

	// market prices must be calculated in a deterministic order
	// this sort order defines the the order we update each market
	// price in the step function
	sort.Strings(p.markets)

	return p
}

// Step walks prices to a current block height from the previously called height
// noop if called more than once for the same height
func (p *PriceGenerator) Step(r *rand.Rand, blockHeight int64) {
	if p.currentBlockHeight == blockHeight {
		// step already called for blockHeight
		return
	}

	if p.currentBlockHeight > blockHeight {
		// step is called with a previous blockHeight
		panic("step out of order")
	}

	for _, marketID := range p.markets {
		lastPrice := p.currentPrice[marketID]
		minPrice := p.minPrice[marketID]
		maxPrice := p.maxPrice[marketID]
		increment := p.increment[marketID]
		lastHeight := p.currentBlockHeight

		for lastHeight < blockHeight {
			upDown := r.Intn(2)

			if upDown == 0 {
				lastPrice = sdk.MinDec(lastPrice.Add(increment), maxPrice)
			} else {
				lastPrice = sdk.MaxDec(lastPrice.Sub(increment), minPrice)
			}

			lastHeight++
		}

		p.currentPrice[marketID] = lastPrice
	}

	p.currentBlockHeight = blockHeight
}

// GetCurrentPrice returns price for last blockHeight set by Step
func (p *PriceGenerator) GetCurrentPrice(marketID string) sdk.Dec {
	price, ok := p.currentPrice[marketID]

	if !ok {
		panic("unknown market")
	}

	return price
}
