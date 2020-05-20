package v0_8

import (
	"encoding/json"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	"github.com/stretchr/testify/require"

	"github.com/kava-labs/kava/app"
)

func TestMigrate_Auth(t *testing.T) {
	config := sdk.GetConfig()
	app.SetBech32AddressPrefixes(config)
	app.SetBip44CoinType(config)
	config.Seal()

	oldGenesisState := genutil.AppMap{
		"auth": []byte(`{
	"params": {
		"max_memo_characters": "256",
		"sig_verify_cost_ed25519": "590",
		"sig_verify_cost_secp256k1": "1000",
		"tx_sig_limit": "7",
		"tx_size_cost_per_byte": "10"
	},
	"accounts": [
		{
			"type": "cosmos-sdk/Account",
			"value": {
				"account_number": "0",
				"address": "kava12jk3szk45afmvjc3xc6kvj4e40tuy2m8ckgs03",
				"coins": [
					{
						"amount": "10000000000",
						"denom": "btc"
					},
					{
						"amount": "9000000000",
						"denom": "ukava"
					},
					{
						"amount": "90000000000",
						"denom": "xrp"
					}
				],
				"public_key": null,
				"sequence": "0"
			}
		}
	]
}`),
	}
	// TODO add other types of accounts

	// Note subtle changes in json marshalling of account sequence, account_number, public_key.
	// sdk v0.38 uses go's json marshaller for accounts, not amino's json marshaller.
	expectedAuthGenState := genutil.AppMap{
		"auth": json.RawMessage(`{
	"params": {
		"max_memo_characters": "256",
		"sig_verify_cost_ed25519": "590",
		"sig_verify_cost_secp256k1": "1000",
		"tx_sig_limit": "7",
		"tx_size_cost_per_byte": "10"
	},
	"accounts": [
		{
			"type": "cosmos-sdk/Account",
			"value": {
				"account_number": 0,
				"address": "kava12jk3szk45afmvjc3xc6kvj4e40tuy2m8ckgs03",
				"coins": [
					{
						"amount": "10000000000",
						"denom": "btc"
					},
					{
						"amount": "9000000000",
						"denom": "ukava"
					},
					{
						"amount": "90000000000",
						"denom": "xrp"
					}
				],
				"public_key": "",
				"sequence": 0
			}
		}
	]
}`),
	}

	newGenesisState := Migrate(oldGenesisState)

	require.JSONEq(t, string(expectedAuthGenState["auth"]), string(newGenesisState["auth"]))
}
