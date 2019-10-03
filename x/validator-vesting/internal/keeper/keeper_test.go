package keeper

import (
	"bytes"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	tmtime "github.com/tendermint/tendermint/types/time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
	stakingexported "github.com/cosmos/cosmos-sdk/x/staking/exported"
)

func TestGetSetValidatorVestingAccounts(t *testing.T) {
	ctx, ak, _, _, _, keeper := CreateTestInput(t, false, 1000)

	vva := ValidatorVestingTestAccount()
	// Add the validator vesting account to the auth store
	ak.SetAccount(ctx, vva)

	// require that the keeper can set the account key without panic
	require.NotPanics(t, func() { keeper.SetValidatorVestingAccountKey(ctx, vva.Address) })

	// require that we can get the account from auth keeper as a validator vesting account.
	require.NotPanics(t, func() { keeper.GetAccountFromAuthKeeper(ctx, vva.Address) })

	// fetching a regular account from the auth keeper does not panic
	require.NotPanics(t, func() { ak.GetAccount(ctx, TestAddrs[0]) })

	// fetching a regular account from the validator vesting keeper panics.
	require.Panics(t, func() { keeper.GetAccountFromAuthKeeper(ctx, TestAddrs[0]) })

	// require that GetAllAccountKeys returns one account
	keys := keeper.GetAllAccountKeys(ctx)
	require.Equal(t, 1, len(keys))
	for _, k := range keys {
		require.NotPanics(t, func() { keeper.GetAccountFromAuthKeeper(ctx, k[1:]) })
	}

	vvAccounts := ValidatorVestingTestAccounts(10)
	for _, a := range vvAccounts {
		ak.SetAccount(ctx, a)
		keeper.SetValidatorVestingAccountKey(ctx, a.Address)
	}

	keys = keeper.GetAllAccountKeys(ctx)
	require.Equal(t, 10, len(keys))

	var ikeys [][]byte
	keeper.IterateAccountKeys(ctx, func(accountKey []byte) bool {
		if bytes.Equal(accountKey, keys[0]) {
			ikeys = append(ikeys, accountKey)
			return true
		}
		return false
	})
	require.Equal(t, 1, len(ikeys))
}

func TestGetSetPreviousBlock(t *testing.T) {
	ctx, _, _, _, _, keeper := CreateTestInput(t, false, 1000)
	now := tmtime.Now()

	// require panic if the previous blocktime was never set
	require.Panics(t, func() { keeper.GetPreviousBlockTime(ctx) })

	// require that passing a valid time to SetPreviousBlockTime does not panic
	require.NotPanics(t, func() { keeper.SetPreviousBlockTime(ctx, now) })

	// require that the value from GetPreviousBlockTime equals what was set
	bpt := keeper.GetPreviousBlockTime(ctx)
	require.Equal(t, now, bpt)

	// require that the zero value is safe
	require.NotPanics(t, func() { keeper.SetPreviousBlockTime(ctx, tmtime.Canonical(time.Unix(0, 0))) })

	bpt = keeper.GetPreviousBlockTime(ctx)
	require.Equal(t, tmtime.Canonical(time.Unix(0, 0)), bpt)

}

func TestGetEndTImes(t *testing.T) {
	ctx, ak, _, _, _, keeper := CreateTestInput(t, false, 1000)

	now := tmtime.Now()

	vva := ValidatorVestingDelegatorTestAccount(now)
	ak.SetAccount(ctx, vva)
	keeper.SetValidatorVestingAccountKey(ctx, vva.Address)

	expectedEndTimes := []int64{
		now.Add(12 * time.Hour).Unix(),
		now.Add(18 * time.Hour).Unix(),
		now.Add(24 * time.Hour).Unix(),
	}

	endTimes := keeper.GetPeriodEndTimes(ctx, vva.Address)

	require.Equal(t, expectedEndTimes, endTimes)
}

func TestSetMissingSignCount(t *testing.T) {
	ctx, ak, _, _, _, keeper := CreateTestInput(t, false, 1000)

	vva := ValidatorVestingTestAccount()
	// Add the validator vesting account to the auth store
	ak.SetAccount(ctx, vva)

	// require empty array after ValidatorVestingAccount is initialized
	require.Equal(t, []int64{0, 0}, vva.MissingSignCount)

	// validator signs a block
	keeper.UpdateMissingSignCount(ctx, vva.Address, false)
	vva = keeper.GetAccountFromAuthKeeper(ctx, vva.Address)
	require.Equal(t, []int64{0, 1}, vva.MissingSignCount)

	// validator misses a block
	keeper.UpdateMissingSignCount(ctx, vva.Address, true)
	vva = keeper.GetAccountFromAuthKeeper(ctx, vva.Address)
	require.Equal(t, []int64{1, 2}, vva.MissingSignCount)

}

func TestUpdateVestedCoinsProgress(t *testing.T) {
	ctx, ak, _, _, _, keeper := CreateTestInput(t, false, 1000)

	vva := ValidatorVestingTestAccount()

	// Add the validator vesting account to the auth store
	ak.SetAccount(ctx, vva)

	// require all vesting period tracking variables to be zero after validator vesting account is initialized
	require.Equal(t, [][]int{{0, 0}, {0, 0}, {0, 0}}, vva.VestingPeriodProgress)

	// period 0 passes with all blocks signed
	vva.MissingSignCount[0] = 0
	vva.MissingSignCount[1] = 100
	ak.SetAccount(ctx, vva)
	vva = keeper.GetAccountFromAuthKeeper(ctx, vva.Address)
	keeper.UpdateVestedCoinsProgress(ctx, vva.Address, 0)
	vva = keeper.GetAccountFromAuthKeeper(ctx, vva.Address)
	// require that debt is zero
	require.Equal(t, sdk.Coins(nil), vva.DebtAfterFailedVesting)
	// require that the first vesting progress variable is 1
	require.Equal(t, [][]int{{1, 1}, {0, 0}, {0, 0}}, vva.VestingPeriodProgress)

	// require that the missing block counter has reset
	require.Equal(t, []int64{0, 0}, vva.MissingSignCount)

	vva = ValidatorVestingTestAccount()
	ak.SetAccount(ctx, vva)
	// period 0 passes with no blocks signed
	// this is an edge case that shouldn't happen,
	// the vest is considered successful in this case.
	vva.MissingSignCount[0] = 0
	vva.MissingSignCount[1] = 0
	ak.SetAccount(ctx, vva)
	vva = keeper.GetAccountFromAuthKeeper(ctx, vva.Address)
	keeper.UpdateVestedCoinsProgress(ctx, vva.Address, 0)
	vva = keeper.GetAccountFromAuthKeeper(ctx, vva.Address)
	// require that debt is zero
	require.Equal(t, sdk.Coins(nil), vva.DebtAfterFailedVesting)
	// require that the first vesting progress variable is 1
	require.Equal(t, [][]int{{1, 1}, {0, 0}, {0, 0}}, vva.VestingPeriodProgress)

	// require that the missing block counter has reset
	require.Equal(t, []int64{0, 0}, vva.MissingSignCount)

	vva = ValidatorVestingTestAccount()
	ak.SetAccount(ctx, vva)
	// period 0 passes with 50% of blocks signed (below threshold)
	vva.MissingSignCount[0] = 50
	vva.MissingSignCount[1] = 100
	ak.SetAccount(ctx, vva)
	keeper.UpdateVestedCoinsProgress(ctx, vva.Address, 0)
	vva = keeper.GetAccountFromAuthKeeper(ctx, vva.Address)
	// require that period 1 coins have become debt
	require.Equal(t, sdk.NewCoins(sdk.NewInt64Coin(feeDenom, 500), sdk.NewInt64Coin(stakeDenom, 50)), vva.DebtAfterFailedVesting)
	// require that the first vesting progress variable is {1,0}
	require.Equal(t, [][]int{{1, 0}, {0, 0}, {0, 0}}, vva.VestingPeriodProgress)
	// require that the missing block counter has reset
	require.Equal(t, []int64{0, 0}, vva.MissingSignCount)
}

func TestHandleVestingDebtNoDebt(t *testing.T) {
	// ctx, ak, bk, stakingKeeper, supplyKeeper, keeper := CreateTestInput(t, false, 1000)

	ctx, ak, _, _, _, keeper := CreateTestInput(t, false, 1000)

	vva := ValidatorVestingTestAccount()
	// Delegate all coins
	origCoins := sdk.Coins{sdk.NewInt64Coin(feeDenom, 1000), sdk.NewInt64Coin(stakeDenom, 100)}
	now := tmtime.Now()
	vva.TrackDelegation(now, origCoins)

	// Add the validator vesting account to the auth store
	ak.SetAccount(ctx, vva)

	// Require that calling HandleVestingDebt when debt is zero doesn't alter the delegation
	keeper.HandleVestingDebt(ctx, vva.Address, now)
	vva = keeper.GetAccountFromAuthKeeper(ctx, vva.Address)
	require.Equal(t, origCoins, vva.DelegatedVesting)
	require.Nil(t, vva.DelegatedFree)

}

func TestHandleVestingDebtForcedUnbond(t *testing.T) {
	// ctx, ak, bk, stakingKeeper, supplyKeeper, keeper := CreateTestInput(t, false, 1000)

	ctx, ak, _, stakingKeeper, _, keeper := CreateTestInput(t, false, 1000)
	now := tmtime.Now()

	// Create validators and a delegation from the validator vesting account
	CreateValidators(ctx, stakingKeeper, []int64{5, 5, 5})

	vva := ValidatorVestingDelegatorTestAccount(now)
	ak.SetAccount(ctx, vva)
	delTokens := sdk.TokensFromConsensusPower(60)
	val1, found := stakingKeeper.GetValidator(ctx, ValOpAddr1)
	require.True(t, found)

	_, err := stakingKeeper.Delegate(ctx, vva.Address, delTokens, sdk.Unbonded, val1, true)
	require.NoError(t, err)

	_ = staking.EndBlocker(ctx, stakingKeeper)

	vva = keeper.GetAccountFromAuthKeeper(ctx, vva.Address)
	// t.Log(vva.GetDelegatedFree())
	t.Log(vva.GetDelegatedVesting())

	// require that there exists one delegation
	var delegations int
	stakingKeeper.IterateDelegations(ctx, vva.Address, func(index int64, d stakingexported.DelegationI) (stop bool) {
		delegations++
		return false
	})

	require.Equal(t, 1, delegations)

	// period 0 passes and the threshold is not met
	vva.MissingSignCount[0] = 50
	vva.MissingSignCount[1] = 100
	ak.SetAccount(ctx, vva)
	keeper.UpdateVestedCoinsProgress(ctx, vva.Address, 0)
	vva = keeper.GetAccountFromAuthKeeper(ctx, vva.Address)

	// require that period 0 coins have become debt
	require.Equal(t, sdk.NewCoins(sdk.NewInt64Coin(stakeDenom, 30000000)), vva.DebtAfterFailedVesting)

	// when there are no additional liquid coins in the account, require that there are no delegations after HandleVestingDebt (ie the account has been force unbonded)
	keeper.HandleVestingDebt(ctx, vva.Address, now.Add(12*time.Hour))

	delegations = 0
	stakingKeeper.IterateDelegations(ctx, vva.Address, func(index int64, d stakingexported.DelegationI) (stop bool) {
		delegations++
		return false
	})
	require.Equal(t, 0, delegations)

}

func TestHandleVestingDebtBurn(t *testing.T) {
	ctx, ak, _, stakingKeeper, supplyKeeper, keeper := CreateTestInput(t, false, 1000)
	CreateValidators(ctx, stakingKeeper, []int64{5, 5, 5})
	now := tmtime.Now()
	vva := ValidatorVestingDelegatorTestAccount(now)
	ak.SetAccount(ctx, vva)
	delTokens := sdk.TokensFromConsensusPower(30)
	val1, found := stakingKeeper.GetValidator(ctx, ValOpAddr1)
	require.True(t, found)
	// delegate half the tokens, which will make the period 1 coins that fail to vest immediately cover the debt.
	_, err := stakingKeeper.Delegate(ctx, vva.Address, delTokens, sdk.Unbonded, val1, true)
	require.NoError(t, err)

	_ = staking.EndBlocker(ctx, stakingKeeper)

	// period 0 passes and the threshold is not met
	vva.MissingSignCount[0] = 50
	vva.MissingSignCount[1] = 100
	ak.SetAccount(ctx, vva)
	keeper.UpdateVestedCoinsProgress(ctx, vva.Address, 0)
	vva = keeper.GetAccountFromAuthKeeper(ctx, vva.Address)
	// require that period 0 coins have become debt
	require.Equal(t, sdk.NewCoins(sdk.NewInt64Coin(stakeDenom, 30000000)), vva.DebtAfterFailedVesting)

	initialSupply := supplyKeeper.GetSupply(ctx).GetTotal()
	expectedSupply := initialSupply.Sub(vva.DebtAfterFailedVesting)
	// Context needs the block time because bank keeper calls 'SpendableCoins' by getting the header from the context.
	ctx = ctx.WithBlockTime(now.Add(12 * time.Hour))
	keeper.HandleVestingDebt(ctx, vva.Address, now.Add(12*time.Hour))
	// in the case when the return address is not set require that the total supply has decreased by the debt amount
	require.Equal(t, expectedSupply, supplyKeeper.GetSupply(ctx).GetTotal())
	// require that there is still one delegation
	delegations := 0
	stakingKeeper.IterateDelegations(ctx, vva.Address, func(index int64, d stakingexported.DelegationI) (stop bool) {
		delegations++
		return false
	})
	require.Equal(t, 1, delegations)
	vva = keeper.GetAccountFromAuthKeeper(ctx, vva.Address)
	//require that debt is now zero
	require.Equal(t, sdk.Coins(nil), vva.DebtAfterFailedVesting)
}

func TestHandleVestingDebtReturn(t *testing.T) {
	ctx, ak, _, stakingKeeper, _, keeper := CreateTestInput(t, false, 1000)
	CreateValidators(ctx, stakingKeeper, []int64{5, 5, 5})
	now := tmtime.Now()
	vva := ValidatorVestingDelegatorTestAccount(now)
	vva.ReturnAddress = TestAddrs[2]
	ak.SetAccount(ctx, vva)
	delTokens := sdk.TokensFromConsensusPower(30)
	val1, found := stakingKeeper.GetValidator(ctx, ValOpAddr1)
	require.True(t, found)
	_, err := stakingKeeper.Delegate(ctx, vva.Address, delTokens, sdk.Unbonded, val1, true)
	require.NoError(t, err)

	_ = staking.EndBlocker(ctx, stakingKeeper)

	// period 0 passes and the threshold is not met
	vva.MissingSignCount[0] = 50
	vva.MissingSignCount[1] = 100
	ak.SetAccount(ctx, vva)
	keeper.UpdateVestedCoinsProgress(ctx, vva.Address, 0)
	vva = keeper.GetAccountFromAuthKeeper(ctx, vva.Address)
	// require that period 0 coins have become debt
	require.Equal(t, sdk.NewCoins(sdk.NewInt64Coin(stakeDenom, 30000000)), vva.DebtAfterFailedVesting)

	initialBalance := ak.GetAccount(ctx, TestAddrs[2]).GetCoins()
	expectedBalance := initialBalance.Add(vva.DebtAfterFailedVesting)
	// Context needs the block time because bank keeper calls 'SpendableCoins' by getting the header from the context.
	ctx = ctx.WithBlockTime(now.Add(12 * time.Hour))
	keeper.HandleVestingDebt(ctx, vva.Address, now.Add(12*time.Hour))
	// in the case when the return address is, set require that return address balance has increased by the debt amount
	require.Equal(t, expectedBalance, ak.GetAccount(ctx, TestAddrs[2]).GetCoins())
	// require that there is still one delegation
	delegations := 0
	stakingKeeper.IterateDelegations(ctx, vva.Address, func(index int64, d stakingexported.DelegationI) (stop bool) {
		delegations++
		return false
	})
	require.Equal(t, 1, delegations)
	vva = keeper.GetAccountFromAuthKeeper(ctx, vva.Address)
	//require that debt is now zero
	require.Equal(t, sdk.Coins(nil), vva.DebtAfterFailedVesting)
}
