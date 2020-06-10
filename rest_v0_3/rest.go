package rest_v0_3

import (
	"net/http"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/gorilla/mux"
	v18de63auth "github.com/kava-labs/kava/migrate/v0_8/sdk/auth/v18de63"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/rest"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authexported "github.com/cosmos/cosmos-sdk/x/auth/exported"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
)

func RegisterRoutes(cliCtx context.CLIContext, r *mux.Router) {
	r.HandleFunc(
		"/v0_3/auth/accounts/{address}", QueryAccountRequestHandlerFn(cliCtx),
	).Methods("GET")
}

// QueryAccountRequestHandlerFn handle auth/accounts queries
// This function is identical to v0.8 except the queried account is cast to the v0.3 account type so it marshals in the old format.
func QueryAccountRequestHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		bech32addr := vars["address"]

		addr, err := sdk.AccAddressFromBech32(bech32addr)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		accGetter := types.NewAccountRetriever(cliCtx)

		account, height, err := accGetter.GetAccountWithHeight(addr)

		oldAccount := rollbackAccountType(account)

		if err != nil {
			// TODO: Handle more appropriately based on the error type.
			// Ref: https://github.com/cosmos/cosmos-sdk/issues/4923
			if err := accGetter.EnsureExists(addr); err != nil {
				cliCtx = cliCtx.WithHeight(height)
				rest.PostProcessResponse(w, cliCtx, v18de63auth.BaseAccount{}) // return empty v18de63 account
				return
			}

			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		cliCtx = cliCtx.WithHeight(height)
		rest.PostProcessResponse(w, cliCtx, oldAccount)
	}
}

func rollbackAccountType(newAccount authexported.Account) v18de63auth.Account {
	// big type switch over all the different types of accounts
	switch acc := newAccount.(type) {
	case *auth.BaseAccount:
		return v18de63auth.BaseAccount(*acc) // TODO pointers??
	// TODO case nil?
	default:
		panic("TODO")
		return nil
	}
}
