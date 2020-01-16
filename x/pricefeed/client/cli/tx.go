package cli

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/client/utils"
	"github.com/kava-labs/kava/x/pricefeed/types"
	tmtime "github.com/tendermint/tendermint/types/time"
)

// GetTxCmd returns the transaction commands for this module
func GetTxCmd(storeKey string, cdc *codec.Codec) *cobra.Command {
	pricefeedTxCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Pricefeed transactions subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	pricefeedTxCmd.AddCommand(client.PostCommands(
		GetCmdPostPrice(cdc),
	)...)

	return pricefeedTxCmd
}

// GetCmdPostPrice cli command for posting prices.
func GetCmdPostPrice(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "postprice [marketID] [price] [expiry]",
		Short: "post the latest price for a particular market",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			txBldr := auth.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))
			// if err := cliCtx.EnsureAccountExists(); err != nil {
			// 	return err
			// }
			price, err := sdk.NewDecFromStr(args[1])
			if err != nil {
				return err
			}
			expiryInt, ok := sdk.NewIntFromString(args[2])
			if !ok {
				fmt.Printf("invalid expiry - %s \n", args[2])
				return nil
			}
			expiry := tmtime.Canonical(time.Unix(expiryInt.Int64(), 0))

			msg := types.NewMsgPostPrice(cliCtx.GetFromAddress(), args[0], price, expiry)
			err = msg.ValidateBasic()
			if err != nil {
				return err
			}
			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}
