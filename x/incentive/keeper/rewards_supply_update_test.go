package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	hardtypes "github.com/kava-labs/kava/x/hard/types"
	"github.com/kava-labs/kava/x/incentive/types"
)

// UpdateHardSupplyIndexDenomsTests runs unit tests for the keeper.UpdateHardSupplyIndexDenoms method
//
// inputs
// - claim in store if it exists (only claim.SupplyRewardIndexes)
// - global indexes in store
// - deposit function arg (only deposit.Amount)
//
// outputs
// - sets a claim
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
	suite.storeClaim(claim)
	suite.storeGlobalSupplyIndexes(claim.SupplyRewardIndexes)

	// remove one denom from the indexes already in the deposit
	expectedIndexes := claim.SupplyRewardIndexes[1:]
	deposit := hardtypes.Deposit{
		Depositor: claim.Owner,
		Amount:    arbitraryCoinsWithDenoms(extractCollateralTypes(expectedIndexes)...),
	}

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
	suite.storeClaim(claim)
	globalIndexes := appendUniqueMultiRewardIndex(claim.SupplyRewardIndexes)
	suite.storeGlobalSupplyIndexes(globalIndexes)

	deposit := hardtypes.Deposit{
		Depositor: claim.Owner,
		Amount:    arbitraryCoinsWithDenoms(extractCollateralTypes(globalIndexes)...),
	}

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
	suite.storeClaim(claim)
	// Set global indexes with same denoms but different values.
	// UpdateHardSupplyIndexDenoms should ignore the new values.
	suite.storeGlobalSupplyIndexes(increaseAllRewardFactors(claim.SupplyRewardIndexes))

	deposit := hardtypes.Deposit{
		Depositor: claim.Owner,
		Amount:    arbitraryCoinsWithDenoms(extractCollateralTypes(claim.SupplyRewardIndexes)...),
	}

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
	suite.storeClaim(claim)
	suite.storeGlobalSupplyIndexes(claim.SupplyRewardIndexes)

	// add a denom to the deposited amount that is not in the global or claim's indexes
	expectedIndexes := appendUniqueEmptyMultiRewardIndex(claim.SupplyRewardIndexes)
	depositedDenoms := extractCollateralTypes(expectedIndexes)
	deposit := hardtypes.Deposit{
		Depositor: claim.Owner,
		Amount:    arbitraryCoinsWithDenoms(depositedDenoms...),
	}

	suite.keeper.UpdateHardSupplyIndexDenoms(suite.ctx, deposit)

	syncedClaim, _ := suite.keeper.GetHardLiquidityProviderClaim(suite.ctx, claim.Owner)
	suite.Equal(expectedIndexes, syncedClaim.SupplyRewardIndexes)
}
