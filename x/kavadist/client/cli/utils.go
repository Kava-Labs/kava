package cli

import (
	"io/ioutil"
	"os"

	"github.com/cosmos/cosmos-sdk/codec"

	"github.com/kava-labs/kava/x/kavadist/types"
)

// ParseCommunityPoolMultiSpendProposalJSON reads and parses a CommunityPoolMultiSpendProposalJSON from a file.
func ParseCommunityPoolMultiSpendProposalJSON(cdc codec.JSONCodec, proposalFile string) (types.CommunityPoolMultiSpendProposalJSON, error) {
	proposal := types.CommunityPoolMultiSpendProposalJSON{}
	contents, err := ioutil.ReadFile(proposalFile)
	if err != nil {
		return proposal, err
	}

	if err := cdc.UnmarshalJSON(contents, &proposal); err != nil {
		return proposal, err
	}

	return proposal, nil
}

// ParseCommunityPoolLendDepositProposal reads a JSON file and parses it to a CommunityPoolLendDepositProposal
func ParseCommunityPoolLendDepositProposal(
	cdc codec.JSONCodec,
	proposalFile string,
) (types.CommunityPoolLendDepositProposal, error) {
	proposal := types.CommunityPoolLendDepositProposal{}
	contents, err := os.ReadFile(proposalFile)
	if err != nil {
		return proposal, err
	}

	err = cdc.UnmarshalJSON(contents, &proposal)
	return proposal, err
}

// ParseCommunityPoolLendWithdrawProposal reads a JSON file and parses it to a CommunityPoolLendWithdrawProposal
func ParseCommunityPoolLendWithdrawProposal(
	cdc codec.JSONCodec,
	proposalFile string,
) (types.CommunityPoolLendWithdrawProposal, error) {
	proposal := types.CommunityPoolLendWithdrawProposal{}
	contents, err := os.ReadFile(proposalFile)
	if err != nil {
		return proposal, err
	}

	err = cdc.UnmarshalJSON(contents, &proposal)
	return proposal, err
}
