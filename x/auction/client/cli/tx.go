package cli

import (
	"fmt"

	"github.com/kava-labs/kava/x/auction/types"
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/client/utils"
)

// GetTxCmd returns the transaction commands for this module
// TODO: Tests, see: https://github.com/cosmos/cosmos-sdk/blob/18de630d0ae1887113e266982b51c2bf1f662edb/x/staking/client/cli/tx_test.go
func GetTxCmd(cdc *codec.Codec) *cobra.Command {
	auctionTxCmd := &cobra.Command{
		Use:   "auction",
		Short: "auction transactions subcommands",
	}

	auctionTxCmd.AddCommand(client.PostCommands(
		GetCmdPlaceBid(cdc),
	)...)

	return auctionTxCmd
}

// GetCmdPlaceBid cli command for creating and modifying cdps.
func GetCmdPlaceBid(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "placebid [AuctionID] [Bidder] [Bid] [Lot]",
		Short: "place a bid on an auction",
		Args:  cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			txBldr := auth.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))

			id, err := types.NewIDFromString(args[0])
			if err != nil {
				fmt.Printf("invalid auction id - %s \n", string(args[0]))
				return err
			}

			bid, err := sdk.ParseCoin(args[2])
			if err != nil {
				fmt.Printf("invalid bid amount - %s \n", string(args[2]))
				return err
			}

			lot, err := sdk.ParseCoin(args[3])
			if err != nil {
				fmt.Printf("invalid lot - %s \n", string(args[3]))
				return err
			}
			msg := types.NewMsgPlaceBid(id, cliCtx.GetFromAddress(), bid, lot)
			err = msg.ValidateBasic()
			if err != nil {
				return err
			}
			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}
