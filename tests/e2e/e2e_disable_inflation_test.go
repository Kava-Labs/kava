package e2e_test

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/kava-labs/kava/tests/util"
	communitytypes "github.com/kava-labs/kava/x/community/types"
	kavadisttypes "github.com/kava-labs/kava/x/kavadist/types"
)

// TestUpgradeHandler can be used to run tests post-upgrade. If an upgrade is enabled, all tests
// are run against the upgraded chain. However, this file is a good place to consolidate all
// acceptance tests for a given set of upgrade handlers.
func (suite *IntegrationTestSuite) TestDisableInflation() {
	// suite.SkipIfUpgradeDisabled()
	fmt.Println("An upgrade has run!")
	// suite.True(true)

	beforeUpgradeCtx := util.CtxAtHeight(suite.UpgradeHeight - 1)
	afterUpgradeCtx := util.CtxAtHeight(suite.UpgradeHeight)

	// Before balances
	kavaDistBalBefore, err := suite.Kava.Kavadist.Balance(beforeUpgradeCtx, &kavadisttypes.QueryBalanceRequest{})
	suite.NoError(err)
	distrBalBefore, err := suite.Kava.Distribution.CommunityPool(beforeUpgradeCtx, &distrtypes.QueryCommunityPoolRequest{})
	suite.NoError(err)
	distrBalCoinsBefore, distrBalDustBefore := distrBalBefore.Pool.TruncateDecimal()
	beforeCommPoolBalance, err := suite.Kava.Community.Balance(beforeUpgradeCtx, &communitytypes.QueryBalanceRequest{})
	suite.NoError(err)

	// After balances
	kavaDistBalAfter, err := suite.Kava.Kavadist.Balance(afterUpgradeCtx, &kavadisttypes.QueryBalanceRequest{})
	suite.NoError(err)
	distrBalAfter, err := suite.Kava.Distribution.CommunityPool(afterUpgradeCtx, &distrtypes.QueryCommunityPoolRequest{})
	suite.NoError(err)
	afterCommPoolBalance, err := suite.Kava.Community.Balance(afterUpgradeCtx, &communitytypes.QueryBalanceRequest{})
	suite.NoError(err)

	// expect empty balances after (ignoring dust in x/distribution)
	suite.Equal(sdk.NewCoins(), kavaDistBalAfter.Coins)
	distrCoinsAfter, distrBalDustAfter := distrBalAfter.Pool.TruncateDecimal()
	suite.Equal(sdk.NewCoins(), distrCoinsAfter)

	// x/kavadist and x/distribution community pools should be moved to x/community
	suite.Equal(
		beforeCommPoolBalance.Coins.
			Add(kavaDistBalBefore.Coins...).
			Add(distrBalCoinsBefore...),
		afterCommPoolBalance.Coins,
	)

	// x/distribution dust should stay in x/distribution
	suite.Equal(distrBalDustBefore, distrBalDustAfter)
}
