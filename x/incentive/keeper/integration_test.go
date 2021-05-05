package keeper_test

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authexported "github.com/cosmos/cosmos-sdk/x/auth/exported"
	"github.com/cosmos/cosmos-sdk/x/auth/vesting"
	"github.com/cosmos/cosmos-sdk/x/staking"
	"github.com/cosmos/cosmos-sdk/x/supply"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/cdp"
	committeetypes "github.com/kava-labs/kava/x/committee/types"
	"github.com/kava-labs/kava/x/hard"
	"github.com/kava-labs/kava/x/incentive/types"
	"github.com/kava-labs/kava/x/pricefeed"
	validatorvesting "github.com/kava-labs/kava/x/validator-vesting"
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

func NewPricefeedGenStateMulti() app.GenesisState {
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
				Expiry:        time.Now().Add(1 * time.Hour),
			},
			{
				MarketID:      "btc:usd",
				OracleAddress: sdk.AccAddress{},
				Price:         sdk.MustNewDecFromStr("8000.00"),
				Expiry:        time.Now().Add(1 * time.Hour),
			},
			{
				MarketID:      "xrp:usd",
				OracleAddress: sdk.AccAddress{},
				Price:         sdk.MustNewDecFromStr("0.25"),
				Expiry:        time.Now().Add(1 * time.Hour),
			},
			{
				MarketID:      "bnb:usd",
				OracleAddress: sdk.AccAddress{},
				Price:         sdk.MustNewDecFromStr("17.25"),
				Expiry:        time.Now().Add(1 * time.Hour),
			},
			{
				MarketID:      "busd:usd",
				OracleAddress: sdk.AccAddress{},
				Price:         sdk.OneDec(),
				Expiry:        time.Now().Add(1 * time.Hour),
			},
			{
				MarketID:      "zzz:usd",
				OracleAddress: sdk.AccAddress{},
				Price:         sdk.MustNewDecFromStr("2.00"),
				Expiry:        time.Now().Add(1 * time.Hour),
			},
		},
	}
	return app.GenesisState{pricefeed.ModuleName: pricefeed.ModuleCdc.MustMarshalJSON(pfGenesis)}
}

func NewHardGenStateMulti() app.GenesisState {
	loanToValue, _ := sdk.NewDecFromStr("0.6")
	borrowLimit := sdk.NewDec(1000000000000000)

	hardGS := hard.NewGenesisState(hard.NewParams(
		hard.MoneyMarkets{
			hard.NewMoneyMarket("usdx", hard.NewBorrowLimit(false, borrowLimit, loanToValue), "usdx:usd", sdk.NewInt(1000000), hard.NewInterestRateModel(sdk.MustNewDecFromStr("0.05"), sdk.MustNewDecFromStr("2"), sdk.MustNewDecFromStr("0.8"), sdk.MustNewDecFromStr("10")), sdk.MustNewDecFromStr("0.05"), sdk.ZeroDec()),
			hard.NewMoneyMarket("ukava", hard.NewBorrowLimit(false, borrowLimit, loanToValue), "kava:usd", sdk.NewInt(1000000), hard.NewInterestRateModel(sdk.MustNewDecFromStr("0.05"), sdk.MustNewDecFromStr("2"), sdk.MustNewDecFromStr("0.8"), sdk.MustNewDecFromStr("10")), sdk.MustNewDecFromStr("0.05"), sdk.ZeroDec()),
			hard.NewMoneyMarket("bnb", hard.NewBorrowLimit(false, borrowLimit, loanToValue), "bnb:usd", sdk.NewInt(1000000), hard.NewInterestRateModel(sdk.MustNewDecFromStr("0.05"), sdk.MustNewDecFromStr("2"), sdk.MustNewDecFromStr("0.8"), sdk.MustNewDecFromStr("10")), sdk.MustNewDecFromStr("0.05"), sdk.ZeroDec()),
			hard.NewMoneyMarket("btcb", hard.NewBorrowLimit(false, borrowLimit, loanToValue), "btc:usd", sdk.NewInt(1000000), hard.NewInterestRateModel(sdk.MustNewDecFromStr("0.05"), sdk.MustNewDecFromStr("2"), sdk.MustNewDecFromStr("0.8"), sdk.MustNewDecFromStr("10")), sdk.MustNewDecFromStr("0.05"), sdk.ZeroDec()),
			hard.NewMoneyMarket("xrp", hard.NewBorrowLimit(false, borrowLimit, loanToValue), "xrp:usd", sdk.NewInt(1000000), hard.NewInterestRateModel(sdk.MustNewDecFromStr("0.05"), sdk.MustNewDecFromStr("2"), sdk.MustNewDecFromStr("0.8"), sdk.MustNewDecFromStr("10")), sdk.MustNewDecFromStr("0.05"), sdk.ZeroDec()),
			hard.NewMoneyMarket("zzz", hard.NewBorrowLimit(false, borrowLimit, loanToValue), "zzz:usd", sdk.NewInt(1000000), hard.NewInterestRateModel(sdk.MustNewDecFromStr("0.05"), sdk.MustNewDecFromStr("2"), sdk.MustNewDecFromStr("0.8"), sdk.MustNewDecFromStr("10")), sdk.MustNewDecFromStr("0.05"), sdk.ZeroDec()),
		},
		sdk.NewDec(10),
	), hard.DefaultAccumulationTimes, hard.DefaultDeposits, hard.DefaultBorrows,
		hard.DefaultTotalSupplied, hard.DefaultTotalBorrowed, hard.DefaultTotalReserves,
	)

	return app.GenesisState{hard.ModuleName: hard.ModuleCdc.MustMarshalJSON(hardGS)}
}

// AuthGenesisBuilder creates an auth genesis state by building it up a default value.
// Helper methods to create basic accounts types and add them to the genesis state.
// All methods are immutable and return updated copies of the builder.
//
// Example:
//     // create a single account genesis state
//     builder := NewAuthGenesisBuilder().WithSimpleAccount(testUserAddress, testCoins)
//     genesisState := builder.Build()
//
type AuthGenesisBuilder struct {
	genesis auth.GenesisState
}

func NewAuthGenesisBuilder() AuthGenesisBuilder {
	return AuthGenesisBuilder{
		genesis: auth.DefaultGenesisState(),
	}
}

// Build assembles and returns the final GenesisState
func (builder AuthGenesisBuilder) Build() auth.GenesisState {
	return builder.genesis
}

// BuildMarshalled assembles the final GenesisState and json encodes it into a universal genesis type.
func (builder AuthGenesisBuilder) BuildMarshalled() app.GenesisState {
	return app.GenesisState{
		auth.ModuleName: auth.ModuleCdc.MustMarshalJSON(builder.Build()),
	}
}

func (builder AuthGenesisBuilder) WithParams(params auth.Params) AuthGenesisBuilder {
	builder.genesis.Params = params
	return builder
}

// WithAccounts adds accounts of any type to the genesis state.
func (builder AuthGenesisBuilder) WithAccounts(account ...authexported.GenesisAccount) AuthGenesisBuilder {
	builder.genesis.Accounts = append(builder.genesis.Accounts, account...)
	return builder
}

// WithSimpleAccount adds a standard account to the genesis state.
func (builder AuthGenesisBuilder) WithSimpleAccount(address sdk.AccAddress, balance sdk.Coins) AuthGenesisBuilder {
	return builder.WithAccounts(auth.NewBaseAccount(address, balance, nil, 0, 0))
}

func (builder AuthGenesisBuilder) WithSimpleModuleAccount(moduleName string, balance sdk.Coins, permissions ...string) AuthGenesisBuilder {
	account := supply.NewEmptyModuleAccount(moduleName, permissions...)
	account.SetCoins(balance)
	return builder.WithAccounts(account)
}

func (builder AuthGenesisBuilder) WithSimplePeriodicVestingAccount(address sdk.AccAddress, balance sdk.Coins, periods vesting.Periods, firstPeriodStartTimestamp int64) AuthGenesisBuilder {
	baseAccount := auth.NewBaseAccount(address, balance, nil, 0, 0)

	originalVesting := sdk.NewCoins()
	for _, p := range periods {
		originalVesting = originalVesting.Add(p.Amount...)
	}

	var totalPeriods int64
	for _, p := range periods {
		totalPeriods += p.Length
	}
	endTime := firstPeriodStartTimestamp + totalPeriods

	baseVestingAccount, err := vesting.NewBaseVestingAccount(baseAccount, originalVesting, endTime)
	if err != nil {
		panic(err.Error())
	}
	periodicVestingAccount := vesting.NewPeriodicVestingAccountRaw(baseVestingAccount, firstPeriodStartTimestamp, periods)

	return builder.WithAccounts(periodicVestingAccount)
}

func (builder AuthGenesisBuilder) WithEmptyValidatorVestingAccount(address sdk.AccAddress) AuthGenesisBuilder {
	// TODO create a validator vesting account builder and remove this method
	bacc := auth.NewBaseAccount(address, nil, nil, 0, 0)
	bva, err := vesting.NewBaseVestingAccount(bacc, nil, 1)
	if err != nil {
		panic(err.Error())
	}
	account := validatorvesting.NewValidatorVestingAccountRaw(bva, 0, nil, sdk.ConsAddress{}, nil, 90)
	return builder.WithAccounts(account)
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
	genState.Committees = []committeetypes.Committee{
		committeetypes.NewCommittee(
			genState.NextProposalID,
			"This committee is for testing.",
			members,
			[]committeetypes.Permission{committeetypes.GodPermission{}},
			d("0.667"),
			time.Hour*24*7,
		)}
	genState.NextProposalID += 1
	return app.GenesisState{
		committeetypes.ModuleName: committeetypes.ModuleCdc.MustMarshalJSON(genState),
	}
}

type incentiveGenesisBuilder struct {
	genesis     types.GenesisState
	genesisTime time.Time
}

func newIncentiveGenesisBuilder() incentiveGenesisBuilder {
	return incentiveGenesisBuilder{
		genesis:     types.DefaultGenesisState(),
		genesisTime: time.Time{},
	}
}

func (builder incentiveGenesisBuilder) build() types.GenesisState {
	return builder.genesis
}

func (builder incentiveGenesisBuilder) buildMarshalled() app.GenesisState {
	return app.GenesisState{
		types.ModuleName: types.ModuleCdc.MustMarshalJSON(builder.build()),
	}
}

func (builder incentiveGenesisBuilder) withGenesisTime(time time.Time) incentiveGenesisBuilder {
	builder.genesisTime = time
	builder.genesis.Params.ClaimEnd = time.Add(5 * oneYear)
	return builder
}

func (builder incentiveGenesisBuilder) withInitializedBorrowRewardPeriod(period types.MultiRewardPeriod) incentiveGenesisBuilder {
	builder.genesis.Params.HardBorrowRewardPeriods = append(builder.genesis.Params.HardBorrowRewardPeriods, period)

	accumulationTimeForPeriod := types.NewGenesisAccumulationTime(period.CollateralType, builder.genesisTime)
	builder.genesis.HardBorrowAccumulationTimes = append(builder.genesis.HardBorrowAccumulationTimes, accumulationTimeForPeriod)
	return builder
}

func (builder incentiveGenesisBuilder) withSimpleBorrowRewardPeriod(ctype string, rewardsPerSecond sdk.Coins) incentiveGenesisBuilder {
	return builder.withInitializedBorrowRewardPeriod(types.NewMultiRewardPeriod(
		true,
		ctype,
		builder.genesisTime,
		builder.genesisTime.Add(4*oneYear),
		rewardsPerSecond,
	))
}
func (builder incentiveGenesisBuilder) withInitializedSupplyRewardPeriod(period types.MultiRewardPeriod) incentiveGenesisBuilder {
	// TODO this could set the start/end times on the period according to builder.genesisTime
	// Then they could be created by a different builder

	builder.genesis.Params.HardSupplyRewardPeriods = append(builder.genesis.Params.HardSupplyRewardPeriods, period)

	accumulationTimeForPeriod := types.NewGenesisAccumulationTime(period.CollateralType, builder.genesisTime)
	builder.genesis.HardSupplyAccumulationTimes = append(builder.genesis.HardSupplyAccumulationTimes, accumulationTimeForPeriod)
	return builder
}

func (builder incentiveGenesisBuilder) withSimpleSupplyRewardPeriod(ctype string, rewardsPerSecond sdk.Coins) incentiveGenesisBuilder {
	return builder.withInitializedSupplyRewardPeriod(types.NewMultiRewardPeriod(
		true,
		ctype,
		builder.genesisTime,
		builder.genesisTime.Add(4*oneYear),
		rewardsPerSecond,
	))
}
func (builder incentiveGenesisBuilder) withInitializedDelegatorRewardPeriod(period types.RewardPeriod) incentiveGenesisBuilder {
	builder.genesis.Params.HardDelegatorRewardPeriods = append(builder.genesis.Params.HardDelegatorRewardPeriods, period)

	accumulationTimeForPeriod := types.NewGenesisAccumulationTime(period.CollateralType, builder.genesisTime)
	builder.genesis.HardDelegatorAccumulationTimes = append(builder.genesis.HardDelegatorAccumulationTimes, accumulationTimeForPeriod)
	return builder
}

func (builder incentiveGenesisBuilder) withSimpleDelegatorRewardPeriod(ctype string, rewardsPerSecond sdk.Coin) incentiveGenesisBuilder {
	return builder.withInitializedDelegatorRewardPeriod(types.NewRewardPeriod(
		true,
		ctype,
		builder.genesisTime,
		builder.genesisTime.Add(4*oneYear),
		rewardsPerSecond,
	))
}
func (builder incentiveGenesisBuilder) withInitializedUSDXRewardPeriod(period types.RewardPeriod) incentiveGenesisBuilder {
	builder.genesis.Params.USDXMintingRewardPeriods = append(builder.genesis.Params.USDXMintingRewardPeriods, period)

	accumulationTimeForPeriod := types.NewGenesisAccumulationTime(period.CollateralType, builder.genesisTime)
	builder.genesis.USDXAccumulationTimes = append(builder.genesis.USDXAccumulationTimes, accumulationTimeForPeriod)
	return builder
}

func (builder incentiveGenesisBuilder) withSimpleUSDXRewardPeriod(ctype string, rewardsPerSecond sdk.Coin) incentiveGenesisBuilder {
	return builder.withInitializedUSDXRewardPeriod(types.NewRewardPeriod(
		true,
		ctype,
		builder.genesisTime,
		builder.genesisTime.Add(4*oneYear),
		rewardsPerSecond,
	))
}

func (builder incentiveGenesisBuilder) withMultipliers(multipliers types.Multipliers) incentiveGenesisBuilder {
	builder.genesis.Params.ClaimMultipliers = multipliers
	return builder
}


