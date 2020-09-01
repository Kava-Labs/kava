package keeper

import (
	"fmt"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingexported "github.com/cosmos/cosmos-sdk/x/staking/exported"

	"github.com/tendermint/tendermint/libs/log"

	"github.com/kava-labs/kava/x/validator-vesting/types"
)

// Keeper of the validatorvesting store
type Keeper struct {
	storeKey      sdk.StoreKey
	cdc           *codec.Codec
	ak            types.AccountKeeper
	bk            types.BankKeeper
	supplyKeeper  types.SupplyKeeper
	stakingKeeper types.StakingKeeper
}

// NewKeeper creates a new Keeper instance
func NewKeeper(cdc *codec.Codec, key sdk.StoreKey, ak types.AccountKeeper, bk types.BankKeeper, sk types.SupplyKeeper, stk types.StakingKeeper) Keeper {

	return Keeper{
		cdc:           cdc,
		storeKey:      key,
		ak:            ak,
		bk:            bk,
		supplyKeeper:  sk,
		stakingKeeper: stk,
	}
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// GetPreviousBlockTime get the blocktime for the previous block
func (k Keeper) GetPreviousBlockTime(ctx sdk.Context) (blockTime time.Time) {
	store := ctx.KVStore(k.storeKey)
	b := store.Get(types.BlocktimeKey)
	if b == nil {
		panic("Previous block time not set")
	}
	k.cdc.MustUnmarshalBinaryLengthPrefixed(b, &blockTime)
	return blockTime
}

// SetPreviousBlockTime set the time of the previous block
func (k Keeper) SetPreviousBlockTime(ctx sdk.Context, blockTime time.Time) {
	store := ctx.KVStore(k.storeKey)
	b := k.cdc.MustMarshalBinaryLengthPrefixed(blockTime)
	store.Set(types.BlocktimeKey, b)
}

// SetValidatorVestingAccountKey stores the account key in the store. This is useful for when we want to iterate over all ValidatorVestingAcounts, so we can avoid iterating over any other accounts stored in the auth keeper.
func (k Keeper) SetValidatorVestingAccountKey(ctx sdk.Context, addr sdk.AccAddress) {
	store := ctx.KVStore(k.storeKey)
	// using empty bytes as value since the only thing we want to do is iterate over the keys.
	store.Set(types.ValidatorVestingAccountKey(addr), []byte{0})
}

// IterateAccountKeys iterates over all the stored account keys and performs a callback function
func (k Keeper) IterateAccountKeys(ctx sdk.Context, cb func(accountKey []byte) (stop bool)) {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.ValidatorVestingAccountPrefix)

	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		accountKey := iterator.Key()

		if cb(accountKey) {
			break
		}
	}
}

// GetAllAccountKeys returns all account keys in the validator vesting keeper.
func (k Keeper) GetAllAccountKeys(ctx sdk.Context) (keys [][]byte) {
	k.IterateAccountKeys(ctx,
		func(key []byte) (stop bool) {
			keys = append(keys, key[1:])
			return false
		})
	return keys
}

// GetAccountFromAuthKeeper returns a ValidatorVestingAccount from the auth keeper
func (k Keeper) GetAccountFromAuthKeeper(ctx sdk.Context, addr sdk.AccAddress) *types.ValidatorVestingAccount {
	acc := k.ak.GetAccount(ctx, addr)
	vv, ok := acc.(*types.ValidatorVestingAccount)
	if ok {
		return vv
	}
	panic("validator vesting account not found")
}

// UpdateMissingSignCount increments the count of blocks missed during the current period
func (k Keeper) UpdateMissingSignCount(ctx sdk.Context, addr sdk.AccAddress, missedBlock bool) {
	vv := k.GetAccountFromAuthKeeper(ctx, addr)
	if missedBlock {
		vv.CurrentPeriodProgress.MissedBlocks++
	}
	vv.CurrentPeriodProgress.TotalBlocks++
	k.ak.SetAccount(ctx, vv)
}

// UpdateVestedCoinsProgress sets the VestingPeriodProgress variable (0 = coins did not vest for the period, 1 = coins did vest for the period) for the given address and period. If coins did not vest, those coins are added to DebtAfterFailedVesting. Finally, MissingSignCount is reset to [0,0], representing that the next period has started and no blocks have been missed.
func (k Keeper) UpdateVestedCoinsProgress(ctx sdk.Context, addr sdk.AccAddress, period int) {
	vv := k.GetAccountFromAuthKeeper(ctx, addr)
	var successfulVest bool
	if sdk.NewDec(vv.CurrentPeriodProgress.TotalBlocks).IsZero() {
		successfulVest = true
	} else {
		successfulVest = vv.CurrentPeriodProgress.SignedPercetageIsOverThreshold(vv.SigningThreshold)
	}

	if successfulVest {
		k.SetVestingProgress(ctx, addr, period, true)
	} else {
		k.SetVestingProgress(ctx, addr, period, false)
		k.AddDebt(ctx, addr, vv.VestingPeriods[period].Amount)
	}
	k.ResetCurrentPeriodProgress(ctx, addr)
}

// SetVestingProgress sets VestingPeriodProgress for the input period
func (k Keeper) SetVestingProgress(ctx sdk.Context, addr sdk.AccAddress, period int, success bool) {
	vv := k.GetAccountFromAuthKeeper(ctx, addr)
	vv.VestingPeriodProgress[period] = types.VestingProgress{PeriodComplete: true, VestingSuccessful: success}
	k.ak.SetAccount(ctx, vv)
}

// AddDebt adds the input amount to DebtAfterFailedVesting field
func (k Keeper) AddDebt(ctx sdk.Context, addr sdk.AccAddress, amount sdk.Coins) {
	vv := k.GetAccountFromAuthKeeper(ctx, addr)
	vv.DebtAfterFailedVesting = vv.DebtAfterFailedVesting.Add(amount...)
	k.ak.SetAccount(ctx, vv)
}

// ResetCurrentPeriodProgress resets CurrentPeriodProgress to zero values
func (k Keeper) ResetCurrentPeriodProgress(ctx sdk.Context, addr sdk.AccAddress) {
	vv := k.GetAccountFromAuthKeeper(ctx, addr)
	vv.CurrentPeriodProgress = types.CurrentPeriodProgress{TotalBlocks: 0, MissedBlocks: 0}
	k.ak.SetAccount(ctx, vv)
}

// HandleVestingDebt removes coins after a vesting period in which the vesting
// threshold was not met. Sends/Burns tokens if there is enough spendable tokens,
// otherwise unbonds all existing tokens.
func (k Keeper) HandleVestingDebt(ctx sdk.Context, addr sdk.AccAddress, blockTime time.Time) {
	vv := k.GetAccountFromAuthKeeper(ctx, addr)

	if vv.DebtAfterFailedVesting.IsZero() {
		return
	}
	spendableCoins := vv.SpendableCoins(blockTime)
	if spendableCoins.IsAllGTE(vv.DebtAfterFailedVesting) {
		if vv.ReturnAddress != nil {
			err := k.bk.SendCoins(ctx, addr, vv.ReturnAddress, vv.DebtAfterFailedVesting)
			if err != nil {
				panic(err)
			}
		} else {
			err := k.supplyKeeper.SendCoinsFromAccountToModule(ctx, addr, types.ModuleName, vv.DebtAfterFailedVesting)
			if err != nil {
				panic(err)
			}
			err = k.supplyKeeper.BurnCoins(ctx, types.ModuleName, vv.DebtAfterFailedVesting)
			if err != nil {
				panic(err)
			}
		}
		k.ResetDebt(ctx, addr)
	} else {
		// iterate over all delegations made from the validator vesting account and undelegate
		// note that we cannot safely undelegate only an amount of shares that covers the debt,
		// because the value of those shares could change if a validator gets slashed.
		k.stakingKeeper.IterateDelegations(ctx, vv.Address, func(index int64, d stakingexported.DelegationI) (stop bool) {
			_, err := k.stakingKeeper.Undelegate(ctx, d.GetDelegatorAddr(), d.GetValidatorAddr(), d.GetShares())
			if err != nil {
				// TODO what should we do instead of panic here?
				panic(err)
			}
			return false
		})
	}
}

// ResetDebt sets DebtAfterFailedVesting to zero
func (k Keeper) ResetDebt(ctx sdk.Context, addr sdk.AccAddress) {
	vv := k.GetAccountFromAuthKeeper(ctx, addr)
	vv.DebtAfterFailedVesting = sdk.NewCoins()
	k.ak.SetAccount(ctx, vv)
}

// GetPeriodEndTimes returns an array of the times when each period ends
func (k Keeper) GetPeriodEndTimes(ctx sdk.Context, addr sdk.AccAddress) []int64 {
	var endTimes []int64
	vv := k.GetAccountFromAuthKeeper(ctx, addr)
	currentEndTime := vv.StartTime
	for _, p := range vv.VestingPeriods {
		currentEndTime += p.Length
		endTimes = append(endTimes, currentEndTime)
	}
	return endTimes
}

// AccountIsVesting returns true if all vesting periods is complete and there is no debt
func (k Keeper) AccountIsVesting(ctx sdk.Context, addr sdk.AccAddress) bool {
	vv := k.GetAccountFromAuthKeeper(ctx, addr)
	if !vv.DebtAfterFailedVesting.IsZero() {
		return true
	}
	for _, p := range vv.VestingPeriodProgress {
		if !p.PeriodComplete {
			return true
		}
	}
	return false
}
