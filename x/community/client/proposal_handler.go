package client

import (
	"net/http"

	"github.com/cosmos/cosmos-sdk/client"
	govclient "github.com/cosmos/cosmos-sdk/x/gov/client"
	govrest "github.com/cosmos/cosmos-sdk/x/gov/client/rest"

	"github.com/kava-labs/kava/x/community/client/cli"
	"github.com/kava-labs/kava/x/community/types"
)

// community-pool proposal handlers
var (
	// Deprecated: Use CommunityPoolProposal instead
	LendDepositProposalHandler = govclient.NewProposalHandler(
		cli.NewCmdSubmitCommunityPoolLendDepositProposal,
		notImplementedRestHandler(types.ProposalTypeCommunityPoolLendDeposit),
	)
	// Deprecated: Use CommunityPoolProposal instead
	LendWithdrawProposalHandler = govclient.NewProposalHandler(
		cli.NewCmdSubmitCommunityPoolLendWithdrawProposal,
		notImplementedRestHandler(types.ProposalTypeCommunityPoolLendDeposit),
	)
	ProposalHandler = govclient.NewProposalHandler(
		cli.GetCmdSubmitCommunityPoolProposal,
		notImplementedRestHandler(types.ProposalTypeCommunityPool),
	)
)

func notImplementedRestHandler(subRoute string) govclient.RESTHandlerFn {
	return func(ctx client.Context) govrest.ProposalRESTHandler {
		return govrest.ProposalRESTHandler{
			SubRoute: subRoute,
			Handler: func(w http.ResponseWriter, r *http.Request) {
				http.Error(w, "Unimplemented", http.StatusNotImplemented)
			},
		}
	}
}
