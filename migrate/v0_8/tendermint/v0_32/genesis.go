package v032

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"time"

	//"github.com/pkg/errors" // replaced this pkg with "errors" to avoid adding a dependency
	tmbytes "github.com/tendermint/tendermint/libs/bytes"
	"github.com/tendermint/tendermint/types"
	tmtime "github.com/tendermint/tendermint/types/time"
)

const (
	// MaxChainIDLen is a maximum length of the chain ID.
	MaxChainIDLen = 50
)

// GenesisDoc defines the initial conditions for a tendermint blockchain, in particular its validator set.
type GenesisDoc struct {
	GenesisTime     time.Time                `json:"genesis_time"`
	ChainID         string                   `json:"chain_id"`
	ConsensusParams *ConsensusParams         `json:"consensus_params,omitempty"`
	Validators      []types.GenesisValidator `json:"validators,omitempty"` // v0.33 GenesisValidator is backwards compatible with v0.32
	AppHash         tmbytes.HexBytes         `json:"app_hash"`             // moved from `common` to `bytes` as they are the same between v0.32 and v0.33
	AppState        json.RawMessage          `json:"app_state,omitempty"`
}

// ValidateAndComplete checks that all necessary fields are present
// and fills in defaults for optional fields left empty
func (genDoc *GenesisDoc) ValidateAndComplete() error {
	if genDoc.ChainID == "" {
		return errors.New("Genesis doc must include non-empty chain_id")
	}
	if len(genDoc.ChainID) > MaxChainIDLen {
		return fmt.Errorf("chain_id in genesis doc is too long (max: %d)", MaxChainIDLen) // replaced errors with fmt
	}

	if genDoc.ConsensusParams == nil {
		genDoc.ConsensusParams = DefaultConsensusParams()
	} /*else if err := genDoc.ConsensusParams.Validate(); err != nil {
		return err
	}*/ // remove validation to avoid having to copy in more types and methods from v0.32

	for i, v := range genDoc.Validators {
		if v.Power == 0 {
			return fmt.Errorf("The genesis file cannot contain validators with no voting power: %v", v) // replaced errors with fmt
		}
		if len(v.Address) > 0 && !bytes.Equal(v.PubKey.Address(), v.Address) {
			return fmt.Errorf("Incorrect address for validator %v in the genesis file, should be %v", v, v.PubKey.Address()) // replaced errors with fmt
		}
		if len(v.Address) == 0 {
			genDoc.Validators[i].Address = v.PubKey.Address()
		}
	}

	if genDoc.GenesisTime.IsZero() {
		genDoc.GenesisTime = tmtime.Now()
	}

	return nil
}

//------------------------------------------------------------
// Make genesis state from file

// GenesisDocFromJSON unmarshalls JSON data into a GenesisDoc.
func GenesisDocFromJSON(jsonBlob []byte) (*GenesisDoc, error) {
	genDoc := GenesisDoc{}
	err := Cdc.UnmarshalJSON(jsonBlob, &genDoc)
	if err != nil {
		return nil, err
	}

	if err := genDoc.ValidateAndComplete(); err != nil {
		return nil, err
	}

	return &genDoc, err
}

// GenesisDocFromFile reads JSON data from a file and unmarshalls it into a GenesisDoc.
func GenesisDocFromFile(genDocFile string) (*GenesisDoc, error) {
	jsonBlob, err := ioutil.ReadFile(genDocFile)
	if err != nil {
		return nil, fmt.Errorf("Couldn't read GenesisDoc file: %w", err)
	}
	genDoc, err := GenesisDocFromJSON(jsonBlob)
	if err != nil {
		return nil, fmt.Errorf("Error reading GenesisDoc at %v: %w", genDocFile, err)
	}
	return genDoc, nil
}
