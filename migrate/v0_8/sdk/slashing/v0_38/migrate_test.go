package v038

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/slashing"

	v18de63slashing "github.com/kava-labs/kava/migrate/v0_8/sdk/slashing/v18de63"
)

func TestMigrate(t *testing.T) {
	oldState := v18de63slashing.GenesisState{
		Params: v18de63slashing.Params{
			DowntimeJailDuration:    10 * time.Minute,
			MaxEvidenceAge:          21 * 24 * time.Hour,
			MinSignedPerWindow:      sdk.MustNewDecFromStr("0.05"),
			SignedBlocksWindow:      10000,
			SlashFractionDoubleSign: sdk.MustNewDecFromStr("0.05"),
			SlashFractionDowntime:   sdk.MustNewDecFromStr("0.0001"),
		},
		SigningInfos: nil,
		MissedBlocks: nil,
	}

	newState := Migrate(oldState)

	// check new genesis state is valid
	require.NoError(t, slashing.ValidateGenesis(newState))
}
