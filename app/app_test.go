package app

import (
	"os"
	"testing"

	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/crisis"

	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/libs/db"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/x/auth"
	distr "github.com/cosmos/cosmos-sdk/x/distribution"
	"github.com/cosmos/cosmos-sdk/x/gov"
	"github.com/cosmos/cosmos-sdk/x/mint"
	"github.com/cosmos/cosmos-sdk/x/slashing"
	"github.com/cosmos/cosmos-sdk/x/staking"

	abci "github.com/tendermint/tendermint/abci/types"
)

func setGenesis(app *App, accs ...*auth.BaseAccount) error {
	genaccs := make([]GenesisAccount, len(accs))
	for i, acc := range accs {
		genaccs[i] = NewGenesisAccount(acc)
	}

	genesisState := NewGenesisState(
		genaccs,
		auth.DefaultGenesisState(),
		bank.DefaultGenesisState(),
		staking.DefaultGenesisState(),
		mint.DefaultGenesisState(),
		distr.DefaultGenesisState(),
		gov.DefaultGenesisState(),
		crisis.DefaultGenesisState(),
		slashing.DefaultGenesisState(),
	)

	stateBytes, err := codec.MarshalJSONIndent(app.cdc, genesisState)
	if err != nil {
		return err
	}

	// Initialize the chain
	vals := []abci.ValidatorUpdate{}
	app.InitChain(abci.RequestInitChain{Validators: vals, AppStateBytes: stateBytes})
	app.Commit()

	return nil
}

func TestExport(t *testing.T) {
	db := db.NewMemDB()
	app := NewApp(log.NewTMLogger(log.NewSyncWriter(os.Stdout)), db, nil, true, 0)
	setGenesis(app)

	// Making a new app object with the db, so that initchain hasn't been called
	newApp := NewApp(log.NewTMLogger(log.NewSyncWriter(os.Stdout)), db, nil, true, 0)
	_, _, err := newApp.ExportAppStateAndValidators(false, []string{})
	require.NoError(t, err, "ExportAppStateAndValidators should not have an error")
}
