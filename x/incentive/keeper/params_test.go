package keeper_test

import (
	"testing"

	earntypes "github.com/kava-labs/kava/x/earn/types"
	"github.com/kava-labs/kava/x/incentive/types"
	"github.com/stretchr/testify/suite"
)

type ParamsTests struct {
	unitTester
}

func TestParamsTests(t *testing.T) {
	suite.Run(t, new(ParamsTests))
}

func (suite *ParamsTests) TestGetMultiplierByDenom() {
	subspace := &fakeParamSubspace{
		params: types.Params{
			ClaimMultipliers: types.MultipliersPerDenoms{
				{
					Denom: "hard",
					Multipliers: types.Multipliers{
						types.NewMultiplier("small", 1, d("0.2")),
					},
				},
				{
					Denom: "ukava",
					Multipliers: types.Multipliers{
						types.NewMultiplier("large", 0, d("1")),
					},
					ModuleName: earntypes.ModuleName,
				},
				{
					Denom: "ukava",
					Multipliers: types.Multipliers{
						types.NewMultiplier("large", 1, d("0.2")),
					},
					ModuleName: "",
				},
			},
		},
	}
	suite.keeper = suite.NewKeeper(subspace, nil, nil, nil, nil, nil, nil, nil, nil, nil)

	suite.T().Logf("params :%v", suite.keeper.GetParams(suite.ctx).ClaimMultipliers)

	tests := []struct {
		name       string
		denom      string
		moduleName string
		multiplier string
		expected   types.Multiplier
	}{
		{
			name:       "hard claim",
			denom:      "hard",
			moduleName: "",
			multiplier: "small",
			expected:   types.NewMultiplier("small", 1, d("0.2")),
		},
		{
			name:       "ukava earn",
			denom:      "ukava",
			moduleName: earntypes.ModuleName,
			multiplier: "large",
			expected:   types.NewMultiplier("large", 0, d("1.0")),
		},
		{
			name:       "ukava non-earn",
			denom:      "ukava",
			moduleName: "",
			multiplier: "large",
			expected:   types.NewMultiplier("large", 1, d("0.2")),
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			multiplier, found := suite.keeper.GetMultiplierByDenom(
				suite.ctx,
				tc.denom,
				tc.multiplier,
				tc.moduleName,
			)
			suite.Require().True(found)
			suite.Require().Equal(tc.expected, multiplier)
		})
	}
}
