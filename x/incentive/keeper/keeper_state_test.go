package keeper_test

import (
	"github.com/kava-labs/kava/x/incentive/types"
)

func (suite *KeeperTestSuite) TestGetSetDeleteClaims() {
	suite.SetupApp()

	for claimTypeName, claimTypeValue := range types.ClaimType_value {
		suite.Run(claimTypeName, func() {
			claimType := types.ClaimType(claimTypeValue)

			c := types.NewClaim(
				claimType,
				suite.addrs[0],
				arbitraryCoins(),
				nonEmptyMultiRewardIndexes,
			)

			_, found := suite.keeper.GetClaim(suite.ctx, claimType, suite.addrs[0])
			suite.Require().False(found)

			suite.Require().NotPanics(func() {
				suite.keeper.SetClaim(suite.ctx, c)
			})
			testC, found := suite.keeper.GetClaim(suite.ctx, claimType, suite.addrs[0])
			suite.Require().True(found)
			suite.Require().Equal(c, testC)

			// Check that other claim types do not exist for the same address
			for otherClaimTypeName, otherClaimTypeValue := range types.ClaimType_value {
				// Skip the current claim type
				if otherClaimTypeValue == claimTypeValue {
					continue
				}

				otherClaimType := types.ClaimType(otherClaimTypeValue)
				_, found := suite.keeper.GetClaim(suite.ctx, otherClaimType, suite.addrs[0])
				suite.Require().False(found, "claim type %s should not exist", otherClaimTypeName)
			}

			suite.Require().NotPanics(func() {
				suite.keeper.DeleteClaim(suite.ctx, claimType, suite.addrs[0])
			})
			_, found = suite.keeper.GetClaim(suite.ctx, claimType, suite.addrs[0])
			suite.Require().False(found)
		})
	}
}

func (suite *KeeperTestSuite) TestIterateClaims() {
	suite.SetupApp()
	var claims types.Claims

	// Add 2 of each type of claim
	for _, claimTypeValue := range types.ClaimType_value {
		c := types.Claims{
			types.NewClaim(types.ClaimType(claimTypeValue), suite.addrs[0], arbitraryCoins(), nonEmptyMultiRewardIndexes),
			types.NewClaim(types.ClaimType(claimTypeValue), suite.addrs[1], nil, nil),
		}
		claims = append(claims, c...)
	}

	for _, claim := range claims {
		suite.keeper.SetClaim(suite.ctx, claim)
	}

	for _, claimTypeValue := range types.ClaimType_value {
		claimType := types.ClaimType(claimTypeValue)

		// Claims of specific claim type only should be returned
		claims := suite.keeper.GetClaims(suite.ctx, claimType)
		suite.Require().Len(claims, 2)
		suite.Require().Equalf(
			claims, types.Claims{
				types.NewClaim(claimType, suite.addrs[0], arbitraryCoins(), nonEmptyMultiRewardIndexes),
				types.NewClaim(claimType, suite.addrs[1], nil, nil),
			},
			"GetClaims(%s) should only return claims of those type", claimType,
		)
	}

	allClaims := suite.keeper.GetAllClaims(suite.ctx)
	suite.Require().Len(allClaims, len(claims))
	suite.Require().ElementsMatch(allClaims, claims, "GetAllClaims() should return claims of all types")
}
