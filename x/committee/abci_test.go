package committee_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov"
	"github.com/cosmos/cosmos-sdk/x/params"
	"github.com/cosmos/cosmos-sdk/x/upgrade"

	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/cdp"
	cdptypes "github.com/kava-labs/kava/x/cdp/types"
	"github.com/kava-labs/kava/x/committee"
)

type ModuleTestSuite struct {
	suite.Suite

	keeper committee.Keeper
	app    app.TestApp
	ctx    sdk.Context

	addresses []sdk.AccAddress
}

func (suite *ModuleTestSuite) SetupTest() {
	suite.app = app.NewTestApp()
	suite.keeper = suite.app.GetCommitteeKeeper()
	suite.ctx = suite.app.NewContext(true, abci.Header{})
	_, suite.addresses = app.GeneratePrivKeyAddressPairs(5)
}

func (suite *ModuleTestSuite) TestBeginBlock_ClosesExpired() {
	suite.app.InitializeFromGenesisStates()

	normalCom := committee.Committee{
		ID:               12,
		Members:          suite.addresses[:2],
		Permissions:      []committee.Permission{committee.GodPermission{}},
		VoteThreshold:    d("0.8"),
		ProposalDuration: time.Hour * 24 * 7,
	}
	suite.keeper.SetCommittee(suite.ctx, normalCom)

	pprop1 := gov.NewTextProposal("Title 1", "A description of this proposal.")
	id1, err := suite.keeper.SubmitProposal(suite.ctx, normalCom.Members[0], normalCom.ID, pprop1)
	suite.NoError(err)

	oneHrLaterCtx := suite.ctx.WithBlockTime(suite.ctx.BlockTime().Add(time.Hour))
	pprop2 := gov.NewTextProposal("Title 2", "A description of this proposal.")
	id2, err := suite.keeper.SubmitProposal(oneHrLaterCtx, normalCom.Members[0], normalCom.ID, pprop2)
	suite.NoError(err)

	// Run BeginBlocker
	proposalDurationLaterCtx := suite.ctx.WithBlockTime(suite.ctx.BlockTime().Add(normalCom.ProposalDuration))
	suite.NotPanics(func() {
		committee.BeginBlocker(proposalDurationLaterCtx, abci.RequestBeginBlock{}, suite.keeper)
	})

	// Check expired proposals are gone
	_, found := suite.keeper.GetProposal(suite.ctx, id1)
	suite.False(found, "expected expired proposal to be closed")
	_, found = suite.keeper.GetProposal(suite.ctx, id2)
	suite.True(found, "expected non expired proposal to be not closed")
}

func (suite *ModuleTestSuite) TestBeginBlock_EnactsPassed() {
	suite.app.InitializeFromGenesisStates()

	// setup committee
	normalCom := committee.Committee{
		ID:               12,
		Members:          suite.addresses[:2],
		Permissions:      []committee.Permission{committee.GodPermission{}},
		VoteThreshold:    d("0.8"),
		ProposalDuration: time.Hour * 24 * 7,
	}
	suite.keeper.SetCommittee(suite.ctx, normalCom)

	// setup 2 proposals
	previousCDPDebtThreshold := suite.app.GetCDPKeeper().GetParams(suite.ctx).DebtAuctionThreshold
	newDebtThreshold := previousCDPDebtThreshold.Add(i(1000000))
	evenNewerDebtThreshold := newDebtThreshold.Add(i(1000000))

	pprop1 := params.NewParameterChangeProposal("Title 1", "A description of this proposal.",
		[]params.ParamChange{{
			Subspace: cdptypes.ModuleName,
			Key:      string(cdp.KeyDebtThreshold),
			Value:    string(cdp.ModuleCdc.MustMarshalJSON(newDebtThreshold)),
		}},
	)
	id1, err := suite.keeper.SubmitProposal(suite.ctx, normalCom.Members[0], normalCom.ID, pprop1)
	suite.NoError(err)

	pprop2 := params.NewParameterChangeProposal("Title 2", "A description of this proposal.",
		[]params.ParamChange{{
			Subspace: cdptypes.ModuleName,
			Key:      string(cdp.KeyDebtThreshold),
			Value:    string(cdp.ModuleCdc.MustMarshalJSON(evenNewerDebtThreshold)),
		}},
	)
	id2, err := suite.keeper.SubmitProposal(suite.ctx, normalCom.Members[0], normalCom.ID, pprop2)
	suite.NoError(err)

	// add enough votes to make the first proposal pass, but not the second
	suite.NoError(suite.keeper.AddVote(suite.ctx, id1, suite.addresses[0]))
	suite.NoError(suite.keeper.AddVote(suite.ctx, id1, suite.addresses[1]))
	suite.NoError(suite.keeper.AddVote(suite.ctx, id2, suite.addresses[0]))

	// Run BeginBlocker
	suite.NotPanics(func() {
		committee.BeginBlocker(suite.ctx, abci.RequestBeginBlock{}, suite.keeper)
	})

	// Check the param has been updated
	suite.Equal(newDebtThreshold, suite.app.GetCDPKeeper().GetParams(suite.ctx).DebtAuctionThreshold)
	// Check the passed proposal has gone
	_, found := suite.keeper.GetProposal(suite.ctx, id1)
	suite.False(found, "expected passed proposal to be enacted and closed")
	_, found = suite.keeper.GetProposal(suite.ctx, id2)
	suite.True(found, "expected non passed proposal to be not closed")
}

func (suite *ModuleTestSuite) TestBeginBlock_DoesntEnactFailed() {
	suite.app.InitializeFromGenesisStates()

	// setup committee
	normalCom := committee.Committee{
		ID:               12,
		Members:          suite.addresses[:1],
		Permissions:      []committee.Permission{committee.SoftwareUpgradePermission{}},
		VoteThreshold:    d("1.0"),
		ProposalDuration: time.Hour * 24 * 7,
	}
	firstBlockTime := time.Date(1998, 1, 1, 0, 0, 0, 0, time.UTC)
	ctx := suite.ctx.WithBlockTime(firstBlockTime)
	suite.keeper.SetCommittee(ctx, normalCom)

	// setup an upgrade proposal
	pprop1 := upgrade.NewSoftwareUpgradeProposal("Title 1", "A description of this proposal.",
		upgrade.Plan{
			Name: "upgrade-version-v0.23.1",
			Time: firstBlockTime.Add(time.Second * 5),
			Info: "some information about the upgrade",
		},
	)
	id1, err := suite.keeper.SubmitProposal(ctx, normalCom.Members[0], normalCom.ID, pprop1)
	suite.NoError(err)

	// add enough votes to make the proposal pass
	suite.NoError(suite.keeper.AddVote(ctx, id1, suite.addresses[0]))

	// Run BeginBlocker 10 seconds later (5 seconds after upgrade expires)
	tenSecLaterCtx := ctx.WithBlockTime(ctx.BlockTime().Add(time.Second * 10))
	suite.NotPanics(func() {
		suite.app.BeginBlocker(tenSecLaterCtx, abci.RequestBeginBlock{})
	})

	// Check the plan has not been stored
	_, found := suite.app.GetUpgradeKeeper().GetUpgradePlan(tenSecLaterCtx)
	suite.False(found)
	// Check the passed proposal has gone
	_, found = suite.keeper.GetProposal(tenSecLaterCtx, id1)
	suite.False(found, "expected failed proposal to be not enacted and closed")

	// Check the chain doesn't halt
	oneMinLaterCtx := ctx.WithBlockTime(ctx.BlockTime().Add(time.Minute).Add(time.Second))
	suite.NotPanics(func() {
		suite.app.BeginBlocker(oneMinLaterCtx, abci.RequestBeginBlock{})
	})
}

func (suite *ModuleTestSuite) TestBeginBlock_EnactsPassedUpgrade() {
	suite.app.InitializeFromGenesisStates()

	// setup committee
	normalCom := committee.Committee{
		ID:               12,
		Members:          suite.addresses[:1],
		Permissions:      []committee.Permission{committee.SoftwareUpgradePermission{}},
		VoteThreshold:    d("1.0"),
		ProposalDuration: time.Hour * 24 * 7,
	}
	firstBlockTime := time.Date(1998, 1, 1, 0, 0, 0, 0, time.UTC)
	ctx := suite.ctx.WithBlockTime(firstBlockTime)
	suite.keeper.SetCommittee(ctx, normalCom)

	// setup an upgrade proposal
	pprop1 := upgrade.NewSoftwareUpgradeProposal("Title 1", "A description of this proposal.",
		upgrade.Plan{
			Name: "upgrade-version-v0.23.1",
			Time: firstBlockTime.Add(time.Minute * 1),
			Info: "some information about the upgrade",
		},
	)
	id1, err := suite.keeper.SubmitProposal(ctx, normalCom.Members[0], normalCom.ID, pprop1)
	suite.NoError(err)

	// add enough votes to make the proposal pass
	suite.NoError(suite.keeper.AddVote(ctx, id1, suite.addresses[0]))

	// Run BeginBlocker
	fiveSecLaterCtx := ctx.WithBlockTime(ctx.BlockTime().Add(time.Second * 5))
	suite.NotPanics(func() {
		suite.app.BeginBlocker(fiveSecLaterCtx, abci.RequestBeginBlock{})
	})

	// Check the plan has been stored
	_, found := suite.app.GetUpgradeKeeper().GetUpgradePlan(fiveSecLaterCtx)
	suite.True(found)
	// Check the passed proposal has gone
	_, found = suite.keeper.GetProposal(fiveSecLaterCtx, id1)
	suite.False(found, "expected passed proposal to be enacted and closed")

	// Check the chain halts
	oneMinLaterCtx := ctx.WithBlockTime(ctx.BlockTime().Add(time.Minute))
	suite.Panics(func() {
		suite.app.BeginBlocker(oneMinLaterCtx, abci.RequestBeginBlock{})
	})
}

func TestModuleTestSuite(t *testing.T) {
	suite.Run(t, new(ModuleTestSuite))
}
