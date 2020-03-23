package cli

import (
	"fmt"
	"io/ioutil"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/client/utils"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/kava-labs/kava/x/committee/types"
)

// // Proposal flags
// const (
// 	FlagTitle        = "title"
// 	FlagDescription  = "description"
// 	flagProposalType = "type"
// 	FlagDeposit      = "deposit"
// 	flagVoter        = "voter"
// 	flagDepositor    = "depositor"
// 	flagStatus       = "status"
// 	flagNumLimit     = "limit"
// 	FlagProposal     = "proposal"
// )

// type proposal struct {
// 	Title       string
// 	Description string
// 	Type        string
// 	Deposit     string
// }

// // ProposalFlags defines the core required fields of a proposal. It is used to
// // verify that these values are not provided in conjunction with a JSON proposal
// // file.
// var ProposalFlags = []string{
// 	FlagTitle,
// 	FlagDescription,
// 	flagProposalType,
// 	FlagDeposit,
// }

// GetTxCmd returns the transaction commands for this module
// governance ModuleClient is slightly different from other ModuleClients in that
// it contains a slice of "proposal" child commands. These commands are respective
// to proposal type handlers that are implemented in other modules but are mounted
// under the governance CLI (eg. parameter change proposals).
func GetTxCmd(storeKey string, cdc *codec.Codec /*, pcmds []*cobra.Command*/) *cobra.Command { // TODO why is storeKey here?
	txCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "committee governance transactions subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmdSubmitProp := GetCmdSubmitProposal(cdc)
	// for _, pcmd := range pcmds {
	// 	cmdSubmitProp.AddCommand(client.PostCommands(pcmd)[0])
	// }

	txCmd.AddCommand(client.PostCommands(
		GetCmdVote(cdc),
		cmdSubmitProp,
	)...)

	return txCmd
}

// // GetCmdSubmitProposal is the root command on which commands for submitting proposals are registered.
// func GetCmdSubmitProposal(cdc *codec.Codec) *cobra.Command {
// 	cmd := &cobra.Command{
// 		Use:                        "submit-proposal [committee-id]",
// 		Short:                      "Submit a governance proposal to a particular committee.", // TODO
// 		DisableFlagParsing:         true,
// 		SuggestionsMinimumDistance: 2,
// 		RunE:                       client.ValidateCmd,
// 	}

// 	return cmd
// }

// GetCmdSubmitProposal
func GetCmdSubmitProposal(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "submit-proposal [committee-id] [proposal-file]",
		Short: "Submit a governance proposal to a particular committee.",
		Long:  "", // TODO
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			txBldr := auth.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))
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

// GetCmdVote implements creating a new vote command.
func GetCmdVote(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "vote [proposal-id]",
		Args:  cobra.ExactArgs(1),
		Short: "Vote for an active proposal", // TODO
		// 		Long: strings.TrimSpace(
		// 			fmt.Sprintf(`Submit a vote for an active proposal. You can
		// find the proposal-id by running "%s query gov proposals".

		// Example:
		// $ %s tx gov vote 1 yes --from mykey
		// `,
		// 				version.ClientName, version.ClientName,
		// 			),
		// 		),
		RunE: func(cmd *cobra.Command, args []string) error {
			txBldr := auth.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			// Get voting address
			from := cliCtx.GetFromAddress()

			// validate that the proposal id is a uint
			proposalID, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("proposal-id %s not a valid int, please input a valid proposal-id", args[0])
			}

			// Build vote message and run basic validation
			msg := types.NewMsgVote(from, proposalID)
			err = msg.ValidateBasic()
			if err != nil {
				return err
			}

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}

// TODO this could replace the whole gov submit-proposal cmd, remove and replace the gov cmd in kvcli main.go
// would want the documentation/examples though
func GetGovCmdSubmitProposal(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "committee [proposal-file] [deposit]",
		Short: "Submit a governance proposal to change a committee.",
		Long:  "This command will work with either CommitteeChange proposals or CommitteeDelete proposals.", // TODO
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			txBldr := auth.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			// Get proposing address
			proposer := cliCtx.GetFromAddress()

			// Get the deposit
			deposit, err := sdk.ParseCoins(args[0])
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
