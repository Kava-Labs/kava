package cli

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"

	"github.com/kava-labs/kava/x/validator-vesting/types"
)

// GetQueryCmd returns the cli query commands for the kavadist module
func GetQueryCmd() *cobra.Command {
	valVestingQueryCmd := &cobra.Command{
		Use:   types.QueryPath,
		Short: "Querying commands for the validator vesting module",
	}

	cmds := []*cobra.Command{
		queryCirculatingSupply(),
		queryTotalSupply(),
		queryCirculatingSupplyHARD(),
		queryCirculatingSupplyUSDX(),
		queryCirculatingSupplySWP(),
		queryTotalSupplyHARD(),
		queryTotalSupplyUSDX(),
	}

	for _, cmd := range cmds {
		flags.AddQueryFlagsToCmd(cmd)
	}

	valVestingQueryCmd.AddCommand(cmds...)
	return valVestingQueryCmd
}

func queryCirculatingSupply() *cobra.Command {
	return &cobra.Command{
		Use:   "circulating-supply",
		Short: "Get circulating supply",
		Long:  "Get the current circulating supply of kava tokens",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(cliCtx)
			res, err := queryClient.CirculatingSupply(context.Background(), &types.QueryCirculatingSupplyRequest{})
			if err != nil {
				return err
			}
			return cliCtx.PrintString(res.Amount.String())
		},
	}
}

func queryTotalSupply() *cobra.Command {
	return &cobra.Command{
		Use:   "total-supply",
		Short: "Get total supply",
		Long:  "Get the current total supply of kava tokens",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(cliCtx)
			res, err := queryClient.TotalSupply(context.Background(), &types.QueryTotalSupplyRequest{})
			if err != nil {
				return err
			}
			return cliCtx.PrintString(res.Amount.String())
		},
	}
}

func queryCirculatingSupplyHARD() *cobra.Command {
	return &cobra.Command{
		Use:   "circulating-supply-hard",
		Short: "Get HARD circulating supply",
		Long:  "Get the current circulating supply of HARD tokens",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(cliCtx)
			res, err := queryClient.CirculatingSupplyHARD(context.Background(), &types.QueryCirculatingSupplyHARDRequest{})
			if err != nil {
				return err
			}
			return cliCtx.PrintString(res.Amount.String())
		},
	}
}

func queryCirculatingSupplyUSDX() *cobra.Command {
	return &cobra.Command{
		Use:   "circulating-supply-usdx",
		Short: "Get USDX circulating supply",
		Long:  "Get the current circulating supply of USDX tokens",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(cliCtx)
			res, err := queryClient.CirculatingSupplyUSDX(context.Background(), &types.QueryCirculatingSupplyUSDXRequest{})
			if err != nil {
				return err
			}
			return cliCtx.PrintString(res.Amount.String())
		},
	}
}

func queryCirculatingSupplySWP() *cobra.Command {
	return &cobra.Command{
		Use:   "circulating-supply-swp",
		Short: "Get SWP circulating supply",
		Long:  "Get the current circulating supply of SWP tokens",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(cliCtx)
			res, err := queryClient.CirculatingSupplySWP(context.Background(), &types.QueryCirculatingSupplySWPRequest{})
			if err != nil {
				return err
			}
			return cliCtx.PrintString(res.Amount.String())
		},
	}
}

func queryTotalSupplyHARD() *cobra.Command {
	return &cobra.Command{
		Use:   "total-supply-hard",
		Short: "Get HARD total supply",
		Long:  "Get the current total supply of HARD tokens",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(cliCtx)
			res, err := queryClient.TotalSupplyHARD(context.Background(), &types.QueryTotalSupplyHARDRequest{})
			if err != nil {
				return err
			}
			return cliCtx.PrintString(res.Amount.String())
		},
	}
}

func queryTotalSupplyUSDX() *cobra.Command {
	return &cobra.Command{
		Use:   "total-supply-usdx",
		Short: "Get USDX total supply",
		Long:  "Get the current total supply of USDX tokens",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(cliCtx)
			res, err := queryClient.TotalSupplyUSDX(context.Background(), &types.QueryTotalSupplyUSDXRequest{})
			if err != nil {
				return err
			}
			return cliCtx.PrintString(res.Amount.String())
		},
	}
}
