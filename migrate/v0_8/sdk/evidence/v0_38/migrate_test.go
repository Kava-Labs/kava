package v038

import (
	"testing"
	"time"

	v18de63slashing "github.com/kava-labs/kava/migrate/v0_8/sdk/slashing/v18de63"
	"github.com/stretchr/testify/require"
)

func TestMigrate(t *testing.T) {
	age := 21 * 24 * time.Hour
	oldSlashingState := v18de63slashing.GenesisState{
		Params: v18de63slashing.Params{MaxEvidenceAge: age},
	}

	newEvidenceState := Migrate(oldSlashingState)

	// check age param was copied over
	require.Equal(t, age, newEvidenceState.Params.MaxEvidenceAge)
	// check new genesis state is valid
	require.NoError(t, newEvidenceState.Validate())
}
