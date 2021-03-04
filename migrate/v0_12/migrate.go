package v0_12

import (
	"time"

	tmtypes "github.com/tendermint/tendermint/types"
)

var (
	GenesisTime = time.Date(2021, 3, 5, 6, 0, 0, 0, time.UTC)
	ChainID     = "kava-6"
)

func Migrate(genDoc tmtypes.GenesisDoc) tmtypes.GenesisDoc {
	genDoc.GenesisTime = GenesisTime
	genDoc.ChainID = ChainID
	return genDoc
}
