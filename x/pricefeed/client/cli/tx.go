package cli

import (
	"bufio"
	"fmt"
	"strconv"
	"time"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/client/utils"

	tmtime "github.com/tendermint/tendermint/types/time"

	"github.com/kava-labs/kava/x/pricefeed/types"
)

const maxExpiry = 253402300799 // 9999-12-31 23:59:59 +0000 UTC

// GetTxCmd returns the transaction commands for this module
func GetTxCmd(storeKey string, cdc *codec.Codec) *cobra.Command {
	pricefeedTxCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Pricefeed transactions subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	pricefeedTxCmd.AddCommand(flags.PostCommands(
		GetCmdPostPrice(cdc),
	)...)

	return pricefeedTxCmd
}

// GetCmdPostPrice cli command for posting prices.
func GetCmdPostPrice(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "postprice [marketID] [price] [expiry]",
		Short: "post the latest price for a particular market with a given expiry as a UNIX time",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			inBuf := bufio.NewReader(cmd.InOrStdin())
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			txBldr := auth.NewTxBuilderFromCLI(inBuf).WithTxEncoder(utils.GetTxEncoder(cdc))

			price, err := sdk.NewDecFromStr(args[1])
			if err != nil {
				return err
			}

			expiryInt, err := strconv.ParseInt(args[2], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid expiry %s: %w", args[2], err)
			}

			if expiryInt > maxExpiry {
				return fmt.Errorf("invalid expiry; got %d, max: %d", expiryInt, maxExpiry)
			}

			expiry := tmtime.Canonical(time.Unix(expiryInt, 0))

			msg := types.NewMsgPostPrice(cliCtx.GetFromAddress(), args[0], price, expiry)
			if err = msg.ValidateBasic(); err != nil {
				return err
			}

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}
