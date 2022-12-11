// Derived from https://github.com/tharsis/evmos/blob/0bfaf0db7be47153bc651e663176ba8deca960b5/contracts/erc20.go
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package types

import (
	// Embed ERC20 JSON files
	_ "embed"
	"encoding/json"

	"github.com/ethereum/go-ethereum/common"
	evmtypes "github.com/tharsis/ethermint/x/evm/types"
)

var (
	//go:embed contracts/CustomERC20.json
	CustomERC20JSON []byte

	// CustomERC20Contract is the compiled erc20 contract
	CustomERC20Contract evmtypes.CompiledContract

	// CustomERC20JSONAddress is the erc20 module address
	CustomERC20JSONAddress common.Address
)

func init() {
	CustomERC20JSONAddress = ModuleEVMAddress

	err := json.Unmarshal(CustomERC20JSON, &CustomERC20Contract)
	if err != nil {
		panic(err)
	}

	if len(CustomERC20Contract.Bin) == 0 {
		panic("load contract failed")
	}
}
