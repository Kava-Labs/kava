package cli

import (
	"fmt"
	"strconv"
	"time"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"

	tmtime "github.com/tendermint/tendermint/types/time"

	"github.com/kava-labs/kava/x/pricefeed/types"
)

// GetTxCmd returns the transaction commands for this module
func GetTxCmd() *cobra.Command {
	pricefeedTxCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Pricefeed transactions subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmds := []*cobra.Command{
		GetCmdPostPrice(),
	}

	for _, cmd := range cmds {
		flags.AddTxFlagsToCmd(cmd)
	}

	pricefeedTxCmd.AddCommand(cmds...)

	return pricefeedTxCmd
}

// GetCmdPostPrice cli command for posting prices.
func GetCmdPostPrice() *cobra.Command {
	return &cobra.Command{
		Use:   "postprice [marketID] [price] [expiry]",
		Short: "post the latest price for a particular market with a given expiry as a UNIX time",
		Example: fmt.Sprintf("%s tx %s postprice bnb:usd 25 9999999999 --from validator",
			version.AppName, types.ModuleName),
		Args: cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			price, err := sdk.NewDecFromStr(args[1])
			if err != nil {
				return err
			}

			expiryInt, err := strconv.ParseInt(args[2], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid expiry %s: %w", args[2], err)
			}

			if expiryInt > types.MaxExpiry {
				return fmt.Errorf("invalid expiry; got %d, max: %d", expiryInt, types.MaxExpiry)
			}

			expiry := tmtime.Canonical(time.Unix(expiryInt, 0))

			from := clientCtx.GetFromAddress()
			msg := types.NewMsgPostPrice(string(from), args[0], price, expiry)
			if err = msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
}
