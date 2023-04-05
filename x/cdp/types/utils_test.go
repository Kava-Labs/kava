package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestSortableDecBytes(t *testing.T) {
	tests := []struct {
		d    sdk.Dec
		want []byte
	}{
		{sdk.NewDec(0), []byte("000000000000000000.000000000000000000")},
		{sdk.NewDec(1), []byte("000000000000000001.000000000000000000")},
		{sdk.MustNewDecFromStr("2.0"), []byte("000000000000000002.000000000000000000")},
		{sdk.MustNewDecFromStr("-2.0"), []byte("-000000000000000002.000000000000000000")},
		{sdk.NewDec(10), []byte("000000000000000010.000000000000000000")},
		{sdk.NewDec(12340), []byte("000000000000012340.000000000000000000")},
		{sdk.NewDecWithPrec(12340, 4), []byte("000000000000000001.234000000000000000")},
		{sdk.NewDecWithPrec(12340, 5), []byte("000000000000000000.123400000000000000")},
		{sdk.NewDecWithPrec(12340, 8), []byte("000000000000000000.000123400000000000")},
		{sdk.NewDecWithPrec(1009009009009009009, 17), []byte("000000000000000010.090090090090090090")},
		{sdk.NewDecWithPrec(-1009009009009009009, 17), []byte("-000000000000000010.090090090090090090")},
		{sdk.NewDec(1000000000000000000), []byte("max")},
		{sdk.NewDec(-1000000000000000000), []byte("--")},
	}
	for tcIndex, tc := range tests {
		assert.Equal(t, tc.want, SortableDecBytes(tc.d), "bad String(), index: %v", tcIndex)
	}

	assert.Panics(t, func() { SortableDecBytes(sdk.NewDec(1000000000000000001)) })
	assert.Panics(t, func() { SortableDecBytes(sdk.NewDec(-1000000000000000001)) })
}

func TestParseSortableDecBytes(t *testing.T) {
	tests := []struct {
		d    sdk.Dec
		want []byte
	}{
		{sdk.NewDec(0), []byte("000000000000000000.000000000000000000")},
		{sdk.NewDec(1), []byte("000000000000000001.000000000000000000")},
		{sdk.MustNewDecFromStr("2.0"), []byte("000000000000000002.000000000000000000")},
		{sdk.MustNewDecFromStr("-2.0"), []byte("-000000000000000002.000000000000000000")},
		{sdk.NewDec(10), []byte("000000000000000010.000000000000000000")},
		{sdk.NewDec(12340), []byte("000000000000012340.000000000000000000")},
		{sdk.NewDecWithPrec(12340, 4), []byte("000000000000000001.234000000000000000")},
		{sdk.NewDecWithPrec(12340, 5), []byte("000000000000000000.123400000000000000")},
		{sdk.NewDecWithPrec(12340, 8), []byte("000000000000000000.000123400000000000")},
		{sdk.NewDecWithPrec(1009009009009009009, 17), []byte("000000000000000010.090090090090090090")},
		{sdk.NewDecWithPrec(-1009009009009009009, 17), []byte("-000000000000000010.090090090090090090")},
		{sdk.NewDec(1000000000000000000), []byte("max")},
		{sdk.NewDec(-1000000000000000000), []byte("--")},
	}
	for tcIndex, tc := range tests {
		b := SortableDecBytes(tc.d)
		r, err := ParseDecBytes(b)
		assert.NoError(t, err)
		assert.Equal(t, tc.d, r, "bad Dec(), index: %v", tcIndex)
	}
}

func TestRelativePow(t *testing.T) {
	tests := []struct {
		args []sdkmath.Int
		want sdkmath.Int
	}{
		{[]sdkmath.Int{sdk.ZeroInt(), sdk.ZeroInt(), sdk.OneInt()}, sdk.OneInt()},
		{[]sdkmath.Int{sdk.ZeroInt(), sdk.ZeroInt(), sdkmath.NewInt(10)}, sdkmath.NewInt(10)},
		{[]sdkmath.Int{sdk.ZeroInt(), sdk.OneInt(), sdkmath.NewInt(10)}, sdk.ZeroInt()},
		{[]sdkmath.Int{sdkmath.NewInt(10), sdkmath.NewInt(2), sdk.OneInt()}, sdkmath.NewInt(100)},
		{[]sdkmath.Int{sdkmath.NewInt(210), sdkmath.NewInt(2), sdkmath.NewInt(100)}, sdkmath.NewInt(441)},
		{[]sdkmath.Int{sdkmath.NewInt(2100), sdkmath.NewInt(2), sdkmath.NewInt(1000)}, sdkmath.NewInt(4410)},
		{[]sdkmath.Int{sdkmath.NewInt(1000000001547125958), sdkmath.NewInt(600), sdkmath.NewInt(1000000000000000000)}, sdkmath.NewInt(1000000928276004850)},
	}
	for i, tc := range tests {
		res := RelativePow(tc.args[0], tc.args[1], tc.args[2])
		require.Equal(t, tc.want, res, "unexpected result for test case %d, input: %v, got: %v", i, tc.args, res)
	}
}
