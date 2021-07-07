package cli

import (
	"bufio"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/client/utils"

	"github.com/kava-labs/kava/x/swap/types"
)

// GetTxCmd returns the transaction commands for this module
func GetTxCmd(cdc *codec.Codec) *cobra.Command {
	swapTxCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("%s transactions subcommands", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	swapTxCmd.AddCommand(flags.PostCommands(
		getCmdDeposit(cdc),
	)...)

	return swapTxCmd
}

func getCmdDeposit(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "deposit [tokenA] [tokenB] [slippage] [deadline]",
		Short: "deposit coins to a swap liquidity pool",
		Example: fmt.Sprintf(
			`%s tx %s deposit 10000000ukava 10000000usdx 0.01 1624224736 --from <key>`, version.ClientName, types.ModuleName,
		),
		Args: cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			inBuf := bufio.NewReader(cmd.InOrStdin())
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			txBldr := auth.NewTxBuilderFromCLI(inBuf).WithTxEncoder(utils.GetTxEncoder(cdc))

			tokenA, err := sdk.ParseCoin(args[0])
			if err != nil {
				return err
			}

			tokenB, err := sdk.ParseCoin(args[1])
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

			msg := types.NewMsgDeposit(cliCtx.GetFromAddress(), tokenA, tokenB, slippage, deadline)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}

func getCmdWithdraw(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "withdraw [shares] [minCoinA] [minCoinB] [deadline]",
		Short: "withdraw coins from a swap liquidity pool",
		Example: fmt.Sprintf(
			`%s tx %s withdraw 153000 10000000ukava 20000000usdx 176293740 --from <key>`, version.ClientName, types.ModuleName,
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			inBuf := bufio.NewReader(cmd.InOrStdin())
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			txBldr := auth.NewTxBuilderFromCLI(inBuf).WithTxEncoder(utils.GetTxEncoder(cdc))

			numShares, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return err
			}
			shares := sdk.NewInt(numShares)

			minTokenA, err := sdk.ParseCoin(args[1])
			if err != nil {
				return err
			}

			minTokenB, err := sdk.ParseCoin(args[2])
			if err != nil {
				return err
			}

			deadline, err := strconv.ParseInt(args[3], 10, 64)
			if err != nil {
				return err
			}

			msg := types.NewMsgWithdraw(cliCtx.GetFromAddress(), shares, minTokenA, minTokenB, deadline)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}
