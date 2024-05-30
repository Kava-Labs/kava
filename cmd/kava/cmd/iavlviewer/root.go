package iavlviewer

import (
	"fmt"
	"strconv"

	dbm "github.com/cometbft/cometbft-db"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/server"
	ethermintserver "github.com/evmos/ethermint/server"
	"github.com/spf13/cobra"

	"github.com/cosmos/iavl"
)

const (
	DefaultCacheSize int = 10000
)

func NewCmd(opts ethermintserver.StartOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "iavlviewer <data|hash|shape|versions> <prefix> [version number]",
		Short: "Output various data, hashes, and calculations for an iavl tree",
	}

	cmd.AddCommand(newDataCmd(opts))
	cmd.AddCommand(newHashCmd(opts))
	cmd.AddCommand(newShapeCmd(opts))
	cmd.AddCommand(newVersionsCmd(opts))

	return cmd
}

func parseVersion(arg string) (int, error) {
	version, err := strconv.Atoi(arg)
	if err != nil {
		return 0, fmt.Errorf("invalid version number: '%s'", arg)
	}
	return version, nil
}

func openPrefixTree(opts ethermintserver.StartOptions, cmd *cobra.Command, prefix string, version int) (*iavl.MutableTree, error) {
	clientCtx := client.GetClientContextFromCmd(cmd)
	ctx := server.GetServerContextFromCmd(cmd)
	ctx.Config.SetRoot(clientCtx.HomeDir)

	db, err := opts.DBOpener(ctx.Viper, clientCtx.HomeDir, server.GetAppDBBackend(ctx.Viper))
	if err != nil {
		return nil, fmt.Errorf("failed to open database at %s: %s", clientCtx.HomeDir, err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			ctx.Logger.Error("error closing db", "error", err.Error())
		}
	}()

	tree, err := readTree(db, version, []byte(prefix))
	if err != nil {
		return nil, fmt.Errorf("failed to read tree with prefix %s: %s", prefix, err)
	}
	return tree, nil
}

// ReadTree loads an iavl tree from the directory
// If version is 0, load latest, otherwise, load named version
// The prefix represents which iavl tree you want to read. The iaviwer will always set a prefix.
func readTree(db dbm.DB, version int, prefix []byte) (*iavl.MutableTree, error) {
	if len(prefix) != 0 {
		db = dbm.NewPrefixDB(db, prefix)
	}

	tree, err := iavl.NewMutableTree(db, DefaultCacheSize, false)
	if err != nil {
		return nil, err
	}
	ver, err := tree.LoadVersion(int64(version))
	if err != nil {
		return nil, err
	}
	fmt.Printf("Latest version: %d\n", ver)
	fmt.Printf("Got version: %d\n", version)
	return tree, err
}
