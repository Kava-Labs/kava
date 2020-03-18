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
	h.BeforeAll(func(t []*trans.Transaction) {
		// fmt.Println("before all modification")

	})
	h.BeforeEach(func(t *trans.Transaction) {
		// fmt.Println("before each modification")
	})

	h.Before("Governance > /gov/proposals/{proposalId} > Query a proposal > 200 > application/json", func(t *trans.Transaction) {
		sendProposal()
	})

	h.Before("Governance > /gov/proposals/{proposalId}/deposits > Query deposits > 200 > application/json", func(t *trans.Transaction) {
		sendDeposit()
	})

	h.Before("Governance > /gov/proposals/{proposalId}/votes > Query voters > 200 > application/json", func(t *trans.Transaction) {
		sendVote()
	})

	h.Before("Transactions > /txs/{hash} > Get a Tx by hash > 200 > application/json", func(t *trans.Transaction) {
		sendCoins()
	})

	h.Before("Staking > /staking/delegators/{delegatorAddr}/delegations > Get all delegations from a delegator > 200 > application/json", func(t *trans.Transaction) {
		sendDelegation()
	})

	h.Before("Staking > /staking/delegators/{delegatorAddr}/unbonding_delegations/{validatorAddr} > Query all unbonding delegations between a delegator and a validator > 200 > application/json", func(t *trans.Transaction) {
		sendUndelegation()
	})

	h.BeforeEachValidation(func(t *trans.Transaction) {
		// fmt.Println("before each validation modification")
	})
	h.BeforeValidation("/message > GET", func(t *trans.Transaction) {
		// fmt.Println("before validation modification")
	})
	h.After("/message > GET", func(t *trans.Transaction) {
		// fmt.Println("after modification")
	})
	h.AfterEach(func(t *trans.Transaction) {
		// fmt.Println("after each modification")
	})
	h.AfterAll(func(t []*trans.Transaction) {
		// fmt.Println("after all modification")
	})
	server.Serve()
	defer server.Listener.Close()
}

func sendProposal() {
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

	// helper methods for transactions
	cdc := app.MakeCodec() // make codec for the app

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

	// the test address - TODO IMPORTANT make sure this lines up with startchain.sh
	address := "kava1ffv7nhd3z6sych2qpqkk03ec6hzkmufy0r2s4c"

	// SEND THE PROPOSAL

	// cast to the generic msg type
	msgToSend := []sdk.Msg{msg}
	keyname := "vlad"      // TODO - IMPORTANT this must match the keys in the startchain.sh script
	password := "password" // TODO - IMPORTANT this must match the keys in the startchain.sh script

	// send the PROPOSAL message to the blockchain
	sendMsgToBlockchain(cdc, address, keyname, password, msgToSend, keybase)
}

func sendDeposit() {
	addr, err := sdk.AccAddressFromBech32("kava1ffv7nhd3z6sych2qpqkk03ec6hzkmufy0r2s4c") // validator
	if err != nil {
		panic(err)
	}

	// helper methods for transactions
	cdc := app.MakeCodec() // make codec for the app

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

	// the test address - TODO IMPORTANT make sure this lines up with startchain.sh
	address := "kava1ffv7nhd3z6sych2qpqkk03ec6hzkmufy0r2s4c"

	keyname := "vlad"      // TODO - IMPORTANT this must match the keys in the startchain.sh script
	password := "password" // TODO - IMPORTANT this must match the keys in the startchain.sh script

	// NOW SEND THE DEPOSIT

	// create a deposit transaction to send to the proposal
	amount := sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 10000000))
	deposit := gov.NewMsgDeposit(addr, 2, amount) // TODO IMPORTANT '2' must match 'x-example' in swagger.yaml
	depositToSend := []sdk.Msg{deposit}

	sendMsgToBlockchain(cdc, address, keyname, password, depositToSend, keybase)

}

func sendVote() {
	addr, err := sdk.AccAddressFromBech32("kava1ffv7nhd3z6sych2qpqkk03ec6hzkmufy0r2s4c") // validator
	if err != nil {
		panic(err)
	}

	// helper methods for transactions
	cdc := app.MakeCodec() // make codec for the app

	// create a keybase
	// TODO - IMPORTANT - this needs to be set manually and does NOT work with tilde i.e. ~/ does NOT work
	keybase, err := keys.NewKeyBaseFromDir("/Users/john/.kvcli/")
	if err != nil {
		panic(err)
	}
	_, err = keybase.List()
	// fmt.Printf("Keys: %s\n\n", all)
	if err != nil {
		panic(err)
	}

	// the test address - TODO IMPORTANT make sure this lines up with startchain.sh
	address := "kava1ffv7nhd3z6sych2qpqkk03ec6hzkmufy0r2s4c"

	keyname := "vlad"      // TODO - IMPORTANT this must match the keys in the startchain.sh script
	password := "password" // TODO - IMPORTANT this must match the keys in the startchain.sh script

	// NOW SEND THE VOTE

	// create a vote on a proposal to send to the blockchain
	vote := gov.NewMsgVote(addr, uint64(2), types.OptionYes) // TODO IMPORTANT '2' must match 'x-example' in swagger.yaml
	fmt.Printf("\nvote:%s\n", vote)

	// send a vote to the blockchain
	voteToSend := []sdk.Msg{vote}
	sendMsgToBlockchain(cdc, address, keyname, password, voteToSend, keybase)

}

// this should send coins from one address to another
func sendCoins() {
	addrFrom, err := sdk.AccAddressFromBech32("kava1ffv7nhd3z6sych2qpqkk03ec6hzkmufy0r2s4c") // validator
	if err != nil {
		panic(err)
	}

	addrTo, err := sdk.AccAddressFromBech32("kava1ls82zzghsx0exkpr52m8vht5jqs3un0ceysshz") // faucet
	if err != nil {
		panic(err)
	}

	// helper methods for transactions
	cdc := app.MakeCodec() // make codec for the app

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

	// the test address - TODO IMPORTANT make sure this lines up with startchain.sh
	address := "kava1ffv7nhd3z6sych2qpqkk03ec6hzkmufy0r2s4c"

	// create coins
	amount := sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 2000000))

	coins := bank.NewMsgSend(addrFrom, addrTo, amount) // TODO IMPORTANT '2' must match 'x-example' in swagger.yaml
	coinsToSend := []sdk.Msg{coins}
	keyname := "vlad"      // TODO - IMPORTANT this must match the keys in the startchain.sh script
	password := "password" // TODO - IMPORTANT this must match the keys in the startchain.sh script

	// NOW SEND THE COINS

	// send the coin message to the blockchain
	sendMsgToBlockchain(cdc, address, keyname, password, coinsToSend, keybase)

}

// this should send a delegation
func sendDelegation() {
	keyname := "vlad"      // TODO - IMPORTANT this must match the keys in the startchain.sh script
	password := "password" // TODO - IMPORTANT this must match the keys in the startchain.sh script

	addrFrom, err := sdk.AccAddressFromBech32("kava1ffv7nhd3z6sych2qpqkk03ec6hzkmufy0r2s4c") // validator
	if err != nil {
		panic(err)
	}

	// helper methods for transactions
	cdc := app.MakeCodec() // make codec for the app

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

	// the test address - TODO IMPORTANT make sure this lines up with startchain.sh
	address := "kava1ffv7nhd3z6sych2qpqkk03ec6hzkmufy0r2s4c"

	// get the validator address for delegation
	valAddr, err := sdk.ValAddressFromBech32("kavavaloper1ffv7nhd3z6sych2qpqkk03ec6hzkmufyz4scd0") // faucet
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
	keyname := "vlad"      // TODO - IMPORTANT this must match the keys in the startchain.sh script
	password := "password" // TODO - IMPORTANT this must match the keys in the startchain.sh script

	addrFrom, err := sdk.AccAddressFromBech32("kava1ffv7nhd3z6sych2qpqkk03ec6hzkmufy0r2s4c") // validator
	if err != nil {
		panic(err)
	}

	// helper methods for transactions
	cdc := app.MakeCodec() // make codec for the app

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

	// the test address - TODO IMPORTANT make sure this lines up with startchain.sh
	address := "kava1ffv7nhd3z6sych2qpqkk03ec6hzkmufy0r2s4c"

	// get the validator address for delegation
	valAddr, err := sdk.ValAddressFromBech32("kavavaloper1ffv7nhd3z6sych2qpqkk03ec6hzkmufyz4scd0") // faucet
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
	// fmt.Printf("\n\naccount: %s\n\n", account)

	return account.GetAccountNumber(), account.GetSequence()

}
