package keeper_test

import (
	"reflect"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov"
	"github.com/cosmos/cosmos-sdk/x/params"

	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/kava-labs/kava/app"
	bep3types "github.com/kava-labs/kava/x/bep3/types"
	cdptypes "github.com/kava-labs/kava/x/cdp/types"
	"github.com/kava-labs/kava/x/committee"
	"github.com/kava-labs/kava/x/committee/types"
	"github.com/kava-labs/kava/x/pricefeed"
)

func newCDPGenesisState(params cdptypes.Params) app.GenesisState {
	genesis := cdptypes.DefaultGenesisState()
	genesis.Params = params
	return app.GenesisState{cdptypes.ModuleName: cdptypes.ModuleCdc.MustMarshalJSON(genesis)}
}

func newBep3GenesisState(params bep3types.Params) app.GenesisState {
	genesis := bep3types.DefaultGenesisState()
	genesis.Params = params
	return app.GenesisState{bep3types.ModuleName: bep3types.ModuleCdc.MustMarshalJSON(genesis)}
}

func newPricefeedGenState(assets []string, prices []sdk.Dec) app.GenesisState {
	if len(assets) != len(prices) {
		panic("assets and prices must be the same length")
	}
	pfGenesis := pricefeed.DefaultGenesisState()

	for i := range assets {
		pfGenesis.Params.Markets = append(
			pfGenesis.Params.Markets,
			pricefeed.Market{
				MarketID: assets[i] + ":usd", BaseAsset: assets[i], QuoteAsset: "usd", Oracles: []sdk.AccAddress{}, Active: true,
			})
		pfGenesis.PostedPrices = append(
			pfGenesis.PostedPrices,
			pricefeed.PostedPrice{
				MarketID:      assets[i] + ":usd",
				OracleAddress: sdk.AccAddress{},
				Price:         prices[i],
				Expiry:        time.Date(1998, 1, 1, 0, 0, 0, 0, time.UTC),
			})
	}
	return app.GenesisState{pricefeed.ModuleName: pricefeed.ModuleCdc.MustMarshalJSON(pfGenesis)}
}

func (suite *KeeperTestSuite) TestSubmitProposal() {
	normalCom := types.Committee{
		ID:               12,
		Description:      "This committee is for testing.",
		Members:          suite.addresses[:2],
		Permissions:      []types.Permission{types.GodPermission{}},
		VoteThreshold:    d("0.667"),
		ProposalDuration: time.Hour * 24 * 7,
	}

	noPermissionsCom := normalCom
	noPermissionsCom.Permissions = []types.Permission{}

	paramChangePermissionsCom := normalCom
	paramChangePermissionsCom.Permissions = []types.Permission{
		types.SubParamChangePermission{
			AllowedParams: types.AllowedParams{
				{Subspace: cdptypes.ModuleName, Key: string(cdptypes.KeyDebtThreshold)},
				{Subspace: cdptypes.ModuleName, Key: string(cdptypes.KeyCollateralParams)},
			},
			AllowedCollateralParams: types.AllowedCollateralParams{
				types.AllowedCollateralParam{
					Denom:        "bnb",
					DebtLimit:    true,
					StabilityFee: true,
				},
			},
		},
	}

	testCP := cdptypes.CollateralParams{{
		Denom:               "bnb",
		LiquidationRatio:    d("1.5"),
		DebtLimit:           c("usdx", 1000000000000),
		StabilityFee:        d("1.000000001547125958"), // %5 apr
		LiquidationPenalty:  d("0.05"),
		AuctionSize:         i(100),
		Prefix:              0x20,
		ConversionFactor:    i(6),
		LiquidationMarketID: "bnb:usd",
		SpotMarketID:        "bnb:usd",
	}}
	testCDPParams := cdptypes.DefaultParams()
	testCDPParams.CollateralParams = testCP
	testCDPParams.GlobalDebtLimit = testCP[0].DebtLimit

	newValidCP := make(cdptypes.CollateralParams, len(testCP))
	copy(newValidCP, testCP)
	newValidCP[0].DebtLimit = c("usdx", 500000000000)

	newInvalidCP := make(cdptypes.CollateralParams, len(testCP))
	copy(newInvalidCP, testCP)
	newInvalidCP[0].SpotMarketID = "btc:usd"

	testcases := []struct {
		name        string
		committee   types.Committee
		pubProposal types.PubProposal
		proposer    sdk.AccAddress
		committeeID uint64
		expectErr   bool
	}{
		{
			name:        "normal text proposal",
			committee:   normalCom,
			pubProposal: gov.NewTextProposal("A Title", "A description of this proposal."),
			proposer:    normalCom.Members[0],
			committeeID: normalCom.ID,
			expectErr:   false,
		},
		{
			name:      "normal param change proposal",
			committee: normalCom,
			pubProposal: params.NewParameterChangeProposal(
				"A Title", "A description of this proposal.",
				[]params.ParamChange{
					{
						Subspace: "cdp", Key: string(cdptypes.KeyDebtThreshold), Value: string(suite.app.Codec().MustMarshalJSON(i(1000000))),
					},
				},
			),
			proposer:    normalCom.Members[0],
			committeeID: normalCom.ID,
			expectErr:   false,
		},
		{
			name:        "invalid proposal",
			committee:   normalCom,
			pubProposal: nil,
			proposer:    normalCom.Members[0],
			committeeID: normalCom.ID,
			expectErr:   true,
		},
		{
			name: "missing committee",
			// no committee
			pubProposal: gov.NewTextProposal("A Title", "A description of this proposal."),
			proposer:    suite.addresses[0],
			committeeID: 0,
			expectErr:   true,
		},
		{
			name:        "not a member",
			committee:   normalCom,
			pubProposal: gov.NewTextProposal("A Title", "A description of this proposal."),
			proposer:    suite.addresses[4],
			committeeID: normalCom.ID,
			expectErr:   true,
		},
		{
			name:        "not enough permissions",
			committee:   noPermissionsCom,
			pubProposal: gov.NewTextProposal("A Title", "A description of this proposal."),
			proposer:    noPermissionsCom.Members[0],
			committeeID: noPermissionsCom.ID,
			expectErr:   true,
		},
		{
			name:      "valid sub param change",
			committee: paramChangePermissionsCom,
			pubProposal: params.NewParameterChangeProposal(
				"A Title", "A description of this proposal.",
				[]params.ParamChange{
					{
						"cdp", string(cdptypes.KeyDebtThreshold), string(suite.app.Codec().MustMarshalJSON(i(1000000000))),
					},
					{
						"cdp", string(cdptypes.KeyCollateralParams), string(suite.app.Codec().MustMarshalJSON(newValidCP)),
					},
				},
			),
			proposer:    paramChangePermissionsCom.Members[0],
			committeeID: paramChangePermissionsCom.ID,
			expectErr:   false,
		},
		{
			name:      "invalid sub param change permission",
			committee: paramChangePermissionsCom,
			pubProposal: params.NewParameterChangeProposal(
				"A Title", "A description of this proposal.",
				[]params.ParamChange{
					{
						"cdp", string(cdptypes.KeyDebtThreshold), string(suite.app.Codec().MustMarshalJSON(i(1000000000))),
					},
					{
						"cdp", string(cdptypes.KeyCollateralParams), string(suite.app.Codec().MustMarshalJSON(newInvalidCP)),
					},
				},
			),
			proposer:    paramChangePermissionsCom.Members[0],
			committeeID: paramChangePermissionsCom.ID,
			expectErr:   true,
		},
	}

	for _, tc := range testcases {
		suite.Run(tc.name, func() {
			// Create local testApp because suite doesn't run the SetupTest function for subtests
			tApp := app.NewTestApp()
			keeper := tApp.GetCommitteeKeeper()
			ctx := tApp.NewContext(true, abci.Header{})
			tApp.InitializeFromGenesisStates(
				newPricefeedGenState([]string{"bnb"}, []sdk.Dec{d("15.01")}),
				newCDPGenesisState(testCDPParams),
			)
			// setup committee (if required)
			if !(reflect.DeepEqual(tc.committee, types.Committee{})) {
				keeper.SetCommittee(ctx, tc.committee)
			}

			id, err := keeper.SubmitProposal(ctx, tc.proposer, tc.committeeID, tc.pubProposal)

			if tc.expectErr {
				suite.NotNil(err)
			} else {
				suite.NoError(err)
				pr, found := keeper.GetProposal(ctx, id)
				suite.True(found)
				suite.Equal(tc.committeeID, pr.CommitteeID)
				suite.Equal(ctx.BlockTime().Add(tc.committee.ProposalDuration), pr.Deadline)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestAddVote() {
	normalCom := types.Committee{
		ID:          12,
		Members:     suite.addresses[:2],
		Permissions: []types.Permission{types.GodPermission{}},
	}
	firstBlockTime := time.Date(1998, time.January, 1, 1, 0, 0, 0, time.UTC)

	testcases := []struct {
		name       string
		proposalID uint64
		voter      sdk.AccAddress
		voteTime   time.Time
		expectErr  bool
	}{
		{
			name:       "normal",
			proposalID: types.DefaultNextProposalID,
			voter:      normalCom.Members[0],
			expectErr:  false,
		},
		{
			name:       "nonexistent proposal",
			proposalID: 9999999,
			voter:      normalCom.Members[0],
			expectErr:  true,
		},
		{
			name:       "voter not committee member",
			proposalID: types.DefaultNextProposalID,
			voter:      suite.addresses[4],
			expectErr:  true,
		},
		{
			name:       "proposal expired",
			proposalID: types.DefaultNextProposalID,
			voter:      normalCom.Members[0],
			voteTime:   firstBlockTime.Add(normalCom.ProposalDuration),
			expectErr:  true,
		},
	}

	for _, tc := range testcases {
		suite.Run(tc.name, func() {
			// Create local testApp because suite doesn't run the SetupTest function for subtests
			tApp := app.NewTestApp()
			keeper := tApp.GetCommitteeKeeper()
			ctx := tApp.NewContext(true, abci.Header{Height: 1, Time: firstBlockTime})
			tApp.InitializeFromGenesisStates()

			// setup the committee and proposal
			keeper.SetCommittee(ctx, normalCom)
			_, err := keeper.SubmitProposal(ctx, normalCom.Members[0], normalCom.ID, gov.NewTextProposal("A Title", "A description of this proposal."))
			suite.NoError(err)

			ctx = ctx.WithBlockTime(tc.voteTime)
			err = keeper.AddVote(ctx, tc.proposalID, tc.voter)

			if tc.expectErr {
				suite.NotNil(err)
			} else {
				suite.NoError(err)
				_, found := keeper.GetVote(ctx, tc.proposalID, tc.voter)
				suite.True(found)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestGetProposalResult() {
	normalCom := types.Committee{
		ID:               12,
		Description:      "This committee is for testing.",
		Members:          suite.addresses[:5],
		Permissions:      []types.Permission{types.GodPermission{}},
		VoteThreshold:    d("0.667"),
		ProposalDuration: time.Hour * 24 * 7,
	}
	var defaultID uint64 = 1
	firstBlockTime := time.Date(1998, time.January, 1, 1, 0, 0, 0, time.UTC)

	testcases := []struct {
		name           string
		committee      types.Committee
		votes          []types.Vote
		proposalPasses bool
		expectErr      bool
	}{
		{
			name:      "enough votes",
			committee: normalCom,
			votes: []types.Vote{
				{ProposalID: defaultID, Voter: suite.addresses[0]},
				{ProposalID: defaultID, Voter: suite.addresses[1]},
				{ProposalID: defaultID, Voter: suite.addresses[2]},
				{ProposalID: defaultID, Voter: suite.addresses[3]},
			},
			proposalPasses: true,
			expectErr:      false,
		},
		{
			name:      "not enough votes",
			committee: normalCom,
			votes: []types.Vote{
				{ProposalID: defaultID, Voter: suite.addresses[0]},
			},
			proposalPasses: false,
			expectErr:      false,
		},
	}

	for _, tc := range testcases {
		suite.Run(tc.name, func() {
			// Create local testApp because suite doesn't run the SetupTest function for subtests
			tApp := app.NewTestApp()
			keeper := tApp.GetCommitteeKeeper()
			ctx := tApp.NewContext(true, abci.Header{Height: 1, Time: firstBlockTime})

			tApp.InitializeFromGenesisStates(
				committeeGenState(
					tApp.Codec(),
					[]types.Committee{tc.committee},
					[]types.Proposal{{
						PubProposal: gov.NewTextProposal("A Title", "A description of this proposal."),
						ID:          defaultID,
						CommitteeID: tc.committee.ID,
						Deadline:    firstBlockTime.Add(time.Hour * 24 * 7),
					}},
					tc.votes,
				),
			)

			proposalPasses, err := keeper.GetProposalResult(ctx, defaultID)

			if tc.expectErr {
				suite.NotNil(err)
			} else {
				suite.NoError(err)
				suite.Equal(tc.proposalPasses, proposalPasses)
			}
		})
	}
}

func committeeGenState(cdc *codec.Codec, committees []types.Committee, proposals []types.Proposal, votes []types.Vote) app.GenesisState {
	gs := types.NewGenesisState(
		uint64(len(proposals)+1),
		committees,
		proposals,
		votes,
	)
	return app.GenesisState{committee.ModuleName: cdc.MustMarshalJSON(gs)}
}

type UnregisteredPubProposal struct {
	gov.TextProposal
}

func (UnregisteredPubProposal) ProposalRoute() string { return "unregistered" }
func (UnregisteredPubProposal) ProposalType() string  { return "unregistered" }

var _ types.PubProposal = UnregisteredPubProposal{}

func (suite *KeeperTestSuite) TestValidatePubProposal() {

	testcases := []struct {
		name        string
		pubProposal types.PubProposal
		expectErr   bool
	}{
		{
			name:        "valid (text proposal)",
			pubProposal: gov.NewTextProposal("A Title", "A description of this proposal."),
			expectErr:   false,
		},
		{
			name: "valid (param change proposal)",
			pubProposal: params.NewParameterChangeProposal(
				"Change the debt limit",
				"This proposal changes the debt limit of the cdp module.",
				[]params.ParamChange{{
					Subspace: cdptypes.ModuleName,
					Key:      string(cdptypes.KeyGlobalDebtLimit),
					Value:    string(types.ModuleCdc.MustMarshalJSON(c("usdx", 100000000000))),
				}},
			),
			expectErr: false,
		},
		{
			name:        "invalid (missing title)",
			pubProposal: gov.TextProposal{Description: "A description of this proposal."},
			expectErr:   true,
		},
		{
			name:        "invalid (unregistered)",
			pubProposal: UnregisteredPubProposal{gov.TextProposal{Title: "A Title", Description: "A description of this proposal."}},
			expectErr:   true,
		},
		{
			name:        "invalid (nil)",
			pubProposal: nil,
			expectErr:   true,
		},
		{
			name: "invalid (proposal handler fails)",
			pubProposal: params.NewParameterChangeProposal(
				"A Title",
				"A description of this proposal.",
				[]params.ParamChange{{
					Subspace: "nonsense-subspace",
					Key:      "nonsense-key",
					Value:    "nonsense-value",
				}},
			),
			expectErr: true,
		},
		{
			name: "invalid (proposal handler panics)",
			pubProposal: params.NewParameterChangeProposal(
				"A Title",
				"A description of this proposal.",
				[]params.ParamChange{{
					Subspace: cdptypes.ModuleName,
					Key:      "nonsense-key", // a valid Subspace but invalid Key will trigger a panic in the paramchange propsal handler
					Value:    "nonsense-value",
				}},
			),
			expectErr: true,
		},
		{
			name: "invalid (proposal handler fails - invalid json)",
			pubProposal: params.NewParameterChangeProposal(
				"A Title",
				"A description of this proposal.",
				[]params.ParamChange{{
					Subspace: cdptypes.ModuleName,
					Key:      string(cdptypes.KeyGlobalDebtLimit),
					Value:    `{"denom": "usdx",`,
				}},
			),
			expectErr: true,
		},
	}

	for _, tc := range testcases {
		suite.Run(tc.name, func() {
			err := suite.keeper.ValidatePubProposal(suite.ctx, tc.pubProposal)
			if tc.expectErr {
				suite.NotNil(err)
			} else {
				suite.NoError(err)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestCloseExpiredProposals() {

	// Setup test state
	firstBlockTime := time.Date(1998, time.January, 1, 1, 0, 0, 0, time.UTC)
	testGenesis := types.NewGenesisState(
		3,
		[]types.Committee{
			{
				ID:               1,
				Description:      "This committee is for testing.",
				Members:          suite.addresses[:3],
				Permissions:      []types.Permission{types.GodPermission{}},
				VoteThreshold:    d("0.667"),
				ProposalDuration: time.Hour * 24 * 7,
			},
			{
				ID:               2,
				Members:          suite.addresses[2:],
				Permissions:      nil,
				VoteThreshold:    d("0.667"),
				ProposalDuration: time.Hour * 24 * 7,
			},
		},
		[]types.Proposal{
			{
				ID:          1,
				CommitteeID: 1,
				PubProposal: gov.NewTextProposal("A Title", "A description of this proposal."),
				Deadline:    firstBlockTime.Add(7 * 24 * time.Hour),
			},
			{
				ID:          2,
				CommitteeID: 1,
				PubProposal: gov.NewTextProposal("Another Title", "A description of this other proposal."),
				Deadline:    firstBlockTime.Add(21 * 24 * time.Hour),
			},
		},
		[]types.Vote{
			{ProposalID: 1, Voter: suite.addresses[0]},
			{ProposalID: 1, Voter: suite.addresses[1]},
			{ProposalID: 2, Voter: suite.addresses[2]},
		},
	)
	suite.app.InitializeFromGenesisStates(
		NewCommitteeGenesisState(suite.app.Codec(), testGenesis),
	)

	// close proposals
	ctx := suite.app.NewContext(true, abci.Header{Height: 1, Time: firstBlockTime})
	suite.keeper.CloseExpiredProposals(ctx)

	// check
	for _, p := range testGenesis.Proposals {
		_, found := suite.keeper.GetProposal(ctx, p.ID)
		votes := getProposalVoteMap(suite.keeper, ctx)

		if ctx.BlockTime().After(p.Deadline) {
			suite.False(found)
			suite.Empty(votes[p.ID])
		} else {
			suite.True(found)
			suite.NotEmpty(votes[p.ID])
		}
	}

	// close (later time)
	ctx = suite.app.NewContext(true, abci.Header{Height: 1, Time: firstBlockTime.Add(7 * 24 * time.Hour)})
	suite.keeper.CloseExpiredProposals(ctx)

	// check
	for _, p := range testGenesis.Proposals {
		_, found := suite.keeper.GetProposal(ctx, p.ID)
		votes := getProposalVoteMap(suite.keeper, ctx)

		if ctx.BlockTime().Equal(p.Deadline) || ctx.BlockTime().After(p.Deadline) {
			suite.False(found)
			suite.Empty(votes[p.ID])
		} else {
			suite.True(found)
			suite.NotEmpty(votes[p.ID])
		}
	}
}
