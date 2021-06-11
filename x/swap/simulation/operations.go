package simulation

import (
	"github.com/cosmos/cosmos-sdk/x/simulation"

	"github.com/kava-labs/kava/x/swap/types"
)

var (
	noOpMsg = simulation.NoOpMsg(types.ModuleName)
)
