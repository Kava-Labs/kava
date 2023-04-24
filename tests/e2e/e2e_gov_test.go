package e2e_test

import (
	"context"
	"fmt"
	"strconv"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	txtypes "github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/cosmos/cosmos-sdk/x/authz"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govtypesv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/tests/e2e/testutil"
	"github.com/kava-labs/kava/tests/util"
	cdptypes "github.com/kava-labs/kava/x/cdp/types"
	communitytypes "github.com/kava-labs/kava/x/community/types"
)

func (suite IntegrationTestSuite) TestGovAuthzMessages() {
	suite.SkipIfUpgradeDisabled()

	communityAcc := suite.Kava.QuerySdkForModuleAccount(communitytypes.ModuleName).GetAddress()
	govAccount := suite.Kava.QuerySdkForModuleAccount(govtypes.ModuleName).GetAddress()
	randomUserAddr := app.RandomAddress()

	// fund community pool since it might not have enough funds accrued yet
	whale := suite.Kava.GetAccount(testutil.FundedAccountName)
	msgRequest := util.KavaMsgRequest{
		Msgs: []sdk.Msg{
			banktypes.NewMsgSend(whale.SdkAddress, communityAcc, sdk.NewCoins(ukava(120e6))),
		},
		GasLimit:  uint64(2e5),
		FeeAmount: sdk.NewCoins(ukava(7500)),
	}
	whale.SignAndBroadcastKavaTx(msgRequest)

	// open cdp & send some community pool funds to another user
	createCdp := cdptypes.NewMsgCreateCDP(
		communityAcc,
		ukava(100e6),
		sdk.NewCoin("usdx", sdkmath.NewInt(80e6)),
		"ukava-a",
	)
	execMsg := authz.NewMsgExec(govAccount, []sdk.Msg{
		banktypes.NewMsgSend(communityAcc, randomUserAddr, sdk.NewCoins(ukava(5e6))),
		&createCdp,
	})
	suite.submitAndPassProposal([]sdk.Msg{&execMsg})

	// validate usdx position
	suite.Eventually(func() bool {
		coins := suite.Kava.QuerySdkForBalances(communityAcc)
		return coins.AmountOf("usdx").Equal(sdkmath.NewInt(80e6))
	}, 10*time.Second, 1*time.Second)

	// validate user funds are sent
	coins := suite.Kava.QuerySdkForBalances(randomUserAddr)
	suite.True(coins.AmountOf("ukava").Equal(sdkmath.NewInt(5e6)))

	// draw debt
	drawDebt := cdptypes.NewMsgDrawDebt(
		communityAcc,
		"ukava-a",
		sdk.NewCoin("usdx", sdkmath.NewInt(20e6)),
	)
	execMsg = authz.NewMsgExec(govAccount, []sdk.Msg{&drawDebt})
	suite.submitAndPassProposal([]sdk.Msg{&execMsg})

	suite.Eventually(func() bool {
		coins := suite.Kava.QuerySdkForBalances(communityAcc)
		return coins.AmountOf("usdx").Equal(sdkmath.NewInt(100e6))
	}, 10*time.Second, 1*time.Second)

	// repay debt
	repayDebt := cdptypes.NewMsgRepayDebt(
		communityAcc,
		"ukava-a",
		sdk.NewCoin("usdx", sdkmath.NewInt(30e6)),
	)
	execMsg = authz.NewMsgExec(govAccount, []sdk.Msg{&repayDebt})
	suite.submitAndPassProposal([]sdk.Msg{&execMsg})

	suite.Eventually(func() bool {
		coins := suite.Kava.QuerySdkForBalances(communityAcc)
		return coins.AmountOf("usdx").Equal(sdkmath.NewInt(70e6))
	}, 10*time.Second, 1*time.Second)
}

func (suite IntegrationTestSuite) submitAndPassProposal(msgs []sdk.Msg) uint64 {
	whale := suite.Kava.GetAccount(testutil.FundedAccountName)

	// submit proposal
	proposal, err := govtypesv1.NewMsgSubmitProposal(
		msgs,
		sdk.NewCoins(ukava(100)),
		whale.SdkAddress.String(),
		"",
	)
	suite.NoError(err)

	msgRequest := util.KavaMsgRequest{
		Msgs:      []sdk.Msg{proposal},
		GasLimit:  uint64(3e5),
		FeeAmount: sdk.NewCoins(ukava(7500)),
	}
	res := whale.SignAndBroadcastKavaTx(msgRequest)
	var events sdk.StringEvents

	// wait until the proposal is committed
	suite.Eventually(func() bool {
		txRes, _ := suite.Kava.Tx.GetTx(
			context.Background(),
			&txtypes.GetTxRequest{Hash: res.Result.TxHash},
		)
		fmt.Printf("txRes: %v", txRes)
		if txRes.TxResponse.Code == 0 {
			events = txRes.TxResponse.Logs[0].Events
			return true
		}
		return false
	}, 15*time.Second, 1*time.Second)

	var proposalId uint64
	for _, event := range events {
		for _, attr := range event.Attributes {
			if attr.Key == "proposal_id" {
				proposalId, err = strconv.ParseUint(attr.Value, 10, 64)
				suite.NoError(err)
			}
		}
	}
	suite.NotEqual(uint64(0), proposalId, "proposal id should not be 0")

	// Vote for proposal
	vote := govtypesv1.NewMsgVote(whale.SdkAddress, proposalId, govtypesv1.VoteOption_VOTE_OPTION_YES, "")
	msgRequest = util.KavaMsgRequest{
		Msgs:      []sdk.Msg{vote},
		GasLimit:  uint64(3e5),
		FeeAmount: sdk.NewCoins(ukava(7500)),
	}
	whale.SignAndBroadcastKavaTx(msgRequest)

	// Wait for proposal to pass
	passedId, err := util.WaitForProposalStatus(
		suite.Kava.Gov, proposalId, govtypesv1.ProposalStatus_PROPOSAL_STATUS_PASSED, 10*time.Second,
	)
	suite.NoError(err)

	return passedId
}
