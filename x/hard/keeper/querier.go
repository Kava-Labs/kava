package keeper

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	supplyexported "github.com/cosmos/cosmos-sdk/x/supply/exported"

	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/kava-labs/kava/x/hard/types"
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
		case types.QueryGetBorrows:
			return queryGetBorrows(ctx, req, k)
		case types.QueryGetBorrow:
			return queryGetBorrow(ctx, req, k)
		case types.QueryGetBorrowed:
			return queryGetBorrowed(ctx, req, k)
		default:
			return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unknown %s query endpoint", types.ModuleName)
		}
	}
}

func queryGetParams(ctx sdk.Context, _ abci.RequestQuery, k Keeper) ([]byte, error) {
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
		acc := k.supplyKeeper.GetModuleAccount(ctx, params.Name)
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

	var deposits []types.Deposit
	switch {
	case depositDenom && owner:
		deposit, found := k.GetDeposit(ctx, params.Owner)
		if found {
			for _, depCoin := range deposit.Amount {
				if depCoin.Denom == params.DepositDenom {
					deposits = append(deposits, deposit)
				}
			}

		}
	case depositDenom:
		k.IterateDeposits(ctx, func(deposit types.Deposit) (stop bool) {
			if deposit.Amount.AmountOf(params.DepositDenom).IsPositive() {
				deposits = append(deposits, deposit)
			}
			return false
		})
	case owner:
		deposit, found := k.GetDeposit(ctx, params.Owner)
		if found {
			deposits = append(deposits, deposit)
		}
	default:
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

func queryGetBorrows(ctx sdk.Context, req abci.RequestQuery, k Keeper) ([]byte, error) {

	var params types.QueryBorrowsParams
	err := types.ModuleCdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
	}

	// TODO: filter query results
	// depositDenom := len(params.BorrowDenom) > 0
	// owner := len(params.Owner) > 0

	var borrows []types.Borrow
	k.IterateBorrows(ctx, func(borrow types.Borrow) (stop bool) {
		borrows = append(borrows, borrow)
		return false
	})

	start, end := client.Paginate(len(borrows), params.Page, params.Limit, 100)
	if start < 0 || end < 0 {
		borrows = []types.Borrow{}
	} else {
		borrows = borrows[start:end]
	}

	bz, err := codec.MarshalJSONIndent(types.ModuleCdc, borrows)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	return bz, nil
}

func queryGetBorrowed(ctx sdk.Context, req abci.RequestQuery, k Keeper) ([]byte, error) {
	var params types.QueryBorrowedParams
	err := types.ModuleCdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
	}

	borrowedCoins, found := k.GetBorrowedCoins(ctx)
	if !found {
		return nil, types.ErrBorrowedCoinsNotFound
	}

	// If user specified a denom only return coins of that denom type
	if len(params.Denom) > 0 {
		borrowedCoins = sdk.NewCoins(sdk.NewCoin(params.Denom, borrowedCoins.AmountOf(params.Denom)))
	}

	bz, err := codec.MarshalJSONIndent(types.ModuleCdc, borrowedCoins)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	return bz, nil
}

func queryGetBorrow(ctx sdk.Context, req abci.RequestQuery, k Keeper) ([]byte, error) {
	var params types.QueryBorrowsParams
	err := types.ModuleCdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
	}

	var borrowBalance sdk.Coins
	if len(params.Owner) > 0 {
		borrowBalance = k.GetBorrowBalance(ctx, params.Owner)
	}

	bz, err := codec.MarshalJSONIndent(types.ModuleCdc, borrowBalance)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	return bz, nil
}
