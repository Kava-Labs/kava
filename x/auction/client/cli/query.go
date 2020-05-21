package cli

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/x/auction/types"
)

// Query auction flags
const (
	flagType  = "type"
	flagDenom = "denom"
	flagPhase = "phase"
)

// GetQueryCmd returns the cli query commands for this module
func GetQueryCmd(queryRoute string, cdc *codec.Codec) *cobra.Command {
	// Group nameservice queries under a subcommand
	auctionQueryCmd := &cobra.Command{
		Use:   "auction",
		Short: "Querying commands for the auction module",
	}

	auctionQueryCmd.AddCommand(flags.GetCommands(
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
	cmd := &cobra.Command{
		Use:   "auctions",
		Short: "query auctions with optional filters",
		Long: strings.TrimSpace(`Query for all paginated auctions that match optional filters:
Example:
$ kvcli q auction auctions --type=(Collateral|Surplus|Debt)
$ kvcli q auction auctions --denom=bnb
$ kvcli q auction auctions --phase=(Forward|Reverse)
$ kvcli q auction auctions --page=2 --limit=100
`,
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			strType := viper.GetString(flagType)
			strDenom := viper.GetString(flagDenom)
			strPhase := viper.GetString(flagPhase)
			page := viper.GetInt(flags.FlagPage)
			limit := viper.GetInt(flags.FlagLimit)

			var auctionType string
			var auctionDenom string
			var auctionPhase string

			params := types.NewQueryAllAuctionParams(page, limit, auctionType, auctionDenom, auctionPhase)

			if len(strType) != 0 {
				auctionType = strings.ToLower(strings.TrimSpace(strType))
				if auctionType != "collateral" && auctionType != "surplus" && auctionType != "debt" {
					return fmt.Errorf("invalid auction type %s", strType)
				}
				params.Type = auctionType
			}

			if len(strDenom) != 0 {
				auctionDenom := strings.TrimSpace(strDenom)
				err := sdk.ValidateDenom(auctionDenom)
				if err != nil {
					return err
				}
				params.Denom = auctionDenom
			}

			if len(strPhase) != 0 {
				auctionPhase := strings.ToLower(strings.TrimSpace(strPhase))
				if auctionType != "collateral" && auctionType != "" {
					return fmt.Errorf("cannot apply phase flag to non-collateral auction type")
				}
				if auctionPhase != "forward" && auctionPhase != "reverse" {
					return fmt.Errorf("invalid auction phase %s", strPhase)
				}
				params.Phase = auctionPhase
			}

			bz, err := cdc.MarshalJSON(params)
			if err != nil {
				return err
			}

			cliCtx := context.NewCLIContext().WithCodec(cdc)

			// Query
			res, height, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s", queryRoute, types.QueryGetAuctions), bz)
			if err != nil {
				return err
			}

			// Decode and print results
			var matchingAuctions types.Auctions
			cdc.MustUnmarshalJSON(res, &matchingAuctions)

			if len(matchingAuctions) == 0 {
				return fmt.Errorf("No matching auctions found")
			}

			auctionsWithPhase := []types.AuctionWithPhase{} // using empty slice so json returns [] instead of null when there's no auctions
			for _, a := range matchingAuctions {
				auctionsWithPhase = append(auctionsWithPhase, types.NewAuctionWithPhase(a))
			}
			cliCtx = cliCtx.WithHeight(height)
			return cliCtx.PrintOutput(auctionsWithPhase) // nolint:errcheck
		},
	}

	cmd.Flags().Int(flags.FlagPage, 1, "pagination page of auctions to to query for")
	cmd.Flags().Int(flags.FlagLimit, 100, "pagination limit of auctions to query for")
	cmd.Flags().String(flagType, "", "(optional) filter by auction type, type: collateral, debt, surplus")
	cmd.Flags().String(flagDenom, "", "(optional) filter by auction denom")
	cmd.Flags().String(flagPhase, "", "(optional) filter by collateral auction phase, phase: forward/reverse")

	return cmd
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
