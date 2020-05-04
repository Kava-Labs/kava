package types

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

var _ PubProposal = UnregisteredPubProposal{}

type UnregisteredPubProposal struct {
	govtypes.TextProposal
}

func (UnregisteredPubProposal) ProposalRoute() string { return "unregistered" }
func (UnregisteredPubProposal) ProposalType() string  { return "unregistered" }

type TypesTestSuite struct {
	suite.Suite
}

func (suite *TypesTestSuite) TestCommittee_HasPermissionsFor() {

	testcases := []struct {
		name                 string
		permissions          []Permission
		pubProposal          PubProposal
		expectHasPermissions bool
	}{
		{
			name: "normal (single permission)",
			permissions: []Permission{ParamChangePermission{
				AllowedParams: AllowedParams{
					{
						Subspace: "cdp",
						Key:      "DebtThreshold",
					},
				}}},
			pubProposal: paramstypes.NewParameterChangeProposal(
				"A Title",
				"A description of this proposal.",
				[]paramstypes.ParamChange{
					{
						Subspace: "cdp",
						Key:      "DebtThreshold",

						Value: `{"denom": "usdx", "amount": "1000000"}`,
					},
				},
			),
			expectHasPermissions: true,
		},
		{
			name: "normal (multiple permissions)",
			permissions: []Permission{
				ParamChangePermission{
					AllowedParams: AllowedParams{
						{
							Subspace: "cdp",
							Key:      "DebtThreshold",
						},
					}},
				TextPermission{},
			},
			pubProposal:          govtypes.NewTextProposal("A Proposal Title", "A description of this proposal"),
			expectHasPermissions: true,
		},
		{
			name: "overruling permission",
			permissions: []Permission{
				ParamChangePermission{
					AllowedParams: AllowedParams{
						{
							Subspace: "cdp",
							Key:      "DebtThreshold",
						},
					}},
				GodPermission{},
			},
			pubProposal: paramstypes.NewParameterChangeProposal(
				"A Title",
				"A description of this proposal.",
				[]paramstypes.ParamChange{
					{
						Subspace: "cdp",
						Key:      "CollateralParams",

						Value: `[]`,
					},
				},
			),
			expectHasPermissions: true,
		},
		{
			name:        "no permissions",
			permissions: nil,
			pubProposal: paramstypes.NewParameterChangeProposal(
				"A Title",
				"A description of this proposal.",
				[]paramstypes.ParamChange{
					{
						Subspace: "cdp",
						Key:      "CollateralParams",

						Value: `[]`,
					},
				},
			),
			expectHasPermissions: false,
		},
		{
			name: "split permissions",
			// These permissions looks like they allow the param change proposal, however a proposal must pass a single permission independently of others.
			permissions: []Permission{
				ParamChangePermission{
					AllowedParams: AllowedParams{
						{
							Subspace: "cdp",
							Key:      "DebtThreshold",
						},
					}},
				ParamChangePermission{
					AllowedParams: AllowedParams{
						{
							Subspace: "cdp",
							Key:      "DebtParams",
						},
					}},
			},
			pubProposal: paramstypes.NewParameterChangeProposal(
				"A Title",
				"A description of this proposal.",
				[]paramstypes.ParamChange{
					{
						Subspace: "cdp",
						Key:      "DebtThreshold",

						Value: `{"denom": "usdx", "amount": "1000000"}`,
					},
					{
						Subspace: "cdp",
						Key:      "DebtParams",

						Value: `[]`,
					},
				},
			),
			expectHasPermissions: false,
		},
		{
			name: "unregistered proposal",
			permissions: []Permission{
				ParamChangePermission{
					AllowedParams: AllowedParams{
						{
							Subspace: "cdp",
							Key:      "DebtThreshold",
						},
					}},
			},
			pubProposal:          UnregisteredPubProposal{govtypes.TextProposal{Title: "A Title", Description: "A description."}},
			expectHasPermissions: false,
		},
	}

	for _, tc := range testcases {
		suite.Run(tc.name, func() {
			com := NewCommittee(
				12,
				"a description of this committee",
				nil,
				tc.permissions,
				d("0.5"),
				24*time.Hour,
			)
			suite.Equal(
				tc.expectHasPermissions,
				com.HasPermissionsFor(tc.pubProposal),
			)
		})
	}
}

func TestTypesTestSuite(t *testing.T) {
	suite.Run(t, new(TypesTestSuite))
}
