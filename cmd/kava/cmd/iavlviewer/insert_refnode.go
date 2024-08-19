package iavlviewer

import (
	"encoding/binary"
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

func newInsertReferenceNode(opts ethermintserver.StartOptions) *cobra.Command {
	flagPruneNode := false

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

			if flagPruneNode {
				// Mark the original data as pruned (nonce = 0)
				// this allows the tree to return the correct earliest version while still keeping the data.
				var newNonce uint32 = 0
				prunedRootKeyBytes := buildNodeKey(rootNodeVersion, newNonce)
				prunedRootNodeKey := nodeKeyFormat.Key(prunedRootKeyBytes)

				fmt.Printf("Saving original root node data from %s to (%d, %d)\n", rootKey.String(), rootNodeVersion, newNonce)
				err = db.Set(prunedRootNodeKey, nodeBytes)
				if err != nil {
					return fmt.Errorf("failed save root node data to (%d, %d): %w", rootNodeVersion, newNonce, err)
				}

				fmt.Printf("Deleting old root node key: %x\n", rootNodeKey)
				err = db.Delete(rootNodeKey)
				if err != nil {
					return fmt.Errorf("failed to delete old root node: %w", err)
				}
			}

			return db.Close()
		},
	}

	cmd.Flags().BoolVar(&flagPruneNode, "prune-node", false, "prune the root node after inserting reference (saves the original data w/ nonce 0 so its version is ignored by tree, then deletes original data)")

	return cmd
}

// buildNodeKey builds the the nodeDb key for the specified version + nonce.
// this is the same logic that NodeKey.GetKey() uses on its private version + nonce variables.
func buildNodeKey(version int64, nonce uint32) []byte {
	b := make([]byte, 12)
	binary.BigEndian.PutUint64(b, uint64(version))
	binary.BigEndian.PutUint32(b[8:], nonce)
	return b
}
