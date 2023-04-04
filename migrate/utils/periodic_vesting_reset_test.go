package utils

import (
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	"github.com/stretchr/testify/assert"
)

func createVestingAccount(balance sdk.Coins, vestingStart time.Time, vestingPeriods vestingtypes.Periods) *vestingtypes.PeriodicVestingAccount {
	key := secp256k1.GenPrivKey()
	pub := key.PubKey()
	addr := sdk.AccAddress(pub.Address())
	acc := authtypes.NewBaseAccount(addr, pub, 1, 1)

	originalVesting := sdk.NewCoins()
	for _, vp := range vestingPeriods {
		originalVesting = originalVesting.Add(vp.Amount...)
	}

	return vestingtypes.NewPeriodicVestingAccount(acc, originalVesting, vestingStart.Unix(), vestingPeriods)
}

func TestResetPeriodVestingAccount_NoVestingPeriods(t *testing.T) {
	vestingStartTime := time.Now().Add(-1 * time.Hour)
	vacc := createVestingAccount(sdk.Coins{}, vestingStartTime, vestingtypes.Periods{})

	newVestingStartTime := vestingStartTime.Add(time.Hour)

	ResetPeriodicVestingAccount(vacc, newVestingStartTime)

	assert.Equal(t, sdk.Coins{}, vacc.OriginalVesting, "expected original vesting to be zero")
	assert.Equal(t, newVestingStartTime.Unix(), vacc.StartTime, "expected vesting start time to be updated")
	assert.Equal(t, newVestingStartTime.Unix(), vacc.EndTime, "expected vesting end time to be updated")
	assert.Equal(t, []vestingtypes.Period{}, vacc.VestingPeriods, "expected vesting periods to be empty")
}

func TestResetPeriodVestingAccount_SingleVestingPeriod_Vested(t *testing.T) {
	balance := sdk.NewCoins(sdk.NewCoin("ukava", sdkmath.NewInt(1e6)))
	vestingStartTime := time.Now().Add(-30 * 24 * time.Hour) // 30 days in past

	periods := vestingtypes.Periods{
		vestingtypes.Period{
			Length: 15 * 24 * 60 * 60, // 15 days (-15 days in past)
			Amount: balance,
		},
	}

	vacc := createVestingAccount(balance, vestingStartTime, periods)

	newVestingStartTime := vestingStartTime.Add(30 * 24 * time.Hour)

	ResetPeriodicVestingAccount(vacc, newVestingStartTime)

	assert.Equal(t, sdk.Coins{}, vacc.OriginalVesting, "expected original vesting to be zero")
	assert.Equal(t, newVestingStartTime.Unix(), vacc.StartTime, "expected vesting start time to be updated")
	assert.Equal(t, newVestingStartTime.Unix(), vacc.EndTime, "expected vesting end time to be updated")
	assert.Equal(t, []vestingtypes.Period{}, vacc.VestingPeriods, "expected vesting periods to be empty")
}

func TestResetPeriodVestingAccount_SingleVestingPeriod_Vesting(t *testing.T) {
	balance := sdk.NewCoins(sdk.NewCoin("ukava", sdkmath.NewInt(1e6)))
	vestingStartTime := time.Now().Add(-30 * 24 * time.Hour) // 30 days in past

	periods := vestingtypes.Periods{
		vestingtypes.Period{
			Length: 45 * 24 * 60 * 60, // 45 days
			Amount: balance,
		},
	}

	vacc := createVestingAccount(balance, vestingStartTime, periods)

	newVestingStartTime := vestingStartTime.Add(30 * 24 * time.Hour)

	ResetPeriodicVestingAccount(vacc, newVestingStartTime)

	// new period length 15 days
	expectedEndtime := newVestingStartTime.Add(15 * 24 * time.Hour).Unix()
	// new period length changed, amount unchanged
	expectedPeriods := []vestingtypes.Period{
		{
			Length: 15 * 24 * 60 * 60, // 15 days
			Amount: balance,
		},
	}

	assert.Equal(t, balance, vacc.OriginalVesting, "expected original vesting to be unchanged")
	assert.Equal(t, newVestingStartTime.Unix(), vacc.StartTime, "expected vesting start time to be updated")
	assert.Equal(t, expectedEndtime, vacc.EndTime, "expected vesting end time end at last period")
	assert.Equal(t, expectedPeriods, vacc.VestingPeriods, "expected vesting periods to be updated")
}

func TestResetPeriodVestingAccount_SingleVestingPeriod_ExactStartTime(t *testing.T) {
	balance := sdk.NewCoins(sdk.NewCoin("ukava", sdkmath.NewInt(1e6)))
	vestingStartTime := time.Now().Add(-30 * 24 * time.Hour) // 30 days in past

	periods := vestingtypes.Periods{
		vestingtypes.Period{
			Length: 30 * 24 * 60 * 60, // 30 days - exact on the start time
			Amount: balance,
		},
	}

	vacc := createVestingAccount(balance, vestingStartTime, periods)

	newVestingStartTime := vestingStartTime.Add(30 * 24 * time.Hour)

	ResetPeriodicVestingAccount(vacc, newVestingStartTime)

	// new period length is 0
	expectedEndtime := newVestingStartTime.Unix()
	// new period length changed, amount unchanged
	expectedPeriods := []vestingtypes.Period{}

	assert.Equal(t, sdk.Coins{}, vacc.OriginalVesting, "expected original vesting to be unchanged")
	assert.Equal(t, newVestingStartTime.Unix(), vacc.StartTime, "expected vesting start time to be updated")
	assert.Equal(t, expectedEndtime, vacc.EndTime, "expected vesting end time end at last period")
	assert.Equal(t, expectedPeriods, vacc.VestingPeriods, "expected vesting periods to be updated")
}

func TestResetPeriodVestingAccount_MultiplePeriods(t *testing.T) {
	balance := sdk.NewCoins(sdk.NewCoin("ukava", sdkmath.NewInt(4e6)))
	vestingStartTime := time.Now().Add(-30 * 24 * time.Hour) // 30 days in past

	periods := vestingtypes.Periods{
		vestingtypes.Period{
			Length: 15 * 24 * 60 * 60, // -15 days - vested
			Amount: sdk.NewCoins(sdk.NewCoin("ukava", sdkmath.NewInt(1e6))),
		},
		vestingtypes.Period{
			Length: 15 * 24 * 60 * 60, // 0 days - exact on the start time
			Amount: sdk.NewCoins(sdk.NewCoin("ukava", sdkmath.NewInt(1e6))),
		},
		vestingtypes.Period{
			Length: 15 * 24 * 60 * 60, // +15 days - vesting
			Amount: sdk.NewCoins(sdk.NewCoin("ukava", sdkmath.NewInt(1e6))),
		},
		vestingtypes.Period{
			Length: 15 * 24 * 60 * 60, // +30 days - vesting
			Amount: sdk.NewCoins(sdk.NewCoin("ukava", sdkmath.NewInt(1e6))),
		},
	}

	vacc := createVestingAccount(balance, vestingStartTime, periods)

	newVestingStartTime := vestingStartTime.Add(30 * 24 * time.Hour)

	ResetPeriodicVestingAccount(vacc, newVestingStartTime)

	// new period length 15 days
	expectedEndtime := newVestingStartTime.Add(30 * 24 * time.Hour).Unix()
	// new period length changed, amount unchanged
	expectedPeriods := []vestingtypes.Period{
		{
			Length: 15 * 24 * 60 * 60, // 15 days
			Amount: sdk.NewCoins(sdk.NewCoin("ukava", sdkmath.NewInt(1e6))),
		},
		{
			Length: 15 * 24 * 60 * 60, // 15 days
			Amount: sdk.NewCoins(sdk.NewCoin("ukava", sdkmath.NewInt(1e6))),
		},
	}

	assert.Equal(t, sdk.NewCoins(sdk.NewCoin("ukava", sdkmath.NewInt(2e6))), vacc.OriginalVesting, "expected original vesting to be updated")
	assert.Equal(t, newVestingStartTime.Unix(), vacc.StartTime, "expected vesting start time to be updated")
	assert.Equal(t, expectedEndtime, vacc.EndTime, "expected vesting end time end at last period")
	assert.Equal(t, expectedPeriods, vacc.VestingPeriods, "expected vesting periods to be updated")
}

func TestResetPeriodVestingAccount_DelegatedVesting_GreaterThanVesting(t *testing.T) {
	balance := sdk.NewCoins(sdk.NewCoin("ukava", sdkmath.NewInt(3e6)))
	vestingStartTime := time.Now().Add(-30 * 24 * time.Hour) // 30 days in past

	periods := vestingtypes.Periods{
		vestingtypes.Period{
			Length: 15 * 24 * 60 * 60, // -15 days - vested
			Amount: sdk.NewCoins(sdk.NewCoin("ukava", sdkmath.NewInt(1e6))),
		},
		vestingtypes.Period{
			Length: 15 * 24 * 60 * 60, // 0 days - exact on the start time
			Amount: sdk.NewCoins(sdk.NewCoin("ukava", sdkmath.NewInt(1e6))),
		},
		vestingtypes.Period{
			Length: 15 * 24 * 60 * 60, // +15 days - vesting
			Amount: sdk.NewCoins(sdk.NewCoin("ukava", sdkmath.NewInt(1e6))),
		},
	}

	vacc := createVestingAccount(balance, vestingStartTime, periods)
	vacc.TrackDelegation(vestingStartTime, balance, balance)

	newVestingStartTime := vestingStartTime.Add(30 * 24 * time.Hour)
	ResetPeriodicVestingAccount(vacc, newVestingStartTime)

	assert.Equal(t, sdk.NewCoins(sdk.NewCoin("ukava", sdkmath.NewInt(2e6))), vacc.DelegatedFree, "expected delegated free to be updated")
	assert.Equal(t, sdk.NewCoins(sdk.NewCoin("ukava", sdkmath.NewInt(1e6))), vacc.DelegatedVesting, "expected delegated vesting to be updated")
}

func TestResetPeriodVestingAccount_DelegatedVesting_LessThanVested(t *testing.T) {
	balance := sdk.NewCoins(sdk.NewCoin("ukava", sdkmath.NewInt(3e6)))
	vestingStartTime := time.Now().Add(-30 * 24 * time.Hour) // 30 days in past

	periods := vestingtypes.Periods{
		vestingtypes.Period{
			Length: 15 * 24 * 60 * 60, // -15 days - vested
			Amount: sdk.NewCoins(sdk.NewCoin("ukava", sdkmath.NewInt(1e6))),
		},
		vestingtypes.Period{
			Length: 15 * 24 * 60 * 60, // 0 days - exact on the start time
			Amount: sdk.NewCoins(sdk.NewCoin("ukava", sdkmath.NewInt(1e6))),
		},
		vestingtypes.Period{
			Length: 15 * 24 * 60 * 60, // +15 days - vesting
			Amount: sdk.NewCoins(sdk.NewCoin("ukava", sdkmath.NewInt(1e6))),
		},
	}

	vacc := createVestingAccount(balance, vestingStartTime, periods)
	vacc.TrackDelegation(vestingStartTime, balance, sdk.NewCoins(sdk.NewCoin("ukava", sdkmath.NewInt(1e6))))

	newVestingStartTime := vestingStartTime.Add(30 * 24 * time.Hour)
	ResetPeriodicVestingAccount(vacc, newVestingStartTime)

	assert.Equal(t, sdk.Coins(nil), vacc.DelegatedFree, "expected delegrated free to be unmodified")
	assert.Equal(t, sdk.NewCoins(sdk.NewCoin("ukava", sdkmath.NewInt(1e6))), vacc.DelegatedVesting, "expected delegated vesting to be unmodified")
}
