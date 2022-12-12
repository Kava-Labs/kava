package client

import (
	"net/http"

	"github.com/cosmos/cosmos-sdk/client"
	govclient "github.com/cosmos/cosmos-sdk/x/gov/client"
	govrest "github.com/cosmos/cosmos-sdk/x/gov/client/rest"

	"github.com/kava-labs/kava/x/community/client/cli"
	"github.com/kava-labs/kava/x/community/types"
)

// community-pool deposit/withdraw lend proposal handlers
var (
	LendDepositProposalHandler = govclient.NewProposalHandler(
		cli.NewCmdSubmitCommunityPoolLendDepositProposal,
		notImplementedRestHandler(types.ProposalTypeCommunityPoolLendDeposit),
	)
	LendWithdrawProposalHandler = govclient.NewProposalHandler(
		cli.NewCmdSubmitCommunityPoolLendWithdrawProposal,
		notImplementedRestHandler(types.ProposalTypeCommunityPoolLendDeposit),
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
