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
	addresses := []string{
		sdk.AccAddress(crypto.AddressHash([]byte("KavaTest1"))).String(),
		sdk.AccAddress(crypto.AddressHash([]byte("KavaTest2"))).String(),
		sdk.AccAddress(crypto.AddressHash([]byte("KavaTest3"))).String(),
	}

	testCases := []struct {
		name            string
		createCommittee func() (*MemberCommittee, error)
		expectPass      bool
	}{
		{
			name: "normal",
			createCommittee: func() (*MemberCommittee, error) {
				return NewMemberCommittee(
					1,
					"This base committee is for testing.",
					addresses[:3],
					[]Permission{&GodPermission{}},
					d("0.667"),
					time.Hour*24*7,
					TALLY_OPTION_FIRST_PAST_THE_POST,
				)
			},
			expectPass: true,
		},
		{
			name: "description length too long",
			createCommittee: func() (*MemberCommittee, error) {
				return NewMemberCommittee(
					1,
					fmt.Sprintln("This base committee has a long description.",
						"This base committee has a long description. This base committee has a long description.",
						"This base committee has a long description. This base committee has a long description.",
						"This base committee has a long description. This base committee has a long description.",
						"This base committee has a long description. This base committee has a long description.",
						"This base committee has a long description. This base committee has a long description.",
						"This base committee has a long description. This base committee has a long description.",
						"This base committee has a long description. This base committee has a long description."),
					addresses[:3],
					[]Permission{&GodPermission{}},
					d("0.667"),
					time.Hour*24*7,
					TALLY_OPTION_FIRST_PAST_THE_POST,
				)
			},
			expectPass: false,
		},
		{
			name: "no members",
			createCommittee: func() (*MemberCommittee, error) {
				return NewMemberCommittee(
					1,
					"This base committee is for testing.",
					[]string{},
					[]Permission{&GodPermission{}},
					d("0.667"),
					time.Hour*24*7,
					TALLY_OPTION_FIRST_PAST_THE_POST,
				)
			},
			expectPass: false,
		},
		{
			name: "duplicate member",
			createCommittee: func() (*MemberCommittee, error) {
				return NewMemberCommittee(
					1,
					"This base committee is for testing.",
					[]string{addresses[2], addresses[2]},
					[]Permission{&GodPermission{}},
					d("0.667"),
					time.Hour*24*7,
					TALLY_OPTION_FIRST_PAST_THE_POST,
				)
			},
			expectPass: false,
		},
		{
			name: "nil permissions",
			createCommittee: func() (*MemberCommittee, error) {
				return NewMemberCommittee(
					1,
					"This base committee is for testing.",
					addresses[:3],
					[]Permission{nil},
					d("0.667"),
					time.Hour*24*7,
					TALLY_OPTION_FIRST_PAST_THE_POST,
				)
			},
			expectPass: false,
		},
		{
			name: "negative proposal duration",
			createCommittee: func() (*MemberCommittee, error) {
				return NewMemberCommittee(
					1,
					"This base committee is for testing.",
					addresses[:3],
					[]Permission{&GodPermission{}},
					d("0.667"),
					time.Hour*24*-7,
					TALLY_OPTION_FIRST_PAST_THE_POST,
				)
			},
			expectPass: false,
		},
		{
			name: "vote threshold is nil",
			createCommittee: func() (*MemberCommittee, error) {
				return NewMemberCommittee(
					1,
					"This base committee is for testing.",
					addresses[:3],
					[]Permission{&GodPermission{}},
					sdk.Dec{},
					time.Hour*24*7,
					TALLY_OPTION_FIRST_PAST_THE_POST,
				)
			},
			expectPass: false,
		},
		{
			name: "vote threshold is 0",
			createCommittee: func() (*MemberCommittee, error) {
				return NewMemberCommittee(
					1,
					"This base committee is for testing.",
					addresses[:3],
					[]Permission{&GodPermission{}},
					d("0"),
					time.Hour*24*7,
					TALLY_OPTION_FIRST_PAST_THE_POST,
				)
			},
			expectPass: false,
		},
		{
			name: "vote threshold above 1",
			createCommittee: func() (*MemberCommittee, error) {
				return NewMemberCommittee(
					1,
					"This base committee is for testing.",
					addresses[:3],
					[]Permission{&GodPermission{}},
					d("1.001"),
					time.Hour*24*7,
					TALLY_OPTION_FIRST_PAST_THE_POST,
				)
			},
			expectPass: false,
		},
		{
			name: "invalid tally option",
			createCommittee: func() (*MemberCommittee, error) {
				return NewMemberCommittee(
					1,
					"This base committee is for testing.",
					addresses[:3],
					[]Permission{&GodPermission{}},
					d("0.667"),
					time.Hour*24*7,
					TALLY_OPTION_UNSPECIFIED,
				)
			},
			expectPass: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			committee, err := tc.createCommittee()
			if err != nil {
				require.False(t, tc.expectPass)
			} else {
				err = committee.BaseCommittee.Validate()
				if tc.expectPass {
					require.NoError(t, err)
				} else {
					require.Error(t, err)
				}
			}
		})
	}
}

// TestMemberCommittee is an alias for BaseCommittee that has 'MemberCommittee' type
func TestMemberCommittee(t *testing.T) {
	addresses := []string{
		sdk.AccAddress(crypto.AddressHash([]byte("KavaTest1"))).String(),
		sdk.AccAddress(crypto.AddressHash([]byte("KavaTest2"))).String(),
		sdk.AccAddress(crypto.AddressHash([]byte("KavaTest3"))).String(),
	}

	testCases := []struct {
		name            string
		createCommittee func() (*MemberCommittee, error)
		expectPass      bool
	}{
		{
			name: "normal",
			createCommittee: func() (*MemberCommittee, error) {
				return NewMemberCommittee(
					1,
					"This member committee is for testing.",
					addresses[:3],
					[]Permission{&GodPermission{}},
					d("0.667"),
					time.Hour*24*7,
					TALLY_OPTION_FIRST_PAST_THE_POST,
				)
			},
			expectPass: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			committee, err := tc.createCommittee()
			require.NoError(t, err)
			require.Equal(t, MemberCommitteeType, committee.GetType())

			err = committee.Validate()
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
	addresses := []string{
		sdk.AccAddress(crypto.AddressHash([]byte("KavaTest1"))).String(),
		sdk.AccAddress(crypto.AddressHash([]byte("KavaTest2"))).String(),
		sdk.AccAddress(crypto.AddressHash([]byte("KavaTest3"))).String(),
	}

	testCases := []struct {
		name            string
		createCommittee func() (*TokenCommittee, error)
		expectPass      bool
	}{
		{
			name: "normal",
			createCommittee: func() (*TokenCommittee, error) {
				return NewTokenCommittee(
					1,
					"This token committee is for testing.",
					addresses[:3],
					[]Permission{&GodPermission{}},
					d("0.667"),
					time.Hour*24*7,
					TALLY_OPTION_FIRST_PAST_THE_POST,
					d("0.4"),
					"hard",
				)
			},
			expectPass: true,
		},
		{
			name: "nil quorum",
			createCommittee: func() (*TokenCommittee, error) {
				return NewTokenCommittee(
					1,
					"This token committee is for testing.",
					addresses[:3],
					[]Permission{&GodPermission{}},
					d("0.667"),
					time.Hour*24*7,
					TALLY_OPTION_FIRST_PAST_THE_POST,
					sdk.Dec{},
					"hard",
				)
			},
			expectPass: false,
		},
		{
			name: "negative quorum",
			createCommittee: func() (*TokenCommittee, error) {
				return NewTokenCommittee(
					1,
					"This token committee is for testing.",
					addresses[:3],
					[]Permission{&GodPermission{}},
					d("0.667"),
					time.Hour*24*7,
					TALLY_OPTION_FIRST_PAST_THE_POST,
					d("-0.1"),
					"hard",
				)
			},
			expectPass: false,
		},
		{
			name: "quroum greater than 1",
			createCommittee: func() (*TokenCommittee, error) {
				return NewTokenCommittee(
					1,
					"This token committee is for testing.",
					addresses[:3],
					[]Permission{&GodPermission{}},
					d("0.667"),
					time.Hour*24*7,
					TALLY_OPTION_FIRST_PAST_THE_POST,
					d("1.001"),
					"hard",
				)
			},
			expectPass: false,
		},
		{
			name: "bond denom as tally denom",
			createCommittee: func() (*TokenCommittee, error) {
				return NewTokenCommittee(
					1,
					"This token committee is for testing.",
					addresses[:3],
					[]Permission{&GodPermission{}},
					d("0.667"),
					time.Hour*24*7,
					TALLY_OPTION_FIRST_PAST_THE_POST,
					d("0.4"),
					BondDenom,
				)
			},
			expectPass: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			committee, err := tc.createCommittee()
			require.NoError(t, err)
			require.Equal(t, TokenCommitteeType, committee.GetType())

			err = committee.Validate()
			if tc.expectPass {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}
