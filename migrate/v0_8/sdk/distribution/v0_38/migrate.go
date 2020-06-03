package v038

import (
	v038dist "github.com/cosmos/cosmos-sdk/x/distribution"

	v18de63dist "github.com/kava-labs/kava/migrate/v0_8/sdk/distribution/v18de63"
)

func Migrate(oldGenState v18de63dist.GenesisState) v038dist.GenesisState {

	// Changes: some fields moved into a params struct, some changes in json tags

	params := v038dist.Params{
		CommunityTax:        oldGenState.CommunityTax,
		BaseProposerReward:  oldGenState.BaseProposerReward,
		BonusProposerReward: oldGenState.BonusProposerReward,
		WithdrawAddrEnabled: oldGenState.WithdrawAddrEnabled,
	}

	withdrawInfos := []v038dist.DelegatorWithdrawInfo{}
	for _, v := range oldGenState.DelegatorWithdrawInfos {
		withdrawInfos = append(withdrawInfos, v038dist.DelegatorWithdrawInfo(v))
	}

	outstandingRewards := []v038dist.ValidatorOutstandingRewardsRecord{}
	for _, v := range oldGenState.OutstandingRewards {
		outstandingRewards = append(outstandingRewards, v038dist.ValidatorOutstandingRewardsRecord(v))
	}

	accumulatedComs := []v038dist.ValidatorAccumulatedCommissionRecord{}
	for _, v := range oldGenState.ValidatorAccumulatedCommissions {
		accumulatedComs = append(accumulatedComs, v038dist.ValidatorAccumulatedCommissionRecord(v))
	}

	histRewards := []v038dist.ValidatorHistoricalRewardsRecord{}
	for _, v := range oldGenState.ValidatorHistoricalRewards {
		histRewards = append(histRewards, v038dist.ValidatorHistoricalRewardsRecord{
			ValidatorAddress: v.ValidatorAddress,
			Period:           v.Period,
			Rewards: v038dist.NewValidatorHistoricalRewards(
				v.Rewards.CumulativeRewardRatio,
				v.Rewards.ReferenceCount,
			),
		})
	}

	currRewards := []v038dist.ValidatorCurrentRewardsRecord{}
	for _, v := range oldGenState.ValidatorCurrentRewards {
		currRewards = append(currRewards, v038dist.ValidatorCurrentRewardsRecord{
			ValidatorAddress: v.ValidatorAddress,
			Rewards: v038dist.NewValidatorCurrentRewards(
				v.Rewards.Rewards,
				v.Rewards.Period,
			),
		})
	}

	startInfos := []v038dist.DelegatorStartingInfoRecord{}
	for _, v := range oldGenState.DelegatorStartingInfos {
		startInfos = append(startInfos, v038dist.DelegatorStartingInfoRecord{
			DelegatorAddress: v.DelegatorAddress,
			ValidatorAddress: v.ValidatorAddress,
			StartingInfo: v038dist.NewDelegatorStartingInfo(
				v.StartingInfo.PreviousPeriod,
				v.StartingInfo.Stake,
				v.StartingInfo.Height,
			),
		})
	}

	slashEvents := []v038dist.ValidatorSlashEventRecord{}
	for _, v := range oldGenState.ValidatorSlashEvents {
		slashEvents = append(slashEvents, v038dist.ValidatorSlashEventRecord{
			ValidatorAddress: v.ValidatorAddress,
			Height:           v.Height,
			Period:           v.Period,
			Event: v038dist.NewValidatorSlashEvent(
				v.Event.ValidatorPeriod,
				v.Event.Fraction,
			),
		})
	}

	newGenState := v038dist.NewGenesisState(
		params,
		v038dist.FeePool(oldGenState.FeePool),
		withdrawInfos,
		oldGenState.PreviousProposer,
		outstandingRewards,
		accumulatedComs,
		histRewards,
		currRewards,
		startInfos,
		slashEvents,
	)

	return newGenState
}
