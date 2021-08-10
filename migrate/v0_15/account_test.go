package v0_15

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	auth "github.com/cosmos/cosmos-sdk/x/auth"
	//"github.com/tendermint/tendermint/crypto"
	vesting "github.com/cosmos/cosmos-sdk/x/auth/vesting"
	"github.com/stretchr/testify/assert"
	"github.com/tendermint/tendermint/crypto/secp256k1"
)

func createVestingAccount(balance sdk.Coins, vestingStart time.Time, vestingPeriods vesting.Periods) *vesting.PeriodicVestingAccount {
	key := secp256k1.GenPrivKey()
	pub := key.PubKey()
	addr := sdk.AccAddress(pub.Address())
	acc := auth.NewBaseAccount(addr, balance, pub, 1, 1)

	return vesting.NewPeriodicVestingAccount(acc, vestingStart.Unix(), vestingPeriods)
}

func TestMigrateAccount_BaseAccount(t *testing.T) {
	balance := sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(1e6)))

	key := secp256k1.GenPrivKey()
	pub := key.PubKey()
	addr := sdk.AccAddress(pub.Address())
	acc := auth.NewBaseAccount(addr, balance, pub, 1, 1)

	migratedAcc := MigrateAccount(acc, time.Now())

	assert.Equal(t, acc, migratedAcc)
}

func TestMigrateAccount_PeriodicVestingAccount_NoPeriods(t *testing.T) {
	balance := sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(1e6)))

	key := secp256k1.GenPrivKey()
	pub := key.PubKey()
	addr := sdk.AccAddress(pub.Address())
	acc := auth.NewBaseAccount(addr, balance, pub, 1, 1)

	vacc := vesting.NewPeriodicVestingAccount(acc, time.Now().Unix(), vesting.Periods{})
	migratedAcc := MigrateAccount(vacc, time.Now())

	assert.Equal(t, acc, migratedAcc)
}

func TestMigrateAccount_PeriodicVestingAccount_Vesting(t *testing.T) {
	balance := sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(3e6)))
	vestingStartTime := time.Now().Add(-30 * 24 * time.Hour) // 30 days in past

	periods := vesting.Periods{
		vesting.Period{
			Length: 15 * 24 * 60 * 60, // -15 days - vested
			Amount: sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(1e6))),
		},
		vesting.Period{
			Length: 45 * 24 * 60 * 60, // +15 days - vested
			Amount: sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(1e6))),
		},
	}

	vacc := createVestingAccount(balance, vestingStartTime, periods)
	genesisTime := vestingStartTime.Add(30 * 24 * time.Hour)

	MigrateAccount(vacc, genesisTime)

	assert.Equal(t, genesisTime.Unix(), vacc.StartTime)
	assert.Equal(t, 1, len(vacc.VestingPeriods))
}

func TestResetPeriodVestingAccount_NoVestingPeriods(t *testing.T) {
	vestingStartTime := time.Now().Add(-1 * time.Hour)
	vacc := createVestingAccount(sdk.Coins{}, vestingStartTime, vesting.Periods{})

	newVestingStartTime := vestingStartTime.Add(time.Hour)
	spendableBefore := vacc.SpendableCoins(newVestingStartTime)

	ResetPeriodicVestingAccount(vacc, newVestingStartTime)

	assert.Equal(t, sdk.Coins{}, vacc.OriginalVesting, "expected original vesting to be zero")
	assert.Equal(t, newVestingStartTime.Unix(), vacc.StartTime, "expected vesting start time to be updated")
	assert.Equal(t, newVestingStartTime.Unix(), vacc.EndTime, "expected vesting end time to be updated")
	assert.Equal(t, vesting.Periods{}, vacc.VestingPeriods, "expected vesting periods to be empty")
	assert.Equal(t, spendableBefore, vacc.SpendableCoins(newVestingStartTime), "expected spendable coins to be unchanged")
}

func TestResetPeriodVestingAccount_SingleVestingPeriod_Vested(t *testing.T) {
	balance := sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(1e6)))
	vestingStartTime := time.Now().Add(-30 * 24 * time.Hour) // 30 days in past

	periods := vesting.Periods{
		vesting.Period{
			Length: 15 * 24 * 60 * 60, // 15 days (-15 days in past)
			Amount: balance,
		},
	}

	vacc := createVestingAccount(balance, vestingStartTime, periods)

	newVestingStartTime := vestingStartTime.Add(30 * 24 * time.Hour)
	spendableBefore := vacc.SpendableCoins(newVestingStartTime)

	ResetPeriodicVestingAccount(vacc, newVestingStartTime)

	assert.Equal(t, sdk.Coins{}, vacc.OriginalVesting, "expected original vesting to be zero")
	assert.Equal(t, newVestingStartTime.Unix(), vacc.StartTime, "expected vesting start time to be updated")
	assert.Equal(t, newVestingStartTime.Unix(), vacc.EndTime, "expected vesting end time to be updated")
	assert.Equal(t, vesting.Periods{}, vacc.VestingPeriods, "expected vesting periods to be empty")
	assert.Equal(t, spendableBefore, vacc.SpendableCoins(newVestingStartTime), "expected spendable coins to be unchanged")
}

func TestResetPeriodVestingAccount_SingleVestingPeriod_Vesting(t *testing.T) {
	balance := sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(1e6)))
	vestingStartTime := time.Now().Add(-30 * 24 * time.Hour) // 30 days in past

	periods := vesting.Periods{
		vesting.Period{
			Length: 45 * 24 * 60 * 60, // 45 days
			Amount: balance,
		},
	}

	vacc := createVestingAccount(balance, vestingStartTime, periods)

	newVestingStartTime := vestingStartTime.Add(30 * 24 * time.Hour)
	spendableBefore := vacc.SpendableCoins(newVestingStartTime)

	ResetPeriodicVestingAccount(vacc, newVestingStartTime)

	// new period length 15 days
	expectedEndtime := newVestingStartTime.Add(15 * 24 * time.Hour).Unix()
	// new period length changed, amount unchanged
	expectedPeriods := vesting.Periods{
		vesting.Period{
			Length: 15 * 24 * 60 * 60, // 15 days
			Amount: balance,
		},
	}

	assert.Equal(t, balance, vacc.OriginalVesting, "expected original vesting to be unchanged")
	assert.Equal(t, newVestingStartTime.Unix(), vacc.StartTime, "expected vesting start time to be updated")
	assert.Equal(t, expectedEndtime, vacc.EndTime, "expected vesting end time end at last period")
	assert.Equal(t, expectedPeriods, vacc.VestingPeriods, "expected vesting periods to be updated")
	assert.Equal(t, spendableBefore, vacc.SpendableCoins(newVestingStartTime), "expected spendable coins to be unchanged")
}

func TestResetPeriodVestingAccount_SingleVestingPeriod_ExactStartTime(t *testing.T) {
	balance := sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(1e6)))
	vestingStartTime := time.Now().Add(-30 * 24 * time.Hour) // 30 days in past

	periods := vesting.Periods{
		vesting.Period{
			Length: 30 * 24 * 60 * 60, // 30 days - exact on the start time
			Amount: balance,
		},
	}

	vacc := createVestingAccount(balance, vestingStartTime, periods)

	newVestingStartTime := vestingStartTime.Add(30 * 24 * time.Hour)
	spendableBefore := vacc.SpendableCoins(newVestingStartTime)

	ResetPeriodicVestingAccount(vacc, newVestingStartTime)

	// new period length is 0
	expectedEndtime := newVestingStartTime.Unix()
	// new period length changed, amount unchanged
	expectedPeriods := vesting.Periods{}

	assert.Equal(t, sdk.Coins{}, vacc.OriginalVesting, "expected original vesting to be unchanged")
	assert.Equal(t, newVestingStartTime.Unix(), vacc.StartTime, "expected vesting start time to be updated")
	assert.Equal(t, expectedEndtime, vacc.EndTime, "expected vesting end time end at last period")
	assert.Equal(t, expectedPeriods, vacc.VestingPeriods, "expected vesting periods to be updated")
	assert.Equal(t, spendableBefore, vacc.SpendableCoins(newVestingStartTime), "expected spendable coins to be unchanged")
}

func TestResetPeriodVestingAccount_MultiplePeriods(t *testing.T) {
	balance := sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(3e6)))
	vestingStartTime := time.Now().Add(-30 * 24 * time.Hour) // 30 days in past

	periods := vesting.Periods{
		vesting.Period{
			Length: 15 * 24 * 60 * 60, // -15 days - vested
			Amount: sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(1e6))),
		},
		vesting.Period{
			Length: 30 * 24 * 60 * 60, // 0 days - exact on the start time
			Amount: sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(1e6))),
		},
		vesting.Period{
			Length: 45 * 24 * 60 * 60, // +15 days - vested
			Amount: sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(1e6))),
		},
	}

	vacc := createVestingAccount(balance, vestingStartTime, periods)

	newVestingStartTime := vestingStartTime.Add(30 * 24 * time.Hour)
	spendableBefore := vacc.SpendableCoins(newVestingStartTime)

	ResetPeriodicVestingAccount(vacc, newVestingStartTime)

	// new period length 15 days
	expectedEndtime := newVestingStartTime.Add(15 * 24 * time.Hour).Unix()
	// new period length changed, amount unchanged
	expectedPeriods := vesting.Periods{
		vesting.Period{
			Length: 0, // 0 days
			Amount: sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(1e6))),
		},
		vesting.Period{
			Length: 15 * 24 * 60 * 60, // 15 days
			Amount: sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(1e6))),
		},
	}

	assert.Equal(t, sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(2e6))), vacc.OriginalVesting, "expected original vesting to be unchanged")
	assert.Equal(t, newVestingStartTime.Unix(), vacc.StartTime, "expected vesting start time to be updated")
	assert.Equal(t, expectedEndtime, vacc.EndTime, "expected vesting end time end at last period")
	assert.Equal(t, expectedPeriods, vacc.VestingPeriods, "expected vesting periods to be updated")
	assert.Equal(t, spendableBefore, vacc.SpendableCoins(newVestingStartTime), "expected spendable coins to be unchanged")
}

func TestResetPeriodVestingAccount_DelegatedVesting_GreaterThanVesting(t *testing.T) {
	balance := sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(3e6)))
	vestingStartTime := time.Now().Add(-30 * 24 * time.Hour) // 30 days in past

	periods := vesting.Periods{
		vesting.Period{
			Length: 15 * 24 * 60 * 60, // -15 days - vested
			Amount: sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(1e6))),
		},
		vesting.Period{
			Length: 30 * 24 * 60 * 60, // 0 days - exact on the start time
			Amount: sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(1e6))),
		},
		vesting.Period{
			Length: 45 * 24 * 60 * 60, // +15 days - vested
			Amount: sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(1e6))),
		},
	}

	vacc := createVestingAccount(balance, vestingStartTime, periods)
	vacc.TrackDelegation(vestingStartTime, balance)

	newVestingStartTime := vestingStartTime.Add(30 * 24 * time.Hour)
	ResetPeriodicVestingAccount(vacc, newVestingStartTime)

	assert.Equal(t, sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(1e6))), vacc.DelegatedFree)
	assert.Equal(t, sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(2e6))), vacc.DelegatedVesting)
}

func TestResetPeriodVestingAccount_DelegatedVesting_LessThanVested(t *testing.T) {
	balance := sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(3e6)))
	vestingStartTime := time.Now().Add(-30 * 24 * time.Hour) // 30 days in past

	periods := vesting.Periods{
		vesting.Period{
			Length: 15 * 24 * 60 * 60, // -15 days - vested
			Amount: sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(1e6))),
		},
		vesting.Period{
			Length: 30 * 24 * 60 * 60, // 0 days - exact on the start time
			Amount: sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(1e6))),
		},
		vesting.Period{
			Length: 45 * 24 * 60 * 60, // +15 days - vested
			Amount: sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(1e6))),
		},
	}

	vacc := createVestingAccount(balance, vestingStartTime, periods)
	vacc.TrackDelegation(vestingStartTime, sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(1e6))))

	newVestingStartTime := vestingStartTime.Add(30 * 24 * time.Hour)
	ResetPeriodicVestingAccount(vacc, newVestingStartTime)

	assert.Equal(t, sdk.Coins(nil), vacc.DelegatedFree)
	assert.Equal(t, sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(1e6))), vacc.DelegatedVesting)
}
