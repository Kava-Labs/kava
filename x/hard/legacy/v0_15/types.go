package v0_15

import (
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// ModuleName name that will be used throughout the module
	ModuleName = "hard"
)

// GenesisState is the state that must be provided at genesis.
type GenesisState struct {
	Params                    Params                   `json:"params" yaml:"params"`
	PreviousAccumulationTimes GenesisAccumulationTimes `json:"previous_accumulation_times" yaml:"previous_accumulation_times"`
	Deposits                  Deposits                 `json:"deposits" yaml:"deposits"`
	Borrows                   Borrows                  `json:"borrows" yaml:"borrows"`
	TotalSupplied             sdk.Coins                `json:"total_supplied" yaml:"total_supplied"`
	TotalBorrowed             sdk.Coins                `json:"total_borrowed" yaml:"total_borrowed"`
	TotalReserves             sdk.Coins                `json:"total_reserves" yaml:"total_reserves"`
}

// Params governance parameters for hard module
type Params struct {
	MoneyMarkets          MoneyMarkets `json:"money_markets" yaml:"money_markets"`
	MinimumBorrowUSDValue sdk.Dec      `json:"minimum_borrow_usd_value" yaml:"minimum_borrow_usd_value"`
}

// MoneyMarkets slice of MoneyMarket
type MoneyMarkets []MoneyMarket

// MoneyMarket is a money market for an individual asset
type MoneyMarket struct {
	Denom                  string            `json:"denom" yaml:"denom"`
	BorrowLimit            BorrowLimit       `json:"borrow_limit" yaml:"borrow_limit"`
	SpotMarketID           string            `json:"spot_market_id" yaml:"spot_market_id"`
	ConversionFactor       sdkmath.Int       `json:"conversion_factor" yaml:"conversion_factor"`
	InterestRateModel      InterestRateModel `json:"interest_rate_model" yaml:"interest_rate_model"`
	ReserveFactor          sdk.Dec           `json:"reserve_factor" yaml:"reserve_factor"`
	KeeperRewardPercentage sdk.Dec           `json:"keeper_reward_percentage" yaml:"keeper_reward_percentages"`
}

// BorrowLimit enforces restrictions on a money market
type BorrowLimit struct {
	HasMaxLimit  bool    `json:"has_max_limit" yaml:"has_max_limit"`
	MaximumLimit sdk.Dec `json:"maximum_limit" yaml:"maximum_limit"`
	LoanToValue  sdk.Dec `json:"loan_to_value" yaml:"loan_to_value"`
}

// InterestRateModel contains information about an asset's interest rate
type InterestRateModel struct {
	BaseRateAPY    sdk.Dec `json:"base_rate_apy" yaml:"base_rate_apy"`
	BaseMultiplier sdk.Dec `json:"base_multiplier" yaml:"base_multiplier"`
	Kink           sdk.Dec `json:"kink" yaml:"kink"`
	JumpMultiplier sdk.Dec `json:"jump_multiplier" yaml:"jump_multiplier"`
}

// GenesisAccumulationTimes slice of GenesisAccumulationTime
type GenesisAccumulationTimes []GenesisAccumulationTime

// GenesisAccumulationTime stores the previous distribution time and its corresponding denom
type GenesisAccumulationTime struct {
	CollateralType           string    `json:"collateral_type" yaml:"collateral_type"`
	PreviousAccumulationTime time.Time `json:"previous_accumulation_time" yaml:"previous_accumulation_time"`
	SupplyInterestFactor     sdk.Dec   `json:"supply_interest_factor" yaml:"supply_interest_factor"`
	BorrowInterestFactor     sdk.Dec   `json:"borrow_interest_factor" yaml:"borrow_interest_factor"`
}

// Deposits is a slice of Deposit
type Deposits []Deposit

// Deposit defines an amount of coins deposited into a hard module account
type Deposit struct {
	Depositor sdk.AccAddress        `json:"depositor" yaml:"depositor"`
	Amount    sdk.Coins             `json:"amount" yaml:"amount"`
	Index     SupplyInterestFactors `json:"index" yaml:"index"`
}

// SupplyInterestFactors is a slice of SupplyInterestFactor, because Amino won't marshal maps
type SupplyInterestFactors []SupplyInterestFactor

// SupplyInterestFactor defines an individual borrow interest factor
type SupplyInterestFactor struct {
	Denom string  `json:"denom" yaml:"denom"`
	Value sdk.Dec `json:"value" yaml:"value"`
}

// Borrows is a slice of Borrow
type Borrows []Borrow

// Borrow defines an amount of coins borrowed from a hard module account
type Borrow struct {
	Borrower sdk.AccAddress        `json:"borrower" yaml:"borrower"`
	Amount   sdk.Coins             `json:"amount" yaml:"amount"`
	Index    BorrowInterestFactors `json:"index" yaml:"index"`
}

// BorrowInterestFactors is a slice of BorrowInterestFactor, because Amino won't marshal maps
type BorrowInterestFactors []BorrowInterestFactor

// BorrowInterestFactor defines an individual borrow interest factor
type BorrowInterestFactor struct {
	Denom string  `json:"denom" yaml:"denom"`
	Value sdk.Dec `json:"value" yaml:"value"`
}
