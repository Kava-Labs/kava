package validatorvesting

import (
	"bytes"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/validator-vesting/internal/keeper"
	abci "github.com/tendermint/tendermint/abci/types"
)

// BeginBlocker updates the vote signing information for each validator vesting account, updates account when period changes, and updates the previousBlockTime value in the store.
func BeginBlocker(ctx sdk.Context, req abci.RequestBeginBlock, k keeper.Keeper) {
	previousBlockTime := time.Time{}
	if ctx.BlockHeight() > 1 {
		previousBlockTime = k.GetPreviousBlockTime(ctx)
	}

	currentBlockTime := req.Header.GetTime()
	var voteInfos VoteInfos
	voteInfos = ctx.VoteInfos()
	validatorVestingKeys := k.GetAllAccountKeys(ctx)
	for _, key := range validatorVestingKeys {
		acc := k.GetAccountFromAuthKeeper(ctx, key)
		if voteInfos.ContainsValidatorAddress(acc.ValidatorAddress) {
			vote := voteInfos.MustFilterByValidatorAddress(acc.ValidatorAddress)
			if !vote.SignedLastBlock {
				// if the validator explicitly missed signing the block, increment the missing sign count
				k.UpdateMissingSignCount(ctx, acc.GetAddress(), true)
			} else {
				k.UpdateMissingSignCount(ctx, acc.GetAddress(), false)
			}
		} else {
			// if the validator was not a voting member of the validator set, increment the missing sign count
			k.UpdateMissingSignCount(ctx, acc.GetAddress(), true)
		}

		// check if a period ended in the last block
		endTimes := k.GetPeriodEndTimes(ctx, key)
		for i, t := range endTimes {
			if currentBlockTime.Unix() >= t && previousBlockTime.Unix() < t {
				k.UpdateVestedCoinsProgress(ctx, key, i)
			}
			k.HandleVestingDebt(ctx, key, currentBlockTime)
		}
	}
	k.SetPreviousBlockTime(ctx, currentBlockTime)
}

// VoteInfos an array of abci.VoteInfo
type VoteInfos []abci.VoteInfo

// ContainsValidatorAddress returns true if the input validator address is found in the VoteInfos array
func (vis VoteInfos) ContainsValidatorAddress(consAddress sdk.ConsAddress) bool {
	for _, vi := range vis {
		votingAddress := sdk.ConsAddress(vi.Validator.Address)
		if bytes.Equal(consAddress, votingAddress) {
			return true
		}
	}
	return false
}

// MustFilterByValidatorAddress returns the VoteInfo that has a validator address matching the input validator address
func (vis VoteInfos) MustFilterByValidatorAddress(consAddress sdk.ConsAddress) abci.VoteInfo {
	for i, vi := range vis {
		votingAddress := sdk.ConsAddress(vi.Validator.Address)
		if bytes.Equal(consAddress, votingAddress) {
			return vis[i]
		}
	}
	panic("validator address not found")
}
