package rest_v0_3

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/rest"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authexported "github.com/cosmos/cosmos-sdk/x/auth/exported"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	"github.com/cosmos/cosmos-sdk/x/supply"

	v18de63auth "github.com/kava-labs/kava/migrate/v0_8/sdk/auth/v18de63"
	v18de63supply "github.com/kava-labs/kava/migrate/v0_8/sdk/supply/v18de63"
	valvesting "github.com/kava-labs/kava/x/validator-vesting"
	v0_3valvesting "github.com/kava-labs/kava/x/validator-vesting/legacy/v0_3"
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

		// convert v0.8 account type into old v0.3 account type so that it json marshals into the v0.3 format
		oldAccount := rollbackAccountType(account)
		// use old codec with old account interface registered
		cliCtx.WithCodec(makeCodecV03())

		if err != nil {
			if err := accGetter.EnsureExists(addr); err != nil {
				cliCtx = cliCtx.WithHeight(height)
				rest.PostProcessResponse(w, cliCtx, v18de63auth.BaseAccount{})
				return
			}

			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		cliCtx = cliCtx.WithHeight(height)
		rest.PostProcessResponse(w, cliCtx, oldAccount)
	}
}

func makeCodecV03() *codec.Codec {
	v0_3Codec := codec.New()
	codec.RegisterCrypto(v0_3Codec)
	v18de63auth.RegisterCodec(v0_3Codec)
	v18de63auth.RegisterCodecVesting(v0_3Codec)
	v18de63supply.RegisterCodec(v0_3Codec)
	v0_3valvesting.RegisterCodec(v0_3Codec)
	return v0_3Codec
}
func rollbackAccountType(newAccount authexported.Account) v18de63auth.Account {
	switch acc := newAccount.(type) {

	case *auth.BaseAccount:
		return v18de63auth.BaseAccount(*acc)

	case *vestingtypes.PeriodicVestingAccount:
		ba := v18de63auth.BaseAccount(*(acc.BaseVestingAccount.BaseAccount))
		bva := v18de63auth.BaseVestingAccount{
			BaseAccount:      &ba,
			OriginalVesting:  acc.BaseVestingAccount.OriginalVesting,
			DelegatedFree:    acc.BaseVestingAccount.DelegatedFree,
			DelegatedVesting: acc.BaseVestingAccount.DelegatedVesting,
			EndTime:          acc.BaseVestingAccount.EndTime,
		}
		var newPeriods v18de63auth.Periods
		for _, p := range acc.VestingPeriods {
			newPeriods = append(newPeriods, v18de63auth.Period(p))
		}
		pva := v18de63auth.PeriodicVestingAccount{
			BaseVestingAccount: &bva,
			StartTime:          acc.StartTime,
			VestingPeriods:     newPeriods,
		}
		return pva

	case *valvesting.ValidatorVestingAccount:
		ba := v18de63auth.BaseAccount(*(acc.PeriodicVestingAccount.BaseVestingAccount.BaseAccount))
		bva := v18de63auth.BaseVestingAccount{
			BaseAccount:      &ba,
			OriginalVesting:  acc.PeriodicVestingAccount.BaseVestingAccount.OriginalVesting,
			DelegatedFree:    acc.PeriodicVestingAccount.BaseVestingAccount.DelegatedFree,
			DelegatedVesting: acc.PeriodicVestingAccount.BaseVestingAccount.DelegatedVesting,
			EndTime:          acc.PeriodicVestingAccount.BaseVestingAccount.EndTime,
		}
		var newPeriods v18de63auth.Periods
		for _, p := range acc.PeriodicVestingAccount.VestingPeriods {
			newPeriods = append(newPeriods, v18de63auth.Period(p))
		}
		pva := v18de63auth.PeriodicVestingAccount{
			BaseVestingAccount: &bva,
			StartTime:          acc.PeriodicVestingAccount.StartTime,
			VestingPeriods:     newPeriods,
		}
		var newVestingProgress []v0_3valvesting.VestingProgress
		for _, p := range acc.VestingPeriodProgress {
			newVestingProgress = append(newVestingProgress, v0_3valvesting.VestingProgress(p))
		}
		vva := v0_3valvesting.ValidatorVestingAccount{
			PeriodicVestingAccount: &pva,
			ValidatorAddress:       acc.ValidatorAddress,
			ReturnAddress:          acc.ReturnAddress,
			SigningThreshold:       acc.SigningThreshold,
			CurrentPeriodProgress:  v0_3valvesting.CurrentPeriodProgress(acc.CurrentPeriodProgress),
			VestingPeriodProgress:  newVestingProgress,
			DebtAfterFailedVesting: acc.DebtAfterFailedVesting,
		}
		return vva

	case supply.ModuleAccount:
		ba := v18de63auth.BaseAccount(*(acc.BaseAccount))
		ma := v18de63supply.ModuleAccount{
			BaseAccount: &ba,
			Name:        acc.Name,
			Permissions: acc.Permissions,
		}
		return ma

	case nil:
		return acc

	default:
		panic(fmt.Errorf("unrecognized account type %+v", acc))
	}
}
