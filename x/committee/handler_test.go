package committee_test

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/distribution"
	"github.com/cosmos/cosmos-sdk/x/params"
	"github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/kava-labs/kava/app"
	cdptypes "github.com/kava-labs/kava/x/cdp/types"
	"github.com/kava-labs/kava/x/committee"
	"github.com/kava-labs/kava/x/committee/keeper"
	"github.com/kava-labs/kava/x/committee/types"
)

// NewDistributionGenesisWithPool creates a default distribution genesis state with some coins in the community pool.
func NewDistributionGenesisWithPool(communityPoolCoins sdk.Coins) app.GenesisState {
	gs := distribution.DefaultGenesisState()
	gs.FeePool = distribution.FeePool{CommunityPool: sdk.NewDecCoins(communityPoolCoins)}
	return app.GenesisState{distribution.ModuleName: distribution.ModuleCdc.MustMarshalJSON(gs)}
}

type HandlerTestSuite struct {
	suite.Suite

	app       app.TestApp
	keeper    keeper.Keeper
	handler   sdk.Handler
	ctx       sdk.Context
	addresses []sdk.AccAddress

	communityPoolAmt sdk.Coins
}

func (suite *HandlerTestSuite) SetupTest() {
	_, suite.addresses = app.GeneratePrivKeyAddressPairs(5)
	suite.app = app.NewTestApp()
	suite.keeper = suite.app.GetCommitteeKeeper()
	suite.handler = committee.NewHandler(suite.keeper)

	firstBlockTime := time.Date(1998, time.January, 1, 1, 0, 0, 0, time.UTC)
	testGenesis := types.NewGenesisState(
		3,
		[]types.Committee{
			{
				ID:               1,
				Description:      "This committee is for testing.",
				Members:          suite.addresses[:3],
				Permissions:      []types.Permission{types.GodPermission{}},
				VoteThreshold:    d("0.5"),
				ProposalDuration: time.Hour * 24 * 7,
			},
		},
		[]types.Proposal{},
		[]types.Vote{},
	)
	suite.communityPoolAmt = cs(c("ukava", 1000))
	suite.app.InitializeFromGenesisStates(
		NewCommitteeGenesisState(suite.app.Codec(), testGenesis),
		NewDistributionGenesisWithPool(suite.communityPoolAmt),
	)
	suite.ctx = suite.app.NewContext(true, abci.Header{Height: 1, Time: firstBlockTime})
}

func (suite *HandlerTestSuite) TestSubmitProposalMsg_Valid() {
	msg := committee.NewMsgSubmitProposal(
		params.NewParameterChangeProposal(
			"A Title",
			"A description of this proposal.",
			[]params.ParamChange{{
				Subspace: cdptypes.ModuleName,
				Key:      string(cdptypes.KeyDebtThreshold),
				Value:    string(types.ModuleCdc.MustMarshalJSON(i(1000000))),
			}},
		),
		suite.addresses[0],
		1,
	)

	res := suite.handler(suite.ctx, msg)

	suite.True(res.IsOK())
	_, found := suite.keeper.GetProposal(suite.ctx, types.Uint64FromBytes(res.Data))
	suite.True(found)
}

func (suite *HandlerTestSuite) TestSubmitProposalMsg_Invalid() {
	msg := types.NewMsgSubmitProposal(
		params.NewParameterChangeProposal(
			"A Title",
			"A description of this proposal.",
			[]params.ParamChange{{
				Subspace: cdptypes.ModuleName,
				Key:      "nonsense-key",
				Value:    "nonsense-value",
			}},
		),
		suite.addresses[0],
		1,
	)

	res := suite.handler(suite.ctx, msg)

	suite.False(res.IsOK())
	suite.keeper.IterateProposals(suite.ctx, func(p types.Proposal) bool {
		suite.Fail("proposal found when none should exist")
		return true
	})
}

func (suite *HandlerTestSuite) TestSubmitProposalMsg_Unregistered() {
	msg := types.NewMsgSubmitProposal(
		UnregisteredPubProposal{},
		suite.addresses[0],
		1,
	)

	res := suite.handler(suite.ctx, msg)

	suite.False(res.IsOK())
	suite.keeper.IterateProposals(suite.ctx, func(p types.Proposal) bool {
		suite.Fail("proposal found when none should exist")
		return true
	})
}

func (suite *HandlerTestSuite) TestMsgAddVote_ProposalPass() {
	previousCDPDebtThreshold := suite.app.GetCDPKeeper().GetParams(suite.ctx).DebtAuctionThreshold
	newDebtThreshold := previousCDPDebtThreshold.Add(i(1000000))
	msg := types.NewMsgSubmitProposal(
		params.NewParameterChangeProposal(
			"A Title",
			"A description of this proposal.",
			[]params.ParamChange{{
				Subspace: cdptypes.ModuleName,
				Key:      string(cdptypes.KeyDebtThreshold),
				Value:    string(types.ModuleCdc.MustMarshalJSON(newDebtThreshold)),
			}},
		),
		suite.addresses[0],
		1,
	)
	res := suite.handler(suite.ctx, msg)
	suite.True(res.IsOK())
	proposalID := types.Uint64FromBytes(res.Data)
	res = suite.handler(suite.ctx, types.NewMsgVote(suite.addresses[0], proposalID))
	suite.True(res.IsOK())

	// Add a vote to make the proposal pass
	res = suite.handler(suite.ctx, types.NewMsgVote(suite.addresses[1], proposalID))

	suite.True(res.IsOK())
	// Check the param has been updated
	suite.Equal(newDebtThreshold, suite.app.GetCDPKeeper().GetParams(suite.ctx).DebtAuctionThreshold)
	// Check proposal and votes are gone
	_, found := suite.keeper.GetProposal(suite.ctx, proposalID)
	suite.False(found)
	suite.keeper.IterateVotes(suite.ctx, proposalID, func(v types.Vote) bool {
		suite.Fail("vote found when there should be none")
		return true
	})
}

func (suite *HandlerTestSuite) TestMsgAddVote_ProposalFail() {
	recipient := suite.addresses[4]
	recipientCoins := suite.app.GetBankKeeper().GetCoins(suite.ctx, recipient)
	msg := types.NewMsgSubmitProposal(
		distribution.NewCommunityPoolSpendProposal(
			"A Title",
			"A description of this proposal.",
			recipient,
			cs(c("ukava", 500)),
		),
		suite.addresses[0],
		1,
	)
	res := suite.handler(suite.ctx, msg)
	suite.True(res.IsOK())
	proposalID := types.Uint64FromBytes(res.Data)
	res = suite.handler(suite.ctx, types.NewMsgVote(suite.addresses[0], proposalID))
	suite.True(res.IsOK())

	// invalidate the proposal by emptying community pool
	suite.app.GetDistrKeeper().DistributeFromFeePool(suite.ctx, suite.communityPoolAmt, suite.addresses[0])

	// Add a vote to make the proposal pass
	res = suite.handler(suite.ctx, types.NewMsgVote(suite.addresses[1], proposalID))

	suite.True(res.IsOK())
	// Check the proposal was not enacted
	suite.Equal(recipientCoins, suite.app.GetBankKeeper().GetCoins(suite.ctx, recipient))
	// Check proposal and votes are gone
	_, found := suite.keeper.GetProposal(suite.ctx, proposalID)
	suite.False(found)
	suite.keeper.IterateVotes(suite.ctx, proposalID, func(v types.Vote) bool {
		suite.Fail("vote found when there should be none")
		return true
	})
}

func TestHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(HandlerTestSuite))
}
