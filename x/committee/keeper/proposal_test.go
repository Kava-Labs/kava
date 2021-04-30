package keeper_test

import (
	"bytes"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov"
	"github.com/cosmos/cosmos-sdk/x/params"
	"github.com/cosmos/cosmos-sdk/x/supply"

	amino "github.com/tendermint/go-amino"
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
	defaultCommitteeID := uint64(12)
	normalCom := types.BaseCommittee{
		ID:               defaultCommitteeID,
		Description:      "This committee is for testing.",
		Members:          suite.addresses[:2],
		Permissions:      []types.Permission{types.GodPermission{}},
		VoteThreshold:    d("0.667"),
		ProposalDuration: time.Hour * 24 * 7,
		TallyOption:      types.FirstPastThePost,
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
					Type:         "bnb-a",
					DebtLimit:    true,
					StabilityFee: true,
				},
			},
		},
	}

	testCP := cdptypes.CollateralParams{{
		Denom:               "bnb",
		Type:                "bnb-a",
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
		committee   types.BaseCommittee
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
						Subspace: "cdp",
						Key:      string(cdptypes.KeyDebtThreshold),
						Value:    string(suite.app.Codec().MustMarshalJSON(i(1000000000))),
					},
					{
						Subspace: "cdp",
						Key:      string(cdptypes.KeyCollateralParams),
						Value:    string(suite.app.Codec().MustMarshalJSON(newValidCP)),
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
						Subspace: "cdp",
						Key:      string(cdptypes.KeyDebtThreshold),
						Value:    string(suite.app.Codec().MustMarshalJSON(i(1000000000))),
					},
					{
						Subspace: "cdp",
						Key:      string(cdptypes.KeyCollateralParams),
						Value:    string(suite.app.Codec().MustMarshalJSON(newInvalidCP)),
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
			// Cast BaseCommittee to MemberCommittee (if required) to meet Committee interface requirement
			if tc.committee.ID == defaultCommitteeID {
				keeper.SetCommittee(ctx, types.MemberCommittee{BaseCommittee: tc.committee})
			}

			id, err := keeper.SubmitProposal(ctx, tc.proposer, tc.committeeID, tc.pubProposal)

			if tc.expectErr {
				suite.NotNil(err)
			} else {
				suite.NoError(err)
				pr, found := keeper.GetProposal(ctx, id)
				suite.True(found)
				suite.Equal(tc.committeeID, pr.CommitteeID)
				suite.Equal(ctx.BlockTime().Add(tc.committee.GetProposalDuration()), pr.Deadline)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestAddVote() {
	memberCom := types.MemberCommittee{
		BaseCommittee: types.BaseCommittee{
			ID:          12,
			Members:     suite.addresses[:2],
			Permissions: []types.Permission{types.GodPermission{}},
		},
	}
	tokenCom := types.TokenCommittee{
		BaseCommittee: types.BaseCommittee{
			ID:          12,
			Members:     suite.addresses[:2],
			Permissions: []types.Permission{types.GodPermission{}},
		},
		Quorum:     d("0.4"),
		TallyDenom: "hard",
	}
	nonMemberAddr := suite.addresses[4]
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
			voteType:   types.Yes,
			expectErr:  false,
		},
		{
			name:       "normal TokenCommittee",
			committee:  tokenCom,
			proposalID: types.DefaultNextProposalID,
			voter:      nonMemberAddr,
			voteType:   types.Yes,
			expectErr:  false,
		},
		{
			name:       "nonexistent proposal",
			committee:  memberCom,
			proposalID: 9999999,
			voter:      memberCom.Members[0],
			voteType:   types.Yes,
			expectErr:  true,
		},
		{
			name:       "proposal expired",
			committee:  memberCom,
			proposalID: types.DefaultNextProposalID,
			voter:      memberCom.Members[0],
			voteTime:   firstBlockTime.Add(memberCom.ProposalDuration),
			voteType:   types.Yes,
			expectErr:  true,
		},
		{
			name:       "MemberCommittee: voter not committee member",
			committee:  memberCom,
			proposalID: types.DefaultNextProposalID,
			voter:      nonMemberAddr,
			voteType:   types.Yes,
			expectErr:  true,
		},
		{
			name:       "MemberCommittee: voter votes no",
			committee:  memberCom,
			proposalID: types.DefaultNextProposalID,
			voter:      memberCom.Members[0],
			voteType:   types.No,
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
			keeper.SetCommittee(ctx, tc.committee)
			_, err := keeper.SubmitProposal(ctx, tc.committee.GetMembers()[0], tc.committee.GetID(), gov.NewTextProposal("A Title", "A description of this proposal."))
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

func (suite *KeeperTestSuite) TestTallyMemberCommitteeVotes() {
	memberCom := types.MemberCommittee{
		BaseCommittee: types.BaseCommittee{
			ID:               12,
			Description:      "This committee is for testing.",
			Members:          suite.addresses[:5],
			Permissions:      []types.Permission{types.GodPermission{}},
			VoteThreshold:    d("0.667"),
			ProposalDuration: time.Hour * 24 * 7,
			TallyOption:      types.Deadline,
		},
	}
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
			expectedVoteCount: d("0"),
		},
		{
			name: "has 1 vote",
			votes: []types.Vote{
				{ProposalID: defaultProposalID, Voter: suite.addresses[0], VoteType: types.Yes},
			},
			expectedVoteCount: d("1"),
		},
		{
			name: "has multiple votes",
			votes: []types.Vote{
				{ProposalID: defaultProposalID, Voter: suite.addresses[0], VoteType: types.Yes},
				{ProposalID: defaultProposalID, Voter: suite.addresses[1], VoteType: types.Yes},
				{ProposalID: defaultProposalID, Voter: suite.addresses[2], VoteType: types.Yes},
				{ProposalID: defaultProposalID, Voter: suite.addresses[3], VoteType: types.Yes},
			},
			expectedVoteCount: d("4"),
		},
	}

	for _, tc := range testcases {
		// Set up test app
		tApp := app.NewTestApp()
		keeper := tApp.GetCommitteeKeeper()
		ctx := tApp.NewContext(true, abci.Header{Height: 1, Time: firstBlockTime})

		// Initialize test app with genesis state
		tApp.InitializeFromGenesisStates(
			committeeGenState(
				tApp.Codec(),
				[]types.Committee{memberCom},
				[]types.Proposal{{
					PubProposal: gov.NewTextProposal("A Title", "A description of this proposal."),
					ID:          defaultProposalID,
					CommitteeID: memberCom.GetID(),
					Deadline:    firstBlockTime.Add(time.Hour * 24 * 7),
				}},
				tc.votes,
			),
		)

		// Check that all votes are counted
		currentVotes := keeper.TallyMemberCommitteeVotes(ctx, defaultProposalID)
		suite.Equal(tc.expectedVoteCount, currentVotes)
	}
}

func (suite *KeeperTestSuite) TestTallyTokenCommitteeVotes() {
	tokenCom := types.TokenCommittee{
		BaseCommittee: types.BaseCommittee{
			ID:               12,
			Description:      "This committee is for testing.",
			Members:          suite.addresses[:5],
			Permissions:      []types.Permission{types.GodPermission{}},
			VoteThreshold:    d("0.667"),
			ProposalDuration: time.Hour * 24 * 7,
			TallyOption:      types.Deadline,
		},
		TallyDenom: "hard",
		Quorum:     d("0.4"),
	}
	var defaultProposalID uint64 = 1
	firstBlockTime := time.Date(1998, time.January, 1, 1, 0, 0, 0, time.UTC)

	genAddrs := suite.addresses[:8]                       // Genesis accounts
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
			expectedYesVoteCount:   d("0"),
			expectedNoVoteCount:    d("0"),
			expectedTotalVoteCount: d("0"),
		},
		{
			name: "counts token holder 'Yes' votes",
			votes: []types.Vote{
				{ProposalID: defaultProposalID, Voter: genAddrs[4], VoteType: types.Yes}, // Token holder
			},
			expectedYesVoteCount:   sdk.NewDec(genCoinCounts[4]),
			expectedNoVoteCount:    d("0"),
			expectedTotalVoteCount: sdk.NewDec(genCoinCounts[4]),
		},
		{
			name: "does not count non-token holder 'Yes' votes",
			votes: []types.Vote{
				{ProposalID: defaultProposalID, Voter: genAddrs[4], VoteType: types.Yes}, // Token holder
				{ProposalID: defaultProposalID, Voter: genAddrs[0], VoteType: types.Yes}, // Non-token holder
			},
			expectedYesVoteCount:   sdk.NewDec(genCoinCounts[4]),
			expectedNoVoteCount:    d("0"),
			expectedTotalVoteCount: sdk.NewDec(genCoinCounts[4]),
		},
		{
			name: "counts multiple 'Yes' votes from token holders",
			votes: []types.Vote{
				{ProposalID: defaultProposalID, Voter: genAddrs[4], VoteType: types.Yes}, // Token holder
				{ProposalID: defaultProposalID, Voter: genAddrs[5], VoteType: types.Yes}, // Token holder
				{ProposalID: defaultProposalID, Voter: genAddrs[6], VoteType: types.Yes}, // Token holder
			},
			expectedYesVoteCount:   sdk.NewDec(genCoinCounts[4] + genCoinCounts[5] + genCoinCounts[6]),
			expectedNoVoteCount:    d("0"),
			expectedTotalVoteCount: sdk.NewDec(genCoinCounts[4] + genCoinCounts[5] + genCoinCounts[6]),
		},
		{
			name: "counts token holder 'No' votes",
			votes: []types.Vote{
				{ProposalID: defaultProposalID, Voter: genAddrs[4], VoteType: types.No}, // Token holder
			},
			expectedYesVoteCount:   d("0"),
			expectedNoVoteCount:    sdk.NewDec(genCoinCounts[4]),
			expectedTotalVoteCount: sdk.NewDec(genCoinCounts[4]),
		},
		{
			name: "does not count non-token holder 'No' votes",
			votes: []types.Vote{
				{ProposalID: defaultProposalID, Voter: genAddrs[4], VoteType: types.No}, // Token holder
				{ProposalID: defaultProposalID, Voter: genAddrs[0], VoteType: types.No}, // Non-token holder
			},
			expectedYesVoteCount:   d("0"),
			expectedNoVoteCount:    sdk.NewDec(genCoinCounts[4]),
			expectedTotalVoteCount: sdk.NewDec(genCoinCounts[4]),
		},
		{
			name: "counts multiple 'No' votes from token holders",
			votes: []types.Vote{
				{ProposalID: defaultProposalID, Voter: genAddrs[4], VoteType: types.No}, // Token holder
				{ProposalID: defaultProposalID, Voter: genAddrs[5], VoteType: types.No}, // Token holder
				{ProposalID: defaultProposalID, Voter: genAddrs[6], VoteType: types.No}, // Token holder
			},
			expectedYesVoteCount:   d("0"),
			expectedNoVoteCount:    sdk.NewDec(genCoinCounts[4] + genCoinCounts[5] + genCoinCounts[6]),
			expectedTotalVoteCount: sdk.NewDec(genCoinCounts[4] + genCoinCounts[5] + genCoinCounts[6]),
		},
		{
			name: "includes token holder 'Abstain' votes in total vote count",
			votes: []types.Vote{
				{ProposalID: defaultProposalID, Voter: genAddrs[4], VoteType: types.Abstain}, // Token holder
			},
			expectedYesVoteCount:   d("0"),
			expectedNoVoteCount:    d("0"),
			expectedTotalVoteCount: sdk.NewDec(genCoinCounts[4]),
		},
	}

	// Convert accounts/token balances into format expected by genesis generation
	var genCoins []sdk.Coins
	var totalSupply sdk.Coins
	for _, amount := range genCoinCounts {
		userCoin := c("hard", amount)
		genCoins = append(genCoins, cs(userCoin))
		totalSupply = totalSupply.Add(userCoin)
	}

	for _, tc := range testcases {
		// Set up test app
		tApp := app.NewTestApp()
		keeper := tApp.GetCommitteeKeeper()
		ctx := tApp.NewContext(true, abci.Header{Height: 1, Time: firstBlockTime})

		// Initialize test app with genesis state
		tApp.InitializeFromGenesisStates(
			committeeGenState(
				tApp.Codec(),
				[]types.Committee{tokenCom},
				[]types.Proposal{{
					PubProposal: gov.NewTextProposal("A Title", "A description of this proposal."),
					ID:          defaultProposalID,
					CommitteeID: tokenCom.GetID(),
					Deadline:    firstBlockTime.Add(time.Hour * 24 * 7),
				}},
				tc.votes,
			),
			supplyGenState(tApp.Codec(), totalSupply),
			app.NewAuthGenState(genAddrs, genCoins),
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

func (suite *KeeperTestSuite) TestGetMemberCommitteeProposalResult() {
	memberCom := types.MemberCommittee{
		BaseCommittee: types.BaseCommittee{
			ID:               12,
			Description:      "This committee is for testing.",
			Members:          suite.addresses[:5],
			Permissions:      []types.Permission{types.GodPermission{}},
			VoteThreshold:    d("0.667"),
			ProposalDuration: time.Hour * 24 * 7,
			TallyOption:      types.Deadline,
		},
	}
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
				{ProposalID: defaultID, Voter: suite.addresses[0], VoteType: types.Yes},
				{ProposalID: defaultID, Voter: suite.addresses[1], VoteType: types.Yes},
				{ProposalID: defaultID, Voter: suite.addresses[2], VoteType: types.Yes},
				{ProposalID: defaultID, Voter: suite.addresses[3], VoteType: types.Yes},
			},
			proposalPasses: true,
		},
		{
			name:      "not enough votes",
			committee: memberCom,
			votes: []types.Vote{
				{ProposalID: defaultID, Voter: suite.addresses[0], VoteType: types.Yes},
			},
			proposalPasses: false,
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
						CommitteeID: tc.committee.GetID(),
						Deadline:    firstBlockTime.Add(time.Hour * 24 * 7),
					}},
					tc.votes,
				),
			)

			proposalPasses := keeper.GetMemberCommitteeProposalResult(ctx, defaultID, tc.committee)
			suite.Equal(tc.proposalPasses, proposalPasses)
		})
	}
}

func (suite *KeeperTestSuite) TestGetTokenCommitteeProposalResult() {
	tokenCom := types.TokenCommittee{
		BaseCommittee: types.BaseCommittee{
			ID:               12,
			Description:      "This committee is for testing.",
			Members:          suite.addresses[:5],
			Permissions:      []types.Permission{types.GodPermission{}},
			VoteThreshold:    d("0.667"),
			ProposalDuration: time.Hour * 24 * 7,
			TallyOption:      types.Deadline,
		},
		TallyDenom: "hard",
		Quorum:     d("0.4"),
	}
	var defaultID uint64 = 1
	firstBlockTime := time.Date(1998, time.January, 1, 1, 0, 0, 0, time.UTC)

	genAddrs := suite.addresses[:8]                       // Genesis accounts
	genCoinCounts := []int64{0, 0, 0, 10, 20, 30, 40, 50} // Genesis token balances

	// ---------------------- Polling information ----------------------
	//	150hard total token supply: 150 possible votes
	//  40% quroum: 60 votes required to meet quroum
	//  66.67% voting threshold: 2/3rds of votes must be Yes votes
	// -----------------------------------------------------------------

	testcases := []struct {
		name           string
		committee      types.TokenCommittee
		votes          []types.Vote
		proposalPasses bool
	}{
		{
			name:      "not enough votes to meet quroum",
			committee: tokenCom,
			votes: []types.Vote{
				{ProposalID: defaultID, Voter: genAddrs[7], VoteType: types.Yes}, // Holds 50 tokens
			},
			proposalPasses: false, // 60 vote quroum; 50 total votes; 50 yes votes. Doesn't pass 40% quroum.
		},
		{
			name:      "enough votes to meet quroum and enough Yes votes to pass voting threshold",
			committee: tokenCom,
			votes: []types.Vote{
				{ProposalID: defaultID, Voter: genAddrs[3], VoteType: types.No},  // Holds 10 tokens
				{ProposalID: defaultID, Voter: genAddrs[7], VoteType: types.Yes}, // Holds 50 tokens
			},
			proposalPasses: true, // 60 vote quroum; 60 total votes; 50 Yes votes. Passes the 66.67% voting threshold.
		},
		{
			name:      "enough votes to meet quroum via Abstain votes and enough Yes votes to pass voting threshold",
			committee: tokenCom,
			votes: []types.Vote{
				{ProposalID: defaultID, Voter: genAddrs[3], VoteType: types.Abstain}, // Holds 10 tokens
				{ProposalID: defaultID, Voter: genAddrs[7], VoteType: types.Yes},     // Holds 50 tokens
			},
			proposalPasses: true, // 60 vote quroum; 60 total votes; 50 Yes votes. Passes the 66.67% voting threshold.
		},
		{
			name:      "enough votes to meet quroum but not enough Yes votes to pass voting threshold",
			committee: tokenCom,
			votes: []types.Vote{
				{ProposalID: defaultID, Voter: genAddrs[4], VoteType: types.Yes}, // Holds 20 tokens
				{ProposalID: defaultID, Voter: genAddrs[6], VoteType: types.No},  // Holds 40 tokens
			},
			proposalPasses: false, // 60 vote quroum; 60 total votes; 20 Yes votes. Doesn't pass 66.67% voting threshold.
		},
		{
			name:      "enough votes to pass voting threshold (multiple Yes votes, multiple No votes)",
			committee: tokenCom,
			votes: []types.Vote{
				{ProposalID: defaultID, Voter: genAddrs[3], VoteType: types.Yes}, // Holds 10 tokens
				{ProposalID: defaultID, Voter: genAddrs[4], VoteType: types.Yes}, // Holds 20 tokens
				{ProposalID: defaultID, Voter: genAddrs[5], VoteType: types.Yes}, // Holds 30 tokens
				{ProposalID: defaultID, Voter: genAddrs[6], VoteType: types.No},  // Holds 40 tokens
				{ProposalID: defaultID, Voter: genAddrs[7], VoteType: types.Yes}, // Holds 50 tokens
			},
			proposalPasses: true, // 60 vote quroum; 150 total votes; 110 Yes votes. Passes the 66.67% voting threshold.
		},
		{
			name:      "not enough votes to pass voting threshold (multiple Yes votes, multiple No votes)",
			committee: tokenCom,
			votes: []types.Vote{
				{ProposalID: defaultID, Voter: genAddrs[3], VoteType: types.Yes}, // Holds 10 tokens
				{ProposalID: defaultID, Voter: genAddrs[4], VoteType: types.Yes}, // Holds 20 tokens
				{ProposalID: defaultID, Voter: genAddrs[5], VoteType: types.Yes}, // Holds 30 tokens
				{ProposalID: defaultID, Voter: genAddrs[6], VoteType: types.Yes}, // Holds 40 tokens
				{ProposalID: defaultID, Voter: genAddrs[7], VoteType: types.No},  // Holds 50 tokens
			},
			proposalPasses: false, // 60 vote quroum; 150 total votes; 100 Yes votes. Doesn't pass 66.67% voting threshold.
		},
	}

	// Convert accounts/token balances into format expected by genesis generation
	var genCoins []sdk.Coins
	var totalSupply sdk.Coins
	for _, amount := range genCoinCounts {
		userCoin := c("hard", amount)
		genCoins = append(genCoins, cs(userCoin))
		totalSupply = totalSupply.Add(userCoin)
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
						CommitteeID: tc.committee.GetID(),
						Deadline:    firstBlockTime.Add(time.Hour * 24 * 7),
					}},
					tc.votes,
				),
				supplyGenState(tApp.Codec(), totalSupply),
				app.NewAuthGenState(genAddrs, genCoins),
			)

			proposalPasses := keeper.GetTokenCommitteeProposalResult(ctx, defaultID, tc.committee)
			suite.Equal(tc.proposalPasses, proposalPasses)
		})
	}
}

func (suite *KeeperTestSuite) TestCloseProposal() {
	memberCom := types.MemberCommittee{
		BaseCommittee: types.BaseCommittee{
			ID:               12,
			Description:      "This committee is for testing.",
			Members:          suite.addresses[:5],
			Permissions:      []types.Permission{types.GodPermission{}},
			VoteThreshold:    d("0.667"),
			ProposalDuration: time.Hour * 24 * 7,
			TallyOption:      types.Deadline,
		},
	}

	var proposalID uint64 = 1
	firstBlockTime := time.Date(1998, time.January, 1, 1, 0, 0, 0, time.UTC)

	tApp := app.NewTestApp()
	keeper := tApp.GetCommitteeKeeper()
	ctx := tApp.NewContext(true, abci.Header{Height: 1, Time: firstBlockTime})

	tApp.InitializeFromGenesisStates(
		committeeGenState(
			tApp.Codec(),
			[]types.Committee{memberCom},
			[]types.Proposal{{
				PubProposal: gov.NewTextProposal("A Title", "A description of this proposal."),
				ID:          proposalID,
				CommitteeID: memberCom.GetID(),
				Deadline:    firstBlockTime.Add(time.Hour * 24 * 7),
			}},
			[]types.Vote{},
		),
	)

	// Confirm proposal exists
	proposal, found := keeper.GetProposal(ctx, proposalID)
	suite.True(found)
	// Close proposal
	keeper.CloseProposal(ctx, proposal, types.Passed)
	// Confirm proposal doesn't exist
	_, found = keeper.GetProposal(ctx, proposalID)
	suite.False(found)
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

func supplyGenState(cdc *codec.Codec, coins sdk.Coins) app.GenesisState {
	gs := supply.NewGenesisState(coins)
	return app.GenesisState{supply.ModuleName: cdc.MustMarshalJSON(gs)}
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

func (suite *KeeperTestSuite) TestProcessProposals() {

	firstBlockTime := time.Date(1998, time.January, 1, 1, 0, 0, 0, time.UTC)

	genAddrs := suite.addresses[:4]      // Genesis accounts
	genCoinCounts := []int64{1, 1, 1, 1} // Genesis token balances
	// Convert accounts/token balances into format expected by genesis generation
	var genCoins []sdk.Coins
	var totalSupply sdk.Coins
	for _, amount := range genCoinCounts {
		userCoin := c("hard", amount)
		genCoins = append(genCoins, cs(userCoin))
		totalSupply = totalSupply.Add(userCoin)
	}

	// Set up committees
	committees := []types.Committee{
		// 	1. FPTP MemberCommmittee
		types.MemberCommittee{
			BaseCommittee: types.BaseCommittee{
				ID:               1,
				Description:      "FTPT MemberCommittee",
				Members:          genAddrs,
				Permissions:      []types.Permission{types.GodPermission{}},
				VoteThreshold:    d("0.667"),
				ProposalDuration: time.Hour * 24 * 7,
				TallyOption:      types.FirstPastThePost,
			},
		},
		// 	2. FPTP TokenCommittee
		types.TokenCommittee{
			BaseCommittee: types.BaseCommittee{
				ID:               2,
				Description:      "FTPT TokenCommittee",
				Members:          genAddrs,
				Permissions:      []types.Permission{types.GodPermission{}},
				VoteThreshold:    d("0.667"),
				ProposalDuration: time.Hour * 24 * 7,
				TallyOption:      types.FirstPastThePost,
			},
			TallyDenom: "hard",
			Quorum:     d("0.30"),
		},
		// 	3. Deadline MemberCommmittee
		types.MemberCommittee{
			BaseCommittee: types.BaseCommittee{
				ID:               3,
				Description:      "Deadline MemberCommittee",
				Members:          genAddrs,
				Permissions:      []types.Permission{types.GodPermission{}},
				VoteThreshold:    d("0.667"),
				ProposalDuration: time.Hour * 24 * 7,
				TallyOption:      types.Deadline,
			},
		},
		// 	4. Deadline TokenCommittee
		types.TokenCommittee{
			BaseCommittee: types.BaseCommittee{
				ID:               4,
				Description:      "Deadline TokenCommittee",
				Members:          genAddrs,
				Permissions:      []types.Permission{types.GodPermission{}},
				VoteThreshold:    d("0.667"),
				ProposalDuration: time.Hour * 24 * 7,
				TallyOption:      types.Deadline,
			},
			TallyDenom: "hard",
			Quorum:     d("0.30"),
		},
		// 	5. PTP MemberCommmittee without permissions
		types.MemberCommittee{
			BaseCommittee: types.BaseCommittee{
				ID:               5,
				Description:      "FTPT MemberCommittee without permissions",
				Members:          genAddrs,
				Permissions:      nil,
				VoteThreshold:    d("0.667"),
				ProposalDuration: time.Hour * 24 * 7,
				TallyOption:      types.FirstPastThePost,
			},
		},
	}

	// Set up proposals that correspond 1:1 with each committee
	proposals := []types.Proposal{
		{
			ID:          1,
			CommitteeID: 1,
			PubProposal: gov.NewTextProposal("Proposal 1", "This proposal is for the FPTP MemberCommmittee."),
			Deadline:    firstBlockTime.Add(7 * 24 * time.Hour),
		},
		{
			ID:          2,
			CommitteeID: 2,
			PubProposal: gov.NewTextProposal("Proposal 2", "This proposal is for the FPTP TokenCommittee."),
			Deadline:    firstBlockTime.Add(7 * 24 * time.Hour),
		},
		{
			ID:          3,
			CommitteeID: 3,
			PubProposal: gov.NewTextProposal("Proposal 3", "This proposal is for the Deadline MemberCommmittee."),
			Deadline:    firstBlockTime.Add(7 * 24 * time.Hour),
		},
		{
			ID:          4,
			CommitteeID: 4,
			PubProposal: gov.NewTextProposal("Proposal 4", "This proposal is for the Deadline TokenCommittee."),
			Deadline:    firstBlockTime.Add(7 * 24 * time.Hour),
		},
		{
			ID:          5,
			CommitteeID: 5,
			PubProposal: gov.NewTextProposal("Proposal 5", "This proposal is for the FPTP MemberCommmittee without permissions."),
			Deadline:    firstBlockTime.Add(7 * 24 * time.Hour),
		},
	}

	// Each test case targets 1 committee/proposal via targeted votes
	testcases := []struct {
		name                             string
		ID                               uint64
		votes                            []types.Vote
		expectedToCompleteBeforeDeadline bool
		expectedOutcome                  types.ProposalOutcome
	}{
		{
			name: "FPTP MemberCommittee proposal does not have enough votes to pass",
			ID:   1,
			votes: []types.Vote{
				{ProposalID: 1, Voter: genAddrs[0], VoteType: types.Yes},
			},
			expectedToCompleteBeforeDeadline: false,
			expectedOutcome:                  types.Failed,
		},
		{
			name: "FPTP MemberCommittee proposal has enough votes to pass before deadline",
			ID:   1,
			votes: []types.Vote{
				{ProposalID: 1, Voter: genAddrs[0], VoteType: types.Yes},
				{ProposalID: 1, Voter: genAddrs[1], VoteType: types.Yes},
				{ProposalID: 1, Voter: genAddrs[2], VoteType: types.Yes},
			},
			expectedToCompleteBeforeDeadline: true,
			expectedOutcome:                  types.Passed,
		},
		{
			name: "FPTP TokenCommittee proposal does not have enough votes to pass",
			ID:   2,
			votes: []types.Vote{
				{ProposalID: 2, Voter: genAddrs[0], VoteType: types.Yes},
			},
			expectedToCompleteBeforeDeadline: false,
			expectedOutcome:                  types.Failed,
		},
		{
			name: "FPTP TokenCommittee proposal has enough votes to pass before deadline",
			ID:   2,
			votes: []types.Vote{
				{ProposalID: 2, Voter: genAddrs[0], VoteType: types.Yes},
				{ProposalID: 2, Voter: genAddrs[1], VoteType: types.Yes},
				{ProposalID: 2, Voter: genAddrs[2], VoteType: types.Yes},
			},
			expectedToCompleteBeforeDeadline: true,
			expectedOutcome:                  types.Passed,
		},
		{
			name: "Deadline MemberCommittee proposal with enough votes to pass only passes after deadline",
			ID:   3,
			votes: []types.Vote{
				{ProposalID: 3, Voter: genAddrs[0], VoteType: types.Yes},
				{ProposalID: 3, Voter: genAddrs[1], VoteType: types.Yes},
				{ProposalID: 3, Voter: genAddrs[2], VoteType: types.Yes},
			},
			expectedOutcome: types.Passed,
		},
		{
			name: "Deadline MemberCommittee proposal doesn't have enough votes to pass",
			ID:   3,
			votes: []types.Vote{
				{ProposalID: 3, Voter: genAddrs[0], VoteType: types.Yes},
			},
			expectedOutcome: types.Failed,
		},
		{
			name: "Deadline TokenCommittee proposal with enough votes to pass only passes after deadline",
			ID:   4,
			votes: []types.Vote{
				{ProposalID: 4, Voter: genAddrs[0], VoteType: types.Yes},
				{ProposalID: 4, Voter: genAddrs[1], VoteType: types.Yes},
				{ProposalID: 4, Voter: genAddrs[2], VoteType: types.Yes},
			},
			expectedOutcome: types.Passed,
		},
		{
			name: "Deadline TokenCommittee proposal doesn't have enough votes to pass",
			ID:   4,
			votes: []types.Vote{
				{ProposalID: 4, Voter: genAddrs[0], VoteType: types.Yes},
			},
			expectedOutcome: types.Failed,
		},
		{
			name: "FPTP MemberCommittee doesn't have permissions to enact passed proposal",
			ID:   5,
			votes: []types.Vote{
				{ProposalID: 5, Voter: genAddrs[0], VoteType: types.Yes},
				{ProposalID: 5, Voter: genAddrs[1], VoteType: types.Yes},
				{ProposalID: 5, Voter: genAddrs[2], VoteType: types.Yes},
			},
			expectedToCompleteBeforeDeadline: true,
			expectedOutcome:                  types.Invalid,
		},
	}

	for _, tc := range testcases {
		suite.Run(tc.name, func() {
			// Create local testApp because suite doesn't run the SetupTest function for subtests
			tApp := app.NewTestApp()
			keeper := tApp.GetCommitteeKeeper()
			ctx := tApp.NewContext(true, abci.Header{Height: 1, Time: firstBlockTime})

			// Initialize all committees, proposals, and votes via Genesis
			tApp.InitializeFromGenesisStates(
				committeeGenState(tApp.Codec(), committees, proposals, tc.votes),
				supplyGenState(tApp.Codec(), totalSupply),
				app.NewAuthGenState(genAddrs, genCoins),
			)

			// Load committee from the store
			committee, found := keeper.GetCommittee(ctx, tc.ID)
			suite.True(found)

			// Process proposals
			ctx = ctx.WithBlockTime(firstBlockTime)
			keeper.ProcessProposals(ctx)

			// Fetch proposal and votes from the store
			votes := getProposalVoteMap(keeper, ctx)
			proposal, found := keeper.GetProposal(ctx, tc.ID)

			if committee.GetTallyOption() == types.FirstPastThePost {
				if tc.expectedToCompleteBeforeDeadline {
					suite.False(found)
					suite.Empty(votes[tc.ID])

					// Check proposal outcome
					outcome, err := getProposalOutcome(tc.ID, ctx.EventManager().Events(), tApp.Codec())
					suite.NoError(err)
					suite.Equal(tc.expectedOutcome, outcome)
					return
				} else {
					suite.True(found)
					suite.NotEmpty(votes[tc.ID])
				}
			}

			// Move block time to deadline
			ctx = ctx.WithBlockTime(proposal.Deadline)
			keeper.ProcessProposals(ctx)

			// Fetch proposal and votes from the store
			votes = getProposalVoteMap(keeper, ctx)
			proposal, found = keeper.GetProposal(ctx, tc.ID)
			suite.False(found)
			suite.Empty(votes[proposal.ID])

			// Check proposal outcome
			outcome, err := getProposalOutcome(tc.ID, ctx.EventManager().Events(), tApp.Codec())
			suite.NoError(err)
			suite.Equal(tc.expectedOutcome, outcome)
		})
	}
}

// getProposalOutcome checks the outcome of a proposal via a `proposal_close` event whose `proposal_id`
// matches argument proposalID
func getProposalOutcome(proposalID uint64, events sdk.Events, cdc *amino.Codec) (types.ProposalOutcome, error) {
	// Marshal proposal ID to match against event attribute
	x, _ := cdc.MarshalJSON(proposalID)
	marshaledID := x[1 : len(x)-1]

	for _, event := range events {
		if event.Type == types.EventTypeProposalClose {
			var proposalOutcome types.ProposalOutcome
			correctProposal := false
			for _, attribute := range event.Attributes {
				// Only get outcome of specific proposal
				if bytes.Compare(attribute.GetKey(), []byte("proposal_id")) == 0 {
					if bytes.Compare(attribute.GetValue(), marshaledID) == 0 {
						correctProposal = true
					}
				}
				// Match event attribute bytes to marshaled outcome
				if bytes.Compare(attribute.GetKey(), []byte(types.AttributeKeyProposalOutcome)) == 0 {
					outcome, err := types.MatchMarshaledOutcome(attribute.GetValue(), cdc)
					if err != nil {
						return -1, err
					}
					proposalOutcome = outcome
				}
			}
			// If this is the desired proposal, return the outcome
			if correctProposal {
				return proposalOutcome, nil
			}
		}
	}
	return -1, nil
}
