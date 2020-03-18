package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/snikch/goodman/hooks"
	trans "github.com/snikch/goodman/transaction"
	"github.com/tendermint/go-amino"

	"github.com/cosmos/cosmos-sdk/client/keys"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/gov"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
	"github.com/kava-labs/kava/app"

	crkeys "github.com/cosmos/cosmos-sdk/crypto/keys"
	sdkrest "github.com/cosmos/cosmos-sdk/types/rest"
	authrest "github.com/cosmos/cosmos-sdk/x/auth/client/rest"
	authclient "github.com/cosmos/cosmos-sdk/x/auth/client/utils"
	authexported "github.com/cosmos/cosmos-sdk/x/auth/exported"
)

func init() {
	config := sdk.GetConfig()
	app.SetBech32AddressPrefixes(config)
	app.SetBip44CoinType(config)
	config.Seal()
}

func main() {
	h := hooks.NewHooks()
	server := hooks.NewServer(hooks.NewHooksRunner(h))

	h.Before("Governance > /gov/proposals/{proposalId} > Query a proposal > 200 > application/json", func(t *trans.Transaction) {
		// send a governance proposal
		sendProposal()
	})

	h.Before("Governance > /gov/proposals/{proposalId}/deposits > Query deposits > 200 > application/json", func(t *trans.Transaction) {
		// send a deposit to the governance proposal
		sendDeposit()
	})

	h.Before("Governance > /gov/proposals/{proposalId}/votes > Query voters > 200 > application/json", func(t *trans.Transaction) {
		// vote on the governance proposal
		sendVote()
	})

	h.Before("Transactions > /txs/{hash} > Get a Tx by hash > 200 > application/json", func(t *trans.Transaction) {
		// send a transaction to the chain
		sendCoins()
	})

	h.Before("Staking > /staking/delegators/{delegatorAddr}/delegations > Get all delegations from a delegator > 200 > application/json", func(t *trans.Transaction) {
		// send a delegation message
		sendDelegation()
	})

	h.Before("Staking > /staking/delegators/{delegatorAddr}/unbonding_delegations/{validatorAddr} > Query all unbonding delegations between a delegator and a validator > 200 > application/json", func(t *trans.Transaction) {
		// send an undelegation
		sendUndelegation()
	})

	server.Serve()
	defer server.Listener.Close()
}

// sendProposal sends a governance proposal to the blockchain
func sendProposal() {
	// get the address
	address := getTestAddress()
	// get the keyname and password
	keyname, password := getKeynameAndPassword()

	proposalContent := gov.ContentFromProposalType("A Test Title", "A test description on this proposal.", gov.ProposalTypeText)
	addr, err := sdk.AccAddressFromBech32(address) // validator address
	if err != nil {
		panic(err)
	}

	// create a message to send to the blockchain
	msg := gov.NewMsgSubmitProposal(
		proposalContent,
		sdk.NewCoins(sdk.NewInt64Coin("stake", 1000)),
		addr,
	)

	// helper methods for transactions
	cdc := app.MakeCodec() // make codec for the app

	// get the keybase
	keybase := getKeybase()

	// SEND THE PROPOSAL
	// cast to the generic msg type
	msgToSend := []sdk.Msg{msg}

	// send the PROPOSAL message to the blockchain
	sendMsgToBlockchain(cdc, address, keyname, password, msgToSend, keybase)
}

func sendDeposit() {
	// get the address
	address := getTestAddress()
	// get the keyname and password
	keyname, password := getKeynameAndPassword()

	addr, err := sdk.AccAddressFromBech32(address) // validator
	if err != nil {
		panic(err)
	}

	// helper methods for transactions
	cdc := app.MakeCodec() // make codec for the app

	// get the keybase
	keybase := getKeybase()

	// NOW SEND THE DEPOSIT

	// create a deposit transaction to send to the proposal
	amount := sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 10000000))
	deposit := gov.NewMsgDeposit(addr, 2, amount) // TODO IMPORTANT '2' must match 'x-example' in swagger.yaml
	depositToSend := []sdk.Msg{deposit}

	sendMsgToBlockchain(cdc, address, keyname, password, depositToSend, keybase)

}

func sendVote() {
	// get the address
	address := getTestAddress()
	// get the keyname and password
	keyname, password := getKeynameAndPassword()

	addr, err := sdk.AccAddressFromBech32(address) // validator
	if err != nil {
		panic(err)
	}

	// helper methods for transactions
	cdc := app.MakeCodec() // make codec for the app

	// get the keybase
	keybase := getKeybase()

	// NOW SEND THE VOTE

	// create a vote on a proposal to send to the blockchain
	vote := gov.NewMsgVote(addr, uint64(2), types.OptionYes) // TODO IMPORTANT '2' must match 'x-example' in swagger.yaml

	// send a vote to the blockchain
	voteToSend := []sdk.Msg{vote}
	sendMsgToBlockchain(cdc, address, keyname, password, voteToSend, keybase)

}

// this should send coins from one address to another
func sendCoins() {
	// get the address
	address := getTestAddress()
	// get the keyname and password
	keyname, password := getKeynameAndPassword()

	addrFrom, err := sdk.AccAddressFromBech32(address) // validator
	if err != nil {
		panic(err)
	}

	addrTo, err := sdk.AccAddressFromBech32("kava1ls82zzghsx0exkpr52m8vht5jqs3un0ceysshz") // TODO IMPORTANT this is the faucet address
	if err != nil {
		panic(err)
	}

	// helper methods for transactions
	cdc := app.MakeCodec() // make codec for the app

	// get the keybase
	keybase := getKeybase()

	// create coins
	amount := sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 2000000))

	coins := bank.NewMsgSend(addrFrom, addrTo, amount) // TODO IMPORTANT '2' must match 'x-example' in swagger.yaml
	coinsToSend := []sdk.Msg{coins}

	// NOW SEND THE COINS

	// send the coin message to the blockchain
	sendMsgToBlockchain(cdc, address, keyname, password, coinsToSend, keybase)

}

func getTestAddress() (address string) {
	// the test address - TODO IMPORTANT make sure this lines up with startchain.sh
	address = "kava1ffv7nhd3z6sych2qpqkk03ec6hzkmufy0r2s4c"
	return address
}

func getKeynameAndPassword() (keyname string, password string) {
	keyname = "vlad"      // TODO - IMPORTANT this must match the keys in the startchain.sh script
	password = "password" // TODO - IMPORTANT this must match the keys in the startchain.sh script
	return keyname, password
}

// this should send a delegation
func sendDelegation() {
	// get the address
	address := getTestAddress()
	// get the keyname and password
	keyname, password := getKeynameAndPassword()

	addrFrom, err := sdk.AccAddressFromBech32(address) // validator
	if err != nil {
		panic(err)
	}

	// helper methods for transactions
	cdc := app.MakeCodec() // make codec for the app

	// get the keybase
	keybase := getKeybase()

	// get the validator address for delegation
	valAddr, err := sdk.ValAddressFromBech32("kavavaloper1ffv7nhd3z6sych2qpqkk03ec6hzkmufyz4scd0") // **FAUCET**
	if err != nil {
		panic(err)
	}

	// create delegation amount
	delAmount := sdk.NewInt64Coin(sdk.DefaultBondDenom, 1000000)
	delegation := staking.NewMsgDelegate(addrFrom, valAddr, delAmount)
	delegationToSend := []sdk.Msg{delegation}

	// send the delegation to the blockchain
	sendMsgToBlockchain(cdc, address, keyname, password, delegationToSend, keybase)
}

// this should send a MsgUndelegate
func sendUndelegation() {
	// get the address
	address := getTestAddress()
	// get the keyname and password
	keyname, password := getKeynameAndPassword()

	addrFrom, err := sdk.AccAddressFromBech32(address) // validator
	if err != nil {
		panic(err)
	}

	// helper methods for transactions
	cdc := app.MakeCodec() // make codec for the app

	// get the keybase
	keybase := getKeybase()

	// get the validator address for delegation
	valAddr, err := sdk.ValAddressFromBech32("kavavaloper1ffv7nhd3z6sych2qpqkk03ec6hzkmufyz4scd0") // **FAUCET**
	if err != nil {
		panic(err)
	}

	// create delegation amount
	undelAmount := sdk.NewInt64Coin(sdk.DefaultBondDenom, 1000000)
	undelegation := staking.NewMsgUndelegate(addrFrom, valAddr, undelAmount)
	delegationToSend := []sdk.Msg{undelegation}

	// send the delegation to the blockchain
	sendMsgToBlockchain(cdc, address, keyname, password, delegationToSend, keybase)

}

func getKeybase() crkeys.Keybase {
	// create a keybase
	// TODO - IMPORTANT - this needs to be set manually and does NOT work with tilde i.e. ~/ does NOT work
	// TODO - QUESTION - should we read the path from a configuration file?
	keybase, err := keys.NewKeyBaseFromDir("/Users/john/.kvcli/")
	if err != nil {
		panic(err)
	}
	_, err = keybase.List()
	// fmt.Printf("Keys: %s\n\n", all)
	if err != nil {
		panic(err)
	}

	return keybase

}

// sendMsgToBlockchain sends a message to the blockchain via the rest api
func sendMsgToBlockchain(cdc *amino.Codec, address string, keyname string,
	password string, msg []sdk.Msg, keybase crkeys.Keybase) {

	// get the account number and sequence number
	accountNumber, sequenceNumber := getAccountNumberAndSequenceNumber(cdc, address)

	txBldr := auth.NewTxBuilderFromCLI().
		WithTxEncoder(authclient.GetTxEncoder(cdc)).WithChainID("testing").
		WithKeybase(keybase).WithAccountNumber(accountNumber).
		WithSequence(sequenceNumber)

		// build and sign the transaction
	// this is the *Amino* encoded version of the transaction
	// fmt.Printf("%+v", txBldr.Keybase())
	txBytes, err := txBldr.BuildAndSign("vlad", "password", msg)
	if err != nil {
		panic(err)
	}
	// fmt.Printf("txBytes: %s", txBytes)

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
	// fmt.Println("post body: ", string(jsonBytes))

	resp, err := http.Post("http://localhost:1317/txs", "application/json", bytes.NewBuffer(jsonBytes))
	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	fmt.Printf("\n\nBody:\n\n")
	fmt.Println(string(body))

}

// getAccountNumberAndSequenceNumber gets an account number and sequence number from the blockchain
func getAccountNumberAndSequenceNumber(cdc *amino.Codec, address string) (accountNumber uint64, sequenceNumber uint64) {

	// we need to setup the account number and sequence in order to have a valid transaction
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

	return account.GetAccountNumber(), account.GetSequence()

}
