package app

import (
	"encoding/json"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/exported" 
)

// GenesisState represents the genesis state of the blockchain. It is a map from module names to module genesis states.
type GenesisState map[string]json.RawMessage

// NewDefaultGenesisState generates the default state for the application.
func NewDefaultGenesisState() GenesisState {
	return ModuleBasics.DefaultGenesis()
}

// TODO remove iterator once merged into sdk master
type GenesisAccountIterator struct{}
// IterateGenesisAccounts iterates over all the genesis accounts found in
// appGenesis and invokes a callback on each genesis account. If any call
// returns true, iteration stops.
func (GenesisAccountIterator) IterateGenesisAccounts(
   cdc *codec.Codec, appGenesis map[string]json.RawMessage, cb func(exported.Account) (stop bool),
) {

	var authGenState auth.GenesisState
	cdc.MustUnmarshalJSON(appGenesis[auth.ModuleName], &authGenState)

	for _, genAcc := range authGenState.Accounts {
	   if cb(genAcc) {
		   break
	   }
   }
}