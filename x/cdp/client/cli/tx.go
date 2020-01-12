package cli

import (
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
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
		Use:   "create [ownerAddress] [collateralChange] [debtChange]",
		Short: "create a new cdp",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			txBldr := auth.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))
			collateral, err := sdk.ParseCoins(args[1])
			if err != nil {
				return err
			}
			debt, err := sdk.ParseCoins(args[2])
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
		Use:   "deposit [ownerAddress] [depositorAddress] [collateralChange]",
		Short: "deposit to an existing cdp",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			txBldr := auth.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))
			collateral, err := sdk.ParseCoins(args[2])
			if err != nil {
				return err
			}
			owner, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}
			depositor, err := sdk.AccAddressFromBech32(args[1])
			if err != nil {
				return err
			}
			msg := types.NewMsgDeposit(owner, depositor, collateral)
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
		Use:   "withdraw [ownerAddress] [depositorAddress] [collateralChange]",
		Short: "withdraw from an existing cdp",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			txBldr := auth.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))
			collateral, err := sdk.ParseCoins(args[2])
			if err != nil {
				return err
			}
			owner, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}
			depositor, err := sdk.AccAddressFromBech32(args[1])
			if err != nil {
				return err
			}
			msg := types.NewMsgWithdraw(owner, depositor, collateral)
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
		Use:   "draw [ownerAddress] [collateralDenom] [debtChange]",
		Short: "draw debt off an existing cdp",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			txBldr := auth.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))
			debt, err := sdk.ParseCoins(args[2])
			if err != nil {
				return err
			}
			owner, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}
			msg := types.NewMsgDrawDebt(owner, args[1], debt)
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
		Use:   "repay [ownerAddress] [collateralDenom] [payment]",
		Short: "repay debt from an existing cdp",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			txBldr := auth.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))
			payment, err := sdk.ParseCoins(args[2])
			if err != nil {
				return err
			}
			owner, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}
			msg := types.NewMsgRepayDebt(owner, args[1], payment)
			err = msg.ValidateBasic()
			if err != nil {
				return err
			}
			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}
