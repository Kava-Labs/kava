package v18de63

import (
	"github.com/tendermint/tendermint/crypto"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type BaseAccount struct {
	Address       sdk.AccAddress `json:"address" yaml:"address"`
	Coins         sdk.Coins      `json:"coins" yaml:"coins"`
	PubKey        crypto.PubKey  `json:"public_key" yaml:"public_key"`
	AccountNumber uint64         `json:"account_number" yaml:"account_number"`
	Sequence      uint64         `json:"sequence" yaml:"sequence"`
}
