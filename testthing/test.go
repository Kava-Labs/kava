package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/cosmos/cosmos-sdk/client/keys"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkrest "github.com/cosmos/cosmos-sdk/types/rest"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authrest "github.com/cosmos/cosmos-sdk/x/auth/client/rest"
	authclient "github.com/cosmos/cosmos-sdk/x/auth/client/utils"
	authexported "github.com/cosmos/cosmos-sdk/x/auth/exported"
	"github.com/cosmos/cosmos-sdk/x/gov"
	"github.com/kava-labs/kava/app"
)

func init() {
	config := sdk.GetConfig()
	app.SetBech32AddressPrefixes(config)
	app.SetBip44CoinType(config)
	config.Seal()
}

func main() {
	proposalContent := gov.ContentFromProposalType("A Test Title", "A test description on this proposal.", gov.ProposalTypeText)
	addr, err := sdk.AccAddressFromBech32("kava1ffv7nhd3z6sych2qpqkk03ec6hzkmufy0r2s4c") // validator
	if err != nil {
		panic(err)
	}

	// create a message to send to the blockchain
	msg := gov.NewMsgSubmitProposal(
		proposalContent,
		sdk.NewCoins(sdk.NewInt64Coin("stake", 1000)),
		addr,
	)

	// creating a transaction
	//	tx := authtypes.NewStdTx([]sdk.Msg{msg}, authtypes.StdFee{}, []authtypes.StdSignature{}, "a test memo")

	// helper methods for transactions
	cdc := app.MakeCodec() // make codec for the app
	// transaction builder
	// create a keybase
	keybase, err := keys.NewKeyBaseFromDir("/Users/john/.kvcli/")
	if err != nil {
		panic(err)
	}
	_, err = keybase.List()
	// fmt.Printf("Keys: %s\n\n", all)
	if err != nil {
		panic(err)
	}

	// we need to setup the account number and sequence in order to have a valid transaction
	address := "kava1ffv7nhd3z6sych2qpqkk03ec6hzkmufy0r2s4c"
	resp, err := http.Get("http://localhost:1317/auth/accounts/" + address)
	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	var bodyUnmarshalled sdkrest.ResponseWithHeight
	err = cdc.UnmarshalJSON(body, &bodyUnmarshalled)
	if err != nil {
		panic(err)
	}

	var account authexported.Account
	err = cdc.UnmarshalJSON(bodyUnmarshalled.Result, &account)
	if err != nil {
		panic(err)
	}
	fmt.Printf("\n\naccount: %s\n\n", account)

	txBldr := auth.NewTxBuilderFromCLI().
		WithTxEncoder(authclient.GetTxEncoder(cdc)).WithChainID("testing").
		WithKeybase(keybase).WithAccountNumber(account.GetAccountNumber()).
		WithSequence(account.GetSequence())

	// build and sign the transaction
	// this is the *Amino* encoded version of the transaction
	fmt.Printf("%+v", txBldr.Keybase())
	txBytes, err := txBldr.BuildAndSign("vlad", "password", []sdk.Msg{msg})
	if err != nil {
		panic(err)
	}
	fmt.Printf("txBytes: %s", txBytes)

	// need to convert the Amino encoded version back to an actual go struct
	var tx auth.StdTx
	cdc.UnmarshalBinaryLengthPrefixed(txBytes, &tx) // might be unmarshal binary bare

	// now we re-marshall it again into json
	jsonBytes, err := cdc.MarshalJSON(
		authrest.BroadcastReq{
			Tx:   tx,
			Mode: "block",
		},
	)
	if err != nil {
		panic(err)
	}
	fmt.Println("post body: ", string(jsonBytes))

	resp, err = http.Post("http://localhost:1317/txs", "application/json", bytes.NewBuffer(jsonBytes))
	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()
	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	fmt.Println(string(body))
}
