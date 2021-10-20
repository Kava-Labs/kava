package cmd

import (
	"io"

	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	"github.com/cosmos/cosmos-sdk/x/crisis"
	"github.com/spf13/cobra"
	"github.com/tendermint/tendermint/libs/log"
	db "github.com/tendermint/tm-db"

	"github.com/kava-labs/kava/app/params"
)

type appCreator struct {
	encodingConfig params.EncodingConfig
}

func (ac appCreator) newApp(
	logger log.Logger,
	db db.DB,
	traceStore io.Writer,
	appOpts servertypes.AppOptions,
) servertypes.Application {
	panic("TODO") // TODO
	return nil
}

func (ac appCreator) appExport(
	logger log.Logger,
	db db.DB,
	traceStore io.Writer,
	height int64,
	forZeroHeight bool,
	jailAllowedAddrs []string,
	appOpts servertypes.AppOptions,
) (servertypes.ExportedApp, error) {
	panic("TODO") // TODO
	return servertypes.ExportedApp{}, nil
}

func (ac appCreator) addStartCmdFlags(startCmd *cobra.Command) {
	crisis.AddModuleInitFlags(startCmd)
}
