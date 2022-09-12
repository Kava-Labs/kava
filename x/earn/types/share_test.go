package types_test

import (
	"strings"
	"testing"

	"github.com/kava-labs/kava/x/earn/types"
	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	testDenom1 = "ukava"
	testDenom2 = "usdx"
)

func d(i int64) sdk.Dec {
	return sdk.NewDec(i)
}

type vaultShareTestSuite struct {
	suite.Suite
}

func TestVaultShareTestSuite(t *testing.T) {
	suite.Run(t, new(vaultShareTestSuite))
}

func (s *vaultShareTestSuite) TestNewVaultShareFromDec() {
	s.Require().NotPanics(func() {
		types.NewVaultShare(testDenom1, sdk.NewDec(5))
	})
	s.Require().NotPanics(func() {
		types.NewVaultShare(testDenom1, sdk.ZeroDec())
	})
	s.Require().NotPanics(func() {
		types.NewVaultShare(strings.ToUpper(testDenom1), sdk.NewDec(5))
	})
	s.Require().Panics(func() {
		types.NewVaultShare(testDenom1, sdk.NewDec(-5))
	})
}

func (s *vaultShareTestSuite) TestAddVaultShare() {
	vaultShareA1 := types.NewVaultShare(testDenom1, sdk.NewDecWithPrec(11, 1))
	vaultShareA2 := types.NewVaultShare(testDenom1, sdk.NewDecWithPrec(22, 1))
	vaultShareB1 := types.NewVaultShare(testDenom2, sdk.NewDecWithPrec(11, 1))

	// regular add
	res := vaultShareA1.Add(vaultShareA1)
	s.Require().Equal(vaultShareA2, res, "sum of shares is incorrect")

	// bad denom add
	s.Require().Panics(func() {
		vaultShareA1.Add(vaultShareB1)
	}, "expected panic on sum of different denoms")
}

func (s *vaultShareTestSuite) TestAddVaultShares() {
	one := sdk.NewDec(1)
	zero := sdk.NewDec(0)
	two := sdk.NewDec(2)

	cases := []struct {
		inputOne types.VaultShares
		inputTwo types.VaultShares
		expected types.VaultShares
	}{
		{
			types.VaultShares{
				{testDenom1, one},
				{testDenom2, one},
			},
			types.VaultShares{
				{testDenom1, one},
				{testDenom2, one},
			},
			types.VaultShares{
				{testDenom1, two},
				{testDenom2, two},
			},
		},
		{
			types.VaultShares{
				{testDenom1, zero},
				{testDenom2, one},
			},
			types.VaultShares{
				{testDenom1, zero},
				{testDenom2, zero},
			},
			types.VaultShares{
				{testDenom2, one},
			},
		},
		{
			types.VaultShares{
				{testDenom1, zero},
				{testDenom2, zero},
			},
			types.VaultShares{
				{testDenom1, zero},
				{testDenom2, zero},
			},
			types.VaultShares(nil),
		},
	}

	for tcIndex, tc := range cases {
		res := tc.inputOne.Add(tc.inputTwo...)
		s.Require().Equal(tc.expected, res, "sum of shares is incorrect, tc #%d", tcIndex)
	}
}

func (s *vaultShareTestSuite) TestFilteredZeroVaultShares() {
	cases := []struct {
		name     string
		input    types.VaultShares
		original string
		expected string
	}{
		{
			name: "all greater than zero",
			input: types.VaultShares{
				{"testa", sdk.NewDec(1)},
				{"testb", sdk.NewDec(2)},
				{"testc", sdk.NewDec(3)},
				{"testd", sdk.NewDec(4)},
				{"teste", sdk.NewDec(5)},
			},
			original: "1.000000000000000000testa,2.000000000000000000testb,3.000000000000000000testc,4.000000000000000000testd,5.000000000000000000teste",
			expected: "1.000000000000000000testa,2.000000000000000000testb,3.000000000000000000testc,4.000000000000000000testd,5.000000000000000000teste",
		},
		{
			name: "zero share in middle",
			input: types.VaultShares{
				{"testa", sdk.NewDec(1)},
				{"testb", sdk.NewDec(2)},
				{"testc", sdk.NewDec(0)},
				{"testd", sdk.NewDec(4)},
				{"teste", sdk.NewDec(5)},
			},
			original: "1.000000000000000000testa,2.000000000000000000testb,0.000000000000000000testc,4.000000000000000000testd,5.000000000000000000teste",
			expected: "1.000000000000000000testa,2.000000000000000000testb,4.000000000000000000testd,5.000000000000000000teste",
		},
		{
			name: "zero share end (unordered)",
			input: types.VaultShares{
				{"teste", sdk.NewDec(5)},
				{"testc", sdk.NewDec(3)},
				{"testa", sdk.NewDec(1)},
				{"testd", sdk.NewDec(4)},
				{"testb", sdk.NewDec(0)},
			},
			original: "5.000000000000000000teste,3.000000000000000000testc,1.000000000000000000testa,4.000000000000000000testd,0.000000000000000000testb",
			expected: "1.000000000000000000testa,3.000000000000000000testc,4.000000000000000000testd,5.000000000000000000teste",
		},
	}

	for _, tt := range cases {
		undertest := types.NewVaultShares(tt.input...)
		s.Require().Equal(tt.expected, undertest.String(), "NewVaultShares must return expected results")
		s.Require().Equal(tt.original, tt.input.String(), "input must be unmodified and match original")
	}
}

func (s *vaultShareTestSuite) TestIsValid() {
	tests := []struct {
		share      types.VaultShare
		expectPass bool
		msg        string
	}{
		{
			types.NewVaultShare("mytoken", sdk.NewDec(10)),
			true,
			"valid shares should have passed",
		},
		{
			types.VaultShare{Denom: "BTC", Amount: sdk.NewDec(10)},
			true,
			"valid uppercase denom",
		},
		{
			types.VaultShare{Denom: "Bitshare", Amount: sdk.NewDec(10)},
			true,
			"valid mixed case denom",
		},
		{
			types.VaultShare{Denom: "btc", Amount: sdk.NewDec(-10)},
			false,
			"negative amount",
		},
	}

	for _, tc := range tests {
		tc := tc
		if tc.expectPass {
			s.Require().True(tc.share.IsValid(), tc.msg)
		} else {
			s.Require().False(tc.share.IsValid(), tc.msg)
		}
	}
}

func (s *vaultShareTestSuite) TestSubVaultShare() {
	tests := []struct {
		share      types.VaultShare
		expectPass bool
		msg        string
	}{
		{
			types.NewVaultShare("mytoken", sdk.NewDec(20)),
			true,
			"valid shares should have passed",
		},
		{
			types.NewVaultShare("othertoken", sdk.NewDec(20)),
			false,
			"denom mismatch",
		},
		{
			types.NewVaultShare("mytoken", sdk.NewDec(9)),
			false,
			"negative amount",
		},
	}

	vaultShare := types.NewVaultShare("mytoken", sdk.NewDec(10))

	for _, tc := range tests {
		tc := tc
		if tc.expectPass {
			equal := tc.share.Sub(vaultShare)
			s.Require().Equal(equal, vaultShare, tc.msg)
		} else {
			s.Require().Panics(func() { tc.share.Sub(vaultShare) }, tc.msg)
		}
	}
}

func (s *vaultShareTestSuite) TestSubVaultShares() {
	tests := []struct {
		shares     types.VaultShares
		expectPass bool
		msg        string
	}{
		{
			types.NewVaultShares(types.NewVaultShare("mytoken", d(10)), types.NewVaultShare("btc", d(20)), types.NewVaultShare("eth", d(30))),
			true,
			"sorted shares should have passed",
		},
		{
			types.VaultShares{types.NewVaultShare("mytoken", d(10)), types.NewVaultShare("btc", d(20)), types.NewVaultShare("eth", d(30))},
			false,
			"unorted shares should panic",
		},
		{
			types.VaultShares{types.VaultShare{Denom: "BTC", Amount: sdk.NewDec(10)}, types.NewVaultShare("eth", d(15)), types.NewVaultShare("mytoken", d(5))},
			false,
			"invalid denoms",
		},
	}

	vaultShares := types.NewVaultShares(types.NewVaultShare("btc", d(10)), types.NewVaultShare("eth", d(15)), types.NewVaultShare("mytoken", d(5)))

	for _, tc := range tests {
		tc := tc
		if tc.expectPass {
			equal := tc.shares.Sub(vaultShares...)
			s.Require().Equal(equal, vaultShares, tc.msg)
		} else {
			s.Require().Panics(func() { tc.shares.Sub(vaultShares...) }, tc.msg)
		}
	}
}

func (s *vaultShareTestSuite) TestSortVaultShares() {
	good := types.VaultShares{
		types.NewVaultShare("gas", d(1)),
		types.NewVaultShare("mineral", d(1)),
		types.NewVaultShare("tree", d(1)),
	}
	empty := types.VaultShares{
		types.NewVaultShare("gold", d(0)),
	}
	badSort1 := types.VaultShares{
		types.NewVaultShare("tree", d(1)),
		types.NewVaultShare("gas", d(1)),
		types.NewVaultShare("mineral", d(1)),
	}
	badSort2 := types.VaultShares{ // both are after the first one, but the second and third are in the wrong order
		types.NewVaultShare("gas", d(1)),
		types.NewVaultShare("tree", d(1)),
		types.NewVaultShare("mineral", d(1)),
	}
	badAmt := types.VaultShares{
		types.NewVaultShare("gas", d(1)),
		types.NewVaultShare("tree", d(0)),
		types.NewVaultShare("mineral", d(1)),
	}
	dup := types.VaultShares{
		types.NewVaultShare("gas", d(1)),
		types.NewVaultShare("gas", d(1)),
		types.NewVaultShare("mineral", d(1)),
	}
	cases := []struct {
		name          string
		shares        types.VaultShares
		before, after bool // valid before/after sort
	}{
		{"valid shares", good, true, true},
		{"empty shares", empty, false, false},
		{"unsorted shares (1)", badSort1, false, true},
		{"unsorted shares (2)", badSort2, false, true},
		{"zero amount shares", badAmt, false, false},
		{"duplicate shares", dup, false, false},
	}

	for _, tc := range cases {
		s.Require().Equal(tc.before, tc.shares.IsValid(), "share validity is incorrect before sorting; %s", tc.name)
		tc.shares.Sort()
		s.Require().Equal(tc.after, tc.shares.IsValid(), "share validity is incorrect after sorting;  %s", tc.name)
	}
}

func (s *vaultShareTestSuite) TestVaultSharesValidate() {
	testCases := []struct {
		input        types.VaultShares
		expectedPass bool
	}{
		{types.VaultShares{}, true},
		{types.VaultShares{types.VaultShare{testDenom1, sdk.NewDec(5)}}, true},
		{types.VaultShares{types.VaultShare{testDenom1, sdk.NewDec(5)}, types.VaultShare{testDenom2, sdk.NewDec(100000)}}, true},
		{types.VaultShares{types.VaultShare{testDenom1, sdk.NewDec(-5)}}, false},
		{types.VaultShares{types.VaultShare{"BTC", sdk.NewDec(5)}}, true},
		{types.VaultShares{types.VaultShare{"0BTC", sdk.NewDec(5)}}, false},
		{types.VaultShares{types.VaultShare{testDenom1, sdk.NewDec(5)}, types.VaultShare{"B", sdk.NewDec(100000)}}, false},
		{types.VaultShares{types.VaultShare{testDenom1, sdk.NewDec(5)}, types.VaultShare{testDenom2, sdk.NewDec(-100000)}}, false},
		{types.VaultShares{types.VaultShare{testDenom1, sdk.NewDec(-5)}, types.VaultShare{testDenom2, sdk.NewDec(100000)}}, false},
		{types.VaultShares{types.VaultShare{"BTC", sdk.NewDec(5)}, types.VaultShare{testDenom2, sdk.NewDec(100000)}}, true},
		{types.VaultShares{types.VaultShare{"0BTC", sdk.NewDec(5)}, types.VaultShare{testDenom2, sdk.NewDec(100000)}}, false},
	}

	for i, tc := range testCases {
		err := tc.input.Validate()
		if tc.expectedPass {
			s.Require().NoError(err, "unexpected result for test case #%d, input: %v", i, tc.input)
		} else {
			s.Require().Error(err, "unexpected result for test case #%d, input: %v", i, tc.input)
		}
	}
}

func (s *vaultShareTestSuite) TestVaultSharesString() {
	testCases := []struct {
		input    types.VaultShares
		expected string
	}{
		{types.VaultShares{}, ""},
		{
			types.VaultShares{
				types.NewVaultShare("atom", sdk.NewDecWithPrec(5040000000000000000, sdk.Precision)),
				types.NewVaultShare("stake", sdk.NewDecWithPrec(4000000000000000, sdk.Precision)),
			},
			"5.040000000000000000atom,0.004000000000000000stake",
		},
	}

	for i, tc := range testCases {
		out := tc.input.String()
		s.Require().Equal(tc.expected, out, "unexpected result for test case #%d, input: %v", i, tc.input)
	}
}

func (s *vaultShareTestSuite) TestNewVaultSharesWithIsValid() {
	fake1 := append(types.NewVaultShares(types.NewVaultShare("mytoken", d(10))), types.VaultShare{Denom: "10BTC", Amount: sdk.NewDec(10)})
	fake2 := append(types.NewVaultShares(types.NewVaultShare("mytoken", d(10))), types.VaultShare{Denom: "BTC", Amount: sdk.NewDec(-10)})

	tests := []struct {
		share      types.VaultShares
		expectPass bool
		msg        string
	}{
		{
			types.NewVaultShares(types.NewVaultShare("mytoken", d(10))),
			true,
			"valid shares should have passed",
		},
		{
			fake1,
			false,
			"invalid denoms",
		},
		{
			fake2,
			false,
			"negative amount",
		},
	}

	for _, tc := range tests {
		tc := tc
		if tc.expectPass {
			s.Require().True(tc.share.IsValid(), tc.msg)
		} else {
			s.Require().False(tc.share.IsValid(), tc.msg)
		}
	}
}

func (s *vaultShareTestSuite) TestVaultShares_AddVaultShareWithIsValid() {
	lengthTestVaultShares := types.NewVaultShares().Add(types.NewVaultShare("mytoken", d(10))).Add(types.VaultShare{Denom: "BTC", Amount: sdk.NewDec(10)})
	s.Require().Equal(2, len(lengthTestVaultShares), "should be 2")

	tests := []struct {
		share      types.VaultShares
		expectPass bool
		msg        string
	}{
		{
			types.NewVaultShares().Add(types.NewVaultShare("mytoken", d(10))),
			true,
			"valid shares should have passed",
		},
		{
			types.NewVaultShares().Add(types.NewVaultShare("mytoken", d(10))).Add(types.VaultShare{Denom: "0BTC", Amount: sdk.NewDec(10)}),
			false,
			"invalid denoms",
		},
		{
			types.NewVaultShares().Add(types.NewVaultShare("mytoken", d(10))).Add(types.VaultShare{Denom: "BTC", Amount: sdk.NewDec(-10)}),
			false,
			"negative amount",
		},
	}

	for _, tc := range tests {
		tc := tc
		if tc.expectPass {
			s.Require().True(tc.share.IsValid(), tc.msg)
		} else {
			s.Require().False(tc.share.IsValid(), tc.msg)
		}
	}
}
