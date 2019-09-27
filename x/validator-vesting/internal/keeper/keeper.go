package keeper

import (
	"fmt"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingexported "github.com/cosmos/cosmos-sdk/x/staking/exported"
	"github.com/cosmos/cosmos-sdk/x/validator-vesting/internal/types"
	"github.com/tendermint/tendermint/libs/log"
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
			keys = append(keys, key)
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
		vv.MissingSignCount[0]++
	}
	vv.MissingSignCount[1]++
	k.ak.SetAccount(ctx, vv)
}

// UpdateVestedCoinsProgress sets the VestingPeriodProgress variable (0 = coins did not vest for the period, 1 = coins did vest for the period) for the given address and period. If coins did not vest, those coins are added to DebtAfterFailedVesting. Finally, MissingSignCount is reset to [0,0], representing that the next period has started and no blocks have been missed.
func (k Keeper) UpdateVestedCoinsProgress(ctx sdk.Context, addr sdk.AccAddress, period int) {
	vv := k.GetAccountFromAuthKeeper(ctx, addr)

	threshold := sdk.NewDec(vv.SigningThreshold)
	blocksMissed := sdk.NewDec(vv.MissingSignCount[0])
	blockCount := sdk.NewDec(vv.MissingSignCount[1])
	var successfulVest bool
	if blockCount.IsZero() {
		successfulVest = true
	} else {
		blocksSigned := blockCount.Sub(blocksMissed)
		percentageBlocksSigned := blocksSigned.Quo(blockCount).Mul(sdk.NewDec(100))
		successfulVest = percentageBlocksSigned.GTE(threshold)
	}

	if successfulVest {
		vv.VestingPeriodProgress[period][1] = 1
	} else {
		vv.VestingPeriodProgress[period][1] = 0
		notVestedTokens := vv.VestingPeriods[period].VestingAmount
		// add the tokens that did not vest to DebtAfterFailedVesting
		vv.DebtAfterFailedVesting = vv.DebtAfterFailedVesting.Add(notVestedTokens)
	}
	vv.VestingPeriodProgress[period][0] = 1
	// reset the number of missed blocks and total number of blocks in the period to zero
	vv.MissingSignCount = []int64{0, 0}
	k.ak.SetAccount(ctx, vv)
}

// HandleVestingDebt removes coins after a vesting period in which the vesting
// threshold was not met. Sends/Burns tokens if there is enough spendable tokens,
// otherwise unbonds all existing tokens.
func (k Keeper) HandleVestingDebt(ctx sdk.Context, addr sdk.AccAddress, blockTime time.Time) {
	vv := k.GetAccountFromAuthKeeper(ctx, addr)
	remainingDebt := !vv.DebtAfterFailedVesting.IsZero()
	if remainingDebt {
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
			vv.DebtAfterFailedVesting = sdk.NewCoins()
			k.ak.SetAccount(ctx, vv)
		} else {
			// iterate over all delegations made from the validator vesting account and undelegate
			// note that we cannot safely undelegate only an amount of shares that covers the debt,
			// because the value of those shares could change if a validator gets slashed.
			k.stakingKeeper.IterateDelegations(ctx, vv.Address, func(index int64, d stakingexported.DelegationI) (stop bool) {
				k.stakingKeeper.Undelegate(ctx, d.GetDelegatorAddr(), d.GetValidatorAddr(), d.GetShares())
				return false
			})
		}
	}
}

// GetPeriodEndTimes returns an array of the times when each period ends
func (k Keeper) GetPeriodEndTimes(ctx sdk.Context, addr sdk.AccAddress) []int64 {
	var endTimes []int64
	vv := k.GetAccountFromAuthKeeper(ctx, addr)
	currentEndTime := vv.StartTime
	for _, p := range vv.VestingPeriods {
		currentEndTime += p.PeriodLength
		endTimes = append(endTimes, currentEndTime)
	}
	return endTimes
}
