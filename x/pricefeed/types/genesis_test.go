package types

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	tmtypes "github.com/tendermint/tendermint/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestGenesisStateValidate(t *testing.T) {
	now := time.Now()
	mockPrivKey := tmtypes.NewMockPV()
	pubkey, err := mockPrivKey.GetPubKey()
	require.NoError(t, err)
	addr := sdk.AccAddress(pubkey.Address())

	testCases := []struct {
		msg          string
		genesisState GenesisState
		expPass      bool
	}{
		{
			msg:          "default",
			genesisState: DefaultGenesisState(),
			expPass:      true,
		},
		{
			msg: "valid genesis",
			genesisState: NewGenesisState(
				NewParams([]Market{
					{"market", "xrp", "bnb", []sdk.AccAddress{addr}, true},
				}),
				[]PostedPrice{NewPostedPrice("xrp", addr, sdk.OneDec(), now)},
			),
			expPass: true,
		},
		{
			msg: "invalid param",
			genesisState: NewGenesisState(
				NewParams([]Market{
					{"", "xrp", "bnb", []sdk.AccAddress{addr}, true},
				}),
				[]PostedPrice{NewPostedPrice("xrp", addr, sdk.OneDec(), now)},
			),
			expPass: false,
		},
		{
			msg: "dup market param",
			genesisState: NewGenesisState(
				NewParams([]Market{
					{"market", "xrp", "bnb", []sdk.AccAddress{addr}, true},
					{"market", "xrp", "bnb", []sdk.AccAddress{addr}, true},
				}),
				[]PostedPrice{NewPostedPrice("xrp", addr, sdk.OneDec(), now)},
			),
			expPass: false,
		},
		{
			msg: "invalid posted price",
			genesisState: NewGenesisState(
				NewParams([]Market{}),
				[]PostedPrice{NewPostedPrice("xrp", nil, sdk.OneDec(), now)},
			),
			expPass: false,
		},
		{
			msg: "duplicated posted price",
			genesisState: NewGenesisState(
				NewParams([]Market{}),
				[]PostedPrice{
					NewPostedPrice("xrp", addr, sdk.OneDec(), now),
					NewPostedPrice("xrp", addr, sdk.OneDec(), now),
				},
			),
			expPass: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.msg, func(t *testing.T) {
			err := tc.genesisState.Validate()
			if tc.expPass {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}
