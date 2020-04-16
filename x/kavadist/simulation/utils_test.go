package simulation

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func TestApproxRoot(t *testing.T) {
	testCases := []struct {
		input    sdk.Dec
		root     uint64
		expected sdk.Dec
	}{
		{sdk.OneDec(), 10, sdk.OneDec()},                                                       // 1.0 ^ (0.1) => 1.0
		{sdk.NewDecWithPrec(25, 2), 2, sdk.NewDecWithPrec(5, 1)},                               // 0.25 ^ (0.5) => 0.5
		{sdk.NewDecWithPrec(4, 2), 2, sdk.NewDecWithPrec(2, 1)},                                // 0.04 => 0.2
		{sdk.NewDecFromInt(sdk.NewInt(27)), 3, sdk.NewDecFromInt(sdk.NewInt(3))},               // 27 ^ (1/3) => 3
		{sdk.NewDecFromInt(sdk.NewInt(-81)), 4, sdk.NewDecFromInt(sdk.NewInt(-3))},             // -81 ^ (0.25) => -3
		{sdk.NewDecFromInt(sdk.NewInt(2)), 2, sdk.NewDecWithPrec(1414213562373095049, 18)},     // 2 ^ (0.5) => 1.414213562373095049
		{sdk.NewDecWithPrec(1005, 3), 31536000, sdk.MustNewDecFromStr("1.000000000158153904")}, // 1.005 ^ (1/31536000) = 1.000000000158153904
	}

	for i, tc := range testCases {
		res, err := approxRoot(tc.input, tc.root)
		require.NoError(t, err)
		require.True(t, tc.expected.Sub(res).Abs().LTE(sdk.SmallestDec()), "unexpected result for test case %d, input: %v", i, tc.input)
	}
}
