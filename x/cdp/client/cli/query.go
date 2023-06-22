package cli

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"

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
func GetQueryCmd() *cobra.Command {
	// Group nameservice queries under a subcommand
	cdpQueryCmd := &cobra.Command{
		Use:   "cdp",
		Short: "Querying commands for the cdp module",
	}

	cmds := []*cobra.Command{
		QueryCdpCmd(),
		QueryGetCdpsCmd(),
		QueryCdpDepositsCmd(),
		QueryParamsCmd(),
		QueryGetAccounts(),
	}

	for _, cmd := range cmds {
		flags.AddQueryFlagsToCmd(cmd)
	}

	cdpQueryCmd.AddCommand(cmds...)

	return cdpQueryCmd
}

// QueryCdpCmd returns the command handler for querying a particular cdp
func QueryCdpCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "cdp [owner-addr] [collateral-type]",
		Short: "get info about a cdp",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Get a CDP by the owner address and the collateral name.

Example:
$ %s query %s cdp kava15qdefkmwswysgg4qxgqpqr35k3m49pkx2jdfnw atom-a
`, version.AppName, types.ModuleName)),
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			_, err = sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			res, err := queryClient.Cdp(context.Background(), &types.QueryCdpRequest{
				Owner:          args[0],
				CollateralType: args[1],
			})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}
}

// QueryGetCdpsCmd queries the cdps in the store
func QueryGetCdpsCmd() *cobra.Command {
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
			strCollateralType, err := cmd.Flags().GetString(flagCollateralType)
			if err != nil {
				return err
			}
			strOwner, err := cmd.Flags().GetString(flagOwner)
			if err != nil {
				return err
			}
			strID, err := cmd.Flags().GetString(flagID)
			if err != nil {
				return err
			}
			strRatio, err := cmd.Flags().GetString(flagRatio)
			if err != nil {
				return err
			}

			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			req := types.QueryCdpsRequest{
				Pagination: pageReq,
			}

			if len(strCollateralType) != 0 {
				req.CollateralType = strings.ToLower(strings.TrimSpace(strCollateralType))
			}

			if len(strOwner) != 0 {
				cdpOwner, err := sdk.AccAddressFromBech32(strings.ToLower(strings.TrimSpace(strOwner)))
				if err != nil {
					return fmt.Errorf("cannot parse address from cdp owner %s", strOwner)
				}
				req.Owner = cdpOwner.String()
			}

			if len(strID) != 0 {
				cdpID, err := strconv.ParseUint(strID, 10, 64)
				if err != nil {
					return fmt.Errorf("cannot parse cdp ID %s", strID)
				}
				req.ID = cdpID
			}

			if len(strRatio) != 0 {
				cdpRatio, err := sdk.NewDecFromStr(strRatio)
				if err != nil {
					return fmt.Errorf("cannot parse cdp ratio %s", strRatio)
				}
				// ratio is also validated on server
				req.Ratio = cdpRatio.String()
			}

			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.Cdps(context.Background(), &req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	cmd.Flags().String(flagCollateralType, "", "(optional) filter by CDP collateral type")
	cmd.Flags().String(flagOwner, "", "(optional) filter by CDP owner")
	cmd.Flags().String(flagID, "", "(optional) filter by CDP ID")
	cmd.Flags().String(flagRatio, "", "(optional) filter by CDP collateralization ratio threshold")

	flags.AddPaginationFlagsToCmd(cmd, "cdps")

	return cmd
}

// QueryCdpDepositsCmd returns the command handler for querying the deposits of a particular cdp
func QueryCdpDepositsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "deposits [owner-addr] [collateral-type]",
		Short: "get deposits for a cdp",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Get the deposits of a CDP.

Example:
$ %s query %s deposits kava15qdefkmwswysgg4qxgqpqr35k3m49pkx2jdfnw atom-a
`, version.AppName, types.ModuleName)),
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			_, err = sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			res, err := queryClient.Deposits(context.Background(), &types.QueryDepositsRequest{
				Owner:          args[0],
				CollateralType: args[1],
			})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}
}

// QueryParamsCmd returns the command handler for cdp parameter querying
func QueryParamsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "params",
		Short: "get the cdp module parameters",
		Long:  "get the current global cdp module parameters.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.Params(context.Background(), &types.QueryParamsRequest{})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(&res.Params)
		},
	}
}

// QueryGetAccounts queries CDP module accounts
func QueryGetAccounts() *cobra.Command {
	return &cobra.Command{
		Use:   "accounts",
		Short: "get module accounts",
		Long:  "get cdp module account addresses",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.Accounts(context.Background(), &types.QueryAccountsRequest{})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}
}
