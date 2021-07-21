package testutil

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/hard"
	hardtypes "github.com/kava-labs/kava/x/hard/types"
	"github.com/kava-labs/kava/x/incentive/types"
)

const (
	oneYear time.Duration = time.Hour * 24 * 365
)

// IncentiveGenesisBuilder is a tool for creating an incentive genesis state.
// Helper methods add values onto a default genesis state.
// All methods are immutable and return updated copies of the builder.
type IncentiveGenesisBuilder struct {
	types.GenesisState
	genesisTime time.Time
}

func NewIncentiveGenesisBuilder() IncentiveGenesisBuilder {
	return IncentiveGenesisBuilder{
		GenesisState: types.DefaultGenesisState(),
		genesisTime:  time.Time{},
	}
}

func (builder IncentiveGenesisBuilder) Build() types.GenesisState {
	return builder.GenesisState
}

func (builder IncentiveGenesisBuilder) BuildMarshalled() app.GenesisState {
	return app.GenesisState{
		types.ModuleName: types.ModuleCdc.MustMarshalJSON(builder.Build()),
	}
}

func (builder IncentiveGenesisBuilder) WithGenesisTime(time time.Time) IncentiveGenesisBuilder {
	builder.genesisTime = time
	builder.Params.ClaimEnd = time.Add(5 * oneYear)
	return builder
}

func (builder IncentiveGenesisBuilder) WithInitializedBorrowRewardPeriod(period types.MultiRewardPeriod) IncentiveGenesisBuilder {
	builder.Params.HardBorrowRewardPeriods = append(builder.Params.HardBorrowRewardPeriods, period)

	accumulationTimeForPeriod := types.NewAccumulationTime(period.CollateralType, builder.genesisTime)
	builder.HardBorrowRewardState.AccumulationTimes = append(
		builder.HardBorrowRewardState.AccumulationTimes,
		accumulationTimeForPeriod,
	)

	builder.HardBorrowRewardState.MultiRewardIndexes = builder.HardBorrowRewardState.MultiRewardIndexes.With(
		period.CollateralType,
		newZeroRewardIndexesFromCoins(period.RewardsPerSecond...),
	)

	return builder
}

func (builder IncentiveGenesisBuilder) WithSimpleBorrowRewardPeriod(ctype string, rewardsPerSecond sdk.Coins) IncentiveGenesisBuilder {
	return builder.WithInitializedBorrowRewardPeriod(builder.simpleRewardPeriod(ctype, rewardsPerSecond))
}

func (builder IncentiveGenesisBuilder) WithInitializedSupplyRewardPeriod(period types.MultiRewardPeriod) IncentiveGenesisBuilder {
	builder.Params.HardSupplyRewardPeriods = append(builder.Params.HardSupplyRewardPeriods, period)

	accumulationTimeForPeriod := types.NewAccumulationTime(period.CollateralType, builder.genesisTime)
	builder.HardSupplyRewardState.AccumulationTimes = append(
		builder.HardSupplyRewardState.AccumulationTimes,
		accumulationTimeForPeriod,
	)

	builder.HardSupplyRewardState.MultiRewardIndexes = builder.HardSupplyRewardState.MultiRewardIndexes.With(
		period.CollateralType,
		newZeroRewardIndexesFromCoins(period.RewardsPerSecond...),
	)

	return builder
}

func (builder IncentiveGenesisBuilder) WithSimpleSupplyRewardPeriod(ctype string, rewardsPerSecond sdk.Coins) IncentiveGenesisBuilder {
	return builder.WithInitializedSupplyRewardPeriod(builder.simpleRewardPeriod(ctype, rewardsPerSecond))
}

func (builder IncentiveGenesisBuilder) WithInitializedDelegatorRewardPeriod(period types.MultiRewardPeriod) IncentiveGenesisBuilder {
	builder.Params.DelegatorRewardPeriods = append(builder.Params.DelegatorRewardPeriods, period)

	accumulationTimeForPeriod := types.NewAccumulationTime(period.CollateralType, builder.genesisTime)
	builder.DelegatorRewardState.AccumulationTimes = append(
		builder.DelegatorRewardState.AccumulationTimes,
		accumulationTimeForPeriod,
	)

	builder.DelegatorRewardState.MultiRewardIndexes = builder.DelegatorRewardState.MultiRewardIndexes.With(
		period.CollateralType,
		newZeroRewardIndexesFromCoins(period.RewardsPerSecond...),
	)

	return builder
}

func (builder IncentiveGenesisBuilder) WithSimpleDelegatorRewardPeriod(ctype string, rewardsPerSecond sdk.Coins) IncentiveGenesisBuilder {
	return builder.WithInitializedDelegatorRewardPeriod(builder.simpleRewardPeriod(ctype, rewardsPerSecond))
}

func (builder IncentiveGenesisBuilder) WithInitializedSwapRewardPeriod(period types.MultiRewardPeriod) IncentiveGenesisBuilder {
	builder.Params.SwapRewardPeriods = append(builder.Params.SwapRewardPeriods, period)

	accumulationTimeForPeriod := types.NewAccumulationTime(period.CollateralType, builder.genesisTime)
	builder.SwapRewardState.AccumulationTimes = append(
		builder.SwapRewardState.AccumulationTimes,
		accumulationTimeForPeriod,
	)

	builder.SwapRewardState.MultiRewardIndexes = builder.SwapRewardState.MultiRewardIndexes.With(
		period.CollateralType,
		newZeroRewardIndexesFromCoins(period.RewardsPerSecond...),
	)

	return builder
}

func (builder IncentiveGenesisBuilder) WithSimpleSwapRewardPeriod(poolID string, rewardsPerSecond sdk.Coins) IncentiveGenesisBuilder {
	return builder.WithInitializedSwapRewardPeriod(builder.simpleRewardPeriod(poolID, rewardsPerSecond))
}

func (builder IncentiveGenesisBuilder) WithInitializedUSDXRewardPeriod(period types.RewardPeriod) IncentiveGenesisBuilder {
	builder.Params.USDXMintingRewardPeriods = append(builder.Params.USDXMintingRewardPeriods, period)

	accumulationTimeForPeriod := types.NewAccumulationTime(period.CollateralType, builder.genesisTime)
	builder.USDXRewardState.AccumulationTimes = append(
		builder.USDXRewardState.AccumulationTimes,
		accumulationTimeForPeriod,
	)

	builder.USDXRewardState.MultiRewardIndexes = builder.USDXRewardState.MultiRewardIndexes.With(
		period.CollateralType,
		newZeroRewardIndexesFromCoins(period.RewardsPerSecond),
	)

	return builder
}

func (builder IncentiveGenesisBuilder) WithSimpleUSDXRewardPeriod(ctype string, rewardsPerSecond sdk.Coin) IncentiveGenesisBuilder {
	return builder.WithInitializedUSDXRewardPeriod(types.NewRewardPeriod(
		true,
		ctype,
		builder.genesisTime,
		builder.genesisTime.Add(4*oneYear),
		rewardsPerSecond,
	))
}

func (builder IncentiveGenesisBuilder) WithMultipliersOLD(multipliers types.Multipliers) IncentiveGenesisBuilder {
	builder.Params.ClaimMultipliersOLD = multipliers
	return builder
}

func (builder IncentiveGenesisBuilder) WithMultipliers(multipliers types.MultipliersPerDenom) IncentiveGenesisBuilder {
	builder.Params.ClaimMultipliers = multipliers
	return builder
}

func (builder IncentiveGenesisBuilder) simpleRewardPeriod(ctype string, rewardsPerSecond sdk.Coins) types.MultiRewardPeriod {
	return types.NewMultiRewardPeriod(
		true,
		ctype,
		builder.genesisTime,
		builder.genesisTime.Add(4*oneYear),
		rewardsPerSecond,
	)
}

func newZeroRewardIndexesFromCoins(coins ...sdk.Coin) types.RewardIndexes {
	var ri types.RewardIndexes
	for _, coin := range coins {
		ri = ri.With(coin.Denom, sdk.ZeroDec())
	}
	return ri
}

// HardGenesisBuilder is a tool for creating a hard genesis state.
// Helper methods add values onto a default genesis state.
// All methods are immutable and return updated copies of the builder.
type HardGenesisBuilder struct {
	hardtypes.GenesisState
	genesisTime time.Time
}

func NewHardGenesisBuilder() HardGenesisBuilder {
	return HardGenesisBuilder{
		GenesisState: hardtypes.DefaultGenesisState(),
	}
}
func (builder HardGenesisBuilder) Build() hardtypes.GenesisState {
	return builder.GenesisState
}
func (builder HardGenesisBuilder) BuildMarshalled() app.GenesisState {
	return app.GenesisState{
		hardtypes.ModuleName: hardtypes.ModuleCdc.MustMarshalJSON(builder.Build()),
	}
}
func (builder HardGenesisBuilder) WithGenesisTime(genTime time.Time) HardGenesisBuilder {
	builder.genesisTime = genTime
	return builder
}
func (builder HardGenesisBuilder) WithInitializedMoneyMarket(market hard.MoneyMarket) HardGenesisBuilder {
	builder.Params.MoneyMarkets = append(builder.Params.MoneyMarkets, market)

	builder.PreviousAccumulationTimes = append(
		builder.PreviousAccumulationTimes,
		hardtypes.NewGenesisAccumulationTime(market.Denom, builder.genesisTime, sdk.OneDec(), sdk.OneDec()),
	)
	return builder
}
func (builder HardGenesisBuilder) WithMinBorrow(minUSDValue sdk.Dec) HardGenesisBuilder {
	builder.Params.MinimumBorrowUSDValue = minUSDValue
	return builder
}
func NewStandardMoneyMarket(denom string) hardtypes.MoneyMarket {
	return hardtypes.NewMoneyMarket(
		denom,
		hard.NewBorrowLimit(
			false,
			sdk.NewDec(1e15),
			sdk.MustNewDecFromStr("0.6"),
		),
		denom+":usd",
		sdk.NewInt(1e6),
		hard.NewInterestRateModel(
			sdk.MustNewDecFromStr("0.05"),
			sdk.MustNewDecFromStr("2"),
			sdk.MustNewDecFromStr("0.8"),
			sdk.MustNewDecFromStr("10"),
		),
		sdk.MustNewDecFromStr("0.05"),
		sdk.ZeroDec(),
	)
}
