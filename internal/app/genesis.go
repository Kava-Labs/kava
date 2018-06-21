package app

import (
	"encoding/json"
	"errors"

	//"github.com/spf13/pflag"
	//"github.com/spf13/viper"
	crypto "github.com/tendermint/go-crypto"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/cosmos/cosmos-sdk/server"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
	"github.com/cosmos/cosmos-sdk/x/auth"
	//"github.com/cosmos/cosmos-sdk/x/stake"

	"github.com/kava-labs/kava/internal/types"
)

// A file, genesis.json, is created with the initial state of the kava network.
// This is done by creating an AppInit object to be handed to the server when it creates commands.
// When `kvd init` is run, a genesis tx is created. Then, from that, an initial app state.

func CreateAppInit() server.AppInit {
	return server.AppInit{
		AppGenTx:    KavaAppGenTx,
		AppGenState: KavaAppGenState,
	}
}

func KavaAppGenTx(cdc *wire.Codec, pk crypto.PubKey) (
	appGenTx, cliPrint json.RawMessage, validator tmtypes.GenesisValidator, err error) {

	// Generate address and secret key for the validator
	var addr sdk.Address
	var secret string
	addr, secret, err = server.GenerateCoinKey()
	if err != nil {
		return
	}

	// Generate appGenTx -------------
	var bz []byte
	genTx := types.GenTx{
		Address: addr,
		PubKey:  pk,
	}
	bz, err = wire.MarshalJSONIndent(cdc, genTx)
	if err != nil {
		return
	}
	appGenTx = json.RawMessage(bz)

	// cliPrint -------------
	mm := map[string]string{"secret": secret}
	bz, err = cdc.MarshalJSON(mm)
	if err != nil {
		return
	}
	cliPrint = json.RawMessage(bz)

	// validator -----------
	validator = tmtypes.GenesisValidator{
		PubKey: pk,
		Power:  10,
	}

	return
}

func KavaAppGenState(cdc *wire.Codec, appGenTxs []json.RawMessage) (appState json.RawMessage, err error) {

	if len(appGenTxs) == 0 {
		err = errors.New("must provide at least 1 genesis transaction")
		return
	}

	genaccs := make([]types.GenesisAccount, len(appGenTxs))
	for i, appGenTx := range appGenTxs {

		var genTx types.GenTx
		err = cdc.UnmarshalJSON(appGenTx, &genTx)
		if err != nil {
			return
		}

		// create the genesis account
		accAuth := auth.NewBaseAccountWithAddress(genTx.Address)
		accAuth.Coins = sdk.Coins{
			{"KVA", 10000000000},
		}
		acc := types.NewGenesisAccount(&accAuth)
		genaccs[i] = acc
	}

	// create the final app state
	genesisState := types.GenesisState{
		Accounts: genaccs,
	}
	appState, err = wire.MarshalJSONIndent(cdc, genesisState)

	return
}

/*

// --------------------- default init --------------------------

// simple default application init
var DefaultAppInit = AppInit{
	AppGenTx:    SimpleAppGenTx,
	AppGenState: SimpleAppGenState,
}

// simple genesis tx
type SimpleGenTx struct {
	Addr sdk.Address `json:"addr"`
}

// Generate a genesis transaction
func SimpleAppGenTx(cdc *wire.Codec, pk crypto.PubKey) (
	appGenTx, cliPrint json.RawMessage, validator tmtypes.GenesisValidator, err error) {

	var addr sdk.Address
	var secret string
	addr, secret, err = GenerateCoinKey()
	if err != nil {
		return
	}

	var bz []byte
	simpleGenTx := SimpleGenTx{addr}
	bz, err = cdc.MarshalJSON(simpleGenTx)
	if err != nil {
		return
	}
	appGenTx = json.RawMessage(bz)

	mm := map[string]string{"secret": secret}
	bz, err = cdc.MarshalJSON(mm)
	if err != nil {
		return
	}
	cliPrint = json.RawMessage(bz)

	validator = tmtypes.GenesisValidator{
		PubKey: pk,
		Power:  10,
	}
	return
}

// create the genesis app state
func SimpleAppGenState(cdc *wire.Codec, appGenTxs []json.RawMessage) (appState json.RawMessage, err error) {

	if len(appGenTxs) != 1 {
		err = errors.New("must provide a single genesis transaction")
		return
	}

	var genTx SimpleGenTx
	err = cdc.UnmarshalJSON(appGenTxs[0], &genTx)
	if err != nil {
		return
	}

	appState = json.RawMessage(fmt.Sprintf(`{
  "accounts": [{
    "address": "%s",
    "coins": [
      {
        "denom": "mycoin",
        "amount": 9007199254740992
      }
    ]
  }]
}`, genTx.Addr.String()))
	return
}


// -------------------- gaia init ----------------------

// simple genesis tx
type GaiaGenTx struct {
	Name    string        `json:"name"`
	Address sdk.Address   `json:"address"`
	PubKey  crypto.PubKey `json:"pub_key"`
}

// Generate a gaia genesis transaction with flags
func GaiaAppGenTx(cdc *wire.Codec, pk crypto.PubKey) (
	appGenTx, cliPrint json.RawMessage, validator tmtypes.GenesisValidator, err error) {
	clientRoot := viper.GetString(flagClientHome)
	overwrite := viper.GetBool(flagOWK)
	name := viper.GetString(flagName)
	if name == "" {
		return nil, nil, tmtypes.GenesisValidator{}, errors.New("Must specify --name (validator moniker)")
	}

	var addr sdk.Address
	var secret string
	addr, secret, err = server.GenerateSaveCoinKey(clientRoot, name, "1234567890", overwrite)
	if err != nil {
		return
	}
	mm := map[string]string{"secret": secret}
	var bz []byte
	bz, err = cdc.MarshalJSON(mm)
	if err != nil {
		return
	}
	cliPrint = json.RawMessage(bz)
	appGenTx,_,validator,err = GaiaAppGenTxNF(cdc, pk, addr, name, overwrite)
	return
}

// Generate a gaia genesis transaction without flags
func GaiaAppGenTxNF(cdc *wire.Codec, pk crypto.PubKey, addr sdk.Address, name string, overwrite bool) (
	appGenTx, cliPrint json.RawMessage, validator tmtypes.GenesisValidator, err error) {

	var bz []byte
	gaiaGenTx := GaiaGenTx{
		Name:    name,
		Address: addr,
		PubKey:  pk,
	}
	bz, err = wire.MarshalJSONIndent(cdc, gaiaGenTx)
	if err != nil {
		return
	}
	appGenTx = json.RawMessage(bz)

	validator = tmtypes.GenesisValidator{
		PubKey: pk,
		Power:  freeFermionVal,
	}
	return
}

// Create the core parameters for genesis initialization for gaia
// note that the pubkey input is this machines pubkey
func GaiaAppGenState(cdc *wire.Codec, appGenTxs []json.RawMessage) (genesisState GenesisState, err error) {

	if len(appGenTxs) == 0 {
		err = errors.New("must provide at least genesis transaction")
		return
	}

	// start with the default staking genesis state
	stakeData := stake.DefaultGenesisState()

	// get genesis flag account information
	genaccs := make([]GenesisAccount, len(appGenTxs))
	for i, appGenTx := range appGenTxs {

		var genTx GaiaGenTx
		err = cdc.UnmarshalJSON(appGenTx, &genTx)
		if err != nil {
			return
		}

		// create the genesis account, give'm few steaks and a buncha token with there name
		accAuth := auth.NewBaseAccountWithAddress(genTx.Address)
		accAuth.Coins = sdk.Coins{
			{genTx.Name + "Token", 1000},
			{"steak", freeFermionsAcc},
		}
		acc := NewGenesisAccount(&accAuth)
		genaccs[i] = acc
		stakeData.Pool.LooseUnbondedTokens += freeFermionsAcc // increase the supply

		// add the validator
		if len(genTx.Name) > 0 {
			desc := stake.NewDescription(genTx.Name, "", "", "")
			validator := stake.NewValidator(genTx.Address, genTx.PubKey, desc)
			validator.PoolShares = stake.NewBondedShares(sdk.NewRat(freeFermionVal))
			stakeData.Validators = append(stakeData.Validators, validator)

			// pool logic
			stakeData.Pool.BondedTokens += freeFermionVal
			stakeData.Pool.BondedShares = sdk.NewRat(stakeData.Pool.BondedTokens)
		}
	}

	// create the final app state
	genesisState = GenesisState{
		Accounts:  genaccs,
		StakeData: stakeData,
	}
	return
}

// GaiaAppGenState but with JSON
func GaiaAppGenStateJSON(cdc *wire.Codec, appGenTxs []json.RawMessage) (appState json.RawMessage, err error) {

	// create the final app state
	genesisState, err := GaiaAppGenState(cdc, appGenTxs)
	if err != nil {
		return nil, err
	}
	appState, err = wire.MarshalJSONIndent(cdc, genesisState)
	return
}
*/
