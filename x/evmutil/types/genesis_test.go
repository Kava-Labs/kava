package types_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/evmutil/types"
)

func TestGenesisState_Validate(t *testing.T) {
	_, addrs := app.GeneratePrivKeyAddressPairs(2)
	tests := []struct {
		name     string
		accounts []types.Account
		success  bool
	}{
		{
			"dup addresses",
			[]types.Account{
				{Address: addrs[0], Balance: sdk.NewInt(100)},
				{Address: addrs[0], Balance: sdk.NewInt(150)},
			},
			false,
		},
		{
			"empty account address",
			[]types.Account{
				{Balance: sdk.NewInt(100)},
			},
			false,
		},
		{
			"negative account balance",
			[]types.Account{
				{Address: addrs[0], Balance: sdk.NewInt(-100)},
			},
			false,
		},
		{
			"valid state",
			[]types.Account{
				{Address: addrs[0], Balance: sdk.NewInt(100)},
				{Address: addrs[1], Balance: sdk.NewInt(150)},
			},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gs := types.NewGenesisState(tt.accounts)
			err := gs.Validate()
			if tt.success {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}
