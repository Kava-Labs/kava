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
		case types.QueryGetTotalDeposited:
			return queryGetTotalDeposited(ctx, req, k)
		case types.QueryGetBorrows:
			return queryGetBorrows(ctx, req, k)
		case types.QueryGetTotalBorrowed:
			return queryGetTotalBorrowed(ctx, req, k)
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

	var params types.QueryDepositsParams
	err := types.ModuleCdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
	}
	denom := len(params.Denom) > 0
	owner := len(params.Owner) > 0

	var deposits types.Deposits
	switch {
	case owner && denom:
		deposit, found := k.GetSyncedDeposit(ctx, params.Owner)
		if found {
			for _, coin := range deposit.Amount {
				if coin.Denom == params.Denom {
					deposits = append(deposits, deposit)
				}
			}
		}
	case owner:
		deposit, found := k.GetSyncedDeposit(ctx, params.Owner)
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

	// If owner param was specified then deposits array already contains the user's synced deposit
	if owner {
		bz, err = codec.MarshalJSONIndent(types.ModuleCdc, deposits)
		if err != nil {
			return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
		}
		return bz, nil
	}

	// Otherwise we need to simulate syncing of each deposit
	var syncedDeposits types.Deposits
	for _, deposit := range deposits {
		syncedDeposit, _ := k.GetSyncedDeposit(ctx, deposit.Depositor)
		syncedDeposits = append(syncedDeposits, syncedDeposit)
	}

	start, end := client.Paginate(len(syncedDeposits), params.Page, params.Limit, 100)
	if start < 0 || end < 0 {
		syncedDeposits = types.Deposits{}
	} else {
		syncedDeposits = syncedDeposits[start:end]
	}

	bz, err = codec.MarshalJSONIndent(types.ModuleCdc, syncedDeposits)
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
	denom := len(params.Denom) > 0
	owner := len(params.Owner) > 0

	var borrows types.Borrows
	switch {
	case owner && denom:
		borrow, found := k.GetSyncedBorrow(ctx, params.Owner)
		if found {
			for _, coin := range borrow.Amount {
				if coin.Denom == params.Denom {
					borrows = append(borrows, borrow)
				}
			}
		}
	case owner:
		borrow, found := k.GetSyncedBorrow(ctx, params.Owner)
		if found {
			borrows = append(borrows, borrow)
		}
	case denom:
		k.IterateBorrows(ctx, func(borrow types.Borrow) (stop bool) {
			if borrow.Amount.AmountOf(params.Denom).IsPositive() {
				borrows = append(borrows, borrow)
			}
			return false
		})
	default:
		k.IterateBorrows(ctx, func(borrow types.Borrow) (stop bool) {
			borrows = append(borrows, borrow)
			return false
		})
	}

	var bz []byte

	// If owner param was specified then borrows array already contains the user's synced borrow
	if owner {
		bz, err = codec.MarshalJSONIndent(types.ModuleCdc, borrows)
		if err != nil {
			return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
		}
		return bz, nil
	}

	// Otherwise we need to simulate syncing of each borrow
	var syncedBorrows types.Borrows
	for _, borrow := range borrows {
		syncedBorrow, _ := k.GetSyncedBorrow(ctx, borrow.Borrower)
		syncedBorrows = append(syncedBorrows, syncedBorrow)
	}

	start, end := client.Paginate(len(syncedBorrows), params.Page, params.Limit, 100)
	if start < 0 || end < 0 {
		syncedBorrows = types.Borrows{}
	} else {
		syncedBorrows = syncedBorrows[start:end]
	}

	bz, err = codec.MarshalJSONIndent(types.ModuleCdc, syncedBorrows)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}
	return bz, nil
}

func queryGetTotalBorrowed(ctx sdk.Context, req abci.RequestQuery, k Keeper) ([]byte, error) {
	var params types.QueryTotalBorrowedParams
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

func queryGetTotalDeposited(ctx sdk.Context, req abci.RequestQuery, k Keeper) ([]byte, error) {
	var params types.QueryTotalDepositedParams
	err := types.ModuleCdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
	}

	suppliedCoins, found := k.GetSuppliedCoins(ctx)
	if !found {
		return nil, types.ErrSuppliedCoinsNotFound
	}

	// If user specified a denom only return coins of that denom type
	if len(params.Denom) > 0 {
		suppliedCoins = sdk.NewCoins(sdk.NewCoin(params.Denom, suppliedCoins.AmountOf(params.Denom)))
	}

	bz, err := codec.MarshalJSONIndent(types.ModuleCdc, suppliedCoins)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	return bz, nil
}
