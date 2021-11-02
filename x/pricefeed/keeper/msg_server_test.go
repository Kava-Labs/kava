package keeper_test

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/pricefeed/keeper"
	"github.com/kava-labs/kava/x/pricefeed/types"
	"github.com/stretchr/testify/require"
	tmprototypes "github.com/tendermint/tendermint/proto/tendermint/types"
)

func TestKeeper_PostPrice(t *testing.T) {
	_, addrs := app.GeneratePrivKeyAddressPairs(4)
	tApp := app.NewTestApp()
	ctx := tApp.NewContext(true, tmprototypes.Header{}).
		WithBlockTime(time.Now().UTC())
	k := tApp.GetPriceFeedKeeper()
	msgSrv := keeper.NewMsgServerImpl(k)

	var strAddrs []string
	for _, a := range addrs {
		strAddrs = append(strAddrs, a.String())
	}
	authorizedOracles := strAddrs[:2]
	unauthorizedAddrs := strAddrs[2:]

	mp := types.Params{
		Markets: []types.Market{
			{MarketID: "tstusd", BaseAsset: "tst", QuoteAsset: "usd", Oracles: authorizedOracles, Active: true},
		},
	}
	k.SetParams(ctx, mp)

	tests := []struct {
		giveOracle   string
		giveMarketId string
		givePrice    sdk.Dec
		giveExpiry   time.Time
		wantAccepted bool
		errorKind    error
	}{
		// Authorized
		{authorizedOracles[0], "tstusd", sdk.MustNewDecFromStr("0.5"), time.Now().UTC().Add(time.Hour * 1), true, nil},
		// Expired
		{authorizedOracles[0], "tstusd", sdk.MustNewDecFromStr("0.5"), time.Now().UTC().Add(-time.Hour * 1), false, types.ErrExpired},
		// Invalid market
		{authorizedOracles[0], "invalid", sdk.MustNewDecFromStr("0.5"), time.Now().UTC().Add(time.Hour * 1), false, types.ErrInvalidMarket},

		// Unauthorized
		{unauthorizedAddrs[0], "tstusd", sdk.MustNewDecFromStr("0.5"), time.Now().UTC().Add(time.Hour * 1), false, types.ErrInvalidOracle},
	}

	for _, tt := range tests {
		// Use MsgServer over keeper methods directly to tests against valid oracles
		msg := types.NewMsgPostPrice(tt.giveOracle, tt.giveMarketId, tt.givePrice, tt.giveExpiry)
		_, err := msgSrv.PostPrice(sdk.WrapSDKContext(ctx), msg)

		if tt.wantAccepted {
			require.NoError(t, err)
		} else {
			require.Error(t, err)
			require.ErrorIs(t, tt.errorKind, err)
		}
	}
}
