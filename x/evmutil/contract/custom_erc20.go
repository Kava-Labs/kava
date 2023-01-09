package contract

import (
	"encoding/json"

	evmtypes "github.com/tharsis/ethermint/x/evm/types"
)

var (
	CustomERC20Contract evmtypes.CompiledContract
)

func init() {
	err := json.Unmarshal([]byte(jsonData), &CustomERC20Contract)
	if err != nil {
		panic(err)
	}

	if len(CustomERC20Contract.Bin) == 0 {
		panic("load contract failed")
	}
}
