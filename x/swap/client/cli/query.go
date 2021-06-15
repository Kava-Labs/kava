package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/x/swap/types"
)

// flags for cli queries
const (
	flagOwner = "owner"
	flagPool  = "pool"
)

// GetQueryCmd returns the cli query commands for the  module
func GetQueryCmd(queryRoute string, cdc *codec.Codec) *cobra.Command {
	swapQueryCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Querying commands for the swap module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	swapQueryCmd.AddCommand(flags.GetCommands(
		queryParamsCmd(queryRoute, cdc),
		queryDepositsCmd(queryRoute, cdc),
		queryPoolCmd(queryRoute, cdc),
		queryPoolsCmd(queryRoute, cdc),
	)...)

	return swapQueryCmd
}

func queryParamsCmd(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "params",
		Short: "get the swap module parameters",
		Long:  "Get the current global swap module parameters.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			// Query
			route := fmt.Sprintf("custom/%s/%s", queryRoute, types.QueryGetParams)
			res, height, err := cliCtx.QueryWithData(route, nil)
			if err != nil {
				return err
			}
			cliCtx = cliCtx.WithHeight(height)

			// Decode and print results
			var params types.Params
			if err := cdc.UnmarshalJSON(res, &params); err != nil {
				return fmt.Errorf("failed to unmarshal params: %w", err)
			}
			return cliCtx.PrintOutput(params)
		},
	}
}

func queryDepositsCmd(queryRoute string, cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deposits",
		Short: "get liquidity provider deposits for a given liquidity pool",
		Long: strings.TrimSpace(`get liquidity provider deposits for a given liquidity pool:

		Example:
		$ kvcli q swap deposits --pool bnb/usdx
		$ kvcli q swap deposits --pool bnb/usdx --owner kava1l0xsq2z7gqd7yly0g40y5836g0appumark77ny`,
		),
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			pool := viper.GetString(flagPool)
			if len(pool) == 0 {
				return fmt.Errorf("must specify param 'pool'")
			}

			var owner sdk.AccAddress
			ownerBech := viper.GetString(flagOwner)
			if len(ownerBech) == 0 {
				return fmt.Errorf("must specify param 'owner'")
			}
			shareOwner, err := sdk.AccAddressFromBech32(ownerBech)
			if err != nil {
				return err
			}
			owner = shareOwner

			// Construct query with params
			params := types.NewQueryDepositsParams(owner, pool)
			bz, err := cdc.MarshalJSON(params)
			if err != nil {
				return err
			}

			// Execute query
			route := fmt.Sprintf("custom/%s/%s", queryRoute, types.QueryGetDeposits)
			res, height, err := cliCtx.QueryWithData(route, bz)
			if err != nil {
				return err
			}
			cliCtx = cliCtx.WithHeight(height)

			// Decode and print results
			var depositCoins sdk.Coins
			if err := cdc.UnmarshalJSON(res, &depositCoins); err != nil {
				return fmt.Errorf("failed to unmarshal coins: %w", err)
			}
			return cliCtx.PrintOutput(depositCoins)
		},
	}
	cmd.Flags().String(flagPool, "", "pool name")
	cmd.Flags().String(flagOwner, "", "share owner, also known as a liquidity provider")
	return cmd
}

func queryPoolCmd(queryRoute string, cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pool",
		Short: "get pool statistics",
		Long: strings.TrimSpace(`get statistics about a given liquidity pool:

		Example:
		$ kvcli q swap pool ukava/usdx`,
		),
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			poolName := args[0]
			if len(poolName) == 0 {
				return fmt.Errorf("must specify pool")
			}

			// Construct query with params
			params := types.NewQueryPoolParams(poolName)
			bz, err := cdc.MarshalJSON(params)
			if err != nil {
				return err
			}

			// Execute query
			route := fmt.Sprintf("custom/%s/%s", queryRoute, types.QueryGetPool)
			res, height, err := cliCtx.QueryWithData(route, bz)
			if err != nil {
				return err
			}
			cliCtx = cliCtx.WithHeight(height)

			var poolStats types.PoolStatsQueryResult
			if err := cdc.UnmarshalJSON(res, &poolStats); err != nil {
				return fmt.Errorf("failed to unmarshal pool stats: %w", err)
			}
			return cliCtx.PrintOutput(poolStats)
		},
	}
	cmd.Flags().String(flagPool, "", "pool name")
	return cmd
}

func queryPoolsCmd(queryRoute string, cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pools",
		Short: "get statistics for all pools",
		Long: strings.TrimSpace(`get statistics for all liquidity pools:

		Example:
		$ kvcli q swap pools`,
		),
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			// Execute query
			route := fmt.Sprintf("custom/%s/%s", queryRoute, types.QueryGetPools)
			res, height, err := cliCtx.QueryWithData(route, nil)
			if err != nil {
				return err
			}
			cliCtx = cliCtx.WithHeight(height)

			var poolStats types.PoolStatsQueryResults
			if err := cdc.UnmarshalJSON(res, &poolStats); err != nil {
				return fmt.Errorf("failed to unmarshal pools' stats: %w", err)
			}
			return cliCtx.PrintOutput(poolStats)
		},
	}
	return cmd
}
