package testutil

import (
	"fmt"
	"log"
	"os"

	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/go-bip39"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/tests/util"
)

type SigningAccount struct {
	name      string
	mnemonic  string
	signer    *util.Signer
	requests  chan<- util.MsgRequest
	responses <-chan util.MsgResponse

	Address sdk.AccAddress

	l *log.Logger
}

// GetAccount returns the account with the given name or fails
func (suite *E2eTestSuite) GetAccount(name string) *SigningAccount {
	acc, found := suite.accounts[name]
	if !found {
		suite.Failf("account does not exist", "failed to find account with name %s", name)
	}
	return acc
}

// AddNewSigningAccount sets up a new account with a signer.
func (suite *E2eTestSuite) AddNewSigningAccount(name string, hdPath *hd.BIP44Params, chainId, mnemonic string) *SigningAccount {
	if _, found := suite.accounts[name]; found {
		suite.Failf("can't create signing account", "account with name %s already exists", name)
	}

	privKeyBytes, err := hd.Secp256k1.Derive()(mnemonic, "", hdPath.String())
	suite.NoErrorf(err, "failed to derive private key from mnemonic for %s: %s", name, err)
	privKey := &secp256k1.PrivKey{Key: privKeyBytes}

	signer := util.NewSigner(
		chainId,
		suite.encodingConfig,
		suite.Auth,
		suite.Tx,
		privKey,
		100,
	)

	requests := make(chan util.MsgRequest)
	responses, err := signer.Run(requests)
	suite.NoErrorf(err, "failed to start signer for account %s: %s", name, err)

	logger := log.New(os.Stdout, fmt.Sprintf("[%s] ", name), log.LstdFlags)

	// TODO: authenticated eth client.
	suite.accounts[name] = &SigningAccount{
		name:      name,
		mnemonic:  mnemonic,
		signer:    signer,
		requests:  requests,
		responses: responses,
		l:         logger,

		Address: signer.Address(),
	}

	return suite.accounts[name]
}

// SignAndBroadcastKavaTx sends a request to the signer and awaits its response.
func (a *SigningAccount) SignAndBroadcastKavaTx(req util.MsgRequest) util.MsgResponse {
	a.l.Printf("broadcasting tx %+v\n", req.Data)
	// send the request to signer
	a.requests <- req

	// block and await response
	// response is not returned until the msg is committed to a block
	res := <-a.responses

	// error will be set if response is not Code 0 (success) or Code 19 (already in mempool)
	if res.Err != nil {
		a.l.Printf("response code: %d error: %s\n", res.Result.Code, res.Result.RawLog)
	} else {
		a.l.Printf("response code: %d, hash %s\n", res.Result.Code, res.Result.TxHash)
	}

	return res
}

func (suite *E2eTestSuite) NewFundedAccount(name string, funds sdk.Coins) *SigningAccount {
	entropy, err := bip39.NewEntropy(128)
	suite.NoErrorf(err, "failed to generate entropy for account %s: %s", name, err)
	mnemonic, err := bip39.NewMnemonic(entropy)
	suite.NoErrorf(err, "failed to create new mnemonic for account %s: %s", name, err)

	acc := suite.AddNewSigningAccount(
		name,
		hd.CreateHDPath(app.Bip44CoinType, 0, 0),
		ChainId,
		mnemonic,
	)

	whale := suite.GetAccount(FundedAccountName)
	res := whale.SignAndBroadcastKavaTx(
		util.MsgRequest{
			Msgs: []sdk.Msg{
				banktypes.NewMsgSend(whale.Address, acc.Address, funds),
			},
			GasLimit:  1e5,
			FeeAmount: sdk.NewCoins(sdk.NewCoin(StakingDenom, sdk.NewInt(75000))),
			Data:      fmt.Sprintf("initial funding of account %s", name),
		},
	)

	suite.NoErrorf(res.Err, "failed to fund new account %s: %s", name, res.Err)

	return acc
}
