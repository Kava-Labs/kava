package cli

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/version"

	"github.com/kava-labs/kava/x/committee/client/common"
	"github.com/kava-labs/kava/x/committee/types"
)

// GetQueryCmd returns the cli query commands for this module
func GetQueryCmd(queryRoute string, cdc *codec.Codec) *cobra.Command {
	queryCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("Querying commands for the %s module", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	queryCmd.AddCommand(flags.GetCommands(
		// committees
		GetCmdQueryCommittee(queryRoute, cdc),
		GetCmdQueryCommittees(queryRoute, cdc),
		// proposals
		GetCmdQueryProposal(queryRoute, cdc),
		GetCmdQueryProposals(queryRoute, cdc),
		// votes
		GetCmdQueryVotes(queryRoute, cdc),
		// other
		GetCmdQueryProposer(queryRoute, cdc),
		GetCmdQueryTally(queryRoute, cdc))...)

	return queryCmd
}

// ------------------------------------------
//				Committees
// ------------------------------------------

// GetCmdQueryCommittee implements a query committee command.
func GetCmdQueryCommittee(queryRoute string, cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "committee [committee-id]",
		Args:    cobra.ExactArgs(1),
		Short:   "Query details of a single committee",
		Example: fmt.Sprintf("%s query %s committee 1", version.ClientName, types.ModuleName),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			// Prepare params for querier
			committeeID, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("committee-id %s not a valid uint", args[0])
			}
			bz, err := cdc.MarshalJSON(types.NewQueryCommitteeParams(committeeID))
			if err != nil {
				return err
			}

			// Query
			res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s", queryRoute, types.QueryCommittee), bz)
			if err != nil {
				return err
			}

			// Decode and print result
			committee := types.Committee{}
			if err = cdc.UnmarshalJSON(res, &committee); err != nil {
				return err
			}
			return cliCtx.PrintOutput(committee)
		},
	}
	return cmd
}

// GetCmdQueryCommittees implements a query committees command.
func GetCmdQueryCommittees(queryRoute string, cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "committees",
		Args:    cobra.NoArgs,
		Short:   "Query all committees",
		Example: fmt.Sprintf("%s query %s committees", version.ClientName, types.ModuleName),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			// Query
			res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s", queryRoute, types.QueryCommittees), nil)
			if err != nil {
				return err
			}

			// Decode and print result
			committees := []types.Committee{} // using empty (not nil) slice so json output returns "[]"" instead of "null" when there's no data
			if err = cdc.UnmarshalJSON(res, &committees); err != nil {
				return err
			}
			return cliCtx.PrintOutput(committees)
		},
	}
	return cmd
}

// ------------------------------------------
//				Proposals
// ------------------------------------------

// GetCmdQueryProposal implements the query proposal command.
func GetCmdQueryProposal(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:     "proposal [proposal-id]",
		Args:    cobra.ExactArgs(1),
		Short:   "Query details of a single proposal",
		Example: fmt.Sprintf("%s query %s proposal 2", version.ClientName, types.ModuleName),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			// Prepare params for querier
			proposalID, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("proposal-id %s not a valid uint", args[0])
			}
			bz, err := cdc.MarshalJSON(types.NewQueryProposalParams(proposalID))
			if err != nil {
				return err
			}

			// Query
			res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s", queryRoute, types.QueryProposal), bz)
			if err != nil {
				return err
			}

			// Decode and print results
			var proposal types.Proposal
			cdc.MustUnmarshalJSON(res, &proposal)
			return cliCtx.PrintOutput(proposal)
		},
	}
}

// GetCmdQueryProposals implements a query proposals command.
func GetCmdQueryProposals(queryRoute string, cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "proposals [committee-id]",
		Short:   "Query all proposals for a committee",
		Args:    cobra.ExactArgs(1),
		Example: fmt.Sprintf("%s query %s proposals 1", version.ClientName, types.ModuleName),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			// Prepare params for querier
			committeeID, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("committee-id %s not a valid uint", args[0])
			}
			bz, err := cdc.MarshalJSON(types.NewQueryCommitteeParams(committeeID))
			if err != nil {
				return err
			}

			// Query
			res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s", queryRoute, types.QueryProposals), bz)
			if err != nil {
				return err
			}

			// Decode and print results
			proposals := []types.Proposal{}
			err = cdc.UnmarshalJSON(res, &proposals)
			if err != nil {
				return err
			}
			return cliCtx.PrintOutput(proposals)
		},
	}
	return cmd
}

// ------------------------------------------
//				Votes
// ------------------------------------------

// GetCmdQueryVotes implements the command to query for proposal votes.
func GetCmdQueryVotes(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:     "votes [proposal-id]",
		Args:    cobra.ExactArgs(1),
		Short:   "Query votes on a proposal",
		Example: fmt.Sprintf("%s query %s votes 2", version.ClientName, types.ModuleName),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			// Prepare params for querier
			proposalID, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("proposal-id %s not a valid int", args[0])
			}
			bz, err := cdc.MarshalJSON(types.NewQueryProposalParams(proposalID))
			if err != nil {
				return err
			}

			// Query
			res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s", queryRoute, types.QueryVotes), bz)
			if err != nil {
				return err
			}

			// Decode and print results
			votes := []types.Vote{} // using empty (not nil) slice so json returns [] instead of null when there's no data
			err = cdc.UnmarshalJSON(res, &votes)
			if err != nil {
				return err
			}
			return cliCtx.PrintOutput(votes)
		},
	}
}

// ------------------------------------------
//				Other
// ------------------------------------------

func GetCmdQueryTally(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:     "tally [proposal-id]",
		Args:    cobra.ExactArgs(1),
		Short:   "Get the current tally of votes on a proposal",
		Long:    "Query the current tally of votes on a proposal to see the progress of the voting.",
		Example: fmt.Sprintf("%s query %s tally 2", version.ClientName, types.ModuleName),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			// Prepare params for querier
			proposalID, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("proposal-id %s not a valid int", args[0])
			}
			bz, err := cdc.MarshalJSON(types.NewQueryProposalParams(proposalID))
			if err != nil {
				return err
			}

			// Query
			res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/tally", queryRoute), bz)
			if err != nil {
				return err
			}

			// Decode and print results
			var tally bool
			if err = cdc.UnmarshalJSON(res, &tally); err != nil {
				return err
			}
			return cliCtx.PrintOutput(tally)
		},
	}
}

func GetCmdQueryProposer(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:     "proposer [proposal-id]",
		Args:    cobra.ExactArgs(1),
		Short:   "Query the proposer of a governance proposal",
		Long:    "Query which address proposed a proposal with a given ID.",
		Example: fmt.Sprintf("%s query %s proposer 2", version.ClientName, types.ModuleName),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			// validate that the proposalID is a uint
			proposalID, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("proposal-id %s is not a valid uint", args[0])
			}

			prop, err := common.QueryProposer(cliCtx, proposalID)
			if err != nil {
				return err
			}

			return cliCtx.PrintOutput(prop)
		},
	}
}
