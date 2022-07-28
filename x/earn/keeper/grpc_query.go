package keeper

import (
	"context"
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"

	"github.com/kava-labs/kava/x/earn/types"
)

type queryServer struct {
	keeper Keeper
}

// NewQueryServerImpl creates a new server for handling gRPC queries.
func NewQueryServerImpl(k Keeper) types.QueryServer {
	return &queryServer{keeper: k}
}

var _ types.QueryServer = queryServer{}

// Params implements the gRPC service handler for querying x/earn parameters.
func (s queryServer) Params(
	ctx context.Context,
	req *types.QueryParamsRequest,
) (*types.QueryParamsResponse, error) {
	if req == nil {
		return nil, status.Errorf(codes.InvalidArgument, "empty request")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	params := s.keeper.GetParams(sdkCtx)

	return &types.QueryParamsResponse{Params: params}, nil
}

// Vaults implements the gRPC service handler for querying x/earn vaults.
func (s queryServer) Vaults(
	ctx context.Context,
	req *types.QueryVaultsRequest,
) (*types.QueryVaultsResponse, error) {
	if req == nil {
		return nil, status.Errorf(codes.InvalidArgument, "empty request")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)

	var queriedAllowedVaults types.AllowedVaults

	if req.Denom != "" {
		// Only 1 vault
		allowedVault, found := s.keeper.GetAllowedVault(sdkCtx, req.Denom)
		if !found {
			return nil, status.Errorf(codes.NotFound, "vault not found with specified denom")
		}

		queriedAllowedVaults = types.AllowedVaults{allowedVault}
	} else {
		// All vaults
		queriedAllowedVaults = s.keeper.GetAllowedVaults(sdkCtx)
	}

	vaults := []types.VaultResponse{}

	for _, allowedVault := range queriedAllowedVaults {
		totalSupplied, err := s.keeper.GetVaultTotalSupplied(sdkCtx, allowedVault.Denom)
		if err != nil {
			// No supply yet, no error just zero
			totalSupplied = sdk.NewCoin(allowedVault.Denom, sdk.ZeroInt())
		}

		totalValue, err := s.keeper.GetVaultTotalValue(sdkCtx, allowedVault.Denom)
		if err != nil {
			return nil, err
		}

		vaults = append(vaults, types.VaultResponse{
			Denom:         allowedVault.Denom,
			VaultStrategy: allowedVault.VaultStrategy,
			TotalSupplied: totalSupplied.Amount,
			TotalValue:    totalValue.Amount,
		})
	}

	return &types.QueryVaultsResponse{
		Vaults: vaults,
	}, nil
}

// Deposits implements the gRPC service handler for querying x/earn deposits.
func (s queryServer) Deposits(
	ctx context.Context,
	req *types.QueryDepositsRequest,
) (*types.QueryDepositsResponse, error) {
	if req == nil {
		return nil, status.Errorf(codes.InvalidArgument, "empty request")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)

	// 1. Specific account and specific vault
	if req.Owner != "" && req.Denom != "" {
		owner, err := sdk.AccAddressFromBech32(req.Owner)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, "Invalid address")
		}

		shareRecord, found := s.keeper.GetVaultShareRecord(sdkCtx, owner)
		if !found {
			return nil, status.Error(codes.NotFound, "No deposit found for owner")
		}

		if shareRecord.AmountSupplied.AmountOf(req.Denom).IsZero() {
			return nil, status.Error(codes.NotFound, fmt.Sprintf("No deposit for denom %s found for owner", req.Denom))
		}

		value, err := getAccountValue(sdkCtx, s.keeper, owner, shareRecord.AmountSupplied)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}

		return &types.QueryDepositsResponse{
			Deposits: []types.DepositResponse{
				{
					Depositor:      owner.String(),
					AmountSupplied: shareRecord.AmountSupplied,
					Value:          value,
				},
			},
			Pagination: nil,
		}, nil
	}

	// 2. All accounts, specific vault
	if req.Owner == "" && req.Denom != "" {
		_, found := s.keeper.GetVaultRecord(sdkCtx, req.Denom)
		if !found {
			return nil, status.Error(codes.NotFound, "Vault record for denom not found")
		}

		deposits := []types.DepositResponse{}
		store := prefix.NewStore(sdkCtx.KVStore(s.keeper.key), types.VaultShareRecordKeyPrefix)

		pageRes, err := query.FilteredPaginate(
			store,
			req.Pagination,
			func(key []byte, value []byte, accumulate bool) (bool, error) {
				var record types.VaultShareRecord
				err := s.keeper.cdc.Unmarshal(value, &record)
				if err != nil {
					return false, err
				}

				// Only those that have amount of requested denom
				if record.AmountSupplied.AmountOf(req.Denom).IsZero() {
					// inform paginate that there was no match on this key
					return false, nil
				}

				if accumulate {
					accValue, err := getAccountValue(sdkCtx, s.keeper, record.Depositor, record.AmountSupplied)
					if err != nil {
						return false, err
					}

					// only add to results if paginate tells us to
					deposits = append(deposits, types.DepositResponse{
						Depositor:      record.Depositor.String(),
						AmountSupplied: record.AmountSupplied,
						Value:          accValue,
					})
				}

				// inform paginate that were was a match on this key
				return true, nil
			},
		)

		if err != nil {
			return nil, err
		}

		return &types.QueryDepositsResponse{
			Deposits:   deposits,
			Pagination: pageRes,
		}, nil
	}

	// 3. Specific account, all vaults
	if req.Owner != "" && req.Denom == "" {
		owner, err := sdk.AccAddressFromBech32(req.Owner)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, "Invalid address")
		}

		deposits := []types.DepositResponse{}

		accountShare, found := s.keeper.GetVaultShareRecord(sdkCtx, owner)
		if !found {
			return nil, status.Error(codes.NotFound, "No deposit found for owner")
		}

		value, err := getAccountValue(sdkCtx, s.keeper, owner, accountShare.AmountSupplied)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}

		deposits = append(deposits, types.DepositResponse{
			Depositor:      owner.String(),
			AmountSupplied: accountShare.AmountSupplied,
			Value:          value,
		})

		return &types.QueryDepositsResponse{
			Deposits:   deposits,
			Pagination: nil,
		}, nil
	}

	// 4. All accounts, all vaults
	deposits := []types.DepositResponse{}
	store := prefix.NewStore(sdkCtx.KVStore(s.keeper.key), types.VaultShareRecordKeyPrefix)

	pageRes, err := query.Paginate(
		store,
		req.Pagination,
		func(key []byte, value []byte) error {
			var record types.VaultShareRecord
			err := s.keeper.cdc.Unmarshal(value, &record)
			if err != nil {
				return err
			}

			accValue, err := getAccountValue(sdkCtx, s.keeper, record.Depositor, record.AmountSupplied)
			if err != nil {
				return err
			}

			// only add to results if paginate tells us to
			deposits = append(deposits, types.DepositResponse{
				Depositor:      record.Depositor.String(),
				AmountSupplied: record.AmountSupplied,
				Value:          accValue,
			})

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return &types.QueryDepositsResponse{
		Deposits:   deposits,
		Pagination: pageRes,
	}, nil
}

// Deposits implements the gRPC service handler for querying x/earn deposits.
func (s queryServer) TotalDeposited(
	ctx context.Context,
	req *types.QueryTotalDepositedRequest,
) (*types.QueryTotalDepositedResponse, error) {
	if req == nil {
		return nil, status.Errorf(codes.InvalidArgument, "empty request")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)

	// Single vault
	if req.Denom != "" {
		totalSupplied, err := s.keeper.GetVaultTotalSupplied(sdkCtx, req.Denom)
		if err != nil {
			return nil, err
		}

		return &types.QueryTotalDepositedResponse{
			SuppliedCoins: sdk.NewCoins(totalSupplied),
		}, nil
	}

	coins := sdk.NewCoins()
	vaults := s.keeper.GetAllVaultRecords(sdkCtx)

	for _, vault := range vaults {
		coins = coins.Add(vault.TotalSupply)
	}

	return &types.QueryTotalDepositedResponse{
		SuppliedCoins: coins,
	}, nil
}

func getAccountValue(
	ctx sdk.Context,
	keeper Keeper,
	account sdk.AccAddress,
	supplied sdk.Coins,
) (sdk.Coins, error) {
	value := sdk.NewCoins()

	for _, coin := range supplied {
		accValue, err := keeper.GetVaultAccountValue(ctx, coin.Denom, account)
		if err != nil {
			return nil, err
		}

		value = value.Add(sdk.NewCoin(coin.Denom, accValue.Amount))
	}

	return value, nil
}
