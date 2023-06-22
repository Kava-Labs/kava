package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/version"
	govv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"

	"github.com/kava-labs/kava/x/kavadist/types"
)

// GetCmdSubmitProposal implements the command to submit a community-pool multi-spend proposal
func GetCmdSubmitProposal() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "community-pool-multi-spend [proposal-file]",
		Args:  cobra.ExactArgs(1),
		Short: "Submit a community pool multi-spend proposal",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Submit a community pool multi-spend proposal along with an initial deposit.
The proposal details must be supplied via a JSON file.

Example:
$ %s tx gov submit-proposal community-pool-multi-spend <path/to/proposal.json> --from=<key_or_address>

Where proposal.json contains:

{
  "title": "Community Pool Multi-Spend",
  "description": "Pay many users some KAVA!",
  "recipient_list": [
		{
			"address": "kava1mz2003lathm95n5vnlthmtfvrzrjkrr53j4464",
			"amount": [
				{
					"denom": "ukava",
					"amount": "1000000"
				}
			]
		},
		{
			"address": "kava1zqezafa0luyetvtj8j67g336vaqtuudnsjq7vm",
			"amount": [
				{
					"denom": "ukava",
					"amount": "1000000"
				}
			]
		}
	],
	"deposit": [
		{
			"denom": "ukava",
			"amount": "1000000000"
		}
	]
}
`,
				version.AppName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			proposal, err := ParseCommunityPoolMultiSpendProposalJSON(clientCtx.Codec, args[0])
			if err != nil {
				return err
			}

			from := clientCtx.GetFromAddress()
			content := types.NewCommunityPoolMultiSpendProposal(proposal.Title, proposal.Description, proposal.RecipientList)
			msg, err := govv1beta1.NewMsgSubmitProposal(content, proposal.Deposit, from)
			if err != nil {
				return err
			}
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	return cmd
}
