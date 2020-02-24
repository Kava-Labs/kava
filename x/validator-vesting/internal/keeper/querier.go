package keeper

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authexported "github.com/cosmos/cosmos-sdk/x/auth/exported"
	"github.com/cosmos/cosmos-sdk/x/auth/vesting"
	supplyexported "github.com/cosmos/cosmos-sdk/x/supply/exported"
	"github.com/kava-labs/kava/x/validator-vesting/internal/types"
	abci "github.com/tendermint/tendermint/abci/types"
)

// NewQuerier returns a new querier function
func NewQuerier(keeper Keeper) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) (res []byte, err sdk.Error) {
		switch path[0] {
		case types.QueryCirculatingSupply:
			return queryGetCirculatingSupply(ctx, req, keeper)
		case types.QueryTotalSupply:
			return queryGetTotalSupply(ctx, req, keeper)
		default:
			return nil, sdk.ErrUnknownRequest("unknown cdp query endpoint")
		}
	}
}

func queryGetTotalSupply(ctx sdk.Context, req abci.RequestQuery, keeper Keeper) ([]byte, sdk.Error) {
	totalSupply := keeper.supplyKeeper.GetSupply(ctx).GetTotal().AmountOf("ukava")
	bz, err := codec.MarshalJSONIndent(keeper.cdc, totalSupply)
	if err != nil {
		return nil, sdk.ErrInternal(err.Error())
	}
	return bz, nil
}

func queryGetCirculatingSupply(ctx sdk.Context, req abci.RequestQuery, keeper Keeper) ([]byte, sdk.Error) {
	circulatingSupply := sdk.ZeroInt()
	keeper.ak.IterateAccounts(ctx,
		func(acc authexported.Account) (stop bool) {
			// exclude module account
			_, ok := acc.(supplyexported.ModuleAccountI)
			if ok {
				return false
			}

			// periodic vesting account
			vacc, ok := acc.(vesting.PeriodicVestingAccount)
			if ok {
				balance := vacc.GetCoins().AmountOf("ukava")
				if balance.IsZero() {
					return false
				}
				spendableBalance := vacc.SpendableCoins(ctx.BlockTime()).AmountOf("ukava")
				circulatingSupply = circulatingSupply.Add(sdk.MinInt(balance, spendableBalance))
				return false
			}

			// base account
			bacc, ok := acc.(*auth.BaseAccount)
			if ok {
				// add all coins
				circulatingSupply = circulatingSupply.Add(bacc.GetCoins().AmountOf("ukava"))
			}
			return false
		})

	bz, err := codec.MarshalJSONIndent(keeper.cdc, circulatingSupply)
	if err != nil {
		return nil, sdk.ErrInternal(err.Error())
	}
	return bz, nil
}
