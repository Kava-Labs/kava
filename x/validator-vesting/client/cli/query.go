package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/x/validator-vesting/types"
)

// GetQueryCmd returns the cli query commands for the kavadist module
func GetQueryCmd(queryRoute string, cdc *codec.Codec) *cobra.Command {
	valVestingQueryCmd := &cobra.Command{
		Use:   types.QueryPath,
		Short: "Querying commands for the validator vesting module",
	}

	valVestingQueryCmd.AddCommand(flags.GetCommands(
		queryCirculatingSupply(queryRoute, cdc),
		queryTotalSupply(queryRoute, cdc),
		queryCirculatingSupplyHARD(queryRoute, cdc),
		queryCirculatingSupplyUSDX(queryRoute, cdc),
		queryTotalSupplyHARD(queryRoute, cdc),
		queryTotalSupplyUSDX(queryRoute, cdc),
		querySpendableBalance(queryRoute, cdc),
	)...)

	return valVestingQueryCmd

}

func queryCirculatingSupply(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "circulating-supply",
		Short: "Get circulating supply",
		Long:  "Get the current circulating supply of kava tokens",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			// Query
			res, height, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s", queryRoute, types.QueryCirculatingSupply), nil)
			if err != nil {
				return err
			}
			cliCtx = cliCtx.WithHeight(height)

			// Decode and print results
			var out int64
			if err := cdc.UnmarshalJSON(res, &out); err != nil {
				return fmt.Errorf("failed to unmarshal supply: %w", err)
			}
			return cliCtx.PrintOutput(out)
		},
	}
}

func queryTotalSupply(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "total-supply",
		Short: "Get total supply",
		Long:  "Get the current total supply of kava tokens",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			// Query
			res, height, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s", queryRoute, types.QueryTotalSupply), nil)
			if err != nil {
				return err
			}
			cliCtx = cliCtx.WithHeight(height)

			// Decode and print results
			var out int64
			if err := cdc.UnmarshalJSON(res, &out); err != nil {
				return fmt.Errorf("failed to unmarshal supply: %w", err)
			}
			return cliCtx.PrintOutput(out)
		},
	}
}

func queryCirculatingSupplyHARD(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "circulating-supply-hard",
		Short: "Get HARD circulating supply",
		Long:  "Get the current circulating supply of HARD tokens",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			// Query
			res, height, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s", queryRoute, types.QueryCirculatingSupplyHARD), nil)
			if err != nil {
				return err
			}
			cliCtx = cliCtx.WithHeight(height)

			// Decode and print results
			var out int64
			if err := cdc.UnmarshalJSON(res, &out); err != nil {
				return fmt.Errorf("failed to unmarshal supply: %w", err)
			}
			return cliCtx.PrintOutput(out)
		},
	}
}

func queryCirculatingSupplyUSDX(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "circulating-supply-usdx",
		Short: "Get USDX circulating supply",
		Long:  "Get the current circulating supply of USDX tokens",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			// Query
			res, height, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s", queryRoute, types.QueryCirculatingSupplyUSDX), nil)
			if err != nil {
				return err
			}
			cliCtx = cliCtx.WithHeight(height)

			// Decode and print results
			var out int64
			if err := cdc.UnmarshalJSON(res, &out); err != nil {
				return fmt.Errorf("failed to unmarshal supply: %w", err)
			}
			return cliCtx.PrintOutput(out)
		},
	}
}

func queryTotalSupplyHARD(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "total-supply-hard",
		Short: "Get HARD total supply",
		Long:  "Get the current total supply of HARD tokens",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			// Query
			res, height, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s", queryRoute, types.QueryTotalSupplyHARD), nil)
			if err != nil {
				return err
			}
			cliCtx = cliCtx.WithHeight(height)

			// Decode and print results
			var out int64
			if err := cdc.UnmarshalJSON(res, &out); err != nil {
				return fmt.Errorf("failed to unmarshal supply: %w", err)
			}
			return cliCtx.PrintOutput(out)
		},
	}
}

func queryTotalSupplyUSDX(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "total-supply-usdx",
		Short: "Get USDX total supply",
		Long:  "Get the current total supply of USDX tokens",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			// Query
			res, height, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s", queryRoute, types.QueryTotalSupplyUSDX), nil)
			if err != nil {
				return err
			}
			cliCtx = cliCtx.WithHeight(height)

			// Decode and print results
			var out int64
			if err := cdc.UnmarshalJSON(res, &out); err != nil {
				return fmt.Errorf("failed to unmarshal supply: %w", err)
			}
			return cliCtx.PrintOutput(out)
		},
	}
}

func querySpendableBalance(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "spendable-balance [address]",
		Short: "get the spendable balance for the input address",
		Long:  "Get the current spendable balance for the input address, defined as the coins in the account which are not currently vesting or delegated (staked).",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			bech32Address := args[0]
			addr, err := sdk.AccAddressFromBech32(bech32Address)
			if err != nil {
				return err
			}
			params := types.NewSpendableBalanceParams(addr)
			bz, err := cdc.MarshalJSON(params)
			if err != nil {
				return err
			}
			// Query
			res, height, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s", queryRoute, types.QuerySpendableBalance), bz)
			if err != nil {
				return err
			}
			cliCtx = cliCtx.WithHeight(height)

			// Decode and print results
			var out sdk.Coins
			if err := cdc.UnmarshalJSON(res, &out); err != nil {
				return fmt.Errorf("failed to unmarshal spendable balance: %w", err)
			}
			return cliCtx.PrintOutput(out)
		},
	}
}
