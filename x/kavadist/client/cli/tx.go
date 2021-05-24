package cli

import (
	"bufio"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/client/utils"
	"github.com/cosmos/cosmos-sdk/x/gov"

	"github.com/kava-labs/kava/x/kavadist/types"
)

// GetCmdSubmitProposal implements the command to submit a community-pool-spend proposal
func GetCmdSubmitProposal(cdc *codec.Codec) *cobra.Command {
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
	]
}
`,
				version.ClientName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			inBuf := bufio.NewReader(cmd.InOrStdin())
			txBldr := auth.NewTxBuilderFromCLI(inBuf).WithTxEncoder(utils.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContextWithInput(inBuf).WithCodec(cdc)

			proposal, err := ParseCommunityPoolMultiSpendProposalJSON(cdc, args[0])
			if err != nil {
				return err
			}

			from := cliCtx.GetFromAddress()
			content := types.NewCommunityPoolMultiSpendProposal(proposal.Title, proposal.Description, proposal.RecipientList)

			msg := gov.NewMsgSubmitProposal(content, proposal.Deposit, from)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}

	return cmd
}
