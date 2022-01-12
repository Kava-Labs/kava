package v0_16

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	v015hard "github.com/kava-labs/kava/x/hard/legacy/v0_15"
	v016hard "github.com/kava-labs/kava/x/hard/types"
)

// Denom generated via: echo -n transfer/channel-0/uatom | shasum -a 256 | awk '{printf "ibc/%s",toupper($1)}'
const UATOM_IBC_DENOM = "ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2"

func migrateParams(params v015hard.Params) v016hard.Params {
	var moneyMarkets []v016hard.MoneyMarket
	for _, mm := range params.MoneyMarkets {
		moneyMarket := v016hard.MoneyMarket{
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
		moneyMarkets = append(moneyMarkets, moneyMarket)
	}

	atomMoneyMarket := v016hard.MoneyMarket{
		Denom: UATOM_IBC_DENOM,
		BorrowLimit: v016hard.BorrowLimit{
			HasMaxLimit:  true,
			MaximumLimit: sdk.NewDec(25000000000),
			LoanToValue:  sdk.MustNewDecFromStr("0.5"),
		},
		SpotMarketID:     "atom:usd:30",
		ConversionFactor: sdk.NewInt(1000000),
		InterestRateModel: v016hard.InterestRateModel{
			BaseRateAPY:    sdk.ZeroDec(),
			BaseMultiplier: sdk.MustNewDecFromStr("0.05"),
			Kink:           sdk.MustNewDecFromStr("0.8"),
			JumpMultiplier: sdk.NewDec(5),
		},
		ReserveFactor:          sdk.MustNewDecFromStr("0.025"),
		KeeperRewardPercentage: sdk.MustNewDecFromStr("0.02"),
	}
	moneyMarkets = append(moneyMarkets, atomMoneyMarket)

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
