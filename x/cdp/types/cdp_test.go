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
			cdp:  types.NewCDP(1, suite.addrs[0], sdk.NewInt64Coin("bnb", 100000), sdk.NewInt64Coin("usdx", 100000), tmtime.Now()),
			errArgs: errArgs{
				expectPass: true,
				contains:   "",
			},
		},
		{
			name: "invalid cdp id",
			cdp:  types.NewCDP(0, suite.addrs[0], sdk.NewInt64Coin("bnb", 100000), sdk.NewInt64Coin("usdx", 100000), tmtime.Now()),
			errArgs: errArgs{
				expectPass: false,
				contains:   "cdp id cannot be 0",
			},
		},
		{
			name: "invalid collateral",
			cdp:  types.CDP{1, suite.addrs[0], sdk.Coin{"", sdk.NewInt(100)}, sdk.Coin{"usdx", sdk.NewInt(100)}, sdk.Coin{"usdx", sdk.NewInt(0)}, tmtime.Now()},
			errArgs: errArgs{
				expectPass: false,
				contains:   "invalid coins: collateral",
			},
		},
		{
			name: "invalid prinicpal",
			cdp:  types.CDP{1, suite.addrs[0], sdk.Coin{"xrp", sdk.NewInt(100)}, sdk.Coin{"", sdk.NewInt(100)}, sdk.Coin{"usdx", sdk.NewInt(0)}, tmtime.Now()},
			errArgs: errArgs{
				expectPass: false,
				contains:   "invalid coins: principal",
			},
		},
		{
			name: "invalid fees",
			cdp:  types.CDP{1, suite.addrs[0], sdk.Coin{"xrp", sdk.NewInt(100)}, sdk.Coin{"usdx", sdk.NewInt(100)}, sdk.Coin{"", sdk.NewInt(0)}, tmtime.Now()},
			errArgs: errArgs{
				expectPass: false,
				contains:   "invalid coins: accumulated fees",
			},
		},
		{
			name: "invalid fees updated",
			cdp:  types.CDP{1, suite.addrs[0], sdk.Coin{"xrp", sdk.NewInt(100)}, sdk.Coin{"usdx", sdk.NewInt(100)}, sdk.Coin{"usdx", sdk.NewInt(0)}, time.Time{}},
			errArgs: errArgs{
				expectPass: false,
				contains:   "cdp updated fee time cannot be zero",
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

func TestCdpValidationSuite(t *testing.T) {
	suite.Run(t, new(CdpValidationSuite))
}
