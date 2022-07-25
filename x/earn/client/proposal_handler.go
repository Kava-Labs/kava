package client

import (
	govclient "github.com/cosmos/cosmos-sdk/x/gov/client"

	"github.com/kava-labs/kava/x/earn/client/cli"
	"github.com/kava-labs/kava/x/earn/client/rest"
)

// community-pool multi-spend proposal handler
var (
	ProposalHandler = govclient.NewProposalHandler(cli.GetCmdSubmitCommunityPoolDepositProposal, rest.ProposalRESTHandler)
)
