package cli

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/kava-labs/kava/x/pricefeed/types"
	"github.com/spf13/cobra"
)

// GetQueryCmd returns the cli query commands for this module
func GetQueryCmd(queryRoute string, cdc *codec.Codec) *cobra.Command {
	// Group nameservice queries under a subcommand
	pricefeedQueryCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Querying commands for the pricefeed module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	pricefeedQueryCmd.AddCommand(client.GetCommands(
		GetCmdCurrentPrice(queryRoute, cdc),
		GetCmdRawPrices(queryRoute, cdc),
		GetCmdMarkets(queryRoute, cdc),
	)...)

	return pricefeedQueryCmd
}

// GetCmdCurrentPrice queries the current price of an asset
func GetCmdCurrentPrice(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "price [marketID]",
		Short: "get the current price for the input market",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			marketID := args[0]

			bz, err := cdc.MarshalJSON(types.QueryPricesParams{
				MarketID: marketID,
			})
			if err != nil {
				return err
			}
			route := fmt.Sprintf("custom/%s/%s", queryRoute, types.QueryCurrentPrice)

			res, _, err := cliCtx.QueryWithData(route, bz)
			if err != nil {
				return err
			}
			var price types.CurrentPrice
			cdc.MustUnmarshalJSON(res, &price)
			return cliCtx.PrintOutput(price)
		},
	}
}

// GetCmdRawPrices queries the current price of an asset
func GetCmdRawPrices(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "rawprices [marketID]",
		Short: "get the raw oracle prices for the input market",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			marketID := args[0]

			bz, err := cdc.MarshalJSON(types.QueryPricesParams{
				MarketID: marketID,
			})
			if err != nil {
				return err
			}
			route := fmt.Sprintf("custom/%s/%s", queryRoute, types.QueryRawPrices)

			res, _, err := cliCtx.QueryWithData(route, bz)

			if err != nil {
				return err
			}
			var prices []types.PostedPrice
			cdc.MustUnmarshalJSON(res, &prices)
			return cliCtx.PrintOutput(prices)
		},
	}
}

// GetCmdMarkets queries list of markets in the pricefeed
func GetCmdMarkets(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "markets",
		Short: "get the markets in the pricefeed",
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			route := fmt.Sprintf("custom/%s/markets", queryRoute)
			res, _, err := cliCtx.QueryWithData(route, nil)
			if err != nil {
				return err
			}
			var markets types.Markets
			cdc.MustUnmarshalJSON(res, &markets)
			return cliCtx.PrintOutput(markets)
		},
	}
}
