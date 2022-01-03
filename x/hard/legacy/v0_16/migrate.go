package v0_16

import (
	v015hard "github.com/kava-labs/kava/x/hard/legacy/v0_15"
	v016hard "github.com/kava-labs/kava/x/hard/types"
)

func migrateParams(params v015hard.Params) v016hard.Params {
	moneyMarkets := make([]v016hard.MoneyMarket, len(params.MoneyMarkets))
	for i, mm := range params.MoneyMarkets {
		moneyMarkets[i] = v016hard.MoneyMarket{
			Denom: mm.Denom,
			BorrowLimit: v016hard.BorrowLimit{
				HasMaxLimit:  mm.BorrowLimit.HasMaxLimit,
				MaximumLimit: mm.BorrowLimit.MaximumLimit,
				LoanToValue:  mm.BorrowLimit.LoanToValue,
			},
			SpotMarketID:     mm.SpotMarketID,
			ConversionFactor: mm.ConversionFactor,
			InterestRateModel: v016hard.InterestRateModel{
				BaseRateAPY:    mm.InterestRateModel.BaseRateAPY,
				BaseMultiplier: mm.InterestRateModel.BaseMultiplier,
				Kink:           mm.InterestRateModel.Kink,
				JumpMultiplier: mm.InterestRateModel.JumpMultiplier,
			},
			ReserveFactor:          mm.ReserveFactor,
			KeeperRewardPercentage: mm.KeeperRewardPercentage,
		}
	}

	return v016hard.Params{
		MoneyMarkets:          moneyMarkets,
		MinimumBorrowUSDValue: params.MinimumBorrowUSDValue,
	}
}

func migrateDeposits(oldDeposits v015hard.Deposits) v016hard.Deposits {
	deposits := make(v016hard.Deposits, len(oldDeposits))
	for i, deposit := range oldDeposits {

		interestFactors := make(v016hard.SupplyInterestFactors, len(deposit.Index))
		for j, interestFactor := range deposit.Index {
			interestFactors[j] = v016hard.SupplyInterestFactor{
				Denom: interestFactor.Denom,
				Value: interestFactor.Value,
			}
		}

		deposits[i] = v016hard.Deposit{
			Depositor: deposit.Depositor,
			Amount:    deposit.Amount,
			Index:     interestFactors,
		}
	}
	return deposits
}

func migratePrevAccTimes(oldPrevAccTimes v015hard.GenesisAccumulationTimes) v016hard.GenesisAccumulationTimes {
	prevAccTimes := make(v016hard.GenesisAccumulationTimes, len(oldPrevAccTimes))
	for i, prevAccTime := range oldPrevAccTimes {
		prevAccTimes[i] = v016hard.GenesisAccumulationTime{
			CollateralType:           prevAccTime.CollateralType,
			PreviousAccumulationTime: prevAccTime.PreviousAccumulationTime,
			SupplyInterestFactor:     prevAccTime.SupplyInterestFactor,
			BorrowInterestFactor:     prevAccTime.BorrowInterestFactor,
		}
	}
	return prevAccTimes
}

func migrateBorrows(oldBorrows v015hard.Borrows) v016hard.Borrows {
	borrows := make(v016hard.Borrows, len(oldBorrows))
	for i, borrow := range oldBorrows {
		interestFactors := make(v016hard.BorrowInterestFactors, len(borrow.Index))
		for j, interestFactor := range borrow.Index {
			interestFactors[j] = v016hard.BorrowInterestFactor{
				Denom: interestFactor.Denom,
				Value: interestFactor.Value,
			}
		}
		borrows[i] = v016hard.Borrow{
			Borrower: borrow.Borrower,
			Amount:   borrow.Amount,
			Index:    interestFactors,
		}
	}
	return borrows
}

// Migrate converts v0.15 hard state and returns it in v0.16 format
func Migrate(oldState v015hard.GenesisState) *v016hard.GenesisState {
	return &v016hard.GenesisState{
		Params:                    migrateParams(oldState.Params),
		PreviousAccumulationTimes: migratePrevAccTimes(oldState.PreviousAccumulationTimes),
		Deposits:                  migrateDeposits(oldState.Deposits),
		Borrows:                   migrateBorrows(oldState.Borrows),
		TotalSupplied:             oldState.TotalSupplied,
		TotalBorrowed:             oldState.TotalBorrowed,
		TotalReserves:             oldState.TotalReserves,
	}
}
