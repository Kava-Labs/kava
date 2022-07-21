package keeper

import (
	"context"

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

	// Specific account and specific vault
	if req.Owner != "" && req.Denom != "" {
		owner, err := sdk.AccAddressFromBech32(req.Owner)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, "Invalid address")
		}

		deposit, found := s.keeper.GetVaultShareRecord(sdkCtx, req.Denom, owner)
		if !found {
			return nil, status.Error(codes.NotFound, "No deposit found for owner and denom")
		}

		value, err := s.keeper.GetVaultAccountValue(sdkCtx, req.Denom, owner)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}

		return &types.QueryDepositsResponse{
			Deposits: []types.DepositResponse{
				{
					Depositor:       owner.String(),
					Denom:           req.Denom,
					AccountSupplied: deposit.AmountSupplied.Amount,
					AccountValue:    value.Amount,
				},
			},
			Pagination: nil,
		}, nil
	}

	// Specific vault, all accounts
	if req.Denom != "" {
		_, found := s.keeper.GetVaultRecord(sdkCtx, req.Denom)
		if !found {
			return nil, status.Error(codes.NotFound, "Vault record for denom not found")
		}

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

				accValue, err := s.keeper.GetVaultAccountValue(sdkCtx, req.Denom, record.Depositor)
				if err != nil {
					return err
				}

				// only add to results if paginate tells us to
				deposits = append(deposits, types.DepositResponse{
					Depositor:       record.Depositor.String(),
					Denom:           req.Denom,
					AccountSupplied: record.AmountSupplied.Amount,
					AccountValue:    accValue.Amount,
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

	// Specific account, all vaults
	if req.Owner != "" && req.Denom == "" {
		owner, err := sdk.AccAddressFromBech32(req.Owner)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, "Invalid address")
		}

		deposits := []types.DepositResponse{}
		vaults := s.keeper.GetAllowedVaults(sdkCtx)

		for _, vault := range vaults {
			deposit, found := s.keeper.GetVaultShareRecord(sdkCtx, vault.Denom, owner)
			if !found {
				// No deposit found for this vault, skip instead of returning error
				continue
			}

			value, err := s.keeper.GetVaultAccountValue(sdkCtx, vault.Denom, owner)
			if err != nil {
				return nil, status.Error(codes.InvalidArgument, err.Error())
			}

			deposits = append(deposits, types.DepositResponse{
				Depositor:       owner.String(),
				Denom:           vault.Denom,
				AccountSupplied: deposit.AmountSupplied.Amount,
				AccountValue:    value.Amount,
			})

			return &types.QueryDepositsResponse{
				Deposits:   deposits,
				Pagination: nil,
			}, nil
		}
	}

	// All accounts, all vaults
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

			accValue, err := s.keeper.GetVaultAccountValue(sdkCtx, record.AmountSupplied.Denom, record.Depositor)
			if err != nil {
				return err
			}

			// only add to results if paginate tells us to
			deposits = append(deposits, types.DepositResponse{
				Depositor:       record.Depositor.String(),
				Denom:           record.AmountSupplied.Denom,
				AccountSupplied: record.AmountSupplied.Amount,
				AccountValue:    accValue.Amount,
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
