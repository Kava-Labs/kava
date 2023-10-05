package e2e_test

import (
	"context"
	"encoding/hex"
	"time"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"

	"github.com/kava-labs/kava/tests/e2e/testutil"
	"github.com/kava-labs/kava/tests/util"
	communitytypes "github.com/kava-labs/kava/x/community/types"
)

func (suite *IntegrationTestSuite) TestCommunityUpdateParams_NonAuthority() {
	// ARRANGE
	// setup kava account
	funds := ukava(1e5) // .1 KAVA
	kavaAcc := suite.Kava.NewFundedAccount("community-non-authority", sdk.NewCoins(funds))

	gasLimit := int64(2e5)
	fee := ukava(200)

	msg := communitytypes.NewMsgUpdateParams(
		kavaAcc.SdkAddress,
		communitytypes.DefaultParams(),
	)

	// ACT
	req := util.KavaMsgRequest{
		Msgs:      []sdk.Msg{&msg},
		GasLimit:  uint64(gasLimit),
		FeeAmount: sdk.NewCoins(fee),
		Memo:      "this is a failure!",
	}
	res := kavaAcc.SignAndBroadcastKavaTx(req)

	// ASSERT
	_, err := util.WaitForSdkTxCommit(suite.Kava.Tx, res.Result.TxHash, 6*time.Second)
	suite.Error(err)
}

func (suite *IntegrationTestSuite) TestCommunityUpdateParams_Authority() {
	// ARRANGE
	govParamsRes, err := suite.Kava.Gov.Params(context.Background(), &govv1.QueryParamsRequest{
		ParamsType: govv1.ParamDeposit,
	})
	suite.NoError(err)

	// Check initial params
	communityParamsResInitial, err := suite.Kava.Community.Params(
		context.Background(),
		&communitytypes.QueryParamsRequest{},
	)
	suite.Require().NoError(err)

	suite.T().Logf("initial params: %v", communityParamsResInitial.Params)

	// setup kava account
	// .1 KAVA + min deposit amount for proposal
	funds := sdk.NewCoins(ukava(1e5)).Add(govParamsRes.DepositParams.MinDeposit...)
	kavaAcc := suite.Kava.NewFundedAccount("community-update-params", funds)

	gasLimit := int64(2e5)
	fee := ukava(200)

	upgradeTime := time.Now().Add(24 * time.Hour).UTC()

	updateParamsMsg := communitytypes.NewMsgUpdateParams(
		authtypes.NewModuleAddress(govtypes.ModuleName), // authority
		communitytypes.NewParams(
			upgradeTime,
			sdkmath.LegacyNewDec(1111), // stakingRewardsPerSecond
			sdkmath.LegacyNewDec(2222), // upgradeTimeSetstakingRewardsPerSecond
		),
	)

	proposalMsg, err := govv1.NewMsgSubmitProposal(
		[]sdk.Msg{&updateParamsMsg},
		govParamsRes.DepositParams.MinDeposit,
		kavaAcc.SdkAddress.String(),
		"community-update-params",
	)
	suite.NoError(err)

	// ACT
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

	// Parse tx response to get proposal id
	var govRes govv1.MsgSubmitProposalResponse
	suite.decodeTxMsgResponse(txRes, &govRes)

	// Vote for proposal from whale account
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

	// Fetch proposal
	proposalRes, err := suite.Kava.Gov.Proposal(context.Background(), &govv1.QueryProposalRequest{
		ProposalId: govRes.ProposalId,
	})
	suite.NoError(err)

	suite.T().Logf("proposal status: %v", proposalRes.Proposal.Status)
	suite.T().Logf("proposal ending: %v", proposalRes.Proposal.VotingEndTime)

	// Wait until proposal passes
	suite.Require().Eventually(func() bool {
		proposalRes, err := suite.Kava.Gov.Proposal(context.Background(), &govv1.QueryProposalRequest{
			ProposalId: govRes.ProposalId,
		})
		suite.NoError(err)

		return proposalRes.Proposal.Status == govv1.StatusPassed
	}, 60*time.Second, 1*time.Second)

	// Check parameters are updated
	communityParamsRes, err := suite.Kava.Community.Params(
		context.Background(),
		&communitytypes.QueryParamsRequest{},
	)
	suite.Require().NoError(err)

	suite.Equal(updateParamsMsg.Params, communityParamsRes.Params)

	suite.T().Logf("new params: %v", communityParamsRes.Params)
}

func (suite *IntegrationTestSuite) decodeTxMsgResponse(txRes *sdk.TxResponse, ptr codec.ProtoMarshaler) {
	// convert txRes.Data hex string to bytes
	txResBytes, err := hex.DecodeString(txRes.Data)
	suite.Require().NoError(err)

	// Unmarshal data to TxMsgData
	var txMsgData sdk.TxMsgData
	suite.Kava.EncodingConfig.Marshaler.MustUnmarshal(txResBytes, &txMsgData)
	suite.T().Logf("txData.MsgResponses: %v", txMsgData.MsgResponses)

	// Parse MsgResponse
	suite.Kava.EncodingConfig.Marshaler.MustUnmarshal(txMsgData.MsgResponses[0].Value, ptr)
	suite.Require().NoError(err)
}
