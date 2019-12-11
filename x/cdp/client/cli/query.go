package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

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
		GetCmdGetCdp(queryRoute, cdc),
		GetCmdGetCdps(queryRoute, cdc),
		GetCmdGetCdpsByCollateralizationRatio(queryRoute, cdc),
		GetCmdGetParams(queryRoute, cdc),
	)...)

	return cdpQueryCmd
}

// GetCmdGetCdp queries the latest info about a particular cdp
func GetCmdGetCdp(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "cdp [ownerAddress] [collateralType]",
		Short: "get info about a cdp",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			// Prepare params for querier
			ownerAddress, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}
			collateralType := args[1] // TODO validation?
			bz, err := cdc.MarshalJSON(types.QueryCdpsParams{
				CollateralDenom: collateralType,
			})
			if err != nil {
				return err
			}

			// Query
			route := fmt.Sprintf("custom/%s/%s", queryRoute, types.QueryGetCdp)
			res, _, err := cliCtx.QueryWithData(route, bz)
			if err != nil {
				fmt.Printf("error when getting cdp info - %s", err)
				fmt.Printf("could not get current cdp info - %s %s \n", string(ownerAddress), string(collateralType))
				return err
			}

			// Decode and print results
			var cdp types.CDP
			cdc.MustUnmarshalJSON(res, &cdp)
			return cliCtx.PrintOutput(cdp)
		},
	}
}

// GetCmdGetCdps queries the store for all cdps for a collateral type
func GetCmdGetCdps(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "cdps [collateralType]",
		Short: "get info about many cdps",
		Long:  "Get all CDPs. Specify a collateral type to get only CDPs with that collateral type.",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			// Prepare params for querier
			bz, err := cdc.MarshalJSON(types.QueryCdpsParams{CollateralDenom: args[0]}) // denom="" returns all CDPs // TODO will this fail if there are no args?
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

func GetCmdGetCdpsByCollateralizationRatio(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "bad-cdps [collateralType] [price]",
		Short: "get under collateralized CDPs",
		Long:  "Get all CDPS of a particular collateral type that will be under collateralized at the specified price. Pass in the current price to get currently under collateralized CDPs.",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			// Prepare params for querier
			price, errSdk := sdk.NewDecFromStr(args[1])
			if errSdk != nil {
				return fmt.Errorf(errSdk.Error()) // TODO check this returns useful output
			}
			bz, err := cdc.MarshalJSON(types.QueryCdpsByRatioParams{
				CollateralDenom: args[0],
				Ratio:           price,
			})
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

func GetCmdGetParams(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "params",
		Short: "get the cdp module parameters",
		Long:  "Get the current global cdp module parameters.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			// Query
			route := fmt.Sprintf("custom/%s/%s", queryRoute, types.QueryGetParams)
			res, _, err := cliCtx.QueryWithData(route, nil) // TODO use cliCtx.QueryStore?
			if err != nil {
				return err
			}

			// Decode and print results
			var out types.QueryCdpParams
			cdc.MustUnmarshalJSON(res, &out)
			return cliCtx.PrintOutput(out)
		},
	}
}
