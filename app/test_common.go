package app

import (
	"encoding/json"
	"math/rand"
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	crisiskeeper "github.com/cosmos/cosmos-sdk/x/crisis/keeper"
	distkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"
	mintkeeper "github.com/cosmos/cosmos-sdk/x/mint/keeper"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	paramskeeper "github.com/cosmos/cosmos-sdk/x/params/keeper"
	slashingkeeper "github.com/cosmos/cosmos-sdk/x/slashing/keeper"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmdb "github.com/tendermint/tm-db"

	bep3keeper "github.com/kava-labs/kava/x/bep3/keeper"
	committeekeeper "github.com/kava-labs/kava/x/committee/keeper"
	issuancekeeper "github.com/kava-labs/kava/x/issuance/keeper"
	kavadistkeeper "github.com/kava-labs/kava/x/kavadist/keeper"
	pricefeedkeeper "github.com/kava-labs/kava/x/pricefeed/keeper"
	swapkeeper "github.com/kava-labs/kava/x/swap/keeper"
)

var (
	emptyTime    time.Time
	emptyChainID string
)

// TestApp is a simple wrapper around an App. It exposes internal keepers for use in integration tests.
// This file also contains test helpers. Ideally they would be in separate package.
// Basic Usage:
// 	Create a test app with NewTestApp, then all keepers and their methods can be accessed for test setup and execution.
// Advanced Usage:
// 	Some tests call for an app to be initialized with some state. This can be achieved through keeper method calls (ie keeper.SetParams(...)).
// 	However this leads to a lot of duplicated logic similar to InitGenesis methods.
// 	So TestApp.InitializeFromGenesisStates() will call InitGenesis with the default genesis state.
//	and TestApp.InitializeFromGenesisStates(authState, cdpState) will do the same but overwrite the auth and cdp sections of the default genesis state
// 	Creating the genesis states can be combersome, but helper methods can make it easier such as NewAuthGenStateFromAccounts below.
type TestApp struct {
	App
}

// NewTestApp creates a new TestApp
//
// Note, it also sets the sdk config with the app's address prefix, coin type, etc.
func NewTestApp() TestApp {
	SetSDKConfig()

	return NewTestAppFromSealed()
}

// NewTestAppFromSealed creates a TestApp without first setting sdk config.
func NewTestAppFromSealed() TestApp {
	db := tmdb.NewMemDB()

	encCfg := MakeEncodingConfig()

	app := NewApp(log.NewNopLogger(), db, nil, encCfg, Options{})
	return TestApp{App: *app}
}

// nolint
func (tApp TestApp) GetAccountKeeper() authkeeper.AccountKeeper { return tApp.accountKeeper }
func (tApp TestApp) GetBankKeeper() bankkeeper.Keeper           { return tApp.bankKeeper }
func (tApp TestApp) GetStakingKeeper() stakingkeeper.Keeper     { return tApp.stakingKeeper }
func (tApp TestApp) GetSlashingKeeper() slashingkeeper.Keeper   { return tApp.slashingKeeper }
func (tApp TestApp) GetMintKeeper() mintkeeper.Keeper           { return tApp.mintKeeper }
func (tApp TestApp) GetDistrKeeper() distkeeper.Keeper          { return tApp.distrKeeper }
func (tApp TestApp) GetGovKeeper() govkeeper.Keeper             { return tApp.govKeeper }
func (tApp TestApp) GetCrisisKeeper() crisiskeeper.Keeper       { return tApp.crisisKeeper }
func (tApp TestApp) GetParamsKeeper() paramskeeper.Keeper       { return tApp.paramsKeeper }

func (tApp TestApp) GetKavadistKeeper() kavadistkeeper.Keeper   { return tApp.kavadistKeeper }
func (tApp TestApp) GetIssuanceKeeper() issuancekeeper.Keeper   { return tApp.issuanceKeeper }
func (tApp TestApp) GetPriceFeedKeeper() pricefeedkeeper.Keeper { return tApp.pricefeedKeeper }
func (tApp TestApp) GetBep3Keeper() bep3keeper.Keeper           { return tApp.bep3Keeper }
func (tApp TestApp) GetSwapKeeper() swapkeeper.Keeper           { return tApp.swapKeeper }
func (tApp TestApp) GetCommitteeKeeper() committeekeeper.Keeper { return tApp.committeeKeeper }

// TODO add back with modules
// func (tApp TestApp) GetVVKeeper() validatorvesting.Keeper { return tApp.vvKeeper }
// func (tApp TestApp) GetAuctionKeeper() auction.Keeper     { return tApp.auctionKeeper }
// func (tApp TestApp) GetCDPKeeper() cdp.Keeper             { return tApp.cdpKeeper }
// func (tApp TestApp) GetIncentiveKeeper() incentive.Keeper { return tApp.incentiveKeeper }
// func (tApp TestApp) GetHardKeeper() hard.Keeper           { return tApp.hardKeeper }
// func (tApp TestApp) GetCommitteeKeeper() committee.Keeper { return tApp.committeeKeeper }

// LegacyAmino returns the app's amino codec.
func (app *App) LegacyAmino() *codec.LegacyAmino {
	return app.legacyAmino
}

// AppCodec returns the app's app codec.
func (app *App) AppCodec() codec.Codec {
	return app.appCodec
}

// InitializeFromGenesisStates calls InitChain on the app using the default genesis state, overwitten with any passed in genesis states
func (tApp TestApp) InitializeFromGenesisStates(genesisStates ...GenesisState) TestApp {
	return tApp.InitializeFromGenesisStatesWithTimeAndChainID(emptyTime, emptyChainID, genesisStates...)
}

// InitializeFromGenesisStatesWithTime calls InitChain on the app using the default genesis state, overwitten with any passed in genesis states and genesis Time
func (tApp TestApp) InitializeFromGenesisStatesWithTime(genTime time.Time, genesisStates ...GenesisState) TestApp {
	return tApp.InitializeFromGenesisStatesWithTimeAndChainID(genTime, emptyChainID, genesisStates...)
}

// InitializeFromGenesisStatesWithTimeAndChainID calls InitChain on the app using the default genesis state, overwitten with any passed in genesis states and genesis Time
func (tApp TestApp) InitializeFromGenesisStatesWithTimeAndChainID(genTime time.Time, chainID string, genesisStates ...GenesisState) TestApp {
	// Create a default genesis state and overwrite with provided values
	genesisState := NewDefaultGenesisState()
	for _, state := range genesisStates {
		for k, v := range state {
			genesisState[k] = v
		}
	}

	// Initialize the chain
	stateBytes, err := json.Marshal(genesisState)
	if err != nil {
		panic(err)
	}
	tApp.InitChain(
		abci.RequestInitChain{
			Time:          genTime,
			Validators:    []abci.ValidatorUpdate{},
			AppStateBytes: stateBytes,
			ChainId:       chainID,
		},
	)
	tApp.Commit()
	tApp.BeginBlock(abci.RequestBeginBlock{Header: tmproto.Header{Height: tApp.LastBlockHeight() + 1, Time: genTime}})
	return tApp
}

// CheckBalance requires the account address has the expected amount of coins.
func (tApp TestApp) CheckBalance(t *testing.T, ctx sdk.Context, owner sdk.AccAddress, expectedCoins sdk.Coins) {
	coins := tApp.GetBankKeeper().GetAllBalances(ctx, owner)
	require.Equal(t, expectedCoins, coins)
}

// FundAccount is a utility function that funds an account by minting and sending the coins to the address.
func (tApp TestApp) FundAccount(ctx sdk.Context, addr sdk.AccAddress, amounts sdk.Coins) error {
	if err := tApp.bankKeeper.MintCoins(ctx, minttypes.ModuleName, amounts); err != nil {
		return err
	}

	return tApp.bankKeeper.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, addr, amounts)
}

// NewQueryServerTestHelper creates a new QueryServiceTestHelper that wraps the provided sdk.Context.
func (tApp TestApp) NewQueryServerTestHelper(ctx sdk.Context) *baseapp.QueryServiceTestHelper {
	return baseapp.NewQueryServerTestHelper(ctx, tApp.interfaceRegistry)
}

// FundModuleAccount is a utility function that funds a module account by minting and sending the coins to the address.
func (tApp TestApp) FundModuleAccount(ctx sdk.Context, recipientMod string, amounts sdk.Coins) error {
	if err := tApp.bankKeeper.MintCoins(ctx, minttypes.ModuleName, amounts); err != nil {
		return err
	}

	return tApp.bankKeeper.SendCoinsFromModuleToModule(ctx, minttypes.ModuleName, recipientMod, amounts)
}

// GeneratePrivKeyAddressPairsFromRand generates (deterministically) a total of n private keys and addresses.
func GeneratePrivKeyAddressPairs(n int) (keys []cryptotypes.PrivKey, addrs []sdk.AccAddress) {
	r := rand.New(rand.NewSource(12345)) // make the generation deterministic
	keys = make([]cryptotypes.PrivKey, n)
	addrs = make([]sdk.AccAddress, n)
	for i := 0; i < n; i++ {
		secret := make([]byte, 32)
		_, err := r.Read(secret)
		if err != nil {
			panic("Could not read randomness")
		}
		keys[i] = secp256k1.GenPrivKeyFromSecret(secret)
		addrs[i] = sdk.AccAddress(keys[i].PubKey().Address())
	}
	return
}

// NewFundedGenStateWithSameCoins creates a (auth and bank) genesis state populated with accounts from the given addresses and balance.
func NewFundedGenStateWithSameCoins(cdc codec.JSONCodec, balance sdk.Coins, addresses []sdk.AccAddress) GenesisState {
	balances := make([]sdk.Coins, len(addresses))
	for i, _ := range addresses {
		balances[i] = balance
	}
	return NewFundedGenStateWithCoins(cdc, balances, addresses)
}

// NewFundedGenStateWithCoins creates a (auth and bank) genesis state populated with accounts from the given addresses and coins.
func NewFundedGenStateWithCoins(cdc codec.JSONCodec, coins []sdk.Coins, addresses []sdk.AccAddress) GenesisState {
	balances := make([]banktypes.Balance, len(addresses))
	for i, addr := range addresses {
		balances[i] = banktypes.Balance{
			Address: addr.String(),
			Coins:   coins[i],
		}
	}

	bankGenesis := banktypes.NewGenesisState(
		banktypes.DefaultParams(),
		balances,
		nil,
		[]banktypes.Metadata{}, // Metadata is not widely used in the sdk or kava
	)

	accounts := make(authtypes.GenesisAccounts, len(addresses))
	for i := range addresses {
		accounts[i] = authtypes.NewBaseAccount(addresses[i], nil, 0, 0)
	}

	authGenesis := authtypes.NewGenesisState(authtypes.DefaultParams(), accounts)

	return GenesisState{
		authtypes.ModuleName: cdc.MustMarshalJSON(authGenesis),
		banktypes.ModuleName: cdc.MustMarshalJSON(bankGenesis),
	}
}

// TODO move auth builder to a new package

// AuthBankGenesisBuilder is a tool for creating a combined auth and bank genesis state.
// Helper methods create basic accounts types and add them to a default genesis state.
// All methods return the builder so method call can be chained together.
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

// TODO upgrade once validator vesting upgraded
// // WithEmptyValidatorVestingAccount adds a stub validator vesting account to the genesis state.
// func (builder AuthGenesisBuilder) WithEmptyValidatorVestingAccount(address sdk.AccAddress) AuthGenesisBuilder {
// 	// TODO create a validator vesting account builder and remove this method
// 	bacc := auth.NewBaseAccount(address, nil, nil, 0, 0)
// 	bva, err := vesting.NewBaseVestingAccount(bacc, nil, 1)
// 	if err != nil {
// 		panic(err.Error())
// 	}
// 	account := validatorvesting.NewValidatorVestingAccountRaw(bva, 0, nil, sdk.ConsAddress{}, nil, 90)
// 	return builder.WithAccounts(account)
// }
