package v033

import (
	"time"

	v032tendermint "github.com/kava-labs/kava/migrate/v0_8/tendermint/v0_32"
	tmtypes "github.com/tendermint/tendermint/types"
)

func Migrate(v032GenDoc v032tendermint.GenesisDoc) tmtypes.GenesisDoc {

	// migrate evidence params

	newConsensusParams := tmtypes.ConsensusParams{
		Block: tmtypes.BlockParams(v032GenDoc.ConsensusParams.Block),
		Evidence: tmtypes.EvidenceParams{
			MaxAgeNumBlocks: v032GenDoc.ConsensusParams.Evidence.MaxAge,
			MaxAgeDuration:  time.Duration(int64(time.Second) * 6 * v032GenDoc.ConsensusParams.Evidence.MaxAge), // assume 6 second block times
		},
		Validator: tmtypes.ValidatorParams(v032GenDoc.ConsensusParams.Validator),
	}

	return tmtypes.GenesisDoc{
		GenesisTime:     v032GenDoc.GenesisTime,
		ChainID:         v032GenDoc.ChainID,
		ConsensusParams: &newConsensusParams,
		Validators:      v032GenDoc.Validators,
		AppHash:         v032GenDoc.AppHash,
		AppState:        v032GenDoc.AppState,
	}
}
