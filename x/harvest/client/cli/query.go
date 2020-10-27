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

	"github.com/kava-labs/kava/x/harvest/types"
)

// flags for cli queries
const (
	flagName         = "name"
	flagDepositDenom = "deposit-denom"
	flagBorrowDenom  = "borrow-denom"
	flagOwner        = "owner"
	flagDepositType  = "deposit-type"
)

// GetQueryCmd returns the cli query commands for the harvest module
func GetQueryCmd(queryRoute string, cdc *codec.Codec) *cobra.Command {
	harvestQueryCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Querying commands for the harvest module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	harvestQueryCmd.AddCommand(flags.GetCommands(
		queryParamsCmd(queryRoute, cdc),
		queryModAccountsCmd(queryRoute, cdc),
		queryDepositsCmd(queryRoute, cdc),
		queryClaimsCmd(queryRoute, cdc),
		queryBorrowsCmd(queryRoute, cdc),
	)...)

	return harvestQueryCmd

}

func queryParamsCmd(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "params",
		Short: "get the harvest module parameters",
		Long:  "Get the current global harvest module parameters.",
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
		Short: "query harvest module accounts with optional filter",
		Long: strings.TrimSpace(`Query for all harvest module accounts or a specific account using the name flag:

		Example:
		$ kvcli q harvest accounts
		$ kvcli q harvest accounts --name harvest|harvest_delegator_distribution|harvest_lp_distribution`,
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
		Short: "query harvest module deposits with optional filters",
		Long: strings.TrimSpace(`query for all harvest module deposits or a specific deposit using flags:

		Example:
		$ kvcli q harvest deposits
		$ kvcli q harvest deposits --owner kava1l0xsq2z7gqd7yly0g40y5836g0appumark77ny --deposit-type lp --deposit-denom bnb
		$ kvcli q harvest deposits --deposit-type stake --deposit-denom ukava
		$ kvcli q harvest deposits --deposit-denom btcb`,
		),
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			var owner sdk.AccAddress
			var depositType types.DepositType

			ownerBech := viper.GetString(flagOwner)
			depositDenom := viper.GetString(flagDepositDenom)
			depositTypeStr := viper.GetString(flagDepositType)

			if len(ownerBech) != 0 {
				depositOwner, err := sdk.AccAddressFromBech32(ownerBech)
				if err != nil {
					return err
				}
				owner = depositOwner
			}

			if len(depositTypeStr) != 0 {
				if err := types.DepositType(depositTypeStr).IsValid(); err != nil {
					return err
				}
				depositType = types.DepositType(depositTypeStr)
			}

			page := viper.GetInt(flags.FlagPage)
			limit := viper.GetInt(flags.FlagLimit)

			params := types.NewQueryDepositParams(page, limit, depositDenom, owner, depositType)
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

			var deposits []types.Deposit
			if err := cdc.UnmarshalJSON(res, &deposits); err != nil {
				return fmt.Errorf("failed to unmarshal deposits: %w", err)
			}
			return cliCtx.PrintOutput(deposits)
		},
	}
	cmd.Flags().Int(flags.FlagPage, 1, "pagination page to query for")
	cmd.Flags().Int(flags.FlagLimit, 100, "pagination limit (max 100)")
	cmd.Flags().String(flagOwner, "", "(optional) filter for deposits by owner address")
	cmd.Flags().String(flagDepositDenom, "", "(optional) filter for deposits by denom")
	cmd.Flags().String(flagDepositType, "", "(optional) filter for deposits by type (lp or staking)")
	return cmd
}

func queryClaimsCmd(queryRoute string, cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "claims",
		Short: "query harvest module claims with optional filters",
		Long: strings.TrimSpace(`query for all harvest module claims or a specific claim using flags:

		Example:
		$ kvcli q harvest claims
		$ kvcli q harvest claims --owner kava1l0xsq2z7gqd7yly0g40y5836g0appumark77ny --deposit-type lp --deposit-denom bnb
		$ kvcli q harvest claims --deposit-type stake --deposit-denom ukava
		$ kvcli q harvest claims --deposit-denom btcb`,
		),
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			var owner sdk.AccAddress
			var depositType types.DepositType

			ownerBech := viper.GetString(flagOwner)
			depositDenom := viper.GetString(flagDepositDenom)
			depositTypeStr := viper.GetString(flagDepositType)

			if len(ownerBech) != 0 {
				claimOwner, err := sdk.AccAddressFromBech32(ownerBech)
				if err != nil {
					return err
				}
				owner = claimOwner
			}

			if len(depositTypeStr) != 0 {
				if err := types.DepositType(depositTypeStr).IsValid(); err != nil {
					return err
				}
				depositType = types.DepositType(depositTypeStr)
			}

			page := viper.GetInt(flags.FlagPage)
			limit := viper.GetInt(flags.FlagLimit)

			params := types.NewQueryDepositParams(page, limit, depositDenom, owner, depositType)
			bz, err := cdc.MarshalJSON(params)
			if err != nil {
				return err
			}

			route := fmt.Sprintf("custom/%s/%s", queryRoute, types.QueryGetClaims)
			res, height, err := cliCtx.QueryWithData(route, bz)
			if err != nil {
				return err
			}
			cliCtx = cliCtx.WithHeight(height)

			var claims []types.Claim
			if err := cdc.UnmarshalJSON(res, &claims); err != nil {
				return fmt.Errorf("failed to unmarshal claims: %w", err)
			}
			return cliCtx.PrintOutput(claims)
		},
	}
	cmd.Flags().Int(flags.FlagPage, 1, "pagination page to query for")
	cmd.Flags().Int(flags.FlagLimit, 100, "pagination limit (max 100)")
	cmd.Flags().String(flagOwner, "", "(optional) filter for claims by owner address")
	cmd.Flags().String(flagDepositDenom, "", "(optional) filter for claims by denom")
	cmd.Flags().String(flagDepositType, "", "(optional) filter for claims by type (lp or staking)")
	return cmd
}

func queryBorrowsCmd(queryRoute string, cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "borrows",
		Short: "query harvest module borrows with optional filters",
		Long: strings.TrimSpace(`query for all harvest module borrows or a specific borrow using flags:

		Example:
		$ kvcli q harvest borrows
		$ kvcli q harvest borrows --borrower kava1l0xsq2z7gqd7yly0g40y5836g0appumark77ny
		$ kvcli q harvest borrows --borrow-denom bnb`,
		),
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			var owner sdk.AccAddress

			ownerBech := viper.GetString(flagOwner)
			depositDenom := viper.GetString(flagDepositDenom)

			if len(ownerBech) != 0 {
				borrowOwner, err := sdk.AccAddressFromBech32(ownerBech)
				if err != nil {
					return err
				}
				owner = borrowOwner
			}

			page := viper.GetInt(flags.FlagPage)
			limit := viper.GetInt(flags.FlagLimit)

			params := types.NewQueryBorrowParams(page, limit, owner, depositDenom)
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

			var borrows []types.Borrow
			if err := cdc.UnmarshalJSON(res, &borrows); err != nil {
				return fmt.Errorf("failed to unmarshal borrows: %w", err)
			}
			return cliCtx.PrintOutput(borrows)
		},
	}
	cmd.Flags().Int(flags.FlagPage, 1, "pagination page to query for")
	cmd.Flags().Int(flags.FlagLimit, 100, "pagination limit (max 100)")
	cmd.Flags().String(flagOwner, "", "(optional) filter for borrows by owner address")
	cmd.Flags().String(flagBorrowDenom, "", "(optional) filter for borrows by denom")
	return cmd
}
