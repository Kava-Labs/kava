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
func GetQueryCmd(storeKey string, cdc *codec.Codec) *cobra.Command {
	// Group nameservice queries under a subcommand
	pricefeedQueryCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Querying commands for the pricefeed module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	pricefeedQueryCmd.AddCommand(client.GetCommands(
		GetCmdCurrentPrice(storeKey, cdc),
		GetCmdRawPrices(storeKey, cdc),
		GetCmdMarkets(storeKey, cdc),
	)...)

	return pricefeedQueryCmd
}

// GetCmdCurrentPrice queries the current price of an asset
func GetCmdCurrentPrice(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "price [assetCode]",
		Short: "get the current price of an asset",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			assetCode := args[0]
			res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/price/%s", queryRoute, assetCode), nil)
			if err != nil {
				fmt.Printf("error when querying current price - %s", err)
				fmt.Printf("could not get current price for - %s \n", assetCode)
				return nil
			}
			var out types.CurrentPrice
			cdc.MustUnmarshalJSON(res, &out)
			return cliCtx.PrintOutput(out)
		},
	}
}

// GetCmdRawPrices queries the current price of an asset
func GetCmdRawPrices(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "rawprices [assetCode]",
		Short: "get the raw oracle prices for an asset",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			assetCode := args[0]
			res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/rawprices/%s", queryRoute, assetCode), nil)
			if err != nil {
				fmt.Printf("could not get raw prices for - %s \n", assetCode)
				return nil
			}
			var out []types.PostedPrice
			cdc.MustUnmarshalJSON(res, &out)
			return cliCtx.PrintOutput(out)
		},
	}
}

// GetCmdMarkets queries list of assets in the pricefeed
func GetCmdMarkets(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "markets",
		Short: "get the markets in the pricefeed",
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/markets", queryRoute), nil)
			if err != nil {
				fmt.Printf("could not get assets")
				return nil
			}
			var out types.Markets
			cdc.MustUnmarshalJSON(res, &out)
			return cliCtx.PrintOutput(out)
		},
	}
}
