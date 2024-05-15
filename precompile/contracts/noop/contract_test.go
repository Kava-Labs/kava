package noop

import (
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/precompile/contract"
	"github.com/stretchr/testify/require"
)

var callerAddr = common.HexToAddress("0xc0ffee254729296a45a3885639AC7E10F9d54979")

type accessibleState struct {
	stateDB contract.StateDB
}

func newAccessibleState(stateDB contract.StateDB) *accessibleState {
	return &accessibleState{
		stateDB: stateDB,
	}
}

func (s *accessibleState) GetStateDB() contract.StateDB {
	return s.stateDB
}

func TestNoopPrecompileGasDeductions(t *testing.T) {
	var (
		stateDB         = state.NewTestStateDB(t)
		accessibleState = newAccessibleState(stateDB)
		contractAddr    = Module.Address
		readOnly        bool
	)

	for _, tc := range []struct {
		desc         string
		suppliedGas  uint64
		remainingGas uint64
		err          string
	}{
		{
			desc:         "not enough gas",
			suppliedGas:  noopGasCost - 1,
			remainingGas: 0,
			err:          "out of gas",
		},
		{
			desc:         "enough gas",
			suppliedGas:  noopGasCost,
			remainingGas: 0,
			err:          "",
		},
		{
			desc:         "more than enough gas",
			suppliedGas:  noopGasCost + 1,
			remainingGas: 1,
			err:          "",
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			input := contract.MustCalculateFunctionSelector("noop()")

			ret, remainingGas, err := Module.Contract.Run(accessibleState, callerAddr, contractAddr, input, tc.suppliedGas, readOnly)
			if tc.err != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.err)
				return
			}
			require.NoError(t, err)
			require.Empty(t, ret)
			require.Equal(t, tc.remainingGas, remainingGas)
		})
	}
}

func TestNoopPrecompileInvalidCalls(t *testing.T) {
	var (
		stateDB         = state.NewTestStateDB(t)
		accessibleState = newAccessibleState(stateDB)
		contractAddr    = Module.Address
		readOnly        bool
	)

	unexistingFuncInput := contract.MustCalculateFunctionSelector("unexistingFunc()")
	invalidArgNumInput := contract.MustCalculateFunctionSelector("noop(uint256)")
	shortFuncSelectorInput := []byte("abc")

	for _, tc := range []struct {
		desc  string
		input []byte
		err   string
	}{
		{
			desc:  "test case #1",
			input: unexistingFuncInput,
			err:   "invalid function selector",
		},
		{
			desc:  "test case #2",
			input: invalidArgNumInput,
			err:   "invalid function selector",
		},
		{
			desc:  "test case #3",
			input: shortFuncSelectorInput,
			err:   "missing function selector to precompile",
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			_, _, err := Module.Contract.Run(accessibleState, callerAddr, contractAddr, tc.input, noopGasCost, readOnly)
			require.Error(t, err)
			require.Contains(t, err.Error(), tc.err)
		})
	}
}
