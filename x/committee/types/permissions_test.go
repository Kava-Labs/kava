package types

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/x/gov"
	"github.com/cosmos/cosmos-sdk/x/params"
	"github.com/stretchr/testify/suite"
)

type PermissionsTestSuite struct {
	suite.Suite

	exampleAllowedParams AllowedParams
}

func (suite *PermissionsTestSuite) SetupTest() {
	suite.exampleAllowedParams = AllowedParams{
		{
			Subspace: "cdp",
			Key:      "DebtThreshold",
		},
		{
			Subspace: "cdp",
			Key:      "SurplusThreshold",
		},
		{
			Subspace: "cdp",
			Key:      "CollateralParams",
		},
		{
			Subspace: "auction",
			Key:      "BidDuration",
		},
	}
}

func (suite *PermissionsTestSuite) TestParamChangePermission_Allows() {
	testcases := []struct {
		name          string
		allowedParams AllowedParams
		pubProposal   PubProposal
		expectAllowed bool
	}{
		{
			name:          "normal (single param)",
			allowedParams: suite.exampleAllowedParams,
			pubProposal: params.NewParameterChangeProposal(
				"A Title",
				"A description for this proposal.",
				[]params.ParamChange{
					{
						Subspace: "cdp",
						Key:      "DebtThreshold",

						Value: `{"denom": "usdx", "amount": "1000000"}`,
					},
				},
			),
			expectAllowed: true,
		},
		{
			name:          "normal (multiple params)",
			allowedParams: suite.exampleAllowedParams,
			pubProposal: params.NewParameterChangeProposal(
				"A Title",
				"A description for this proposal.",
				[]params.ParamChange{
					{
						Subspace: "cdp",
						Key:      "DebtThreshold",

						Value: `{"denom": "usdx", "amount": "1000000"}`,
					},
					{
						Subspace: "cdp",
						Key:      "CollateralParams",

						Value: `[]`,
					},
				},
			),
			expectAllowed: true,
		},
		{
			name:          "not allowed (not in list)",
			allowedParams: suite.exampleAllowedParams,
			pubProposal: params.NewParameterChangeProposal(
				"A Title",
				"A description for this proposal.",
				[]params.ParamChange{
					{
						Subspace: "cdp",
						Key:      "GlobalDebtLimit",

						Value: `{"denom": "usdx", "amount": "1000000000"}`,
					},
				},
			),
			expectAllowed: false,
		},
		{
			name:          "not allowed (nil allowed params)",
			allowedParams: nil,
			pubProposal: params.NewParameterChangeProposal(
				"A Title",
				"A description for this proposal.",
				[]params.ParamChange{
					{
						Subspace: "cdp",
						Key:      "DebtThreshold",

						Value: `[{"denom": "usdx", "amount": "1000000"}]`,
					},
				},
			),
			expectAllowed: false,
		},
		{
			name:          "not allowed (mismatched pubproposal type)",
			allowedParams: suite.exampleAllowedParams,
			pubProposal:   gov.NewTextProposal("A Title", "A description of this proposal."),
			expectAllowed: false,
		},
		{
			name:          "not allowed (nil pubproposal)",
			allowedParams: suite.exampleAllowedParams,
			pubProposal:   nil,
			expectAllowed: false,
		},
	}

	for _, tc := range testcases {
		suite.Run(tc.name, func() {
			permission := ParamChangePermission{
				AllowedParams: tc.allowedParams,
			}
			suite.Equal(
				tc.expectAllowed,
				permission.Allows(tc.pubProposal),
			)
		})
	}
}

func (suite *PermissionsTestSuite) TestAllowedParams_Contains() {
	testcases := []struct {
		name            string
		allowedParams   AllowedParams
		testParam       params.ParamChange
		expectContained bool
	}{
		{
			name:          "normal",
			allowedParams: suite.exampleAllowedParams,
			testParam: params.ParamChange{
				Subspace: "cdp",
				Key:      "DebtThreshold",

				Value: `{"denom": "usdx", "amount": "1000000"}`,
			},
			expectContained: true,
		},
		{
			name:          "missing subspace",
			allowedParams: suite.exampleAllowedParams,
			testParam: params.ParamChange{
				Subspace: "",
				Key:      "DebtThreshold",

				Value: `{"denom": "usdx", "amount": "1000000"}`,
			},
			expectContained: false,
		},
		{
			name:          "missing key",
			allowedParams: suite.exampleAllowedParams,
			testParam: params.ParamChange{
				Subspace: "cdp",
				Key:      "",

				Value: `{"denom": "usdx", "amount": "1000000"}`,
			},
			expectContained: false,
		},
		{
			name:          "empty list",
			allowedParams: AllowedParams{},
			testParam: params.ParamChange{
				Subspace: "cdp",
				Key:      "DebtThreshold",

				Value: `{"denom": "usdx", "amount": "1000000"}`,
			},
			expectContained: false,
		},
		{
			name:          "nil list",
			allowedParams: nil,
			testParam: params.ParamChange{
				Subspace: "cdp",
				Key:      "DebtThreshold",

				Value: `{"denom": "usdx", "amount": "1000000"}`,
			},
			expectContained: false,
		},
		{
			name:            "no param change",
			allowedParams:   suite.exampleAllowedParams,
			testParam:       params.ParamChange{},
			expectContained: false,
		},
		{
			name:            "empty list and no param change",
			allowedParams:   AllowedParams{},
			testParam:       params.ParamChange{},
			expectContained: false,
		},
	}

	for _, tc := range testcases {
		suite.Run(tc.name, func() {
			suite.Require().Equal(
				tc.expectContained,
				tc.allowedParams.Contains(tc.testParam),
			)
		})
	}
}

func (suite *PermissionsTestSuite) TestTextPermission_Allows() {
	testcases := []struct {
		name          string
		pubProposal   PubProposal
		expectAllowed bool
	}{
		{
			name: "normal",
			pubProposal: gov.NewTextProposal(
				"A Title",
				"A description for this proposal.",
			),
			expectAllowed: true,
		},
		{
			name: "not allowed (wrong pubproposal type)",
			pubProposal: params.NewParameterChangeProposal(
				"A Title",
				"A description for this proposal.",
				[]params.ParamChange{
					{
						Subspace: "cdp",
						Key:      "DebtThreshold",
						Value:    `{"denom": "usdx", "amount": "1000000"}`,
					},
					{
						Subspace: "cdp",
						Key:      "CollateralParams",
						Value:    `[]`,
					},
				},
			),
			expectAllowed: false,
		},
		{
			name:          "not allowed (nil pubproposal)",
			pubProposal:   nil,
			expectAllowed: false,
		},
	}

	for _, tc := range testcases {
		suite.Run(tc.name, func() {
			permission := TextPermission{}
			suite.Equal(
				tc.expectAllowed,
				permission.Allows(tc.pubProposal),
			)
		})
	}
}
func TestPermissionsTestSuite(t *testing.T) {
	suite.Run(t, new(PermissionsTestSuite))
}
