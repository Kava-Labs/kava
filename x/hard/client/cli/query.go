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
	supplyexported "github.com/cosmos/cosmos-sdk/x/supply/exported"

	"github.com/kava-labs/kava/x/hard/types"
)

// flags for cli queries
const (
	flagName  = "name"
	flagDenom = "denom"
	flagOwner = "owner"
)

// GetQueryCmd returns the cli query commands for the  module
func GetQueryCmd(queryRoute string, cdc *codec.Codec) *cobra.Command {
	hardQueryCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Querying commands for the hard module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	hardQueryCmd.AddCommand(flags.GetCommands(
		queryParamsCmd(queryRoute, cdc),
		queryModAccountsCmd(queryRoute, cdc),
		queryDepositsCmd(queryRoute, cdc),
		queryTotalDepositedCmd(queryRoute, cdc),
		queryBorrowsCmd(queryRoute, cdc),
		queryTotalBorrowedCmd(queryRoute, cdc),
		queryInterestRateCmd(queryRoute, cdc),
	)...)

	return hardQueryCmd
}

func queryParamsCmd(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "params",
		Short: "get the hard module parameters",
		Long:  "Get the current global hard module parameters.",
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

func queryModAccountsCmd(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "accounts",
		Short: "query hard module accounts with optional filter",
		Long: strings.TrimSpace(`Query for all hard module accounts or a specific account using the name flag:

		Example:
		$ kvcli q hard accounts
		$ kvcli q hard accounts --name hard|hard_delegator_distribution|hard_lp_distribution`,
		),
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			name := viper.GetString(flagName)
			page := viper.GetInt(flags.FlagPage)
			limit := viper.GetInt(flags.FlagLimit)

			params := types.NewQueryAccountParams(page, limit, name)
			bz, err := cdc.MarshalJSON(params)
			if err != nil {
				return err
			}

			route := fmt.Sprintf("custom/%s/%s", queryRoute, types.QueryGetModuleAccounts)
			res, height, err := cliCtx.QueryWithData(route, bz)
			if err != nil {
				return err
			}
			cliCtx = cliCtx.WithHeight(height)

			var modAccounts []supplyexported.ModuleAccountI
			if err := cdc.UnmarshalJSON(res, &modAccounts); err != nil {
				return fmt.Errorf("failed to unmarshal module accounts: %w", err)
			}
			return cliCtx.PrintOutput(modAccounts)
		},
	}
}

func queryDepositsCmd(queryRoute string, cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deposits",
		Short: "query hard module deposits with optional filters",
		Long: strings.TrimSpace(`query for all hard module deposits or a specific deposit using flags:

		Example:
		$ kvcli q hard deposits
		$ kvcli q hard deposits --owner kava1l0xsq2z7gqd7yly0g40y5836g0appumark77ny --denom bnb
		$ kvcli q hard deposits --denom ukava
		$ kvcli q hard deposits --denom btcb`,
		),
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			var owner sdk.AccAddress

			ownerBech := viper.GetString(flagOwner)
			denom := viper.GetString(flagDenom)

			if len(ownerBech) != 0 {
				depositOwner, err := sdk.AccAddressFromBech32(ownerBech)
				if err != nil {
					return err
				}
				owner = depositOwner
			}

			page := viper.GetInt(flags.FlagPage)
			limit := viper.GetInt(flags.FlagLimit)

			params := types.NewQueryDepositsParams(page, limit, denom, owner)
			bz, err := cdc.MarshalJSON(params)
			if err != nil {
				return err
			}

			route := fmt.Sprintf("custom/%s/%s", queryRoute, types.QueryGetDeposits)
			res, height, err := cliCtx.QueryWithData(route, bz)
			if err != nil {
				return err
			}
			cliCtx = cliCtx.WithHeight(height)

			var deposits types.Deposits
			if err := cdc.UnmarshalJSON(res, &deposits); err != nil {
				return fmt.Errorf("failed to unmarshal deposits: %w", err)
			}
			return cliCtx.PrintOutput(deposits)
		},
	}

	cmd.Flags().Int(flags.FlagPage, 1, "pagination page to query for")
	cmd.Flags().Int(flags.FlagLimit, 100, "pagination limit (max 100)")
	cmd.Flags().String(flagOwner, "", "(optional) filter for deposits by owner address")
	cmd.Flags().String(flagDenom, "", "(optional) filter for deposits by denom")
	return cmd
}

func queryBorrowsCmd(queryRoute string, cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "borrows",
		Short: "query hard module borrows with optional filters",
		Long: strings.TrimSpace(`query for all hard module borrows or a specific borrow using flags:

		Example:
		$ kvcli q hard borrows
		$ kvcli q hard borrows --owner kava1l0xsq2z7gqd7yly0g40y5836g0appumark77ny
		$ kvcli q hard borrows --denom bnb`,
		),
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			var owner sdk.AccAddress

			ownerBech := viper.GetString(flagOwner)
			denom := viper.GetString(flagDenom)

			if len(ownerBech) != 0 {
				borrowOwner, err := sdk.AccAddressFromBech32(ownerBech)
				if err != nil {
					return err
				}
				owner = borrowOwner
			}

			page := viper.GetInt(flags.FlagPage)
			limit := viper.GetInt(flags.FlagLimit)

			params := types.NewQueryBorrowsParams(page, limit, owner, denom)
			bz, err := cdc.MarshalJSON(params)
			if err != nil {
				return err
			}

			route := fmt.Sprintf("custom/%s/%s", queryRoute, types.QueryGetBorrows)
			res, height, err := cliCtx.QueryWithData(route, bz)
			if err != nil {
				return err
			}
			cliCtx = cliCtx.WithHeight(height)

			var borrows types.Borrows
			if err := cdc.UnmarshalJSON(res, &borrows); err != nil {
				return fmt.Errorf("failed to unmarshal borrows: %w", err)
			}
			return cliCtx.PrintOutput(borrows)
		},
	}
	cmd.Flags().Int(flags.FlagPage, 1, "pagination page to query for")
	cmd.Flags().Int(flags.FlagLimit, 100, "pagination limit (max 100)")
	cmd.Flags().String(flagOwner, "", "(optional) filter for borrows by owner address")
	cmd.Flags().String(flagDenom, "", "(optional) filter for borrows by denom")
	return cmd
}

func queryTotalBorrowedCmd(queryRoute string, cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "total-borrowed",
		Short: "get total current borrowed amount",
		Long: strings.TrimSpace(`get the total amount of coins currently borrowed using flags:

		Example:
		$ kvcli q hard total-borrowed
		$ kvcli q hard total-borrowed --denom bnb`,
		),
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			denom := viper.GetString(flagDenom)

			// Construct query with params
			params := types.NewQueryTotalBorrowedParams(denom)
			bz, err := cdc.MarshalJSON(params)
			if err != nil {
				return err
			}

			// Execute query
			route := fmt.Sprintf("custom/%s/%s", queryRoute, types.QueryGetTotalBorrowed)
			res, height, err := cliCtx.QueryWithData(route, bz)
			if err != nil {
				return err
			}
			cliCtx = cliCtx.WithHeight(height)

			// Decode and print results
			var totalBorrowedCoins sdk.Coins
			if err := cdc.UnmarshalJSON(res, &totalBorrowedCoins); err != nil {
				return fmt.Errorf("failed to unmarshal total borrowed coins: %w", err)
			}
			return cliCtx.PrintOutput(totalBorrowedCoins)
		},
	}
	cmd.Flags().String(flagDenom, "", "(optional) filter total borrowed coins by denom")
	return cmd
}

func queryTotalDepositedCmd(queryRoute string, cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "total-deposited",
		Short: "get total current deposited amount",
		Long: strings.TrimSpace(`get the total amount of coins currently deposited using flags:

		Example:
		$ kvcli q hard total-deposited
		$ kvcli q hard total-deposited --denom bnb`,
		),
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			denom := viper.GetString(flagDenom)

			// Construct query with params
			params := types.NewQueryTotalDepositedParams(denom)
			bz, err := cdc.MarshalJSON(params)
			if err != nil {
				return err
			}

			// Execute query
			route := fmt.Sprintf("custom/%s/%s", queryRoute, types.QueryGetTotalDeposited)
			res, height, err := cliCtx.QueryWithData(route, bz)
			if err != nil {
				return err
			}
			cliCtx = cliCtx.WithHeight(height)

			// Decode and print results
			var totalSuppliedCoins sdk.Coins
			if err := cdc.UnmarshalJSON(res, &totalSuppliedCoins); err != nil {
				return fmt.Errorf("failed to unmarshal total deposited coins: %w", err)
			}
			return cliCtx.PrintOutput(totalSuppliedCoins)
		},
	}
	cmd.Flags().String(flagDenom, "", "(optional) filter total deposited coins by denom")
	return cmd
}

func queryInterestRateCmd(queryRoute string, cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "interest-rate",
		Short: "get current money market interest rates",
		Long: strings.TrimSpace(`get current money market interest rates:

		Example:
		$ kvcli q hard interest-rate
		$ kvcli q hard interest-rate --denom bnb`,
		),
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			denom := viper.GetString(flagDenom)

			// Construct query with params
			params := types.NewQueryInterestRateParams(denom)
			bz, err := cdc.MarshalJSON(params)
			if err != nil {
				return err
			}

			// Execute query
			route := fmt.Sprintf("custom/%s/%s", queryRoute, types.QueryGetInterestRate)
			res, height, err := cliCtx.QueryWithData(route, bz)
			if err != nil {
				return err
			}
			cliCtx = cliCtx.WithHeight(height)

			// Decode and print results
			var moneyMarketInterestRates types.MoneyMarketInterestRates
			if err := cdc.UnmarshalJSON(res, &moneyMarketInterestRates); err != nil {
				return fmt.Errorf("failed to unmarshal money market interest rates: %w", err)
			}
			return cliCtx.PrintOutput(moneyMarketInterestRates)
		},
	}
	cmd.Flags().String(flagDenom, "", "(optional) filter interest rates by denom")
	return cmd
}
