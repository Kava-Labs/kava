package keeper_test

import (
	"time"

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

func (suite *KeeperTestSuite) TestGetSetRewardAccrualTimes() {
	testCases := []struct {
		name        string
		subKey      string
		accrualTime time.Time
		panics      bool
	}{
		{
			name:        "normal time can be written and read",
			subKey:      "btc/usdx",
			accrualTime: time.Date(1998, 1, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name:        "zero time can be written and read",
			subKey:      "btc/usdx",
			accrualTime: time.Time{},
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.SetupApp()

			_, found := suite.keeper.GetRewardAccrualTime(suite.ctx, types.CLAIM_TYPE_USDX_MINTING, tc.subKey)
			suite.False(found)

			setFunc := func() {
				suite.keeper.SetRewardAccrualTime(suite.ctx, types.CLAIM_TYPE_USDX_MINTING, tc.subKey, tc.accrualTime)
			}
			if tc.panics {
				suite.Panics(setFunc)
				return
			} else {
				suite.NotPanics(setFunc)
			}

			for _, claimTypeValue := range types.ClaimType_value {
				claimType := types.ClaimType(claimTypeValue)

				if claimType == types.CLAIM_TYPE_USDX_MINTING {
					continue
				}

				_, found := suite.keeper.GetRewardAccrualTime(suite.ctx, claimType, tc.subKey)
				suite.False(found, "reward accrual time for claim type %s should not exist", claimType)
			}

			storedTime, found := suite.keeper.GetRewardAccrualTime(suite.ctx, types.CLAIM_TYPE_USDX_MINTING, tc.subKey)
			suite.True(found)
			suite.Equal(tc.accrualTime, storedTime)
		})
	}
}

func (suite *KeeperTestSuite) TestIterateRewardAccrualTimes() {
	suite.SetupApp()

	expectedAccrualTimes := nonEmptyAccrualTimes

	for _, at := range expectedAccrualTimes {
		suite.keeper.SetRewardAccrualTime(suite.ctx, types.CLAIM_TYPE_USDX_MINTING, at.denom, at.time)
	}

	var actualAccrualTimes []accrualtime
	suite.keeper.IterateRewardAccrualTimesByClaimType(suite.ctx, types.CLAIM_TYPE_USDX_MINTING, func(denom string, accrualTime time.Time) bool {
		actualAccrualTimes = append(actualAccrualTimes, accrualtime{denom: denom, time: accrualTime})
		return false
	})

	suite.ElementsMatch(expectedAccrualTimes, actualAccrualTimes)
}

func (suite *KeeperTestSuite) TestIterateAllRewardAccrualTimes() {
	suite.SetupApp()

	var expectedAccrualTimes types.AccrualTimes

	for _, claimTypeValue := range types.ClaimType_value {
		claimType := types.ClaimType(claimTypeValue)

		// Skip invalid claim type
		if claimType.Validate() != nil {
			continue
		}

		for _, at := range nonEmptyAccrualTimes {
			suite.keeper.SetRewardAccrualTime(suite.ctx, claimType, at.denom, at.time)

			expectedAccrualTimes = append(expectedAccrualTimes, types.NewAccrualTime(
				claimType,

				at.denom,
				at.time,
			))
		}
	}

	var actualAccrualTimes types.AccrualTimes
	suite.keeper.IterateRewardAccrualTimes(
		suite.ctx,
		func(accrualTime types.AccrualTime) bool {
			actualAccrualTimes = append(actualAccrualTimes, accrualTime)
			return false
		},
	)

	suite.ElementsMatch(expectedAccrualTimes, actualAccrualTimes)
}
