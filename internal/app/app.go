package app

import (
	"encoding/json"

	abci "github.com/tendermint/abci/types"
	tmtypes "github.com/tendermint/tendermint/types"
	cmn "github.com/tendermint/tmlibs/common"
	dbm "github.com/tendermint/tmlibs/db"
	"github.com/tendermint/tmlibs/log"

	bam "github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/bank"
	//"github.com/cosmos/cosmos-sdk/x/ibc"
	//"github.com/cosmos/cosmos-sdk/x/slashing"
	//"github.com/cosmos/cosmos-sdk/x/stake"
	"github.com/kava-labs/kava/internal/types"
	"github.com/kava-labs/kava/internal/x/paychan"
)

const (
	appName = "KavaApp"
)

// Extended ABCI application
type KavaApp struct {
	*bam.BaseApp
	cdc *wire.Codec

	// keys to access the substores
	keyMain    *sdk.KVStoreKey
	keyAccount *sdk.KVStoreKey
	keyPaychan *sdk.KVStoreKey
	//keyIBC      *sdk.KVStoreKey
	//keyStake    *sdk.KVStoreKey
	//keySlashing *sdk.KVStoreKey

	// Manage getting and setting accounts
	accountMapper       auth.AccountMapper
	feeCollectionKeeper auth.FeeCollectionKeeper
	coinKeeper          bank.Keeper
	paychanKeeper       paychan.Keeper
	//ibcMapper           ibc.Mapper
	//stakeKeeper         stake.Keeper
	//slashingKeeper      slashing.Keeper
}

func NewKavaApp(logger log.Logger, db dbm.DB) *KavaApp {

	// Create app-level codec for txs and accounts.
	var cdc = MakeCodec()

	// Create your application object.
	var app = &KavaApp{
		BaseApp:    bam.NewBaseApp(appName, cdc, logger, db),
		cdc:        cdc,
		keyMain:    sdk.NewKVStoreKey("main"),
		keyAccount: sdk.NewKVStoreKey("acc"),
		keyPaychan: sdk.NewKVStoreKey("paychan"),
		//keyIBC:      sdk.NewKVStoreKey("ibc"),
		//keyStake:    sdk.NewKVStoreKey("stake"),
		//keySlashing: sdk.NewKVStoreKey("slashing"),
	}

	// Define the accountMapper.
	app.accountMapper = auth.NewAccountMapper(
		cdc,
		app.keyAccount, // target store
		&auth.BaseAccount{},
	)

	// add accountMapper/handlers
	app.coinKeeper = bank.NewKeeper(app.accountMapper)
	app.paychanKeeper = paychan.NewKeeper(app.cdc, app.keyPaychan, app.coinKeeper)
	//app.ibcMapper = ibc.NewMapper(app.cdc, app.keyIBC, app.RegisterCodespace(ibc.DefaultCodespace))
	//app.stakeKeeper = stake.NewKeeper(app.cdc, app.keyStake, app.coinKeeper, app.RegisterCodespace(stake.DefaultCodespace))
	//app.slashingKeeper = slashing.NewKeeper(app.cdc, app.keySlashing, app.stakeKeeper, app.RegisterCodespace(slashing.DefaultCodespace))

	// register message routes
	app.Router().
		AddRoute("auth", auth.NewHandler(app.accountMapper)).
		AddRoute("bank", bank.NewHandler(app.coinKeeper)).
		AddRoute("paychan", paychan.NewHandler(app.paychanKeeper))
		//AddRoute("ibc", ibc.NewHandler(app.ibcMapper, app.coinKeeper)).
		//AddRoute("stake", stake.NewHandler(app.stakeKeeper))

	// Initialize BaseApp.
	app.SetInitChainer(app.initChainer)
	app.SetBeginBlocker(app.BeginBlocker)
	app.SetEndBlocker(app.EndBlocker)
	app.SetAnteHandler(auth.NewAnteHandler(app.accountMapper, app.feeCollectionKeeper))
	app.MountStoresIAVL(app.keyMain, app.keyAccount, app.keyPaychan) //, app.keyIBC, app.keyStake, app.keySlashing)
	err := app.LoadLatestVersion(app.keyMain)
	if err != nil {
		cmn.Exit(err.Error())
	}
	return app
}

// Custom tx codec
func MakeCodec() *wire.Codec {
	var cdc = wire.NewCodec()
	wire.RegisterCrypto(cdc) // Register crypto.
	sdk.RegisterWire(cdc)    // Register Msgs
	bank.RegisterWire(cdc)
	paychan.RegisterWire(cdc)
	//stake.RegisterWire(cdc)
	//slashing.RegisterWire(cdc)
	//ibc.RegisterWire(cdc)
	auth.RegisterWire(cdc)

	// register custom AppAccount
	//cdc.RegisterInterface((*auth.Account)(nil), nil)
	//cdc.RegisterConcrete(&types.BaseAccount{}, "kava/Account", nil)
	return cdc
}

// application updates every end block
func (app *KavaApp) BeginBlocker(ctx sdk.Context, req abci.RequestBeginBlock) abci.ResponseBeginBlock {
	//tags := slashing.BeginBlocker(ctx, req, app.slashingKeeper)

	//return abci.ResponseBeginBlock{
	//	Tags: tags.ToKVPairs(),
	//}
	return abci.ResponseBeginBlock{}
}

// application updates every end block
func (app *KavaApp) EndBlocker(ctx sdk.Context, req abci.RequestEndBlock) abci.ResponseEndBlock {
	//validatorUpdates := stake.EndBlocker(ctx, app.stakeKeeper)

	//return abci.ResponseEndBlock{
	//	ValidatorUpdates: validatorUpdates,
	//}
	return abci.ResponseEndBlock{}
}

// Custom logic for initialization
func (app *KavaApp) initChainer(ctx sdk.Context, req abci.RequestInitChain) abci.ResponseInitChain {
	stateJSON := req.AppStateBytes

	genesisState := new(types.GenesisState)
	err := app.cdc.UnmarshalJSON(stateJSON, genesisState)
	if err != nil {
		panic(err) // TODO https://github.com/cosmos/cosmos-sdk/issues/468
		// return sdk.ErrGenesisParse("").TraceCause(err, "")
	}

	for _, gacc := range genesisState.Accounts {
		acc, err := gacc.ToAppAccount()
		if err != nil {
			panic(err) // TODO https://github.com/cosmos/cosmos-sdk/issues/468
			//	return sdk.ErrGenesisParse("").TraceCause(err, "")
		}
		acc.AccountNumber = app.accountMapper.GetNextAccountNumber(ctx)
		app.accountMapper.SetAccount(ctx, acc)
	}

	// load the initial stake information
	//stake.InitGenesis(ctx, app.stakeKeeper, genesisState.StakeData)

	return abci.ResponseInitChain{}
}

// Custom logic for state export
func (app *KavaApp) ExportAppStateAndValidators() (appState json.RawMessage, validators []tmtypes.GenesisValidator, err error) {
	ctx := app.NewContext(true, abci.Header{})

	// iterate to get the accounts
	accounts := []types.GenesisAccount{}
	appendAccount := func(acc auth.Account) (stop bool) {
		account := types.GenesisAccount{
			Address: acc.GetAddress(),
			Coins:   acc.GetCoins(),
		}
		accounts = append(accounts, account)
		return false
	}
	app.accountMapper.IterateAccounts(ctx, appendAccount)

	genState := types.GenesisState{
		Accounts: accounts,
	}
	appState, err = wire.MarshalJSONIndent(app.cdc, genState)
	if err != nil {
		return nil, nil, err
	}

	validators = make([]tmtypes.GenesisValidator, 0) // TODO export the actual validators

	return appState, validators, err
}
