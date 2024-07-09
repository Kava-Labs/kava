package types_test

import (
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"

	"github.com/kava-labs/kava/x/evmutil/types"
)

func TestGenesisState_Validate(t *testing.T) {
	tests := []struct {
		name    string
		success bool
		params  types.Params
	}{
		{
			name: "invalid params",
			params: types.NewParams(
				types.NewConversionPairs(
					types.NewConversionPair(types.NewInternalEVMAddress(common.HexToAddress("0xinvalidaddress")), "weth"),
				),
				types.NewAllowedCosmosCoinERC20Tokens(),
			),
			success: false,
		},
		{
			name:    "valid state",
			success: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gs := types.NewGenesisState(tt.params)
			err := gs.Validate()
			if tt.success {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}
