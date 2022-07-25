package cli

import (
	"io/ioutil"

	"github.com/cosmos/cosmos-sdk/codec"

	"github.com/kava-labs/kava/x/earn/types"
)

// ParseCommunityPoolDepositProposalJSON reads and parses a CommunityPoolDepositProposalJSON from a file.
func ParseCommunityPoolDepositProposalJSON(cdc codec.JSONCodec, proposalFile string) (types.CommunityPoolDepositProposalJSON, error) {
	proposal := types.CommunityPoolDepositProposalJSON{}
	contents, err := ioutil.ReadFile(proposalFile)
	if err != nil {
		return proposal, err
	}

	if err := cdc.UnmarshalJSON(contents, &proposal); err != nil {
		return proposal, err
	}

	return proposal, nil
}
