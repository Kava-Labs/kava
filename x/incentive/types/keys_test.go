package types_test

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/kava-labs/kava/x/incentive/types"
	"github.com/stretchr/testify/require"
)

func TestDecodeKeyPrefix(t *testing.T) {
	tests := []struct {
		name          string
		key           []byte
		wantClaimType types.ClaimType
		wantSubKey    string
		wantErr       error
	}{
		{
			"valid Claim key - empty subkey",
			types.GetClaimKeyPrefix(types.CLAIM_TYPE_USDX_MINTING),
			types.CLAIM_TYPE_USDX_MINTING,
			"",
			nil,
		},
		{
			"valid Claim key - with subkey",
			append(types.GetClaimKeyPrefix(types.CLAIM_TYPE_USDX_MINTING), []byte("usdx")...),
			types.CLAIM_TYPE_USDX_MINTING,
			"usdx",
			nil,
		},
		{
			"valid key without data type",
			bytes.TrimPrefix(
				append(
					types.GetClaimKeyPrefix(types.CLAIM_TYPE_USDX_MINTING),
					[]byte("usdx")...,
				),
				types.ClaimKeyPrefix, // remove the claim key prefix
			),
			types.CLAIM_TYPE_USDX_MINTING,
			"usdx",
			nil,
		},
		{
			"invalid key prefix",
			[]byte{1, 2, 3},
			types.CLAIM_TYPE_UNSPECIFIED,
			"",
			fmt.Errorf("invalid key prefix length to decode ClaimType: %v", string([]byte{1, 2, 3})),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotClaimType, gotSubKey, err := types.DecodeKeyPrefix(tt.key)

			if tt.wantErr != nil {
				require.Error(t, err)
				require.Equal(t, tt.wantErr.Error(), err.Error())
				return
			}

			require.NoError(t, err)
			require.Equal(t, tt.wantClaimType, gotClaimType)
			require.Equal(t, tt.wantSubKey, gotSubKey)
		})
	}
}
