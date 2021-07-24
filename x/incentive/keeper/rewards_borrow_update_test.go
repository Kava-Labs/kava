package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/kava-labs/kava/x/incentive/types"
)

// UpdateHardBorrowIndexDenomsTests runs unit tests for the keeper.UpdateHardBorrowIndexDenoms method
type UpdateHardBorrowIndexDenomsTests struct {
	unitTester
}

func TestUpdateHardBorrowIndexDenoms(t *testing.T) {
	suite.Run(t, new(UpdateHardBorrowIndexDenomsTests))
}

func (suite *UpdateHardBorrowIndexDenomsTests) TestClaimIndexesAreRemovedForDenomsNoLongerBorrowed() {
	claim := types.HardLiquidityProviderClaim{
		BaseMultiClaim: types.BaseMultiClaim{
			Owner: arbitraryAddress(),
		},
		BorrowRewardIndexes: nonEmptyMultiRewardIndexes,
	}
	suite.storeHardClaim(claim)
	suite.storeGlobalBorrowIndexes(claim.BorrowRewardIndexes)

	// remove one denom from the indexes already in the borrow
	expectedIndexes := claim.BorrowRewardIndexes[1:]
	borrow := NewBorrowBuilder(claim.Owner).
		WithArbitrarySourceShares(extractCollateralTypes(expectedIndexes)...).
		Build()

	suite.keeper.UpdateHardBorrowIndexDenoms(suite.ctx, borrow)

	syncedClaim, _ := suite.keeper.GetHardLiquidityProviderClaim(suite.ctx, claim.Owner)
	suite.Equal(expectedIndexes, syncedClaim.BorrowRewardIndexes)
}

func (suite *UpdateHardBorrowIndexDenomsTests) TestClaimIndexesAreAddedForNewlyBorrowedDenoms() {
	claim := types.HardLiquidityProviderClaim{
		BaseMultiClaim: types.BaseMultiClaim{
			Owner: arbitraryAddress(),
		},
		BorrowRewardIndexes: nonEmptyMultiRewardIndexes,
	}
	suite.storeHardClaim(claim)
	globalIndexes := appendUniqueMultiRewardIndex(claim.BorrowRewardIndexes)
	suite.storeGlobalBorrowIndexes(globalIndexes)

	borrow := NewBorrowBuilder(claim.Owner).
		WithArbitrarySourceShares(extractCollateralTypes(globalIndexes)...).
		Build()

	suite.keeper.UpdateHardBorrowIndexDenoms(suite.ctx, borrow)

	syncedClaim, _ := suite.keeper.GetHardLiquidityProviderClaim(suite.ctx, claim.Owner)
	suite.Equal(globalIndexes, syncedClaim.BorrowRewardIndexes)
}

func (suite *UpdateHardBorrowIndexDenomsTests) TestClaimIndexesAreUnchangedWhenBorrowedDenomsUnchanged() {
	claim := types.HardLiquidityProviderClaim{
		BaseMultiClaim: types.BaseMultiClaim{
			Owner: arbitraryAddress(),
		},
		BorrowRewardIndexes: nonEmptyMultiRewardIndexes,
	}
	suite.storeHardClaim(claim)
	// Set global indexes with same denoms but different values.
	// UpdateHardBorrowIndexDenoms should ignore the new values.
	suite.storeGlobalBorrowIndexes(increaseAllRewardFactors(claim.BorrowRewardIndexes))

	borrow := NewBorrowBuilder(claim.Owner).
		WithArbitrarySourceShares(extractCollateralTypes(claim.BorrowRewardIndexes)...).
		Build()

	suite.keeper.UpdateHardBorrowIndexDenoms(suite.ctx, borrow)

	syncedClaim, _ := suite.keeper.GetHardLiquidityProviderClaim(suite.ctx, claim.Owner)
	suite.Equal(claim.BorrowRewardIndexes, syncedClaim.BorrowRewardIndexes)
}

func (suite *UpdateHardBorrowIndexDenomsTests) TestEmptyClaimIndexesAreAddedForNewlyBorrowedButNotRewardedDenoms() {
	claim := types.HardLiquidityProviderClaim{
		BaseMultiClaim: types.BaseMultiClaim{
			Owner: arbitraryAddress(),
		},
		BorrowRewardIndexes: nonEmptyMultiRewardIndexes,
	}
	suite.storeHardClaim(claim)
	suite.storeGlobalBorrowIndexes(claim.BorrowRewardIndexes)

	// add a denom to the borrowed amount that is not in the global or claim's indexes
	expectedIndexes := appendUniqueEmptyMultiRewardIndex(claim.BorrowRewardIndexes)
	borrowedDenoms := extractCollateralTypes(expectedIndexes)
	borrow := NewBorrowBuilder(claim.Owner).
		WithArbitrarySourceShares(borrowedDenoms...).
		Build()

	suite.keeper.UpdateHardBorrowIndexDenoms(suite.ctx, borrow)

	syncedClaim, _ := suite.keeper.GetHardLiquidityProviderClaim(suite.ctx, claim.Owner)
	suite.Equal(expectedIndexes, syncedClaim.BorrowRewardIndexes)
}
