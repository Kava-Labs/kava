package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/kava-labs/kava/x/incentive/types"
)

// InitializeHardSupplyRewardTests runs unit tests for the keeper.InitializeHardSupplyReward method
type InitializeHardSupplyRewardTests struct {
	unitTester
}

func TestInitializeHardSupplyReward(t *testing.T) {
	suite.Run(t, new(InitializeHardSupplyRewardTests))
}

func (suite *InitializeHardSupplyRewardTests) TestClaimIndexesAreSetWhenClaimExists() {
	claim := types.HardLiquidityProviderClaim{
		BaseMultiClaim: types.BaseMultiClaim{
			Owner: arbitraryAddress(),
		},
		// Indexes should always be empty when initialize is called.
		// If initialize is called then the user must have repaid their deposit positions,
		// which means UpdateHardSupplyIndexDenoms was called and should have remove indexes.
		SupplyRewardIndexes: types.MultiRewardIndexes{},
	}
	suite.storeHardClaim(claim)

	globalIndexes := nonEmptyMultiRewardIndexes
	suite.storeGlobalSupplyIndexes(globalIndexes)

	deposit := NewDepositBuilder(claim.Owner).
		WithArbitrarySourceShares(extractCollateralTypes(globalIndexes)...).
		Build()

	suite.keeper.InitializeHardSupplyReward(suite.ctx, deposit)

	syncedClaim, _ := suite.keeper.GetHardLiquidityProviderClaim(suite.ctx, claim.Owner)
	suite.Equal(globalIndexes, syncedClaim.SupplyRewardIndexes)
}
func (suite *InitializeHardSupplyRewardTests) TestClaimIndexesAreSetWhenClaimDoesNotExist() {
	globalIndexes := nonEmptyMultiRewardIndexes
	suite.storeGlobalSupplyIndexes(globalIndexes)

	owner := arbitraryAddress()
	deposit := NewDepositBuilder(owner).
		WithArbitrarySourceShares(extractCollateralTypes(globalIndexes)...).
		Build()

	suite.keeper.InitializeHardSupplyReward(suite.ctx, deposit)

	syncedClaim, found := suite.keeper.GetHardLiquidityProviderClaim(suite.ctx, owner)
	suite.True(found)
	suite.Equal(globalIndexes, syncedClaim.SupplyRewardIndexes)
}
func (suite *InitializeHardSupplyRewardTests) TestClaimIndexesAreSetEmptyForMissingIndexes() {

	globalIndexes := nonEmptyMultiRewardIndexes
	suite.storeGlobalSupplyIndexes(globalIndexes)

	owner := arbitraryAddress()
	// Supply a denom that is not in the global indexes.
	// This happens when a deposit denom has no rewards associated with it.
	expectedIndexes := appendUniqueEmptyMultiRewardIndex(globalIndexes)
	depositedDenoms := extractCollateralTypes(expectedIndexes)
	deposit := NewDepositBuilder(owner).
		WithArbitrarySourceShares(depositedDenoms...).
		Build()

	suite.keeper.InitializeHardSupplyReward(suite.ctx, deposit)

	syncedClaim, _ := suite.keeper.GetHardLiquidityProviderClaim(suite.ctx, owner)
	suite.Equal(expectedIndexes, syncedClaim.SupplyRewardIndexes)
}
