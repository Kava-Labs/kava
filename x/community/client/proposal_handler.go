package client

import (
	"github.com/kava-labs/kava/x/community/client/cli"

	govclient "github.com/cosmos/cosmos-sdk/x/gov/client"
)

// community-pool deposit/withdraw lend proposal handlers
var (
	LendDepositProposalHandler = govclient.NewProposalHandler(
		cli.NewCmdSubmitCommunityPoolLendDepositProposal,
	)
	LendWithdrawProposalHandler = govclient.NewProposalHandler(
		cli.NewCmdSubmitCommunityPoolLendWithdrawProposal,
	)
)
