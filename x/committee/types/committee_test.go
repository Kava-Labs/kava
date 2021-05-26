package types

import (
	"fmt"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto"
)

func TestBaseCommittee(t *testing.T) {
	addresses := []sdk.AccAddress{
		sdk.AccAddress(crypto.AddressHash([]byte("KavaTest1"))),
		sdk.AccAddress(crypto.AddressHash([]byte("KavaTest2"))),
		sdk.AccAddress(crypto.AddressHash([]byte("KavaTest3"))),
	}

	testCases := []struct {
		name       string
		committee  BaseCommittee
		expectPass bool
	}{
		{
			name: "normal",
			committee: BaseCommittee{
				ID:               1,
				Description:      "This base committee is for testing.",
				Members:          addresses[:3],
				Permissions:      []Permission{GodPermission{}},
				VoteThreshold:    d("0.667"),
				ProposalDuration: time.Hour * 24 * 7,
				TallyOption:      FirstPastThePost,
			},
			expectPass: true,
		},
		{
			name: "description length too long",
			committee: BaseCommittee{
				ID: 1,
				Description: fmt.Sprintln("This base committee has a long description.",
					"This base committee has a long description. This base committee has a long description.",
					"This base committee has a long description. This base committee has a long description.",
					"This base committee has a long description. This base committee has a long description.",
					"This base committee has a long description. This base committee has a long description.",
					"This base committee has a long description. This base committee has a long description.",
					"This base committee has a long description. This base committee has a long description.",
					"This base committee has a long description. This base committee has a long description."),
				Members:          addresses[:3],
				Permissions:      []Permission{GodPermission{}},
				VoteThreshold:    d("0.667"),
				ProposalDuration: time.Hour * 24 * 7,
				TallyOption:      FirstPastThePost,
			},
			expectPass: false,
		},
		{
			name: "no members",
			committee: BaseCommittee{
				ID:               1,
				Description:      "This base committee is for testing.",
				Members:          []sdk.AccAddress{},
				Permissions:      []Permission{GodPermission{}},
				VoteThreshold:    d("0.667"),
				ProposalDuration: time.Hour * 24 * 7,
				TallyOption:      FirstPastThePost,
			},
			expectPass: false,
		},
		{
			name: "duplicate member",
			committee: BaseCommittee{
				ID:               1,
				Description:      "This base committee is for testing.",
				Members:          []sdk.AccAddress{addresses[2], addresses[2]},
				Permissions:      []Permission{GodPermission{}},
				VoteThreshold:    d("0.667"),
				ProposalDuration: time.Hour * 24 * 7,
				TallyOption:      FirstPastThePost,
			},
			expectPass: false,
		},
		{
			name: "nil permissions",
			committee: BaseCommittee{
				ID:               1,
				Description:      "This base committee is for testing.",
				Members:          addresses[:3],
				Permissions:      []Permission{nil},
				VoteThreshold:    d("0.667"),
				ProposalDuration: time.Hour * 24 * 7,
				TallyOption:      FirstPastThePost,
			},
			expectPass: false,
		},
		{
			name: "negative proposal duration",
			committee: BaseCommittee{
				ID:               1,
				Description:      "This base committee is for testing.",
				Members:          addresses[:3],
				Permissions:      []Permission{GodPermission{}},
				VoteThreshold:    d("0.667"),
				ProposalDuration: time.Hour * 24 * -7,
				TallyOption:      FirstPastThePost,
			},
			expectPass: false,
		},
		{
			name: "vote threshold is nil",
			committee: BaseCommittee{
				ID:               1,
				Description:      "This base committee is for testing.",
				Members:          addresses[:3],
				Permissions:      []Permission{GodPermission{}},
				VoteThreshold:    sdk.Dec{Int: nil},
				ProposalDuration: time.Hour * 24 * 7,
				TallyOption:      FirstPastThePost,
			},
			expectPass: false,
		},
		{
			name: "vote threshold is 0",
			committee: BaseCommittee{
				ID:               1,
				Description:      "This base committee is for testing.",
				Members:          addresses[:3],
				Permissions:      []Permission{GodPermission{}},
				VoteThreshold:    d("0"),
				ProposalDuration: time.Hour * 24 * 7,
				TallyOption:      FirstPastThePost,
			},
			expectPass: false,
		},
		{
			name: "vote threshold above 1",
			committee: BaseCommittee{
				ID:               1,
				Description:      "This base committee is for testing.",
				Members:          addresses[:3],
				Permissions:      []Permission{GodPermission{}},
				VoteThreshold:    d("1.001"),
				ProposalDuration: time.Hour * 24 * 7,
				TallyOption:      FirstPastThePost,
			},
			expectPass: false,
		},
		{
			name: "invalid tally option",
			committee: BaseCommittee{
				ID:               1,
				Description:      "This base committee is for testing.",
				Members:          addresses[:3],
				Permissions:      []Permission{GodPermission{}},
				VoteThreshold:    d("0.667"),
				ProposalDuration: time.Hour * 24 * 7,
				TallyOption:      NullTallyOption,
			},
			expectPass: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			err := tc.committee.Validate()

			if tc.expectPass {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

// TestMemberCommittee is an alias for BaseCommittee that has 'MemberCommittee' type
func TestMemberCommittee(t *testing.T) {
	addresses := []sdk.AccAddress{
		sdk.AccAddress(crypto.AddressHash([]byte("KavaTest1"))),
		sdk.AccAddress(crypto.AddressHash([]byte("KavaTest2"))),
		sdk.AccAddress(crypto.AddressHash([]byte("KavaTest3"))),
	}

	testCases := []struct {
		name       string
		committee  MemberCommittee
		expectPass bool
	}{
		{
			name: "normal",
			committee: MemberCommittee{
				BaseCommittee: BaseCommittee{
					ID:               1,
					Description:      "This member committee is for testing.",
					Members:          addresses[:3],
					Permissions:      []Permission{GodPermission{}},
					VoteThreshold:    d("0.667"),
					ProposalDuration: time.Hour * 24 * 7,
					TallyOption:      FirstPastThePost,
				},
			},
			expectPass: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			require.Equal(t, MemberCommitteeType, tc.committee.GetType())

			err := tc.committee.Validate()
			if tc.expectPass {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

// TestTokenCommittee tests unique TokenCommittee functionality
func TestTokenCommittee(t *testing.T) {
	addresses := []sdk.AccAddress{
		sdk.AccAddress(crypto.AddressHash([]byte("KavaTest1"))),
		sdk.AccAddress(crypto.AddressHash([]byte("KavaTest2"))),
		sdk.AccAddress(crypto.AddressHash([]byte("KavaTest3"))),
	}

	testCases := []struct {
		name       string
		committee  TokenCommittee
		expectPass bool
	}{
		{
			name: "normal",
			committee: TokenCommittee{
				BaseCommittee: BaseCommittee{
					ID:               1,
					Description:      "This token committee is for testing.",
					Members:          addresses[:3],
					Permissions:      []Permission{GodPermission{}},
					VoteThreshold:    d("0.667"),
					ProposalDuration: time.Hour * 24 * 7,
					TallyOption:      FirstPastThePost,
				},
				Quorum:     d("0.4"),
				TallyDenom: "hard",
			},
			expectPass: true,
		},
		{
			name: "nil quorum",
			committee: TokenCommittee{
				BaseCommittee: BaseCommittee{
					ID:               1,
					Description:      "This token committee is for testing.",
					Members:          addresses[:3],
					Permissions:      []Permission{GodPermission{}},
					VoteThreshold:    d("0.667"),
					ProposalDuration: time.Hour * 24 * 7,
					TallyOption:      FirstPastThePost,
				},
				Quorum:     sdk.Dec{Int: nil},
				TallyDenom: "hard",
			},
			expectPass: false,
		},
		{
			name: "negative quorum",
			committee: TokenCommittee{
				BaseCommittee: BaseCommittee{
					ID:               1,
					Description:      "This token committee is for testing.",
					Members:          addresses[:3],
					Permissions:      []Permission{GodPermission{}},
					VoteThreshold:    d("0.667"),
					ProposalDuration: time.Hour * 24 * 7,
					TallyOption:      FirstPastThePost,
				},
				Quorum:     d("-0.1"),
				TallyDenom: "hard",
			},
			expectPass: false,
		},
		{
			name: "quroum greater than 1",
			committee: TokenCommittee{
				BaseCommittee: BaseCommittee{
					ID:               1,
					Description:      "This token committee is for testing.",
					Members:          addresses[:3],
					Permissions:      []Permission{GodPermission{}},
					VoteThreshold:    d("0.667"),
					ProposalDuration: time.Hour * 24 * 7,
					TallyOption:      FirstPastThePost,
				},
				Quorum:     d("1.001"),
				TallyDenom: "hard",
			},
			expectPass: false,
		},
		{
			name: "bond denom as tally denom",
			committee: TokenCommittee{
				BaseCommittee: BaseCommittee{
					ID:               1,
					Description:      "This token committee is for testing.",
					Members:          addresses[:3],
					Permissions:      []Permission{GodPermission{}},
					VoteThreshold:    d("0.667"),
					ProposalDuration: time.Hour * 24 * 7,
					TallyOption:      FirstPastThePost,
				},
				Quorum:     d("0.4"),
				TallyDenom: BondDenom,
			},
			expectPass: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			require.Equal(t, TokenCommitteeType, tc.committee.GetType())

			err := tc.committee.Validate()
			if tc.expectPass {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}
