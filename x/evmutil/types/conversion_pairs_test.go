package types_test

import (
	"testing"

	"github.com/kava-labs/kava/x/evmutil/testutil"
	"github.com/kava-labs/kava/x/evmutil/types"
	"github.com/stretchr/testify/require"
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
		errArgs     errArgs
	}{
		{
			"valid",
			testutil.MustNewInternalEVMAddressFromString("0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2"),
			"weth",
			errArgs{
				expectPass: true,
			},
		},
		{
			"invalid - empty denom",
			testutil.MustNewInternalEVMAddressFromString("0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2"),
			"",
			errArgs{
				expectPass: false,
				contains:   "denom cannot be empty",
			},
		},
		{
			"invalid - zero address",
			testutil.MustNewInternalEVMAddressFromString("0x0000000000000000000000000000000000000000"),
			"weth",
			errArgs{
				expectPass: false,
				contains:   "address cannot be zero value",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			pair := types.NewConversionPair(tc.giveAddress, tc.giveDenom)

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
