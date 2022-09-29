package client

import (
	govclient "github.com/cosmos/cosmos-sdk/x/gov/client"

	"github.com/kava-labs/kava/x/earn/client/cli"
	"github.com/kava-labs/kava/x/earn/client/rest"
)

// community-pool deposit/withdraw proposal handlers
var (
	DepositProposalHandler  = govclient.NewProposalHandler(cli.GetCmdSubmitCommunityPoolDepositProposal, rest.DepositProposalRESTHandler)
	WithdrawProposalHandler = govclient.NewProposalHandler(cli.GetCmdSubmitCommunityPoolWithdrawProposal, rest.WithdrawProposalRESTHandler)
)
