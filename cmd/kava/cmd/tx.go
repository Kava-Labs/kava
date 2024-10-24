package cmd

import (
	"cosmossdk.io/log"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	authcmd "github.com/cosmos/cosmos-sdk/x/auth/client/cli"
	"github.com/spf13/cobra"

	"github.com/kava-labs/kava/app"
)

// newTxCmd creates all commands for submitting blockchain transactions.
func newTxCmd() *cobra.Command {
	_, testAddresses := app.GeneratePrivKeyAddressPairs(10)
	manual := testAddresses[6:]
	encodingConfig := app.MakeEncodingConfig()
	opts := app.DefaultOptions
	opts.MempoolEnableAuth = true
	opts.MempoolAuthAddresses = manual

	tempApp := app.NewApp(
		log.NewNopLogger(),
		dbm.NewMemDB(),
		app.DefaultNodeHome,
		nil,
		encodingConfig,
		opts,
	)

	cmd := &cobra.Command{
		Use:                        "tx",
		Short:                      "Transactions subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		authcmd.GetSignCommand(),
		authcmd.GetSignBatchCommand(),
		authcmd.GetMultiSignCommand(),
		authcmd.GetMultiSignBatchCmd(),
		authcmd.GetValidateSignaturesCommand(),
		authcmd.GetBroadcastCommand(),
		authcmd.GetEncodeCommand(),
		authcmd.GetDecodeCommand(),
	)

	tempApp.BasicModuleManager.AddTxCommands(cmd)

	//app.ModuleBasics.AddTxCommands(cmd)
	cmd.PersistentFlags().String(flags.FlagChainID, "", "The network chain ID")

	return cmd
}
