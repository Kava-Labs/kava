package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/kava-labs/kava/x/incentive/types"
)

// InitializeHardBorrowRewardTests runs unit tests for the keeper.InitializeHardBorrowReward method
type InitializeHardBorrowRewardTests struct {
	unitTester
}

func TestInitializeHardBorrowReward(t *testing.T) {
	suite.Run(t, new(InitializeHardBorrowRewardTests))
}

func (suite *InitializeHardBorrowRewardTests) TestClaimIndexesAreSetWhenClaimExists() {
	claim := types.HardLiquidityProviderClaim{
		BaseMultiClaim: types.BaseMultiClaim{
			Owner: arbitraryAddress(),
		},
		// Indexes should always be empty when initialize is called.
		// If initialize is called then the user must have repaid their borrow positions,
		// which means UpdateHardBorrowIndexDenoms was called and should have remove indexes.
		BorrowRewardIndexes: types.MultiRewardIndexes{},
	}
	suite.storeHardClaim(claim)

	globalIndexes := nonEmptyMultiRewardIndexes
	suite.storeGlobalBorrowIndexes(globalIndexes)

	borrow := NewBorrowBuilder(claim.Owner).
		WithArbitrarySourceShares(extractCollateralTypes(globalIndexes)...).
		Build()

	suite.keeper.InitializeHardBorrowReward(suite.ctx, borrow)

	syncedClaim, _ := suite.keeper.GetHardLiquidityProviderClaim(suite.ctx, claim.Owner)
	suite.Equal(globalIndexes, syncedClaim.BorrowRewardIndexes)
}
func (suite *InitializeHardBorrowRewardTests) TestClaimIndexesAreSetWhenClaimDoesNotExist() {
	globalIndexes := nonEmptyMultiRewardIndexes
	suite.storeGlobalBorrowIndexes(globalIndexes)

	owner := arbitraryAddress()
	borrow := NewBorrowBuilder(owner).
		WithArbitrarySourceShares(extractCollateralTypes(globalIndexes)...).
		Build()

	suite.keeper.InitializeHardBorrowReward(suite.ctx, borrow)

	syncedClaim, found := suite.keeper.GetHardLiquidityProviderClaim(suite.ctx, owner)
	suite.True(found)
	suite.Equal(globalIndexes, syncedClaim.BorrowRewardIndexes)
}
func (suite *InitializeHardBorrowRewardTests) TestClaimIndexesAreSetEmptyForMissingIndexes() {

	globalIndexes := nonEmptyMultiRewardIndexes
	suite.storeGlobalBorrowIndexes(globalIndexes)

	owner := arbitraryAddress()
	// Borrow a denom that is not in the global indexes.
	// This happens when a borrow denom has no rewards associated with it.
	expectedIndexes := appendUniqueEmptyMultiRewardIndex(globalIndexes)
	borrowedDenoms := extractCollateralTypes(expectedIndexes)
	borrow := NewBorrowBuilder(owner).
		WithArbitrarySourceShares(borrowedDenoms...).
		Build()

	suite.keeper.InitializeHardBorrowReward(suite.ctx, borrow)

	syncedClaim, _ := suite.keeper.GetHardLiquidityProviderClaim(suite.ctx, owner)
	suite.Equal(expectedIndexes, syncedClaim.BorrowRewardIndexes)
}
