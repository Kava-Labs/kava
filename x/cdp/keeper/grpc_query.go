package keeper

import (
	"context"
	"sort"

	"github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/query"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/kava-labs/kava/x/cdp/types"
)

type QueryServer struct {
	keeper Keeper
}

// NewQueryServer returns an implementation of the pricefeed MsgServer interface
// for the provided Keeper.
func NewQueryServerImpl(keeper Keeper) types.QueryServer {
	return &QueryServer{keeper: keeper}
}

var _ types.QueryServer = QueryServer{}

// Params queries all parameters of the cdp module.
func (s QueryServer) Params(c context.Context, req *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	if req == nil {
		return nil, status.Errorf(codes.InvalidArgument, "empty request")
	}

	sdkCtx := sdk.UnwrapSDKContext(c)
	params := s.keeper.GetParams(sdkCtx)

	return &types.QueryParamsResponse{Params: params}, nil
}

// Accounts queries the CDP module accounts.
func (s QueryServer) Accounts(c context.Context, req *types.QueryAccountsRequest) (*types.QueryAccountsResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	cdpAccAccount := s.keeper.accountKeeper.GetModuleAccount(ctx, types.ModuleName)
	liquidatorAccAccount := s.keeper.accountKeeper.GetModuleAccount(ctx, types.LiquidatorMacc)

	accounts := []authtypes.ModuleAccount{
		*cdpAccAccount.(*authtypes.ModuleAccount),
		*liquidatorAccAccount.(*authtypes.ModuleAccount),
	}

	return &types.QueryAccountsResponse{Accounts: accounts}, nil
}

// TotalPrincipal queries the total principal of a given collateral type.
func (s QueryServer) TotalPrincipal(c context.Context, req *types.QueryTotalPrincipalRequest) (*types.QueryTotalPrincipalResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	var queryCollateralTypes []string

	if req.CollateralType != "" {
		// Single collateralType provided
		queryCollateralTypes = append(queryCollateralTypes, req.CollateralType)
	} else {
		// No collateralType provided, respond with all of them
		keeperParams := s.keeper.GetParams(ctx)

		for _, collateral := range keeperParams.CollateralParams {
			queryCollateralTypes = append(queryCollateralTypes, collateral.Type)
		}
	}

	var collateralPrincipals types.TotalPrincipals

	for _, queryType := range queryCollateralTypes {
		// Hardcoded to default USDX
		principalAmount := s.keeper.GetTotalPrincipal(ctx, queryType, types.DefaultStableDenom)
		// Wrap it in an sdk.Coin
		totalAmountCoin := sdk.NewCoin(types.DefaultStableDenom, principalAmount)

		totalPrincipal := types.NewTotalPrincipal(queryType, totalAmountCoin)
		collateralPrincipals = append(collateralPrincipals, totalPrincipal)
	}

	return &types.QueryTotalPrincipalResponse{
		TotalPrincipal: collateralPrincipals,
	}, nil
}

// TotalCollateral queries the total collateral of a given collateral type.
func (s QueryServer) TotalCollateral(c context.Context, req *types.QueryTotalCollateralRequest) (*types.QueryTotalCollateralResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	params := s.keeper.GetParams(ctx)
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
	cdpAccount := s.keeper.accountKeeper.GetModuleAccount(ctx, types.ModuleName)
	totalCdpCollateral := s.keeper.bankKeeper.GetAllBalances(ctx, cdpAccount.GetAddress())

	var totalCollaterals types.TotalCollaterals

	for denom, collateralTypes := range denomCollateralTypes {
		// skip any denoms that do not match the requested collateral type
		if req.CollateralType != "" {
			match := false
			for _, ctype := range collateralTypes {
				if ctype == req.CollateralType {
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
			cdps := s.keeper.GetAllCdpsByCollateralType(ctx, collateralTypes[i])

			collateral := sdk.ZeroInt()

			for _, cdp := range cdps {
				collateral = collateral.Add(cdp.Collateral.Amount)
			}

			totalCollateral = totalCollateral.Sub(collateral)

			// if we have no collateralType filter, or the filter matches, include it in the response
			if req.CollateralType == "" || collateralTypes[i] == req.CollateralType {
				totalCollaterals = append(totalCollaterals, types.NewTotalCollateral(collateralTypes[i], sdk.NewCoin(denom, collateral)))
			}

			// skip the rest of the cdp queries if we have a matching filter
			if collateralTypes[i] == req.CollateralType {
				break
			}
		}

		if req.CollateralType == "" || collateralTypes[0] == req.CollateralType {
			// all leftover total collateral belongs to the first collateral type
			totalCollaterals = append(totalCollaterals, types.NewTotalCollateral(collateralTypes[0], sdk.NewCoin(denom, totalCollateral)))
		}
	}

	// sort to ensure deterministic response
	sort.Slice(totalCollaterals, func(i int, j int) bool {
		return totalCollaterals[i].CollateralType < totalCollaterals[j].CollateralType
	})

	return &types.QueryTotalCollateralResponse{
		TotalCollateral: totalCollaterals,
	}, nil
}

// Cdps queries all active CDPs.
func (s QueryServer) Cdps(c context.Context, req *types.QueryCdpsRequest) (*types.QueryCdpsResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	// Filter CDPs
	filteredCDPs, err := FilterCDPs(ctx, s.keeper, *req)
	if err != nil {
		status.Errorf(codes.InvalidArgument, "empty request")
	}

	return &types.QueryCdpsResponse{
		Cdps: filteredCDPs,
		// TODO: Use built in pagination and respond
		Pagination: nil,
	}, nil
}

// Cdp queries a CDP with the input owner address and collateral type.
func (s QueryServer) Cdp(c context.Context, req *types.QueryCdpRequest) (*types.QueryCdpResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	owner, err := sdk.AccAddressFromBech32(req.Owner)
	if err != nil {
		status.Errorf(codes.InvalidArgument, "invalid address")
	}

	_, valid := s.keeper.GetCollateralTypePrefix(ctx, req.CollateralType)
	if !valid {
		return nil, sdkerrors.Wrap(types.ErrInvalidCollateral, req.CollateralType)
	}

	cdp, found := s.keeper.GetCdpByOwnerAndCollateralType(ctx, owner, req.CollateralType)
	if !found {
		return nil, sdkerrors.Wrapf(types.ErrCdpNotFound, "owner %s, denom %s", req.Owner, req.CollateralType)
	}

	augmentedCDP := s.keeper.LoadAugmentedCDP(ctx, cdp)

	return &types.QueryCdpResponse{
		Cdp: augmentedCDP,
	}, nil
}

// Deposits queries deposits associated with the CDP owned by an address for a collateral type.
func (s QueryServer) Deposits(c context.Context, req *types.QueryDepositsRequest) (*types.QueryDepositsResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	owner, err := sdk.AccAddressFromBech32(req.Owner)
	if err != nil {
		status.Errorf(codes.InvalidArgument, "invalid address")
	}

	_, valid := s.keeper.GetCollateralTypePrefix(ctx, req.CollateralType)
	if !valid {
		return nil, sdkerrors.Wrap(types.ErrInvalidCollateral, req.CollateralType)
	}

	cdp, found := s.keeper.GetCdpByOwnerAndCollateralType(ctx, owner, req.CollateralType)
	if !found {
		return nil, sdkerrors.Wrapf(types.ErrCdpNotFound, "owner %s, denom %s", req.Owner, req.CollateralType)
	}

	deposits := s.keeper.GetDeposits(ctx, cdp.ID)

	return &types.QueryDepositsResponse{
		Deposits: deposits,
	}, nil
}

// CdpsByCollateralType queries all CDPs with the collateral type equal to the input collateral type.
func (s QueryServer) CdpsByCollateralType(c context.Context, req *types.QueryCdpsByCollateralTypeRequest) (*types.QueryCdpsByCollateralTypeResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	_, valid := s.keeper.GetCollateralTypePrefix(ctx, req.CollateralType)
	if !valid {
		return nil, sdkerrors.Wrap(types.ErrInvalidCollateral, req.CollateralType)
	}

	cdps := s.keeper.GetAllCdpsByCollateralType(ctx, req.CollateralType)
	// augment CDPs by adding collateral value and collateralization ratio
	var augmentedCDPs types.AugmentedCDPs
	for _, cdp := range cdps {
		augmentedCDP := s.keeper.LoadAugmentedCDP(ctx, cdp)
		augmentedCDPs = append(augmentedCDPs, augmentedCDP)
	}

	return &types.QueryCdpsByCollateralTypeResponse{
		Cdps: augmentedCDPs,
	}, nil
}

// CdpsByRatio queries all CDPs with the collateral type equal to the input
// colalteral type and collateralization ratio strictly less than the input
// ratio.
func (s QueryServer) CdpsByRatio(c context.Context, req *types.QueryCdpsByRatioRequest) (*types.QueryCdpsByRatioResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	_, valid := s.keeper.GetCollateralTypePrefix(ctx, req.CollateralType)
	if !valid {
		return nil, sdkerrors.Wrap(types.ErrInvalidCollateral, req.CollateralType)
	}

	ratio, err := s.keeper.CalculateCollateralizationRatioFromAbsoluteRatio(ctx, req.CollateralType, req.Ratio, "liquidation")
	if err != nil {
		return nil, sdkerrors.Wrap(err, "couldn't get collateralization ratio from absolute ratio")
	}

	cdps := s.keeper.GetAllCdpsByCollateralTypeAndRatio(ctx, req.CollateralType, ratio)
	// augment CDPs by adding collateral value and collateralization ratio
	var augmentedCDPs types.AugmentedCDPs
	for _, cdp := range cdps {
		augmentedCDP := s.keeper.LoadAugmentedCDP(ctx, cdp)
		augmentedCDPs = append(augmentedCDPs, augmentedCDP)
	}

	return &types.QueryCdpsByRatioResponse{
		Cdps: augmentedCDPs,
	}, nil
}

// FilterCDPs queries the store for all CDPs that match query req
func FilterCDPs(ctx sdk.Context, k Keeper, req types.QueryCdpsRequest) (types.AugmentedCDPs, error) {
	var matchCollateralType, matchOwner, matchID, matchRatio types.CDPs

	owner, err := sdk.AccAddressFromBech32(req.Owner)
	if err != nil {
		return nil, err
	}

	// match cdp owner (if supplied)
	if len(req.Owner) > 0 {
		denoms := k.GetCollateralTypes(ctx)
		for _, denom := range denoms {
			cdp, found := k.GetCdpByOwnerAndCollateralType(ctx, owner, denom)
			if found {
				matchOwner = append(matchOwner, cdp)
			}
		}
	}

	// match cdp collateral denom (if supplied)
	if len(req.CollateralType) > 0 {
		// if owner is specified only iterate over already matched cdps for efficiency
		if len(req.Owner) > 0 {
			for _, cdp := range matchOwner {
				if cdp.Type == req.CollateralType {
					matchCollateralType = append(matchCollateralType, cdp)
				}
			}
		} else {
			matchCollateralType = k.GetAllCdpsByCollateralType(ctx, req.CollateralType)
		}
	}

	// match cdp ID (if supplied)
	if req.ID != 0 {
		denoms := k.GetCollateralTypes(ctx)
		for _, denom := range denoms {
			cdp, found := k.GetCDP(ctx, denom, req.ID)
			if found {
				matchID = append(matchID, cdp)
			}
		}
	}

	// match cdp ratio (if supplied)
	if req.Ratio.GT(sdk.ZeroDec()) {
		denoms := k.GetCollateralTypes(ctx)
		for _, denom := range denoms {
			ratio, err := k.CalculateCollateralizationRatioFromAbsoluteRatio(ctx, denom, req.Ratio, "liquidation")
			if err != nil {
				continue
			}
			cdpsUnderRatio := k.GetAllCdpsByCollateralTypeAndRatio(ctx, denom, ratio)
			matchRatio = append(matchRatio, cdpsUnderRatio...)
		}
	}

	var commonCDPs types.CDPs
	// If no params specified, fetch all CDPs
	if len(req.CollateralType) == 0 && len(req.Owner) == 0 && req.ID == 0 && req.Ratio.Equal(sdk.ZeroDec()) {
		commonCDPs = k.GetAllCdps(ctx)
	}

	// Find the intersection of any matched CDPs
	if len(req.CollateralType) > 0 {
		if len(matchCollateralType) > 0 {
			commonCDPs = matchCollateralType
		} else {
			return types.AugmentedCDPs{}, nil
		}
	}

	if len(req.Owner) > 0 {
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
	if req.ID != 0 {
		if len(matchID) > 0 {
			if len(commonCDPs) > 0 {
				commonCDPs = FindIntersection(commonCDPs, matchID)
			} else {
				commonCDPs = matchID
			}
		} else {
			return types.AugmentedCDPs{}, nil
		}
	}
	if req.Ratio.GT(sdk.ZeroDec()) {
		if len(matchRatio) > 0 {
			if len(commonCDPs) > 0 {
				commonCDPs = FindIntersection(commonCDPs, matchRatio)
			} else {
				commonCDPs = matchRatio
			}
		} else {
			return types.AugmentedCDPs{}, nil
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

	// TODO: Use query.Paginate()? May be difficult to use special indexed keeper methods
	page, limit, err := query.ParsePagination(req.Pagination)
	if err != nil {
		return nil, err
	}

	start, end := client.Paginate(len(augmentedCDPs), page, limit, 100)
	if start < 0 || end < 0 {
		paginatedCDPs = types.AugmentedCDPs{}
	} else {
		paginatedCDPs = augmentedCDPs[start:end]
	}

	return paginatedCDPs, nil
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
