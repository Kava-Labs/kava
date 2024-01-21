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
	_, err := util.WaitForSdkTxCommit(suite.Kava.Grpc.Query.Tx, res.Result.TxHash, 6*time.Second)
	suite.Require().Error(err)
	suite.Require().ErrorContains(
		err,
		govtypes.ErrInvalidSigner.Error(),
		"should return with authority check error",
	)
}

func (suite *IntegrationTestSuite) TestCommunityUpdateParams_Authority() {
	// ARRANGE
	govParamsRes, err := suite.Kava.Grpc.Query.Gov.Params(context.Background(), &govv1.QueryParamsRequest{
		ParamsType: govv1.ParamDeposit,
	})
	suite.NoError(err)

	// Check initial params
	communityParamsResInitial, err := suite.Kava.Grpc.Query.Community.Params(
		context.Background(),
		&communitytypes.QueryParamsRequest{},
	)
	suite.Require().NoError(err)

	// setup kava account
	// .1 KAVA + min deposit amount for proposal
	funds := sdk.NewCoins(ukava(1e5)).Add(govParamsRes.DepositParams.MinDeposit...)
	kavaAcc := suite.Kava.NewFundedAccount("community-update-params", funds)

	gasLimit := int64(2e5)
	fee := ukava(200)

	// Wait until switchover actually happens - When testing without the upgrade
	// handler that sets a relative switchover time, the switchover time in
	// genesis should be set in the past so it runs immediately.
	suite.Require().Eventually(
		func() bool {
			params, err := suite.Kava.Grpc.Query.Community.Params(
				context.Background(),
				&communitytypes.QueryParamsRequest{},
			)
			suite.Require().NoError(err)

			return params.Params.UpgradeTimeDisableInflation.Equal(time.Time{})
		},
		20*time.Second,
		1*time.Second,
		"switchover should happen",
	)

	// Add 1 to the staking rewards per second
	newStakingRewardsPerSecond := communityParamsResInitial.Params.
		StakingRewardsPerSecond.
		Add(sdkmath.LegacyNewDec(1))

	// 1. Proposal
	// Only modify stakingRewardsPerSecond, as to not re-run the switchover and
	// to not influence other tests
	updateParamsMsg := communitytypes.NewMsgUpdateParams(
		authtypes.NewModuleAddress(govtypes.ModuleName), // authority
		communitytypes.NewParams(
			time.Time{},                // after switchover, is empty
			newStakingRewardsPerSecond, // only modify stakingRewardsPerSecond
			communityParamsResInitial.Params.UpgradeTimeSetStakingRewardsPerSecond,
		),
	)

	// Make sure we're actually changing the params
	suite.NotEqual(
		updateParamsMsg.Params,
		communityParamsResInitial.Params,
		"new params should be different from existing",
	)

	proposalMsg, err := govv1.NewMsgSubmitProposal(
		[]sdk.Msg{&updateParamsMsg},
		govParamsRes.Params.MinDeposit,
		kavaAcc.SdkAddress.String(),
		"community-update-params",
		"title",
		"summary",
	)
	suite.NoError(err)

	req := util.KavaMsgRequest{
		Msgs:      []sdk.Msg{proposalMsg},
		GasLimit:  uint64(gasLimit),
		FeeAmount: sdk.NewCoins(fee),
		Memo:      "this is a proposal please accept me",
	}
	res := kavaAcc.SignAndBroadcastKavaTx(req)
	suite.Require().NoError(res.Err)

	// Wait for proposal to be submitted
	txRes, err := util.WaitForSdkTxCommit(suite.Kava.Grpc.Query.Tx, res.Result.TxHash, 6*time.Second)
	suite.Require().NoError(err)

	// Parse tx response to get proposal id
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

	_, err = util.WaitForSdkTxCommit(suite.Kava.Grpc.Query.Tx, voteRes.Result.TxHash, 6*time.Second)
	suite.Require().NoError(err)

	// 3. Wait until proposal passes
	suite.Require().Eventually(func() bool {
		proposalRes, err := suite.Kava.Grpc.Query.Gov.Proposal(context.Background(), &govv1.QueryProposalRequest{
			ProposalId: govRes.ProposalId,
		})
		suite.NoError(err)

		return proposalRes.Proposal.Status == govv1.StatusPassed
	}, 60*time.Second, 1*time.Second)

	// Check parameters are updated
	communityParamsRes, err := suite.Kava.Grpc.Query.Community.Params(
		context.Background(),
		&communitytypes.QueryParamsRequest{},
	)
	suite.Require().NoError(err)

	suite.Equal(updateParamsMsg.Params, communityParamsRes.Params)
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
