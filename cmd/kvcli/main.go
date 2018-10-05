// Copyright 2016 All in Bits, inc
// Modifications copyright 2018 Kava Labs

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

	slashingcmd "github.com/cosmos/cosmos-sdk/x/slashing/client/cli"
	stakecmd "github.com/cosmos/cosmos-sdk/x/stake/client/cli"
	paychancmd "github.com/kava-labs/kava/internal/x/paychan/client/cli"

	"github.com/kava-labs/kava/internal/app"
	"github.com/kava-labs/kava/internal/lcd"
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

	// Add state commands
	tendermintCmd := &cobra.Command{
		Use:   "tendermint",
		Short: "Tendermint state querying subcommands",
	}
	tendermintCmd.AddCommand(
		rpc.BlockCommand(),
		rpc.ValidatorCommand(),
	)
	tx.AddCommands(tendermintCmd, cdc)

	advancedCmd := &cobra.Command{
		Use:   "advanced",
		Short: "Advanced subcommands",
	}

	advancedCmd.AddCommand(
		tendermintCmd,
		lcd.ServeCommand(cdc),
	)
	rootCmd.AddCommand(
		advancedCmd,
		client.LineBreak,
	)

	// Add stake commands
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
			stakecmd.GetCmdQueryUnbondingDelegation("stake", cdc),
			stakecmd.GetCmdQueryUnbondingDelegations("stake", cdc),
			stakecmd.GetCmdQueryRedelegation("stake", cdc),
			stakecmd.GetCmdQueryRedelegations("stake", cdc),
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

	// Add auth and bank commands
	rootCmd.AddCommand(
		client.GetCommands(
			authcmd.GetAccountCmd("acc", cdc, authcmd.GetAccountDecoder(cdc)),
		)...)
	rootCmd.AddCommand(
		client.PostCommands( // this just wraps the input cmds with common flags
			bankcmd.SendTxCmd(cdc),
		)...)

	// Add paychan commands
	paychanCmd := &cobra.Command{
		Use:   "paychan",
		Short: "Payment channel subcommand",
	}
	paychanCmd.AddCommand(
		client.PostCommands(
			paychancmd.CreateChannelCmd(cdc),
			paychancmd.GetChannelCmd(cdc, "paychan"), // pass in storeKey
			paychancmd.GeneratePaymentCmd(cdc),
			paychancmd.VerifyPaymentCmd(cdc, "paychan"), // pass in storeKey
			paychancmd.SubmitPaymentCmd(cdc),
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
