package rest

import (
	"github.com/gorilla/mux"

	"github.com/cosmos/cosmos-sdk/client/context"
)

// REST Variable names
const (
	RestProposalID  = "proposal-id"
	RestCommitteeID = "committee-id"
	RestVoter       = "voter"
	//RestProposalStatus = "status"
	//RestNumLimit       = "limit"
)

// // ProposalRESTHandler defines a REST handler implemented in another module. The
// // sub-route is mounted on the governance REST handler.
// type ProposalRESTHandler struct {
// 	SubRoute string
// 	Handler  func(http.ResponseWriter, *http.Request)
// }

// RegisterRoutes - Central function to define routes that get registered by the main application
func RegisterRoutes(cliCtx context.CLIContext, r *mux.Router /*, phs []ProposalRESTHandler*/) {
	registerQueryRoutes(cliCtx, r)
	registerTxRoutes(cliCtx, r /* , phs*/)
}
