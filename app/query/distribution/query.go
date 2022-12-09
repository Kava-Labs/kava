package distribution

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	distrkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"

	communityKeeper "github.com/kava-labs/kava/x/community/keeper"
)

var _ distrtypes.QueryServer = &queryServer{}

type queryServer struct {
	distrkeeper.Keeper

	communityKeeper communityKeeper.Keeper
}

// NewQueryServer returns a grpc query server for the distribution module.
// It forwards most requests to the distribution keeper, except for the community pool request which is mapped to the kava community module.
func NewQueryServer(distrKeeper distrkeeper.Keeper, commKeeper communityKeeper.Keeper) distrtypes.QueryServer {
	return &queryServer{
		Keeper:          distrKeeper,
		communityKeeper: commKeeper,
	}
}

// CommunityPool queries the kava community module
func (q queryServer) CommunityPool(c context.Context, req *distrtypes.QueryCommunityPoolRequest) (*distrtypes.QueryCommunityPoolResponse, error) {

	// TODO fetch the community module and convert to the correct format
	// pool := distrtypes.Pool{q.communityKeeper.Balance()}

	return &types.QueryCommunityPoolResponse{Pool: sdk.NewDecCoins(sdk.NewDecCoin("fake-coin", sdk.OneInt()))}, nil
}
