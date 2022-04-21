package migrate_test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/migrate"
	"github.com/stretchr/testify/require"
)

func TestMigrateGenesisCmd_V17_Success(t *testing.T) {
	ctx := newCmdContext()
	cmd := migrate.MigrateGenesisCmd()
	file := filepath.Join("v0_17", "testdata", "genesis-v16.json")
	cmd.SetArgs([]string{file})
	err := cmd.ExecuteContext(ctx)
	require.NoError(t, err)
}

func newCmdContext() context.Context {
	config := app.MakeEncodingConfig()
	clientCtx := client.Context{}.
		WithCodec(config.Marshaler).
		WithLegacyAmino(config.Amino)
	ctx := context.Background()
	ctx = context.WithValue(ctx, client.ClientContextKey, &clientCtx)
	return ctx
}
