package ante_test

import (
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/simapp/helpers"
	sdk "github.com/cosmos/cosmos-sdk/types"
	staking "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/app/ante"
)

var (
	prik1              = ed25519.GenPrivKey()
	pk1                = prik1.PubKey()
	addr1, _           = sdk.Bech32ifyAddressBytes(sdk.Bech32PrefixAccAddr, pk1.Address().Bytes())
	valAddr1           = sdk.ValAddress(pk1.Address())
	commission1        = staking.NewCommissionRates(sdk.NewDecWithPrec(4, 2), sdk.NewDecWithPrec(4, 2), sdk.NewDecWithPrec(4, 2))
	newCommissionRate1 = sdk.NewDecWithPrec(4, 2)
	newSelfDelegation1 = sdk.OneInt()
	prik2              = ed25519.GenPrivKey()
	pk2                = prik2.PubKey()
	addr2, _           = sdk.Bech32ifyAddressBytes(sdk.Bech32PrefixAccAddr, pk1.Address().Bytes())
	valAddr2           = sdk.ValAddress(pk1.Address())
	commission2        = staking.NewCommissionRates(sdk.NewDecWithPrec(5, 2), sdk.NewDecWithPrec(5, 2), sdk.NewDecWithPrec(5, 2))
	newCommissionRate2 = sdk.NewDecWithPrec(5, 2)
	newSelfDelegation2 = sdk.OneInt()
	coinPos            = sdk.NewInt64Coin(sdk.DefaultBondDenom, 1000)
	coinZero           = sdk.NewInt64Coin(sdk.DefaultBondDenom, 0)
	description        = staking.NewDescription("a", "b", "c", "d", "e")
	cdc                = app.MakeEncodingConfig().Marshaler
)

func TestMinCommision_MsgCreateValidator(t *testing.T) {
	txConfig := app.MakeEncodingConfig().TxConfig
	decorator := ante.NewMinCommissionDecorator(cdc)
	createValidatorMsg1, _ := staking.NewMsgCreateValidator(valAddr1, pk1, coinPos, description, commission1, sdk.OneInt())
	tx1, err := helpers.GenSignedMockTx(
		rand.New(rand.NewSource(time.Now().UnixNano())),
		txConfig,
		[]sdk.Msg{createValidatorMsg1},
		sdk.NewCoins(),
		helpers.DefaultGenTxGas,
		"testing-chain-id",
		[]uint64{0},
		[]uint64{0},
		prik1,
	)
	require.NoError(t, err)
	mmd := MockAnteHandler{}
	ctx := sdk.Context{}.WithIsCheckTx(true)
	_, err = decorator.AnteHandle(ctx, tx1, false, mmd.AnteHandle)
	require.Error(t, err)
	require.Contains(t, err.Error(), "commission can't be lower than 5%")

	// success test
	createValidatorMsg2, _ := staking.NewMsgCreateValidator(valAddr2, pk2, coinPos, description, commission2, sdk.OneInt())
	tx2, err := helpers.GenSignedMockTx(
		rand.New(rand.NewSource(time.Now().UnixNano())),
		txConfig,
		[]sdk.Msg{createValidatorMsg2},
		sdk.NewCoins(),
		helpers.DefaultGenTxGas,
		"testing-chain-id",
		[]uint64{0},
		[]uint64{0},
		prik1,
	)
	_, err = decorator.AnteHandle(ctx, tx2, false, mmd.AnteHandle)
	require.NoError(t, err)
}

func TestMinCommision_MsgEditValidator(t *testing.T) {
	txConfig := app.MakeEncodingConfig().TxConfig
	decorator := ante.NewMinCommissionDecorator(cdc)
	editValidatorMsg := staking.NewMsgEditValidator(valAddr1, description, &newCommissionRate1, &newSelfDelegation1)
	tx, err := helpers.GenSignedMockTx(
		rand.New(rand.NewSource(time.Now().UnixNano())),
		txConfig,
		[]sdk.Msg{editValidatorMsg},
		sdk.NewCoins(),
		helpers.DefaultGenTxGas,
		"testing-chain-id",
		[]uint64{0},
		[]uint64{0},
		prik1,
	)
	require.NoError(t, err)
	mmd := MockAnteHandler{}
	ctx := sdk.Context{}.WithIsCheckTx(true)
	_, err = decorator.AnteHandle(ctx, tx, false, mmd.AnteHandle)
	require.Error(t, err)
	require.Contains(t, err.Error(), "commission can't be lower than 5%")

	//edit success test
	editValidatorMsg2 := staking.NewMsgEditValidator(valAddr2, description, &newCommissionRate2, &newSelfDelegation2)
	tx2, err := helpers.GenSignedMockTx(
		rand.New(rand.NewSource(time.Now().UnixNano())),
		txConfig,
		[]sdk.Msg{editValidatorMsg2},
		sdk.NewCoins(),
		helpers.DefaultGenTxGas,
		"testing-chain-id",
		[]uint64{0},
		[]uint64{0},
		prik1,
	)
	_, err = decorator.AnteHandle(ctx, tx2, false, mmd.AnteHandle)
	require.NoError(t, err)
}
