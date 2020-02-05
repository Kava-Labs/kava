package cli

import (
	"github.com/spf13/cobra"

	// "github.com/cosmos/cosmos-sdk/client"
	// "github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	// sdk "github.com/cosmos/cosmos-sdk/types"
	// "github.com/cosmos/cosmos-sdk/x/auth"
	// "github.com/cosmos/cosmos-sdk/x/auth/client/utils"
	// "github.com/kava-labs/kava/x/bep3/types"
)

// GetTxCmd returns the transaction commands for this module
func GetTxCmd(cdc *codec.Codec) *cobra.Command {
	bep3TxCmd := &cobra.Command{
		Use:   "bep3",
		Short: "bep3 transactions subcommands",
	}

	// bep3TxCmd.AddCommand(client.PostCommands(
	// 	GetCmdCreateHtlt(cdc),
	// )...)

	return bep3TxCmd
}

// // GetCmdCreateHtlt cli command for creating htlts
// func GetCmdCreateHtlt(cdc *codec.Codec) *cobra.Command {
// 	return &cobra.Command{
// 		Use:   "create [coin]",
// 		Short: "create a new Hashed Time Locked Transaction (HTLT)",
// 		Args:  cobra.MinimumNArgs(0),
// 		RunE: func(cmd *cobra.Command, args []string) error {
// 			cliCtx := context.NewCLIContext().WithCodec(cdc)
// 			txBldr := auth.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))

// 			// id, err := strconv.ParseUint(args[0], 10, 64)
// 			// if err != nil {
// 			// 	fmt.Printf("invalid auction id - %s \n", string(args[0]))
// 			// 	return err
// 			// }

// 			amt, err := bnb.ParseCoin(args[2])
// 			if err != nil {
// 				fmt.Printf("invalid amount - %s \n", string(args[2]))
// 				return err
// 			}

// 			msg := types.NewMsgCreateHtlt(cliCtx.GetFromAddress(), amt)
// 			err = msg.ValidateBasic()
// 			if err != nil {
// 				return err
// 			}
// 			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
// 		},
// 	}
// }
