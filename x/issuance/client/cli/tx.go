package cli

import (
	"bufio"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/client/utils"

	"github.com/kava-labs/kava/x/issuance/types"
)

// GetTxCmd returns the transaction cli commands for the issuance module
func GetTxCmd(cdc *codec.Codec) *cobra.Command {
	issuanceTxCmd := &cobra.Command{
		Use:   types.ModuleName,
		Short: "transaction commands for the issuance module",
	}

	issuanceTxCmd.AddCommand(flags.PostCommands(
		getCmdIssueTokens(cdc),
		getCmdRedeemTokens(cdc),
		getCmdBlockAddress(cdc),
		getCmdUnblockAddress(cdc),
		getCmdPauseAsset(cdc),
	)...)

	return issuanceTxCmd

}

func getCmdIssueTokens(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "issue [tokens] [receiver]",
		Short: "issue new tokens to the receiver address",
		Long:  "The asset owner issues new tokens that will be credited to the receiver address",
		Example: fmt.Sprintf(`$ %s tx %s issue 20000000usdtoken kava15qdefkmwswysgg4qxgqpqr35k3m49pkx2jdfnw
		`, version.ClientName, types.ModuleName),
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			inBuf := bufio.NewReader(cmd.InOrStdin())
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			txBldr := auth.NewTxBuilderFromCLI(inBuf).WithTxEncoder(utils.GetTxEncoder(cdc))
			tokens, err := sdk.ParseCoin(args[0])
			if err != nil {
				return err
			}
			receiver, err := sdk.AccAddressFromBech32(args[1])
			if err != nil {
				return err
			}

			msg := types.NewMsgIssueTokens(cliCtx.GetFromAddress(), tokens, receiver)
			err = msg.ValidateBasic()
			if err != nil {
				return err
			}
			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}

func getCmdRedeemTokens(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "redeem [tokens]",
		Short: "redeem tokens",
		Long:  "The asset owner redeems (burns) tokens, removing them from the circulating supply",
		Example: fmt.Sprintf(`$ %s tx %s redeem 20000000usdtoken
		`, version.ClientName, types.ModuleName),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			inBuf := bufio.NewReader(cmd.InOrStdin())
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			txBldr := auth.NewTxBuilderFromCLI(inBuf).WithTxEncoder(utils.GetTxEncoder(cdc))
			tokens, err := sdk.ParseCoin(args[0])
			if err != nil {
				return err
			}

			msg := types.NewMsgRedeemTokens(cliCtx.GetFromAddress(), tokens)
			err = msg.ValidateBasic()
			if err != nil {
				return err
			}
			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}

func getCmdBlockAddress(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "block [address] [denom]",
		Short: "block an address for the input denom",
		Long:  "The asset owner blocks an address from holding coins of that denomination. Any tokens of the input denomination held by the address will be sent to the owner address",
		Example: fmt.Sprintf(`$ %s tx %s block kava15qdefkmwswysgg4qxgqpqr35k3m49pkx2jdfnw usdtoken
		`, version.ClientName, types.ModuleName),
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			inBuf := bufio.NewReader(cmd.InOrStdin())
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			txBldr := auth.NewTxBuilderFromCLI(inBuf).WithTxEncoder(utils.GetTxEncoder(cdc))
			address, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			err = sdk.ValidateDenom(args[1])
			if err != nil {
				return err
			}
			msg := types.NewMsgBlockAddress(cliCtx.GetFromAddress(), args[1], address)
			err = msg.ValidateBasic()
			if err != nil {
				return err
			}
			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}

func getCmdUnblockAddress(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "unblock [address] [denom]",
		Short: "unblock an address for the input denom",
		Long:  "The asset owner unblocks an address from holding coins of that denomination.",
		Example: fmt.Sprintf(`$ %s tx %s unblock kava15qdefkmwswysgg4qxgqpqr35k3m49pkx2jdfnw usdtoken
		`, version.ClientName, types.ModuleName),
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			inBuf := bufio.NewReader(cmd.InOrStdin())
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			txBldr := auth.NewTxBuilderFromCLI(inBuf).WithTxEncoder(utils.GetTxEncoder(cdc))
			address, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			err = sdk.ValidateDenom(args[1])
			if err != nil {
				return err
			}
			msg := types.NewMsgUnblockAddress(cliCtx.GetFromAddress(), args[1], address)
			err = msg.ValidateBasic()
			if err != nil {
				return err
			}
			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}

func getCmdPauseAsset(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "set-pause-status [denom] [status]",
		Short: "pause or unpause an asset",
		Long:  "The asset owner pauses or unpauses the input asset, halting new issuance and redemption",
		Example: fmt.Sprintf(`$ %s tx %s pause usdtoken true
		`, version.ClientName, types.ModuleName),
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			inBuf := bufio.NewReader(cmd.InOrStdin())
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			txBldr := auth.NewTxBuilderFromCLI(inBuf).WithTxEncoder(utils.GetTxEncoder(cdc))
			err := sdk.ValidateDenom(args[0])
			if err != nil {
				return err
			}
			var status bool
			if args[1] == "true" {
				status = true
			} else if args[1] == "false" {
				status = false
			} else {
				return fmt.Errorf(fmt.Sprintf("status must be true or false, got %s", args[1]))
			}

			msg := types.NewMsgSetPauseStatus(cliCtx.GetFromAddress(), args[0], status)
			err = msg.ValidateBasic()
			if err != nil {
				return err
			}
			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}
