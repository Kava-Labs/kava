package cli

import (
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"

	tmtime "github.com/tendermint/tendermint/types/time"

	"github.com/kava-labs/kava/x/bep3/types"
)

// GetTxCmd returns the transaction commands for this module
func GetTxCmd() *cobra.Command {
	bep3TxCmd := &cobra.Command{
		Use:                        "bep3",
		Short:                      "bep3 transactions subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmds := []*cobra.Command{
		GetCmdCreateAtomicSwap(),
		GetCmdClaimAtomicSwap(),
		GetCmdRefundAtomicSwap(),
	}

	for _, cmd := range cmds {
		flags.AddTxFlagsToCmd(cmd)
	}

	bep3TxCmd.AddCommand(cmds...)

	return bep3TxCmd
}

// GetCmdCreateAtomicSwap cli command for creating atomic swaps
func GetCmdCreateAtomicSwap() *cobra.Command {
	return &cobra.Command{
		Use:   "create [to] [recipient-other-chain] [sender-other-chain] [timestamp] [coins] [height-span]",
		Short: "create a new atomic swap",
		Example: fmt.Sprintf("%s tx %s create kava1xy7hrjy9r0algz9w3gzm8u6mrpq97kwta747gj bnb1urfermcg92dwq36572cx4xg84wpk3lfpksr5g7 bnb1uky3me9ggqypmrsvxk7ur6hqkzq7zmv4ed4ng7 now 100bnb 270 --from validator",
			version.AppName, types.ModuleName),
		Args: cobra.ExactArgs(6),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			from := clientCtx.GetFromAddress() // same as Kava executor's deputy address
			to, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			recipientOtherChain := args[1] // same as the other executor's deputy address
			senderOtherChain := args[2]

			// Timestamp defaults to time.Now() unless it's explicitly set
			var timestamp int64
			if strings.Compare(args[3], "now") == 0 {
				timestamp = tmtime.Now().Unix()
			} else {
				timestamp, err = strconv.ParseInt(args[3], 10, 64)
				if err != nil {
					return err
				}
			}

			// Generate cryptographically strong pseudo-random number
			randomNumber, err := types.GenerateSecureRandomNumber()
			if err != nil {
				return err
			}

			randomNumberHash := types.CalculateRandomHash(randomNumber, timestamp)

			// Print random number, timestamp, and hash to user's console
			fmt.Printf("\nRandom number: %s\n", hex.EncodeToString(randomNumber))
			fmt.Printf("Timestamp: %d\n", timestamp)
			fmt.Printf("Random number hash: %s\n\n", hex.EncodeToString(randomNumberHash))

			coins, err := sdk.ParseCoinsNormalized(args[4])
			if err != nil {
				return err
			}

			heightSpan, err := strconv.ParseUint(args[5], 10, 64)
			if err != nil {
				return err
			}

			msg := types.NewMsgCreateAtomicSwap(
				from.String(), to.String(), recipientOtherChain, senderOtherChain,
				randomNumberHash, timestamp, coins, heightSpan,
			)

			err = msg.ValidateBasic()
			if err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), &msg)
		},
	}
}

// GetCmdClaimAtomicSwap cli command for claiming an atomic swap
func GetCmdClaimAtomicSwap() *cobra.Command {
	return &cobra.Command{
		Use:   "claim [swap-id] [random-number]",
		Short: "claim coins in an atomic swap using the secret number",
		Example: fmt.Sprintf(
			"%s tx %s claim 6682c03cc3856879c8fb98c9733c6b0c30758299138166b6523fe94628b1d3af 56f13e6a5cd397447f8b5f8c82fdb5bbf56127db75269f5cc14e50acd8ac9a4c --from accA",
			version.AppName, types.ModuleName,
		),
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			from := clientCtx.GetFromAddress()

			swapID, err := hex.DecodeString(args[0])
			if err != nil {
				return err
			}

			if len(strings.TrimSpace(args[1])) == 0 {
				return fmt.Errorf("random-number cannot be empty")
			}
			randomNumber, err := hex.DecodeString(args[1])
			if err != nil {
				return err
			}

			msg := types.NewMsgClaimAtomicSwap(from.String(), swapID, randomNumber)

			err = msg.ValidateBasic()
			if err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), &msg)
		},
	}
}

// GetCmdRefundAtomicSwap cli command for claiming an atomic swap
func GetCmdRefundAtomicSwap() *cobra.Command {
	return &cobra.Command{
		Use:   "refund [swap-id]",
		Short: "refund the coins in an atomic swap",
		Example: fmt.Sprintf(
			"%s tx %s refund 6682c03cc3856879c8fb98c9733c6b0c30758299138166b6523fe94628b1d3af --from accA",
			version.AppName, types.ModuleName,
		),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			from := clientCtx.GetFromAddress()

			swapID, err := hex.DecodeString(args[0])
			if err != nil {
				return err
			}

			msg := types.NewMsgRefundAtomicSwap(from.String(), swapID)

			err = msg.ValidateBasic()
			if err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), &msg)
		},
	}
}
