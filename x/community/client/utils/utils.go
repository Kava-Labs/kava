package utils

import (
	"os"

	"github.com/cosmos/cosmos-sdk/codec"

	"github.com/kava-labs/kava/x/community/types"
)

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
