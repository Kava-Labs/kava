package v0_11

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	v39_1auth "github.com/cosmos/cosmos-sdk/x/auth"
	v39_1authexported "github.com/cosmos/cosmos-sdk/x/auth/exported"
	v39_1vesting "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	v39_1gov "github.com/cosmos/cosmos-sdk/x/gov"
	v39_1supply "github.com/cosmos/cosmos-sdk/x/supply"

	v38_5auth "github.com/kava-labs/kava/migrate/v0_11/legacy/cosmos-sdk/v0.38.5/auth"
	v38_5supply "github.com/kava-labs/kava/migrate/v0_11/legacy/cosmos-sdk/v0.38.5/supply"
	v0_11bep3 "github.com/kava-labs/kava/x/bep3/legacy/v0_11"
	v0_9bep3 "github.com/kava-labs/kava/x/bep3/legacy/v0_9"
	v0_11harvest "github.com/kava-labs/kava/x/harvest"
	v0_11incentive "github.com/kava-labs/kava/x/incentive"
	v0_9incentive "github.com/kava-labs/kava/x/incentive/legacy/v0_9"
	v0_11pricefeed "github.com/kava-labs/kava/x/pricefeed"
	v0_9pricefeed "github.com/kava-labs/kava/x/pricefeed/legacy/v0_9"
	v0_11validator_vesting "github.com/kava-labs/kava/x/validator-vesting"
	v0_9validator_vesting "github.com/kava-labs/kava/x/validator-vesting/legacy/v0_9"
)

var deputyBnbBalance sdk.Coin
var hardBalance sdk.Coin

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
	deputyAddr, err := sdk.AccAddressFromBech32("kava1r4v2zdhdalfj2ydazallqvrus9fkphmglhn6u6")
	if err != nil {
		panic(err)
	}
	for _, account := range oldGenState.Accounts {
		switch acc := account.(type) {
		case *v38_5auth.BaseAccount:
			a := v39_1auth.BaseAccount(*acc)
			// Remove deputy bnb
			if a.GetAddress().Equals(deputyAddr) {
				deputyBnbBalance = sdk.NewCoin("bnb", a.GetCoins().AmountOf("bnb"))
				err := a.SetCoins(a.GetCoins().Sub(sdk.NewCoins(sdk.NewCoin("bnb", a.GetCoins().AmountOf("bnb")))))
				if err != nil {
					panic(err)
				}
			}
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
	err = lpMacc.SetCoins(sdk.NewCoins(sdk.NewCoin("hard", sdk.NewInt(80000000000000))))
	if err != nil {
		panic(err)
	}
	newAccounts = append(newAccounts, v39_1authexported.GenesisAccount(lpMacc))
	delegatorMacc := v39_1supply.NewEmptyModuleAccount(v0_11harvest.DelegatorAccount, v39_1supply.Minter, v39_1supply.Burner)
	err = delegatorMacc.SetCoins(sdk.NewCoins(sdk.NewCoin("hard", sdk.NewInt(40000000000000))))
	if err != nil {
		panic(err)
	}
	newAccounts = append(newAccounts, v39_1authexported.GenesisAccount(delegatorMacc))
	hardBalance = sdk.NewCoin("hard", sdk.NewInt(120000000000000))

	return v39_1auth.NewGenesisState(v39_1auth.Params(oldGenState.Params), newAccounts)
}

// MigrateSupply reconciles supply from kava-3 to kava-4
// deputy balance of bnb coins is removed (deputy now mints coins)
// hard token supply is added
func MigrateSupply(oldGenState v39_1supply.GenesisState, deputyBalance sdk.Coin, hardBalance sdk.Coin) v39_1supply.GenesisState {
	oldGenState.Supply = oldGenState.Supply.Sub(sdk.Coins{deputyBalance}).Add(hardBalance)
	return oldGenState
}

// MigrateGov migrates gov genesis state
func MigrateGov(oldGenState v39_1gov.GenesisState) v39_1gov.GenesisState {
	oldGenState.VotingParams.VotingPeriod = time.Hour * 24 * 7
	return oldGenState
}

// MigrateIncentive migrates from a v0.9 (or v0.10) incentive genesis state to a v0.11 incentive genesis state
func MigrateIncentive(oldGenState v0_9incentive.GenesisState) v0_11incentive.GenesisState {
	var newRewards v0_11incentive.Rewards
	var newRewardPeriods v0_11incentive.RewardPeriods
	var newClaimPeriods v0_11incentive.ClaimPeriods
	var newClaims v0_11incentive.Claims
	var newClaimPeriodIds v0_11incentive.GenesisClaimPeriodIDs

	newMultiplier := v0_11incentive.NewMultiplier(v0_11incentive.Large, 12, sdk.OneDec())
	smallMultiplier := v0_11incentive.NewMultiplier(v0_11incentive.Small, 1, sdk.MustNewDecFromStr("0.25"))

	for _, oldReward := range oldGenState.Params.Rewards {
		newReward := v0_11incentive.NewReward(oldReward.Active, oldReward.Denom+"-a", oldReward.AvailableRewards, oldReward.Duration, v0_11incentive.Multipliers{smallMultiplier, newMultiplier}, oldReward.ClaimDuration)
		newRewards = append(newRewards, newReward)
	}
	newParams := v0_11incentive.NewParams(true, newRewards)

	for _, oldRewardPeriod := range oldGenState.RewardPeriods {

		newRewardPeriod := v0_11incentive.NewRewardPeriod(oldRewardPeriod.Denom+"-a", oldRewardPeriod.Start, oldRewardPeriod.End, oldRewardPeriod.Reward, oldRewardPeriod.ClaimEnd, v0_11incentive.Multipliers{smallMultiplier, newMultiplier})
		newRewardPeriods = append(newRewardPeriods, newRewardPeriod)
	}

	for _, oldClaimPeriod := range oldGenState.ClaimPeriods {
		newClaimPeriod := v0_11incentive.NewClaimPeriod(oldClaimPeriod.Denom+"-a", oldClaimPeriod.ID, oldClaimPeriod.End, v0_11incentive.Multipliers{smallMultiplier, newMultiplier})
		newClaimPeriods = append(newClaimPeriods, newClaimPeriod)
	}

	for _, oldClaim := range oldGenState.Claims {
		newClaim := v0_11incentive.NewClaim(oldClaim.Owner, oldClaim.Reward, oldClaim.Denom+"-a", oldClaim.ClaimPeriodID)
		newClaims = append(newClaims, newClaim)
	}

	for _, oldClaimPeriodID := range oldGenState.NextClaimPeriodIDs {
		newClaimPeriodID := v0_11incentive.GenesisClaimPeriodID{
			CollateralType: oldClaimPeriodID.Denom + "-a",
			ID:             oldClaimPeriodID.ID,
		}
		newClaimPeriodIds = append(newClaimPeriodIds, newClaimPeriodID)
	}

	return v0_11incentive.NewGenesisState(newParams, oldGenState.PreviousBlockTime, newRewardPeriods, newClaimPeriods, newClaims, newClaimPeriodIds)
}

// MigratePricefeed migrates from a v0.9 (or v0.10) pricefeed genesis state to a v0.11 pricefeed genesis state
func MigratePricefeed(oldGenState v0_9pricefeed.GenesisState) v0_11pricefeed.GenesisState {
	var newMarkets v0_11pricefeed.Markets
	var newPostedPrices v0_11pricefeed.PostedPrices
	var oracles []sdk.AccAddress

	for _, market := range oldGenState.Params.Markets {
		newMarket := v0_11pricefeed.NewMarket(market.MarketID, market.BaseAsset, market.QuoteAsset, market.Oracles, market.Active)
		newMarkets = append(newMarkets, newMarket)
		oracles = market.Oracles
	}
	// ------- add btc, xrp, busd markets --------
	btcSpotMarket := v0_11pricefeed.NewMarket("btc:usd", "btc", "usd", oracles, true)
	btcLiquidationMarket := v0_11pricefeed.NewMarket("btc:usd:30", "btc", "usd", oracles, true)
	xrpSpotMarket := v0_11pricefeed.NewMarket("xrp:usd", "xrp", "usd", oracles, true)
	xrpLiquidationMarket := v0_11pricefeed.NewMarket("xrp:usd:30", "xrp", "usd", oracles, true)
	busdSpotMarket := v0_11pricefeed.NewMarket("busd:usd", "busd", "usd", oracles, true)
	busdLiquidationMarket := v0_11pricefeed.NewMarket("busd:usd:30", "busd", "usd", oracles, true)
	newMarkets = append(newMarkets, btcSpotMarket, btcLiquidationMarket, xrpSpotMarket, xrpLiquidationMarket, busdSpotMarket, busdLiquidationMarket)

	for _, price := range oldGenState.PostedPrices {
		newPrice := v0_11pricefeed.NewPostedPrice(price.MarketID, price.OracleAddress, price.Price, price.Expiry)
		newPostedPrices = append(newPostedPrices, newPrice)
	}
	newParams := v0_11pricefeed.NewParams(newMarkets)

	return v0_11pricefeed.NewGenesisState(newParams, newPostedPrices)
}
