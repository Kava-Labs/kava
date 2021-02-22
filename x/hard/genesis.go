package hard

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/x/hard/types"
)

// InitGenesis initializes the store state from a genesis state.
func InitGenesis(ctx sdk.Context, k Keeper, supplyKeeper types.SupplyKeeper, gs GenesisState) {
	if err := gs.Validate(); err != nil {
		panic(fmt.Sprintf("failed to validate %s genesis state: %s", ModuleName, err))
	}

	k.SetParams(ctx, gs.Params)

	for _, mm := range gs.Params.MoneyMarkets {
		k.SetMoneyMarket(ctx, mm.Denom, mm)
	}

	for _, gat := range gs.PreviousAccumulationTimes {
		k.SetPreviousAccrualTime(ctx, gat.CollateralType, gat.PreviousAccumulationTime)
		k.SetSupplyInterestFactor(ctx, gat.CollateralType, gat.SupplyInterestFactor)
		k.SetBorrowInterestFactor(ctx, gat.CollateralType, gat.BorrowInterestFactor)
	}

	for _, deposit := range gs.Deposits {
		k.SetDeposit(ctx, deposit)
	}

	for _, borrow := range gs.Borrows {
		k.SetBorrow(ctx, borrow)
	}

	k.SetSuppliedCoins(ctx, gs.TotalSupplied)
	k.SetBorrowedCoins(ctx, gs.TotalBorrowed)
	k.SetTotalReserves(ctx, gs.TotalReserves)

	// check if the module account exists
	DepositModuleAccount := supplyKeeper.GetModuleAccount(ctx, ModuleAccountName)
	if DepositModuleAccount == nil {
		panic(fmt.Sprintf("%s module account has not been set", DepositModuleAccount))
	}
}

// ExportGenesis export genesis state for hard module
func ExportGenesis(ctx sdk.Context, k Keeper) GenesisState {
	params := k.GetParams(ctx)

	gats := types.GenesisAccumulationTimes{}
	deposits := types.Deposits{}
	borrows := types.Borrows{}

	k.IterateDeposits(ctx, func(d types.Deposit) bool {
		k.BeforeDepositModified(ctx, d)
		syncedDeposit, found := k.GetSyncedDeposit(ctx, d.Depositor)
		if !found {
			panic(fmt.Sprintf("syncable deposit not found for %s", d.Depositor))
		}
		deposits = append(deposits, syncedDeposit)
		return false
	})

	k.IterateBorrows(ctx, func(b types.Borrow) bool {
		k.BeforeBorrowModified(ctx, b)
		syncedBorrow, found := k.GetSyncedBorrow(ctx, b.Borrower)
		if !found {
			panic(fmt.Sprintf("syncable borrow not found for %s", b.Borrower))
		}
		borrows = append(borrows, syncedBorrow)
		return false
	})

	totalSupplied, found := k.GetSuppliedCoins(ctx)
	if !found {
		totalSupplied = DefaultTotalSupplied
	}
	totalBorrowed, found := k.GetBorrowedCoins(ctx)
	if !found {
		totalBorrowed = DefaultTotalBorrowed
	}
	totalReserves, found := k.GetTotalReserves(ctx)
	if !found {
		totalReserves = DefaultTotalReserves
	}

	for _, mm := range params.MoneyMarkets {
		supplyFactor, f := k.GetSupplyInterestFactor(ctx, mm.Denom)
		if !f {
			supplyFactor = sdk.ZeroDec()
		}
		borrowFactor, f := k.GetBorrowInterestFactor(ctx, mm.Denom)
		if !f {
			borrowFactor = sdk.ZeroDec()
		}
		previousAccrualTime, f := k.GetPreviousAccrualTime(ctx, mm.Denom)
		if !f {
			previousAccrualTime = ctx.BlockTime()
		}
		gat := types.NewGenesisAccumulationTime(mm.Denom, previousAccrualTime, supplyFactor, borrowFactor)
		gats = append(gats, gat)

	}
	return NewGenesisState(
		params, gats, deposits, borrows,
		totalSupplied, totalBorrowed, totalReserves,
	)
}
