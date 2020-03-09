package main

import (
	"fmt"
	"net/http"
	"bytes"

	"github.com/snikch/goodman/hooks"
	trans "github.com/snikch/goodman/transaction"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	authrest "github.com/cosmos/cosmos-sdk/x/auth/client/rest"
	"github.com/kava-labs/kava/app"
)

// TODO - NOTE should rename to 'main()' if you want to use this instead of 'main_hooks()'
func main_hooks() {
	h := hooks.NewHooks()
	server := hooks.NewServer(hooks.NewHooksRunner(h))
	h.BeforeAll(func(t []*trans.Transaction) {
		fmt.Println("before all modification")
	})
	h.BeforeEach(func(t *trans.Transaction) {
		fmt.Println("before each modification")
	})
	h.Before("/message > GET", func(t *trans.Transaction) {
		// Create Proposal
		// - MsgSubmitProposal
		// - Create TX
		// - Sign Tx
		// - Broadcast Tx

		msg := gov.NewMsgSubmitProposal(
			nil,
			sdk.NewCoins(sdk.NewInt64Coin("kava", 1000)),
			sdk.AccAddress{}
		)

		tx := authtypes.NewStdTx([]sdk.Msg{msg}, authtypes.StdFee{}, []authtypes.StdSignature{}, "a test memo")
		
		cdc := app.MakeCodec()
		jsonBytes, err := cdc.MarshalJSON(
			authrest.BroadcastReq{
				Tx: tx,
				Mode: "block",
			}
		)
		if err != nil {
			panic(err)
		}

		resp, err := http.Post("http://localhost:1317/txs", "application/json", bytes.NewBuffer(jsonBytes))
		if err != nil {
			panic(err)
		}
		

		fmt.Println("before modification")
	})
	h.BeforeEachValidation(func(t *trans.Transaction) {
		fmt.Println("before each validation modification")
	})
	h.BeforeValidation("/message > GET", func(t *trans.Transaction) {
		fmt.Println("before validation modification")
	})
	h.After("/message > GET", func(t *trans.Transaction) {
		fmt.Println("after modification")
	})
	h.AfterEach(func(t *trans.Transaction) {
		fmt.Println("after each modification")
	})
	h.AfterAll(func(t []*trans.Transaction) {
		fmt.Println("after all modification")
	})
	server.Serve()
	defer server.Listener.Close()
}
