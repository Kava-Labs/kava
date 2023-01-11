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
			// TODO: Removed to be replaced with updated queries for new store claims
			return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "query for %s is not supported", types.QueryGetSavingsRewards)
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
	stakingAPR, err := GetStakingAPR(ctx, k, params)
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
		apy, err := GetAPYFromMultiRewardPeriod(ctx, k, param.CollateralType, param, vaultTotalValue.Amount)
		if err != nil {
			return nil, err
		}

		apys = append(apys, types.NewAPY(param.CollateralType, apy))
	}

	// Marshal APYs
	res := types.NewQueryGetAPYsResponse(apys)
	bz, err := codec.MarshalJSONIndent(legacyQuerierCdc, res)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}
	return bz, nil
}

// GetStakingAPR returns the total APR for staking and incentive rewards
func GetStakingAPR(ctx sdk.Context, k Keeper, params types.Params) (sdk.Dec, error) {
	// Get staking APR
	stakingAPR := k.kavamintKeeper.GetStakingApy(ctx)

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
	// Override collateral type to use "kava" instead of "bkava" when fetching
	incentiveAPY, err := GetAPYFromMultiRewardPeriod(ctx, k, types.BondDenom, bkavaRewardPeriod, totalEarnBkavaDeposited)
	if err != nil {
		return sdk.ZeroDec(), err
	}

	totalAPY := stakingAPR.Add(incentiveAPY)
	return totalAPY, nil
}

// GetAPYFromMultiRewardPeriod calculates the APY for a given MultiRewardPeriod
func GetAPYFromMultiRewardPeriod(
	ctx sdk.Context,
	k Keeper,
	collateralType string,
	rewardPeriod types.MultiRewardPeriod,
	totalSupply sdk.Int,
) (sdk.Dec, error) {
	if totalSupply.IsZero() {
		return sdk.ZeroDec(), nil
	}

	// Get USD value of collateral type
	collateralUSDValue, err := k.pricefeedKeeper.GetCurrentPrice(ctx, getMarketID(collateralType))
	if err != nil {
		return sdk.ZeroDec(), fmt.Errorf(
			"failed to get price for incentive collateralType %s with market ID %s: %w",
			collateralType, getMarketID(collateralType), err,
		)
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
			return sdk.ZeroDec(), fmt.Errorf("failed to get price for RewardsPerSecond asset %s: %w", reward.Denom, err)
		}

		rewardPerSecond := reward.Amount.ToDec().Mul(rewardDenomUSDValue.Price)
		totalUSDRewardsPerSecond = totalUSDRewardsPerSecond.Add(rewardPerSecond)
	}

	// APY = USD rewards per second * seconds per year / USD total supplied
	apy := totalUSDRewardsPerSecond.
		MulInt64(SecondsPerYear).
		Quo(totalSupplyUSDValue)

	return apy, nil
}

func getMarketID(denom string) string {
	// Rewrite denoms as pricefeed has different names for some assets,
	// e.g. "ukava" -> "kava", "erc20/multichain/usdc" -> "usdc"
	// bkava is not included as it is handled separately

	// TODO: Replace hardcoded conversion with possible params set somewhere
	// to be more flexible. E.g. a map of denoms to pricefeed market denoms in
	// pricefeed params.
	switch denom {
	case types.BondDenom:
		denom = "kava"
	case "erc20/multichain/usdc":
		denom = "usdc"
	case "erc20/multichain/usdt":
		denom = "usdt"
	case "erc20/multichain/dai":
		denom = "dai"
	}

	return fmt.Sprintf("%s:usd:30", denom)
}
