package keeper

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	supplyexported "github.com/cosmos/cosmos-sdk/x/supply/exported"

	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/kava-labs/kava/x/harvest/types"
)

// NewQuerier is the module level router for state queries
func NewQuerier(k Keeper) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) (res []byte, err error) {
		switch path[0] {
		case types.QueryGetParams:
			return queryGetParams(ctx, req, k)
		case types.QueryGetModuleAccounts:
			return queryGetModAccounts(ctx, req, k)
		case types.QueryGetDeposits:
			return queryGetDeposits(ctx, req, k)
		case types.QueryGetClaims:
			return queryGetClaims(ctx, req, k)
		default:
			return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unknown %s query endpoint", types.ModuleName)
		}
	}
}

func queryGetParams(ctx sdk.Context, req abci.RequestQuery, k Keeper) ([]byte, error) {
	// Get params
	params := k.GetParams(ctx)

	// Encode results
	bz, err := codec.MarshalJSONIndent(k.cdc, params)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}
	return bz, nil
}

func queryGetModAccounts(ctx sdk.Context, req abci.RequestQuery, k Keeper) ([]byte, error) {

	var params types.QueryAccountParams
	err := types.ModuleCdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
	}
	var accs []supplyexported.ModuleAccountI
	if len(params.Name) > 0 {
		acc := k.supplyKeeper.GetModuleAccount(ctx, types.LPAccount)
		accs = append(accs, acc)
	} else {
		acc := k.supplyKeeper.GetModuleAccount(ctx, types.ModuleAccountName)
		accs = append(accs, acc)
		acc = k.supplyKeeper.GetModuleAccount(ctx, types.LPAccount)
		accs = append(accs, acc)
		acc = k.supplyKeeper.GetModuleAccount(ctx, types.DelegatorAccount)
		accs = append(accs, acc)
	}

	bz, err := codec.MarshalJSONIndent(k.cdc, accs)

	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	return bz, nil
}

func queryGetDeposits(ctx sdk.Context, req abci.RequestQuery, k Keeper) ([]byte, error) {

	var params types.QueryDepositParams
	err := types.ModuleCdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
	}
	depositDenom := len(params.DepositDenom) > 0
	owner := len(params.Owner) > 0
	depositType := len(params.DepositType) > 0

	var deposits []types.Deposit
	if depositDenom && owner && depositType {
		deposit, found := k.GetDeposit(ctx, params.Owner, params.DepositDenom, params.DepositType)
		if found {
			deposits = append(deposits, deposit)
		}
	} else if depositDenom && owner {
		for _, dt := range types.DepositTypesDepositQuery {
			deposit, found := k.GetDeposit(ctx, params.Owner, params.DepositDenom, dt)
			if found {
				deposits = append(deposits, deposit)
			}
		}
	} else if depositDenom && depositType {
		k.IterateDepositsByTypeAndDenom(ctx, params.DepositType, params.DepositDenom, func(deposit types.Deposit) (stop bool) {
			deposits = append(deposits, deposit)
			return false
		})
	} else if owner && depositType {
		schedules := k.GetParams(ctx).LiquidityProviderSchedules
		for _, lps := range schedules {
			deposit, found := k.GetDeposit(ctx, params.Owner, lps.DepositDenom, params.DepositType)
			if found {
				deposits = append(deposits, deposit)
			}
		}
	} else if depositDenom {
		for _, dt := range types.DepositTypesDepositQuery {
			k.IterateDepositsByTypeAndDenom(ctx, dt, params.DepositDenom, func(deposit types.Deposit) (stop bool) {
				deposits = append(deposits, deposit)
				return false
			})
		}
	} else if owner {
		schedules := k.GetParams(ctx).LiquidityProviderSchedules
		for _, lps := range schedules {
			for _, dt := range types.DepositTypesDepositQuery {
				deposit, found := k.GetDeposit(ctx, params.Owner, lps.DepositDenom, dt)
				if found {
					deposits = append(deposits, deposit)
				}
			}
		}
	} else if depositType {
		schedules := k.GetParams(ctx).LiquidityProviderSchedules
		for _, lps := range schedules {
			k.IterateDepositsByTypeAndDenom(ctx, params.DepositType, lps.DepositDenom, func(deposit types.Deposit) (stop bool) {
				deposits = append(deposits, deposit)
				return false
			})
		}
	} else {
		k.IterateDeposits(ctx, func(deposit types.Deposit) (stop bool) {
			deposits = append(deposits, deposit)
			return false
		})
	}

	start, end := client.Paginate(len(deposits), params.Page, params.Limit, 100)
	if start < 0 || end < 0 {
		deposits = []types.Deposit{}
	} else {
		deposits = deposits[start:end]
	}

	bz, err := codec.MarshalJSONIndent(types.ModuleCdc, deposits)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	return bz, nil
}

func queryGetClaims(ctx sdk.Context, req abci.RequestQuery, k Keeper) ([]byte, error) {

	var params types.QueryClaimParams
	err := types.ModuleCdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
	}
	depositDenom := false
	owner := false
	depositType := false

	if len(params.DepositDenom) > 0 {
		depositDenom = true
	}
	if len(params.Owner) > 0 {
		owner = true
	}
	if len(params.DepositType) > 0 {
		depositType = true
	}

	var claims []types.Claim
	if depositDenom && owner && depositType {
		claim, found := k.GetClaim(ctx, params.Owner, params.DepositDenom, params.DepositType)
		if found {
			claims = append(claims, claim)
		}
	} else if depositDenom && owner {
		for _, dt := range types.DepositTypesClaimQuery {
			claim, found := k.GetClaim(ctx, params.Owner, params.DepositDenom, dt)
			if found {
				claims = append(claims, claim)
			}
		}
	} else if depositDenom && depositType {
		k.IterateClaimsByTypeAndDenom(ctx, params.DepositType, params.DepositDenom, func(claim types.Claim) (stop bool) {
			claims = append(claims, claim)
			return false
		})
	} else if owner && depositType {
		harvestParams := k.GetParams(ctx)
		for _, lps := range harvestParams.LiquidityProviderSchedules {
			claim, found := k.GetClaim(ctx, params.Owner, lps.DepositDenom, params.DepositType)
			if found {
				claims = append(claims, claim)
			}
		}
		for _, dss := range harvestParams.DelegatorDistributionSchedules {
			claim, found := k.GetClaim(ctx, params.Owner, dss.DistributionSchedule.DepositDenom, params.DepositType)
			if found {
				claims = append(claims, claim)
			}
		}
	} else if depositDenom {
		for _, dt := range types.DepositTypesClaimQuery {
			k.IterateClaimsByTypeAndDenom(ctx, dt, params.DepositDenom, func(claim types.Claim) (stop bool) {
				claims = append(claims, claim)
				return false
			})
		}
	} else if owner {
		harvestParams := k.GetParams(ctx)
		for _, lps := range harvestParams.LiquidityProviderSchedules {
			claim, found := k.GetClaim(ctx, params.Owner, lps.DepositDenom, types.LP)
			if found {
				claims = append(claims, claim)
			}
		}
		for _, dds := range harvestParams.DelegatorDistributionSchedules {
			claim, found := k.GetClaim(ctx, params.Owner, dds.DistributionSchedule.DepositDenom, types.Stake)
			if found {
				claims = append(claims, claim)
			}
		}
	} else if depositType {
		harvestParams := k.GetParams(ctx)
		for _, lps := range harvestParams.LiquidityProviderSchedules {
			k.IterateClaimsByTypeAndDenom(ctx, params.DepositType, lps.DepositDenom, func(claim types.Claim) (stop bool) {
				claims = append(claims, claim)
				return false
			})
		}
		for _, dds := range harvestParams.DelegatorDistributionSchedules {
			k.IterateClaimsByTypeAndDenom(ctx, params.DepositType, dds.DistributionSchedule.DepositDenom, func(claim types.Claim) (stop bool) {
				claims = append(claims, claim)
				return false
			})
		}
	} else {
		k.IterateClaims(ctx, func(claim types.Claim) (stop bool) {
			claims = append(claims, claim)
			return false
		})
	}

	start, end := client.Paginate(len(claims), params.Page, params.Limit, 100)
	if start < 0 || end < 0 {
		claims = []types.Claim{}
	} else {
		claims = claims[start:end]
	}

	bz, err := codec.MarshalJSONIndent(types.ModuleCdc, claims)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	return bz, nil
}
