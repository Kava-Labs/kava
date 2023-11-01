package e2e_test

import (
	"context"
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	query "github.com/cosmos/cosmos-sdk/types/query"
	"github.com/cosmos/cosmos-sdk/x/authz"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"

	"github.com/kava-labs/kava/tests/e2e/testutil"
	"github.com/kava-labs/kava/tests/util"
)

const (
	govModuleAcc       = "kava10d07y265gmmuvt4z0w9aw880jnsr700jxh8cq5"
	communityModuleAcc = "kava17d2wax0zhjrrecvaszuyxdf5wcu5a0p4qlx3t5"
	kavadistModuleAcc  = "kava1cj7njkw2g9fqx4e768zc75dp9sks8u9znxrf0w"
)

func (suite *IntegrationTestSuite) TestGovParamChanges() {
	suite.SkipIfUpgradeDisabled()

	beforeUpgradeCtx := util.CtxAtHeight(suite.UpgradeHeight - 1)
	afterUpgradeCtx := util.CtxAtHeight(suite.UpgradeHeight)

	// fetch gov parameters before upgrade
	govBeforeParams, err := suite.Kava.Gov.Params(beforeUpgradeCtx, &govv1.QueryParamsRequest{ParamsType: "tallying"})
	suite.Require().NoError(err)

	// assert expected gov quorum before upgrade
	suite.NotEqual(govBeforeParams.TallyParams.Quorum, "0.200000000000000000")

	govAfterParams, err := suite.Kava.Gov.Params(afterUpgradeCtx, &govv1.QueryParamsRequest{ParamsType: "tallying"})
	suite.Require().NoError(err)

	// assert expected gov quorum after upgrade
	suite.Equal(govAfterParams.TallyParams.Quorum, "0.200000000000000000")

}

func (suite *IntegrationTestSuite) TestAuthzParamChanges() {
	suite.SkipIfUpgradeDisabled()

	beforeUpgradeCtx := util.CtxAtHeight(suite.UpgradeHeight - 1)
	afterUpgradeCtx := util.CtxAtHeight(suite.UpgradeHeight)

	// fetch authz grants before upgrade
	authzBeforeGrants, err := suite.Kava.Authz.Grants(beforeUpgradeCtx, &authz.QueryGrantsRequest{Granter: kavadistModuleAcc, Grantee: govModuleAcc, Pagination: &query.PageRequest{Limit: 1000, CountTotal: true}})
	suite.Require().NoError(err)
	suite.Require().Equal(authzBeforeGrants.Pagination.Total, uint64(len(authzBeforeGrants.Grants)), "expected all grants to have been requested")

	// no kavadist -> gov grants
	suite.Equal(0, len(authzBeforeGrants.Grants))

	// fetch authz grants after upgrade
	authzAfterGrants, err := suite.Kava.Authz.Grants(afterUpgradeCtx, &authz.QueryGrantsRequest{Granter: kavadistModuleAcc, Grantee: govModuleAcc, Pagination: &query.PageRequest{Limit: 1000, CountTotal: true}})
	suite.Require().NoError(err)
	suite.Require().Equal(authzAfterGrants.Pagination.Total, uint64(len(authzAfterGrants.Grants)), "expected all grants to have been requested")

	// one kavadist -> gov grants
	suite.Require().Equal(1, len(authzAfterGrants.Grants))

	grant := authzAfterGrants.Grants[0]

	var authorization authz.Authorization
	suite.Kava.EncodingConfig.Marshaler.UnpackAny(grant.Authorization, &authorization)

	genericAuthorization, ok := authorization.(*authz.GenericAuthorization)
	suite.Require().True(ok, "expected generic authorization")

	// kavadist allows gov to MsgSend it's funds
	suite.Equal(sdk.MsgTypeURL(&banktypes.MsgSend{}), genericAuthorization.Msg)
	// no expiration
	var expectedExpiration *time.Time
	suite.Equal(expectedExpiration, grant.Expiration)
}

func (suite *IntegrationTestSuite) TestModuleAccountGovTransfers() {
	suite.SkipIfUpgradeDisabled()
	suite.SkipIfKvtoolDisabled()

	// the module account (authority) that executes the transfers
	govAcc := sdk.MustAccAddressFromBech32(govModuleAcc)

	// module accounts for gov transfer test cases
	communityAcc := sdk.MustAccAddressFromBech32(communityModuleAcc)
	kavadistAcc := sdk.MustAccAddressFromBech32(kavadistModuleAcc)

	testCases := []struct {
		name     string
		sender   sdk.AccAddress
		receiver sdk.AccAddress
		amount   sdk.Coin
	}{
		{
			name:     "transfer from community to kavadist for incentive rewards",
			sender:   communityAcc,
			receiver: kavadistAcc,
			amount:   ukava(100e6),
		},
		{
			name:     "transfer from kavadist to community",
			sender:   kavadistAcc,
			receiver: communityAcc,
			amount:   ukava(50e6),
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			// create msg exec for transfer between modules
			msg := banktypes.NewMsgSend(
				tc.sender,
				tc.receiver,
				sdk.NewCoins(tc.amount),
			)
			execMsg := authz.NewMsgExec(govAcc, []sdk.Msg{msg})

			// ensure proposal passes
			passBlock := suite.submitAndPassProposal([]sdk.Msg{&execMsg})
			transfers := suite.getBankTransferAmountAtBlock(passBlock, tc.sender, tc.receiver)

			suite.Require().Containsf(
				transfers,
				tc.amount,
				"expected transfer of %s to be included in bank transfer events: %s",
				tc.amount,
				transfers,
			)
		})
	}
}

func (suite *IntegrationTestSuite) submitAndPassProposal(msgs []sdk.Msg) int64 {
	govParamsRes, err := suite.Kava.Gov.Params(context.Background(), &govv1.QueryParamsRequest{
		ParamsType: govv1.ParamDeposit,
	})
	suite.NoError(err)

	kavaAcc := suite.Kava.GetAccount(testutil.FundedAccountName)

	proposalMsg, err := govv1.NewMsgSubmitProposal(
		msgs,
		govParamsRes.DepositParams.MinDeposit,
		kavaAcc.SdkAddress.String(),
		"",
	)
	suite.NoError(err)

	gasLimit := 1e6
	fee := ukava(1000)

	req := util.KavaMsgRequest{
		Msgs:      []sdk.Msg{proposalMsg},
		GasLimit:  uint64(gasLimit),
		FeeAmount: sdk.NewCoins(fee),
		Memo:      "this is a proposal please accept me",
	}
	res := kavaAcc.SignAndBroadcastKavaTx(req)
	suite.Require().NoError(res.Err)

	// Wait for proposal to be submitted
	txRes, err := util.WaitForSdkTxCommit(suite.Kava.Tx, res.Result.TxHash, 6*time.Second)
	suite.Require().NoError(err)

	var govRes govv1.MsgSubmitProposalResponse
	suite.decodeTxMsgResponse(txRes, &govRes)

	// 2. Vote for proposal from whale account
	whale := suite.Kava.GetAccount(testutil.FundedAccountName)
	voteMsg := govv1.NewMsgVote(
		whale.SdkAddress,
		govRes.ProposalId,
		govv1.OptionYes,
		"",
	)

	voteReq := util.KavaMsgRequest{
		Msgs:      []sdk.Msg{voteMsg},
		GasLimit:  uint64(gasLimit),
		FeeAmount: sdk.NewCoins(fee),
		Memo:      "voting",
	}
	voteRes := whale.SignAndBroadcastKavaTx(voteReq)
	suite.Require().NoError(voteRes.Err)

	_, err = util.WaitForSdkTxCommit(suite.Kava.Tx, voteRes.Result.TxHash, 6*time.Second)
	suite.Require().NoError(err)

	// 3. Wait until proposal passes
	suite.Require().Eventually(func() bool {
		proposalRes, err := suite.Kava.Gov.Proposal(context.Background(), &govv1.QueryProposalRequest{
			ProposalId: govRes.ProposalId,
		})
		suite.NoError(err)

		switch status := proposalRes.Proposal.Status; status {
		case govv1.StatusDepositPeriod, govv1.StatusVotingPeriod:
			return false
		case govv1.StatusPassed:
			return true
		case govv1.StatusFailed, govv1.StatusRejected:
			suite.Failf("proposal failed", "proposal failed with status %s", status.String())
			return true
		}

		return false
	}, 60*time.Second, 1*time.Second)

	page := 1
	perPage := 100

	// Get the block the proposal was passed in
	passBlock, err := suite.Kava.TmSignClient.BlockSearch(
		context.Background(),
		fmt.Sprintf(
			"active_proposal.proposal_result = 'proposal_passed' AND active_proposal.proposal_id = %d",
			govRes.ProposalId,
		),
		&page,
		&perPage,
		"asc",
	)
	suite.Require().NoError(err)
	suite.Require().Equal(1, len(passBlock.Blocks), "passed proposal should be searchable")

	return passBlock.Blocks[len(passBlock.Blocks)-1].Block.Height
}

// getBankTransferAmountAtBlock returns the amount of coins transferred between
// the given accounts in the block at the given height. Note that this returns
// a slice of sdk.Coin that can contain multiple coins of the SAME denom -- ie. NOT sdk.Coins
func (suite *IntegrationTestSuite) getBankTransferAmountAtBlock(
	blockHeight int64,
	sender sdk.AccAddress,
	receiver sdk.AccAddress,
) []sdk.Coin {
	// Fetch block results for paid staking rewards in the block
	blockRes, err := suite.Kava.TmSignClient.BlockResults(
		context.Background(),
		&blockHeight,
	)
	suite.Require().NoError(err)

	transferEvents := util.FilterEventsByType(
		blockRes.EndBlockEvents, // gov proposals applied in EndBlocker
		banktypes.EventTypeTransfer,
	)
	suite.Require().NotEmpty(transferEvents, "there should be at least 1 bank transfer event")

	transfers := []sdk.Coin{}

event:
	for _, event := range transferEvents {
		if event.Type != banktypes.EventTypeTransfer {
			suite.FailNowf(
				"unexpected event type %s in block results",
				event.Type,
			)
		}

		for _, attr := range event.Attributes {
			suite.T().Logf("event attr: %s = %s", string(attr.Key), string(attr.Value))

			if string(attr.Key) == banktypes.AttributeKeyRecipient {
				if string(attr.Value) != receiver.String() {
					continue event
				}
			}

			if string(attr.Key) == banktypes.AttributeKeySender {
				if string(attr.Value) != sender.String() {
					continue event
				}
			}

			if string(attr.Key) == sdk.AttributeKeyAmount {
				amount, err := sdk.ParseCoinNormalized(string(attr.Value))
				suite.Require().NoError(err)

				transfers = append(transfers, amount)
			}
		}
	}

	return transfers
}
