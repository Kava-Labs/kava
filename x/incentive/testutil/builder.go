package testutil

import (
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/app"
	hardtypes "github.com/kava-labs/kava/x/hard/types"
	"github.com/kava-labs/kava/x/incentive/types"
	savingstypes "github.com/kava-labs/kava/x/savings/types"
)

const (
	oneYear time.Duration = time.Hour * 24 * 365
)

type GenesisBuilder interface {
	BuildMarshalled(cdc codec.JSONCodec) app.GenesisState
}

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

func (builder IncentiveGenesisBuilder) BuildMarshalled(cdc codec.JSONCodec) app.GenesisState {
	built := builder.Build()

	return app.GenesisState{
		types.ModuleName: cdc.MustMarshalJSON(&built),
	}
}

func (builder IncentiveGenesisBuilder) WithGenesisTime(time time.Time) IncentiveGenesisBuilder {
	builder.genesisTime = time
	builder.Params.ClaimEnd = time.Add(5 * oneYear)
	return builder
}

// WithInitializedBorrowRewardPeriod sets the genesis time as the previous accumulation time for the specified period.
// This can be helpful in tests. With no prev time set, the first block accrues no rewards as it just sets the prev time to the current.
func (builder IncentiveGenesisBuilder) WithInitializedBorrowRewardPeriod(period types.MultiRewardPeriod) IncentiveGenesisBuilder {
	builder.Params.HardBorrowRewardPeriods = append(builder.Params.HardBorrowRewardPeriods, period)

	accumulationTimeForPeriod := types.NewAccumulationTime(period.CollateralType, builder.genesisTime)
	builder.HardBorrowRewardState.AccumulationTimes = append(
		builder.HardBorrowRewardState.AccumulationTimes,
		accumulationTimeForPeriod,
	)

	// TODO remove to better reflect real states
	builder.HardBorrowRewardState.MultiRewardIndexes = builder.HardBorrowRewardState.MultiRewardIndexes.With(
		period.CollateralType,
		newZeroRewardIndexesFromCoins(period.RewardsPerSecond...),
	)

	return builder
}

func (builder IncentiveGenesisBuilder) WithSimpleBorrowRewardPeriod(ctype string, rewardsPerSecond sdk.Coins) IncentiveGenesisBuilder {
	return builder.WithInitializedBorrowRewardPeriod(builder.simpleRewardPeriod(ctype, rewardsPerSecond))
}

// WithInitializedSupplyRewardPeriod sets the genesis time as the previous accumulation time for the specified period.
// This can be helpful in tests. With no prev time set, the first block accrues no rewards as it just sets the prev time to the current.
func (builder IncentiveGenesisBuilder) WithInitializedSupplyRewardPeriod(period types.MultiRewardPeriod) IncentiveGenesisBuilder {
	builder.Params.HardSupplyRewardPeriods = append(builder.Params.HardSupplyRewardPeriods, period)

	accumulationTimeForPeriod := types.NewAccumulationTime(period.CollateralType, builder.genesisTime)
	builder.HardSupplyRewardState.AccumulationTimes = append(
		builder.HardSupplyRewardState.AccumulationTimes,
		accumulationTimeForPeriod,
	)

	// TODO remove to better reflect real states
	builder.HardSupplyRewardState.MultiRewardIndexes = builder.HardSupplyRewardState.MultiRewardIndexes.With(
		period.CollateralType,
		newZeroRewardIndexesFromCoins(period.RewardsPerSecond...),
	)

	return builder
}

func (builder IncentiveGenesisBuilder) WithSimpleSupplyRewardPeriod(ctype string, rewardsPerSecond sdk.Coins) IncentiveGenesisBuilder {
	return builder.WithInitializedSupplyRewardPeriod(builder.simpleRewardPeriod(ctype, rewardsPerSecond))
}

// WithInitializedDelegatorRewardPeriod sets the genesis time as the previous accumulation time for the specified period.
// This can be helpful in tests. With no prev time set, the first block accrues no rewards as it just sets the prev time to the current.
func (builder IncentiveGenesisBuilder) WithInitializedDelegatorRewardPeriod(period types.MultiRewardPeriod) IncentiveGenesisBuilder {
	builder.Params.DelegatorRewardPeriods = append(builder.Params.DelegatorRewardPeriods, period)

	accumulationTimeForPeriod := types.NewAccumulationTime(period.CollateralType, builder.genesisTime)
	builder.DelegatorRewardState.AccumulationTimes = append(
		builder.DelegatorRewardState.AccumulationTimes,
		accumulationTimeForPeriod,
	)

	// TODO remove to better reflect real states
	builder.DelegatorRewardState.MultiRewardIndexes = builder.DelegatorRewardState.MultiRewardIndexes.With(
		period.CollateralType,
		newZeroRewardIndexesFromCoins(period.RewardsPerSecond...),
	)

	return builder
}

func (builder IncentiveGenesisBuilder) WithSimpleDelegatorRewardPeriod(ctype string, rewardsPerSecond sdk.Coins) IncentiveGenesisBuilder {
	return builder.WithInitializedDelegatorRewardPeriod(builder.simpleRewardPeriod(ctype, rewardsPerSecond))
}

// WithInitializedSwapRewardPeriod sets the genesis time as the previous accumulation time for the specified period.
// This can be helpful in tests. With no prev time set, the first block accrues no rewards as it just sets the prev time to the current.
func (builder IncentiveGenesisBuilder) WithInitializedSwapRewardPeriod(period types.MultiRewardPeriod) IncentiveGenesisBuilder {
	builder.Params.SwapRewardPeriods = append(builder.Params.SwapRewardPeriods, period)

	accumulationTimeForPeriod := types.NewAccumulationTime(period.CollateralType, builder.genesisTime)
	builder.SwapRewardState.AccumulationTimes = append(
		builder.SwapRewardState.AccumulationTimes,
		accumulationTimeForPeriod,
	)

	return builder
}

func (builder IncentiveGenesisBuilder) WithSimpleSwapRewardPeriod(poolID string, rewardsPerSecond sdk.Coins) IncentiveGenesisBuilder {
	return builder.WithInitializedSwapRewardPeriod(builder.simpleRewardPeriod(poolID, rewardsPerSecond))
}

// WithInitializedUSDXRewardPeriod sets the genesis time as the previous accumulation time for the specified period.
// This can be helpful in tests. With no prev time set, the first block accrues no rewards as it just sets the prev time to the current.
func (builder IncentiveGenesisBuilder) WithInitializedUSDXRewardPeriod(period types.RewardPeriod) IncentiveGenesisBuilder {
	builder.Params.USDXMintingRewardPeriods = append(builder.Params.USDXMintingRewardPeriods, period)

	accumulationTimeForPeriod := types.NewAccumulationTime(period.CollateralType, builder.genesisTime)
	builder.USDXRewardState.AccumulationTimes = append(
		builder.USDXRewardState.AccumulationTimes,
		accumulationTimeForPeriod,
	)

	// TODO remove to better reflect real states
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

// WithInitializedEarnRewardPeriod sets the genesis time as the previous accumulation time for the specified period.
// This can be helpful in tests. With no prev time set, the first block accrues no rewards as it just sets the prev time to the current.
func (builder IncentiveGenesisBuilder) WithInitializedEarnRewardPeriod(period types.MultiRewardPeriod) IncentiveGenesisBuilder {
	builder.Params.EarnRewardPeriods = append(builder.Params.EarnRewardPeriods, period)

	accumulationTimeForPeriod := types.NewAccumulationTime(period.CollateralType, builder.genesisTime)
	builder.EarnRewardState.AccumulationTimes = append(
		builder.EarnRewardState.AccumulationTimes,
		accumulationTimeForPeriod,
	)

	return builder
}

func (builder IncentiveGenesisBuilder) WithSimpleEarnRewardPeriod(ctype string, rewardsPerSecond sdk.Coins) IncentiveGenesisBuilder {
	return builder.WithInitializedEarnRewardPeriod(builder.simpleRewardPeriod(ctype, rewardsPerSecond))
}

func (builder IncentiveGenesisBuilder) WithMultipliers(multipliers types.MultipliersPerDenoms) IncentiveGenesisBuilder {
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

func (builder HardGenesisBuilder) BuildMarshalled(cdc codec.JSONCodec) app.GenesisState {
	built := builder.Build()

	return app.GenesisState{
		hardtypes.ModuleName: cdc.MustMarshalJSON(&built),
	}
}

func (builder HardGenesisBuilder) WithGenesisTime(genTime time.Time) HardGenesisBuilder {
	builder.genesisTime = genTime
	return builder
}

func (builder HardGenesisBuilder) WithInitializedMoneyMarket(market hardtypes.MoneyMarket) HardGenesisBuilder {
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
		hardtypes.NewBorrowLimit(
			false,
			sdk.NewDec(1e15),
			sdk.MustNewDecFromStr("0.6"),
		),
		denom+":usd",
		sdk.NewInt(1e6),
		hardtypes.NewInterestRateModel(
			sdk.MustNewDecFromStr("0.05"),
			sdk.MustNewDecFromStr("2"),
			sdk.MustNewDecFromStr("0.8"),
			sdk.MustNewDecFromStr("10"),
		),
		sdk.MustNewDecFromStr("0.05"),
		sdk.ZeroDec(),
	)
}

// WithInitializedSavingsRewardPeriod sets the genesis time as the previous accumulation time for the specified period.
// This can be helpful in tests. With no prev time set, the first block accrues no rewards as it just sets the prev time to the current.
func (builder IncentiveGenesisBuilder) WithInitializedSavingsRewardPeriod(period types.MultiRewardPeriod) IncentiveGenesisBuilder {
	builder.Params.SavingsRewardPeriods = append(builder.Params.SavingsRewardPeriods, period)

	accumulationTimeForPeriod := types.NewAccumulationTime(period.CollateralType, builder.genesisTime)
	builder.SavingsRewardState.AccumulationTimes = append(
		builder.SavingsRewardState.AccumulationTimes,
		accumulationTimeForPeriod,
	)

	builder.SavingsRewardState.MultiRewardIndexes = builder.SavingsRewardState.MultiRewardIndexes.With(
		period.CollateralType,
		newZeroRewardIndexesFromCoins(period.RewardsPerSecond...),
	)

	return builder
}

func (builder IncentiveGenesisBuilder) WithSimpleSavingsRewardPeriod(ctype string, rewardsPerSecond sdk.Coins) IncentiveGenesisBuilder {
	return builder.WithInitializedSavingsRewardPeriod(builder.simpleRewardPeriod(ctype, rewardsPerSecond))
}

// SavingsGenesisBuilder is a tool for creating a savings genesis state.
// Helper methods add values onto a default genesis state.
// All methods are immutable and return updated copies of the builder.
type SavingsGenesisBuilder struct {
	savingstypes.GenesisState
	genesisTime time.Time
}

func NewSavingsGenesisBuilder() SavingsGenesisBuilder {
	return SavingsGenesisBuilder{
		GenesisState: savingstypes.DefaultGenesisState(),
	}
}

func (builder SavingsGenesisBuilder) Build() savingstypes.GenesisState {
	return builder.GenesisState
}

func (builder SavingsGenesisBuilder) BuildMarshalled(cdc codec.JSONCodec) app.GenesisState {
	built := builder.Build()

	return app.GenesisState{
		savingstypes.ModuleName: cdc.MustMarshalJSON(&built),
	}
}

func (builder SavingsGenesisBuilder) WithGenesisTime(genTime time.Time) SavingsGenesisBuilder {
	builder.genesisTime = genTime
	return builder
}

func (builder SavingsGenesisBuilder) WithSupportedDenoms(denoms ...string) SavingsGenesisBuilder {
	builder.Params.SupportedDenoms = append(builder.Params.SupportedDenoms, denoms...)
	return builder
}
