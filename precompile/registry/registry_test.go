package registry_test

import (
	"testing"

	"github.com/ethereum/go-ethereum/precompile/modules"
	"github.com/stretchr/testify/assert"
)

// TestRegisteredPrecompiles asserts precompiles are registered
//
// In addition, this serves as an integration test to
//  1. Ensure modules.RegisteredModules() is returning addresses in the correct ascending order
//  2. Ensure that that the address defined in the module is correct. Since we use common.HexToAddress and
//     then back to 0x encoded string, we can be certain that the string defined in the module is the
//     expected length, not missing 0's, etc.
func TestRegisteredPrecompilesAddresses(t *testing.T) {
	// build list of 0x addresses that are registered
	registeredModules := modules.RegisteredModules()
	registeredPrecompiles := make([]string, 0, len(registeredModules))
	for _, rp := range registeredModules {
		registeredPrecompiles = append(registeredPrecompiles, rp.Address.String())
	}

	expectedPrecompiles := []string{
		// 0x9 address space used for e2e & integration tests
		"0x9000000000000000000000000000000000000001", // noop
		"0x9000000000000000000000000000000000000002", // noop (duplicated for testing)
	}

	assert.Equal(t, expectedPrecompiles, registeredPrecompiles,
		"expected registered precompile address list to match to match expected")
}
