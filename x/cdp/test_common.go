package cdp

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/mock"

	"github.com/kava-labs/kava/x/pricefeed"
)

// Mock app is an ABCI app with an in memory database.
// This function creates an app, setting up the keepers, routes, begin and end blockers.
// But leaves it to the tests to call InitChain (done by calling mock.SetGenesis)
// The app works by submitting ABCI messages.
//  - InitChain sets up the app db from genesis.
//  - BeginBlock starts the delivery of a new block
//  - DeliverTx delivers a tx
//  - EndBlock signals the end of a block
//  - Commit ?
func setUpMockAppWithoutGenesis() (*mock.App, Keeper, PricefeedKeeper) {
	// Create uninitialized mock app
	mapp := mock.NewApp()

	// Register codecs
	RegisterCodec(mapp.Cdc)

	// Create keepers
	keyCDP := sdk.NewKVStoreKey("cdp")
	keyPriceFeed := sdk.NewKVStoreKey(pricefeed.StoreKey)
	pk := mapp.ParamsKeeper
	priceFeedKeeper := pricefeed.NewKeeper(keyPriceFeed, mapp.Cdc, pk.Subspace(pricefeed.DefaultParamspace).WithKeyTable(pricefeed.ParamKeyTable()), pricefeed.DefaultCodespace)
	blacklistedAddrs := make(map[string]bool)
	bankKeeper := bank.NewBaseKeeper(mapp.AccountKeeper, pk.Subspace(bank.DefaultParamspace), bank.DefaultCodespace, blacklistedAddrs)
	cdpKeeper := NewKeeper(mapp.Cdc, keyCDP, pk.Subspace(DefaultParamspace), priceFeedKeeper, bankKeeper)

	// Register routes
	mapp.Router().AddRoute("cdp", NewHandler(cdpKeeper))

	// Mount and load the stores
	err := mapp.CompleteSetup(keyPriceFeed, keyCDP)
	if err != nil {
		panic("mock app setup failed")
	}

	return mapp, cdpKeeper, priceFeedKeeper
}

// Avoid cluttering test cases with long function name
func i(in int64) sdk.Int                    { return sdk.NewInt(in) }
func d(str string) sdk.Dec                  { return sdk.MustNewDecFromStr(str) }
func c(denom string, amount int64) sdk.Coin { return sdk.NewInt64Coin(denom, amount) }
func cs(coins ...sdk.Coin) sdk.Coins        { return sdk.NewCoins(coins...) }
