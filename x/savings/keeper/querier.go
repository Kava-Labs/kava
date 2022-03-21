package keeper

import (
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/kava-labs/kava/x/savings/types"
)

// NewQuerier is the module level router for state queries
func NewQuerier(keeper Keeper, legacyQuerierCdc *codec.LegacyAmino) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) (res []byte, err error) {
		switch path[0] {
		case types.QueryGetParams:
			return queryGetParams(ctx, req, keeper, legacyQuerierCdc)
		case types.QueryGetDeposits:
			return queryGetDeposits(ctx, req, keeper, legacyQuerierCdc)
		default:
			return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unknown %s query endpoint", types.ModuleName)
		}
	}

}

// query params in the pricefeed store
func queryGetParams(ctx sdk.Context, req abci.RequestQuery, keeper Keeper, legacyQuerierCdc *codec.LegacyAmino) ([]byte, error) {
	params := keeper.GetParams(ctx)

	// Encode results
	bz, err := codec.MarshalJSONIndent(legacyQuerierCdc, params)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}
	return bz, nil
}

func queryGetDeposits(ctx sdk.Context, req abci.RequestQuery, k Keeper, legacyQuerierCdc *codec.LegacyAmino) ([]byte, error) {

	var params types.QueryDepositsParams
	err := legacyQuerierCdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
	}
	denom := len(params.Denom) > 0
	owner := len(params.Owner) > 0

	var deposits types.Deposits
	switch {
	case owner && denom:
		deposit, found := k.GetDeposit(ctx, params.Owner)
		if found {
			for _, coin := range deposit.Amount {
				if coin.Denom == params.Denom {
					deposits = append(deposits, deposit)
				}
			}
		}
	case owner:
		deposit, found := k.GetDeposit(ctx, params.Owner)
		if found {
			deposits = append(deposits, deposit)
		}
	case denom:
		k.IterateDeposits(ctx, func(deposit types.Deposit) (stop bool) {
			if deposit.Amount.AmountOf(params.Denom).IsPositive() {
				deposits = append(deposits, deposit)
			}
			return false
		})
	default:
		k.IterateDeposits(ctx, func(deposit types.Deposit) (stop bool) {
			deposits = append(deposits, deposit)
			return false
		})
	}

	var bz []byte

	start, end := client.Paginate(len(deposits), params.Page, params.Limit, 100)
	if start < 0 || end < 0 {
		deposits = types.Deposits{}
	} else {
		deposits = deposits[start:end]
	}

	bz, err = codec.MarshalJSONIndent(legacyQuerierCdc, deposits)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}
	return bz, nil
}
