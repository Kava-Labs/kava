package cli

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"
	supply "github.com/cosmos/cosmos-sdk/x/supply"

	"github.com/kava-labs/kava/x/cdp/types"
)

// Query CDP flags
const (
	flagCollateralType = "collateral-type"
	flagOwner          = "owner"
	flagID             = "id"
	flagRatio          = "ratio" // returns CDPs under the given collateralization ratio threshold
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
		QueryGetCdpsCmd(queryRoute, cdc),
		QueryCdpDepositsCmd(queryRoute, cdc),
		QueryParamsCmd(queryRoute, cdc),
		QueryGetAccounts(queryRoute, cdc),
		QueryGetSavingsRateDistributed(queryRoute, cdc),
		QueryGetSavingsRateDistTime(queryRoute, cdc),
		QueryGetFees(queryRoute, cdc),
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

// QueryGetCdpsCmd queries the cdps in the store
func QueryGetCdpsCmd(queryRoute string, cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cdps",
		Short: "query cdps with optional filters",
		Long: strings.TrimSpace(`Query for all paginated cdps that match optional filters:
Example:
$ kvcli q cdp cdps --collateral-type=bnb
$ kvcli q cdp cdps --owner=kava1hatdq32u5x4wnxrtv5wzjzmq49sxgjgsj0mffm
$ kvcli q cdp cdps --id=21
$ kvcli q cdp cdps --ratio=2.75
$ kvcli q cdp cdps --page=2 --limit=100
`,
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			strCollateralType := viper.GetString(flagCollateralType)
			strOwner := viper.GetString(flagOwner)
			strID := viper.GetString(flagID)
			strRatio := viper.GetString(flagRatio)
			page := viper.GetInt(flags.FlagPage)
			limit := viper.GetInt(flags.FlagLimit)

			var (
				cdpCollateralType string
				cdpOwner          sdk.AccAddress
				cdpID             uint64
				cdpRatio          sdk.Dec
			)

			params := types.NewQueryCdpsParams(page, limit, cdpCollateralType, cdpOwner, cdpID, cdpRatio)

			if len(strCollateralType) != 0 {
				cdpCollateralType = strings.ToLower(strings.TrimSpace(strCollateralType))
				params.CollateralType = cdpCollateralType
			}

			if len(strOwner) != 0 {
				cdpOwner, err := sdk.AccAddressFromBech32(strings.ToLower(strings.TrimSpace(strOwner)))
				if err != nil {
					return fmt.Errorf("cannot parse address from cdp owner %s", strOwner)
				}
				params.Owner = cdpOwner
			}

			if len(strID) != 0 {
				cdpID, err := strconv.ParseUint(strID, 10, 64)
				if err != nil {
					return fmt.Errorf("cannot parse cdp ID %s", strID)
				}
				params.ID = cdpID
			}

			params.Ratio = sdk.ZeroDec()
			if len(strRatio) != 0 {
				cdpRatio, err := sdk.NewDecFromStr(strRatio)
				if err != nil {
					return fmt.Errorf("cannot parse cdp ratio %s", strRatio)
				}
				params.Ratio = cdpRatio
			} else {
				// Set to sdk.Dec(0) so that if not specified in params it doesn't panic when unmarshaled
				params.Ratio = sdk.ZeroDec()
			}

			bz, err := cdc.MarshalJSON(params)
			if err != nil {
				return err
			}

			cliCtx := context.NewCLIContext().WithCodec(cdc)

			// Query
			res, height, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s", queryRoute, types.QueryGetCdps), bz)
			if err != nil {
				return err
			}

			// Decode and print results
			var matchingCDPs types.AugmentedCDPs
			cdc.MustUnmarshalJSON(res, &matchingCDPs)
			if len(matchingCDPs) == 0 {
				return fmt.Errorf("No matching CDPs found")
			}

			cliCtx = cliCtx.WithHeight(height)
			return cliCtx.PrintOutput(matchingCDPs) // nolint:errcheck
		},
	}

	cmd.Flags().Int(flags.FlagPage, 1, "pagination page of CDPs to to query for")
	cmd.Flags().Int(flags.FlagLimit, 100, "pagination limit of CDPs to query for")
	cmd.Flags().String(flagCollateralType, "", "(optional) filter by CDP collateral type")
	cmd.Flags().String(flagOwner, "", "(optional) filter by CDP owner")
	cmd.Flags().String(flagID, "", "(optional) filter by CDP ID")
	cmd.Flags().String(flagRatio, "", "(optional) filter by CDP collateralization ratio threshold")

	return cmd
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
		Long:  "get the current global cdp module parameters.",
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

// QueryGetAccounts queries CDP module accounts
func QueryGetAccounts(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "accounts",
		Short: "get module accounts",
		Long:  "get cdp module account addresses",
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

// QueryGetSavingsRateDistributed queries the total amount of savings rate distributed in USDX
func QueryGetSavingsRateDistributed(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "savings-rate-dist",
		Short: "get total amount of savings rate distributed in USDX",
		Long:  "get total amount of savings rate distributed",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			// Query
			res, height, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s", queryRoute, types.QueryGetSavingsRateDistributed), nil)
			if err != nil {
				return err
			}
			cliCtx = cliCtx.WithHeight(height)

			// Decode and print results
			var out sdk.Int
			if err := cdc.UnmarshalJSON(res, &out); err != nil {
				return fmt.Errorf("failed to unmarshal sdk.Int: %w", err)
			}
			return cliCtx.PrintOutput(out)
		},
	}
}

// QueryGetSavingsRateDistTime queries the total amount of savings rate distributed in USDX
func QueryGetSavingsRateDistTime(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "savings-rate-dist-time",
		Short: "get the previous savings rate distribution time",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			// Query
			res, height, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s", queryRoute, types.QueryGetPreviousSavingsDistributionTime), nil)
			if err != nil {
				return err
			}
			cliCtx = cliCtx.WithHeight(height)

			// Decode and print results
			var out time.Time
			if err := cdc.UnmarshalJSON(res, &out); err != nil {
				return fmt.Errorf("failed to unmarshal time.Time: %w", err)
			}
			return cliCtx.PrintOutput(out)
		},
	}
}

// QueryGetFees queries the total amount of fees collected of a fee coin's denom
func QueryGetFees(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "fees",
		Short: "get the total fees paid in the specified coin denom",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Get the total paid fees for a given coin denom

Example:
$ %s query %s fees usdx
`, version.ClientName, types.ModuleName)),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			// Prepare params for querier
			bz, err := cdc.MarshalJSON(types.NewQueryFees(args[0]))
			if err != nil {
				return err
			}

			// Query
			route := fmt.Sprintf("custom/%s/%s", queryRoute, types.QueryGetFees)
			res, height, err := cliCtx.QueryWithData(route, bz)
			if err != nil {
				return err
			}
			cliCtx = cliCtx.WithHeight(height)

			// Decode and print results
			var out sdk.Int
			if err := cdc.UnmarshalJSON(res, &out); err != nil {
				return fmt.Errorf("failed to unmarshal sdk.Int: %w", err)
			}
			return cliCtx.PrintOutput(out)
		},
	}
}
