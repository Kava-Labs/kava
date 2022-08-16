package keeper_test

import (
	"testing"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/earn/keeper"
	"github.com/kava-labs/kava/x/earn/testutil"
	"github.com/kava-labs/kava/x/earn/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"
)

type invariantTestSuite struct {
	testutil.Suite

	invariants map[string]map[string]sdk.Invariant
	addrs      []sdk.AccAddress
}

func TestInvariantTestSuite(t *testing.T) {
	suite.Run(t, new(invariantTestSuite))
}

func (suite *invariantTestSuite) SetupTest() {
	suite.Suite.SetupTest()
	suite.invariants = make(map[string]map[string]sdk.Invariant)
	keeper.RegisterInvariants(suite, suite.Keeper)

	_, addrs := app.GeneratePrivKeyAddressPairs(4)
	suite.addrs = addrs
}

func (suite *invariantTestSuite) SetupValidState() {
	suite.Keeper.SetVaultRecord(suite.Ctx, types.NewVaultRecord(
		"usdx",
		sdk.MustNewDecFromStr("100"),
	))
	suite.Keeper.SetVaultRecord(suite.Ctx, types.NewVaultRecord(
		"ukava",
		sdk.MustNewDecFromStr("250.123456"),
	))

	vaultShare1 := types.NewVaultShareRecord(
		suite.addrs[0],
		types.NewVaultShares(
			types.NewVaultShare("usdx", sdk.MustNewDecFromStr("50")),
			types.NewVaultShare("ukava", sdk.MustNewDecFromStr("105.123")),
		),
	)
	vaultShare2 := types.NewVaultShareRecord(
		suite.addrs[1],
		types.NewVaultShares(
			types.NewVaultShare("usdx", sdk.MustNewDecFromStr("50")),
			types.NewVaultShare("ukava", sdk.MustNewDecFromStr("145.000456")),
		),
	)

	suite.Require().NoError(vaultShare1.Validate())
	suite.Require().NoError(vaultShare2.Validate())

	suite.Keeper.SetVaultShareRecord(suite.Ctx, vaultShare1)
	suite.Keeper.SetVaultShareRecord(suite.Ctx, vaultShare2)
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
	suite.Require().Equalf(dBroken, aBroken, "expected all invariant broken to match, direct %v != all %v", dBroken, aBroken)
	if dBroken {
		suite.Require().Equal(dMessage, aMessage, "expected all invariant message to match")
	}

	// return message, broken
	return dMessage, dBroken
}

func (suite *invariantTestSuite) TestVaultRecordsInvariant() {
	// default state is valid
	message, broken := suite.runInvariant("vault-records", keeper.VaultRecordsInvariant)
	suite.Equal("earn: validate vault records broken invariant\nvault record invalid\n", message)
	suite.Equal(false, broken)

	suite.SetupValidState()
	message, broken = suite.runInvariant("vault-records", keeper.VaultRecordsInvariant)
	suite.Equal("earn: validate vault records broken invariant\nvault record invalid\n", message)
	suite.Equal(false, broken)

	// broken with invalid vault record
	suite.Keeper.SetVaultRecord(suite.Ctx, types.VaultRecord{
		TotalShares: types.VaultShare{
			Denom:  "invalid denom",
			Amount: sdk.MustNewDecFromStr("101"),
		},
	})
	message, broken = suite.runInvariant("vault-records", keeper.VaultRecordsInvariant)
	suite.Equal("earn: validate vault records broken invariant\nvault record invalid\n", message)
	suite.Equal(true, broken)
}

func (suite *invariantTestSuite) TestShareRecordsInvariant() {
	message, broken := suite.runInvariant("share-records", keeper.ShareRecordsInvariant)
	suite.Equal("earn: validate share records broken invariant\nshare record invalid\n", message)
	suite.Equal(false, broken)

	suite.SetupValidState()
	message, broken = suite.runInvariant("share-records", keeper.ShareRecordsInvariant)
	suite.Equal("earn: validate share records broken invariant\nshare record invalid\n", message)
	suite.Equal(false, broken)

	// broken with invalid share record
	suite.Keeper.SetVaultShareRecord(suite.Ctx, types.NewVaultShareRecord(
		suite.addrs[0],
		// Directly create vaultshares instead of NewVaultShares() to avoid sanitization
		types.VaultShares{
			types.NewVaultShare("ukava", sdk.MustNewDecFromStr("50")),
			types.NewVaultShare("ukava", sdk.MustNewDecFromStr("105.123")),
		},
	))
	message, broken = suite.runInvariant("share-records", keeper.ShareRecordsInvariant)
	suite.Equal("earn: validate share records broken invariant\nshare record invalid\n", message)
	suite.Equal(true, broken)
}

func (suite *invariantTestSuite) TestVaultSharesInvariant() {
	message, broken := suite.runInvariant("vault-shares", keeper.VaultSharesInvariant)
	suite.Equal("earn: vault shares broken invariant\nvault shares do not match depositor shares\n", message)
	suite.Equal(false, broken)

	suite.SetupValidState()
	message, broken = suite.runInvariant("vault-shares", keeper.VaultSharesInvariant)
	suite.Equal("earn: vault shares broken invariant\nvault shares do not match depositor shares\n", message)
	suite.Equal(false, broken)

	// broken when total shares are greater than depositor shares
	suite.Keeper.SetVaultRecord(suite.Ctx, types.NewVaultRecord(
		"usdx",
		sdk.MustNewDecFromStr("101"),
	))
	message, broken = suite.runInvariant("vault-shares", keeper.VaultSharesInvariant)
	suite.Equal("earn: vault shares broken invariant\nvault shares do not match depositor shares\n", message)
	suite.Equal(true, broken)

	// broken when total shares are less than the depositor shares
	suite.Keeper.SetVaultRecord(suite.Ctx, types.NewVaultRecord(
		"usdx",
		sdk.MustNewDecFromStr("99.999"),
	))
	message, broken = suite.runInvariant("vault-shares", keeper.VaultSharesInvariant)
	suite.Equal("earn: vault shares broken invariant\nvault shares do not match depositor shares\n", message)
	suite.Equal(true, broken)

	// broken when vault record is missing
	suite.Keeper.DeleteVaultRecord(suite.Ctx, "usdx")
	message, broken = suite.runInvariant("vault-shares", keeper.VaultSharesInvariant)
	suite.Equal("earn: vault shares broken invariant\nvault shares do not match depositor shares\n", message)
	suite.Equal(true, broken)
}
