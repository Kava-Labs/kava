package cli

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"

	"github.com/kava-labs/kava/x/committee/client/common"
	"github.com/kava-labs/kava/x/committee/types"
)

// GetQueryCmd returns the cli query commands for this module
func GetQueryCmd(queryRoute string, cdc *codec.Codec) *cobra.Command {
	// Group gov queries under a subcommand
	govQueryCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Querying commands for the governance module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	govQueryCmd.AddCommand(client.GetCommands(
		//GetCmdQueryCommittee(queryRoute, cdc),
		GetCmdQueryCommittees(queryRoute, cdc),
		GetCmdQueryProposal(queryRoute, cdc),
		GetCmdQueryProposals(queryRoute, cdc),
		//GetCmdQueryVote(queryRoute, cdc),
		GetCmdQueryVotes(queryRoute, cdc),
		//GetCmdQueryParams(queryRoute, cdc),
		GetCmdQueryProposer(queryRoute, cdc),
		GetCmdQueryTally(queryRoute, cdc))...)

	return govQueryCmd
}

// GetCmdQueryProposals implements a query proposals command.
func GetCmdQueryCommittees(queryRoute string, cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "committees",
		Short: "Query all committees",
		Long:  "", // TODO
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			// Query
			res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s", queryRoute, types.QueryCommittees), nil)
			if err != nil {
				return err
			}

			// Decode and print result
			committees := []types.Committee{}
			if err = cdc.UnmarshalJSON(res, &committees); err != nil {
				return err
			}
			return cliCtx.PrintOutput(committees)
		},
	}
	return cmd
}

// GetCmdQueryProposal implements the query proposal command.
func GetCmdQueryProposal(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "proposal [proposal-id]",
		Args:  cobra.ExactArgs(1),
		Short: "Query details of a single proposal",
		// 		Long: strings.TrimSpace(
		// 			fmt.Sprintf(`Query details for a proposal. You can find the
		// proposal-id by running "%s query gov proposals".

		// Example:
		// $ %s query gov proposal 1
		// `,
		// 				version.ClientName, version.ClientName,
		// 			),
		// 		),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			// Prepare params for querier
			proposalID, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("proposal-id %s not a valid uint", args[0])
			}
			bz, err := cdc.MarshalJSON(types.NewQueryCommitteeParams(proposalID))
			if err != nil {
				return err
			}

			// Query
			//res, err := gcutils.QueryProposalByID(proposalID, cliCtx, queryRoute)
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
		Use:   "proposals [committee-id]",
		Short: "Query proposals by committee.",
		Args:  cobra.ExactArgs(1),
		// 		Long: strings.TrimSpace(
		// 			fmt.Sprintf(`Query for a all proposals. You can filter the returns with the following flags.

		// Example:
		// $ %s query gov proposals --depositor cosmos1skjwj5whet0lpe65qaq4rpq03hjxlwd9nf39lk
		// $ %s query gov proposals --voter cosmos1skjwj5whet0lpe65qaq4rpq03hjxlwd9nf39lk
		// $ %s query gov proposals --status (DepositPeriod|VotingPeriod|Passed|Rejected)
		// `,
		// 				version.ClientName, version.ClientName, version.ClientName,
		// 			),
		// 		),
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
			res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/proposals", queryRoute), bz)
			if err != nil {
				return err
			}

			// Decode and print results
			proposals := []types.Proposal{} // using empty (not nil) slice so json returns [] instead of null when there's no data // TODO check
			err = cdc.UnmarshalJSON(res, &proposals)
			if err != nil {
				return err
			}
			return cliCtx.PrintOutput(proposals)
		},
	}
	return cmd
}

// // Command to Get a Proposal Information
// // GetCmdQueryVote implements the query proposal vote command.
// func GetCmdQueryVote(queryRoute string, cdc *codec.Codec) *cobra.Command {
// 	return &cobra.Command{
// 		Use:   "vote [proposal-id] [voter-addr]",
// 		Args:  cobra.ExactArgs(2),
// 		Short: "Query details of a single vote",
// 		Long: strings.TrimSpace(
// 			fmt.Sprintf(`Query details for a single vote on a proposal given its identifier.

// Example:
// $ %s query gov vote 1 cosmos1skjwj5whet0lpe65qaq4rpq03hjxlwd9nf39lk
// `,
// 				version.ClientName,
// 			),
// 		),
// 		RunE: func(cmd *cobra.Command, args []string) error {
// 			cliCtx := context.NewCLIContext().WithCodec(cdc)

// 			// validate that the proposal id is a uint
// 			proposalID, err := strconv.ParseUint(args[0], 10, 64)
// 			if err != nil {
// 				return fmt.Errorf("proposal-id %s not a valid int, please input a valid proposal-id", args[0])
// 			}

// 			// check to see if the proposal is in the store
// 			_, err = gcutils.QueryProposalByID(proposalID, cliCtx, queryRoute)
// 			if err != nil {
// 				return fmt.Errorf("failed to fetch proposal-id %d: %s", proposalID, err)
// 			}

// 			voterAddr, err := sdk.AccAddressFromBech32(args[1])
// 			if err != nil {
// 				return err
// 			}

// 			params := types.NewQueryVoteParams(proposalID, voterAddr)
// 			bz, err := cdc.MarshalJSON(params)
// 			if err != nil {
// 				return err
// 			}

// 			res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/vote", queryRoute), bz)
// 			if err != nil {
// 				return err
// 			}

// 			var vote types.Vote

// 			// XXX: Allow the decoding to potentially fail as the vote may have been
// 			// pruned from state. If so, decoding will fail and so we need to check the
// 			// Empty() case. Consider updating Vote JSON decoding to not fail when empty.
// 			_ = cdc.UnmarshalJSON(res, &vote)

// 			if vote.Empty() {
// 				res, err = gcutils.QueryVoteByTxQuery(cliCtx, params)
// 				if err != nil {
// 					return err
// 				}

// 				if err := cdc.UnmarshalJSON(res, &vote); err != nil {
// 					return err
// 				}
// 			}

// 			return cliCtx.PrintOutput(vote)
// 		},
// 	}
// }

// GetCmdQueryVotes implements the command to query for proposal votes.
func GetCmdQueryVotes(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "votes [proposal-id]",
		Args:  cobra.ExactArgs(1),
		Short: "Query votes on a proposal",
		// 		Long: strings.TrimSpace(
		// 			fmt.Sprintf(`Query vote details for a single proposal by its identifier.

		// Example:
		// $ %s query gov votes 1
		// `,
		// 				version.ClientName,
		// 			),
		// 		),
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
			votes := []types.Vote{} // using empty (not nil) slice so json returns [] instead of null when there's no data // TODO check
			err = cdc.UnmarshalJSON(res, &votes)
			if err != nil {
				return err
			}
			return cliCtx.PrintOutput(votes)
		},
	}
}

// GetCmdQueryTally implements the command to query for proposal tally result.
func GetCmdQueryTally(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "tally [proposal-id]",
		Args:  cobra.ExactArgs(1),
		Short: "Get the tally of a proposal vote",
		// 		Long: strings.TrimSpace(
		// 			fmt.Sprintf(`Query tally of votes on a proposal. You can find
		// the proposal-id by running "%s query gov proposals".

		// Example:
		// $ %s query gov tally 1
		// `,
		// 				version.ClientName, version.ClientName,
		// 			),
		// 		),
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
			cdc.MustUnmarshalJSON(res, &tally) // TODO must or normal, what's the difference on the cli?
			return cliCtx.PrintOutput(tally)
		},
	}
}

// // GetCmdQueryProposal implements the query proposal command.
// func GetCmdQueryParams(queryRoute string, cdc *codec.Codec) *cobra.Command {
// 	return &cobra.Command{
// 		Use:   "params",
// 		Short: "Query the parameters of the governance process",
// 		Long: strings.TrimSpace(
// 			fmt.Sprintf(`Query the all the parameters for the governance process.

// Example:
// $ %s query gov params
// `,
// 				version.ClientName,
// 			),
// 		),
// 		Args: cobra.NoArgs,
// 		RunE: func(cmd *cobra.Command, args []string) error {
// 			cliCtx := context.NewCLIContext().WithCodec(cdc)
// 			tp, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/params/tallying", queryRoute), nil)
// 			if err != nil {
// 				return err
// 			}
// 			dp, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/params/deposit", queryRoute), nil)
// 			if err != nil {
// 				return err
// 			}
// 			vp, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/params/voting", queryRoute), nil)
// 			if err != nil {
// 				return err
// 			}

// 			var tallyParams types.TallyParams
// 			cdc.MustUnmarshalJSON(tp, &tallyParams)
// 			var depositParams types.DepositParams
// 			cdc.MustUnmarshalJSON(dp, &depositParams)
// 			var votingParams types.VotingParams
// 			cdc.MustUnmarshalJSON(vp, &votingParams)

// 			return cliCtx.PrintOutput(types.NewParams(votingParams, tallyParams, depositParams))
// 		},
// 	}
// }

// GetCmdQueryProposer implements the query proposer command.
func GetCmdQueryProposer(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "proposer [proposal-id]",
		Args:  cobra.ExactArgs(1),
		Short: "Query the proposer of a governance proposal",
		// 		Long: strings.TrimSpace(
		// 			fmt.Sprintf(`Query which address proposed a proposal with a given ID.

		// Example:
		// $ %s query gov proposer 1
		// `,
		// 				version.ClientName,
		// 			),
		// 		),
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
