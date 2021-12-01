package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/kava-labs/kava/x/incentive/types"
)

// UpdateHardSupplyIndexDenomsTests runs unit tests for the keeper.UpdateHardSupplyIndexDenoms method
type UpdateHardSupplyIndexDenomsTests struct {
	unitTester
}

func TestUpdateHardSupplyIndexDenoms(t *testing.T) {
	suite.Run(t, new(UpdateHardSupplyIndexDenomsTests))
}

func (suite *UpdateHardSupplyIndexDenomsTests) TestClaimIndexesAreRemovedForDenomsNoLongerSupplied() {
	claim := types.HardLiquidityProviderClaim{
		BaseMultiClaim: types.BaseMultiClaim{
			Owner: arbitraryAddress(),
		},
		SupplyRewardIndexes: nonEmptyMultiRewardIndexes,
	}
	suite.storeHardClaim(claim)
	suite.storeGlobalSupplyIndexes(claim.SupplyRewardIndexes)

	// remove one denom from the indexes already in the deposit
	expectedIndexes := claim.SupplyRewardIndexes[1:]
	deposit := NewDepositBuilder(claim.Owner).
		WithArbitrarySourceShares(extractCollateralTypes(expectedIndexes)...).
		Build()

	suite.keeper.UpdateHardSupplyIndexDenoms(suite.ctx, deposit)

	syncedClaim, _ := suite.keeper.GetHardLiquidityProviderClaim(suite.ctx, claim.Owner)
	suite.Equal(expectedIndexes, syncedClaim.SupplyRewardIndexes)
}

func (suite *UpdateHardSupplyIndexDenomsTests) TestClaimIndexesAreAddedForNewlySuppliedDenoms() {
	claim := types.HardLiquidityProviderClaim{
		BaseMultiClaim: types.BaseMultiClaim{
			Owner: arbitraryAddress(),
		},
		SupplyRewardIndexes: nonEmptyMultiRewardIndexes,
	}
	suite.storeHardClaim(claim)
	globalIndexes := appendUniqueMultiRewardIndex(claim.SupplyRewardIndexes)
	suite.storeGlobalSupplyIndexes(globalIndexes)

	deposit := NewDepositBuilder(claim.Owner).
		WithArbitrarySourceShares(extractCollateralTypes(globalIndexes)...).
		Build()

	suite.keeper.UpdateHardSupplyIndexDenoms(suite.ctx, deposit)

	syncedClaim, _ := suite.keeper.GetHardLiquidityProviderClaim(suite.ctx, claim.Owner)
	suite.Equal(globalIndexes, syncedClaim.SupplyRewardIndexes)
}

func (suite *UpdateHardSupplyIndexDenomsTests) TestClaimIndexesAreUnchangedWhenSuppliedDenomsUnchanged() {
	claim := types.HardLiquidityProviderClaim{
		BaseMultiClaim: types.BaseMultiClaim{
			Owner: arbitraryAddress(),
		},
		SupplyRewardIndexes: nonEmptyMultiRewardIndexes,
	}
	suite.storeHardClaim(claim)
	// Set global indexes with same denoms but different values.
	// UpdateHardSupplyIndexDenoms should ignore the new values.
	suite.storeGlobalSupplyIndexes(increaseAllRewardFactors(claim.SupplyRewardIndexes))

	deposit := NewDepositBuilder(claim.Owner).
		WithArbitrarySourceShares(extractCollateralTypes(claim.SupplyRewardIndexes)...).
		Build()

	suite.keeper.UpdateHardSupplyIndexDenoms(suite.ctx, deposit)

	syncedClaim, _ := suite.keeper.GetHardLiquidityProviderClaim(suite.ctx, claim.Owner)
	suite.Equal(claim.SupplyRewardIndexes, syncedClaim.SupplyRewardIndexes)
}

func (suite *UpdateHardSupplyIndexDenomsTests) TestEmptyClaimIndexesAreAddedForNewlySuppliedButNotRewardedDenoms() {
	claim := types.HardLiquidityProviderClaim{
		BaseMultiClaim: types.BaseMultiClaim{
			Owner: arbitraryAddress(),
		},
		SupplyRewardIndexes: nonEmptyMultiRewardIndexes,
	}
	suite.storeHardClaim(claim)
	suite.storeGlobalSupplyIndexes(claim.SupplyRewardIndexes)

	// add a denom to the deposited amount that is not in the global or claim's indexes
	expectedIndexes := appendUniqueEmptyMultiRewardIndex(claim.SupplyRewardIndexes)
	depositedDenoms := extractCollateralTypes(expectedIndexes)
	deposit := NewDepositBuilder(claim.Owner).
		WithArbitrarySourceShares(depositedDenoms...).
		Build()

	suite.keeper.UpdateHardSupplyIndexDenoms(suite.ctx, deposit)

	syncedClaim, _ := suite.keeper.GetHardLiquidityProviderClaim(suite.ctx, claim.Owner)
	suite.Equal(expectedIndexes, syncedClaim.SupplyRewardIndexes)
}
