package distribution

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	distrkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
)

var _ distrtypes.QueryServer = &queryServer{}

type queryServer struct {
	distrkeeper.Keeper

	communityKeeper CommunityKeeper
}

// NewQueryServer returns a grpc query server for the distribution module.
// It forwards most requests to the distribution keeper, except for the community pool request
// which is mapped to the kava community module.
func NewQueryServer(distrKeeper distrkeeper.Keeper, commKeeper CommunityKeeper) distrtypes.QueryServer {
	return &queryServer{
		Keeper:          distrKeeper,
		communityKeeper: commKeeper,
	}
}

// CommunityPool queries the kava community module
// The original community pool, which is a separately accounted for portion of x/auth's fee pool
// is replaces with the x/community module account.
// TODO: implement legacy community pool balance query in x/community
// To query the original community pool, including historical values, use x/community's LegacyCommunityPoolBalance
func (q queryServer) CommunityPool(c context.Context, req *distrtypes.QueryCommunityPoolRequest) (*distrtypes.QueryCommunityPoolResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	balance := q.communityKeeper.GetModuleAccountBalance(ctx)
	return &distrtypes.QueryCommunityPoolResponse{Pool: sdk.NewDecCoinsFromCoins(balance...)}, nil
}
