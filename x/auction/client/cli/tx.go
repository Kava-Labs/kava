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
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/client/utils"

	"github.com/kava-labs/kava/x/auction/types"
)

// GetTxCmd returns the transaction commands for this module
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

// GetCmdPlaceBid cli command for placing bids on auctions
func GetCmdPlaceBid(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "bid [auction-id] [amount]",
		Short: "place a bid on an auction",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Place a bid on any type of auction. Collateral auctions must be bid up to their maxbid before entering reverse phase.

Example:
$ %s tx %s bid 34 1000usdx --from myKeyName
`, version.ClientName, types.ModuleName)),
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			txBldr := auth.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))

			id, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("auction-id '%s' not a valid uint", args[0])
			}

			amt, err := sdk.ParseCoin(args[1])
			if err != nil {
				return err
			}

			msg := types.NewMsgPlaceBid(id, cliCtx.GetFromAddress(), amt)
			err = msg.ValidateBasic()
			if err != nil {
				return err
			}
			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}
