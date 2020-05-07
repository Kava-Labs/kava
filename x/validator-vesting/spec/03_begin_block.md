# Begin Block

At each `BeginBlock`, all validator vesting accounts are iterated over to update the status of the current vesting period. Note that the address of each account is retrieved by iterating over the keys in the `validator-vesting` store, while the account objects are stored and accessed using the `auth` module's `AccountKeeper`. For each account, the block count is incremented, the missed sign count is incremented if the validator did not sign the block or was not found in the validator set. By comparing the blocktime of the current `BeginBlock`, with the value of `previousBlockTime` stored in the `validator-vesting` store, it is determined if the end of the current period has been reached. If the current period has ended, the `VestingPeriodProgress` field is updated to reflect if the coins for the ending period successfully vested or not. After updates are made regarding the status of the current vesting period, any outstanding debt on the account is attempted to be collected. If there is enough `SpendableBalance` on the account to cover the debt, coins are sent to the `ReturnAdress` or burned. If there is not enough `SpendableBalance` to cover the debt, all delegations of the account are `Unbonded`. Once those unbonding events reach maturity, the coins freed from the unbonding will be used to cover the debt. Finally, the time of the previous block is stored in the validator vesting account keeper, which is used to determine when a period has ended.

```go
func BeginBlocker(ctx sdk.Context, req abci.RequestBeginBlock, k keeper.Keeper) {
	previousBlockTime := time.Time{}
	if ctx.BlockHeight() > 1 {
		previousBlockTime = k.GetPreviousBlockTime(ctx)
	}

	currentBlockTime := ctx.BlockTime()
	var voteInfos VoteInfos
	voteInfos = req.LastCommitInfo.GetVotes()
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
		}
		// handle any new/remaining debt on the account
		k.HandleVestingDebt(ctx, key, currentBlockTime)
	}
	k.SetPreviousBlockTime(ctx, currentBlockTime)
}
```

