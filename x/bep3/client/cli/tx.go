package cli

import (
	"github.com/spf13/cobra"

	binance "github.com/binance-chain/go-sdk/common/types"
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
	)...)

	return bep3TxCmd
}

// GetCmdCreateHtlt cli command for creating htlts
func GetCmdCreateHtlt(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "create [from] [to] [recipientOtherChain] [senderOtherChain] [randNumHash] [timestamp] [amount] [expectedIncome] [heightSpan] [crosschain]",
		Short: "create a new Hashed Time Locked Transaction (HTLT)",
		Args:  cobra.MinimumNArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {

			cliCtx := context.NewCLIContext().WithCodec(cdc)
			txBldr := auth.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))

			// We need...
			// 1. ReceiverAddr == KavaExecutor.DeputyAddress
			//		- kava15qdefkmwswysgg4qxgqpqr35k3m49pkx2jdfnw
			// 2. OtherChainAddr == OtherExecutor.DeputyAddress
			//		- 0x9eB05a790e2De0a047a57a22199D8CccEA6d6D5A
			// 3. txLog.OutCoin == deputy.Config.KavaConfig.Symbol (both are converted to uppercase)
			//		- kava

			// Acc A -> passed with '--from accA'
			from := cliCtx.GetFromAddress()
			// from := sdk.AccAddress("kava1xy7hrjy9r0algz9w3gzm8u6mrpq97kwta747gj")

			// Acc B
			to := sdk.AccAddress("kava15qdefkmwswysgg4qxgqpqr35k3m49pkx2jdfnw")
			recipientOtherChain := "0x9eB05a790e2De0a047a57a22199D8CccEA6d6D5A" //current depty eth address
			senderOtherChain := ""

			randomNumberBytes := []byte{15}
			timestampInt64 := int64(9988776655)
			randomNumberHash := types.CalculateRandomHash(randomNumberBytes, timestampInt64)

			amount := binance.Coins{binance.Coin{Denom: "kava", Amount: 100}}
			expectedIncome := "kava100"
			heightSpan := int64(500000)
			crossChain := true

			msg := types.NewMsgCreateHTLT(from, to, recipientOtherChain, senderOtherChain,
				randomNumberHash, timestampInt64, amount, expectedIncome, heightSpan, crossChain)

			err := msg.ValidateBasic()
			if err != nil {
				return err
			}

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}
