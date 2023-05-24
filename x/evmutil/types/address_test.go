package types_test

import (
	"testing"

	"github.com/kava-labs/kava/x/evmutil/testutil"
	"github.com/kava-labs/kava/x/evmutil/types"
	"github.com/stretchr/testify/require"
)

func TestInternalEVMAddress_BytesToInternalEVMAddress(t *testing.T) {
	addr := testutil.RandomEvmAddress()
	require.Equal(t,
		types.NewInternalEVMAddress(addr),
		types.BytesToInternalEVMAddress(addr.Bytes()),
	)
}

func TestInternalEVMAddress_IsNil(t *testing.T) {
	addr := types.InternalEVMAddress{}
	require.True(t, addr.IsNil())
	addr.Address = testutil.RandomEvmAddress()
	require.False(t, addr.IsNil())
}

func TestInternalEVMAddress_NewInternalEVMAddressFromString(t *testing.T) {
	t.Run("works with valid address string", func(t *testing.T) {
		validAddr := testutil.RandomEvmAddress()
		addr, err := types.NewInternalEVMAddressFromString(validAddr.Hex())
		require.NoError(t, err)
		require.Equal(t, types.NewInternalEVMAddress(validAddr), addr)
	})

	t.Run("fails with invalid hex string", func(t *testing.T) {
		_, err := types.NewInternalEVMAddressFromString("0xinvalid-address")
		require.ErrorContains(t, err, "string is not a hex address")
	})
}
