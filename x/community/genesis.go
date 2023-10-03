// the community module has no genesis state but must init its module account on init
package community

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/x/community/keeper"
	"github.com/kava-labs/kava/x/community/types"
)

// InitGenesis initializes the community module account and stores the genesis state
func InitGenesis(ctx sdk.Context, k keeper.Keeper, ak types.AccountKeeper, gs types.GenesisState) {
	// check if the module account exists
	if moduleAcc := ak.GetModuleAccount(ctx, types.ModuleAccountName); moduleAcc == nil {
		panic(fmt.Sprintf("%s module account has not been set", types.ModuleAccountName))
	}

	k.SetParams(ctx, gs.Params)
	k.SetStakingRewardsState(ctx, gs.StakingRewardsState)
}

// ExportGenesis exports the store to a genesis state
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) types.GenesisState {
	params, found := k.GetParams(ctx)
	if !found {
		params = types.Params{}
	}

	stakingRewardsState := k.GetStakingRewardsState(ctx)

	return types.NewGenesisState(params, stakingRewardsState)
}
