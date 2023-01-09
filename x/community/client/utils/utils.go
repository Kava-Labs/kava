package utils

import (
	"encoding/json"
	"os"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
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

// communityPoolJSONProposal defines the new Msg-based proposal.
type communityPoolJSONProposal struct {
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Messages    []json.RawMessage `json:"messages,omitempty"`
	Deposit     string            `json:"deposit"`
}

// ParseCommunityPoolProposalJSON reads and parses a CommunityPoolProposalJSON from a file.
func ParseCommunityPoolProposalJSON(cdc codec.JSONCodec, proposalFile string) (*types.CommunityPoolProposal, sdk.Coins, error) {
	var jsonProposal communityPoolJSONProposal
	contents, err := os.ReadFile(proposalFile)
	if err != nil {
		return nil, nil, err
	}

	err = json.Unmarshal(contents, &jsonProposal)
	if err != nil {
		return nil, nil, err
	}

	msgs := make([]sdk.Msg, len(jsonProposal.Messages))
	for i, anyJSON := range jsonProposal.Messages {
		var msg sdk.Msg
		err := cdc.UnmarshalInterfaceJSON(anyJSON, &msg)
		if err != nil {
			return nil, nil, err
		}

		msgs[i] = msg
	}

	deposit, err := sdk.ParseCoinsNormalized(jsonProposal.Deposit)
	if err != nil {
		return nil, nil, err
	}

	communityPoolProposal, err := types.NewCommunityPoolProposal(jsonProposal.Title, jsonProposal.Description, msgs)
	if err != nil {
		return nil, nil, err
	}

	return communityPoolProposal, deposit, nil
}
