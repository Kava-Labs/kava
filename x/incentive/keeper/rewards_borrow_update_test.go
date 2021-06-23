package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	hardtypes "github.com/kava-labs/kava/x/hard/types"
	"github.com/kava-labs/kava/x/incentive/types"
)

// UpdateHardBorrowIndexDenomsTests runs unit tests for the keeper.UpdateHardBorrowIndexDenoms method
//
// inputs
// - claim in store if it exists (only claim.BorrowRewardIndexes)
// - global indexes in store
// - borrow function arg (only borrow.Amount)
//
// outputs
// - sets a claim
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
	suite.storeClaim(claim)
	suite.storeGlobalBorrowIndexes(claim.BorrowRewardIndexes)

	// remove one denom from the indexes already in the borrow
	expectedIndexes := claim.BorrowRewardIndexes[1:]
	borrow := hardtypes.Borrow{
		Borrower: claim.Owner,
		Amount:   arbitraryCoinsWithDenoms(extractCollateralTypes(expectedIndexes)...),
	}

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
	suite.storeClaim(claim)
	globalIndexes := appendUniqueMultiRewardIndex(claim.BorrowRewardIndexes)
	suite.storeGlobalBorrowIndexes(globalIndexes)

	borrow := hardtypes.Borrow{
		Borrower: claim.Owner,
		Amount:   arbitraryCoinsWithDenoms(extractCollateralTypes(globalIndexes)...),
	}

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
	suite.storeClaim(claim)
	// Set global indexes with same denoms but different values.
	// UpdateHardBorrowIndexDenoms should ignore the new values.
	suite.storeGlobalBorrowIndexes(increaseAllRewardFactors(claim.BorrowRewardIndexes))

	borrow := hardtypes.Borrow{
		Borrower: claim.Owner,
		Amount:   arbitraryCoinsWithDenoms(extractCollateralTypes(claim.BorrowRewardIndexes)...),
	}

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
	suite.storeClaim(claim)
	suite.storeGlobalBorrowIndexes(claim.BorrowRewardIndexes)

	// add a denom to the borrowed amount that is not in the global or claim's indexes
	expectedIndexes := appendUniqueEmptyMultiRewardIndex(claim.BorrowRewardIndexes)
	borrowedDenoms := extractCollateralTypes(expectedIndexes)
	borrow := hardtypes.Borrow{
		Borrower: claim.Owner,
		Amount:   arbitraryCoinsWithDenoms(borrowedDenoms...),
	}

	suite.keeper.UpdateHardBorrowIndexDenoms(suite.ctx, borrow)

	syncedClaim, _ := suite.keeper.GetHardLiquidityProviderClaim(suite.ctx, claim.Owner)
	suite.Equal(expectedIndexes, syncedClaim.BorrowRewardIndexes)
}
