package cli

import (
	"fmt"

	"github.com/kava-labs/kava/x/validator-vesting/internal/types"
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
)

// GetQueryValidatorVesting returns the cli query commands for this module
func GetQueryValidatorVesting(queryRoute string, cdc *codec.Codec) *cobra.Command {
	// Group nameservice queries under a subcommand
	queryValidatorVestingCmd := &cobra.Command{
		Use:   "validator-vesting",
		Short: "Querying commands for the validator vesting module",
	}

	queryValidatorVestingCmd.AddCommand(client.GetCommands(
		QueryCirculatingSupplyCmd(queryRoute, cdc),
	)...)
	return queryValidatorVestingCmd
}

// QueryCirculatingSupplyCmd queries the total circulating supply
func QueryCirculatingSupplyCmd(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "circulating-supply",
		Short: "Query circulating supply information",

		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s", queryRoute, types.QueryCirculatingSupply), nil)
			if err != nil {
				fmt.Printf("could not get total circulating supply\n")
				return nil
			}

			var out types.TotalCirculatingSupply
			cdc.MustUnmarshalJSON(res, &out)
			return cliCtx.PrintOutput(out)
		},
	}
}
