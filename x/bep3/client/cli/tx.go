package cli

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"time"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/client/utils"
	"github.com/kava-labs/kava/x/bep3/types"
	"github.com/spf13/cobra"
	tmtime "github.com/tendermint/tendermint/types/time"
)

// GetTxCmd returns the transaction commands for this module
func GetTxCmd(cdc *codec.Codec) *cobra.Command {
	bep3TxCmd := &cobra.Command{
		Use:   "bep3",
		Short: "bep3 transactions subcommands",
	}

	bep3TxCmd.AddCommand(client.PostCommands(
		GetCmdCreateAtomicSwap(cdc),
		GetCmdClaimAtomicSwap(cdc),
		GetCmdRefundAtomicSwap(cdc),
	)...)

	return bep3TxCmd
}

// GetCmdCreateAtomicSwap cli command for creating atomic swaps
func GetCmdCreateAtomicSwap(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "create [to] [recipient-other-chain] [sender-other-chain] [timestamp] [coins] [expected-income] [height-span] [cross-chain]",
		Short: "create a new atomic swap",
		Example: fmt.Sprintf("%s tx %s create kava1xy7hrjy9r0algz9w3gzm8u6mrpq97kwta747gj bnb1urfermcg92dwq36572cx4xg84wpk3lfpksr5g7 bnb1uky3me9ggqypmrsvxk7ur6hqkzq7zmv4ed4ng7 now 100bnb 100bnb 360 true --from validator",
			version.ClientName, types.ModuleName),
		Args: cobra.ExactArgs(8),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			txBldr := auth.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))

			from := cliCtx.GetFromAddress() // same as KavaExecutor.DeputyAddress (for cross-chain)
			to, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			recipientOtherChain := args[1] // same as OtherExecutor.DeputyAddress
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

			randomNumber, err := types.GenerateSecureRandomNumber()
			if err != nil {
				return err
			}

			randomNumberHash := types.CalculateRandomHash(randomNumber.Bytes(), timestamp)

			// Print random number, timestamp, and hash to user's console
			fmt.Printf("\nRandom number: %s\n", randomNumber.Text(16))
			fmt.Printf("Timestamp: %d\n", timestamp)
			fmt.Printf("Random number hash: %s\n\n", hex.EncodeToString(randomNumberHash))

			coins, err := sdk.ParseCoins(args[4])
			if err != nil {
				return err
			}

			expectedIncome := args[5]

			heightSpan, err := strconv.ParseInt(args[6], 10, 64)
			if err != nil {
				return err
			}

			crossChain, err := strconv.ParseBool(args[7])
			if err != nil {
				return err
			}

			msg := types.NewMsgCreateAtomicSwap(
				from, to, recipientOtherChain, senderOtherChain, randomNumberHash,
				timestamp, coins, expectedIncome, heightSpan, crossChain,
			)

			err = msg.ValidateBasic()
			if err != nil {
				return err
			}

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}

// GetCmdClaimAtomicSwap cli command for claiming an atomic swap
func GetCmdClaimAtomicSwap(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:     "claim [swap-id] [random-number]",
		Short:   "claim coins in an atomic swap using the secret number",
		Example: fmt.Sprintf("%s tx %s claim 6682c03cc3856879c8fb98c9733c6b0c30758299138166b6523fe94628b1d3af 56f13e6a5cd397447f8b5f8c82fdb5bbf56127db75269f5cc14e50acd8ac9a4c --from accA", version.ClientName, types.ModuleName),
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			txBldr := auth.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))

			from := cliCtx.GetFromAddress()

			swapID, err := types.HexToBytes(args[0])
			if err != nil {
				return err
			}

			if len(strings.TrimSpace(args[1])) == 0 {
				return fmt.Errorf("random-number cannot be empty")
			}

			randomNumber, err := types.HexToBytes(args[1])
			if err != nil {
				return err
			}

			msg := types.NewMsgClaimAtomicSwap(from, swapID, randomNumber)

			err = msg.ValidateBasic()
			if err != nil {
				return err
			}

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}

// GetCmdRefundAtomicSwap cli command for claiming an atomic swap
func GetCmdRefundAtomicSwap(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:     "refund [swap-id]",
		Short:   "refund the coins in an atomic swap",
		Example: fmt.Sprintf("%s tx %s refund 6682c03cc3856879c8fb98c9733c6b0c30758299138166b6523fe94628b1d3af --from accA", version.ClientName, types.ModuleName),
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			txBldr := auth.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))

			from := cliCtx.GetFromAddress()

			swapID, err := types.HexToBytes(args[0])
			if err != nil {
				return err
			}

			msg := types.NewMsgRefundAtomicSwap(from, swapID)

			err = msg.ValidateBasic()
			if err != nil {
				return err
			}

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}

func generateSecureRandomNumber() (*big.Int, error) {
	max := new(big.Int)
	max.Exp(big.NewInt(2), big.NewInt(256), nil).Sub(max, big.NewInt(1)) // 256-bits integer i.e. 2^256 - 1

	// Generate number between 0 - max
	randomNumber, err := rand.Int(rand.Reader, max)
	if err != nil {
		return big.NewInt(0), errors.New("random number generation error")
	}

	// Catch random numbers that encode to hexadecimal poorly
	if len(randomNumber.Text(16)) != 64 {
		return generateSecureRandomNumber()
	}

	return randomNumber, nil
}
