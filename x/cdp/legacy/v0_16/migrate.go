package v0_16

import (
	v015cdp "github.com/kava-labs/kava/x/cdp/legacy/v0_15"
	v016cdp "github.com/kava-labs/kava/x/cdp/types"
)

func migrateParams(params v015cdp.Params) v016cdp.Params {
	// migrate collateral params
	collateralParams := make(v016cdp.CollateralParams, len(params.CollateralParams))
	for i, cp := range params.CollateralParams {
		collateralParams[i] = v016cdp.CollateralParam{
			Denom:                            cp.Denom,
			Type:                             cp.Type,
			LiquidationRatio:                 cp.LiquidationRatio,
			DebtLimit:                        cp.DebtLimit,
			StabilityFee:                     cp.StabilityFee,
			AuctionSize:                      cp.AuctionSize,
			LiquidationPenalty:               cp.LiquidationPenalty,
			SpotMarketID:                     cp.SpotMarketID,
			LiquidationMarketID:              cp.LiquidationMarketID,
			KeeperRewardPercentage:           cp.KeeperRewardPercentage,
			CheckCollateralizationIndexCount: cp.CheckCollateralizationIndexCount,
			ConversionFactor:                 cp.ConversionFactor,
		}
	}

	return v016cdp.Params{
		CollateralParams: collateralParams,
		DebtParam: v016cdp.DebtParam{
			Denom:            params.DebtParam.Denom,
			ReferenceAsset:   params.DebtParam.ReferenceAsset,
			ConversionFactor: params.DebtParam.ConversionFactor,
			DebtFloor:        params.DebtParam.DebtFloor,
		},
		GlobalDebtLimit:         params.GlobalDebtLimit,
		SurplusAuctionThreshold: params.SurplusAuctionThreshold,
		SurplusAuctionLot:       params.SurplusAuctionLot,
		DebtAuctionThreshold:    params.DebtAuctionThreshold,
		DebtAuctionLot:          params.DebtAuctionLot,
		CircuitBreaker:          params.CircuitBreaker,
	}
}

func migrateCDPs(oldCDPs v015cdp.CDPs) v016cdp.CDPs {
	cdps := make(v016cdp.CDPs, len(oldCDPs))
	for i, cdp := range oldCDPs {
		cdps[i] = v016cdp.CDP{
			ID:              cdp.ID,
			Owner:           cdp.Owner,
			Type:            cdp.Type,
			Collateral:      cdp.Collateral,
			Principal:       cdp.Principal,
			AccumulatedFees: cdp.AccumulatedFees,
			FeesUpdated:     cdp.FeesUpdated,
			InterestFactor:  cdp.InterestFactor,
		}
	}
	return cdps
}

func migrateDeposits(oldDeposits v015cdp.Deposits) v016cdp.Deposits {
	deposits := make(v016cdp.Deposits, len(oldDeposits))
	for i, deposit := range oldDeposits {
		deposits[i] = v016cdp.Deposit{
			CdpID:     deposit.CdpID,
			Depositor: deposit.Depositor,
			Amount:    deposit.Amount,
		}
	}
	return deposits
}

func migratePrevAccTimes(oldPrevAccTimes v015cdp.GenesisAccumulationTimes) v016cdp.GenesisAccumulationTimes {
	prevAccTimes := make(v016cdp.GenesisAccumulationTimes, len(oldPrevAccTimes))
	for i, prevAccTime := range oldPrevAccTimes {
		prevAccTimes[i] = v016cdp.GenesisAccumulationTime{
			CollateralType:           prevAccTime.CollateralType,
			PreviousAccumulationTime: prevAccTime.PreviousAccumulationTime,
			InterestFactor:           prevAccTime.InterestFactor,
		}
	}
	return prevAccTimes
}

func migrateTotalPrincipals(oldTotalPrincipals v015cdp.GenesisTotalPrincipals) v016cdp.GenesisTotalPrincipals {
	totalPrincipals := make(v016cdp.GenesisTotalPrincipals, len(oldTotalPrincipals))
	for i, tp := range oldTotalPrincipals {
		totalPrincipals[i] = v016cdp.GenesisTotalPrincipal{
			CollateralType: tp.CollateralType,
			TotalPrincipal: tp.TotalPrincipal,
		}
	}
	return totalPrincipals
}

// Migrate converts v0.15 cdp state and returns it in v0.16 format
func Migrate(oldState v015cdp.GenesisState) *v016cdp.GenesisState {
	return &v016cdp.GenesisState{
		Params:                    migrateParams(oldState.Params),
		CDPs:                      migrateCDPs(oldState.CDPs),
		Deposits:                  migrateDeposits(oldState.Deposits),
		StartingCdpID:             oldState.StartingCdpID,
		DebtDenom:                 oldState.DebtDenom,
		GovDenom:                  oldState.GovDenom,
		PreviousAccumulationTimes: migratePrevAccTimes(oldState.PreviousAccumulationTimes),
		TotalPrincipals:           migrateTotalPrincipals(oldState.TotalPrincipals),
	}
}
