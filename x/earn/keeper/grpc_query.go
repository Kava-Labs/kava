package keeper

import (
	"context"
	"fmt"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	sdk "github.com/cosmos/cosmos-sdk/types"

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

	allowedVaults := s.keeper.GetAllowedVaults(sdkCtx)
	allowedVaultsMap := make(map[string]types.AllowedVault)
	visitedMap := make(map[string]bool)
	for _, av := range allowedVaults {
		allowedVaultsMap[av.Denom] = av
		visitedMap[av.Denom] = false
	}

	vaults := []types.VaultResponse{}

	var vaultRecordsErr error

	// Iterate over vault records instead of AllowedVaults to get all bkava-*
	// vaults
	s.keeper.IterateVaultRecords(sdkCtx, func(record types.VaultRecord) bool {
		// Check if bkava, use allowed vault
		allowedVaultDenom := record.TotalShares.Denom
		if strings.HasPrefix(record.TotalShares.Denom, bkavaPrefix) {
			allowedVaultDenom = bkavaDenom
		}

		allowedVault, found := allowedVaultsMap[allowedVaultDenom]
		if !found {
			vaultRecordsErr = fmt.Errorf("vault record not found for vault record denom %s", record.TotalShares.Denom)
			return true
		}

		totalValue, err := s.keeper.GetVaultTotalValue(sdkCtx, record.TotalShares.Denom)
		if err != nil {
			vaultRecordsErr = err
			// Stop iterating if error
			return true
		}

		vaults = append(vaults, types.VaultResponse{
			Denom:             record.TotalShares.Denom,
			Strategies:        allowedVault.Strategies,
			IsPrivateVault:    allowedVault.IsPrivateVault,
			AllowedDepositors: addressSliceToStringSlice(allowedVault.AllowedDepositors),
			TotalShares:       record.TotalShares.Amount.String(),
			TotalValue:        totalValue.Amount,
		})

		// Mark this allowed vault as visited
		visitedMap[allowedVaultDenom] = true

		return false
	})

	if vaultRecordsErr != nil {
		return nil, vaultRecordsErr
	}

	// Add the allowed vaults that have not been visited yet
	// These are always empty vaults, as the vault would have been visited
	// earlier if there are any deposits
	for denom, visited := range visitedMap {
		if visited {
			continue
		}

		allowedVault, found := allowedVaultsMap[denom]
		if !found {
			return nil, fmt.Errorf("vault record not found for vault record denom %s", denom)
		}

		vaults = append(vaults, types.VaultResponse{
			Denom:             denom,
			Strategies:        allowedVault.Strategies,
			IsPrivateVault:    allowedVault.IsPrivateVault,
			AllowedDepositors: addressSliceToStringSlice(allowedVault.AllowedDepositors),
			// No shares, no value
			TotalShares: sdk.ZeroDec().String(),
			TotalValue:  sdk.ZeroInt(),
		})
	}

	// Does not include vaults that have no deposits, only iterates over vault
	// records which exists only for those with deposits.
	return &types.QueryVaultsResponse{
		Vaults: vaults,
	}, nil
}

// Vaults implements the gRPC service handler for querying x/earn vaults.
func (s queryServer) Vault(
	ctx context.Context,
	req *types.QueryVaultRequest,
) (*types.QueryVaultResponse, error) {
	if req == nil {
		return nil, status.Errorf(codes.InvalidArgument, "empty request")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)

	if req.Denom == "" {
		return nil, status.Errorf(codes.InvalidArgument, "empty denom")
	}

	// Only 1 vault
	allowedVault, found := s.keeper.GetAllowedVault(sdkCtx, req.Denom)
	if !found {
		return nil, status.Errorf(codes.NotFound, "vault not found with specified denom")
	}

	// Handle bkava separately to get total of **all** bkava vaults
	if req.Denom == "bkava" {
		return s.getAggregateBkavaVault(sdkCtx, allowedVault)
	}

	// Must be req.Denom and not allowedVault.Denom to get full "bkava" denom
	vaultRecord, found := s.keeper.GetVaultRecord(sdkCtx, req.Denom)
	if !found {
		// No supply yet, no error just set it to zero
		vaultRecord.TotalShares = types.NewVaultShare(req.Denom, sdk.ZeroDec())
	}

	totalValue, err := s.keeper.GetVaultTotalValue(sdkCtx, req.Denom)
	if err != nil {
		return nil, err
	}

	vault := types.VaultResponse{
		// VaultRecord denom instead of AllowedVault.Denom for full bkava denom
		Denom:             vaultRecord.TotalShares.Denom,
		Strategies:        allowedVault.Strategies,
		IsPrivateVault:    allowedVault.IsPrivateVault,
		AllowedDepositors: addressSliceToStringSlice(allowedVault.AllowedDepositors),
		TotalShares:       vaultRecord.TotalShares.Amount.String(),
		TotalValue:        totalValue.Amount,
	}

	return &types.QueryVaultResponse{
		Vault: vault,
	}, nil
}

// getAggregateBkavaVault returns a VaultResponse of the total of all bkava
// vaults.
func (s queryServer) getAggregateBkavaVault(
	ctx sdk.Context,
	allowedVault types.AllowedVault,
) (*types.QueryVaultResponse, error) {
	totalValue := sdk.NewInt(0)

	var iterErr error
	s.keeper.IterateVaultRecords(ctx, func(record types.VaultRecord) (stop bool) {
		// Skip non bkava vaults
		if !strings.HasPrefix(record.TotalShares.Denom, "bkava") {
			return false
		}

		vaultValue, err := s.keeper.GetVaultTotalValue(ctx, record.TotalShares.Denom)
		if err != nil {
			iterErr = err
			return false
		}

		totalValue = totalValue.Add(vaultValue.Amount)

		return false
	})

	if iterErr != nil {
		return nil, iterErr
	}

	return &types.QueryVaultResponse{
		Vault: types.VaultResponse{
			Denom:             "bkava",
			Strategies:        allowedVault.Strategies,
			IsPrivateVault:    allowedVault.IsPrivateVault,
			AllowedDepositors: addressSliceToStringSlice(allowedVault.AllowedDepositors),
			// Empty for shares, as adding up all shares is not useful information
			TotalShares: "0",
			TotalValue:  totalValue,
		},
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

	if req.Depositor == "" {
		return nil, status.Errorf(codes.InvalidArgument, "depositor is required")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)

	// bkava aggregate total
	if req.Denom == "bkava" {
		return s.getOneAccountBkavaVaultDeposit(sdkCtx, req)
	}

	// specific vault
	if req.Denom != "" {
		return s.getOneAccountOneVaultDeposit(sdkCtx, req)
	}

	// all vaults
	return s.getOneAccountAllDeposits(sdkCtx, req)
}

// getOneAccountOneVaultDeposit returns deposits for a specific vault and a specific
// account
func (s queryServer) getOneAccountOneVaultDeposit(
	ctx sdk.Context,
	req *types.QueryDepositsRequest,
) (*types.QueryDepositsResponse, error) {
	depositor, err := sdk.AccAddressFromBech32(req.Depositor)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "Invalid address")
	}

	shareRecord, found := s.keeper.GetVaultShareRecord(ctx, depositor)
	if !found {
		return nil, status.Error(codes.NotFound, "No deposit found for owner")
	}

	// Only requesting the value of the specified denom
	value, err := s.keeper.GetVaultAccountValue(ctx, req.Denom, depositor)
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}

	return &types.QueryDepositsResponse{
		Deposits: []types.DepositResponse{
			{
				Depositor: depositor.String(),
				// Only respond with requested denom shares
				Shares: types.NewVaultShares(
					types.NewVaultShare(req.Denom, shareRecord.Shares.AmountOf(req.Denom)),
				),
				Value: sdk.NewCoins(value),
			},
		},
		Pagination: nil,
	}, nil
}

// getOneAccountBkavaVaultDeposit returns deposits for the aggregated bkava vault
// and a specific account
func (s queryServer) getOneAccountBkavaVaultDeposit(
	ctx sdk.Context,
	req *types.QueryDepositsRequest,
) (*types.QueryDepositsResponse, error) {
	depositor, err := sdk.AccAddressFromBech32(req.Depositor)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "Invalid address")
	}

	shareRecord, found := s.keeper.GetVaultShareRecord(ctx, depositor)
	if !found {
		return nil, status.Error(codes.NotFound, "No deposit found for owner")
	}

	// Get all account deposit values to add up bkava
	totalAccountValue, err := getAccountTotalValue(ctx, s.keeper, depositor, shareRecord.Shares)
	if err != nil {
		return nil, err
	}

	// Use account value with only the aggregate bkava
	bkavaValue := getTotalBkava(totalAccountValue)

	return &types.QueryDepositsResponse{
		Deposits: []types.DepositResponse{
			{
				Depositor: depositor.String(),
				// Only respond with requested denom shares
				Shares: types.NewVaultShares(
					types.NewVaultShare(req.Denom, shareRecord.Shares.AmountOf(req.Denom)),
				),
				Value: sdk.NewCoins(bkavaValue),
			},
		},
		Pagination: nil,
	}, nil
}

// getOneAccountAllDeposits returns deposits for all vaults for a specific account
func (s queryServer) getOneAccountAllDeposits(
	ctx sdk.Context,
	req *types.QueryDepositsRequest,
) (*types.QueryDepositsResponse, error) {
	depositor, err := sdk.AccAddressFromBech32(req.Depositor)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "Invalid address")
	}

	deposits := []types.DepositResponse{}

	accountShare, found := s.keeper.GetVaultShareRecord(ctx, depositor)
	if !found {
		return nil, status.Error(codes.NotFound, "No deposit found for depositor")
	}

	value, err := getAccountTotalValue(ctx, s.keeper, depositor, accountShare.Shares)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	deposits = append(deposits, types.DepositResponse{
		Depositor: depositor.String(),
		Shares:    accountShare.Shares,
		Value:     value,
	})

	return &types.QueryDepositsResponse{
		Deposits:   deposits,
		Pagination: nil,
	}, nil
}

// getAccountTotalValue returns the total value for all vaults for a specific
// account based on their shares.
func getAccountTotalValue(
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

func getTotalBkava(coins sdk.Coins) sdk.Coin {
	bkavaTotal := sdk.NewCoin("bkava", sdk.ZeroInt())

	for _, coin := range coins {
		if strings.HasPrefix(coin.Denom, "bkava") {
			bkavaTotal = bkavaTotal.AddAmount(coin.Amount)
		}
	}

	return bkavaTotal
}
