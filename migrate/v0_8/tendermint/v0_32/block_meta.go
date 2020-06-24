package v032

import (
	tmtypes "github.com/tendermint/tendermint/types"
)

// BlockMeta contains meta information about a block - namely, it's ID and Header.
type BlockMeta struct {
	BlockID tmtypes.BlockID `json:"block_id"` // the block hash and partsethash
	Header  Header          `json:"header"`   // The block's Header
}
