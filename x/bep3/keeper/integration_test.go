package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kava-labs/kava/app"

	// TODO: update alias
	"github.com/kava-labs/kava/x/bep3"
	"github.com/kava-labs/kava/x/bep3/types"
)

func NewBep3GenStateMulti() app.GenesisState {
	deputy, _ := sdk.AccAddressFromBech32("kava1xy7hrjy9r0algz9w3gzm8u6mrpq97kwta747gj")

	bep3Genesis := types.GenesisState{
		Params: bep3.Params{
			BnbDeputyAddress: deputy,
			MinBlockLock:     types.DefaultMinBlockLock, // 80
			MaxBlockLock:     types.DefaultMaxBlockLock, // 360
			SupportedAssets: types.AssetParams{
				types.AssetParam{
					Denom:  "bnb",
					CoinID: "714",
					Limit:  sdk.NewInt(100000000000),
					Active: true,
				},
			},
		},
	}

	return app.GenesisState{bep3.ModuleName: bep3.ModuleCdc.MustMarshalJSON(bep3Genesis)}
}
