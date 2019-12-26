package types

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto/ed25519"
)

var addr = sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address())

func TestKeys(t *testing.T) {
	key := CdpKey(0x01, 2)
	db, id := SplitCdpKey(key)
	require.Equal(t, int(id), 2)
	require.Equal(t, byte(0x01), db)

	denomKey := DenomIterKey(0x01)
	db = SplitDenomIterKey(denomKey)
	require.Equal(t, byte(0x01), db)

	depositKey := DepositKey(2, addr)
	id, a := SplitDepositKey(depositKey)
	require.Equal(t, 2, int(id))
	require.Equal(t, a, addr)

	collateralKey := CollateralRatioKey(0x01, 2, sdk.MustNewDecFromStr("1.50"))
	db, id, ratio := SplitCollateralRatioKey(collateralKey)
	require.Equal(t, byte(0x01), db)
	require.Equal(t, int(id), 2)
	require.Equal(t, ratio, sdk.MustNewDecFromStr("1.50"))

	bigRatio := sdk.OneDec().Quo(sdk.SmallestDec()).Mul(sdk.OneDec().Add(sdk.OneDec()))
	collateralKey = CollateralRatioKey(0x01, 2, bigRatio)
	db, id, ratio = SplitCollateralRatioKey(collateralKey)
	require.Equal(t, ratio, MaxSortableDec)

	collateralIterKey := CollateralRatioIterKey(0x01, sdk.MustNewDecFromStr("1.50"))
	db, ratio = SplitCollateralRatioIterKey(collateralIterKey)
	require.Equal(t, byte(0x01), db)
	require.Equal(t, ratio, sdk.MustNewDecFromStr("1.50"))

	require.Panics(t, func() { SplitCollateralRatioKey(badRatioKey()) })
	require.Panics(t, func() { SplitCollateralRatioIterKey(badRatioIterKey()) })

}

func badRatioKey() []byte {
	r := append(append(append(append([]byte{0x01}, sep...), []byte("nonsense")...), sep...), []byte{0xff}...)
	return r
}

func badRatioIterKey() []byte {
	r := append(append([]byte{0x01}, sep...), []byte("nonsense")...)
	return r
}
