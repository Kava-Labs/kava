package keeper

import (
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/kava-labs/kava/x/cdp/types"
)

// NewQuerier returns a new querier function
func NewQuerier(keeper Keeper) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) (res []byte, err error) {
		switch path[0] {
		case types.QueryGetCdp:
			return queryGetCdp(ctx, req, keeper)
		case types.QueryGetCdps:
			return queryGetCdpsByDenom(ctx, req, keeper)
		case types.QueryGetCdpsByCollateralization:
			return queryGetCdpsByRatio(ctx, req, keeper)
		case types.QueryGetParams:
			return queryGetParams(ctx, req, keeper)
		case types.QueryGetCdpDeposits:
			return queryGetDeposits(ctx, req, keeper)
		case types.QueryTotalSupply:
			return queryGetTotalSupply(ctx, req, keeper)
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

	_, valid := keeper.GetDenomPrefix(ctx, requestParams.CollateralDenom)
	if !valid {
		return nil, sdkerrors.Wrap(types.ErrCollateralNotSupported, requestParams.CollateralDenom)
	}

	cdp, found := keeper.GetCdpByOwnerAndDenom(ctx, requestParams.Owner, requestParams.CollateralDenom)
	if !found {
		return nil, sdkerrors.Wrapf(types.ErrCdpNotFound, "owner %s, denom %s", requestParams.Owner, requestParams.CollateralDenom)
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

	_, valid := keeper.GetDenomPrefix(ctx, requestParams.CollateralDenom)
	if !valid {
		return nil, sdkerrors.Wrap(types.ErrCollateralNotSupported, requestParams.CollateralDenom)
	}

	cdp, found := keeper.GetCdpByOwnerAndDenom(ctx, requestParams.Owner, requestParams.CollateralDenom)
	if !found {
		return nil, sdkerrors.Wrapf(types.ErrCdpNotFound, "owner %s, denom %s", requestParams.Owner, requestParams.CollateralDenom)
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
	_, valid := keeper.GetDenomPrefix(ctx, requestParams.CollateralDenom)
	if !valid {
		return nil, sdkerrors.Wrap(types.ErrCollateralNotSupported, requestParams.CollateralDenom)
	}

	ratio, err := keeper.CalculateCollateralizationRatioFromAbsoluteRatio(ctx, requestParams.CollateralDenom, requestParams.Ratio, "liquidation")
	if err != nil {
		return nil, sdkerrors.Wrap(err, "couldn't get collateralization ratio from absolute ratio")
	}

	cdps := keeper.GetAllCdpsByDenomAndRatio(ctx, requestParams.CollateralDenom, ratio)
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
func queryGetCdpsByDenom(ctx sdk.Context, req abci.RequestQuery, keeper Keeper) ([]byte, error) {
	var requestParams types.QueryCdpsParams
	err := types.ModuleCdc.UnmarshalJSON(req.Data, &requestParams)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
	}
	_, valid := keeper.GetDenomPrefix(ctx, requestParams.CollateralDenom)
	if !valid {
		return nil, sdkerrors.Wrap(types.ErrCollateralNotSupported, requestParams.CollateralDenom)
	}

	cdps := keeper.GetAllCdpsByDenom(ctx, requestParams.CollateralDenom)
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

func queryGetTotalSupply(ctx sdk.Context, req abci.RequestQuery, keeper Keeper) ([]byte, error) {
	totalSupply := keeper.supplyKeeper.GetSupply(ctx).GetTotal().AmountOf("usdx")
	supplyInt := sdk.NewDecFromInt(totalSupply).Mul(sdk.MustNewDecFromStr("0.000001")).TruncateInt64()
	bz, err := types.ModuleCdc.MarshalJSON(supplyInt)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
	}
	return bz, nil
}

// query cdp module accounts
func queryGetAccounts(ctx sdk.Context, req abci.RequestQuery, keeper Keeper) ([]byte, error) {
	cdpAccAddress := keeper.supplyKeeper.GetModuleAddress(types.ModuleName)
	liquidatorAccAddress := keeper.supplyKeeper.GetModuleAddress(types.LiquidatorMacc)
	savingsRateAccAddress := keeper.supplyKeeper.GetModuleAddress(types.SavingsRateMacc)

	res := types.QueryGetAccountsResponse{
		Cdp:         cdpAccAddress,
		Liquidator:  liquidatorAccAddress,
		SavingsRate: savingsRateAccAddress,
	}

	// Encode results
	bz, err := codec.MarshalJSONIndent(types.ModuleCdc, res)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}
	return bz, nil
}
