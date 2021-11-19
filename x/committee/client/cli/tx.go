package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/tendermint/tendermint/crypto"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	paramsproposal "github.com/cosmos/cosmos-sdk/x/params/types/proposal"

	"github.com/kava-labs/kava/x/committee/types"
)

const PARAMS_CHANGE_PROPOSAL_EXAMPLE = `
{
  "title": "title",
  "description": "description",
  "changes": [{ "subspace": "subspace", "key": "key", "value": "value" }]
}
`

const COMMITTEE_CHANGE_PROPOSAL_EXAMPLE = `
{
  "title": "A Title",
  "description": "A proposal description.",
  "new_committee": {
    "@type": "/kava.committee.v1beta1.MemberCommittee",
    "base_committee": {
      "id": "34",
      "description": "member committee",
      "members": ["kava1ze7y9qwdddejmy7jlw4cymqqlt2wh05yhwmrv2"],
      "permissions": [],
      "vote_threshold": "1.000000000000000000",
      "proposal_duration": "86400s",
      "tally_option": "TALLY_OPTION_DEADLINE"
    }
  }
}
`

const COMMITTEE_DELETE_PROPOSAL_EXAMPLE = `
{
  "title": "A Title",
  "description": "A proposal description.",
  "committee_id": "1"
}
`

func GetTxCmd() *cobra.Command {
	txCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "committee governance transactions subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmds := []*cobra.Command{
		getCmdVote(),
		getCmdSubmitProposal(),
	}

	for _, cmd := range cmds {
		flags.AddTxFlagsToCmd(cmd)
	}

	txCmd.AddCommand(cmds...)

	return txCmd
}

// getCmdSubmitProposal returns the command to submit a proposal to a committee
func getCmdSubmitProposal() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "submit-proposal [committee-id] [proposal-file]",
		Short: "Submit a governance proposal to a particular committee",
		Long: fmt.Sprintf(`Submit a proposal to a committee so they can vote on it.

The proposal file must be the json encoded forms of the proposal type you want to submit.
For example:
%s
`, PARAMS_CHANGE_PROPOSAL_EXAMPLE),
		Args:    cobra.ExactArgs(2),
		Example: fmt.Sprintf("%s tx %s submit-proposal 1 your-proposal.json", version.AppName, types.ModuleName),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			// Get proposing address
			proposer := clientCtx.GetFromAddress()

			// Get committee ID
			committeeID, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("committee-id %s not a valid int", args[0])
			}

			// Get proposal content
			var pubProposal types.PubProposal
			contents, err := ioutil.ReadFile(args[1])
			if err != nil {
				return err
			}
			if err := clientCtx.Codec.UnmarshalInterface(contents, &pubProposal); err != nil {
				return err
			}
			if err = pubProposal.ValidateBasic(); err != nil {
				return err
			}

			// Build message and run basic validation
			msg, err := types.NewMsgSubmitProposal(pubProposal, proposer, committeeID)
			if err != nil {
				return err
			}
			err = msg.ValidateBasic()
			if err != nil {
				return err
			}

			// Sign and broadcast message
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	return cmd
}

// getCmdVote returns the command to vote on a proposal.
func getCmdVote() *cobra.Command {
	return &cobra.Command{
		Use:     "vote [proposal-id] [vote]",
		Args:    cobra.ExactArgs(2),
		Short:   "Vote for an active proposal",
		Long:    "Submit a [yes/no/abstain] vote for the proposal with id [proposal-id].",
		Example: fmt.Sprintf("%s tx %s vote 2 yes", version.AppName, types.ModuleName),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			// Get voting address
			from := clientCtx.GetFromAddress()

			// validate that the proposal id is a uint
			proposalID, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("proposal-id %s not a valid int, please input a valid proposal-id", args[0])
			}

			rawVote := strings.ToLower(strings.TrimSpace(args[1]))
			if len(rawVote) == 0 {
				return fmt.Errorf("must specify a vote")
			}

			var vote types.VoteType
			switch rawVote {
			case "yes", "y":
				vote = types.VOTE_TYPE_YES
			case "no", "n":
				vote = types.VOTE_TYPE_NO
			case "abstain", "a":
				vote = types.VOTE_TYPE_ABSTAIN
			default:
				return fmt.Errorf("must specify a valid vote type: (yes/y, no/n, abstain/a)")
			}

			// Build vote message and run basic validation
			msg := types.NewMsgVote(from, proposalID, vote)
			err = msg.ValidateBasic()
			if err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
}

// GetGovCmdSubmitProposal returns a command to submit a proposal to the gov module. It is passed to the gov module for use on its command subtree.
func GetGovCmdSubmitProposal() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "committee [proposal-file] [deposit]",
		Short: "Submit a governance proposal to change a committee.",
		Long: fmt.Sprintf(`Submit a governance proposal to create, alter, or delete a committee.

The proposal file must be the json encoded form of the proposal type you want to submit.
For example, to create or update a committee:
%s

and to delete a committee:
%s
`, COMMITTEE_CHANGE_PROPOSAL_EXAMPLE, COMMITTEE_DELETE_PROPOSAL_EXAMPLE),
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			// Get proposing address
			proposer := clientCtx.GetFromAddress()

			// Get the deposit
			deposit, err := sdk.ParseCoinsNormalized(args[1])
			if err != nil {
				return err
			}

			// Get the proposal
			bz, err := ioutil.ReadFile(args[0])
			if err != nil {
				return err
			}
			var content govtypes.Content
			if err := clientCtx.Codec.UnmarshalInterface(bz, &content); err != nil {
				return err
			}
			if err = content.ValidateBasic(); err != nil {
				return err
			}

			// Build message and run basic validation
			msg, err := govtypes.NewMsgSubmitProposal(content, deposit, proposer)
			if err != nil {
				return err
			}
			err = msg.ValidateBasic()
			if err != nil {
				return err
			}

			// Sign and broadcast message
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	return cmd
}

// MustGetExampleCommitteeChangeProposal is a helper function to return an example json proposal
func MustGetExampleCommitteeChangeProposal(cdc codec.Codec) string {
	exampleChangeProposal, err := types.NewCommitteeChangeProposal(
		"A Title",
		"A description of this proposal.",
		types.MustNewMemberCommittee(
			1,
			"The description of this committee.",
			[]sdk.AccAddress{sdk.AccAddress(crypto.AddressHash([]byte("exampleAddress")))},
			[]types.Permission{
				&types.GodPermission{},
			},
			sdk.MustNewDecFromStr("0.8"),
			time.Hour*24*7,
			types.TALLY_OPTION_FIRST_PAST_THE_POST,
		),
	)
	if err != nil {
		panic(err)
	}
	exampleChangeProposalBz, err := cdc.MarshalJSON(&exampleChangeProposal)
	if err != nil {
		panic(err)
	}
	var out bytes.Buffer
	if err = json.Indent(&out, exampleChangeProposalBz, "", "  "); err != nil {
		panic(err)
	}
	return out.String()
}

// MustGetExampleCommitteeDeleteProposal is a helper function to return an example json proposal
func MustGetExampleCommitteeDeleteProposal(cdc codec.Codec) string {
	exampleDeleteProposal := types.NewCommitteeDeleteProposal(
		"A Title",
		"A description of this proposal.",
		1,
	)
	bz, err := cdc.MarshalJSON(&exampleDeleteProposal)
	if err != nil {
		panic(err)
	}
	var out bytes.Buffer
	if err = json.Indent(&out, bz, "", "  "); err != nil {
		panic(err)
	}
	return out.String()
}

// MustGetExampleParameterChangeProposal is a helper function to return an example json proposal
func MustGetExampleParameterChangeProposal(cdc codec.Codec) string {
	value := fmt.Sprintf("\"%d\"", 1000000000)
	exampleParameterChangeProposal := paramsproposal.NewParameterChangeProposal(
		"A Title",
		"A description of this proposal.",
		[]paramsproposal.ParamChange{paramsproposal.NewParamChange("cdp", "SurplusAuctionThreshold", value)},
	)
	bz, err := cdc.MarshalJSON(exampleParameterChangeProposal)
	if err != nil {
		panic(err)
	}
	var out bytes.Buffer
	if err = json.Indent(&out, bz, "", "  "); err != nil {
		panic(err)
	}
	return out.String()
}
