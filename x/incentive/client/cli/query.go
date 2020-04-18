package cli

import (
	"fmt"
	"strings"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/kava-labs/kava/x/incentive/types"
	"github.com/spf13/cobra"
)

// GetQueryCmd returns the cli query commands for the incentive module
func GetQueryCmd(queryRoute string, cdc *codec.Codec) *cobra.Command {
	incentiveQueryCmd := &cobra.Command{
		Use:   types.ModuleName,
		Short: "Querying commands for the incentive module",
	}

	incentiveQueryCmd.AddCommand(client.GetCommands(
		queryParamsCmd(queryRoute, cdc),
		queryClaimsCmd(queryRoute, cdc),
	)...)

	return incentiveQueryCmd

}

func queryClaimsCmd(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "claims [owner-addr] [denom]",
		Short: "get claims by onwer and denom",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Get all claims owned by the owner address for the particular collateral type.

			Example:
			$ %s query %s claims kava15qdefkmwswysgg4qxgqpqr35k3m49pkx2jdfnw bnb`, version.ClientName, types.ModuleName)),
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			// Prepare params for querier
			ownerAddress, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}
			bz, err := cdc.MarshalJSON(types.QueryClaimsParams{
				Owner: ownerAddress,
				Denom: args[1],
			})
			if err != nil {
				return err
			}

			// Query
			route := fmt.Sprintf("custom/%s/%s", queryRoute, types.QueryGetClaims)
			res, _, err := cliCtx.QueryWithData(route, bz)
			if err != nil {
				return err
			}

			var claims types.Claims
			if err := cdc.UnmarshalJSON(res, &claims); err != nil {
				return fmt.Errorf("failed to unmarshal claims: %w", err)
			}
			return cliCtx.PrintOutput(claims)

		},
	}
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
			res, _, err := cliCtx.QueryWithData(route, nil)
			if err != nil {
				return err
			}

			// Decode and print results
			var params types.Params
			if err := cdc.UnmarshalJSON(res, &params); err != nil {
				return fmt.Errorf("failed to unmarshal params: %w", err)
			}
			return cliCtx.PrintOutput(params)
		},
	}
}
