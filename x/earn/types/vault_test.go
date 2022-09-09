package types_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/earn/types"
)

func TestVaultRecordValidate(t *testing.T) {
	type errArgs struct {
		expectPass bool
		contains   string
	}

	tests := []struct {
		name         string
		vaultRecords types.VaultRecords
		errArgs      errArgs
	}{
		{
			name: "valid vault records",
			vaultRecords: types.VaultRecords{
				{
					TotalShares: types.NewVaultShare("usdx", sdk.NewDec(0)),
				},
				{
					TotalShares: types.NewVaultShare("ukava", sdk.NewDec(5)),
				},
			},
			errArgs: errArgs{
				expectPass: true,
			},
		},
		{
			name: "invalid - duplicate denom",
			vaultRecords: types.VaultRecords{
				{
					TotalShares: types.NewVaultShare("usdx", sdk.NewDec(0)),
				},
				{
					TotalShares: types.NewVaultShare("usdx", sdk.NewDec(5)),
				},
			},
			errArgs: errArgs{
				expectPass: false,
				contains:   "duplicate vault denom usdx",
			},
		},
		{
			name: "invalid - invalid denom",
			vaultRecords: types.VaultRecords{
				{
					TotalShares: types.VaultShare{Denom: "", Amount: sdk.NewDec(0)},
				},
			},
			errArgs: errArgs{
				expectPass: false,
				contains:   "invalid denom",
			},
		},
		{
			name: "invalid - negative",
			vaultRecords: types.VaultRecords{
				{
					TotalShares: types.VaultShare{"usdx", sdk.NewDec(-5)},
				},
			},
			errArgs: errArgs{
				expectPass: false,
				contains:   "vault share amount -5.000000000000000000 is negative",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.vaultRecords.Validate()

			if test.errArgs.expectPass {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				require.Contains(t, err.Error(), test.errArgs.contains)
			}
		})
	}
}

func TestVaultShareRecordsValidate(t *testing.T) {
	_, addrs := app.GeneratePrivKeyAddressPairs(2)

	type errArgs struct {
		expectPass bool
		contains   string
	}

	tests := []struct {
		name         string
		vaultRecords types.VaultShareRecords
		errArgs      errArgs
	}{
		{
			name: "valid vault share records",
			vaultRecords: types.VaultShareRecords{
				{
					Depositor: addrs[0],
					Shares: types.NewVaultShares(
						types.NewVaultShare("usdx", sdk.NewDec(0)),
					),
				},
				{
					Depositor: addrs[1],
					Shares: types.NewVaultShares(
						types.NewVaultShare("usdx", sdk.NewDec(0)),
						types.NewVaultShare("ukava", sdk.NewDec(5)),
					),
				},
			},
			errArgs: errArgs{
				expectPass: true,
			},
		},
		{
			name: "invalid - duplicate address",
			vaultRecords: types.VaultShareRecords{
				{
					Depositor: addrs[0],
					Shares: types.NewVaultShares(
						types.NewVaultShare("usdx", sdk.NewDec(0)),
					),
				},
				{
					Depositor: addrs[0],
					Shares: types.NewVaultShares(
						types.NewVaultShare("usdx", sdk.NewDec(0)),
						types.NewVaultShare("ukava", sdk.NewDec(5)),
					),
				},
			},
			errArgs: errArgs{
				expectPass: false,
				contains:   "duplicate address",
			},
		},
		{
			name: "invalid - invalid address",
			vaultRecords: types.VaultShareRecords{
				{
					Depositor: sdk.AccAddress{},
					Shares: types.NewVaultShares(
						types.NewVaultShare("usdx", sdk.NewDec(0)),
					),
				},
			},
			errArgs: errArgs{
				expectPass: false,
				contains:   "depositor is empty",
			},
		},
		{
			name: "invalid - negative",
			vaultRecords: types.VaultShareRecords{
				{
					Depositor: addrs[0],
					// Direct slice, not NewVaultShares() which panics
					Shares: types.VaultShares{
						types.VaultShare{"usdx", sdk.NewDec(-5)},
					},
				},
			},
			errArgs: errArgs{
				expectPass: false,
				contains:   "invalid vault share record shares: share -5.000000000000000000usdx amount is not positive",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.vaultRecords.Validate()

			if test.errArgs.expectPass {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				require.Contains(t, err.Error(), test.errArgs.contains)
			}
		})
	}
}

func TestAllowedVaultsValidate(t *testing.T) {
	type errArgs struct {
		expectPass bool
		contains   string
	}

	tests := []struct {
		name         string
		vaultRecords types.AllowedVaults
		errArgs      errArgs
	}{
		{
			name: "valid vault share records",
			vaultRecords: types.AllowedVaults{
				{
					Denom:         "usdx",
					VaultStrategy: types.STRATEGY_TYPE_HARD,
				},
				{
					Denom:         "busd",
					VaultStrategy: types.STRATEGY_TYPE_HARD,
				},
			},
			errArgs: errArgs{
				expectPass: true,
			},
		},
		{
			name: "invalid - duplicate denom",
			vaultRecords: types.AllowedVaults{
				{
					Denom:         "usdx",
					VaultStrategy: types.STRATEGY_TYPE_HARD,
				},
				{
					Denom:         "usdx",
					VaultStrategy: types.STRATEGY_TYPE_HARD,
				},
			},
			errArgs: errArgs{
				expectPass: false,
				contains:   "duplicate vault denom usdx",
			},
		},
		{
			name: "invalid - invalid denom",
			vaultRecords: types.AllowedVaults{
				{
					Denom:         "",
					VaultStrategy: types.STRATEGY_TYPE_HARD,
				},
			},
			errArgs: errArgs{
				expectPass: false,
				contains:   "invalid denom",
			},
		},
		{
			name: "invalid - invalid strategy",
			vaultRecords: types.AllowedVaults{
				{
					Denom:         "usdx",
					VaultStrategy: types.STRATEGY_TYPE_UNSPECIFIED,
				},
			},
			errArgs: errArgs{
				expectPass: false,
				contains:   "invalid vault strategy",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.vaultRecords.Validate()

			if test.errArgs.expectPass {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				require.Contains(t, err.Error(), test.errArgs.contains)
			}
		})
	}
}

func TestNewVaultShareRecord(t *testing.T) {
	_, addrs := app.GeneratePrivKeyAddressPairs(1)

	shares := types.NewVaultShares(
		types.NewVaultShare("usdx", sdk.NewDec(0)),
		types.NewVaultShare("ukava", sdk.NewDec(5)),
	)

	shareRecord := types.NewVaultShareRecord(addrs[0], shares)
	require.Equal(t, shares, shareRecord.Shares)
}
