package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"

	"github.com/kava-labs/kava/x/kavadist/types"
)

// GetQueryCmd returns the cli query commands for the kavadist module
func GetQueryCmd(queryRoute string, cdc *codec.Codec) *cobra.Command {
	kavadistQueryCmd := &cobra.Command{
		Use:   types.ModuleName,
		Short: "Querying commands for the kavadist module",
	}

	kavadistQueryCmd.AddCommand(flags.GetCommands(
		queryParamsCmd(queryRoute, cdc),
	)...)

	return kavadistQueryCmd

}

func queryParamsCmd(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "params",
		Short: "get the kavadist module parameters",
		Long:  "Get the current global kavadist module parameters.",
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
