package cli

import (
	"context"
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/spf13/cobra"

	"github.com/kava-labs/kava/x/evmutil/types"
)

// GetQueryCmd returns the cli query commands for this module
func GetQueryCmd() *cobra.Command {
	evmutilQueryCmd := &cobra.Command{
		Use:                        "evmutil",
		Short:                      "Querying commands for the evmutil module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmds := []*cobra.Command{
		QueryParamsCmd(),
		QueryDeployedCosmosCoinContractsCmd(),
	}

	for _, cmd := range cmds {
		flags.AddQueryFlagsToCmd(cmd)
	}

	evmutilQueryCmd.AddCommand(cmds...)

	return evmutilQueryCmd
}

// QueryParamsCmd queries the evmutil module parameters
func QueryParamsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "params",
		Short: "Query the evmutil module parameters",
		Example: fmt.Sprintf(
			"%[1]s q %[2]s params",
			version.AppName, types.ModuleName,
		),
		Args: cobra.NoArgs,
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

func QueryDeployedCosmosCoinContractsCmd() *cobra.Command {
	var cosmosDenoms []string
	cmdName := "deployed-cosmos-coin-contracts"
	q := fmt.Sprintf("%[1]s q %[2]s %s", version.AppName, types.ModuleName, cmdName)
	cmd := &cobra.Command{
		Use:     fmt.Sprintf("%s [--denoms denom1,denom2] [flags]", cmdName),
		Short:   "Query for deployed ERC20 contract addresses representing cosmos coins in the EVM",
		Example: fmt.Sprintf("Query all:\n  %s\n\nQuery by denom:\n  %s --denoms denom1,denom2", q, q),
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			page, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)
			request := types.QueryDeployedCosmosCoinContractsRequest{
				CosmosDenoms: cosmosDenoms,
				Pagination:   page,
			}
			res, err := queryClient.DeployedCosmosCoinContracts(context.Background(), &request)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddPaginationFlagsToCmd(cmd, cmdName)
	cmd.Flags().StringSliceVar(&cosmosDenoms, "denoms", []string{}, fmt.Sprintf("(optional) Cosmos denoms to get addresses for. If no contract is deployed, the result will be omitted. Limit %d per query.", query.DefaultLimit))

	return cmd
}
