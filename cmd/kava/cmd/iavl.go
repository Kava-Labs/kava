package cmd

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"strconv"
	"strings"

	"cosmossdk.io/log"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/store/wrapper"
	iavldb "github.com/cosmos/iavl/db"
	ethermintserver "github.com/evmos/ethermint/server"
	"github.com/spf13/cobra"

	"github.com/cosmos/iavl"
)

const (
	DefaultCacheSize int = 10000
)

func newIavlViewerCmd(opts ethermintserver.StartOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "iavlviewer <data|shape|versions> <prefix> [version number]",
		Short: "View iavl tree data, shape, and versions.",
		Long:  "View iavl tree data, shape, and versions.",
		RunE: func(cmd *cobra.Command, args []string) error {
			version := 0
			if len(args) == 3 {
				var err error
				version, err = strconv.Atoi(args[2])
				if err != nil {
					fmt.Fprintf(os.Stderr, "Invalid version number: %s\n", err)
					os.Exit(1)
				}
			}

			clientCtx := client.GetClientContextFromCmd(cmd)
			ctx := server.GetServerContextFromCmd(cmd)
			ctx.Config.SetRoot(clientCtx.HomeDir)

			db, err := opts.DBOpener(ctx.Viper, clientCtx.HomeDir, server.GetAppDBBackend(ctx.Viper))
			if err != nil {
				return err
			}
			defer func() {
				if err := db.Close(); err != nil {
					ctx.Logger.Error("error closing db", "error", err.Error())
				}
			}()

			cosmosdb := wrapper.NewCosmosDB(db)

			tree, err := readTree(cosmosdb, version, []byte(args[1]))
			if err != nil {
				return err
			}

			switch args[0] {
			case "data":
				printKeys(tree)
				hash := tree.Hash()
				fmt.Printf("Hash: %X\n", hash)
				fmt.Printf("Size: %X\n", tree.Size())
			case "hash":
				fmt.Printf("Hash: %X\n", tree.Hash())
			case "shape":
				printShape(tree)
			case "versions":
				printVersions(tree)
			}

			return nil
		},
	}

	return cmd
}

func PrintDBStats(db dbm.DB) {
	count := 0
	prefix := map[string]int{}
	itr, err := db.Iterator(nil, nil)
	if err != nil {
		panic(err)
	}

	defer itr.Close()
	for ; itr.Valid(); itr.Next() {
		key := itr.Key()[:1]
		prefix[string(key)]++
		count++
	}
	if err := itr.Error(); err != nil {
		panic(err)
	}
	fmt.Printf("DB contains %d entries\n", count)
	for k, v := range prefix {
		fmt.Printf("  %s: %d\n", k, v)
	}
}

// ReadTree loads an iavl tree from the directory
// If version is 0, load latest, otherwise, load named version
// The prefix represents which iavl tree you want to read. The iaviwer will always set a prefix.
func readTree(db dbm.DB, version int, prefix []byte) (*iavl.MutableTree, error) {
	if len(prefix) != 0 {
		db = dbm.NewPrefixDB(db, prefix)
	}

	iavldb := iavldb.NewWrapper(db)

	tree := iavl.NewMutableTree(iavldb, DefaultCacheSize, false, log.NewLogger(os.Stdout))
	ver, err := tree.LoadVersion(int64(version))
	if err != nil {
		return nil, err
	}
	fmt.Printf("Latest version: %d\n", ver)
	fmt.Printf("Got version: %d\n", version)
	return tree, err
}

func printKeys(tree *iavl.MutableTree) {
	fmt.Println("Printing all keys with hashed values (to detect diff)")
	tree.Iterate(func(key []byte, value []byte) bool { //nolint:errcheck
		printKey := parseWeaveKey(key)
		digest := sha256.Sum256(value)
		fmt.Printf("  %s\n    %X\n", printKey, digest)
		return false
	})
}

// parseWeaveKey assumes a separating : where all in front should be ascii,
// and all afterwards may be ascii or binary
func parseWeaveKey(key []byte) string {
	cut := bytes.IndexRune(key, ':')
	if cut == -1 {
		return encodeID(key)
	}
	prefix := key[:cut]
	id := key[cut+1:]
	return fmt.Sprintf("%s:%s", encodeID(prefix), encodeID(id))
}

// casts to a string if it is printable ascii, hex-encodes otherwise
func encodeID(id []byte) string {
	for _, b := range id {
		if b < 0x20 || b >= 0x80 {
			return strings.ToUpper(hex.EncodeToString(id))
		}
	}
	return string(id)
}

func printShape(tree *iavl.MutableTree) {
	// shape := tree.RenderShape("  ", nil)
	// TODO: handle this error
	shape, _ := tree.RenderShape("  ", nodeEncoder)
	fmt.Println(strings.Join(shape, "\n"))
}

func nodeEncoder(id []byte, depth int, isLeaf bool) string {
	prefix := fmt.Sprintf("-%d ", depth)
	if isLeaf {
		prefix = fmt.Sprintf("*%d ", depth)
	}
	if len(id) == 0 {
		return fmt.Sprintf("%s<nil>", prefix)
	}
	return fmt.Sprintf("%s%s", prefix, parseWeaveKey(id))
}

func printVersions(tree *iavl.MutableTree) {
	versions := tree.AvailableVersions()
	fmt.Println("Available versions:")
	for _, v := range versions {
		fmt.Printf("  %d\n", v)
	}
}
