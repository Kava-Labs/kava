package e2e_test

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	ibctransfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	ibcchanneltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"

	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

func (suite *IntegrationTestSuite) getEscrowAccountBalances(
	ctx context.Context,
) (map[string]sdk.Coins, sdk.Coins) {
	suite.T().Helper()

	grpcClient := suite.Kava.Grpc

	channelsRes, err := grpcClient.Query.IbcChannel.Channels(
		ctx,
		&ibcchanneltypes.QueryChannelsRequest{
			Pagination: &query.PageRequest{
				Limit: 1000,
			},
		},
	)
	suite.Require().NoError(err)

	suite.T().Logf("found %d IBC channels", len(channelsRes.Channels))

	escrowBals := make(map[string]sdk.Coins)
	totalEscrowBal := sdk.Coins{}

	for _, channel := range channelsRes.Channels {
		escrowAddress, err := grpcClient.Query.IbcTransfer.EscrowAddress(
			ctx,
			&ibctransfertypes.QueryEscrowAddressRequest{
				ChannelId: channel.ChannelId,
				PortId:    channel.PortId,
			},
		)
		suite.Require().NoError(err)

		// Check if escrow addr already exists in map
		if _, ok := escrowBals[escrowAddress.String()]; ok {
			suite.T().Logf("escrow address %s already exists in map", escrowAddress.String())
			continue
		}

		escrowBalances, err := grpcClient.Query.Bank.AllBalances(ctx, &banktypes.QueryAllBalancesRequest{
			Address: escrowAddress.EscrowAddress,
		})
		suite.Require().NoError(err)

		escrowBals[escrowAddress.String()] = escrowBalances.Balances
		totalEscrowBal = totalEscrowBal.Add(escrowBalances.Balances...)
	}

	return escrowBals, totalEscrowBal
}

func (suite *IntegrationTestSuite) getEscrowStateBalances(
	ctx context.Context,
	totalBankEscrow sdk.Coins,
) sdk.Coins {
	grpcClient := suite.Kava.Grpc
	escrowState := sdk.Coins{}

	for _, escrowCoin := range totalBankEscrow {
		escrowStateRes, err := grpcClient.Query.IbcTransfer.TotalEscrowForDenom(
			ctx,
			&ibctransfertypes.QueryTotalEscrowForDenomRequest{
				Denom: escrowCoin.Denom,
			},
		)
		suite.Require().NoError(err)

		escrowState = escrowState.Add(escrowStateRes.Amount)
	}

	return escrowState
}

func (suite *IntegrationTestSuite) TestPfmUpgrade() {
	suite.SkipIfUpgradeDisabled()

	// Ensure pre-upgrade version has the state we want to test from

	beforeUpgradeCtx := suite.Kava.Grpc.CtxAtHeight(suite.UpgradeHeight - 1)
	afterUpgradeCtx := suite.Kava.Grpc.CtxAtHeight(suite.UpgradeHeight)

	// Balances BANK
	escrowBankBalsBefore, totalBankEscrowBefore := suite.getEscrowAccountBalances(beforeUpgradeCtx)

	suite.T().Logf("totalBankEscrowBefore: %s", totalBankEscrowBefore)
	suite.T().Logf("escrowBankBalsBefore: %s", escrowBankBalsBefore)

	// Balances in escrow STATE, not bank
	escrowStateBefore := suite.getEscrowStateBalances(beforeUpgradeCtx, totalBankEscrowBefore)
	suite.T().Logf("escrowStateBefore: %s", escrowStateBefore)

	// Pre-upgrade, escrow bank balances should NOT match escrow state
	suite.Require().NotEqual(escrowStateBefore, totalBankEscrowBefore, "escrow mismatch before upgrade")

	// ------------------------------------------------------------------------
	// Check again after upgrade
	// Migration should have run to re-sync the escrow state with the escrow bank balances
	escrowBankBalsAfter, totalBankEscrowAfter := suite.getEscrowAccountBalances(afterUpgradeCtx)

	suite.T().Logf("totalBankEscrowAfter: %s", totalBankEscrowAfter)
	suite.T().Logf("escrowBankBalsAfter: %s", escrowBankBalsAfter)

	// Balances in escrow STATE, not bank
	escrowStateAfter := suite.getEscrowStateBalances(afterUpgradeCtx, totalBankEscrowAfter)
	suite.T().Logf("escrowStateAfter: %s", escrowStateAfter)

	// Bank balances should stay unchanged before/after upgrade
	suite.Require().Equal(totalBankEscrowBefore, totalBankEscrowAfter, "bank balances should stay the same after upgrade")

	// Post-upgrade, escrow bank balances should match escrow state
	suite.Require().Equal(escrowStateAfter, totalBankEscrowAfter, "escrow state/balance mismatch after upgrade")
}
