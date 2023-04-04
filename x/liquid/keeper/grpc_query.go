package keeper

import (
	"context"
	"fmt"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	vestingexported "github.com/cosmos/cosmos-sdk/x/auth/vesting/exported"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
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

	bondDenom := s.keeper.stakingKeeper.BondDenom(ctx)
	vesting := s.getVesting(ctx, delegator).AmountOf(bondDenom)

	vestingDelegated := sdk.MinInt(vesting, delegated)
	vestedDelegated := delegated.Sub(vestingDelegated)

	res := types.QueryDelegatedBalanceResponse{
		Vested:  sdk.NewCoin(bondDenom, vestedDelegated),
		Vesting: sdk.NewCoin(bondDenom, vestingDelegated),
	}
	return &res, nil
}

func (s queryServer) TotalSupply(
	goCtx context.Context,
	req *types.QueryTotalSupplyRequest,
) (*types.QueryTotalSupplyResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	totalValue, err := s.keeper.GetTotalDerivativeValue(ctx)
	if err != nil {
		return nil, err
	}

	return &types.QueryTotalSupplyResponse{
		Height: ctx.BlockHeight(),
		Result: []sdk.Coin{totalValue},
	}, nil
}

func (s queryServer) getDelegatedBalance(ctx sdk.Context, delegator sdk.AccAddress) sdkmath.Int {
	balance := sdk.ZeroDec()

	s.keeper.stakingKeeper.IterateDelegatorDelegations(ctx, delegator, func(delegation stakingtypes.Delegation) bool {
		validator, found := s.keeper.stakingKeeper.GetValidator(ctx, delegation.GetValidatorAddr())
		if !found {
			panic(fmt.Sprintf("validator %s for delegation not found", delegation.GetValidatorAddr()))
		}
		tokens := validator.TokensFromSharesTruncated(delegation.GetShares())
		balance = balance.Add(tokens)

		return false
	})
	return balance.TruncateInt()
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
