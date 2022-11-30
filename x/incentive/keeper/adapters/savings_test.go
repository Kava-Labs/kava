package adapters_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kava-labs/kava/x/incentive/keeper/adapters/savings"
	savingskeeper "github.com/kava-labs/kava/x/savings/keeper"
	savingstypes "github.com/kava-labs/kava/x/savings/types"
)

type SavingsSourceTester struct {
	savings.SourceAdapter

	keeper savingskeeper.Keeper
}

var _ AdapterSourceTester = SavingsSourceTester{}

func NewSavingsSourceTester(keeper savingskeeper.Keeper) SavingsSourceTester {
	return SavingsSourceTester{
		SourceAdapter: savings.NewSourceAdapter(keeper),
		keeper:        keeper,
	}
}

func (s SavingsSourceTester) Initialize(ctx sdk.Context) error {
	s.keeper.SetParams(ctx, savingstypes.NewParams(
		testDenoms,
	))

	return nil
}

func (s SavingsSourceTester) Deposit(ctx sdk.Context, owner sdk.AccAddress, amount sdk.Coins) error {
	return s.keeper.Deposit(
		ctx,
		owner,
		amount,
	)
}
