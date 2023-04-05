package keeper_test

import (
	"math/big"
	"testing"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	"github.com/kava-labs/kava/x/evmutil/keeper"
	"github.com/kava-labs/kava/x/evmutil/testutil"
	"github.com/kava-labs/kava/x/evmutil/types"
)

type invariantTestSuite struct {
	testutil.Suite
	invariants   map[string]map[string]sdk.Invariant
	contractAddr types.InternalEVMAddress
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
}

// RegisterRoutes implements sdk.InvariantRegistry
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
