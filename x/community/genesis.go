// the community module has no genesis state but must init its module account on init
package community

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/x/community/keeper"
	"github.com/kava-labs/kava/x/community/types"
)

// InitGenesis initializes the community module account
func InitGenesis(ctx sdk.Context, k keeper.Keeper, ak types.AccountKeeper) {
	// check if the module account exists
	if moduleAcc := ak.GetModuleAccount(ctx, types.ModuleAccountName); moduleAcc == nil {
		panic(fmt.Sprintf("%s module account has not been set", types.ModuleAccountName))
	}
}
