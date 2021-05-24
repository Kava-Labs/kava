package client

import (
	govclient "github.com/cosmos/cosmos-sdk/x/gov/client"

	"github.com/kava-labs/kava/x/kavadist/client/cli"
	"github.com/kava-labs/kava/x/kavadist/client/rest"
)

// param change proposal handler
var (
	ProposalHandler = govclient.NewProposalHandler(cli.GetCmdSubmitProposal, rest.ProposalRESTHandler)
)
