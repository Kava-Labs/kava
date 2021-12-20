/**
 * The v0_16 x/auth migration logic is adapted from
 * https://github.com/cosmos/cosmos-sdk/blob/b75c29fc15d3320ec0c7596dbd7c787c48dccad8/x/auth/legacy/v040/migrate.go
 *
 * The original migration code is changed here to support the following custom Account from the kava modules.
 * - `x/validator-vesting/ValidatorVestingAccount`
 */
package v0_16

import (
	"time"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	v039auth "github.com/cosmos/cosmos-sdk/x/auth/legacy/v039"
	v040auth "github.com/cosmos/cosmos-sdk/x/auth/types"
	v040vesting "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"

	v015validatorvesting "github.com/kava-labs/kava/x/validator-vesting/legacy/v0_15"
)

// convertBaseAccount converts a 0.39 BaseAccount to a 0.40 BaseAccount.
func convertBaseAccount(old *v039auth.BaseAccount) *v040auth.BaseAccount {
	var any *codectypes.Any

	if old.PubKey != nil {
		var err error
		any, err = codectypes.NewAnyWithValue(old.PubKey)
		if err != nil {
			panic(err)
		}
	}

	return &v040auth.BaseAccount{
		Address:       old.Address.String(),
		PubKey:        any,
		AccountNumber: old.AccountNumber,
		Sequence:      old.Sequence,
	}
}

// convertBaseVestingAccount converts a 0.39 BaseVestingAccount to a 0.40 BaseVestingAccount.
func convertBaseVestingAccount(old *v039auth.BaseVestingAccount) *v040vesting.BaseVestingAccount {
	baseAccount := convertBaseAccount(old.BaseAccount)

	return &v040vesting.BaseVestingAccount{
		BaseAccount:      baseAccount,
		OriginalVesting:  old.OriginalVesting,
		DelegatedFree:    old.DelegatedFree,
		DelegatedVesting: old.DelegatedVesting,
		EndTime:          old.EndTime,
	}
}

// ResetPeriodicVestingAccount resets a periodic vesting account to a new start
// time.  The account is modified in place, and vesting periods before the new
// start time are removed from the account.
func ResetPeriodicVestingAccount(vacc *v040vesting.PeriodicVestingAccount, startTime time.Time) {
	currentPeriod := vacc.StartTime

	newOriginalVesting := sdk.Coins{}
	newStartTime := startTime.Unix()
	newPeriods := v040vesting.Periods{}

	for _, period := range vacc.VestingPeriods {
		currentPeriod = currentPeriod + period.Length

		// Periods less than the newStartTime are still vesting,
		// so adjust their length and add them to the newPeriods
		if newStartTime < currentPeriod {
			// adjust the length of the first vesting period
			// to be relative to the new start time
			if len(newPeriods) == 0 {
				period.Length = currentPeriod - newStartTime
			}

			newOriginalVesting = newOriginalVesting.Add(period.Amount...)
			newPeriods = append(newPeriods, period)
		}
	}

	// If the new original vesting amount is less than the delegated vesting amount, set delegated vesting
	// to the new original vesting amount, and add the difference to the delegated free amount
	for _, delegatedVestingCoin := range vacc.DelegatedVesting {
		newDelegatedVestingCoin := sdk.NewCoin(delegatedVestingCoin.Denom, sdk.MinInt(delegatedVestingCoin.Amount, newOriginalVesting.AmountOf(delegatedVestingCoin.Denom)))
		delegationAdjustment := delegatedVestingCoin.Sub(newDelegatedVestingCoin)

		if !delegationAdjustment.IsZero() {
			vacc.DelegatedVesting = vacc.DelegatedVesting.Sub(sdk.NewCoins(delegationAdjustment))
			vacc.DelegatedFree = vacc.DelegatedFree.Add(delegationAdjustment)
		}
	}

	// update vesting account
	vacc.StartTime = newStartTime
	vacc.OriginalVesting = newOriginalVesting
	vacc.VestingPeriods = newPeriods

	// ensure end time is >= start time
	if vacc.StartTime >= vacc.EndTime {
		vacc.EndTime = vacc.StartTime
	}
}

// Migrate accepts exported x/auth genesis state from v0.38/v0.39 and migrates
// it to v0.40 x/auth genesis state. The migration includes:
//
// - Removing coins from account encoding.
// - Re-encode in v0.40 GenesisState.
func MigrateAuthV040(authGenState v039auth.GenesisState, genesisTime time.Time) *v040auth.GenesisState {
	// Convert v0.39 accounts to v0.40 ones.
	var v040Accounts = make([]v040auth.GenesisAccount, len(authGenState.Accounts))
	for i, v039Account := range authGenState.Accounts {
		switch v039Account := v039Account.(type) {
		case *v039auth.BaseAccount:
			{
				v040Accounts[i] = convertBaseAccount(v039Account)
			}
		case *v039auth.ModuleAccount:
			{
				v040Accounts[i] = &v040auth.ModuleAccount{
					BaseAccount: convertBaseAccount(v039Account.BaseAccount),
					Name:        v039Account.Name,
					Permissions: v039Account.Permissions,
				}
			}
		case *v039auth.BaseVestingAccount:
			{
				v040Accounts[i] = convertBaseVestingAccount(v039Account)
			}
		case *v039auth.ContinuousVestingAccount:
			{
				v040Accounts[i] = &v040vesting.ContinuousVestingAccount{
					BaseVestingAccount: convertBaseVestingAccount(v039Account.BaseVestingAccount),
					StartTime:          v039Account.StartTime,
				}
			}
		case *v039auth.DelayedVestingAccount:
			{
				v040Accounts[i] = &v040vesting.DelayedVestingAccount{
					BaseVestingAccount: convertBaseVestingAccount(v039Account.BaseVestingAccount),
				}
			}
		case *v039auth.PeriodicVestingAccount:
			{
				vestingPeriods := make([]v040vesting.Period, len(v039Account.VestingPeriods))
				for j, period := range v039Account.VestingPeriods {
					vestingPeriods[j] = v040vesting.Period{
						Length: period.Length,
						Amount: period.Amount,
					}
				}
				vacc := v040vesting.PeriodicVestingAccount{
					BaseVestingAccount: convertBaseVestingAccount(v039Account.BaseVestingAccount),
					StartTime:          v039Account.StartTime,
					VestingPeriods:     vestingPeriods,
				}

				ResetPeriodicVestingAccount(&vacc, genesisTime)

				v040Accounts[i] = &vacc
			}
		case *v015validatorvesting.ValidatorVestingAccount:
			{
				// Convert validator vesting accounts to base accounts since no more vesting is needed
				v040Accounts[i] = convertBaseAccount(v039Account.BaseAccount)
			}
		default:
			panic(sdkerrors.Wrapf(sdkerrors.ErrInvalidType, "got invalid type %T", v039Account))
		}

	}

	// Convert v0.40 accounts into Anys.
	anys := make([]*codectypes.Any, len(v040Accounts))
	for i, v040Account := range v040Accounts {
		any, err := codectypes.NewAnyWithValue(v040Account)
		if err != nil {
			panic(err)
		}

		anys[i] = any
	}

	return &v040auth.GenesisState{
		Params: v040auth.Params{
			MaxMemoCharacters:      authGenState.Params.MaxMemoCharacters,
			TxSigLimit:             authGenState.Params.TxSigLimit,
			TxSizeCostPerByte:      authGenState.Params.TxSizeCostPerByte,
			SigVerifyCostED25519:   authGenState.Params.SigVerifyCostED25519,
			SigVerifyCostSecp256k1: authGenState.Params.SigVerifyCostSecp256k1,
		},
		Accounts: anys,
	}
}
