package cmd

import (
	"os"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/config"
	"github.com/cosmos/cosmos-sdk/client/debug"
	"github.com/cosmos/cosmos-sdk/client/keys"
	"github.com/cosmos/cosmos-sdk/client/rpc"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	genutilcli "github.com/cosmos/cosmos-sdk/x/genutil/client/cli"
	"github.com/spf13/cobra"
	tmcli "github.com/tendermint/tendermint/libs/cli"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/app/params"
)

// TODO set sdk config somewhere

// NewRootCmd creates a new root command for kava.
func NewRootCmd() *cobra.Command {
	encodingConfig := app.MakeEncodingConfig()

	initClientCtx := client.Context{}.
		WithCodec(encodingConfig.Marshaler).
		WithInterfaceRegistry(encodingConfig.InterfaceRegistry).
		WithTxConfig(encodingConfig.TxConfig).
		WithLegacyAmino(encodingConfig.Amino).
		WithInput(os.Stdin).
		WithAccountRetriever(types.AccountRetriever{}).
		WithHomeDir(app.DefaultNodeHome). // TODO why is this in app ?
		WithViper("")                     // TODO this sets the env prefix

	rootCmd := &cobra.Command{
		Use:   "kava",
		Short: "Daemon and CLI for the Kava blockchain.",
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			initClientCtx = client.ReadHomeFlag(initClientCtx, cmd)
			initClientCtx, err := config.ReadFromClientConfig(initClientCtx)
			if err != nil {
				return err
			}

			if err = client.SetCmdClientContextHandler(initClientCtx, cmd); err != nil {
				return err
			}

			return server.InterceptConfigsPreRunHandler(cmd, "", nil)
		},
	}

	addSubCmds(rootCmd, encodingConfig)

	return rootCmd
}

func addSubCmds(rootCmd *cobra.Command, encodingConfig params.EncodingConfig) {
	rootCmd.AddCommand(
		genutilcli.InitCmd(app.ModuleBasics, app.DefaultNodeHome),
		genutilcli.CollectGenTxsCmd(banktypes.GenesisBalancesIterator{}, app.DefaultNodeHome),
		// app.MigrateGenesisCmd(), // TODO
		genutilcli.GenTxCmd(app.ModuleBasics, encodingConfig.TxConfig, banktypes.GenesisBalancesIterator{}, app.DefaultNodeHome),
		genutilcli.ValidateGenesisCmd(app.ModuleBasics),
		// AddGenesisAccountCmd(app.DefaultNodeHome), // TODO
		tmcli.NewCompletionCmd(rootCmd, true), // TODO add other shells
		// testnetCmd(app.ModuleBasics, banktypes.GenesisBalancesIterator{}), // TODO
		debug.Cmd(),
		config.Cmd(),
	)

	ac := appCreator{
		encodingConfig: encodingConfig,
	}
	server.AddCommands(rootCmd, app.DefaultNodeHome, ac.newApp, ac.appExport, ac.addStartCmdFlags)

	// add keybase, auxiliary RPC, query, and tx child commands
	rootCmd.AddCommand(
		rpc.StatusCommand(),
		newQueryCmd(),
		newTxCmd(),
		keys.Commands(app.DefaultNodeHome),
	)
}
