package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	paramsproposal "github.com/cosmos/cosmos-sdk/x/params/types/proposal"

	"github.com/kava-labs/kava/x/committee/types"
	communitytypes "github.com/kava-labs/kava/x/community/types"
)

func TestPackPermissions_Success(t *testing.T) {
	_, err := types.PackPermissions([]types.Permission{&types.GodPermission{}})
	require.NoError(t, err)
}

func TestPackPermissions_Failure(t *testing.T) {
	_, err := types.PackPermissions([]types.Permission{nil})
	require.Error(t, err)
}

func TestUnpackPermissions_Success(t *testing.T) {
	packedPermissions, err := types.PackPermissions([]types.Permission{&types.GodPermission{}})
	require.NoError(t, err)
	unpackedPermissions, err := types.UnpackPermissions(packedPermissions)
	require.NoError(t, err)
	require.Len(t, unpackedPermissions, 1)
	_, ok := unpackedPermissions[0].(*types.GodPermission)
	require.True(t, ok)
}

func TestUnpackPermissions_Failure(t *testing.T) {
	vote, err := codectypes.NewAnyWithValue(&types.Vote{ProposalID: 1})
	require.NoError(t, err)
	_, err = types.UnpackPermissions([]*codectypes.Any{vote})
	require.Error(t, err)
}

func TestCommunityCDPRepayDebtPermission_Allows(t *testing.T) {
	permission := types.CommunityCDPRepayDebtPermission{}
	testcases := []struct {
		name     string
		proposal types.PubProposal
		allowed  bool
	}{
		{
			name: "allowed for correct proposal",
			proposal: communitytypes.NewCommunityCDPRepayDebtProposal(
				"repay x/community cdp debt",
				"repays debt on a cdp position",
				"collateral-type",
				sdk.NewInt64Coin("ukava", 1e10),
			),
			allowed: true,
		},
		{
			name:     "fails for nil proposal",
			proposal: nil,
			allowed:  false,
		},
		{
			name: "fails for wrong proposal",
			proposal: newTestParamsChangeProposalWithChanges([]paramsproposal.ParamChange{
				{Subspace: "cdp", Key: "DebtThreshold", Value: `test`},
			}),
			allowed: false,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.allowed, permission.Allows(sdk.Context{}, nil, tc.proposal))
		})
	}
}

func TestCommunityPoolLendWithdrawPermission_Allows(t *testing.T) {
	permission := types.CommunityPoolLendWithdrawPermission{}
	testcases := []struct {
		name     string
		proposal types.PubProposal
		allowed  bool
	}{
		{
			name: "allowed for correct proposal",
			proposal: communitytypes.NewCommunityPoolLendWithdrawProposal(
				"withdraw lend position",
				"this fake proposal withdraws a lend position for the community pool",
				sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(1e10))),
			),
			allowed: true,
		},
		{
			name:     "fails for nil proposal",
			proposal: nil,
			allowed:  false,
		},
		{
			name: "fails for wrong proposal",
			proposal: newTestParamsChangeProposalWithChanges([]paramsproposal.ParamChange{
				{Subspace: "cdp", Key: "DebtThreshold", Value: `test`},
			}),
			allowed: false,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.allowed, permission.Allows(sdk.Context{}, nil, tc.proposal))
		})
	}
}

func TestCommunityCDPWithdrawCollateralPermission_Allows(t *testing.T) {
	permission := types.CommunityCDPWithdrawCollateralPermission{}
	testcases := []struct {
		name     string
		proposal types.PubProposal
		allowed  bool
	}{
		{
			name: "allowed for correct proposal",
			proposal: communitytypes.NewCommunityCDPWithdrawCollateralProposal(
				"withdraw x/community cdp collateral",
				"yes",
				"collateral-type",
				sdk.NewInt64Coin("ukava", 1e10),
			),
			allowed: true,
		},
		{
			name:     "fails for nil proposal",
			proposal: nil,
			allowed:  false,
		},
		{
			name: "fails for wrong proposal",
			proposal: newTestParamsChangeProposalWithChanges([]paramsproposal.ParamChange{
				{Subspace: "cdp", Key: "DebtThreshold", Value: `test`},
			}),
			allowed: false,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.allowed, permission.Allows(sdk.Context{}, nil, tc.proposal))
		})
	}
}

func TestParamsChangePermission_SimpleParamsChange_Allows(t *testing.T) {
	testPermission := types.ParamsChangePermission{
		AllowedParamsChanges: types.AllowedParamsChanges{
			types.AllowedParamsChange{
				Subspace: "cdp",
				Key:      "DebtThreshold",
			},
			types.AllowedParamsChange{
				Subspace: "cdp",
				Key:      "SurplusThreshold",
			},
			types.AllowedParamsChange{
				Subspace: "auction",
				Key:      "BidDuration",
			},
			types.AllowedParamsChange{
				Subspace: "bep3",
				Key:      "MinAmount",
			},
		},
	}

	testcases := []struct {
		name          string
		permission    types.ParamsChangePermission
		pubProposal   types.PubProposal
		expectAllowed bool
	}{
		{
			name:       "normal (single param)",
			permission: testPermission,
			pubProposal: newTestParamsChangeProposalWithChanges(
				[]paramsproposal.ParamChange{
					{
						Subspace: "cdp",
						Key:      "DebtThreshold",
						Value:    `test`,
					},
				},
			),
			expectAllowed: true,
		},
		{
			name:       "not allowed (no allowed params change)",
			permission: testPermission,
			pubProposal: newTestParamsChangeProposalWithChanges(
				[]paramsproposal.ParamChange{
					{
						Subspace: "kavadist",
						Key:      "TestKey",
						Value:    `100`,
					},
				},
			),
			expectAllowed: false,
		},
		{
			name:       "allowed (multiple params)",
			permission: testPermission,
			pubProposal: newTestParamsChangeProposalWithChanges(
				[]paramsproposal.ParamChange{
					{
						Subspace: "cdp",
						Key:      "DebtThreshold",
						Value:    `test`,
					},
					{
						Subspace: "cdp",
						Key:      "SurplusThreshold",
						Value:    `100`,
					},
					{
						Subspace: "bep3",
						Key:      "MinAmount",
						Value:    `test`,
					},
				},
			),
			expectAllowed: true,
		},
		{
			name:       "not allowed (multiple params)",
			permission: testPermission,
			pubProposal: newTestParamsChangeProposalWithChanges(
				[]paramsproposal.ParamChange{
					{
						Subspace: "cdp",
						Key:      "DebtThreshold",
						Value:    `test`,
					},
					{
						Subspace: "cdp",
						Key:      "SurplusThreshold",
						Value:    `100`,
					},
					{
						Subspace: "bep3",
						Key:      "Duration",
						Value:    `test`,
					},
				},
			),
			expectAllowed: false,
		},
		{
			name:       "not allowed (empty allowed params)",
			permission: types.ParamsChangePermission{AllowedParamsChanges: types.AllowedParamsChanges{}},
			pubProposal: newTestParamsChangeProposalWithChanges(
				[]paramsproposal.ParamChange{
					{
						Subspace: "cdp",
						Key:      "DebtThreshold",
						Value:    `test`,
					},
				},
			),
			expectAllowed: false,
		},
		{
			name:          "not allowed (mismatched pubproposal type)",
			permission:    testPermission,
			pubProposal:   govv1beta1.NewTextProposal("A Title", "A description of this proposal."),
			expectAllowed: false,
		},
		{
			name:          "not allowed (nil pubproposal)",
			permission:    testPermission,
			pubProposal:   nil,
			expectAllowed: false,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.expectAllowed, tc.permission.Allows(sdk.Context{}, nil, tc.pubProposal))
		})
	}
}

func newTestParamsChangeProposalWithChanges(changes []paramsproposal.ParamChange) types.PubProposal {
	return paramsproposal.NewParameterChangeProposal(
		"A Title",
		"A description for this proposal.",
		changes,
	)
}
