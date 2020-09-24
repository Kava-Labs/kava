package v0_11

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	v39_1auth "github.com/cosmos/cosmos-sdk/x/auth"
	v39_1authexported "github.com/cosmos/cosmos-sdk/x/auth/exported"
	v39_1vesting "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	v39_1supply "github.com/cosmos/cosmos-sdk/x/supply"

	v38_5auth "github.com/kava-labs/kava/migrate/v0_11/legacy/cosmos-sdk/v0.38.5/auth"
	v38_5supply "github.com/kava-labs/kava/migrate/v0_11/legacy/cosmos-sdk/v0.38.5/supply"
	v0_11bep3 "github.com/kava-labs/kava/x/bep3/legacy/v0_11"
	v0_9bep3 "github.com/kava-labs/kava/x/bep3/legacy/v0_9"
	v0_11harvest "github.com/kava-labs/kava/x/hvt"
	v0_11validator_vesting "github.com/kava-labs/kava/x/validator-vesting"
	v0_9validator_vesting "github.com/kava-labs/kava/x/validator-vesting/legacy/v0_9"
)

// MigrateBep3 migrates from a v0.9 (or v0.10) bep3 genesis state to a v0.11 bep3 genesis state
func MigrateBep3(oldGenState v0_9bep3.GenesisState) v0_11bep3.GenesisState {
	var assetParams v0_11bep3.AssetParams
	var assetSupplies v0_11bep3.AssetSupplies
	v0_9Params := oldGenState.Params

	for _, asset := range v0_9Params.SupportedAssets {
		v10AssetParam := v0_11bep3.AssetParam{
			Active:        asset.Active,
			Denom:         asset.Denom,
			CoinID:        asset.CoinID,
			DeputyAddress: v0_9Params.BnbDeputyAddress,
			FixedFee:      v0_9Params.BnbDeputyFixedFee,
			MinSwapAmount: sdk.OneInt(), // set min swap to one - prevents accounts that hold zero bnb from creating spam txs
			MaxSwapAmount: v0_9Params.MaxAmount,
			MinBlockLock:  v0_9Params.MinBlockLock,
			MaxBlockLock:  v0_9Params.MaxBlockLock,
			SupplyLimit: v0_11bep3.SupplyLimit{
				Limit:          asset.Limit,
				TimeLimited:    false,
				TimePeriod:     time.Duration(0),
				TimeBasedLimit: sdk.ZeroInt(),
			},
		}
		assetParams = append(assetParams, v10AssetParam)
	}
	for _, supply := range oldGenState.AssetSupplies {
		newSupply := v0_11bep3.NewAssetSupply(supply.IncomingSupply, supply.OutgoingSupply, supply.CurrentSupply, sdk.NewCoin(supply.CurrentSupply.Denom, sdk.ZeroInt()), time.Duration(0))
		assetSupplies = append(assetSupplies, newSupply)
	}
	var swaps v0_11bep3.AtomicSwaps
	for _, oldSwap := range oldGenState.AtomicSwaps {
		newSwap := v0_11bep3.AtomicSwap{
			Amount:              oldSwap.Amount,
			RandomNumberHash:    oldSwap.RandomNumberHash,
			ExpireHeight:        oldSwap.ExpireHeight,
			Timestamp:           oldSwap.Timestamp,
			Sender:              oldSwap.Sender,
			Recipient:           oldSwap.Recipient,
			SenderOtherChain:    oldSwap.SenderOtherChain,
			RecipientOtherChain: oldSwap.RecipientOtherChain,
			ClosedBlock:         oldSwap.ClosedBlock,
			Status:              v0_11bep3.SwapStatus(oldSwap.Status),
			CrossChain:          oldSwap.CrossChain,
			Direction:           v0_11bep3.SwapDirection(oldSwap.Direction),
		}
		swaps = append(swaps, newSwap)
	}
	return v0_11bep3.GenesisState{
		Params: v0_11bep3.Params{
			AssetParams: assetParams},
		AtomicSwaps:       swaps,
		Supplies:          assetSupplies,
		PreviousBlockTime: v0_11bep3.DefaultPreviousBlockTime,
	}
}

// MigrateAuth migrates from a v0.38.5 auth genesis state to a v0.39.1 auth genesis state
func MigrateAuth(oldGenState v38_5auth.GenesisState) v39_1auth.GenesisState {
	var newAccounts v39_1authexported.GenesisAccounts

	for _, account := range oldGenState.Accounts {
		switch acc := account.(type) {
		case *v38_5auth.BaseAccount:
			a := v39_1auth.BaseAccount(*acc)
			newAccounts = append(newAccounts, v39_1authexported.GenesisAccount(&a))

		case *v38_5auth.BaseVestingAccount:
			ba := v39_1auth.BaseAccount(*(acc.BaseAccount))
			bva := v39_1vesting.BaseVestingAccount{
				BaseAccount:      &ba,
				OriginalVesting:  acc.OriginalVesting,
				DelegatedFree:    acc.DelegatedFree,
				DelegatedVesting: acc.DelegatedVesting,
				EndTime:          acc.EndTime,
			}
			newAccounts = append(newAccounts, v39_1authexported.GenesisAccount(&bva))

		case *v38_5auth.ContinuousVestingAccount:
			ba := v39_1auth.BaseAccount(*(acc.BaseVestingAccount.BaseAccount))
			bva := v39_1vesting.BaseVestingAccount{
				BaseAccount:      &ba,
				OriginalVesting:  acc.BaseVestingAccount.OriginalVesting,
				DelegatedFree:    acc.BaseVestingAccount.DelegatedFree,
				DelegatedVesting: acc.BaseVestingAccount.DelegatedVesting,
				EndTime:          acc.BaseVestingAccount.EndTime,
			}
			cva := v39_1vesting.ContinuousVestingAccount{
				BaseVestingAccount: &bva,
				StartTime:          acc.StartTime,
			}
			newAccounts = append(newAccounts, v39_1authexported.GenesisAccount(&cva))

		case *v38_5auth.DelayedVestingAccount:
			ba := v39_1auth.BaseAccount(*(acc.BaseVestingAccount.BaseAccount))
			bva := v39_1vesting.BaseVestingAccount{
				BaseAccount:      &ba,
				OriginalVesting:  acc.BaseVestingAccount.OriginalVesting,
				DelegatedFree:    acc.BaseVestingAccount.DelegatedFree,
				DelegatedVesting: acc.BaseVestingAccount.DelegatedVesting,
				EndTime:          acc.BaseVestingAccount.EndTime,
			}
			dva := v39_1vesting.DelayedVestingAccount{
				BaseVestingAccount: &bva,
			}
			newAccounts = append(newAccounts, v39_1authexported.GenesisAccount(&dva))

		case *v38_5auth.PeriodicVestingAccount:
			ba := v39_1auth.BaseAccount(*(acc.BaseVestingAccount.BaseAccount))
			bva := v39_1vesting.BaseVestingAccount{
				BaseAccount:      &ba,
				OriginalVesting:  acc.BaseVestingAccount.OriginalVesting,
				DelegatedFree:    acc.BaseVestingAccount.DelegatedFree,
				DelegatedVesting: acc.BaseVestingAccount.DelegatedVesting,
				EndTime:          acc.BaseVestingAccount.EndTime,
			}
			var newPeriods v39_1vesting.Periods
			for _, p := range acc.VestingPeriods {
				newPeriods = append(newPeriods, v39_1vesting.Period(p))
			}
			pva := v39_1vesting.PeriodicVestingAccount{
				BaseVestingAccount: &bva,
				StartTime:          acc.StartTime,
				VestingPeriods:     newPeriods,
			}
			newAccounts = append(newAccounts, v39_1authexported.GenesisAccount(&pva))

		case *v38_5supply.ModuleAccount:
			ba := v39_1auth.BaseAccount(*(acc.BaseAccount))
			ma := v39_1supply.ModuleAccount{
				BaseAccount: &ba,
				Name:        acc.Name,
				Permissions: acc.Permissions,
			}
			newAccounts = append(newAccounts, v39_1authexported.GenesisAccount(&ma))

		case *v0_9validator_vesting.ValidatorVestingAccount:
			ba := v39_1auth.BaseAccount(*(acc.BaseAccount))
			bva := v39_1vesting.BaseVestingAccount{
				BaseAccount:      &ba,
				OriginalVesting:  acc.BaseVestingAccount.OriginalVesting,
				DelegatedFree:    acc.BaseVestingAccount.DelegatedFree,
				DelegatedVesting: acc.BaseVestingAccount.DelegatedVesting,
				EndTime:          acc.BaseVestingAccount.EndTime,
			}
			var newPeriods v39_1vesting.Periods
			for _, p := range acc.VestingPeriods {
				newPeriods = append(newPeriods, v39_1vesting.Period(p))
			}
			pva := v39_1vesting.PeriodicVestingAccount{
				BaseVestingAccount: &bva,
				StartTime:          acc.StartTime,
				VestingPeriods:     newPeriods,
			}
			var newVestingProgress []v0_11validator_vesting.VestingProgress
			for _, p := range acc.VestingPeriodProgress {
				newVestingProgress = append(newVestingProgress, v0_11validator_vesting.VestingProgress(p))
			}
			vva := v0_11validator_vesting.ValidatorVestingAccount{
				PeriodicVestingAccount: &pva,
				ValidatorAddress:       acc.ValidatorAddress,
				ReturnAddress:          acc.ReturnAddress,
				SigningThreshold:       acc.SigningThreshold,
				CurrentPeriodProgress:  v0_11validator_vesting.CurrentPeriodProgress(acc.CurrentPeriodProgress),
				VestingPeriodProgress:  newVestingProgress,
				DebtAfterFailedVesting: acc.DebtAfterFailedVesting,
			}
			newAccounts = append(newAccounts, v39_1authexported.GenesisAccount(&vva))

		default:
			panic(fmt.Sprintf("unrecognized account type: %T", acc))
		}
	}

	// ---- add harvest module accounts -------
	lpMacc := v39_1supply.NewEmptyModuleAccount(v0_11harvest.LPAccount, v39_1supply.Minter, v39_1supply.Burner)
	lpMacc.SetCoins(sdk.NewCoins(sdk.NewCoin("hard", sdk.NewInt(80000000000000))))
	newAccounts = append(newAccounts, v39_1authexported.GenesisAccount(lpMacc))
	delegatorMacc := v39_1supply.NewEmptyModuleAccount(v0_11harvest.DelegatorAccount, v39_1supply.Minter, v39_1supply.Burner)
	delegatorMacc.SetCoins(sdk.NewCoins(sdk.NewCoin("hard", sdk.NewInt(40000000000000))))
	newAccounts = append(newAccounts, v39_1authexported.GenesisAccount(delegatorMacc))

	return v39_1auth.NewGenesisState(v39_1auth.Params(oldGenState.Params), newAccounts)
}
