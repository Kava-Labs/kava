package iavlviewer

import (
	"crypto/sha256"
	"fmt"

	"github.com/cosmos/iavl"
	ethermintserver "github.com/evmos/ethermint/server"
	"github.com/spf13/cobra"
)

func newDataCmd(opts ethermintserver.StartOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "data <prefix> [version number]",
		Short: "View all keys, hash, & size of tree.",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			prefix := args[0]
			version := 0
			if len(args) == 2 {
				var err error
				version, err = parseVersion(args[1])
				if err != nil {
					return err
				}
			}

			tree, err := openPrefixTree(opts, cmd, prefix, version)
			if err != nil {
				return err
			}

			printKeys(tree)
			hash, err := tree.Hash()
			if err != nil {
				return err
			}
			fmt.Printf("Hash: %X\n", hash)
			fmt.Printf("Size: %X\n", tree.Size())

			return nil
		},
	}

	return cmd
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
