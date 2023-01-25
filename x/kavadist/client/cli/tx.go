package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/kava-labs/kava/x/kavadist/types"
)

const (
	flagDeposit = "deposit"
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
			msg, err := govtypes.NewMsgSubmitProposal(content, proposal.Deposit, from)
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

// NewCmdSubmitCommunityPoolLendDepositProposal implements the command to submit a community-pool lend deposit proposal
func NewCmdSubmitCommunityPoolLendDepositProposal() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "community-pool-lend-deposit [proposal-file]",
		Args:  cobra.ExactArgs(1),
		Short: "Submit a community pool lend deposit proposal",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Submit a community pool lend deposit proposal along with an initial deposit.
The proposal details must be supplied via a JSON file.
Note that --deposit below is the initial proposal deposit submitted along with the proposal.
Example:
$ %s tx gov submit-proposal community-pool-lend-deposit <path/to/proposal.json> --deposit 1000000000ukava --from=<key_or_address>
Where proposal.json contains:
{
  "title": "Community Pool Deposit",
  "description": "Deposit some KAVA from community pool!",
  "amount": [
    {
      "denom": "ukava",
      "amount": "100000000000"
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
			// parse proposal
			proposal, err := ParseCommunityPoolLendDepositProposal(clientCtx.Codec, args[0])
			if err != nil {
				return err
			}

			deposit, err := parseInitialDeposit(cmd)
			if err != nil {
				return err
			}

			from := clientCtx.GetFromAddress()
			msg, err := govtypes.NewMsgSubmitProposal(&proposal, deposit, from)
			if err != nil {
				return err
			}
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().String(flagDeposit, "", "Initial deposit for the proposal")

	return cmd
}

// NewCmdSubmitCommunityPoolLendWithdrawProposal implements the command to submit a community-pool lend withdraw proposal
func NewCmdSubmitCommunityPoolLendWithdrawProposal() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "community-pool-lend-withdraw [proposal-file]",
		Args:  cobra.ExactArgs(1),
		Short: "Submit a community pool lend withdraw proposal",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Submit a community pool lend withdraw proposal along with an initial deposit.
The proposal details must be supplied via a JSON file.
Note that --deposit below is the initial proposal deposit submitted along with the proposal.
Example:
$ %s tx gov submit-proposal community-pool-lend-withdraw <path/to/proposal.json> --deposit 1000000000ukava --from=<key_or_address>
Where proposal.json contains:
{
  "title": "Community Pool Withdrawal",
  "description": "Withdraw some KAVA from community pool!",
  "amount": [
    {
      "denom": "ukava",
      "amount": "100000000000"
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
			// parse proposal
			proposal, err := ParseCommunityPoolLendWithdrawProposal(clientCtx.Codec, args[0])
			if err != nil {
				return err
			}

			deposit, err := parseInitialDeposit(cmd)
			if err != nil {
				return err
			}
			from := clientCtx.GetFromAddress()
			msg, err := govtypes.NewMsgSubmitProposal(&proposal, deposit, from)
			if err != nil {
				return err
			}
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().String(flagDeposit, "", "Initial deposit for the proposal")

	return cmd
}

func parseInitialDeposit(cmd *cobra.Command) (sdk.Coins, error) {
	// parse initial deposit
	depositStr, err := cmd.Flags().GetString(flagDeposit)
	if err != nil {
		return nil, fmt.Errorf("no initial deposit found. did you set --deposit? %s", err)
	}
	deposit, err := sdk.ParseCoinsNormalized(depositStr)
	if err != nil {
		return nil, fmt.Errorf("unable to parse deposit: %s", err)
	}
	if !deposit.IsValid() || deposit.IsZero() {
		return nil, fmt.Errorf("no initial deposit set, use --deposit flag")
	}
	return deposit, nil
}
