package app

import (
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/ethereum/go-ethereum/common"
	ethermint "github.com/tharsis/ethermint/types"
	evmtypes "github.com/tharsis/ethermint/x/evm/types"
)

var _ authkeeper.ProtoAccountConstructor = newProtoAccount

func newProtoAccount(currentBlockheight int64) authtypes.AccountI {
	if currentBlockheight < FixDefaultAccountUpgradeHeight {

		return &ethermint.EthAccount{
			BaseAccount: &authtypes.BaseAccount{},
			CodeHash:    common.BytesToHash(evmtypes.EmptyCodeHash).String(),
		}
	}
	return &authtypes.BaseAccount{}
}
