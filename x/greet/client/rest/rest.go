package rest

import (
	"fmt"
	"net/http"

	// "strconv"
	// "strings"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/kava-labs/kava/x/greet/types"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/rest"
	"github.com/gorilla/mux"
)

type CreatGreetReq struct {
	BaseReq        rest.BaseReq   `json:"base_req" yaml:"base_req"`
	GreetMessage 	   string         `json:"greetMessage" yaml:greetMessage`
}

func registerQueryRoutes(cliCtx client.Context, r *mux.Router) {
	r.HandleFunc("/greetings", getGreetingsHandler(cliCtx)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/greetings/{%s}",types.QueryGetGreeting), getGreetingHandler(cliCtx)).Methods("GET")
}


func registerTxRoutes(cliCtx client.Context, r*mux.Router) { 
	r.HandleFunc("/greeting", postCreateGreeting(cliCtx)).Methods("POST")
}

func getGreetingsHandler(cliCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}
		res, height, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s", types.ModuleName, types.QueryListGreetings), nil)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		cliCtx = cliCtx.WithHeight(height)
		rest.PostProcessResponse(w, cliCtx, res)
	}
}
// query 
func getGreetingHandler(cliCtx client.Context) http.HandlerFunc{
	return func(w http.ResponseWriter, r *http.Request) {
	
		_, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}
		vars := mux.Vars(r)
		gid := vars[types.QueryGetGreeting]
		var q = types.QueryGetGreetRequest{Id: gid}
		bz, err := cliCtx.LegacyAmino.MarshalJSON(q)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}
		
		res, height, err := cliCtx.QueryWithData(fmt.Sprintf("custom/greet/%s", types.QueryGetGreeting), bz)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}
		
		cliCtx = cliCtx.WithHeight(height)
		rest.PostProcessResponse(w, cliCtx, res)
		}
}

func postCreateGreeting(cliCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req CreatGreetReq
		if !rest.ReadRESTReq(w, r, cliCtx.LegacyAmino, &req) {
			rest.WriteErrorResponse(w, http.StatusBadRequest, "failed to parse request")
			return
		}

		baseReq := req.BaseReq.Sanitize()
		if !baseReq.ValidateBasic(w) {
			return
		}

		_, err := sdk.AccAddressFromBech32(baseReq.From)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}
		
		var greet = types.NewMsgCreateGreet(baseReq.From, req.GreetMessage)

		if err := greet.ValidateBasic(); err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		tx.WriteGeneratedTxResponse(cliCtx, w, baseReq, greet)

	
	}
}

func RegisterRoutes(cliCtx client.Context, r *mux.Router) {
	registerQueryRoutes(cliCtx, r)
	registerTxRoutes(cliCtx, r)
}

