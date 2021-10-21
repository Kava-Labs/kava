package types

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	upgrade "github.com/cosmos/cosmos-sdk/x/upgrade"
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
		{
			Subspace: "bep3",
			Key:      "MinAmount",
		},
	}
}

func (suite *PermissionsTestSuite) TestSimpleParamChangePermission_Allows() {
	testcases := []struct {
		name          string
		allowedParams AllowedParams
		pubProposal   PubProposal
		expectAllowed bool
	}{
		{
			name:          "normal (single param)",
			allowedParams: suite.exampleAllowedParams,
			pubProposal: paramstypes.NewParameterChangeProposal(
				"A Title",
				"A description for this proposal.",
				[]paramstypes.ParamChange{
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
			pubProposal: paramstypes.NewParameterChangeProposal(
				"A Title",
				"A description for this proposal.",
				[]paramstypes.ParamChange{
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
					{
						Subspace: "bep3",
						Key:      "MinAmount",

						Value: `2000`,
					},
				},
			),
			expectAllowed: true,
		},
		{
			name:          "not allowed (not in list)",
			allowedParams: suite.exampleAllowedParams,
			pubProposal: paramstypes.NewParameterChangeProposal(
				"A Title",
				"A description for this proposal.",
				[]paramstypes.ParamChange{
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
			pubProposal: paramstypes.NewParameterChangeProposal(
				"A Title",
				"A description for this proposal.",
				[]paramstypes.ParamChange{
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
			pubProposal:   govtypes.NewTextProposal("A Title", "A description of this proposal."),
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
			permission := SimpleParamChangePermission{
				AllowedParams: tc.allowedParams,
			}
			suite.Equal(
				tc.expectAllowed,
				permission.Allows(sdk.Context{}, nil, nil, tc.pubProposal),
			)
		})
	}
}

func (suite *PermissionsTestSuite) TestAllowedParams_Contains() {
	testcases := []struct {
		name            string
		allowedParams   AllowedParams
		testParam       paramstypes.ParamChange
		expectContained bool
	}{
		{
			name:          "normal",
			allowedParams: suite.exampleAllowedParams,
			testParam: paramstypes.ParamChange{
				Subspace: "cdp",
				Key:      "DebtThreshold",

				Value: `{"denom": "usdx", "amount": "1000000"}`,
			},
			expectContained: true,
		},
		{
			name:          "missing subspace",
			allowedParams: suite.exampleAllowedParams,
			testParam: paramstypes.ParamChange{
				Subspace: "",
				Key:      "DebtThreshold",

				Value: `{"denom": "usdx", "amount": "1000000"}`,
			},
			expectContained: false,
		},
		{
			name:          "missing key",
			allowedParams: suite.exampleAllowedParams,
			testParam: paramstypes.ParamChange{
				Subspace: "cdp",
				Key:      "",

				Value: `{"denom": "usdx", "amount": "1000000"}`,
			},
			expectContained: false,
		},
		{
			name:          "empty list",
			allowedParams: AllowedParams{},
			testParam: paramstypes.ParamChange{
				Subspace: "cdp",
				Key:      "DebtThreshold",

				Value: `{"denom": "usdx", "amount": "1000000"}`,
			},
			expectContained: false,
		},
		{
			name:          "nil list",
			allowedParams: nil,
			testParam: paramstypes.ParamChange{
				Subspace: "cdp",
				Key:      "DebtThreshold",

				Value: `{"denom": "usdx", "amount": "1000000"}`,
			},
			expectContained: false,
		},
		{
			name:            "no param change",
			allowedParams:   suite.exampleAllowedParams,
			testParam:       paramstypes.ParamChange{},
			expectContained: false,
		},
		{
			name:            "empty list and no param change",
			allowedParams:   AllowedParams{},
			testParam:       paramstypes.ParamChange{},
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
			pubProposal: govtypes.NewTextProposal(
				"A Title",
				"A description for this proposal.",
			),
			expectAllowed: true,
		},
		{
			name: "not allowed (wrong pubproposal type)",
			pubProposal: paramstypes.NewParameterChangeProposal(
				"A Title",
				"A description for this proposal.",
				[]paramstypes.ParamChange{
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
				permission.Allows(sdk.Context{}, nil, nil, tc.pubProposal),
			)
		})
	}
}

func (suite *PermissionsTestSuite) TestSoftwareUpgradePermission_Allows() {
	testcases := []struct {
		name          string
		pubProposal   PubProposal
		expectAllowed bool
	}{
		{
			name: "normal",
			pubProposal: upgrade.NewSoftwareUpgradeProposal(
				"A Title",
				"A description for this proposal.",
				upgrade.Plan{
					Name: "upgrade v0.12.1",
					Time: time.Date(1998, 1, 1, 0, 0, 0, 0, time.UTC),
					Info: "some information",
				},
			),
			expectAllowed: true,
		},
		{
			name: "not allowed (wrong pubproposal type)",
			pubProposal: paramstypes.NewParameterChangeProposal(
				"A Title",
				"A description for this proposal.",
				[]paramstypes.ParamChange{
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
			permission := SoftwareUpgradePermission{}
			suite.Equal(
				tc.expectAllowed,
				permission.Allows(sdk.Context{}, nil, nil, tc.pubProposal),
			)
		})
	}
}

func TestPermissionsTestSuite(t *testing.T) {
	suite.Run(t, new(PermissionsTestSuite))
}
