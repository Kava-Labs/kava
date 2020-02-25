package cli

import (
	"fmt"

	"github.com/kava-labs/kava/x/validator-vesting/internal/types"
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GetQueryCmd returns the cli query commands for this module
func GetQueryCmd(queryRoute string, cdc *codec.Codec) *cobra.Command {
	// Group nameservice queries under a subcommand
	queryValidatorVestingCmd := &cobra.Command{
		Use:   "validator-vesting",
		Short: "Querying commands for the validator vesting module",
	}

	queryValidatorVestingCmd.AddCommand(client.GetCommands(
		QueryCirculatingSupplyCmd(queryRoute, cdc),
		QueryTotalSupplyCmd(queryRoute, cdc),
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
				return err
			}

			var out sdk.Dec
			cdc.MustUnmarshalJSON(res, &out)
			return cliCtx.PrintOutput(out)
		},
	}
}

// QueryTotalSupplyCmd queries the total supply of ukava
func QueryTotalSupplyCmd(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "total-supply",
		Short: "Query total supply information",

		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s", queryRoute, types.QueryTotalSupply), nil)
			if err != nil {
				fmt.Printf("could not get total supply\n")
				return err
			}

			var out sdk.Dec
			cdc.MustUnmarshalJSON(res, &out)
			return cliCtx.PrintOutput(out)
		},
	}
}
