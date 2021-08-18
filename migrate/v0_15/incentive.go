package v0_15

import (
	"fmt"
	"sort"

	sdk "github.com/cosmos/cosmos-sdk/types"

	v0_15hard "github.com/kava-labs/kava/x/hard/types"
	v0_14incentive "github.com/kava-labs/kava/x/incentive/legacy/v0_14"
	v0_15incentive "github.com/kava-labs/kava/x/incentive/types"
)

// Incentive migrates from a v0.14 incentive genesis state to a v0.15 incentive genesis state
func Incentive(incentiveGS v0_14incentive.GenesisState, hardGS v0_15hard.GenesisState) v0_15incentive.GenesisState {
	// Migrate params
	claimMultipliers := v0_15incentive.Multipliers{}
	for _, m := range incentiveGS.Params.ClaimMultipliers {
		newMultiplier := v0_15incentive.NewMultiplier(v0_15incentive.MultiplierName(m.Name), m.MonthsLockup, m.Factor)
		claimMultipliers = append(claimMultipliers, newMultiplier)
	}
	newMultipliers := v0_15incentive.MultipliersPerDenom{
		{
			Denom:       "hard",
			Multipliers: claimMultipliers,
		},
		{
			Denom:       "ukava",
			Multipliers: claimMultipliers,
		},
		{
			Denom: "swp",
			Multipliers: v0_15incentive.Multipliers{
				{
					Name:         v0_15incentive.Small,
					MonthsLockup: 1,
					Factor:       sdk.MustNewDecFromStr("0.1"),
				},
				{
					Name:         v0_15incentive.Large,
					MonthsLockup: 12,
					Factor:       sdk.OneDec(),
				},
			},
		},
	}

	usdxMintingRewardPeriods := v0_15incentive.RewardPeriods{}
	for _, rp := range incentiveGS.Params.USDXMintingRewardPeriods {
		usdxMintingRewardPeriod := v0_15incentive.NewRewardPeriod(rp.Active,
			rp.CollateralType, rp.Start, rp.End, rp.RewardsPerSecond)
		usdxMintingRewardPeriods = append(usdxMintingRewardPeriods, usdxMintingRewardPeriod)
	}

	delegatorRewardPeriods := v0_15incentive.MultiRewardPeriods{}
	for _, rp := range incentiveGS.Params.HardDelegatorRewardPeriods {
		rewardsPerSecond := sdk.NewCoins(rp.RewardsPerSecond, SwpDelegatorRewardsPerSecond)
		delegatorRewardPeriod := v0_15incentive.NewMultiRewardPeriod(rp.Active,
			rp.CollateralType, rp.Start, rp.End, rewardsPerSecond)
		delegatorRewardPeriods = append(delegatorRewardPeriods, delegatorRewardPeriod)
	}

	// TODO: finalize swap reward pool IDs, rewards per second, start/end times. Should swap rewards start active?
	swapRewardPeriods := v0_15incentive.MultiRewardPeriods{}

	// Build new params from migrated values
	params := v0_15incentive.NewParams(
		usdxMintingRewardPeriods,
		migrateMultiRewardPeriods(incentiveGS.Params.HardSupplyRewardPeriods),
		migrateMultiRewardPeriods(incentiveGS.Params.HardBorrowRewardPeriods),
		delegatorRewardPeriods,
		swapRewardPeriods,
		newMultipliers,
		incentiveGS.Params.ClaimEnd,
	)

	// Migrate accumulation times and reward indexes
	usdxGenesisRewardState := migrateGenesisRewardState(incentiveGS.USDXAccumulationTimes, incentiveGS.USDXRewardIndexes)
	hardSupplyGenesisRewardState := migrateGenesisRewardState(incentiveGS.HardSupplyAccumulationTimes, incentiveGS.HardSupplyRewardIndexes)
	hardBorrowGenesisRewardState := migrateGenesisRewardState(incentiveGS.HardBorrowAccumulationTimes, incentiveGS.HardBorrowRewardIndexes)
	delegatorGenesisRewardState := migrateGenesisRewardState(incentiveGS.HardDelegatorAccumulationTimes, incentiveGS.HardDelegatorRewardIndexes)
	swapGenesisRewardState := v0_15incentive.DefaultGenesisRewardState // There is no previous swap rewards so accumulation starts at genesis time.

	// Migrate USDX minting claims
	usdxMintingClaims := v0_15incentive.USDXMintingClaims{}
	for _, claim := range incentiveGS.USDXMintingClaims {
		rewardIndexes := migrateRewardIndexes(claim.RewardIndexes)
		usdxMintingClaim := v0_15incentive.NewUSDXMintingClaim(claim.Owner, claim.Reward, rewardIndexes)
		usdxMintingClaims = append(usdxMintingClaims, usdxMintingClaim)
	}

	// Migrate Hard protocol claims (includes creating new Delegator claims)
	hardClaims, delegatorClaims, err := migrateHardLiquidityProviderClaims(incentiveGS.HardLiquidityProviderClaims)
	if err != nil {
		panic(fmt.Sprintf("could not migrate hard claims: %v", err.Error()))
	}
	hardClaims = addMissingHardClaims(
		hardClaims,
		hardGS.Deposits, hardGS.Borrows,
		hardSupplyGenesisRewardState.MultiRewardIndexes, hardBorrowGenesisRewardState.MultiRewardIndexes,
	)
	hardClaims = alignClaimIndexes(
		hardClaims,
		hardGS.Deposits, hardGS.Borrows,
		hardSupplyGenesisRewardState.MultiRewardIndexes, hardBorrowGenesisRewardState.MultiRewardIndexes,
	)

	// Add Swap Claims
	swapClaims := v0_15incentive.DefaultSwapClaims

	return v0_15incentive.NewGenesisState(
		params,
		usdxGenesisRewardState,
		hardSupplyGenesisRewardState,
		hardBorrowGenesisRewardState,
		delegatorGenesisRewardState,
		swapGenesisRewardState,
		usdxMintingClaims,
		hardClaims,
		delegatorClaims,
		swapClaims,
	)
}

func migrateHardLiquidityProviderClaims(oldClaims v0_14incentive.HardLiquidityProviderClaims) (v0_15incentive.HardLiquidityProviderClaims, v0_15incentive.DelegatorClaims, error) {

	hardClaims := v0_15incentive.HardLiquidityProviderClaims{}
	delegatorClaims := v0_15incentive.DelegatorClaims{}

	for _, claim := range oldClaims {
		// Migrate supply multi reward indexes
		supplyMultiRewardIndexes := migrateMultiRewardIndexes(claim.SupplyRewardIndexes)

		// Migrate borrow multi reward indexes
		borrowMultiRewardIndexes := migrateMultiRewardIndexes(claim.BorrowRewardIndexes)

		// Migrate delegator reward indexes to multi reward indexes inside DelegatorClaims
		delegatorMultiRewardIndexes, err := migrateDelegatorRewardIndexes(claim.DelegatorRewardIndexes)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid claim found '%s': %w", claim.Owner, err)
		}

		// It's impossible to distinguish between rewards from delegation vs. liquidity provisioning
		// as they're all combined inside claim.Reward, so put them all inside the hard claim to
		// avoid duplicating rewards.
		delegatorClaim := v0_15incentive.NewDelegatorClaim(claim.Owner, sdk.NewCoins(), delegatorMultiRewardIndexes)
		delegatorClaims = append(delegatorClaims, delegatorClaim)

		hardClaim := v0_15incentive.NewHardLiquidityProviderClaim(claim.Owner, claim.Reward,
			supplyMultiRewardIndexes, borrowMultiRewardIndexes)
		hardClaims = append(hardClaims, hardClaim)
	}

	return hardClaims, delegatorClaims, nil
}

func migrateDelegatorRewardIndexes(oldIndexes v0_14incentive.RewardIndexes) (v0_15incentive.MultiRewardIndexes, error) {
	newRIs := v0_15incentive.RewardIndexes{}

	if len(oldIndexes) > 1 {
		return nil, fmt.Errorf("delegator claims should not have more than one rewarded denom")
	}

	if len(oldIndexes) > 0 {
		oldRI := oldIndexes[0]

		if oldRI.CollateralType != v0_15incentive.BondDenom {
			return nil, fmt.Errorf("delegator claims should only reward staked '%s', not '%s'", v0_15incentive.BondDenom, oldRI.CollateralType)
		}
		newRIs = newRIs.With(v0_14incentive.HardLiquidityRewardDenom, oldRI.RewardFactor)
	}

	newIndexes := v0_15incentive.MultiRewardIndexes{
		v0_15incentive.NewMultiRewardIndex(v0_15incentive.BondDenom, newRIs),
	}

	return newIndexes, nil
}

func migrateMultiRewardPeriods(oldPeriods v0_14incentive.MultiRewardPeriods) v0_15incentive.MultiRewardPeriods {
	newPeriods := v0_15incentive.MultiRewardPeriods{}
	for _, rp := range oldPeriods {
		newPeriod := v0_15incentive.NewMultiRewardPeriod(
			rp.Active,
			rp.CollateralType,
			rp.Start,
			rp.End,
			rp.RewardsPerSecond,
		)
		newPeriods = append(newPeriods, newPeriod)
	}
	return newPeriods
}

func migrateGenesisRewardState(oldAccumulationTimes v0_14incentive.GenesisAccumulationTimes, oldIndexes v0_14incentive.GenesisRewardIndexesSlice) v0_15incentive.GenesisRewardState {
	accumulationTimes := v0_15incentive.AccumulationTimes{}
	for _, t := range oldAccumulationTimes {
		newAccumulationTime := v0_15incentive.NewAccumulationTime(t.CollateralType, t.PreviousAccumulationTime)
		accumulationTimes = append(accumulationTimes, newAccumulationTime)
	}
	multiRewardIndexes := v0_15incentive.MultiRewardIndexes{}
	for _, gri := range oldIndexes {
		multiRewardIndex := v0_15incentive.NewMultiRewardIndex(gri.CollateralType, migrateRewardIndexes(gri.RewardIndexes))
		multiRewardIndexes = append(multiRewardIndexes, multiRewardIndex)
	}
	return v0_15incentive.NewGenesisRewardState(
		accumulationTimes,
		multiRewardIndexes,
	)
}

func migrateMultiRewardIndexes(oldIndexes v0_14incentive.MultiRewardIndexes) v0_15incentive.MultiRewardIndexes {
	newIndexes := v0_15incentive.MultiRewardIndexes{}
	for _, mri := range oldIndexes {
		multiRewardIndex := v0_15incentive.NewMultiRewardIndex(
			mri.CollateralType,
			migrateRewardIndexes(mri.RewardIndexes),
		)
		newIndexes = append(newIndexes, multiRewardIndex)
	}
	return newIndexes
}

func migrateRewardIndexes(oldIndexes v0_14incentive.RewardIndexes) v0_15incentive.RewardIndexes {
	newIndexes := v0_15incentive.RewardIndexes{}
	for _, ri := range oldIndexes {
		rewardIndex := v0_15incentive.NewRewardIndex(ri.CollateralType, ri.RewardFactor)
		newIndexes = append(newIndexes, rewardIndex)
	}
	return newIndexes
}

// addMissingHardClaims checks for hard deposits and borrows without claims. If found it creates new
// claims with indexes set to the global indexes so rewards start accumulating from launch.
func addMissingHardClaims(hardClaims v0_15incentive.HardLiquidityProviderClaims, deposits v0_15hard.Deposits, borrows v0_15hard.Borrows, depositGlobalIndexes, borrowGlobalIndexes v0_15incentive.MultiRewardIndexes) v0_15incentive.HardLiquidityProviderClaims {

	missingClaims := map[string]v0_15incentive.HardLiquidityProviderClaim{}

	for _, deposit := range deposits {
		_, found := getHardClaimByOwner(hardClaims, deposit.Depositor)

		if !found {
			missingClaims[deposit.Depositor.String()] = v0_15incentive.NewHardLiquidityProviderClaim(
				deposit.Depositor,
				sdk.NewCoins(), // do not calculate missing rewards
				getIndexesForCoins(deposit.Amount, depositGlobalIndexes),
				v0_15incentive.MultiRewardIndexes{}, // the depositor may also have a borrow, this will be added below
			)
		}
	}

	for _, borrow := range borrows {
		_, found := getHardClaimByOwner(hardClaims, borrow.Borrower)

		if !found {
			c, ok := missingClaims[borrow.Borrower.String()]
			borrowIndexes := getIndexesForCoins(borrow.Amount, borrowGlobalIndexes)
			if ok {
				c.BorrowRewardIndexes = borrowIndexes
				missingClaims[borrow.Borrower.String()] = c
			} else {
				missingClaims[borrow.Borrower.String()] = v0_15incentive.NewHardLiquidityProviderClaim(
					borrow.Borrower,
					sdk.NewCoins(),                      // do not calculate missing rewards
					v0_15incentive.MultiRewardIndexes{}, // this borrow does not have any deposits as it would have been found above
					borrowIndexes,
				)
			}
		}
	}

	// New claims need to be sorted to ensure the output genesis state is the same for everyone.
	// Sorting by address should be give deterministic order as the addresses are unique.
	sortedNewClaims := make(v0_15incentive.HardLiquidityProviderClaims, 0, len(missingClaims))
	for _, claim := range missingClaims {
		sortedNewClaims = append(sortedNewClaims, claim)
	}
	sort.Slice(sortedNewClaims, func(i, j int) bool {
		return sortedNewClaims[i].Owner.String() < sortedNewClaims[j].Owner.String()
	})

	return append(hardClaims, sortedNewClaims...)
}

// getIndexesForCoins returns indexes with collateral types matching the coins denoms. RewardIndexes values are taken
// from the provided indexes, or left empty if not found.
func getIndexesForCoins(coins sdk.Coins, indexes v0_15incentive.MultiRewardIndexes) v0_15incentive.MultiRewardIndexes {
	newIndexes := v0_15incentive.MultiRewardIndexes{}
	for _, c := range coins {
		ri, found := indexes.Get(c.Denom)
		if !found {
			ri = v0_15incentive.RewardIndexes{}
		}
		newIndexes = newIndexes.With(c.Denom, ri)
	}
	return newIndexes
}

// getHardClaimByOwner picks out the first claim matching an address.
func getHardClaimByOwner(claims v0_15incentive.HardLiquidityProviderClaims, owner sdk.AccAddress) (v0_15incentive.HardLiquidityProviderClaim, bool) {
	for _, claim := range claims {
		if claim.Owner.Equals(owner) {
			return claim, true
		}
	}
	return v0_15incentive.HardLiquidityProviderClaim{}, false
}

// getHardDepositByOwner picks out the first deposit matching an address.
func getHardDepositByOwner(deposits v0_15hard.Deposits, owner sdk.AccAddress) (v0_15hard.Deposit, bool) {
	for _, deposit := range deposits {
		if deposit.Depositor.Equals(owner) {
			return deposit, true
		}
	}
	return v0_15hard.Deposit{}, false
}

// getHardBorrowByOwner picks out the first borrow matching an address.
func getHardBorrowByOwner(borrows v0_15hard.Borrows, owner sdk.AccAddress) (v0_15hard.Borrow, bool) {
	for _, borrow := range borrows {
		if borrow.Borrower.Equals(owner) {
			return borrow, true
		}
	}
	return v0_15hard.Borrow{}, false
}

// alignClaimIndexes fixes the supply and borrow indexes on the hard claims to ensure they match the deposits and borrows.
func alignClaimIndexes(claims v0_15incentive.HardLiquidityProviderClaims, deposits v0_15hard.Deposits, borrows v0_15hard.Borrows, depositGlobalIndexes, borrowGlobalIndexes v0_15incentive.MultiRewardIndexes) v0_15incentive.HardLiquidityProviderClaims {
	newClaims := make(v0_15incentive.HardLiquidityProviderClaims, 0, len(claims))

	for _, claim := range claims {

		deposit, found := getHardDepositByOwner(deposits, claim.Owner)
		amt := deposit.Amount
		if !found {
			amt = sdk.NewCoins()
		}
		claim.SupplyRewardIndexes = alignIndexes(claim.SupplyRewardIndexes, amt, depositGlobalIndexes)

		borrow, found := getHardBorrowByOwner(borrows, claim.Owner)
		amt = borrow.Amount
		if !found {
			amt = sdk.NewCoins()
		}
		claim.BorrowRewardIndexes = alignIndexes(claim.BorrowRewardIndexes, amt, depositGlobalIndexes)

		newClaims = append(newClaims, claim)
	}

	return newClaims
}

// alignIndexes adds or remove items from indexes so all the collateral types match the coins denoms.
// Missing index values are filled from globalIndexes if found. Otherwise they're set to empty.
// It preserves the order of the original index to avoid unnecessary churn in the migrated claims.
func alignIndexes(indexes v0_15incentive.MultiRewardIndexes, coins sdk.Coins, globalIndexes v0_15incentive.MultiRewardIndexes) v0_15incentive.MultiRewardIndexes {
	newIndexes := indexes

	// add missing indexes
	for _, coin := range coins {
		if _, found := indexes.Get(coin.Denom); !found {
			ri, f := globalIndexes.Get(coin.Denom)
			if !f {
				ri = v0_15incentive.RewardIndexes{}
			}
			newIndexes = newIndexes.With(coin.Denom, ri)
		}
	}

	// remove extra indexes
	for _, index := range indexes {
		if coins.AmountOf(index.CollateralType).Equal(sdk.ZeroInt()) {
			// RemoveRewardIndex returns a copy of the underlying array so the loop is not distupted
			newIndexes = newIndexes.RemoveRewardIndex(index.CollateralType)
		}
	}

	return newIndexes
}
