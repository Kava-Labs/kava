package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/client/utils"

	"github.com/kava-labs/kava/x/cdp/types"
)

// GetTxCmd returns the transaction commands for this module
func GetTxCmd(cdc *codec.Codec) *cobra.Command {
	cdpTxCmd := &cobra.Command{
		Use:   "cdp",
		Short: "cdp transactions subcommands",
	}

	cdpTxCmd.AddCommand(client.PostCommands(
		GetCmdCreateCdp(cdc),
		GetCmdDeposit(cdc),
		GetCmdWithdraw(cdc),
		GetCmdDraw(cdc),
		GetCmdRepay(cdc),
	)...)

	return cdpTxCmd
}

// GetCmdCreateCdp returns the command handler for creating a cdp
func GetCmdCreateCdp(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "create [collateral] [debt]",
		Short: "create a new cdp",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Create a new cdp, depositing some collateral and drawing some debt.

Example:
$ %s tx %s create 10000000uatom 1000usdx --from myKeyName
`, version.ClientName, types.ModuleName)),
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			txBldr := auth.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))

			collateral, err := sdk.ParseCoins(args[0])
			if err != nil {
				return err
			}
			debt, err := sdk.ParseCoins(args[1])
			if err != nil {
				return err
			}
			msg := types.NewMsgCreateCDP(cliCtx.GetFromAddress(), collateral, debt)
			err = msg.ValidateBasic()
			if err != nil {
				return err
			}
			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}

// GetCmdDeposit cli command for depositing to a cdp.
func GetCmdDeposit(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "deposit [owner-addr] [collateral]",
		Short: "deposit collateral to an existing cdp",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Add collateral to an existing cdp.

Example:
$ %s tx %s deposit kava15qdefkmwswysgg4qxgqpqr35k3m49pkx2jdfnw 10000000uatom --from myKeyName
`, version.ClientName, types.ModuleName)),
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			txBldr := auth.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))

			collateral, err := sdk.ParseCoins(args[1])
			if err != nil {
				return err
			}
			owner, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}
			msg := types.NewMsgDeposit(owner, cliCtx.GetFromAddress(), collateral)
			err = msg.ValidateBasic()
			if err != nil {
				return err
			}
			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}

// GetCmdWithdraw cli command for withdrawing from a cdp.
func GetCmdWithdraw(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "withdraw [owner-addr] [collateral]",
		Short: "withdraw collateral from an existing cdp",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Remove collateral from an existing cdp.

Example:
$ %s tx %s withdraw kava15qdefkmwswysgg4qxgqpqr35k3m49pkx2jdfnw 10000000uatom --from myKeyName
`, version.ClientName, types.ModuleName)),
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			txBldr := auth.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))

			collateral, err := sdk.ParseCoins(args[1])
			if err != nil {
				return err
			}
			owner, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}
			msg := types.NewMsgWithdraw(owner, cliCtx.GetFromAddress(), collateral)
			err = msg.ValidateBasic()
			if err != nil {
				return err
			}
			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}

// GetCmdDraw cli command for depositing to a cdp.
func GetCmdDraw(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "draw [collateral-name] [debt]",
		Short: "draw debt off an existing cdp",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Create debt in an existing cdp and send the newly minted asset to your account.

Example:
$ %s tx %s draw uatom 1000usdx --from myKeyName
`, version.ClientName, types.ModuleName)),
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			txBldr := auth.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))

			debt, err := sdk.ParseCoins(args[1])
			if err != nil {
				return err
			}
			msg := types.NewMsgDrawDebt(cliCtx.GetFromAddress(), args[0], debt)
			err = msg.ValidateBasic()
			if err != nil {
				return err
			}
			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}

// GetCmdRepay cli command for depositing to a cdp.
func GetCmdRepay(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "repay [collateral-name] [debt]",
		Short: "repay debt to an existing cdp",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Cancel out debt in an existing cdp by paying it down.

Example:
$ %s tx %s repay uatom 1000usdx --from myKeyName
`, version.ClientName, types.ModuleName)),
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			txBldr := auth.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))

			payment, err := sdk.ParseCoins(args[1])
			if err != nil {
				return err
			}
			msg := types.NewMsgRepayDebt(cliCtx.GetFromAddress(), args[0], payment)
			err = msg.ValidateBasic()
			if err != nil {
				return err
			}
			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}
