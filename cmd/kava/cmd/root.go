package cmd

import (
	"os"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/config"
	"github.com/cosmos/cosmos-sdk/client/debug"
	"github.com/cosmos/cosmos-sdk/client/keys"
	"github.com/cosmos/cosmos-sdk/server"
	simdcmd "github.com/cosmos/cosmos-sdk/simapp/simd/cmd"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	genutilcli "github.com/cosmos/cosmos-sdk/x/genutil/client/cli"
	"github.com/spf13/cobra"
	tmcli "github.com/tendermint/tendermint/libs/cli"
	ethermintserver "github.com/tharsis/ethermint/server"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/app/params"
	"github.com/kava-labs/kava/migrate"
)

// NewRootCmd creates a new root command for the kava blockchain.
func NewRootCmd() *cobra.Command {
	app.SetSDKConfig().Seal()

	encodingConfig := app.MakeEncodingConfig()

	initClientCtx := client.Context{}.
		WithCodec(encodingConfig.Marshaler).
		WithInterfaceRegistry(encodingConfig.InterfaceRegistry).
		WithTxConfig(encodingConfig.TxConfig).
		WithLegacyAmino(encodingConfig.Amino).
		WithInput(os.Stdin).
		WithAccountRetriever(types.AccountRetriever{}).
		WithHomeDir(app.DefaultNodeHome).
		WithViper("") // TODO this sets the env prefix

	rootCmd := &cobra.Command{
		Use:   "kava",
		Short: "Daemon and CLI for the Kava blockchain.",
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			cmd.SetOut(cmd.OutOrStdout())
			cmd.SetErr(cmd.ErrOrStderr())

			initClientCtx, err := client.ReadPersistentCommandFlags(initClientCtx, cmd.Flags())
			if err != nil {
				return err
			}

			initClientCtx, err = config.ReadFromClientConfig(initClientCtx)
			if err != nil {
				return err
			}

			if err = client.SetCmdClientContextHandler(initClientCtx, cmd); err != nil {
				return err
			}

			return server.InterceptConfigsPreRunHandler(cmd, "", nil)
		},
	}

	addSubCmds(rootCmd, encodingConfig, app.DefaultNodeHome)

	return rootCmd
}

// addSubCmds registers all the sub commands used by kava.
func addSubCmds(rootCmd *cobra.Command, encodingConfig params.EncodingConfig, defaultNodeHome string) {
	rootCmd.AddCommand(
		// TODO: Need to wrap InitCmd with `ethermintclient.ValidateChainID`
		genutilcli.InitCmd(app.ModuleBasics, defaultNodeHome),
		genutilcli.CollectGenTxsCmd(banktypes.GenesisBalancesIterator{}, defaultNodeHome),
		migrate.MigrateGenesisCmd(),
		migrate.AssertInvariantsCmd(encodingConfig),
		genutilcli.GenTxCmd(app.ModuleBasics, encodingConfig.TxConfig, banktypes.GenesisBalancesIterator{}, defaultNodeHome),
		genutilcli.ValidateGenesisCmd(app.ModuleBasics),
		simdcmd.AddGenesisAccountCmd(defaultNodeHome), // TODO use kava version with vesting accounts
		tmcli.NewCompletionCmd(rootCmd, true),         // TODO add other shells, drop tmcli dependency, unhide?
		// testnetCmd(app.ModuleBasics, banktypes.GenesisBalancesIterator{}), // TODO add
		debug.Cmd(),
		config.Cmd(),
	)

	ac := appCreator{
		encodingConfig: encodingConfig,
	}

	// ethermintserver adds additional flags to start the JSON-RPC server for evm support
	ethermintserver.AddCommands(rootCmd, defaultNodeHome, ac.newApp, ac.appExport, ac.addStartCmdFlags)

	// add keybase, auxiliary RPC, query, and tx child commands
	rootCmd.AddCommand(
		StatusCommand(),
		newQueryCmd(),
		newTxCmd(),
		keys.Commands(defaultNodeHome),
	)
}
