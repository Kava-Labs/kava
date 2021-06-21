package keeper_test

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/cdp"
	committeetypes "github.com/kava-labs/kava/x/committee/types"
	"github.com/kava-labs/kava/x/hard"
	hardtypes "github.com/kava-labs/kava/x/hard/types"
	"github.com/kava-labs/kava/x/incentive/types"
	"github.com/kava-labs/kava/x/pricefeed"
)

const (
	oneYear time.Duration = time.Hour * 24 * 365
)

// Avoid cluttering test cases with long function names
func i(in int64) sdk.Int                    { return sdk.NewInt(in) }
func d(str string) sdk.Dec                  { return sdk.MustNewDecFromStr(str) }
func c(denom string, amount int64) sdk.Coin { return sdk.NewInt64Coin(denom, amount) }
func cs(coins ...sdk.Coin) sdk.Coins        { return sdk.NewCoins(coins...) }

func NewCDPGenStateMulti() app.GenesisState {
	cdpGenesis := cdp.GenesisState{
		Params: cdp.Params{
			GlobalDebtLimit:         sdk.NewInt64Coin("usdx", 2000000000000),
			SurplusAuctionThreshold: cdp.DefaultSurplusThreshold,
			SurplusAuctionLot:       cdp.DefaultSurplusLot,
			DebtAuctionThreshold:    cdp.DefaultDebtThreshold,
			DebtAuctionLot:          cdp.DefaultDebtLot,
			CollateralParams: cdp.CollateralParams{
				{
					Denom:               "xrp",
					Type:                "xrp-a",
					LiquidationRatio:    sdk.MustNewDecFromStr("2.0"),
					DebtLimit:           sdk.NewInt64Coin("usdx", 500000000000),
					StabilityFee:        sdk.MustNewDecFromStr("1.000000001547125958"), // %5 apr
					LiquidationPenalty:  d("0.05"),
					AuctionSize:         i(7000000000),
					Prefix:              0x20,
					SpotMarketID:        "xrp:usd",
					LiquidationMarketID: "xrp:usd",
					ConversionFactor:    i(6),
				},
				{
					Denom:               "btc",
					Type:                "btc-a",
					LiquidationRatio:    sdk.MustNewDecFromStr("1.5"),
					DebtLimit:           sdk.NewInt64Coin("usdx", 500000000000),
					StabilityFee:        sdk.MustNewDecFromStr("1.000000000782997609"), // %2.5 apr
					LiquidationPenalty:  d("0.025"),
					AuctionSize:         i(10000000),
					Prefix:              0x21,
					SpotMarketID:        "btc:usd",
					LiquidationMarketID: "btc:usd",
					ConversionFactor:    i(8),
				},
				{
					Denom:               "bnb",
					Type:                "bnb-a",
					LiquidationRatio:    sdk.MustNewDecFromStr("1.5"),
					DebtLimit:           sdk.NewInt64Coin("usdx", 500000000000),
					StabilityFee:        sdk.MustNewDecFromStr("1.000000001547125958"), // %5 apr
					LiquidationPenalty:  d("0.05"),
					AuctionSize:         i(50000000000),
					Prefix:              0x22,
					SpotMarketID:        "bnb:usd",
					LiquidationMarketID: "bnb:usd",
					ConversionFactor:    i(8),
				},
				{
					Denom:               "busd",
					Type:                "busd-a",
					LiquidationRatio:    d("1.01"),
					DebtLimit:           sdk.NewInt64Coin("usdx", 500000000000),
					StabilityFee:        sdk.OneDec(), // %0 apr
					LiquidationPenalty:  d("0.05"),
					AuctionSize:         i(10000000000),
					Prefix:              0x23,
					SpotMarketID:        "busd:usd",
					LiquidationMarketID: "busd:usd",
					ConversionFactor:    i(8),
				},
			},
			DebtParam: cdp.DebtParam{
				Denom:            "usdx",
				ReferenceAsset:   "usd",
				ConversionFactor: i(6),
				DebtFloor:        i(10000000),
			},
		},
		StartingCdpID: cdp.DefaultCdpStartingID,
		DebtDenom:     cdp.DefaultDebtDenom,
		GovDenom:      cdp.DefaultGovDenom,
		CDPs:          cdp.CDPs{},
		PreviousAccumulationTimes: cdp.GenesisAccumulationTimes{
			cdp.NewGenesisAccumulationTime("btc-a", time.Time{}, sdk.OneDec()),
			cdp.NewGenesisAccumulationTime("xrp-a", time.Time{}, sdk.OneDec()),
			cdp.NewGenesisAccumulationTime("busd-a", time.Time{}, sdk.OneDec()),
			cdp.NewGenesisAccumulationTime("bnb-a", time.Time{}, sdk.OneDec()),
		},
		TotalPrincipals: cdp.GenesisTotalPrincipals{
			cdp.NewGenesisTotalPrincipal("btc-a", sdk.ZeroInt()),
			cdp.NewGenesisTotalPrincipal("xrp-a", sdk.ZeroInt()),
			cdp.NewGenesisTotalPrincipal("busd-a", sdk.ZeroInt()),
			cdp.NewGenesisTotalPrincipal("bnb-a", sdk.ZeroInt()),
		},
	}
	return app.GenesisState{cdp.ModuleName: cdp.ModuleCdc.MustMarshalJSON(cdpGenesis)}
}

func NewPricefeedGenStateMultiFromTime(t time.Time) app.GenesisState {
	pfGenesis := pricefeed.GenesisState{
		Params: pricefeed.Params{
			Markets: []pricefeed.Market{
				{MarketID: "kava:usd", BaseAsset: "kava", QuoteAsset: "usd", Oracles: []sdk.AccAddress{}, Active: true},
				{MarketID: "btc:usd", BaseAsset: "btc", QuoteAsset: "usd", Oracles: []sdk.AccAddress{}, Active: true},
				{MarketID: "xrp:usd", BaseAsset: "xrp", QuoteAsset: "usd", Oracles: []sdk.AccAddress{}, Active: true},
				{MarketID: "bnb:usd", BaseAsset: "bnb", QuoteAsset: "usd", Oracles: []sdk.AccAddress{}, Active: true},
				{MarketID: "busd:usd", BaseAsset: "busd", QuoteAsset: "usd", Oracles: []sdk.AccAddress{}, Active: true},
				{MarketID: "zzz:usd", BaseAsset: "zzz", QuoteAsset: "usd", Oracles: []sdk.AccAddress{}, Active: true},
			},
		},
		PostedPrices: []pricefeed.PostedPrice{
			{
				MarketID:      "kava:usd",
				OracleAddress: sdk.AccAddress{},
				Price:         sdk.MustNewDecFromStr("2.00"),
				Expiry:        t.Add(1 * time.Hour),
			},
			{
				MarketID:      "btc:usd",
				OracleAddress: sdk.AccAddress{},
				Price:         sdk.MustNewDecFromStr("8000.00"),
				Expiry:        t.Add(1 * time.Hour),
			},
			{
				MarketID:      "xrp:usd",
				OracleAddress: sdk.AccAddress{},
				Price:         sdk.MustNewDecFromStr("0.25"),
				Expiry:        t.Add(1 * time.Hour),
			},
			{
				MarketID:      "bnb:usd",
				OracleAddress: sdk.AccAddress{},
				Price:         sdk.MustNewDecFromStr("17.25"),
				Expiry:        t.Add(1 * time.Hour),
			},
			{
				MarketID:      "busd:usd",
				OracleAddress: sdk.AccAddress{},
				Price:         sdk.OneDec(),
				Expiry:        t.Add(1 * time.Hour),
			},
			{
				MarketID:      "zzz:usd",
				OracleAddress: sdk.AccAddress{},
				Price:         sdk.MustNewDecFromStr("2.00"),
				Expiry:        t.Add(1 * time.Hour),
			},
		},
	}
	return app.GenesisState{pricefeed.ModuleName: pricefeed.ModuleCdc.MustMarshalJSON(pfGenesis)}
}

func NewHardGenStateMulti(genTime time.Time) HardGenesisBuilder {
	kavaMM := NewStandardMoneyMarket("ukava")
	kavaMM.SpotMarketID = "kava:usd"
	btcMM := NewStandardMoneyMarket("btcb")
	btcMM.SpotMarketID = "btc:usd"

	builder := NewHardGenesisBuilder().WithGenesisTime(genTime).
		WithInitializedMoneyMarket(NewStandardMoneyMarket("usdx")).
		WithInitializedMoneyMarket(kavaMM).
		WithInitializedMoneyMarket(NewStandardMoneyMarket("bnb")).
		WithInitializedMoneyMarket(btcMM).
		WithInitializedMoneyMarket(NewStandardMoneyMarket("xrp")).
		WithInitializedMoneyMarket(NewStandardMoneyMarket("zzz"))
	return builder
}

func NewStakingGenesisState() app.GenesisState {
	genState := staking.DefaultGenesisState()
	genState.Params.BondDenom = "ukava"
	return app.GenesisState{
		staking.ModuleName: staking.ModuleCdc.MustMarshalJSON(genState),
	}
}

func NewCommitteeGenesisState(members []sdk.AccAddress) app.GenesisState {
	genState := committeetypes.DefaultGenesisState()
	genState.Committees = committeetypes.Committees{
		committeetypes.MemberCommittee{
			BaseCommittee: committeetypes.BaseCommittee{
				ID:               genState.NextProposalID,
				Description:      "This committee is for testing.",
				Members:          members,
				Permissions:      []committeetypes.Permission{committeetypes.GodPermission{}},
				VoteThreshold:    d("0.667"),
				ProposalDuration: time.Hour * 24 * 7,
				TallyOption:      committeetypes.FirstPastThePost,
			},
		},
	}
	genState.NextProposalID += 1
	return app.GenesisState{
		committeetypes.ModuleName: committeetypes.ModuleCdc.MustMarshalJSON(genState),
	}
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

	accumulationTimeForPeriod := types.NewGenesisAccumulationTime(period.CollateralType, builder.genesisTime)
	builder.HardBorrowAccumulationTimes = append(builder.HardBorrowAccumulationTimes, accumulationTimeForPeriod)
	return builder
}

func (builder IncentiveGenesisBuilder) WithSimpleBorrowRewardPeriod(ctype string, rewardsPerSecond sdk.Coins) IncentiveGenesisBuilder {
	return builder.WithInitializedBorrowRewardPeriod(types.NewMultiRewardPeriod(
		true,
		ctype,
		builder.genesisTime,
		builder.genesisTime.Add(4*oneYear),
		rewardsPerSecond,
	))
}
func (builder IncentiveGenesisBuilder) WithInitializedSupplyRewardPeriod(period types.MultiRewardPeriod) IncentiveGenesisBuilder {
	builder.Params.HardSupplyRewardPeriods = append(builder.Params.HardSupplyRewardPeriods, period)

	accumulationTimeForPeriod := types.NewGenesisAccumulationTime(period.CollateralType, builder.genesisTime)
	builder.HardSupplyAccumulationTimes = append(builder.HardSupplyAccumulationTimes, accumulationTimeForPeriod)
	return builder
}

func (builder IncentiveGenesisBuilder) WithSimpleSupplyRewardPeriod(ctype string, rewardsPerSecond sdk.Coins) IncentiveGenesisBuilder {
	return builder.WithInitializedSupplyRewardPeriod(types.NewMultiRewardPeriod(
		true,
		ctype,
		builder.genesisTime,
		builder.genesisTime.Add(4*oneYear),
		rewardsPerSecond,
	))
}
func (builder IncentiveGenesisBuilder) WithInitializedDelegatorRewardPeriod(period types.RewardPeriod) IncentiveGenesisBuilder {
	builder.Params.HardDelegatorRewardPeriods = append(builder.Params.HardDelegatorRewardPeriods, period)

	accumulationTimeForPeriod := types.NewGenesisAccumulationTime(period.CollateralType, builder.genesisTime)
	builder.HardDelegatorAccumulationTimes = append(builder.HardDelegatorAccumulationTimes, accumulationTimeForPeriod)
	return builder
}

func (builder IncentiveGenesisBuilder) WithSimpleDelegatorRewardPeriod(ctype string, rewardsPerSecond sdk.Coin) IncentiveGenesisBuilder {
	return builder.WithInitializedDelegatorRewardPeriod(types.NewRewardPeriod(
		true,
		ctype,
		builder.genesisTime,
		builder.genesisTime.Add(4*oneYear),
		rewardsPerSecond,
	))
}
func (builder IncentiveGenesisBuilder) WithInitializedUSDXRewardPeriod(period types.RewardPeriod) IncentiveGenesisBuilder {
	builder.Params.USDXMintingRewardPeriods = append(builder.Params.USDXMintingRewardPeriods, period)

	accumulationTimeForPeriod := types.NewGenesisAccumulationTime(period.CollateralType, builder.genesisTime)
	builder.USDXAccumulationTimes = append(builder.USDXAccumulationTimes, accumulationTimeForPeriod)
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

func (builder IncentiveGenesisBuilder) WithMultipliers(multipliers types.Multipliers) IncentiveGenesisBuilder {
	builder.Params.ClaimMultipliers = multipliers
	return builder
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
			d("0.6"),
		),
		denom+":usd",
		i(1e6),
		hard.NewInterestRateModel(d("0.05"), d("2"), d("0.8"), d("10")),
		d("0.05"),
		sdk.ZeroDec(),
	)
}
