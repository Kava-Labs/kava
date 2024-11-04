package util

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/go-bip39"
	"github.com/ethereum/go-ethereum/common"
)

const (
	MnemonicEntropyBits = 128
)

func SdkToEvmAddress(addr sdk.AccAddress) common.Address {
	return common.BytesToAddress(addr.Bytes())
}

func EvmToSdkAddress(addr common.Address) sdk.AccAddress {
	return sdk.AccAddress(addr.Bytes())
}

// RandomMnemonic generates a random BIP39 mnemonic from 128 bits of entropy
func RandomMnemonic() (string, error) {
	entropy, err := bip39.NewEntropy(MnemonicEntropyBits)
	if err != nil {
		return "", errorsmod.Wrap(err, "failed to generate entropy for new mnemonic")
	}
	mnemonic, err := bip39.NewMnemonic(entropy)
	if err != nil {
		return mnemonic, errorsmod.Wrap(err, "failed to create mnemonic")
	}
	return mnemonic, nil
}
