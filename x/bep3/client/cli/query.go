package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/kava-labs/kava/x/bep3/types"
)

// GetQueryCmd returns the cli query commands for this module
func GetQueryCmd(queryRoute string, cdc *codec.Codec) *cobra.Command {
	// Group nameservice queries under a subcommand
	bep3QueryCmd := &cobra.Command{
		Use:   "bep3",
		Short: "Querying commands for the bep3 module",
	}

	bep3QueryCmd.AddCommand(client.GetCommands(
		QueryGetHtltsCmd(queryRoute, cdc),
		QueryParamsCmd(queryRoute, cdc),
	)...)

	return bep3QueryCmd
}

// QueryGetHtltsCmd queries the htlts in the store
func QueryGetHtltsCmd(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "htlts",
		Short: "get a list of active htlts",
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/htlts", queryRoute), nil)
			if err != nil {
				fmt.Printf("error when getting htlts - %s", err)
				return nil
			}
			var out types.QueryResHTLTs
			cdc.MustUnmarshalJSON(res, &out)
			if len(out) == 0 {
				out = append(out, "There are currently no htlts")
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
