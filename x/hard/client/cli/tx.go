package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"

	"github.com/kava-labs/kava/x/hard/types"
)

// GetTxCmd returns the transaction commands for this module
func GetTxCmd() *cobra.Command {
	hardTxCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("%s transactions subcommands", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmds := []*cobra.Command{
		getCmdDeposit(),
		getCmdWithdraw(),
		getCmdBorrow(),
		getCmdRepay(),
		getCmdLiquidate(),
	}

	for _, cmd := range cmds {
		flags.AddTxFlagsToCmd(cmd)
	}

	hardTxCmd.AddCommand(cmds...)

	return hardTxCmd
}

func getCmdDeposit() *cobra.Command {
	return &cobra.Command{
		Use:   "deposit [amount]",
		Short: "deposit coins to hard",
		Example: fmt.Sprintf(
			`%s tx %s deposit 10000000bnb --from <key>`, version.AppName, types.ModuleName,
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			amount, err := sdk.ParseCoinsNormalized(args[0])
			if err != nil {
				return err
			}
			msg := types.NewMsgDeposit(clientCtx.GetFromAddress(), amount)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), &msg)
		},
	}
}

func getCmdWithdraw() *cobra.Command {
	return &cobra.Command{
		Use:   "withdraw [amount]",
		Short: "withdraw coins from hard",
		Args:  cobra.ExactArgs(1),
		Example: fmt.Sprintf(
			`%s tx %s withdraw 10000000bnb --from <key>`, version.AppName, types.ModuleName,
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			amount, err := sdk.ParseCoinsNormalized(args[0])
			if err != nil {
				return err
			}
			msg := types.NewMsgWithdraw(clientCtx.GetFromAddress(), amount)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), &msg)
		},
	}
}

func getCmdBorrow() *cobra.Command {
	return &cobra.Command{
		Use:   "borrow [amount]",
		Short: "borrow tokens from the hard protocol",
		Long:  strings.TrimSpace(`borrows tokens from the hard protocol`),
		Args:  cobra.ExactArgs(1),
		Example: fmt.Sprintf(
			`%s tx %s borrow 1000000000ukava --from <key>`, version.AppName, types.ModuleName,
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			coins, err := sdk.ParseCoinsNormalized(args[0])
			if err != nil {
				return err
			}

			msg := types.NewMsgBorrow(clientCtx.GetFromAddress(), coins)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), &msg)
		},
	}
}

func getCmdRepay() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "repay [amount]",
		Short: "repay tokens to the hard protocol",
		Long:  strings.TrimSpace(`repay tokens to the hard protocol with optional --owner param to repay another account's loan`),
		Args:  cobra.ExactArgs(1),
		Example: strings.TrimSpace(`
kvcli tx hard repay 1000000000ukava --from <key>
kvcli tx hard repay 1000000000ukava,25000000000bnb --from <key>
kvcli tx hard repay 1000000000ukava,25000000000bnb --owner <owner-address> --from <key>
		`),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			var owner sdk.AccAddress
			ownerStr, err := cmd.Flags().GetString(flagOwner)
			if err != nil {
				return err
			}

			// Parse optional owner argument or default to sender
			if len(ownerStr) > 0 {
				ownerAddr, err := sdk.AccAddressFromBech32(ownerStr)
				if err != nil {
					return err
				}
				owner = ownerAddr
			} else {
				owner = clientCtx.GetFromAddress()
			}

			coins, err := sdk.ParseCoinsNormalized(args[0])
			if err != nil {
				return err
			}

			msg := types.NewMsgRepay(clientCtx.GetFromAddress(), owner, coins)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), &msg)
		},
	}

	cmd.Flags().String(flagOwner, "", "original borrower's address whose loan will be repaid")

	return cmd
}

func getCmdLiquidate() *cobra.Command {
	return &cobra.Command{
		Use:   "liquidate [borrower-addr]",
		Short: "liquidate a borrower that's over their loan-to-value ratio",
		Long:  strings.TrimSpace(`liquidate a borrower that's over their loan-to-value ratio`),
		Args:  cobra.ExactArgs(1),
		Example: fmt.Sprintf(
			`%s tx %s liquidate kava1hgcfsuwc889wtdmt8pjy7qffua9dd2tralu64j --from <key>`, version.AppName, types.ModuleName,
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			borrower, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			msg := types.NewMsgLiquidate(clientCtx.GetFromAddress(), borrower)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), &msg)
		},
	}
}
