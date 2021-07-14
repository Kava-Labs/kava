package keeper_test

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/kava-labs/kava/x/incentive/types"
)

// ClaimTests runs unit tests for the keeper Claim methods
type ClaimTests struct {
	unitTester
}

func TestClaim(t *testing.T) {
	suite.Run(t, new(ClaimTests))
}

func (suite *ClaimTests) ErrorIs(err, target error) bool {
	return suite.Truef(errors.Is(err, target), "err didn't match: %s, it was: %s", target, err)
}

func (suite *ClaimTests) TestCannotClaimWhenMultiplierNotRecognised() {
	subspace := &fakeParamSubspace{
		params: types.Params{
			ClaimMultipliers: types.Multipliers{
				types.NewMultiplier(types.Small, 1, d("0.2")),
			},
		},
	}
	suite.keeper = suite.NewKeeper(subspace, nil, nil, nil, nil, nil, nil)

	claim := types.DelegatorClaim{
		BaseMultiClaim: types.BaseMultiClaim{
			Owner: arbitraryAddress(),
		},
	}
	suite.storeDelegatorClaim(claim)

	// multiplier not in params
	err := suite.keeper.ClaimDelegatorReward(suite.ctx, claim.Owner, claim.Owner, types.Large)
	suite.ErrorIs(err, types.ErrInvalidMultiplier)

	// invalid multiplier name
	err = suite.keeper.ClaimDelegatorReward(suite.ctx, claim.Owner, claim.Owner, "")
	suite.ErrorIs(err, types.ErrInvalidMultiplier)

	// invalid multiplier name
	const zeroWidthSpace = "​"
	err = suite.keeper.ClaimDelegatorReward(suite.ctx, claim.Owner, claim.Owner, types.Small+zeroWidthSpace)
	suite.ErrorIs(err, types.ErrInvalidMultiplier)
}

func (suite *ClaimTests) TestCannotClaimAfterEndTime() {
	endTime := time.Date(1998, 1, 1, 0, 0, 0, 0, time.UTC)

	subspace := &fakeParamSubspace{
		params: types.Params{
			ClaimMultipliers: types.Multipliers{
				types.NewMultiplier(types.Small, 1, d("0.2")),
			},
			ClaimEnd: endTime,
		},
	}
	suite.keeper = suite.NewKeeper(subspace, nil, nil, nil, nil, nil, nil)

	suite.ctx = suite.ctx.WithBlockTime(endTime.Add(time.Nanosecond))

	claim := types.DelegatorClaim{
		BaseMultiClaim: types.BaseMultiClaim{
			Owner: arbitraryAddress(),
		},
	}
	suite.storeDelegatorClaim(claim)

	err := suite.keeper.ClaimDelegatorReward(suite.ctx, claim.Owner, claim.Owner, types.Small)
	suite.ErrorIs(err, types.ErrClaimExpired)
}
