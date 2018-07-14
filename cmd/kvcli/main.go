package main

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/tendermint/tmlibs/cli"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/keys"
	"github.com/cosmos/cosmos-sdk/client/rpc"
	"github.com/cosmos/cosmos-sdk/client/tx"

	"github.com/cosmos/cosmos-sdk/version"
	authcmd "github.com/cosmos/cosmos-sdk/x/auth/client/cli"
	bankcmd "github.com/cosmos/cosmos-sdk/x/bank/client/cli"
	//ibccmd "github.com/cosmos/cosmos-sdk/x/ibc/client/cli"
	//stakecmd "github.com/cosmos/cosmos-sdk/x/stake/client/cli"
	paychancmd "github.com/kava-labs/kava/internal/x/paychan/client/cli"

	"github.com/kava-labs/kava/internal/app"
	"github.com/kava-labs/kava/internal/lcd"
	//"github.com/kava-labs/kava/internal/types"
)

// rootCmd is the entry point for this binary
var (
	rootCmd = &cobra.Command{
		Use:   "kvcli",
		Short: "Kava Light-Client",
	}
)

func main() {
	// disable sorting
	cobra.EnableCommandSorting = false

	// get the codec
	cdc := app.MakeCodec()

	// TODO: setup keybase, viper object, etc. to be passed into
	// the below functions and eliminate global vars, like we do
	// with the cdc

	// add standard rpc, and tx commands
	rpc.AddCommands(rootCmd)
	rootCmd.AddCommand(client.LineBreak)
	tx.AddCommands(rootCmd, cdc)
	rootCmd.AddCommand(client.LineBreak)

	// add query/post commands (custom to binary)
	rootCmd.AddCommand(
		client.GetCommands(
			authcmd.GetAccountCmd("acc", cdc, authcmd.GetAccountDecoder(cdc)),
		)...)

	rootCmd.AddCommand(
		client.PostCommands( // this just wraps the input cmds with common flags
			bankcmd.SendTxCmd(cdc),
			//ibccmd.IBCTransferCmd(cdc),
			//ibccmd.IBCRelayCmd(cdc),
			//stakecmd.GetCmdCreateValidator(cdc),
			//stakecmd.GetCmdEditValidator(cdc),
			//stakecmd.GetCmdDelegate(cdc),
			//stakecmd.GetCmdUnbond(cdc),
		)...)

	paychanCmd := &cobra.Command{
		Use:   "paychan",
		Short: "Payment channel subcommands",
	}
	stakeCmd.AddCommand(
		client.PostCommands(
			paychancmd.CreatePaychanCmd(cdc),
			paychancmd.GenerateNewStateCmd(cdc),
			paychancmd.ClosePaychanCmd(cdc),
		)...)
	rootCmd.AddCommand(
		paychanCmd,
	)
	// add proxy, version and key info
	rootCmd.AddCommand(
		client.LineBreak,
		lcd.ServeCommand(cdc),
		keys.Commands(),
		client.LineBreak,
		version.VersionCmd,
	)

	// prepare and add flags
	executor := cli.PrepareMainCmd(rootCmd, "KV", os.ExpandEnv("$HOME/.kvcli"))
	executor.Execute()
}
