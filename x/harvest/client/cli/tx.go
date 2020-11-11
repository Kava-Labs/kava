package cli

import (
	"bufio"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/client/utils"

	"github.com/kava-labs/kava/x/harvest/types"
)

// GetTxCmd returns the transaction commands for this module
func GetTxCmd(cdc *codec.Codec) *cobra.Command {
	harvestTxCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("%s transactions subcommands", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	harvestTxCmd.AddCommand(flags.PostCommands(
		getCmdDeposit(cdc),
		getCmdWithdraw(cdc),
		getCmdClaimReward(cdc),
		getCmdBorrow(cdc),
	)...)

	return harvestTxCmd
}

func getCmdDeposit(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "deposit [amount]",
		Short: "deposit coins to harvest",
		Example: fmt.Sprintf(
			`%s tx %s deposit 10000000bnb --from <key>`, version.ClientName, types.ModuleName,
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			inBuf := bufio.NewReader(cmd.InOrStdin())
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			txBldr := auth.NewTxBuilderFromCLI(inBuf).WithTxEncoder(utils.GetTxEncoder(cdc))
			amount, err := sdk.ParseCoin(args[0])
			if err != nil {
				return err
			}
			msg := types.NewMsgDeposit(cliCtx.GetFromAddress(), amount)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}
			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}

func getCmdWithdraw(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "withdraw [amount]",
		Short: "withdraw coins from harvest",
		Args:  cobra.ExactArgs(1),
		Example: fmt.Sprintf(
			`%s tx %s withdraw 10000000bnb --from <key>`, version.ClientName, types.ModuleName,
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			inBuf := bufio.NewReader(cmd.InOrStdin())
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			txBldr := auth.NewTxBuilderFromCLI(inBuf).WithTxEncoder(utils.GetTxEncoder(cdc))
			amount, err := sdk.ParseCoin(args[0])
			if err != nil {
				return err
			}
			msg := types.NewMsgWithdraw(cliCtx.GetFromAddress(), amount)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}
			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}

func getCmdClaimReward(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "claim [receiver-addr] [deposit-denom] [deposit-type] [multiplier]",
		Short: "claim HARD tokens to receiver address",
		Long: strings.TrimSpace(
			`sends accumulated HARD tokens from the harvest module account to the receiver address.
			Note that receiver address should match the sender address,
			unless the sender is a validator-vesting account`),
		Args: cobra.ExactArgs(4),
		Example: fmt.Sprintf(
			`%s tx %s claim kava1hgcfsuwc889wtdmt8pjy7qffua9dd2tralu64j bnb lp large --from <key>`, version.ClientName, types.ModuleName,
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			inBuf := bufio.NewReader(cmd.InOrStdin())
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			txBldr := auth.NewTxBuilderFromCLI(inBuf).WithTxEncoder(utils.GetTxEncoder(cdc))
			receiver, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}
			msg := types.NewMsgClaimReward(cliCtx.GetFromAddress(), receiver, args[1], args[2], args[3])
			if err := msg.ValidateBasic(); err != nil {
				return err
			}
			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}

func getCmdBorrow(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "borrow [1000000000ukava]",
		Short: "borrow tokens from the harvest protocol",
		Long:  strings.TrimSpace(`borrows tokens from the harvest protocol`),
		Args:  cobra.ExactArgs(1),
		Example: fmt.Sprintf(
			`%s tx %s borrow 1000000000ukava --from <key>`, version.ClientName, types.ModuleName,
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			inBuf := bufio.NewReader(cmd.InOrStdin())
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			txBldr := auth.NewTxBuilderFromCLI(inBuf).WithTxEncoder(utils.GetTxEncoder(cdc))

			coins, err := sdk.ParseCoins(args[0])
			if err != nil {
				return err
			}

			msg := types.NewMsgBorrow(cliCtx.GetFromAddress(), coins)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}
			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}
