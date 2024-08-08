package noop_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kava-labs/kava/precompile/contracts/noop"
)

// TestContractConstructor ensures we have a valid constructor. This will fail
// if we attempt to define invalid or duplicate function selectors.
func TestContractConstructor(t *testing.T) {
	precompile, err := noop.NewContract()
	require.NoError(t, err, "expected precompile not error when created")
	assert.NotNil(t, precompile, "expected precompile contract to be defined")
}
