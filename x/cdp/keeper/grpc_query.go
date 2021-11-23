package keeper

import (
	"context"
	"sort"

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
	filteredCDPs, err := GrpcFilterCDPs(ctx, s.keeper, *req)
	if err != nil {
		return nil, err
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
		return nil, status.Errorf(codes.InvalidArgument, "invalid address")
	}

	_, valid := s.keeper.GetCollateral(ctx, req.CollateralType)
	if !valid {
		return nil, sdkerrors.Wrap(types.ErrInvalidCollateral, req.CollateralType)
	}

	cdp, found := s.keeper.GetCdpByOwnerAndCollateralType(ctx, owner, req.CollateralType)
	if !found {
		return nil, sdkerrors.Wrapf(types.ErrCdpNotFound, "owner %s, denom %s", req.Owner, req.CollateralType)
	}

	cdpResponse := s.keeper.LoadCDPResponse(ctx, cdp)

	return &types.QueryCdpResponse{
		Cdp: cdpResponse,
	}, nil
}

// Deposits queries deposits associated with the CDP owned by an address for a collateral type.
func (s QueryServer) Deposits(c context.Context, req *types.QueryDepositsRequest) (*types.QueryDepositsResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	owner, err := sdk.AccAddressFromBech32(req.Owner)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid address")
	}

	_, valid := s.keeper.GetCollateral(ctx, req.CollateralType)
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

// FilterCDPs queries the store for all CDPs that match query req
func GrpcFilterCDPs(ctx sdk.Context, k Keeper, req types.QueryCdpsRequest) (types.CDPResponses, error) {
	// TODO: Ideally use query.Paginate() here over existing FilterCDPs. However
	// This is difficult to use different CDP indices and specific keeper
	// methods without iterating over all CDPs.
	page, limit, err := query.ParsePagination(req.Pagination)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	// Owner address is optional, only parse if it's provided otherwise it will
	// respond with an error
	var owner sdk.AccAddress
	if req.Owner != "" {
		owner, err = sdk.AccAddressFromBech32(req.Owner)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid owner address")
		}
	}

	ratio := sdk.ZeroDec()

	if req.Ratio != "" {
		ratio, err = sdk.NewDecFromStr(req.Ratio)
		if err != nil {
			if err != nil {
				return nil, status.Errorf(codes.InvalidArgument, "invalid ratio")
			}
		}
	}

	legacyParams := types.NewQueryCdpsParams(page, limit, req.CollateralType, owner, req.ID, ratio)

	cdps, err := FilterCDPs(ctx, k, legacyParams)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	var cdpResponses types.CDPResponses
	for _, cdp := range cdps {
		cdpResponse := types.CDPResponse{
			ID:                     cdp.ID,
			Owner:                  cdp.Owner.String(),
			Type:                   cdp.Type,
			Collateral:             cdp.Collateral,
			Principal:              cdp.Principal,
			AccumulatedFees:        cdp.AccumulatedFees,
			FeesUpdated:            cdp.FeesUpdated,
			InterestFactor:         cdp.InterestFactor.String(),
			CollateralValue:        cdp.CollateralValue,
			CollateralizationRatio: cdp.CollateralizationRatio.String(),
		}
		cdpResponses = append(cdpResponses, cdpResponse)
	}

	return cdpResponses, nil
}
