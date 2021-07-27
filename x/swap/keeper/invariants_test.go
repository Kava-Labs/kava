package keeper_test

import (
	"testing"

	"github.com/kava-labs/kava/x/swap/keeper"
	"github.com/kava-labs/kava/x/swap/testutil"
	"github.com/kava-labs/kava/x/swap/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"
)

type invariantTestSuite struct {
	testutil.Suite
	invariants map[string]map[string]sdk.Invariant
}

func (suite *invariantTestSuite) SetupTest() {
	suite.Suite.SetupTest()
	suite.invariants = make(map[string]map[string]sdk.Invariant)
	keeper.RegisterInvariants(suite, suite.Keeper)
}

func (suite *invariantTestSuite) SetupValidState() {
	suite.Keeper.SetPool(suite.Ctx, types.NewPoolRecord(
		sdk.NewCoins(
			sdk.NewCoin("ukava", sdk.NewInt(1e6)),
			sdk.NewCoin("usdx", sdk.NewInt(5e6)),
		),
		sdk.NewInt(3e6),
	))
	suite.AddCoinsToModule(
		sdk.NewCoins(
			sdk.NewCoin("ukava", sdk.NewInt(1e6)),
			sdk.NewCoin("usdx", sdk.NewInt(5e6)),
		),
	)
	suite.Keeper.SetDepositorShares(suite.Ctx, types.NewShareRecord(
		sdk.AccAddress("depositor 1"),
		types.PoolID("ukava", "usdx"),
		sdk.NewInt(2e6),
	))
	suite.Keeper.SetDepositorShares(suite.Ctx, types.NewShareRecord(
		sdk.AccAddress("depositor 2"),
		types.PoolID("ukava", "usdx"),
		sdk.NewInt(1e6),
	))

	suite.Keeper.SetPool(suite.Ctx, types.NewPoolRecord(
		sdk.NewCoins(
			sdk.NewCoin("hard", sdk.NewInt(1e6)),
			sdk.NewCoin("usdx", sdk.NewInt(2e6)),
		),
		sdk.NewInt(1e6),
	))
	suite.AddCoinsToModule(
		sdk.NewCoins(
			sdk.NewCoin("hard", sdk.NewInt(1e6)),
			sdk.NewCoin("usdx", sdk.NewInt(2e6)),
		),
	)
	suite.Keeper.SetDepositorShares(suite.Ctx, types.NewShareRecord(
		sdk.AccAddress("depositor 1"),
		types.PoolID("hard", "usdx"),
		sdk.NewInt(1e6),
	))
}

func (suite *invariantTestSuite) RegisterRoute(moduleName string, route string, invariant sdk.Invariant) {
	_, exists := suite.invariants[moduleName]

	if !exists {
		suite.invariants[moduleName] = make(map[string]sdk.Invariant)
	}

	suite.invariants[moduleName][route] = invariant
}

func (suite *invariantTestSuite) runInvariant(route string, invariant func(k keeper.Keeper) sdk.Invariant) (string, bool) {
	ctx := suite.Ctx
	registeredInvariant := suite.invariants[types.ModuleName][route]
	suite.Require().NotNil(registeredInvariant)

	// direct call
	dMessage, dBroken := invariant(suite.Keeper)(ctx)
	// registered call
	rMessage, rBroken := registeredInvariant(ctx)
	// all call
	aMessage, aBroken := keeper.AllInvariants(suite.Keeper)(ctx)

	// require matching values for direct call and registered call
	suite.Require().Equal(dMessage, rMessage, "expected registered invariant message to match")
	suite.Require().Equal(dBroken, rBroken, "expected registered invariant broken to match")
	// require matching values for direct call and all invariants call if broken
	suite.Require().Equal(dBroken, aBroken, "expected all invariant broken to match")
	if dBroken {
		suite.Require().Equal(dMessage, aMessage, "expected all invariant message to match")
	}

	// return message, broken
	return dMessage, dBroken
}

func (suite *invariantTestSuite) TestPoolRecordsInvariant() {

	// default state is valid
	message, broken := suite.runInvariant("pool-records", keeper.PoolRecordsInvariant)
	suite.Equal("swap: validate pool records broken invariant\npool record invalid\n", message)
	suite.Equal(false, broken)

	suite.SetupValidState()
	message, broken = suite.runInvariant("pool-records", keeper.PoolRecordsInvariant)
	suite.Equal("swap: validate pool records broken invariant\npool record invalid\n", message)
	suite.Equal(false, broken)

	// broken with invalid pool record
	suite.Keeper.SetPool_Raw(suite.Ctx, types.NewPoolRecord(
		sdk.NewCoins(
			sdk.NewCoin("ukava", sdk.NewInt(1e6)),
			sdk.NewCoin("usdx", sdk.NewInt(5e6)),
		),
		sdk.NewInt(-1e6),
	))
	message, broken = suite.runInvariant("pool-records", keeper.PoolRecordsInvariant)
	suite.Equal("swap: validate pool records broken invariant\npool record invalid\n", message)
	suite.Equal(true, broken)
}

func (suite *invariantTestSuite) TestShareRecordsInvariant() {
	message, broken := suite.runInvariant("share-records", keeper.ShareRecordsInvariant)
	suite.Equal("swap: validate share records broken invariant\nshare record invalid\n", message)
	suite.Equal(false, broken)

	suite.SetupValidState()
	message, broken = suite.runInvariant("share-records", keeper.ShareRecordsInvariant)
	suite.Equal("swap: validate share records broken invariant\nshare record invalid\n", message)
	suite.Equal(false, broken)

	// broken with invalid share record
	suite.Keeper.SetDepositorShares_Raw(suite.Ctx, types.NewShareRecord(
		sdk.AccAddress("depositor 1"),
		types.PoolID("ukava", "usdx"),
		sdk.NewInt(-1e6),
	))
	message, broken = suite.runInvariant("share-records", keeper.ShareRecordsInvariant)
	suite.Equal("swap: validate share records broken invariant\nshare record invalid\n", message)
	suite.Equal(true, broken)
}

func (suite *invariantTestSuite) TestPoolReservesInvariant() {
	message, broken := suite.runInvariant("pool-reserves", keeper.PoolReservesInvariant)
	suite.Equal("swap: pool reserves broken invariant\npool reserves do not match module account\n", message)
	suite.Equal(false, broken)

	suite.SetupValidState()
	message, broken = suite.runInvariant("pool-reserves", keeper.PoolReservesInvariant)
	suite.Equal("swap: pool reserves broken invariant\npool reserves do not match module account\n", message)
	suite.Equal(false, broken)

	// broken when reserves are greater than module balance
	suite.Keeper.SetPool(suite.Ctx, types.NewPoolRecord(
		sdk.NewCoins(
			sdk.NewCoin("ukava", sdk.NewInt(2e6)),
			sdk.NewCoin("usdx", sdk.NewInt(10e6)),
		),
		sdk.NewInt(5e6),
	))
	message, broken = suite.runInvariant("pool-reserves", keeper.PoolReservesInvariant)
	suite.Equal("swap: pool reserves broken invariant\npool reserves do not match module account\n", message)
	suite.Equal(true, broken)

	// broken when reserves are less than the module balance
	suite.Keeper.SetPool(suite.Ctx, types.NewPoolRecord(
		sdk.NewCoins(
			sdk.NewCoin("ukava", sdk.NewInt(1e5)),
			sdk.NewCoin("usdx", sdk.NewInt(5e5)),
		),
		sdk.NewInt(3e5),
	))
	message, broken = suite.runInvariant("pool-reserves", keeper.PoolReservesInvariant)
	suite.Equal("swap: pool reserves broken invariant\npool reserves do not match module account\n", message)
	suite.Equal(true, broken)
}

func (suite *invariantTestSuite) TestPoolSharesInvariant() {
	message, broken := suite.runInvariant("pool-shares", keeper.PoolSharesInvariant)
	suite.Equal("swap: pool shares broken invariant\npool shares do not match depositor shares\n", message)
	suite.Equal(false, broken)

	suite.SetupValidState()
	message, broken = suite.runInvariant("pool-shares", keeper.PoolSharesInvariant)
	suite.Equal("swap: pool shares broken invariant\npool shares do not match depositor shares\n", message)
	suite.Equal(false, broken)

	// broken when total shares are greater than depositor shares
	suite.Keeper.SetPool(suite.Ctx, types.NewPoolRecord(
		sdk.NewCoins(
			sdk.NewCoin("ukava", sdk.NewInt(1e6)),
			sdk.NewCoin("usdx", sdk.NewInt(5e6)),
		),
		sdk.NewInt(5e6),
	))
	message, broken = suite.runInvariant("pool-shares", keeper.PoolSharesInvariant)
	suite.Equal("swap: pool shares broken invariant\npool shares do not match depositor shares\n", message)
	suite.Equal(true, broken)

	// broken when total shares are less than the depositor shares
	suite.Keeper.SetPool(suite.Ctx, types.NewPoolRecord(
		sdk.NewCoins(
			sdk.NewCoin("ukava", sdk.NewInt(1e6)),
			sdk.NewCoin("usdx", sdk.NewInt(5e6)),
		),
		sdk.NewInt(1e5),
	))
	message, broken = suite.runInvariant("pool-shares", keeper.PoolSharesInvariant)
	suite.Equal("swap: pool shares broken invariant\npool shares do not match depositor shares\n", message)
	suite.Equal(true, broken)

	// broken when pool record is missing
	suite.Keeper.DeletePool(suite.Ctx, types.PoolID("ukava", "usdx"))
	suite.RemoveCoinsFromModule(
		sdk.NewCoins(
			sdk.NewCoin("ukava", sdk.NewInt(1e6)),
			sdk.NewCoin("usdx", sdk.NewInt(5e6)),
		),
	)
	message, broken = suite.runInvariant("pool-shares", keeper.PoolSharesInvariant)
	suite.Equal("swap: pool shares broken invariant\npool shares do not match depositor shares\n", message)
	suite.Equal(true, broken)
}

func TestInvariantTestSuite(t *testing.T) {
	suite.Run(t, new(invariantTestSuite))
}
