package types_test

import (
	"math/rand"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/tendermint/tendermint/crypto/secp256k1"
	tmtime "github.com/tendermint/tendermint/types/time"

	"github.com/kava-labs/kava/x/cdp/types"
)

type CdpValidationSuite struct {
	suite.Suite

	addrs []sdk.AccAddress
}

func (suite *CdpValidationSuite) SetupTest() {
	r := rand.New(rand.NewSource(12345))
	privkeySeed := make([]byte, 15)
	r.Read(privkeySeed)
	addr := sdk.AccAddress(secp256k1.GenPrivKeySecp256k1(privkeySeed).PubKey().Address())
	suite.addrs = []sdk.AccAddress{addr}
}

func (suite *CdpValidationSuite) TestCdpValidation() {
	type errArgs struct {
		expectPass bool
		contains   string
	}
	testCases := []struct {
		name    string
		cdp     types.CDP
		errArgs errArgs
	}{
		{
			name: "valid cdp",
			cdp:  types.NewCDP(1, suite.addrs[0], sdk.NewInt64Coin("bnb", 100000), "bnb-a", sdk.NewInt64Coin("usdx", 100000), tmtime.Now(), sdk.OneDec()),
			errArgs: errArgs{
				expectPass: true,
				contains:   "",
			},
		},
		{
			name: "invalid cdp id",
			cdp:  types.NewCDP(0, suite.addrs[0], sdk.NewInt64Coin("bnb", 100000), "bnb-a", sdk.NewInt64Coin("usdx", 100000), tmtime.Now(), sdk.OneDec()),
			errArgs: errArgs{
				expectPass: false,
				contains:   "cdp id cannot be 0",
			},
		},
		{
			name: "invalid collateral",
			cdp:  types.CDP{1, suite.addrs[0], "bnb-a", sdk.Coin{"", sdk.NewInt(100)}, sdk.Coin{"usdx", sdk.NewInt(100)}, sdk.Coin{"usdx", sdk.NewInt(0)}, tmtime.Now(), sdk.OneDec()},
			errArgs: errArgs{
				expectPass: false,
				contains:   "invalid coins: collateral",
			},
		},
		{
			name: "invalid prinicpal",
			cdp:  types.CDP{1, suite.addrs[0], "xrp-a", sdk.Coin{"xrp", sdk.NewInt(100)}, sdk.Coin{"", sdk.NewInt(100)}, sdk.Coin{"usdx", sdk.NewInt(0)}, tmtime.Now(), sdk.OneDec()},
			errArgs: errArgs{
				expectPass: false,
				contains:   "invalid coins: principal",
			},
		},
		{
			name: "invalid fees",
			cdp:  types.CDP{1, suite.addrs[0], "xrp-a", sdk.Coin{"xrp", sdk.NewInt(100)}, sdk.Coin{"usdx", sdk.NewInt(100)}, sdk.Coin{"", sdk.NewInt(0)}, tmtime.Now(), sdk.OneDec()},
			errArgs: errArgs{
				expectPass: false,
				contains:   "invalid coins: accumulated fees",
			},
		},
		{
			name: "invalid fees updated",
			cdp:  types.CDP{1, suite.addrs[0], "xrp-a", sdk.Coin{"xrp", sdk.NewInt(100)}, sdk.Coin{"usdx", sdk.NewInt(100)}, sdk.Coin{"usdx", sdk.NewInt(0)}, time.Time{}, sdk.OneDec()},
			errArgs: errArgs{
				expectPass: false,
				contains:   "cdp updated fee time cannot be zero",
			},
		},
		{
			name: "invalid type",
			cdp:  types.CDP{1, suite.addrs[0], "", sdk.Coin{"xrp", sdk.NewInt(100)}, sdk.Coin{"usdx", sdk.NewInt(100)}, sdk.Coin{"usdx", sdk.NewInt(0)}, tmtime.Now(), sdk.OneDec()},
			errArgs: errArgs{
				expectPass: false,
				contains:   "cdp type cannot be empty",
			},
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			err := tc.cdp.Validate()
			if tc.errArgs.expectPass {
				suite.Require().NoError(err, tc.name)
			} else {
				suite.Require().Error(err, tc.name)
				suite.Require().True(strings.Contains(err.Error(), tc.errArgs.contains))
			}
		})
	}
}

func (suite *CdpValidationSuite) TestDepositValidation() {
	type errArgs struct {
		expectPass bool
		contains   string
	}
	testCases := []struct {
		name    string
		deposit types.Deposit
		errArgs errArgs
	}{
		{
			name:    "valid deposit",
			deposit: types.NewDeposit(1, suite.addrs[0], sdk.NewInt64Coin("bnb", 1000000)),
			errArgs: errArgs{
				expectPass: true,
				contains:   "",
			},
		},
		{
			name:    "invalid cdp id",
			deposit: types.NewDeposit(0, suite.addrs[0], sdk.NewInt64Coin("bnb", 1000000)),
			errArgs: errArgs{
				expectPass: false,
				contains:   "deposit's cdp id cannot be 0",
			},
		},
		{
			name:    "empty depositor",
			deposit: types.NewDeposit(1, sdk.AccAddress{}, sdk.NewInt64Coin("bnb", 1000000)),
			errArgs: errArgs{
				expectPass: false,
				contains:   "depositor cannot be empty",
			},
		},
		{
			name:    "invalid deposit coins",
			deposit: types.NewDeposit(1, suite.addrs[0], sdk.Coin{"Invalid Denom", sdk.NewInt(1000000)}),
			errArgs: errArgs{
				expectPass: false,
				contains:   "invalid coins: deposit",
			},
		},
	}
	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			err := tc.deposit.Validate()
			if tc.errArgs.expectPass {
				suite.Require().NoError(err, tc.name)
			} else {
				suite.Require().Error(err, tc.name)
				suite.Require().True(strings.Contains(err.Error(), tc.errArgs.contains))
			}
		})
	}
}

func (suite *CdpValidationSuite) TestCdpGetTotalPrinciple() {
	principal := sdk.Coin{"usdx", sdk.NewInt(100500)}
	accumulatedFees := sdk.Coin{"usdx", sdk.NewInt(25000)}

	cdp := types.CDP{Principal: principal, AccumulatedFees: accumulatedFees}

	suite.Require().Equal(cdp.GetTotalPrincipal(), principal.Add(accumulatedFees))
}

func (suite *CdpValidationSuite) TestCDPGetNormalizedPrincipal() {
	type expectedErr struct {
		expectPass bool
		contains   string
	}
	testCases := []struct {
		name        string
		cdp         types.CDP
		expected    sdk.Dec
		expectedErr expectedErr
	}{
		{
			name: "principal + fees is divided by factor correctly",
			cdp: types.CDP{
				Principal:       sdk.NewInt64Coin("usdx", 1e9),
				AccumulatedFees: sdk.NewInt64Coin("usdx", 1e6),
				InterestFactor:  sdk.MustNewDecFromStr("2"),
			},
			expected: sdk.MustNewDecFromStr("500500000"),
			expectedErr: expectedErr{
				expectPass: true,
			},
		},
		{
			name: "factor < 1 returns error",
			cdp: types.CDP{
				Principal:       sdk.NewInt64Coin("usdx", 1e9),
				AccumulatedFees: sdk.NewInt64Coin("usdx", 1e6),
				InterestFactor:  sdk.MustNewDecFromStr("0.999999999999999999"),
			},
			expectedErr: expectedErr{
				contains: "must be ≥ 1",
			},
		},
		{
			name: "0 factor returns error rather than div by 0 panic",
			cdp: types.CDP{
				Principal:       sdk.NewInt64Coin("usdx", 1e9),
				AccumulatedFees: sdk.NewInt64Coin("usdx", 1e6),
				InterestFactor:  sdk.MustNewDecFromStr("0"),
			},
			expectedErr: expectedErr{
				contains: "must be ≥ 1",
			},
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			np, err := tc.cdp.GetNormalizedPrincipal()

			if tc.expectedErr.expectPass {
				suite.Require().NoError(err, tc.name)
				suite.Equal(tc.expected, np)
			} else {
				suite.Require().Error(err, tc.name)
				suite.Contains(err.Error(), tc.expectedErr.contains)
			}
		})
	}
}

func TestCdpValidationSuite(t *testing.T) {
	suite.Run(t, new(CdpValidationSuite))
}
