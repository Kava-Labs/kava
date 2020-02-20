package cli

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/client/utils"
	"github.com/kava-labs/kava/x/bep3/types"
)

// GetTxCmd returns the transaction commands for this module
func GetTxCmd(cdc *codec.Codec) *cobra.Command {
	bep3TxCmd := &cobra.Command{
		Use:   "bep3",
		Short: "bep3 transactions subcommands",
	}

	bep3TxCmd.AddCommand(client.PostCommands(
		GetCmdCreateHtlt(cdc),
		GetCmdDepositHtlt(cdc),
		GetCmdClaimHtlt(cdc),
		GetCmdRefundHtlt(cdc),
	)...)

	return bep3TxCmd
}

// GetCmdCreateHtlt cli command for creating htlts
func GetCmdCreateHtlt(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:     "create [to] [recipient-other-chain] [sender-other-chain] [hashed-secret] [timestamp] [coins] [expected-income] [height-span] [cross-chain]",
		Short:   "create a new Hashed Time Locked Transaction (HTLT)",
		Example: "bep3 create kava15qdefkmwswysgg4qxgqpqr35k3m49pkx2jdfnw 0x9eB05a790e2De0a047a57a22199D8CccEA6d6D5A '' 0677bd8a303dd981810f34d8e5cc6507f13b391899b84d3c1be6c6045a17d747 9988776655 100xrp 99xrp 1000 true --from accA",
		Args:    cobra.ExactArgs(9),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			txBldr := auth.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))

			from := cliCtx.GetFromAddress() // same as KavaExecutor.DeputyAddress (for cross-chain HTLTs)
			to, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			recipientOtherChain := args[1] // same as OtherExecutor.DeputyAddress
			senderOtherChain := args[2]

			randomNumberHash, err := types.HexToBytes(args[3])
			if err != nil {
				return err
			}

			timeStamp, err := strconv.ParseInt(args[4], 10, 64)
			if err != nil {
				return err
			}

			coins, err := sdk.ParseCoins(args[5])
			if err != nil {
				return err
			}

			expectedIncome := args[6]

			heightSpan, err := strconv.ParseInt(args[7], 10, 64)
			if err != nil {
				return err
			}

			crossChain, err := strconv.ParseBool(args[8])
			if err != nil {
				return err
			}

			msg := types.NewMsgCreateHTLT(
				from, to, recipientOtherChain, senderOtherChain, randomNumberHash,
				timeStamp, coins, expectedIncome, heightSpan, crossChain,
			)

			err = msg.ValidateBasic()
			if err != nil {
				return err
			}

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}

// GetCmdDepositHtlt cli command for depositing into an htlt
func GetCmdDepositHtlt(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:     "deposit [swap-id] [coins]",
		Short:   "deposit coins into an existing HTLT",
		Example: "bep3 deposit 6682c03cc3856879c8fb98c9733c6b0c30758299138166b6523fe94628b1d3af 10btc  --from accA",
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
				return fmt.Errorf("coins cannot be empty")
			}
			coins, err := sdk.ParseCoins(args[1])
			if err != nil {
				return err
			}

			msg := types.NewMsgDepositHTLT(from, swapID, coins)

			err = msg.ValidateBasic()
			if err != nil {
				return err
			}

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}

// GetCmdClaimHtlt cli command for claiming an htlt
func GetCmdClaimHtlt(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:     "claim [swap-id] [random-number]",
		Short:   "claim coins in an HTLT using the secret random number",
		Example: "bep3 claim 6682c03cc3856879c8fb98c9733c6b0c30758299138166b6523fe94628b1d3af 123456789 --from accA",
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
			randomNumber := []byte(args[1])

			msg := types.NewMsgClaimHTLT(from, swapID, randomNumber)

			err = msg.ValidateBasic()
			if err != nil {
				return err
			}

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}

// GetCmdRefundHtlt cli command for claiming an htlt
func GetCmdRefundHtlt(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:     "refund [swap-id]",
		Short:   "refund the coins in an HTLT",
		Example: "bep3 refund 6682c03cc3856879c8fb98c9733c6b0c30758299138166b6523fe94628b1d3af --from accA",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			txBldr := auth.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))

			from := cliCtx.GetFromAddress()

			swapID, err := types.HexToBytes(args[0])
			if err != nil {
				return err
			}

			msg := types.NewMsgRefundHTLT(from, swapID)

			err = msg.ValidateBasic()
			if err != nil {
				return err
			}

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}
