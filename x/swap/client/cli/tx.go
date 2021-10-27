package cli

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"

	"github.com/kava-labs/kava/x/swap/types"
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
		getCmdSwapExactForTokens(),
		getCmdSwapForExactTokens(),
	}

	for _, cmd := range cmds {
		flags.AddTxFlagsToCmd(cmd)
	}

	swapTxCmd.AddCommand(cmds...)

	return swapTxCmd
}

func getCmdDeposit() *cobra.Command {
	return &cobra.Command{
		Use:   "deposit [tokenA] [tokenB] [slippage] [deadline]",
		Short: "deposit coins to a swap liquidity pool",
		Example: fmt.Sprintf(
			`%s tx %s deposit 10000000ukava 10000000usdx 0.01 1624224736 --from <key>`,
			version.AppName, types.ModuleName,
		),
		Args: cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			tokenA, err := sdk.ParseCoinNormalized(args[0])
			if err != nil {
				return err
			}

			tokenB, err := sdk.ParseCoinNormalized(args[1])
			if err != nil {
				return err
			}

			slippage, err := sdk.NewDecFromStr(args[2])
			if err != nil {
				return err
			}

			deadline, err := strconv.ParseInt(args[3], 10, 64)
			if err != nil {
				return err
			}

			signer := clientCtx.GetFromAddress()
			msg := types.NewMsgDeposit(signer.String(), tokenA, tokenB, slippage, deadline)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
}

func getCmdWithdraw() *cobra.Command {
	return &cobra.Command{
		Use:   "withdraw [shares] [minCoinA] [minCoinB] [deadline]",
		Short: "withdraw coins from a swap liquidity pool",
		Example: fmt.Sprintf(
			`%s tx %s withdraw 153000 10000000ukava 20000000usdx 176293740 --from <key>`,
			version.AppName, types.ModuleName,
		),
		Args: cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			numShares, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return err
			}
			shares := sdk.NewInt(numShares)

			minTokenA, err := sdk.ParseCoinNormalized(args[1])
			if err != nil {
				return err
			}

			minTokenB, err := sdk.ParseCoinNormalized(args[2])
			if err != nil {
				return err
			}

			deadline, err := strconv.ParseInt(args[3], 10, 64)
			if err != nil {
				return err
			}

			fromAddr := clientCtx.GetFromAddress()
			msg := types.NewMsgWithdraw(fromAddr.String(), shares, minTokenA, minTokenB, deadline)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
}

func getCmdSwapExactForTokens() *cobra.Command {
	return &cobra.Command{
		Use:   "swap-exact-for-tokens [exactCoinA] [coinB] [slippage] [deadline]",
		Short: "swap an exact amount of token a for token b",
		Example: fmt.Sprintf(
			`%s tx %s swap-exact-for-tokens 1000000ukava 5000000usdx 0.01 1624224736 --from <key>`,
			version.AppName, types.ModuleName,
		),
		Args: cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			exactTokenA, err := sdk.ParseCoinNormalized(args[0])
			if err != nil {
				return err
			}

			tokenB, err := sdk.ParseCoinNormalized(args[1])
			if err != nil {
				return err
			}

			slippage, err := sdk.NewDecFromStr(args[2])
			if err != nil {
				return err
			}

			deadline, err := strconv.ParseInt(args[3], 10, 64)
			if err != nil {
				return err
			}

			fromAddr := clientCtx.GetFromAddress()
			msg := types.NewMsgSwapExactForTokens(fromAddr.String(), exactTokenA, tokenB, slippage, deadline)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
}

func getCmdSwapForExactTokens() *cobra.Command {
	return &cobra.Command{
		Use:   "swap-for-exact-tokens [coinA] [exactCoinB] [slippage] [deadline]",
		Short: "swap token a for exact amount of token b",
		Example: fmt.Sprintf(
			`%s tx %s swap-for-exact-tokens 1000000ukava 5000000usdx 0.01 1624224736 --from <key>`,
			version.AppName, types.ModuleName,
		),
		Args: cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			tokenA, err := sdk.ParseCoinNormalized(args[0])
			if err != nil {
				return err
			}

			exactTokenB, err := sdk.ParseCoinNormalized(args[1])
			if err != nil {
				return err
			}

			slippage, err := sdk.NewDecFromStr(args[2])
			if err != nil {
				return err
			}

			deadline, err := strconv.ParseInt(args[3], 10, 64)
			if err != nil {
				return err
			}

			fromAddr := clientCtx.GetFromAddress()
			msg := types.NewMsgSwapForExactTokens(fromAddr.String(), tokenA, exactTokenB, slippage, deadline)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
}
