package main

import (
	"github.com/spf13/cobra"

	"github.com/tendermint/tendermint/libs/cli"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/keys"
	"github.com/cosmos/cosmos-sdk/client/rpc"
	"github.com/cosmos/cosmos-sdk/client/tx"

	"github.com/cosmos/cosmos-sdk/version"
	authcmd "github.com/cosmos/cosmos-sdk/x/auth/client/cli"
	bankcmd "github.com/cosmos/cosmos-sdk/x/bank/client/cli"
	//govcmd "github.com/cosmos/cosmos-sdk/x/gov/client/cli"
	//ibccmd "github.com/cosmos/cosmos-sdk/x/ibc/client/cli"
	slashingcmd "github.com/cosmos/cosmos-sdk/x/slashing/client/cli"
	stakecmd "github.com/cosmos/cosmos-sdk/x/stake/client/cli"
	paychancmd "github.com/kava-labs/kava/internal/x/paychan/client/cmd"

	"github.com/kava-labs/kava/internal/app"
	//"github.com/kava-labs/kava/internal/lcd"
)

var (
	rootCmd = &cobra.Command{
		Use:   "kvcli",
		Short: "Kava Light-Client",
	}
)

func main() {
	cobra.EnableCommandSorting = false

	// get the codec
	cdc := app.CreateKavaAppCodec()

	// add standard rpc commands
	rpc.AddCommands(rootCmd)

	//Add state commands
	tendermintCmd := &cobra.Command{
		Use:   "tendermint",
		Short: "Tendermint state querying subcommands",
	}
	tendermintCmd.AddCommand(
		rpc.BlockCommand(),
		rpc.ValidatorCommand(),
	)
	tx.AddCommands(tendermintCmd, cdc)

	//Add IBC commands
	// ibcCmd := &cobra.Command{
	// 	Use:   "ibc",
	// 	Short: "Inter-Blockchain Communication subcommands",
	// }
	// ibcCmd.AddCommand(
	// 	client.PostCommands(
	// 		ibccmd.IBCTransferCmd(cdc),
	// 		ibccmd.IBCRelayCmd(cdc),
	// 	)...)

	advancedCmd := &cobra.Command{
		Use:   "advanced",
		Short: "Advanced subcommands",
	}

	advancedCmd.AddCommand(
		tendermintCmd,
		//ibcCmd,
		//lcd.ServeCommand(cdc),
	)
	rootCmd.AddCommand(
		advancedCmd,
		client.LineBreak,
	)

	//Add stake commands
	stakeCmd := &cobra.Command{
		Use:   "stake",
		Short: "Stake and validation subcommands",
	}
	stakeCmd.AddCommand(
		client.GetCommands(
			stakecmd.GetCmdQueryValidator("stake", cdc),
			stakecmd.GetCmdQueryValidators("stake", cdc),
			stakecmd.GetCmdQueryDelegation("stake", cdc),
			stakecmd.GetCmdQueryDelegations("stake", cdc),
			slashingcmd.GetCmdQuerySigningInfo("slashing", cdc),
		)...)
	stakeCmd.AddCommand(
		client.PostCommands(
			stakecmd.GetCmdCreateValidator(cdc),
			stakecmd.GetCmdEditValidator(cdc),
			stakecmd.GetCmdDelegate(cdc),
			stakecmd.GetCmdUnbond("stake", cdc),
			stakecmd.GetCmdRedelegate("stake", cdc),
			slashingcmd.GetCmdUnrevoke(cdc),
		)...)
	rootCmd.AddCommand(
		stakeCmd,
	)

	//Add stake commands
	// govCmd := &cobra.Command{
	// 	Use:   "gov",
	// 	Short: "Governance and voting subcommands",
	// }
	// govCmd.AddCommand(
	// 	client.GetCommands(
	// 		govcmd.GetCmdQueryProposal("gov", cdc),
	// 		govcmd.GetCmdQueryVote("gov", cdc),
	// 		govcmd.GetCmdQueryVotes("gov", cdc),
	// 	)...)
	// govCmd.AddCommand(
	// 	client.PostCommands(
	// 		govcmd.GetCmdSubmitProposal(cdc),
	// 		govcmd.GetCmdDeposit(cdc),
	// 		govcmd.GetCmdVote(cdc),
	// 	)...)
	// rootCmd.AddCommand(
	// 	govCmd,
	// )

	//Add auth and bank commands
	rootCmd.AddCommand(
		client.GetCommands(
			authcmd.GetAccountCmd("acc", cdc, authcmd.GetAccountDecoder(cdc)),
		)...)
	rootCmd.AddCommand(
		client.PostCommands( // this just wraps the input cmds with common flags
			bankcmd.SendTxCmd(cdc),
		)...)

	paychanCmd := &cobra.Command{
		Use:   "paychan",
		Short: "Payment channel subcommands",
	}
	paychanCmd.AddCommand(
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
		keys.Commands(),
		client.LineBreak,
		version.VersionCmd,
	)

	// prepare and add flags
	executor := cli.PrepareMainCmd(rootCmd, "KV", app.DefaultCLIHome)
	err := executor.Execute()
	if err != nil {
		panic(err)
	}
}
