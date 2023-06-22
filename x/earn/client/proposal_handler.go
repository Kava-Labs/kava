package client

import (
	"github.com/kava-labs/kava/x/earn/client/cli"

	govclient "github.com/cosmos/cosmos-sdk/x/gov/client"
)

// community-pool deposit/withdraw proposal handlers
var (
	DepositProposalHandler  = govclient.NewProposalHandler(cli.GetCmdSubmitCommunityPoolDepositProposal)
	WithdrawProposalHandler = govclient.NewProposalHandler(cli.GetCmdSubmitCommunityPoolWithdrawProposal)
)
