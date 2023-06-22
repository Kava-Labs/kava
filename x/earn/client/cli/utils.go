package cli

import (
	"os"

	"github.com/cosmos/cosmos-sdk/codec"

	"github.com/kava-labs/kava/x/earn/types"
)

// ParseCommunityPoolDepositProposalJSON reads and parses a CommunityPoolDepositProposalJSON from a file.
func ParseCommunityPoolDepositProposalJSON(cdc codec.JSONCodec, proposalFile string) (types.CommunityPoolDepositProposalJSON, error) {
	proposal := types.CommunityPoolDepositProposalJSON{}
	contents, err := os.ReadFile(proposalFile)
	if err != nil {
		return proposal, err
	}

	if err := cdc.UnmarshalJSON(contents, &proposal); err != nil {
		return proposal, err
	}

	return proposal, nil
}

// ParseCommunityPoolWithdrawProposalJSON reads and parses a CommunityPoolWithdrawProposalJSON from a file.
func ParseCommunityPoolWithdrawProposalJSON(cdc codec.JSONCodec, proposalFile string) (types.CommunityPoolWithdrawProposalJSON, error) {
	proposal := types.CommunityPoolWithdrawProposalJSON{}
	contents, err := os.ReadFile(proposalFile)
	if err != nil {
		return proposal, err
	}

	if err := cdc.UnmarshalJSON(contents, &proposal); err != nil {
		return proposal, err
	}

	return proposal, nil
}
