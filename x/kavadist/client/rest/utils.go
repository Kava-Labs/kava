package rest

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/rest"

	"github.com/kava-labs/kava/x/kavadist/types"
)

type (
	// CommunityPoolMultiSpendProposalReq defines a community pool multi-spend proposal request body.
	CommunityPoolMultiSpendProposalReq struct {
		BaseReq rest.BaseReq `json:"base_req" yaml:"base_req"`

		Title         string                     `json:"title" yaml:"title"`
		Description   string                     `json:"description" yaml:"description"`
		RecipientList types.MultiSpendRecipients `json:"recipient_list" yaml:"recipient_list"`
		Deposit       sdk.Coins                  `json:"deposit" yaml:"deposit"`
		Proposer      sdk.AccAddress             `json:"proposer" yaml:"proposer"`
	}
)
