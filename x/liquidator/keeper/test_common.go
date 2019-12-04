package keeper

import (
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/params"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	tmtime "github.com/tendermint/tendermint/types/time"
	dbm "github.com/tendermint/tm-db"

	"github.com/kava-labs/kava/x/auction"
	"github.com/kava-labs/kava/x/cdp"
	"github.com/kava-labs/kava/x/liquidator/types"
	"github.com/kava-labs/kava/x/pricefeed"
)

// Avoid cluttering test cases with long function name
func i(in int64) sdk.Int                    { return sdk.NewInt(in) }
func c(denom string, amount int64) sdk.Coin { return sdk.NewInt64Coin(denom, amount) }
func cs(coins ...sdk.Coin) sdk.Coins        { return sdk.NewCoins(coins...) }

type keepers struct {
	paramsKeeper     params.Keeper
	accountKeeper    auth.AccountKeeper
	bankKeeper       bank.Keeper
	pricefeedKeeper  pricefeed.Keeper
	auctionKeeper    auction.Keeper
	cdpKeeper        cdp.Keeper
	liquidatorKeeper Keeper
}

func setupTestKeepers() (sdk.Context, keepers) {

	// Setup in memory database
	keyParams := sdk.NewKVStoreKey(params.StoreKey)
	tkeyParams := sdk.NewTransientStoreKey(params.TStoreKey)
	keyAcc := sdk.NewKVStoreKey(auth.StoreKey)
	keyPriceFeed := sdk.NewKVStoreKey(pricefeed.StoreKey)
	keyCDP := sdk.NewKVStoreKey("cdp")
	keyAuction := sdk.NewKVStoreKey("auction")
	keyLiquidator := sdk.NewKVStoreKey("liquidator")

	db := dbm.NewMemDB()
	ms := store.NewCommitMultiStore(db)
	ms.MountStoreWithDB(keyParams, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(tkeyParams, sdk.StoreTypeTransient, db)
	ms.MountStoreWithDB(keyAcc, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(keyPriceFeed, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(keyCDP, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(keyAuction, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(keyLiquidator, sdk.StoreTypeIAVL, db)
	err := ms.LoadLatestVersion()
	if err != nil {
		panic(err)
	}

	// Create Codec
	cdc := makeTestCodec()

	// Create Keepers
	paramsKeeper := params.NewKeeper(cdc, keyParams, tkeyParams, params.DefaultCodespace)
	accountKeeper := auth.NewAccountKeeper(
		cdc,
		keyAcc,
		paramsKeeper.Subspace(auth.DefaultParamspace),
		auth.ProtoBaseAccount,
	)
	blacklistedAddrs := make(map[string]bool)
	bankKeeper := bank.NewBaseKeeper(
		accountKeeper,
		paramsKeeper.Subspace(bank.DefaultParamspace),
		bank.DefaultCodespace,
		blacklistedAddrs,
	)
	pricefeedKeeper := pricefeed.NewKeeper(keyPriceFeed, cdc, paramsKeeper.Subspace(pricefeed.DefaultParamspace), pricefeed.DefaultCodespace)
	cdpKeeper := cdp.NewKeeper(
		cdc,
		keyCDP,
		paramsKeeper.Subspace(cdp.DefaultParamspace),
		pricefeedKeeper,
		bankKeeper,
	)
	auctionKeeper := auction.NewKeeper(cdc, cdpKeeper, keyAuction, paramsKeeper.Subspace(auction.DefaultParamspace)) // Note: cdp keeper stands in for bank keeper
	liquidatorKeeper := NewKeeper(
		cdc,
		keyLiquidator,
		paramsKeeper.Subspace(types.DefaultParamspace),
		cdpKeeper,
		auctionKeeper,
		cdpKeeper,
	) // Note: cdp keeper stands in for bank keeper

	// Create context
	ctx := sdk.NewContext(ms, abci.Header{ChainID: "testchain"}, false, log.NewNopLogger())

	return ctx, keepers{
		paramsKeeper,
		accountKeeper,
		bankKeeper,
		pricefeedKeeper,
		auctionKeeper,
		cdpKeeper,
		liquidatorKeeper,
	}
}

func makeTestCodec() *codec.Codec {
	var cdc = codec.New()
	auth.RegisterCodec(cdc)
	bank.RegisterCodec(cdc)
	pricefeed.RegisterCodec(cdc)
	auction.RegisterCodec(cdc)
	cdp.RegisterCodec(cdc)
	types.RegisterCodec(cdc)
	sdk.RegisterCodec(cdc)
	codec.RegisterCrypto(cdc)
	return cdc
}

func defaultParams() types.LiquidatorParams {
	return types.LiquidatorParams{
		DebtAuctionSize: sdk.NewInt(1000),
		CollateralParams: []types.CollateralParams{
			{
				Denom:       "btc",
				AuctionSize: sdk.NewInt(1),
			},
		},
	}
}

func cdpDefaultGenesis() cdp.GenesisState {
	return cdp.GenesisState{
		cdp.CdpParams{
			GlobalDebtLimit: sdk.NewInt(1000000),
			CollateralParams: []cdp.CollateralParams{
				{
					Denom:            "btc",
					LiquidationRatio: sdk.MustNewDecFromStr("1.5"),
					DebtLimit:        sdk.NewInt(500000),
				},
			},
		},
		sdk.ZeroInt(),
		cdp.CDPs{},
	}
}

func pricefeedGenesis() pricefeed.GenesisState {
	ap := pricefeed.Params{
		Markets: []pricefeed.Market{
			pricefeed.Market{MarketID: "btc", BaseAsset: "btc", QuoteAsset: "usd", Oracles: pricefeed.Oracles{}, Active: true}},
	}
	return pricefeed.GenesisState{
		Params: ap,
		PostedPrices: []pricefeed.PostedPrice{
			pricefeed.PostedPrice{
				MarketID:      "btc",
				OracleAddress: sdk.AccAddress{},
				Price:         sdk.MustNewDecFromStr("8000.00"),
				Expiry:        tmtime.Now().Add(1 * time.Hour),
			},
		},
	}
}
