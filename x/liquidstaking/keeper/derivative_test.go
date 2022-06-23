package keeper_test

import (
	"fmt"
	"testing"

	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/require"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/liquidstaking/types"
)

func TestMintDerivative(t *testing.T) {
	_, addrs := app.GeneratePrivKeyAddressPairs(5)

	tApp := app.NewTestApp()

	gen := app.NewAuthBankGenesisBuilder().
		WithSimpleAccount(addrs[0], sdk.NewCoins(sdk.NewInt64Coin("stake", 1e9))).
		BuildMarshalled(tApp.AppCodec())
	tApp.InitializeFromGenesisStates(gen)

	ctx := tApp.NewContext(false, tmproto.Header{Height: 1})

	msgCreate, err := stakingtypes.NewMsgCreateValidator(
		sdk.ValAddress(addrs[0]),
		ed25519.GenPrivKey().PubKey(),
		sdk.NewInt64Coin("stake", 1e9),
		stakingtypes.Description{},
		stakingtypes.NewCommissionRates(sdk.ZeroDec(), sdk.ZeroDec(), sdk.ZeroDec()),
		sdk.NewInt(1e6),
	)
	require.NoError(t, err)

	_, err = tApp.MsgServiceRouter().Handler(msgCreate)(ctx, msgCreate)
	require.NoError(t, err)

	msgMint := types.NewMsgMintDerivative(
		addrs[0],
		sdk.ValAddress(addrs[0]),
		sdk.NewDec(1e6),
	)
	_, err = tApp.MsgServiceRouter().Handler(&msgMint)(ctx, &msgMint)
	require.NoError(t, err)

	denom := fmt.Sprintf("%s-%s", "stake", sdk.ValAddress(addrs[0]).String())
	bal := tApp.GetBankKeeper().GetBalance(ctx, addrs[0], denom)

	require.Equal(t, sdk.NewInt64Coin(denom, 1e6), bal)

	// delegation
}
