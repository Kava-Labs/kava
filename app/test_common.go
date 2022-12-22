package app

import (
	"encoding/json"
	"math/rand"
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	crisiskeeper "github.com/cosmos/cosmos-sdk/x/crisis/keeper"
	distkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"
	paramskeeper "github.com/cosmos/cosmos-sdk/x/params/keeper"
	slashingkeeper "github.com/cosmos/cosmos-sdk/x/slashing/keeper"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmdb "github.com/tendermint/tm-db"
	evmkeeper "github.com/tharsis/ethermint/x/evm/keeper"
	feemarketkeeper "github.com/tharsis/ethermint/x/feemarket/keeper"

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
	kavamintkeeper "github.com/kava-labs/kava/x/kavamint/keeper"
	kavaminttypes "github.com/kava-labs/kava/x/kavamint/types"
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

	app := NewApp(log.NewNopLogger(), db, DefaultNodeHome, nil, encCfg, DefaultOptions)
	return TestApp{App: *app}
}

// nolint
func (tApp TestApp) GetAccountKeeper() authkeeper.AccountKeeper { return tApp.accountKeeper }
func (tApp TestApp) GetBankKeeper() bankkeeper.Keeper           { return tApp.bankKeeper }
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
func (tApp TestApp) GetKavamintKeeper() kavamintkeeper.Keeper   { return tApp.kavamintKeeper }
func (tApp TestApp) GetCommunityKeeper() communitykeeper.Keeper { return tApp.communityKeeper }

// LegacyAmino returns the app's amino codec.
func (app *App) LegacyAmino() *codec.LegacyAmino {
	return app.legacyAmino
}

// AppCodec returns the app's app codec.
func (app *App) AppCodec() codec.Codec {
	return app.appCodec
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
func (tApp TestApp) InitializeFromGenesisStatesWithTimeAndChainIDAndHeight(genTime time.Time, chainID string, initialHeight int64, genesisStates ...GenesisState) TestApp {
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

// CheckBalance requires the account address has the expected amount of coins.
func (tApp TestApp) CheckBalance(t *testing.T, ctx sdk.Context, owner sdk.AccAddress, expectedCoins sdk.Coins) {
	coins := tApp.GetBankKeeper().GetAllBalances(ctx, owner)
	require.Equal(t, expectedCoins, coins)
}

// GetModuleAccountBalance gets the current balance of the denom for a module account
func (tApp TestApp) GetModuleAccountBalance(ctx sdk.Context, moduleName string, denom string) sdk.Int {
	moduleAcc := tApp.accountKeeper.GetModuleAccount(ctx, moduleName)
	balance := tApp.bankKeeper.GetBalance(ctx, moduleAcc.GetAddress(), denom)
	return balance.Amount
}

// FundAccount is a utility function that funds an account by minting and sending the coins to the address.
func (tApp TestApp) FundAccount(ctx sdk.Context, addr sdk.AccAddress, amounts sdk.Coins) error {
	if err := tApp.bankKeeper.MintCoins(ctx, kavaminttypes.ModuleAccountName, amounts); err != nil {
		return err
	}

	return tApp.bankKeeper.SendCoinsFromModuleToAccount(ctx, kavaminttypes.ModuleAccountName, addr, amounts)
}

// NewQueryServerTestHelper creates a new QueryServiceTestHelper that wraps the provided sdk.Context.
func (tApp TestApp) NewQueryServerTestHelper(ctx sdk.Context) *baseapp.QueryServiceTestHelper {
	return baseapp.NewQueryServerTestHelper(ctx, tApp.interfaceRegistry)
}

// FundModuleAccount is a utility function that funds a module account by minting and sending the coins to the address.
func (tApp TestApp) FundModuleAccount(ctx sdk.Context, recipientMod string, amounts sdk.Coins) error {
	if err := tApp.bankKeeper.MintCoins(ctx, kavaminttypes.ModuleAccountName, amounts); err != nil {
		return err
	}

	return tApp.bankKeeper.SendCoinsFromModuleToModule(ctx, kavaminttypes.ModuleAccountName, recipientMod, amounts)
}

// CreateNewUnbondedValidator creates a new validator in the staking module.
// New validators are unbonded until the end blocker is run.
func (tApp TestApp) CreateNewUnbondedValidator(ctx sdk.Context, valAddress sdk.ValAddress, selfDelegation sdk.Int) error {
	msg, err := stakingtypes.NewMsgCreateValidator(
		valAddress,
		ed25519.GenPrivKey().PubKey(),
		sdk.NewCoin(tApp.stakingKeeper.BondDenom(ctx), selfDelegation),
		stakingtypes.Description{},
		stakingtypes.NewCommissionRates(sdk.ZeroDec(), sdk.ZeroDec(), sdk.ZeroDec()),
		sdk.NewInt(1e6),
	)
	if err != nil {
		return err
	}

	msgServer := stakingkeeper.NewMsgServerImpl(tApp.stakingKeeper)
	_, err = msgServer.CreateValidator(sdk.WrapSDKContext(ctx), msg)
	return err
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
