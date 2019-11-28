package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/mock"

	"github.com/kava-labs/kava/x/cdp/types"
	"github.com/kava-labs/kava/x/pricefeed"
	"github.com/tendermint/tendermint/crypto"
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
func setUpMockAppWithoutGenesis() (*mock.App, Keeper, []sdk.AccAddress, []crypto.PrivKey) {
	// Create uninitialized mock app
	mapp := mock.NewApp()

	// Register codecs
	types.RegisterCodec(mapp.Cdc)

	// Create keepers
	keyCDP := sdk.NewKVStoreKey("cdp")
	keyPriceFeed := sdk.NewKVStoreKey(pricefeed.StoreKey)
	pk := mapp.ParamsKeeper
	priceFeedKeeper := pricefeed.NewKeeper(keyPriceFeed, mapp.Cdc, pk.Subspace(pricefeed.DefaultParamspace), pricefeed.DefaultCodespace)
	blacklistedAddrs := make(map[string]bool)
	bankKeeper := bank.NewBaseKeeper(mapp.AccountKeeper, pk.Subspace(bank.DefaultParamspace), bank.DefaultCodespace, blacklistedAddrs)
	cdpKeeper := NewKeeper(mapp.Cdc, keyCDP, pk.Subspace(types.DefaultParamspace), priceFeedKeeper, bankKeeper)

	// Mount and load the stores
	err := mapp.CompleteSetup(keyPriceFeed, keyCDP)
	if err != nil {
		panic("mock app setup failed")
	}

	// Create a bunch (ie 10) of pre-funded accounts to use for tests
	genAccs, addrs, _, privKeys := mock.CreateGenAccounts(10, sdk.NewCoins(sdk.NewInt64Coin("token1", 100), sdk.NewInt64Coin("token2", 100)))
	mock.SetGenesis(mapp, genAccs)

	return mapp, cdpKeeper, addrs, privKeys
}

// Avoid cluttering test cases with long function name
func i(in int64) sdk.Int                    { return sdk.NewInt(in) }
func d(str string) sdk.Dec                  { return sdk.MustNewDecFromStr(str) }
func c(denom string, amount int64) sdk.Coin { return sdk.NewInt64Coin(denom, amount) }
func cs(coins ...sdk.Coin) sdk.Coins        { return sdk.NewCoins(coins...) }

func defaultParamsMulti() types.CdpParams {
	return types.CdpParams{
		GlobalDebtLimit: sdk.NewInt(1000000),
		CollateralParams: []types.CollateralParams{
			{
				Denom:            "btc",
				LiquidationRatio: sdk.MustNewDecFromStr("1.5"),
				DebtLimit:        sdk.NewInt(500000),
			},
			{
				Denom:            "xrp",
				LiquidationRatio: sdk.MustNewDecFromStr("2.0"),
				DebtLimit:        sdk.NewInt(500000),
			},
		},
		StableDenoms: []string{"usdx"},
	}
}

func defaultParamsSingle() types.CdpParams {
	return types.CdpParams{
		GlobalDebtLimit: sdk.NewInt(1000000),
		CollateralParams: []types.CollateralParams{
			{
				Denom:            "xrp",
				LiquidationRatio: sdk.MustNewDecFromStr("2.0"),
				DebtLimit:        sdk.NewInt(500000),
			},
		},
		StableDenoms: []string{"usdx"},
	}
}
