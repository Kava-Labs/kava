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
	if req.Denom == bkavaDenom {
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
	allBkava := sdk.NewCoins()

	var iterErr error
	s.keeper.IterateVaultRecords(ctx, func(record types.VaultRecord) (stop bool) {
		// Skip non bkava vaults
		if !strings.HasPrefix(record.TotalShares.Denom, bkavaPrefix) {
			return false
		}

		vaultValue, err := s.keeper.GetVaultTotalValue(ctx, record.TotalShares.Denom)
		if err != nil {
			iterErr = err
			return false
		}

		allBkava = allBkava.Add(vaultValue)

		return false
	})

	if iterErr != nil {
		return nil, iterErr
	}

	vaultValue, err := s.keeper.liquidKeeper.GetStakedTokensForDerivatives(ctx, allBkava)
	if err != nil {
		return nil, err
	}

	return &types.QueryVaultResponse{
		Vault: types.VaultResponse{
			Denom:             bkavaDenom,
			Strategies:        allowedVault.Strategies,
			IsPrivateVault:    allowedVault.IsPrivateVault,
			AllowedDepositors: addressSliceToStringSlice(allowedVault.AllowedDepositors),
			// Empty for shares, as adding up all shares is not useful information
			TotalShares: "0",
			TotalValue:  vaultValue.Amount,
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
	if req.Denom == bkavaDenom {
		return s.getOneAccountBkavaVaultDeposit(sdkCtx, req)
	}

	// specific vault
	if req.Denom != "" {
		return s.getOneAccountOneVaultDeposit(sdkCtx, req)
	}

	// all vaults
	return s.getOneAccountAllDeposits(sdkCtx, req)
}

// TotalSupply implements the gRPC service handler for querying x/earn total supply (TVL)
func (s queryServer) TotalSupply(
	ctx context.Context,
	req *types.QueryTotalSupplyRequest,
) (*types.QueryTotalSupplyResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	totalSupply := sdk.NewCoins()
	liquidStakedDerivatives := sdk.NewCoins()

	// allowed vaults param contains info on allowed strategies, but bkava is aggregated
	allowedVaults := s.keeper.GetAllowedVaults(sdkCtx)
	allowedVaultByDenom := make(map[string]types.AllowedVault)
	for _, av := range allowedVaults {
		allowedVaultByDenom[av.Denom] = av
	}

	var vaultRecordErr error
	// iterate actual records to properly enumerate all denoms
	s.keeper.IterateVaultRecords(sdkCtx, func(vault types.VaultRecord) (stop bool) {
		isLiquidStakingDenom := false
		// find allowed vault to get parameters. handle translating bkava denoms to allowed vault denom
		allowedVaultDenom := vault.TotalShares.Denom
		if strings.HasPrefix(vault.TotalShares.Denom, bkavaPrefix) {
			isLiquidStakingDenom = true
			allowedVaultDenom = bkavaDenom
		}
		allowedVault, found := allowedVaultByDenom[allowedVaultDenom]
		if !found {
			vaultRecordErr = fmt.Errorf("vault record not found for vault record denom %s", vault.TotalShares.Denom)
			return true
		}

		// only consider savings strategy vaults when determining supply
		if !allowedVault.IsStrategyAllowed(types.STRATEGY_TYPE_SAVINGS) {
			return false
		}

		// vault has savings strategy! determine total value of vault and add to sum
		vaultSupply, err := s.keeper.GetVaultTotalValue(sdkCtx, vault.TotalShares.Denom)
		if err != nil {
			vaultRecordErr = err
			return true
		}

		// liquid staked tokens must be converted to their underlying value
		// aggregate them here and then we can convert to underlying values all at once at the end
		if isLiquidStakingDenom {
			liquidStakedDerivatives = liquidStakedDerivatives.Add(vaultSupply)
		} else {
			totalSupply = totalSupply.Add(vaultSupply)
		}
		return false
	})

	// determine underlying value of bkava denoms
	if len(liquidStakedDerivatives) > 0 {
		underlyingValue, err := s.keeper.liquidKeeper.GetStakedTokensForDerivatives(
			sdkCtx,
			liquidStakedDerivatives,
		)
		if err != nil {
			return nil, err
		}
		totalSupply = totalSupply.Add(sdk.NewCoin(bkavaDenom, underlyingValue.Amount))
	}

	return &types.QueryTotalSupplyResponse{
		Height: sdkCtx.BlockHeight(),
		Result: totalSupply,
	}, vaultRecordErr
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
		return &types.QueryDepositsResponse{
			Deposits: []types.DepositResponse{
				{
					Depositor: depositor.String(),
					// Zero shares and zero value for no deposits
					Shares: types.NewVaultShares(types.NewVaultShare(req.Denom, sdk.ZeroDec())),
					Value:  sdk.NewCoins(sdk.NewCoin(req.Denom, sdk.ZeroInt())),
				},
			},
			Pagination: nil,
		}, nil
	}

	// Only requesting the value of the specified denom
	value, err := s.keeper.GetVaultAccountValue(ctx, req.Denom, depositor)
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}

	if req.ValueInStakedTokens {
		// Get underlying ukava amount if denom is a derivative
		if !s.keeper.liquidKeeper.IsDerivativeDenom(ctx, req.Denom) {
			return nil, status.Errorf(
				codes.InvalidArgument,
				"denom %s is not a derivative, ValueInStakedTokens can only be used with liquid derivatives",
				req.Denom,
			)
		}

		ukavaValue, err := s.keeper.liquidKeeper.GetStakedTokensForDerivatives(ctx, sdk.NewCoins(value))
		if err != nil {
			// This should "never" happen if IsDerivativeDenom is true
			panic("Error getting ukava value for " + req.Denom)
		}

		value = ukavaValue
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
		return &types.QueryDepositsResponse{
			Deposits: []types.DepositResponse{
				{
					Depositor: depositor.String(),
					// Zero shares and zero value for no deposits
					Shares: types.NewVaultShares(types.NewVaultShare(req.Denom, sdk.ZeroDec())),
					Value:  sdk.NewCoins(sdk.NewCoin(req.Denom, sdk.ZeroInt())),
				},
			},
			Pagination: nil,
		}, nil
	}

	// Get all account deposit values to add up bkava
	totalAccountValue, err := getAccountTotalValue(ctx, s.keeper, depositor, shareRecord.Shares)
	if err != nil {
		return nil, err
	}

	// Remove non-bkava coins, GetStakedTokensForDerivatives expects only bkava
	totalBkavaValue := sdk.NewCoins()
	for _, coin := range totalAccountValue {
		if s.keeper.liquidKeeper.IsDerivativeDenom(ctx, coin.Denom) {
			totalBkavaValue = totalBkavaValue.Add(coin)
		}
	}

	// Use account value with only the aggregate bkava converted to underlying staked tokens
	stakedValue, err := s.keeper.liquidKeeper.GetStakedTokensForDerivatives(ctx, totalBkavaValue)
	if err != nil {
		return nil, err
	}

	return &types.QueryDepositsResponse{
		Deposits: []types.DepositResponse{
			{
				Depositor: depositor.String(),
				// Only respond with requested denom shares
				Shares: types.NewVaultShares(
					types.NewVaultShare(req.Denom, shareRecord.Shares.AmountOf(req.Denom)),
				),
				Value: sdk.NewCoins(stakedValue),
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
		return &types.QueryDepositsResponse{
			Deposits:   []types.DepositResponse{},
			Pagination: nil,
		}, nil
	}

	value, err := getAccountTotalValue(ctx, s.keeper, depositor, accountShare.Shares)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if req.ValueInStakedTokens {
		// Plain slice to not sum ukava amounts together. This is not a valid
		// sdk.Coin due to multiple coins of the same denom, but we need them to
		// be separate in the response to not be an aggregate amount.
		var valueInStakedTokens []sdk.Coin

		for _, coin := range value {
			// Skip non-bkava coins
			if !s.keeper.liquidKeeper.IsDerivativeDenom(ctx, coin.Denom) {
				continue
			}

			// Derivative coins are converted to underlying staked tokens
			ukavaValue, err := s.keeper.liquidKeeper.GetStakedTokensForDerivatives(ctx, sdk.NewCoins(coin))
			if err != nil {
				// This should "never" happen if IsDerivativeDenom is true
				panic("Error getting ukava value for " + coin.Denom)
			}
			valueInStakedTokens = append(valueInStakedTokens, ukavaValue)
		}

		var filteredShares types.VaultShares
		for _, share := range accountShare.Shares {
			// Remove non-bkava coins from shares as they are used to
			// determine which value is mapped to which denom
			// These should be in the same order as valueInStakedTokens
			if !s.keeper.liquidKeeper.IsDerivativeDenom(ctx, share.Denom) {
				continue
			}

			filteredShares = append(filteredShares, share)
		}

		value = valueInStakedTokens
		accountShare.Shares = filteredShares
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
