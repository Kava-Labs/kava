package app

import (
	"encoding/json"
	"io"
	"os"

	abci "github.com/tendermint/tendermint/abci/types"
	cmn "github.com/tendermint/tendermint/libs/common"
	dbm "github.com/tendermint/tendermint/libs/db"
	"github.com/tendermint/tendermint/libs/log"
	tmtypes "github.com/tendermint/tendermint/types"

	bam "github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/bank"
	//"github.com/cosmos/cosmos-sdk/x/gov"
	//"github.com/cosmos/cosmos-sdk/x/ibc"
	"github.com/cosmos/cosmos-sdk/x/slashing"
	"github.com/cosmos/cosmos-sdk/x/stake"
	"github.com/kava-labs/kava/internal/x/paychan"
)

const (
	appName = "KavaApp"
)

// Set default directories for data
var (
	DefaultCLIHome  = os.ExpandEnv("$HOME/.kvcli")
	DefaultNodeHome = os.ExpandEnv("$HOME/.kvd")
)

type KavaApp struct {
	*bam.BaseApp
	cdc *wire.Codec

	// keys to access the substores
	keyMain    *sdk.KVStoreKey
	keyAccount *sdk.KVStoreKey

	//keyIBC           *sdk.KVStoreKey
	keyStake    *sdk.KVStoreKey
	keySlashing *sdk.KVStoreKey
	//keyGov           *sdk.KVStoreKey
	keyFeeCollection *sdk.KVStoreKey
	keyPaychan       *sdk.KVStoreKey

	// keepers
	accountMapper       auth.AccountMapper
	feeCollectionKeeper auth.FeeCollectionKeeper
	coinKeeper          bank.Keeper
	paychanKeeper       paychan.Keeper
	//ibcMapper           ibc.Mapper
	stakeKeeper    stake.Keeper
	slashingKeeper slashing.Keeper
	//govKeeper           gov.Keeper
}

// Creates a new KavaApp.
func NewKavaApp(logger log.Logger, db dbm.DB, traceStore io.Writer, baseAppOptions ...func(*bam.BaseApp)) *KavaApp {

	// Create a codec for use across the whole app
	cdc := CreateKavaAppCodec()

	// Create a new base app
	bApp := bam.NewBaseApp(appName, cdc, logger, db, baseAppOptions...)
	bApp.SetCommitMultiStoreTracer(traceStore)

	// Create the kava app, extending baseApp
	var app = &KavaApp{
		BaseApp:    bApp,
		cdc:        cdc,
		keyMain:    sdk.NewKVStoreKey("main"),
		keyAccount: sdk.NewKVStoreKey("acc"),
		keyPaychan: sdk.NewKVStoreKey("paychan"),
		//keyIBC:      sdk.NewKVStoreKey("ibc"),
		keyStake:    sdk.NewKVStoreKey("stake"),
		keySlashing: sdk.NewKVStoreKey("slashing"),
		//keyGov:           sdk.NewKVStoreKey("gov"),
		keyFeeCollection: sdk.NewKVStoreKey("fee"),
	}

	// Define the accountMapper and base account
	app.accountMapper = auth.NewAccountMapper(
		cdc,
		app.keyAccount,
		auth.ProtoBaseAccount,
	)

	// Create the keepers
	app.coinKeeper = bank.NewKeeper(app.accountMapper)
	app.paychanKeeper = paychan.NewKeeper(app.cdc, app.keyPaychan, app.coinKeeper)
	//app.ibcMapper = ibc.NewMapper(app.cdc, app.keyIBC, app.RegisterCodespace(ibc.DefaultCodespace))
	app.stakeKeeper = stake.NewKeeper(app.cdc, app.keyStake, app.coinKeeper, app.RegisterCodespace(stake.DefaultCodespace))
	app.slashingKeeper = slashing.NewKeeper(app.cdc, app.keySlashing, app.stakeKeeper, app.RegisterCodespace(slashing.DefaultCodespace))
	//app.govKeeper = gov.NewKeeper(app.cdc, app.keyGov, app.coinKeeper, app.stakeKeeper, app.RegisterCodespace(gov.DefaultCodespace))
	app.feeCollectionKeeper = auth.NewFeeCollectionKeeper(app.cdc, app.keyFeeCollection)

	// Register the message handlers
	app.Router().
		AddRoute("bank", bank.NewHandler(app.coinKeeper)).
		//AddRoute("ibc", ibc.NewHandler(app.ibcMapper, app.coinKeeper)).
		AddRoute("stake", stake.NewHandler(app.stakeKeeper)).
		AddRoute("slashing", slashing.NewHandler(app.slashingKeeper)).
		AddRoute("paychan", paychan.NewHandler(app.paychanKeeper))
		//AddRoute("gov", gov.NewHandler(app.govKeeper))

	// Set func to initialze the chain from appState in genesis file
	app.SetInitChainer(app.initChainer)

	// Set functions that run before and after txs / blocks
	app.SetBeginBlocker(app.BeginBlocker)
	app.SetEndBlocker(app.EndBlocker)
	app.SetAnteHandler(auth.NewAnteHandler(app.accountMapper, app.feeCollectionKeeper))

	// Mount stores
	app.MountStoresIAVL(app.keyMain, app.keyAccount, app.keyStake, app.keySlashing, app.keyFeeCollection, app.keyPaychan)
	err := app.LoadLatestVersion(app.keyMain)
	if err != nil {
		cmn.Exit(err.Error())
	}
	return app
}

// Creates a codec for use across the whole app.
func CreateKavaAppCodec() *wire.Codec {
	cdc := wire.NewCodec()
	paychan.RegisterWire(cdc)
	//ibc.RegisterWire(cdc)
	bank.RegisterWire(cdc)
	stake.RegisterWire(cdc)
	slashing.RegisterWire(cdc)
	//gov.RegisterWire(cdc)
	auth.RegisterWire(cdc)
	sdk.RegisterWire(cdc)
	wire.RegisterCrypto(cdc)
	return cdc
}

// The function baseapp runs on receipt of a BeginBlock ABCI message
func (app *KavaApp) BeginBlocker(ctx sdk.Context, req abci.RequestBeginBlock) abci.ResponseBeginBlock {
	tags := slashing.BeginBlocker(ctx, req, app.slashingKeeper)

	return abci.ResponseBeginBlock{
		Tags: tags.ToKVPairs(),
	}
}

// The function baseapp runs on receipt of a EndBlock ABCI message
func (app *KavaApp) EndBlocker(ctx sdk.Context, req abci.RequestEndBlock) abci.ResponseEndBlock {
	paychan.EndBlocker(ctx, app.paychanKeeper)
	validatorUpdates := stake.EndBlocker(ctx, app.stakeKeeper)

	//tags, _ := gov.EndBlocker(ctx, app.govKeeper)

	return abci.ResponseEndBlock{
		ValidatorUpdates: validatorUpdates,
		//Tags:             tags,
	}
}

// Initialzes the app db from the appState in the genesis file. Baseapp runs this on receipt of an InitChain ABCI message
func (app *KavaApp) initChainer(ctx sdk.Context, req abci.RequestInitChain) abci.ResponseInitChain {
	stateJSON := req.AppStateBytes

	var genesisState GenesisState
	err := app.cdc.UnmarshalJSON(stateJSON, &genesisState)
	if err != nil {
		panic(err)
	}

	// load the accounts
	for _, gacc := range genesisState.Accounts {
		acc := gacc.ToAccount()
		acc.AccountNumber = app.accountMapper.GetNextAccountNumber(ctx)
		app.accountMapper.SetAccount(ctx, acc)
	}

	// load the initial stake information
	err = stake.InitGenesis(ctx, app.stakeKeeper, genesisState.StakeData)
	if err != nil {
		panic(err)
	}

	//gov.InitGenesis(ctx, app.govKeeper, gov.DefaultGenesisState())

	return abci.ResponseInitChain{}
}

//
func (app *KavaApp) ExportAppStateAndValidators() (appState json.RawMessage, validators []tmtypes.GenesisValidator, err error) {
	ctx := app.NewContext(true, abci.Header{})

	// iterate to get the accounts
	accounts := []GenesisAccount{}
	appendAccount := func(acc auth.Account) (stop bool) {
		account := NewGenesisAccountI(acc)
		accounts = append(accounts, account)
		return false
	}
	app.accountMapper.IterateAccounts(ctx, appendAccount)

	genState := GenesisState{
		Accounts:  accounts,
		StakeData: stake.WriteGenesis(ctx, app.stakeKeeper),
	}
	appState, err = wire.MarshalJSONIndent(app.cdc, genState)
	if err != nil {
		return nil, nil, err
	}
	validators = stake.WriteValidators(ctx, app.stakeKeeper)
	return appState, validators, nil
}
