package v0_13

import (
	"fmt"
	"sort"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authexported "github.com/cosmos/cosmos-sdk/x/auth/exported"
	"github.com/cosmos/cosmos-sdk/x/supply"

	v0_13cdp "github.com/kava-labs/kava/x/cdp"
	v0_11cdp "github.com/kava-labs/kava/x/cdp/legacy/v0_11"
	v0_13hard "github.com/kava-labs/kava/x/hard"
	v0_11hard "github.com/kava-labs/kava/x/hard/legacy/v0_11"
	v0_13incentive "github.com/kava-labs/kava/x/incentive"
	v0_11incentive "github.com/kava-labs/kava/x/incentive/legacy/v0_11"
)

var (
	GenesisTime   = time.Date(2021, 3, 4, 14, 0, 0, 0, time.UTC)
	RewardEndTime = time.Date(2022, 2, 25, 14, 0, 0, 0, time.UTC)
	ClaimEndTime  = time.Date(2026, 2, 25, 14, 0, 0, 0, time.UTC)
)

// MigrateCDP migrates from a v0.11 cdp genesis state to a v0.13 cdp genesis state
func MigrateCDP(oldGenState v0_11cdp.GenesisState) v0_13cdp.GenesisState {
	var newCDPs v0_13cdp.CDPs
	var newDeposits v0_13cdp.Deposits
	var newCollateralParams v0_13cdp.CollateralParams
	var newGenesisAccumulationTimes v0_13cdp.GenesisAccumulationTimes
	var previousAccumulationTime time.Time
	var totalPrincipals v0_13cdp.GenesisTotalPrincipals
	newStartingID := oldGenState.StartingCdpID

	totalPrincipalMap := make(map[string]sdk.Int)

	for _, cdp := range oldGenState.CDPs {
		newCDP := v0_13cdp.NewCDPWithFees(cdp.ID, cdp.Owner, cdp.Collateral, cdp.Type, cdp.Principal, cdp.AccumulatedFees, cdp.FeesUpdated, sdk.OneDec())
		if previousAccumulationTime.Before(cdp.FeesUpdated) {
			previousAccumulationTime = cdp.FeesUpdated
		}
		_, found := totalPrincipalMap[cdp.Type]
		if !found {
			totalPrincipalMap[cdp.Type] = sdk.ZeroInt()
		}
		totalPrincipalMap[cdp.Type] = totalPrincipalMap[cdp.Type].Add(newCDP.GetTotalPrincipal().Amount)

		newCDPs = append(newCDPs, newCDP)
	}

	for _, cp := range oldGenState.Params.CollateralParams {
		newCollateralParam := v0_13cdp.NewCollateralParam(cp.Denom, cp.Type, cp.LiquidationRatio, cp.DebtLimit, cp.StabilityFee, cp.AuctionSize, cp.LiquidationPenalty, cp.Prefix, cp.SpotMarketID, cp.LiquidationMarketID, sdk.MustNewDecFromStr("0.01"), sdk.NewInt(10), cp.ConversionFactor)
		newCollateralParams = append(newCollateralParams, newCollateralParam)
		newGenesisAccumulationTime := v0_13cdp.NewGenesisAccumulationTime(cp.Type, previousAccumulationTime, sdk.OneDec())
		newGenesisAccumulationTimes = append(newGenesisAccumulationTimes, newGenesisAccumulationTime)
	}

	for _, dep := range oldGenState.Deposits {
		newDep := v0_13cdp.NewDeposit(dep.CdpID, dep.Depositor, dep.Amount)
		newDeposits = append(newDeposits, newDep)
	}

	for ctype, tp := range totalPrincipalMap {
		totalPrincipal := v0_13cdp.NewGenesisTotalPrincipal(ctype, tp)
		totalPrincipals = append(totalPrincipals, totalPrincipal)
	}

	sort.Slice(totalPrincipals, func(i, j int) bool { return totalPrincipals[i].CollateralType < totalPrincipals[j].CollateralType })

	oldDebtParam := oldGenState.Params.DebtParam

	newDebtParam := v0_13cdp.NewDebtParam(oldDebtParam.Denom, oldDebtParam.ReferenceAsset, oldDebtParam.ConversionFactor, oldDebtParam.DebtFloor)

	newGlobalDebtLimit := oldGenState.Params.GlobalDebtLimit

	newParams := v0_13cdp.NewParams(newGlobalDebtLimit, newCollateralParams, newDebtParam, oldGenState.Params.SurplusAuctionThreshold, oldGenState.Params.SurplusAuctionLot, oldGenState.Params.DebtAuctionThreshold, oldGenState.Params.DebtAuctionLot, false)

	return v0_13cdp.NewGenesisState(
		newParams,
		newCDPs,
		newDeposits,
		newStartingID,
		oldGenState.DebtDenom,
		oldGenState.GovDenom,
		newGenesisAccumulationTimes,
		totalPrincipals,
	)
}

// MigrateHard migrates from a v0.11 hard (harvest) genesis state to a v0.13 hard genesis state
func MigrateHard(genesisState v0_11hard.GenesisState) v0_13hard.GenesisState {
	v13Deposits := v0_13hard.Deposits{}
	v13DepositorMap := make(map[string]v0_13hard.Deposit)
	v13GenesisAccumulationTimes := v0_13hard.GenesisAccumulationTimes{}
	v13TotalSupplied := sdk.NewCoins()

	for _, dep := range genesisState.Deposits {
		v13Deposit, ok := v13DepositorMap[dep.Depositor.String()]
		if !ok {
			v13Deposit := v0_13hard.NewDeposit(dep.Depositor, sdk.NewCoins(dep.Amount), v0_13hard.SupplyInterestFactors{v0_13hard.NewSupplyInterestFactor(dep.Amount.Denom, sdk.OneDec())})
			v13DepositorMap[dep.Depositor.String()] = v13Deposit
		} else {
			v13Deposit.Amount = v13Deposit.Amount.Add(dep.Amount)
			v13Deposit.Index = append(v13Deposit.Index, v0_13hard.NewSupplyInterestFactor(dep.Amount.Denom, sdk.OneDec()))
			v13DepositorMap[dep.Depositor.String()] = v13Deposit
		}
	}

	newParams := v0_13hard.NewParams(
		v0_13hard.MoneyMarkets{
			// btcb money market - TODO:
			// 1. is auction size actually used/enforced?
			// 2. What even is this interest rate model
			v0_13hard.NewMoneyMarket("btcb", v0_13hard.NewBorrowLimit(true, sdk.ZeroDec(), sdk.MustNewDecFromStr("0.5")), "btc:usd", sdk.NewInt(8), sdk.NewInt(100000000),
				v0_13hard.NewInterestRateModel(sdk.ZeroDec(), sdk.MustNewDecFromStr("0.002"), sdk.MustNewDecFromStr("0.8"), sdk.MustNewDecFromStr("0.01")),
				sdk.MustNewDecFromStr("0.025"), sdk.MustNewDecFromStr("0.01"),
			),
			// xrpb
			v0_13hard.NewMoneyMarket("xrpb", v0_13hard.NewBorrowLimit(true, sdk.ZeroDec(), sdk.MustNewDecFromStr("0.5")), "xrp:usd", sdk.NewInt(8), sdk.NewInt(10000000000000),
				v0_13hard.NewInterestRateModel(sdk.ZeroDec(), sdk.MustNewDecFromStr("0.002"), sdk.MustNewDecFromStr("0.8"), sdk.MustNewDecFromStr("0.01")),
				sdk.MustNewDecFromStr("0.025"), sdk.MustNewDecFromStr("0.01"),
			),
			// busd
			v0_13hard.NewMoneyMarket("busd", v0_13hard.NewBorrowLimit(true, sdk.MustNewDecFromStr("100000000000000"), sdk.MustNewDecFromStr("0.5")), "busd:usd", sdk.NewInt(8), sdk.NewInt(5000000000000),
				v0_13hard.NewInterestRateModel(sdk.ZeroDec(), sdk.MustNewDecFromStr("0.002"), sdk.MustNewDecFromStr("0.8"), sdk.MustNewDecFromStr("0.01")),
				sdk.MustNewDecFromStr("0.025"), sdk.MustNewDecFromStr("0.01"),
			),
			// usdx
			v0_13hard.NewMoneyMarket("usdx", v0_13hard.NewBorrowLimit(true, sdk.ZeroDec(), sdk.ZeroDec()), "usdx:usd", sdk.NewInt(6), sdk.NewInt(50000000000),
				v0_13hard.NewInterestRateModel(sdk.ZeroDec(), sdk.MustNewDecFromStr("0.002"), sdk.MustNewDecFromStr("0.8"), sdk.MustNewDecFromStr("0.01")),
				sdk.MustNewDecFromStr("0.025"), sdk.MustNewDecFromStr("0.01"),
			),
			// ukava
			v0_13hard.NewMoneyMarket("ukava", v0_13hard.NewBorrowLimit(true, sdk.ZeroDec(), sdk.MustNewDecFromStr("0.5")), "kava:usd", sdk.NewInt(6), sdk.NewInt(10000000000),
				v0_13hard.NewInterestRateModel(sdk.ZeroDec(), sdk.MustNewDecFromStr("0.002"), sdk.MustNewDecFromStr("0.8"), sdk.MustNewDecFromStr("0.01")),
				sdk.MustNewDecFromStr("0.025"), sdk.MustNewDecFromStr("0.01"),
			),
			// hard
			v0_13hard.NewMoneyMarket("hard", v0_13hard.NewBorrowLimit(true, sdk.ZeroDec(), sdk.MustNewDecFromStr("0.5")), "hard:usd", sdk.NewInt(6), sdk.NewInt(25000000000),
				v0_13hard.NewInterestRateModel(sdk.ZeroDec(), sdk.MustNewDecFromStr("0.002"), sdk.MustNewDecFromStr("0.8"), sdk.MustNewDecFromStr("0.01")),
				sdk.MustNewDecFromStr("0.025"), sdk.MustNewDecFromStr("0.01"),
			),
		},
		10,
	)

	for _, newDep := range v13DepositorMap {
		v13Deposits = append(v13Deposits, newDep)
		v13TotalSupplied = v13TotalSupplied.Add(newDep.Amount...)
	}

	for _, mm := range newParams.MoneyMarkets {
		genAccumulationTime := v0_13hard.NewGenesisAccumulationTime(mm.Denom, GenesisTime, sdk.OneDec(), sdk.OneDec())
		v13GenesisAccumulationTimes = append(v13GenesisAccumulationTimes, genAccumulationTime)
	}

	return v0_13hard.NewGenesisState(newParams, v13GenesisAccumulationTimes, v13Deposits, v0_13hard.DefaultBorrows, v13TotalSupplied, v0_13hard.DefaultTotalBorrowed, v0_13hard.DefaultTotalReserves)
}

// MigrateIncentive migrates from a v0.11 incentive genesis state to a v0.13 incentive genesis state
func MigrateIncentive(hardGS v0_11hard.GenesisState, incentiveGS v0_11incentive.GenesisState) v0_13incentive.GenesisState {
	usdxMintingRewardPeriods := v0_13incentive.RewardPeriods{}
	usdxRewardsPerSecondMap := make(map[string]sdk.Coin)
	usdxRewardsPerSecondMap["bnb-a"] = sdk.NewCoin("ukava", sdk.NewInt(122354))
	usdxRewardsPerSecondMap["btcb-a"] = sdk.NewCoin("ukava", sdk.NewInt(158730))
	usdxRewardsPerSecondMap["busd-a"] = sdk.NewCoin("ukava", sdk.NewInt(30588))
	usdxRewardsPerSecondMap["hard-a"] = sdk.NewCoin("ukava", sdk.NewInt(23809))
	usdxRewardsPerSecondMap["ukava-a"] = sdk.NewCoin("ukava", sdk.NewInt(31746))
	usdxRewardsPerSecondMap["xrpb-a"] = sdk.NewCoin("ukava", sdk.NewInt(31746))

	for _, rp := range incentiveGS.RewardPeriods {
		rewardsPerSecond, ok := usdxRewardsPerSecondMap[rp.CollateralType]
		if !ok {
			panic(fmt.Sprintf("No rewards per second for collateral type: %s\n", rp.CollateralType))
		}
		newRP := v0_13incentive.NewRewardPeriod(true, rp.CollateralType, GenesisTime, RewardEndTime, rewardsPerSecond)
		usdxMintingRewardPeriods = append(usdxMintingRewardPeriods, newRP)
	}

	hardSupplyRewardPeriods := v0_13incentive.MultiRewardPeriods{}
	for _, rp := range hardGS.Params.LiquidityProviderSchedules {
		newRP := v0_13incentive.NewMultiRewardPeriod(true, rp.DepositDenom, rp.Start, rp.End, sdk.NewCoins(rp.RewardsPerSecond))
		hardSupplyRewardPeriods = append(hardSupplyRewardPeriods, newRP)
	}
	hardBorrowRewardPeriods := v0_13incentive.MultiRewardPeriods{}
	hardDelegatorRewardPeriods := v0_13incentive.RewardPeriods{}
	for _, rp := range hardGS.Params.DelegatorDistributionSchedules {
		newRP := v0_13incentive.NewRewardPeriod(rp.DistributionSchedule.Active, rp.DistributionSchedule.DepositDenom, rp.DistributionSchedule.Start, rp.DistributionSchedule.End, rp.DistributionSchedule.RewardsPerSecond)
		hardDelegatorRewardPeriods = append(hardDelegatorRewardPeriods, newRP)
	}
	params := v0_13incentive.NewParams(usdxMintingRewardPeriods, hardSupplyRewardPeriods, hardBorrowRewardPeriods, hardDelegatorRewardPeriods, v0_13incentive.Multipliers{v0_13incentive.NewMultiplier(v0_13incentive.Small, 1, sdk.MustNewDecFromStr("0.25")), v0_13incentive.NewMultiplier(v0_13incentive.Large, 12, sdk.MustNewDecFromStr("1.0"))}, ClaimEndTime)

	usdxGenAccumulationTimes := v0_13incentive.GenesisAccumulationTimes{}

	for _, rp := range params.USDXMintingRewardPeriods {
		gat := v0_13incentive.NewGenesisAccumulationTime(rp.CollateralType, GenesisTime, sdk.ZeroDec())
		usdxGenAccumulationTimes = append(usdxGenAccumulationTimes, gat)
	}
	hardSupplyGenAccumulationTimes := v0_13incentive.GenesisAccumulationTimes{}
	for _, rp := range params.HardSupplyRewardPeriods {
		gat := v0_13incentive.NewGenesisAccumulationTime(rp.CollateralType, GenesisTime, sdk.ZeroDec())
		hardSupplyGenAccumulationTimes = append(hardSupplyGenAccumulationTimes, gat)
	}
	hardBorrowGenAccumulationTimes := v0_13incentive.GenesisAccumulationTimes{}
	for _, rp := range params.HardBorrowRewardPeriods {
		gat := v0_13incentive.NewGenesisAccumulationTime(rp.CollateralType, GenesisTime, sdk.ZeroDec())
		hardBorrowGenAccumulationTimes = append(hardBorrowGenAccumulationTimes, gat)
	}

	hardDelegatorGenAccumulationTimes := v0_13incentive.GenesisAccumulationTimes{}

	for _, rp := range params.HardDelegatorRewardPeriods {
		gat := v0_13incentive.NewGenesisAccumulationTime(rp.CollateralType, GenesisTime, sdk.ZeroDec())
		hardDelegatorGenAccumulationTimes = append(hardDelegatorGenAccumulationTimes, gat)
	}

	usdxClaims := v0_13incentive.USDXMintingClaims{}
	usdxClaimMap := make(map[string]v0_13incentive.USDXMintingClaim)
	claimEndMap := make(map[uint64]time.Time)

	for _, cp := range incentiveGS.ClaimPeriods {
		claimEndMap[cp.ID] = cp.End
	}

	for _, claim := range incentiveGS.Claims {
		if claimEndMap[claim.ClaimPeriodID].Before(GenesisTime) {
			continue
		}
		newClaim, ok := usdxClaimMap[claim.Owner.String()]
		if !ok {
			newClaim = v0_13incentive.NewUSDXMintingClaim(claim.Owner, claim.Reward, v0_13incentive.RewardIndexes{v0_13incentive.NewRewardIndex(claim.CollateralType, sdk.ZeroDec())})
		} else {
			newClaim.Reward = newClaim.Reward.Add(claim.Reward)
			_, found := newClaim.RewardIndexes.GetRewardIndex(claim.CollateralType)
			if !found {
				newClaim.RewardIndexes = append(newClaim.RewardIndexes, v0_13incentive.NewRewardIndex(claim.CollateralType, sdk.ZeroDec()))
			}
		}
		usdxClaimMap[newClaim.Owner.String()] = newClaim
	}

	for _, claim := range usdxClaimMap {
		usdxClaims = append(usdxClaims, claim)
	}

	hardClaims := v0_13incentive.HardLiquidityProviderClaims{}
	hardClaimMap := make(map[string]v0_13incentive.HardLiquidityProviderClaim)

	for _, claim := range hardGS.Claims {
		newClaim, ok := hardClaimMap[claim.Owner.String()]
		if !ok {
			// if claim.Type == "lp" -- hard supply
			// if claim.Type == "stake" -- hard delegator
			// hard barrow always empty
			delegatorIndexes := v0_13incentive.RewardIndexes{}
			supplyIndexes := v0_13incentive.MultiRewardIndexes{}
			borrowIndexes := v0_13incentive.MultiRewardIndexes{}
			if claim.Type == v0_11hard.Stake {
				delegatorIndexes = v0_13incentive.RewardIndexes{v0_13incentive.NewRewardIndex(claim.DepositDenom, sdk.ZeroDec())}
			}
			if claim.Type == v0_11hard.LP {
				supplyIndexes = v0_13incentive.MultiRewardIndexes{v0_13incentive.NewMultiRewardIndex(claim.DepositDenom, v0_13incentive.RewardIndexes{v0_13incentive.NewRewardIndex("hard", sdk.ZeroDec())})}
			}
			newClaim = v0_13incentive.NewHardLiquidityProviderClaim(claim.Owner, sdk.NewCoins(claim.Amount), supplyIndexes, borrowIndexes, delegatorIndexes)
		} else {
			newClaim.Reward = sdk.NewCoins(claim.Amount)
			if claim.Type == v0_11hard.Stake {
				newClaim.DelegatorRewardIndexes = v0_13incentive.RewardIndexes{v0_13incentive.NewRewardIndex(claim.DepositDenom, sdk.ZeroDec())}
			}
			if claim.Type == v0_11hard.LP {
				_, found := newClaim.SupplyRewardIndexes.GetRewardIndex(claim.DepositDenom)
				if !found {
					newClaim.SupplyRewardIndexes = append(newClaim.SupplyRewardIndexes, v0_13incentive.NewMultiRewardIndex(claim.DepositDenom, v0_13incentive.RewardIndexes{v0_13incentive.NewRewardIndex("hard", sdk.ZeroDec())}))
				}
			}
		}
		hardClaimMap[claim.Owner.String()] = newClaim
	}

	for _, claim := range hardClaimMap {
		hardClaims = append(hardClaims, claim)
	}

	return v0_13incentive.NewGenesisState(
		params,
		usdxGenAccumulationTimes,
		hardSupplyGenAccumulationTimes,
		hardBorrowGenAccumulationTimes,
		hardDelegatorGenAccumulationTimes,
		usdxClaims,
		hardClaims,
	)
}

// MigrateAuth migrates from a v0.11 auth genesis state to a v0.13
func MigrateAuth(genesisState auth.GenesisState) auth.GenesisState {
	savingsRateMaccCoins := sdk.NewCoins()
	savingsMaccAddr := supply.NewModuleAddress(v0_11cdp.SavingsRateMacc)
	savingsRateMaccIndex := 0
	liquidatorMaccIndex := 0
	for idx, acc := range genesisState.Accounts {
		if acc.GetAddress().Equals(savingsMaccAddr) {
			savingsRateMaccCoins = acc.GetCoins()
			savingsRateMaccIndex = idx
			err := acc.SetCoins(acc.GetCoins().Sub(acc.GetCoins()))
			if err != nil {
				panic(err)
			}
		}
		if acc.GetAddress().Equals(supply.NewModuleAddress(v0_13cdp.LiquidatorMacc)) {
			liquidatorMaccIndex = idx
		}
	}
	liquidatorAcc := genesisState.Accounts[liquidatorMaccIndex]
	err := liquidatorAcc.SetCoins(liquidatorAcc.GetCoins().Add(savingsRateMaccCoins...))
	if err != nil {
		panic(err)
	}
	genesisState.Accounts[liquidatorMaccIndex] = liquidatorAcc

	genesisState.Accounts = removeIndex(genesisState.Accounts, savingsRateMaccIndex)
	return genesisState
}

func removeIndex(accs authexported.GenesisAccounts, index int) authexported.GenesisAccounts {
	ret := make(authexported.GenesisAccounts, 0)
	ret = append(ret, accs[:index]...)
	return append(ret, accs[index+1:]...)
}
