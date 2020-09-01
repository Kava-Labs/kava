package rpccore

import (
	types "github.com/kava-labs/kava/migrate/v0_8/tendermint/v0_32"
)

// Single block (with meta)
type ResultBlock struct {
	BlockMeta *types.BlockMeta `json:"block_meta"`
	Block     *types.Block     `json:"block"`
}
