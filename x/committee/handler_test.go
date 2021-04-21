package committee_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/distribution"
	"github.com/cosmos/cosmos-sdk/x/params"
	"github.com/cosmos/cosmos-sdk/x/upgrade"

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
	gs.FeePool = distribution.FeePool{CommunityPool: sdk.NewDecCoinsFromCoins(communityPoolCoins...)}
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
			types.MemberCommittee{
				BaseCommittee: types.BaseCommittee{
					ID:               1,
					Description:      "This committee is for testing.",
					Members:          suite.addresses[:3],
					Permissions:      []types.Permission{types.GodPermission{}},
					VoteThreshold:    d("0.5"),
					ProposalDuration: time.Hour * 24 * 7,
				},
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

	res, err := suite.handler(suite.ctx, msg)

	suite.NoError(err)
	_, found := suite.keeper.GetProposal(suite.ctx, types.Uint64FromBytes(res.Data))
	suite.True(found)
}

func (suite *HandlerTestSuite) TestSubmitProposalMsg_Invalid() {
	var committeeID uint64 = 1
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
		committeeID,
	)

	_, err := suite.handler(suite.ctx, msg)

	suite.Error(err)
	suite.Empty(
		suite.keeper.GetProposalsByCommittee(suite.ctx, committeeID),
		"proposal found when none should exist",
	)

}

func (suite *HandlerTestSuite) TestSubmitProposalMsg_ValidUpgrade() {
	msg := committee.NewMsgSubmitProposal(
		upgrade.NewSoftwareUpgradeProposal(
			"A Title",
			"A description of this proposal.",
			upgrade.Plan{
				Name: "emergency-shutdown-1",                      // identifier for the upgrade
				Time: suite.ctx.BlockTime().Add(time.Minute * 10), // time after which to implement plan
				Info: "Some information about the shutdown.",
			},
		),
		suite.addresses[0],
		1,
	)

	res, err := suite.handler(suite.ctx, msg)

	suite.NoError(err)
	_, found := suite.keeper.GetProposal(suite.ctx, types.Uint64FromBytes(res.Data))
	suite.True(found)
}

func (suite *HandlerTestSuite) TestSubmitProposalMsg_Unregistered() {
	var committeeID uint64 = 1
	msg := types.NewMsgSubmitProposal(
		UnregisteredPubProposal{},
		suite.addresses[0],
		committeeID,
	)

	_, err := suite.handler(suite.ctx, msg)

	suite.Error(err)
	suite.Empty(
		suite.keeper.GetProposalsByCommittee(suite.ctx, committeeID),
		"proposal found when none should exist",
	)
}

func TestHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(HandlerTestSuite))
}
