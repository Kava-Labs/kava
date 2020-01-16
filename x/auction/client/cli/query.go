package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/kava-labs/kava/x/auction/types"
)

// GetQueryCmd returns the cli query commands for this module
func GetQueryCmd(queryRoute string, cdc *codec.Codec) *cobra.Command {
	// Group nameservice queries under a subcommand
	auctionQueryCmd := &cobra.Command{
		Use:   "auction",
		Short: "Querying commands for the auction module",
	}

	auctionQueryCmd.AddCommand(client.GetCommands(
		QueryGetAuctionsCmd(queryRoute, cdc),
		QueryParamsCmd(queryRoute, cdc),
	)...)

	return auctionQueryCmd
}

// QueryGetAuctionsCmd queries the auctions in the store
func QueryGetAuctionsCmd(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "getauctions",
		Short: "get a list of active auctions",
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/getauctions", queryRoute), nil)
			if err != nil {
				fmt.Printf("error when getting auctions - %s", err)
				return nil
			}
			var out types.QueryResAuctions
			cdc.MustUnmarshalJSON(res, &out)
			if len(out) == 0 {
				out = append(out, "There are currently no auctions")
			}
			return cliCtx.PrintOutput(out)
		},
	}
}

// QueryParamsCmd queries the auction module parameters
func QueryParamsCmd(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "params",
		Short: "get the auction module parameters",
		Long:  "Get the current global auction module parameters.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			// Query
			route := fmt.Sprintf("custom/%s/%s", queryRoute, types.QueryGetParams)
			res, _, err := cliCtx.QueryWithData(route, nil)
			if err != nil {
				return err
			}

			// Decode and print results
			var out types.Params
			cdc.MustUnmarshalJSON(res, &out)
			return cliCtx.PrintOutput(out)
		},
	}
}
