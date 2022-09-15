package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/version"

	"github.com/kava-labs/kava/x/liquid/types"
)

// GetTxCmd returns the transaction commands for this module
func GetTxCmd() *cobra.Command {
	liquidTxCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "liquid transactions subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmds := []*cobra.Command{
		getCmdMintDerivative(),
		getCmdBurnDerivative(),
	}

	for _, cmd := range cmds {
		flags.AddTxFlagsToCmd(cmd)
	}

	liquidTxCmd.AddCommand(cmds...)

	return liquidTxCmd
}

func getCmdMintDerivative() *cobra.Command {
	return &cobra.Command{
		Use:   "mint [validator-addr] [amount]",
		Short: "mints staking derivative from a delegation",
		Long:  "Mint removes a portion of a user's staking delegation and issues them validator specific staking derivative tokens.",
		Args:  cobra.ExactArgs(2),
		Example: fmt.Sprintf(
			`%s tx %s mint kavavaloper16lnfpgn6llvn4fstg5nfrljj6aaxyee9z59jqd 10000000ukava --from <key>`, version.AppName, types.ModuleName,
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

			msg := types.NewMsgMintDerivative(clientCtx.GetFromAddress(), valAddr, coin)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), &msg)
		},
	}
}

func getCmdBurnDerivative() *cobra.Command {
	return &cobra.Command{
		Use:   "burn [amount]",
		Short: "burns staking derivative to redeem a delegation",
		Long:  "Burn removes some staking derivative from a user's account and converts it back to a staking delegation.",
		Example: fmt.Sprintf(
			`%s tx %s burn 10000000bkava-kavavaloper16lnfpgn6llvn4fstg5nfrljj6aaxyee9z59jqd --from <key>`, version.AppName, types.ModuleName,
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

			valAddr, err := parseLiquidStakingTokenDenom(amount.Denom)
			if err != nil {
				return sdkerrors.Wrap(types.ErrInvalidDenom, err.Error())
			}

			msg := types.NewMsgBurnDerivative(clientCtx.GetFromAddress(), valAddr, amount)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), &msg)
		},
	}
}

// parseLiquidStakingTokenDenom extracts a validator address from a derivative denom.
func parseLiquidStakingTokenDenom(denom string) (sdk.ValAddress, error) {
	elements := strings.Split(denom, types.DenomSeparator)
	if len(elements) != 2 {
		return nil, fmt.Errorf("cannot parse denom %s", denom)
	}
	addr, err := sdk.ValAddressFromBech32(elements[1])
	if err != nil {
		return nil, err
	}
	return addr, nil
}
