package cli

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"

	"github.com/kava-labs/kava/x/evmutil/types"
)

// GetTxCmd returns the transaction commands for this module
func GetTxCmd() *cobra.Command {
	txCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("%s transactions subcommands", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmds := []*cobra.Command{
		getCmdMsgConvertCoinToERC20(),
		getCmdConvertERC20ToCoin(),
	}

	for _, cmd := range cmds {
		flags.AddTxFlagsToCmd(cmd)
	}

	txCmd.AddCommand(cmds...)

	return txCmd
}

func getCmdMsgConvertCoinToERC20() *cobra.Command {
	return &cobra.Command{
		Use:   "convert-coin-to-erc20 [receiver] [coin]",
		Short: "converts sdk.Coin to erc20 tokens on Kava eth co-chain",
		Example: fmt.Sprintf(
			`%s tx %s convert-coin-to-erc20 0x6B1088f788b412Ad1280F95240d56B886A64bc05 100000000weth --from <key>`,
			version.AppName, types.ModuleName,
		),
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			receiver := args[0]
			if !common.IsHexAddress(receiver) {
				return fmt.Errorf("receiver '%s' is an invalid hex address", args[0])
			}

			coin, err := sdk.ParseCoinNormalized(args[1])
			if err != nil {
				return err
			}

			signer := clientCtx.GetFromAddress()
			msg := types.NewMsgConvertCoinToERC20(signer.String(), receiver, coin)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), &msg)
		},
	}
}

func getCmdConvertERC20ToCoin() *cobra.Command {
	return &cobra.Command{
		Use:   "convert-erc20-to-coin [Kava receiver address] [Kava ERC20 address or Denom] [amount]",
		Short: "burns ERC20 tokens on Kava EVM co-chain and unlocks on Ethereum",
		Example: fmt.Sprintf(`
%[1]s tx %[2]s convert-erc20-to-coin 0x8223259205A3E31C54469fCbfc9F7Cf83D515ff6 0x21E360e198Cde35740e88572B59f2CAdE421E6b1 1000000000000000 --from <key>
%[1]s tx %[2]s convert-erc20-to-coin kava10wlnqzyss4accfqmyxwx5jy5x9nfkwh6qm7n4t erc20/weth 1000000000000000 --from <key>
`,
			version.AppName, types.ModuleName,
		),
		Args: cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			if err := CanSignEthTx(clientCtx); err != nil {
				return err
			}

			receiver, err := ParseAddrFromHexOrBech32(args[0])
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)
			contractAddr, err := ParseOrQueryConversionPairAddress(queryClient, args[1])
			if err != nil {
				return err
			}

			amount, ok := new(big.Int).SetString(args[2], 10)
			if !ok {
				return fmt.Errorf("amount '%s' is invalid", args[2])
			}

			data, err := PackContractCallData(
				types.ERC20MintableBurnableContract.ABI,
				"convertToCoin",
				receiver,
				amount,
			)
			if err != nil {
				return err
			}

			ethTx, err := CreateEthCallContractTx(
				clientCtx,
				&contractAddr,
				data,
			)
			if err != nil {
				return err
			}

			return GenerateOrBroadcastTx(clientCtx, ethTx)
		},
	}
}
