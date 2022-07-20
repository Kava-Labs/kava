package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"

	"github.com/kava-labs/kava/x/earn/types"
)

// GetTxCmd returns the transaction commands for this module
func GetTxCmd() *cobra.Command {
	swapTxCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("%s transactions subcommands", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmds := []*cobra.Command{
		getCmdDeposit(),
		getCmdWithdraw(),
	}

	for _, cmd := range cmds {
		flags.AddTxFlagsToCmd(cmd)
	}

	swapTxCmd.AddCommand(cmds...)

	return swapTxCmd
}

func getCmdDeposit() *cobra.Command {
	return &cobra.Command{
		Use:   "deposit [amount]",
		Short: "deposit coins to an earn vault",
		Example: fmt.Sprintf(
			`%s tx %s deposit 10000000ukava --from <key>`,
			version.AppName, types.ModuleName,
		),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			amount, err := sdk.ParseCoinNormalized(args[0])
			if err != nil {
				return err
			}

			signer := clientCtx.GetFromAddress()
			msg := types.NewMsgDeposit(signer.String(), amount)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
}

func getCmdWithdraw() *cobra.Command {
	return &cobra.Command{
		Use:   "withdraw [amount]",
		Short: "withdraw coins from an earn vault",
		Example: fmt.Sprintf(
			`%s tx %s withdraw 10000000ukava --from <key>`,
			version.AppName, types.ModuleName,
		),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			amount, err := sdk.ParseCoinNormalized(args[0])
			if err != nil {
				return err
			}

			fromAddr := clientCtx.GetFromAddress()
			msg := types.NewMsgWithdraw(fromAddr.String(), amount)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
}
