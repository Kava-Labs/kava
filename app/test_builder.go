package app

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

// AuthBankGenesisBuilder is a tool for creating a combined auth and bank genesis state.
// Helper methods create basic accounts types and add them to a default genesis state.
// All methods return the builder so method calls can be chained together.
//
// Example:
//     // create a single account genesis state
//     builder := NewAuthBankGenesisBuilder().WithSimpleAccount(testUserAddress, testCoins)
//     genesisState := builder.BuildMarshalled()
//
type AuthBankGenesisBuilder struct {
	AuthGenesis authtypes.GenesisState
	BankGenesis banktypes.GenesisState
}

// NewAuthBankGenesisBuilder creates a AuthBankGenesisBuilder containing default genesis states.
func NewAuthBankGenesisBuilder() *AuthBankGenesisBuilder {
	return &AuthBankGenesisBuilder{
		AuthGenesis: *authtypes.DefaultGenesisState(),
		BankGenesis: *banktypes.DefaultGenesisState(),
	}
}

// BuildMarshalled assembles the final GenesisState and json encodes it into a generic genesis type.
func (builder *AuthBankGenesisBuilder) BuildMarshalled(cdc codec.JSONCodec) GenesisState {
	return GenesisState{
		authtypes.ModuleName: cdc.MustMarshalJSON(&builder.AuthGenesis),
		banktypes.ModuleName: cdc.MustMarshalJSON(&builder.BankGenesis),
	}
}

// WithAccounts adds accounts of any type to the genesis state.
func (builder *AuthBankGenesisBuilder) WithAccounts(account ...authtypes.GenesisAccount) *AuthBankGenesisBuilder {
	existing, err := authtypes.UnpackAccounts(builder.AuthGenesis.Accounts)
	if err != nil {
		panic(err)
	}
	existing = append(existing, account...)

	existingPacked, err := authtypes.PackAccounts(existing)
	if err != nil {
		panic(err)
	}
	builder.AuthGenesis.Accounts = existingPacked
	return builder
}

// WithBalances adds balances to the bank genesis state.
// It does not check the new denom is in the genesis state denom metadata.
func (builder *AuthBankGenesisBuilder) WithBalances(balance ...banktypes.Balance) *AuthBankGenesisBuilder {
	builder.BankGenesis.Balances = append(builder.BankGenesis.Balances, balance...)
	if !builder.BankGenesis.Supply.Empty() {
		for _, b := range balance {
			builder.BankGenesis.Supply = builder.BankGenesis.Supply.Add(b.Coins...)
		}
	}
	return builder
}

// WithSimpleAccount adds a standard account to the genesis state.
func (builder *AuthBankGenesisBuilder) WithSimpleAccount(address sdk.AccAddress, balance sdk.Coins) *AuthBankGenesisBuilder {
	return builder.
		WithAccounts(authtypes.NewBaseAccount(address, nil, 0, 0)).
		WithBalances(banktypes.Balance{Address: address.String(), Coins: balance})
}

// WithSimpleModuleAccount adds a module account to the genesis state.
func (builder *AuthBankGenesisBuilder) WithSimpleModuleAccount(moduleName string, balance sdk.Coins, permissions ...string) *AuthBankGenesisBuilder {
	account := authtypes.NewEmptyModuleAccount(moduleName, permissions...)

	return builder.
		WithAccounts(account).
		WithBalances(banktypes.Balance{Address: account.Address, Coins: balance})
}

// WithSimplePeriodicVestingAccount adds a periodic vesting account to the genesis state.
func (builder *AuthBankGenesisBuilder) WithSimplePeriodicVestingAccount(address sdk.AccAddress, balance sdk.Coins, periods vestingtypes.Periods, firstPeriodStartTimestamp int64) *AuthBankGenesisBuilder {
	vestingAccount := newPeriodicVestingAccount(address, periods, firstPeriodStartTimestamp)

	return builder.
		WithAccounts(vestingAccount).
		WithBalances(banktypes.Balance{Address: address.String(), Coins: balance})
}

// newPeriodicVestingAccount creates a periodic vesting account from a set of vesting periods.
func newPeriodicVestingAccount(address sdk.AccAddress, periods vestingtypes.Periods, firstPeriodStartTimestamp int64) *vestingtypes.PeriodicVestingAccount {
	baseAccount := authtypes.NewBaseAccount(address, nil, 0, 0)

	originalVesting := sdk.NewCoins()
	for _, p := range periods {
		originalVesting = originalVesting.Add(p.Amount...)
	}

	var totalPeriods int64
	for _, p := range periods {
		totalPeriods += p.Length
	}
	endTime := firstPeriodStartTimestamp + totalPeriods

	baseVestingAccount := vestingtypes.NewBaseVestingAccount(baseAccount, originalVesting, endTime)
	return vestingtypes.NewPeriodicVestingAccountRaw(baseVestingAccount, firstPeriodStartTimestamp, periods)
}
