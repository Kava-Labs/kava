package keeper

import (
	errorsmod "cosmossdk.io/errors"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/kava-labs/kava/x/kavadist/types"
)

// NewQuerier is the module level router for state queries
func NewQuerier(k Keeper, legacyQuerierCdc *codec.LegacyAmino) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) (res []byte, err error) {
		switch path[0] {
		case types.QueryGetParams:
			return queryGetParams(ctx, req, k, legacyQuerierCdc)
		case types.QueryGetBalance:
			return queryGetBalance(ctx, req, k, legacyQuerierCdc)
		default:
			return nil, errorsmod.Wrapf(sdkerrors.ErrUnknownRequest, "unknown %s query endpoint: %s", types.ModuleName, path[0])
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
		return nil, errorsmod.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}
	return bz, nil
}

// queryGetBalance returns current balance of kavadist module account
func queryGetBalance(ctx sdk.Context, req abci.RequestQuery, k Keeper, legacyQuerierCdc *codec.LegacyAmino) ([]byte, error) {
	acc := k.accountKeeper.GetModuleAccount(ctx, types.KavaDistMacc)
	balance := k.bankKeeper.GetAllBalances(ctx, acc.GetAddress())
	bz, err := codec.MarshalJSONIndent(legacyQuerierCdc, balance)
	if err != nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	return bz, nil
}
