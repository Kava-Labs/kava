package client

import (
	"github.com/kava-labs/kava/x/kavadist/client/cli"

	govclient "github.com/cosmos/cosmos-sdk/x/gov/client"
)

// community-pool multi-spend proposal handler
var (
	ProposalHandler = govclient.NewProposalHandler(cli.GetCmdSubmitProposal)
)
