package cli

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"

	"github.com/kava-labs/kava/x/liquid/types"
)

// flags for cli queries
const (
	flagDenom = "denom"
	flagOwner = "owner"
)

// GetQueryCmd returns the cli query commands for this module
func GetQueryCmd() *cobra.Command {
	liquidQueryCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Querying commands for the liquid module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmds := []*cobra.Command{
		GetCmdQueryParams(),
	}

	for _, cmd := range cmds {
		flags.AddQueryFlagsToCmd(cmd)
	}

	liquidQueryCmd.AddCommand(cmds...)

	return liquidQueryCmd
}

// GetCmdQueryParams queries the liquid module parameters
func GetCmdQueryParams() *cobra.Command {
	return &cobra.Command{
		Use:   "params",
		Short: "get the liquid module parameters",
		Long:  "Get the current global liquid module parameters.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.Params(context.Background(), &types.QueryParamsRequest{})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(&res.Params)
		},
	}
}
