package cli

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"

	"github.com/kava-labs/kava/x/incentive/keeper"
	"github.com/kava-labs/kava/x/incentive/types"
)

const (
	flagOwner    = "owner"
	flagType     = "type"
	flagUnsynced = "unsynced"
	flagDenom    = "denom"
)

var rewardTypes = []string{
	keeper.RewardTypeHard,
	keeper.RewardTypeUSDXMinting,
	keeper.RewardTypeDelegator,
	keeper.RewardTypeSwap,
	keeper.RewardTypeSavings,
	keeper.RewardTypeEarn,
}

// GetQueryCmd returns the cli query commands for the incentive module
func GetQueryCmd() *cobra.Command {
	incentiveQueryCmd := &cobra.Command{
		Use:   types.ModuleName,
		Short: "Querying commands for the incentive module",
	}

	cmds := []*cobra.Command{
		queryParamsCmd(),
		queryRewardsCmd(),
		queryRewardFactorsCmd(),
		queryApyCmd(),
	}

	for _, cmd := range cmds {
		flags.AddQueryFlagsToCmd(cmd)
	}

	incentiveQueryCmd.AddCommand(cmds...)

	return incentiveQueryCmd
}

func queryRewardsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rewards",
		Short: "query claimable rewards",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query rewards with optional flags for owner and type

			Example:
			$ %[1]s query %[2]s rewards
			$ %[1]s query %[2]s rewards --owner kava15qdefkmwswysgg4qxgqpqr35k3m49pkx2jdfnw
			$ %[1]s query %[2]s rewards --type hard
			$ %[1]s query %[2]s rewards --type usdx-minting
			$ %[1]s query %[2]s rewards --type delegator
			$ %[1]s query %[2]s rewards --type swap
			$ %[1]s query %[2]s rewards --type savings
			$ %[1]s query %[2]s rewards --type earn
			$ %[1]s query %[2]s rewards --type hard --owner kava15qdefkmwswysgg4qxgqpqr35k3m49pkx2jdfnw
			$ %[1]s query %[2]s rewards --type hard --unsynced
			`,
				version.AppName, types.ModuleName)),
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			strOwner, _ := cmd.Flags().GetString(flagOwner)
			strType, _ := cmd.Flags().GetString(flagType)
			boolUnsynced, _ := cmd.Flags().GetBool(flagUnsynced)

			// Prepare params for querier
			var owner sdk.AccAddress
			if strOwner != "" {
				if owner, err = sdk.AccAddressFromBech32(strOwner); err != nil {
					return err
				}
			}

			rewardType := strings.ToLower(strType)
			queryClient := types.NewQueryClient(cliCtx)
			request := types.QueryRewardsRequest{
				RewardType:     rewardType,
				Owner:          owner.String(),
				Unsynchronized: boolUnsynced,
			}
			rewards, err := queryClient.Rewards(context.Background(), &request)
			if err != nil {
				return err
			}
			return cliCtx.PrintProto(rewards)
		},
	}
	cmd.Flags().String(flagOwner, "", "(optional) filter by owner address")
	cmd.Flags().String(flagType, "", fmt.Sprintf("(optional) filter by a reward type: %s", strings.Join(rewardTypes, "|")))
	cmd.Flags().Bool(flagUnsynced, false, "(optional) get unsynced claims")
	cmd.Flags().Int(flags.FlagPage, 1, "pagination page rewards of to query for")
	cmd.Flags().Int(flags.FlagLimit, 100, "pagination limit of rewards to query for")
	return cmd
}

func queryParamsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "params",
		Short: "get the incentive module parameters",
		Long:  "Get the current global incentive module parameters.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(cliCtx)
			res, err := queryClient.Params(context.Background(), &types.QueryParamsRequest{})
			if err != nil {
				return err
			}
			return cliCtx.PrintProto(res)
		},
	}
}

func queryRewardFactorsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "reward-factors",
		Short: "get current global reward factors",
		Long:  `Get current global reward factors for all reward types.`,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(cliCtx)
			res, err := queryClient.RewardFactors(context.Background(), &types.QueryRewardFactorsRequest{})
			if err != nil {
				return err
			}
			return cliCtx.PrintProto(res)
		},
	}
	cmd.Flags().String(flagDenom, "", "(optional) filter reward factors by denom")
	return cmd
}

func queryApyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "apy",
		Short: "queries incentive reward apy for a reward",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(cliCtx)
			res, err := queryClient.Apy(context.Background(), &types.QueryApyRequest{})
			if err != nil {
				return err
			}
			return cliCtx.PrintProto(res)
		},
	}
	return cmd
}
