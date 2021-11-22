package committee_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	abci "github.com/tendermint/tendermint/abci/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/kava-labs/kava/app"
	// "github.com/kava-labs/kava/x/cdp"
	// cdptypes "github.com/kava-labs/kava/x/cdp/types"
	"github.com/kava-labs/kava/x/committee"
	"github.com/kava-labs/kava/x/committee/keeper"
	"github.com/kava-labs/kava/x/committee/testutil"
	"github.com/kava-labs/kava/x/committee/types"
)

type ModuleTestSuite struct {
	suite.Suite

	keeper keeper.Keeper
	app    app.TestApp
	ctx    sdk.Context

	addresses []sdk.AccAddress
}

func (suite *ModuleTestSuite) SetupTest() {
	suite.app = app.NewTestApp()
	suite.keeper = suite.app.GetCommitteeKeeper()
	suite.ctx = suite.app.NewContext(true, tmproto.Header{})
	_, suite.addresses = app.GeneratePrivKeyAddressPairs(5)
}

func (suite *ModuleTestSuite) TestBeginBlock_ClosesExpired() {
	suite.app.InitializeFromGenesisStates()

	memberCom := types.MustNewMemberCommittee(
		12,
		"This committee is for testing.",
		suite.addresses[:2],
		[]types.Permission{&types.GodPermission{}},
		testutil.D("0.8"),
		time.Hour*24*7,
		types.TALLY_OPTION_DEADLINE,
	)
	suite.keeper.SetCommittee(suite.ctx, memberCom)

	pprop1 := govtypes.NewTextProposal("Title 1", "A description of this proposal.")
	id1, err := suite.keeper.SubmitProposal(suite.ctx, memberCom.Members[0], memberCom.ID, pprop1)
	suite.NoError(err)

	oneHrLaterCtx := suite.ctx.WithBlockTime(suite.ctx.BlockTime().Add(time.Hour))
	pprop2 := govtypes.NewTextProposal("Title 2", "A description of this proposal.")
	id2, err := suite.keeper.SubmitProposal(oneHrLaterCtx, memberCom.Members[0], memberCom.ID, pprop2)
	suite.NoError(err)

	// Run BeginBlocker
	proposalDurationLaterCtx := suite.ctx.WithBlockTime(suite.ctx.BlockTime().Add(memberCom.ProposalDuration))
	suite.NotPanics(func() {
		committee.BeginBlocker(proposalDurationLaterCtx, abci.RequestBeginBlock{}, suite.keeper)
	})

	// Check expired proposals are gone
	_, found := suite.keeper.GetProposal(suite.ctx, id1)
	suite.False(found, "expected expired proposal to be closed")
	_, found = suite.keeper.GetProposal(suite.ctx, id2)
	suite.True(found, "expected non expired proposal to be not closed")
}

// func (suite *ModuleTestSuite) TestBeginBlock_EnactsPassed() {
// 	suite.app.InitializeFromGenesisStates()

// 	// setup committee
// 	normalCom := types.MustNewMemberCommittee(12, "committee description", suite.addresses[:2],
// 		[]types.Permission{&types.GodPermission{}}, testutil.D("0.8"), time.Hour*24*7, types.TALLY_OPTION_FIRST_PAST_THE_POST)

// 	suite.keeper.SetCommittee(suite.ctx, normalCom)

// 	// setup 2 proposals
// 	previousCDPDebtThreshold := suite.app.GetCDPKeeper().GetParams(suite.ctx).DebtAuctionThreshold
// 	newDebtThreshold := previousCDPDebtThreshold.Add(i(1000000))
// 	evenNewerDebtThreshold := newDebtThreshold.Add(i(1000000))

// 	pprop1 := params.NewParameterChangeProposal("Title 1", "A description of this proposal.",
// 		[]params.ParamChange{{
// 			Subspace: cdptypes.ModuleName,
// 			Key:      string(cdp.KeyDebtThreshold),
// 			Value:    string(cdp.ModuleCdc.MustMarshalJSON(newDebtThreshold)),
// 		}},
// 	)
// 	id1, err := suite.keeper.SubmitProposal(suite.ctx, normalCom.Members[0], normalCom.ID, pprop1)
// 	suite.NoError(err)

// 	pprop2 := params.NewParameterChangeProposal("Title 2", "A description of this proposal.",
// 		[]params.ParamChange{{
// 			Subspace: cdptypes.ModuleName,
// 			Key:      string(cdp.KeyDebtThreshold),
// 			Value:    string(cdp.ModuleCdc.MustMarshalJSON(evenNewerDebtThreshold)),
// 		}},
// 	)
// 	id2, err := suite.keeper.SubmitProposal(suite.ctx, normalCom.Members[0], normalCom.ID, pprop2)
// 	suite.NoError(err)

// 	// add enough votes to make the first proposal pass, but not the second
// 	suite.NoError(suite.keeper.AddVote(suite.ctx, id1, suite.addresses[0], types.Yes))
// 	suite.NoError(suite.keeper.AddVote(suite.ctx, id1, suite.addresses[1], types.Yes))
// 	suite.NoError(suite.keeper.AddVote(suite.ctx, id2, suite.addresses[0], types.Yes))

// 	// Run BeginBlocker
// 	suite.NotPanics(func() {
// 		committee.BeginBlocker(suite.ctx, abci.RequestBeginBlock{}, suite.keeper)
// 	})

// 	// Check the param has been updated
// 	suite.Equal(newDebtThreshold, suite.app.GetCDPKeeper().GetParams(suite.ctx).DebtAuctionThreshold)
// 	// Check the passed proposal has gone
// 	_, found := suite.keeper.GetProposal(suite.ctx, id1)
// 	suite.False(found, "expected passed proposal to be enacted and closed")
// 	_, found = suite.keeper.GetProposal(suite.ctx, id2)
// 	suite.True(found, "expected non passed proposal to be not closed")
// }

// func (suite *ModuleTestSuite) TestBeginBlock_DoesntEnactFailed() {
// 	suite.app.InitializeFromGenesisStates()

// 	// setup committee
// 	memberCom := types.MustNewMemberCommittee(12, "committee description", suite.addresses[:1],
// 		[]types.Permission{types.SoftwareUpgradePermission{}}, testutil.D("1.0"), time.Hour*24*7, types.TALLY_OPTION_FIRST_PAST_THE_POST)

// 	firstBlockTime := time.Date(1998, 1, 1, 0, 0, 0, 0, time.UTC)
// 	ctx := suite.ctx.WithBlockTime(firstBlockTime)
// 	suite.keeper.SetCommittee(ctx, memberCom)

// 	// setup an upgrade proposal
// 	pprop1 := upgradetypes.NewSoftwareUpgradeProposal("Title 1", "A description of this proposal.",
// 		upgradetypes.Plan{
// 			Name: "upgrade-version-v0.23.1",
// 			Time: firstBlockTime.Add(time.Second * 5),
// 			Info: "some information about the upgrade",
// 		},
// 	)
// 	id1, err := suite.keeper.SubmitProposal(ctx, memberCom.Members[0], memberCom.ID, pprop1)
// 	suite.NoError(err)

// 	// add enough votes to make the proposal pass
// 	suite.NoError(suite.keeper.AddVote(ctx, id1, suite.addresses[0], types.Yes))

// 	// Run BeginBlocker 10 seconds later (5 seconds after upgrade expires)
// 	tenSecLaterCtx := ctx.WithBlockTime(ctx.BlockTime().Add(time.Second * 10))
// 	suite.NotPanics(func() {
// 		suite.app.BeginBlocker(tenSecLaterCtx, abci.RequestBeginBlock{})
// 	})

// 	// Check the plan has not been stored
// 	_, found := suite.app.GetUpgradeKeeper().GetUpgradePlan(tenSecLaterCtx)
// 	suite.False(found)
// 	// Check the passed proposal has gone
// 	_, found = suite.keeper.GetProposal(tenSecLaterCtx, id1)
// 	suite.False(found, "expected failed proposal to be not enacted and closed")

// 	// Check the chain doesn't halt
// 	oneMinLaterCtx := ctx.WithBlockTime(ctx.BlockTime().Add(time.Minute).Add(time.Second))
// 	suite.NotPanics(func() {
// 		suite.app.BeginBlocker(oneMinLaterCtx, abci.RequestBeginBlock{})
// 	})
// }

// func (suite *ModuleTestSuite) TestBeginBlock_EnactsPassedUpgrade() {
// 	suite.app.InitializeFromGenesisStates()

// 	// setup committee
// 	memberCom := types.MustNewMemberCommittee(
// 		12,
// 		"committee description",
// 		suite.addresses[:1],
// 		[]types.Permission{types.SoftwareUpgradePermission{}},
// 		testutil.D("1.0"),
// 		time.Hour*24*7,
// 		types.TALLY_OPTION_FIRST_PAST_THE_POST,
// 	)

// 	firstBlockTime := time.Date(1998, 1, 1, 0, 0, 0, 0, time.UTC)
// 	ctx := suite.ctx.WithBlockTime(firstBlockTime)
// 	suite.keeper.SetCommittee(ctx, memberCom)

// 	// setup an upgrade proposal
// 	pprop1 := upgradetypes.NewSoftwareUpgradeProposal("Title 1", "A description of this proposal.",
// 		upgradetypes.Plan{
// 			Name: "upgrade-version-v0.23.1",
// 			Time: firstBlockTime.Add(time.Minute * 1),
// 			Info: "some information about the upgrade",
// 		},
// 	)
// 	id1, err := suite.keeper.SubmitProposal(ctx, memberCom.Members[0], memberCom.ID, pprop1)
// 	suite.NoError(err)

// 	// add enough votes to make the proposal pass
// 	suite.NoError(suite.keeper.AddVote(ctx, id1, suite.addresses[0], types.Yes))

// 	// Run BeginBlocker
// 	fiveSecLaterCtx := ctx.WithBlockTime(ctx.BlockTime().Add(time.Second * 5))
// 	suite.NotPanics(func() {
// 		suite.app.BeginBlocker(fiveSecLaterCtx, abci.RequestBeginBlock{})
// 	})

// 	// Check the plan has been stored
// 	_, found := suite.app.GetUpgradeKeeper().GetUpgradePlan(fiveSecLaterCtx)
// 	suite.True(found)
// 	// Check the passed proposal has gone
// 	_, found = suite.keeper.GetProposal(fiveSecLaterCtx, id1)
// 	suite.False(found, "expected passed proposal to be enacted and closed")

// 	// Check the chain halts
// 	oneMinLaterCtx := ctx.WithBlockTime(ctx.BlockTime().Add(time.Minute))
// 	suite.Panics(func() {
// 		suite.app.BeginBlocker(oneMinLaterCtx, abci.RequestBeginBlock{})
// 	})
// }

func TestModuleTestSuite(t *testing.T) {
	suite.Run(t, new(ModuleTestSuite))
}
