package iavlviewer

import (
	"fmt"
	"strconv"

	dbm "github.com/cometbft/cometbft-db"
	"github.com/cosmos/iavl"
	"github.com/cosmos/iavl/keyformat"
	ethermintserver "github.com/evmos/ethermint/server"
	"github.com/spf13/cobra"
)

const (
	int32Size = 4
	int64Size = 8
)

var (
	nodeKeyFormat = keyformat.NewFastPrefixFormatter('s', int64Size+int32Size) // s<version><nonce>
)

func isReferenceRoot(bz []byte) (bool, int) {
	if bz[0] == nodeKeyFormat.Prefix()[0] {
		return true, len(bz)
	}
	return false, 0
}

func newInsertReferenceNode(opts ethermintserver.StartOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "insert-reference-node [prefix] [target node version] [reference key version]",
		Short: "Add a new reference node that points to another version.",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			prefix := args[0]

			rootNodeVersion, err := strconv.ParseInt(args[1], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid root node version: %w", err)
			}

			rootKeyBytes := iavl.GetRootKey(rootNodeVersion)
			rootKey := iavl.GetNodeKey(rootKeyBytes)

			rootNodeKey := nodeKeyFormat.Key(rootKeyBytes)

			newVersion, err := strconv.ParseInt(args[2], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid new version: %w", err)
			}

			db, err := openDB(opts, cmd)
			if err != nil {
				return err
			}

			// Prefix by module
			db = dbm.NewPrefixDB(db, []byte(prefix))

			fmt.Printf("looking for root node\n")
			fmt.Printf("\tRoot key: %x\n", rootKey.String())
			fmt.Printf("\tRoot node DB key: %x\n", rootNodeKey)

			fmt.Printf("---------------------\n")
			fmt.Printf("Node key format: (version, nonce) - root nodes have nonce of 1\n")

			iter, err := db.Iterator(nil, nil)
			if err != nil {
				return fmt.Errorf("failed to create iterator: %w", err)
			}
			defer iter.Close()

			// Print all nodes
			for ; iter.Valid(); iter.Next() {
				key := iter.Key()
				keyPrefix := string(key[:1])
				keyRest := key[1:]

				if keyPrefix == "s" {
					var nk []byte
					nodeKeyFormat.Scan(key, &nk)
					nodeKey := iavl.GetNodeKey(nk)

					isRef, _ := isReferenceRoot(iter.Value())
					if isRef {
						var nk []byte
						nodeKeyFormat.Scan(iter.Value(), &nk)
						referredNodeKey := iavl.GetNodeKey(nk)

						fmt.Printf(
							"[Reference Node] %s -> %s\n",
							nodeKey.String(),
							referredNodeKey.String(),
						)
						continue
					}

					fmt.Printf("[Node] %s -> %x\n", nodeKey.String(), iter.Value())
					continue
				}

				keyPrefixName := ""
				switch keyPrefix {
				case "f":
					keyPrefixName = "fastKey"
				case "m":
					keyPrefixName = "metadata"
				default:
					keyPrefixName = keyPrefix
				}

				fmt.Printf("%s: %x -> %x\n", keyPrefixName, keyRest, iter.Value())
			}

			// How to get node db, not IAVL tree
			nodeBytes, err := db.Get(rootNodeKey)
			if err != nil {
				return fmt.Errorf("failed to get root node: %w", err)
			}

			if len(nodeBytes) == 0 {
				return fmt.Errorf("root node not found at key: %s", rootKey.String())
			}

			// Print node
			fmt.Printf("Key found: %x\n", rootNodeKey)
			fmt.Printf("\t Node value: %x\n", nodeBytes)

			newRootKey := iavl.GetRootKey(newVersion)
			newRootNodeKey := nodeKeyFormat.Key(newRootKey)

			// Create a reference node from newVersion to old version
			fmt.Printf("Creating a new reference node from %d to %d\n", newVersion, rootNodeVersion)
			fmt.Printf("New key (unprefixed): %s -> %s\n", iavl.GetNodeKey(newRootKey), rootKey.String())

			err = db.Set(newRootNodeKey, rootNodeKey)
			if err != nil {
				return fmt.Errorf("failed to set reference node at version %d: %w", newVersion, err)
			}
			//
			// fmt.Printf("Deleting old root node key: %x\n", rootNodeKey)
			// err = db.Delete(rootNodeKey)
			// if err != nil {
			// 	return fmt.Errorf("failed to delete old root node: %w", err)
			// }

			return db.Close()
		},
	}

	return cmd
}
