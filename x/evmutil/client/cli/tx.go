package cli

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"

	"github.com/spf13/cobra"

	sdkmath "cosmossdk.io/math"
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
		getCmdMsgConvertCosmosCoinToERC20(),
	}

	for _, cmd := range cmds {
		flags.AddTxFlagsToCmd(cmd)
	}

	txCmd.AddCommand(cmds...)

	return txCmd
}

func getCmdMsgConvertCoinToERC20() *cobra.Command {
	return &cobra.Command{
		Use:   "convert-coin-to-erc20 [Kava EVM address] [coin]",
		Short: "converts sdk.Coin to erc20 tokens on Kava eth co-chain",
		Example: fmt.Sprintf(
			`%s tx %s convert-coin-to-erc20 0x7Bbf300890857b8c241b219C6a489431669b3aFA 500000000erc20/usdc --from <key> --gas 2000000`,
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
		Use:   "convert-erc20-to-coin [Kava receiver address] [Kava ERC20 address] [amount]",
		Short: "burns ERC20 tokens on Kava EVM co-chain and unlocks on Ethereum",
		Example: fmt.Sprintf(`
%[1]s tx %[2]s convert-erc20-to-coin kava10wlnqzyss4accfqmyxwx5jy5x9nfkwh6qm7n4t 0xeA7100edA2f805356291B0E55DaD448599a72C6d 1000000000000000 --from <key> --gas 1000000
`, version.AppName, types.ModuleName,
		),
		Args: cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			receiver, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return fmt.Errorf("receiver '%s' is not a bech32 address", args[0])
			}

			signer := clientCtx.GetFromAddress()
			initiator, err := ParseAddrFromHexOrBech32(signer.String())
			if err != nil {
				return err
			}

			amount, ok := sdkmath.NewIntFromString(args[2])
			if !ok {
				return fmt.Errorf("amount '%s' is invalid", args[2])
			}

			if !common.IsHexAddress(args[1]) {
				return fmt.Errorf("contractAddr '%s' is not a hex address", args[1])
			}
			contractAddr := types.NewInternalEVMAddress(common.HexToAddress(args[1]))
			msg := types.NewMsgConvertERC20ToCoin(types.NewInternalEVMAddress(initiator), receiver, contractAddr, amount)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), &msg)
		},
	}
}

func getCmdMsgConvertCosmosCoinToERC20() *cobra.Command {
	return &cobra.Command{
		Use:   "convert-cosmos-coin-to-erc20 [receiver_0x_address] [amount] [flags]",
		Short: "converts asset native to Cosmos Co-chain to an ERC20 on the EVM Co-chain",
		Example: fmt.Sprintf(
			`Convert 500 ATOM and send ERC20 to 0x03db6b11F47d074a532b9eb8a98aB7AdA5845087:
  %s tx %s convert-cosmos-coin-to-erc20 0x03db6b11F47d074a532b9eb8a98aB7AdA5845087 500000000ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2 --from <key> --gas 2000000`,
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

			amount, err := sdk.ParseCoinNormalized(args[1])
			if err != nil {
				return err
			}

			signer := clientCtx.GetFromAddress()
			msg := types.NewMsgConvertCosmosCoinToERC20(signer.String(), receiver, amount)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), &msg)
		},
	}
}
