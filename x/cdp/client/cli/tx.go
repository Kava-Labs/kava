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

	"github.com/kava-labs/kava/x/cdp/types"
)

// GetTxCmd returns the transaction commands for this module
func GetTxCmd() *cobra.Command {
	cdpTxCmd := &cobra.Command{
		Use:   "cdp",
		Short: "cdp transactions subcommands",
	}

	cmds := []*cobra.Command{
		GetCmdCreateCdp(),
		GetCmdDeposit(),
		GetCmdWithdraw(),
		GetCmdDraw(),
		GetCmdRepay(),
		GetCmdLiquidate(),
	}

	for _, cmd := range cmds {
		flags.AddTxFlagsToCmd(cmd)
	}

	cdpTxCmd.AddCommand(cmds...)

	return cdpTxCmd
}

// GetCmdCreateCdp returns the command handler for creating a cdp
func GetCmdCreateCdp() *cobra.Command {
	return &cobra.Command{
		Use:   "create [collateral] [debt] [collateral-type]",
		Short: "create a new cdp",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Create a new cdp, depositing some collateral and drawing some debt.

Example:
$ %s tx %s create 10000000uatom 1000usdx atom-a --from myKeyName
`, version.AppName, types.ModuleName)),
		Args: cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			collateral, err := sdk.ParseCoinNormalized(args[0])
			if err != nil {
				return err
			}
			debt, err := sdk.ParseCoinNormalized(args[1])
			if err != nil {
				return err
			}
			msg := types.NewMsgCreateCDP(clientCtx.GetFromAddress(), collateral, debt, args[2])
			err = msg.ValidateBasic()
			if err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), &msg)
		},
	}
}

// GetCmdDeposit cli command for depositing to a cdp.
func GetCmdDeposit() *cobra.Command {
	return &cobra.Command{
		Use:   "deposit [owner-addr] [collateral] [collateral-type]",
		Short: "deposit collateral to an existing cdp",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Add collateral to an existing cdp.

Example:
$ %s tx %s deposit kava15qdefkmwswysgg4qxgqpqr35k3m49pkx2jdfnw 10000000uatom atom-a --from myKeyName
`, version.AppName, types.ModuleName)),
		Args: cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			collateral, err := sdk.ParseCoinNormalized(args[1])
			if err != nil {
				return err
			}
			owner, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}
			msg := types.NewMsgDeposit(owner, clientCtx.GetFromAddress(), collateral, args[2])
			err = msg.ValidateBasic()
			if err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), &msg)
		},
	}
}

// GetCmdWithdraw cli command for withdrawing from a cdp.
func GetCmdWithdraw() *cobra.Command {
	return &cobra.Command{
		Use:   "withdraw [owner-addr] [collateral] [collateral-type]",
		Short: "withdraw collateral from an existing cdp",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Remove collateral from an existing cdp.

Example:
$ %s tx %s withdraw kava15qdefkmwswysgg4qxgqpqr35k3m49pkx2jdfnw 10000000uatom atom-a --from myKeyName
`, version.AppName, types.ModuleName)),
		Args: cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			collateral, err := sdk.ParseCoinNormalized(args[1])
			if err != nil {
				return err
			}
			owner, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}
			msg := types.NewMsgWithdraw(owner, clientCtx.GetFromAddress(), collateral, args[2])
			err = msg.ValidateBasic()
			if err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), &msg)
		},
	}
}

// GetCmdDraw cli command for depositing to a cdp.
func GetCmdDraw() *cobra.Command {
	return &cobra.Command{
		Use:   "draw [collateral-type] [debt]",
		Short: "draw debt off an existing cdp",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Create debt in an existing cdp and send the newly minted asset to your account.

Example:
$ %s tx %s draw atom-a 1000usdx --from myKeyName
`, version.AppName, types.ModuleName)),
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			debt, err := sdk.ParseCoinNormalized(args[1])
			if err != nil {
				return err
			}
			msg := types.NewMsgDrawDebt(clientCtx.GetFromAddress(), args[0], debt)
			err = msg.ValidateBasic()
			if err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), &msg)
		},
	}
}

// GetCmdRepay cli command for depositing to a cdp.
func GetCmdRepay() *cobra.Command {
	return &cobra.Command{
		Use:   "repay [collateral-name] [debt]",
		Short: "repay debt to an existing cdp",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Cancel out debt in an existing cdp.

Example:
$ %s tx %s repay atom-a 1000usdx --from myKeyName
`, version.AppName, types.ModuleName)),
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			payment, err := sdk.ParseCoinNormalized(args[1])
			if err != nil {
				return err
			}
			msg := types.NewMsgRepayDebt(clientCtx.GetFromAddress(), args[0], payment)
			err = msg.ValidateBasic()
			if err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), &msg)
		},
	}
}

// GetCmdLiquidate cli command for liquidating a cdp.
func GetCmdLiquidate() *cobra.Command {
	return &cobra.Command{
		Use:   "liquidate [cdp-owner-address] [collateral-type]",
		Short: "liquidate a cdp",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Liquidate a cdp if it is below the required liquidation ratio

Example:
$ %s tx %s liquidate kava1y70y90wzmnf00e63efk2lycgqwepthdmyzsfzm  btcb-a --from myKeyName
`, version.AppName, types.ModuleName)),
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			addr, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}
			msg := types.NewMsgLiquidate(clientCtx.GetFromAddress(), addr, args[1])
			err = msg.ValidateBasic()
			if err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), &msg)
		},
	}
}
