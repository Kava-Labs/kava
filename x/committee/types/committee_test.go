package types

import (
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/x/gov"
	"github.com/cosmos/cosmos-sdk/x/params"
	"github.com/stretchr/testify/suite"
)

var _ PubProposal = UnregisteredPubProposal{}

type UnregisteredPubProposal struct {
	gov.TextProposal
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
						Subkey:   "",
					},
				}}},
			pubProposal: params.NewParameterChangeProposal(
				"A Title",
				"A description of this proposal.",
				[]params.ParamChange{
					{
						Subspace: "cdp",
						Key:      "DebtThreshold",
						Subkey:   "",
						Value:    `{"denom": "usdx", "amount": "1000000"}`,
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
							Subkey:   "",
						},
					}},
				TextPermission{},
			},
			pubProposal:          gov.NewTextProposal("A Proposal Title", "A description of this proposal"),
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
							Subkey:   "",
						},
					}},
				GodPermission{},
			},
			pubProposal: params.NewParameterChangeProposal(
				"A Title",
				"A description of this proposal.",
				[]params.ParamChange{
					{
						Subspace: "cdp",
						Key:      "CollateralParams",
						Subkey:   "",
						Value:    `[]`,
					},
				},
			),
			expectHasPermissions: true,
		},
		{
			name:        "no permissions",
			permissions: nil,
			pubProposal: params.NewParameterChangeProposal(
				"A Title",
				"A description of this proposal.",
				[]params.ParamChange{
					{
						Subspace: "cdp",
						Key:      "CollateralParams",
						Subkey:   "",
						Value:    `[]`,
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
							Subkey:   "",
						},
					}},
				ParamChangePermission{
					AllowedParams: AllowedParams{
						{
							Subspace: "cdp",
							Key:      "DebtParams",
							Subkey:   "",
						},
					}},
			},
			pubProposal: params.NewParameterChangeProposal(
				"A Title",
				"A description of this proposal.",
				[]params.ParamChange{
					{
						Subspace: "cdp",
						Key:      "DebtThreshold",
						Subkey:   "",
						Value:    `{"denom": "usdx", "amount": "1000000"}`,
					},
					{
						Subspace: "cdp",
						Key:      "DebtParams",
						Subkey:   "",
						Value:    `[]`,
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
							Subkey:   "",
						},
					}},
			},
			pubProposal:          UnregisteredPubProposal{gov.TextProposal{"A Title", "A description."}},
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
