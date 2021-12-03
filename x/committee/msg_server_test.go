package committee_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"
	proposal "github.com/cosmos/cosmos-sdk/x/params/types/proposal"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	"github.com/gogo/protobuf/proto"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/committee"
	"github.com/kava-labs/kava/x/committee/keeper"
	"github.com/kava-labs/kava/x/committee/types"
	swaptypes "github.com/kava-labs/kava/x/swap/types"
)

var _ types.PubProposal = &UnregisteredPubProposal{}

// UnregisteredPubProposal is a pubproposal type that is not registered on the amino codec.
type UnregisteredPubProposal struct {
	proto.Message
}

func (*UnregisteredPubProposal) GetTitle() string       { return "unregistered" }
func (*UnregisteredPubProposal) GetDescription() string { return "unregistered" }
func (*UnregisteredPubProposal) ProposalRoute() string  { return "unregistered" }
func (*UnregisteredPubProposal) ProposalType() string   { return "unregistered" }
func (*UnregisteredPubProposal) ValidateBasic() error   { return nil }
func (*UnregisteredPubProposal) String() string         { return "unregistered" }

//NewDistributionGenesisWithPool creates a default distribution genesis state with some coins in the community pool.
//func NewDistributionGenesisWithPool(communityPoolCoins sdk.Coins) app.GenesisState {
//gs := distribution.DefaultGenesisState()
//gs.FeePool = distribution.FeePool{CommunityPool: sdk.NewDecCoinsFromCoins(communityPoolCoins...)}
//return app.GenesisState{distribution.ModuleName: distribution.ModuleCdc.MustMarshalJSON(gs)}
//}

type MsgServerTestSuite struct {
	suite.Suite

	app       app.TestApp
	keeper    keeper.Keeper
	msgServer types.MsgServer
	ctx       sdk.Context
	addresses []sdk.AccAddress

	communityPoolAmt sdk.Coins
}

func (suite *MsgServerTestSuite) SetupTest() {
	_, suite.addresses = app.GeneratePrivKeyAddressPairs(5)
	suite.app = app.NewTestApp()
	suite.keeper = suite.app.GetCommitteeKeeper()
	suite.msgServer = committee.NewMsgServerImpl(suite.keeper)
	encodingCfg := app.MakeEncodingConfig()
	cdc := encodingCfg.Marshaler

	memberCommittee, err := types.NewMemberCommittee(
		1,
		"This committee is for testing.",
		suite.addresses[:3],
		[]types.Permission{&types.GodPermission{}},
		sdk.MustNewDecFromStr("0.5"),
		time.Hour*24*7,
		types.TALLY_OPTION_FIRST_PAST_THE_POST,
	)
	suite.Require().NoError(err)

	firstBlockTime := time.Date(1998, time.January, 1, 1, 0, 0, 0, time.UTC)
	testGenesis := types.NewGenesisState(
		3,
		[]types.Committee{memberCommittee},
		[]types.Proposal{},
		[]types.Vote{},
	)
	suite.communityPoolAmt = sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(1000)))
	suite.app.InitializeFromGenesisStates(
		app.GenesisState{types.ModuleName: cdc.MustMarshalJSON(testGenesis)},
		// TODO: not used?
		//NewDistributionGenesisWithPool(suite.communityPoolAmt),
	)
	suite.ctx = suite.app.NewContext(true, tmproto.Header{Height: 1, Time: firstBlockTime})
}

func (suite *MsgServerTestSuite) TestSubmitProposalMsg_Valid() {
	msg, err := types.NewMsgSubmitProposal(
		proposal.NewParameterChangeProposal(
			"A Title",
			"A description of this proposal.",
			[]proposal.ParamChange{{
				Subspace: swaptypes.ModuleName,
				Key:      string(swaptypes.KeySwapFee),
				Value:    "\"0.001500000000000000\"",
			}},
		),
		suite.addresses[0],
		1,
	)
	suite.Require().NoError(err)

	res, err := suite.msgServer.SubmitProposal(sdk.WrapSDKContext(suite.ctx), msg)

	suite.NoError(err)
	_, found := suite.keeper.GetProposal(suite.ctx, res.ProposalID)
	suite.True(found)
}

func (suite *MsgServerTestSuite) TestSubmitProposalMsg_Invalid() {
	var committeeID uint64 = 1
	msg, err := types.NewMsgSubmitProposal(
		proposal.NewParameterChangeProposal(
			"A Title",
			"A description of this proposal.",
			[]proposal.ParamChange{{
				Subspace: swaptypes.ModuleName,
				Key:      "nonsense-key",
				Value:    "nonsense-value",
			}},
		),
		suite.addresses[0],
		committeeID,
	)
	suite.Require().NoError(err)

	_, err = suite.msgServer.SubmitProposal(sdk.WrapSDKContext(suite.ctx), msg)

	suite.Error(err)
	suite.Empty(
		suite.keeper.GetProposalsByCommittee(suite.ctx, committeeID),
		"proposal found when none should exist",
	)

}

func (suite *MsgServerTestSuite) TestSubmitProposalMsg_ValidUpgrade() {
	msg, err := types.NewMsgSubmitProposal(
		upgradetypes.NewSoftwareUpgradeProposal(
			"A Title",
			"A description of this proposal.",
			upgradetypes.Plan{
				Name:   "emergency-shutdown-1", // identifier for the upgrade
				Height: 100000,
				Info:   "Some information about the shutdown.",
			},
		),
		suite.addresses[0],
		1,
	)
	suite.Require().NoError(err)

	res, err := suite.msgServer.SubmitProposal(sdk.WrapSDKContext(suite.ctx), msg)

	suite.NoError(err)
	_, found := suite.keeper.GetProposal(suite.ctx, res.ProposalID)
	suite.True(found)
}

// TODO: create a unregisted proto for tests?
func (suite *MsgServerTestSuite) TestSubmitProposalMsg_Unregistered() {
	var committeeID uint64 = 1
	msg, err := types.NewMsgSubmitProposal(
		&UnregisteredPubProposal{},
		suite.addresses[0],
		committeeID,
	)
	suite.Require().NoError(err)

	_, err = suite.msgServer.SubmitProposal(sdk.WrapSDKContext(suite.ctx), msg)

	suite.Error(err)
	suite.Empty(
		suite.keeper.GetProposalsByCommittee(suite.ctx, committeeID),
		"proposal found when none should exist",
	)
}

func (suite *MsgServerTestSuite) TestSubmitProposalMsgAndVote() {
	msg, err := types.NewMsgSubmitProposal(
		proposal.NewParameterChangeProposal(
			"A Title",
			"A description of this proposal.",
			[]proposal.ParamChange{{
				Subspace: swaptypes.ModuleName,
				Key:      string(swaptypes.KeySwapFee),
				Value:    "\"0.001500000000000000\"",
			}},
		),
		suite.addresses[0],
		1,
	)
	suite.Require().NoError(err)

	res, err := suite.msgServer.SubmitProposal(sdk.WrapSDKContext(suite.ctx), msg)
	suite.Require().NoError(err)

	proposal, found := suite.keeper.GetProposal(suite.ctx, res.ProposalID)
	suite.Require().True(found)

	msgVote := types.NewMsgVote(suite.addresses[0], proposal.ID, types.VOTE_TYPE_YES)
	_, err = suite.msgServer.Vote(sdk.WrapSDKContext(suite.ctx), msgVote)
	suite.Require().NoError(err)
}

func TestMsgServerTestSuite(t *testing.T) {
	suite.Run(t, new(MsgServerTestSuite))
}
