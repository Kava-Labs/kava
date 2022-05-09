package v0_16

import (
	"io/ioutil"
	"path/filepath"
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	v036distr "github.com/cosmos/cosmos-sdk/x/distribution/legacy/v036"
	v040distr "github.com/cosmos/cosmos-sdk/x/distribution/types"
	v036gov "github.com/cosmos/cosmos-sdk/x/gov/legacy/v036"
	v040gov "github.com/cosmos/cosmos-sdk/x/gov/types"
	v036params "github.com/cosmos/cosmos-sdk/x/params/legacy/v036"
	v040params "github.com/cosmos/cosmos-sdk/x/params/types/proposal"
	v038upgrade "github.com/cosmos/cosmos-sdk/x/upgrade/legacy/v038"
	v040upgrade "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	"github.com/stretchr/testify/suite"

	app "github.com/kava-labs/kava/app"
	v015committee "github.com/kava-labs/kava/x/committee/legacy/v0_15"
	v016committee "github.com/kava-labs/kava/x/committee/types"
	v015kavadist "github.com/kava-labs/kava/x/kavadist/legacy/v0_15"
	v016kavadist "github.com/kava-labs/kava/x/kavadist/types"
	v015pricefeed "github.com/kava-labs/kava/x/pricefeed/legacy/v0_15"
	v016pricefeedtypes "github.com/kava-labs/kava/x/pricefeed/types"
)

type migrateTestSuite struct {
	suite.Suite

	addresses            []sdk.AccAddress
	v15genstate          v015committee.GenesisState
	v15pricefeedgenstate v015pricefeed.GenesisState
	cdc                  codec.Codec
	legacyCdc            *codec.LegacyAmino
}

func (s *migrateTestSuite) SetupTest() {
	app.SetSDKConfig()

	s.v15genstate = v015committee.GenesisState{
		Committees:     v015committee.Committees{},
		Proposals:      []v015committee.Proposal{},
		NextProposalID: 1,
		Votes:          []v015committee.Vote{},
	}

	config := app.MakeEncodingConfig()
	s.cdc = config.Marshaler

	legacyCodec := codec.NewLegacyAmino()
	v015committee.RegisterLegacyAminoCodec(legacyCodec)
	v036distr.RegisterLegacyAminoCodec(legacyCodec)
	v038upgrade.RegisterLegacyAminoCodec(legacyCodec)
	v036params.RegisterLegacyAminoCodec(legacyCodec)
	v015kavadist.RegisterLegacyAminoCodec(legacyCodec)

	s.legacyCdc = legacyCodec

	_, accAddresses := app.GeneratePrivKeyAddressPairs(10)
	s.addresses = accAddresses
}

func (s *migrateTestSuite) TestMigrate_JSON() {
	file := filepath.Join("testdata", "v15-committee.json")
	data, err := ioutil.ReadFile(file)
	s.Require().NoError(err)

	err = s.legacyCdc.UnmarshalJSON(data, &s.v15genstate)
	s.Require().NoError(err)

	pricefeedFile := filepath.Join("testdata", "v15-pricefeed.json")
	pricefeedData, err := ioutil.ReadFile(pricefeedFile)
	s.Require().NoError(err)

	err = s.legacyCdc.UnmarshalJSON(pricefeedData, &s.v15pricefeedgenstate)
	s.Require().NoError(err)

	genstate := Migrate(s.v15genstate, s.v15pricefeedgenstate)
	actual := s.cdc.MustMarshalJSON(genstate)

	file = filepath.Join("testdata", "v16-committee.json")
	expected, err := ioutil.ReadFile(file)

	s.Require().NoError(err)
	s.Require().JSONEq(string(expected), string(actual))
}

func (s *migrateTestSuite) TestMigrate_PricefeedPermissions() {
	file := filepath.Join("testdata", "v15-committee.json")
	data, err := ioutil.ReadFile(file)
	s.Require().NoError(err)

	err = s.legacyCdc.UnmarshalJSON(data, &s.v15genstate)
	s.Require().NoError(err)

	pricefeedFile := filepath.Join("testdata", "v15-pricefeed.json")
	pricefeedData, err := ioutil.ReadFile(pricefeedFile)
	s.Require().NoError(err)

	err = s.legacyCdc.UnmarshalJSON(pricefeedData, &s.v15pricefeedgenstate)
	s.Require().NoError(err)

	genstate := Migrate(s.v15genstate, s.v15pricefeedgenstate)

	uniqueMarkets := make(map[string]bool)

	for _, permission := range genstate.GetCommittees()[0].GetPermissions() {
		paramChangePermission, ok := permission.(v016committee.ParamsChangePermission)
		if !ok {
			continue
		}

		for _, allowedParamChanges := range paramChangePermission.AllowedParamsChanges {
			if allowedParamChanges.Subspace == v016pricefeedtypes.ModuleName {
				for _, req := range allowedParamChanges.MultiSubparamsRequirements {
					_, found := uniqueMarkets[req.Val]
					s.Require().Falsef(
						found,
						"pricefeed market MultiSubparamsRequirement for %v is duplicated",
						req.Val,
					)

					uniqueMarkets[req.Val] = true
				}
			}
		}
	}
}

func (s *migrateTestSuite) TestMigrate_TokenCommittee() {
	oldTokenCommittee := v015committee.TokenCommittee{
		BaseCommittee: v015committee.BaseCommittee{
			ID:               1,
			Description:      "test",
			Members:          s.addresses,
			Permissions:      []v015committee.Permission{},
			VoteThreshold:    sdk.NewDec(40),
			ProposalDuration: time.Hour * 24 * 7,
			TallyOption:      v015committee.Deadline,
		},
		Quorum:     sdk.NewDec(40),
		TallyDenom: "ukava",
	}

	expectedTokenCommittee, err := v016committee.NewTokenCommittee(1, "test", s.addresses, []v016committee.Permission{}, oldTokenCommittee.VoteThreshold, oldTokenCommittee.ProposalDuration, v016committee.TALLY_OPTION_DEADLINE, oldTokenCommittee.Quorum, oldTokenCommittee.TallyDenom)
	s.Require().NoError(err)

	s.v15genstate.Committees = []v015committee.Committee{oldTokenCommittee}
	genState := Migrate(s.v15genstate, s.v15pricefeedgenstate)
	s.Require().Len(genState.Committees, 1)
	s.Equal(expectedTokenCommittee, genState.GetCommittees()[0])
}

func (s *migrateTestSuite) TestMigrate_Committee_TallyOption() {
	testcases := []struct {
		name            string
		v015tallyOption v015committee.TallyOption
		v016tallyOption v016committee.TallyOption
	}{
		{
			name:            "null tally option",
			v015tallyOption: v015committee.NullTallyOption,
			v016tallyOption: v016committee.TALLY_OPTION_UNSPECIFIED,
		},
		{
			name:            "first past the post tally option",
			v015tallyOption: v015committee.FirstPastThePost,
			v016tallyOption: v016committee.TALLY_OPTION_FIRST_PAST_THE_POST,
		},
		{
			name:            "deadline tally",
			v015tallyOption: v015committee.Deadline,
			v016tallyOption: v016committee.TALLY_OPTION_DEADLINE,
		},
	}

	for _, tc := range testcases {
		s.Run(tc.name, func() {
			oldCommittee := v015committee.MemberCommittee{
				BaseCommittee: v015committee.BaseCommittee{
					ID:               2,
					Description:      "test",
					Members:          s.addresses,
					Permissions:      []v015committee.Permission{},
					VoteThreshold:    sdk.NewDec(40),
					ProposalDuration: time.Hour * 24 * 7,
					TallyOption:      tc.v015tallyOption,
				},
			}
			expectedProposal, err := v016committee.NewMemberCommittee(2, "test", s.addresses, []v016committee.Permission{}, oldCommittee.VoteThreshold, oldCommittee.ProposalDuration, tc.v016tallyOption)
			s.Require().NoError(err)
			s.v15genstate.Committees = []v015committee.Committee{oldCommittee}
			genState := Migrate(s.v15genstate, s.v15pricefeedgenstate)
			s.Require().Len(genState.Committees, 1)
			s.Equal(expectedProposal.GetTallyOption(), genState.GetCommittees()[0].GetTallyOption())
		})
	}
}

func (s *migrateTestSuite) TestMigrate_Committee_Permissions() {
	testcases := []struct {
		name           string
		v015permission v015committee.Permission
		v016permission v016committee.Permission
	}{
		{
			name:           "god permission",
			v015permission: v015committee.GodPermission{},
			v016permission: &v016committee.GodPermission{},
		},
		{
			name:           "text permission",
			v015permission: v015committee.TextPermission{},
			v016permission: &v016committee.TextPermission{},
		},
		{
			name:           "software upgrade permission",
			v015permission: v015committee.SoftwareUpgradePermission{},
			v016permission: &v016committee.SoftwareUpgradePermission{},
		},
		{
			name: "simple param change permission",
			v015permission: v015committee.SimpleParamChangePermission{
				AllowedParams: []v015committee.AllowedParam{
					{Subspace: "staking", Key: "bondDenom"},
					{Subspace: "test", Key: "testkey"},
				},
			},
			v016permission: &v016committee.ParamsChangePermission{
				AllowedParamsChanges: []v016committee.AllowedParamsChange{
					{Subspace: "staking", Key: "bondDenom"},
					{Subspace: "test", Key: "testkey"},
				},
			},
		},
		{
			name: "sub param change permission",
			v015permission: v015committee.SubParamChangePermission{
				AllowedParams: []v015committee.AllowedParam{
					{Subspace: "staking", Key: "bondDenom"},
					{Subspace: "test", Key: "testkey"},
				},
			},
			v016permission: &v016committee.ParamsChangePermission{
				AllowedParamsChanges: []v016committee.AllowedParamsChange{
					{Subspace: "staking", Key: "bondDenom"},
					{Subspace: "test", Key: "testkey"},
				},
			},
		},
	}

	for _, tc := range testcases {
		s.Run(tc.name, func() {
			oldCommittee := v015committee.MemberCommittee{
				BaseCommittee: v015committee.BaseCommittee{
					ID:               2,
					Description:      "test",
					Members:          s.addresses,
					Permissions:      []v015committee.Permission{tc.v015permission},
					VoteThreshold:    sdk.NewDec(40),
					ProposalDuration: time.Hour * 24 * 7,
					TallyOption:      v015committee.FirstPastThePost,
				},
			}
			expectedProposal, err := v016committee.NewMemberCommittee(2, "test", s.addresses, []v016committee.Permission{tc.v016permission}, oldCommittee.VoteThreshold, oldCommittee.ProposalDuration, v016committee.TALLY_OPTION_FIRST_PAST_THE_POST)
			s.Require().NoError(err)
			s.v15genstate.Committees = []v015committee.Committee{oldCommittee}
			genState := Migrate(s.v15genstate, s.v15pricefeedgenstate)
			s.Require().Len(genState.Committees, 1)
			s.Equal(expectedProposal, genState.GetCommittees()[0])
		})
	}
}

func (s *migrateTestSuite) TestMigrate_Proposals() {
	testcases := []struct {
		name         string
		v015proposal v015committee.PubProposal
		v016proposal v016committee.PubProposal
	}{
		{
			name:         "text proposal",
			v015proposal: v036gov.NewTextProposal("A Title", "A description of this proposal."),
			v016proposal: v040gov.NewTextProposal("A Title", "A description of this proposal."),
		},
		{
			name: "community poll spend proposal",
			v015proposal: v036distr.CommunityPoolSpendProposal{
				Title:       "Community Pool Spend",
				Description: "Fund the community pool.",
				Recipient:   s.addresses[0],
				Amount:      sdk.NewCoins(sdk.NewInt64Coin("ukava", 10)),
			},
			v016proposal: &v040distr.CommunityPoolSpendProposal{
				Title:       "Community Pool Spend",
				Description: "Fund the community pool.",
				Recipient:   s.addresses[0].String(),
				Amount:      sdk.NewCoins(sdk.NewInt64Coin("ukava", 10)),
			},
		},
		{
			name: "software upgrade with deprecated plan time",
			v015proposal: v038upgrade.SoftwareUpgradeProposal{
				Title:       "Test",
				Description: "Test",
				Plan:        v038upgrade.Plan{Name: "Test", Height: 100, Time: time.Now()},
			},
			v016proposal: &v040upgrade.SoftwareUpgradeProposal{
				Title:       "Test",
				Description: "Test",
				Plan:        v040upgrade.Plan{Name: "Test", Height: 100},
			},
		},
		{
			name: "param change proposal",
			v015proposal: v036params.ParameterChangeProposal{
				Title:       "Test",
				Description: "Test",
				Changes: []v036params.ParamChange{
					{Subspace: "Test", Key: "Test", Value: "Test"},
				},
			},
			v016proposal: &v040params.ParameterChangeProposal{
				Title:       "Test",
				Description: "Test",
				Changes: []v040params.ParamChange{
					{Subspace: "Test", Key: "Test", Value: "Test"},
				},
			},
		},
		{
			name: "kavadist community pool multi spend proposal",
			v015proposal: v015kavadist.CommunityPoolMultiSpendProposal{
				Title:       "Test",
				Description: "Test",
				RecipientList: v015kavadist.MultiSpendRecipients{
					v015kavadist.MultiSpendRecipient{
						Address: s.addresses[0],
						Amount:  sdk.NewCoins(sdk.NewInt64Coin("ukava", 10)),
					},
				},
			},
			v016proposal: &v016kavadist.CommunityPoolMultiSpendProposal{
				Title:       "Test",
				Description: "Test",
				RecipientList: []v016kavadist.MultiSpendRecipient{
					{
						Address: s.addresses[0].String(),
						Amount:  sdk.NewCoins(sdk.NewInt64Coin("ukava", 10)),
					},
				},
			},
		},
		{
			name: "cancel software upgrade proposal",
			v015proposal: v038upgrade.CancelSoftwareUpgradeProposal{
				Title:       "Test",
				Description: "Test",
			},
			v016proposal: &v040upgrade.CancelSoftwareUpgradeProposal{
				Title:       "Test",
				Description: "Test",
			},
		},
	}

	for _, tc := range testcases {
		s.Run(tc.name, func() {
			deadline := time.Now().Add(2 * time.Hour)
			oldProposal := v015committee.Proposal{
				PubProposal: tc.v015proposal,
				ID:          1,
				CommitteeID: 2,
				Deadline:    deadline,
			}
			expectedProposal, err := v016committee.NewProposal(tc.v016proposal, 1, 2, deadline)
			s.Require().NoError(err)
			s.v15genstate.Proposals = []v015committee.Proposal{oldProposal}
			genState := Migrate(s.v15genstate, s.v15pricefeedgenstate)
			s.Require().Len(genState.Proposals, 1)
			s.Equal(expectedProposal, genState.Proposals[0])
		})
	}
}

func (s *migrateTestSuite) TestMigrate_Votes() {
	testcases := []struct {
		name         string
		v015voteType v015committee.VoteType
		v016VoteType v016committee.VoteType
	}{
		{
			name:         "yes vote",
			v015voteType: v015committee.Yes,
			v016VoteType: v016committee.VOTE_TYPE_YES,
		},
		{
			name:         "no vote",
			v015voteType: v015committee.No,
			v016VoteType: v016committee.VOTE_TYPE_NO,
		},
		{
			name:         "null vote",
			v015voteType: v015committee.NullVoteType,
			v016VoteType: v016committee.VOTE_TYPE_UNSPECIFIED,
		},
		{
			name:         "abstain vote",
			v015voteType: v015committee.Abstain,
			v016VoteType: v016committee.VOTE_TYPE_ABSTAIN,
		},
	}

	for _, tc := range testcases {
		s.Run(tc.name, func() {
			oldVote := v015committee.Vote{
				ProposalID: 1,
				Voter:      s.addresses[0],
				VoteType:   tc.v015voteType,
			}
			expectedVote := v016committee.Vote{
				ProposalID: 1,
				Voter:      s.addresses[0],
				VoteType:   tc.v016VoteType,
			}
			s.v15genstate.Votes = []v015committee.Vote{oldVote}
			genState := Migrate(s.v15genstate, s.v15pricefeedgenstate)
			s.Require().Len(genState.Votes, 1)
			s.Equal(expectedVote, genState.Votes[0])
		})
	}
}

func TestMigrateTestSuite(t *testing.T) {
	suite.Run(t, new(migrateTestSuite))
}
