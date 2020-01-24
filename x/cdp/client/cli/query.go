package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"

	"github.com/kava-labs/kava/x/cdp/types"
)

// GetQueryCmd returns the cli query commands for this module
func GetQueryCmd(queryRoute string, cdc *codec.Codec) *cobra.Command {
	// Group nameservice queries under a subcommand
	cdpQueryCmd := &cobra.Command{
		Use:   "cdp",
		Short: "Querying commands for the cdp module",
	}

	cdpQueryCmd.AddCommand(client.GetCommands(
		QueryCdpCmd(queryRoute, cdc),
		QueryCdpsByDenomCmd(queryRoute, cdc),
		QueryCdpsByDenomAndRatioCmd(queryRoute, cdc),
		QueryCdpDepositsCmd(queryRoute, cdc),
		QueryParamsCmd(queryRoute, cdc),
	)...)

	return cdpQueryCmd
}

// QueryCdpCmd returns the command handler for querying a particular cdp
func QueryCdpCmd(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "cdp [owner-addr] [collateral-name]",
		Short: "get info about a cdp",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Get a CDP by the owner address and the collateral name.

Example:
$ %s query %s cdp kava15qdefkmwswysgg4qxgqpqr35k3m49pkx2jdfnw uatom
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
				CollateralDenom: args[1],
				Owner:           ownerAddress,
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
			var cdp types.CDP
			cdc.MustUnmarshalJSON(res, &cdp)
			return cliCtx.PrintOutput(cdp)
		},
	}
}

// QueryCdpsByDenomCmd returns the command handler for querying cdps for a collateral type
func QueryCdpsByDenomCmd(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "cdps [collateral-name]",
		Short: "query CDPs by collateral",
		Long: strings.TrimSpace(
			fmt.Sprintf(`List all CDPs collateralized with the specified asset.

Example:
$ %s query %s cdps uatom
`, version.ClientName, types.ModuleName)),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			// Prepare params for querier
			bz, err := cdc.MarshalJSON(types.QueryCdpsParams{CollateralDenom: args[0]})
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
			var out types.CDPs
			cdc.MustUnmarshalJSON(res, &out)
			return cliCtx.PrintOutput(out)
		},
	}
}

// QueryCdpsByDenomAndRatioCmd returns the command handler for querying cdps
// that are under the specified collateral ratio
func QueryCdpsByDenomAndRatioCmd(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "cdps-by-ratio [collateral-name] [collateralization-ratio]",
		Short: "get cdps under a collateralization ratio",
		Long: strings.TrimSpace(
			fmt.Sprintf(`List all CDPs under a collateralization ratios.
Collateralization ratio is: collateral * price / debt.

Example:
$ %s query %s cdps-by-ratio uatom 1.5
`, version.ClientName, types.ModuleName)),
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			// Prepare params for querier
			ratio, errSdk := sdk.NewDecFromStr(args[1])
			if errSdk != nil {
				return fmt.Errorf(errSdk.Error())
			}
			bz, err := cdc.MarshalJSON(types.QueryCdpsByRatioParams{
				CollateralDenom: args[0],
				Ratio:           ratio,
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
			var out types.CDPs
			cdc.MustUnmarshalJSON(res, &out)
			return cliCtx.PrintOutput(out)
		},
	}
}

// QueryCdpDepositsCmd returns the command handler for querying the deposits of a particular cdp
func QueryCdpDepositsCmd(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "cdp deposits [owner-addr] [collateral-name]",
		Short: "get deposits for a cdp",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Get the deposits of a CDP.

Example:
$ %s query %s cdp deposits kava15qdefkmwswysgg4qxgqpqr35k3m49pkx2jdfnw uatom
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
				CollateralDenom: args[1],
				Owner:           ownerAddress,
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
