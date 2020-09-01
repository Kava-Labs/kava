package keeper

import (
	abci "github.com/tendermint/tendermint/abci/types"

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
			return queryGetCdpsByCollateralType(ctx, req, keeper)
		case types.QueryGetCdpsByCollateralization:
			return queryGetCdpsByRatio(ctx, req, keeper)
		case types.QueryGetParams:
			return queryGetParams(ctx, req, keeper)
		case types.QueryGetCdpDeposits:
			return queryGetDeposits(ctx, req, keeper)
		case types.QueryGetAccounts:
			return queryGetAccounts(ctx, req, keeper)
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
		return nil, sdkerrors.Wrap(types.ErrCollateralNotSupported, requestParams.CollateralType)
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
		return nil, sdkerrors.Wrap(types.ErrCollateralNotSupported, requestParams.CollateralType)
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
		return nil, sdkerrors.Wrap(types.ErrCollateralNotSupported, requestParams.CollateralType)
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

// query all cdps with matching collateral denom
func queryGetCdpsByCollateralType(ctx sdk.Context, req abci.RequestQuery, keeper Keeper) ([]byte, error) {
	var requestParams types.QueryCdpsParams
	err := types.ModuleCdc.UnmarshalJSON(req.Data, &requestParams)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
	}
	_, valid := keeper.GetCollateralTypePrefix(ctx, requestParams.CollateralType)
	if !valid {
		return nil, sdkerrors.Wrap(types.ErrCollateralNotSupported, requestParams.CollateralType)
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
