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

			_, found := suite.keeper.Store.GetClaim(suite.ctx, claimType, suite.addrs[0])
			suite.Require().False(found)

			suite.Require().NotPanics(func() {
				suite.keeper.Store.SetClaim(suite.ctx, c)
			})
			testC, found := suite.keeper.Store.GetClaim(suite.ctx, claimType, suite.addrs[0])
			suite.Require().True(found)
			suite.Require().Equal(c, testC)

			// Check that other claim types do not exist for the same address
			for otherClaimTypeName, otherClaimTypeValue := range types.ClaimType_value {
				// Skip the current claim type
				if otherClaimTypeValue == claimTypeValue {
					continue
				}

				otherClaimType := types.ClaimType(otherClaimTypeValue)
				_, found := suite.keeper.Store.GetClaim(suite.ctx, otherClaimType, suite.addrs[0])
				suite.Require().False(found, "claim type %s should not exist", otherClaimTypeName)
			}

			suite.Require().NotPanics(func() {
				suite.keeper.Store.DeleteClaim(suite.ctx, claimType, suite.addrs[0])
			})
			_, found = suite.keeper.Store.GetClaim(suite.ctx, claimType, suite.addrs[0])
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
		suite.keeper.Store.SetClaim(suite.ctx, claim)
	}

	for _, claimTypeValue := range types.ClaimType_value {
		claimType := types.ClaimType(claimTypeValue)

		// Claims of specific claim type only should be returned
		claims := suite.keeper.Store.GetClaims(suite.ctx, claimType)
		suite.Require().Len(claims, 2)
		suite.Require().Equalf(
			claims, types.Claims{
				types.NewClaim(claimType, suite.addrs[0], arbitraryCoins(), nonEmptyMultiRewardIndexes),
				types.NewClaim(claimType, suite.addrs[1], nil, nil),
			},
			"GetClaims(%s) should only return claims of those type", claimType,
		)
	}

	allClaims := suite.keeper.Store.GetAllClaims(suite.ctx)
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

			_, found := suite.keeper.Store.GetRewardAccrualTime(suite.ctx, types.CLAIM_TYPE_USDX_MINTING, tc.subKey)
			suite.False(found)

			setFunc := func() {
				suite.keeper.Store.SetRewardAccrualTime(suite.ctx, types.CLAIM_TYPE_USDX_MINTING, tc.subKey, tc.accrualTime)
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

				_, found := suite.keeper.Store.GetRewardAccrualTime(suite.ctx, claimType, tc.subKey)
				suite.False(found, "reward accrual time for claim type %s should not exist", claimType)
			}

			storedTime, found := suite.keeper.Store.GetRewardAccrualTime(suite.ctx, types.CLAIM_TYPE_USDX_MINTING, tc.subKey)
			suite.True(found)
			suite.Equal(tc.accrualTime, storedTime)
		})
	}
}

func (suite *KeeperTestSuite) TestGetSetRewardIndexes() {
	testCases := []struct {
		name           string
		collateralType string
		indexes        types.RewardIndexes
		wantIndex      types.RewardIndexes
		panics         bool
	}{
		{
			name:           "two factors can be written and read",
			collateralType: "btc/usdx",
			indexes: types.RewardIndexes{
				{
					CollateralType: "hard",
					RewardFactor:   d("0.02"),
				},
				{
					CollateralType: "ukava",
					RewardFactor:   d("0.04"),
				},
			},
			wantIndex: types.RewardIndexes{
				{
					CollateralType: "hard",
					RewardFactor:   d("0.02"),
				},
				{
					CollateralType: "ukava",
					RewardFactor:   d("0.04"),
				},
			},
		},
		{
			name:           "indexes with empty pool name panics",
			collateralType: "",
			indexes: types.RewardIndexes{
				{
					CollateralType: "hard",
					RewardFactor:   d("0.02"),
				},
				{
					CollateralType: "ukava",
					RewardFactor:   d("0.04"),
				},
			},
			panics: true,
		},
		{
			// this test is to detect any changes in behavior
			name:           "setting empty indexes does not panic",
			collateralType: "btc/usdx",
			// Marshalling empty slice results in [] bytes, unmarshalling the []
			// empty bytes results in a nil slice instead of an empty slice
			indexes:   types.RewardIndexes{},
			wantIndex: nil,
			panics:    false,
		},
		{
			// this test is to detect any changes in behavior
			name:           "setting nil indexes does not panic",
			collateralType: "btc/usdx",
			indexes:        nil,
			wantIndex:      nil,
			panics:         false,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.SetupApp()

			_, found := suite.keeper.Store.GetRewardIndexesOfClaimType(suite.ctx, types.CLAIM_TYPE_SWAP, tc.collateralType)
			suite.False(found)

			setFunc := func() {
				suite.keeper.Store.SetRewardIndexes(suite.ctx, types.CLAIM_TYPE_SWAP, tc.collateralType, tc.indexes)
			}
			if tc.panics {
				suite.Panics(setFunc)
				return
			} else {
				suite.NotPanics(setFunc)
			}

			storedIndexes, found := suite.keeper.Store.GetRewardIndexesOfClaimType(suite.ctx, types.CLAIM_TYPE_SWAP, tc.collateralType)
			suite.True(found)
			suite.Equal(tc.wantIndex, storedIndexes)

			for _, otherClaimTypeValue := range types.ClaimType_value {
				// Skip swap
				if types.ClaimType(otherClaimTypeValue) == types.CLAIM_TYPE_SWAP {
					continue
				}

				otherClaimType := types.ClaimType(otherClaimTypeValue)

				// Other claim types should not be affected
				_, found := suite.keeper.Store.GetRewardIndexesOfClaimType(suite.ctx, otherClaimType, tc.collateralType)
				suite.False(found)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestIterateRewardAccrualTimes() {
	suite.SetupApp()

	expectedAccrualTimes := nonEmptyAccrualTimes

	for _, at := range expectedAccrualTimes {
		suite.keeper.Store.SetRewardAccrualTime(suite.ctx, types.CLAIM_TYPE_USDX_MINTING, at.denom, at.time)
	}

	var actualAccrualTimes []accrualtime
	suite.keeper.Store.IterateRewardAccrualTimesByClaimType(suite.ctx, types.CLAIM_TYPE_USDX_MINTING, func(denom string, accrualTime time.Time) bool {
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
			suite.keeper.Store.SetRewardAccrualTime(suite.ctx, claimType, at.denom, at.time)

			expectedAccrualTimes = append(expectedAccrualTimes, types.NewAccrualTime(
				claimType,

				at.denom,
				at.time,
			))
		}
	}

	var actualAccrualTimes types.AccrualTimes
	suite.keeper.Store.IterateRewardAccrualTimes(
		suite.ctx,
		func(accrualTime types.AccrualTime) bool {
			actualAccrualTimes = append(actualAccrualTimes, accrualTime)
			return false
		},
	)

	suite.ElementsMatch(expectedAccrualTimes, actualAccrualTimes)
}

func (suite *KeeperTestSuite) TestIterateRewardIndexes() {
	suite.SetupApp()
	swapMultiIndexes := types.MultiRewardIndexes{
		{
			CollateralType: "bnb",
			RewardIndexes: types.RewardIndexes{
				{
					CollateralType: "swap",
					RewardFactor:   d("0.0000002"),
				},
				{
					CollateralType: "ukava",
					RewardFactor:   d("0.04"),
				},
			},
		},
		{
			CollateralType: "btcb",
			RewardIndexes: types.RewardIndexes{
				{
					CollateralType: "hard",
					RewardFactor:   d("0.02"),
				},
			},
		},
	}

	earnMultiIndexes := types.MultiRewardIndexes{
		{
			CollateralType: "usdc",
			RewardIndexes: types.RewardIndexes{
				{
					CollateralType: "usdc",
					RewardFactor:   d("0.0000002"),
				},
				{
					CollateralType: "ukava",
					RewardFactor:   d("0.04"),
				},
			},
		},
		{
			CollateralType: "ukava",
			RewardIndexes: types.RewardIndexes{
				{
					CollateralType: "ukava",
					RewardFactor:   d("0.02"),
				},
			},
		},
	}

	for _, mi := range swapMultiIndexes {
		suite.keeper.Store.SetRewardIndexes(suite.ctx, types.CLAIM_TYPE_SWAP, mi.CollateralType, mi.RewardIndexes)
	}

	for _, mi := range earnMultiIndexes {
		// These should be excluded when iterating over swap indexes
		suite.keeper.Store.SetRewardIndexes(suite.ctx, types.CLAIM_TYPE_EARN, mi.CollateralType, mi.RewardIndexes)
	}

	actualMultiIndexesMap := make(map[types.ClaimType]types.MultiRewardIndexes)
	suite.keeper.Store.IterateRewardIndexesByClaimType(suite.ctx, types.CLAIM_TYPE_SWAP, func(rewardIndex types.TypedRewardIndexes) bool {
		actualMultiIndexesMap[rewardIndex.ClaimType] = actualMultiIndexesMap[rewardIndex.ClaimType].With(rewardIndex.CollateralType, rewardIndex.RewardIndexes)
		return false
	})

	suite.Require().Len(actualMultiIndexesMap, 1, "iteration should only include 1 claim type")
	suite.Require().Equal(swapMultiIndexes, actualMultiIndexesMap[types.CLAIM_TYPE_SWAP])
}

func (suite *KeeperTestSuite) TestIterateAllRewardIndexes() {
	suite.SetupApp()
	multiIndexes := types.MultiRewardIndexes{
		{
			CollateralType: "ukava",
			RewardIndexes: types.RewardIndexes{
				{
					CollateralType: "swap",
					RewardFactor:   d("0.0000002"),
				},
				{
					CollateralType: "ukava",
					RewardFactor:   d("0.04"),
				},
			},
		},
		{
			CollateralType: "usdc",
			RewardIndexes: types.RewardIndexes{
				{
					CollateralType: "hard",
					RewardFactor:   d("0.02"),
				},
			},
		},
	}

	for _, claimTypeValue := range types.ClaimType_value {
		if types.ClaimType(claimTypeValue) == types.CLAIM_TYPE_UNSPECIFIED {
			continue
		}

		claimType := types.ClaimType(claimTypeValue)

		for _, mi := range multiIndexes {
			suite.keeper.Store.SetRewardIndexes(suite.ctx, claimType, mi.CollateralType, mi.RewardIndexes)
		}
	}

	actualMultiIndexesMap := make(map[types.ClaimType]types.MultiRewardIndexes)
	suite.keeper.Store.IterateRewardIndexes(suite.ctx, func(rewardIndex types.TypedRewardIndexes) bool {
		actualMultiIndexesMap[rewardIndex.ClaimType] = actualMultiIndexesMap[rewardIndex.ClaimType].With(rewardIndex.CollateralType, rewardIndex.RewardIndexes)
		return false
	})

	// -1 to exclude the unspecified type
	suite.Require().Len(actualMultiIndexesMap, len(types.ClaimType_value)-1)

	for claimType, actualMultiIndexes := range actualMultiIndexesMap {
		suite.Require().NoError(claimType.Validate())
		suite.Require().Equal(multiIndexes, actualMultiIndexes)
	}
}
