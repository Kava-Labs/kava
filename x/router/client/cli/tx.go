package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/version"

	"github.com/kava-labs/kava/x/router/types"
)

// GetTxCmd returns the transaction commands for this module
func GetTxCmd() *cobra.Command {
	liquidTxCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "router transactions subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmds := []*cobra.Command{
		getCmdMintDeposit(),
		getCmdDelegateMintDeposit(),
		getCmdWithdrawBurn(),
		getCmdWithdrawBurnUndelegate(),
	}

	for _, cmd := range cmds {
		flags.AddTxFlagsToCmd(cmd)
	}

	liquidTxCmd.AddCommand(cmds...)

	return liquidTxCmd
}

func getCmdMintDeposit() *cobra.Command {
	return &cobra.Command{
		Use:   "mint-deposit [validator-addr] [amount]",
		Short: "mints staking derivative from a delegation and deposits them to earn",
		Args:  cobra.ExactArgs(2),
		Example: fmt.Sprintf(
			`%s tx %s mint-deposit kavavaloper16lnfpgn6llvn4fstg5nfrljj6aaxyee9z59jqd 10000000ukava --from <key>`, version.AppName, types.ModuleName,
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			valAddr, err := sdk.ValAddressFromBech32(args[0])
			if err != nil {
				return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, err.Error())
			}

			coin, err := sdk.ParseCoinNormalized(args[1])
			if err != nil {
				return err
			}

			msg := types.NewMsgMintDeposit(clientCtx.GetFromAddress(), valAddr, coin)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
}

func getCmdDelegateMintDeposit() *cobra.Command {
	return &cobra.Command{
		Use:   "delegate-mint-deposit [validator-addr] [amount]",
		Short: "delegates tokens, mints staking derivative from a them, and deposits them to earn",
		Args:  cobra.ExactArgs(2),
		Example: fmt.Sprintf(
			`%s tx %s delegate-mint-deposit kavavaloper16lnfpgn6llvn4fstg5nfrljj6aaxyee9z59jqd 10000000ukava --from <key>`, version.AppName, types.ModuleName,
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			valAddr, err := sdk.ValAddressFromBech32(args[0])
			if err != nil {
				return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, err.Error())
			}

			coin, err := sdk.ParseCoinNormalized(args[1])
			if err != nil {
				return err
			}

			msg := types.NewMsgDelegateMintDeposit(clientCtx.GetFromAddress(), valAddr, coin)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
}

func getCmdWithdrawBurn() *cobra.Command {
	return &cobra.Command{
		Use:   "withdraw-burn [validator-addr] [amount]",
		Short: "withdraws staking derivatives from earn and burns them to redeem a delegation",
		Example: fmt.Sprintf(
			`%s tx %s withdraw-burn kavavaloper16lnfpgn6llvn4fstg5nfrljj6aaxyee9z59jqd 10000000ukava --from <key>`, version.AppName, types.ModuleName,
		),
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			valAddr, err := sdk.ValAddressFromBech32(args[0])
			if err != nil {
				return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, err.Error())
			}

			amount, err := sdk.ParseCoinNormalized(args[1])
			if err != nil {
				return err
			}

			msg := types.NewMsgWithdrawBurn(clientCtx.GetFromAddress(), valAddr, amount)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
}

func getCmdWithdrawBurnUndelegate() *cobra.Command {
	return &cobra.Command{
		Use:   "withdraw-burn-undelegate [validator-addr] [amount]",
		Short: "withdraws staking derivatives from earn, burns them to redeem a delegation, then unstakes the delegation",
		Example: fmt.Sprintf(
			`%s tx %s withdraw-burn-undelegate kavavaloper16lnfpgn6llvn4fstg5nfrljj6aaxyee9z59jqd 10000000ukava --from <key>`, version.AppName, types.ModuleName,
		),
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			valAddr, err := sdk.ValAddressFromBech32(args[0])
			if err != nil {
				return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, err.Error())
			}

			amount, err := sdk.ParseCoinNormalized(args[1])
			if err != nil {
				return err
			}

			msg := types.NewMsgWithdrawBurnUndelegate(clientCtx.GetFromAddress(), valAddr, amount)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
}
