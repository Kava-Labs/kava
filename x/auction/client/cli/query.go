package cli

import (
	"fmt"
	"strconv"

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
		QueryGetAuctionCmd(queryRoute, cdc),
		QueryGetAuctionsCmd(queryRoute, cdc),
		QueryParamsCmd(queryRoute, cdc),
	)...)

	return auctionQueryCmd
}

// QueryGetAuctionCmd queries one auction in the store
func QueryGetAuctionCmd(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "auction [auction-id]",
		Short: "get a info about an auction",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			// Prepare params for querier
			id, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("auction-id '%s' not a valid uint", args[0])
			}
			bz, err := cdc.MarshalJSON(types.QueryAuctionParams{
				AuctionID: id,
			})
			if err != nil {
				return err
			}

			// Query
			res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s", queryRoute, types.QueryGetAuction), bz)
			if err != nil {
				return err
			}

			// Decode and print results
			var auction types.Auction
			cdc.MustUnmarshalJSON(res, &auction)
			auctionWithPhase := types.NewAuctionWithPhase(auction)
			return cliCtx.PrintOutput(auctionWithPhase)
		},
	}
}

// QueryGetAuctionsCmd queries the auctions in the store
func QueryGetAuctionsCmd(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "auctions",
		Short: "get a list of active auctions",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			// Query
			res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s", queryRoute, types.QueryGetAuctions), nil)
			if err != nil {
				return err
			}

			// Decode and print results
			var auctions types.Auctions
			cdc.MustUnmarshalJSON(res, &auctions)

			auctionsWithPhase := []types.AuctionWithPhase{} // using empty slice so json returns [] instead of null when there's no auctions
			for _, a := range auctions {
				auctionsWithPhase = append(auctionsWithPhase, types.NewAuctionWithPhase(a))
			}
			return cliCtx.PrintOutput(auctionsWithPhase)
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
