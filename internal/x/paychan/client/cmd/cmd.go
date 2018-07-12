package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	//"github.com/cosmos/cosmos-sdk/client/context"
	//sdk "github.com/cosmos/cosmos-sdk/types"
	//"github.com/cosmos/cosmos-sdk/wire"
	//"github.com/cosmos/cosmos-sdk/x/auth"
)

// list of functions that return pointers to cobra commands
// No local storage needed for cli acting as a sender

// create paychan
// close paychan
// get paychan(s)
// send paychan payment
// get balance from receiver

// minimum
// create paychan (sender signs)
// create state update (sender signs) (just a half signed close tx, (json encoded?))
// close paychan (receiver signs) (provide state update as arg)

// example from x/auth
/*
func GetAccountCmd(storeName string, cdc *wire.Codec, decoder auth.AccountDecoder) *cobra.Command {
	return &cobra.Command{
		Use:   "account [address]",
		Short: "Query account balance",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {

			// find the key to look up the account
			addr := args[0]

			key, err := sdk.GetAccAddressBech32(addr)
			if err != nil {
				return err
			}

			// perform query
			ctx := context.NewCoreContextFromViper()
			res, err := ctx.Query(auth.AddressStoreKey(key), storeName)
			if err != nil {
				return err
			}

			// Check if account was found
			if res == nil {
				return sdk.ErrUnknownAddress("No account with address " + addr +
					" was found in the state.\nAre you sure there has been a transaction involving it?")
			}

			// decode the value
			account, err := decoder(res)
			if err != nil {
				return err
			}

			// print out whole account
			output, err := wire.MarshalJSONIndent(cdc, account)
			if err != nil {
				return err
			}
			fmt.Println(string(output))
			return nil
		},
	}
}
*/
