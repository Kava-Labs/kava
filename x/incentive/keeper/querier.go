package keeper

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	abci "github.com/tendermint/tendermint/abci/types"

	earntypes "github.com/kava-labs/kava/x/earn/types"
	"github.com/kava-labs/kava/x/incentive/types"
	kavadisttypes "github.com/kava-labs/kava/x/kavadist/types"
	liquidtypes "github.com/kava-labs/kava/x/liquid/types"
)

const (
	SecondsPerYear = 31536000
)

// NewQuerier is the module level router for state queries
func NewQuerier(k Keeper, legacyQuerierCdc *codec.LegacyAmino) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) (res []byte, err error) {
		switch path[0] {
		case types.QueryGetParams:
			return queryGetParams(ctx, req, k, legacyQuerierCdc)

		case types.QueryGetHardRewards:
			return queryGetHardRewards(ctx, req, k, legacyQuerierCdc)
		case types.QueryGetUSDXMintingRewards:
			return queryGetUSDXMintingRewards(ctx, req, k, legacyQuerierCdc)
		case types.QueryGetDelegatorRewards:
			return queryGetDelegatorRewards(ctx, req, k, legacyQuerierCdc)
		case types.QueryGetSwapRewards:
			return queryGetSwapRewards(ctx, req, k, legacyQuerierCdc)
		case types.QueryGetSavingsRewards:
			return queryGetSavingsRewards(ctx, req, k, legacyQuerierCdc)
		case types.QueryGetRewardFactors:
			return queryGetRewardFactors(ctx, req, k, legacyQuerierCdc)
		case types.QueryGetEarnRewards:
			return queryGetEarnRewards(ctx, req, k, legacyQuerierCdc)
		case types.QueryGetAPYs:
			return queryGetAPYs(ctx, req, k, legacyQuerierCdc)
		default:
			return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unknown %s query endpoint", types.ModuleName)
		}
	}
}

// query params in the store
func queryGetParams(ctx sdk.Context, req abci.RequestQuery, k Keeper, legacyQuerierCdc *codec.LegacyAmino) ([]byte, error) {
	// Get params
	params := k.GetParams(ctx)

	// Encode results
	bz, err := codec.MarshalJSONIndent(legacyQuerierCdc, params)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}
	return bz, nil
}

func queryGetHardRewards(ctx sdk.Context, req abci.RequestQuery, k Keeper, legacyQuerierCdc *codec.LegacyAmino) ([]byte, error) {
	var params types.QueryRewardsParams
	err := legacyQuerierCdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
	}
	owner := len(params.Owner) > 0

	var hardClaims types.HardLiquidityProviderClaims
	switch {
	case owner:
		hardClaim, foundHardClaim := k.GetHardLiquidityProviderClaim(ctx, params.Owner)
		if foundHardClaim {
			hardClaims = append(hardClaims, hardClaim)
		}
	default:
		hardClaims = k.GetAllHardLiquidityProviderClaims(ctx)
	}

	var paginatedHardClaims types.HardLiquidityProviderClaims
	startH, endH := client.Paginate(len(hardClaims), params.Page, params.Limit, 100)
	if startH < 0 || endH < 0 {
		paginatedHardClaims = types.HardLiquidityProviderClaims{}
	} else {
		paginatedHardClaims = hardClaims[startH:endH]
	}

	if !params.Unsynchronized {
		for i, claim := range paginatedHardClaims {
			paginatedHardClaims[i] = k.SimulateHardSynchronization(ctx, claim)
		}
	}

	// Marshal Hard claims
	bz, err := codec.MarshalJSONIndent(legacyQuerierCdc, paginatedHardClaims)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}
	return bz, nil
}

func queryGetUSDXMintingRewards(ctx sdk.Context, req abci.RequestQuery, k Keeper, legacyQuerierCdc *codec.LegacyAmino) ([]byte, error) {
	var params types.QueryRewardsParams
	err := legacyQuerierCdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
	}
	owner := len(params.Owner) > 0

	var usdxMintingClaims types.USDXMintingClaims
	switch {
	case owner:
		usdxMintingClaim, foundUsdxMintingClaim := k.GetUSDXMintingClaim(ctx, params.Owner)
		if foundUsdxMintingClaim {
			usdxMintingClaims = append(usdxMintingClaims, usdxMintingClaim)
		}
	default:
		usdxMintingClaims = k.GetAllUSDXMintingClaims(ctx)
	}

	var paginatedUsdxMintingClaims types.USDXMintingClaims
	startU, endU := client.Paginate(len(usdxMintingClaims), params.Page, params.Limit, 100)
	if startU < 0 || endU < 0 {
		paginatedUsdxMintingClaims = types.USDXMintingClaims{}
	} else {
		paginatedUsdxMintingClaims = usdxMintingClaims[startU:endU]
	}

	if !params.Unsynchronized {
		for i, claim := range paginatedUsdxMintingClaims {
			paginatedUsdxMintingClaims[i] = k.SimulateUSDXMintingSynchronization(ctx, claim)
		}
	}

	// Marshal USDX minting claims
	bz, err := codec.MarshalJSONIndent(legacyQuerierCdc, paginatedUsdxMintingClaims)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}
	return bz, nil
}

func queryGetDelegatorRewards(ctx sdk.Context, req abci.RequestQuery, k Keeper, legacyQuerierCdc *codec.LegacyAmino) ([]byte, error) {
	var params types.QueryRewardsParams
	err := legacyQuerierCdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
	}
	owner := len(params.Owner) > 0

	var delegatorClaims types.DelegatorClaims
	switch {
	case owner:
		delegatorClaim, foundDelegatorClaim := k.GetDelegatorClaim(ctx, params.Owner)
		if foundDelegatorClaim {
			delegatorClaims = append(delegatorClaims, delegatorClaim)
		}
	default:
		delegatorClaims = k.GetAllDelegatorClaims(ctx)
	}

	var paginatedDelegatorClaims types.DelegatorClaims
	startH, endH := client.Paginate(len(delegatorClaims), params.Page, params.Limit, 100)
	if startH < 0 || endH < 0 {
		paginatedDelegatorClaims = types.DelegatorClaims{}
	} else {
		paginatedDelegatorClaims = delegatorClaims[startH:endH]
	}

	if !params.Unsynchronized {
		for i, claim := range paginatedDelegatorClaims {
			paginatedDelegatorClaims[i] = k.SimulateDelegatorSynchronization(ctx, claim)
		}
	}

	// Marshal Hard claims
	bz, err := codec.MarshalJSONIndent(legacyQuerierCdc, paginatedDelegatorClaims)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}
	return bz, nil
}

func queryGetSwapRewards(ctx sdk.Context, req abci.RequestQuery, k Keeper, legacyQuerierCdc *codec.LegacyAmino) ([]byte, error) {
	var params types.QueryRewardsParams
	err := legacyQuerierCdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
	}
	owner := len(params.Owner) > 0

	var claims types.SwapClaims
	switch {
	case owner:
		claim, found := k.GetSwapClaim(ctx, params.Owner)
		if found {
			claims = append(claims, claim)
		}
	default:
		claims = k.GetAllSwapClaims(ctx)
	}

	var paginatedClaims types.SwapClaims
	startH, endH := client.Paginate(len(claims), params.Page, params.Limit, 100)
	if startH < 0 || endH < 0 {
		paginatedClaims = types.SwapClaims{}
	} else {
		paginatedClaims = claims[startH:endH]
	}

	if !params.Unsynchronized {
		for i, claim := range paginatedClaims {
			syncedClaim, found := k.GetSynchronizedSwapClaim(ctx, claim.Owner)
			if !found {
				panic("previously found claim should still be found")
			}
			paginatedClaims[i] = syncedClaim
		}
	}

	// Marshal claims
	bz, err := codec.MarshalJSONIndent(legacyQuerierCdc, paginatedClaims)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}
	return bz, nil
}

func queryGetSavingsRewards(ctx sdk.Context, req abci.RequestQuery, k Keeper, legacyQuerierCdc *codec.LegacyAmino) ([]byte, error) {
	var params types.QueryRewardsParams
	err := legacyQuerierCdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
	}
	owner := len(params.Owner) > 0

	var claims types.SavingsClaims
	switch {
	case owner:
		claim, found := k.GetSavingsClaim(ctx, params.Owner)
		if found {
			claims = append(claims, claim)
		}
	default:
		claims = k.GetAllSavingsClaims(ctx)
	}

	var paginatedClaims types.SavingsClaims
	startH, endH := client.Paginate(len(claims), params.Page, params.Limit, 100)
	if startH < 0 || endH < 0 {
		paginatedClaims = types.SavingsClaims{}
	} else {
		paginatedClaims = claims[startH:endH]
	}

	if !params.Unsynchronized {
		for i, claim := range paginatedClaims {
			syncedClaim, found := k.GetSynchronizedSavingsClaim(ctx, claim.Owner)
			if !found {
				panic("previously found claim should still be found")
			}
			paginatedClaims[i] = syncedClaim
		}
	}

	// Marshal claims
	bz, err := codec.MarshalJSONIndent(legacyQuerierCdc, paginatedClaims)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}
	return bz, nil
}

func queryGetEarnRewards(ctx sdk.Context, req abci.RequestQuery, k Keeper, legacyQuerierCdc *codec.LegacyAmino) ([]byte, error) {
	var params types.QueryRewardsParams
	err := legacyQuerierCdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
	}
	owner := len(params.Owner) > 0

	var claims types.EarnClaims
	switch {
	case owner:
		claim, found := k.GetEarnClaim(ctx, params.Owner)
		if found {
			claims = append(claims, claim)
		}
	default:
		claims = k.GetAllEarnClaims(ctx)
	}

	var paginatedClaims types.EarnClaims
	startH, endH := client.Paginate(len(claims), params.Page, params.Limit, 100)
	if startH < 0 || endH < 0 {
		paginatedClaims = types.EarnClaims{}
	} else {
		paginatedClaims = claims[startH:endH]
	}

	if !params.Unsynchronized {
		for i, claim := range paginatedClaims {
			syncedClaim, found := k.GetSynchronizedEarnClaim(ctx, claim.Owner)
			if !found {
				panic("previously found claim should still be found")
			}
			paginatedClaims[i] = syncedClaim
		}
	}

	// Marshal claims
	bz, err := codec.MarshalJSONIndent(legacyQuerierCdc, paginatedClaims)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}
	return bz, nil
}

func queryGetRewardFactors(ctx sdk.Context, req abci.RequestQuery, k Keeper, legacyQuerierCdc *codec.LegacyAmino) ([]byte, error) {
	var usdxFactors types.RewardIndexes
	k.IterateUSDXMintingRewardFactors(ctx, func(collateralType string, factor sdk.Dec) (stop bool) {
		usdxFactors = usdxFactors.With(collateralType, factor)
		return false
	})

	var supplyFactors types.MultiRewardIndexes
	k.IterateHardSupplyRewardIndexes(ctx, func(denom string, indexes types.RewardIndexes) (stop bool) {
		supplyFactors = supplyFactors.With(denom, indexes)
		return false
	})

	var borrowFactors types.MultiRewardIndexes
	k.IterateHardBorrowRewardIndexes(ctx, func(denom string, indexes types.RewardIndexes) (stop bool) {
		borrowFactors = borrowFactors.With(denom, indexes)
		return false
	})

	var delegatorFactors types.MultiRewardIndexes
	k.IterateDelegatorRewardIndexes(ctx, func(denom string, indexes types.RewardIndexes) (stop bool) {
		delegatorFactors = delegatorFactors.With(denom, indexes)
		return false
	})

	var swapFactors types.MultiRewardIndexes
	k.IterateSwapRewardIndexes(ctx, func(poolID string, indexes types.RewardIndexes) (stop bool) {
		swapFactors = swapFactors.With(poolID, indexes)
		return false
	})

	var savingsFactors types.MultiRewardIndexes
	k.IterateSavingsRewardIndexes(ctx, func(denom string, indexes types.RewardIndexes) (stop bool) {
		savingsFactors = savingsFactors.With(denom, indexes)
		return false
	})

	var earnFactors types.MultiRewardIndexes
	k.IterateEarnRewardIndexes(ctx, func(denom string, indexes types.RewardIndexes) (stop bool) {
		earnFactors = earnFactors.With(denom, indexes)
		return false
	})

	response := types.NewQueryGetRewardFactorsResponse(
		usdxFactors,
		supplyFactors,
		borrowFactors,
		delegatorFactors,
		swapFactors,
		savingsFactors,
		earnFactors,
	)

	bz, err := codec.MarshalJSONIndent(legacyQuerierCdc, response)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	return bz, nil
}

func queryGetAPYs(ctx sdk.Context, req abci.RequestQuery, k Keeper, legacyQuerierCdc *codec.LegacyAmino) ([]byte, error) {
	params := k.GetParams(ctx)
	var apys types.APYs

	// bkava APY (staking + incentive rewards)
	stakingAPR, err := getStakingAPR(ctx, k, params)
	if err != nil {
		return nil, err
	}

	apys = append(apys, types.NewAPY(liquidtypes.DefaultDerivativeDenom, stakingAPR))

	// Incentive only APYs
	for _, param := range params.EarnRewardPeriods {
		// Skip bkava as it's calculated earlier with staking rewards
		if param.CollateralType == liquidtypes.DefaultDerivativeDenom {
			continue
		}

		// Value in the vault in the same denom as CollateralType
		vaultTotalValue, err := k.earnKeeper.GetVaultTotalValue(ctx, param.CollateralType)
		if err != nil {
			return nil, err
		}
		apy, err := getAPYFromMultiRewardPeriod(ctx, k, param, vaultTotalValue.Amount)
		if err != nil {
			return nil, err
		}

		apys = append(apys, apy)
	}

	// Marshal APYs
	res := types.NewQueryGetAPYsResponse(apys)
	bz, err := codec.MarshalJSONIndent(legacyQuerierCdc, res)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}
	return bz, nil
}

func getStakingAPR(ctx sdk.Context, k Keeper, params types.Params) (sdk.Dec, error) {
	// Get staking APR + incentive APR
	inflationRate := k.mintKeeper.GetMinter(ctx).Inflation
	communityTax := k.distrKeeper.GetCommunityTax(ctx)
	infrastructureTaxPeriods := k.kavadistKeeper.GetParams(ctx).InfrastructureParams.InfrastructurePeriods
	infrastructureTax := GetTotalInfrastructureInflation(ctx, infrastructureTaxPeriods)

	bondedTokens := k.stakingKeeper.TotalBondedTokens(ctx)
	circulatingSupply := k.bankKeeper.GetSupply(ctx, types.BondDenom)

	// Staking APR = (Inflation Rate - Community Tax - Infrastructure Tax) / (Bonded Tokens / Circulating Supply)
	stakingAPR := inflationRate.
		Sub(communityTax).
		Sub(infrastructureTax).
		Quo(bondedTokens.ToDec().
			Quo(circulatingSupply.Amount.ToDec()))

	// Get incentive APR
	bkavaRewardPeriod, found := params.EarnRewardPeriods.GetMultiRewardPeriod(liquidtypes.DefaultDerivativeDenom)
	if !found {
		// No incentive rewards for bkava, only staking rewards
		return stakingAPR, nil
	}

	// Total amount of bkava in earn vaults, this may be lower than total bank
	// supply of bkava as some bkava may not be deposited in earn vaults
	totalEarnBkavaDeposited := sdk.ZeroInt()

	var iterErr error
	k.earnKeeper.IterateVaultRecords(ctx, func(record earntypes.VaultRecord) (stop bool) {
		if !k.liquidKeeper.IsDerivativeDenom(ctx, record.TotalShares.Denom) {
			return false
		}

		vaultValue, err := k.earnKeeper.GetVaultTotalValue(ctx, record.TotalShares.Denom)
		if err != nil {
			iterErr = err
			return false
		}

		totalEarnBkavaDeposited = totalEarnBkavaDeposited.Add(vaultValue.Amount)

		return false
	})

	if iterErr != nil {
		return sdk.ZeroDec(), iterErr
	}

	// Incentive APR = rewards per second * seconds per year / total supplied to earn vaults
	incentiveAPY := bkavaRewardPeriod.RewardsPerSecond.AmountOf(types.BondDenom).ToDec().
		MulInt64(SecondsPerYear).
		QuoInt(totalEarnBkavaDeposited)

	totalAPY := stakingAPR.Add(incentiveAPY)
	return totalAPY, nil
}

func GetTotalInfrastructureInflation(
	ctx sdk.Context,
	infrastructureTaxPeriods kavadisttypes.Periods,
) sdk.Dec {
	totalPerSecondInflation := sdk.ZeroDec()

	for _, period := range infrastructureTaxPeriods {
		// Skip periods that already ended or haven't started yet.
		// Only consider periods that are active **now** at this given blocktime
		// and not those that expired between past block and this block.
		if period.End.Before(ctx.BlockTime()) || period.Start.After(ctx.BlockTime()) {
			continue
		}

		// Inflation is represented as a new percent of *total* tokens in supply,
		// not just a percentage of what to mint.
		// For example: 1.000000003022265980 as 10% inflation
		// So we have to subtract 1.0 to get only the percentage of minted tokens.
		totalPerSecondInflation = totalPerSecondInflation.Add(period.Inflation.Sub(sdk.OneDec()))
	}

	// Per second minting  = 0.000000003022265980
	// Convert to yearly % = 0.000000003022265980 * seconds per year (31536000)
	//                     = 0.09531017994528 per year (9.531%)
	totalAPR := totalPerSecondInflation.MulInt64(SecondsPerYear)
	return totalAPR
}

func getAPYFromMultiRewardPeriod(
	ctx sdk.Context,
	k Keeper,
	rewardPeriod types.MultiRewardPeriod,
	totalSupply sdk.Int,
) (types.APY, error) {
	// Get USD value of collateral type
	collateralUSDValue, err := k.pricefeedKeeper.GetCurrentPrice(ctx, getMarketID(rewardPeriod.CollateralType))
	if err != nil {
		return types.APY{}, err
	}

	// Total USD value of the collateral type total supply
	totalSupplyUSDValue := totalSupply.ToDec().Mul(collateralUSDValue.Price)

	totalUSDRewardsPerSecond := sdk.ZeroDec()

	// In many cases, RewardsPerSecond are assets that are different from the
	// CollateralType, so we need to use the USD value of CollateralType and
	// RewardsPerSecond to determine the APY.
	for _, reward := range rewardPeriod.RewardsPerSecond {
		// Get USD value of 1 unit of reward asset type, using TWAP
		rewardDenomUSDValue, err := k.pricefeedKeeper.GetCurrentPrice(ctx, getMarketID(reward.Denom))
		if err != nil {
			return types.APY{}, fmt.Errorf("failed to get price for RewardsPerSecond asset %s: %w", reward.Denom, err)
		}

		rewardPerSecond := reward.Amount.ToDec().Mul(rewardDenomUSDValue.Price)
		totalUSDRewardsPerSecond = totalUSDRewardsPerSecond.Add(rewardPerSecond)
	}

	// APY = USD rewards per second * seconds per year / USD total supplied
	apy := totalUSDRewardsPerSecond.
		MulInt64(SecondsPerYear).
		Quo(totalSupplyUSDValue)

	return types.NewAPY(rewardPeriod.CollateralType, apy), nil
}

func getMarketID(denom string) string {
	return fmt.Sprintf("%s:usd:30", denom)
}
