package app

import (
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	ethermint "github.com/tharsis/ethermint/types"
)

var _ authkeeper.ProtoAccountConstructor = newProtoAccount

func newProtoAccount(currentBlockheight int64) authtypes.AccountI {
	if currentBlockheight < FixDefaultAccountUpgradeHeight {

		emptyCodeHash := crypto.Keccak256(nil)

		return &ethermint.EthAccount{
			BaseAccount: &authtypes.BaseAccount{},
			CodeHash:    common.BytesToHash(emptyCodeHash).String(),
		}
	}
	return &authtypes.BaseAccount{}
}
