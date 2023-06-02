package types_test

import (
	"fmt"
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

func TestInternalEVMAddress_ProtobufMarshaller(t *testing.T) {
	t.Run("Equal", func(t *testing.T) {
		addr := testutil.RandomEvmAddress()
		require.True(t, types.NewInternalEVMAddress(addr).Equal(testutil.MustNewInternalEVMAddressFromString(addr.Hex())))
	})

	t.Run("MarshalTo", func(t *testing.T) {
		addr := testutil.RandomInternalEVMAddress()
		expectedBytes := addr.Bytes()

		data := make([]byte, len(expectedBytes))
		n, err := addr.MarshalTo(data)
		require.NoError(t, err)
		// check length
		require.Equal(t, len(expectedBytes), n)
		// check data
		require.Equal(t, expectedBytes, data)
	})

	t.Run("MarshalJSON", func(t *testing.T) {
		addr := testutil.RandomInternalEVMAddress()
		expected := fmt.Sprintf("\"%s\"", addr.Hex())
		marshalled, err := addr.MarshalJSON()
		require.NoError(t, err)
		require.Equal(t, expected, string(marshalled))
	})

	t.Run("Size", func(t *testing.T) {
		addr := testutil.RandomInternalEVMAddress()
		require.Equal(t, 20, addr.Size())
	})

	t.Run("Unmarshal", func(t *testing.T) {
		addr := types.InternalEVMAddress{}
		expectedAddress := testutil.RandomEvmAddress()
		data := expectedAddress.Bytes()

		err := addr.Unmarshal(data)
		require.NoError(t, err)

		// check address is properly set
		require.Equal(t, expectedAddress, addr.Address)

		// fails with invalid data length
		invalidData := []byte{0xbe, 0xef}
		err = addr.Unmarshal(invalidData)
		require.ErrorContains(t, err, "invalid data length for InternalEVMAddress")
	})
}
