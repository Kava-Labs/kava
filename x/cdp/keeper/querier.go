package keeper

import (
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	supply "github.com/cosmos/cosmos-sdk/x/supply"

	"github.com/kava-labs/kava/x/cdp/types"
)

// NewQuerier returns a new querier function
func NewQuerier(keeper Keeper) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) (res []byte, err error) {
		switch path[0] {
		case types.QueryGetCdp:
			return queryGetCdp(ctx, req, keeper)
		case types.QueryGetCdps:
			return queryGetCdps(ctx, req, keeper)
		case types.QueryGetCdpDeposits:
			return queryGetDeposits(ctx, req, keeper)
		case types.QueryGetCdpsByCollateralType: // legacy, maintained for REST API
			return queryGetCdpsByCollateralType(ctx, req, keeper)
		case types.QueryGetCdpsByCollateralization: // legacy, maintained for REST API
			return queryGetCdpsByRatio(ctx, req, keeper)
		case types.QueryGetParams:
			return queryGetParams(ctx, req, keeper)
		case types.QueryGetAccounts:
			return queryGetAccounts(ctx, req, keeper)
		case types.QueryGetSavingsRateDistributed:
			return queryGetSavingsRateDistributed(ctx, req, keeper)
		default:
			return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unknown %s query endpoint %s", types.ModuleName, path[0])
		}
	}
}

// query a specific cdp
func queryGetCdp(ctx sdk.Context, req abci.RequestQuery, keeper Keeper) ([]byte, error) {
	var requestParams types.QueryCdpParams
	err := types.ModuleCdc.UnmarshalJSON(req.Data, &requestParams)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
	}

	_, valid := keeper.GetCollateralTypePrefix(ctx, requestParams.CollateralType)
	if !valid {
		return nil, sdkerrors.Wrap(types.ErrInvalidCollateral, requestParams.CollateralType)
	}

	cdp, found := keeper.GetCdpByOwnerAndCollateralType(ctx, requestParams.Owner, requestParams.CollateralType)
	if !found {
		return nil, sdkerrors.Wrapf(types.ErrCdpNotFound, "owner %s, denom %s", requestParams.Owner, requestParams.CollateralType)
	}

	augmentedCDP := keeper.LoadAugmentedCDP(ctx, cdp)

	bz, err := codec.MarshalJSONIndent(types.ModuleCdc, augmentedCDP)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}
	return bz, nil

}

// query deposits on a particular cdp
func queryGetDeposits(ctx sdk.Context, req abci.RequestQuery, keeper Keeper) ([]byte, error) {
	var requestParams types.QueryCdpDeposits
	err := types.ModuleCdc.UnmarshalJSON(req.Data, &requestParams)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
	}

	_, valid := keeper.GetCollateralTypePrefix(ctx, requestParams.CollateralType)
	if !valid {
		return nil, sdkerrors.Wrap(types.ErrInvalidCollateral, requestParams.CollateralType)
	}

	cdp, found := keeper.GetCdpByOwnerAndCollateralType(ctx, requestParams.Owner, requestParams.CollateralType)
	if !found {
		return nil, sdkerrors.Wrapf(types.ErrCdpNotFound, "owner %s, denom %s", requestParams.Owner, requestParams.CollateralType)
	}

	deposits := keeper.GetDeposits(ctx, cdp.ID)

	bz, err := codec.MarshalJSONIndent(types.ModuleCdc, deposits)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}
	return bz, nil

}

// query cdps with matching denom and ratio LESS THAN the input ratio
func queryGetCdpsByRatio(ctx sdk.Context, req abci.RequestQuery, keeper Keeper) ([]byte, error) {
	var requestParams types.QueryCdpsByRatioParams
	err := types.ModuleCdc.UnmarshalJSON(req.Data, &requestParams)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
	}
	_, valid := keeper.GetCollateralTypePrefix(ctx, requestParams.CollateralType)
	if !valid {
		return nil, sdkerrors.Wrap(types.ErrInvalidCollateral, requestParams.CollateralType)
	}

	ratio, err := keeper.CalculateCollateralizationRatioFromAbsoluteRatio(ctx, requestParams.CollateralType, requestParams.Ratio, "liquidation")
	if err != nil {
		return nil, sdkerrors.Wrap(err, "couldn't get collateralization ratio from absolute ratio")
	}

	cdps := keeper.GetAllCdpsByCollateralTypeAndRatio(ctx, requestParams.CollateralType, ratio)
	// augment CDPs by adding collateral value and collateralization ratio
	var augmentedCDPs types.AugmentedCDPs
	for _, cdp := range cdps {
		augmentedCDP := keeper.LoadAugmentedCDP(ctx, cdp)
		augmentedCDPs = append(augmentedCDPs, augmentedCDP)
	}
	bz, err := codec.MarshalJSONIndent(types.ModuleCdc, augmentedCDPs)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}
	return bz, nil
}

// query all cdps with matching collateral type
func queryGetCdpsByCollateralType(ctx sdk.Context, req abci.RequestQuery, keeper Keeper) ([]byte, error) {
	var requestParams types.QueryCdpsByCollateralTypeParams
	err := types.ModuleCdc.UnmarshalJSON(req.Data, &requestParams)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
	}
	_, valid := keeper.GetCollateralTypePrefix(ctx, requestParams.CollateralType)
	if !valid {
		return nil, sdkerrors.Wrap(types.ErrInvalidCollateral, requestParams.CollateralType)
	}

	cdps := keeper.GetAllCdpsByCollateralType(ctx, requestParams.CollateralType)
	// augment CDPs by adding collateral value and collateralization ratio
	var augmentedCDPs types.AugmentedCDPs
	for _, cdp := range cdps {
		augmentedCDP := keeper.LoadAugmentedCDP(ctx, cdp)
		augmentedCDPs = append(augmentedCDPs, augmentedCDP)
	}
	bz, err := codec.MarshalJSONIndent(types.ModuleCdc, augmentedCDPs)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}
	return bz, nil
}

// query params in the cdp store
func queryGetParams(ctx sdk.Context, req abci.RequestQuery, keeper Keeper) ([]byte, error) {
	// Get params
	params := keeper.GetParams(ctx)

	// Encode results
	bz, err := codec.MarshalJSONIndent(types.ModuleCdc, params)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}
	return bz, nil
}

// query cdp module accounts
func queryGetAccounts(ctx sdk.Context, req abci.RequestQuery, keeper Keeper) ([]byte, error) {
	cdpAccAccount := keeper.supplyKeeper.GetModuleAccount(ctx, types.ModuleName)
	liquidatorAccAccount := keeper.supplyKeeper.GetModuleAccount(ctx, types.LiquidatorMacc)
	savingsRateAccAccount := keeper.supplyKeeper.GetModuleAccount(ctx, types.SavingsRateMacc)

	accounts := []supply.ModuleAccount{
		*cdpAccAccount.(*supply.ModuleAccount),
		*liquidatorAccAccount.(*supply.ModuleAccount),
		*savingsRateAccAccount.(*supply.ModuleAccount),
	}

	// Encode results
	bz, err := codec.MarshalJSONIndent(supply.ModuleCdc, accounts)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}
	return bz, nil
}

// query get savings rate distributed in the cdp store
func queryGetSavingsRateDistributed(ctx sdk.Context, req abci.RequestQuery, keeper Keeper) ([]byte, error) {
	// Get savings rate distributed
	savingsRateDist := keeper.GetSavingsRateDistributed(ctx)

	// Encode results
	bz, err := codec.MarshalJSONIndent(types.ModuleCdc, savingsRateDist)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}
	return bz, nil
}

// query cdps in store and filter by request params
func queryGetCdps(ctx sdk.Context, req abci.RequestQuery, keeper Keeper) ([]byte, error) {
	var params types.QueryCdpsParams
	err := types.ModuleCdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
	}

	// Filter CDPs
	filteredCDPs := FilterCDPs(ctx, keeper, params)

	bz, err := codec.MarshalJSONIndent(keeper.cdc, filteredCDPs)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	return bz, nil
}

// FilterCDPs queries the store for all CDPs that match query params
func FilterCDPs(ctx sdk.Context, k Keeper, params types.QueryCdpsParams) types.AugmentedCDPs {
	var matchCollateralType, matchOwner, matchID, matchRatio types.CDPs

	// match cdp owner (if supplied)
	if len(params.Owner) > 0 {
		denoms := k.GetCollateralTypes(ctx)
		for _, denom := range denoms {
			cdp, found := k.GetCdpByOwnerAndCollateralType(ctx, params.Owner, denom)
			if found {
				matchOwner = append(matchOwner, cdp)
			}
		}
	}

	// match cdp collateral denom (if supplied)
	if len(params.CollateralType) > 0 {
		// if owner is specified only iterate over already matched cdps for efficiency
		if len(params.Owner) > 0 {
			for _, cdp := range matchOwner {
				if cdp.Type == params.CollateralType {
					matchCollateralType = append(matchCollateralType, cdp)
				}
			}
		} else {
			matchCollateralType = k.GetAllCdpsByCollateralType(ctx, params.CollateralType)
		}
	}

	// match cdp ID (if supplied)
	if params.ID != 0 {
		denoms := k.GetCollateralTypes(ctx)
		for _, denom := range denoms {
			cdp, found := k.GetCDP(ctx, denom, params.ID)
			if found {
				matchID = append(matchID, cdp)
			}
		}
	}

	// match cdp ratio (if supplied)
	if params.Ratio.GT(sdk.ZeroDec()) {
		denoms := k.GetCollateralTypes(ctx)
		for _, denom := range denoms {
			ratio, err := k.CalculateCollateralizationRatioFromAbsoluteRatio(ctx, denom, params.Ratio, "liquidation")
			if err != nil {
				continue
			}
			cdpsUnderRatio := k.GetAllCdpsByCollateralTypeAndRatio(ctx, denom, ratio)
			matchRatio = append(matchRatio, cdpsUnderRatio...)
		}
	}

	var commonCDPs types.CDPs
	// If no params specified, fetch all CDPs
	if len(params.CollateralType) == 0 && len(params.Owner) == 0 && params.ID == 0 && params.Ratio.Equal(sdk.ZeroDec()) {
		commonCDPs = k.GetAllCdps(ctx)
	}

	// Find the intersection of any matched CDPs
	if len(params.CollateralType) > 0 {
		if len(matchCollateralType) > 0 {
			commonCDPs = matchCollateralType
		} else {
			return types.AugmentedCDPs{}
		}
	}

	if len(params.Owner) > 0 {
		if len(matchCollateralType) > 0 {
			if len(commonCDPs) > 0 {
				commonCDPs = FindIntersection(commonCDPs, matchOwner)
			} else {
				commonCDPs = matchOwner
			}
		} else {
			commonCDPs = matchOwner
		}
	}
	if params.ID != 0 {
		if len(matchID) > 0 {
			if len(commonCDPs) > 0 {
				commonCDPs = FindIntersection(commonCDPs, matchID)
			} else {
				commonCDPs = matchID
			}
		} else {
			return types.AugmentedCDPs{}
		}
	}
	if params.Ratio.GT(sdk.ZeroDec()) {
		if len(matchRatio) > 0 {
			if len(commonCDPs) > 0 {
				commonCDPs = FindIntersection(commonCDPs, matchRatio)
			} else {
				commonCDPs = matchRatio
			}
		} else {
			return types.AugmentedCDPs{}
		}
	}
	// Load augmented CDPs
	var augmentedCDPs types.AugmentedCDPs
	for _, cdp := range commonCDPs {
		augmentedCDP := k.LoadAugmentedCDP(ctx, cdp)
		augmentedCDPs = append(augmentedCDPs, augmentedCDP)
	}

	// Apply page and limit params
	var paginatedCDPs types.AugmentedCDPs
	start, end := client.Paginate(len(augmentedCDPs), params.Page, params.Limit, 100)
	if start < 0 || end < 0 {
		paginatedCDPs = types.AugmentedCDPs{}
	} else {
		paginatedCDPs = augmentedCDPs[start:end]
	}

	return paginatedCDPs
}

// FindIntersection finds the intersection of two CDP arrays in linear time complexity O(n + n)
func FindIntersection(x types.CDPs, y types.CDPs) types.CDPs {
	cdpSet := make(types.CDPs, 0)
	cdpMap := make(map[uint64]bool)

	for i := 0; i < len(x); i++ {
		cdp := x[i]
		cdpMap[cdp.ID] = true
	}

	for i := 0; i < len(y); i++ {
		cdp := y[i]
		if _, found := cdpMap[cdp.ID]; found {
			cdpSet = append(cdpSet, cdp)
		}
	}

	return cdpSet
}
