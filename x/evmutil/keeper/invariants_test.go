package keeper_test

import (
	"math/big"
	"testing"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/stretchr/testify/suite"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/evmutil/keeper"
	"github.com/kava-labs/kava/x/evmutil/testutil"
	"github.com/kava-labs/kava/x/evmutil/types"
)

type invariantTestSuite struct {
	testutil.Suite
	invariants   map[string]map[string]sdk.Invariant
	contractAddr types.InternalEVMAddress

	cosmosCoin             types.AllowedCosmosCoinERC20Token
	cosmosCoinContractAddr types.InternalEVMAddress
}

func TestInvariantTestSuite(t *testing.T) {
	suite.Run(t, new(invariantTestSuite))
}

func (suite *invariantTestSuite) SetupTest() {
	suite.Suite.SetupTest()

	suite.contractAddr = suite.DeployERC20()

	suite.invariants = make(map[string]map[string]sdk.Invariant)
	keeper.RegisterInvariants(suite, suite.BankKeeper, suite.Keeper)
}

func (suite *invariantTestSuite) SetupValidState() {
	for i := 0; i < 4; i++ {
		suite.Keeper.SetAccount(suite.Ctx, *types.NewAccount(
			suite.Addrs[i],
			keeper.ConversionMultiplier.QuoRaw(2),
		))
	}
	suite.FundModuleAccountWithKava(
		types.ModuleName,
		sdk.NewCoins(
			sdk.NewCoin("ukava", sdkmath.NewInt(2)), // ( sum of all minor balances ) / conversion multiplier
		),
	)

	err := suite.Keeper.MintERC20(suite.Ctx, suite.contractAddr, suite.Key1Addr, big.NewInt(1000000))
	suite.Require().NoError(err)

	// key1 ERC20 bal -10000, sdk.Coin +1000
	// Module account balance 0 -> 1000
	_, err = suite.Keeper.CallEVM(
		suite.Ctx,
		types.ERC20MintableBurnableContract.ABI,
		suite.Key1Addr.Address,
		suite.contractAddr,
		"convertToCoin",
		// convertToCoin ERC20 args
		suite.Key1Addr.Address,
		big.NewInt(1000),
	)
	suite.Require().NoError(err)

	// setup a cosmos coin erc20 with supply
	tokenInfo := types.NewAllowedCosmosCoinERC20Token("magic", "Magic coin", "MAGIC", 6)
	suite.cosmosCoin = tokenInfo
	params := suite.Keeper.GetParams(suite.Ctx)
	params.AllowedCosmosDenoms = append(params.AllowedCosmosDenoms, tokenInfo)
	suite.Keeper.SetParams(suite.Ctx, params)

	suite.cosmosCoinContractAddr, err = suite.Keeper.GetOrDeployCosmosCoinERC20Contract(suite.Ctx, tokenInfo)
	suite.NoError(err)

	// setup converted coin position
	err = suite.Keeper.MintERC20(
		suite.Ctx,
		suite.cosmosCoinContractAddr,
		testutil.RandomInternalEVMAddress(),
		big.NewInt(1e12),
	)
	suite.NoError(err)
	err = suite.App.FundModuleAccount(
		suite.Ctx,
		types.ModuleName,
		sdk.NewCoins(sdk.NewInt64Coin(tokenInfo.CosmosDenom, 1e12)),
	)
	suite.NoError(err)
}

// RegisterRoute implements sdk.InvariantRegistry
func (suite *invariantTestSuite) RegisterRoute(moduleName string, route string, invariant sdk.Invariant) {
	_, exists := suite.invariants[moduleName]

	if !exists {
		suite.invariants[moduleName] = make(map[string]sdk.Invariant)
	}

	suite.invariants[moduleName][route] = invariant
}

func (suite *invariantTestSuite) runInvariant(route string, invariant func(bankKeeper types.BankKeeper, k keeper.Keeper) sdk.Invariant) (string, bool) {
	ctx := suite.Ctx
	registeredInvariant := suite.invariants[types.ModuleName][route]
	suite.Require().NotNil(registeredInvariant)

	// direct call
	dMessage, dBroken := invariant(suite.BankKeeper, suite.Keeper)(ctx)
	// registered call
	rMessage, rBroken := registeredInvariant(ctx)
	// all call
	aMessage, aBroken := keeper.AllInvariants(suite.BankKeeper, suite.Keeper)(ctx)

	// require matching values for direct call and registered call
	suite.Require().Equal(dMessage, rMessage, "expected registered invariant message to match")
	suite.Require().Equal(dBroken, rBroken, "expected registered invariant broken to match")
	// require matching values for direct call and all invariants call if broken
	suite.Require().Equal(dBroken, aBroken, "expected all invariant broken to match")
	if dBroken {
		suite.Require().Equal(dMessage, aMessage, "expected all invariant message to match")
	}

	return dMessage, dBroken
}

func (suite *invariantTestSuite) TestFullyBackedInvariant() {
	// default state is valid
	_, broken := suite.runInvariant("fully-backed", keeper.FullyBackedInvariant)
	suite.Equal(false, broken)

	suite.SetupValidState()
	_, broken = suite.runInvariant("fully-backed", keeper.FullyBackedInvariant)
	suite.Equal(false, broken)

	// break invariant by increasing total minor balances above module balance
	suite.Keeper.AddBalance(suite.Ctx, suite.Addrs[0], sdk.OneInt())

	message, broken := suite.runInvariant("fully-backed", keeper.FullyBackedInvariant)
	suite.Equal("evmutil: fully backed broken invariant\nsum of minor balances greater than module account\n", message)
	suite.Equal(true, broken)
}

func (suite *invariantTestSuite) TestSmallBalances() {
	// default state is valid
	_, broken := suite.runInvariant("small-balances", keeper.SmallBalancesInvariant)
	suite.Equal(false, broken)

	suite.SetupValidState()
	_, broken = suite.runInvariant("small-balances", keeper.SmallBalancesInvariant)
	suite.Equal(false, broken)

	// increase minor balance at least above conversion multiplier
	suite.Keeper.AddBalance(suite.Ctx, suite.Addrs[0], keeper.ConversionMultiplier)
	// add same number of ukava to avoid breaking other invariants
	amt := sdk.NewCoins(sdk.NewInt64Coin(keeper.CosmosDenom, 1))
	suite.Require().NoError(
		suite.App.FundModuleAccount(suite.Ctx, types.ModuleName, amt),
	)

	message, broken := suite.runInvariant("small-balances", keeper.SmallBalancesInvariant)
	suite.Equal("evmutil: small balances broken invariant\nminor balances not all less than overflow\n", message)
	suite.Equal(true, broken)
}

// the cosmos-coins-fully-backed invariant depends on 1-to-1 mapping of module balance to erc20s
// if coins can be sent directly to the module account, this assumption is broken.
// this test verifies that coins cannot be directly sent to the module account.
func (suite *invariantTestSuite) TestSendToModuleAccountNotAllowed() {
	bKeeper := suite.App.GetBankKeeper()
	maccAddress := authtypes.NewModuleAddress(types.ModuleName)
	suite.True(bKeeper.BlockedAddr(maccAddress))

	coins := sdk.NewCoins(sdk.NewInt64Coin(suite.cosmosCoin.CosmosDenom, 1e7))
	addr := app.RandomAddress()

	err := suite.App.FundAccount(suite.Ctx, addr, coins)
	suite.NoError(err)

	bankMsgServer := bankkeeper.NewMsgServerImpl(bKeeper)
	_, err = bankMsgServer.Send(suite.Ctx, &banktypes.MsgSend{
		FromAddress: addr.String(),
		ToAddress:   maccAddress.String(),
		Amount:      coins,
	})
	suite.ErrorContains(err, "kava1w9vxuke5dz6hyza2j932qgmxltnfxwl78u920k is not allowed to receive funds: unauthorized")
}

func (suite *invariantTestSuite) TestCosmosCoinsFullyBackedInvariant() {
	invariantName := "cosmos-coins-fully-backed"
	// default state is valid
	_, broken := suite.runInvariant(invariantName, keeper.CosmosCoinsFullyBackedInvariant)
	suite.False(broken)

	suite.SetupValidState()
	_, broken = suite.runInvariant(invariantName, keeper.CosmosCoinsFullyBackedInvariant)
	suite.False(broken)

	// break the invariant by removing module account balance without adjusting token supply
	err := suite.BankKeeper.SendCoinsFromModuleToAccount(
		suite.Ctx,
		types.ModuleName,
		app.RandomAddress(),
		sdk.NewCoins(sdk.NewInt64Coin(suite.cosmosCoin.CosmosDenom, 1e5)),
	)
	suite.NoError(err)

	_, broken = suite.runInvariant(invariantName, keeper.CosmosCoinsFullyBackedInvariant)
	suite.True(broken)
}
