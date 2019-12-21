package types

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

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
		args []sdk.Int
		want sdk.Int
	}{
		{[]sdk.Int{sdk.ZeroInt(), sdk.ZeroInt(), sdk.OneInt()}, sdk.OneInt()},
		{[]sdk.Int{sdk.ZeroInt(), sdk.ZeroInt(), sdk.NewInt(10)}, sdk.NewInt(10)},
		{[]sdk.Int{sdk.ZeroInt(), sdk.OneInt(), sdk.NewInt(10)}, sdk.ZeroInt()},
		{[]sdk.Int{sdk.NewInt(10), sdk.NewInt(2), sdk.OneInt()}, sdk.NewInt(100)},
		{[]sdk.Int{sdk.NewInt(210), sdk.NewInt(2), sdk.NewInt(100)}, sdk.NewInt(441)},
		{[]sdk.Int{sdk.NewInt(2100), sdk.NewInt(2), sdk.NewInt(1000)}, sdk.NewInt(4410)},
		{[]sdk.Int{sdk.NewInt(1000000001547125958), sdk.NewInt(600), sdk.NewInt(1000000000000000000)}, sdk.NewInt(1000000928276004850)},
	}
	for i, tc := range tests {
		res := RelativePow(tc.args[0], tc.args[1], tc.args[2])
		require.Equal(t, tc.want, res, "unexpected result for test case %d, input: %v, got: %v", i, tc.args, res)
	}
}
