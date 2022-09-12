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
	flagDenom = "denom"
	flagOwner = "owner"
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
		queryDepositsCmd(),
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
		Use:   "vaults",
		Short: "get the earn vaults",
		Long:  "Get the earn module vaults.",
		Args:  cobra.MaximumNArgs(1),
		Example: fmt.Sprintf(`%[1]s q %[2]s vaults
%[1]s q %[2]s vaults
%[1]s q %[2]s vaults usdx`, version.AppName, types.ModuleName),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			vaultDenom := ""
			if len(args) > 1 {
				vaultDenom = args[0]
			}

			req := types.NewQueryVaultsRequest(vaultDenom)
			res, err := queryClient.Vaults(context.Background(), req)
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

			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			req := types.NewQueryDepositsRequest(
				ownerBech,
				denom,
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

	return cmd
}
