package kava

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/crypto/keys"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/client/utils"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/stretchr/testify/require"

	"github.com/kava-labs/kava/app"
)

func TestMultiTxs(t *testing.T) {
	/*
		- run kava chain with an account with coins
		- create many signed sequential txs (query chain)
		- submit them all one after another on sync mode
		- check they all get through, at least print out all results
	*/
	config := sdk.GetConfig()
	app.SetBech32AddressPrefixes(config)
	app.SetBip44CoinType(config)

	cdc := app.MakeCodec()
	txBldr := auth.NewTxBuilder(
		utils.GetTxEncoder(cdc),
		0, 0, 300000,
		1.0, false, "testing", "",
		sdk.Coins{}, sdk.DecCoins{},
	)
	mnemonic := "crash sort dwarf disease change advice attract clump avoid mobile clump right junior axis book fresh mask tube front require until face effort vault"
	keybase := keys.NewInMemory()
	bip39Password := ""
	keyPassword := "password"
	keyName := "faucet"
	_, err := keybase.CreateAccount(keyName, mnemonic, bip39Password, keyPassword, keys.CreateHDPath(0, 0).String(), keys.Secp256k1)
	require.NoError(t, err)

	txBldr = txBldr.WithKeybase(keybase)

	senderKeyName := "faucet"
	senderAddressString := "kava1adkm6svtzjsxxvg7g6rshg6kj9qwej8gwqadqd"
	senderAddress, err := sdk.AccAddressFromBech32(senderAddressString)
	require.NoError(t, err)
	validatorAddressString := "kava1ffv7nhd3z6sych2qpqkk03ec6hzkmufy0r2s4c"
	validatorAddress, err := sdk.AccAddressFromBech32(validatorAddressString)
	require.NoError(t, err)

	msg := bank.NewMsgSend(
		senderAddress,
		validatorAddress,
		sdk.NewCoins(sdk.NewInt64Coin("ukava", 1000000)),
	)

	cliCtx := context.CLIContext{}.WithNodeURI("tcp://localhost:26657").WithTrustNode(true)
	accNum, seq, err := authtypes.NewAccountRetriever(cliCtx).GetAccountNumberSequence(senderAddress)
	require.NoError(t, err)

	txBldr = txBldr.WithAccountNumber(accNum)
	var numTxsToSend uint64 = 100
	var txHashes []string
	for i := uint64(0); i < numTxsToSend; i++ {
		txBldr = txBldr.WithSequence(seq + i)
		txBytes, err := txBldr.BuildAndSign(senderKeyName, "password", []sdk.Msg{msg})
		require.NoError(t, err)

		// broadcast to a Tendermint node
		res, err := cliCtx.BroadcastTxSync(txBytes)
		require.NoError(t, err)
		require.EqualValues(t, res.Code, 0)
		txHashes = append(txHashes, res.TxHash)
		t.Logf("result: %+v", res)
	}
}
