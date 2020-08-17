package v032

import (
	"sync"
	"time"

	tmbytes "github.com/tendermint/tendermint/libs/bytes"
	tmtypes "github.com/tendermint/tendermint/types"
)

// Block defines the atomic unit of a Tendermint blockchain.
type Block struct {
	mtx        sync.Mutex
	Header     `json:"header"`
	Data       tmtypes.Data         `json:"data"`
	Evidence   tmtypes.EvidenceData `json:"evidence"`
	LastCommit *Commit              `json:"last_commit"` // not using for trust wallet
}

// Header defines the structure of a Tendermint block header.
type Header struct {
	// basic block info
	Version  Consensus `json:"version"`
	ChainID  string    `json:"chain_id"`
	Height   int64     `json:"height"`
	Time     time.Time `json:"time"`
	NumTxs   int64     `json:"num_txs"`
	TotalTxs int64     `json:"total_txs"`

	// prev block info
	LastBlockID tmtypes.BlockID `json:"last_block_id"`

	// hashes of block data
	LastCommitHash tmbytes.HexBytes `json:"last_commit_hash"` // commit from validators from the last block
	DataHash       tmbytes.HexBytes `json:"data_hash"`        // transactions

	// hashes from the app output from the prev block
	ValidatorsHash     tmbytes.HexBytes `json:"validators_hash"`      // validators for the current block
	NextValidatorsHash tmbytes.HexBytes `json:"next_validators_hash"` // validators for the next block
	ConsensusHash      tmbytes.HexBytes `json:"consensus_hash"`       // consensus params for current block
	AppHash            tmbytes.HexBytes `json:"app_hash"`             // state after txs from the previous block
	// root hash of all results from the txs from the previous block
	LastResultsHash tmbytes.HexBytes `json:"last_results_hash"`

	// consensus info
	EvidenceHash    tmbytes.HexBytes `json:"evidence_hash"`    // evidence included in the block
	ProposerAddress tmtypes.Address  `json:"proposer_address"` // original proposer of the block
}

// Commit contains the evidence that a block was committed by a set of validators.
type Commit struct {
}
