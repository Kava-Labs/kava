package app

import (
	"encoding/json"
	"errors"

	"github.com/spf13/pflag"
	"github.com/tendermint/tendermint/crypto"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/server/config"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/stake"
)

var (
	// bonded tokens given to genesis validators/accounts
	freeFermionVal  = int64(100)
	freeFermionsAcc = int64(50)
)

// Initial app state to be written to (and read from) genesis file
type GenesisState struct {
	Accounts  []GenesisAccount   `json:"accounts"`
	StakeData stake.GenesisState `json:"stake"`
}

// A simplified version of a normal account. It doesn't have pubkey or sequence.
type GenesisAccount struct {
	Address sdk.AccAddress `json:"address"`
	Coins   sdk.Coins      `json:"coins"`
}

// TODO remove?
func NewGenesisAccount(acc *auth.BaseAccount) GenesisAccount {
	return GenesisAccount{
		Address: acc.Address,
		Coins:   acc.Coins,
	}
}

// TODO remove?
func NewGenesisAccountI(acc auth.Account) GenesisAccount {
	return GenesisAccount{
		Address: acc.GetAddress(),
		Coins:   acc.GetCoins(),
	}
}

// Converts a GenesisAccount to auth.BaseAccount TODO rename
func (ga *GenesisAccount) ToAccount() (acc *auth.BaseAccount) {
	return &auth.BaseAccount{
		Address: ga.Address,
		Coins:   ga.Coins.Sort(),
	}
}

// Create the appInit stuct for server init command
func KavaAppInit() server.AppInit {
	fsAppGenState := pflag.NewFlagSet("", pflag.ContinueOnError)
	fsAppGenTx := pflag.NewFlagSet("", pflag.ContinueOnError)

	fsAppGenTx.String(server.FlagName, "", "validator moniker, required")
	fsAppGenTx.String(server.FlagClientHome, DefaultCLIHome,
		"home directory for the client, used for key generation")
	fsAppGenTx.Bool(server.FlagOWK, false, "overwrite the accounts created")

	return server.AppInit{
		FlagsAppGenState: fsAppGenState,
		FlagsAppGenTx:    fsAppGenTx,
		AppGenTx:         KavaAppGenTx,
		AppGenState:      KavaAppGenStateJSON,
	}
}

// Define format for GenTx json
type KavaGenTx struct {
	Name    string         `json:"name"`
	Address sdk.AccAddress `json:"address"`
	PubKey  string         `json:"pub_key"`
}

// Generate a genesis transsction
func KavaAppGenTx(cdc *wire.Codec, pk crypto.PubKey, genTxConfig config.GenTx) (
	appGenTx, cliPrint json.RawMessage, validator tmtypes.GenesisValidator, err error) {

	// Generate address and secret key for the validator
	if genTxConfig.Name == "" {
		return nil, nil, tmtypes.GenesisValidator{}, errors.New("Must specify --name (validator moniker)")
	}
	var addr sdk.AccAddress
	var secret string
	addr, secret, err = server.GenerateSaveCoinKey(genTxConfig.CliRoot, genTxConfig.Name, "password", genTxConfig.Overwrite)
	if err != nil {
		return
	}

	// Create string to print out
	mm := map[string]string{"secret": secret}
	var bz []byte
	bz, err = cdc.MarshalJSON(mm)
	if err != nil {
		return
	}
	cliPrint = json.RawMessage(bz)

	// Create genTx and validator
	appGenTx, _, validator, err = KavaAppGenTxNF(cdc, pk, addr, genTxConfig.Name)

	return
}

// TODO combine with KavaAppGenTx
func KavaAppGenTxNF(cdc *wire.Codec, pk crypto.PubKey, addr sdk.AccAddress, name string) (
	appGenTx, cliPrint json.RawMessage, validator tmtypes.GenesisValidator, err error) {

	// Create the gentx
	var bz []byte
	genTx := KavaGenTx{
		Name:    name,
		Address: addr,
		PubKey:  sdk.MustBech32ifyAccPub(pk),
	}
	bz, err = wire.MarshalJSONIndent(cdc, genTx)
	if err != nil {
		return
	}
	appGenTx = json.RawMessage(bz)

	// Create the validator
	validator = tmtypes.GenesisValidator{
		PubKey: pk,
		Power:  freeFermionVal,
	}
	return
}

// Create the core parameters for genesis initialization
// note that the pubkey input is this machines pubkey
func KavaAppGenState(cdc *wire.Codec, appGenTxs []json.RawMessage) (genesisState GenesisState, err error) {

	if len(appGenTxs) == 0 {
		err = errors.New("must provide at least 1 genesis transaction")
		return
	}

	// start with the default staking genesis state
	stakeData := stake.DefaultGenesisState()

	// get genesis flag account information
	genaccs := make([]GenesisAccount, len(appGenTxs))
	for i, appGenTx := range appGenTxs {

		var genTx KavaGenTx
		err = cdc.UnmarshalJSON(appGenTx, &genTx)
		if err != nil {
			return
		}

		// create the genesis account, give'm few steaks and a buncha token with there name
		accAuth := auth.NewBaseAccountWithAddress(genTx.Address)
		accAuth.Coins = sdk.Coins{
			{genTx.Name + "Token", sdk.NewInt(1000)},
			{"steak", sdk.NewInt(freeFermionsAcc)},
		}
		acc := NewGenesisAccount(&accAuth)
		genaccs[i] = acc
		stakeData.Pool.LooseTokens = stakeData.Pool.LooseTokens.Add(sdk.NewRat(freeFermionsAcc)) // increase the supply

		// add the validator
		if len(genTx.Name) > 0 {
			desc := stake.NewDescription(genTx.Name, "", "", "")
			validator := stake.NewValidator(genTx.Address,
				sdk.MustGetAccPubKeyBech32(genTx.PubKey), desc)

			stakeData.Pool.LooseTokens = stakeData.Pool.LooseTokens.Add(sdk.NewRat(freeFermionVal)) // increase the supply

			// add some new shares to the validator
			var issuedDelShares sdk.Rat
			validator, stakeData.Pool, issuedDelShares = validator.AddTokensFromDel(stakeData.Pool, freeFermionVal)
			stakeData.Validators = append(stakeData.Validators, validator)

			// create the self-delegation from the issuedDelShares
			delegation := stake.Delegation{
				DelegatorAddr: validator.Owner,
				ValidatorAddr: validator.Owner,
				Shares:        issuedDelShares,
				Height:        0,
			}

			stakeData.Bonds = append(stakeData.Bonds, delegation)
		}
	}

	// create the final app state
	genesisState = GenesisState{
		Accounts:  genaccs,
		StakeData: stakeData,
	}
	return
}

// Run KavaAppGenState then convert to JSON
func KavaAppGenStateJSON(cdc *wire.Codec, appGenTxs []json.RawMessage) (appState json.RawMessage, err error) {

	// create the final app state
	genesisState, err := KavaAppGenState(cdc, appGenTxs)
	if err != nil {
		return nil, err
	}
	appState, err = wire.MarshalJSONIndent(cdc, genesisState)
	return
}

/*

// A file, genesis.json, is created with the initial state of the kava network.
// This is done by creating an AppInit object to be handed to the server when it creates commands.
// When `kvd init` is run, a genesis tx is created. Then, from that, an initial app state.

func CreateAppInit() server.AppInit {
	return server.AppInit{
		AppGenTx:    KavaAppGenTx,
		AppGenState: KavaAppGenState,
	}
}

func KavaAppGenTx(cdc *wire.Codec, pk crypto.PubKey, genTxConfig config.GenTx) (
	appGenTx, cliPrint json.RawMessage, validator tmtypes.GenesisValidator, err error) {

	// Generate address and secret key for the validator
	var addr sdk.AccAddress
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

*/
