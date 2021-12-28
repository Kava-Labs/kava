package cli

import (
	"context"
	"strings"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"

	"github.com/kava-labs/kava/x/swap/types"
)

// flags for cli queries
const (
	flagOwner = "owner"
	flagPool  = "pool"
)

// GetQueryCmd returns the cli query commands for the  module
func GetQueryCmd(queryRoute string) *cobra.Command {
	swapQueryCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Querying commands for the swap module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmds := []*cobra.Command{
		queryParamsCmd(queryRoute),
		queryDepositsCmd(queryRoute),
		queryPoolsCmd(queryRoute),
	}

	for _, cmd := range cmds {
		flags.AddQueryFlagsToCmd(cmd)
	}

	swapQueryCmd.AddCommand(cmds...)

	return swapQueryCmd
}

func queryParamsCmd(queryRoute string) *cobra.Command {
	return &cobra.Command{
		Use:   "params",
		Short: "get the swap module parameters",
		Long:  "Get the current global swap module parameters.",
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

func queryDepositsCmd(queryRoute string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deposits",
		Short: "get liquidity provider deposits",
		Long: strings.TrimSpace(`get liquidity provider deposits:
 		Example:
 		$ kvcli q swap deposits --pool bnb:usdx
 		$ kvcli q swap deposits --owner kava1l0xsq2z7gqd7yly0g40y5836g0appumark77ny
 		$ kvcli q swap deposits --pool bnb:usdx --owner kava1l0xsq2z7gqd7yly0g40y5836g0appumark77ny
 		$ kvcli q swap deposits --page=2 --limit=100
 		`,
		),
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			bechOwnerAddr, err := cmd.Flags().GetString(flagOwner)
			if err != nil {
				return err
			}
			pool, err := cmd.Flags().GetString(flagPool)
			if err != nil {
				return err
			}

			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			params := types.QueryDepositsRequest{
				Owner:      bechOwnerAddr,
				PoolId:     pool,
				Pagination: pageReq,
			}
			res, err := queryClient.Deposits(context.Background(), &params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddPaginationFlagsToCmd(cmd, "deposits")

	cmd.Flags().String(flagPool, "", "pool name")
	cmd.Flags().String(flagOwner, "", "owner, also known as a liquidity provider")

	return cmd
}

func queryPoolsCmd(queryRoute string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pools",
		Short: "get statistics for all pools",
		Long: strings.TrimSpace(`get statistics for all liquidity pools:
 		Example:
 		$ kvcli q swap pools`,
		),
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			params := types.QueryPoolsRequest{}
			res, err := queryClient.Pools(context.Background(), &params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}
	return cmd
}
