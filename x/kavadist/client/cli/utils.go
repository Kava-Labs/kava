package cli

import (
	"io/ioutil"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/x/kavadist/types"
)

type (
	// CommunityPoolMultiSpendProposalJSON defines a CommunityPoolMultiSpendProposal with a deposit
	CommunityPoolMultiSpendProposalJSON struct {
		Title         string                     `json:"title" yaml:"title"`
		Description   string                     `json:"description" yaml:"description"`
		RecipientList types.MultiSpendRecipients `json:"recipient_list" yaml:"recipient_list"`
		Deposit       sdk.Coins                  `json:"deposit" yaml:"deposit"`
	}
)

// ParseCommunityPoolMultiSpendProposalJSON reads and parses a CommunityPoolMultiSpendProposalJSON from a file.
func ParseCommunityPoolMultiSpendProposalJSON(cdc *codec.Codec, proposalFile string) (CommunityPoolMultiSpendProposalJSON, error) {
	proposal := CommunityPoolMultiSpendProposalJSON{}

	contents, err := ioutil.ReadFile(proposalFile)
	if err != nil {
		return proposal, err
	}

	if err := cdc.UnmarshalJSON(contents, &proposal); err != nil {
		return proposal, err
	}

	return proposal, nil
}
