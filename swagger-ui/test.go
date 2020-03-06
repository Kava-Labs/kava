package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authrest "github.com/cosmos/cosmos-sdk/x/auth/client/rest"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
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
	msg := gov.NewMsgSubmitProposal(
		proposalContent,
		sdk.NewCoins(sdk.NewInt64Coin("kava", 1000)),
		addr,
	)

	tx := authtypes.NewStdTx([]sdk.Msg{msg}, authtypes.StdFee{}, []authtypes.StdSignature{}, "a test memo")

	cdc := app.MakeCodec()
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

	resp, err := http.Post("http://localhost:1317/txs", "application/json", bytes.NewBuffer(jsonBytes))
	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	fmt.Println(string(body))
}
