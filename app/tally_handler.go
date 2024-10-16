package app

import (
	"context"
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	earnkeeper "github.com/kava-labs/kava/x/earn/keeper"
	liquidkeeper "github.com/kava-labs/kava/x/liquid/keeper"
	liquidtypes "github.com/kava-labs/kava/x/liquid/types"
	savingskeeper "github.com/kava-labs/kava/x/savings/keeper"
)

var _ govv1.TallyHandler = TallyHandler{}

// TallyHandler is the tally handler for kava
type TallyHandler struct {
	gk  govkeeper.Keeper
	stk stakingkeeper.Keeper
	svk savingskeeper.Keeper
	ek  earnkeeper.Keeper
	lk  liquidkeeper.Keeper
	bk  bankkeeper.Keeper
}

// NewTallyHandler creates a new tally handler.
func NewTallyHandler(
	gk govkeeper.Keeper, stk stakingkeeper.Keeper, svk savingskeeper.Keeper,
	ek earnkeeper.Keeper, lk liquidkeeper.Keeper, bk bankkeeper.Keeper,
) TallyHandler {
	return TallyHandler{
		gk:  gk,
		stk: stk,
		svk: svk,
		ek:  ek,
		lk:  lk,
		bk:  bk,
	}
}

// need the method: Tally(context.Context, Proposal)                    (passes bool, burnDeposits bool, tallyResults TallyResult, err error)
// have the method: Tally(ctx context.Context, proposal govv1.Proposal) (passes bool, burnDeposits bool, tallyResults govv1.TallyResult)

func (th TallyHandler) Tally(
	ctx context.Context,
	proposal govv1.Proposal,
) (passes bool, burnDeposits bool, tallyResults govv1.TallyResult, err error) {
	results := make(map[govv1.VoteOption]sdkmath.LegacyDec)
	results[govv1.OptionYes] = sdkmath.LegacyZeroDec()
	results[govv1.OptionAbstain] = sdkmath.LegacyZeroDec()
	results[govv1.OptionNo] = sdkmath.LegacyZeroDec()
	results[govv1.OptionNoWithVeto] = sdkmath.LegacyZeroDec()

	totalVotingPower := sdkmath.LegacyZeroDec()
	currValidators := make(map[string]govv1.ValidatorGovInfo)

	// fetch all the bonded validators, insert them into currValidators
	th.stk.IterateBondedValidatorsByPower(ctx, func(index int64, validator stakingtypes.ValidatorI) (stop bool) {
		currValidators[validator.GetOperator()] = govv1.NewValidatorGovInfo(
			[]byte(validator.GetOperator()),
			validator.GetBondedTokens(),
			validator.GetDelegatorShares(),
			sdkmath.LegacyZeroDec(),
			govv1.WeightedVoteOptions{},
		)

		return false
	})

	// Cannot use 'func(key []byte, value []byte) bool { }' (type func(key []byte, value []byte) bool) as the type Order
	iterator, err := th.gk.Votes.Iterate(ctx, nil)
	if err != nil {
		panic(err)
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)

	for ; iterator.Valid(); iterator.Next() {
		vote, err := iterator.Value()
		if err != nil {
			panic(err)
		}

		// if validator, just record it in the map
		voter, err := sdk.AccAddressFromBech32(vote.Voter)

		if err != nil {
			panic(err)
		}

		valAddrStr := sdk.ValAddress(voter.Bytes()).String()
		if val, ok := currValidators[valAddrStr]; ok {
			val.Vote = vote.Options
			currValidators[valAddrStr] = val
		}

		// iterate over all delegations from voter, deduct from any delegated-to validators
		th.stk.IterateDelegations(ctx, voter, func(index int64, delegation stakingtypes.DelegationI) (stop bool) {
			valAddrStr := delegation.GetValidatorAddr()

			if val, ok := currValidators[valAddrStr]; ok {
				// There is no need to handle the special case that validator address equal to voter address.
				// Because voter's voting power will tally again even if there will deduct voter's voting power from validator.
				val.DelegatorDeductions = val.DelegatorDeductions.Add(delegation.GetShares())
				currValidators[valAddrStr] = val

				// delegation shares * bonded / total shares
				votingPower := delegation.GetShares().MulInt(val.BondedTokens).Quo(val.DelegatorShares)

				for _, option := range vote.Options {
					subPower := votingPower.Mul(sdkmath.LegacyMustNewDecFromStr(option.Weight))
					results[option.Option] = results[option.Option].Add(subPower)
				}
				totalVotingPower = totalVotingPower.Add(votingPower)
			}

			return false
		})

		// get voter bkava and update total voting power and results
		addrBkava := th.getAddrBkava(sdkCtx, voter).toCoins()
		for _, coin := range addrBkava {
			valAddr, err := liquidtypes.ParseLiquidStakingTokenDenom(coin.Denom)
			if err != nil {
				break
			}

			// reduce delegator shares by the amount of voter bkava for the validator
			valAddrStr := valAddr.String()
			if val, ok := currValidators[valAddrStr]; ok {
				val.DelegatorDeductions = val.DelegatorDeductions.Add(sdkmath.LegacyNewDecFromInt(coin.Amount))
				currValidators[valAddrStr] = val
			}

			// votingPower = amount of ukava coin
			stakedCoins, err := th.lk.GetStakedTokensForDerivatives(sdkCtx, sdk.NewCoins(coin))
			if err != nil {
				// error is returned only if the bkava denom is incorrect, which should never happen here.
				panic(err)
			}
			votingPower := sdkmath.LegacyNewDecFromInt(stakedCoins.Amount)

			for _, option := range vote.Options {
				subPower := votingPower.Mul(sdkmath.LegacyMustNewDecFromStr(option.Weight))
				results[option.Option] = results[option.Option].Add(subPower)
			}
			totalVotingPower = totalVotingPower.Add(votingPower)
		}

		// TODO(boodyvo): check here. This one deletes all votes from proposal ID, not just particular one
		th.gk.DeleteVote(ctx, vote.ProposalId)
	}

	iterator, err = th.gk.Votes.Iterate(ctx, nil)
	if err != nil {
		panic(err)
	}

	for ; iterator.Valid(); iterator.Next() {
		vote, err := iterator.Value()
		if err != nil {
			panic(err)
		}
		// if validator, just record it in the map
		voter, err := sdk.AccAddressFromBech32(vote.Voter)

		if err != nil {
			panic(err)
		}

		valAddrStr := sdk.ValAddress(voter.Bytes()).String()
		if val, ok := currValidators[valAddrStr]; ok {
			val.Vote = vote.Options
			currValidators[valAddrStr] = val
		}

		// iterate over all delegations from voter, deduct from any delegated-to validators
		th.stk.IterateDelegations(ctx, voter, func(index int64, delegation stakingtypes.DelegationI) (stop bool) {
			valAddrStr := delegation.GetValidatorAddr()

			if val, ok := currValidators[valAddrStr]; ok {
				// There is no need to handle the special case that validator address equal to voter address.
				// Because voter's voting power will tally again even if there will deduct voter's voting power from validator.
				val.DelegatorDeductions = val.DelegatorDeductions.Add(delegation.GetShares())
				currValidators[valAddrStr] = val

				// delegation shares * bonded / total shares
				votingPower := delegation.GetShares().MulInt(val.BondedTokens).Quo(val.DelegatorShares)

				for _, option := range vote.Options {
					subPower := votingPower.Mul(sdkmath.LegacyMustNewDecFromStr(option.Weight))
					results[option.Option] = results[option.Option].Add(subPower)
				}
				totalVotingPower = totalVotingPower.Add(votingPower)
			}

			return false
		})

		// get voter bkava and update total voting power and results
		addrBkava := th.getAddrBkava(sdkCtx, voter).toCoins()
		for _, coin := range addrBkava {
			valAddr, err := liquidtypes.ParseLiquidStakingTokenDenom(coin.Denom)
			if err != nil {
				break
			}

			// reduce delegator shares by the amount of voter bkava for the validator
			valAddrStr := valAddr.String()
			if val, ok := currValidators[valAddrStr]; ok {
				val.DelegatorDeductions = val.DelegatorDeductions.Add(sdkmath.LegacyNewDecFromInt(coin.Amount))
				currValidators[valAddrStr] = val
			}

			// votingPower = amount of ukava coin
			stakedCoins, err := th.lk.GetStakedTokensForDerivatives(sdkCtx, sdk.NewCoins(coin))
			if err != nil {
				// error is returned only if the bkava denom is incorrect, which should never happen here.
				panic(err)
			}
			votingPower := sdkmath.LegacyNewDecFromInt(stakedCoins.Amount)

			for _, option := range vote.Options {
				subPower := votingPower.Mul(sdkmath.LegacyMustNewDecFromStr(option.Weight))
				results[option.Option] = results[option.Option].Add(subPower)
			}
			totalVotingPower = totalVotingPower.Add(votingPower)
		}

		// TODO(boodyvo): check this one. This switch to removed all votes,
		//th.gk.DeleteVote(ctx, vote.ProposalId, voter)
		th.gk.DeleteVote(ctx, vote.ProposalId)
		//return false
	}

	// iterate over the validators again to tally their voting power
	for _, val := range currValidators {
		if len(val.Vote) == 0 {
			continue
		}

		sharesAfterDeductions := val.DelegatorShares.Sub(val.DelegatorDeductions)
		votingPower := sharesAfterDeductions.MulInt(val.BondedTokens).Quo(val.DelegatorShares)

		for _, option := range val.Vote {
			subPower := votingPower.Mul(sdkmath.LegacyMustNewDecFromStr(option.Weight))
			results[option.Option] = results[option.Option].Add(subPower)
		}
		totalVotingPower = totalVotingPower.Add(votingPower)
	}

	tallyParams, err := th.gk.Params.Get(ctx)
	if err != nil {
		return false, false, govv1.TallyResult{}, err
	}
	tallyResults = govv1.NewTallyResultFromMap(results)

	totalBondedTokens, err := th.stk.TotalBondedTokens(ctx)
	if err != nil {
		return false, false, tallyResults, err
	}
	// TODO: Upgrade the spec to cover all of these cases & remove pseudocode.
	// If there is no staked coins, the proposal fails
	if totalBondedTokens.IsZero() {
		// TODO(boodyvo): return particular error
		return false, false, tallyResults, nil
	}

	// If there is not enough quorum of votes, the proposal fails
	percentVoting := totalVotingPower.Quo(sdkmath.LegacyNewDecFromInt(totalBondedTokens))
	if percentVoting.LT(sdkmath.LegacyMustNewDecFromStr(tallyParams.Quorum)) {
		// TODO(boodyvo): return particular error
		return false, tallyParams.BurnVoteQuorum, tallyResults, nil
	}

	// If no one votes (everyone abstains), proposal fails
	if totalVotingPower.Sub(results[govv1.OptionAbstain]).Equal(sdkmath.LegacyZeroDec()) {
		// TODO(boodyvo): return particular error
		return false, false, tallyResults, nil
	}

	// If more than 1/3 of voters veto, proposal fails
	if results[govv1.OptionNoWithVeto].Quo(totalVotingPower).GT(sdkmath.LegacyMustNewDecFromStr(tallyParams.VetoThreshold)) {
		// TODO(boodyvo): return particular error
		return false, tallyParams.BurnVoteVeto, tallyResults, nil
	}

	// If more than 1/2 of non-abstaining voters vote Yes, proposal passes
	if results[govv1.OptionYes].Quo(totalVotingPower.Sub(results[govv1.OptionAbstain])).GT(sdkmath.LegacyMustNewDecFromStr(tallyParams.Threshold)) {
		// TODO(boodyvo): return particular error
		return true, false, tallyResults, nil
	}

	// If more than 1/2 of non-abstaining voters vote No, proposal fails
	return false, false, tallyResults, nil
}

// bkavaByDenom a map of the bkava denom and the amount of bkava for that denom.
type bkavaByDenom map[string]sdkmath.Int

func (bkavaMap bkavaByDenom) add(coin sdk.Coin) {
	_, found := bkavaMap[coin.Denom]
	if !found {
		bkavaMap[coin.Denom] = sdkmath.ZeroInt()
	}
	bkavaMap[coin.Denom] = bkavaMap[coin.Denom].Add(coin.Amount)
}

func (bkavaMap bkavaByDenom) toCoins() sdk.Coins {
	coins := sdk.Coins{}
	for denom, amt := range bkavaMap {
		coins = coins.Add(sdk.NewCoin(denom, amt))
	}
	return coins.Sort()
}

// getAddrBkava returns a map of validator address & the amount of bkava
// of the addr for each validator.
func (th TallyHandler) getAddrBkava(ctx sdk.Context, addr sdk.AccAddress) bkavaByDenom {
	results := make(bkavaByDenom)
	th.addBkavaFromWallet(ctx, addr, results)
	th.addBkavaFromSavings(ctx, addr, results)
	th.addBkavaFromEarn(ctx, addr, results)
	return results
}

// addBkavaFromWallet adds all addr balances of bkava in x/bank.
func (th TallyHandler) addBkavaFromWallet(ctx sdk.Context, addr sdk.AccAddress, bkava bkavaByDenom) {
	coins := th.bk.GetAllBalances(ctx, addr)
	for _, coin := range coins {
		if th.lk.IsDerivativeDenom(ctx, coin.Denom) {
			bkava.add(coin)
		}
	}
}

// addBkavaFromSavings adds all addr deposits of bkava in x/savings.
func (th TallyHandler) addBkavaFromSavings(ctx sdk.Context, addr sdk.AccAddress, bkava bkavaByDenom) {
	deposit, found := th.svk.GetDeposit(ctx, addr)
	if !found {
		return
	}
	for _, coin := range deposit.Amount {
		if th.lk.IsDerivativeDenom(ctx, coin.Denom) {
			bkava.add(coin)
		}
	}
}

// addBkavaFromEarn adds all addr deposits of bkava in x/earn.
func (th TallyHandler) addBkavaFromEarn(ctx sdk.Context, addr sdk.AccAddress, bkava bkavaByDenom) {
	shares, found := th.ek.GetVaultAccountShares(ctx, addr)
	if !found {
		return
	}
	for _, share := range shares {
		if th.lk.IsDerivativeDenom(ctx, share.Denom) {
			if coin, err := th.ek.ConvertToAssets(ctx, share); err == nil {
				bkava.add(coin)
			}
		}
	}
}
