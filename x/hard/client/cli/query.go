package cli

import (
	"context"
	"strings"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/x/hard/types"
)

// flags for cli queries
const (
	flagName  = "name"
	flagDenom = "denom"
	flagOwner = "owner"
)

// GetQueryCmd returns the cli query commands for the  module
func GetQueryCmd() *cobra.Command {
	hardQueryCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Querying commands for the hard module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmds := []*cobra.Command{
		queryParamsCmd(),
		queryModAccountsCmd(),
		queryDepositsCmd(),
		queryUnsyncedDepositsCmd(),
		queryTotalDepositedCmd(),
		queryBorrowsCmd(),
		queryUnsyncedBorrowsCmd(),
		queryTotalBorrowedCmd(),
		queryInterestRateCmd(),
		queryReserves(),
		queryInterestFactorsCmd(),
	}

	for _, cmd := range cmds {
		flags.AddQueryFlagsToCmd(cmd)
	}

	hardQueryCmd.AddCommand(cmds...)

	return hardQueryCmd
}

func queryParamsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "params",
		Short: "get the hard module parameters",
		Long:  "Get the current global hard module parameters.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.Params(context.Background(), &types.QueryParamsRequest{})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}
}

func queryModAccountsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "accounts",
		Short: "query hard module accounts with optional filter",
		Long: strings.TrimSpace(`Query for all hard module accounts or a specific account using the name flag:

		Example:
		$ kvcli q hard accounts
		$ kvcli q hard accounts --name hard|hard_delegator_distribution|hard_lp_distribution`,
		),
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			name, err := cmd.Flags().GetString(flagName)
			if err != nil {
				return err
			}

			req := &types.QueryAccountsRequest{
				Pagination: pageReq,
			}

			if name != "" {
				req.Name = name
			}

			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.Accounts(context.Background(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddPaginationFlagsToCmd(cmd, "accounts")

	return cmd
}

func queryUnsyncedDepositsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unsynced-deposits",
		Short: "query hard module unsynced deposits with optional filters",
		Long: strings.TrimSpace(`query for all hard module unsynced deposits or a specific unsynced deposit using flags:

		Example:
		$ kvcli q hard unsynced-deposits
		$ kvcli q hard unsynced-deposits --owner kava1l0xsq2z7gqd7yly0g40y5836g0appumark77ny --denom bnb
		$ kvcli q hard unsynced-deposits --denom ukava
		$ kvcli q hard unsynced-deposits --denom btcb`,
		),
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			ownerBech, err := cmd.Flags().GetString(flagOwner)
			if err != nil {
				return err
			}
			denom, err := cmd.Flags().GetString(flagDenom)
			if err != nil {
				return err
			}

			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			req := &types.QueryUnsyncedDepositsRequest{
				Denom:      denom,
				Pagination: pageReq,
			}

			if len(ownerBech) != 0 {
				depositOwner, err := sdk.AccAddressFromBech32(ownerBech)
				if err != nil {
					return err
				}
				req.Owner = depositOwner.String()
			}

			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.UnsyncedDeposits(context.Background(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddPaginationFlagsToCmd(cmd, "unsynced-deposits")

	cmd.Flags().String(flagOwner, "", "(optional) filter for unsynced deposits by owner address")
	cmd.Flags().String(flagDenom, "", "(optional) filter for unsynced deposits by denom")

	return cmd
}

func queryDepositsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deposits",
		Short: "query hard module deposits with optional filters",
		Long: strings.TrimSpace(`query for all hard module deposits or a specific deposit using flags:

		Example:
		$ kvcli q hard deposits
		$ kvcli q hard deposits --owner kava1l0xsq2z7gqd7yly0g40y5836g0appumark77ny --denom bnb
		$ kvcli q hard deposits --denom ukava
		$ kvcli q hard deposits --denom btcb`,
		),
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			ownerBech, err := cmd.Flags().GetString(flagOwner)
			if err != nil {
				return err
			}
			denom, err := cmd.Flags().GetString(flagDenom)
			if err != nil {
				return err
			}

			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			req := &types.QueryDepositsRequest{
				Denom:      denom,
				Pagination: pageReq,
			}

			if len(ownerBech) != 0 {
				depositOwner, err := sdk.AccAddressFromBech32(ownerBech)
				if err != nil {
					return err
				}
				req.Owner = depositOwner.String()
			}

			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.Deposits(context.Background(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddPaginationFlagsToCmd(cmd, "deposits")

	cmd.Flags().String(flagOwner, "", "(optional) filter for deposits by owner address")
	cmd.Flags().String(flagDenom, "", "(optional) filter for deposits by denom")

	return cmd
}

func queryUnsyncedBorrowsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unsynced-borrows",
		Short: "query hard module unsynced borrows with optional filters",
		Long: strings.TrimSpace(`query for all hard module unsynced borrows or a specific unsynced borrow using flags:

		Example:
		$ kvcli q hard unsynced-borrows
		$ kvcli q hard unsynced-borrows --owner kava1l0xsq2z7gqd7yly0g40y5836g0appumark77ny
		$ kvcli q hard unsynced-borrows --denom bnb`,
		),
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			ownerBech, err := cmd.Flags().GetString(flagOwner)
			if err != nil {
				return err
			}
			denom, err := cmd.Flags().GetString(flagDenom)
			if err != nil {
				return err
			}

			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			req := &types.QueryUnsyncedBorrowsRequest{
				Denom:      denom,
				Pagination: pageReq,
			}

			if len(ownerBech) != 0 {
				borrowOwner, err := sdk.AccAddressFromBech32(ownerBech)
				if err != nil {
					return err
				}
				req.Owner = borrowOwner.String()
			}

			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.UnsyncedBorrows(context.Background(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddPaginationFlagsToCmd(cmd, "unsynced borrows")

	cmd.Flags().String(flagOwner, "", "(optional) filter for unsynced borrows by owner address")
	cmd.Flags().String(flagDenom, "", "(optional) filter for unsynced borrows by denom")

	return cmd
}

func queryBorrowsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "borrows",
		Short: "query hard module borrows with optional filters",
		Long: strings.TrimSpace(`query for all hard module borrows or a specific borrow using flags:

		Example:
		$ kvcli q hard borrows
		$ kvcli q hard borrows --owner kava1l0xsq2z7gqd7yly0g40y5836g0appumark77ny
		$ kvcli q hard borrows --denom bnb`,
		),
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			ownerBech, err := cmd.Flags().GetString(flagOwner)
			if err != nil {
				return err
			}
			denom, err := cmd.Flags().GetString(flagDenom)
			if err != nil {
				return err
			}

			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			req := &types.QueryBorrowsRequest{
				Denom:      denom,
				Pagination: pageReq,
			}

			if len(ownerBech) != 0 {
				borrowOwner, err := sdk.AccAddressFromBech32(ownerBech)
				if err != nil {
					return err
				}
				req.Owner = borrowOwner.String()
			}

			queryClient := types.NewQueryClient(clientCtx)
			res, err := queryClient.Borrows(context.Background(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddPaginationFlagsToCmd(cmd, "borrows")

	cmd.Flags().String(flagOwner, "", "(optional) filter for borrows by owner address")
	cmd.Flags().String(flagDenom, "", "(optional) filter for borrows by denom")

	return cmd
}

func queryTotalBorrowedCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "total-borrowed",
		Short: "get total current borrowed amount",
		Long: strings.TrimSpace(`get the total amount of coins currently borrowed using flags:

		Example:
		$ kvcli q hard total-borrowed
		$ kvcli q hard total-borrowed --denom bnb`,
		),
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			denom, err := cmd.Flags().GetString(flagDenom)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)
			res, err := queryClient.TotalBorrowed(context.Background(), &types.QueryTotalBorrowedRequest{
				Denom: denom,
			})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	cmd.Flags().String(flagDenom, "", "(optional) filter total borrowed coins by denom")

	return cmd
}

func queryTotalDepositedCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "total-deposited",
		Short: "get total current deposited amount",
		Long: strings.TrimSpace(`get the total amount of coins currently deposited using flags:

		Example:
		$ kvcli q hard total-deposited
		$ kvcli q hard total-deposited --denom bnb`,
		),
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			denom, err := cmd.Flags().GetString(flagDenom)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)
			res, err := queryClient.TotalDeposited(context.Background(), &types.QueryTotalDepositedRequest{
				Denom: denom,
			})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	cmd.Flags().String(flagDenom, "", "(optional) filter total deposited coins by denom")

	return cmd
}

func queryInterestRateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "interest-rate",
		Short: "get current money market interest rates",
		Long: strings.TrimSpace(`get current money market interest rates:

		Example:
		$ kvcli q hard interest-rate
		$ kvcli q hard interest-rate --denom bnb`,
		),
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			denom, err := cmd.Flags().GetString(flagDenom)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)
			res, err := queryClient.InterestRate(context.Background(), &types.QueryInterestRateRequest{
				Denom: denom,
			})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	cmd.Flags().String(flagDenom, "", "(optional) filter interest rates by denom")

	return cmd
}

func queryReserves() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "reserves",
		Short: "get total current Hard module reserves",
		Long: strings.TrimSpace(`get the total amount of coins currently held as reserve by the Hard module:

		Example:
		$ kvcli q hard reserves
		$ kvcli q hard reserves --denom bnb`,
		),
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			denom, err := cmd.Flags().GetString(flagDenom)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)
			res, err := queryClient.Reserves(context.Background(), &types.QueryReservesRequest{
				Denom: denom,
			})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	cmd.Flags().String(flagDenom, "", "(optional) filter reserve coins by denom")

	return cmd
}

func queryInterestFactorsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "interest-factors",
		Short: "get current global interest factors",
		Long: strings.TrimSpace(`get current global interest factors:

		Example:
		$ kvcli q hard interest-factors
		$ kvcli q hard interest-factors --denom bnb`,
		),
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			denom, err := cmd.Flags().GetString(flagDenom)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)
			res, err := queryClient.InterestFactors(context.Background(), &types.QueryInterestFactorsRequest{
				Denom: denom,
			})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	cmd.Flags().String(flagDenom, "", "(optional) filter interest factors by denom")

	return cmd
}
