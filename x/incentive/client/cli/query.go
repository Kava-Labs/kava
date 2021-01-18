package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"

	"github.com/kava-labs/kava/x/incentive/types"
)

// GetQueryCmd returns the cli query commands for the incentive module
func GetQueryCmd(queryRoute string, cdc *codec.Codec) *cobra.Command {
	incentiveQueryCmd := &cobra.Command{
		Use:   types.ModuleName,
		Short: "Querying commands for the incentive module",
	}

	incentiveQueryCmd.AddCommand(flags.GetCommands(
		queryParamsCmd(queryRoute, cdc),
		queryClaimsCmd(queryRoute, cdc),
	)...)

	return incentiveQueryCmd
}

const (
	flagOwner = "owner"
)

func queryClaimsCmd(queryRoute string, cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "claims ",
		Short: "query USDX minting claims",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query USDX minting claims with optional flag for finding claims for a specifc owner

			Example:
			$ %s query %s claims
			$ %s query %s claims --owner kava15qdefkmwswysgg4qxgqpqr35k3m49pkx2jdfnw
			`,
				version.ClientName, types.ModuleName, version.ClientName, types.ModuleName)),
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			strOwner := viper.GetString(flagOwner)
			page := viper.GetInt(flags.FlagPage)
			limit := viper.GetInt(flags.FlagLimit)

			// Prepare params for querier
			owner, err := sdk.AccAddressFromBech32(strOwner)
			if err != nil {
				return err
			}
			params := types.NewQueryClaimsParams(page, limit, owner)
			bz, err := cdc.MarshalJSON(params)
			if err != nil {
				return err
			}

			// Query
			route := fmt.Sprintf("custom/%s/%s", queryRoute, types.QueryGetClaims)
			res, height, err := cliCtx.QueryWithData(route, bz)
			if err != nil {
				return err
			}
			cliCtx = cliCtx.WithHeight(height)

			var claims types.USDXMintingClaims
			if err := cdc.UnmarshalJSON(res, &claims); err != nil {
				return fmt.Errorf("failed to unmarshal claims: %w", err)
			}
			return cliCtx.PrintOutput(claims)

		},
	}
	cmd.Flags().String(flagOwner, "", "(optional) filter by claim owner address")
	cmd.Flags().Int(flags.FlagPage, 1, "pagination page of CDPs to to query for")
	cmd.Flags().Int(flags.FlagLimit, 100, "pagination limit of CDPs to query for")
	return cmd
}

func queryParamsCmd(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "params",
		Short: "get the incentive module parameters",
		Long:  "Get the current global incentive module parameters.",
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
