package keeper

import (
	"context"

	"github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/query"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/kava-labs/kava/x/hard/types"
)

type QueryServer struct {
	keeper Keeper
}

// NewQueryServer returns an implementation of the hard MsgServer interface
// for the provided Keeper.
func NewQueryServerImpl(keeper Keeper) types.QueryServer {
	return &QueryServer{keeper: keeper}
}

var _ types.QueryServer = QueryServer{}

func (qs QueryServer) Params(ctx context.Context, req *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	if req == nil {
		return nil, status.Errorf(codes.InvalidArgument, "empty request")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)

	// Get params
	params := qs.keeper.GetParams(sdkCtx)

	return &types.QueryParamsResponse{
		Params: params,
	}, nil
}

func (qs QueryServer) Accounts(ctx context.Context, req *types.QueryAccountsRequest) (*types.QueryAccountsResponse, error) {
	if req == nil {
		return nil, status.Errorf(codes.InvalidArgument, "empty request")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)

	var accs []authtypes.ModuleAccount
	if len(req.Name) > 0 {
		acc := qs.keeper.accountKeeper.GetModuleAccount(sdkCtx, req.Name)
		if acc == nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid account name")
		}

		accs = append(accs, *acc.(*authtypes.ModuleAccount))
	} else {
		acc := qs.keeper.accountKeeper.GetModuleAccount(sdkCtx, types.ModuleAccountName)
		accs = append(accs, *acc.(*authtypes.ModuleAccount))
	}

	return &types.QueryAccountsResponse{
		Accounts: accs,
	}, nil
}

func (qs QueryServer) Deposits(ctx context.Context, req *types.QueryDepositsRequest) (*types.QueryDepositsResponse, error) {
	if req == nil {
		return nil, status.Errorf(codes.InvalidArgument, "empty request")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)

	hasDenom := len(req.Denom) > 0
	hasOwner := len(req.Owner) > 0

	var owner sdk.AccAddress
	var err error
	if hasOwner {
		owner, err = sdk.AccAddressFromBech32(req.Owner)
		if err != nil {
			return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, err.Error())
		}
	}

	var deposits types.Deposits
	switch {
	case hasOwner && hasDenom:
		deposit, found := qs.keeper.GetSyncedDeposit(sdkCtx, owner)
		if found {
			for _, coin := range deposit.Amount {
				if coin.Denom == req.Denom {
					deposits = append(deposits, deposit)
				}
			}
		}
	case hasOwner:
		deposit, found := qs.keeper.GetSyncedDeposit(sdkCtx, owner)
		if found {
			deposits = append(deposits, deposit)
		}
	case hasDenom:
		qs.keeper.IterateDeposits(sdkCtx, func(deposit types.Deposit) (stop bool) {
			if deposit.Amount.AmountOf(req.Denom).IsPositive() {
				deposits = append(deposits, deposit)
			}
			return false
		})
	default:
		qs.keeper.IterateDeposits(sdkCtx, func(deposit types.Deposit) (stop bool) {
			deposits = append(deposits, deposit)
			return false
		})
	}

	// If owner param was specified then deposits array already contains the user's synced deposit
	if hasOwner {
		return &types.QueryDepositsResponse{
			Deposits:   deposits.ToResponse(),
			Pagination: nil,
		}, nil
	}

	// Otherwise we need to simulate syncing of each deposit
	var syncedDeposits types.Deposits
	for _, deposit := range deposits {
		syncedDeposit, _ := qs.keeper.GetSyncedDeposit(sdkCtx, deposit.Depositor)
		syncedDeposits = append(syncedDeposits, syncedDeposit)
	}

	// TODO: Use more optimal FilteredPaginate to directly iterate over the store
	// and not fetch everything. This currently also ignores certain fields in
	// the pagination request like Key, CountTotal, Reverse.
	page, limit, err := query.ParsePagination(req.Pagination)
	if err != nil {
		return nil, err
	}

	start, end := client.Paginate(len(syncedDeposits), page, limit, 100)
	if start < 0 || end < 0 {
		syncedDeposits = types.Deposits{}
	} else {
		syncedDeposits = syncedDeposits[start:end]
	}

	return &types.QueryDepositsResponse{
		Deposits:   syncedDeposits.ToResponse(),
		Pagination: nil,
	}, nil
}

func (qs QueryServer) UnsyncedDeposits(ctx context.Context, req *types.QueryUnsyncedDepositsRequest) (*types.QueryUnsyncedDepositsResponse, error) {
	if req == nil {
		return nil, status.Errorf(codes.InvalidArgument, "empty request")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)

	hasDenom := len(req.Denom) > 0
	hasOwner := len(req.Owner) > 0

	var owner sdk.AccAddress
	var err error
	if hasOwner {
		owner, err = sdk.AccAddressFromBech32(req.Owner)
		if err != nil {
			return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, err.Error())
		}
	}

	var deposits types.Deposits
	switch {
	case hasOwner && hasDenom:
		deposit, found := qs.keeper.GetDeposit(sdkCtx, owner)
		if found {
			for _, coin := range deposit.Amount {
				if coin.Denom == req.Denom {
					deposits = append(deposits, deposit)
				}
			}
		}
	case hasOwner:
		deposit, found := qs.keeper.GetDeposit(sdkCtx, owner)
		if found {
			deposits = append(deposits, deposit)
		}
	case hasDenom:
		qs.keeper.IterateDeposits(sdkCtx, func(deposit types.Deposit) (stop bool) {
			if deposit.Amount.AmountOf(req.Denom).IsPositive() {
				deposits = append(deposits, deposit)
			}
			return false
		})
	default:
		qs.keeper.IterateDeposits(sdkCtx, func(deposit types.Deposit) (stop bool) {
			deposits = append(deposits, deposit)
			return false
		})
	}

	page, limit, err := query.ParsePagination(req.Pagination)
	if err != nil {
		return nil, err
	}

	start, end := client.Paginate(len(deposits), page, limit, 100)
	if start < 0 || end < 0 {
		deposits = types.Deposits{}
	} else {
		deposits = deposits[start:end]
	}

	return &types.QueryUnsyncedDepositsResponse{
		Deposits:   deposits.ToResponse(),
		Pagination: nil,
	}, nil
}

func (qs QueryServer) Borrows(ctx context.Context, req *types.QueryBorrowsRequest) (*types.QueryBorrowsResponse, error) {
	if req == nil {
		return nil, status.Errorf(codes.InvalidArgument, "empty request")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)

	hasDenom := len(req.Denom) > 0
	hasOwner := len(req.Owner) > 0

	var owner sdk.AccAddress
	var err error
	if hasOwner {
		owner, err = sdk.AccAddressFromBech32(req.Owner)
		if err != nil {
			return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, err.Error())
		}
	}

	var borrows types.Borrows
	switch {
	case hasOwner && hasDenom:
		borrow, found := qs.keeper.GetSyncedBorrow(sdkCtx, owner)
		if found {
			for _, coin := range borrow.Amount {
				if coin.Denom == req.Denom {
					borrows = append(borrows, borrow)
				}
			}
		}
	case hasOwner:
		borrow, found := qs.keeper.GetSyncedBorrow(sdkCtx, owner)
		if found {
			borrows = append(borrows, borrow)
		}
	case hasDenom:
		qs.keeper.IterateBorrows(sdkCtx, func(borrow types.Borrow) (stop bool) {
			if borrow.Amount.AmountOf(req.Denom).IsPositive() {
				borrows = append(borrows, borrow)
			}
			return false
		})
	default:
		qs.keeper.IterateBorrows(sdkCtx, func(borrow types.Borrow) (stop bool) {
			borrows = append(borrows, borrow)
			return false
		})
	}

	// If owner param was specified then borrows array already contains the user's synced borrow
	if hasOwner {
		return &types.QueryBorrowsResponse{
			Borrows:    borrows.ToResponse(),
			Pagination: nil,
		}, nil
	}

	// Otherwise we need to simulate syncing of each borrow
	var syncedBorrows types.Borrows
	for _, borrow := range borrows {
		syncedBorrow, _ := qs.keeper.GetSyncedBorrow(sdkCtx, borrow.Borrower)
		syncedBorrows = append(syncedBorrows, syncedBorrow)
	}

	page, limit, err := query.ParsePagination(req.Pagination)
	if err != nil {
		return nil, err
	}

	start, end := client.Paginate(len(syncedBorrows), page, limit, 100)
	if start < 0 || end < 0 {
		syncedBorrows = types.Borrows{}
	} else {
		syncedBorrows = syncedBorrows[start:end]
	}

	return &types.QueryBorrowsResponse{
		Borrows: syncedBorrows.ToResponse(),
	}, nil
}

func (qs QueryServer) UnsyncedBorrows(ctx context.Context, req *types.QueryUnsyncedBorrowsRequest) (*types.QueryUnsyncedBorrowsResponse, error) {
	if req == nil {
		return nil, status.Errorf(codes.InvalidArgument, "empty request")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)

	hasDenom := len(req.Denom) > 0
	hasOwner := len(req.Owner) > 0

	var owner sdk.AccAddress
	var err error
	if hasOwner {
		owner, err = sdk.AccAddressFromBech32(req.Owner)
		if err != nil {
			return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, err.Error())
		}
	}

	var borrows types.Borrows
	switch {
	case hasOwner && hasDenom:
		borrow, found := qs.keeper.GetBorrow(sdkCtx, owner)
		if found {
			for _, coin := range borrow.Amount {
				if coin.Denom == req.Denom {
					borrows = append(borrows, borrow)
				}
			}
		}
	case hasOwner:
		borrow, found := qs.keeper.GetBorrow(sdkCtx, owner)
		if found {
			borrows = append(borrows, borrow)
		}
	case hasDenom:
		qs.keeper.IterateBorrows(sdkCtx, func(borrow types.Borrow) (stop bool) {
			if borrow.Amount.AmountOf(req.Denom).IsPositive() {
				borrows = append(borrows, borrow)
			}
			return false
		})
	default:
		qs.keeper.IterateBorrows(sdkCtx, func(borrow types.Borrow) (stop bool) {
			borrows = append(borrows, borrow)
			return false
		})
	}

	page, limit, err := query.ParsePagination(req.Pagination)
	if err != nil {
		return nil, err
	}

	start, end := client.Paginate(len(borrows), page, limit, 100)
	if start < 0 || end < 0 {
		borrows = types.Borrows{}
	} else {
		borrows = borrows[start:end]
	}

	return &types.QueryUnsyncedBorrowsResponse{
		Borrows:    borrows.ToResponse(),
		Pagination: nil,
	}, nil
}

func (qs QueryServer) TotalBorrowed(ctx context.Context, req *types.QueryTotalBorrowedRequest) (*types.QueryTotalBorrowedResponse, error) {
	if req == nil {
		return nil, status.Errorf(codes.InvalidArgument, "empty request")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)

	borrowedCoins, found := qs.keeper.GetBorrowedCoins(sdkCtx)
	if !found {
		return nil, types.ErrBorrowedCoinsNotFound
	}

	// If user specified a denom only return coins of that denom type
	if len(req.Denom) > 0 {
		borrowedCoins = sdk.NewCoins(sdk.NewCoin(req.Denom, borrowedCoins.AmountOf(req.Denom)))
	}

	return &types.QueryTotalBorrowedResponse{
		BorrowedCoins: borrowedCoins,
	}, nil
}

func (qs QueryServer) TotalDeposited(ctx context.Context, req *types.QueryTotalDepositedRequest) (*types.QueryTotalDepositedResponse, error) {
	if req == nil {
		return nil, status.Errorf(codes.InvalidArgument, "empty request")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)

	suppliedCoins, found := qs.keeper.GetSuppliedCoins(sdkCtx)
	if !found {
		return nil, types.ErrSuppliedCoinsNotFound
	}

	// If user specified a denom only return coins of that denom type
	if len(req.Denom) > 0 {
		suppliedCoins = sdk.NewCoins(sdk.NewCoin(req.Denom, suppliedCoins.AmountOf(req.Denom)))
	}

	return &types.QueryTotalDepositedResponse{
		SuppliedCoins: suppliedCoins,
	}, nil
}

func (qs QueryServer) InterestRate(ctx context.Context, req *types.QueryInterestRateRequest) (*types.QueryInterestRateResponse, error) {
	if req == nil {
		return nil, status.Errorf(codes.InvalidArgument, "empty request")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)

	var moneyMarketInterestRates types.MoneyMarketInterestRates
	var moneyMarkets types.MoneyMarkets
	if len(req.Denom) > 0 {
		moneyMarket, found := qs.keeper.GetMoneyMarket(sdkCtx, req.Denom)
		if !found {
			return nil, types.ErrMoneyMarketNotFound
		}
		moneyMarkets = append(moneyMarkets, moneyMarket)
	} else {
		moneyMarkets = qs.keeper.GetAllMoneyMarkets(sdkCtx)
	}

	// Calculate the borrow and supply APY interest rates for each money market
	for _, moneyMarket := range moneyMarkets {
		denom := moneyMarket.Denom
		macc := qs.keeper.accountKeeper.GetModuleAccount(sdkCtx, types.ModuleName)
		cash := qs.keeper.bankKeeper.GetBalance(sdkCtx, macc.GetAddress(), denom).Amount

		borrowed := sdk.NewCoin(denom, sdk.ZeroInt())
		borrowedCoins, foundBorrowedCoins := qs.keeper.GetBorrowedCoins(sdkCtx)
		if foundBorrowedCoins {
			borrowed = sdk.NewCoin(denom, borrowedCoins.AmountOf(denom))
		}

		reserves, foundReserves := qs.keeper.GetTotalReserves(sdkCtx)
		if !foundReserves {
			reserves = sdk.NewCoins()
		}

		// CalculateBorrowRate calculates the current interest rate based on utilization (the fraction of supply that has ien borrowed)
		borrowAPY, err := CalculateBorrowRate(moneyMarket.InterestRateModel, sdk.NewDecFromInt(cash), sdk.NewDecFromInt(borrowed.Amount), sdk.NewDecFromInt(reserves.AmountOf(denom)))
		if err != nil {
			return nil, err
		}

		utilRatio := CalculateUtilizationRatio(sdk.NewDecFromInt(cash), sdk.NewDecFromInt(borrowed.Amount), sdk.NewDecFromInt(reserves.AmountOf(denom)))
		fullSupplyAPY := borrowAPY.Mul(utilRatio)
		realSupplyAPY := fullSupplyAPY.Mul(sdk.OneDec().Sub(moneyMarket.ReserveFactor))

		moneyMarketInterestRate := types.MoneyMarketInterestRate{
			Denom:              denom,
			SupplyInterestRate: realSupplyAPY.String(),
			BorrowInterestRate: borrowAPY.String(),
		}

		moneyMarketInterestRates = append(moneyMarketInterestRates, moneyMarketInterestRate)
	}

	return &types.QueryInterestRateResponse{
		InterestRates: moneyMarketInterestRates,
	}, nil
}

func (qs QueryServer) Reserves(ctx context.Context, req *types.QueryReservesRequest) (*types.QueryReservesResponse, error) {
	if req == nil {
		return nil, status.Errorf(codes.InvalidArgument, "empty request")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)

	reserveCoins, found := qs.keeper.GetTotalReserves(sdkCtx)
	if !found {
		reserveCoins = sdk.Coins{}
	}

	// If user specified a denom only return coins of that denom type
	if len(req.Denom) > 0 {
		reserveCoins = sdk.NewCoins(sdk.NewCoin(req.Denom, reserveCoins.AmountOf(req.Denom)))
	}

	return &types.QueryReservesResponse{
		Amount: reserveCoins,
	}, nil
}

func (qs QueryServer) InterestFactors(ctx context.Context, req *types.QueryInterestFactorsRequest) (*types.QueryInterestFactorsResponse, error) {
	if req == nil {
		return nil, status.Errorf(codes.InvalidArgument, "empty request")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)

	var interestFactors types.InterestFactors
	if len(req.Denom) > 0 {
		// Fetch supply/borrow interest factors for a single denom
		interestFactor := types.InterestFactor{}
		interestFactor.Denom = req.Denom
		supplyInterestFactor, found := qs.keeper.GetSupplyInterestFactor(sdkCtx, req.Denom)
		if found {
			interestFactor.SupplyInterestFactor = supplyInterestFactor.String()
		}
		borrowInterestFactor, found := qs.keeper.GetBorrowInterestFactor(sdkCtx, req.Denom)
		if found {
			interestFactor.BorrowInterestFactor = borrowInterestFactor.String()
		}
		interestFactors = append(interestFactors, interestFactor)
	} else {
		interestFactorMap := make(map[string]types.InterestFactor)
		// Populate mapping with supply interest factors
		qs.keeper.IterateSupplyInterestFactors(sdkCtx, func(denom string, factor sdk.Dec) (stop bool) {
			interestFactor := types.InterestFactor{Denom: denom, SupplyInterestFactor: factor.String()}
			interestFactorMap[denom] = interestFactor
			return false
		})
		// Populate mapping with borrow interest factors
		qs.keeper.IterateBorrowInterestFactors(sdkCtx, func(denom string, factor sdk.Dec) (stop bool) {
			interestFactor, ok := interestFactorMap[denom]
			if !ok {
				newInterestFactor := types.InterestFactor{Denom: denom, BorrowInterestFactor: factor.String()}
				interestFactorMap[denom] = newInterestFactor
			} else {
				interestFactor.BorrowInterestFactor = factor.String()
				interestFactorMap[denom] = interestFactor
			}
			return false
		})
		// Translate mapping to slice
		for _, val := range interestFactorMap {
			interestFactors = append(interestFactors, val)
		}
	}

	return &types.QueryInterestFactorsResponse{
		InterestFactors: interestFactors,
	}, nil
}
