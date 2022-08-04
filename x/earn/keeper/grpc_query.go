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
		vaultTotalShares, found := s.keeper.GetVaultTotalShares(sdkCtx, allowedVault.Denom)
		if !found {
			// No supply yet, no error just zero
			vaultTotalShares = types.NewVaultShare(allowedVault.Denom, sdk.ZeroDec())
		}

		totalValue, err := s.keeper.GetVaultTotalValue(sdkCtx, allowedVault.Denom)
		if err != nil {
			return nil, err
		}

		vaults = append(vaults, types.VaultResponse{
			Denom:             allowedVault.Denom,
			Strategies:        allowedVault.Strategies,
			IsPrivateVault:    allowedVault.IsPrivateVault,
			AllowedDepositors: addressSliceToStringSlice(allowedVault.AllowedDepositors),
			TotalShares:       vaultTotalShares.Amount.String(),
			TotalValue:        totalValue.Amount,
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
		return s.getAccountVaultDeposit(sdkCtx, req)
	}

	// 2. All accounts, specific vault
	if req.Owner == "" && req.Denom != "" {
		return s.getVaultAllDeposits(sdkCtx, req)
	}

	// 3. Specific account, all vaults
	if req.Owner != "" && req.Denom == "" {
		return s.getAccountAllDeposits(sdkCtx, req)
	}

	// 4. All accounts, all vaults
	return s.getAllDeposits(sdkCtx, req)
}

// getAccountVaultDeposit returns deposits for a specific vault and a specific
// account
func (s queryServer) getAccountVaultDeposit(
	ctx sdk.Context,
	req *types.QueryDepositsRequest,
) (*types.QueryDepositsResponse, error) {
	owner, err := sdk.AccAddressFromBech32(req.Owner)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "Invalid address")
	}

	shareRecord, found := s.keeper.GetVaultShareRecord(ctx, owner)
	if !found {
		return nil, status.Error(codes.NotFound, "No deposit found for owner")
	}

	if shareRecord.Shares.AmountOf(req.Denom).IsZero() {
		return nil, status.Error(codes.NotFound, fmt.Sprintf("No deposit for denom %s found for owner", req.Denom))
	}

	value, err := getAccountValue(ctx, s.keeper, owner, shareRecord.Shares)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return &types.QueryDepositsResponse{
		Deposits: []types.DepositResponse{
			{
				Depositor: owner.String(),
				Shares:    shareRecord.Shares,
				Value:     value,
			},
		},
		Pagination: nil,
	}, nil
}

// getVaultAllDeposits returns all deposits for a specific vault
func (s queryServer) getVaultAllDeposits(
	ctx sdk.Context,
	req *types.QueryDepositsRequest,
) (*types.QueryDepositsResponse, error) {
	_, found := s.keeper.GetVaultRecord(ctx, req.Denom)
	if !found {
		return nil, status.Error(codes.NotFound, "Vault record for denom not found")
	}

	deposits := []types.DepositResponse{}
	store := prefix.NewStore(ctx.KVStore(s.keeper.key), types.VaultShareRecordKeyPrefix)

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
			if record.Shares.AmountOf(req.Denom).IsZero() {
				// inform paginate that there was no match on this key
				return false, nil
			}

			if accumulate {
				accValue, err := getAccountValue(ctx, s.keeper, record.Depositor, record.Shares)
				if err != nil {
					return false, err
				}

				// only add to results if paginate tells us to
				deposits = append(deposits, types.DepositResponse{
					Depositor: record.Depositor.String(),
					Shares:    record.Shares,
					Value:     accValue,
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

// getAccountAllDeposits returns deposits for all vaults for a specific account
func (s queryServer) getAccountAllDeposits(
	ctx sdk.Context,
	req *types.QueryDepositsRequest,
) (*types.QueryDepositsResponse, error) {
	owner, err := sdk.AccAddressFromBech32(req.Owner)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "Invalid address")
	}

	deposits := []types.DepositResponse{}

	accountShare, found := s.keeper.GetVaultShareRecord(ctx, owner)
	if !found {
		return nil, status.Error(codes.NotFound, "No deposit found for owner")
	}

	value, err := getAccountValue(ctx, s.keeper, owner, accountShare.Shares)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	deposits = append(deposits, types.DepositResponse{
		Depositor: owner.String(),
		Shares:    accountShare.Shares,
		Value:     value,
	})

	return &types.QueryDepositsResponse{
		Deposits:   deposits,
		Pagination: nil,
	}, nil
}

// getAllDeposits returns all deposits for all vaults
func (s queryServer) getAllDeposits(
	ctx sdk.Context,
	req *types.QueryDepositsRequest,
) (*types.QueryDepositsResponse, error) {
	deposits := []types.DepositResponse{}
	store := prefix.NewStore(ctx.KVStore(s.keeper.key), types.VaultShareRecordKeyPrefix)

	pageRes, err := query.Paginate(
		store,
		req.Pagination,
		func(key []byte, value []byte) error {
			var record types.VaultShareRecord
			err := s.keeper.cdc.Unmarshal(value, &record)
			if err != nil {
				return err
			}

			accValue, err := getAccountValue(ctx, s.keeper, record.Depositor, record.Shares)
			if err != nil {
				return err
			}

			// only add to results if paginate tells us to
			deposits = append(deposits, types.DepositResponse{
				Depositor: record.Depositor.String(),
				Shares:    record.Shares,
				Value:     accValue,
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

func getAccountValue(
	ctx sdk.Context,
	keeper Keeper,
	account sdk.AccAddress,
	shares types.VaultShares,
) (sdk.Coins, error) {
	value := sdk.NewCoins()

	for _, share := range shares {
		accValue, err := keeper.GetVaultAccountValue(ctx, share.Denom, account)
		if err != nil {
			return nil, err
		}

		value = value.Add(sdk.NewCoin(share.Denom, accValue.Amount))
	}

	return value, nil
}

func addressSliceToStringSlice(addresses []sdk.AccAddress) []string {
	var strings []string
	for _, address := range addresses {
		strings = append(strings, address.String())
	}

	return strings
}
