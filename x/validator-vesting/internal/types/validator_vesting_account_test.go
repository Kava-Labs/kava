package types

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	tmtime "github.com/tendermint/tendermint/types/time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authexported "github.com/cosmos/cosmos-sdk/x/auth/exported"
	"github.com/cosmos/cosmos-sdk/x/auth/vesting"
)

var (
	stakeDenom = "stake"
	feeDenom   = "fee"
)

func TestGetVestedCoinsValidatorVestingAcc(t *testing.T) {
	now := tmtime.Now()
	periods := vesting.VestingPeriods{
		vesting.VestingPeriod{PeriodLength: int64(12 * 60 * 60), VestingAmount: sdk.Coins{sdk.NewInt64Coin(feeDenom, 500), sdk.NewInt64Coin(stakeDenom, 50)}},
		vesting.VestingPeriod{PeriodLength: int64(6 * 60 * 60), VestingAmount: sdk.Coins{sdk.NewInt64Coin(feeDenom, 250), sdk.NewInt64Coin(stakeDenom, 25)}},
		vesting.VestingPeriod{PeriodLength: int64(6 * 60 * 60), VestingAmount: sdk.Coins{sdk.NewInt64Coin(feeDenom, 250), sdk.NewInt64Coin(stakeDenom, 25)}},
	}

	testAddr := CreateTestAddrs(1)[0]
	testPk := CreateTestPubKeys(1)[0]
	testConsAddr := sdk.ConsAddress(testPk.Address())
	origCoins := sdk.Coins{sdk.NewInt64Coin(feeDenom, 1000), sdk.NewInt64Coin(stakeDenom, 100)}
	bacc := auth.NewBaseAccountWithAddress(testAddr)
	bacc.SetCoins(origCoins)
	vva := NewValidatorVestingAccount(&bacc, now.Unix(), periods, testConsAddr, nil, 90)

	// require no coins vested at the beginning of the vesting schedule
	vestedCoins := vva.GetVestedCoins(now)
	require.Nil(t, vestedCoins)

	// require no coins vested during first vesting period
	vestedCoins = vva.GetVestedCoins(now.Add(6 * time.Hour))
	require.Nil(t, vestedCoins)

	// require 50% of coins vested after successful period 1 vesting
	vva.VestingPeriodProgress[0] = 1
	vestedCoins = vva.GetVestedCoins(now.Add(12 * time.Hour))
	require.Equal(t, sdk.Coins{sdk.NewInt64Coin(feeDenom, 500), sdk.NewInt64Coin(stakeDenom, 50)}, vestedCoins)

	// require  no coins vested after unsuccessful period 1 vesting
	vva.VestingPeriodProgress[0] = 0
	vestedCoins = vva.GetVestedCoins(now.Add(12 * time.Hour))
	require.Nil(t, vestedCoins)

	// require period 2 coins don't vest until period is over
	vva.VestingPeriodProgress[0] = 1
	// even if the vesting period was somehow successful, should still only return 50% of coins as vested, since the second vesting period hasn't completed.
	vva.VestingPeriodProgress[1] = 1
	vestedCoins = vva.GetVestedCoins(now.Add(15 * time.Hour))
	require.Equal(t, sdk.Coins{sdk.NewInt64Coin(feeDenom, 500), sdk.NewInt64Coin(stakeDenom, 50)}, vestedCoins)

	// require 75% of coins vested after successful period 2
	vva.VestingPeriodProgress[0] = 1
	vva.VestingPeriodProgress[1] = 1
	vestedCoins = vva.GetVestedCoins(now.Add(18 * time.Hour))
	require.Equal(t,
		sdk.Coins{
			sdk.NewInt64Coin(feeDenom, 750), sdk.NewInt64Coin(stakeDenom, 75)}, vestedCoins)

	// require 50% of coins vested after successful period 1 and unsuccessful period 2
	vva.VestingPeriodProgress[0] = 1
	vva.VestingPeriodProgress[1] = 0
	vestedCoins = vva.GetVestedCoins(now.Add(18 * time.Hour))
	require.Equal(t, sdk.Coins{sdk.NewInt64Coin(feeDenom, 500), sdk.NewInt64Coin(stakeDenom, 50)}, vestedCoins)

	// require 100% of coins vested after all periods complete successfully
	vva.VestingPeriodProgress[0] = 1
	vva.VestingPeriodProgress[1] = 1
	vva.VestingPeriodProgress[2] = 1

	vestedCoins = vva.GetVestedCoins(now.Add(48 * time.Hour))
	require.Equal(t, origCoins, vestedCoins)
}

func TestGetVestingCoinsValidatorVestingAcc(t *testing.T) {
	now := tmtime.Now()
	periods := vesting.VestingPeriods{
		vesting.VestingPeriod{PeriodLength: int64(12 * 60 * 60), VestingAmount: sdk.Coins{sdk.NewInt64Coin(feeDenom, 500), sdk.NewInt64Coin(stakeDenom, 50)}},
		vesting.VestingPeriod{PeriodLength: int64(6 * 60 * 60), VestingAmount: sdk.Coins{sdk.NewInt64Coin(feeDenom, 250), sdk.NewInt64Coin(stakeDenom, 25)}},
		vesting.VestingPeriod{PeriodLength: int64(6 * 60 * 60), VestingAmount: sdk.Coins{sdk.NewInt64Coin(feeDenom, 250), sdk.NewInt64Coin(stakeDenom, 25)}},
	}

	testAddr := CreateTestAddrs(1)[0]
	testPk := CreateTestPubKeys(1)[0]
	testConsAddr := sdk.ConsAddress(testPk.Address())
	origCoins := sdk.Coins{sdk.NewInt64Coin(feeDenom, 1000), sdk.NewInt64Coin(stakeDenom, 100)}
	bacc := auth.NewBaseAccountWithAddress(testAddr)
	bacc.SetCoins(origCoins)
	vva := NewValidatorVestingAccount(&bacc, now.Unix(), periods, testConsAddr, nil, 90)

	// require all coins vesting at the beginning of the vesting schedule
	vestingCoins := vva.GetVestingCoins(now)
	require.Equal(t, origCoins, vestingCoins)

	// require all coins vesting during first vesting period
	vestingCoins = vva.GetVestingCoins(now.Add(6 * time.Hour))
	require.Equal(t, origCoins, vestingCoins)

	// require 50% of coins vesting after successful period 1 vesting
	vva.VestingPeriodProgress[0] = 1
	vestingCoins = vva.GetVestingCoins(now.Add(12 * time.Hour))
	require.Equal(t, sdk.Coins{sdk.NewInt64Coin(feeDenom, 500), sdk.NewInt64Coin(stakeDenom, 50)}, vestingCoins)

	// require 50% of coins vesting after unsuccessful period 1 vesting
	vva.VestingPeriodProgress[0] = 0
	vestingCoins = vva.GetVestingCoins(now.Add(12 * time.Hour))
	require.Equal(t, sdk.Coins{sdk.NewInt64Coin(feeDenom, 500), sdk.NewInt64Coin(stakeDenom, 50)}, vestingCoins)

	// require period 2 coins still vesting until period is over
	vva.VestingPeriodProgress[0] = 1
	// should never happen, but still won't affect vesting balance
	vva.VestingPeriodProgress[1] = 1
	vestingCoins = vva.GetVestingCoins(now.Add(15 * time.Hour))
	require.Equal(t, sdk.Coins{sdk.NewInt64Coin(feeDenom, 500), sdk.NewInt64Coin(stakeDenom, 50)}, vestingCoins)

	// require 25% of coins vesting after successful period 2
	vva.VestingPeriodProgress[0] = 1
	vva.VestingPeriodProgress[1] = 1
	vestingCoins = vva.GetVestingCoins(now.Add(18 * time.Hour))
	require.Equal(t,
		sdk.Coins{
			sdk.NewInt64Coin(feeDenom, 250), sdk.NewInt64Coin(stakeDenom, 25)}, vestingCoins)

	// require 25% of coins vesting after successful period 1 and unsuccessful period 2
	vva.VestingPeriodProgress[0] = 1
	vva.VestingPeriodProgress[1] = 0
	vestingCoins = vva.GetVestingCoins(now.Add(18 * time.Hour))
	require.Equal(t, sdk.Coins{sdk.NewInt64Coin(feeDenom, 250), sdk.NewInt64Coin(stakeDenom, 25)}, vestingCoins)

	// require no coins vesting after all periods complete successfully
	vva.VestingPeriodProgress[0] = 1
	vva.VestingPeriodProgress[1] = 1
	vva.VestingPeriodProgress[2] = 1

	vestingCoins = vva.GetVestingCoins(now.Add(48 * time.Hour))
	require.Nil(t, vestingCoins)

	// require no coins vesting after all periods complete unsuccessfully
	vva.VestingPeriodProgress[0] = 0
	vva.VestingPeriodProgress[1] = 0
	vva.VestingPeriodProgress[2] = 0

	vestingCoins = vva.GetVestingCoins(now.Add(48 * time.Hour))
	require.Nil(t, vestingCoins)
}

func TestSpendableCoinsValidatorVestingAccount(t *testing.T) {
	now := tmtime.Now()
	periods := vesting.VestingPeriods{
		vesting.VestingPeriod{PeriodLength: int64(12 * 60 * 60), VestingAmount: sdk.Coins{sdk.NewInt64Coin(feeDenom, 500), sdk.NewInt64Coin(stakeDenom, 50)}},
		vesting.VestingPeriod{PeriodLength: int64(6 * 60 * 60), VestingAmount: sdk.Coins{sdk.NewInt64Coin(feeDenom, 250), sdk.NewInt64Coin(stakeDenom, 25)}},
		vesting.VestingPeriod{PeriodLength: int64(6 * 60 * 60), VestingAmount: sdk.Coins{sdk.NewInt64Coin(feeDenom, 250), sdk.NewInt64Coin(stakeDenom, 25)}},
	}

	testAddr := CreateTestAddrs(1)[0]
	testPk := CreateTestPubKeys(1)[0]
	testConsAddr := sdk.ConsAddress(testPk.Address())
	origCoins := sdk.Coins{sdk.NewInt64Coin(feeDenom, 1000), sdk.NewInt64Coin(stakeDenom, 100)}
	bacc := auth.NewBaseAccountWithAddress(testAddr)
	bacc.SetCoins(origCoins)
	vva := NewValidatorVestingAccount(&bacc, now.Unix(), periods, testConsAddr, nil, 90)

	// require that there exist no spendable coins at the beginning of the vesting schedule
	spendableCoins := vva.SpendableCoins(now)
	require.Nil(t, spendableCoins)

	// require that all vested coins (50%) are spendable when period 1 completes successfully
	vva.VestingPeriodProgress[0] = 1
	spendableCoins = vva.SpendableCoins(now.Add(12 * time.Hour))
	require.Equal(t, sdk.Coins{sdk.NewInt64Coin(feeDenom, 500), sdk.NewInt64Coin(stakeDenom, 50)}, spendableCoins)

	// require that there exist no spendable coins after period 1 completes unsuccessfully.
	vva.VestingPeriodProgress[0] = 0
	spendableCoins = vva.SpendableCoins(now)
	require.Nil(t, spendableCoins)

	// receive some coins
	recvAmt := sdk.Coins{sdk.NewInt64Coin(stakeDenom, 50)}
	vva.SetCoins(vva.GetCoins().Add(recvAmt))

	// require that all vested coins (50%) are spendable plus any received after period 1 completes unsuccessfully
	vva.VestingPeriodProgress[0] = 1
	spendableCoins = vva.SpendableCoins(now.Add(12 * time.Hour))
	require.Equal(t, sdk.Coins{sdk.NewInt64Coin(feeDenom, 500), sdk.NewInt64Coin(stakeDenom, 100)}, spendableCoins)

	// spend all spendable coins
	vva.SetCoins(vva.GetCoins().Sub(spendableCoins))

	// require that no more coins are spendable
	spendableCoins = vva.SpendableCoins(now.Add(12 * time.Hour))
	require.Nil(t, spendableCoins)
}

func TestTrackDelegationValidatorVestingAcc(t *testing.T) {
	now := tmtime.Now()
	periods := vesting.VestingPeriods{
		vesting.VestingPeriod{PeriodLength: int64(12 * 60 * 60), VestingAmount: sdk.Coins{sdk.NewInt64Coin(feeDenom, 500), sdk.NewInt64Coin(stakeDenom, 50)}},
		vesting.VestingPeriod{PeriodLength: int64(6 * 60 * 60), VestingAmount: sdk.Coins{sdk.NewInt64Coin(feeDenom, 250), sdk.NewInt64Coin(stakeDenom, 25)}},
		vesting.VestingPeriod{PeriodLength: int64(6 * 60 * 60), VestingAmount: sdk.Coins{sdk.NewInt64Coin(feeDenom, 250), sdk.NewInt64Coin(stakeDenom, 25)}},
	}

	testAddr := CreateTestAddrs(1)[0]
	testPk := CreateTestPubKeys(1)[0]
	testConsAddr := sdk.ConsAddress(testPk.Address())
	origCoins := sdk.Coins{sdk.NewInt64Coin(feeDenom, 1000), sdk.NewInt64Coin(stakeDenom, 100)}
	bacc := auth.NewBaseAccountWithAddress(testAddr)
	bacc.SetCoins(origCoins)
	vva := NewValidatorVestingAccount(&bacc, now.Unix(), periods, testConsAddr, nil, 90)

	vva.TrackDelegation(now, origCoins)
	require.Equal(t, origCoins, vva.DelegatedVesting)
	require.Nil(t, vva.DelegatedFree)
	require.Nil(t, vva.GetCoins())

	// require the ability to delegate all vesting coins (50%) and all vested coins (50%)
	bacc.SetCoins(origCoins)
	vva = NewValidatorVestingAccount(&bacc, now.Unix(), periods, testConsAddr, nil, 90)
	vva.TrackDelegation(now.Add(12*time.Hour), sdk.Coins{sdk.NewInt64Coin(stakeDenom, 50)})
	require.Equal(t, sdk.Coins{sdk.NewInt64Coin(stakeDenom, 50)}, vva.DelegatedVesting)
	require.Nil(t, vva.DelegatedFree)

	vva.VestingPeriodProgress[0] = 1
	vva.TrackDelegation(now.Add(12*time.Hour), sdk.Coins{sdk.NewInt64Coin(stakeDenom, 50)})
	require.Equal(t, sdk.Coins{sdk.NewInt64Coin(stakeDenom, 50)}, vva.DelegatedVesting)
	require.Equal(t, sdk.Coins{sdk.NewInt64Coin(stakeDenom, 50)}, vva.DelegatedFree)
	require.Equal(t, sdk.Coins{sdk.NewInt64Coin(feeDenom, 1000)}, vva.GetCoins())

	// require no modifications when delegation amount is zero or not enough funds
	bacc.SetCoins(origCoins)
	vva = NewValidatorVestingAccount(&bacc, now.Unix(), periods, testConsAddr, nil, 90)
	require.Panics(t, func() {
		vva.TrackDelegation(now.Add(24*time.Hour), sdk.Coins{sdk.NewInt64Coin(stakeDenom, 1000000)})
	})
	require.Nil(t, vva.DelegatedVesting)
	require.Nil(t, vva.DelegatedFree)
	require.Equal(t, origCoins, vva.GetCoins())
}

func TestTrackUndelegationPeriodicVestingAcc(t *testing.T) {
	now := tmtime.Now()
	periods := vesting.VestingPeriods{
		vesting.VestingPeriod{PeriodLength: int64(12 * 60 * 60), VestingAmount: sdk.Coins{sdk.NewInt64Coin(feeDenom, 500), sdk.NewInt64Coin(stakeDenom, 50)}},
		vesting.VestingPeriod{PeriodLength: int64(6 * 60 * 60), VestingAmount: sdk.Coins{sdk.NewInt64Coin(feeDenom, 250), sdk.NewInt64Coin(stakeDenom, 25)}},
		vesting.VestingPeriod{PeriodLength: int64(6 * 60 * 60), VestingAmount: sdk.Coins{sdk.NewInt64Coin(feeDenom, 250), sdk.NewInt64Coin(stakeDenom, 25)}},
	}

	testAddr := CreateTestAddrs(1)[0]
	testPk := CreateTestPubKeys(1)[0]
	testConsAddr := sdk.ConsAddress(testPk.Address())
	origCoins := sdk.Coins{sdk.NewInt64Coin(feeDenom, 1000), sdk.NewInt64Coin(stakeDenom, 100)}
	bacc := auth.NewBaseAccountWithAddress(testAddr)
	bacc.SetCoins(origCoins)
	vva := NewValidatorVestingAccount(&bacc, now.Unix(), periods, testConsAddr, nil, 90)

	// require ability to delegate then undelegate all coins.
	vva.TrackDelegation(now, origCoins)
	vva.TrackUndelegation(origCoins)
	require.Nil(t, vva.DelegatedFree)
	require.Nil(t, vva.DelegatedVesting)
	require.Equal(t, origCoins, vva.GetCoins())

	// require the ability to delegate all coins after they have successfully vested
	bacc.SetCoins(origCoins)
	vva = NewValidatorVestingAccount(&bacc, now.Unix(), periods, testConsAddr, nil, 90)
	vva.VestingPeriodProgress[0] = 1
	vva.VestingPeriodProgress[1] = 1
	vva.VestingPeriodProgress[2] = 1
	vva.TrackDelegation(now.Add(24*time.Hour), origCoins)
	vva.TrackUndelegation(origCoins)
	require.Nil(t, vva.DelegatedFree)
	require.Nil(t, vva.DelegatedVesting)
	require.Equal(t, origCoins, vva.GetCoins())

	// require panic and no modifications when attempting to undelegate zero coins
	bacc.SetCoins(origCoins)
	vva = NewValidatorVestingAccount(&bacc, now.Unix(), periods, testConsAddr, nil, 90)
	require.Panics(t, func() {
		vva.TrackUndelegation(sdk.Coins{sdk.NewInt64Coin(stakeDenom, 0)})
	})
	require.Nil(t, vva.DelegatedFree)
	require.Nil(t, vva.DelegatedVesting)
	require.Equal(t, origCoins, vva.GetCoins())

	// successfuly vest period 1 and delegate to two validators
	vva = NewValidatorVestingAccount(&bacc, now.Unix(), periods, testConsAddr, nil, 90)
	vva.VestingPeriodProgress[0] = 1
	vva.TrackDelegation(now.Add(12*time.Hour), sdk.Coins{sdk.NewInt64Coin(stakeDenom, 50)})
	vva.TrackDelegation(now.Add(12*time.Hour), sdk.Coins{sdk.NewInt64Coin(stakeDenom, 50)})

	// undelegate from one validator that got slashed 50%
	vva.TrackUndelegation(sdk.Coins{sdk.NewInt64Coin(stakeDenom, 25)})
	require.Equal(t, sdk.Coins{sdk.NewInt64Coin(stakeDenom, 25)}, vva.DelegatedFree)
	require.Equal(t, sdk.Coins{sdk.NewInt64Coin(stakeDenom, 50)}, vva.DelegatedVesting)
	require.Equal(t, sdk.Coins{sdk.NewInt64Coin(feeDenom, 1000), sdk.NewInt64Coin(stakeDenom, 25)}, vva.GetCoins())

	// undelegate from the other validator that did not get slashed
	vva.TrackUndelegation(sdk.Coins{sdk.NewInt64Coin(stakeDenom, 50)})
	require.Nil(t, vva.DelegatedFree)
	require.Equal(t, sdk.Coins{sdk.NewInt64Coin(stakeDenom, 25)}, vva.DelegatedVesting)
	require.Equal(t, sdk.Coins{sdk.NewInt64Coin(feeDenom, 1000), sdk.NewInt64Coin(stakeDenom, 75)}, vva.GetCoins())
}

func TestGenesisAccountValidate(t *testing.T) {
	now := tmtime.Now()
	periods := vesting.VestingPeriods{
		vesting.VestingPeriod{PeriodLength: int64(12 * 60 * 60), VestingAmount: sdk.Coins{sdk.NewInt64Coin(feeDenom, 500), sdk.NewInt64Coin(stakeDenom, 50)}},
		vesting.VestingPeriod{PeriodLength: int64(6 * 60 * 60), VestingAmount: sdk.Coins{sdk.NewInt64Coin(feeDenom, 250), sdk.NewInt64Coin(stakeDenom, 25)}},
		vesting.VestingPeriod{PeriodLength: int64(6 * 60 * 60), VestingAmount: sdk.Coins{sdk.NewInt64Coin(feeDenom, 250), sdk.NewInt64Coin(stakeDenom, 25)}},
	}

	testAddr := CreateTestAddrs(1)[0]
	testPk := CreateTestPubKeys(1)[0]
	testConsAddr := sdk.ConsAddress(testPk.Address())
	origCoins := sdk.Coins{sdk.NewInt64Coin(feeDenom, 1000), sdk.NewInt64Coin(stakeDenom, 100)}
	bacc := auth.NewBaseAccountWithAddress(testAddr)
	bacc.SetCoins(origCoins)
	tests := []struct {
		name   string
		acc    authexported.GenesisAccount
		expErr error
	}{
		{
			"valid validator vesting account",
			NewValidatorVestingAccount(&bacc, now.Unix(), periods, testConsAddr, nil, 100),
			nil,
		},
		{
			"invalid signing threshold",
			NewValidatorVestingAccount(&bacc, now.Unix(), periods, testConsAddr, nil, -1),
			errors.New("signing threshold must be between 0 and 100"),
		},
		{
			"invalid signing threshold",
			NewValidatorVestingAccount(&bacc, now.Unix(), periods, testConsAddr, nil, 120),
			errors.New("signing threshold must be between 0 and 100"),
		},
		{
			"invalid return address",
			NewValidatorVestingAccount(&bacc, now.Unix(), periods, testConsAddr, testAddr, 90),
			errors.New("return address cannot be the same as the account address"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.acc.Validate()
			require.Equal(t, tt.expErr, err)
		})
	}
}
