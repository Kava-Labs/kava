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
		Use:   "deposit [tokenA] [tokenB] [deadline]",
		Short: "deposit coins to a swap liquidity pool",
		Example: fmt.Sprintf(
			`%s tx %s deposit 10000000ukava 10000000usdx --from <key>`, version.ClientName, types.ModuleName,
		),
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

			deadline, err := strconv.ParseInt(args[2], 10, 64)
			if err != nil {
				return err
			}

			msg := types.NewMsgDeposit(cliCtx.GetFromAddress(), tokenA, tokenB, deadline)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}

func getCmdWithdraw(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "withdraw [pool-id] [shares] [slippage] [expected-coin-a] [expected-coin-b] [deadline]",
		Short: "withdraw coins from a swap liquidity pool",
		Example: fmt.Sprintf(
			`%s tx %s withdraw ukava/usdx 153000 0.05 10000000ukava 20000000usdx 176293740 --from <key>`, version.ClientName, types.ModuleName,
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			inBuf := bufio.NewReader(cmd.InOrStdin())
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			txBldr := auth.NewTxBuilderFromCLI(inBuf).WithTxEncoder(utils.GetTxEncoder(cdc))

			poolID := args[0]

			numShares, err := strconv.ParseInt(args[1], 10, 64)
			if err != nil {
				return err
			}
			shares := sdk.NewInt(numShares)

			slippage := sdk.MustNewDecFromStr(args[2])

			deadline, err := strconv.ParseInt(args[3], 10, 64)
			if err != nil {
				return err
			}

			expectedCoinA, err := sdk.ParseCoin(args[4])
			if err != nil {
				return err
			}

			expectedCoinB, err := sdk.ParseCoin(args[5])
			if err != nil {
				return err
			}

			msg := types.NewMsgWithdraw(cliCtx.GetFromAddress(), poolID, shares, slippage, expectedCoinA, expectedCoinB, deadline)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}
