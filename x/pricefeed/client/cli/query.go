package cli

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/spf13/cobra"

	"github.com/kava-labs/kava/x/pricefeed/types"
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

	pricefeedQueryCmd.AddCommand(flags.GetCommands(
		GetCmdPrice(queryRoute, cdc),
		GetCmdRawPrices(queryRoute, cdc),
		GetCmdOracles(queryRoute, cdc),
		GetCmdMarkets(queryRoute, cdc),
		GetCmdQueryParams(queryRoute, cdc),
	)...)

	return pricefeedQueryCmd
}

// GetCmdOracles queries the oracle set of an asset
func GetCmdOracles(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "oracles [marketID]",
		Short: "get the oracle set for a market",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			marketID := args[0]

			bz, err := cdc.MarshalJSON(types.QueryWithMarketIDParams{
				MarketID: marketID,
			})
			if err != nil {
				return err
			}
			route := fmt.Sprintf("custom/%s/%s", queryRoute, types.QueryOracles)

			res, _, err := cliCtx.QueryWithData(route, bz)
			if err != nil {
				return err
			}
			var oracles []sdk.AccAddress
			cdc.MustUnmarshalJSON(res, &oracles)
			return cliCtx.PrintOutput(oracles)
		},
	}
}

// GetCmdPrice queries the current price of an asset
func GetCmdPrice(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "price [marketID]",
		Short: "get the current price for the input market",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			marketID := args[0]

			bz, err := cdc.MarshalJSON(types.QueryWithMarketIDParams{
				MarketID: marketID,
			})
			if err != nil {
				return err
			}
			route := fmt.Sprintf("custom/%s/%s", queryRoute, types.QueryPrice)

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

			bz, err := cdc.MarshalJSON(types.QueryWithMarketIDParams{
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

// GetCmdQueryParams queries the pricefeed module parameters
func GetCmdQueryParams(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "params",
		Short: "get the pricefeed module parameters",
		Long:  "Get the current global pricefeed module parameters.",
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
