package swap

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/x/swap/keeper"
	"github.com/kava-labs/kava/x/swap/types"
)

// InitGenesis initializes story state from genesis file
func InitGenesis(ctx sdk.Context, k keeper.Keeper, gs types.GenesisState) {
	if err := gs.Validate(); err != nil {
		panic(fmt.Sprintf("failed to validate %s genesis state: %s", types.ModuleName, err))
	}

	k.SetParams(ctx, gs.Params)
	for _, pr := range gs.PoolRecords {
		k.SetPool(ctx, pr)
	}
	for _, sh := range gs.ShareRecords {
		k.SetDepositorShares(ctx, sh)
	}
}

// ExportGenesis exports the genesis state
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) types.GenesisState {
	params := k.GetParams(ctx)
	pools := k.GetAllPools(ctx)
	shares := k.GetAllDepositorShares(ctx)

	return types.NewGenesisState(params, pools, shares)
}
