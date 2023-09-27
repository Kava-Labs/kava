package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kava-labs/kava/x/community/types"

	kavadisttypes "github.com/kava-labs/kava/x/kavadist/types"

	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
)

// StartCommunityFundConsolidation consolidates the community funds from
// x/distribution and x/kavadist into the x/community module account
func (k Keeper) StartCommunityFundConsolidation(ctx sdk.Context) error {
	logger := k.Logger(ctx)
	logger.Info("community fund consolidation upgrade started")

	// Consolidate x/distribution community pool
	if err := k.consolidateCommunityDistribution(ctx); err != nil {
		return err
	}

	// Consolidate x/kavadist account
	if err := k.consolidateCommunityKavadist(ctx); err != nil {
		return err
	}

	// Log new x/community balance
	communityCoins := k.GetModuleAccountBalance(ctx)
	logger.Info(fmt.Sprintf("community funds consolidated, x/community balance is now %s", communityCoins))

	return nil
}

// consolidateCommunityDistribution transfers all coins from the x/distribution
// community pool to the x/community module account
func (k Keeper) consolidateCommunityDistribution(ctx sdk.Context) error {
	logger := k.Logger(ctx)

	// Get community coins with leftover leftoverDust
	truncatedCoins, leftoverDust := k.distrKeeper.
		GetFeePoolCommunityCoins(ctx).
		TruncateDecimal()

	// Transfer to x/community
	err := k.bankKeeper.SendCoinsFromModuleToModule(
		ctx,
		distrtypes.ModuleName, // sender
		types.ModuleName,      // recipient
		truncatedCoins,
	)
	if err != nil {
		return fmt.Errorf("failed to transfer x/distribution coins to x/community: %w", err)
	}

	logger.Info(fmt.Sprintf("transferred %s from x/distribution to x/community", truncatedCoins))

	// Set x/distribution community pool to remaining dust amounts
	feePool := k.distrKeeper.GetFeePool(ctx)
	feePool.CommunityPool = leftoverDust
	k.distrKeeper.SetFeePool(ctx, feePool)

	logger.Info(fmt.Sprintf("remaining x/distribution community pool dust: %s", leftoverDust))

	return nil
}

// consolidateCommunityKavadist transfers all coins from the x/kavadist module
// account to the x/community module account
func (k Keeper) consolidateCommunityKavadist(ctx sdk.Context) error {
	logger := k.Logger(ctx)

	kavadistAcc := k.accountKeeper.GetModuleAccount(ctx, kavadisttypes.KavaDistMacc)
	transferCoins := k.bankKeeper.GetAllBalances(ctx, kavadistAcc.GetAddress())

	// Remove ukava from transfer coins - ony transfer non-ukava coins
	found, kavaCoins := transferCoins.Find("ukava")
	if found {
		transferCoins = transferCoins.Sub(kavaCoins)
	}

	// Transfer remaining coins to x/community
	err := k.bankKeeper.SendCoinsFromModuleToModule(
		ctx,
		kavadisttypes.ModuleName, // sender
		types.ModuleName,         // recipient
		transferCoins,
	)
	if err != nil {
		return fmt.Errorf("failed to transfer x/kavadist coins to x/community: %w", err)
	}

	kavadistRemainingCoins := k.bankKeeper.GetAllBalances(ctx, kavadistAcc.GetAddress())

	logger.Info(fmt.Sprintf(
		"transferred %s from x/kavadist to x/community, remaining x/kavadist balance: %s",
		transferCoins,
		kavadistRemainingCoins,
	))

	return nil
}
