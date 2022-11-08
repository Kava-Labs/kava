package cli

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/version"

	"github.com/kava-labs/kava/x/earn/types"
)

// flags for cli queries
const (
	flagDenom               = "denom"
	flagOwner               = "owner"
	flagValueInStakedTokens = "value_in_staked_tokens"
)

// GetQueryCmd returns the cli query commands for the earn module
func GetQueryCmd() *cobra.Command {
	earnQueryCommand := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Querying commands for the earn module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmds := []*cobra.Command{
		queryParamsCmd(),
		queryVaultsCmd(),
		queryVaultCmd(),
		queryDepositsCmd(),
		queryTotalSupplyCmd(),
	}

	for _, cmd := range cmds {
		flags.AddQueryFlagsToCmd(cmd)
	}

	earnQueryCommand.AddCommand(cmds...)

	return earnQueryCommand
}

func queryParamsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "params",
		Short: "get the earn module parameters",
		Long:  "Get the current earn module parameters.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			req := types.NewQueryParamsRequest()
			res, err := queryClient.Params(context.Background(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(&res.Params)
		},
	}
}

func queryVaultsCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "vaults",
		Short:   "get all earn vaults",
		Long:    "Get all earn module vaults.",
		Args:    cobra.NoArgs,
		Example: fmt.Sprintf(`%[1]s q %[2]s vaults`, version.AppName, types.ModuleName),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			req := types.NewQueryVaultsRequest()
			res, err := queryClient.Vaults(context.Background(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}
}

func queryVaultCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "vault",
		Short:   "get a earn vault",
		Long:    "Get a specific earn module vault by denom.",
		Args:    cobra.ExactArgs(1),
		Example: fmt.Sprintf(`%[1]s q %[2]s vault usdx`, version.AppName, types.ModuleName),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			req := types.NewQueryVaultRequest(args[0])
			res, err := queryClient.Vault(context.Background(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}
}

func queryDepositsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deposits",
		Short: "get earn vault deposits",
		Long:  "Get earn vault deposits for all or specific accounts and vaults.",
		Args:  cobra.NoArgs,
		Example: fmt.Sprintf(`%[1]s q %[2]s deposits
%[1]s q %[2]s deposits --owner kava1l0xsq2z7gqd7yly0g40y5836g0appumark77ny --denom usdx
%[1]s q %[2]s deposits --denom usdx`, version.AppName, types.ModuleName),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			ownerBech, err := cmd.Flags().GetString(flagOwner)
			if err != nil {
				return err
			}
			denom, err := cmd.Flags().GetString(flagDenom)
			if err != nil {
				return err
			}
			valueInStakedTokens, err := cmd.Flags().GetBool(flagValueInStakedTokens)
			if err != nil {
				return err
			}

			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			req := types.NewQueryDepositsRequest(
				ownerBech,
				denom,
				valueInStakedTokens,
				pageReq,
			)
			res, err := queryClient.Deposits(context.Background(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddPaginationFlagsToCmd(cmd, "deposits")

	cmd.Flags().String(flagOwner, "", "(optional) filter for deposits by owner address")
	cmd.Flags().String(flagDenom, "", "(optional) filter for deposits by vault denom")
	cmd.Flags().Bool(flagValueInStakedTokens, false, "(optional) get underlying staked tokens for staking derivative vaults")

	return cmd
}

func queryTotalSupplyCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "total-supply",
		Short: "get total supply across all savings strategy vaults",
		Long:  "Get the sum of all denoms locking into vaults that allow the savings strategy.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.TotalSupply(context.Background(), &types.QueryTotalSupplyRequest{})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}
}
