package app

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/testutil/mock"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
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
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	evmkeeper "github.com/evmos/ethermint/x/evm/keeper"
	feemarketkeeper "github.com/evmos/ethermint/x/feemarket/keeper"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmtypes "github.com/tendermint/tendermint/types"
	tmdb "github.com/tendermint/tm-db"

	auctionkeeper "github.com/kava-labs/kava/x/auction/keeper"
	bep3keeper "github.com/kava-labs/kava/x/bep3/keeper"
	cdpkeeper "github.com/kava-labs/kava/x/cdp/keeper"
	committeekeeper "github.com/kava-labs/kava/x/committee/keeper"
	communitykeeper "github.com/kava-labs/kava/x/community/keeper"
	earnkeeper "github.com/kava-labs/kava/x/earn/keeper"
	evmutilkeeper "github.com/kava-labs/kava/x/evmutil/keeper"
	hardkeeper "github.com/kava-labs/kava/x/hard/keeper"
	incentivekeeper "github.com/kava-labs/kava/x/incentive/keeper"
	issuancekeeper "github.com/kava-labs/kava/x/issuance/keeper"
	kavadistkeeper "github.com/kava-labs/kava/x/kavadist/keeper"
	liquidkeeper "github.com/kava-labs/kava/x/liquid/keeper"
	pricefeedkeeper "github.com/kava-labs/kava/x/pricefeed/keeper"
	routerkeeper "github.com/kava-labs/kava/x/router/keeper"
	savingskeeper "github.com/kava-labs/kava/x/savings/keeper"
	swapkeeper "github.com/kava-labs/kava/x/swap/keeper"
)

var (
	emptyTime            time.Time
	testChainID                = "kavatest_1-1"
	defaultInitialHeight int64 = 1
)

// TestApp is a simple wrapper around an App. It exposes internal keepers for use in integration tests.
// This file also contains test helpers. Ideally they would be in separate package.
// Basic Usage:
//
//	Create a test app with NewTestApp, then all keepers and their methods can be accessed for test setup and execution.
//
// Advanced Usage:
//
//	Some tests call for an app to be initialized with some state. This can be achieved through keeper method calls (ie keeper.SetParams(...)).
//	However this leads to a lot of duplicated logic similar to InitGenesis methods.
//	So TestApp.InitializeFromGenesisStates() will call InitGenesis with the default genesis state.
//	and TestApp.InitializeFromGenesisStates(authState, cdpState) will do the same but overwrite the auth and cdp sections of the default genesis state
//	Creating the genesis states can be combersome, but helper methods can make it easier such as NewAuthGenStateFromAccounts below.
type TestApp struct {
	App

	GenesisAddrs []sdk.AccAddress
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

	app := NewApp(log.NewNopLogger(), db, DefaultNodeHome, nil, encCfg, DefaultOptions)
	return TestApp{App: *app}
}

// nolint
func (tApp TestApp) GetAccountKeeper() authkeeper.AccountKeeper { return tApp.accountKeeper }
func (tApp TestApp) GetBankKeeper() bankkeeper.Keeper           { return tApp.bankKeeper }
func (tApp TestApp) GetMintKeeper() mintkeeper.Keeper           { return tApp.mintKeeper }
func (tApp TestApp) GetStakingKeeper() stakingkeeper.Keeper     { return tApp.stakingKeeper }
func (tApp TestApp) GetSlashingKeeper() slashingkeeper.Keeper   { return tApp.slashingKeeper }
func (tApp TestApp) GetDistrKeeper() distkeeper.Keeper          { return tApp.distrKeeper }
func (tApp TestApp) GetGovKeeper() govkeeper.Keeper             { return tApp.govKeeper }
func (tApp TestApp) GetCrisisKeeper() crisiskeeper.Keeper       { return tApp.crisisKeeper }
func (tApp TestApp) GetParamsKeeper() paramskeeper.Keeper       { return tApp.paramsKeeper }

func (tApp TestApp) GetKavadistKeeper() kavadistkeeper.Keeper   { return tApp.kavadistKeeper }
func (tApp TestApp) GetAuctionKeeper() auctionkeeper.Keeper     { return tApp.auctionKeeper }
func (tApp TestApp) GetIssuanceKeeper() issuancekeeper.Keeper   { return tApp.issuanceKeeper }
func (tApp TestApp) GetBep3Keeper() bep3keeper.Keeper           { return tApp.bep3Keeper }
func (tApp TestApp) GetPriceFeedKeeper() pricefeedkeeper.Keeper { return tApp.pricefeedKeeper }
func (tApp TestApp) GetSwapKeeper() swapkeeper.Keeper           { return tApp.swapKeeper }
func (tApp TestApp) GetCDPKeeper() cdpkeeper.Keeper             { return tApp.cdpKeeper }
func (tApp TestApp) GetHardKeeper() hardkeeper.Keeper           { return tApp.hardKeeper }
func (tApp TestApp) GetCommitteeKeeper() committeekeeper.Keeper { return tApp.committeeKeeper }
func (tApp TestApp) GetIncentiveKeeper() incentivekeeper.Keeper { return tApp.incentiveKeeper }
func (tApp TestApp) GetEvmutilKeeper() evmutilkeeper.Keeper     { return tApp.evmutilKeeper }
func (tApp TestApp) GetEvmKeeper() *evmkeeper.Keeper            { return tApp.evmKeeper }
func (tApp TestApp) GetSavingsKeeper() savingskeeper.Keeper     { return tApp.savingsKeeper }
func (tApp TestApp) GetFeeMarketKeeper() feemarketkeeper.Keeper { return tApp.feeMarketKeeper }
func (tApp TestApp) GetLiquidKeeper() liquidkeeper.Keeper       { return tApp.liquidKeeper }
func (tApp TestApp) GetEarnKeeper() earnkeeper.Keeper           { return tApp.earnKeeper }
func (tApp TestApp) GetRouterKeeper() routerkeeper.Keeper       { return tApp.routerKeeper }
func (tApp TestApp) GetCommunityKeeper() communitykeeper.Keeper { return tApp.communityKeeper }

// LegacyAmino returns the app's amino codec.
func (app *App) LegacyAmino() *codec.LegacyAmino {
	return app.legacyAmino
}

// AppCodec returns the app's app codec.
func (app *App) AppCodec() codec.Codec {
	return app.appCodec
}

// InitializeDefaultGenesis runs InitGenesis for all specified modules using a given context. It does not initialize a chain from Genesis (as happens during InitChain).
// It is intended as a shorthand to initialize modules without calling keeper methods.
func (tApp TestApp) InitializeDefaultGenesis(ctx sdk.Context, overrideGenesisStates ...GenesisState) {

	genesisState := NewDefaultGenesisState()
	modifiedStates := make(map[string]bool)

	for _, state := range overrideGenesisStates {
		for k, v := range state {
			genesisState[k] = v

			// Ensure that the same module genesis state is not set more than once.
			// Multiple GenesisStates can have the same module genesis state, but
			// the same module genesis state will be overwritten.
			if previouslyModified := modifiedStates[k]; previouslyModified {
				panic(fmt.Sprintf("genesis state for module %s was set more than once, this overrides previous state", k))
			}

			modifiedStates[k] = true
		}
	}

	// This doesn't call ModuleManager.InitGenesis, or app.InitChainer to avoid the requirement to return validator updates.
	for _, module := range tApp.mm.OrderInitGenesis {
		_ = tApp.mm.Modules[module].InitGenesis(ctx, tApp.appCodec, genesisState[module])
		// discards val updates
	}
}

// SetupWithGenesisValSet initializes GenesisState with a single validator and genesis accounts
// that also act as delegators.
func GenesisStateWithSingleValidator(
	app *TestApp,
	genesisState GenesisState,
) GenesisState {
	privVal := mock.NewPV()
	pubKey, err := privVal.GetPubKey()
	if err != nil {
		panic(fmt.Errorf("error getting pubkey: %w", err))
	}

	// create validator set with single validator
	validator := tmtypes.NewValidator(pubKey, 1)
	valSet := tmtypes.NewValidatorSet([]*tmtypes.Validator{validator})

	// generate genesis account
	senderPrivKey := secp256k1.GenPrivKey()
	acc := authtypes.NewBaseAccount(senderPrivKey.PubKey().Address().Bytes(), senderPrivKey.PubKey(), 0, 0)
	app.GenesisAddrs = append(app.GenesisAddrs, acc.GetAddress())

	balances := []banktypes.Balance{
		{
			Address: acc.GetAddress().String(),
			Coins:   sdk.NewCoins(sdk.NewCoin("ukava", sdkmath.NewInt(100000000000000))),
		},
	}

	return genesisStateWithValSet(app, genesisState, valSet, []authtypes.GenesisAccount{acc}, balances...)
}

// genesisStateWithValSet applies the provided validator set and genesis accounts
// to the provided genesis state, appending to any existing state without replacement.
func genesisStateWithValSet(
	app *TestApp,
	genesisState GenesisState,
	valSet *tmtypes.ValidatorSet,
	genAccs []authtypes.GenesisAccount,
	balances ...banktypes.Balance,
) GenesisState {
	// set genesis accounts
	currentAuthGenesis := authtypes.GetGenesisStateFromAppState(app.appCodec, genesisState)
	currentAccs, err := authtypes.UnpackAccounts(currentAuthGenesis.Accounts)
	if err != nil {
		panic(fmt.Errorf("error unpacking accounts: %w", err))
	}

	// Add the new accounts to the existing ones
	authGenesis := authtypes.NewGenesisState(currentAuthGenesis.Params, append(currentAccs, genAccs...))

	validators := make([]stakingtypes.Validator, 0, len(valSet.Validators))
	delegations := make([]stakingtypes.Delegation, 0, len(valSet.Validators))

	bondAmt := sdk.DefaultPowerReduction

	for _, val := range valSet.Validators {
		pk, err := cryptocodec.FromTmPubKeyInterface(val.PubKey)
		if err != nil {
			panic(fmt.Errorf("error converting validator public key: %w", err))
		}

		pkAny, err := codectypes.NewAnyWithValue(pk)
		if err != nil {
			panic(fmt.Errorf("can't pack public key into Any: %w", err))
		}

		validator := stakingtypes.Validator{
			OperatorAddress: sdk.ValAddress(val.Address).String(),
			ConsensusPubkey: pkAny,
			Jailed:          false,
			Status:          stakingtypes.Bonded,
			Tokens:          bondAmt,
			DelegatorShares: sdk.OneDec(),
			Description: stakingtypes.Description{
				Moniker: "genesis validator",
			},
			UnbondingHeight:   int64(0),
			UnbondingTime:     time.Unix(0, 0).UTC(),
			Commission:        stakingtypes.NewCommission(sdk.ZeroDec(), sdk.ZeroDec(), sdk.ZeroDec()),
			MinSelfDelegation: sdk.ZeroInt(),
		}
		validators = append(validators, validator)
		delegations = append(delegations, stakingtypes.NewDelegation(genAccs[0].GetAddress(), val.Address.Bytes(), sdk.OneDec()))

	}
	// set validators and delegations
	currentStakingGenesis := stakingtypes.GetGenesisStateFromAppState(app.appCodec, genesisState)
	currentStakingGenesis.Params.BondDenom = "ukava"

	stakingGenesis := stakingtypes.NewGenesisState(
		currentStakingGenesis.Params,
		append(currentStakingGenesis.Validators, validators...),
		append(currentStakingGenesis.Delegations, delegations...),
	)

	// Add the new balances to the existing ones
	currentBankGenesis := banktypes.GetGenesisStateFromAppState(app.appCodec, genesisState)
	balances = append(currentBankGenesis.Balances, balances...)

	totalSupply := sdk.NewCoins()
	for _, b := range balances {
		// add genesis acc tokens to total supply
		totalSupply = totalSupply.Add(b.Coins...)
	}

	for range delegations {
		// add delegated tokens to total supply
		totalSupply = totalSupply.Add(sdk.NewCoin("ukava", bondAmt))
	}

	// add bonded amount to bonded pool module account
	balances = append(balances, banktypes.Balance{
		Address: authtypes.NewModuleAddress(stakingtypes.BondedPoolName).String(),
		Coins:   sdk.Coins{sdk.NewCoin("ukava", bondAmt)},
	})

	bankGenesis := banktypes.NewGenesisState(
		currentBankGenesis.Params,
		balances,
		totalSupply,
		currentBankGenesis.DenomMetadata,
	)

	// set genesis state
	genesisState[authtypes.ModuleName] = app.appCodec.MustMarshalJSON(authGenesis)
	genesisState[banktypes.ModuleName] = app.appCodec.MustMarshalJSON(bankGenesis)
	genesisState[stakingtypes.ModuleName] = app.appCodec.MustMarshalJSON(stakingGenesis)

	return genesisState
}

// InitializeFromGenesisStates calls InitChain on the app using the provided genesis states.
// If any module genesis states are missing, defaults are used.
func (tApp TestApp) InitializeFromGenesisStates(genesisStates ...GenesisState) TestApp {
	return tApp.InitializeFromGenesisStatesWithTimeAndChainIDAndHeight(emptyTime, testChainID, defaultInitialHeight, genesisStates...)
}

// InitializeFromGenesisStatesWithTime calls InitChain on the app using the provided genesis states and time.
// If any module genesis states are missing, defaults are used.
func (tApp TestApp) InitializeFromGenesisStatesWithTime(genTime time.Time, genesisStates ...GenesisState) TestApp {
	return tApp.InitializeFromGenesisStatesWithTimeAndChainIDAndHeight(genTime, testChainID, defaultInitialHeight, genesisStates...)
}

// InitializeFromGenesisStatesWithTimeAndChainID calls InitChain on the app using the provided genesis states, time, and chain id.
// If any module genesis states are missing, defaults are used.
func (tApp TestApp) InitializeFromGenesisStatesWithTimeAndChainID(genTime time.Time, chainID string, genesisStates ...GenesisState) TestApp {
	return tApp.InitializeFromGenesisStatesWithTimeAndChainIDAndHeight(genTime, chainID, defaultInitialHeight, genesisStates...)
}

// InitializeFromGenesisStatesWithTimeAndChainIDAndHeight calls InitChain on the app using the provided genesis states and other parameters.
// If any module genesis states are missing, defaults are used.
func (tApp TestApp) InitializeFromGenesisStatesWithTimeAndChainIDAndHeight(
	genTime time.Time,
	chainID string,
	initialHeight int64,
	genesisStates ...GenesisState,
) TestApp {
	// Create a default genesis state and overwrite with provided values
	genesisState := NewDefaultGenesisState()
	modifiedStates := make(map[string]bool)

	for _, state := range genesisStates {
		for k, v := range state {
			genesisState[k] = v

			// Ensure that the same module genesis state is not set more than once.
			// Multiple GenesisStates can have the same module genesis state, but
			// the same module genesis state will be overwritten.
			if previouslyModified := modifiedStates[k]; previouslyModified {
				panic(fmt.Sprintf("genesis state for module %s was set more than once, this overrides previous state", k))
			}

			modifiedStates[k] = true
		}
	}

	// Add default genesis states for at least 1 validator
	genesisState = GenesisStateWithSingleValidator(
		&tApp,
		genesisState,
	)

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
			// Set consensus params, which is needed by x/feemarket
			ConsensusParams: &abci.ConsensusParams{
				Block: &abci.BlockParams{
					MaxBytes: 200000,
					MaxGas:   20000000,
				},
			},
			InitialHeight: initialHeight,
		},
	)
	tApp.Commit()
	tApp.BeginBlock(abci.RequestBeginBlock{
		Header: tmproto.Header{
			Height: tApp.LastBlockHeight() + 1, Time: genTime, ChainID: chainID,
		},
	})

	return tApp
}

// DeleteGenesisValidator deletes the genesis validator from the staking module.
// This is useful for testing with validators, but only want to consider the
// validators added in the test. InitGenesis requires at least 1 validator, so
// it must be deleted additional validators are created.
func (tApp TestApp) DeleteGenesisValidator(t *testing.T, ctx sdk.Context) {
	sk := tApp.GetStakingKeeper()
	vals := sk.GetAllValidators(ctx)

	var genVal stakingtypes.Validator
	found := false
	for _, val := range vals {
		if val.GetMoniker() == "genesis validator" {
			genVal = val
			found = true
			break
		}
	}

	require.True(t, found, "genesis validator not found")

	delegations := sk.GetValidatorDelegations(ctx, genVal.GetOperator())
	for _, delegation := range delegations {
		_, err := sk.Undelegate(ctx, delegation.GetDelegatorAddr(), genVal.GetOperator(), delegation.Shares)
		require.NoError(t, err)
	}
}

func (tApp TestApp) DeleteGenesisValidatorCoins(t *testing.T, ctx sdk.Context) {
	ak := tApp.GetAccountKeeper()
	bk := tApp.GetBankKeeper()

	notBondedAcc := ak.GetModuleAccount(ctx, stakingtypes.NotBondedPoolName)

	// Burn genesis account balance - use staking module to burn
	genAccBal := bk.GetAllBalances(ctx, tApp.GenesisAddrs[0])
	err := bk.SendCoinsFromAccountToModule(ctx, tApp.GenesisAddrs[0], stakingtypes.NotBondedPoolName, genAccBal)
	require.NoError(t, err)

	// Burn coins from the module account
	err = bk.BurnCoins(
		ctx,
		stakingtypes.NotBondedPoolName,
		bk.GetAllBalances(ctx, notBondedAcc.GetAddress()),
	)
	require.NoError(t, err)
}

// CheckBalance requires the account address has the expected amount of coins.
func (tApp TestApp) CheckBalance(t *testing.T, ctx sdk.Context, owner sdk.AccAddress, expectedCoins sdk.Coins) {
	coins := tApp.GetBankKeeper().GetAllBalances(ctx, owner)
	require.Equal(t, expectedCoins, coins)
}

// GetModuleAccountBalance gets the current balance of the denom for a module account
func (tApp TestApp) GetModuleAccountBalance(ctx sdk.Context, moduleName string, denom string) sdkmath.Int {
	moduleAcc := tApp.accountKeeper.GetModuleAccount(ctx, moduleName)
	balance := tApp.bankKeeper.GetBalance(ctx, moduleAcc.GetAddress(), denom)
	return balance.Amount
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

// CreateNewUnbondedValidator creates a new validator in the staking module.
// New validators are unbonded until the end blocker is run.
func (tApp TestApp) CreateNewUnbondedValidator(ctx sdk.Context, valAddress sdk.ValAddress, selfDelegation sdkmath.Int) error {
	msg, err := stakingtypes.NewMsgCreateValidator(
		valAddress,
		ed25519.GenPrivKey().PubKey(),
		sdk.NewCoin(tApp.stakingKeeper.BondDenom(ctx), selfDelegation),
		stakingtypes.Description{},
		stakingtypes.NewCommissionRates(sdk.ZeroDec(), sdk.ZeroDec(), sdk.ZeroDec()),
		sdkmath.NewInt(1e6),
	)
	if err != nil {
		return err
	}

	msgServer := stakingkeeper.NewMsgServerImpl(tApp.stakingKeeper)
	_, err = msgServer.CreateValidator(sdk.WrapSDKContext(ctx), msg)
	return err
}

func (tApp TestApp) SetInflation(ctx sdk.Context, value sdk.Dec) {
	mk := tApp.GetMintKeeper()

	mintParams := mk.GetParams(ctx)
	mintParams.InflationMax = sdk.ZeroDec()
	mintParams.InflationMin = sdk.ZeroDec()

	if err := mintParams.Validate(); err != nil {
		panic(err)
	}

	mk.SetParams(ctx, mintParams)
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

// RandomAddress non-deterministically generates a new address, discarding the private key.
func RandomAddress() sdk.AccAddress {
	secret := make([]byte, 32)
	_, err := rand.Read(secret)
	if err != nil {
		panic("Could not read randomness")
	}
	key := secp256k1.GenPrivKeyFromSecret(secret)
	return sdk.AccAddress(key.PubKey().Address())
}

// NewFundedGenStateWithSameCoins creates a (auth and bank) genesis state populated with accounts from the given addresses and balance.
func NewFundedGenStateWithSameCoins(cdc codec.JSONCodec, balance sdk.Coins, addresses []sdk.AccAddress) GenesisState {
	builder := NewAuthBankGenesisBuilder()
	for _, address := range addresses {
		builder.WithSimpleAccount(address, balance)
	}
	return builder.BuildMarshalled(cdc)
}

// NewFundedGenStateWithCoins creates a (auth and bank) genesis state populated with accounts from the given addresses and coins.
func NewFundedGenStateWithCoins(cdc codec.JSONCodec, coins []sdk.Coins, addresses []sdk.AccAddress) GenesisState {
	builder := NewAuthBankGenesisBuilder()
	for i, address := range addresses {
		builder.WithSimpleAccount(address, coins[i])
	}
	return builder.BuildMarshalled(cdc)
}

// NewFundedGenStateWithSameCoinsWithModuleAccount creates a (auth and bank) genesis state populated with accounts from the given addresses and balance along with an empty module account
func NewFundedGenStateWithSameCoinsWithModuleAccount(cdc codec.JSONCodec, coins sdk.Coins, addresses []sdk.AccAddress, modAcc *authtypes.ModuleAccount) GenesisState {
	builder := NewAuthBankGenesisBuilder()

	for _, address := range addresses {
		builder.WithSimpleAccount(address, coins)
	}

	builder.WithSimpleModuleAccount(modAcc.Address, nil)

	return builder.BuildMarshalled(cdc)
}
