package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"
	supply "github.com/cosmos/cosmos-sdk/x/supply"

	"github.com/kava-labs/kava/x/cdp/types"
)

// GetQueryCmd returns the cli query commands for this module
func GetQueryCmd(queryRoute string, cdc *codec.Codec) *cobra.Command {
	// Group nameservice queries under a subcommand
	cdpQueryCmd := &cobra.Command{
		Use:   "cdp",
		Short: "Querying commands for the cdp module",
	}

	cdpQueryCmd.AddCommand(flags.GetCommands(
		QueryCdpCmd(queryRoute, cdc),
		QueryCdpsByCollateralTypeCmd(queryRoute, cdc),
		QueryCdpsByCollateralTypeAndRatioCmd(queryRoute, cdc),
		QueryCdpDepositsCmd(queryRoute, cdc),
		QueryParamsCmd(queryRoute, cdc),
		QueryGetAccounts(queryRoute, cdc),
	)...)

	return cdpQueryCmd
}

// QueryCdpCmd returns the command handler for querying a particular cdp
func QueryCdpCmd(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "cdp [owner-addr] [collateral-type]",
		Short: "get info about a cdp",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Get a CDP by the owner address and the collateral name.

Example:
$ %s query %s cdp kava15qdefkmwswysgg4qxgqpqr35k3m49pkx2jdfnw atom-a
`, version.ClientName, types.ModuleName)),
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			// Prepare params for querier
			ownerAddress, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}
			bz, err := cdc.MarshalJSON(types.QueryCdpParams{
				CollateralType: args[1],
				Owner:          ownerAddress,
			})
			if err != nil {
				return err
			}

			// Query
			route := fmt.Sprintf("custom/%s/%s", queryRoute, types.QueryGetCdp)
			res, _, err := cliCtx.QueryWithData(route, bz)
			if err != nil {
				return err
			}

			// Decode and print results
			var cdp types.AugmentedCDP
			cdc.MustUnmarshalJSON(res, &cdp)
			return cliCtx.PrintOutput(cdp)
		},
	}
}

// QueryCdpsByCollateralTypeCmd returns the command handler for querying cdps for a collateral type
func QueryCdpsByCollateralTypeCmd(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "cdps [collateral-type]",
		Short: "query CDPs by collateral",
		Long: strings.TrimSpace(
			fmt.Sprintf(`List all CDPs collateralized with the specified asset.

Example:
$ %s query %s cdps atom-a
`, version.ClientName, types.ModuleName)),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			// Prepare params for querier
			bz, err := cdc.MarshalJSON(types.QueryCdpsParams{CollateralType: args[0]})
			if err != nil {
				return err
			}

			// Query
			route := fmt.Sprintf("custom/%s/%s", queryRoute, types.QueryGetCdps)
			res, _, err := cliCtx.QueryWithData(route, bz)
			if err != nil {
				return err
			}

			// Decode and print results
			var cdps types.AugmentedCDPs
			cdc.MustUnmarshalJSON(res, &cdps)
			return cliCtx.PrintOutput(cdps)
		},
	}
}

// QueryCdpsByCollateralTypeAndRatioCmd returns the command handler for querying cdps
// that are under the specified collateral ratio
func QueryCdpsByCollateralTypeAndRatioCmd(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "cdps-by-ratio [collateral-type] [collateralization-ratio]",
		Short: "get cdps under a collateralization ratio",
		Long: strings.TrimSpace(
			fmt.Sprintf(`List all CDPs under a specified collateralization ratio.
Collateralization ratio is: collateral * price / debt.

Example:
$ %s query %s cdps-by-ratio atom-a 1.6
`, version.ClientName, types.ModuleName)),
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			// Prepare params for querier
			ratio, err := sdk.NewDecFromStr(args[1])
			if err != nil {
				return err
			}
			bz, err := cdc.MarshalJSON(types.QueryCdpsByRatioParams{
				CollateralType: args[0],
				Ratio:          ratio,
			})
			if err != nil {
				return err
			}

			// Query
			route := fmt.Sprintf("custom/%s/%s", queryRoute, types.QueryGetCdpsByCollateralization)
			res, _, err := cliCtx.QueryWithData(route, bz)
			if err != nil {
				return err
			}

			// Decode and print results
			var cdps types.AugmentedCDPs
			cdc.MustUnmarshalJSON(res, &cdps)
			return cliCtx.PrintOutput(cdps)
		},
	}
}

// QueryCdpDepositsCmd returns the command handler for querying the deposits of a particular cdp
func QueryCdpDepositsCmd(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "deposits [owner-addr] [collateral-type]",
		Short: "get deposits for a cdp",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Get the deposits of a CDP.

Example:
$ %s query %s deposits kava15qdefkmwswysgg4qxgqpqr35k3m49pkx2jdfnw atom-a
`, version.ClientName, types.ModuleName)),
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			// Prepare params for querier
			ownerAddress, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}
			bz, err := cdc.MarshalJSON(types.QueryCdpParams{
				CollateralType: args[1],
				Owner:          ownerAddress,
			})
			if err != nil {
				return err
			}

			// Query
			route := fmt.Sprintf("custom/%s/%s", queryRoute, types.QueryGetCdpDeposits)
			res, _, err := cliCtx.QueryWithData(route, bz)
			if err != nil {
				return err
			}

			// Decode and print results
			var deposits types.Deposits
			cdc.MustUnmarshalJSON(res, &deposits)
			return cliCtx.PrintOutput(deposits)
		},
	}
}

// QueryParamsCmd returns the command handler for cdp parameter querying
func QueryParamsCmd(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "params",
		Short: "get the cdp module parameters",
		Long:  "Get the current global cdp module parameters.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			// Query
			route := fmt.Sprintf("custom/%s/%s", queryRoute, types.QueryGetParams)
			res, _, err := cliCtx.QueryWithData(route, nil)
			if err != nil {
				return err
			}

			// Decode and print results
			var out types.Params
			cdc.MustUnmarshalJSON(res, &out)
			return cliCtx.PrintOutput(out)
		},
	}
}

func QueryGetAccounts(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "accounts",
		Short: "Get module accounts",
		Long:  "Get cdp module account addresses",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			// Query
			res, height, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s", queryRoute, types.QueryGetAccounts), nil)
			if err != nil {
				return err
			}
			cliCtx = cliCtx.WithHeight(height)

			// Decode and print results
			var out []supply.ModuleAccount
			if err := cdc.UnmarshalJSON(res, &out); err != nil {
				return fmt.Errorf("failed to unmarshal accounts: %w", err)
			}
			return cliCtx.PrintOutput(out)
		},
	}
}
