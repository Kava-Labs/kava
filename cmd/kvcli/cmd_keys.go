package main

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/keys"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/kava-labs/kava/app"
)

/*
NOTE TO FUTURE IMPLEMENTERS
This monkey patches the sdk `keys` command, therefore needs to be reviewed on any sdk updates.

The patch adds support for using kava's legacy bip44 coin type to the cli.
Coin types are used to create a bip44 derivation path, which is used as a mapping from mnemonic to private key.

In cosmos-sdk v0.38.3, all private keys are stored without reference to the mnemonic or bip44 derivation path, except ledger keys.
Ledger keys are just references to a private key on a ledger device. They contain the bip44 derivation path.
To patch the cli, we only need to modify:
- when new ledger references are created
- anything to do with converting a mnemonic to a private key.

These only happen in `kvcli keys add` cmd.
For private key generation, use a --legacy-hd-path flag to enable old coin type.
The current cosmos ledger app (v1.5.3) only supports the legacy coin type. So we only need to ensure ledger reference creation doesn't use the new coin type.

Signing txs:
- with local keys
	- the stored the priv key is used to sign, mnemonics or bip44 paths not involved
- with ledger
	- the stored bip44 path is used to instruct the ledger which key to sign with
*/

const flagLegacyHDPath = "legacy-hd-path"
const flagHDPath = "hd-path" // this is copied from keys add cmd because it's not exported

// getModifiedKeysCmd returns the standard cosmos-sdk/client/keys cmd but modified to support new and old bip44 coin types supported by kava.
func getModifiedKeysCmd() *cobra.Command {
	keysCmd := keys.Commands()
	for _, c := range keysCmd.Commands() {
		if c.Name() == "add" {
			monkeyPatchCmdKeysAdd(c)
			break
		}
	}
	return keysCmd
}

// monkeyPatchCmdKeysAdd modifies the `keys add` command to use the old bip44 coin type when a flag is passed.
func monkeyPatchCmdKeysAdd(keysAddCmd *cobra.Command) {
	// add flag
	keysAddCmd.Flags().Bool(flagLegacyHDPath, false, fmt.Sprintf("Use the old bip44 coin type (%d) to derive addresses from mnemonics.", sdk.CoinType))

	// replace description
	keysAddCmd.Long = fmt.Sprintf(`Derive a new private key and encrypt to disk.
	Optionally specify a BIP39 mnemonic, a BIP39 passphrase to further secure the mnemonic,
	and BIP44 account/index numbers to derive a specific key. The key will be stored under the given name
	and encrypted with the given password.

	NOTE: This cli defaults to Kava's BIP44 coin type %d. Use the --%s flag to use the old one (%d).
	
	The flag --recover allows one to recover a key from a seed passphrase.
	If run with --dry-run, a key would be generated (or recovered) but not stored to the
	local keystore.
	Use the --pubkey flag to add arbitrary public keys to the keystore for constructing
	multisig transactions.
	
	You can add a multisig key by passing the list of key names you want the public
	key to be composed of to the --multisig flag and the minimum number of signatures
	required through --multisig-threshold. The keys are sorted by address, unless
	the flag --nosort is set.
	`, app.Bip44CoinType, flagLegacyHDPath, sdk.CoinType)

	// replace the run function with a wrapped version that sets the old coin type in the global config
	oldRun := keysAddCmd.RunE
	keysAddCmd.RunE = func(cmd *cobra.Command, args []string) error {

		if !viper.GetBool(flagLegacyHDPath) && viper.GetBool(flags.FlagUseLedger) {
			return fmt.Errorf("cosmos ledger app only supports legacy bip44 coin type, must use --%s flag when adding ledger key", flagLegacyHDPath)
		}
		if viper.GetBool(flagLegacyHDPath) && viper.IsSet(flagHDPath) {
			return fmt.Errorf("cannot use a custom hd path (--%s) and legacy bip44 coin type (--%s) at the same time", flagHDPath, flagLegacyHDPath)
		}
		if viper.GetBool(flagLegacyHDPath) {
			preExistingCoinType := sdk.GetConfig().GetCoinType()
			sdk.GetConfig().SetCoinType(sdk.CoinType) // set old coin type
			err := oldRun(cmd, args)
			sdk.GetConfig().SetCoinType(preExistingCoinType) // revert to pre-existing coin type
			return err
		}
		return oldRun(cmd, args)
	}
}
