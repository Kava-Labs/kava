package keeper

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	vestingexported "github.com/cosmos/cosmos-sdk/x/auth/vesting/exported"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/kava-labs/kava/x/liquid/types"
)

type queryServer struct {
	keeper Keeper
}

// NewQueryServerImpl creates a new server for handling gRPC queries.
func NewQueryServerImpl(k Keeper) types.QueryServer {
	return &queryServer{keeper: k}
}

var _ types.QueryServer = queryServer{}

func (s queryServer) DelegatedBalance(
	goCtx context.Context,
	req *types.QueryDelegatedBalanceRequest,
) (*types.QueryDelegatedBalanceResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	delegator, err := sdk.AccAddressFromBech32(req.Delegator)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid delegator address: %s", err)
	}

	delegated := s.getDelegatedBalance(ctx, delegator)
	unbonding := s.getUnbondingBalance(ctx, delegator)

	bondDenom := s.keeper.stakingKeeper.BondDenom(ctx)
	vesting := s.getVesting(ctx, delegator).AmountOf(bondDenom)

	totalStaked := delegated.Add(unbonding)
	vestingDelegated := sdk.MinInt(vesting, totalStaked)
	vestedDelegated := totalStaked.Sub(vestingDelegated)

	res := types.QueryDelegatedBalanceResponse{
		Vested:  sdk.NewCoin(bondDenom, vestedDelegated),
		Vesting: sdk.NewCoin(bondDenom, vestingDelegated),
	}
	return &res, nil
}

func (s queryServer) getDelegatedBalance(ctx sdk.Context, delegator sdk.AccAddress) sdk.Int {
	balance := sdk.ZeroInt()

	delegations := s.keeper.stakingKeeper.GetDelegatorDelegations(ctx, delegator, 1000) // TODO what to set max to?
	for _, delegation := range delegations {
		validator, found := s.keeper.stakingKeeper.GetValidator(ctx, delegation.GetValidatorAddr())
		if !found {
			panic(fmt.Sprintf("validator %s for delegation not found", delegation.GetValidatorAddr()))
		}
		balance = balance.Add(validator.TokensFromShares(delegation.GetShares()).TruncateInt())
	}
	return balance
}

func (s queryServer) getUnbondingBalance(ctx sdk.Context, delegator sdk.AccAddress) sdk.Int {
	balance := sdk.ZeroInt()

	ubds := s.keeper.stakingKeeper.GetUnbondingDelegations(ctx, delegator, 1000) // TODO what to set max to?
	for _, ubd := range ubds {
		for _, entry := range ubd.Entries {
			balance = balance.Add(entry.Balance)
		}
	}
	return balance
}

func (s queryServer) getVesting(ctx sdk.Context, delegator sdk.AccAddress) sdk.Coins {
	acc := s.keeper.accountKeeper.GetAccount(ctx, delegator)
	if acc == nil {
		// account doesn't exist so amount vesting is 0
		return nil
	}
	vestAcc, ok := acc.(vestingexported.VestingAccount)
	if !ok {
		// account is not vesting type, so amount vesting is 0
		return nil
	}
	return vestAcc.GetVestingCoins(ctx.BlockTime())
}
