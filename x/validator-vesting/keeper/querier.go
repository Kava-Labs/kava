package keeper

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

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
	totalSupply := keeper.bankKeeper.GetSupply(ctx).GetTotal().AmountOf("ukava")
	supplyInt := sdk.NewDecFromInt(totalSupply).Mul(sdk.MustNewDecFromStr("0.000001")).TruncateInt64()
	bz, err := types.ModuleCdc.MarshalJSON(supplyInt)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
	}
	return bz, nil
}

func queryGetCirculatingSupply(ctx sdk.Context, req abci.RequestQuery, keeper Keeper) ([]byte, error) {
	supplyInt := getCirculatingSupply(ctx.BlockTime())
	bz, err := keeper.cdc.MarshalJSON(supplyInt)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
	}
	return bz, nil
}

func getCirculatingSupply(blockTime time.Time) sdk.Int {
	vestingDates := []time.Time{
		time.Date(2020, 9, 5, 14, 0, 0, 0, time.UTC),
		time.Date(2020, 11, 5, 14, 0, 0, 0, time.UTC),
		time.Date(2021, 2, 5, 14, 0, 0, 0, time.UTC),
		time.Date(2021, 5, 5, 14, 0, 0, 0, time.UTC),
		time.Date(2021, 8, 5, 14, 0, 0, 0, time.UTC),
		time.Date(2021, 11, 5, 14, 0, 0, 0, time.UTC),
		time.Date(2022, 2, 5, 14, 0, 0, 0, time.UTC),
		time.Date(2022, 5, 5, 14, 0, 0, 0, time.UTC),
		time.Date(2022, 8, 5, 14, 0, 0, 0, time.UTC),
		time.Date(2022, 11, 5, 14, 0, 0, 0, time.UTC),
	}

	switch {
	case blockTime.Before(vestingDates[0]):
		return sdk.NewInt(27190672)
	case blockTime.After(vestingDates[0]) && blockTime.Before(vestingDates[1]) || blockTime.Equal(vestingDates[0]):
		return sdk.NewInt(29442227)
	case blockTime.After(vestingDates[1]) && blockTime.Before(vestingDates[2]) || blockTime.Equal(vestingDates[1]):
		return sdk.NewInt(46876230)
	case blockTime.After(vestingDates[2]) && blockTime.Before(vestingDates[3]) || blockTime.Equal(vestingDates[2]):
		return sdk.NewInt(58524186)
	case blockTime.After(vestingDates[3]) && blockTime.Before(vestingDates[4]) || blockTime.Equal(vestingDates[3]):
		return sdk.NewInt(70172142)
	case blockTime.After(vestingDates[4]) && blockTime.Before(vestingDates[5]) || blockTime.Equal(vestingDates[4]):
		return sdk.NewInt(81443180)
	case blockTime.After(vestingDates[5]) && blockTime.Before(vestingDates[6]) || blockTime.Equal(vestingDates[5]):
		return sdk.NewInt(90625000)
	case blockTime.After(vestingDates[6]) && blockTime.Before(vestingDates[7]) || blockTime.Equal(vestingDates[6]):
		return sdk.NewInt(92968750)
	case blockTime.After(vestingDates[7]) && blockTime.Before(vestingDates[8]) || blockTime.Equal(vestingDates[7]):
		return sdk.NewInt(95312500)
	case blockTime.After(vestingDates[8]) && blockTime.Before(vestingDates[9]) || blockTime.Equal(vestingDates[8]):
		return sdk.NewInt(97656250)
	default:
		return sdk.NewInt(100000000)
	}

}
