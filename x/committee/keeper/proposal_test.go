package keeper_test

import (
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/kava-labs/kava/app"
	// bep3types "github.com/kava-labs/kava/x/bep3/types"
	// cdptypes "github.com/kava-labs/kava/x/cdp/types"

	"github.com/kava-labs/kava/x/committee/testutil"
	"github.com/kava-labs/kava/x/committee/types"
	// "github.com/kava-labs/kava/x/pricefeed"
)

// func newCDPGenesisState(params cdptypes.Params) app.GenesisState {
// 	genesis := cdptypes.DefaultGenesisState()
// 	genesis.Params = params
// 	return app.GenesisState{cdptypes.ModuleName: cdptypes.ModuleCdc.MustMarshalJSON(genesis)}
// }

// func newBep3GenesisState(params bep3types.Params) app.GenesisState {
// 	genesis := bep3types.DefaultGenesisState()
// 	genesis.Params = params
// 	return app.GenesisState{bep3types.ModuleName: bep3types.ModuleCdc.MustMarshalJSON(genesis)}
// }

// func newPricefeedGenState(assets []string, prices []sdk.Dec) app.GenesisState {
// 	if len(assets) != len(prices) {
// 		panic("assets and prices must be the same length")
// 	}
// 	pfGenesis := pricefeed.DefaultGenesisState()

// 	for i := range assets {
// 		pfGenesis.Params.Markets = append(
// 			pfGenesis.Params.Markets,
// 			pricefeed.Market{
// 				MarketID: assets[i] + ":usd", BaseAsset: assets[i], QuoteAsset: "usd", Oracles: []sdk.AccAddress{}, Active: true,
// 			})
// 		pfGenesis.PostedPrices = append(
// 			pfGenesis.PostedPrices,
// 			pricefeed.PostedPrice{
// 				MarketID:      assets[i] + ":usd",
// 				OracleAddress: sdk.AccAddress{},
// 				Price:         prices[i],
// 				Expiry:        time.Date(1998, 1, 1, 0, 0, 0, 0, time.UTC),
// 			})
// 	}
// 	return app.GenesisState{pricefeed.ModuleName: pricefeed.ModuleCdc.MustMarshalJSON(pfGenesis)}
// }

// func (suite *keeperTestSuite) TestSubmitProposal() {
// 	defaultCommitteeID := uint64(12)
// 	normalCom := types.BaseCommittee{
// 		ID:               defaultCommitteeID,
// 		Description:      "This committee is for testing.",
// 		Members:          suite.Addresses[:2],
// 		Permissions:      []types.Permission{&types.GodPermission{}},
// 		VoteThreshold:    testutil.D("0.667"),
// 		ProposalDuration: time.Hour * 24 * 7,
// 		TallyOption:      types.TALLY_OPTION_FIRST_PAST_THE_POST,
// 	}

// 	noPermissionsCom := normalCom
// 	noPermissionsCom.Permissions = []types.Permission{}

// 	paramChangePermissionsCom := normalCom
// 	paramChangePermissionsCom.Permissions = []types.Permission{
// 		types.paramsproposalbParamChangePermission{
// 			AllowedParams: types.AllowedParams{
// 				{Subspace: cdptypes.ModuleName, Key: string(cdptypes.KeyDebtThreshold)},
// 				{Subspace: cdptypes.ModuleName, Key: string(cdptypes.KeyCollateralParams)},
// 			},
// 			AllowedCollateralParams: types.AllowedCollateralParams{
// 				types.AllowedCollateralParam{
// 					Type:         "bnb-a",
// 					DebtLimit:    true,
// 					StabilityFee: true,
// 				},
// 			},
// 		},
// 	}

// 	testCP := cdptypes.CollateralParams{{
// 		Denom:               "bnb",
// 		Type:                "bnb-a",
// 		LiquidationRatio:    testutil.D("1.5"),
// 		DebtLimit:           testutil.C("usdx", 1000000000000),
// 		StabilityFee:        testutil.D("1.000000001547125958"), // %5 apr
// 		LiquidationPenalty:  testutil.D("0.05"),
// 		AuctionSize:         i(100),
// 		Prefix:              0x20,
// 		ConversionFactor:    i(6),
// 		LiquidationMarketID: "bnb:usd",
// 		SpotMarketID:        "bnb:usd",
// 	}}
// 	testCDPParams := cdptypes.DefaultParams()
// 	testCDPParams.CollateralParams = testCP
// 	testCDPParams.GlobalDebtLimit = testCP[0].DebtLimit

// 	newValidCP := make(cdptypes.CollateralParams, len(testCP))
// 	copy(newValidCP, testCP)
// 	newValidCP[0].DebtLimit = testutil.C("usdx", 500000000000)

// 	newInvalidCP := make(cdptypes.CollateralParams, len(testCP))
// 	copy(newInvalidCP, testCP)
// 	newInvalidCP[0].SpotMarketID = "btc:usd"

// 	testcases := []struct {
// 		name        string
// 		committee   types.BaseCommittee
// 		pubProposal types.PubProposal
// 		proposer    sdk.AccAddress
// 		committeeID uint64
// 		expectErr   bool
// 	}{
// 		{
// 			name:        "normal text proposal",
// 			committee:   normalCom,
// 			pubProposal: govtypes.NewTextProposal("A Title", "A description of this proposal."),
// 			proposer:    normalCom.Members[0],
// 			committeeID: normalCom.ID,
// 			expectErr:   false,
// 		},
// 		{
// 			name:      "normal param change proposal",
// 			committee: normalCom,
// 			pubProposal: paramsproposal.NewParameterChangeProposal(
// 				"A Title", "A description of this proposal.",
// 				[]paramsproposal.ParamChange{
// 					{
// 						Subspace: "cdp", Key: string(cdptypes.KeyDebtThreshold), Value: string(suite.app.Codec().MustMarshalJSON(i(1000000))),
// 					},
// 				},
// 			),
// 			proposer:    normalCom.Members[0],
// 			committeeID: normalCom.ID,
// 			expectErr:   false,
// 		},
// 		{
// 			name:        "invalid proposal",
// 			committee:   normalCom,
// 			pubProposal: nil,
// 			proposer:    normalCom.Members[0],
// 			committeeID: normalCom.ID,
// 			expectErr:   true,
// 		},
// 		{
// 			name: "missing committee",
// 			// no committee
// 			pubProposal: govtypes.NewTextProposal("A Title", "A description of this proposal."),
// 			proposer:    suite.Addresses[0],
// 			committeeID: 0,
// 			expectErr:   true,
// 		},
// 		{
// 			name:        "not a member",
// 			committee:   normalCom,
// 			pubProposal: govtypes.NewTextProposal("A Title", "A description of this proposal."),
// 			proposer:    suite.Addresses[4],
// 			committeeID: normalCom.ID,
// 			expectErr:   true,
// 		},
// 		{
// 			name:        "not enough permissions",
// 			committee:   noPermissionsCom,
// 			pubProposal: govtypes.NewTextProposal("A Title", "A description of this proposal."),
// 			proposer:    noPermissionsCom.Members[0],
// 			committeeID: noPermissionsCom.ID,
// 			expectErr:   true,
// 		},
// 		{
// 			name:      "valid sub param change",
// 			committee: paramChangePermissionsCom,
// 			pubProposal: paramsproposal.NewParameterChangeProposal(
// 				"A Title", "A description of this proposal.",
// 				[]paramsproposal.ParamChange{
// 					{
// 						Subspace: "cdp",
// 						Key:      string(cdptypes.KeyDebtThreshold),
// 						Value:    string(suite.app.Codec().MustMarshalJSON(i(1000000000))),
// 					},
// 					{
// 						Subspace: "cdp",
// 						Key:      string(cdptypes.KeyCollateralParams),
// 						Value:    string(suite.app.Codec().MustMarshalJSON(newValidCP)),
// 					},
// 				},
// 			),
// 			proposer:    paramChangePermissionsCom.Members[0],
// 			committeeID: paramChangePermissionsCom.ID,
// 			expectErr:   false,
// 		},
// 		{
// 			name:      "invalid sub param change permission",
// 			committee: paramChangePermissionsCom,
// 			pubProposal: paramsproposal.NewParameterChangeProposal(
// 				"A Title", "A description of this proposal.",
// 				[]paramsproposal.ParamChange{
// 					{
// 						Subspace: "cdp",
// 						Key:      string(cdptypes.KeyDebtThreshold),
// 						Value:    string(suite.app.Codec().MustMarshalJSON(i(1000000000))),
// 					},
// 					{
// 						Subspace: "cdp",
// 						Key:      string(cdptypes.KeyCollateralParams),
// 						Value:    string(suite.app.Codec().MustMarshalJSON(newInvalidCP)),
// 					},
// 				},
// 			),
// 			proposer:    paramChangePermissionsCom.Members[0],
// 			committeeID: paramChangePermissionsCom.ID,
// 			expectErr:   true,
// 		},
// 	}

// 	for _, tc := range testcases {
// 		suite.Run(tc.name, func() {
// 			// Create local testApp because suite doesn't run the SetupTest function for subtests
// 			tApp := app.NewTestApp()
// 			keeper := tApp.GetCommitteeKeeper()
// 			ctx := tApp.NewContext(true, tmproto.Header{})
// 			tApp.InitializeFromGenesisStates(
// 				newPricefeedGenState([]string{"bnb"}, []sdk.Dec{testutil.D("15.01")}),
// 				newCDPGenesisState(testCDPParams),
// 			)
// 			// Cast BaseCommittee to MemberCommittee (if required) to meet Committee interface requirement
// 			if tc.committee.ID == defaultCommitteeID {
// 				keeper.SetCommittee(ctx, types.MustNewMemberCommittee(
// 			}

// 			id, err := keeper.SubmitProposal(ctx, tc.proposer, tc.committeeID, tc.pubProposal)

// 			if tc.expectErr {
// 				suite.NotNil(err)
// 			} else {
// 				suite.NoError(err)
// 				pr, found := keeper.GetProposal(ctx, id)
// 				suite.True(found)
// 				suite.Equal(tc.committeeID, pr.CommitteeID)
// 				suite.Equal(ctx.BlockTime().Add(tc.committee.GetProposalDuration()), pr.Deadline)
// 			}
// 		})
// 	}
// }

func (suite *keeperTestSuite) TestAddVote() {
	memberCom := types.MustNewMemberCommittee(
		1,
		"This member committee is for testing.",
		suite.Addresses[:2],
		[]types.Permission{&types.GodPermission{}},
		testutil.D("0.667"),
		time.Hour*24*7,
		types.TALLY_OPTION_FIRST_PAST_THE_POST,
	)
	tokenCom := types.MustNewTokenCommittee(
		12,
		"This token committee is for testing.",
		suite.Addresses[:2],
		[]types.Permission{&types.GodPermission{}},
		testutil.D("0.4"),
		time.Hour*24*7,
		types.TALLY_OPTION_FIRST_PAST_THE_POST,
		sdk.Dec{},
		"hard",
	)
	nonMemberAddr := suite.Addresses[4]
	firstBlockTime := time.Date(1998, time.January, 1, 1, 0, 0, 0, time.UTC)

	testcases := []struct {
		name       string
		proposalID uint64
		committee  types.Committee
		voter      sdk.AccAddress
		voteType   types.VoteType
		voteTime   time.Time
		expectErr  bool
	}{
		{
			name:       "normal MemberCommittee",
			committee:  memberCom,
			proposalID: types.DefaultNextProposalID,
			voter:      memberCom.Members[0],
			voteType:   types.VOTE_TYPE_YES,
			expectErr:  false,
		},
		{
			name:       "normal TokenCommittee",
			committee:  tokenCom,
			proposalID: types.DefaultNextProposalID,
			voter:      nonMemberAddr,
			voteType:   types.VOTE_TYPE_YES,
			expectErr:  false,
		},
		{
			name:       "nonexistent proposal",
			committee:  memberCom,
			proposalID: 9999999,
			voter:      memberCom.Members[0],
			voteType:   types.VOTE_TYPE_YES,
			expectErr:  true,
		},
		{
			name:       "proposal expired",
			committee:  memberCom,
			proposalID: types.DefaultNextProposalID,
			voter:      memberCom.Members[0],
			voteTime:   firstBlockTime.Add(memberCom.ProposalDuration),
			voteType:   types.VOTE_TYPE_YES,
			expectErr:  true,
		},
		{
			name:       "MemberCommittee: voter not committee member",
			committee:  memberCom,
			proposalID: types.DefaultNextProposalID,
			voter:      nonMemberAddr,
			voteType:   types.VOTE_TYPE_YES,
			expectErr:  true,
		},
		{
			name:       "MemberCommittee: voter votes no",
			committee:  memberCom,
			proposalID: types.DefaultNextProposalID,
			voter:      memberCom.Members[0],
			voteType:   types.VOTE_TYPE_NO,
			expectErr:  true,
		},
	}

	for _, tc := range testcases {
		suite.Run(tc.name, func() {
			// Create local testApp because suite doesn't run the SetupTest function for subtests
			tApp := app.NewTestApp()
			keeper := tApp.GetCommitteeKeeper()
			ctx := tApp.NewContext(true, tmproto.Header{Height: 1, Time: firstBlockTime})
			tApp.InitializeFromGenesisStates()

			// setup the committee and proposal
			keeper.SetCommittee(ctx, tc.committee)
			_, err := keeper.SubmitProposal(ctx, tc.committee.GetMembers()[0], tc.committee.GetID(), govtypes.NewTextProposal("A Title", "A description of this proposal."))
			suite.NoError(err)

			ctx = ctx.WithBlockTime(tc.voteTime)
			err = keeper.AddVote(ctx, tc.proposalID, tc.voter, tc.voteType)

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

func (suite *keeperTestSuite) TestTallyMemberCommitteeVotes() {
	memberCom := types.MustNewMemberCommittee(
		12,
		"This committee is for testing.",
		suite.Addresses[:5],
		[]types.Permission{&types.GodPermission{}},
		testutil.D("0.667"),
		time.Hour*24*7,
		types.TALLY_OPTION_DEADLINE,
	)
	var defaultProposalID uint64 = 1
	firstBlockTime := time.Date(1998, time.January, 1, 1, 0, 0, 0, time.UTC)

	testcases := []struct {
		name              string
		votes             []types.Vote
		expectedVoteCount sdk.Dec
	}{
		{
			name:              "has 0 votes",
			votes:             []types.Vote{},
			expectedVoteCount: testutil.D("0"),
		},
		{
			name: "has 1 vote",
			votes: []types.Vote{
				{ProposalID: defaultProposalID, Voter: suite.Addresses[0], VoteType: types.VOTE_TYPE_YES},
			},
			expectedVoteCount: testutil.D("1"),
		},
		{
			name: "has multiple votes",
			votes: []types.Vote{
				{ProposalID: defaultProposalID, Voter: suite.Addresses[0], VoteType: types.VOTE_TYPE_YES},
				{ProposalID: defaultProposalID, Voter: suite.Addresses[1], VoteType: types.VOTE_TYPE_YES},
				{ProposalID: defaultProposalID, Voter: suite.Addresses[2], VoteType: types.VOTE_TYPE_YES},
				{ProposalID: defaultProposalID, Voter: suite.Addresses[3], VoteType: types.VOTE_TYPE_YES},
			},
			expectedVoteCount: testutil.D("4"),
		},
	}

	for _, tc := range testcases {
		// Set up test app
		tApp := app.NewTestApp()
		keeper := tApp.GetCommitteeKeeper()
		ctx := tApp.NewContext(true, tmproto.Header{Height: 1, Time: firstBlockTime})

		// Initialize test app with genesis state
		tApp.InitializeFromGenesisStates(
			committeeGenState(
				tApp.AppCodec(),
				[]types.Committee{memberCom},
				[]types.Proposal{types.MustNewProposal(
					govtypes.NewTextProposal("A Title", "A description of this proposal."),
					defaultProposalID,
					memberCom.GetID(),
					firstBlockTime.Add(time.Hour*24*7),
				)},
				tc.votes,
			),
		)

		// Check that all votes are counted
		currentVotes := keeper.TallyMemberCommitteeVotes(ctx, defaultProposalID)
		suite.Equal(tc.expectedVoteCount, currentVotes)
	}
}

func (suite *keeperTestSuite) TestTallyTokenCommitteeVotes() {
	tokenCom := types.MustNewTokenCommittee(
		12,
		"This committee is for testing.",
		suite.Addresses[:5],
		[]types.Permission{&types.GodPermission{}},
		testutil.D("0.667"),
		time.Hour*24*7,
		types.TALLY_OPTION_DEADLINE,
		testutil.D("0.4"),
		"hard",
	)
	var defaultProposalID uint64 = 1
	firstBlockTime := time.Date(1998, time.January, 1, 1, 0, 0, 0, time.UTC)

	genAddrs := suite.Addresses[:8]                       // Genesis accounts
	genCoinCounts := []int64{0, 0, 0, 10, 20, 30, 40, 50} // Genesis token balances

	testcases := []struct {
		name                   string
		votes                  []types.Vote
		expectedYesVoteCount   sdk.Dec
		expectedNoVoteCount    sdk.Dec
		expectedTotalVoteCount sdk.Dec
	}{
		{
			name:                   "has 0 votes",
			votes:                  []types.Vote{},
			expectedYesVoteCount:   testutil.D("0"),
			expectedNoVoteCount:    testutil.D("0"),
			expectedTotalVoteCount: testutil.D("0"),
		},
		{
			name: "counts token holder 'Yes' votes",
			votes: []types.Vote{
				{ProposalID: defaultProposalID, Voter: genAddrs[4], VoteType: types.VOTE_TYPE_YES}, // Token holder
			},
			expectedYesVoteCount:   sdk.NewDec(genCoinCounts[4]),
			expectedNoVoteCount:    testutil.D("0"),
			expectedTotalVoteCount: sdk.NewDec(genCoinCounts[4]),
		},
		{
			name: "does not count non-token holder 'Yes' votes",
			votes: []types.Vote{
				{ProposalID: defaultProposalID, Voter: genAddrs[4], VoteType: types.VOTE_TYPE_YES}, // Token holder
				{ProposalID: defaultProposalID, Voter: genAddrs[0], VoteType: types.VOTE_TYPE_YES}, // Non-token holder
			},
			expectedYesVoteCount:   sdk.NewDec(genCoinCounts[4]),
			expectedNoVoteCount:    testutil.D("0"),
			expectedTotalVoteCount: sdk.NewDec(genCoinCounts[4]),
		},
		{
			name: "counts multiple 'Yes' votes from token holders",
			votes: []types.Vote{
				{ProposalID: defaultProposalID, Voter: genAddrs[4], VoteType: types.VOTE_TYPE_YES}, // Token holder
				{ProposalID: defaultProposalID, Voter: genAddrs[5], VoteType: types.VOTE_TYPE_YES}, // Token holder
				{ProposalID: defaultProposalID, Voter: genAddrs[6], VoteType: types.VOTE_TYPE_YES}, // Token holder
			},
			expectedYesVoteCount:   sdk.NewDec(genCoinCounts[4] + genCoinCounts[5] + genCoinCounts[6]),
			expectedNoVoteCount:    testutil.D("0"),
			expectedTotalVoteCount: sdk.NewDec(genCoinCounts[4] + genCoinCounts[5] + genCoinCounts[6]),
		},
		{
			name: "counts token holder 'No' votes",
			votes: []types.Vote{
				{ProposalID: defaultProposalID, Voter: genAddrs[4], VoteType: types.VOTE_TYPE_NO}, // Token holder
			},
			expectedYesVoteCount:   testutil.D("0"),
			expectedNoVoteCount:    sdk.NewDec(genCoinCounts[4]),
			expectedTotalVoteCount: sdk.NewDec(genCoinCounts[4]),
		},
		{
			name: "does not count non-token holder 'No' votes",
			votes: []types.Vote{
				{ProposalID: defaultProposalID, Voter: genAddrs[4], VoteType: types.VOTE_TYPE_NO}, // Token holder
				{ProposalID: defaultProposalID, Voter: genAddrs[0], VoteType: types.VOTE_TYPE_NO}, // Non-token holder
			},
			expectedYesVoteCount:   testutil.D("0"),
			expectedNoVoteCount:    sdk.NewDec(genCoinCounts[4]),
			expectedTotalVoteCount: sdk.NewDec(genCoinCounts[4]),
		},
		{
			name: "counts multiple 'No' votes from token holders",
			votes: []types.Vote{
				{ProposalID: defaultProposalID, Voter: genAddrs[4], VoteType: types.VOTE_TYPE_NO}, // Token holder
				{ProposalID: defaultProposalID, Voter: genAddrs[5], VoteType: types.VOTE_TYPE_NO}, // Token holder
				{ProposalID: defaultProposalID, Voter: genAddrs[6], VoteType: types.VOTE_TYPE_NO}, // Token holder
			},
			expectedYesVoteCount:   testutil.D("0"),
			expectedNoVoteCount:    sdk.NewDec(genCoinCounts[4] + genCoinCounts[5] + genCoinCounts[6]),
			expectedTotalVoteCount: sdk.NewDec(genCoinCounts[4] + genCoinCounts[5] + genCoinCounts[6]),
		},
		{
			name: "includes token holder 'Abstain' votes in total vote count",
			votes: []types.Vote{
				{ProposalID: defaultProposalID, Voter: genAddrs[4], VoteType: types.VOTE_TYPE_ABSTAIN}, // Token holder
			},
			expectedYesVoteCount:   testutil.D("0"),
			expectedNoVoteCount:    testutil.D("0"),
			expectedTotalVoteCount: sdk.NewDec(genCoinCounts[4]),
		},
	}

	// Convert accounts/token balances into format expected by genesis generation
	var genCoins []sdk.Coins
	var totalSupply sdk.Coins
	for _, amount := range genCoinCounts {
		userCoin := testutil.C("hard", amount)
		genCoins = append(genCoins, testutil.Cs(userCoin))
		totalSupply = totalSupply.Add(userCoin)
	}

	for _, tc := range testcases {
		// Set up test app
		tApp := app.NewTestApp()
		keeper := tApp.GetCommitteeKeeper()
		ctx := tApp.NewContext(true, tmproto.Header{Height: 1, Time: firstBlockTime})

		// Initialize test app with genesis state
		tApp.InitializeFromGenesisStates(
			committeeGenState(
				tApp.AppCodec(),
				[]types.Committee{tokenCom},
				[]types.Proposal{types.MustNewProposal(
					govtypes.NewTextProposal("A Title", "A description of this proposal."),
					defaultProposalID,
					tokenCom.GetID(),
					firstBlockTime.Add(time.Hour*24*7),
				)},
				tc.votes,
			),
			bankGenState(tApp.AppCodec(), totalSupply),
			app.NewFundedGenStateWithCoins(tApp.AppCodec(), genCoins, genAddrs),
		)

		yesVotes, noVotes, currVotes, possibleVotes := keeper.TallyTokenCommitteeVotes(ctx, defaultProposalID, tokenCom.TallyDenom)

		// Check that all Yes votes are counted according to their weight
		suite.Equal(tc.expectedYesVoteCount, yesVotes)
		// Check that all No votes are counted according to their weight
		suite.Equal(tc.expectedNoVoteCount, noVotes)
		// Check that all non-Yes votes are counted according to their weight
		suite.Equal(tc.expectedTotalVoteCount, currVotes)
		// Check that possible votes equals the number of members on the committee
		suite.Equal(totalSupply.AmountOf(tokenCom.GetTallyDenom()).ToDec(), possibleVotes)
	}
}

func (suite *keeperTestSuite) TestGetMemberCommitteeProposalResult() {
	memberCom := types.MustNewMemberCommittee(

		12,
		"This committee is for testing.",
		suite.Addresses[:5],
		[]types.Permission{&types.GodPermission{}},
		testutil.D("0.667"),
		time.Hour*24*7,
		types.TALLY_OPTION_DEADLINE,
	)
	var defaultID uint64 = 1
	firstBlockTime := time.Date(1998, time.January, 1, 1, 0, 0, 0, time.UTC)

	testcases := []struct {
		name           string
		committee      types.Committee
		votes          []types.Vote
		proposalPasses bool
	}{
		{
			name:      "enough votes",
			committee: memberCom,
			votes: []types.Vote{
				{ProposalID: defaultID, Voter: suite.Addresses[0], VoteType: types.VOTE_TYPE_YES},
				{ProposalID: defaultID, Voter: suite.Addresses[1], VoteType: types.VOTE_TYPE_YES},
				{ProposalID: defaultID, Voter: suite.Addresses[2], VoteType: types.VOTE_TYPE_YES},
				{ProposalID: defaultID, Voter: suite.Addresses[3], VoteType: types.VOTE_TYPE_YES},
			},
			proposalPasses: true,
		},
		{
			name:      "not enough votes",
			committee: memberCom,
			votes: []types.Vote{
				{ProposalID: defaultID, Voter: suite.Addresses[0], VoteType: types.VOTE_TYPE_YES},
			},
			proposalPasses: false,
		},
	}

	for _, tc := range testcases {
		suite.Run(tc.name, func() {
			// Create local testApp because suite doesn't run the SetupTest function for subtests
			tApp := app.NewTestApp()
			keeper := tApp.GetCommitteeKeeper()
			ctx := tApp.NewContext(true, tmproto.Header{Height: 1, Time: firstBlockTime})

			tApp.InitializeFromGenesisStates(
				committeeGenState(
					tApp.AppCodec(),
					[]types.Committee{tc.committee},
					[]types.Proposal{types.MustNewProposal(
						govtypes.NewTextProposal("A Title", "A description of this proposal."),
						defaultID,
						tc.committee.GetID(),
						firstBlockTime.Add(time.Hour*24*7),
					)},
					tc.votes,
				),
			)

			proposalPasses := keeper.GetMemberCommitteeProposalResult(ctx, defaultID, tc.committee)
			suite.Equal(tc.proposalPasses, proposalPasses)
		})
	}
}

func (suite *keeperTestSuite) TestGetTokenCommitteeProposalResult() {
	tokenCom := types.MustNewTokenCommittee(
		12,
		"This committee is for testing.",
		suite.Addresses[:5],
		[]types.Permission{&types.GodPermission{}},
		testutil.D("0.667"),
		time.Hour*24*7,
		types.TALLY_OPTION_DEADLINE,
		testutil.D("0.4"),
		"hard",
	)
	var defaultID uint64 = 1
	firstBlockTime := time.Date(1998, time.January, 1, 1, 0, 0, 0, time.UTC)

	genAddrs := suite.Addresses[:8]                       // Genesis accounts
	genCoinCounts := []int64{0, 0, 0, 10, 20, 30, 40, 50} // Genesis token balances

	// ---------------------- Polling information ----------------------
	//	150hard total token supply: 150 possible votes
	//  40% quroum: 60 votes required to meet quroum
	//  66.67% voting threshold: 2/3rds of votes must be Yes votes
	// -----------------------------------------------------------------

	testcases := []struct {
		name           string
		committee      *types.TokenCommittee
		votes          []types.Vote
		proposalPasses bool
	}{
		{
			name:      "not enough votes to meet quroum",
			committee: tokenCom,
			votes: []types.Vote{
				{ProposalID: defaultID, Voter: genAddrs[7], VoteType: types.VOTE_TYPE_YES}, // Holds 50 tokens
			},
			proposalPasses: false, // 60 vote quroum; 50 total votes; 50 yes votes. Doesn't pass 40% quroum.
		},
		{
			name:      "enough votes to meet quroum and enough Yes votes to pass voting threshold",
			committee: tokenCom,
			votes: []types.Vote{
				{ProposalID: defaultID, Voter: genAddrs[3], VoteType: types.VOTE_TYPE_NO},  // Holds 10 tokens
				{ProposalID: defaultID, Voter: genAddrs[7], VoteType: types.VOTE_TYPE_YES}, // Holds 50 tokens
			},
			proposalPasses: true, // 60 vote quroum; 60 total votes; 50 Yes votes. Passes the 66.67% voting threshold.
		},
		{
			name:      "enough votes to meet quroum via Abstain votes and enough Yes votes to pass voting threshold",
			committee: tokenCom,
			votes: []types.Vote{
				{ProposalID: defaultID, Voter: genAddrs[3], VoteType: types.VOTE_TYPE_ABSTAIN}, // Holds 10 tokens
				{ProposalID: defaultID, Voter: genAddrs[7], VoteType: types.VOTE_TYPE_YES},     // Holds 50 tokens
			},
			proposalPasses: true, // 60 vote quroum; 60 total votes; 50 Yes votes. Passes the 66.67% voting threshold.
		},
		{
			name:      "enough votes to meet quroum but not enough Yes votes to pass voting threshold",
			committee: tokenCom,
			votes: []types.Vote{
				{ProposalID: defaultID, Voter: genAddrs[4], VoteType: types.VOTE_TYPE_YES}, // Holds 20 tokens
				{ProposalID: defaultID, Voter: genAddrs[6], VoteType: types.VOTE_TYPE_NO},  // Holds 40 tokens
			},
			proposalPasses: false, // 60 vote quroum; 60 total votes; 20 Yes votes. Doesn't pass 66.67% voting threshold.
		},
		{
			name:      "enough votes to pass voting threshold (multiple Yes votes, multiple No votes)",
			committee: tokenCom,
			votes: []types.Vote{
				{ProposalID: defaultID, Voter: genAddrs[3], VoteType: types.VOTE_TYPE_YES}, // Holds 10 tokens
				{ProposalID: defaultID, Voter: genAddrs[4], VoteType: types.VOTE_TYPE_YES}, // Holds 20 tokens
				{ProposalID: defaultID, Voter: genAddrs[5], VoteType: types.VOTE_TYPE_YES}, // Holds 30 tokens
				{ProposalID: defaultID, Voter: genAddrs[6], VoteType: types.VOTE_TYPE_NO},  // Holds 40 tokens
				{ProposalID: defaultID, Voter: genAddrs[7], VoteType: types.VOTE_TYPE_YES}, // Holds 50 tokens
			},
			proposalPasses: true, // 60 vote quroum; 150 total votes; 110 Yes votes. Passes the 66.67% voting threshold.
		},
		{
			name:      "not enough votes to pass voting threshold (multiple Yes votes, multiple No votes)",
			committee: tokenCom,
			votes: []types.Vote{
				{ProposalID: defaultID, Voter: genAddrs[3], VoteType: types.VOTE_TYPE_YES}, // Holds 10 tokens
				{ProposalID: defaultID, Voter: genAddrs[4], VoteType: types.VOTE_TYPE_YES}, // Holds 20 tokens
				{ProposalID: defaultID, Voter: genAddrs[5], VoteType: types.VOTE_TYPE_YES}, // Holds 30 tokens
				{ProposalID: defaultID, Voter: genAddrs[6], VoteType: types.VOTE_TYPE_YES}, // Holds 40 tokens
				{ProposalID: defaultID, Voter: genAddrs[7], VoteType: types.VOTE_TYPE_NO},  // Holds 50 tokens
			},
			proposalPasses: false, // 60 vote quroum; 150 total votes; 100 Yes votes. Doesn't pass 66.67% voting threshold.
		},
	}

	// Convert accounts/token balances into format expected by genesis generation
	var genCoins []sdk.Coins
	var totalSupply sdk.Coins
	for _, amount := range genCoinCounts {
		userCoin := testutil.C("hard", amount)
		genCoins = append(genCoins, testutil.Cs(userCoin))
		totalSupply = totalSupply.Add(userCoin)
	}

	for _, tc := range testcases {
		suite.Run(tc.name, func() {
			// Create local testApp because suite doesn't run the SetupTest function for subtests
			tApp := app.NewTestApp()
			keeper := tApp.GetCommitteeKeeper()
			ctx := tApp.NewContext(true, tmproto.Header{Height: 1, Time: firstBlockTime})

			tApp.InitializeFromGenesisStates(
				committeeGenState(
					tApp.AppCodec(),
					[]types.Committee{tc.committee},
					[]types.Proposal{types.MustNewProposal(
						govtypes.NewTextProposal("A Title", "A description of this proposal."),
						defaultID,
						tc.committee.GetID(),
						firstBlockTime.Add(time.Hour*24*7),
					)},
					tc.votes,
				),
				bankGenState(tApp.AppCodec(), totalSupply),
				app.NewFundedGenStateWithCoins(tApp.AppCodec(), genCoins, genAddrs),
			)

			proposalPasses := keeper.GetTokenCommitteeProposalResult(ctx, defaultID, tc.committee)
			suite.Equal(tc.proposalPasses, proposalPasses)
		})
	}
}

func (suite *keeperTestSuite) TestCloseProposal() {
	memberCom := types.MustNewMemberCommittee(
		12,
		"This committee is for testing.",
		suite.Addresses[:5],
		[]types.Permission{&types.GodPermission{}},
		testutil.D("0.667"),
		time.Hour*24*7,
		types.TALLY_OPTION_DEADLINE,
	)

	var proposalID uint64 = 1
	firstBlockTime := time.Date(1998, time.January, 1, 1, 0, 0, 0, time.UTC)

	tApp := app.NewTestApp()
	keeper := tApp.GetCommitteeKeeper()
	ctx := tApp.NewContext(true, tmproto.Header{Height: 1, Time: firstBlockTime})

	tApp.InitializeFromGenesisStates(
		committeeGenState(
			tApp.AppCodec(),
			[]types.Committee{memberCom},
			[]types.Proposal{types.MustNewProposal(
				govtypes.NewTextProposal("A Title", "A description of this proposal."),
				proposalID,
				memberCom.GetID(),
				firstBlockTime.Add(time.Hour*24*7),
			)},
			[]types.Vote{},
		),
	)

	// Confirm proposal exists
	proposal, found := keeper.GetProposal(ctx, proposalID)
	suite.True(found)

	// Close proposal
	keeper.CloseProposal(ctx, proposal, types.Passed)

	events := ctx.EventManager().Events()
	event := events[0]
	suite.Require().Equal("proposal_close", event.Type)

	hasProposalTallyAttr := false
	for _, attr := range event.Attributes {
		if string(attr.GetKey()) == "proposal_tally" {
			hasProposalTallyAttr = true
			valueStr := string(attr.GetValue())
			suite.Contains(valueStr, "proposal_id")
			suite.Contains(valueStr, "yes_votes")
			suite.Contains(valueStr, "current_votes")
			suite.Contains(valueStr, "possible_votes")
			suite.Contains(valueStr, "vote_threshold")
			suite.Contains(valueStr, "quorum")
		}
	}
	suite.Require().True(hasProposalTallyAttr)

	// Confirm proposal doesn't exist
	_, found = keeper.GetProposal(ctx, proposalID)
	suite.False(found)
}

func committeeGenState(cdc codec.Codec, committees []types.Committee, proposals []types.Proposal, votes []types.Vote) app.GenesisState {
	gs := types.NewGenesisState(
		uint64(len(proposals)+1),
		committees,
		proposals,
		votes,
	)
	return app.GenesisState{types.ModuleName: cdc.MustMarshalJSON(gs)}
}

func bankGenState(cdc codec.Codec, coins sdk.Coins) app.GenesisState {
	gs := banktypes.DefaultGenesisState()
	gs.Supply = coins
	return app.GenesisState{banktypes.ModuleName: cdc.MustMarshalJSON(gs)}
}

type UnregisteredPubProposal struct {
	govtypes.TextProposal
}

func (UnregisteredPubProposal) ProposalRoute() string { return "unregistered" }
func (UnregisteredPubProposal) ProposalType() string  { return "unregistered" }

var _ types.PubProposal = &UnregisteredPubProposal{}

// func (suite *keeperTestSuite) TestValidatePubProposal() {

// 	testcases := []struct {
// 		name        string
// 		pubProposal types.PubProposal
// 		expectErr   bool
// 	}{
// 		{
// 			name:        "valid (text proposal)",
// 			pubProposal: govtypes.NewTextProposal("A Title", "A description of this proposal."),
// 			expectErr:   false,
// 		},
// 		{
// 			name: "valid (param change proposal)",
// 			pubProposal: paramsproposal.NewParameterChangeProposal(
// 				"Change the debt limit",
// 				"This proposal changes the debt limit of the cdp module.",
// 				[]paramsproposal.ParamChange{{
// 					Subspace: cdptypes.ModuleName,
// 					Key:      string(cdptypes.KeyGlobalDebtLimit),
// 					Value:    string(types.ModuleCdc.MustMarshalJSON(c("usdx", 100000000000))),
// 				}},
// 			),
// 			expectErr: false,
// 		},
// 		{
// 			name:        "invalid (missing title)",
// 			pubProposal: govtypes.TextProposal{Description: "A description of this proposal."},
// 			expectErr:   true,
// 		},
// 		{
// 			name:        "invalid (unregistered)",
// 			pubProposal: UnregisteredPubProposal{govtypes.TextProposal{Title: "A Title", Description: "A description of this proposal."}},
// 			expectErr:   true,
// 		},
// 		{
// 			name:        "invalid (nil)",
// 			pubProposal: nil,
// 			expectErr:   true,
// 		},
// 		{
// 			name: "invalid (proposal handler fails)",
// 			pubProposal: paramsproposal.NewParameterChangeProposal(
// 				"A Title",
// 				"A description of this proposal.",
// 				[]paramsproposal.ParamChange{{
// 					Subspace: "nonsense-subspace",
// 					Key:      "nonsense-key",
// 					Value:    "nonsense-value",
// 				}},
// 			),
// 			expectErr: true,
// 		},
// 		{
// 			name: "invalid (proposal handler panics)",
// 			pubProposal: paramsproposal.NewParameterChangeProposal(
// 				"A Title",
// 				"A description of this proposal.",
// 				[]paramsproposal.ParamChange{{
// 					Subspace: cdptypes.ModuleName,
// 					Key:      "nonsense-key", // a valid Subspace but invalid Key will trigger a panic in the paramchange propsal handler
// 					Value:    "nonsense-value",
// 				}},
// 			),
// 			expectErr: true,
// 		},
// 		{
// 			name: "invalid (proposal handler fails - invalid json)",
// 			pubProposal: paramsproposal.NewParameterChangeProposal(
// 				"A Title",
// 				"A description of this proposal.",
// 				[]paramsproposal.ParamChange{{
// 					Subspace: cdptypes.ModuleName,
// 					Key:      string(cdptypes.KeyGlobalDebtLimit),
// 					Value:    `{"denom": "usdx",`,
// 				}},
// 			),
// 			expectErr: true,
// 		},
// 	}

// 	for _, tc := range testcases {
// 		suite.Run(tc.name, func() {
// 			err := suite.keeper.ValidatePubProposal(suite.ctx, tc.pubProposal)
// 			if tc.expectErr {
// 				suite.NotNil(err)
// 			} else {
// 				suite.NoError(err)
// 			}
// 		})
// 	}
// }

// func (suite *keeperTestSuite) TestProcessProposals() {

// 	firstBlockTime := time.Date(1998, time.January, 1, 1, 0, 0, 0, time.UTC)

// 	genAddrs := suite.Addresses[:4]      // Genesis accounts
// 	genCoinCounts := []int64{1, 1, 1, 1} // Genesis token balances
// 	// Convert accounts/token balances into format expected by genesis generation
// 	var genCoins []sdk.Coins
// 	var totalSupply sdk.Coins
// 	for _, amount := range genCoinCounts {
// 		userCoin := testutil.C("hard", amount)
// 		genCoins = append(genCoins, testutil.Cs(userCoin))
// 		totalSupply = totalSupply.Add(userCoin)
// 	}

// 	// Set up committees
// 	committees := []types.Committee{
// 		// 	1. FPTP MemberCommmittee
// 		types.MustNewMemberCommittee(

// 			1,
// 			"FTPT MemberCommittee",
// 			genAddrs,
// 			[]types.Permission{&types.GodPermission{}},
// 			testutil.D("0.667"),
// 			time.Hour*24*7,
// 			types.TALLY_OPTION_FIRST_PAST_THE_POST,
// 		),
// 		// 	2. FPTP TokenCommittee
// 		types.MustNewTokenCommittee(

// 			2,
// 			"FTPT TokenCommittee",
// 			genAddrs,
// 			[]types.Permission{&types.GodPermission{}},
// 			testutil.D("0.667"),
// 			time.Hour*24*7,
// 			types.TALLY_OPTION_FIRST_PAST_THE_POST,
// 			testutil.D("0.30"),
// 			"hard",
// 		),
// 		// 	3. Deadline MemberCommmittee
// 		types.MustNewMemberCommittee(

// 			3,
// 			"Deadline MemberCommittee",
// 			genAddrs,
// 			[]types.Permission{&types.GodPermission{}},
// 			testutil.D("0.667"),
// 			time.Hour*24*7,
// 			types.TALLY_OPTION_DEADLINE,
// 		),
// 		// 	4. Deadline TokenCommittee
// 		types.MustNewTokenCommittee(
// 			4,
// 			"Deadline TokenCommittee",
// 			genAddrs,
// 			[]types.Permission{&types.GodPermission{}},
// 			testutil.D("0.667"),
// 			time.Hour*24*7,
// 			types.TALLY_OPTION_DEADLINE,
// 			testutil.D("0.30"),
// 			"hard",
// 		),
// 		// 	5. PTP MemberCommmittee without permissions
// 		types.MustNewMemberCommittee(

// 			5,
// 			"FTPT MemberCommittee without permissions",
// 			genAddrs,
// 			nil,
// 			testutil.D("0.667"),
// 			time.Hour*24*7,
// 			types.TALLY_OPTION_FIRST_PAST_THE_POST,
// 		),
// 	}

// 	// Set up proposals that correspond 1:1 with each committee
// 	proposals := []types.Proposal{
// 		types.MustNewProposal(
// 			govtypes.NewTextProposal("Proposal 1", "This proposal is for the FPTP MemberCommmittee."),
// 			1,
// 			1,
// 			firstBlockTime.Add(7*24*time.Hour),
// 		),
// 		types.MustNewProposal(
// 			govtypes.NewTextProposal("Proposal 2", "This proposal is for the FPTP TokenCommittee."),
// 			2,
// 			2,

// 			firstBlockTime.Add(7*24*time.Hour),
// 		),
// 		types.MustNewProposal(
// 			govtypes.NewTextProposal("Proposal 3", "This proposal is for the Deadline MemberCommmittee."),
// 			3,
// 			3,

// 			firstBlockTime.Add(7*24*time.Hour),
// 		),
// 		types.MustNewProposal(
// 			govtypes.NewTextProposal("Proposal 4", "This proposal is for the Deadline TokenCommittee."),
// 			4,
// 			4,

// 			firstBlockTime.Add(7*24*time.Hour),
// 		),
// 		types.MustNewProposal(
// 			govtypes.NewTextProposal("Proposal 5", "This proposal is for the FPTP MemberCommmittee without permissions."),
// 			5,
// 			5,
// 			firstBlockTime.Add(7*24*time.Hour),
// 		),
// 	}

// 	// Each test case targets 1 committee/proposal via targeted votes
// 	testcases := []struct {
// 		name                             string
// 		ID                               uint64
// 		votes                            []types.Vote
// 		expectedToCompleteBeforeDeadline bool
// 		expectedOutcome                  types.ProposalOutcome
// 	}{
// 		{
// 			name: "FPTP MemberCommittee proposal does not have enough votes to pass",
// 			ID:   1,
// 			votes: []types.Vote{
// 				{ProposalID: 1, Voter: genAddrs[0], VoteType: types.VOTE_TYPE_YES},
// 			},
// 			expectedToCompleteBeforeDeadline: false,
// 			expectedOutcome:                  types.Failed,
// 		},
// 		{
// 			name: "FPTP MemberCommittee proposal has enough votes to pass before deadline",
// 			ID:   1,
// 			votes: []types.Vote{
// 				{ProposalID: 1, Voter: genAddrs[0], VoteType: types.VOTE_TYPE_YES},
// 				{ProposalID: 1, Voter: genAddrs[1], VoteType: types.VOTE_TYPE_YES},
// 				{ProposalID: 1, Voter: genAddrs[2], VoteType: types.VOTE_TYPE_YES},
// 			},
// 			expectedToCompleteBeforeDeadline: true,
// 			expectedOutcome:                  types.Passed,
// 		},
// 		{
// 			name: "FPTP TokenCommittee proposal does not have enough votes to pass",
// 			ID:   2,
// 			votes: []types.Vote{
// 				{ProposalID: 2, Voter: genAddrs[0], VoteType: types.VOTE_TYPE_YES},
// 			},
// 			expectedToCompleteBeforeDeadline: false,
// 			expectedOutcome:                  types.Failed,
// 		},
// 		{
// 			name: "FPTP TokenCommittee proposal has enough votes to pass before deadline",
// 			ID:   2,
// 			votes: []types.Vote{
// 				{ProposalID: 2, Voter: genAddrs[0], VoteType: types.VOTE_TYPE_YES},
// 				{ProposalID: 2, Voter: genAddrs[1], VoteType: types.VOTE_TYPE_YES},
// 				{ProposalID: 2, Voter: genAddrs[2], VoteType: types.VOTE_TYPE_YES},
// 			},
// 			expectedToCompleteBeforeDeadline: true,
// 			expectedOutcome:                  types.Passed,
// 		},
// 		{
// 			name: "Deadline MemberCommittee proposal with enough votes to pass only passes after deadline",
// 			ID:   3,
// 			votes: []types.Vote{
// 				{ProposalID: 3, Voter: genAddrs[0], VoteType: types.VOTE_TYPE_YES},
// 				{ProposalID: 3, Voter: genAddrs[1], VoteType: types.VOTE_TYPE_YES},
// 				{ProposalID: 3, Voter: genAddrs[2], VoteType: types.VOTE_TYPE_YES},
// 			},
// 			expectedOutcome: types.Passed,
// 		},
// 		{
// 			name: "Deadline MemberCommittee proposal doesn't have enough votes to pass",
// 			ID:   3,
// 			votes: []types.Vote{
// 				{ProposalID: 3, Voter: genAddrs[0], VoteType: types.VOTE_TYPE_YES},
// 			},
// 			expectedOutcome: types.Failed,
// 		},
// 		{
// 			name: "Deadline TokenCommittee proposal with enough votes to pass only passes after deadline",
// 			ID:   4,
// 			votes: []types.Vote{
// 				{ProposalID: 4, Voter: genAddrs[0], VoteType: types.VOTE_TYPE_YES},
// 				{ProposalID: 4, Voter: genAddrs[1], VoteType: types.VOTE_TYPE_YES},
// 				{ProposalID: 4, Voter: genAddrs[2], VoteType: types.VOTE_TYPE_YES},
// 			},
// 			expectedOutcome: types.Passed,
// 		},
// 		{
// 			name: "Deadline TokenCommittee proposal doesn't have enough votes to pass",
// 			ID:   4,
// 			votes: []types.Vote{
// 				{ProposalID: 4, Voter: genAddrs[0], VoteType: types.VOTE_TYPE_YES},
// 			},
// 			expectedOutcome: types.Failed,
// 		},
// 		{
// 			name: "FPTP MemberCommittee doesn't have permissions to enact passed proposal",
// 			ID:   5,
// 			votes: []types.Vote{
// 				{ProposalID: 5, Voter: genAddrs[0], VoteType: types.VOTE_TYPE_YES},
// 				{ProposalID: 5, Voter: genAddrs[1], VoteType: types.VOTE_TYPE_YES},
// 				{ProposalID: 5, Voter: genAddrs[2], VoteType: types.VOTE_TYPE_YES},
// 			},
// 			expectedToCompleteBeforeDeadline: true,
// 			expectedOutcome:                  types.Invalid,
// 		},
// 	}

// 	for _, tc := range testcases {
// 		suite.Run(tc.name, func() {
// 			// Create local testApp because suite doesn't run the SetupTest function for subtests
// 			tApp := app.NewTestApp()
// 			keeper := tApp.GetCommitteeKeeper()
// 			ctx := tApp.NewContext(true, tmproto.Header{Height: 1, Time: firstBlockTime})

// 			// Initialize all committees, proposals, and votes via Genesis
// 			tApp.InitializeFromGenesisStates(
// 				committeeGenState(tApp.AppCodec(), committees, proposals, tc.votes),
// 				bankGenState(tApp.AppCodec(), totalSupply),
// 				app.NewFundedGenStateWithCoins(tApp.AppCodec(), genCoins, genAddrs),
// 			)

// 			// Load committee from the store
// 			committee, found := keeper.GetCommittee(ctx, tc.ID)
// 			suite.True(found)

// 			// Process proposals
// 			ctx = ctx.WithBlockTime(firstBlockTime)
// 			keeper.ProcessProposals(ctx)

// 			// Fetch proposal and votes from the store
// 			votes := getProposalVoteMap(keeper, ctx)
// 			proposal, found := keeper.GetProposal(ctx, tc.ID)

// 			if committee.GetTallyOption() == types.TALLY_OPTION_FIRST_PAST_THE_POST {
// 				if tc.expectedToCompleteBeforeDeadline {
// 					suite.False(found)
// 					suite.Empty(votes[tc.ID])

// 					// Check proposal outcome
// 					outcome, err := getProposalOutcome(tc.ID, ctx.EventManager().Events(), tApp.LegacyAmino())
// 					suite.NoError(err)
// 					suite.Equal(tc.expectedOutcome, outcome)
// 					return
// 				} else {
// 					suite.True(found)
// 					suite.NotEmpty(votes[tc.ID])
// 				}
// 			}

// 			// Move block time to deadline
// 			ctx = ctx.WithBlockTime(proposal.Deadline)
// 			keeper.ProcessProposals(ctx)

// 			// Fetch proposal and votes from the store
// 			votes = getProposalVoteMap(keeper, ctx)
// 			proposal, found = keeper.GetProposal(ctx, tc.ID)
// 			suite.False(found)
// 			suite.Empty(votes[proposal.ID])

// 			// Check proposal outcome
// 			outcome, err := getProposalOutcome(tc.ID, ctx.EventManager().Events(), tApp.LegacyAmino())
// 			suite.NoError(err)
// 			suite.Equal(tc.expectedOutcome, outcome)
// 		})
// 	}
// }

// // getProposalOutcome checks the outcome of a proposal via a `proposal_close` event whose `proposal_id`
// // matches argument proposalID
// func getProposalOutcome(proposalID uint64, events sdk.Events, cdc *codec.LegacyAmino) (types.ProposalOutcome, error) {
// 	// Marshal proposal ID to match against event attribute
// 	x, _ := cdc.MarshalJSON(proposalID)
// 	marshaledID := x[1 : len(x)-1]

// 	for _, event := range events {
// 		if event.Type == types.EventTypeProposalClose {
// 			var proposalOutcome types.ProposalOutcome
// 			correctProposal := false
// 			for _, attribute := range event.Attributes {
// 				// Only get outcome of specific proposal
// 				if bytes.Compare(attribute.GetKey(), []byte("proposal_id")) == 0 {
// 					if bytes.Compare(attribute.GetValue(), marshaledID) == 0 {
// 						correctProposal = true
// 					}
// 				}
// 				// Match event attribute bytes to marshaled outcome
// 				if bytes.Compare(attribute.GetKey(), []byte(types.AttributeKeyProposalOutcome)) == 0 {
// 					outcome, err := types.MatchMarshaledOutcome(attribute.GetValue(), cdc)
// 					if err != nil {
// 						return 0, err
// 					}
// 					proposalOutcome = outcome
// 				}
// 			}
// 			// If this is the desired proposal, return the outcome
// 			if correctProposal {
// 				return proposalOutcome, nil
// 			}
// 		}
// 	}
// 	return 0, nil
// }
