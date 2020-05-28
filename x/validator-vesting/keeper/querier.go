package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authexported "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/auth/vesting"

	"github.com/kava-labs/kava/x/validator-vesting/types"

	abci "github.com/tendermint/tendermint/abci/types"
)

// NewQuerier returns a new querier function
func NewQuerier(keeper Keeper) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) (res []byte, err error) {
		switch path[0] {
		case types.QueryCirculatingSupply:
			return queryGetCirculatingSupply(ctx, req, keeper)
		case types.QueryTotalSupply:
			return queryGetTotalSupply(ctx, req, keeper)
		default:
			return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unknown %s query endpoint: %s", types.ModuleName, path[0])
		}
	}
}

func queryGetTotalSupply(ctx sdk.Context, req abci.RequestQuery, keeper Keeper) ([]byte, error) {
	totalSupply := keeper.supplyKeeper.GetSupply(ctx).GetTotal().AmountOf("ukava")
	supplyInt := sdk.NewDecFromInt(totalSupply).Mul(sdk.MustNewDecFromStr("0.000001")).TruncateInt64()
	bz, err := types.ModuleCdc.MarshalJSON(supplyInt)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
	}
	return bz, nil
}

func queryGetCirculatingSupply(ctx sdk.Context, req abci.RequestQuery, keeper Keeper) ([]byte, error) {
	circulatingSupply := keeper.supplyKeeper.GetSupply(ctx).GetTotal().AmountOf("ukava")
	keeper.ak.IterateAccounts(ctx,
		func(acc authexported.Account) (stop bool) {

			// validator vesting account
			vvacc, ok := acc.(*types.ValidatorVestingAccount)
			if ok {
				vestedBalance := vvacc.GetVestingCoins(ctx.BlockTime()).AmountOf("ukava")
				circulatingSupply = circulatingSupply.Sub(vestedBalance)
				return false
			}
			// periodic vesting account
			pvacc, ok := acc.(*vesting.PeriodicVestingAccount)
			if ok {
				vestedBalance := pvacc.GetVestingCoins(ctx.BlockTime()).AmountOf("ukava")
				circulatingSupply = circulatingSupply.Sub(vestedBalance)
				return false
			}
			return false
		})
	supplyInt := sdk.NewDecFromInt(circulatingSupply).Mul(sdk.MustNewDecFromStr("0.000001")).TruncateInt64()
	bz, err := types.ModuleCdc.MarshalJSON(supplyInt)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
	}
	return bz, nil
}
