package keeper

import (
	"fmt"
	"sort"

	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"github.com/kava-labs/kava/x/cdp/types"
)

// NewQuerier returns a new querier function
func NewQuerier(keeper Keeper, legacyQuerierCdc *codec.LegacyAmino) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) (res []byte, err error) {
		switch path[0] {
		case types.QueryGetCdp:
			return queryGetCdp(ctx, req, keeper, legacyQuerierCdc)
		case types.QueryGetCdps:
			return queryGetCdps(ctx, req, keeper, legacyQuerierCdc)
		case types.QueryGetCdpDeposits:
			return queryGetDeposits(ctx, req, keeper, legacyQuerierCdc)
		case types.QueryGetCdpsByCollateralType: // legacy, maintained for REST API
			return queryGetCdpsByCollateralType(ctx, req, keeper, legacyQuerierCdc)
		case types.QueryGetCdpsByCollateralization: // legacy, maintained for REST API
			return queryGetCdpsByRatio(ctx, req, keeper, legacyQuerierCdc)
		case types.QueryGetParams:
			return queryGetParams(ctx, req, keeper, legacyQuerierCdc)
		case types.QueryGetAccounts:
			return queryGetAccounts(ctx, req, keeper, legacyQuerierCdc)
		case types.QueryGetTotalPrincipal:
			return queryGetTotalPrincipal(ctx, req, keeper, legacyQuerierCdc)
		case types.QueryGetTotalCollateral:
			return queryGetTotalCollateral(ctx, req, keeper, legacyQuerierCdc)
		default:
			return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unknown %s query endpoint %s", types.ModuleName, path[0])
		}
	}
}

// query a specific cdp
func queryGetCdp(ctx sdk.Context, req abci.RequestQuery, keeper Keeper, legacyQuerierCdc *codec.LegacyAmino) ([]byte, error) {
	var requestParams types.QueryCdpParams
	err := legacyQuerierCdc.UnmarshalJSON(req.Data, &requestParams)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
	}

	_, valid := keeper.GetCollateral(ctx, requestParams.CollateralType)
	if !valid {
		return nil, sdkerrors.Wrap(types.ErrInvalidCollateral, requestParams.CollateralType)
	}

	cdp, found := keeper.GetCdpByOwnerAndCollateralType(ctx, requestParams.Owner, requestParams.CollateralType)
	if !found {
		return nil, sdkerrors.Wrapf(types.ErrCdpNotFound, "owner %s, denom %s", requestParams.Owner, requestParams.CollateralType)
	}

	augmentedCDP := keeper.LoadAugmentedCDP(ctx, cdp)

	bz, err := codec.MarshalJSONIndent(legacyQuerierCdc, augmentedCDP)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}
	return bz, nil

}

// query deposits on a particular cdp
func queryGetDeposits(ctx sdk.Context, req abci.RequestQuery, keeper Keeper, legacyQuerierCdc *codec.LegacyAmino) ([]byte, error) {
	var requestParams types.QueryCdpDeposits
	err := legacyQuerierCdc.UnmarshalJSON(req.Data, &requestParams)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
	}

	_, valid := keeper.GetCollateral(ctx, requestParams.CollateralType)
	if !valid {
		return nil, sdkerrors.Wrap(types.ErrInvalidCollateral, requestParams.CollateralType)
	}

	cdp, found := keeper.GetCdpByOwnerAndCollateralType(ctx, requestParams.Owner, requestParams.CollateralType)
	if !found {
		return nil, sdkerrors.Wrapf(types.ErrCdpNotFound, "owner %s, denom %s", requestParams.Owner, requestParams.CollateralType)
	}

	deposits := keeper.GetDeposits(ctx, cdp.ID)

	bz, err := codec.MarshalJSONIndent(legacyQuerierCdc, deposits)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}
	return bz, nil

}

// query cdps with matching denom and ratio LESS THAN the input ratio
func queryGetCdpsByRatio(ctx sdk.Context, req abci.RequestQuery, keeper Keeper, legacyQuerierCdc *codec.LegacyAmino) ([]byte, error) {
	var requestParams types.QueryCdpsByRatioParams
	err := legacyQuerierCdc.UnmarshalJSON(req.Data, &requestParams)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
	}
	_, valid := keeper.GetCollateral(ctx, requestParams.CollateralType)
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
	bz, err := codec.MarshalJSONIndent(legacyQuerierCdc, augmentedCDPs)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}
	return bz, nil
}

// query all cdps with matching collateral type
func queryGetCdpsByCollateralType(ctx sdk.Context, req abci.RequestQuery, keeper Keeper, legacyQuerierCdc *codec.LegacyAmino) ([]byte, error) {
	var requestParams types.QueryCdpsByCollateralTypeParams
	err := legacyQuerierCdc.UnmarshalJSON(req.Data, &requestParams)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
	}
	_, valid := keeper.GetCollateral(ctx, requestParams.CollateralType)
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
	bz, err := codec.MarshalJSONIndent(legacyQuerierCdc, augmentedCDPs)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}
	return bz, nil
}

// query params in the cdp store
func queryGetParams(ctx sdk.Context, req abci.RequestQuery, keeper Keeper, legacyQuerierCdc *codec.LegacyAmino) ([]byte, error) {
	// Get params
	params := keeper.GetParams(ctx)

	// Encode results
	bz, err := codec.MarshalJSONIndent(legacyQuerierCdc, params)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}
	return bz, nil
}

// query cdp module accounts
func queryGetAccounts(ctx sdk.Context, req abci.RequestQuery, keeper Keeper, legacyQuerierCdc *codec.LegacyAmino) ([]byte, error) {
	cdpAccAccount := keeper.accountKeeper.GetModuleAccount(ctx, types.ModuleName)
	liquidatorAccAccount := keeper.accountKeeper.GetModuleAccount(ctx, types.LiquidatorMacc)

	accounts := []authtypes.ModuleAccount{
		*cdpAccAccount.(*authtypes.ModuleAccount),
		*liquidatorAccAccount.(*authtypes.ModuleAccount),
	}

	// Encode results
	bz, err := codec.MarshalJSONIndent(legacyQuerierCdc, accounts)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}
	return bz, nil
}

// query cdps in store and filter by request params
func queryGetCdps(ctx sdk.Context, req abci.RequestQuery, keeper Keeper, legacyQuerierCdc *codec.LegacyAmino) ([]byte, error) {
	var params types.QueryCdpsParams
	err := legacyQuerierCdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
	}

	// Filter CDPs
	filteredCDPs, err := FilterCDPs(ctx, keeper, params)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	bz, err := codec.MarshalJSONIndent(legacyQuerierCdc, filteredCDPs)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	return bz, nil
}

// query total amount of principal (ie. usdx) that has been minted with a particular collateral type
func queryGetTotalPrincipal(ctx sdk.Context, req abci.RequestQuery, keeper Keeper, legacyQuerierCdc *codec.LegacyAmino) ([]byte, error) {
	var params types.QueryGetTotalPrincipalParams
	err := legacyQuerierCdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
	}

	var queryCollateralTypes []string

	if params.CollateralType != "" {
		// Single collateralType provided
		queryCollateralTypes = append(queryCollateralTypes, params.CollateralType)
	} else {
		// No collateralType provided, respond with all of them
		keeperParams := keeper.GetParams(ctx)

		for _, collateral := range keeperParams.CollateralParams {
			queryCollateralTypes = append(queryCollateralTypes, collateral.Type)
		}
	}

	var collateralPrincipals types.TotalPrincipals

	for _, queryType := range queryCollateralTypes {
		// Hardcoded to default USDX
		principalAmount := keeper.GetTotalPrincipal(ctx, queryType, types.DefaultStableDenom)
		// Wrap it in an sdk.Coin
		totalAmountCoin := sdk.NewCoin(types.DefaultStableDenom, principalAmount)

		totalPrincipal := types.NewTotalPrincipal(queryType, totalAmountCoin)
		collateralPrincipals = append(collateralPrincipals, totalPrincipal)
	}

	bz, err := codec.MarshalJSONIndent(legacyQuerierCdc, collateralPrincipals)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	return bz, nil
}

// query total amount of collateral (ie. btcb) that has been deposited with a particular collateral type
func queryGetTotalCollateral(ctx sdk.Context, req abci.RequestQuery, keeper Keeper, legacyQuerierCdc *codec.LegacyAmino) ([]byte, error) {
	var request types.QueryGetTotalCollateralParams
	err := legacyQuerierCdc.UnmarshalJSON(req.Data, &request)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
	}

	params := keeper.GetParams(ctx)
	denomCollateralTypes := make(map[string][]string)

	// collect collateral types for each denom
	for _, collateralParam := range params.CollateralParams {
		denomCollateralTypes[collateralParam.Denom] =
			append(denomCollateralTypes[collateralParam.Denom], collateralParam.Type)
	}

	// sort collateral types alphabetically
	for _, collateralTypes := range denomCollateralTypes {
		sort.Slice(collateralTypes, func(i int, j int) bool {
			return collateralTypes[i] < collateralTypes[j]
		})
	}

	// get total collateral in all cdps
	cdpAccount := keeper.accountKeeper.GetModuleAccount(ctx, types.ModuleName)
	totalCdpCollateral := keeper.bankKeeper.GetAllBalances(ctx, cdpAccount.GetAddress())

	var response types.TotalCollaterals

	for denom, collateralTypes := range denomCollateralTypes {
		// skip any denoms that do not match the requested collateral type
		if request.CollateralType != "" {
			match := false
			for _, ctype := range collateralTypes {
				if ctype == request.CollateralType {
					match = true
				}
			}

			if !match {
				continue
			}
		}

		totalCollateral := totalCdpCollateral.AmountOf(denom)

		// we need to query individual cdps for denoms with more than one collateral type
		for i := len(collateralTypes) - 1; i > 0; i-- {
			cdps := keeper.GetAllCdpsByCollateralType(ctx, collateralTypes[i])

			collateral := sdk.ZeroInt()

			for _, cdp := range cdps {
				collateral = collateral.Add(cdp.Collateral.Amount)
			}

			totalCollateral = totalCollateral.Sub(collateral)

			// if we have no collateralType filter, or the filter matches, include it in the response
			if request.CollateralType == "" || collateralTypes[i] == request.CollateralType {
				response = append(response, types.NewTotalCollateral(collateralTypes[i], sdk.NewCoin(denom, collateral)))
			}

			// skip the rest of the cdp queries if we have a matching filter
			if collateralTypes[i] == request.CollateralType {
				break
			}
		}

		if request.CollateralType == "" || collateralTypes[0] == request.CollateralType {
			// all leftover total collateral belongs to the first collateral type
			response = append(response, types.NewTotalCollateral(collateralTypes[0], sdk.NewCoin(denom, totalCollateral)))
		}
	}

	// sort to ensure deterministic response
	sort.Slice(response, func(i int, j int) bool {
		return response[i].CollateralType < response[j].CollateralType
	})

	// encode response
	bz, err := codec.MarshalJSONIndent(legacyQuerierCdc, response)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	return bz, nil
}

// FilterCDPs queries the store for all CDPs that match query params
func FilterCDPs(ctx sdk.Context, k Keeper, params types.QueryCdpsParams) (types.AugmentedCDPs, error) {
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
			_, found := k.GetCollateral(ctx, params.CollateralType)
			if !found {
				return nil, fmt.Errorf("invalid collateral type")
			}
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
	if !params.Ratio.IsNil() && params.Ratio.GT(sdk.ZeroDec()) {
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
			return nil, nil
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
			return nil, nil
		}
	}
	if !params.Ratio.IsNil() && params.Ratio.GT(sdk.ZeroDec()) {
		if len(matchRatio) > 0 {
			if len(commonCDPs) > 0 {
				commonCDPs = FindIntersection(commonCDPs, matchRatio)
			} else {
				commonCDPs = matchRatio
			}
		} else {
			return nil, nil
		}
	}
	// Load augmented CDPs
	var augmentedCDPs types.AugmentedCDPs
	for _, cdp := range commonCDPs {
		augmentedCDP := k.LoadAugmentedCDP(ctx, cdp)
		augmentedCDPs = append(augmentedCDPs, augmentedCDP)
	}

	// Apply page and limit params
	start, end := client.Paginate(len(augmentedCDPs), params.Page, params.Limit, 100)
	if start < 0 || end < 0 {
		return nil, nil
	}

	return augmentedCDPs[start:end], nil
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
