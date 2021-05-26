package cli

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/client/utils"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/cosmos/cosmos-sdk/x/params"

	"github.com/tendermint/tendermint/crypto"

	"github.com/kava-labs/kava/x/committee/types"
)

func GetTxCmd(storeKey string, cdc *codec.Codec) *cobra.Command {
	txCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "committee governance transactions subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	txCmd.AddCommand(flags.PostCommands(
		GetCmdVote(cdc),
		GetCmdSubmitProposal(cdc),
	)...)

	return txCmd
}

// GetCmdSubmitProposal returns the command to submit a proposal to a committee
func GetCmdSubmitProposal(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "submit-proposal [committee-id] [proposal-file]",
		Short: "Submit a governance proposal to a particular committee",
		Long: fmt.Sprintf(`Submit a proposal to a committee so they can vote on it.

The proposal file must be the json encoded forms of the proposal type you want to submit.
For example:
%s
`, MustGetExampleParameterChangeProposal(cdc)),
		Args:    cobra.ExactArgs(2),
		Example: fmt.Sprintf("%s tx %s submit-proposal 1 your-proposal.json", version.ClientName, types.ModuleName),
		RunE: func(cmd *cobra.Command, args []string) error {
			inBuf := bufio.NewReader(cmd.InOrStdin())
			txBldr := auth.NewTxBuilderFromCLI(inBuf).WithTxEncoder(utils.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			// Get proposing address
			proposer := cliCtx.GetFromAddress()

			// Get committee ID
			committeeID, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("committee-id %s not a valid int", args[0])
			}

			// Get the proposal
			bz, err := ioutil.ReadFile(args[1])
			if err != nil {
				return err
			}
			var pubProposal types.PubProposal
			if err := cdc.UnmarshalJSON(bz, &pubProposal); err != nil {
				return err
			}
			if err = pubProposal.ValidateBasic(); err != nil {
				return err
			}

			// Build message and run basic validation
			msg := types.NewMsgSubmitProposal(pubProposal, proposer, committeeID)
			err = msg.ValidateBasic()
			if err != nil {
				return err
			}

			// Sign and broadcast message
			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}

	return cmd
}

// GetCmdVote returns the command to vote on a proposal.
func GetCmdVote(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:     "vote [proposal-id] [vote]",
		Args:    cobra.ExactArgs(2),
		Short:   "Vote for an active proposal",
		Long:    "Submit a vote for the proposal with id [proposal-id].",
		Example: fmt.Sprintf("%s tx %s vote 2 yes", version.ClientName, types.ModuleName),
		RunE: func(cmd *cobra.Command, args []string) error {
			inBuf := bufio.NewReader(cmd.InOrStdin())
			txBldr := auth.NewTxBuilderFromCLI(inBuf).WithTxEncoder(utils.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			// Get voting address
			from := cliCtx.GetFromAddress()

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
				vote = types.Yes
			case "no", "n":
				vote = types.No
			case "abstain", "a":
				vote = types.Abstain
			default:
				return fmt.Errorf("must specify a valid vote type: (yes/y, no/n, abstain/a)")
			}

			// Build vote message and run basic validation
			msg := types.NewMsgVote(from, proposalID, vote)
			err = msg.ValidateBasic()
			if err != nil {
				return err
			}

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}

// GetGovCmdSubmitProposal returns a command to submit a proposal to the gov module. It is passed to the gov module for use on its command subtree.
func GetGovCmdSubmitProposal(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "committee [proposal-file] [deposit]",
		Short: "Submit a governance proposal to change a committee.",
		Long: fmt.Sprintf(`Submit a governance proposal to create, alter, or delete a committee.

The proposal file must be the json encoded form of the proposal type you want to submit.
For example, to create or update a committee:
%s

and to delete a committee:
%s
`, MustGetExampleCommitteeChangeProposal(cdc), MustGetExampleCommitteeDeleteProposal(cdc)),
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			inBuf := bufio.NewReader(cmd.InOrStdin())
			txBldr := auth.NewTxBuilderFromCLI(inBuf).WithTxEncoder(utils.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			// Get proposing address
			proposer := cliCtx.GetFromAddress()

			// Get the deposit
			deposit, err := sdk.ParseCoins(args[1])
			if err != nil {
				return err
			}

			// Get the proposal
			bz, err := ioutil.ReadFile(args[0])
			if err != nil {
				return err
			}
			var content govtypes.Content
			if err := cdc.UnmarshalJSON(bz, &content); err != nil {
				return err
			}
			if err = content.ValidateBasic(); err != nil {
				return err
			}

			// Build message and run basic validation
			msg := govtypes.NewMsgSubmitProposal(content, deposit, proposer)
			err = msg.ValidateBasic()
			if err != nil {
				return err
			}

			// Sign and broadcast message
			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
	return cmd
}

// MustGetExampleCommitteeChangeProposal is a helper function to return an example json proposal
func MustGetExampleCommitteeChangeProposal(cdc *codec.Codec) string {
	exampleChangeProposal := types.NewCommitteeChangeProposal(
		"A Title",
		"A description of this proposal.",
		types.NewMemberCommittee(
			1,
			"The description of this committee.",
			[]sdk.AccAddress{sdk.AccAddress(crypto.AddressHash([]byte("exampleAddress")))},
			[]types.Permission{
				types.SimpleParamChangePermission{
					AllowedParams: types.AllowedParams{{Subspace: "cdp", Key: "CircuitBreaker"}},
				},
			},
			sdk.MustNewDecFromStr("0.8"),
			time.Hour*24*7,
			types.FirstPastThePost,
		),
	)
	exampleChangeProposalBz, err := cdc.MarshalJSONIndent(exampleChangeProposal, "", "  ")
	if err != nil {
		panic(err)
	}
	return string(exampleChangeProposalBz)
}

// MustGetExampleCommitteeDeleteProposal is a helper function to return an example json proposal
func MustGetExampleCommitteeDeleteProposal(cdc *codec.Codec) string {
	exampleDeleteProposal := types.NewCommitteeDeleteProposal(
		"A Title",
		"A description of this proposal.",
		1,
	)
	exampleDeleteProposalBz, err := cdc.MarshalJSONIndent(exampleDeleteProposal, "", "  ")
	if err != nil {
		panic(err)
	}
	return string(exampleDeleteProposalBz)
}

// MustGetExampleParameterChangeProposal is a helper function to return an example json proposal
func MustGetExampleParameterChangeProposal(cdc *codec.Codec) string {
	value := fmt.Sprintf("\"%d\"", 1000000000)
	exampleParameterChangeProposal := params.NewParameterChangeProposal(
		"A Title",
		"A description of this proposal.",
		[]params.ParamChange{params.NewParamChange("cdp", "SurplusAuctionThreshold", value)},
	)
	exampleParameterChangeProposalBz, err := cdc.MarshalJSONIndent(exampleParameterChangeProposal, "", "  ")
	if err != nil {
		panic(err)
	}
	return string(exampleParameterChangeProposalBz)
}
