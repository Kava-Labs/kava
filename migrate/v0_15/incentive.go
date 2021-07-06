package v0_15

// import (
// 	v0_14incentive "github.com/kava-labs/kava/x/incentive/legacy/v0_14"
// 	v0_15incentive "github.com/kava-labs/kava/x/incentive/types"
// )

// // Incentive migrates from a v0.14 incentive genesis state to a v0.15 incentive genesis state
// func Incentive(oldGenesis v0_14incentive.GenesisState) v0_15incentive.GenesisState {
// 	return v0_15incentive.NewGenesisState(
// 		convert14to15IncentiveParams(oldGenesis.Params),
// 		convert14to15GenesisAccumulationTimes(oldGenesis.USDXAccumulationTimes),
// 		convert14to15GenesisAccumulationTimes(oldGenesis.HardSupplyAccumulationTimes),
// 		convert14to15GenesisAccumulationTimes(oldGenesis.HardBorrowAccumulationTimes),
// 		convert14to15GenesisAccumulationTimes(oldGenesis.HardDelegatorAccumulationTimes),
// 		convert14to15USDXMintingClaims(oldGenesis.USDXMintingClaims),
// 		convert14to15HardLiquidityProviderClaims(oldGenesis.HardLiquidityProviderClaims),
// 	)
// }

// func convert14to15IncentiveParams(oldParams v0_14incentive.Params) v0_15incentive.Params {
// 	params := v0_15incentive.NewParams(
// 		convert14to15RewardPeriods(oldParams.USDXMintingRewardPeriods),
// 		convert14to15MultiRewardPeriods(oldParams.HardSupplyRewardPeriods),
// 		convert14to15MultiRewardPeriods(oldParams.HardBorrowRewardPeriods),
// 		convert14to15RewardPeriods(oldParams.HardDelegatorRewardPeriods),
// 		v0_15incentive.DefaultMultiRewardPeriods,
// 		convert14to15ClaimMultipliers(oldParams.ClaimMultipliers),
// 		oldParams.ClaimEnd,
// 	)
// 	return params
// }

// func convert14to15RewardPeriods(oldPeriods v0_14incentive.RewardPeriods) v0_15incentive.RewardPeriods {
// 	newPeriods := v0_15incentive.RewardPeriods{}
// 	for _, p := range oldPeriods {
// 		newPeriods = append(newPeriods, v0_15incentive.RewardPeriod(p))
// 	}
// 	return newPeriods
// }

// func convert14to15MultiRewardPeriods(oldPeriods v0_14incentive.MultiRewardPeriods) v0_15incentive.MultiRewardPeriods {
// 	newPeriods := v0_15incentive.MultiRewardPeriods{}
// 	for _, p := range oldPeriods {
// 		newPeriods = append(newPeriods, v0_15incentive.MultiRewardPeriod(p))
// 	}
// 	return newPeriods
// }

// func convert14to15GenesisAccumulationTimes(oldTimes v0_14incentive.GenesisAccumulationTimes) v0_15incentive.GenesisAccumulationTimes {
// 	newTimes := v0_15incentive.GenesisAccumulationTimes{}
// 	for _, t := range oldTimes {
// 		newTimes = append(newTimes, v0_15incentive.GenesisAccumulationTime(t))
// 	}
// 	return newTimes
// }

// func convert14to15ClaimMultipliers(oldMultipliers v0_14incentive.Multipliers) v0_15incentive.Multipliers {
// 	newMultipliers := v0_15incentive.Multipliers{}
// 	for _, m := range oldMultipliers {
// 		newMultipliers = append(newMultipliers, v0_15incentive.NewMultiplier(
// 			v0_15incentive.MultiplierName(m.Name),
// 			m.MonthsLockup,
// 			m.Factor,
// 		))
// 	}
// 	return newMultipliers
// }

// func convert14to15RewardIndexes(oldIndexes v0_14incentive.RewardIndexes) v0_15incentive.RewardIndexes {
// 	newIndexes := v0_15incentive.RewardIndexes{}
// 	for _, p := range oldIndexes {
// 		newIndexes = append(newIndexes, v0_15incentive.RewardIndex(p))
// 	}
// 	return newIndexes
// }

// func convert14to15MultiRewardIndexes(oldIndexes v0_14incentive.MultiRewardIndexes) v0_15incentive.MultiRewardIndexes {
// 	newIndexes := v0_15incentive.MultiRewardIndexes{}
// 	for _, p := range oldIndexes {
// 		newIndexes = append(newIndexes, v0_15incentive.NewMultiRewardIndex(
// 			p.CollateralType,
// 			convert14to15RewardIndexes(p.RewardIndexes),
// 		))
// 	}
// 	return newIndexes
// }

// func convert14to15HardLiquidityProviderClaims(oldClaims v0_14incentive.HardLiquidityProviderClaims) v0_15incentive.HardLiquidityProviderClaims {
// 	newClaims := v0_15incentive.HardLiquidityProviderClaims{}
// 	for _, c := range oldClaims {
// 		newClaims = append(newClaims, v0_15incentive.NewHardLiquidityProviderClaim(
// 			c.Owner,
// 			c.Reward,
// 			convert14to15MultiRewardIndexes(c.SupplyRewardIndexes),
// 			convert14to15MultiRewardIndexes(c.BorrowRewardIndexes),
// 			convert14to15RewardIndexes(c.DelegatorRewardIndexes),
// 		))
// 	}
// 	return newClaims
// }
// func convert14to15USDXMintingClaims(oldClaims v0_14incentive.USDXMintingClaims) v0_15incentive.USDXMintingClaims {
// 	newClaims := v0_15incentive.USDXMintingClaims{}
// 	for _, c := range oldClaims {
// 		newClaims = append(newClaims, v0_15incentive.NewUSDXMintingClaim(
// 			c.Owner,
// 			c.Reward,
// 			convert14to15RewardIndexes(c.RewardIndexes),
// 		))
// 	}
// 	return newClaims
// }
