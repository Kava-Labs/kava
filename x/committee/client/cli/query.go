package cli

import (
	"context"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/kava-labs/kava/x/committee/client/common"
	"github.com/kava-labs/kava/x/committee/types"
)

// GetQueryCmd returns the cli query commands for this module
func GetQueryCmd() *cobra.Command {
	queryCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("Querying commands for the %s module", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmds := []*cobra.Command{
		// committees
		getCmdQueryCommittee(),
		getCmdQueryCommittees(),
		// proposals
		getCmdQueryNextProposalID(),
		getCmdQueryProposal(),
		getCmdQueryProposals(),
		// votes
		getCmdQueryVotes(),
		// other
		getCmdQueryProposer(),
		getCmdQueryTally(),
		getCmdQueryRawParams(),
	}

	for _, cmd := range cmds {
		flags.AddQueryFlagsToCmd(cmd)
	}

	queryCmd.AddCommand(cmds...)

	return queryCmd
}

// ------------------------------------------
//				Committees
// ------------------------------------------

// getCmdQueryCommittee implements a query committee command.
func getCmdQueryCommittee() *cobra.Command {
	return &cobra.Command{
		Use:     "committee [committee-id]",
		Args:    cobra.ExactArgs(1),
		Short:   "Query details of a single committee",
		Example: fmt.Sprintf("%s query %s committee 1", version.AppName, types.ModuleName),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			// validate that the committee id is a uint
			committeeID, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("committee-id %s not a valid uint, please input a valid committee-id", args[0])
			}

			queryClient := types.NewQueryClient(clientCtx)
			res, err := queryClient.Committee(context.Background(), &types.QueryCommitteeRequest{CommitteeId: committeeID})
			if err != nil {
				return err
			}
			return clientCtx.PrintProto(res)
		},
	}
}

// getCmdQueryCommittees implements a query committees command.
func getCmdQueryCommittees() *cobra.Command {
	return &cobra.Command{
		Use:     "committees",
		Args:    cobra.NoArgs,
		Short:   "Query all committees",
		Example: fmt.Sprintf("%s query %s committees", version.AppName, types.ModuleName),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)
			res, err := queryClient.Committees(context.Background(), &types.QueryCommitteesRequest{})
			if err != nil {
				return err
			}
			return clientCtx.PrintProto(res)
		},
	}
}

// ------------------------------------------
//				Proposals
// ------------------------------------------

// getCmdQueryNextProposalID implements a query next proposal ID command.
func getCmdQueryNextProposalID() *cobra.Command {
	return &cobra.Command{
		Use:     "next-proposal-id",
		Short:   "Query the next proposal ID",
		Args:    cobra.ExactArgs(0),
		Example: fmt.Sprintf("%s query %s next-proposal-id", version.AppName, types.ModuleName),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)
			res, err := queryClient.NextProposalID(context.Background(), &types.QueryNextProposalIDRequest{})
			if err != nil {
				return err
			}
			return clientCtx.PrintProto(res)
		},
	}
}

// getCmdQueryProposal implements the query proposal command.
func getCmdQueryProposal() *cobra.Command {
	return &cobra.Command{
		Use:     "proposal [proposal-id]",
		Args:    cobra.ExactArgs(1),
		Short:   "Query details of a single proposal",
		Example: fmt.Sprintf("%s query %s proposal 2", version.AppName, types.ModuleName),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			// Prepare params for querier
			proposalID, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("proposal-id %s not a valid uint", args[0])
			}

			queryClient := types.NewQueryClient(clientCtx)
			res, err := queryClient.Proposal(context.Background(), &types.QueryProposalRequest{
				ProposalId: proposalID,
			})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}
}

// getCmdQueryProposals implements a query proposals command.
func getCmdQueryProposals() *cobra.Command {
	return &cobra.Command{
		Use:     "proposals [committee-id]",
		Short:   "Query all proposals for a committee",
		Args:    cobra.ExactArgs(1),
		Example: fmt.Sprintf("%s query %s proposals 1", version.AppName, types.ModuleName),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			// Prepare params for querier
			committeeID, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("committee-id %s not a valid uint", args[0])
			}

			queryClient := types.NewQueryClient(clientCtx)
			res, err := queryClient.Proposals(context.Background(), &types.QueryProposalsRequest{
				CommitteeId: committeeID,
			})
			if err != nil {
				return err
			}
			return clientCtx.PrintProto(res)
		},
	}
}

// ------------------------------------------
//				Votes
// ------------------------------------------

// getCmdQueryVotes implements the command to query for proposal votes.
func getCmdQueryVotes() *cobra.Command {
	return &cobra.Command{
		Use:     "votes [proposal-id]",
		Args:    cobra.ExactArgs(1),
		Short:   "Query votes on a proposal",
		Example: fmt.Sprintf("%s query %s votes 2", version.AppName, types.ModuleName),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			// Prepare params for querier
			proposalID, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("proposal-id %s not a valid int", args[0])
			}

			queryClient := types.NewQueryClient(clientCtx)
			res, err := queryClient.Votes(context.Background(), &types.QueryVotesRequest{
				ProposalId: proposalID,
			})
			if err != nil {
				return err
			}
			return clientCtx.PrintProto(res)
		},
	}
}

// ------------------------------------------
//				Other
// ------------------------------------------

func getCmdQueryTally() *cobra.Command {
	return &cobra.Command{
		Use:     "tally [proposal-id]",
		Args:    cobra.ExactArgs(1),
		Short:   "Get the current tally of votes on a proposal",
		Long:    "Query the current tally of votes on a proposal to see the progress of the voting.",
		Example: fmt.Sprintf("%s query %s tally 2", version.AppName, types.ModuleName),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			// Prepare params for querier
			proposalID, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("proposal-id %s not a valid int", args[0])
			}

			queryClient := types.NewQueryClient(clientCtx)
			res, err := queryClient.Tally(context.Background(), &types.QueryTallyRequest{
				ProposalId: proposalID,
			})
			if err != nil {
				return err
			}
			return clientCtx.PrintProto(res)
		},
	}
}

func getCmdQueryProposer() *cobra.Command {
	return &cobra.Command{
		Use:     "proposer [proposal-id]",
		Args:    cobra.ExactArgs(1),
		Short:   "Query the proposer of a governance proposal",
		Long:    "Query which address proposed a proposal with a given ID.",
		Example: fmt.Sprintf("%s query %s proposer 2", version.AppName, types.ModuleName),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			// validate that the proposalID is a uint
			proposalID, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("proposal-id %s is not a valid uint", args[0])
			}
			prop, err := common.QueryProposer(clientCtx, proposalID)
			if err != nil {
				return err
			}

			return clientCtx.PrintObjectLegacy(prop)
		},
	}
}

func getCmdQueryRawParams() *cobra.Command {
	return &cobra.Command{
		Use:     "raw-params [subspace] [key]",
		Args:    cobra.ExactArgs(2),
		Short:   "Query raw parameter values from any module.",
		Long:    "Query the byte value of any module's parameters. Useful in debugging and verifying governance proposals.",
		Example: fmt.Sprintf("%s query %s raw-params cdp CollateralParams", version.AppName, types.ModuleName),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)
			res, err := queryClient.RawParams(context.Background(), &types.QueryRawParamsRequest{
				Subspace: args[0],
				Key:      args[1],
			})
			if err != nil {
				return err
			}
			return clientCtx.PrintProto(res)
		},
	}
}
