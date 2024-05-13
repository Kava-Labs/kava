package precisebank

import (
	"fmt"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/x/precisebank/keeper"
	"github.com/kava-labs/kava/x/precisebank/types"
)

// InitGenesis initializes the store state from a genesis state.
func InitGenesis(
	ctx sdk.Context,
	keeper keeper.Keeper,
	ak types.AccountKeeper,
	bk types.BankKeeper,
	gs *types.GenesisState,
) {
	// Ensure the genesis state is valid
	if err := gs.Validate(); err != nil {
		panic(fmt.Sprintf("failed to validate %s genesis state: %s", types.ModuleName, err))
	}

	// Initialize module account
	if moduleAcc := ak.GetModuleAccount(ctx, types.ModuleName); moduleAcc == nil {
		panic(fmt.Sprintf("%s module account has not been set", types.ModuleName))
	}

	// Check module balance matches sum of fractional balances + remainder
	totalAmt := gs.TotalAmountWithRemainder()
	splitAmt := types.NewSplitBalanceFromFullAmount(totalAmt)

	moduleAddr := ak.GetModuleAddress(types.ModuleName)
	moduleBal := bk.GetBalance(ctx, moduleAddr, "ukava")

	if !moduleBal.Amount.Equal(splitAmt.IntegerAmount) {
		panic(fmt.Sprintf(
			"module account balance does not match sum of fractional balances and remainder, balance is %s but expected %s",
			moduleBal.Amount, splitAmt.IntegerAmount,
		))
	}

	// TODO: After keeper methods are implemented
	// - Set account FractionalBalances
}

// ExportGenesis returns a GenesisState for a given context and keeper.
func ExportGenesis(ctx sdk.Context, keeper keeper.Keeper) *types.GenesisState {
	return types.NewGenesisState(types.FractionalBalances{}, sdkmath.ZeroInt())
}
