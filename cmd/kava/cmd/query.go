package cmd

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/rpc"
	authcmd "github.com/cosmos/cosmos-sdk/x/auth/client/cli"
	"github.com/spf13/cobra"

	"github.com/kava-labs/kava/app"
)

// newQueryCmd creates all the commands for querying blockchain state.
func newQueryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "query",
		Aliases:                    []string{"q"},
		Short:                      "Querying subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	// TODO(boodyvo): Add query commands
	cmd.AddCommand(
		//authcmd.GetAccountCmd(),
		rpc.ValidatorCommand(),
		//rpc.BlockCommand(),
		authcmd.QueryTxsByEventsCmd(),
		authcmd.QueryTxCmd(),
	)

	//encodingConfig := app.MakeEncodingConfig()
	//tempApp := app.NewApp(log.NewNopLogger(), dbm.NewMemDB(), app.DefaultNodeHome, nil, encodingConfig, app.DefaultOptions)

	app.ModuleBasics.AddQueryCommands(cmd)
	cmd.PersistentFlags().String(flags.FlagChainID, "", "The network chain ID")

	return cmd
}
