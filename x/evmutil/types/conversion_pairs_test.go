package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/kava-labs/kava/x/evmutil/testutil"
	"github.com/kava-labs/kava/x/evmutil/types"
)

func TestConversionPairValidate(t *testing.T) {
	type errArgs struct {
		expectPass bool
		contains   string
	}
	tests := []struct {
		name        string
		giveAddress types.InternalEVMAddress
		giveDenom   string
		givDecimals int32
		errArgs     errArgs
	}{
		{
			"valid",
			testutil.MustNewInternalEVMAddressFromString("0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2"),
			"weth",
			0,
			errArgs{
				expectPass: true,
			},
		},
		{
			"valid - negative decimals",
			testutil.MustNewInternalEVMAddressFromString("0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2"),
			"weth",
			-9999999,
			errArgs{
				expectPass: true,
			},
		},
		{
			"valid - positive decimals",
			testutil.MustNewInternalEVMAddressFromString("0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2"),
			"weth",
			9999999, // TODO limit this to the max range supported by sdk.Int and the evm
			errArgs{
				expectPass: true,
			},
		},
		{
			"invalid - empty denom",
			testutil.MustNewInternalEVMAddressFromString("0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2"),
			"",
			0,
			errArgs{
				expectPass: false,
				contains:   "conversion pair denom invalid: invalid denom",
			},
		},
		{
			"invalid - zero address",
			testutil.MustNewInternalEVMAddressFromString("0x0000000000000000000000000000000000000000"),
			"weth",
			0,
			errArgs{
				expectPass: false,
				contains:   "address cannot be zero value",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			pair := types.NewConversionPairWithDecimal(tc.giveAddress, tc.giveDenom, tc.givDecimals)

			err := pair.Validate()

			if tc.errArgs.expectPass {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.errArgs.contains)
			}
		})
	}
}

func TestConversionPairValidate_Direct(t *testing.T) {
	type errArgs struct {
		expectPass bool
		contains   string
	}
	tests := []struct {
		name     string
		givePair types.ConversionPair
		errArgs  errArgs
	}{
		{
			"valid",
			types.ConversionPair{
				KavaERC20Address: testutil.MustNewInternalEVMAddressFromString("0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2").Bytes(),
				Denom:            "weth",
			},
			errArgs{
				expectPass: true,
			},
		},

		{
			"invalid - length",
			types.ConversionPair{
				KavaERC20Address: []byte{1},
				Denom:            "weth",
			},
			errArgs{
				expectPass: false,
				contains:   "address length is 1 but expected 20",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.givePair.Validate()

			if tc.errArgs.expectPass {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.errArgs.contains)
			}
		})
	}
}

func TestConversionPair_GetAddress(t *testing.T) {
	addr := testutil.MustNewInternalEVMAddressFromString("0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2")

	pair := types.NewConversionPair(
		addr,
		"weth",
		0,
	)

	require.Equal(t, types.HexBytes(addr.Bytes()), pair.KavaERC20Address, "struct address should match input bytes")
	require.Equal(t, addr, pair.GetAddress(), "get internal address should match input bytes")
}

func TestConversionPairs_Validate(t *testing.T) {
	type errArgs struct {
		expectPass bool
		contains   string
	}
	tests := []struct {
		name      string
		givePairs types.ConversionPairs
		errArgs   errArgs
	}{
		{
			"valid",
			types.NewConversionPairs(
				types.NewConversionPair(
					testutil.MustNewInternalEVMAddressFromString("0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2"),
					"weth",
				),
				types.NewConversionPair(
					testutil.MustNewInternalEVMAddressFromString("0x000000000000000000000000000000000000000A"),
					"kava",
				),
				types.NewConversionPair(
					testutil.MustNewInternalEVMAddressFromString("0x000000000000000000000000000000000000000B"),
					"usdc",
				),
			),
			errArgs{
				expectPass: true,
			},
		},
		{
			"invalid - duplicate address",
			types.NewConversionPairs(
				types.NewConversionPair(
					testutil.MustNewInternalEVMAddressFromString("0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2"),
					"weth",
				),
				types.NewConversionPair(
					testutil.MustNewInternalEVMAddressFromString("0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2"),
					"kava",
				),
				types.NewConversionPair(
					testutil.MustNewInternalEVMAddressFromString("0x000000000000000000000000000000000000000B"),
					"usdc",
				),
			),
			errArgs{
				expectPass: false,
				contains:   "found duplicate enabled conversion pair internal ERC20 address",
			},
		},
		{
			"invalid - duplicate denom",
			types.NewConversionPairs(
				types.NewConversionPair(
					testutil.MustNewInternalEVMAddressFromString("0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2"),
					"weth",
				),
				types.NewConversionPair(
					testutil.MustNewInternalEVMAddressFromString("0x000000000000000000000000000000000000000A"),
					"kava",
				),
				types.NewConversionPair(
					testutil.MustNewInternalEVMAddressFromString("0x000000000000000000000000000000000000000B"),
					"kava",
				),
			),
			errArgs{
				expectPass: false,
				contains:   "found duplicate enabled conversion pair denom kava",
			},
		},
		{
			"invalid - invalid pair",
			types.NewConversionPairs(
				types.NewConversionPair(
					testutil.MustNewInternalEVMAddressFromString("0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2"),
					"weth",
				),
				types.NewConversionPair(
					testutil.MustNewInternalEVMAddressFromString("0x0000000000000000000000000000000000000000"),
					"usdc",
				),
				types.NewConversionPair(
					testutil.MustNewInternalEVMAddressFromString("0x000000000000000000000000000000000000000B"),
					"kava",
				),
			),
			errArgs{
				expectPass: false,
				contains:   "address cannot be zero value",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.givePairs.Validate()

			if tc.errArgs.expectPass {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.errArgs.contains)
			}
		})
	}
}

func TestAllowedCosmosCoinERC20Token_Validate(t *testing.T) {
	testCases := []struct {
		name   string
		token  types.AllowedCosmosCoinERC20Token
		expErr string
	}{
		{
			name:   "valid token",
			token:  types.NewAllowedCosmosCoinERC20Token("uatom", "Kava-wrapped ATOM", "kATOM", 6),
			expErr: "",
		},
		{
			name:   "valid - highest allowed decimals",
			token:  types.NewAllowedCosmosCoinERC20Token("uatom", "Kava-wrapped ATOM", "kATOM", 255),
			expErr: "",
		},
		{
			name: "invalid - Empty SdkDenom",
			token: types.AllowedCosmosCoinERC20Token{
				CosmosDenom: "",
				Name:        "Example Token",
				Symbol:      "ETK",
				Decimals:    0,
			},
			expErr: "sdk denom is invalid",
		},
		{
			name: "invalid - Empty Name",
			token: types.AllowedCosmosCoinERC20Token{
				CosmosDenom: "example_denom",
				Name:        "",
				Symbol:      "ETK",
				Decimals:    6,
			},
			expErr: "name cannot be empty",
		},
		{
			name: "invalid - Empty Symbol",
			token: types.AllowedCosmosCoinERC20Token{
				CosmosDenom: "example_denom",
				Name:        "Example Token",
				Symbol:      "",
				Decimals:    6,
			},
			expErr: "symbol cannot be empty",
		},
		{
			name:   "invalid - decimals higher than uint8",
			token:  types.NewAllowedCosmosCoinERC20Token("uatom", "Kava-wrapped ATOM", "kATOM", 256),
			expErr: "decimals must be less than 256",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.token.Validate()
			if tc.expErr != "" {
				require.ErrorContains(t, err, tc.expErr, "Expected validation error")
			} else {
				require.NoError(t, err, "Expected no validation error")
			}
		})
	}
}

func TestAllowedCosmosCoinERC20Tokens_Validate(t *testing.T) {
	token1 := types.NewAllowedCosmosCoinERC20Token("denom1", "Token 1", "TK1", 6)
	token2 := types.NewAllowedCosmosCoinERC20Token("denom2", "Token 2", "TK2", 0)
	invalidToken := types.NewAllowedCosmosCoinERC20Token("", "No SDK Denom Token", "TK3", 18)

	testCases := []struct {
		name   string
		tokens types.AllowedCosmosCoinERC20Tokens
		expErr string
	}{
		{
			name:   "valid - no tokens",
			tokens: types.NewAllowedCosmosCoinERC20Tokens(),
			expErr: "",
		},
		{
			name:   "valid - one token",
			tokens: types.NewAllowedCosmosCoinERC20Tokens(token1),
			expErr: "",
		},
		{
			name:   "valid - multiple tokens",
			tokens: types.NewAllowedCosmosCoinERC20Tokens(token1, token2),
			expErr: "",
		},
		{
			name:   "invalid - contains invalid token",
			tokens: types.NewAllowedCosmosCoinERC20Tokens(token1, token2, invalidToken),
			expErr: "invalid token at index 2",
		},
		{
			name:   "invalid - duplicate denoms",
			tokens: types.NewAllowedCosmosCoinERC20Tokens(token1, token2, token1),
			expErr: "found duplicate token with sdk denom denom1",
		},
		{
			name: "invalid - duplicate symbol",
			tokens: types.NewAllowedCosmosCoinERC20Tokens(
				token1,
				types.NewAllowedCosmosCoinERC20Token("diff", "Diff Denom, Same Symbol", "TK1", 6),
			),
			expErr: "found duplicate token with symbol TK1",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.tokens.Validate()
			if tc.expErr != "" {
				require.ErrorContains(t, err, tc.expErr, "Expected validation error")
			} else {
				require.NoError(t, err, "Expected no validation error")
			}
		})
	}
}
