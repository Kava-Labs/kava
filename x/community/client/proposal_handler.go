package client

import (
	govclient "github.com/cosmos/cosmos-sdk/x/gov/client"

	"github.com/kava-labs/kava/x/community/client/cli"
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
